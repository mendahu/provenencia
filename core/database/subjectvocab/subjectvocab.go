// Package subjectvocab owns Subject-type↔Property bindings and the provenencia
// Interpretation subject registry (create-time Install).
//
// Install upserts shipped Subject types, Properties, and bindings once at
// catalog create. Call it only from onboarding.createCatalog — not on open.
// Capabilities (type-level), presentation tokens, and non-bridge locked bindings
// live in this compiled registry. The connect matrix lives in core/connectrules;
// Connect / ListConnectRules / EdgeEndpoint / Install bridge bindings loop All()
// / BridgeBindings(). Term capabilities (birthday / tree-edge) are deferred.
// Future plugin:<id> modules call connectrules.Register and extend the same shape.
package subjectvocab

import (
	"database/sql"
	"errors"
	"sort"
	"strings"

	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/vocabkey"
)

var ErrInvalid = apperr.New(apperr.CodeSubjectVocabInvalid, apperr.KindUser)

// ErrLocked is returned when removing or assigning against a registry-locked binding.
var ErrLocked = apperr.New(apperr.CodeSubjectVocabLocked, apperr.KindConflict)

// bindingEntity is the audit entity type of a subject_type_properties row,
// keyed by its property id.
const bindingEntity = "subject_type_property"

const (
	sqlEnsureBinding = `INSERT INTO subject_type_properties (subject_type_id, property_id, sort_order)
		VALUES (?, ?, ?)
		ON CONFLICT(subject_type_id, property_id) DO UPDATE SET sort_order = excluded.sort_order`
	sqlListBindings = `SELECT p.id, p.key, p.origin, p.label, COALESCE(p.description, ''), p.value_type, j.sort_order
		FROM subject_type_properties j
		JOIN properties p ON p.id = j.property_id
		WHERE j.subject_type_id = ?
		ORDER BY j.sort_order ASC, p.label COLLATE NOCASE, p.key`
	sqlDeleteBinding = `DELETE FROM subject_type_properties
		WHERE subject_type_id = ? AND property_id = ?`
	sqlAppendBinding = `INSERT INTO subject_type_properties (subject_type_id, property_id, sort_order)
		SELECT ?, ?, COALESCE(MAX(sort_order), -1) + 1
		FROM subject_type_properties WHERE subject_type_id = ?
		ON CONFLICT(subject_type_id, property_id) DO NOTHING`
	sqlBindingOrder    = `SELECT sort_order FROM subject_type_properties WHERE subject_type_id = ? AND property_id = ?`
	sqlBindingOrders   = `SELECT property_id, sort_order FROM subject_type_properties WHERE subject_type_id = ?`
	sqlSetBindingOrder = `UPDATE subject_type_properties SET sort_order = ? WHERE subject_type_id = ? AND property_id = ?`
	sqlCountBindings   = `SELECT COUNT(*) FROM subject_type_properties WHERE subject_type_id = ?`
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

// ConnectEdge is one cited edge on a bridge: property key binds an endpoint type.
type ConnectEdge struct {
	PropertyKey     string
	EndpointTypeKey string
}

// ConnectRule describes one allowed (or refused) connect endpoint pair.
type ConnectRule struct {
	FromTypeKey      string
	ToTypeKey        string
	BridgeTypeKey    string
	EdgePropertyKeys []string
	Edges            []ConnectEdge
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

// AppendBinding attaches a Property to a type at the end of its order and
// returns the audited create change. A pair that is already bound returns no
// change. The caller records the changes (writes.Run).
func AppendBinding(tx *database.Tx, userID, typeID, propertyID []byte) ([]rowchange.Change, error) {
	if tx == nil || len(typeID) != 16 || len(propertyID) != 16 {
		return nil, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return nil, err
	}
	if _, err := subjecttypes.GetByIDTx(tx.Tx, typeID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalid
		}
		return nil, err
	}
	if _, err := properties.GetByIDTx(tx.Tx, propertyID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalid
		}
		return nil, err
	}
	res, err := tx.Exec(sqlAppendBinding, typeID, propertyID, typeID)
	if err != nil {
		return nil, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, nil
	}
	var order int
	if err := tx.QueryRow(sqlBindingOrder, typeID, propertyID).Scan(&order); err != nil {
		return nil, err
	}
	return []rowchange.Change{{
		EntityType: bindingEntity,
		EntityID:   append([]byte(nil), propertyID...),
		Action:     rowchange.ActionCreate,
		Fields:     rowchange.FullRow(bindingFields(typeID, propertyID, order)),
	}}, nil
}

