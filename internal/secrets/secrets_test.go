package secrets

import (
	"errors"
	"testing"

	"github.com/zalando/go-keyring"
)

func TestKeyringStoresReadsAndForgetsASecret(t *testing.T) {
	keyring.MockInit()
	stores := map[string]Secrets{"keyring": NewKeyring("coolssh-test"), "memory": NewMemory()}
	for name, store := range stores {
		t.Run(name, func(t *testing.T) { roundTrip(t, store) })
	}
}

// roundTrip stores a secret, reads it back and forgets it, as many times as Delete is repeated.
func roundTrip(t *testing.T, store Secrets) {
	t.Helper()
	if _, err := store.Get("password/a"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get of a missing secret = %v, want ErrNotFound", err)
	}
	if err := store.Set("password/a", "hunter2"); err != nil {
		t.Fatal(err)
	}
	if got, err := store.Get("password/a"); err != nil || got != "hunter2" {
		t.Fatalf("Get = %q, %v", got, err)
	}
	for range 2 {
		if err := store.Delete("password/a"); err != nil {
			t.Fatalf("Delete: %v", err)
		}
	}
	if _, err := store.Get("password/a"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get after Delete = %v, want ErrNotFound", err)
	}
}
