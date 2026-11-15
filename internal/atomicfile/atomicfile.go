// Package atomicfile replaces a file whole, so a crash during a write leaves the previous file.
package atomicfile

import (
	"os"
	"path/filepath"
)

// Replace writes raw to a file beside path, readable only by the user, and renames it over path,
// creating path's folder when missing.
func Replace(path string, raw []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*")
	if err != nil {
		return err
	}
	defer os.Remove(temporary.Name())
	if _, err := temporary.Write(raw); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporary.Name(), path)
}
