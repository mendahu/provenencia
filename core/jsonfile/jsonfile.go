// Package jsonfile writes small install-local JSON files (mkdir, indent, 0o600).
package jsonfile

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/mendahu/provenencia/core/apperr"
)

var ErrNotFound = apperr.New(apperr.CodeFileNotFound, apperr.KindNotFound)

// Read returns the file bytes. A missing path is ErrNotFound.
func Read(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return b, nil
}

// Write marshals v with indent, a trailing newline, and mode 0o600.
func Write(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(path, b, 0o600)
}

// Remove deletes path. A missing file is not an error.
func Remove(path string) error {
	err := os.Remove(path)
	if err != nil && os.IsNotExist(err) {
		return nil
	}
	return err
}
