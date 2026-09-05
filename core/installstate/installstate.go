// Package installstate is the last opened project for this install (not an auth session).
package installstate

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/jsonfile"
)

const FileName = "active-project.json"

var (
	ErrNotFound = apperr.New(apperr.CodeInstallNotFound, apperr.KindNotFound)
	ErrInvalid  = apperr.New(apperr.CodeInstallInvalid, apperr.KindUser)
)

// Active is the last project this install opened (not a catalog row).
type Active struct {
	ProjectDir string `json:"project_dir"`
}

func path(dir string) string {
	return filepath.Join(dir, FileName)
}

func validate(a Active) error {
	if strings.TrimSpace(a.ProjectDir) == "" {
		return ErrInvalid
	}
	return nil
}

// Load reads active-project.json from dir. A missing file is ErrNotFound.
func Load(dir string) (*Active, error) {
	b, err := jsonfile.Read(path(dir))
	if err != nil {
		if errors.Is(err, jsonfile.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	var a Active
	if err := json.Unmarshal(b, &a); err != nil {
		return nil, fmt.Errorf("installstate: %w", err)
	}
	a.ProjectDir = strings.TrimSpace(a.ProjectDir)
	if err := validate(a); err != nil {
		return nil, err
	}
	return &a, nil
}

// Save writes active-project.json under dir (creating dir if needed).
func Save(dir string, a Active) error {
	a.ProjectDir = strings.TrimSpace(a.ProjectDir)
	if err := validate(a); err != nil {
		return err
	}
	return jsonfile.Write(path(dir), a)
}

// Remove deletes active-project.json under dir. Missing file is not an error.
func Remove(dir string) error {
	return jsonfile.Remove(path(dir))
}
