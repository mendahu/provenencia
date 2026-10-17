// Package namevalues stores shared genealogical NameValue rows. The value type,
// part-type registry and validation live in the pure package core/namevalue;
// this package holds the queries.
//
// A NameValue always has a full-form normalized reading (`form`). Ordered
// parts are optional. namevalue.Part `type` is a product-known key from the compiled
// registry (see registry.go) when set; empty type means an untyped segment.
// Insert rejects unknown non-empty types. User-minted part-type vocabulary
// (DB rows) is deferred — keep DDL as open TEXT.
//
// NameValues are value objects (UUID for persistence only). Observation saves
// call UpdateTx to rewrite a row an observation already points at, so
// value_name_id stays stable. Other callers attach a new row.
// Format profiles and project defaults live elsewhere (structured-name-model §4).
package namevalues

import (
	"database/sql"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/sqlbatch"
	"github.com/mendahu/provenencia/core/namevalue"
)

const (
	sqlInsertValue = `INSERT INTO name_values (id, form) VALUES (?, ?)`

	sqlInsertPart = `INSERT INTO name_value_parts (id, name_value_id, idx, value, type)
		VALUES (?, ?, ?, ?, ?)`

	sqlUpdateValue = `UPDATE name_values SET form = ? WHERE id = ?`

	sqlDeleteParts = `DELETE FROM name_value_parts WHERE name_value_id = ?`

	sqlValueExists = `SELECT 1 FROM name_values WHERE id = ?`

	sqlLookupValue = `SELECT form FROM name_values WHERE id = ?`

	sqlLookupParts = `SELECT id, idx, value, type FROM name_value_parts
		WHERE name_value_id = ? ORDER BY idx`

	sqlLookupValueID = `SELECT id, form FROM name_values WHERE id IN (`

	sqlLookupPartsByParents = `SELECT id, name_value_id, idx, value, type
		FROM name_value_parts WHERE name_value_id IN (`
)

// Insert mints UUIDv7 ids, validates, inserts form + parts in one transaction,
// and returns the parent id.
func Insert(c *database.Catalog, v namevalue.Value) ([]byte, error) {
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
func InsertTx(tx *sql.Tx, v namevalue.Value) ([]byte, error) {
	if tx == nil {
		return nil, namevalue.ErrInvalid
	}
	v.Form = strings.TrimSpace(v.Form)
	parts := make([]namevalue.Part, len(v.Parts))
	for i, p := range v.Parts {
		parts[i] = namevalue.Part{
			Idx:   p.Idx,
			Value: strings.TrimSpace(p.Value),
			Type:  strings.TrimSpace(p.Type),
		}
	}
	if err := namevalue.Validate(namevalue.Value{Form: v.Form, Parts: parts}); err != nil {
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

// UpdateTx rewrites form and replaces parts for an existing name_values id.
func UpdateTx(tx *sql.Tx, id []byte, v namevalue.Value) error {
	if tx == nil || len(id) != 16 {
		return namevalue.ErrInvalid
	}
	v.Form = strings.TrimSpace(v.Form)
	parts := make([]namevalue.Part, len(v.Parts))
	for i, p := range v.Parts {
		parts[i] = namevalue.Part{
			Idx:   p.Idx,
			Value: strings.TrimSpace(p.Value),
			Type:  strings.TrimSpace(p.Type),
		}
	}
	if err := namevalue.Validate(namevalue.Value{Form: v.Form, Parts: parts}); err != nil {
		return err
	}
	var one int
	if err := tx.QueryRow(sqlValueExists, id).Scan(&one); err != nil {
		if err == sql.ErrNoRows {
			return namevalue.ErrInvalid
		}
		return err
	}
	if _, err := tx.Exec(sqlUpdateValue, v.Form, id); err != nil {
		return err
	}
	if _, err := tx.Exec(sqlDeleteParts, id); err != nil {
		return err
	}
	for _, p := range parts {
		partID, err := uuid.NewV7()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(
			sqlInsertPart,
			partID[:],
			id,
			p.Idx,
			p.Value,
			nullIfEmpty(p.Type),
		); err != nil {
			return err
		}
	}
	return nil
}

// Lookup returns the name_values row and ordered parts for id, or sql.ErrNoRows.
func Lookup(c *database.Catalog, id []byte) (namevalue.Value, error) {
	db, err := c.DB()
	if err != nil {
		return namevalue.Value{}, err
	}
	return LookupTx(db, id)
}

// LookupTx returns the name_values row and ordered parts on an existing connection or transaction.
func LookupTx(q interface {
	QueryRow(query string, args ...any) *sql.Row
	Query(query string, args ...any) (*sql.Rows, error)
}, id []byte) (namevalue.Value, error) {
	if q == nil || len(id) != 16 {
		return namevalue.Value{}, namevalue.ErrInvalid
	}
	var v namevalue.Value
	v.ID = append([]byte(nil), id...)
	if err := q.QueryRow(sqlLookupValue, id).Scan(&v.Form); err != nil {
		return namevalue.Value{}, err
	}
	rows, err := q.Query(sqlLookupParts, id)
	if err != nil {
		return namevalue.Value{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			p    namevalue.Part
			typ  sql.NullString
			part []byte
		)
		if err := rows.Scan(&part, &p.Idx, &p.Value, &typ); err != nil {
			return namevalue.Value{}, err
		}
		p.ID = append([]byte(nil), part...)
		p.Type = typ.String
		v.Parts = append(v.Parts, p)
	}
	if err := rows.Err(); err != nil {
		return namevalue.Value{}, err
	}
	return v, nil
}

// LookupMany returns name_values rows and ordered parts for the given ids.
// Missing ids are omitted. Empty ids returns an empty map.
func LookupMany(c *database.Catalog, ids [][]byte) (map[string]namevalue.Value, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	return LookupManyTx(db, ids)
}

// LookupManyTx is LookupMany on an existing connection or transaction: two
// queries per sqlbatch.Size ids.
func LookupManyTx(db interface {
	Query(query string, args ...any) (*sql.Rows, error)
}, ids [][]byte) (map[string]namevalue.Value, error) {
	ids = database.UniqueBlobIDs(ids)
	out := make(map[string]namevalue.Value, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	err := sqlbatch.Query(db, sqlLookupValueID, `)`, nil, ids, func(rows *sql.Rows) error {
		var (
			v  namevalue.Value
			id []byte
		)
		if err := rows.Scan(&id, &v.Form); err != nil {
			return err
		}
		v.ID = append([]byte(nil), id...)
		out[string(v.ID)] = v
		return nil
	})
	if err != nil {
		return nil, err
	}
	parentIDs := make([][]byte, 0, len(out))
	for _, v := range out {
		parentIDs = append(parentIDs, v.ID)
	}
	// A parent's parts all come back in its batch, in idx order.
	err = sqlbatch.Query(db, sqlLookupPartsByParents, `) ORDER BY name_value_id, idx`, nil, parentIDs, func(rows *sql.Rows) error {
		var (
			p        namevalue.Part
			parentID []byte
			typ      sql.NullString
			part     []byte
		)
		if err := rows.Scan(&part, &parentID, &p.Idx, &p.Value, &typ); err != nil {
			return err
		}
		p.ID = append([]byte(nil), part...)
		p.Type = typ.String
		k := string(parentID)
		v := out[k]
		v.Parts = append(v.Parts, p)
		out[k] = v
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
