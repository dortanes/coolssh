package session

import (
	"io"

	"golang.org/x/crypto/ssh/agent"
)

// Agent is the user's SSH agent, connected for one session, and how to let it go.
type Agent interface {
	agent.Agent
	io.Closer
}

type systemAgent struct {
	agent.ExtendedAgent
	io.Closer
}

// SystemAgent connects to the user's SSH agent; nil, without an error, when none runs.
func SystemAgent() (Agent, error) {
	conn, err := dialAgent()
	if conn == nil || err != nil {
		return nil, err
	}
	return systemAgent{ExtendedAgent: agent.NewClient(conn), Closer: conn}, nil
}
