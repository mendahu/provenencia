// Package subjectvocab owns Subject-type↔Property bindings and the provenencia
// Interpretation subject registry (create-time Install).
//
// Install upserts shipped Subject types, Properties, and bindings once at
// catalog create. Call it only from onboarding.createCatalog — not on open.
// Capabilities, presentation tokens, locked bindings, and the connect matrix
// live in the compiled registry and are exposed via lookup helpers (no SQLite
// JSON). Future plugin:<id> modules extend the same registry shape.
package subjectvocab

import (
	"database/sql"
	"errors"
	"sort"
	"strings"

	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
)

var ErrInvalid = apperr.New(apperr.CodeSubjectVocabInvalid, apperr.KindUser)

// ErrLocked is returned when removing or assigning against a registry-locked binding.
var ErrLocked = apperr.New(apperr.CodeSubjectVocabLocked, apperr.KindConflict)

const (
	sqlEnsureBinding = `INSERT INTO subject_type_fields (subject_type_id, property_id, sort_order)
		VALUES (?, ?, ?)
		ON CONFLICT(subject_type_id, property_id) DO UPDATE SET sort_order = excluded.sort_order`
	sqlListBindings = `SELECT p.id, p.key, p.origin, p.label, COALESCE(p.description, ''), p.value_type, j.sort_order
		FROM subject_type_fields j
		JOIN properties p ON p.id = j.property_id
		WHERE j.subject_type_id = ?
		ORDER BY j.sort_order ASC, p.label COLLATE NOCASE, p.key`
	sqlDeleteBinding = `DELETE FROM subject_type_fields
		WHERE subject_type_id = ? AND property_id = ?`
	sqlAppendBinding = `INSERT INTO subject_type_fields (subject_type_id, property_id, sort_order)
		SELECT ?, ?, COALESCE(MAX(sort_order), -1) + 1
		FROM subject_type_fields WHERE subject_type_id = ?
		ON CONFLICT(subject_type_id, property_id) DO NOTHING`
	sqlCountBindings = `SELECT COUNT(*) FROM subject_type_fields WHERE subject_type_id = ?`
)

// Binding is one ordered Property attached to a Subject type.
type Binding struct {
	Property  properties.Property
	SortOrder int
	Locked    bool
}

// TypeInfo is a registry type with capabilities and presentation (compiled, not a row).
type TypeInfo struct {
	Key                      string
	Label                    string
	Description              string
	RefPrefix                string
	CandidateRefPrefix       string
	Role                     string
	Placeable                bool
	PaletteSort              int
	RequiresCitationAtCreate bool
	Presentation             Presentation
}

// Presentation holds design-system token names for Evidence graph chrome.
type Presentation struct {
	L10nKey       string
	IconSymbol    string
	InkToken      string
	TintToken     string
	ChipToken     string
	LineToken     string
	EdgeFromToken string
	EdgeToToken   string
}

// ConnectRule describes one allowed (or refused) connect endpoint pair.
type ConnectRule struct {
	FromTypeKey      string
	ToTypeKey        string
	BridgeTypeKey    string
	EdgePropertyKeys []string
	Disambiguation   string
	Refuse           bool
}

// EnsureBinding inserts or updates the join row's sort_order.
func EnsureBinding(c *database.Catalog, typeID, propertyID []byte, sortOrder int) error {
	db, err := c.DB()
	if err != nil {
		return err
	}
	if len(typeID) != 16 || len(propertyID) != 16 {
		return ErrInvalid
	}
	_, err = db.Exec(sqlEnsureBinding, typeID, propertyID, sortOrder)
	return err
}

// AppendBinding attaches a Property to a type at the end of its order.
// Refuses when the registry marks the proveniencia (type_key, property_key) pair locked
// and it is already bound — callers still use RemoveBinding for detach.
// New user assign of a locked seed pair is allowed only if not already the locked seed
// (locked means cannot remove / cannot drop integrity — assign of same keys is OK for Install).
func AppendBinding(c *database.Catalog, typeID, propertyID []byte) error {
	db, err := c.DB()
	if err != nil {
		return err
	}
	if len(typeID) != 16 || len(propertyID) != 16 {
		return ErrInvalid
	}
	if _, err := subjecttypes.GetByID(c, typeID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalid
		}
		return err
	}
	if _, err := properties.GetByID(c, propertyID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalid
		}
		return err
	}
	_, err = db.Exec(sqlAppendBinding, typeID, propertyID, typeID)
	return err
}

