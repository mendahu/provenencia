// Package sourcemetadata stores descriptive Source metadata values, plus the
// per-Source layout (dismissed suggestions and field order) that the Source
// page metadata editor reads back through ListWorkspace. Set, Clear,
// DismissSuggestion, and Reorder run inside writes.Run and return the row
// changes that Run records.
package sourcemetadata

import (
	"database/sql"
	"errors"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/metadatafields"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcevocab"
	"github.com/mendahu/provenencia/core/urlshape"
)

var ErrInvalid = apperr.New(apperr.CodeSourceMetadataInvalid, apperr.KindUser)

const (
	sqlInsert = `INSERT INTO source_metadata (id, source_id, field_id, value_text)
		VALUES (?, ?, ?, ?)`
	sqlUpdate = `UPDATE source_metadata SET value_text = ?
		WHERE id = ?`
	sqlDelete  = `DELETE FROM source_metadata WHERE source_id = ? AND field_id = ?`
	sqlGetPair = `SELECT id, source_id, field_id, COALESCE(value_text, '')
		FROM source_metadata WHERE source_id = ? AND field_id = ?`
	sqlListBySource = `SELECT id, source_id, field_id, COALESCE(value_text, '')
		FROM source_metadata WHERE source_id = ?
		ORDER BY field_id`
	sqlSourceExists = `SELECT 1 FROM sources WHERE id = ?`
	sqlFieldGet     = `SELECT id, key, origin, label, data_type, COALESCE(description, '')
		FROM source_metadata_fields WHERE id = ?`

	// Layout rows append at the end of the Source's existing order. The next
	// sort_order is computed inside the INSERT so two appends cannot read the
	// same MAX and share a slot; the WHERE clause is also what lets SQLite
	// parse INSERT…SELECT with an upsert clause.
	sqlLayoutAppend = `INSERT INTO source_metadata_layout (source_id, field_id, sort_order, dismissed)
		SELECT ?, ?, COALESCE(MAX(sort_order), -1) + 1, 0
		FROM source_metadata_layout WHERE source_id = ?
		ON CONFLICT(source_id, field_id) DO NOTHING`
	sqlLayoutDismiss = `INSERT INTO source_metadata_layout (source_id, field_id, sort_order, dismissed)
		SELECT ?, ?, COALESCE(MAX(sort_order), -1) + 1, 1
		FROM source_metadata_layout WHERE source_id = ?
		ON CONFLICT(source_id, field_id) DO UPDATE SET dismissed = 1`
	// Reorder must not disturb dismissed, so the upsert touches sort_order only.
	sqlLayoutOrder = `INSERT INTO source_metadata_layout (source_id, field_id, sort_order)
		VALUES (?, ?, ?)
		ON CONFLICT(source_id, field_id) DO UPDATE SET sort_order = excluded.sort_order`
	sqlLayoutGet  = `SELECT sort_order, dismissed FROM source_metadata_layout WHERE source_id = ? AND field_id = ?`
	sqlLayoutList = `SELECT field_id, sort_order, dismissed FROM source_metadata_layout WHERE source_id = ?`
)

// Row is one source_metadata value.
type Row struct {
	ID        []byte
	SourceID  []byte
	FieldID   []byte
	ValueText string
}

// Input is the payload for Set.
type Input struct {
	SourceID  []byte
	FieldID   []byte
	ValueText string
}

// WorkspaceEntry is a suggested or extra field for Source edit UI.
type WorkspaceEntry struct {
	Field     metadatafields.Field
	Value     *Row // nil when suggested but unset
	Suggested bool
	// SortOrder is the position ListWorkspace returned the entry in: the
	// source_metadata_layout order once the Source has any layout row,
	// otherwise the type's suggestion order (0 for extras).
	SortOrder int
}

// layout is one source_metadata_layout row for a field.
type layout struct {
	SortOrder int
	Dismissed bool
}

