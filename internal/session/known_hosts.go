package session

import (
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"sync"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// errUnknownHostKey reports a host key neither file knows.
var errUnknownHostKey = errors.New("unknown host key")

// KnownHosts checks host keys against the user's known_hosts, which it never writes, and the
// app's own, where keys the user trusts go. Both are read again on every check.
type KnownHosts struct {
	user string
	own  string
	mu   sync.Mutex
	// changed is called after the app's own file changes, with no lock held.
	changed func()
}

// Follow calls changed after every change to the app's own known_hosts.
func (k *KnownHosts) Follow(changed func()) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.changed = changed
}

func (k *KnownHosts) notify() {
	k.mu.Lock()
	changed := k.changed
	k.mu.Unlock()
	if changed != nil {
		changed()
	}
}

// NewKnownHosts checks keys against the files at user and own, either of which may not exist yet.
func NewKnownHosts(user, own string) *KnownHosts {
	return &KnownHosts{user: user, own: own}
}

func (k *KnownHosts) callback() (ssh.HostKeyCallback, error) {
	var files []string
	for _, file := range []string{k.user, k.own} {
		if _, err := os.Stat(file); err == nil {
			files = append(files, file)
		} else if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
	}
	return knownhosts.New(files...)
}

// check answers nil for a trusted key, errUnknownHostKey for one neither file knows, and a
// Problem for a key that differs from a trusted one or is revoked.
func (k *KnownHosts) check(host, address string, remote net.Addr, key ssh.PublicKey) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	callback, err := k.callback()
	if err != nil {
		return err
	}
	err = callback(address, remote, key)
	var mismatch *knownhosts.KeyError
	var revoked *knownhosts.RevokedError
	switch {
	case errors.As(err, &mismatch) && len(mismatch.Want) == 0:
		return errUnknownHostKey
	case errors.As(err, &mismatch):
		return problem("session_host_key_changed", host, "file", mismatch.Want[0].Filename, "line", strconv.Itoa(mismatch.Want[0].Line))
	case errors.As(err, &revoked):
		return problem("session_host_key_revoked", host, "file", revoked.Revoked.Filename, "line", strconv.Itoa(revoked.Revoked.Line))
	}
	return err
}

// algorithms are the host key algorithms of the keys trusted for address, so a server holding
// several keys presents a trusted one; nil, the defaults, when none is trusted.
func (k *KnownHosts) algorithms(address string) []string {
	k.mu.Lock()
	defer k.mu.Unlock()
	callback, err := k.callback()
	if err != nil {
		return nil
	}
	// A key no file holds makes the callback list every key trusted for the address.
	var mismatch *knownhosts.KeyError
	if !errors.As(callback(address, placeholderAddr, placeholderKey{}), &mismatch) {
		return nil
	}
	var algorithms []string
	for _, known := range mismatch.Want {
		for _, algorithm := range keyAlgorithms(known.Key.Type()) {
			if !slices.Contains(algorithms, algorithm) {
				algorithms = append(algorithms, algorithm)
			}
		}
	}
	return algorithms
}

// keyAlgorithms are the signature algorithms a key type is presented with.
func keyAlgorithms(keyType string) []string {
	if keyType == ssh.KeyAlgoRSA {
		return []string{ssh.KeyAlgoRSASHA512, ssh.KeyAlgoRSASHA256, ssh.KeyAlgoRSA}
	}
	return []string{keyType}
}

// trust adds key for address to the app's file.
func (k *KnownHosts) trust(address string, key ssh.PublicKey) error {
	if err := k.append(address, key); err != nil {
		return err
	}
	k.notify()
	return nil
}

func (k *KnownHosts) append(address string, key ssh.PublicKey) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	if err := os.MkdirAll(filepath.Dir(k.own), 0o700); err != nil {
		return err
	}
	file, err := os.OpenFile(k.own, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintln(file, knownhosts.Line([]string{address}, key)); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}

var placeholderAddr = &net.TCPAddr{IP: net.IPv4zero}

// placeholderKey is a key of a type no host has.
type placeholderKey struct{}

func (placeholderKey) Type() string                        { return "coolssh-placeholder" }
func (placeholderKey) Marshal() []byte                     { return []byte("coolssh-placeholder") }
func (placeholderKey) Verify([]byte, *ssh.Signature) error { return errors.New("placeholder key") }