// CountBindings reports how many Properties a type binds.
func CountBindings(c *database.Catalog, typeID []byte) (int, error) {
	db, err := c.DB()
	if err != nil {
		return 0, err
	}
	if len(typeID) != 16 {
		return 0, ErrInvalid
	}
	var n int
	if err := db.QueryRow(sqlCountBindings, typeID).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

// ListBindings returns Properties for a type, ordered by sort_order, with Locked from registry.
func ListBindings(c *database.Catalog, typeID []byte) ([]Binding, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	if len(typeID) != 16 {
		return nil, ErrInvalid
	}
	st, err := subjecttypes.GetByID(c, typeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalid
		}
		return nil, err
	}
	rows, err := db.Query(sqlListBindings, typeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Binding
	for rows.Next() {
		var b Binding
		if err := rows.Scan(
			&b.Property.ID, &b.Property.Key, &b.Property.Origin, &b.Property.Label,
			&b.Property.Description, &b.Property.ValueType, &b.SortOrder,
		); err != nil {
			return nil, err
		}
		b.Locked = LockedBinding(st.Key, b.Property.Key)
		out = append(out, b)
	}
	return out, rows.Err()
}

// DeleteBinding removes one join row. Refuses registry-locked proveniencia seed bindings.
func DeleteBinding(c *database.Catalog, typeID, propertyID []byte) error {
	db, err := c.DB()
	if err != nil {
		return err
	}
	if len(typeID) != 16 || len(propertyID) != 16 {
		return ErrInvalid
	}
	st, err := subjecttypes.GetByID(c, typeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalid
		}
		return err
	}
	prop, err := properties.GetByID(c, propertyID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalid
		}
		return err
	}
	if LockedBinding(st.Key, prop.Key) && prop.Origin == properties.OriginProvenencia {
		return ErrLocked
	}
	_, err = db.Exec(sqlDeleteBinding, typeID, propertyID)
	return err
}

// Install writes the proveniencia Subject type + Property + binding registry.
// Call only at create time (onboarding.createCatalog).
func Install(c *database.Catalog) error {
	if _, err := c.DB(); err != nil {
		return err
	}
	typeIDs := make(map[string][]byte, len(seedTypes))
	for _, t := range seedTypes {
		id, err := subjecttypes.Upsert(c, subjecttypes.Type{
			Key:                t.Key,
			Origin:             subjecttypes.OriginProvenencia,
			Label:              t.Label,
			Description:        t.Description,
			RefPrefix:          t.RefPrefix,
			CandidateRefPrefix: t.CandidateRefPrefix,
		})
		if err != nil {
			return err
		}
		typeIDs[t.Key] = id
	}
	propIDs := make(map[string][]byte, len(seedProperties))
	for _, p := range seedProperties {
		id, err := properties.Upsert(c, properties.Property{
			Key:         p.Key,
			Origin:      properties.OriginProvenencia,
			Label:       p.Label,
			Description: p.Description,
			ValueType:   p.ValueType,
		})
		if err != nil {
			return err
		}
		propIDs[p.Key] = id
	}
	for _, b := range seedBindings {
		typeID := typeIDs[b.TypeKey]
		propID := propIDs[b.PropertyKey]
		if len(typeID) == 0 || len(propID) == 0 {
			return ErrInvalid
		}
		if err := EnsureBinding(c, typeID, propID, b.SortOrder); err != nil {
			return err
		}
	}
	return nil
}

// PlaceableTypes returns registry types placeable on the Evidence graph, palette order.
func PlaceableTypes() []TypeInfo {
	var out []TypeInfo
	for _, t := range seedTypes {
		if !t.Placeable {
			continue
		}
		out = append(out, typeInfoFromSeed(t))
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].PaletteSort < out[j].PaletteSort
	})
	return out
}

