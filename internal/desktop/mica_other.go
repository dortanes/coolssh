//go:build !windows

package desktop

func micaSupported() bool {
	return false
}
