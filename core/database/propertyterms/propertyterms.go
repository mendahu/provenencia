// Package propertyterms accesses the property_terms vocabulary table.
// Create, Update, and Delete run inside writes.Run and return the row changes
// that Run records. Upsert is the un-audited seed path.
package propertyterms

import (
	"bytes"
	"database/sql"
	"errors"
	"strings"

	"github.com/mendahu/provenencia/core/database/catalogmodel"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/deleteimpact"
	"github.com/mendahu/provenencia/core/slug"
)

var ErrInvalid = apperr.New(apperr.CodePropertyTermsInvalid, apperr.KindUser)

// ErrDuplicateKey is returned by Create and Update when a term of the Property
// already reads the same as the label. Its param is that term's label.
var ErrDuplicateKey = apperr.New(apperr.CodePropertyTermsDuplicateKey, apperr.KindConflict)

// ErrLocked is returned when mutating a proveniencia or plugin-origin term.
var ErrLocked = apperr.New(apperr.CodePropertyTermsLocked, apperr.KindUser)

// ErrInUse is returned by Delete when Observations still reference the term.
var ErrInUse = apperr.New(apperr.CodePropertyTermsInUse, apperr.KindConflict)

const (
	// KeyPlaceRelationshipType is product-only: researchers cannot mint terms.
	KeyPlaceRelationshipType = "place_relationship_type"
	// Locked place_relationship_type keys. Walk / cycle behaviour is keyed
	// here in Go (part_of = containment; succeeded_by = succession).
	KeyPartOf      = string(vocabkey.TermPartOf)
	KeySucceededBy = "succeeded_by"

	sqlUpsert = `INSERT INTO property_terms (id, property_id, key, origin, label, description, directed, inverse_key, neutral)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(property_id, key, origin) DO UPDATE SET
			label = excluded.label,
			description = excluded.description,
			directed = excluded.directed,
			inverse_key = excluded.inverse_key,
			neutral = excluded.neutral`
	sqlTermCols = `id, property_id, key, origin, label, COALESCE(description, ''), directed, COALESCE(inverse_key, ''), neutral`
	sqlLookup   = `SELECT ` + sqlTermCols + `
		FROM property_terms WHERE property_id = ? AND key = ? AND origin = ?`
	sqlGetByID = `SELECT ` + sqlTermCols + `
		FROM property_terms WHERE id = ?`
	sqlListByProperty = `SELECT ` + sqlTermCols + `
		FROM property_terms WHERE property_id = ?
		ORDER BY label COLLATE NOCASE, origin, key`
	sqlUpdate      = `UPDATE property_terms SET label = ?, description = ? WHERE id = ?`
	sqlDelete      = `DELETE FROM property_terms WHERE id = ?`
	sqlPropertyKey = `SELECT key FROM properties WHERE id = ? AND origin = ?`
)

