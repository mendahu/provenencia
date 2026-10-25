package promote

import (
	"bytes"
	"database/sql"
	"errors"

	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/vocabkey"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/evidencegraphs"
	"github.com/mendahu/provenencia/core/database/identityclaims"
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
// Handle joins EntityID. PinSets are the evidence the researcher confirmed
// (join only; a New row with pin sets is invalid).
type BatchRow struct {
	SubjectID         []byte
	Target            string
	EntityID          []byte
	ConfidenceGradeID []byte
	Argument          string
	PinSets           []PinSet
}

// Batch is one Done for an Evidence graph. SeenRevision is the audit revision
// the proposal was read at.
type Batch struct {
	GraphID      []byte
	SeenRevision int64
	Rows         []BatchRow
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

// SaveBatch files one Done in a single transaction: claims, pin sets,
// and bridges whose ends are both handles. The caller records the returned
// changes as promote_batch. Revision on the result is SeenRevision; the
// caller replaces it with the revision it recorded when that is non-zero,
// so an all-skip batch still reports the revision the proposal was read at.
// A catalog write since SeenRevision is ErrStale and writes nothing.
func SaveBatch(tx *database.Tx, userID []byte, in Batch) (BatchResult, []rowchange.Change, error) {
	if tx == nil {
		return BatchResult{}, nil, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return BatchResult{}, nil, err
	}

	q := tx.Tx
	if _, err := evidencegraphs.SourceOf(q, in.GraphID); err != nil {
		if errors.Is(err, evidencegraphs.ErrInvalid) {
			return BatchResult{}, nil, ErrInvalid
		}
		return BatchResult{}, nil, err
	}
	rev, err := latestRevision(q)
	if err != nil {
		return BatchResult{}, nil, err
	}
	if rev != in.SeenRevision {
		return BatchResult{}, nil, ErrStale
	}

	var (
		changes []rowchange.Change
		written []Written
	)
	for _, row := range in.Rows {
		w, rowChanges, _, err := applyRow(q, in.GraphID, row)
		if err != nil {
			return BatchResult{}, nil, err
		}
		if w == nil {
			continue
		}
		changes = append(changes, rowChanges...)
		written = append(written, *w)
	}

	_, filed, err := identityclaims.FileEvidenceGraphBridgesTx(q, in.GraphID)
	if err != nil {
		return BatchResult{}, nil, err
	}
	changes = append(changes, filed...)
	return BatchResult{Revision: rev, Written: written}, changes, nil
}

func latestRevision(tx *sql.Tx) (int64, error) {
	var rev int64
	err := tx.QueryRow(`SELECT COALESCE(MAX(revision), 0) FROM audit_transactions`).Scan(&rev)
	return rev, err
}

// applyRow files one claim. A nil Written is an anchor or a skip.
func applyRow(tx *sql.Tx, graphID []byte, row BatchRow) (*Written, []rowchange.Change, []byte, error) {
	switch row.Target {
	case TargetSkip:
		if len(row.EntityID) != 0 || len(row.PinSets) != 0 || len(row.ConfidenceGradeID) != 0 || row.Argument != "" {
			return nil, nil, nil, ErrInvalid
		}
		return nil, nil, nil, nil
	case TargetNew:
		if len(row.EntityID) != 0 || len(row.PinSets) != 0 {
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
	if !bytes.Equal(subject.GraphID, graphID) {
		return nil, nil, nil, ErrInvalid
	}
	st, err := subjecttypes.GetByIDTx(tx, subject.SubjectTypeID)
	if err != nil {
		return nil, nil, nil, err
	}
	if !vocabkey.IsPrimaryType(st.Key, st.Origin) {
		return nil, nil, nil, ErrUnsupportedType
	}

	existing, err := identityclaims.AcceptedEntityForSubjectTx(tx, subject.ID)
	if err == nil {
		if row.Target == TargetHandle && bytes.Equal(existing.EntityID, row.EntityID) && len(row.PinSets) == 0 {
			return nil, nil, nil, nil
		}
		return nil, nil, nil, identityclaims.ErrAlreadyMember
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, nil, nil, err
	}

	var (
		entity  canonicalentities.Entity
		changes []rowchange.Change
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
	pins, pinChanges, err := pinSets(tx, claim, row.PinSets)
	if err != nil {
		return nil, nil, nil, err
	}
	changes = append(changes, pinChanges...)
	return &Written{Entity: entity, Claim: claim, Pins: pins}, changes, entity.ID, nil
}
