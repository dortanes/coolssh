package library

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"errors"
	"testing"

	"golang.org/x/crypto/ssh"
)

// newKey returns an OpenSSH private key, encrypted when passphrase is not empty, with its
// public key's fingerprint.
func newKey(t *testing.T, passphrase string) (text []byte, fingerprint string) {
	t.Helper()
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	var block *pem.Block
	if passphrase == "" {
		block, err = ssh.MarshalPrivateKey(private, "")
	} else {
		block, err = ssh.MarshalPrivateKeyWithPassphrase(private, "", []byte(passphrase))
	}
	if err != nil {
		t.Fatal(err)
	}
	sshPublic, err := ssh.NewPublicKey(public)
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(block), ssh.FingerprintSHA256(sshPublic)
}

func TestInspectReadsTypeFingerprintAndEncryption(t *testing.T) {
	plain, plainFingerprint := newKey(t, "")
	encrypted, encryptedFingerprint := newKey(t, "correct horse")
	tests := []struct {
		name       string
		text       []byte
		passphrase string
		want       Key
		err        error
	}{
		{"a plain key", plain, "", Key{Type: ssh.KeyAlgoED25519, Fingerprint: plainFingerprint}, nil},
		{"an encrypted key without its passphrase", encrypted, "", Key{Type: ssh.KeyAlgoED25519, Fingerprint: encryptedFingerprint, Encrypted: true}, nil},
		{"an encrypted key with its passphrase", encrypted, "correct horse", Key{Type: ssh.KeyAlgoED25519, Fingerprint: encryptedFingerprint, Encrypted: true}, nil},
		{"a wrong passphrase", encrypted, "wrong", Key{}, errWrongPassphrase},
		{"not a key", []byte("hello"), "", Key{}, errNotAKey},
		{"a public key", []byte("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIB"), "", Key{}, errNotAKey},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := inspectKey(test.text, test.passphrase)
			if !errors.Is(err, test.err) {
				t.Fatalf("err = %v, want %v", err, test.err)
			}
			if got != test.want {
				t.Errorf("key = %+v, want %+v", got, test.want)
			}
		})
	}
}

func TestDeletingAKeyLetsItsUsersInheritAgain(t *testing.T) {
	l, ids := tree(t)
	id, err := l.AddKey(Key{Name: "work", Type: ssh.KeyAlgoED25519})
	if err != nil {
		t.Fatal(err)
	}
	if err := l.SetSettings(ids["web"], Settings{KeyID: &id}); err != nil {
		t.Fatal(err)
	}
	if err := l.DeleteKey(id); err != nil {
		t.Fatal(err)
	}
	if len(l.Keys) != 0 || l.Hosts[ids["web"]].Settings.KeyID != nil {
		t.Errorf("keys %v, web's key %v", l.Keys, l.Hosts[ids["web"]].Settings.KeyID)
	}
}
