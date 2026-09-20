// Package properties accesses the properties vocabulary table with audited mutations.
package properties

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

var ErrInvalid = apperr.New(apperr.CodePropertiesInvalid, apperr.KindUser)

// ErrDuplicateKey is returned by Create when the label-derived key already
// names a user-origin Property.
var ErrDuplicateKey = apperr.New(apperr.CodePropertiesDuplicateKey, apperr.KindConflict)

// ErrLocked is returned by Update when the target is plugin-origin.
var ErrLocked = apperr.New(apperr.CodePropertiesInvalid, apperr.KindUser)

// ErrInUse is returned by Delete when subject_type_fields still references the Property.
var ErrInUse = apperr.New(apperr.CodePropertiesInUse, apperr.KindConflict)

const (
	OriginProvenencia = "provenencia"
	OriginUser        = "user"

	ValueTypeText     = "text"
	ValueTypeInteger  = "integer"
	ValueTypeDate     = "date"
	ValueTypeName     = "name"
	ValueTypeSubject  = "subject"
	ValueTypeTerm     = "term"

	sqlUpsert = `INSERT INTO properties (id, key, origin, label, description, value_type)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(key, origin) DO UPDATE SET
			label = excluded.label,
			description = excluded.description,
			value_type = excluded.value_type`
	sqlLookup = `SELECT id, key, origin, label, COALESCE(description, ''), value_type
		FROM properties WHERE key = ? AND origin = ?`
	sqlGetByID = `SELECT id, key, origin, label, COALESCE(description, ''), value_type
		FROM properties WHERE id = ?`
	sqlList = `SELECT p.id, p.key, p.origin, p.label, COALESCE(p.description, ''), p.value_type,
			(SELECT COUNT(*) FROM subject_type_fields j WHERE j.property_id = p.id)
		FROM properties p ORDER BY p.label COLLATE NOCASE, p.origin, p.key`
	sqlUpdate = `UPDATE properties SET label = ?, description = ? WHERE id = ?`
	sqlDelete = `DELETE FROM properties WHERE id = ?`
	sqlInUse  = `SELECT 1 FROM (
			SELECT 1 AS x FROM subject_type_fields WHERE property_id = ?
			UNION ALL
			SELECT 1 FROM observations WHERE property_id = ?
		) LIMIT 1`
	sqlUsedBy = `SELECT COUNT(*) FROM subject_type_fields WHERE property_id = ?`
	sqlCountByOrigin = `SELECT origin, COUNT(*) FROM properties GROUP BY origin`
)

// Property is one properties row.
type Property struct {
	ID          []byte
	Key         string
	Origin      string
	Label       string
	Description string
	ValueType   string
	// UsedBy is how many subject_type_fields rows reference this Property.
	// Only List and Update populate it; other readers leave it 0.
	UsedBy int
}

// OriginCounts is the workspace nav / vocabulary-header split.
type OriginCounts struct {
	Total  int
	Seeded int
	User   int
	Plugin int
}

