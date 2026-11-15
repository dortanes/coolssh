package netbird

import (
	"context"
	"net"
	"os"
	"path/filepath"

	"github.com/Microsoft/go-winio"
	"google.golang.org/grpc"
)

// daemonPipe is where NetBird's daemon, running as a service, serves its API on Windows: a name
// only an administrator can create, so no other process can stand in for it.
const daemonPipe = `\\.\pipe\ProtectedPrefix\Administrators\netbird`

// installedClients are where NetBird's client is installed on Windows.
var installedClients = []string{
	filepath.Join(os.Getenv("ProgramFiles"), "NetBird", "netbird.exe"),
}

func daemonTarget() (string, []grpc.DialOption) {
	return "passthrough:///netbird", []grpc.DialOption{
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return winio.DialPipeContext(ctx, daemonPipe)
		}),
	}
}

func daemonPresent() bool {
	_, err := os.Stat(daemonPipe)
	return err == nil
}