// Term is one property_terms row.
type Term struct {
	ID          []byte
	PropertyID  []byte
	Key         string
	Origin      string
	Label       string
	Description string
	Directed    bool // kinship / place-relationship: order is part of the key
	// InverseKey names the term (same Property and origin) that reads the
	// same relationship from the other end: parent ↔ child. Empty when none.
	InverseKey string
	// Neutral terms are no evidence either way ("unknown" sex at birth):
	// reconciliation never lets one win, and matching never compares one.
	Neutral bool
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
	if len(t.PropertyID) != 16 || t.Key == "" || t.Label == "" || !vocabkey.ValidOrigin(t.Origin) {
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
	var inverse any
	if k := strings.TrimSpace(t.InverseKey); k != "" {
		inverse = k
	}
	if _, err := db.Exec(sqlUpsert, id, t.PropertyID, t.Key, t.Origin, t.Label, desc, directedBit(t.Directed), inverse, directedBit(t.Neutral)); err != nil {
		return nil, err
	}
	return append([]byte(nil), id...), nil
}

// Create mints a kebab-case key from label and inserts a user-origin term.
// Refuses place_relationship_type (product-locked vocabulary). The caller
// records the returned changes.
func Create(tx *database.Tx, userID, propertyID []byte, label, description string) (Term, []rowchange.Change, error) {
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return Term{}, nil, err
	}
	if tx == nil || len(propertyID) != 16 {
		return Term{}, nil, ErrInvalid
	}
	if locked, err := propertyTermsLocked(tx.Tx, propertyID); err != nil {
		return Term{}, nil, err
	} else if locked {
		return Term{}, nil, ErrLocked
	}
	key := slug.Kebab(label)
	if key == "" {
		return Term{}, nil, ErrInvalid
	}
	if dup, found, err := conflictingTerm(tx, propertyID, nil, label, key); err != nil {
		return Term{}, nil, err
	} else if found {
		return Term{}, nil, ErrDuplicateKey.WithParams(dup.Label)
	}

	uid, err := uuid.NewV7()
	if err != nil {
		return Term{}, nil, err
	}
	id := uid[:]
	label = strings.TrimSpace(label)
	description = strings.TrimSpace(description)
	if label == "" {
		return Term{}, nil, ErrInvalid
	}
	var desc any
	if description == "" {
		desc = nil
	} else {
		desc = description
	}
	if _, err := tx.Exec(sqlUpsert, id, propertyID, key, vocabkey.OriginUser, label, desc, 0, nil, 0); err != nil {
		return Term{}, nil, err
	}
	fields := map[string]rowchange.FieldDiff{
		"id":          {Old: nil, New: uid.String()},
		"property_id": {Old: nil, New: uuidString(propertyID)},
		"key":         {Old: nil, New: key},
		"origin":      {Old: nil, New: vocabkey.OriginUser},
		"label":       {Old: nil, New: label},
	}
	if description != "" {
		fields["description"] = rowchange.FieldDiff{Old: nil, New: description}
	}
	term, err := scanTerm(tx.QueryRow(sqlGetByID, id))
	if err != nil {
		return Term{}, nil, err
	}
	return term, []rowchange.Change{{
		EntityType: "property_term",
		EntityID:   append([]byte(nil), id...),
		Action:     rowchange.ActionCreate,
		Fields:     fields,
	}}, nil
}

// Update patches label and description for a user-origin term.
// Key, origin, and property_id are immutable. Returns ErrLocked for product/plugin terms.
// An unchanged label and description returns the existing term and no changes.
func Update(tx *database.Tx, userID, id []byte, label, description string) (Term, []rowchange.Change, error) {
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return Term{}, nil, err
	}
	if tx == nil {
		return Term{}, nil, ErrInvalid
	}
	label = strings.TrimSpace(label)
	description = strings.TrimSpace(description)
	if len(id) != 16 || label == "" {
		return Term{}, nil, ErrInvalid
	}
	existing, err := scanTerm(tx.QueryRow(sqlGetByID, id))
	if err != nil {
		return Term{}, nil, err
	}
	if existing.Origin != vocabkey.OriginUser {
		return Term{}, nil, ErrLocked
	}

	if existing.Label != label {
		if dup, found, err := conflictingTerm(tx, existing.PropertyID, id, label, ""); err != nil {
			return Term{}, nil, err
		} else if found {
			return Term{}, nil, ErrDuplicateKey.WithParams(dup.Label)
		}
	}

	fields := map[string]rowchange.FieldDiff{}
	if existing.Label != label {
		fields["label"] = rowchange.FieldDiff{Old: nullJSON(existing.Label), New: nullJSON(label)}
	}
	if existing.Description != description {
		fields["description"] = rowchange.FieldDiff{Old: nullJSON(existing.Description), New: nullJSON(description)}
	}
	if len(fields) == 0 {
		return existing, nil, nil
	}
	var desc any
	if description == "" {
		desc = nil
	} else {
		desc = description
	}
	if _, err := tx.Exec(sqlUpdate, label, desc, id); err != nil {
		return Term{}, nil, err
	}
	term, err := scanTerm(tx.QueryRow(sqlGetByID, id))
	if err != nil {
		return Term{}, nil, err
	}
	return term, []rowchange.Change{{
		EntityType: "property_term",
		EntityID:   append([]byte(nil), id...),
		Action:     rowchange.ActionUpdate,
		Fields:     fields,
	}}, nil
}

