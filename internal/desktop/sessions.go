package desktop

import (
	"errors"
	"strconv"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"

	"github.com/dortanes/coolssh/internal/session"
)

// SessionsEvent carries every session's state to the windows each time one changes.
const SessionsEvent = "sessions"

// terminalStream is the stream a window opens per terminal; its first frame names the session.
const terminalStream = "terminal"

func init() {
	application.RegisterEvent[[]session.State](SessionsEvent)
}

func (a *App) openSessions() {
	a.sessions = session.New(session.Options{
		Library:    a.options.Library,
		KnownHosts: a.options.KnownHosts,
		Agent:      session.SystemAgent,
		Changed:    a.publishSessions,
		Remember: func(hostID, password string) {
			updated, err := a.options.Library.SetPassword(hostID, password)
			_ = a.published(updated, err, "secrets_not_saved")
		},
		Word: func(problem *session.Problem) string {
			return a.options.Catalogs.Format(a.language(), problem.Code, problem.Values)
		},
	})
	a.application.HandleStream(terminalStream, func(window *application.StreamConn) {
		defer window.Close()
		id, err := window.Receive()
		if err != nil {
			return
		}
		_ = a.sessions.Show(string(id), window)
	})
}

func (a *App) publishSessions() {
	a.application.Event.Emit(SessionsEvent, a.sessions.States())
}

// confirmQuit asks before quitting with sessions connected and quits once the user agrees.
func (a *App) confirmQuit() {
	dialog := a.application.Dialog.Question().
		SetTitle(a.text("session_quit_title")).
		SetMessage(a.options.Catalogs.Format(a.language(), "session_quit_body", map[string]string{
			"count": strconv.Itoa(a.sessions.Connected()),
		}))
	quit := dialog.AddButton(a.text("session_quit_confirm")).OnClick(func() {
		a.sessions.CloseAll()
		a.application.Quit()
	})
	cancel := dialog.AddButton(a.text("common_cancel"))
	dialog.SetDefaultButton(quit).SetCancelButton(cancel)
	if main, ok := a.application.Window.GetByName(mainWindowName); ok {
		dialog.AttachToWindow(main)
	}
	dialog.Show()
}

// asksBeforeDisconnecting reports whether ending the connected sessions needs the user's yes.
func (a *App) asksBeforeDisconnecting() bool {
	return a.options.Preferences.Current().ConfirmDisconnect && a.sessions.Connected() > 0
}

// shouldQuit lets the app quit at once unless it must ask before disconnecting sessions.
func (a *App) shouldQuit() bool {
	if !a.asksBeforeDisconnecting() {
		a.sessions.CloseAll()
		return true
	}
	// The dialog runs on the main thread, which is busy asking whether to quit.
	go a.confirmQuit()
	return false
}

// guardClosing asks before the main window closes with sessions connected.
func (a *App) guardClosing(window *application.WebviewWindow) {
	window.RegisterHook(events.Common.WindowClosing, func(event *application.WindowEvent) {
		if a.asksBeforeDisconnecting() {
			event.Cancel()
			go a.confirmQuit()
		}
	})
}

// Connect opens a session to a host and returns its id; a host that cannot be reached as
// configured is refused.
func (s *Service) Connect(hostID string) (string, error) {
	id, err := s.app.sessions.Open(hostID)
	if err != nil {
		return "", s.app.refuse(err, "session_not_opened")
	}
	return id, nil
}

// ConnectAddress opens a session to a typed [user@]host[:port] without adding a host, and returns
// its id; an address that cannot be read is refused.
func (s *Service) ConnectAddress(target string) (string, error) {
	id, err := s.app.sessions.OpenAddress(target)
	if err != nil {
		return "", s.app.refuse(err, "session_not_opened")
	}
	return id, nil
}

// Sessions returns every open session's state, in the order they opened.
func (s *Service) Sessions() []session.State {
	return s.app.sessions.States()
}

// Answer replies to what a session asked.
func (s *Service) Answer(id string, prompt int, answer session.Answer) error {
	return s.app.sessionCall(s.app.sessions.Answer(id, prompt, answer))
}

// Resize gives a session's terminal a size in columns and rows.
func (s *Service) Resize(id string, columns, rows int) error {
	return s.app.sessionCall(s.app.sessions.Resize(id, columns, rows))
}

// Reconnect connects a closed session again in its tab.
func (s *Service) Reconnect(id string) error {
	return s.app.sessionCall(s.app.sessions.Reconnect(id))
}

// CloseSession ends a session and forgets it.
func (s *Service) CloseSession(id string) error {
	return s.app.sessionCall(s.app.sessions.Close(id))
}

// sessionCall words the failure of a call about a session that is no longer open.
func (a *App) sessionCall(err error) error {
	if err == nil {
		return nil
	}
	return errors.New(a.text("session_gone"))
}
