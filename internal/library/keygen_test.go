package library

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestAGeneratedKeyIsStoredAndGivesItsPublicKey(t *testing.T) {
	for _, c := range []struct {
		kind       KeyKind
		passphrase string
		keyType    string
	}{
		{Ed25519Kind, "", ssh.KeyAlgoED25519},
		{Ed25519Kind, "correct horse", ssh.KeyAlgoED25519},
		{RSAKind, "", ssh.KeyAlgoRSA},
	} {
		store, _, _ := openStore(t)
		id, library, err := store.GenerateKey("laptop", c.kind, c.passphrase)
		if err != nil {
			t.Fatalf("%s: %v", c.kind, err)
		}
		key := library.Keys[0]
		if key.ID != id || key.Type != c.keyType || key.Source != StoredKey || key.Encrypted != (c.passphrase != "") {
			t.Errorf("%s: key %+v", c.kind, key)
		}
		line, err := store.PublicKey(id)
		if err != nil {
			t.Fatalf("%s: %v", c.kind, err)
		}
		public, _, _, _, err := ssh.ParseAuthorizedKey([]byte(line))
		if err != nil || ssh.FingerprintSHA256(public) != key.Fingerprint {
			t.Errorf("%s: public key %q does not match %s: %v", c.kind, line, key.Fingerprint, err)
		}
	}
	store, _, _ := openStore(t)
	if _, _, err := store.GenerateKey("laptop", "dsa", ""); !errors.Is(err, errUnknownKeyKind) {
		t.Errorf("an unknown kind: %v", err)
	}
}

func TestAKeyFilesPublicKeyComesFromThePubFileBesideIt(t *testing.T) {
	store, _, _ := openStore(t)
	text, _ := newKey(t, "")
	other, _ := newKey(t, "")
	path := filepath.Join(t.TempDir(), "id_ed25519")
	signer, err := ssh.ParsePrivateKey(other)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, text, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path+".pub", ssh.MarshalAuthorizedKey(signer.PublicKey()), 0o600); err != nil {
		t.Fatal(err)
	}
	id, _, err := store.AddKeyFile(path, "id_ed25519", "")
	if err != nil {
		t.Fatal(err)
	}
	line, err := store.PublicKey(id)
	if err != nil || line != strings.TrimSpace(string(ssh.MarshalAuthorizedKey(signer.PublicKey()))) {
		t.Errorf("got %q, %v; want the .pub file's key", line, err)
	}
}
