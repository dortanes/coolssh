package desktop

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/wailsapp/wails/v3/pkg/application"
	"golang.org/x/text/language"

	"github.com/dortanes/coolssh/internal/catalog"
	"github.com/dortanes/coolssh/internal/library"
	"github.com/dortanes/coolssh/internal/preferences"
	"github.com/dortanes/coolssh/internal/product"
	"github.com/dortanes/coolssh/internal/session"
)

func TestSolidWindowsMatchThePagesWindowColour(t *testing.T) {
	css, err := os.ReadFile("../../frontend/src/styles/tailwind.css")
	if err != nil {
		t.Fatal(err)
	}
	blocks := map[bool]string{false: ":root", true: ".dark"}
	for dark, selector := range blocks {
		surface := surfaces[dark]
		hex := fmt.Sprintf("#%02x%02x%02x", surface.Red, surface.Green, surface.Blue)
		block := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(selector) + ` \{[^}]*--window: (#[0-9a-f]{6});`)
		match := block.FindSubmatch(css)
		if match == nil {
			t.Fatalf("%s sets no --window", selector)
		}
		if string(match[1]) != hex {
			t.Errorf("%s --window is %s, the window shows %s", selector, match[1], hex)
		}
	}
}

func TestTheMenuFindsEveryLabelInEveryLanguage(t *testing.T) {
	catalogs, err := catalog.Load(os.DirFS("../../frontend/messages"))
	if err != nil {
		t.Fatal(err)
	}
	for _, tag := range catalog.Supported {
		t.Run(tag.String(), func(t *testing.T) {
			store, err := preferences.Open(filepath.Join(t.TempDir(), "preferences.json"))
			if err != nil {
				t.Fatal(err)
			}
			chosen := preferences.Defaults()
			chosen.Language = preferences.Language(tag.String())
			if err := store.Save(chosen); err != nil {
				t.Fatal(err)
			}
			app := &App{
				// Role items read the running application's name.
				application: application.New(application.Options{Name: "coolssh"}),
				options:     Options{Product: product.Product{Name: "coolssh"}, Preferences: store, Catalogs: catalogs, SystemLanguages: []language.Tag{language.German}},
			}
			app.menu()
			if got := app.language(); got != tag {
				t.Errorf("the menu was built in %v", got)
			}
		})
	}
}

func TestEveryRefusalAndSessionProblemIsWorded(t *testing.T) {
	catalogs, err := catalog.Load(os.DirFS("../../frontend/messages"))
	if err != nil {
		t.Fatal(err)
	}
	keys := session.Problems()
	for _, refusal := range library.Refusals() {
		keys = append(keys, string(refusal))
	}
	for _, key := range keys {
		if !catalogs.Has(key) {
			t.Errorf("a catalog has no message %s", key)
		}
	}
}
