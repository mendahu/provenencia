// Package sourcemetadata stores descriptive Source metadata values with audited
// set/clear, plus the per-Source layout (dismissed suggestions and field order)
// that the Source page metadata editor reads back through ListWorkspace.
package sourcemetadata

import (
	"database/sql"
	"errors"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/audit"
	"github.com/mendahu/provenencia/core/database/project"
	"github.com/mendahu/provenencia/core/database/searchindex"
	"github.com/mendahu/provenencia/core/database/sourcefields"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcevocab"
)

var ErrInvalid = apperr.New(apperr.CodeSourceMetadataInvalid, apperr.KindUser)

const (
	sqlInsert = `INSERT INTO source_metadata (id, source_id, field_id, value_text, date_value_id)
		VALUES (?, ?, ?, ?, ?)`
	sqlUpdate = `UPDATE source_metadata SET value_text = ?, date_value_id = ?
		WHERE id = ?`
	sqlDelete = `DELETE FROM source_metadata WHERE source_id = ? AND field_id = ?`
	sqlGetPair = `SELECT id, source_id, field_id, COALESCE(value_text, ''), date_value_id
		FROM source_metadata WHERE source_id = ? AND field_id = ?`
	sqlListBySource = `SELECT id, source_id, field_id, COALESCE(value_text, ''), date_value_id
		FROM source_metadata WHERE source_id = ?
		ORDER BY field_id`
	sqlSourceExists = `SELECT 1 FROM sources WHERE id = ?`
	sqlFieldGet     = `SELECT id, key, origin, label, data_type, COALESCE(description, '')
		FROM source_metadata_fields WHERE id = ?`
	sqlDateExists = `SELECT 1 FROM date_values WHERE id = ?`

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
	ID          []byte
	SourceID    []byte
	FieldID     []byte
	ValueText   string
	DateValueID []byte // nil when unset
}

// Input is the payload for Set.
type Input struct {
	SourceID    []byte
	FieldID     []byte
	ValueText   string
	DateValueID []byte // nil/empty when unset
}

