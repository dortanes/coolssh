// Package preferences keeps the user's settings and the window layout in a JSON file.
package preferences

import (
	"errors"
	"fmt"
	"io/fs"
	"sync"

	"github.com/dortanes/coolssh/internal/jsonfile"
)

// Language is the language the user chose: System, or a supported language's tag.
type Language string

// Appearance is the theme the user chose.
type Appearance string

// Style is the look the windows take: a platform's materials and controls, or the stock theme.
type Style string

// Cursor is the shape of the terminal's cursor.
type Cursor string

const (
	SystemLanguage Language = "system"
	English        Language = "en"
	Russian        Language = "ru"

	SystemAppearance Appearance = "system"
	LightAppearance  Appearance = "light"
	DarkAppearance   Appearance = "dark"

	// AutomaticStyle is the platform's own style.
	AutomaticStyle   Style = "auto"
	LiquidGlassStyle Style = "liquid-glass"
	MicaStyle        Style = "mica"
	ShadcnStyle      Style = "shadcn"

	BlockCursor     Cursor = "block"
	BarCursor       Cursor = "bar"
	UnderlineCursor Cursor = "underline"
)

// The bounds of the terminal's numeric settings.
const (
	minFontSize   = 10
	maxFontSize   = 20
	minScrollback = 1_000
	maxScrollback = 100_000
)

// Terminal is how session terminals look and behave.
type Terminal struct {
	// FontSize is in CSS pixels; 0 takes the style's own.
	FontSize    int    `json:"fontSize"`
	Cursor      Cursor `json:"cursor"`
	CursorBlink bool   `json:"cursorBlink"`
	// Scrollback is how many lines a terminal keeps above its screen.
	Scrollback   int  `json:"scrollback"`
	CopyOnSelect bool `json:"copyOnSelect"`
	// OptionAsMeta makes macOS's Option key send Meta rather than type special characters.
	OptionAsMeta bool `json:"optionAsMeta"`
	// HardwareAcceleration draws terminals with WebGL rather than the DOM.
	HardwareAcceleration bool `json:"hardwareAcceleration"`
}

// defaultSidebarWidth is the sidebar's width before the user drags it.
const defaultSidebarWidth = 260

// Preferences are the user's settings and the window layout.
type Preferences struct {
	Language   Language   `json:"language"`
	Appearance Appearance `json:"appearance"`
	Style      Style      `json:"style"`
	// SidebarWidth is in CSS pixels; the window keeps it within the limits its layout sets.
	SidebarWidth int      `json:"sidebarWidth"`
	Terminal     Terminal `json:"terminal"`
	// ConfirmDisconnect asks before closing a connected session's tab and before quitting with
	// sessions connected.
	ConfirmDisconnect bool `json:"confirmDisconnect"`
	// Shortcuts are the shortcuts the user set, an empty one for none; other actions take the
	// platform's default.
	Shortcuts map[Action]string `json:"shortcuts,omitempty"`
}

// Defaults are the preferences before the user changes any.
func Defaults() Preferences {
	return Preferences{
		Language:     SystemLanguage,
		Appearance:   SystemAppearance,
		Style:        AutomaticStyle,
		SidebarWidth: defaultSidebarWidth,
		Terminal: Terminal{
			Cursor:               BlockCursor,
			CursorBlink:          true,
			Scrollback:           10_000,
			HardwareAcceleration: true,
		},
		ConfirmDisconnect: true,
	}
}

// Validate reports the first value outside its allowed set.
func (p Preferences) Validate() error {
	switch p.Language {
	case SystemLanguage, English, Russian:
	default:
		return fmt.Errorf("unknown language %q", p.Language)
	}
	switch p.Appearance {
	case SystemAppearance, LightAppearance, DarkAppearance:
	default:
		return fmt.Errorf("unknown appearance %q", p.Appearance)
	}
	switch p.Style {
	case AutomaticStyle, LiquidGlassStyle, MicaStyle, ShadcnStyle:
	default:
		return fmt.Errorf("unknown style %q", p.Style)
	}
	if p.SidebarWidth <= 0 {
		return fmt.Errorf("sidebar width %d is not positive", p.SidebarWidth)
	}
	if err := validateShortcuts(p.Shortcuts); err != nil {
		return err
	}
	return p.Terminal.validate()
}

func (t Terminal) validate() error {
	if t.FontSize != 0 && (t.FontSize < minFontSize || t.FontSize > maxFontSize) {
		return fmt.Errorf("font size %d is outside %d to %d", t.FontSize, minFontSize, maxFontSize)
	}
	switch t.Cursor {
	case BlockCursor, BarCursor, UnderlineCursor:
	default:
		return fmt.Errorf("unknown cursor %q", t.Cursor)
	}
	if t.Scrollback < minScrollback || t.Scrollback > maxScrollback {
		return fmt.Errorf("scrollback %d is outside %d to %d", t.Scrollback, minScrollback, maxScrollback)
	}
	return nil
}

// Store holds the preferences in effect and the file they are saved to. It is safe for concurrent
// use.
type Store struct {
	path    string
	mu      sync.Mutex
	current Preferences
}

// Open reads path; a missing file means the defaults, and an unreadable or invalid one leaves them in
// effect, untouched until the next Save, with an error alongside a usable Store.
func Open(path string) (*Store, error) {
	s := &Store{path: path, current: Defaults()}
	loaded := Defaults()
	err := jsonfile.Read(path, &loaded)
	if errors.Is(err, fs.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return s, fmt.Errorf("preferences %s: %w", path, err)
	}
	if err := loaded.Validate(); err != nil {
		return s, fmt.Errorf("preferences %s: %w", path, err)
	}
	s.current = loaded
	return s, nil
}

// Current returns the preferences in effect.
func (s *Store) Current() Preferences {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.current
}

// Save validates p, writes it and puts it in effect. The file is replaced whole, so a crash during
// the write leaves the previous one.
func (s *Store) Save(p Preferences) error {
	if err := p.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := jsonfile.Write(s.path, p); err != nil {
		return fmt.Errorf("preferences %s: %w", s.path, err)
	}
	s.current = p
	return nil
}
