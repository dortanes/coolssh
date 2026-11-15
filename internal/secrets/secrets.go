// Package secrets keeps the app's secrets in the operating system's keyring: the Keychain on macOS,
// the Credential Manager on Windows, the Secret Service on Linux.
package secrets

import (
	"errors"
	"sync"

	"github.com/zalando/go-keyring"
)

// ErrNotFound reports that no secret is stored under a name.
var ErrNotFound = errors.New("no secret under this name")

// Secrets stores secrets by name.
type Secrets interface {
	Set(name, secret string) error
	// Get returns ErrNotFound when nothing is stored under name.
	Get(name string) (string, error)
	// Delete succeeds when nothing is stored under name.
	Delete(name string) error
}

// Keyring keeps secrets in the operating system's keyring under one service name.
type Keyring struct {
	service string
}

// NewKeyring keeps secrets under service.
func NewKeyring(service string) Keyring {
	return Keyring{service: service}
}

func (k Keyring) Set(name, secret string) error {
	return keyring.Set(k.service, name, secret)
}

func (k Keyring) Get(name string) (string, error) {
	secret, err := keyring.Get(k.service, name)
	if errors.Is(err, keyring.ErrNotFound) {
		return "", ErrNotFound
	}
	return secret, err
}

func (k Keyring) Delete(name string) error {
	err := keyring.Delete(k.service, name)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}

// Memory keeps secrets in memory, for tests. It is safe for concurrent use.
type Memory struct {
	mu      sync.Mutex
	secrets map[string]string
}

// NewMemory is an empty Memory.
func NewMemory() *Memory {
	return &Memory{secrets: map[string]string{}}
}

func (m *Memory) Set(name, secret string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.secrets[name] = secret
	return nil
}

func (m *Memory) Get(name string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	secret, ok := m.secrets[name]
	if !ok {
		return "", ErrNotFound
	}
	return secret, nil
}

func (m *Memory) Delete(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.secrets, name)
	return nil
}
