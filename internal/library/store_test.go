package library

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dortanes/coolssh/internal/secrets"
)

func openStore(t *testing.T) (*Store, string, *secrets.Memory) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "library.json")
	memory := secrets.NewMemory()
	store, err := Open(path, memory, defaults)
	if err != nil {
		t.Fatal(err)
	}
	return store, path, memory
}

// addFolderWithHost saves folder "prod" holding host "web" and returns their ids.
func addFolderWithHost(t *testing.T, store *Store) (folder, host string) {
	t.Helper()
	if _, err := store.Change(func(l *Library) error {
		var err error
		if folder, err = l.CreateFolder("", "prod"); err != nil {
			return err
		}
		host, err = l.CreateHost(folder, "web")
		return err
	}); err != nil {
		t.Fatal(err)
	}
	return folder, host
}

func TestTheLibraryOpensAsItWasSaved(t *testing.T) {
	store, path, memory := openStore(t)
	folder, host := addFolderWithHost(t, store)
	reopened, err := Open(path, memory, defaults)
	if err != nil {
		t.Fatal(err)
	}
	library := reopened.Library()
	if len(library.Root) != 1 || library.Root[0] != folder || library.Folders[folder].Children[0] != host {
		t.Errorf("reopened %+v", library)
	}
}

func TestADamagedLibraryIsLeftAsItIs(t *testing.T) {
	tests := []struct {
		name     string
		contents string
	}{
		{"invalid JSON", "{"},
		{"another version", `{"version":9,"root":[],"folders":{},"hosts":{},"keys":[]}`},
		{"an item listed twice", `{"version":1,"root":["a","a"],"folders":{},"hosts":{"a":{"id":"a","name":"a"}},"keys":[]}`},
		{"a folder inside itself", `{"version":1,"root":[],"folders":{"a":{"id":"a","name":"a","children":["a"]}},"hosts":{},"keys":[]}`},
		{"a missing item", `{"version":1,"root":["a"],"folders":{},"hosts":{},"keys":[]}`},
		{"a missing default key", `{"version":1,"root":[],"folders":{},"hosts":{},"keys":[],"defaultKeyId":"k"}`},
		{"a port out of range", `{"version":1,"root":["a"],"folders":{},"hosts":{"a":{"id":"a","name":"a","settings":{"port":0}}},"keys":[]}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "library.json")
			if err := os.WriteFile(path, []byte(test.contents), 0o600); err != nil {
				t.Fatal(err)
			}
			store, err := Open(path, secrets.NewMemory(), defaults)
			if err == nil {
				t.Error("Open reported no problem")
			}
			if library := store.Library(); len(library.Root) != 0 {
				t.Errorf("a damaged file opened as %+v", library)
			}
			if raw, _ := os.ReadFile(path); string(raw) != test.contents {
				t.Errorf("the damaged file changed to %q", raw)
			}
		})
	}
}

func TestAPastedKeyLivesOnlyInTheKeyring(t *testing.T) {
	store, path, memory := openStore(t)
	text, _ := newKey(t, "correct horse")
	id, _, err := store.AddKeyText("work", string(text), "correct horse")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "PRIVATE KEY") || strings.Contains(string(raw), "correct horse") {
		t.Error("the library file holds the key's text or passphrase")
	}
	if stored, err := memory.Get(keySecret(id)); err != nil || stored != string(text) {
		t.Errorf("keyring holds %q, %v", stored, err)
	}
	if _, err := store.DeleteKey(id); err != nil {
		t.Fatal(err)
	}
	if _, err := memory.Get(passphraseSecret(id)); !errors.Is(err, secrets.ErrNotFound) {
		t.Error("the passphrase outlived its key")
	}
}

func TestAKeyFileIsReferencedNotCopied(t *testing.T) {
	store, path, _ := openStore(t)
	text, fingerprint := newKey(t, "")
	file := filepath.Join(t.TempDir(), "id_ed25519")
	if err := os.WriteFile(file, text, 0o600); err != nil {
		t.Fatal(err)
	}
	id, library, err := store.AddKeyFile(file, "id_ed25519", "")
	if err != nil {
		t.Fatal(err)
	}
	key := library.Keys[0]
	if key.ID != id || key.Path != file || key.Source != FileKey || key.Fingerprint != fingerprint {
		t.Errorf("key = %+v", key)
	}
	raw, _ := os.ReadFile(path)
	if strings.Contains(string(raw), "PRIVATE KEY") {
		t.Error("the library file holds the key's text")
	}
	if _, _, err := store.AddKeyFile(filepath.Join(t.TempDir(), "missing"), "missing", ""); !errors.Is(err, errUnreadableKey) {
		t.Errorf("a missing file: %v", err)
	}
}

func TestDeletingAFolderForgetsThePasswordsInside(t *testing.T) {
	store, _, memory := openStore(t)
	folder, host := addFolderWithHost(t, store)
	if _, err := store.SetPassword(host, "hunter2"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Delete(folder); err != nil {
		t.Fatal(err)
	}
	if _, err := memory.Get(passwordSecret(host)); !errors.Is(err, secrets.ErrNotFound) {
		t.Error("the deleted host's password is still in the keyring")
	}
}
