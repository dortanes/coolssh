package desktop

import (
	"runtime"

	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/dortanes/coolssh/internal/preferences"
)

// Look is how the windows are dressed.
type Look struct {
	// Style is the style in effect; Automatic is already replaced by the platform's own.
	Style preferences.Style `json:"style"`
	// Material reports whether the windows show the platform's material behind the page; a style
	// that does not use it paints over it.
	Material bool `json:"material"`
}

// lookFor resolves the chosen style on a platform named as runtime.GOOS names it. The material
// depends on the platform alone: macOS always, Windows where it supports Mica.
func lookFor(chosen preferences.Style, platform string, micaSupported bool) Look {
	style := chosen
	if style == preferences.AutomaticStyle {
		switch platform {
		case "darwin":
			style = preferences.LiquidGlassStyle
		case "windows":
			style = preferences.MicaStyle
		default:
			style = preferences.ShadcnStyle
		}
	}
	material := platform == "darwin" || (platform == "windows" && micaSupported)
	return Look{Style: style, Material: material}
}

func (a *App) look() Look {
	return lookFor(a.options.Preferences.Current().Style, runtime.GOOS, micaSupported())
}

// dress sets the options a window's frame decides: the macOS toolbar, and either the material or
// the solid colour shown until the page paints.
func (a *App) dress(options *application.WebviewWindowOptions) {
	options.Mac.TitleBar = application.MacTitleBarHiddenInsetUnified
	if !a.look().Material {
		options.BackgroundColour = a.surface()
		return
	}
	options.BackgroundType = application.BackgroundTypeTranslucent
	options.BackgroundColour = application.NewRGBA(0, 0, 0, 0)
	// The sidebar's behind-window blur, which follows the app's appearance; the Liquid Glass backdrop
	// leaves the desktop readable through a whole window.
	options.Mac.Backdrop = application.MacBackdropTranslucent
	options.Windows.BackdropType = application.Mica
	switch a.options.Preferences.Current().Appearance {
	case preferences.LightAppearance:
		options.Windows.Theme = application.Light
	case preferences.DarkAppearance:
		options.Windows.Theme = application.Dark
	}
}
