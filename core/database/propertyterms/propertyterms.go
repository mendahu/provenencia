// Package propertyterms accesses the property_terms vocabulary table with audited mutations.
package propertyterms

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/audit"
	"github.com/mendahu/provenencia/core/database/project"
	"github.com/mendahu/provenencia/core/slug"
)

var ErrInvalid = apperr.New(apperr.CodePropertyTermsInvalid, apperr.KindUser)

// ErrDuplicateKey is returned by Create when the label-derived key already
// names a user-origin term under the same Property.
var ErrDuplicateKey = apperr.New(apperr.CodePropertyTermsDuplicateKey, apperr.KindConflict)

// ErrLocked is returned when mutating a proveniencia or plugin-origin term.
var ErrLocked = apperr.New(apperr.CodePropertyTermsLocked, apperr.KindUser)

// ErrInUse is returned by Delete when Observations still reference the term.
// Until S7-03 adds observations.value_term_id, InUse always reports unused.
var ErrInUse = apperr.New(apperr.CodePropertyTermsInUse, apperr.KindConflict)

const (
	OriginProvenencia = "provenencia"
	OriginUser        = "user"

	sqlUpsert = `INSERT INTO property_terms (id, property_id, key, origin, label, description)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(property_id, key, origin) DO UPDATE SET
			label = excluded.label,
			description = excluded.description`
	sqlLookup = `SELECT id, property_id, key, origin, label, COALESCE(description, '')
		FROM property_terms WHERE property_id = ? AND key = ? AND origin = ?`
	sqlGetByID = `SELECT id, property_id, key, origin, label, COALESCE(description, '')
		FROM property_terms WHERE id = ?`
	sqlListByProperty = `SELECT id, property_id, key, origin, label, COALESCE(description, '')
		FROM property_terms WHERE property_id = ?
		ORDER BY label COLLATE NOCASE, origin, key`
	sqlUpdate = `UPDATE property_terms SET label = ?, description = ? WHERE id = ?`
	sqlDelete = `DELETE FROM property_terms WHERE id = ?`
	// TODO(S7-03): SELECT 1 FROM observations WHERE value_term_id = ? LIMIT 1
	sqlInUse = `SELECT 0 WHERE 0`
)

// Term is one property_terms row.
type Term struct {
	ID          []byte
	PropertyID  []byte
	Key         string
	Origin      string
	Label       string
	Description string
}

// Upsert inserts or updates by (property_id, key, origin). Mints a UUIDv7 id when ID is empty on insert.
// Used by create-time seed; does not write audit.
func Upsert(c *database.Catalog, t Term) ([]byte, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	t.Key = strings.TrimSpace(t.Key)
	t.Origin = strings.TrimSpace(t.Origin)
	t.Label = strings.TrimSpace(t.Label)
	t.Description = strings.TrimSpace(t.Description)
	if len(t.PropertyID) != 16 || t.Key == "" || t.Label == "" || !originOK(t.Origin) {
		return nil, ErrInvalid
	}
	id := t.ID
	if len(id) == 0 {
		existing, err := Lookup(c, t.PropertyID, t.Key, t.Origin)
		if err == nil {
			id = existing.ID
		} else if errors.Is(err, sql.ErrNoRows) {
			uid, err := uuid.NewV7()
			if err != nil {
				return nil, err
			}
			id = uid[:]
		} else {
			return nil, err
		}
	} else if len(id) != 16 {
		return nil, ErrInvalid
	}
	var desc any
	if t.Description == "" {
		desc = nil
	} else {
		desc = t.Description
	}
	if _, err := db.Exec(sqlUpsert, id, t.PropertyID, t.Key, t.Origin, t.Label, desc); err != nil {
		return nil, err
	}
	return append([]byte(nil), id...), nil
}

