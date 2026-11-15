package session

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"slices"
	"strings"

	"golang.org/x/crypto/ssh"

	"github.com/dortanes/coolssh/internal/atomicfile"
)

// TrustedKey is a host key the user trusted in the app, one per host it was trusted for.
type TrustedKey struct {
	// Host is how known_hosts names it: a host name or address, in brackets with a port other than 22.
	Host        string `json:"host"`
	Type        string `json:"type"`
	Fingerprint string `json:"fingerprint"`
}

// Trusted lists the keys in the app's own known_hosts, in the order they were trusted. Keys in the
// user's known_hosts are theirs to edit and are not listed.
func (k *KnownHosts) Trusted() ([]TrustedKey, error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	lines, err := k.ownLines()
	if err != nil {
		return nil, err
	}
	trusted := []TrustedKey{}
	for _, line := range lines {
		hosts, key, ok := parseLine(line)
		if !ok {
			continue
		}
		for _, host := range hosts {
			trusted = append(trusted, TrustedKey{Host: host, Type: key.Type(), Fingerprint: ssh.FingerprintSHA256(key)})
		}
	}
	return trusted, nil
}

// Forget removes the lines trusting a key for host from the app's own known_hosts, so the next
// connection asks again.
func (k *KnownHosts) Forget(host, fingerprint string) error {
	if err := k.forget(host, fingerprint); err != nil {
		return err
	}
	k.notify()
	return nil
}

func (k *KnownHosts) forget(host, fingerprint string) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	lines, err := k.ownLines()
	if err != nil {
		return err
	}
	var kept bytes.Buffer
	for _, line := range lines {
		hosts, key, ok := parseLine(line)
		if ok && slices.Contains(hosts, host) && ssh.FingerprintSHA256(key) == fingerprint {
			continue
		}
		kept.WriteString(line + "\n")
	}
	return atomicfile.Replace(k.own, kept.Bytes())
}

func (k *KnownHosts) ownLines() ([]string, error) {
	text, err := os.ReadFile(k.own)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimRight(string(text), "\n"), "\n"), nil
}

// parseLine reads a plain known_hosts line; comments, markers and blank lines are not one.
func parseLine(line string) ([]string, ssh.PublicKey, bool) {
	marker, hosts, key, _, _, err := ssh.ParseKnownHosts([]byte(line))
	if err != nil || marker != "" || key == nil {
		return nil, nil, false
	}
	return hosts, key, true
}
