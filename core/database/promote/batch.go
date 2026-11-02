package promote

import (
	"bytes"
	"database/sql"
	"errors"

	"github.com/mendahu/provenencia/core/database/rowchange"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/evidencegraphs"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/subjects"
)

// Row targets for one Done. They match graphalign.Target strings.
const (
	TargetHandle = "handle"
	TargetNew    = "new"
	TargetSkip   = "skip"
)

// BatchRow is one Subject on a Done. Skip writes nothing. New mints a handle.
// Handle joins EntityID.
type BatchRow struct {
	SubjectID []byte
	Target    string
	EntityID  []byte
}

// Batch is one Done for an Evidence graph. SeenRevision is the audit revision
// the proposal was read at.
type Batch struct {
	GraphID      []byte
	SeenRevision int64
	Rows         []BatchRow
}

// SaveBatch files one Done in a single transaction: claims and bridges whose
// ends are both handles. The caller records the returned changes as
// promote_batch. A catalog write since SeenRevision is ErrStale and writes
// nothing.
func SaveBatch(tx *database.Tx, userID []byte, in Batch) (struct{}, []rowchange.Change, error) {
	if tx == nil {
		return struct{}{}, nil, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return struct{}{}, nil, err
	}

	q := tx.Tx
	if _, err := evidencegraphs.SourceOf(q, in.GraphID); err != nil {
		if errors.Is(err, evidencegraphs.ErrInvalid) {
			return struct{}{}, nil, ErrInvalid
		}
		return struct{}{}, nil, err
	}
	rev, err := latestRevision(q)
	if err != nil {
		return struct{}{}, nil, err
	}
	if rev != in.SeenRevision {
		return struct{}{}, nil, ErrStale
	}

	var changes []rowchange.Change
	for _, row := range in.Rows {
		rowChanges, err := applyRow(q, in.GraphID, row)
		if err != nil {
			return struct{}{}, nil, err
		}
		changes = append(changes, rowChanges...)
	}

	_, filed, err := identityclaims.FileEvidenceGraphBridgesTx(q, in.GraphID)
	if err != nil {
		return struct{}{}, nil, err
	}
	changes = append(changes, filed...)
	return struct{}{}, changes, nil
}

func latestRevision(tx *sql.Tx) (int64, error) {
	var rev int64
	err := tx.QueryRow(`SELECT COALESCE(MAX(revision), 0) FROM audit_transactions`).Scan(&rev)
	return rev, err
}

// applyRow files one claim and returns its changes;
// an anchor or a skip files nothing and returns nil.
func applyRow(tx *sql.Tx, graphID []byte, row BatchRow) ([]rowchange.Change, error) {
	switch row.Target {
	case TargetSkip:
		if len(row.EntityID) != 0 {
			return nil, ErrInvalid
		}
		return nil, nil
	case TargetNew:
		if len(row.EntityID) != 0 {
			return nil, ErrInvalid
		}
	case TargetHandle:
		if len(row.EntityID) != 16 {
			return nil, ErrInvalid
		}
	default:
		return nil, ErrInvalid
	}
	if len(row.SubjectID) != 16 {
		return nil, ErrInvalid
	}

	subject, err := subjects.GetTx(tx, row.SubjectID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInvalid
	}
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(subject.GraphID, graphID) {
		return nil, ErrInvalid
	}
	if err := requirePrimary(tx, subject); err != nil {
		return nil, err
	}
	existing, err := identityclaims.AcceptedEntityForSubjectTx(tx, subject.ID)
	if err == nil {
		if row.Target == TargetHandle && bytes.Equal(existing.EntityID, row.EntityID) {
			return nil, nil
		}
		return nil, identityclaims.ErrAlreadyMember
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	var entityID []byte
	if row.Target == TargetHandle {
		entityID = row.EntityID
	}
	_, _, changes, err := fileSubject(tx, subject, entityID)
	if err != nil {
		return nil, err
	}
	return changes, nil
}
