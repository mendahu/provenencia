// Package namevalues stores shared genealogical NameValue rows. The value type,
// part-type registry and validation live in the pure package core/namevalue;
// this package holds the queries.
//
// NameValues are value objects (UUID for persistence only). Observation saves
// call UpdateTx to rewrite a row an observation already points at, so
// value_name_id stays stable. Other callers attach a new row.
package namevalues

import (
	"database/sql"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/rowchange"
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
	v = normalize(v)
	parts := v.Parts
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
			database.NullString(p.Type),
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
	v = normalize(v)
	parts := v.Parts
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
			database.NullString(p.Type),
		); err != nil {
			return err
		}
	}
	return nil
}

// normalize trims the form and each part's value and type: the form every
// write stores, compares and audits. Parts keep their Idx and carry no id.
func normalize(v namevalue.Value) namevalue.Value {
	v.Form = strings.TrimSpace(v.Form)
	parts := make([]namevalue.Part, len(v.Parts))
	for i, p := range v.Parts {
		parts[i] = namevalue.Part{
			Idx:   p.Idx,
			Value: strings.TrimSpace(p.Value),
			Type:  strings.TrimSpace(p.Type),
		}
	}
	v.Parts = parts
	return v
}

// PutTx stores v for an owner. With a 16-byte prevID (the name the owner holds)
// it rewrites that row in place, so the id stays, and returns prevID with an
// update change naming the fields that differ; when none differs it writes
// nothing and returns a nil change. With no prevID it inserts a new row and
// returns its id with the create change. v is normalized before it is compared
// and audited. An invalid v is namevalue.ErrInvalid and nothing is written.
func PutTx(tx *sql.Tx, prevID []byte, v namevalue.Value) (id []byte, ch *rowchange.Change, err error) {
	if tx == nil {
		return nil, nil, namevalue.ErrInvalid
	}
	v = normalize(v)
	if err := namevalue.Validate(namevalue.Value{Form: v.Form, Parts: v.Parts}); err != nil {
		return nil, nil, err
	}
	if len(prevID) != 16 {
		id, err := InsertTx(tx, v)
		if err != nil {
			return nil, nil, err
		}
		c := CreateChange(id, v)
		return id, &c, nil
	}
	prev, err := LookupTx(tx, prevID)
	if err != nil {
		return nil, nil, err
	}
	next := v
	next.ID = append([]byte(nil), prevID...)
	ch = UpdateChange(prev, next)
	if ch == nil {
		return append([]byte(nil), prevID...), nil, nil
	}
	if err := UpdateTx(tx, prevID, v); err != nil {
		return nil, nil, err
	}
	return append([]byte(nil), prevID...), ch, nil
}

// RowMap is the audit row of a name_values row: its id, form and parts.
func RowMap(id []byte, v namevalue.Value) map[string]any {
	return map[string]any{
		"id":    database.UUIDJSON(id),
		"form":  v.Form,
		"parts": partsJSON(v.Parts),
	}
}

// CreateChange is the audit change of a new name_values row.
func CreateChange(id []byte, v namevalue.Value) rowchange.Change {
	return rowchange.Change{
		EntityType: "name_value",
		EntityID:   append([]byte(nil), id...),
		Action:     rowchange.ActionCreate,
		Fields:     rowchange.FullRow(RowMap(id, v)),
	}
}

// UpdateChange is the audit change from prev to next, naming only the fields
// that differ (form, parts); nil when none does.
func UpdateChange(prev, next namevalue.Value) *rowchange.Change {
	fields := map[string]rowchange.FieldDiff{}
	if prev.Form != next.Form {
		fields["form"] = rowchange.FieldDiff{Old: prev.Form, New: next.Form}
	}
	if !partsEqual(prev.Parts, next.Parts) {
		fields["parts"] = rowchange.FieldDiff{Old: partsJSON(prev.Parts), New: partsJSON(next.Parts)}
	}
	if len(fields) == 0 {
		return nil
	}
	id := prev.ID
	if len(id) != 16 {
		id = next.ID
	}
	return &rowchange.Change{
		EntityType: "name_value",
		EntityID:   append([]byte(nil), id...),
		Action:     rowchange.ActionUpdate,
		Fields:     fields,
	}
}

func partsEqual(a, b []namevalue.Part) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Idx != b[i].Idx || a[i].Value != b[i].Value || a[i].Type != b[i].Type {
			return false
		}
	}
	return true
}

func partsJSON(parts []namevalue.Part) []map[string]any {
	out := make([]map[string]any, len(parts))
	for i, p := range parts {
		out[i] = map[string]any{
			"idx":   p.Idx,
			"value": p.Value,
			"type":  database.NullString(p.Type),
		}
	}
	return out
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

// RowMapTx is the release row-map of the name_values row id for
// deleteimpact.ReleaseReplaced: it looks the row up when called, then returns
// its audit row.
func RowMapTx(tx *sql.Tx, id []byte) func() (map[string]any, error) {
	return func() (map[string]any, error) {
		v, err := LookupTx(tx, id)
		return RowMap(id, v), err
	}
}