// bindingFields is a binding row's audited columns.
func bindingFields(typeID, propertyID []byte, sortOrder int) map[string]any {
	return map[string]any{
		"subject_type_id": database.UUIDJSON(typeID),
		"property_id":     database.UUIDJSON(propertyID),
		"sort_order":      sortOrder,
	}
}

// ReorderBindings rewrites a type's binding order as 0..n-1 over propertyIDs,
// which must be exactly the type's bound Properties, each once. It returns one
// update change per row whose order moved; an order that already matches
// returns none. The caller records the changes (writes.Run).
func ReorderBindings(tx *database.Tx, userID, typeID []byte, propertyIDs [][]byte) ([]rowchange.Change, error) {
	if tx == nil || len(typeID) != 16 || len(propertyIDs) == 0 {
		return nil, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return nil, err
	}
	if _, err := subjecttypes.GetByIDTx(tx.Tx, typeID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalid
		}
		return nil, err
	}
	rows, err := tx.Query(sqlBindingOrders, typeID)
	if err != nil {
		return nil, err
	}
	bound := map[string]int{}
	for rows.Next() {
		var id []byte
		var order int
		if err := rows.Scan(&id, &order); err != nil {
			_ = rows.Close()
			return nil, err
		}
		bound[string(id)] = order
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if len(propertyIDs) != len(bound) {
		return nil, ErrInvalid
	}
	seen := make(map[string]struct{}, len(propertyIDs))
	for _, id := range propertyIDs {
		if _, ok := bound[string(id)]; !ok {
			return nil, ErrInvalid
		}
		if _, dup := seen[string(id)]; dup {
			return nil, ErrInvalid
		}
		seen[string(id)] = struct{}{}
	}
	var changes []rowchange.Change
	for i, id := range propertyIDs {
		if bound[string(id)] == i {
			continue
		}
		if _, err := tx.Exec(sqlSetBindingOrder, i, typeID, id); err != nil {
			return nil, err
		}
		changes = append(changes, rowchange.Change{
			EntityType: bindingEntity,
			EntityID:   append([]byte(nil), id...),
			Action:     rowchange.ActionUpdate,
			Fields: map[string]rowchange.FieldDiff{
				"subject_type_id": {Old: database.UUIDJSON(typeID), New: database.UUIDJSON(typeID)},
				"property_id":     {Old: database.UUIDJSON(id), New: database.UUIDJSON(id)},
				"sort_order":      {Old: bound[string(id)], New: i},
			},
		})
	}
	return changes, nil
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

// DeleteBinding removes one join row and returns the audited delete change.
// Refuses registry-locked proveniencia seed bindings. An unbound pair returns
// no change. The caller records the changes (writes.Run).
func DeleteBinding(tx *database.Tx, userID, typeID, propertyID []byte) ([]rowchange.Change, error) {
	if tx == nil || len(typeID) != 16 || len(propertyID) != 16 {
		return nil, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return nil, err
	}
	st, err := subjecttypes.GetByIDTx(tx.Tx, typeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalid
		}
		return nil, err
	}
	prop, err := properties.GetByIDTx(tx.Tx, propertyID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalid
		}
		return nil, err
	}
	if LockedBinding(st.Key, prop.Key) && prop.Origin == vocabkey.OriginProvenencia {
		return nil, ErrLocked
	}
	var order int
	err = tx.QueryRow(sqlBindingOrder, typeID, propertyID).Scan(&order)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if _, err := tx.Exec(sqlDeleteBinding, typeID, propertyID); err != nil {
		return nil, err
	}
	return []rowchange.Change{{
		EntityType: bindingEntity,
		EntityID:   append([]byte(nil), propertyID...),
		Action:     rowchange.ActionDelete,
		Fields:     rowchange.DeletedRow(bindingFields(typeID, propertyID, order)),
	}}, nil
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
			Origin:             vocabkey.OriginProvenencia,
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
			Origin:      vocabkey.OriginProvenencia,
			Label:       p.Label,
			Description: p.Description,
			ValueType:   p.ValueType,
			Cardinality: p.Cardinality,
		})
		if err != nil {
			return err
		}
		propIDs[p.Key] = id
	}
	for _, term := range seedTerms {
		propID := propIDs[term.PropertyKey]
		if len(propID) == 0 {
			return ErrInvalid
		}
		if _, err := propertyterms.Upsert(c, propertyterms.Term{
			PropertyID:  propID,
			Key:         term.Key,
			Origin:      vocabkey.OriginProvenencia,
			Label:       term.Label,
			Description: term.Description,
			Directed:    term.Directed,
			InverseKey:  term.Inverse,
			Neutral:     term.Neutral,
		}); err != nil {
			return err
		}
	}
	for _, b := range allBindings() {
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

// BindingsForType returns seed bindings for a type key (compiled registry +
// connectrules-derived bridge bindings).
func BindingsForType(typeKey string) []seedBinding {
	typeKey = strings.TrimSpace(typeKey)
	var out []seedBinding
	for _, b := range allBindings() {
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
	for _, b := range allBindings() {
		if b.TypeKey == typeKey && b.PropertyKey == propertyKey {
			return b.Locked
		}
	}
	return false
}

func allBindings() []seedBinding {
	out := append([]seedBinding(nil), seedBindings...)
	for _, b := range connectrules.BridgeBindings() {
		// Create-time Install only materializes product-origin bridges.
		// Plugin modules will Install their own rows when enabled.
		if b.Origin != vocabkey.OriginProvenencia {
			continue
		}
		out = append(out, seedBinding{
			TypeKey: b.TypeKey, PropertyKey: b.PropertyKey,
			SortOrder: b.SortOrder, Locked: b.Locked,
		})
	}
	return out
}

// Connect returns the connect rule for an ordered pair, or refuse-by-default.
func Connect(fromTypeKey, toTypeKey string) ConnectRule {
	fromTypeKey = strings.TrimSpace(fromTypeKey)
	toTypeKey = strings.TrimSpace(toTypeKey)
	for _, r := range connectrules.All() {
		if r.FromTypeKey == fromTypeKey && r.ToTypeKey == toTypeKey {
			return ruleFromSeed(r)
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
	all := connectrules.All()
	out := make([]ConnectRule, 0, len(all))
	for _, r := range all {
		out = append(out, ruleFromSeed(r))
	}
	return out
}

func ruleFromSeed(r connectrules.Rule) ConnectRule {
	return ConnectRule{
		FromTypeKey:      r.FromTypeKey,
		ToTypeKey:        r.ToTypeKey,
		BridgeTypeKey:    r.BridgeTypeKey,
		EdgePropertyKeys: r.EdgePropertyKeys(),
		Edges:            edgesFromRule(r),
		Disambiguation:   r.Disambiguation,
		Refuse:           r.Refuse,
	}
}

func edgesFromRule(r connectrules.Rule) []ConnectEdge {
	out := make([]ConnectEdge, 0, len(r.Endpoints))
	for _, e := range r.Endpoints {
		out = append(out, ConnectEdge{PropertyKey: e.PropertyKey, EndpointTypeKey: e.TypeKey})
	}
	return out
}

// EdgeEndpoint returns the endpoint type a bridge edge property binds, if the
// pair is a non-refused connect rule.
func EdgeEndpoint(bridgeTypeKey, propertyKey string) (endpointTypeKey string, ok bool) {
	return connectrules.Edge(bridgeTypeKey, propertyKey)
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
