//go:build !darwin || !cgo

package appicon

func draw(int) []byte { return nil }