// Create mints a kebab-case key from label and inserts a user-origin term with audit.
func Create(c *database.Catalog, userID, propertyID []byte, label, description string) (Term, error) {
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return Term{}, err
	}
	if len(propertyID) != 16 {
		return Term{}, ErrInvalid
	}
	key := slug.Kebab(label)
	if key == "" {
		return Term{}, ErrInvalid
	}
	if _, err := Lookup(c, propertyID, key, OriginUser); err == nil {
		return Term{}, ErrDuplicateKey.WithParams(key)
	} else if !errors.Is(err, sql.ErrNoRows) {
		return Term{}, err
	}

	db, err := c.DB()
	if err != nil {
		return Term{}, err
	}
	tx, err := db.Begin()
	if err != nil {
		return Term{}, err
	}
	defer func() { _ = tx.Rollback() }()

	uid, err := uuid.NewV7()
	if err != nil {
		return Term{}, err
	}
	id := uid[:]
	label = strings.TrimSpace(label)
	description = strings.TrimSpace(description)
	if label == "" {
		return Term{}, ErrInvalid
	}
	var desc any
	if description == "" {
		desc = nil
	} else {
		desc = description
	}
	if _, err := tx.Exec(sqlUpsert, id, propertyID, key, OriginUser, label, desc); err != nil {
		return Term{}, err
	}
	fields := map[string]audit.FieldDiff{
		"id":          {Old: nil, New: uid.String()},
		"property_id": {Old: nil, New: uuidString(propertyID)},
		"key":         {Old: nil, New: key},
		"origin":      {Old: nil, New: OriginUser},
		"label":       {Old: nil, New: label},
	}
	if description != "" {
		fields["description"] = audit.FieldDiff{Old: nil, New: description}
	}
	if _, err := audit.Record(tx, audit.Revision{
		UserID:     userID,
		ActionType: "create_property_term",
		CreatedAt:  project.NowUTC(),
		Changes: []audit.Change{{
			EntityType: "property_term",
			EntityID:   id,
			Action:     audit.ActionCreate,
			Fields:     fields,
		}},
	}); err != nil {
		return Term{}, err
	}
	if err := tx.Commit(); err != nil {
		return Term{}, err
	}
	return Lookup(c, propertyID, key, OriginUser)
}

// Update patches label and description for a user-origin term.
// Key, origin, and property_id are immutable. Returns ErrLocked for product/plugin terms.
func Update(c *database.Catalog, userID, id []byte, label, description string) (Term, error) {
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return Term{}, err
	}
	db, err := c.DB()
	if err != nil {
		return Term{}, err
	}
	label = strings.TrimSpace(label)
	description = strings.TrimSpace(description)
	if len(id) != 16 || label == "" {
		return Term{}, ErrInvalid
	}
	existing, err := GetByID(c, id)
	if err != nil {
		return Term{}, err
	}
	if existing.Origin != OriginUser {
		return Term{}, ErrLocked
	}

	tx, err := db.Begin()
	if err != nil {
		return Term{}, err
	}
	defer func() { _ = tx.Rollback() }()

	fields := map[string]audit.FieldDiff{}
	if existing.Label != label {
		fields["label"] = audit.FieldDiff{Old: nullJSON(existing.Label), New: nullJSON(label)}
	}
	if existing.Description != description {
		fields["description"] = audit.FieldDiff{Old: nullJSON(existing.Description), New: nullJSON(description)}
	}
	if len(fields) == 0 {
		_ = tx.Rollback()
		return existing, nil
	}
	var desc any
	if description == "" {
		desc = nil
	} else {
		desc = description
	}
	if _, err := tx.Exec(sqlUpdate, label, desc, id); err != nil {
		return Term{}, err
	}
	if _, err := audit.Record(tx, audit.Revision{
		UserID:     userID,
		ActionType: "update_property_term",
		CreatedAt:  project.NowUTC(),
		Changes: []audit.Change{{
			EntityType: "property_term",
			EntityID:   id,
			Action:     audit.ActionUpdate,
			Fields:     fields,
		}},
	}); err != nil {
		return Term{}, err
	}
	if err := tx.Commit(); err != nil {
		return Term{}, err
	}
	return GetByID(c, id)
}

