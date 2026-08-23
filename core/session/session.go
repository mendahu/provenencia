package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const FileName = "active-project.json"

var (
	ErrNotFound = errors.New("active project file not found")
	ErrInvalid  = errors.New("project dir is empty")
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
	b, err := os.ReadFile(path(dir))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	var a Active
	if err := json.Unmarshal(b, &a); err != nil {
		return nil, fmt.Errorf("session: %w", err)
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
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(path(dir), b, 0o600)
}

// Remove deletes active-project.json under dir. Missing file is not an error.
func Remove(dir string) error {
	err := os.Remove(path(dir))
	if err != nil && os.IsNotExist(err) {
		return nil
	}
	return err
}
