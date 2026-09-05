package onboarding

import (
	"errors"
	"strings"

	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/project"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/identity"
	"github.com/mendahu/provenencia/core/installstate"
)

var ErrBlankName = apperr.New(apperr.CodeOnboardingBlankName, apperr.KindUser)

// Result is the created/opened project path and the install identity used.
type Result struct {
	ProjectDir string
	Identity   identity.Identity
	Project    project.Info
}

// Complete mints or loads install identity, creates a *.provenencia folder, and upserts users.
// identityDir and parent are supplied by the caller (Swift: Application Support / Documents).
// familyName is the project label; the folder basename is a kebab slug of that label.
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

	id, err := loadOrMint(identityDir, displayName, true)
	if err != nil {
		return Result{}, err
	}

	proj, err := database.Create(parent, folder)
	if err != nil {
		return Result{}, err
	}
	uid := id.UserID
	if err := users.Upsert(proj, uid[:], displayName, id.Ref); err != nil {
		_ = proj.Close()
		return Result{}, err
	}
	now := project.NowUTC()
	info := project.Info{
		Label:     familyName,
		CreatedAt: now,
		UpdatedAt: now,
		UpdatedBy: uid[:],
	}
	if err := project.Upsert(proj, info); err != nil {
		_ = proj.Close()
		return Result{}, err
	}
	dir, err := closeCatalog(proj)
	if err != nil {
		return Result{}, err
	}
	if err := rememberActive(identityDir, dir); err != nil {
		return Result{}, err
	}
	return Result{ProjectDir: dir, Identity: id, Project: info}, nil
}

func rememberActive(identityDir, projectDir string) error {
	return installstate.Save(identityDir, installstate.Active{ProjectDir: projectDir})
}

func closeCatalog(proj *database.Catalog) (string, error) {
	dir := proj.Dir()
	if err := proj.Close(); err != nil {
		return "", err
	}
	return dir, nil
}

func persistIdentityAndActive(identityDir, dir string, id identity.Identity, info project.Info) (Result, error) {
	if err := identity.Save(identityDir, id); err != nil {
		return Result{}, err
	}
	if err := rememberActive(identityDir, dir); err != nil {
		return Result{}, err
	}
	return Result{ProjectDir: dir, Identity: id, Project: info}, nil
}

// persistNow writes identity.json immediately (Complete). Open mint passes false (H2).
func loadOrMint(identityDir, displayName string, persistNow bool) (identity.Identity, error) {
	got, err := identity.Load(identityDir)
	if errors.Is(err, identity.ErrNotFound) {
		if displayName == "" {
			return identity.Identity{}, ErrBlankName
		}
		minted, err := identity.Mint(displayName)
		if err != nil {
			return identity.Identity{}, err
		}
		if persistNow {
			if err := identity.Save(identityDir, minted); err != nil {
				return identity.Identity{}, err
			}
		}
		return minted, nil
	}
	if err != nil {
		return identity.Identity{}, err
	}
	if displayName != "" {
		got.DisplayName = displayName
	}
	if persistNow {
		if err := identity.Save(identityDir, *got); err != nil {
			return identity.Identity{}, err
		}
	}
	return *got, nil
}
