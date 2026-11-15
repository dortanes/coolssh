package session

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	"github.com/dortanes/coolssh/internal/library"
)

// Status is where a session is in its life.
type Status string

const (
	Connecting Status = "connecting"
	Connected  Status = "connected"
	Closed     Status = "closed"
)

// The terminal a session asks for before the window reports its size.
const (
	defaultColumns = 80
	defaultRows    = 24
)

// missedKeepalives is how many keepalives may go unanswered before a session gives the host up.
const missedKeepalives = 3

// Session is one connection to a host, shown in one tab. Its output outlives a reconnect.
type Session struct {
	id string
	// hostID is the library's host; empty for a typed address.
	hostID  string
	resolve func() (library.Connection, error)
	// task is run instead of a shell; nil for a shell.
	task   *Task
	owner  *Sessions
	screen screen

	mu      sync.Mutex
	title   string
	status  Status
	prompt  *Prompt
	answers chan Answer
	problem *Problem
	// outcome is what a finished task's session closed with; nil for none.
	outcome *Problem
	cancel  context.CancelFunc
	shell   *ssh.Session
	input   io.Writer
	columns int
	rows    int
}

// State is what a window shows of a session.
type State struct {
	ID string `json:"id"`
	// HostID is the library's host; empty for a typed address.
	HostID string `json:"hostId"`
	// Title is the host's name when the session started.
	Title  string  `json:"title"`
	Status Status  `json:"status"`
	Prompt *Prompt `json:"prompt"`
	// Problem is why the session closed, worded in the language in effect; empty when nothing went
	// wrong.
	Problem string `json:"problem"`
	// Outcome says how a task the session ran ended, worded; empty for a shell.
	Outcome string `json:"outcome,omitempty"`
}

func (s *Session) state() State {
	s.mu.Lock()
	defer s.mu.Unlock()
	state := State{ID: s.id, HostID: s.hostID, Title: s.title, Status: s.status, Prompt: s.prompt}
	if s.problem != nil {
		state.Problem = s.owner.options.Word(s.problem)
	}
	if s.outcome != nil {
		state.Outcome = s.owner.options.Word(s.outcome)
	}
	return state
}

// start connects the session on its own goroutine.
func (s *Session) start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.connect()
}

// restart connects a closed session again; false when it is still open.
func (s *Session) restart() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.status != Closed {
		return false
	}
	s.connect()
	return true
}

// connect starts the connection's goroutine; s.mu is held.
func (s *Session) connect() {
	ctx, cancel := context.WithCancel(context.Background())
	s.status, s.problem, s.outcome, s.prompt, s.cancel = Connecting, nil, nil, nil, cancel
	go func() {
		defer cancel()
		err := s.run(ctx)
		s.mu.Lock()
		s.status, s.prompt, s.shell, s.input = Closed, nil, nil, nil
		if err != nil && !errors.Is(err, errDeclined) && ctx.Err() == nil {
			s.problem = asProblem(err, s.title)
		}
		s.mu.Unlock()
		s.owner.changed()
	}()
}

func (s *Session) stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *Session) run(ctx context.Context) error {
	connection, err := s.resolve()
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.title = connection.Name
	s.mu.Unlock()
	s.owner.changed()
	attempt := &attempt{session: s}
	if held, err := s.owner.options.Agent(); err == nil && held != nil {
		attempt.agent = held
		defer held.Close()
	}
	defer func() {
		for _, client := range attempt.clients {
			client.Close()
		}
	}()
	client, err := attempt.dial(ctx, connection)
	if err != nil {
		return err
	}
	defer client.Close()
	shell, err := s.openShell(client, attempt.agent, connection.AgentForwarding)
	if err != nil {
		return problem("session_shell_refused", connection.Name)
	}
	defer shell.Close()
	finished := make(chan error, 1)
	go func() { finished <- shell.Wait() }()
	lost := make(chan error, 1)
	if connection.KeepaliveSeconds > 0 {
		go func() {
			lost <- keepAlive(ctx, client, time.Duration(connection.KeepaliveSeconds)*s.owner.options.KeepaliveUnit, connection.Name)
		}()
	}
	select {
	case err := <-finished:
		return s.ended(ctx, err, connection.Name)
	case err := <-lost:
		return err
	case <-ctx.Done():
		return nil
	}
}

