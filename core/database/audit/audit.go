// Package audit records append-only research revisions in the catalog.
package audit

import (
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
)

var ErrInvalid = apperr.New(apperr.CodeAuditInvalid, apperr.KindUser)

const (
	ActionCreate = "create"
	ActionUpdate = "update"
	ActionDelete = "delete"

	sqlNextRevision = `SELECT COALESCE(MAX(revision), 0) + 1 FROM audit_transactions`
	sqlInsertTx     = `INSERT INTO audit_transactions (id, revision, user_id, action_type, description, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`
	sqlInsertChange = `INSERT INTO audit_changes (id, audit_transaction_id, entity_type, entity_id, action, changes_json)
		VALUES (?, ?, ?, ?, ?, ?)`
)

// FieldDiff is one field's previous and resulting value in changes_json.
type FieldDiff struct {
	Old any `json:"old"`
	New any `json:"new"`
}

// Change is one row-level mutation under a revision.
type Change struct {
	EntityType string
	EntityID   []byte
	Action     string
	Fields     map[string]FieldDiff
}

// Revision is one logical application action and its row-level changes.
type Revision struct {
	UserID      []byte // 16-byte users.id; nil or empty → SQL NULL
	ActionType  string
	Description string
	CreatedAt   string // RFC3339 UTC
	Changes     []Change
}

// Record allocates the next revision and inserts the transaction + changes on tx.
// The caller owns BEGIN/COMMIT; Record must run inside that transaction.
func Record(tx *sql.Tx, rev Revision) (int64, error) {
	if tx == nil {
		return 0, ErrInvalid
	}
	createdAt := strings.TrimSpace(rev.CreatedAt)
	if createdAt == "" {
		return 0, ErrInvalid
	}
	var userID any
	switch {
	case len(rev.UserID) == 0:
		userID = nil
	case len(rev.UserID) == 16:
		userID = rev.UserID
	default:
		return 0, ErrInvalid
	}
	for _, ch := range rev.Changes {
		if err := validateChange(ch); err != nil {
			return 0, err
		}
	}

	var revision int64
	if err := tx.QueryRow(sqlNextRevision).Scan(&revision); err != nil {
		return 0, err
	}

	txID, err := uuid.NewV7()
	if err != nil {
		return 0, err
	}
	if _, err := tx.Exec(
		sqlInsertTx,
		txID[:],
		revision,
		userID,
		nullIfEmpty(rev.ActionType),
		nullIfEmpty(rev.Description),
		createdAt,
	); err != nil {
		return 0, err
	}

	for _, ch := range rev.Changes {
		changesJSON, err := json.Marshal(ch.Fields)
		if err != nil {
			return 0, err
		}
		changeID, err := uuid.NewV7()
		if err != nil {
			return 0, err
		}
		if _, err := tx.Exec(
			sqlInsertChange,
			changeID[:],
			txID[:],
			strings.TrimSpace(ch.EntityType),
			ch.EntityID,
			ch.Action,
			string(changesJSON),
		); err != nil {
			return 0, err
		}
	}
	return revision, nil
}

func validateChange(ch Change) error {
	if strings.TrimSpace(ch.EntityType) == "" || len(ch.EntityID) != 16 || ch.Fields == nil {
		return ErrInvalid
	}
	switch ch.Action {
	case ActionCreate, ActionUpdate, ActionDelete:
	default:
		return ErrInvalid
	}
	if _, err := json.Marshal(ch.Fields); err != nil {
		return ErrInvalid
	}
	return nil
}

func nullIfEmpty(s string) any {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return s
}
