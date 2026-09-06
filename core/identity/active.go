package identity

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mendahu/provenencia/core/apperr"
)

const ActiveFileName = "active-project.json"

var (
	ErrActiveNotFound = apperr.New(apperr.CodeInstallNotFound, apperr.KindNotFound)
	ErrActiveInvalid  = apperr.New(apperr.CodeInstallInvalid, apperr.KindUser)
)

// Active is the last project this install opened (not a catalog row).
type Active struct {
	ProjectDir string `json:"project_dir"`
}

func activePath(dir string) string {
	return filepath.Join(dir, ActiveFileName)
}

func validateActive(a Active) error {
	if strings.TrimSpace(a.ProjectDir) == "" {
		return ErrActiveInvalid
	}
	return nil
}

// LoadActive reads active-project.json from dir. A missing file is ErrActiveNotFound.
func LoadActive(dir string) (*Active, error) {
	b, err := readJSONFile(activePath(dir))
	if err != nil {
		if errors.Is(err, errFileNotFound) {
			return nil, ErrActiveNotFound
		}
		return nil, err
	}
	var a Active
	if err := json.Unmarshal(b, &a); err != nil {
		return nil, fmt.Errorf("identity: active: %w", err)
	}
	a.ProjectDir = strings.TrimSpace(a.ProjectDir)
	if err := validateActive(a); err != nil {
		return nil, err
	}
	return &a, nil
}

// SaveActive writes active-project.json under dir (creating dir if needed).
func SaveActive(dir string, a Active) error {
	a.ProjectDir = strings.TrimSpace(a.ProjectDir)
	if err := validateActive(a); err != nil {
		return err
	}
	return writeJSONFile(activePath(dir), a)
}

// RemoveActive deletes active-project.json under dir. Missing file is not an error.
func RemoveActive(dir string) error {
	return removeJSONFile(activePath(dir))
}
