// Package appearance gives the app's windows, materials and menus the appearance chosen in the
// settings, while they run.
package appearance

import "github.com/dortanes/coolssh/internal/preferences"

// Apply puts the chosen appearance in effect for the whole app; System follows the operating system.
func Apply(chosen preferences.Appearance) {
	apply(chosen)
}
