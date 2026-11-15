package session

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/dortanes/coolssh/internal/library"
	"github.com/dortanes/coolssh/internal/secrets"
)

// testLibrary answers connections and secrets from maps.
type testLibrary struct {
	connections map[string]library.Connection
	addresses   map[string]library.Connection
	passwords   map[string]string
	keys        map[string][]byte
	passphrases map[string]string
}

func (l *testLibrary) Connection(hostID string) (library.Connection, error) {
	connection, ok := l.connections[hostID]
	if !ok {
		return library.Connection{}, library.Refusal("library_no_address")
	}
	return connection, nil
}

func (l *testLibrary) AddressConnection(target string) (library.Connection, error) {
	connection, ok := l.addresses[target]
	if !ok {
		return library.Connection{}, library.Refusal("library_no_address")
	}
	return connection, nil
}

func (l *testLibrary) Password(item string) (string, error) {
	password, ok := l.passwords[item]
	if !ok {
		return "", secrets.ErrNotFound
	}
	return password, nil
}

func (l *testLibrary) KeyText(key library.Key) ([]byte, error) {
	return l.keys[key.ID], nil
}

func (l *testLibrary) Passphrase(keyID string) (string, bool, error) {
	passphrase, ok := l.passphrases[keyID]
	return passphrase, ok, nil
}

// remembered records the passwords sessions asked to save.
type remembered struct {
	mu        sync.Mutex
	passwords map[string]string
}

func (r *remembered) save(hostID, password string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.passwords[hostID] = password
}

type fixture struct {
	library    *testLibrary
	sessions   *Sessions
	knownHosts string
	remembered *remembered
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	folder := t.TempDir()
	f := &fixture{
		library:    &testLibrary{connections: map[string]library.Connection{}, addresses: map[string]library.Connection{}, passwords: map[string]string{}, keys: map[string][]byte{}, passphrases: map[string]string{}},
		knownHosts: filepath.Join(folder, "known_hosts"),
		remembered: &remembered{passwords: map[string]string{}},
	}
	f.sessions = New(Options{
		Library:       f.library,
		KnownHosts:    NewKnownHosts(filepath.Join(folder, "user_known_hosts"), f.knownHosts),
		Agent:         func() (Agent, error) { return nil, nil },
		Remember:      f.remembered.save,
		Word:          func(p *Problem) string { return p.Code },
		KeepaliveUnit: 20 * time.Millisecond,
	})
	t.Cleanup(f.sessions.CloseAll)
	return f
}

// host adds a host reaching server as user tester.
func (f *fixture) host(id string, server *testServer) library.Connection {
	connection := library.Connection{HostID: id, Name: id, Address: "127.0.0.1", Port: server.port(), User: "tester"}
	f.library.connections[id] = connection
	return connection
}

// trust adds server's key to the app's known hosts.
func (f *fixture) trust(t *testing.T, server *testServer) {
	t.Helper()
	if err := NewKnownHosts("", f.knownHosts).trust(server.address(), server.hostKey.PublicKey()); err != nil {
		t.Fatal(err)
	}
}

// openWeb opens a session to host web, on a trusted server that takes the password "secret".
func (f *fixture) openWeb(t *testing.T) string {
	t.Helper()
	server := newServer(t, acceptPassword("secret"))
	f.trust(t, server)
	f.host("web", server)
	return f.open(t, "web")
}

