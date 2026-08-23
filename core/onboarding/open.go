package onboarding

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenance/core/database"
	"github.com/mendahu/provenance/core/database/users"
	"github.com/mendahu/provenance/core/identity"
)

var ErrUnknownUser = errors.New("user not in project")

// ListContributors opens the catalog briefly and returns users rows as identities.
func ListContributors(projectDir string) ([]identity.Identity, error) {
	projectDir = strings.TrimSpace(projectDir)
	if projectDir == "" {
		return nil, database.ErrNotAProject
	}
	proj, err := database.Open(projectDir)
	if err != nil {
		return nil, err
	}
	defer proj.Close()
	rows, err := users.List(proj)
	if err != nil {
		return nil, err
	}
	out := make([]identity.Identity, 0, len(rows))
	for _, row := range rows {
		id, err := uuid.FromBytes(row.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, identity.Identity{UserID: id, DisplayName: row.DisplayName})
	}
	return out, nil
}

// Open opens an existing *.provenance folder and remembers it as the active project.
// If adoptUserID is set, identity.json is written from that catalog users row (new Mac).
// Otherwise it mints or loads install identity and upserts users. Corrupt identity.json is not overwritten.
func Open(identityDir, projectDir, displayName, adoptUserID string) (Result, error) {
	projectDir = strings.TrimSpace(projectDir)
	displayName = strings.TrimSpace(displayName)
	adoptUserID = strings.TrimSpace(adoptUserID)
	if projectDir == "" {
		return Result{}, database.ErrNotAProject
	}
	if adoptUserID != "" {
		return adopt(identityDir, projectDir, adoptUserID)
	}
	return openMint(identityDir, projectDir, displayName)
}

func adopt(identityDir, projectDir, adoptUserID string) (Result, error) {
	uid, err := uuid.Parse(adoptUserID)
	if err != nil || uid.Version() != 7 {
		return Result{}, identity.ErrInvalidID
	}
	if _, err := identity.Load(identityDir); err != nil && !errors.Is(err, identity.ErrNotFound) {
		return Result{}, err
	}

	proj, err := database.Open(projectDir)
	if err != nil {
		return Result{}, err
	}
	name, err := users.Lookup(proj, uid[:])
	if errors.Is(err, sql.ErrNoRows) {
		_ = proj.Close()
		return Result{}, ErrUnknownUser
	}
	if err != nil {
		_ = proj.Close()
		return Result{}, err
	}
	dir := proj.Dir()
	if err := proj.Close(); err != nil {
		return Result{}, err
	}

	id := identity.Identity{UserID: uid, DisplayName: name}
	if err := identity.Save(identityDir, id); err != nil {
		return Result{}, err
	}
	if err := rememberActive(identityDir, dir); err != nil {
		return Result{}, err
	}
	return Result{ProjectDir: dir, Identity: id}, nil
}

func openMint(identityDir, projectDir, displayName string) (Result, error) {
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
