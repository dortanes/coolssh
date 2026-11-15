package library

import (
	"errors"
	"slices"

	"golang.org/x/crypto/ssh"
)

// KeySource is where a key's text lives.
type KeySource string

const (
	// FileKey is read from its file each time it is used; the library keeps the path.
	FileKey KeySource = "file"
	// StoredKey was pasted and lives in the keyring.
	StoredKey KeySource = "stored"
)

// Key is a private key hosts and folders can use. Its text is never part of the library.
type Key struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Fingerprint string    `json:"fingerprint"`
	Source      KeySource `json:"source"`
	Path        string    `json:"path,omitempty"`
	// Encrypted keys need their passphrase to be used.
	Encrypted bool `json:"encrypted"`
}

// inspectKey reads a key's type, SHA-256 fingerprint and encryption, opening it with passphrase when
// given; without one, an encrypted key's format alone tells them, empty where it does not.
func inspectKey(text []byte, passphrase string) (Key, error) {
	raw, err := ssh.ParseRawPrivateKey(text)
	var missing *ssh.PassphraseMissingError
	switch {
	case errors.As(err, &missing) && passphrase == "":
		return describe(Key{Encrypted: true}, missing.PublicKey), nil
	case errors.As(err, &missing):
		raw, err = ssh.ParseRawPrivateKeyWithPassphrase(text, []byte(passphrase))
		if err != nil {
			return Key{}, errWrongPassphrase
		}
		return signerKey(Key{Encrypted: true}, raw)
	case err != nil:
		return Key{}, errNotAKey
	}
	return signerKey(Key{}, raw)
}

func signerKey(key Key, raw any) (Key, error) {
	signer, err := ssh.NewSignerFromKey(raw)
	if err != nil {
		return Key{}, errNotAKey
	}
	return describe(key, signer.PublicKey()), nil
}

func describe(key Key, public ssh.PublicKey) Key {
	if public != nil {
		key.Type = public.Type()
		key.Fingerprint = ssh.FingerprintSHA256(public)
	}
	return key
}

// AddKey puts key at the end of the keys under a new id and returns it.
func (l *Library) AddKey(key Key) (string, error) {
	name, err := checkName(key.Name)
	if err != nil {
		return "", err
	}
	key.Name = name
	key.ID = newID()
	l.Keys = append(l.Keys, key)
	return key.ID, nil
}

// RenameKey gives a key a new name.
func (l *Library) RenameKey(id, name string) error {
	name, err := checkName(name)
	if err != nil {
		return err
	}
	index := slices.IndexFunc(l.Keys, func(key Key) bool { return key.ID == id })
	if index < 0 {
		return errUnknownKey
	}
	l.Keys[index].Name = name
	return nil
}

// SetDefaultKey makes a key the one hosts use when neither they nor their folders set one; empty
// sets none.
func (l *Library) SetDefaultKey(id string) error {
	if id != "" && !l.hasKey(id) {
		return errUnknownKey
	}
	l.DefaultKeyID = id
	return nil
}

func (l *Library) hasKey(id string) bool {
	return slices.ContainsFunc(l.Keys, func(key Key) bool { return key.ID == id })
}

// DeleteKey removes a key; hosts and folders that used it, and the library when it was the
// default, inherit their key again.
func (l *Library) DeleteKey(id string) error {
	index := slices.IndexFunc(l.Keys, func(key Key) bool { return key.ID == id })
	if index < 0 {
		return errUnknownKey
	}
	l.Keys = slices.Delete(l.Keys, index, index+1)
	if l.DefaultKeyID == id {
		l.DefaultKeyID = ""
	}
	l.forEachSettings(func(settings *Settings) {
		if settings.KeyID != nil && *settings.KeyID == id {
			settings.KeyID = nil
		}
	})
	return nil
}
