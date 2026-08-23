package onboarding

import (
	"errors"
	"strings"

	"github.com/mendahu/provenance/core/database"
	"github.com/mendahu/provenance/core/database/users"
	"github.com/mendahu/provenance/core/identity"
)

// Open opens an existing *.provenance folder, upserts the install identity into users,
// and remembers it as the active project. It mints identity only when none exists
// (displayName required then). Corrupt identity.json is not overwritten.
func Open(identityDir, projectDir, displayName string) (Result, error) {
	projectDir = strings.TrimSpace(projectDir)
	displayName = strings.TrimSpace(displayName)
	if projectDir == "" {
		return Result{}, database.ErrNotAProject
	}

	id, err := loadOrMintOpen(identityDir, displayName)
	if err != nil {
		return Result{}, err
	}

	proj, err := database.Open(projectDir)
	if err != nil {
		return Result{}, err
	}
	uid := id.UserID
	if err := users.Upsert(proj, uid[:], id.DisplayName); err != nil {
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

func loadOrMintOpen(identityDir, displayName string) (identity.Identity, error) {
	got, err := identity.Load(identityDir)
	if errors.Is(err, identity.ErrNotFound) {
		if displayName == "" {
			return identity.Identity{}, ErrBlankName
		}
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
	if displayName != "" {
		got.DisplayName = displayName
		if err := identity.Save(identityDir, *got); err != nil {
			return identity.Identity{}, err
		}
	}
	return *got, nil
}