// valueWords are the shipped registry's word rules, never catalog data: a text
// key reads nothing else.
var valueWords = matchregistry.Default().Value.Words

// conflictingTerm is the first term of the Property, other than selfID, that
// reads the same as label (the same text key at Equivalent, any origin) or,
// when slugKey is set, is a user term with that key: the upsert would
// overwrite it. A label with no text key is ErrInvalid.
func conflictingTerm(tx *database.Tx, propertyID, selfID []byte, label, slugKey string) (Term, bool, error) {
	want, ok := valuecompare.TextKey(label, valuecompare.Equivalent, valueWords)
	if !ok {
		return Term{}, false, ErrInvalid
	}
	rows, err := tx.Query(sqlListByProperty, propertyID)
	if err != nil {
		return Term{}, false, err
	}
	defer rows.Close()
	for rows.Next() {
		t, err := scanTermRows(rows)
		if err != nil {
			return Term{}, false, err
		}
		if bytes.Equal(t.ID, selfID) {
			continue
		}
		if k, ok := valuecompare.TextKey(t.Label, valuecompare.Equivalent, valueWords); (ok && k == want) ||
			(slugKey != "" && t.Origin == vocabkey.OriginUser && t.Key == slugKey) {
			return t, true, nil
		}
	}
	return Term{}, false, rows.Err()
}

// Delete removes a user-origin term when unused. Product/plugin terms are locked.
// The caller records the returned changes.
func Delete(tx *database.Tx, userID, id []byte) ([]rowchange.Change, error) {
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return nil, err
	}
	if tx == nil || len(id) != 16 {
		return nil, ErrInvalid
	}
	existing, err := scanTerm(tx.QueryRow(sqlGetByID, id))
	if err != nil {
		return nil, err
	}

	report, err := deleteimpact.Impact(tx.Tx, catalogmodel.KindPropertyTerm, id)
	if err != nil {
		return nil, err
	}
	if err := deleteimpact.Refuse(report, deleteimpact.Codes{
		InUse: ErrInUse, OriginLocked: ErrLocked, NotFound: ErrInvalid,
	}); err != nil {
		return nil, err
	}

	if _, err := tx.Exec(sqlDelete, id); err != nil {
		return nil, err
	}
	fields := map[string]rowchange.FieldDiff{
		"id":          {Old: uuidString(id), New: nil},
		"property_id": {Old: uuidString(existing.PropertyID), New: nil},
		"key":         {Old: existing.Key, New: nil},
		"origin":      {Old: existing.Origin, New: nil},
		"label":       {Old: existing.Label, New: nil},
	}
	if existing.Description != "" {
		fields["description"] = rowchange.FieldDiff{Old: existing.Description, New: nil}
	}
	return []rowchange.Change{{
		EntityType: "property_term",
		EntityID:   append([]byte(nil), id...),
		Action:     rowchange.ActionDelete,
		Fields:     fields,
	}}, nil
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

func scanTerm(row *sql.Row) (Term, error) {
	var t Term
	var directed, neutral int
	err := row.Scan(&t.ID, &t.PropertyID, &t.Key, &t.Origin, &t.Label, &t.Description, &directed, &t.InverseKey, &neutral)
	if err != nil {
		return Term{}, err
	}
	t.Directed = directed != 0
	t.Neutral = neutral != 0
	return t, nil
}

func scanTermRows(rows *sql.Rows) (Term, error) {
	var t Term
	var directed, neutral int
	err := rows.Scan(&t.ID, &t.PropertyID, &t.Key, &t.Origin, &t.Label, &t.Description, &directed, &t.InverseKey, &neutral)
	if err != nil {
		return Term{}, err
	}
	t.Directed = directed != 0
	t.Neutral = neutral != 0
	return t, nil
}

func directedBit(on bool) int {
	if on {
		return 1
	}
	return 0
}

func propertyTermsLocked(tx *sql.Tx, propertyID []byte) (bool, error) {
	var key string
	err := tx.QueryRow(sqlPropertyKey, propertyID, vocabkey.OriginProvenencia).Scan(&key)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return key == KeyPlaceRelationshipType, nil
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
