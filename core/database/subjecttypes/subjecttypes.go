// Package subjecttypes accesses the subject_types vocabulary table.
//
// Subject types are product-seeded (and later plugin-seeded), not researcher-
// authored. Upsert is for create-time Install / plugin Install only — there is
// no user-origin Create/Update/Delete path.
package subjecttypes

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/ref"
)

var ErrInvalid = apperr.New(apperr.CodeSubjectTypesInvalid, apperr.KindUser)

// ErrDuplicatePrefix is returned when a ref_prefix or candidate_ref_prefix
// collides with any existing value in either column (shared namespace).
var ErrDuplicatePrefix = apperr.New(apperr.CodeSubjectTypesDuplicatePrefix, apperr.KindConflict)

const (
	OriginProvenencia = "provenencia"

	sqlUpsert = `INSERT INTO subject_types (
			id, key, origin, label, description, ref_prefix, candidate_ref_prefix
		) VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(key, origin) DO UPDATE SET
			label = excluded.label,
			description = excluded.description,
			ref_prefix = excluded.ref_prefix,
			candidate_ref_prefix = excluded.candidate_ref_prefix`
	sqlLookup = `SELECT id, key, origin, label, COALESCE(description, ''),
			ref_prefix, candidate_ref_prefix
		FROM subject_types WHERE key = ? AND origin = ?`
	sqlGetByID = `SELECT id, key, origin, label, COALESCE(description, ''),
			ref_prefix, candidate_ref_prefix
		FROM subject_types WHERE id = ?`
	sqlList = `SELECT id, key, origin, label, COALESCE(description, ''),
			ref_prefix, candidate_ref_prefix
		FROM subject_types
		ORDER BY label COLLATE NOCASE, origin, key`
	// Shared three-letter namespace across both prefix columns, excluding this row.
	sqlPrefixTaken = `SELECT 1 FROM subject_types
		WHERE (ref_prefix = ? OR candidate_ref_prefix = ? OR ref_prefix = ? OR candidate_ref_prefix = ?)
			AND id != ?
		LIMIT 1`
)

// Type is one subject_types row.
type Type struct {
	ID                 []byte
	Key                string
	Origin             string
	Label              string
	Description        string
	RefPrefix          string
	CandidateRefPrefix string
}

// Upsert inserts or updates by (key, origin). Mints a UUIDv7 id when ID is empty on insert.
// Origin must be proveniencia or plugin:<id>. Validates both prefixes via ref.ValidatePrefix
// and enforces cross-column uniqueness.
func Upsert(c *database.Catalog, t Type) ([]byte, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	t.Key = strings.TrimSpace(t.Key)
	t.Origin = strings.TrimSpace(t.Origin)
	t.Label = strings.TrimSpace(t.Label)
	t.Description = strings.TrimSpace(t.Description)
	if t.Key == "" || t.Label == "" || !originOK(t.Origin) {
		return nil, ErrInvalid
	}

	refPrefix, err := ref.ValidatePrefix(t.RefPrefix)
	if err != nil {
		return nil, err
	}
	candPrefix, err := ref.ValidatePrefix(t.CandidateRefPrefix)
	if err != nil {
		return nil, err
	}
	if refPrefix == candPrefix {
		return nil, ErrDuplicatePrefix.WithParams(refPrefix)
	}
	t.RefPrefix = refPrefix
	t.CandidateRefPrefix = candPrefix

	id := t.ID
	if len(id) == 0 {
		existing, err := Lookup(c, t.Key, t.Origin)
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

	var taken int
	err = db.QueryRow(sqlPrefixTaken, refPrefix, refPrefix, candPrefix, candPrefix, id).Scan(&taken)
	if err == nil {
		return nil, ErrDuplicatePrefix.WithParams(refPrefix)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	var desc any
	if t.Description == "" {
		desc = nil
	} else {
		desc = t.Description
	}
	if _, err := db.Exec(sqlUpsert, id, t.Key, t.Origin, t.Label, desc, t.RefPrefix, t.CandidateRefPrefix); err != nil {
		return nil, err
	}
	return append([]byte(nil), id...), nil
}

// Lookup returns the type for (key, origin), or sql.ErrNoRows.
func Lookup(c *database.Catalog, key, origin string) (Type, error) {
	db, err := c.DB()
	if err != nil {
		return Type{}, err
	}
	key = strings.TrimSpace(key)
	origin = strings.TrimSpace(origin)
	if key == "" || origin == "" {
		return Type{}, ErrInvalid
	}
	return scanType(db.QueryRow(sqlLookup, key, origin))
}

// LookupTx returns the type for (key, origin) on an open transaction.
func LookupTx(tx *sql.Tx, key, origin string) (Type, error) {
	key = strings.TrimSpace(key)
	origin = strings.TrimSpace(origin)
	if key == "" || origin == "" {
		return Type{}, ErrInvalid
	}
	return scanType(tx.QueryRow(sqlLookup, key, origin))
}

// GetByID returns a type by id, or sql.ErrNoRows.
func GetByID(c *database.Catalog, id []byte) (Type, error) {
	db, err := c.DB()
	if err != nil {
		return Type{}, err
	}
	if len(id) != 16 {
		return Type{}, ErrInvalid
	}
	return scanType(db.QueryRow(sqlGetByID, id))
}

// GetByIDTx returns a type by id on an open transaction.
func GetByIDTx(tx *sql.Tx, id []byte) (Type, error) {
	if len(id) != 16 {
		return Type{}, ErrInvalid
	}
	return scanType(tx.QueryRow(sqlGetByID, id))
}

// List returns all subject_types rows.
func List(c *database.Catalog) ([]Type, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(sqlList)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Type
	for rows.Next() {
		t, err := scanType(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanType(row rowScanner) (Type, error) {
	var t Type
	if err := row.Scan(
		&t.ID, &t.Key, &t.Origin, &t.Label, &t.Description,
		&t.RefPrefix, &t.CandidateRefPrefix,
	); err != nil {
		return Type{}, err
	}
	return t, nil
}

func originOK(origin string) bool {
	if origin == OriginProvenencia {
		return true
	}
	// Future plugin:<id> Install modules use the same Upsert path.
	return strings.HasPrefix(origin, "plugin:") && len(origin) > len("plugin:")
}
