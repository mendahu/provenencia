// Package sourcevocab owns type↔field suggestions per level and provenencia create-time seed.
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
	"github.com/mendahu/provenencia/core/database/metadatafields"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/vocabkey"
)

var ErrInvalid = apperr.New(apperr.CodeSourceVocabInvalid, apperr.KindUser)

const (
	sqlEnsureSuggestion = `INSERT INTO source_type_metadata_fields (source_type_id, field_id, level, sort_order)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(source_type_id, field_id, level) DO UPDATE SET sort_order = excluded.sort_order`
	sqlListSuggestions = `SELECT f.id, f.key, f.origin, f.label, f.data_type, COALESCE(f.description, ''), j.sort_order
		FROM source_type_metadata_fields j
		JOIN source_metadata_fields f ON f.id = j.field_id
		WHERE j.source_type_id = ? AND j.level = ?
		ORDER BY j.sort_order ASC, f.label COLLATE NOCASE, f.key`
	sqlDeleteSuggestion = `DELETE FROM source_type_metadata_fields
		WHERE source_type_id = ? AND field_id = ? AND level = ?`
	// The next sort_order is computed inside the INSERT so concurrent
	// appends cannot read the same MAX and share a slot. (The WHERE clause
	// is also what lets SQLite parse INSERT…SELECT with an upsert clause.)
	sqlAppendSuggestion = `INSERT INTO source_type_metadata_fields (source_type_id, field_id, level, sort_order)
		SELECT ?, ?, ?, COALESCE(MAX(sort_order), -1) + 1
		FROM source_type_metadata_fields WHERE source_type_id = ? AND level = ?
		ON CONFLICT(source_type_id, field_id, level) DO NOTHING`
)

// Suggestion is one ordered field attached to a source type at a level.
type Suggestion struct {
	Field     metadatafields.Field
	SortOrder int
}

// EnsureSuggestion inserts or updates the join row's sort_order. A level
// outside vocabkey.MetadataLevels is ErrInvalid.
func EnsureSuggestion(c *database.Catalog, typeID, fieldID []byte, level vocabkey.Level, sortOrder int) error {
	if !level.HoldsMetadata() {
		return ErrInvalid
	}
	db, err := c.DB()
	if err != nil {
		return err
	}
	if len(typeID) != 16 || len(fieldID) != 16 {
		return ErrInvalid
	}
	_, err = db.Exec(sqlEnsureSuggestion, typeID, fieldID, string(level), sortOrder)
	return err
}

// AppendSuggestion attaches a field to a type at the end of its existing
// order at one level. Re-attaching a field already suggested for the type at
// that level leaves its place alone — the join is a set, so assigning twice is
// not an error. A level outside vocabkey.MetadataLevels is ErrInvalid.
func AppendSuggestion(c *database.Catalog, typeID, fieldID []byte, level vocabkey.Level) error {
	if !level.HoldsMetadata() {
		return ErrInvalid
	}
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
	if _, err := metadatafields.GetByID(c, fieldID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalid
		}
		return err
	}
	_, err = db.Exec(sqlAppendSuggestion, typeID, fieldID, string(level), typeID, string(level))
	return err
}

// ListSuggestions returns a type's suggested fields at one level, ordered by
// sort_order. A level outside vocabkey.MetadataLevels is ErrInvalid.
func ListSuggestions(c *database.Catalog, typeID []byte, level vocabkey.Level) ([]Suggestion, error) {
	if !level.HoldsMetadata() {
		return nil, ErrInvalid
	}
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	if len(typeID) != 16 {
		return nil, ErrInvalid
	}
	rows, err := db.Query(sqlListSuggestions, typeID, string(level))
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

// DeleteSuggestion removes one join row at one level — the detach behind the
// RemoveTypeField RPC. The field itself stays in the vocabulary, and its rows
// at other levels stay. A level outside vocabkey.MetadataLevels is ErrInvalid.
func DeleteSuggestion(c *database.Catalog, typeID, fieldID []byte, level vocabkey.Level) error {
	if !level.HoldsMetadata() {
		return ErrInvalid
	}
	db, err := c.DB()
	if err != nil {
		return err
	}
	if len(typeID) != 16 || len(fieldID) != 16 {
		return ErrInvalid
	}
	_, err = db.Exec(sqlDeleteSuggestion, typeID, fieldID, string(level))
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
		id, err := metadatafields.Upsert(c, metadatafields.Field{
			Key:         f.Key,
			Origin:      vocabkey.OriginProvenencia,
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
			Origin:      vocabkey.OriginProvenencia,
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
		for order, key := range s.FieldKeys {
			fieldID := fieldIDs[key]
			if len(typeID) == 0 || len(fieldID) == 0 {
				return ErrInvalid
			}
			if err := EnsureSuggestion(c, typeID, fieldID, s.Level, order); err != nil {
				return err
			}
		}
	}
	return nil
}
