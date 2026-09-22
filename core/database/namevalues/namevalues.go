// Package namevalues stores shared genealogical NameValue rows.
//
// A NameValue always has a full-form normalized reading (`form`). Ordered
// parts are optional. Part `type` is a product-known key from the compiled
// registry (see registry.go) when set; empty type means an untyped segment.
// Insert rejects unknown non-empty types. User-minted part-type vocabulary
// (DB rows) is deferred — keep DDL as open TEXT.
//
// NameValues are value objects (UUID for persistence only). Mutation is not
// supported; consumers attach new rows when a name assertion changes.
// Format profiles and project defaults live elsewhere (structured-name-model §4).
package namevalues

import (
	"database/sql"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
)

var ErrInvalid = apperr.New(apperr.CodeNameValuesInvalid, apperr.KindUser)

// Product part-type keys (seeded-vocabulary §4.1). Same strings as PartTypes().
const (
	PartTypePrefix        = "prefix"
	PartTypeGiven         = "given"
	PartTypeInitial       = "initial"
	PartTypeNick          = "nick"
	PartTypeSurnamePrefix = "surname_prefix"
	PartTypeSurname       = "surname"
	PartTypeSuffix        = "suffix"
	PartTypeUndetermined  = "undetermined"

	sqlInsertValue = `INSERT INTO name_values (id, form) VALUES (?, ?)`

	sqlInsertPart = `INSERT INTO name_value_parts (id, name_value_id, idx, value, type)
		VALUES (?, ?, ?, ?, ?)`

	sqlLookupValue = `SELECT form FROM name_values WHERE id = ?`

	sqlLookupParts = `SELECT id, idx, value, type FROM name_value_parts
		WHERE name_value_id = ? ORDER BY idx`
)

// Part is one ordered segment of a NameValue.
type Part struct {
	ID    []byte
	Idx   int
	Value string
	Type  string // empty or a KnownPartType key
}

// Value is one name_values row plus optional ordered parts.
type Value struct {
	ID    []byte
	Form  string
	Parts []Part // ordered by idx on Lookup
}

// Insert mints UUIDv7 ids, validates, inserts form + parts in one transaction,
// and returns the parent id.
func Insert(c *database.Catalog, v Value) ([]byte, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	id, err := InsertTx(tx, v)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return id, nil
}

// InsertTx validates and inserts form + parts on an existing transaction.
func InsertTx(tx *sql.Tx, v Value) ([]byte, error) {
	if tx == nil {
		return nil, ErrInvalid
	}
	v.Form = strings.TrimSpace(v.Form)
	parts := make([]Part, len(v.Parts))
	for i, p := range v.Parts {
		parts[i] = Part{
			Idx:   p.Idx,
			Value: strings.TrimSpace(p.Value),
			Type:  strings.TrimSpace(p.Type),
		}
	}
	if err := validate(v.Form, parts); err != nil {
		return nil, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	idBytes := id[:]
	if _, err := tx.Exec(sqlInsertValue, idBytes, v.Form); err != nil {
		return nil, err
	}
	for _, p := range parts {
		partID, err := uuid.NewV7()
		if err != nil {
			return nil, err
		}
		if _, err := tx.Exec(
			sqlInsertPart,
			partID[:],
			idBytes,
			p.Idx,
			p.Value,
			nullIfEmpty(p.Type),
		); err != nil {
			return nil, err
		}
	}
	return idBytes, nil
}

// Lookup returns the name_values row and ordered parts for id, or sql.ErrNoRows.
func Lookup(c *database.Catalog, id []byte) (Value, error) {
	db, err := c.DB()
	if err != nil {
		return Value{}, err
	}
	if len(id) != 16 {
		return Value{}, ErrInvalid
	}
	var v Value
	v.ID = append([]byte(nil), id...)
	if err := db.QueryRow(sqlLookupValue, id).Scan(&v.Form); err != nil {
		return Value{}, err
	}
	rows, err := db.Query(sqlLookupParts, id)
	if err != nil {
		return Value{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			p    Part
			typ  sql.NullString
			part []byte
		)
		if err := rows.Scan(&part, &p.Idx, &p.Value, &typ); err != nil {
			return Value{}, err
		}
		p.ID = append([]byte(nil), part...)
		p.Type = typ.String
		v.Parts = append(v.Parts, p)
	}
	if err := rows.Err(); err != nil {
		return Value{}, err
	}
	return v, nil
}

func validate(form string, parts []Part) error {
	if form == "" {
		return ErrInvalid
	}
	seen := make(map[int]struct{}, len(parts))
	for _, p := range parts {
		if p.Idx < 0 {
			return ErrInvalid
		}
		if p.Value == "" {
			return ErrInvalid
		}
		if p.Type != "" && !KnownPartType(p.Type) {
			return ErrInvalid
		}
		if _, dup := seen[p.Idx]; dup {
			return ErrInvalid
		}
		seen[p.Idx] = struct{}{}
	}
	return nil
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
