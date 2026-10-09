// Package propertyterms accesses the property_terms vocabulary table.
// Create, Update, and Delete run inside writes.Run and return the row changes
// that Run records. Upsert is the un-audited seed path.
package propertyterms

import (
	"database/sql"
	"errors"
	"github.com/mendahu/provenencia/core/database/catalogmodel"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/deleteimpact"
	"github.com/mendahu/provenencia/core/slug"
)

var ErrInvalid = apperr.New(apperr.CodePropertyTermsInvalid, apperr.KindUser)

// ErrDuplicateKey is returned by Create when the label-derived key already
// names a user-origin term under the same Property.
var ErrDuplicateKey = apperr.New(apperr.CodePropertyTermsDuplicateKey, apperr.KindConflict)

// ErrLocked is returned when mutating a proveniencia or plugin-origin term.
var ErrLocked = apperr.New(apperr.CodePropertyTermsLocked, apperr.KindUser)

// ErrInUse is returned by Delete when Observations still reference the term.
var ErrInUse = apperr.New(apperr.CodePropertyTermsInUse, apperr.KindConflict)

const (
	OriginProvenencia = "provenencia"
	OriginUser        = "user"

	// KeyPlaceRelationshipType is product-only: researchers cannot mint terms.
	KeyPlaceRelationshipType = "place_relationship_type"
	// Locked place_relationship_type keys. Walk / cycle behaviour is keyed
	// here in Go (part_of = containment; succeeded_by = succession).
	KeyPartOf      = "part_of"
	KeySucceededBy = "succeeded_by"

	sqlUpsert = `INSERT INTO property_terms (id, property_id, key, origin, label, description, directed, inverse_key)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(property_id, key, origin) DO UPDATE SET
			label = excluded.label,
			description = excluded.description,
			directed = excluded.directed,
			inverse_key = excluded.inverse_key`
	sqlTermCols = `id, property_id, key, origin, label, COALESCE(description, ''), directed, COALESCE(inverse_key, '')`
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
	var inverse any
	if k := strings.TrimSpace(t.InverseKey); k != "" {
		inverse = k
	}
	if _, err := db.Exec(sqlUpsert, id, t.PropertyID, t.Key, t.Origin, t.Label, desc, directedBit(t.Directed), inverse); err != nil {
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
	if _, err := scanTerm(tx.QueryRow(sqlLookup, propertyID, key, OriginUser)); err == nil {
		return Term{}, nil, ErrDuplicateKey.WithParams(key)
	} else if !errors.Is(err, sql.ErrNoRows) {
		return Term{}, nil, err
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
	if _, err := tx.Exec(sqlUpsert, id, propertyID, key, OriginUser, label, desc, 0, nil); err != nil {
		return Term{}, nil, err
	}
	fields := map[string]rowchange.FieldDiff{
		"id":          {Old: nil, New: uid.String()},
		"property_id": {Old: nil, New: uuidString(propertyID)},
		"key":         {Old: nil, New: key},
		"origin":      {Old: nil, New: OriginUser},
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
	if existing.Origin != OriginUser {
		return Term{}, nil, ErrLocked
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
	var directed int
	err := row.Scan(&t.ID, &t.PropertyID, &t.Key, &t.Origin, &t.Label, &t.Description, &directed, &t.InverseKey)
	if err != nil {
		return Term{}, err
	}
	t.Directed = directed != 0
	return t, nil
}

func scanTermRows(rows *sql.Rows) (Term, error) {
	var t Term
	var directed int
	err := rows.Scan(&t.ID, &t.PropertyID, &t.Key, &t.Origin, &t.Label, &t.Description, &directed, &t.InverseKey)
	if err != nil {
		return Term{}, err
	}
	t.Directed = directed != 0
	return t, nil
}

func directedBit(directed bool) int {
	if directed {
		return 1
	}
	return 0
}

func propertyTermsLocked(tx *sql.Tx, propertyID []byte) (bool, error) {
	var key string
	err := tx.QueryRow(sqlPropertyKey, propertyID, "provenencia").Scan(&key)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return key == KeyPlaceRelationshipType, nil
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
