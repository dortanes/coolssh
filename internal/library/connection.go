package library

import (
	"errors"
	"fmt"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/dortanes/coolssh/internal/secrets"
)

// Connection is everything needed to reach a host: its own settings over those it inherits.
type Connection struct {
	HostID  string
	Name    string
	Address string
	Port    int
	User    string
	// Key is nil when no key is set.
	Key *Key
	// PasswordItem names the host or folder whose saved password the host uses; empty for none.
	PasswordItem     string
	AgentForwarding  bool
	KeepaliveSeconds int
	// Jump is the host this one is reached through; nil for none.
	Jump *Connection
	// HostKey is a host key its source vouches for, in authorized_keys form; empty for none.
	HostKey string
}

// Connection resolves how hostID is reached, through its jump hosts.
func (l *Library) Connection(hostID string, defaults Settings) (Connection, error) {
	return l.connection(hostID, defaults, nil)
}

func (l *Library) connection(hostID string, defaults Settings, through []string) (Connection, error) {
	host, ok := l.Hosts[hostID]
	if !ok {
		return Connection{}, errNotAHost
	}
	if slices.Contains(through, hostID) {
		return Connection{}, errJumpLoop
	}
	address := strings.TrimSpace(host.Address)
	if address == "" {
		return Connection{}, errNoAddress
	}
	inherited, err := l.Inherited(hostID, defaults)
	if err != nil {
		return Connection{}, err
	}
	own := host.Settings
	settings := inherited.Values
	settings.take(own)
	resolved := Connection{
		HostID:           hostID,
		Name:             host.Name,
		Address:          address,
		HostKey:          host.HostKey,
		Port:             *settings.Port,
		User:             *settings.User,
		AgentForwarding:  *settings.AgentForwarding,
		KeepaliveSeconds: *settings.KeepaliveSeconds,
	}
	if *settings.Password {
		resolved.PasswordItem = hostID
		if own.Password == nil {
			resolved.PasswordItem = inherited.Origins.Password
		}
	}
	if index := slices.IndexFunc(l.Keys, func(key Key) bool { return key.ID == *settings.KeyID }); index >= 0 {
		key := l.Keys[index]
		resolved.Key = &key
	}
	if *settings.JumpHostID != "" {
		jump, err := l.connection(*settings.JumpHostID, defaults, append(through, hostID))
		if err != nil {
			return Connection{}, err
		}
		resolved.Jump = &jump
	}
	return resolved, nil
}

// AddressConnection resolves how a typed [user@]host[:port] is reached without a host in the library:
// the address's user and port over the defaults, and the default key.
func (l *Library) AddressConnection(target string, defaults Settings) (Connection, error) {
	target = strings.TrimSpace(target)
	user, address, named := strings.Cut(target, "@")
	if !named {
		user, address = "", target
	}
	var port *int
	if host, number, err := net.SplitHostPort(address); err == nil {
		parsed, err := strconv.Atoi(number)
		if err != nil || parsed < minPort || parsed > maxPort {
			return Connection{}, errPortRange
		}
		address, port = host, &parsed
	}
	address = strings.Trim(address, "[]")
	if address == "" || strings.ContainsAny(address, " /@") {
		return Connection{}, errNoAddress
	}
	var own Settings
	own.Port = port
	if user != "" {
		own.User = &user
	}
	var inherited Inherited
	inherited.take(own, "")
	l.fallBack(&inherited, defaults)
	settings := inherited.Values
	resolved := Connection{
		Name:             target,
		Address:          address,
		Port:             *settings.Port,
		User:             *settings.User,
		AgentForwarding:  *settings.AgentForwarding,
		KeepaliveSeconds: *settings.KeepaliveSeconds,
	}
	if settings.KeyID != nil {
		if index := slices.IndexFunc(l.Keys, func(key Key) bool { return key.ID == *settings.KeyID }); index >= 0 {
			key := l.Keys[index]
			resolved.Key = &key
		}
	}
	return resolved, nil
}

// take overrides every setting own sets.
func (s *Settings) take(own Settings) {
	override(&s.User, own.User)
	override(&s.Port, own.Port)
	override(&s.KeyID, own.KeyID)
	override(&s.Password, own.Password)
	override(&s.JumpHostID, own.JumpHostID)
	override(&s.AgentForwarding, own.AgentForwarding)
	override(&s.KeepaliveSeconds, own.KeepaliveSeconds)
}

func override[T any](value **T, own *T) {
	if own != nil {
		*value = clonePointer(own)
	}
}

// Connection resolves how hostID is reached, through its jump hosts.
func (s *Store) Connection(hostID string) (Connection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.library.Connection(hostID, s.defaults)
}

// AddressConnection resolves how a typed [user@]host[:port] is reached without a host in the library.
func (s *Store) AddressConnection(target string) (Connection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.library.AddressConnection(target, s.defaults)
}

// Password returns the password saved for a host or folder.
func (s *Store) Password(item string) (string, error) {
	return s.secrets.Get(passwordSecret(item))
}

// KeyText returns a key's private text: its file's content, or the pasted text in the keyring.
func (s *Store) KeyText(key Key) ([]byte, error) {
	if key.Source == FileKey {
		text, err := os.ReadFile(key.Path)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", errUnreadableKey, err)
		}
		return text, nil
	}
	text, err := s.secrets.Get(keySecret(key.ID))
	return []byte(text), err
}

// Passphrase returns a key's saved passphrase; false when none is saved.
func (s *Store) Passphrase(keyID string) (string, bool, error) {
	passphrase, err := s.secrets.Get(passphraseSecret(keyID))
	if errors.Is(err, secrets.ErrNotFound) {
		return "", false, nil
	}
	return passphrase, err == nil, err
}