// Delete removes a user-origin term when unused. Product/plugin terms are locked.
func Delete(c *database.Catalog, userID, id []byte) error {
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return err
	}
	if len(id) != 16 {
		return ErrInvalid
	}
	existing, err := GetByID(c, id)
	if err != nil {
		return err
	}
	if existing.Origin != OriginUser {
		return ErrLocked
	}
	inUse, err := InUse(c, id)
	if err != nil {
		return err
	}
	if inUse {
		return ErrInUse
	}

	db, err := c.DB()
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(sqlDelete, id); err != nil {
		return err
	}
	fields := map[string]audit.FieldDiff{
		"id":          {Old: uuidString(id), New: nil},
		"property_id": {Old: uuidString(existing.PropertyID), New: nil},
		"key":         {Old: existing.Key, New: nil},
		"origin":      {Old: existing.Origin, New: nil},
		"label":       {Old: existing.Label, New: nil},
	}
	if existing.Description != "" {
		fields["description"] = audit.FieldDiff{Old: existing.Description, New: nil}
	}
	if _, err := audit.Record(tx, audit.Revision{
		UserID:     userID,
		ActionType: "delete_property_term",
		CreatedAt:  project.NowUTC(),
		Changes: []audit.Change{{
			EntityType: "property_term",
			EntityID:   id,
			Action:     audit.ActionDelete,
			Fields:     fields,
		}},
	}); err != nil {
		return err
	}
	return tx.Commit()
}

// Lookup returns the term for (propertyID, key, origin).
func Lookup(c *database.Catalog, propertyID []byte, key, origin string) (Term, error) {
	db, err := c.DB()
	if err != nil {
		return Term{}, err
	}
	return scanTerm(db.QueryRow(sqlLookup, propertyID, strings.TrimSpace(key), strings.TrimSpace(origin)))
}

// GetByID returns the term by primary key.
func GetByID(c *database.Catalog, id []byte) (Term, error) {
	db, err := c.DB()
	if err != nil {
		return Term{}, err
	}
	if len(id) != 16 {
		return Term{}, ErrInvalid
	}
	return scanTerm(db.QueryRow(sqlGetByID, id))
}

// ListByProperty returns all terms for a Property, label order.
func ListByProperty(c *database.Catalog, propertyID []byte) ([]Term, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	if len(propertyID) != 16 {
		return nil, ErrInvalid
	}
	rows, err := db.Query(sqlListByProperty, propertyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Term
	for rows.Next() {
		t, err := scanTermRows(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// InUse reports whether any Observation references this term.
// Until S7-03, always false.
func InUse(c *database.Catalog, id []byte) (bool, error) {
	db, err := c.DB()
	if err != nil {
		return false, err
	}
	if len(id) != 16 {
		return false, ErrInvalid
	}
	var one int
	err = db.QueryRow(sqlInUse).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func scanTerm(row *sql.Row) (Term, error) {
	var t Term
	err := row.Scan(&t.ID, &t.PropertyID, &t.Key, &t.Origin, &t.Label, &t.Description)
	if err != nil {
		return Term{}, err
	}
	return t, nil
}

func scanTermRows(rows *sql.Rows) (Term, error) {
	var t Term
	err := rows.Scan(&t.ID, &t.PropertyID, &t.Key, &t.Origin, &t.Label, &t.Description)
	if err != nil {
		return Term{}, err
	}
	return t, nil
}

func originOK(origin string) bool {
	if origin == OriginProvenencia || origin == OriginUser {
		return true
	}
	return strings.HasPrefix(origin, "plugin:") && len(origin) > len("plugin:")
}

func uuidString(id []byte) string {
	if len(id) != 16 {
		return ""
	}
	var u uuid.UUID
	copy(u[:], id)
	return u.String()
}

func nullJSON(s string) any {
	if s == "" {
		return nil
	}
	return s
}