// Set upserts a metadata value for (source_id, field_id). The caller records
// the returned changes. An unchanged value returns the existing row and no changes.
func Set(tx *database.Tx, userID []byte, in Input) (Row, []rowchange.Change, error) {
	in.ValueText = strings.TrimSpace(in.ValueText)
	if tx == nil || len(in.SourceID) != 16 || len(in.FieldID) != 16 {
		return Row{}, nil, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return Row{}, nil, err
	}

	if err := requireSource(tx.Tx, in.SourceID); err != nil {
		return Row{}, nil, err
	}
	field, err := getFieldTx(tx.Tx, in.FieldID)
	if err != nil {
		return Row{}, nil, err
	}

	prev, err := getPairTx(tx.Tx, in.SourceID, in.FieldID)
	creating := errors.Is(err, sql.ErrNoRows)
	if err != nil && !creating {
		return Row{}, nil, err
	}

	normalized, err := normalizeValue(field.DataType, in.ValueText)
	if err != nil {
		return Row{}, nil, err
	}
	in.ValueText = normalized

	var row Row
	var action string
	fields := map[string]rowchange.FieldDiff{}

	if creating {
		id, err := uuid.NewV7()
		if err != nil {
			return Row{}, nil, err
		}
		idBytes := id[:]
		if _, err := tx.Exec(sqlInsert, idBytes, in.SourceID, in.FieldID, nullStr(in.ValueText)); err != nil {
			return Row{}, nil, mapConstraint(err)
		}
		// A field the researcher just filled needs a place in the Source's
		// order; appending keeps an existing hand-sorted layout intact.
		if _, err := tx.Exec(sqlLayoutAppend, in.SourceID, in.FieldID, in.SourceID); err != nil {
			return Row{}, nil, mapConstraint(err)
		}
		row = Row{
			ID:        append([]byte(nil), idBytes...),
			SourceID:  append([]byte(nil), in.SourceID...),
			FieldID:   append([]byte(nil), in.FieldID...),
			ValueText: in.ValueText,
		}
		action = rowchange.ActionCreate
		fields["id"] = rowchange.FieldDiff{Old: nil, New: id.String()}
		fields["source_id"] = rowchange.FieldDiff{Old: nil, New: uuidString(in.SourceID)}
		fields["field_id"] = rowchange.FieldDiff{Old: nil, New: uuidString(in.FieldID)}
		if in.ValueText != "" {
			fields["value_text"] = rowchange.FieldDiff{Old: nil, New: in.ValueText}
		}
	} else {
		if prev.ValueText == in.ValueText {
			return prev, nil, nil
		}
		if _, err := tx.Exec(sqlUpdate, nullStr(in.ValueText), prev.ID); err != nil {
			return Row{}, nil, mapConstraint(err)
		}
		row = Row{
			ID:        append([]byte(nil), prev.ID...),
			SourceID:  append([]byte(nil), in.SourceID...),
			FieldID:   append([]byte(nil), in.FieldID...),
			ValueText: in.ValueText,
		}
		action = rowchange.ActionUpdate
		fields["value_text"] = rowchange.FieldDiff{Old: nullJSON(prev.ValueText), New: nullJSON(in.ValueText)}
	}

	return row, []rowchange.Change{{
		EntityType: "source_metadata",
		EntityID:   row.ID,
		Action:     action,
		Fields:     fields,
	}}, nil
}

