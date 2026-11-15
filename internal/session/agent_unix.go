//go:build !windows

package session

import (
	"net"
	"os"
)

// dialAgent connects to the SSH agent at SSH_AUTH_SOCK; nil when none is set.
func dialAgent() (net.Conn, error) {
	socket := os.Getenv("SSH_AUTH_SOCK")
	if socket == "" {
		return nil, nil
	}
	return net.Dial("unix", socket)
}
