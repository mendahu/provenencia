package promote

import (
	"bytes"
	"database/sql"
	"errors"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/audit"
	"github.com/mendahu/provenencia/core/database/autoreconciler"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/project"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
)

// Row targets for one Done. They match graphalign.Target strings.
const (
	TargetHandle = "handle"
	TargetNew    = "new"
	TargetSkip   = "skip"
)

// BatchRow is one Subject on a Done. Skip writes nothing. New mints a handle.
// Handle joins EntityID. Pairs are confirmed comparisons (join and new-onto
// existing only; a mint with pairs is invalid, same as Save).
type BatchRow struct {
	SubjectID         []byte
	Target            string
	EntityID          []byte
	ConfidenceGradeID []byte
	Argument          string
	Pairs             []Pair
}

// Batch is one Done for a Source. SeenRevision is the audit revision the
// proposal was read at. SkipBridgeIDs are Evidence bridges the researcher
// switched off.
type Batch struct {
	SourceID      []byte
	SeenRevision  int64
	Rows          []BatchRow
	SkipBridgeIDs [][]byte
}

// Written is one claim the batch filed.
type Written struct {
	Entity canonicalentities.Entity
	Claim  identityclaims.Claim
	Pins   int
}

// BatchResult is the revision the batch wrote (or SeenRevision when nothing
// changed) and the claims it filed. Anchors and skips are absent.
type BatchResult struct {
	Revision int64
	Written  []Written
}

// SaveBatch files one Done in a single transaction: claims, one-hop pins,
// bridges (except switched-off ones), one audit revision, and one recompute.
// A catalog write since SeenRevision is ErrStale and writes nothing.
func SaveBatch(c *database.Catalog, userID []byte, in Batch) (BatchResult, error) {
	db, err := c.DB()
	if err != nil {
		return BatchResult{}, err
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return BatchResult{}, err
	}
	if len(in.SourceID) != 16 {
		return BatchResult{}, ErrInvalid
	}
	skip := map[string]struct{}{}
	for _, id := range in.SkipBridgeIDs {
		if len(id) != 16 {
			return BatchResult{}, ErrInvalid
		}
		skip[string(id)] = struct{}{}
	}

	tx, err := db.Begin()
	if err != nil {
		return BatchResult{}, err
	}
	defer func() { _ = tx.Rollback() }()

	var exists int
	if err := tx.QueryRow(`SELECT 1 FROM sources WHERE id = ?`, in.SourceID).Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return BatchResult{}, ErrInvalid
		}
		return BatchResult{}, err
	}
	rev, err := latestRevision(tx)
	if err != nil {
		return BatchResult{}, err
	}
	if rev != in.SeenRevision {
		return BatchResult{}, ErrStale
	}

	var (
		changes  []audit.Change
		touched  [][]byte
		subjects [][]byte
		written  []Written
	)
	for _, row := range in.Rows {
		w, rowChanges, entityID, err := applyRow(tx, in.SourceID, row)
		if err != nil {
			return BatchResult{}, err
		}
		if w == nil {
			continue
		}
		changes = append(changes, rowChanges...)
		touched = append(touched, entityID)
		subjects = append(subjects, row.SubjectID)
		written = append(written, *w)
	}

	assocs, filed, err := identityclaims.FileSourceBridgesTx(tx, in.SourceID, skip)
	if err != nil {
		return BatchResult{}, err
	}
	changes = append(changes, filed...)
	touched = append(touched, assocs...)

	outRev := rev
	if len(changes) > 0 {
		outRev, err = audit.Record(tx, audit.Revision{
			UserID:     userID,
			ActionType: "promote_batch",
			CreatedAt:  project.NowUTC(),
			Changes:    changes,
		})
		if err != nil {
			return BatchResult{}, err
		}
		for _, sid := range subjects {
			extra, err := autoreconciler.HandlesObservingSubject(tx, sid)
			if err != nil {
				return BatchResult{}, err
			}
			touched = append(touched, extra...)
		}
		if err := autoreconciler.RecomputeTx(tx, touched); err != nil {
			return BatchResult{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return BatchResult{}, err
	}
	return BatchResult{Revision: outRev, Written: written}, nil
}

func latestRevision(tx *sql.Tx) (int64, error) {
	var rev int64
	err := tx.QueryRow(`SELECT COALESCE(MAX(revision), 0) FROM audit_transactions`).Scan(&rev)
	return rev, err
}

// applyRow files one claim. A nil Written is an anchor or a skip.
func applyRow(tx *sql.Tx, sourceID []byte, row BatchRow) (*Written, []audit.Change, []byte, error) {
	switch row.Target {
	case TargetSkip:
		if len(row.EntityID) != 0 || len(row.Pairs) != 0 || len(row.ConfidenceGradeID) != 0 || row.Argument != "" {
			return nil, nil, nil, ErrInvalid
		}
		return nil, nil, nil, nil
	case TargetNew:
		if len(row.EntityID) != 0 || len(row.Pairs) != 0 {
			return nil, nil, nil, ErrInvalid
		}
	case TargetHandle:
		if len(row.EntityID) != 16 {
			return nil, nil, nil, ErrInvalid
		}
	default:
		return nil, nil, nil, ErrInvalid
	}
	if len(row.SubjectID) != 16 {
		return nil, nil, nil, ErrInvalid
	}
	if len(row.ConfidenceGradeID) != 0 && len(row.ConfidenceGradeID) != 16 {
		return nil, nil, nil, ErrInvalid
	}

	subject, err := subjects.GetTx(tx, row.SubjectID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, nil, ErrInvalid
	}
	if err != nil {
		return nil, nil, nil, err
	}
	if !bytes.Equal(subject.SourceID, sourceID) {
		return nil, nil, nil, ErrInvalid
	}
	st, err := subjecttypes.GetByIDTx(tx, subject.SubjectTypeID)
	if err != nil {
		return nil, nil, nil, err
	}
	if !PrimaryKind(st.Key, st.Origin) {
		return nil, nil, nil, ErrUnsupportedType
	}

	existing, err := identityclaims.AcceptedEntityForSubjectTx(tx, subject.ID)
	if err == nil {
		if row.Target == TargetHandle && bytes.Equal(existing.EntityID, row.EntityID) && len(row.Pairs) == 0 {
			return nil, nil, nil, nil
		}
		return nil, nil, nil, identityclaims.ErrAlreadyMember
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, nil, nil, err
	}

	var (
		entity  canonicalentities.Entity
		changes []audit.Change
	)
	if row.Target == TargetNew {
		minted, change, err := canonicalentities.InsertTx(tx, canonicalentities.CreateInput{
			SubjectTypeID: subject.SubjectTypeID,
		})
		if err != nil {
			return nil, nil, nil, err
		}
		entity, changes = minted, append(changes, change)
	} else {
		entity, err = joinTarget(tx, row.EntityID, subject.SubjectTypeID)
		if err != nil {
			return nil, nil, nil, err
		}
	}
	claim, claimChange, err := identityclaims.InsertTx(tx, identityclaims.CreateInput{
		SubjectID:         subject.ID,
		EntityID:          entity.ID,
		Status:            identityclaims.StatusAccepted,
		ConfidenceGradeID: row.ConfidenceGradeID,
		Argument:          row.Argument,
	})
	if err != nil {
		return nil, nil, nil, err
	}
	changes = append(changes, claimChange)
	pins, pinChanges, err := pinPairs(tx, claim, row.Pairs)
	if err != nil {
		return nil, nil, nil, err
	}
	changes = append(changes, pinChanges...)
	return &Written{Entity: entity, Claim: claim, Pins: pins}, changes, entity.ID, nil
}
