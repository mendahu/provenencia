package onboarding

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/catalogsession"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/project"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/identity"
)

var ErrUnknownUser = apperr.New(apperr.CodeOnboardingUnknownUser, apperr.KindNotFound)

// ListContributors lists contributors on the held catalog session for projectDir.
func ListContributors(projectDir string) ([]users.User, error) {
	var out []users.User
	err := catalogsession.Do(projectDir, func(proj *database.Catalog) error {
		rows, err := users.List(proj)
		if err != nil {
			return err
		}
		out = rows
		return nil
	})
	return out, err
}

// Open opens an existing *.provenencia folder and remembers it as the active project.
// If adoptUserID is set, identity.json is written from that catalog users row
// (replacing a different UUID already on this Mac). Otherwise it mints or loads
// install identity and upserts users after the catalog opens successfully.
// Corrupt identity.json is not overwritten.
// The catalog session stays held after a successful open.
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

	var (
		u        users.User
		resolved ResolvedInfo
		dir      string
	)
	err = catalogsession.Do(projectDir, func(proj *database.Catalog) error {
		got, err := users.Lookup(proj, uid[:])
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUnknownUser
		}
		if err != nil {
			return err
		}
		u = got
		info, err := ensureProject(proj, projectDir, uid[:])
		if err != nil {
			return err
		}
		resolved, err = resolveUpdatedBy(proj, info)
		if err != nil {
			return err
		}
		dir = proj.Dir()
		return nil
	})
	if err != nil {
		return Result{}, err
	}

	id := identity.Identity{UserID: uid, DisplayName: u.DisplayName, Ref: u.Ref}
	return persistIdentityAndActive(identityDir, dir, id, resolved)
}

func openMint(identityDir, projectDir, displayName string) (Result, error) {
	id, err := loadOrMint(identityDir, displayName, false)
	if err != nil {
		return Result{}, err
	}

	var (
		resolved ResolvedInfo
		dir      string
	)
	uid := id.UserID
	err = catalogsession.Do(projectDir, func(proj *database.Catalog) error {
		if err := users.Upsert(proj, uid[:], id.DisplayName, id.Ref); err != nil {
			return err
		}
		info, err := ensureProject(proj, projectDir, uid[:])
		if err != nil {
			return err
		}
		resolved, err = resolveUpdatedBy(proj, info)
		if err != nil {
			return err
		}
		dir = proj.Dir()
		return nil
	})
	if err != nil {
		return Result{}, err
	}
	return persistIdentityAndActive(identityDir, dir, id, resolved)
}

func ensureProject(proj *database.Catalog, projectDir string, updatedBy []byte) (project.Info, error) {
	info, err := project.Get(proj)
	if err == nil {
		if err := project.EnsureUUID(proj); err != nil {
			return project.Info{}, err
		}
		return project.Get(proj)
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
	return project.Get(proj)
}

// resolveUpdatedBy fills updated-by display fields from the users row while the catalog is open.
func resolveUpdatedBy(proj *database.Catalog, info project.Info) (ResolvedInfo, error) {
	out := ResolvedInfo{Info: info}
	if len(info.UpdatedBy) != 16 {
		return out, nil
	}
	if uid, err := uuid.FromBytes(info.UpdatedBy); err == nil {
		out.UpdatedByUserID = uid.String()
	}
	u, err := users.Lookup(proj, info.UpdatedBy)
	if err != nil {
		return ResolvedInfo{}, err
	}
	out.UpdatedByDisplayName = u.DisplayName
	out.UpdatedByRef = u.Ref
	return out, nil
}

// ProjectInfo loads project bookkeeping from the held catalog session
// and resolves updated-by display fields in the same op.
func ProjectInfo(projectDir string) (ResolvedInfo, error) {
	var out ResolvedInfo
	err := catalogsession.Do(projectDir, func(proj *database.Catalog) error {
		if err := project.EnsureUUID(proj); err != nil {
			return err
		}
		info, err := project.Get(proj)
		if errors.Is(err, project.ErrMissing) {
			rows, listErr := users.List(proj)
			if listErr != nil {
				return listErr
			}
			var updatedBy []byte
			if len(rows) > 0 {
				updatedBy = rows[0].ID
			}
			if len(updatedBy) != 16 {
				out = ResolvedInfo{Info: project.Info{Label: project.LabelFromDir(projectDir)}}
				return nil
			}
			info, err = ensureProject(proj, projectDir, updatedBy)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		out, err = resolveUpdatedBy(proj, info)
		return err
	})
	return out, err
}
