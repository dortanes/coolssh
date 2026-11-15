package desktop

import (
	"github.com/dortanes/coolssh/internal/appicon"
	"github.com/dortanes/coolssh/internal/preferences"
	"github.com/dortanes/coolssh/internal/product"
)

// Service is what the windows call in Go.
type Service struct {
	app *App
}

// Snapshot returns the preferences in effect and what they resolve to on this platform.
func (s *Service) Snapshot() Snapshot {
	return s.app.snapshot()
}

// SavePreferences puts p in effect in every window and saves it; invalid preferences are refused.
func (s *Service) SavePreferences(p preferences.Preferences) (Snapshot, error) {
	return s.app.savePreferences(p)
}

// OpenSettings brings the settings window to the front, opening it when it is closed.
func (s *Service) OpenSettings() {
	s.app.openSettings()
}

// aboutIconSide is twice the 64 px the About section shows the icon at.
const aboutIconSide = 128

// Product returns what the app says about itself: name, version, copyright, license, source.
func (s *Service) Product() product.Product {
	about := s.app.options.Product
	about.Icon = appicon.PNG(aboutIconSide, about.Icon)
	return about
}
