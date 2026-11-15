package session

import (
	"net"
	"time"

	"github.com/Microsoft/go-winio"
)

// agentPipe is the named pipe Windows' OpenSSH agent listens on.
const agentPipe = `\\.\pipe\openssh-ssh-agent`

// dialAgent connects to Windows' OpenSSH agent.
func dialAgent() (net.Conn, error) {
	timeout := time.Second
	return winio.DialPipe(agentPipe, &timeout)
}
