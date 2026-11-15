// Package appicon gives the app's icon as the operating system draws it.
package appicon

// PNG is the app's icon as a square PNG side pixels wide: on macOS the one the system renders from
// the bundle's Icon Composer icon, elsewhere fallback, the icon the platform packages are made from.
func PNG(side int, fallback []byte) []byte {
	if drawn := draw(side); drawn != nil {
		return drawn
	}
	return fallback
}
