// Package session connects to hosts over SSH: authentication, host keys, jump hosts, the agent,
// keepalive, and the terminal each session runs.
package session

import (
	"crypto/rand"
	"errors"
	"slices"
	"sync"
	"time"

	"github.com/dortanes/coolssh/internal/library"
)

// errUnknownSession reports a session id that is not open.
var errUnknownSession = errors.New("no such session")

// Library is what sessions read from the host library: how a host is reached, and its secrets.
type Library interface {
	Connection(hostID string) (library.Connection, error)
	AddressConnection(target string) (library.Connection, error)
	Password(item string) (string, error)
	KeyText(key library.Key) ([]byte, error)
	Passphrase(keyID string) (string, bool, error)
}

// Options configures Sessions.
type Options struct {
	Library    Library
	KnownHosts *KnownHosts
	// Agent connects to the user's SSH agent for one session; nil when none runs.
	Agent func() (Agent, error)
	// Changed is called after any session changes, with no lock held.
	Changed func()
	// Remember saves a password the user typed for a host and asked to keep, once it was accepted.
	Remember func(hostID, password string)
	// Word turns a problem into the sentence a window shows.
	Word func(*Problem) string
	// KeepaliveUnit is what a host's keepalive setting counts; a second outside tests.
	KeepaliveUnit time.Duration
}

// Sessions are the open sessions, in the order they opened. It is safe for concurrent use.
type Sessions struct {
	options Options

	mu       sync.Mutex
	sessions []*Session
	prompts  int
}

// New holds no sessions until Open.
func New(options Options) *Sessions {
	if options.KeepaliveUnit == 0 {
		options.KeepaliveUnit = time.Second
	}
	return &Sessions{options: options}
}

// Open starts a session to a host and returns its id. A host that cannot be reached as configured
// is refused before any session opens.
func (s *Sessions) Open(hostID string) (string, error) {
	return s.open(hostID, func() (library.Connection, error) { return s.options.Library.Connection(hostID) }, nil)
}

// OpenAddress starts a session to a typed [user@]host[:port] that is not in the library and returns
// its id.
func (s *Sessions) OpenAddress(target string) (string, error) {
	return s.open("", func() (library.Connection, error) { return s.options.Library.AddressConnection(target) }, nil)
}

// OpenTask starts a session to a host that runs task instead of a shell, and returns its id.
func (s *Sessions) OpenTask(hostID string, task Task) (string, error) {
	return s.open(hostID, func() (library.Connection, error) { return s.options.Library.Connection(hostID) }, &task)
}

// open starts a session reached through resolve, called again on every reconnect, running task
// when given.
func (s *Sessions) open(hostID string, resolve func() (library.Connection, error), task *Task) (string, error) {
	connection, err := resolve()
	if err != nil {
		return "", err
	}
	session := &Session{
		id:      rand.Text(),
		hostID:  hostID,
		resolve: resolve,
		task:    task,
		owner:   s,
		title:   connection.Name,
		columns: defaultColumns,
		rows:    defaultRows,
	}
	s.mu.Lock()
	s.sessions = append(s.sessions, session)
	s.mu.Unlock()
	session.start()
	s.changed()
	return session.id, nil
}

// States are what the windows show of every session, in order.
func (s *Sessions) States() []State {
	s.mu.Lock()
	sessions := slices.Clone(s.sessions)
	s.mu.Unlock()
	states := make([]State, len(sessions))
	for i, session := range sessions {
		states[i] = session.state()
	}
	return states
}

// Connected counts the sessions whose shell is running.
func (s *Sessions) Connected() int {
	count := 0
	for _, state := range s.States() {
		if state.Status == Connected {
			count++
		}
	}
	return count
}

func (s *Sessions) find(id string) (*Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	index := slices.IndexFunc(s.sessions, func(session *Session) bool { return session.id == id })
	if index < 0 {
		return nil, errUnknownSession
	}
	return s.sessions[index], nil
}

// Answer replies to a session's prompt; an answer to a prompt no longer shown is ignored.
func (s *Sessions) Answer(id string, prompt int, answer Answer) error {
	session, err := s.find(id)
	if err != nil {
		return err
	}
	if session.answer(prompt, answer) {
		s.changed()
	}
	return nil
}

// Resize gives a session's terminal a size in columns and rows.
func (s *Sessions) Resize(id string, columns, rows int) error {
	if columns < 1 || rows < 1 {
		return nil
	}
	session, err := s.find(id)
	if err != nil {
		return err
	}
	session.resize(columns, rows)
	return nil
}

// Reconnect connects a closed session again; a session still open is left as it is.
func (s *Sessions) Reconnect(id string) error {
	session, err := s.find(id)
	if err != nil {
		return err
	}
	if session.restart() {
		s.changed()
	}
	return nil
}

// Close ends a session and forgets it.
func (s *Sessions) Close(id string) error {
	session, err := s.find(id)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.sessions = slices.DeleteFunc(s.sessions, func(open *Session) bool { return open == session })
	s.mu.Unlock()
	session.stop()
	s.changed()
	return nil
}

// CloseAll ends every session.
func (s *Sessions) CloseAll() {
	s.mu.Lock()
	sessions := s.sessions
	s.sessions = nil
	s.mu.Unlock()
	for _, session := range sessions {
		session.stop()
	}
}

// Show sends a session's output to window and its input to the session until the window goes
// away; a window shown later takes the place of this one.
func (s *Sessions) Show(id string, window Window) error {
	session, err := s.find(id)
	if err != nil {
		return err
	}
	session.show(window)
	return nil
}

func (s *Sessions) nextPrompt() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.prompts++
	return s.prompts
}

func (s *Sessions) changed() {
	if s.options.Changed != nil {
		s.options.Changed()
	}
}
