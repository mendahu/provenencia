package onboarding

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/project"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/identity"
)

var ErrUnknownUser = apperr.New(apperr.CodeOnboardingUnknownUser, apperr.KindNotFound)

// ListContributors opens the catalog (migrates if needed), ensures user refs, and lists contributors.
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
	if err := users.EnsureRefs(proj); err != nil {
		return nil, err
	}
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
		out = append(out, identity.Identity{UserID: id, DisplayName: row.DisplayName, Ref: row.Ref})
	}
	return out, nil
}

// Open opens an existing *.provenencia folder and remembers it as the active project.
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
	if err := users.EnsureRefs(proj); err != nil {
		_ = proj.Close()
		return Result{}, err
	}
	u, err := users.Lookup(proj, uid[:])
	if errors.Is(err, sql.ErrNoRows) {
		_ = proj.Close()
		return Result{}, ErrUnknownUser
	}
	if err != nil {
		_ = proj.Close()
		return Result{}, err
	}
	info, err := ensureProject(proj, projectDir, uid[:])
	if err != nil {
		_ = proj.Close()
		return Result{}, err
	}
	dir, err := closeCatalog(proj)
	if err != nil {
		return Result{}, err
	}

	id := identity.Identity{UserID: uid, DisplayName: u.DisplayName, Ref: u.Ref}
	return persistIdentityAndActive(identityDir, dir, id, info)
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
	if err := users.EnsureRefs(proj); err != nil {
		_ = proj.Close()
		return Result{}, err
	}
	uid := id.UserID
	if err := users.Upsert(proj, uid[:], id.DisplayName, id.Ref); err != nil {
		_ = proj.Close()
		return Result{}, err
	}
	info, err := ensureProject(proj, projectDir, uid[:])
	if err != nil {
		_ = proj.Close()
		return Result{}, err
	}
	dir, err := closeCatalog(proj)
	if err != nil {
		return Result{}, err
	}
	return persistIdentityAndActive(identityDir, dir, id, info)
}

func ensureProject(proj *database.Catalog, projectDir string, updatedBy []byte) (project.Info, error) {
	info, err := project.Get(proj)
	if err == nil {
		return info, nil
	}
	if !errors.Is(err, project.ErrMissing) {
		return project.Info{}, err
	}
	now := project.NowUTC()
	info = project.Info{
		Label:     project.LabelFromDir(projectDir),
		CreatedAt: now,
		UpdatedAt: now,
		UpdatedBy: updatedBy,
	}
	if info.Label == "" {
		info.Label = "Untitled"
	}
	if err := project.Upsert(proj, info); err != nil {
		return project.Info{}, err
	}
	return info, nil
}

// ProjectInfo loads project bookkeeping from an existing catalog (migrates if needed).
func ProjectInfo(projectDir string) (project.Info, error) {
	projectDir = strings.TrimSpace(projectDir)
	if projectDir == "" {
		return project.Info{}, database.ErrNotAProject
	}
	proj, err := database.Open(projectDir)
	if err != nil {
		return project.Info{}, err
	}
	defer proj.Close()
	if err := users.EnsureRefs(proj); err != nil {
		return project.Info{}, err
	}
	info, err := project.Get(proj)
	if errors.Is(err, project.ErrMissing) {
		rows, listErr := users.List(proj)
		if listErr != nil {
			return project.Info{}, listErr
		}
		var updatedBy []byte
		if len(rows) > 0 {
			updatedBy = rows[0].ID
		}
		if len(updatedBy) != 16 {
			return project.Info{Label: project.LabelFromDir(projectDir)}, nil
		}
		return ensureProject(proj, projectDir, updatedBy)
	}
	return info, err
}
