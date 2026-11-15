package preferences

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestOpenWithoutAFileGivesTheDefaults(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "preferences.json"))
	if err != nil {
		t.Fatal(err)
	}
	if got := store.Current(); !reflect.DeepEqual(got, Defaults()) {
		t.Errorf("Current() = %+v, want the defaults", got)
	}
}

func TestSavedPreferencesOpenAgain(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "preferences.json")
	store, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	saved := Defaults()
	saved.Language, saved.Appearance, saved.Style, saved.SidebarWidth = Russian, DarkAppearance, MicaStyle, 300
	saved.Terminal.FontSize, saved.Terminal.Cursor, saved.ConfirmDisconnect = 15, BarCursor, false
	saved.Shortcuts = map[Action]string{NewHostAction: "Meta+Shift+KeyH", ClosePaneAction: ""}
	if err := store.Save(saved); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := reopened.Current(); !reflect.DeepEqual(got, saved) {
		t.Errorf("reopened %+v, want %+v", got, saved)
	}
}

func TestAFileWithoutLaterSettingsOpensWithTheirDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "preferences.json")
	if err := os.WriteFile(path, []byte(`{"language":"ru","appearance":"dark","sidebarWidth":300}`), 0o600); err != nil {
		t.Fatal(err)
	}
	store, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	want := Defaults()
	want.Language, want.Appearance, want.SidebarWidth = Russian, DarkAppearance, 300
	if got := store.Current(); !reflect.DeepEqual(got, want) {
		t.Errorf("Current() = %+v, want %+v", got, want)
	}
}

func TestADamagedFileKeepsTheDefaultsAndStaysUntouched(t *testing.T) {
	tests := []struct {
		name     string
		contents string
	}{
		{"invalid JSON", "{"},
		{"an unknown appearance", `{"language":"en","appearance":"sepia","sidebarWidth":260}`},
		{"a sidebar without width", `{"language":"en","appearance":"dark","sidebarWidth":0}`},
		{"an unknown style", `{"language":"en","appearance":"dark","style":"aqua","sidebarWidth":260}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "preferences.json")
			if err := os.WriteFile(path, []byte(test.contents), 0o600); err != nil {
				t.Fatal(err)
			}
			store, err := Open(path)
			if err == nil {
				t.Error("Open reported no problem with the damaged file")
			}
			if got := store.Current(); !reflect.DeepEqual(got, Defaults()) {
				t.Errorf("Current() = %+v, want the defaults", got)
			}
			if raw, _ := os.ReadFile(path); string(raw) != test.contents {
				t.Errorf("the damaged file changed to %q", raw)
			}
		})
	}
}

func TestSaveRefusesInvalidPreferences(t *testing.T) {
	path := filepath.Join(t.TempDir(), "preferences.json")
	store, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	invalid := Defaults()
	invalid.Language = "de"
	if err := store.Save(invalid); err == nil {
		t.Error("Save accepted an unsupported language")
	}
	if got := store.Current(); !reflect.DeepEqual(got, Defaults()) {
		t.Errorf("Current() = %+v after a refused save", got)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("a refused save wrote the file")
	}
}

func TestTerminalSettingsOutsideTheirBoundsAreRefused(t *testing.T) {
	cases := map[string]func(*Terminal){
		"font size below 10":       func(t *Terminal) { t.FontSize = 9 },
		"font size above 20":       func(t *Terminal) { t.FontSize = 21 },
		"an unknown cursor":        func(t *Terminal) { t.Cursor = "beam" },
		"scrollback below 1 000":   func(t *Terminal) { t.Scrollback = 999 },
		"scrollback above 100 000": func(t *Terminal) { t.Scrollback = 100_001 },
	}
	for name, spoil := range cases {
		t.Run(name, func(t *testing.T) {
			p := Defaults()
			spoil(&p.Terminal)
			if err := p.Validate(); err == nil {
				t.Error("Validate accepted it")
			}
		})
	}
}
