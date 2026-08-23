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

// ListContributors peeks at users without migrating or taking the exclusive lock.
func ListContributors(projectDir string) ([]identity.Identity, error) {
	projectDir = strings.TrimSpace(projectDir)
	if projectDir == "" {
		return nil, database.ErrNotAProject
	}
	proj, err := database.OpenReadOnly(projectDir)
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
// If adoptUserID is set, identity.json is written from that catalog users row
// (replacing a different UUID already on this Mac). Otherwise it mints or loads
// install identity and upserts users after the catalog opens successfully.
// Corrupt identity.json is not overwritten.
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
	dir, err := closeCatalog(proj)
	if err != nil {
		return Result{}, err
	}

	// A different UUID on disk is replaced: this Mac is now that catalog contributor.
	id := identity.Identity{UserID: uid, DisplayName: name}
	return persistIdentityAndActive(identityDir, dir, id)
}

func openMint(identityDir, projectDir, displayName string) (Result, error) {
	id, err := loadOrMint(identityDir, displayName, false)
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
	dir, err := closeCatalog(proj)
	if err != nil {
		return Result{}, err
	}
	return persistIdentityAndActive(identityDir, dir, id)
}