// Clear deletes the metadata row for (source_id, field_id) when one exists.
// A missing row returns no changes. The caller records the returned changes.
func Clear(tx *database.Tx, userID, sourceID, fieldID []byte) ([]rowchange.Change, error) {
	if tx == nil || len(sourceID) != 16 || len(fieldID) != 16 {
		return nil, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return nil, err
	}

	prev, err := getPairTx(tx.Tx, sourceID, fieldID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if _, err := tx.Exec(sqlDelete, sourceID, fieldID); err != nil {
		return nil, err
	}
	fields := map[string]rowchange.FieldDiff{
		"id":        {Old: uuidString(prev.ID), New: nil},
		"source_id": {Old: uuidString(prev.SourceID), New: nil},
		"field_id":  {Old: uuidString(prev.FieldID), New: nil},
	}
	if prev.ValueText != "" {
		fields["value_text"] = rowchange.FieldDiff{Old: prev.ValueText, New: nil}
	}
	return []rowchange.Change{{
		EntityType: "source_metadata",
		EntityID:   prev.ID,
		Action:     rowchange.ActionDelete,
		Fields:     fields,
	}}, nil
}

// DismissSuggestion hides one of the type's metadata suggestions for a single
// Source. The caller records the returned changes.
//
// Dismissing is only meaningful while the field is an empty suggestion. A field
// that already holds a value stays visible, so the call is a no-op rather than
// an error — the editor's X is offered on empty suggestion rows only. An
// already-dismissed suggestion is also a no-op. Both return no changes.
func DismissSuggestion(tx *database.Tx, userID, sourceID, fieldID []byte) ([]rowchange.Change, error) {
	if tx == nil || len(sourceID) != 16 || len(fieldID) != 16 {
		return nil, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return nil, err
	}

	if err := requireSource(tx.Tx, sourceID); err != nil {
		return nil, err
	}
	if _, err := getFieldTx(tx.Tx, fieldID); err != nil {
		return nil, err
	}

	_, err := getPairTx(tx.Tx, sourceID, fieldID)
	switch {
	case err == nil:
		return nil, nil
	case !errors.Is(err, sql.ErrNoRows):
		return nil, err
	}

	prev, err := getLayoutTx(tx.Tx, sourceID, fieldID)
	had := true
	if errors.Is(err, sql.ErrNoRows) {
		had = false
	} else if err != nil {
		return nil, err
	}
	if had && prev.Dismissed {
		return nil, nil
	}

	if _, err := tx.Exec(sqlLayoutDismiss, sourceID, fieldID, sourceID); err != nil {
		return nil, mapConstraint(err)
	}
	action := rowchange.ActionCreate
	var oldDismissed any
	if had {
		action = rowchange.ActionUpdate
		oldDismissed = false
	}
	return []rowchange.Change{{
		EntityType: "source_metadata_layout",
		EntityID:   fieldID,
		Action:     action,
		Fields: map[string]rowchange.FieldDiff{
			"source_id": {Old: uuidString(sourceID), New: uuidString(sourceID)},
			"field_id":  {Old: uuidString(fieldID), New: uuidString(fieldID)},
			"dismissed": {Old: oldDismissed, New: true},
		},
	}}, nil
}

// Reorder rewrites the Source's field order as 0..n-1 over fieldIDs. The caller
// records the returned changes. Fields left out of the list keep their rows and
// sort after the listed ones in ListWorkspace. Dismissed flags are preserved.
// An order that already matches returns no changes.
func Reorder(tx *database.Tx, userID, sourceID []byte, fieldIDs [][]byte) ([]rowchange.Change, error) {
	if tx == nil || len(sourceID) != 16 || len(fieldIDs) == 0 {
		return nil, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return nil, err
	}
	seen := make(map[string]struct{}, len(fieldIDs))
	for _, fieldID := range fieldIDs {
		if len(fieldID) != 16 {
			return nil, ErrInvalid
		}
		if _, dup := seen[string(fieldID)]; dup {
			return nil, ErrInvalid
		}
		seen[string(fieldID)] = struct{}{}
	}

	if err := requireSource(tx.Tx, sourceID); err != nil {
		return nil, err
	}

	var changes []rowchange.Change
	for i, fieldID := range fieldIDs {
		if _, err := getFieldTx(tx.Tx, fieldID); err != nil {
			return nil, err
		}
		prev, err := getLayoutTx(tx.Tx, sourceID, fieldID)
		had := true
		if errors.Is(err, sql.ErrNoRows) {
			had = false
		} else if err != nil {
			return nil, err
		}
		if had && prev.SortOrder == i {
			continue
		}
		if _, err := tx.Exec(sqlLayoutOrder, sourceID, fieldID, i); err != nil {
			return nil, mapConstraint(err)
		}
		action := rowchange.ActionCreate
		var oldOrder any
		if had {
			action = rowchange.ActionUpdate
			oldOrder = prev.SortOrder
		}
		changes = append(changes, rowchange.Change{
			EntityType: "source_metadata_layout",
			EntityID:   fieldID,
			Action:     action,
			Fields: map[string]rowchange.FieldDiff{
				"source_id":  {Old: uuidString(sourceID), New: uuidString(sourceID)},
				"field_id":   {Old: uuidString(fieldID), New: uuidString(fieldID)},
				"sort_order": {Old: oldOrder, New: i},
			},
		})
	}
	return changes, nil
}

// ListBySource returns all metadata rows for a Source.
func ListBySource(c *database.Catalog, sourceID []byte) ([]Row, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	if len(sourceID) != 16 {
		return nil, ErrInvalid
	}
	rows, err := db.Query(sqlListBySource, sourceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Row
	for rows.Next() {
		r, err := scanRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ListWorkspace returns suggested fields for the Source’s type plus any extra
// stored values, filtered and ordered by the Source's layout.
//
// Dismissed suggestions are omitted while they hold no value. Once the Source
// has any layout row the entries come back in layout order, with fields that
// have no row yet sorting after them in suggestion-then-extra order; a Source
// with no layout at all keeps the original suggestion-then-extra order.
func ListWorkspace(c *database.Catalog, sourceID []byte) ([]WorkspaceEntry, error) {
	if len(sourceID) != 16 {
		return nil, ErrInvalid
	}
	src, err := sources.Get(c, sourceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalid
		}
		return nil, err
	}
	suggestions, err := sourcevocab.ListSuggestions(c, src.SourceTypeID)
	if err != nil {
		return nil, err
	}
	values, err := ListBySource(c, sourceID)
	if err != nil {
		return nil, err
	}
	layoutRows, maxOrder, err := listLayout(c, sourceID)
	if err != nil {
		return nil, err
	}
	byField := make(map[string]Row, len(values))
	for _, v := range values {
		byField[string(v.FieldID)] = v
	}

	out := make([]WorkspaceEntry, 0, len(suggestions)+len(values))
	seen := make(map[string]struct{}, len(suggestions))
	for _, s := range suggestions {
		key := string(s.Field.ID)
		seen[key] = struct{}{}
		entry := WorkspaceEntry{Field: s.Field, Suggested: true, SortOrder: s.SortOrder}
		if v, ok := byField[key]; ok {
			vv := v
			entry.Value = &vv
		}
		if entry.Value == nil && layoutRows[key].Dismissed {
			continue
		}
		out = append(out, entry)
	}
	for _, v := range values {
		key := string(v.FieldID)
		if _, ok := seen[key]; ok {
			continue
		}
		field, err := lookupField(c, v.FieldID)
		if err != nil {
			return nil, err
		}
		vv := v
		out = append(out, WorkspaceEntry{
			Field:     field,
			Value:     &vv,
			Suggested: false,
			SortOrder: 0,
		})
	}

	if len(layoutRows) == 0 {
		return out, nil
	}
	fallback := maxOrder + 1
	for i := range out {
		if l, ok := layoutRows[string(out[i].Field.ID)]; ok {
			out[i].SortOrder = l.SortOrder
			continue
		}
		out[i].SortOrder = fallback + i
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].SortOrder < out[j].SortOrder })
	return out, nil
}

// listLayout reads the Source's layout rows keyed by field id, with the highest
// sort_order in use (-1 when the Source has no layout).
func listLayout(c *database.Catalog, sourceID []byte) (map[string]layout, int, error) {
	db, err := c.DB()
	if err != nil {
		return nil, 0, err
	}
	rows, err := db.Query(sqlLayoutList, sourceID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := map[string]layout{}
	maxOrder := -1
	for rows.Next() {
		var fieldID []byte
		var sortOrder, dismissed int
		if err := rows.Scan(&fieldID, &sortOrder, &dismissed); err != nil {
			return nil, 0, err
		}
		out[string(fieldID)] = layout{SortOrder: sortOrder, Dismissed: dismissed == 1}
		if sortOrder > maxOrder {
			maxOrder = sortOrder
		}
	}
	return out, maxOrder, rows.Err()
}

func normalizeValue(dataType, valueText string) (string, error) {
	if valueText == "" {
		return "", ErrInvalid
	}
	switch dataType {
	case metadatafields.DataTypeText:
		return valueText, nil
	case metadatafields.DataTypeURL:
		canon, err := urlshape.Canonical(valueText)
		if err != nil {
			return "", ErrInvalid
		}
		return canon, nil
	default:
		return "", ErrInvalid
	}
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanRow(row rowScanner) (Row, error) {
	var r Row
	if err := row.Scan(&r.ID, &r.SourceID, &r.FieldID, &r.ValueText); err != nil {
		return Row{}, err
	}
	return r, nil
}

func getPairTx(tx *sql.Tx, sourceID, fieldID []byte) (Row, error) {
	return scanRow(tx.QueryRow(sqlGetPair, sourceID, fieldID))
}

func getLayoutTx(tx *sql.Tx, sourceID, fieldID []byte) (layout, error) {
	var l layout
	var dismissed int
	if err := tx.QueryRow(sqlLayoutGet, sourceID, fieldID).Scan(&l.SortOrder, &dismissed); err != nil {
		return layout{}, err
	}
	l.Dismissed = dismissed == 1
	return l, nil
}

func getFieldTx(tx *sql.Tx, fieldID []byte) (metadatafields.Field, error) {
	var f metadatafields.Field
	err := tx.QueryRow(sqlFieldGet, fieldID).Scan(
		&f.ID, &f.Key, &f.Origin, &f.Label, &f.DataType, &f.Description,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return metadatafields.Field{}, ErrInvalid
	}
	return f, err
}

func lookupField(c *database.Catalog, fieldID []byte) (metadatafields.Field, error) {
	db, err := c.DB()
	if err != nil {
		return metadatafields.Field{}, err
	}
	var f metadatafields.Field
	err = db.QueryRow(sqlFieldGet, fieldID).Scan(
		&f.ID, &f.Key, &f.Origin, &f.Label, &f.DataType, &f.Description,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return metadatafields.Field{}, ErrInvalid
	}
	return f, err
}

func requireSource(tx *sql.Tx, sourceID []byte) error {
	var one int
	err := tx.QueryRow(sqlSourceExists, sourceID).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrInvalid
	}
	return err
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullJSON(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func uuidString(id []byte) string {
	u, err := uuid.FromBytes(id)
	if err != nil {
		return ""
	}
	return u.String()
}

func mapConstraint(err error) error {
	if database.IsConstraintViolation(err) {
		return ErrInvalid
	}
	return err
}
