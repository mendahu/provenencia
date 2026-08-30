package identity

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/jsonfile"
)

const FileName = "identity.json"

var (
	ErrNotFound    = errors.New("identity file not found")
	ErrInvalidName = errors.New("display name is empty")
	ErrInvalidID   = errors.New("user id must be UUIDv7")
)

// Identity is the install-local contributor (not a project catalog row).
type Identity struct {
	UserID      uuid.UUID `json:"user_id"`
	DisplayName string    `json:"display_name"`
}

func path(dir string) string {
	return filepath.Join(dir, FileName)
}

func validate(id Identity) error {
	if strings.TrimSpace(id.DisplayName) == "" {
		return ErrInvalidName
	}
	if id.UserID.Version() != 7 {
		return ErrInvalidID
	}
	return nil
}

// Mint returns a new UUIDv7 identity. It does not write disk.
func Mint(displayName string) (Identity, error) {
	name := strings.TrimSpace(displayName)
	if name == "" {
		return Identity{}, ErrInvalidName
	}
	id, err := uuid.NewV7()
	if err != nil {
		return Identity{}, err
	}
	return Identity{UserID: id, DisplayName: name}, nil
}

// Load reads identity.json from dir. A missing file is ErrNotFound; it does not mint.
func Load(dir string) (*Identity, error) {
	b, err := jsonfile.Read(path(dir))
	if err != nil {
		if errors.Is(err, jsonfile.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	var id Identity
	if err := json.Unmarshal(b, &id); err != nil {
		return nil, fmt.Errorf("identity: %w", err)
	}
	if err := validate(id); err != nil {
		return nil, err
	}
	return &id, nil
}

// Save writes identity.json under dir (creating dir if needed).
func Save(dir string, id Identity) error {
	if err := validate(id); err != nil {
		return err
	}
	id.DisplayName = strings.TrimSpace(id.DisplayName)
	return jsonfile.Write(path(dir), id)
}

// Remove deletes identity.json under dir. Missing file is not an error.
func Remove(dir string) error {
	return jsonfile.Remove(path(dir))
}