func (f *fixture) open(t *testing.T, hostID string) string {
	t.Helper()
	id, err := f.sessions.Open(hostID)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

// await waits until the session's state satisfies done, and returns it.
func (f *fixture) await(t *testing.T, id string, what string, done func(State) bool) State {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		for _, state := range f.sessions.States() {
			if state.ID == id && done(state) {
				return state
			}
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("the session never got %s: %+v", what, f.sessions.States())
	return State{}
}

func (f *fixture) awaitPrompt(t *testing.T, id string, kind PromptKind) Prompt {
	t.Helper()
	state := f.await(t, id, "a "+string(kind)+" prompt", func(s State) bool { return s.Prompt != nil && s.Prompt.Kind == kind })
	return *state.Prompt
}

func (f *fixture) answer(t *testing.T, id string, prompt Prompt, answer Answer) {
	t.Helper()
	if err := f.sessions.Answer(id, prompt.ID, answer); err != nil {
		t.Fatal(err)
	}
}

func connected(s State) bool { return s.Status == Connected }
func closed(s State) bool    { return s.Status == Closed }

func TestATypedPasswordIsSavedOnceTheServerAcceptsIt(t *testing.T) {
	f := newFixture(t)
	id := f.openWeb(t)
	prompt := f.awaitPrompt(t, id, PasswordPrompt)
	f.answer(t, id, prompt, Answer{Accept: true, Values: []string{"wrong"}, Remember: true})
	prompt = f.awaitPrompt(t, id, PasswordPrompt)
	if !prompt.Retry {
		t.Error("the second password prompt does not say the first one failed")
	}
	f.answer(t, id, prompt, Answer{Accept: true, Values: []string{"secret"}, Remember: true})
	f.await(t, id, "connected", connected)
	if got := f.remembered.passwords["web"]; got != "secret" {
		t.Errorf("remembered %q, want the accepted password", got)
	}
}

func TestAPasswordForATypedAddressIsNeverSaved(t *testing.T) {
	f := newFixture(t)
	server := newServer(t, acceptPassword("secret"))
	f.trust(t, server)
	target := "tester@" + server.address()
	f.library.addresses[target] = library.Connection{Name: target, Address: "127.0.0.1", Port: server.port(), User: "tester"}
	id, err := f.sessions.OpenAddress(target)
	if err != nil {
		t.Fatal(err)
	}
	prompt := f.awaitPrompt(t, id, PasswordPrompt)
	if prompt.Savable {
		t.Error("the prompt offers to save the password")
	}
	f.answer(t, id, prompt, Answer{Accept: true, Values: []string{"secret"}, Remember: true})
	state := f.await(t, id, "connected", connected)
	if state.Title != target || state.HostID != "" || len(f.remembered.passwords) != 0 {
		t.Errorf("state %+v, remembered %v", state, f.remembered.passwords)
	}
}

func TestASavedPasswordIsUsedWithoutAsking(t *testing.T) {
	f := newFixture(t)
	server := newServer(t, acceptPassword("secret"))
	f.trust(t, server)
	connection := f.host("web", server)
	connection.PasswordItem = "folder"
	f.library.connections["web"] = connection
	f.library.passwords["folder"] = "secret"
	id := f.open(t, "web")
	state := f.await(t, id, "connected", func(s State) bool { return s.Status == Connected || s.Prompt != nil })
	if state.Prompt != nil {
		t.Fatalf("asked %s with a saved password", state.Prompt.Kind)
	}
}

func TestAnUnknownHostKeyIsTrustedOnlyOnceAccepted(t *testing.T) {
	f := newFixture(t)
	server := newServer(t, acceptPassword("secret"))
	f.host("web", server)
	f.library.connections["web"] = withPassword(f.library.connections["web"], f.library, "secret")
	id := f.open(t, "web")
	prompt := f.awaitPrompt(t, id, HostKeyPrompt)
	if prompt.Fingerprint != ssh.FingerprintSHA256(server.hostKey.PublicKey()) {
		t.Errorf("fingerprint %q, want the server's", prompt.Fingerprint)
	}
	f.answer(t, id, prompt, Answer{Accept: true})
	f.await(t, id, "connected", connected)
	again := f.open(t, "web")
	state := f.await(t, again, "connected", func(s State) bool { return s.Status == Connected || s.Prompt != nil })
	if state.Prompt != nil {
		t.Errorf("asked %s about a trusted key", state.Prompt.Kind)
	}
}

func TestAHostKeyTheSourceVouchesForIsTrustedWithoutAsking(t *testing.T) {
	f := newFixture(t)
	server := newServer(t, acceptPassword("secret"))
	connection := withPassword(f.host("web", server), f.library, "secret")
	connection.HostKey = string(ssh.MarshalAuthorizedKey(server.hostKey.PublicKey()))
	f.library.connections["web"] = connection
	id := f.open(t, "web")
	state := f.await(t, id, "connected", func(s State) bool { return s.Status == Connected || s.Prompt != nil })
	if state.Prompt != nil {
		t.Errorf("asked %s about a vouched key", state.Prompt.Kind)
	}
	if _, err := os.Stat(f.knownHosts); !errors.Is(err, os.ErrNotExist) {
		t.Error("a vouched key was written to the known hosts")
	}
}

func TestADeclinedHostKeyClosesTheSessionWithoutAProblem(t *testing.T) {
	f := newFixture(t)
	server := newServer(t, acceptPassword("secret"))
	f.host("web", server)
	id := f.open(t, "web")
	f.answer(t, id, f.awaitPrompt(t, id, HostKeyPrompt), Answer{Accept: false})
	if state := f.await(t, id, "closed", closed); state.Problem != "" {
		t.Errorf("problem %q, want none", state.Problem)
	}
	if _, err := os.Stat(f.knownHosts); !errors.Is(err, os.ErrNotExist) {
		t.Error("a declined key was written to the known hosts")
	}
}

func TestAChangedHostKeyStopsTheConnection(t *testing.T) {
	f := newFixture(t)
	server := newServer(t, acceptPassword("secret"))
	if err := NewKnownHosts("", f.knownHosts).trust(server.address(), newSigner(t).PublicKey()); err != nil {
		t.Fatal(err)
	}
	f.host("web", server)
	id := f.open(t, "web")
	if state := f.await(t, id, "closed", closed); state.Problem != "session_host_key_changed" {
		t.Errorf("problem %q, want the changed key", state.Problem)
	}
}

func TestAnEncryptedKeyAsksForItsPassphrase(t *testing.T) {
	f := newFixture(t)
	_, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	block, err := ssh.MarshalPrivateKeyWithPassphrase(private, "", []byte("open sesame"))
	if err != nil {
		t.Fatal(err)
	}
	public, err := ssh.NewPublicKey(private.Public())
	if err != nil {
		t.Fatal(err)
	}
	server := newServer(t, acceptKey(public))
	f.trust(t, server)
	connection := f.host("web", server)
	connection.Key = &library.Key{ID: "key", Name: "laptop", Encrypted: true}
	f.library.connections["web"] = connection
	f.library.keys["key"] = pem.EncodeToMemory(block)
	id := f.open(t, "web")
	prompt := f.awaitPrompt(t, id, PassphrasePrompt)
	if prompt.Key != "laptop" {
		t.Errorf("asked for %q's passphrase, want laptop's", prompt.Key)
	}
	f.answer(t, id, prompt, Answer{Accept: true, Values: []string{"open sesame"}})
	f.await(t, id, "connected", connected)
}

func TestAServerAcceptingNothingFailsAuthentication(t *testing.T) {
	f := newFixture(t)
	server := newServer(t, acceptKey(newSigner(t).PublicKey()))
	f.trust(t, server)
	f.host("web", server)
	id := f.open(t, "web")
	if state := f.await(t, id, "closed", closed); state.Problem != "session_authentication_failed" {
		t.Errorf("problem %q, want failed authentication", state.Problem)
	}
}

func TestASessionIsReachedThroughItsJumpHost(t *testing.T) {
	f := newFixture(t)
	jump := newServer(t, acceptPassword("jump"))
	target := newServer(t, acceptPassword("target"))
	f.trust(t, jump)
	f.trust(t, target)
	f.host("bastion", jump)
	connection := f.host("db", target)
	bastion := withPassword(f.library.connections["bastion"], f.library, "jump")
	connection.Jump = &bastion
	f.library.connections["db"] = withPassword(connection, f.library, "target")
	id := f.open(t, "db")
	f.await(t, id, "connected", connected)
}

func TestOutputIsHeldUntilAWindowShowsIt(t *testing.T) {
	f := newFixture(t)
	server := newServer(t, acceptPassword("secret"))
	f.trust(t, server)
	f.host("web", server)
	f.library.connections["web"] = withPassword(f.library.connections["web"], f.library, "secret")
	id := f.open(t, "web")
	f.await(t, id, "connected", connected)
	window := newTestWindow()
	go func() { _ = f.sessions.Show(id, window) }()
	window.expect(t, "ready")
	window.input <- []byte("hello")
	window.expect(t, "hello")
	window.input <- []byte("exit")
	if state := f.await(t, id, "closed", closed); state.Problem != "" {
		t.Errorf("problem %q after exit, want none", state.Problem)
	}
}

func TestAHostThatStopsAnsweringIsGivenUp(t *testing.T) {
	f := newFixture(t)
	server := newServer(t, acceptPassword("secret"))
	server.silent = true
	f.trust(t, server)
	connection := withPassword(f.host("web", server), f.library, "secret")
	connection.KeepaliveSeconds = 1
	f.library.connections["web"] = connection
	id := f.open(t, "web")
	if state := f.await(t, id, "closed", closed); state.Problem != "session_not_answering" {
		t.Errorf("problem %q, want not answering", state.Problem)
	}
}

func TestClosingASessionStopsItsPrompt(t *testing.T) {
	f := newFixture(t)
	id := f.openWeb(t)
	f.awaitPrompt(t, id, PasswordPrompt)
	if err := f.sessions.Close(id); err != nil {
		t.Fatal(err)
	}
	if states := f.sessions.States(); len(states) != 0 {
		t.Errorf("sessions after closing: %+v", states)
	}
}

// openClosed opens a session to web and closes it by declining its password prompt.
func (f *fixture) openClosed(t *testing.T) string {
	t.Helper()
	id := f.openWeb(t)
	f.answer(t, id, f.awaitPrompt(t, id, PasswordPrompt), Answer{Accept: false})
	f.await(t, id, "closed", closed)
	return id
}

func TestAClosedSessionReconnectsUnderTheSameID(t *testing.T) {
	f := newFixture(t)
	id := f.openClosed(t)
	if err := f.sessions.Reconnect(id); err != nil {
		t.Fatal(err)
	}
	f.answer(t, id, f.awaitPrompt(t, id, PasswordPrompt), Answer{Accept: true, Values: []string{"secret"}})
	f.await(t, id, "connected", connected)
}

func TestReconnectingTwiceAtOnceConnectsOnce(t *testing.T) {
	f := newFixture(t)
	session, err := f.sessions.find(f.openClosed(t))
	if err != nil {
		t.Fatal(err)
	}
	var started atomic.Int32
	var clicks sync.WaitGroup
	for range 8 {
		clicks.Go(func() {
			if session.restart() {
				started.Add(1)
			}
		})
	}
	clicks.Wait()
	if got := started.Load(); got != 1 {
		t.Errorf("%d connections started, want 1", got)
	}
}

// withPassword saves password for connection's host.
func withPassword(connection library.Connection, l *testLibrary, password string) library.Connection {
	connection.PasswordItem = connection.HostID
	l.passwords[connection.HostID] = password
	return connection
}

// testWindow is a window's terminal stream in memory.
type testWindow struct {
	output chan []byte
	input  chan []byte
	seen   []byte
}

func newTestWindow() *testWindow {
	return &testWindow{output: make(chan []byte, 64), input: make(chan []byte, 64)}
}

func (w *testWindow) Send(frame []byte) error {
	w.output <- frame
	return nil
}

func (w *testWindow) Receive() ([]byte, error) {
	return <-w.input, nil
}

func (w *testWindow) expect(t *testing.T, text string) {
	t.Helper()
	deadline := time.After(5 * time.Second)
	for !slices.Contains(strings.Split(string(w.seen), "\r\n"), text) && !strings.Contains(string(w.seen), text) {
		select {
		case frame := <-w.output:
			w.seen = append(w.seen, frame...)
		case <-deadline:
			t.Fatalf("the window never showed %q; it showed %q", text, w.seen)
		}
	}
}

func TestATaskClosesWithItsOutcome(t *testing.T) {
	for _, c := range []struct {
		task Task
		want string
	}{
		{InstallKeyTask("ssh-ed25519 AAAA"), "session_key_installed"},
		{Task{Command: "false", Done: "session_key_installed", Failed: "session_key_not_installed"}, "session_key_not_installed"},
	} {
		f := newFixture(t)
		server := newServer(t, acceptPassword("secret"))
		f.trust(t, server)
		f.host("web", server)
		id, err := f.sessions.OpenTask("web", c.task)
		if err != nil {
			t.Fatal(err)
		}
		f.answer(t, id, f.awaitPrompt(t, id, PasswordPrompt), Answer{Accept: true, Values: []string{"secret"}})
		state := f.await(t, id, "closed", closed)
		if state.Outcome != c.want || state.Problem != "" {
			t.Errorf("%s: outcome %q, problem %q", c.task.Command, state.Outcome, state.Problem)
		}
	}
}
