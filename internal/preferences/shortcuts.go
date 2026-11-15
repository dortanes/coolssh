package preferences

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// Action is something the main window does on a keyboard shortcut.
type Action string

const (
	CommandPaletteAction Action = "commandPalette"
	NewHostAction        Action = "newHost"
	NewFolderAction      Action = "newFolder"
	ClosePaneAction      Action = "closePane"
	NextTabAction        Action = "nextTab"
	PreviousTabAction    Action = "previousTab"
	SplitRightAction     Action = "splitRight"
	SplitDownAction      Action = "splitDown"
	NextPaneAction       Action = "nextPane"
	PreviousPaneAction   Action = "previousPane"
)

// actions lists every Action a shortcut can be set for.
var actions = []Action{
	CommandPaletteAction, NewHostAction, NewFolderAction, ClosePaneAction, NextTabAction, PreviousTabAction,
	SplitRightAction, SplitDownAction, NextPaneAction, PreviousPaneAction,
}

// maxShortcut bounds a shortcut's length; the longest real one is well under it.
const maxShortcut = 64

// A shortcut is modifiers and a physical key joined by "+", as KeyboardEvent names them
// ("Meta+Shift+KeyK"); empty for none.
var (
	macShortcuts = map[Action]string{
		CommandPaletteAction: "Meta+KeyK",
		NewHostAction:        "Meta+KeyN",
		NewFolderAction:      "Meta+Shift+KeyN",
		ClosePaneAction:      "Meta+KeyW",
		NextTabAction:        "Meta+Shift+BracketRight",
		PreviousTabAction:    "Meta+Shift+BracketLeft",
		SplitRightAction:     "Meta+KeyD",
		SplitDownAction:      "Meta+Shift+KeyD",
		NextPaneAction:       "Meta+BracketRight",
		PreviousPaneAction:   "Meta+BracketLeft",
	}
	// Elsewhere Ctrl alone belongs to the shell (Ctrl+K, Ctrl+N, Ctrl+W edit the command line).
	otherShortcuts = map[Action]string{
		CommandPaletteAction: "Control+Shift+KeyK",
		NewHostAction:        "Control+Shift+KeyN",
		NewFolderAction:      "Control+Alt+Shift+KeyN",
		ClosePaneAction:      "Control+Shift+KeyW",
		NextTabAction:        "Control+Tab",
		PreviousTabAction:    "Control+Shift+Tab",
		SplitRightAction:     "Control+Shift+KeyD",
		SplitDownAction:      "Control+Alt+Shift+KeyD",
		NextPaneAction:       "Control+Alt+BracketRight",
		PreviousPaneAction:   "Control+Alt+BracketLeft",
	}
)

// DefaultShortcuts are the shortcuts on an operating system named as runtime.GOOS names it.
func DefaultShortcuts(goos string) map[Action]string {
	if goos == "darwin" {
		return maps.Clone(macShortcuts)
	}
	return maps.Clone(otherShortcuts)
}

// ShortcutsIn are the shortcuts in effect on goos: the user's own over the defaults.
func (p Preferences) ShortcutsIn(goos string) map[Action]string {
	shortcuts := DefaultShortcuts(goos)
	maps.Copy(shortcuts, p.Shortcuts)
	return shortcuts
}

func validateShortcuts(shortcuts map[Action]string) error {
	for action, shortcut := range shortcuts {
		if !slices.Contains(actions, action) {
			return fmt.Errorf("unknown action %q", action)
		}
		if len(shortcut) > maxShortcut || strings.Contains(shortcut, " ") {
			return fmt.Errorf("shortcut %q for %s is not one key combination", shortcut, action)
		}
	}
	return nil
}
