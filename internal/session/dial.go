package session

import (
	"context"
	"errors"
	"net"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/dortanes/coolssh/internal/library"
)

// reachTimeout bounds opening the TCP connection to a host, not the handshake, which may wait
// for the user.
const reachTimeout = 15 * time.Second

// passwordTries is how many times a password or keyboard-interactive answer may be typed.
const passwordTries = 3

// attempt is one run of a session's connection: the hosts it dials and what the user answered.
type attempt struct {
	session *Session
	agent   Agent
	// reason is why a callback stopped the handshake, which ssh reports only as text.
	reason error
	// clients are the jump hosts dialled, to close with the session.
	clients []*ssh.Client
}

func (a *attempt) stop(err error) error {
	if a.reason == nil {
		a.reason = err
	}
	return err
}

// dial connects and authenticates to c, through its jump host when it has one.
func (a *attempt) dial(ctx context.Context, c library.Connection) (*ssh.Client, error) {
	address := net.JoinHostPort(c.Address, strconv.Itoa(c.Port))
	var conn net.Conn
	var err error
	if c.Jump != nil {
		jump, err := a.dial(ctx, *c.Jump)
		if err != nil {
			return nil, err
		}
		a.clients = append(a.clients, jump)
		conn, err = jump.DialContext(ctx, "tcp", address)
		if err != nil {
			return nil, problem("session_unreachable", c.Name)
		}
	} else {
		dialer := net.Dialer{Timeout: reachTimeout}
		conn, err = dialer.DialContext(ctx, "tcp", address)
		if err != nil {
			return nil, problem("session_unreachable", c.Name)
		}
	}
	context.AfterFunc(ctx, func() { conn.Close() })
	login := &login{attempt: a, ctx: ctx, connection: c}
	config := &ssh.ClientConfig{
		User:              c.User,
		Auth:              login.methods(),
		HostKeyCallback:   login.checkHostKey,
		HostKeyAlgorithms: a.session.owner.options.KnownHosts.algorithms(address),
	}
	clientConn, channels, requests, err := ssh.NewClientConn(conn, address, config)
	if err != nil {
		conn.Close()
		return nil, a.handshakeProblem(ctx, err, c.Name)
	}
	if login.typed != "" && login.remember && c.HostID != "" {
		a.session.owner.options.Remember(c.HostID, login.typed)
	}
	return ssh.NewClient(clientConn, channels, requests), nil
}

func (a *attempt) handshakeProblem(ctx context.Context, err error, host string) error {
	switch {
	case a.reason != nil:
		return a.reason
	case ctx.Err() != nil:
		return errDeclined
	// x/crypto reports exhausted authentication only as text.
	case strings.Contains(err.Error(), "unable to authenticate"):
		return problem("session_authentication_failed", host)
	case errors.Is(err, net.ErrClosed):
		return problem("session_lost", host)
	}
	return problem("session_failed", host, "reason", err.Error())
}
