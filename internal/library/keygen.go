package library

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"encoding/pem"
	"errors"
	"os"
	"slices"
	"strings"

	"golang.org/x/crypto/ssh"
)

// KeyKind is the algorithm of a key the app generates.
type KeyKind string

const (
	Ed25519Kind KeyKind = "ed25519"
	RSAKind     KeyKind = "rsa"
)

// rsaBits is the size of a generated RSA key, the size OpenSSH's ssh-keygen picks by default.
const rsaBits = 3072

// generateKey returns a new private key of kind in OpenSSH's format, encrypted with passphrase
// when one is given.
func generateKey(kind KeyKind, comment, passphrase string) (string, error) {
	var private crypto.PrivateKey
	switch kind {
	case Ed25519Kind:
		_, key, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return "", err
		}
		private = key
	case RSAKind:
		key, err := rsa.GenerateKey(rand.Reader, rsaBits)
		if err != nil {
			return "", err
		}
		private = key
	default:
		return "", errUnknownKeyKind
	}
	var block *pem.Block
	var err error
	if passphrase == "" {
		block, err = ssh.MarshalPrivateKey(private, comment)
	} else {
		block, err = ssh.MarshalPrivateKeyWithPassphrase(private, comment, []byte(passphrase))
	}
	if err != nil {
		return "", err
	}
	return string(pem.EncodeToMemory(block)), nil
}

// GenerateKey makes a new key of kind named name, kept in the keyring like a pasted one, and
// returns its id.
func (s *Store) GenerateKey(name string, kind KeyKind, passphrase string) (string, Library, error) {
	text, err := generateKey(kind, strings.TrimSpace(name), passphrase)
	if err != nil {
		return "", Library{}, err
	}
	return s.AddKeyText(name, text, passphrase)
}

// PublicKey returns a key's authorized_keys line without a comment, from the .pub file beside a key
// file when there is one, else from the private key, opened with its saved passphrase when needed.
func (s *Store) PublicKey(id string) (string, error) {
	s.mu.Lock()
	index := slices.IndexFunc(s.library.Keys, func(key Key) bool { return key.ID == id })
	var key Key
	if index >= 0 {
		key = s.library.Keys[index]
	}
	s.mu.Unlock()
	if index < 0 {
		return "", errUnknownKey
	}
	if key.Source == FileKey {
		if text, err := os.ReadFile(key.Path + ".pub"); err == nil {
			if public, _, _, _, err := ssh.ParseAuthorizedKey(text); err == nil {
				return authorizedLine(public), nil
			}
		}
	}
	text, err := s.KeyText(key)
	if err != nil {
		return "", err
	}
	public, err := s.publicOf(key, text)
	if err != nil {
		return "", err
	}
	return authorizedLine(public), nil
}

// HostPublicKey returns the public half of the key a host uses, as PublicKey does; a host without a
// key is refused.
func (s *Store) HostPublicKey(hostID string) (string, error) {
	connection, err := s.Connection(hostID)
	if err != nil {
		return "", err
	}
	if connection.Key == nil {
		return "", errNoKey
	}
	return s.PublicKey(connection.Key.ID)
}

func (s *Store) publicOf(key Key, text []byte) (ssh.PublicKey, error) {
	signer, err := ssh.ParsePrivateKey(text)
	var missing *ssh.PassphraseMissingError
	switch {
	case err == nil:
		return signer.PublicKey(), nil
	case !errors.As(err, &missing):
		return nil, errNotAKey
	case missing.PublicKey != nil:
		return missing.PublicKey, nil
	}
	passphrase, saved, err := s.Passphrase(key.ID)
	if err != nil || !saved {
		return nil, errNoPublicKey
	}
	signer, err = ssh.ParsePrivateKeyWithPassphrase(text, []byte(passphrase))
	if err != nil {
		return nil, errNoPublicKey
	}
	return signer.PublicKey(), nil
}

func authorizedLine(public ssh.PublicKey) string {
	return strings.TrimSpace(string(ssh.MarshalAuthorizedKey(public)))
}