// Upsert inserts or updates by (key, origin). Mints a UUIDv7 id when ID is empty on insert.
// Used by create-time seed; does not write audit.
func Upsert(c *database.Catalog, p Property) ([]byte, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	p.Key = strings.TrimSpace(p.Key)
	p.Origin = strings.TrimSpace(p.Origin)
	p.Label = strings.TrimSpace(p.Label)
	p.Description = strings.TrimSpace(p.Description)
	p.ValueType = strings.TrimSpace(p.ValueType)
	if p.Key == "" || p.Label == "" || !originOK(p.Origin) || !valueTypeOK(p.ValueType) {
		return nil, ErrInvalid
	}
	id := p.ID
	if len(id) == 0 {
		existing, err := Lookup(c, p.Key, p.Origin)
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
	if p.Description == "" {
		desc = nil
	} else {
		desc = p.Description
	}
	if _, err := db.Exec(sqlUpsert, id, p.Key, p.Origin, p.Label, desc, p.ValueType); err != nil {
		return nil, err
	}
	return append([]byte(nil), id...), nil
}

// Create mints a kebab-case key from label and inserts a user-origin Property with audit.
// value_type = term is registry/Install only — Create refuses it for researchers.
func Create(c *database.Catalog, userID []byte, label, valueType, description string) (Property, error) {
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return Property{}, err
	}
	valueType = strings.TrimSpace(valueType)
	if valueType == ValueTypeTerm {
		return Property{}, ErrInvalid
	}
	key := slug.Kebab(label)
	if key == "" {
		return Property{}, ErrInvalid
	}
	if _, err := Lookup(c, key, OriginUser); err == nil {
		return Property{}, ErrDuplicateKey.WithParams(key)
	} else if !errors.Is(err, sql.ErrNoRows) {
		return Property{}, err
	}

	db, err := c.DB()
	if err != nil {
		return Property{}, err
	}
	tx, err := db.Begin()
	if err != nil {
		return Property{}, err
	}
	defer func() { _ = tx.Rollback() }()

	uid, err := uuid.NewV7()
	if err != nil {
		return Property{}, err
	}
	id := uid[:]
	label = strings.TrimSpace(label)
	description = strings.TrimSpace(description)
	if label == "" || !valueTypeOK(valueType) {
		return Property{}, ErrInvalid
	}
	var desc any
	if description == "" {
		desc = nil
	} else {
		desc = description
	}
	if _, err := tx.Exec(sqlUpsert, id, key, OriginUser, label, desc, valueType); err != nil {
		return Property{}, err
	}
	fields := map[string]audit.FieldDiff{
		"id":         {Old: nil, New: uid.String()},
		"key":        {Old: nil, New: key},
		"origin":     {Old: nil, New: OriginUser},
		"label":      {Old: nil, New: label},
		"value_type": {Old: nil, New: valueType},
	}
	if description != "" {
		fields["description"] = audit.FieldDiff{Old: nil, New: description}
	}
	if _, err := audit.Record(tx, audit.Revision{
		UserID:     userID,
		ActionType: "create_property",
		CreatedAt:  project.NowUTC(),
		Changes: []audit.Change{{
			EntityType: "property",
			EntityID:   id,
			Action:     audit.ActionCreate,
			Fields:     fields,
		}},
	}); err != nil {
		return Property{}, err
	}
	if err := tx.Commit(); err != nil {
		return Property{}, err
	}
	return Lookup(c, key, OriginUser)
}

// Update patches label and description for a project Property (user or provenencia).
// Key, origin, and value_type are immutable. Returns ErrLocked for plugin-origin.
func Update(c *database.Catalog, userID, id []byte, label, valueType, description string) (Property, error) {
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return Property{}, err
	}
	db, err := c.DB()
	if err != nil {
		return Property{}, err
	}
	label = strings.TrimSpace(label)
	valueType = strings.TrimSpace(valueType)
	description = strings.TrimSpace(description)
	if len(id) != 16 || label == "" || !valueTypeOK(valueType) {
		return Property{}, ErrInvalid
	}
	existing, err := GetByID(c, id)
	if err != nil {
		return Property{}, err
	}
	if existing.Origin != OriginUser && existing.Origin != OriginProvenencia {
		return Property{}, ErrLocked
	}
	if valueType != existing.ValueType {
		return Property{}, ErrInvalid
	}

	tx, err := db.Begin()
	if err != nil {
		return Property{}, err
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
		existing.UsedBy, _ = UsedBy(c, id)
		return existing, nil
	}
	var desc any
	if description == "" {
		desc = nil
	} else {
		desc = description
	}
	if _, err := tx.Exec(sqlUpdate, label, desc, id); err != nil {
		return Property{}, err
	}
	if _, err := audit.Record(tx, audit.Revision{
		UserID:     userID,
		ActionType: "update_property",
		CreatedAt:  project.NowUTC(),
		Changes: []audit.Change{{
			EntityType: "property",
			EntityID:   id,
			Action:     audit.ActionUpdate,
			Fields:     fields,
		}},
	}); err != nil {
		return Property{}, err
	}
	if err := tx.Commit(); err != nil {
		return Property{}, err
	}
	got, err := GetByID(c, id)
	if err != nil {
		return Property{}, err
	}
	got.UsedBy, err = UsedBy(c, id)
	return got, err
}

