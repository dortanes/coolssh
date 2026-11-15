// Package desktop is the app: its windows, the native menu and the service the windows call.
package desktop

import (
	"io/fs"
	"runtime"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	"golang.org/x/text/language"

	"github.com/dortanes/coolssh/internal/appearance"
	"github.com/dortanes/coolssh/internal/catalog"
	"github.com/dortanes/coolssh/internal/hostsource"
	"github.com/dortanes/coolssh/internal/library"
	"github.com/dortanes/coolssh/internal/preferences"
	"github.com/dortanes/coolssh/internal/product"
	"github.com/dortanes/coolssh/internal/session"
)

// PreferencesEvent carries the Snapshot to every window each time the preferences change.
const PreferencesEvent = "preferences"

// The binding generator finds events only in direct calls.
func init() {
	application.RegisterEvent[Snapshot](PreferencesEvent)
}

const (
	mainWindowName     = "main"
	settingsWindowName = "settings"
)

// surfaces are the colour of a solid window in the light and the dark theme, `--window` in
// frontend/src/styles/tailwind.css; the window shows it until the page paints.
var surfaces = map[bool]application.RGBA{
	false: application.NewRGB(0xff, 0xff, 0xff),
	true:  application.NewRGB(0x0a, 0x0a, 0x0a),
}

// Options configures the app.
type Options struct {
	Product     product.Product
	Preferences *preferences.Store
	Library     *library.Store
	// KnownHosts checks the keys of the hosts sessions connect to.
	KnownHosts *session.KnownHosts
	// Sources are the networks folders can take their hosts from, in the order they are offered.
	Sources  []hostsource.Source
	Catalogs *catalog.Catalogs
	// SystemLanguages are the operating system's preferred languages, most preferred first.
	SystemLanguages []language.Tag
	// Assets holds the built frontend.
	Assets fs.FS
}

// App runs the main window and, on demand, the settings window.
type App struct {
	application *application.App
	options     Options
	sessions    *session.Sessions
	syncer      *hostsource.Syncer
	tokens      *sourceTokens
}

// New builds the app and its main window; Run shows them.
func New(options Options) *App {
	a := &App{options: options}
	a.application = application.New(application.Options{
		Name:       options.Product.Name,
		Services:   []application.Service{application.NewService(&Service{app: a})},
		Assets:     application.AssetOptions{Handler: application.AssetFileServerFS(options.Assets)},
		Mac:        application.MacOptions{ApplicationShouldTerminateAfterLastWindowClosed: true},
		ShouldQuit: func() bool { return a.shouldQuit() },
	})
	a.openSessions()
	a.followTrustedKeys()
	a.openSources()
	appearance.Apply(options.Preferences.Current().Appearance)
	main, _ := a.windowOptions(mainWindowName)
	a.guardClosing(a.application.Window.NewWithOptions(main))
	a.application.Menu.Set(a.menu())
	a.application.Event.OnApplicationEvent(events.Common.ThemeChanged, func(*application.ApplicationEvent) {
		a.paintWindows()
	})
	return a
}

// Run runs the app until it quits.
func (a *App) Run() error {
	return a.application.Run()
}

// Snapshot is the preferences in effect and what they resolve to.
type Snapshot struct {
	Preferences preferences.Preferences `json:"preferences"`
	// Locale is the resolved language's tag, the name of the catalog the windows read.
	Locale string `json:"locale"`
	Look   Look   `json:"look"`
	// Platform names the operating system as Go's runtime.GOOS does.
	Platform string `json:"platform"`
	// Shortcuts are the shortcuts in effect for every action, an empty one for none.
	Shortcuts map[preferences.Action]string `json:"shortcuts"`
}

func (a *App) snapshot() Snapshot {
	current := a.options.Preferences.Current()
	return Snapshot{
		Preferences: current,
		Locale:      a.language().String(),
		Look:        a.look(),
		Platform:    runtime.GOOS,
		Shortcuts:   current.ShortcutsIn(runtime.GOOS),
	}
}

func (a *App) language() language.Tag {
	chosen := a.options.Preferences.Current().Language
	return a.options.Catalogs.Resolve(string(chosen), a.options.SystemLanguages)
}

func (a *App) text(key string) string {
	return a.options.Catalogs.Text(a.language(), key)
}

// savePreferences puts p in effect and tells every window, the menu and the windows' frames.
func (a *App) savePreferences(p preferences.Preferences) (Snapshot, error) {
	if err := a.options.Preferences.Save(p); err != nil {
		a.application.Logger.Error("the preferences were not saved", "error", err)
		return Snapshot{}, err
	}
	snapshot := a.snapshot()
	a.application.Menu.Set(a.menu())
	if settings, ok := a.application.Window.GetByName(settingsWindowName); ok {
		settings.SetTitle(a.text("settings_title"))
	}
	appearance.Apply(p.Appearance)
	a.paintWindows()
	a.application.Event.Emit(PreferencesEvent, snapshot)
	// Problems are worded in the language in effect.
	a.publishSessions()
	return snapshot, nil
}

// openSettings brings the settings window to the front, opening it when it is closed.
func (a *App) openSettings() {
	if settings, ok := a.application.Window.GetByName(settingsWindowName); ok {
		settings.Show()
		settings.Focus()
		return
	}
	options, _ := a.windowOptions(settingsWindowName)
	a.application.Window.NewWithOptions(options)
}

// windowOptions are the options of the app's window with the given name, dressed for the frame in
// effect; false for a name the app does not open.
func (a *App) windowOptions(name string) (application.WebviewWindowOptions, bool) {
	var options application.WebviewWindowOptions
	switch name {
	case mainWindowName:
		options = application.WebviewWindowOptions{
			Name:      mainWindowName,
			Title:     a.options.Product.Name,
			Width:     1180,
			Height:    760,
			MinWidth:  720,
			MinHeight: 460,
		}
	case settingsWindowName:
		options = application.WebviewWindowOptions{
			Name:      settingsWindowName,
			Title:     a.text("settings_title"),
			URL:       "/settings.html",
			Width:     960,
			Height:    760,
			MinWidth:  620,
			MinHeight: 420,
		}
	default:
		return options, false
	}
	a.dress(&options)
	return options, true
}

func (a *App) dark() bool {
	switch a.options.Preferences.Current().Appearance {
	case preferences.LightAppearance:
		return false
	case preferences.DarkAppearance:
		return true
	default:
		return a.application.Env.IsDarkMode()
	}
}

func (a *App) surface() application.RGBA {
	return surfaces[a.dark()]
}

// paintWindows gives solid windows the colour of the theme in effect; a material paints itself.
func (a *App) paintWindows() {
	if a.look().Material {
		return
	}
	surface := a.surface()
	for _, window := range a.application.Window.GetAll() {
		window.SetBackgroundColour(surface)
	}
}
