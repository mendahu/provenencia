// Package sourcevocab owns type↔field suggestions and provenencia create-time seed.
//
// Install upserts the shipped vocabulary registry (origin=provenencia) once at
// catalog create. Call it only from onboarding.createCatalog — not on open.
// It does not heal deleted rows on later opens; calling it twice would refresh
// labels via Upsert (create path invokes it once).
package sourcevocab

import (
	"database/sql"
	"errors"

	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/sourcefields"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
)

var ErrInvalid = apperr.New(apperr.CodeSourceVocabInvalid, apperr.KindUser)

const (
	sqlEnsureSuggestion = `INSERT INTO source_type_metadata_fields (source_type_id, field_id, sort_order)
		VALUES (?, ?, ?)
		ON CONFLICT(source_type_id, field_id) DO UPDATE SET sort_order = excluded.sort_order`
	sqlListSuggestions = `SELECT f.id, f.key, f.origin, f.label, f.data_type, COALESCE(f.description, ''), j.sort_order
		FROM source_type_metadata_fields j
		JOIN source_metadata_fields f ON f.id = j.field_id
		WHERE j.source_type_id = ?
		ORDER BY j.sort_order ASC, f.label COLLATE NOCASE, f.key`
	sqlDeleteSuggestion = `DELETE FROM source_type_metadata_fields
		WHERE source_type_id = ? AND field_id = ?`
	// The next sort_order is computed inside the INSERT so concurrent
	// appends cannot read the same MAX and share a slot. (The WHERE clause
	// is also what lets SQLite parse INSERT…SELECT with an upsert clause.)
	sqlAppendSuggestion = `INSERT INTO source_type_metadata_fields (source_type_id, field_id, sort_order)
		SELECT ?, ?, COALESCE(MAX(sort_order), -1) + 1
		FROM source_type_metadata_fields WHERE source_type_id = ?
		ON CONFLICT(source_type_id, field_id) DO NOTHING`
	sqlCountSuggestions = `SELECT COUNT(*) FROM source_type_metadata_fields WHERE source_type_id = ?`
)

// Suggestion is one ordered field attached to a source type.
type Suggestion struct {
	Field     sourcefields.Field
	SortOrder int
}

// EnsureSuggestion inserts or updates the join row's sort_order.
func EnsureSuggestion(c *database.Catalog, typeID, fieldID []byte, sortOrder int) error {
	db, err := c.DB()
	if err != nil {
		return err
	}
	if len(typeID) != 16 || len(fieldID) != 16 {
		return ErrInvalid
	}
	_, err = db.Exec(sqlEnsureSuggestion, typeID, fieldID, sortOrder)
	return err
}

// AppendSuggestion attaches a field to a type at the end of its existing
// order. Re-attaching a field already suggested for the type leaves its
// place alone — the join is a set, so assigning twice is not an error.
func AppendSuggestion(c *database.Catalog, typeID, fieldID []byte) error {
	db, err := c.DB()
	if err != nil {
		return err
	}
	if len(typeID) != 16 || len(fieldID) != 16 {
		return ErrInvalid
	}
	// Verify both join sides name real rows. Without this a stale id dies
	// on the FK constraint, which the FFI layer can only report as an
	// internal error rather than a user one.
	if _, err := sourcetypes.GetByID(c, typeID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalid
		}
		return err
	}
	if _, err := sourcefields.GetByID(c, fieldID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalid
		}
		return err
	}
	_, err = db.Exec(sqlAppendSuggestion, typeID, fieldID, typeID)
	return err
}

// CountSuggestions reports how many fields a type suggests, without
// materializing the join rows the way ListSuggestions does.
func CountSuggestions(c *database.Catalog, typeID []byte) (int, error) {
	db, err := c.DB()
	if err != nil {
		return 0, err
	}
	if len(typeID) != 16 {
		return 0, ErrInvalid
	}
	var n int
	if err := db.QueryRow(sqlCountSuggestions, typeID).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

// ListSuggestions returns suggested fields for a type, ordered by sort_order.
func ListSuggestions(c *database.Catalog, typeID []byte) ([]Suggestion, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	if len(typeID) != 16 {
		return nil, ErrInvalid
	}
	rows, err := db.Query(sqlListSuggestions, typeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Suggestion
	for rows.Next() {
		var s Suggestion
		var sort sql.NullInt64
		if err := rows.Scan(
			&s.Field.ID, &s.Field.Key, &s.Field.Origin, &s.Field.Label, &s.Field.DataType, &s.Field.Description, &sort,
		); err != nil {
			return nil, err
		}
		if sort.Valid {
			s.SortOrder = int(sort.Int64)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// DeleteSuggestion removes one join row — the detach behind the
// RemoveTypeField RPC. The field itself stays in the vocabulary.
func DeleteSuggestion(c *database.Catalog, typeID, fieldID []byte) error {
	db, err := c.DB()
	if err != nil {
		return err
	}
	if len(typeID) != 16 || len(fieldID) != 16 {
		return ErrInvalid
	}
	_, err = db.Exec(sqlDeleteSuggestion, typeID, fieldID)
	return err
}

// Install writes the provenencia seed registry into a new catalog.
// Call only at create time (onboarding.createCatalog).
func Install(c *database.Catalog) error {
	if _, err := c.DB(); err != nil {
		return err
	}
	fieldIDs := make(map[string][]byte, len(seedFields))
	for _, f := range seedFields {
		id, err := sourcefields.Upsert(c, sourcefields.Field{
			Key:         f.Key,
			Origin:      sourcefields.OriginProvenencia,
			Label:       f.Label,
			DataType:    f.DataType,
			Description: f.Description,
		})
		if err != nil {
			return err
		}
		fieldIDs[f.Key] = id
	}
	typeIDs := make(map[string][]byte, len(seedTypes))
	for _, t := range seedTypes {
		id, err := sourcetypes.Upsert(c, sourcetypes.Type{
			Key:         t.Key,
			Origin:      sourcetypes.OriginProvenencia,
			Label:       t.Label,
			Description: t.Description,
			IconKey:     t.IconKey,
		})
		if err != nil {
			return err
		}
		typeIDs[t.Key] = id
	}
	for _, s := range seedSuggestions {
		typeID := typeIDs[s.TypeKey]
		fieldID := fieldIDs[s.FieldKey]
		if len(typeID) == 0 || len(fieldID) == 0 {
			return ErrInvalid
		}
		if err := EnsureSuggestion(c, typeID, fieldID, s.SortOrder); err != nil {
			return err
		}
	}
	return nil
}