// GetByID returns the Property with the given id, or sql.ErrNoRows.
func GetByID(c *database.Catalog, id []byte) (Property, error) {
	db, err := c.DB()
	if err != nil {
		return Property{}, err
	}
	if len(id) != 16 {
		return Property{}, ErrInvalid
	}
	var p Property
	err = db.QueryRow(sqlGetByID, id).Scan(
		&p.ID, &p.Key, &p.Origin, &p.Label, &p.Description, &p.ValueType,
	)
	if err != nil {
		return Property{}, err
	}
	return p, nil
}

// Lookup returns the Property for (key, origin), or sql.ErrNoRows.
func Lookup(c *database.Catalog, key, origin string) (Property, error) {
	db, err := c.DB()
	if err != nil {
		return Property{}, err
	}
	key = strings.TrimSpace(key)
	origin = strings.TrimSpace(origin)
	if key == "" || origin == "" {
		return Property{}, ErrInvalid
	}
	var p Property
	err = db.QueryRow(sqlLookup, key, origin).Scan(
		&p.ID, &p.Key, &p.Origin, &p.Label, &p.Description, &p.ValueType,
	)
	if err != nil {
		return Property{}, err
	}
	return p, nil
}

// List returns all properties rows.
func List(c *database.Catalog) ([]Property, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(sqlList)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Property
	for rows.Next() {
		var p Property
		if err := rows.Scan(&p.ID, &p.Key, &p.Origin, &p.Label, &p.Description, &p.ValueType, &p.UsedBy); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// UsedBy reports how many subject_type_fields rows reference the Property.
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

// CountByOrigin returns vocabulary counts for the workspace sidebar.
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

// Delete removes a Property by id when no subject_type_fields rows reference it.
func Delete(c *database.Catalog, userID, id []byte) error {
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return err
	}
	db, err := c.DB()
	if err != nil {
		return err
	}
	if len(id) != 16 {
		return ErrInvalid
	}
	existing, err := GetByID(c, id)
	if err != nil {
		return err
	}
	var one int
	err = db.QueryRow(sqlInUse, id, id).Scan(&one)
	if err == nil {
		return ErrInUse
	}
	if !errors.Is(err, sql.ErrNoRows) {
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
		"id":         {Old: uuidString(id), New: nil},
		"key":        {Old: existing.Key, New: nil},
		"origin":     {Old: existing.Origin, New: nil},
		"label":      {Old: existing.Label, New: nil},
		"value_type": {Old: existing.ValueType, New: nil},
	}
	if existing.Description != "" {
		fields["description"] = audit.FieldDiff{Old: existing.Description, New: nil}
	}
	if _, err := audit.Record(tx, audit.Revision{
		UserID:     userID,
		ActionType: "delete_property",
		CreatedAt:  project.NowUTC(),
		Changes: []audit.Change{{
			EntityType: "property",
			EntityID:   id,
			Action:     audit.ActionDelete,
			Fields:     fields,
		}},
	}); err != nil {
		return err
	}
	return tx.Commit()
}

func originOK(origin string) bool {
	if origin == OriginProvenencia || origin == OriginUser {
		return true
	}
	return strings.HasPrefix(origin, "plugin:") && len(origin) > len("plugin:")
}

func valueTypeOK(vt string) bool {
	switch vt {
	case ValueTypeText, ValueTypeInteger, ValueTypeDate, ValueTypeName, ValueTypeSubject, ValueTypeTerm:
		return true
	default:
		return false
	}
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
