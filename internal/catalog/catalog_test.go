package catalog

import (
	"os"
	"slices"
	"testing"

	"golang.org/x/text/language"
)

func loadFrontendCatalogs(t *testing.T) *Catalogs {
	t.Helper()
	catalogs, err := Load(os.DirFS("../../frontend/messages"))
	if err != nil {
		t.Fatal(err)
	}
	return catalogs
}

func TestEveryCatalogHasTheSourceKeysAndNoOthers(t *testing.T) {
	catalogs := loadFrontendCatalogs(t)
	source := keysOf(catalogs.messages[Source])
	for _, tag := range Supported[1:] {
		if keys := keysOf(catalogs.messages[tag]); !slices.Equal(keys, source) {
			t.Errorf("%s has keys %q, the source has %q", tag, keys, source)
		}
	}
}

func keysOf(messages map[string]string) []string {
	keys := make([]string, 0, len(messages))
	for key := range messages {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

func TestResolvePicksTheLanguageTheUserReads(t *testing.T) {
	catalogs := loadFrontendCatalogs(t)
	tests := []struct {
		name      string
		chosen    string
		preferred []language.Tag
		want      language.Tag
	}{
		{"a chosen language wins over the system", "ru", []language.Tag{language.English}, language.Russian},
		{"the system's first supported language", "system", []language.Tag{language.German, language.Russian}, language.Russian},
		{"a regional variant matches its language", "system", []language.Tag{language.MustParse("ru-RU")}, language.Russian},
		{"no supported system language falls back to the source", "system", []language.Tag{language.German}, Source},
		{"no system languages fall back to the source", "system", nil, Source},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := catalogs.Resolve(test.chosen, test.preferred); got != test.want {
				t.Errorf("Resolve(%q, %v) = %v, want %v", test.chosen, test.preferred, got, test.want)
			}
		})
	}
}

func TestTextFallsBackToTheSourceMessage(t *testing.T) {
	catalogs := &Catalogs{messages: map[language.Tag]map[string]string{
		Source:           {"greeting": "Hello", "farewell": "Bye"},
		language.Russian: {"greeting": "Привет"},
	}}
	if got := catalogs.Text(language.Russian, "greeting"); got != "Привет" {
		t.Errorf("translated message = %q", got)
	}
	if got := catalogs.Text(language.Russian, "farewell"); got != "Bye" {
		t.Errorf("missing translation = %q, want the source message", got)
	}
}

func TestFormatFillsEveryPlaceholder(t *testing.T) {
	catalogs := &Catalogs{messages: map[language.Tag]map[string]string{
		Source: {"changed": "Remove line {line} of {file}, then reconnect to {file}."},
	}}
	got := catalogs.Format(Source, "changed", map[string]string{"file": "known_hosts", "line": "12"})
	if want := "Remove line 12 of known_hosts, then reconnect to known_hosts."; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
