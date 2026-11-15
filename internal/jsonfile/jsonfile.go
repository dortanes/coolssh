// Package jsonfile keeps a value in a JSON file that is replaced whole on every write, so a crash
// during a write leaves the previous file.
package jsonfile

import (
	"encoding/json"
	"os"

	"github.com/dortanes/coolssh/internal/atomicfile"
)

// Read decodes the file at path into v. A missing file is an error that matches fs.ErrNotExist.
func Read(path string, v any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, v)
}

// Write encodes v and replaces the file at path with it, creating its folder when missing.
func Write(path string, v any) error {
	raw, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return atomicfile.Replace(path, raw)
}