// ended is what a session closes with once its shell or task exits with err: a lost connection's
// problem, or nothing, recording how a task ended.
func (s *Session) ended(ctx context.Context, err error, host string) error {
	var missing *ssh.ExitMissingError
	if errors.As(err, &missing) && ctx.Err() == nil {
		return problem("session_lost", host)
	}
	if s.task != nil {
		code := s.task.Done
		if err != nil {
			code = s.task.Failed
		}
		s.mu.Lock()
		s.outcome = problem(code, host)
		s.mu.Unlock()
	}
	return nil
}

// openShell asks for a terminal of the window's size and a shell in it, forwarding the agent when
// asked to.
func (s *Session) openShell(client *ssh.Client, held agent.Agent, forward bool) (*ssh.Session, error) {
	shell, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	if forward && held != nil {
		if err := agent.ForwardToAgent(client, held); err == nil {
			_ = agent.RequestAgentForwarding(shell)
		}
	}
	shell.Stdout, shell.Stderr = &s.screen, &s.screen
	input, err := shell.StdinPipe()
	if err != nil {
		shell.Close()
		return nil, err
	}
	s.mu.Lock()
	columns, rows := s.columns, s.rows
	s.mu.Unlock()
	modes := ssh.TerminalModes{ssh.ECHO: 1, ssh.TTY_OP_ISPEED: 38400, ssh.TTY_OP_OSPEED: 38400}
	if err := shell.RequestPty("xterm-256color", rows, columns, modes); err != nil {
		shell.Close()
		return nil, err
	}
	if s.task != nil {
		err = shell.Start(s.task.Command)
	} else {
		err = shell.Shell()
	}
	if err != nil {
		shell.Close()
		return nil, err
	}
	s.mu.Lock()
	s.status, s.shell, s.input = Connected, shell, input
	// The window may have resized while the terminal was being asked for.
	if s.columns != columns || s.rows != rows {
		_ = shell.WindowChange(s.rows, s.columns)
	}
	s.mu.Unlock()
	s.owner.changed()
	return shell, nil
}

// keepAlive asks the host for a reply every interval until ctx ends, and returns a problem once
// missedKeepalives go unanswered or the connection fails.
func keepAlive(ctx context.Context, client *ssh.Client, interval time.Duration, host string) error {
	missed := 0
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	replies := make(chan error, missedKeepalives+1)
	waiting := false
	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-replies:
			waiting = false
			if err != nil {
				return problem("session_lost", host)
			}
			missed = 0
		case <-ticker.C:
			if waiting {
				missed++
				if missed >= missedKeepalives {
					return problem("session_not_answering", host)
				}
				continue
			}
			waiting = true
			go func() {
				_, _, err := client.SendRequest("keepalive@openssh.com", true, nil)
				replies <- err
			}()
		}
	}
}

// ask shows prompt and waits for the user's answer; errDeclined when the user declines or the
// session closes first.
func (s *Session) ask(ctx context.Context, prompt Prompt) (Answer, error) {
	answers := make(chan Answer, 1)
	s.mu.Lock()
	prompt.ID = s.owner.nextPrompt()
	s.prompt, s.answers = &prompt, answers
	s.mu.Unlock()
	s.owner.changed()
	defer func() {
		s.mu.Lock()
		if s.answers == answers {
			s.prompt, s.answers = nil, nil
		}
		s.mu.Unlock()
		s.owner.changed()
	}()
	select {
	case answer := <-answers:
		if !answer.Accept {
			return Answer{}, errDeclined
		}
		return answer, nil
	case <-ctx.Done():
		return Answer{}, errDeclined
	}
}

// answer delivers the user's reply to the prompt with the given id; false when it is not the one
// shown.
func (s *Session) answer(prompt int, answer Answer) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.prompt == nil || s.prompt.ID != prompt {
		return false
	}
	s.answers <- answer
	s.prompt, s.answers = nil, nil
	return true
}

func (s *Session) resize(columns, rows int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.columns, s.rows = columns, rows
	if s.shell != nil {
		_ = s.shell.WindowChange(rows, columns)
	}
}

// show sends the session's output to window and the window's input to the shell until the window
// goes away.
func (s *Session) show(window Window) {
	s.screen.show(window)
	defer s.screen.hide(window)
	for {
		input, err := window.Receive()
		if err != nil {
			return
		}
		s.mu.Lock()
		writer := s.input
		s.mu.Unlock()
		if writer != nil {
			_, _ = writer.Write(input)
		}
	}
}