// TypeByKey returns registry TypeInfo for a proveniencia type key, or false.
func TypeByKey(key string) (TypeInfo, bool) {
	key = strings.TrimSpace(key)
	for _, t := range seedTypes {
		if t.Key == key {
			return typeInfoFromSeed(t), true
		}
	}
	return TypeInfo{}, false
}

// PresentationFor returns presentation tokens for a type key.
func PresentationFor(key string) (Presentation, bool) {
	info, ok := TypeByKey(key)
	if !ok {
		return Presentation{}, false
	}
	return info.Presentation, true
}

// BindingsForType returns seed bindings for a type key (compiled registry).
func BindingsForType(typeKey string) []seedBinding {
	typeKey = strings.TrimSpace(typeKey)
	var out []seedBinding
	for _, b := range seedBindings {
		if b.TypeKey == typeKey {
			out = append(out, b)
		}
	}
	return out
}

// LockedBinding reports whether the registry locks (typeKey, propertyKey).
func LockedBinding(typeKey, propertyKey string) bool {
	typeKey = strings.TrimSpace(typeKey)
	propertyKey = strings.TrimSpace(propertyKey)
	for _, b := range seedBindings {
		if b.TypeKey == typeKey && b.PropertyKey == propertyKey {
			return b.Locked
		}
	}
	return false
}

// Connect returns the connect rule for an ordered pair, or refuse-by-default.
func Connect(fromTypeKey, toTypeKey string) ConnectRule {
	fromTypeKey = strings.TrimSpace(fromTypeKey)
	toTypeKey = strings.TrimSpace(toTypeKey)
	for _, r := range seedConnect {
		if r.FromTypeKey == fromTypeKey && r.ToTypeKey == toTypeKey {
			return ConnectRule{
				FromTypeKey:      r.FromTypeKey,
				ToTypeKey:        r.ToTypeKey,
				BridgeTypeKey:    r.BridgeTypeKey,
				EdgePropertyKeys: append([]string(nil), r.EdgePropertyKeys...),
				Disambiguation:   r.Disambiguation,
				Refuse:           r.Refuse,
			}
		}
	}
	return ConnectRule{
		FromTypeKey: fromTypeKey,
		ToTypeKey:   toTypeKey,
		Refuse:      true,
	}
}

// ListConnectRules returns the full seed connect matrix.
func ListConnectRules() []ConnectRule {
	out := make([]ConnectRule, 0, len(seedConnect))
	for _, r := range seedConnect {
		out = append(out, ConnectRule{
			FromTypeKey:      r.FromTypeKey,
			ToTypeKey:        r.ToTypeKey,
			BridgeTypeKey:    r.BridgeTypeKey,
			EdgePropertyKeys: append([]string(nil), r.EdgePropertyKeys...),
			Disambiguation:   r.Disambiguation,
			Refuse:           r.Refuse,
		})
	}
	return out
}

// AllTypes returns every seeded type's registry info.
func AllTypes() []TypeInfo {
	out := make([]TypeInfo, 0, len(seedTypes))
	for _, t := range seedTypes {
		out = append(out, typeInfoFromSeed(t))
	}
	return out
}

func typeInfoFromSeed(t seedType) TypeInfo {
	return TypeInfo{
		Key:                      t.Key,
		Label:                    t.Label,
		Description:              t.Description,
		RefPrefix:                t.RefPrefix,
		CandidateRefPrefix:       t.CandidateRefPrefix,
		Role:                     t.Role,
		Placeable:                t.Placeable,
		PaletteSort:              t.PaletteSort,
		RequiresCitationAtCreate: t.RequiresCitationAtCreate,
		Presentation: Presentation{
			L10nKey:       t.Presentation.L10nKey,
			IconSymbol:    t.Presentation.IconSymbol,
			InkToken:      t.Presentation.InkToken,
			TintToken:     t.Presentation.TintToken,
			ChipToken:     t.Presentation.ChipToken,
			LineToken:     t.Presentation.LineToken,
			EdgeFromToken: t.Presentation.EdgeFromToken,
			EdgeToToken:   t.Presentation.EdgeToToken,
		},
	}
}
