// Package sourcetypes accesses the source_types vocabulary table.
package sourcetypes

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/slug"
)

var ErrInvalid = apperr.New(apperr.CodeSourceTypesInvalid, apperr.KindUser)

// ErrDuplicateKey is returned by Create when the label-derived key already
// names a user-origin type. Carries the colliding key as a param.
var ErrDuplicateKey = apperr.New(apperr.CodeSourceTypesDuplicateKey, apperr.KindConflict)

// ErrLocked is returned by Update when the target type is plugin-origin
// (project types — user and provenencia starters — are editable).
var ErrLocked = apperr.New(apperr.CodeSourceTypesInvalid, apperr.KindUser)

// ErrInUse is returned by Delete when sources still reference the type.
var ErrInUse = apperr.New(apperr.CodeSourceTypesInUse, apperr.KindConflict)

const (
	OriginProvenencia = "provenencia"
	OriginUser        = "user"

	sqlUpsert = `INSERT INTO source_types (id, key, origin, label, description, icon_key)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(key, origin) DO UPDATE SET
			label = excluded.label,
			description = excluded.description,
			icon_key = excluded.icon_key`
	sqlLookup = `SELECT id, key, origin, label, COALESCE(description, ''), icon_key
		FROM source_types WHERE key = ? AND origin = ?`
	sqlGetByID = `SELECT id, key, origin, label, COALESCE(description, ''), icon_key
		FROM source_types WHERE id = ?`
	sqlList = `SELECT t.id, t.key, t.origin, t.label, COALESCE(t.description, ''), t.icon_key,
			(SELECT COUNT(*) FROM sources s WHERE s.source_type_id = t.id),
			(SELECT COUNT(*) FROM source_type_metadata_fields j WHERE j.source_type_id = t.id)
		FROM source_types t ORDER BY t.label COLLATE NOCASE, t.origin, t.key`
	sqlUpdate = `UPDATE source_types SET label = ?, description = ?, icon_key = ? WHERE id = ?`
	sqlDelete = `DELETE FROM source_types WHERE id = ?`
	sqlInUse  = `SELECT 1 FROM sources WHERE source_type_id = ? LIMIT 1`
	sqlUsedBy = `SELECT COUNT(*) FROM sources WHERE source_type_id = ?`
	sqlCountByOrigin = `SELECT origin, COUNT(*) FROM source_types GROUP BY origin`
)

// Type is one source_types row.
type Type struct {
	ID          []byte
	Key         string
	Origin      string
	Label       string
	Description string
	// IconKey references the closed design-system type_* set.
	IconKey string
	// UsedBy is how many sources reference this type. Only List and
	// Update populate it; the other readers leave it 0.
	UsedBy int
	// SuggestedFields is how many metadata fields this type suggests. Only
	// List and Update populate it; the other readers leave it 0.
	SuggestedFields int
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
func Upsert(c *database.Catalog, t Type) ([]byte, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	t.Key = strings.TrimSpace(t.Key)
	t.Origin = strings.TrimSpace(t.Origin)
	t.Label = strings.TrimSpace(t.Label)
	t.Description = strings.TrimSpace(t.Description)
	iconKey, err := NormalizeIconKey(t.IconKey)
	if err != nil {
		return nil, err
	}
	t.IconKey = iconKey
	if t.Key == "" || t.Label == "" || !originOK(t.Origin) {
		return nil, ErrInvalid
	}
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
	var desc any
	if t.Description == "" {
		desc = nil
	} else {
		desc = t.Description
	}
	if _, err := db.Exec(sqlUpsert, id, t.Key, t.Origin, t.Label, desc, t.IconKey); err != nil {
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
	var t Type
	err = db.QueryRow(sqlLookup, key, origin).Scan(
		&t.ID, &t.Key, &t.Origin, &t.Label, &t.Description, &t.IconKey,
	)
	if err != nil {
		return Type{}, err
	}
	return t, nil
}

// List returns all source_types rows.
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
		var t Type
		if err := rows.Scan(
			&t.ID, &t.Key, &t.Origin, &t.Label, &t.Description, &t.IconKey,
			&t.UsedBy, &t.SuggestedFields,
		); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// Create mints a kebab-case key from label (see core/slug.Kebab) and
// inserts a new user-origin type. Returns ErrInvalid if the label cannot
// form a key or icon_key is unknown, or ErrDuplicateKey (params: the
// colliding key) if a user-origin type with that key already exists.
func Create(c *database.Catalog, label, description, iconKey string) (Type, error) {
	key := slug.Kebab(label)
	if key == "" {
		return Type{}, ErrInvalid
	}
	if _, err := Lookup(c, key, OriginUser); err == nil {
		return Type{}, ErrDuplicateKey.WithParams(key)
	} else if !errors.Is(err, sql.ErrNoRows) {
		return Type{}, err
	}
	if _, err := Upsert(c, Type{
		Key: key, Origin: OriginUser, Label: label, Description: description, IconKey: iconKey,
	}); err != nil {
		return Type{}, err
	}
	return Lookup(c, key, OriginUser)
}

// Update patches label, description, and icon_key for a project type (user
// or provenencia) by id. Key and origin are immutable after create, so a
// rename keeps existing sources attached. Empty icon_key keeps the row's
// current icon. Returns ErrLocked for plugin-origin types.
func Update(c *database.Catalog, id []byte, label, description, iconKey string) (Type, error) {
	db, err := c.DB()
	if err != nil {
		return Type{}, err
	}
	label = strings.TrimSpace(label)
	description = strings.TrimSpace(description)
	if len(id) != 16 || label == "" {
		return Type{}, ErrInvalid
	}
	existing, err := GetByID(c, id)
	if err != nil {
		return Type{}, err
	}
	if existing.Origin != OriginUser && existing.Origin != OriginProvenencia {
		return Type{}, ErrLocked
	}
	normalizedIcon := existing.IconKey
	if strings.TrimSpace(iconKey) != "" {
		normalizedIcon, err = NormalizeIconKey(iconKey)
		if err != nil {
			return Type{}, err
		}
	} else if normalizedIcon == "" {
		normalizedIcon = DefaultIconKey
	}
	var desc any
	if description == "" {
		desc = nil
	} else {
		desc = description
	}
	if _, err := db.Exec(sqlUpdate, label, desc, normalizedIcon, id); err != nil {
		return Type{}, err
	}
	return GetByID(c, id)
}

// GetByID returns the type with the given id, or sql.ErrNoRows.
func GetByID(c *database.Catalog, id []byte) (Type, error) {
	db, err := c.DB()
	if err != nil {
		return Type{}, err
	}
	if len(id) != 16 {
		return Type{}, ErrInvalid
	}
	var t Type
	err = db.QueryRow(sqlGetByID, id).Scan(
		&t.ID, &t.Key, &t.Origin, &t.Label, &t.Description, &t.IconKey,
	)
	if err != nil {
		return Type{}, err
	}
	return t, nil
}

// UsedBy reports how many sources reference the type.
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
// materializing every type row.
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

// Delete removes a type by id when no sources reference it.
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
	return err
}

func originOK(origin string) bool {
	if origin == OriginProvenencia || origin == OriginUser {
		return true
	}
	return strings.HasPrefix(origin, "plugin:") && len(origin) > len("plugin:")
}
