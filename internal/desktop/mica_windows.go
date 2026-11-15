package desktop

import "github.com/wailsapp/wails/v3/pkg/w32"

// micaSupported reports whether this Windows build draws backdrop materials (Windows 11 22H2).
func micaSupported() bool {
	return w32.SupportsBackdropTypes()
}
