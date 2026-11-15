package preferences

import "testing"

func TestTheUsersShortcutsOverTheDefaults(t *testing.T) {
	p := Defaults()
	p.Shortcuts = map[Action]string{NewHostAction: "Meta+Shift+KeyH", ClosePaneAction: ""}
	shortcuts := p.ShortcutsIn("darwin")
	for action, want := range map[Action]string{
		NewHostAction:        "Meta+Shift+KeyH",
		ClosePaneAction:      "",
		CommandPaletteAction: "Meta+KeyK",
	} {
		if shortcuts[action] != want {
			t.Errorf("%s = %q, want %q", action, shortcuts[action], want)
		}
	}
	if got := p.ShortcutsIn("linux")[CommandPaletteAction]; got != "Control+Shift+KeyK" {
		t.Errorf("the palette on Linux = %q", got)
	}
}

func TestEveryActionHasADefaultOnEveryPlatform(t *testing.T) {
	for _, goos := range []string{"darwin", "windows", "linux"} {
		shortcuts := DefaultShortcuts(goos)
		seen := map[string]Action{}
		for _, action := range actions {
			shortcut := shortcuts[action]
			if shortcut == "" {
				t.Errorf("%s has no default on %s", action, goos)
			}
			if other, taken := seen[shortcut]; taken {
				t.Errorf("%s and %s share %q on %s", action, other, shortcut, goos)
			}
			seen[shortcut] = action
		}
	}
}

func TestShortcutsForUnknownActionsAreRefused(t *testing.T) {
	for name, shortcuts := range map[string]map[Action]string{
		"unknown action": {"launchRockets": "Meta+KeyR"},
		"a sequence":     {NewHostAction: "Meta+KeyK Meta+KeyN"},
	} {
		p := Defaults()
		p.Shortcuts = shortcuts
		if p.Validate() == nil {
			t.Errorf("%s was accepted", name)
		}
	}
}
