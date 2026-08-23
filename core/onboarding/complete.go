package onboarding

import (
	"errors"
	"strings"

	"github.com/mendahu/provenance/core/database"
	"github.com/mendahu/provenance/core/database/users"
	"github.com/mendahu/provenance/core/identity"
	"github.com/mendahu/provenance/core/session"
)

var ErrBlankName = errors.New("name is empty")

// Result is the created project path and the install identity used.
type Result struct {
	ProjectDir string
	Identity   identity.Identity
}

// Complete mints or loads install identity, creates a *.provenance folder, and upserts users.
// identityDir and parent are supplied by the caller (Swift: Application Support / Documents).
func Complete(identityDir, parent, displayName, familyName string) (Result, error) {
	displayName = strings.TrimSpace(displayName)
	familyName = strings.TrimSpace(familyName)
	if displayName == "" || familyName == "" {
		return Result{}, ErrBlankName
	}
	folder, err := FolderName(familyName)
	if err != nil {
		return Result{}, err
	}

	id, err := loadOrMint(identityDir, displayName)
	if err != nil {
		return Result{}, err
	}

	proj, err := database.Create(parent, folder)
	if err != nil {
		return Result{}, err
	}
	uid := id.UserID
	if err := users.Upsert(proj, uid[:], displayName); err != nil {
		_ = proj.Close()
		return Result{}, err
	}
	dir := proj.Dir()
	if err := proj.Close(); err != nil {
		return Result{}, err
	}
	if err := rememberActive(identityDir, dir); err != nil {
		return Result{}, err
	}
	return Result{ProjectDir: dir, Identity: id}, nil
}

func rememberActive(identityDir, projectDir string) error {
	return session.Save(identityDir, session.Active{ProjectDir: projectDir})
}

func loadOrMint(identityDir, displayName string) (identity.Identity, error) {
	got, err := identity.Load(identityDir)
	if errors.Is(err, identity.ErrNotFound) {
		minted, err := identity.Mint(displayName)
		if err != nil {
			return identity.Identity{}, err
		}
		if err := identity.Save(identityDir, minted); err != nil {
			return identity.Identity{}, err
		}
		return minted, nil
	}
	if err != nil {
		return identity.Identity{}, err
	}
	got.DisplayName = displayName
	if err := identity.Save(identityDir, *got); err != nil {
		return identity.Identity{}, err
	}
	return *got, nil
}
