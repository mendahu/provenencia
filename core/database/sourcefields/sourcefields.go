// Package sourcefields accesses the source_metadata_fields vocabulary table.
package sourcefields

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/searchindex"
	"github.com/mendahu/provenencia/core/slug"
)

var ErrInvalid = apperr.New(apperr.CodeSourceFieldsInvalid, apperr.KindUser)

// ErrDuplicateKey is returned by Create when the label-derived key already
// names a user-origin field. Carries the colliding key as a param.
var ErrDuplicateKey = apperr.New(apperr.CodeSourceFieldsDuplicateKey, apperr.KindConflict)

// ErrLocked is returned by Update when the target field is plugin-origin
// (project fields — user and provenencia starters — are editable).
var ErrLocked = apperr.New(apperr.CodeSourceFieldsInvalid, apperr.KindUser)

// ErrInUse is returned by Delete when source_metadata still references the field.
var ErrInUse = apperr.New(apperr.CodeSourceFieldsInUse, apperr.KindConflict)

const (
	OriginProvenencia = "provenencia"
	OriginUser        = "user"

	DataTypeText = "text"
	DataTypeDate = "date"

	sqlUpsert = `INSERT INTO source_metadata_fields (id, key, origin, label, data_type, description)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(key, origin) DO UPDATE SET
			label = excluded.label,
			data_type = excluded.data_type,
			description = excluded.description`
	sqlLookup = `SELECT id, key, origin, label, data_type, COALESCE(description, '')
		FROM source_metadata_fields WHERE key = ? AND origin = ?`
	sqlGetByID = `SELECT id, key, origin, label, data_type, COALESCE(description, '')
		FROM source_metadata_fields WHERE id = ?`
	sqlList = `SELECT f.id, f.key, f.origin, f.label, f.data_type, COALESCE(f.description, ''),
			(SELECT COUNT(*) FROM source_metadata m WHERE m.field_id = f.id)
		FROM source_metadata_fields f ORDER BY f.label COLLATE NOCASE, f.origin, f.key`
	sqlUpdate = `UPDATE source_metadata_fields SET label = ?, description = ? WHERE id = ?`
	sqlDelete = `DELETE FROM source_metadata_fields WHERE id = ?`
	sqlInUse  = `SELECT 1 FROM source_metadata WHERE field_id = ? LIMIT 1`
	sqlUsedBy = `SELECT COUNT(*) FROM source_metadata WHERE field_id = ?`
	sqlCountByOrigin = `SELECT origin, COUNT(*) FROM source_metadata_fields GROUP BY origin`
)

// Field is one source_metadata_fields row.
type Field struct {
	ID          []byte
	Key         string
	Origin      string
	Label       string
	DataType    string
	Description string
	// UsedBy is how many source_metadata rows reference this field. Only
	// List and Update populate it; the other readers leave it 0.
	UsedBy int
}

// OriginCounts is the workspace nav / vocabulary-header split for this
// vocabulary. Total is always Seeded + User + Plugin.
type OriginCounts struct {
	Total  int
	Seeded int
	User   int
	Plugin int
}

