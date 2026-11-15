package session

import (
	"path/filepath"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestTrustedKeysAreListedAndForgotten(t *testing.T) {
	own := filepath.Join(t.TempDir(), "known_hosts")
	known := NewKnownHosts("", own)
	first := newServer(t, acceptPassword("secret"))
	second := newServer(t, acceptPassword("secret"))
	for _, server := range []*testServer{first, second} {
		if err := known.trust(server.address(), server.hostKey.PublicKey()); err != nil {
			t.Fatal(err)
		}
	}
	trusted, err := known.Trusted()
	if err != nil || len(trusted) != 2 {
		t.Fatalf("trusted %+v, %v", trusted, err)
	}
	fingerprint := ssh.FingerprintSHA256(first.hostKey.PublicKey())
	if trusted[0].Fingerprint != fingerprint || trusted[0].Type != first.hostKey.PublicKey().Type() {
		t.Errorf("first %+v", trusted[0])
	}
	if err := known.Forget(trusted[0].Host, fingerprint); err != nil {
		t.Fatal(err)
	}
	trusted, err = known.Trusted()
	if err != nil || len(trusted) != 1 || trusted[0].Fingerprint != ssh.FingerprintSHA256(second.hostKey.PublicKey()) {
		t.Errorf("after forgetting: %+v, %v", trusted, err)
	}
	if err := known.check("web", first.address(), placeholderAddr, first.hostKey.PublicKey()); err != errUnknownHostKey {
		t.Errorf("the forgotten key is still trusted: %v", err)
	}
}