// WorkspaceEntry is a suggested or extra field for Source edit UI.
type WorkspaceEntry struct {
	Field     sourcefields.Field
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

// Set upserts a metadata value for (source_id, field_id) and records update_source_metadata.
func Set(c *database.Catalog, userID []byte, in Input) (Row, error) {
	db, err := c.DB()
	if err != nil {
		return Row{}, err
	}
	in.ValueText = strings.TrimSpace(in.ValueText)
	if len(in.SourceID) != 16 || len(in.FieldID) != 16 {
		return Row{}, ErrInvalid
	}
	if len(in.DateValueID) != 0 && len(in.DateValueID) != 16 {
		return Row{}, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return Row{}, err
	}

	tx, err := db.Begin()
	if err != nil {
		return Row{}, err
	}
	defer func() { _ = tx.Rollback() }()

	if err := requireSource(tx, in.SourceID); err != nil {
		return Row{}, err
	}
	field, err := getFieldTx(tx, in.FieldID)
	if err != nil {
		return Row{}, err
	}
	if err := validateValue(field.DataType, in.ValueText, in.DateValueID); err != nil {
		return Row{}, err
	}
	if len(in.DateValueID) == 16 {
		if err := requireDate(tx, in.DateValueID); err != nil {
			return Row{}, err
		}
	}

	prev, err := getPairTx(tx, in.SourceID, in.FieldID)
	creating := errors.Is(err, sql.ErrNoRows)
	if err != nil && !creating {
		return Row{}, err
	}

	var row Row
	var action string
	fields := map[string]audit.FieldDiff{}

	if creating {
		id, err := uuid.NewV7()
		if err != nil {
			return Row{}, err
		}
		idBytes := id[:]
		if _, err := tx.Exec(sqlInsert, idBytes, in.SourceID, in.FieldID, nullStr(in.ValueText), nullBlob(in.DateValueID)); err != nil {
			return Row{}, mapConstraint(err)
		}
		// A field the researcher just filled needs a place in the Source's
		// order; appending keeps an existing hand-sorted layout intact.
		if _, err := tx.Exec(sqlLayoutAppend, in.SourceID, in.FieldID, in.SourceID); err != nil {
			return Row{}, mapConstraint(err)
		}
		row = Row{
			ID:          append([]byte(nil), idBytes...),
			SourceID:    append([]byte(nil), in.SourceID...),
			FieldID:     append([]byte(nil), in.FieldID...),
			ValueText:   in.ValueText,
			DateValueID: copyBlob(in.DateValueID),
		}
		action = audit.ActionCreate
		fields["id"] = audit.FieldDiff{Old: nil, New: id.String()}
		fields["source_id"] = audit.FieldDiff{Old: nil, New: uuidString(in.SourceID)}
		fields["field_id"] = audit.FieldDiff{Old: nil, New: uuidString(in.FieldID)}
		if in.ValueText != "" {
			fields["value_text"] = audit.FieldDiff{Old: nil, New: in.ValueText}
		}
		if len(in.DateValueID) == 16 {
			fields["date_value_id"] = audit.FieldDiff{Old: nil, New: uuidString(in.DateValueID)}
		}
	} else {
		if prev.ValueText == in.ValueText && bytesEqual(prev.DateValueID, in.DateValueID) {
			_ = tx.Commit()
			return prev, nil
		}
		if _, err := tx.Exec(sqlUpdate, nullStr(in.ValueText), nullBlob(in.DateValueID), prev.ID); err != nil {
			return Row{}, mapConstraint(err)
		}
		row = Row{
			ID:          append([]byte(nil), prev.ID...),
			SourceID:    append([]byte(nil), in.SourceID...),
			FieldID:     append([]byte(nil), in.FieldID...),
			ValueText:   in.ValueText,
			DateValueID: copyBlob(in.DateValueID),
		}
		action = audit.ActionUpdate
		if prev.ValueText != in.ValueText {
			fields["value_text"] = audit.FieldDiff{Old: nullJSON(prev.ValueText), New: nullJSON(in.ValueText)}
		}
		if !bytesEqual(prev.DateValueID, in.DateValueID) {
			fields["date_value_id"] = audit.FieldDiff{Old: uuidJSON(prev.DateValueID), New: uuidJSON(in.DateValueID)}
		}
	}

	if _, err := audit.Record(tx, audit.Revision{
		UserID:     userID,
		ActionType: "update_source_metadata",
		CreatedAt:  project.NowUTC(),
		Changes: []audit.Change{{
			EntityType: "source_metadata",
			EntityID:   row.ID,
			Action:     action,
			Fields:     fields,
		}},
	}); err != nil {
		return Row{}, err
	}
	if err := searchindex.ReprojectSource(tx, in.SourceID); err != nil {
		return Row{}, err
	}
	if err := tx.Commit(); err != nil {
		return Row{}, err
	}
	return row, nil
}

// Clear deletes the metadata row for (source_id, field_id) if present.
func Clear(c *database.Catalog, userID, sourceID, fieldID []byte) error {
	db, err := c.DB()
	if err != nil {
		return err
	}
	if len(sourceID) != 16 || len(fieldID) != 16 {
		return ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	prev, err := getPairTx(tx, sourceID, fieldID)
	if errors.Is(err, sql.ErrNoRows) {
		return tx.Commit()
	}
	if err != nil {
		return err
	}

	if _, err := tx.Exec(sqlDelete, sourceID, fieldID); err != nil {
		return err
	}
	fields := map[string]audit.FieldDiff{
		"id":        {Old: uuidString(prev.ID), New: nil},
		"source_id": {Old: uuidString(prev.SourceID), New: nil},
		"field_id":  {Old: uuidString(prev.FieldID), New: nil},
	}
	if prev.ValueText != "" {
		fields["value_text"] = audit.FieldDiff{Old: prev.ValueText, New: nil}
	}
	if len(prev.DateValueID) == 16 {
		fields["date_value_id"] = audit.FieldDiff{Old: uuidString(prev.DateValueID), New: nil}
	}
	if _, err := audit.Record(tx, audit.Revision{
		UserID:     userID,
		ActionType: "update_source_metadata",
		CreatedAt:  project.NowUTC(),
		Changes: []audit.Change{{
			EntityType: "source_metadata",
			EntityID:   prev.ID,
			Action:     audit.ActionDelete,
			Fields:     fields,
		}},
	}); err != nil {
		return err
	}
	if err := searchindex.ReprojectSource(tx, sourceID); err != nil {
		return err
	}
	return tx.Commit()
}

// DismissSuggestion hides one of the type's metadata suggestions for a single
// Source and records dismiss_source_metadata_suggestion.
//
// Dismissing is only meaningful while the field is an empty suggestion. A field
// that already holds a value stays visible, so the call is a no-op rather than
// an error — the editor's X is offered on empty suggestion rows only.
func DismissSuggestion(c *database.Catalog, userID, sourceID, fieldID []byte) error {
	db, err := c.DB()
	if err != nil {
		return err
	}
	if len(sourceID) != 16 || len(fieldID) != 16 {
		return ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := requireSource(tx, sourceID); err != nil {
		return err
	}
	if _, err := getFieldTx(tx, fieldID); err != nil {
		return err
	}

	_, err = getPairTx(tx, sourceID, fieldID)
	switch {
	case err == nil:
		return tx.Commit()
	case !errors.Is(err, sql.ErrNoRows):
		return err
	}

	prev, err := getLayoutTx(tx, sourceID, fieldID)
	had := true
	if errors.Is(err, sql.ErrNoRows) {
		had = false
	} else if err != nil {
		return err
	}
	if had && prev.Dismissed {
		return tx.Commit()
	}

	if _, err := tx.Exec(sqlLayoutDismiss, sourceID, fieldID, sourceID); err != nil {
		return mapConstraint(err)
	}
	action := audit.ActionCreate
	var oldDismissed any
	if had {
		action = audit.ActionUpdate
		oldDismissed = false
	}
	if _, err := audit.Record(tx, audit.Revision{
		UserID:     userID,
		ActionType: "dismiss_source_metadata_suggestion",
		CreatedAt:  project.NowUTC(),
		Changes: []audit.Change{{
			EntityType: "source_metadata_layout",
			EntityID:   fieldID,
			Action:     action,
			Fields: map[string]audit.FieldDiff{
				"source_id": {Old: uuidString(sourceID), New: uuidString(sourceID)},
				"field_id":  {Old: uuidString(fieldID), New: uuidString(fieldID)},
				"dismissed": {Old: oldDismissed, New: true},
			},
		}},
	}); err != nil {
		return err
	}
	return tx.Commit()
}

// Reorder rewrites the Source's field order as 0..n-1 over fieldIDs and records
// reorder_source_metadata. Fields left out of the list keep their rows and sort
// after the listed ones in ListWorkspace. Dismissed flags are preserved.
func Reorder(c *database.Catalog, userID, sourceID []byte, fieldIDs [][]byte) error {
	db, err := c.DB()
	if err != nil {
		return err
	}
	if len(sourceID) != 16 || len(fieldIDs) == 0 {
		return ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(fieldIDs))
	for _, fieldID := range fieldIDs {
		if len(fieldID) != 16 {
			return ErrInvalid
		}
		if _, dup := seen[string(fieldID)]; dup {
			return ErrInvalid
		}
		seen[string(fieldID)] = struct{}{}
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := requireSource(tx, sourceID); err != nil {
		return err
	}

	var changes []audit.Change
	for i, fieldID := range fieldIDs {
		if _, err := getFieldTx(tx, fieldID); err != nil {
			return err
		}
		prev, err := getLayoutTx(tx, sourceID, fieldID)
		had := true
		if errors.Is(err, sql.ErrNoRows) {
			had = false
		} else if err != nil {
			return err
		}
		if had && prev.SortOrder == i {
			continue
		}
		if _, err := tx.Exec(sqlLayoutOrder, sourceID, fieldID, i); err != nil {
			return mapConstraint(err)
		}
		action := audit.ActionCreate
		var oldOrder any
		if had {
			action = audit.ActionUpdate
			oldOrder = prev.SortOrder
		}
		changes = append(changes, audit.Change{
			EntityType: "source_metadata_layout",
			EntityID:   fieldID,
			Action:     action,
			Fields: map[string]audit.FieldDiff{
				"source_id":  {Old: uuidString(sourceID), New: uuidString(sourceID)},
				"field_id":   {Old: uuidString(fieldID), New: uuidString(fieldID)},
				"sort_order": {Old: oldOrder, New: i},
			},
		})
	}
	if len(changes) == 0 {
		return tx.Commit()
	}
	if _, err := audit.Record(tx, audit.Revision{
		UserID:     userID,
		ActionType: "reorder_source_metadata",
		CreatedAt:  project.NowUTC(),
		Changes:    changes,
	}); err != nil {
		return err
	}
	return tx.Commit()
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

func validateValue(dataType, valueText string, dateValueID []byte) error {
	hasDate := len(dateValueID) == 16
	hasText := valueText != ""
	switch dataType {
	case sourcefields.DataTypeText:
		if hasDate {
			return ErrInvalid
		}
		if !hasText {
			return ErrInvalid
		}
		return nil
	case sourcefields.DataTypeDate:
		if !hasText && !hasDate {
			return ErrInvalid
		}
		return nil
	default:
		return ErrInvalid
	}
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanRow(row rowScanner) (Row, error) {
	var r Row
	var dateID []byte
	if err := row.Scan(&r.ID, &r.SourceID, &r.FieldID, &r.ValueText, &dateID); err != nil {
		return Row{}, err
	}
	r.DateValueID = dateID
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

func getFieldTx(tx *sql.Tx, fieldID []byte) (sourcefields.Field, error) {
	var f sourcefields.Field
	err := tx.QueryRow(sqlFieldGet, fieldID).Scan(
		&f.ID, &f.Key, &f.Origin, &f.Label, &f.DataType, &f.Description,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return sourcefields.Field{}, ErrInvalid
	}
	return f, err
}

func lookupField(c *database.Catalog, fieldID []byte) (sourcefields.Field, error) {
	db, err := c.DB()
	if err != nil {
		return sourcefields.Field{}, err
	}
	var f sourcefields.Field
	err = db.QueryRow(sqlFieldGet, fieldID).Scan(
		&f.ID, &f.Key, &f.Origin, &f.Label, &f.DataType, &f.Description,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return sourcefields.Field{}, ErrInvalid
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

func requireDate(tx *sql.Tx, dateID []byte) error {
	var one int
	err := tx.QueryRow(sqlDateExists, dateID).Scan(&one)
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

func nullBlob(b []byte) any {
	if len(b) == 0 {
		return nil
	}
	return b
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

func uuidJSON(id []byte) any {
	if len(id) == 0 {
		return nil
	}
	return uuidString(id)
}

func copyBlob(b []byte) []byte {
	if len(b) == 0 {
		return nil
	}
	return append([]byte(nil), b...)
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func mapConstraint(err error) error {
	if database.IsConstraintViolation(err) {
		return ErrInvalid
	}
	return err
}