// Upsert inserts or updates by (key, origin). Mints a UUIDv7 id when ID is empty on insert.
func Upsert(c *database.Catalog, f Field) ([]byte, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	f.Key = strings.TrimSpace(f.Key)
	f.Origin = strings.TrimSpace(f.Origin)
	f.Label = strings.TrimSpace(f.Label)
	f.DataType = strings.TrimSpace(f.DataType)
	f.Description = strings.TrimSpace(f.Description)
	if f.Key == "" || f.Label == "" || !originOK(f.Origin) || !dataTypeOK(f.DataType) {
		return nil, ErrInvalid
	}
	id := f.ID
	if len(id) == 0 {
		existing, err := Lookup(c, f.Key, f.Origin)
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
	if f.Description == "" {
		desc = nil
	} else {
		desc = f.Description
	}
	if _, err := db.Exec(sqlUpsert, id, f.Key, f.Origin, f.Label, f.DataType, desc); err != nil {
		return nil, err
	}
	if err := searchindex.ReprojectSourceField(db, id); err != nil {
		return nil, err
	}
	return append([]byte(nil), id...), nil
}

// Create mints a kebab-case key from label (see core/slug.Kebab) and
// inserts a new user-origin field. Returns ErrInvalid if the label cannot
// form a key, or ErrDuplicateKey (params: the colliding key) if a
// user-origin field with that key already exists.
func Create(c *database.Catalog, label, dataType, description string) (Field, error) {
	key := slug.Kebab(label)
	if key == "" {
		return Field{}, ErrInvalid
	}
	if _, err := Lookup(c, key, OriginUser); err == nil {
		return Field{}, ErrDuplicateKey.WithParams(key)
	} else if !errors.Is(err, sql.ErrNoRows) {
		return Field{}, err
	}
	if _, err := Upsert(c, Field{
		Key: key, Origin: OriginUser, Label: label, DataType: dataType, Description: description,
	}); err != nil {
		return Field{}, err
	}
	return Lookup(c, key, OriginUser)
}

// Update patches label and description for a project field (user or
// provenencia) by id. Key, origin, and data_type are immutable after create.
// dataType must match the existing value (callers still pass it for clarity).
// Returns ErrLocked for plugin-origin fields.
func Update(c *database.Catalog, id []byte, label, dataType, description string) (Field, error) {
	db, err := c.DB()
	if err != nil {
		return Field{}, err
	}
	label = strings.TrimSpace(label)
	dataType = strings.TrimSpace(dataType)
	description = strings.TrimSpace(description)
	if len(id) != 16 || label == "" || !dataTypeOK(dataType) {
		return Field{}, ErrInvalid
	}
	existing, err := GetByID(c, id)
	if err != nil {
		return Field{}, err
	}
	if existing.Origin != OriginUser && existing.Origin != OriginProvenencia {
		return Field{}, ErrLocked
	}
	if dataType != existing.DataType {
		return Field{}, ErrInvalid
	}
	var desc any
	if description == "" {
		desc = nil
	} else {
		desc = description
	}
	if _, err := db.Exec(sqlUpdate, label, desc, id); err != nil {
		return Field{}, err
	}
	if err := searchindex.ReprojectSourceField(db, id); err != nil {
		return Field{}, err
	}
	return GetByID(c, id)
}

// GetByID returns the field with the given id, or sql.ErrNoRows.
func GetByID(c *database.Catalog, id []byte) (Field, error) {
	db, err := c.DB()
	if err != nil {
		return Field{}, err
	}
	if len(id) != 16 {
		return Field{}, ErrInvalid
	}
	var f Field
	err = db.QueryRow(sqlGetByID, id).Scan(
		&f.ID, &f.Key, &f.Origin, &f.Label, &f.DataType, &f.Description,
	)
	if err != nil {
		return Field{}, err
	}
	return f, nil
}

// Lookup returns the field for (key, origin), or sql.ErrNoRows.
func Lookup(c *database.Catalog, key, origin string) (Field, error) {
	db, err := c.DB()
	if err != nil {
		return Field{}, err
	}
	key = strings.TrimSpace(key)
	origin = strings.TrimSpace(origin)
	if key == "" || origin == "" {
		return Field{}, ErrInvalid
	}
	var f Field
	err = db.QueryRow(sqlLookup, key, origin).Scan(
		&f.ID, &f.Key, &f.Origin, &f.Label, &f.DataType, &f.Description,
	)
	if err != nil {
		return Field{}, err
	}
	return f, nil
}

// List returns all source_metadata_fields rows.
func List(c *database.Catalog) ([]Field, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(sqlList)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Field
	for rows.Next() {
		var f Field
		if err := rows.Scan(&f.ID, &f.Key, &f.Origin, &f.Label, &f.DataType, &f.Description, &f.UsedBy); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

// UsedBy reports how many source_metadata rows reference the field.
func UsedBy(c *database.Catalog, id []byte) (int, error) {
	db, err := c.DB()
	if err != nil {
		return 0, err
	}
	var n int
	if err := db.QueryRow(sqlUsedBy, id).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

// CountByOrigin returns vocabulary counts for the workspace sidebar without
// materializing every field row.
func CountByOrigin(c *database.Catalog) (OriginCounts, error) {
	db, err := c.DB()
	if err != nil {
		return OriginCounts{}, err
	}
	rows, err := db.Query(sqlCountByOrigin)
	if err != nil {
		return OriginCounts{}, err
	}
	defer rows.Close()
	var out OriginCounts
	for rows.Next() {
		var origin string
		var n int
		if err := rows.Scan(&origin, &n); err != nil {
			return OriginCounts{}, err
		}
		switch {
		case origin == OriginProvenencia:
			out.Seeded += n
		case origin == OriginUser:
			out.User += n
		default:
			out.Plugin += n
		}
	}
	if err := rows.Err(); err != nil {
		return OriginCounts{}, err
	}
	out.Total = out.Seeded + out.User + out.Plugin
	return out, nil
}

// Delete removes a field by id when no source_metadata rows reference it.
// Cascades suggestion joins. Any origin may be deleted when unused.
func Delete(c *database.Catalog, id []byte) error {
	db, err := c.DB()
	if err != nil {
		return err
	}
	if len(id) != 16 {
		return ErrInvalid
	}
	var one int
	err = db.QueryRow(sqlInUse, id).Scan(&one)
	if err == nil {
		return ErrInUse
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	_, err = db.Exec(sqlDelete, id)
	if err != nil {
		return err
	}
	return searchindex.Delete(db, searchindex.KindSourceField, uuidString(id))
}

func originOK(origin string) bool {
	if origin == OriginProvenencia || origin == OriginUser {
		return true
	}
	return strings.HasPrefix(origin, "plugin:") && len(origin) > len("plugin:")
}

func dataTypeOK(dt string) bool {
	return dt == DataTypeText || dt == DataTypeDate
}

func uuidString(id []byte) string {
	if len(id) != 16 {
		return ""
	}
	var u uuid.UUID
	copy(u[:], id)
	return u.String()
}
