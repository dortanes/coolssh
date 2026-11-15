//go:build !windows

package netbird

import (
	"os"

	"google.golang.org/grpc"
)

// daemonSocket is where NetBird's daemon serves its API on macOS and Linux.
const daemonSocket = "/var/run/netbird.sock"

// installedClients are where NetBird's client is installed on macOS and Linux.
var installedClients = []string{
	"/Applications/NetBird.app",
	"/usr/local/bin/netbird",
	"/opt/homebrew/bin/netbird",
	"/usr/bin/netbird",
}

func daemonTarget() (string, []grpc.DialOption) {
	return "unix://" + daemonSocket, nil
}

func daemonPresent() bool {
	_, err := os.Stat(daemonSocket)
	return err == nil
}
