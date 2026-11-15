package session

import (
	"slices"
	"sync"
)

// held is how much output a session keeps while no window shows it, bytes.
const held = 1 << 20

// Window is a window's terminal stream: output goes out with Send, input comes in with Receive.
type Window interface {
	Send(frame []byte) error
	Receive() ([]byte, error)
}

// screen passes a session's output to the window showing it, holding the latest while none does.
// A window that falls behind blocks Write, which holds the server back rather than drop output.
type screen struct {
	mu      sync.Mutex
	window  Window
	pending []byte
}

func (s *screen) Write(output []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.window != nil {
		if err := s.window.Send(slices.Clone(output)); err == nil {
			return len(output), nil
		}
		s.window = nil
	}
	s.pending = append(s.pending, output...)
	if excess := len(s.pending) - held; excess > 0 {
		s.pending = slices.Delete(s.pending, 0, excess)
	}
	return len(output), nil
}

// show sends window what is held, then everything written after it, in place of the last window.
func (s *screen) show(window Window) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.pending) > 0 {
		if err := window.Send(s.pending); err != nil {
			return
		}
		s.pending = nil
	}
	s.window = window
}

// hide stops sending to window, if it is still the one shown.
func (s *screen) hide(window Window) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.window == window {
		s.window = nil
	}
}
