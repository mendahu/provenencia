// Package sourcemetadata stores descriptive metadata values for an owner (a
// Source, an Artifact or an Evidence graph, which share the same functions),
// plus the per-owner layout (dismissed suggestions and field order)
// that the metadata editor reads back through ListWorkspace. Every function
// takes an Owner, a vocabkey.Level and an id; a level with no store entry is
// ErrInvalid. Set, Clear, DismissSuggestion, Reorder, CreateAll, and Apply run
// inside writes.Run and return the row changes that Run records.
package sourcemetadata

import (
	"bytes"
	"database/sql"
	"errors"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/metadatafields"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sourcevocab"
	"github.com/mendahu/provenencia/core/database/sqlbatch"
	"github.com/mendahu/provenencia/core/urlshape"
	"github.com/mendahu/provenencia/core/vocabkey"
)

var ErrInvalid = apperr.New(apperr.CodeSourceMetadataInvalid, apperr.KindUser)

// Row is one stored metadata value.
type Row struct {
	ID        []byte
	Owner     Owner
	FieldID   []byte
	ValueText string
}

// Input is one value of a Set or a Draft.
type Input struct {
	FieldID   []byte
	ValueText string
}

// Draft is what an owner's metadata form or copy holds: one value per field,
// and layout rows written as given.
type Draft struct {
	Values []Input     // one per field
	Layout []LayoutRow // written verbatim
}

// WorkspaceEntry is a suggested or extra field for the owner's edit UI.
type WorkspaceEntry struct {
	Field     metadatafields.Field
	Value     *Row // nil when suggested but unset
	Suggested bool
	// SortOrder is the position ListWorkspace returned the entry in: the
	// layout order once the owner has any layout row, otherwise the type's
	// suggestion order (0 for extras).
	SortOrder int
}

// OpenHref is the web link to open for this entry: the browser-ready form of a
// url-typed value (urlshape.OpenHref), or "" when the entry has no value, is not
// a url field, or holds a value that is not a web address (a damaged row). It
// is the only link the app opens from metadata.
func (e WorkspaceEntry) OpenHref() string {
	if e.Value == nil || e.Field.DataType != metadatafields.DataTypeURL {
		return ""
	}
	href, err := urlshape.OpenHref(e.Value.ValueText)
	if err != nil {
		return ""
	}
	return href
}

// LayoutRow is one layout row for a field.
type LayoutRow struct {
	FieldID   []byte
	SortOrder int
	Dismissed bool
}

// Set upserts a metadata value for (owner, field). The caller records the
// returned changes. An unchanged value returns the existing row and no changes.
func Set(tx *database.Tx, userID []byte, owner Owner, in Input) (Row, []rowchange.Change, error) {
	st, err := storeFor(owner)
	if err != nil {
		return Row{}, nil, err
	}
	in.ValueText = strings.TrimSpace(in.ValueText)
	if tx == nil || len(in.FieldID) != 16 {
		return Row{}, nil, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return Row{}, nil, err
	}

	if err := requireOwner(tx.Tx, st, owner); err != nil {
		return Row{}, nil, err
	}
	fields, err := loadFields(tx.Tx, [][]byte{in.FieldID})
	if err != nil {
		return Row{}, nil, err
	}
	field := fields[string(in.FieldID)]

	normalized, err := normalizeValue(field.DataType, in.ValueText)
	if err != nil {
		return Row{}, nil, err
	}

	prev, err := getPairTx(tx.Tx, st, owner, in.FieldID)
	creating := errors.Is(err, sql.ErrNoRows)
	if err != nil && !creating {
		return Row{}, nil, err
	}

	if creating {
		row, change, err := insertValueTx(tx.Tx, st, owner, field.ID, normalized)
		if err != nil {
			return Row{}, nil, err
		}
		// A field the researcher just filled needs a place in the owner's
		// order; appending keeps an existing hand-sorted layout intact.
		if _, err := tx.Exec(st.layoutAppend, owner.ID, in.FieldID, owner.ID); err != nil {
			return Row{}, nil, mapConstraint(err)
		}
		return row, []rowchange.Change{change}, nil
	}

	if prev.ValueText == normalized {
		return prev, nil, nil
	}
	row, change, err := updateValueTx(tx.Tx, st, prev, normalized)
	if err != nil {
		return Row{}, nil, err
	}
	return row, []rowchange.Change{change}, nil
}

// insertValueTx inserts the already normalized text as the owner's value for
// the field and returns the row with its create change. It writes no layout
// row; Set appends one and CreateAll writes the caller's.
func insertValueTx(tx *sql.Tx, st *store, owner Owner, fieldID []byte, normalized string) (Row, rowchange.Change, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return Row{}, rowchange.Change{}, err
	}
	idBytes := id[:]
	if _, err := tx.Exec(st.insert, idBytes, owner.ID, fieldID, database.NullString(normalized)); err != nil {
		return Row{}, rowchange.Change{}, mapConstraint(err)
	}
	row := Row{
		ID:        append([]byte(nil), idBytes...),
		Owner:     Owner{Level: owner.Level, ID: append([]byte(nil), owner.ID...)},
		FieldID:   append([]byte(nil), fieldID...),
		ValueText: normalized,
	}
	fields := map[string]rowchange.FieldDiff{
		"id":           {Old: nil, New: id.String()},
		st.ownerColumn: {Old: nil, New: database.UUIDString(owner.ID)},
		"field_id":     {Old: nil, New: database.UUIDString(fieldID)},
	}
	if normalized != "" {
		fields["value_text"] = rowchange.FieldDiff{Old: nil, New: normalized}
	}
	return row, rowchange.Change{
		EntityType: st.values,
		EntityID:   row.ID,
		Action:     rowchange.ActionCreate,
		Fields:     fields,
	}, nil
}

// updateValueTx writes the new, already normalized text over prev's row.
func updateValueTx(tx *sql.Tx, st *store, prev Row, normalized string) (Row, rowchange.Change, error) {
	if _, err := tx.Exec(st.update, database.NullString(normalized), prev.ID); err != nil {
		return Row{}, rowchange.Change{}, mapConstraint(err)
	}
	row := Row{
		ID:        append([]byte(nil), prev.ID...),
		Owner:     Owner{Level: prev.Owner.Level, ID: append([]byte(nil), prev.Owner.ID...)},
		FieldID:   append([]byte(nil), prev.FieldID...),
		ValueText: normalized,
	}
	return row, rowchange.Change{
		EntityType: st.values,
		EntityID:   row.ID,
		Action:     rowchange.ActionUpdate,
		Fields: map[string]rowchange.FieldDiff{
			"value_text": {Old: database.NullString(prev.ValueText), New: database.NullString(normalized)},
		},
	}, nil
}

// deleteValueTx deletes prev's row and returns its delete change.
func deleteValueTx(tx *sql.Tx, st *store, prev Row) (rowchange.Change, error) {
	if _, err := tx.Exec(st.delete, prev.Owner.ID, prev.FieldID); err != nil {
		return rowchange.Change{}, err
	}
	fields := map[string]rowchange.FieldDiff{
		"id":           {Old: database.UUIDString(prev.ID), New: nil},
		st.ownerColumn: {Old: database.UUIDString(prev.Owner.ID), New: nil},
		"field_id":     {Old: database.UUIDString(prev.FieldID), New: nil},
	}
	if prev.ValueText != "" {
		fields["value_text"] = rowchange.FieldDiff{Old: prev.ValueText, New: nil}
	}
	return rowchange.Change{
		EntityType: st.values,
		EntityID:   prev.ID,
		Action:     rowchange.ActionDelete,
		Fields:     fields,
	}, nil
}

// layoutChange builds every layout row change: the owner column and field_id
// are always written ({nil, id} on a create, {id, id} on an update, the shape
// effects scopes by), with diffs beside them. EntityID is the field id.
func layoutChange(st *store, owner Owner, fieldID []byte, action string, diffs map[string]rowchange.FieldDiff) rowchange.Change {
	ownerID, field := database.UUIDString(owner.ID), database.UUIDString(fieldID)
	fields := make(map[string]rowchange.FieldDiff, len(diffs)+2)
	if action == rowchange.ActionCreate {
		fields[st.ownerColumn] = rowchange.FieldDiff{Old: nil, New: ownerID}
		fields["field_id"] = rowchange.FieldDiff{Old: nil, New: field}
	} else {
		fields[st.ownerColumn] = rowchange.FieldDiff{Old: ownerID, New: ownerID}
		fields["field_id"] = rowchange.FieldDiff{Old: field, New: field}
	}
	for k, v := range diffs {
		fields[k] = v
	}
	return rowchange.Change{
		EntityType: st.layout,
		EntityID:   append([]byte(nil), fieldID...),
		Action:     action,
		Fields:     fields,
	}
}

// insertLayoutTx writes a layout row as given and returns its create change.
func insertLayoutTx(tx *sql.Tx, st *store, owner Owner, l LayoutRow) (rowchange.Change, error) {
	if _, err := tx.Exec(st.layoutInsert, owner.ID, l.FieldID, l.SortOrder, boolInt(l.Dismissed)); err != nil {
		return rowchange.Change{}, mapConstraint(err)
	}
	return layoutChange(st, owner, l.FieldID, rowchange.ActionCreate, map[string]rowchange.FieldDiff{
		"sort_order": {Old: nil, New: l.SortOrder},
		"dismissed":  {Old: nil, New: l.Dismissed},
	}), nil
}

// updateLayoutTx writes next over prev's layout row and returns its update
// change, which carries only the columns that differ.
func updateLayoutTx(tx *sql.Tx, st *store, owner Owner, prev, next LayoutRow) (rowchange.Change, error) {
	if _, err := tx.Exec(st.layoutUpdate, next.SortOrder, boolInt(next.Dismissed), owner.ID, next.FieldID); err != nil {
		return rowchange.Change{}, mapConstraint(err)
	}
	diffs := map[string]rowchange.FieldDiff{}
	if prev.SortOrder != next.SortOrder {
		diffs["sort_order"] = rowchange.FieldDiff{Old: prev.SortOrder, New: next.SortOrder}
	}
	if prev.Dismissed != next.Dismissed {
		diffs["dismissed"] = rowchange.FieldDiff{Old: prev.Dismissed, New: next.Dismissed}
	}
	return layoutChange(st, owner, next.FieldID, rowchange.ActionUpdate, diffs), nil
}

// CreateAll writes values and layout rows verbatim onto an owner that has
// neither, so a copy reads back in the original's order with the original's
// dismissed flags. Unlike Set it appends no layout row of its own. It is Apply
// on an empty owner: changes are the value changes in input order, then the
// layout changes in input order. The caller records them.
func CreateAll(tx *database.Tx, userID []byte, owner Owner, d Draft) ([]rowchange.Change, error) {
	st, err := storeFor(owner)
	if err != nil {
		return nil, err
	}
	if tx == nil {
		return nil, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return nil, err
	}
	if len(d.Values) == 0 && len(d.Layout) == 0 {
		return nil, nil
	}
	have, err := readOwnerTx(tx.Tx, st, owner)
	if err != nil {
		return nil, err
	}
	if len(have.values) > 0 || len(have.layout) > 0 {
		return nil, ErrInvalid
	}
	return apply(tx, userID, st, owner, have, d)
}

// Apply saves a draft: it makes the owner's stored values equal d.Values and
// makes each layout row d.Layout lists equal that row, as given. A value the
// draft leaves out is deleted (an empty draft clears every value); a layout row
// the draft leaves out stays, so a draft never undoes a dismiss by omission.
// Apply appends no layout row for a new value: a caller with a board sends the
// whole order as d.Layout. Only differing rows are written. The caller records
// the returned changes.
func Apply(tx *database.Tx, userID []byte, owner Owner, d Draft) ([]rowchange.Change, error) {
	st, err := storeFor(owner)
	if err != nil {
		return nil, err
	}
	if tx == nil {
		return nil, ErrInvalid
	}
	have, err := readOwnerTx(tx.Tx, st, owner)
	if err != nil {
		return nil, err
	}
	return apply(tx, userID, st, owner, have, d)
}

// checkFieldIDs is ErrInvalid unless every id is 16 bytes and listed once. It
// returns the set of ids.
func checkFieldIDs(ids [][]byte) (map[string]struct{}, error) {
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		if len(id) != 16 {
			return nil, ErrInvalid
		}
		if _, dup := seen[string(id)]; dup {
			return nil, ErrInvalid
		}
		seen[string(id)] = struct{}{}
	}
	return seen, nil
}

// apply checks the whole draft before it writes anything, then writes the
// value changes (deletes in field-id order, then inserts and updates in draft
// order) and the layout changes (inserts and updates in draft order).
func apply(tx *database.Tx, userID []byte, st *store, owner Owner, have stored, d Draft) ([]rowchange.Change, error) {
	if tx == nil {
		return nil, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return nil, err
	}
	if err := requireOwner(tx.Tx, st, owner); err != nil {
		return nil, err
	}

	valueIDs := make([][]byte, len(d.Values))
	for i, in := range d.Values {
		valueIDs[i] = in.FieldID
	}
	inValues, err := checkFieldIDs(valueIDs)
	if err != nil {
		return nil, err
	}
	layoutIDs := make([][]byte, len(d.Layout))
	for i, l := range d.Layout {
		layoutIDs[i] = l.FieldID
	}
	if _, err := checkFieldIDs(layoutIDs); err != nil {
		return nil, err
	}
	ids := append(valueIDs, layoutIDs...)
	var fields map[string]metadatafields.Field
	if len(ids) > 0 {
		if fields, err = loadFields(tx.Tx, ids); err != nil {
			return nil, err
		}
	}
	normalized := make([]string, len(d.Values))
	for i, in := range d.Values {
		n, err := normalizeValue(fields[string(in.FieldID)].DataType, strings.TrimSpace(in.ValueText))
		if err != nil {
			return nil, err
		}
		normalized[i] = n
	}

	haveValue := make(map[string]Row, len(have.values))
	for _, r := range have.values {
		haveValue[string(r.FieldID)] = r
	}
	haveLayout := make(map[string]LayoutRow, len(have.layout))
	for _, l := range have.layout {
		haveLayout[string(l.FieldID)] = l
	}

	var changes []rowchange.Change
	for _, prev := range have.values {
		if _, kept := inValues[string(prev.FieldID)]; kept {
			continue
		}
		change, err := deleteValueTx(tx.Tx, st, prev)
		if err != nil {
			return nil, err
		}
		changes = append(changes, change)
	}
	for i, in := range d.Values {
		prev, ok := haveValue[string(in.FieldID)]
		switch {
		case !ok:
			_, change, err := insertValueTx(tx.Tx, st, owner, in.FieldID, normalized[i])
			if err != nil {
				return nil, err
			}
			changes = append(changes, change)
		case prev.ValueText != normalized[i]:
			_, change, err := updateValueTx(tx.Tx, st, prev, normalized[i])
			if err != nil {
				return nil, err
			}
			changes = append(changes, change)
		}
	}
	for _, l := range d.Layout {
		prev, ok := haveLayout[string(l.FieldID)]
		switch {
		case !ok:
			change, err := insertLayoutTx(tx.Tx, st, owner, l)
			if err != nil {
				return nil, err
			}
			changes = append(changes, change)
		case prev.SortOrder != l.SortOrder || prev.Dismissed != l.Dismissed:
			change, err := updateLayoutTx(tx.Tx, st, owner, prev, l)
			if err != nil {
				return nil, err
			}
			changes = append(changes, change)
		}
	}
	return changes, nil
}

// Clear deletes the metadata row for (owner, field) when one exists. A missing
// row, or a missing owner, returns no changes. The caller records the returned
// changes.
func Clear(tx *database.Tx, userID []byte, owner Owner, fieldID []byte) ([]rowchange.Change, error) {
	st, err := storeFor(owner)
	if err != nil {
		return nil, err
	}
	if tx == nil || len(fieldID) != 16 {
		return nil, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return nil, err
	}

	prev, err := getPairTx(tx.Tx, st, owner, fieldID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	change, err := deleteValueTx(tx.Tx, st, prev)
	if err != nil {
		return nil, err
	}
	return []rowchange.Change{change}, nil
}

// DismissSuggestion hides one of the type's metadata suggestions for a single
// owner. The caller records the returned changes.
//
// Dismissing is only meaningful while the field is an empty suggestion. A field
// that already holds a value stays visible, so the call is a no-op rather than
// an error — the editor's X is offered on empty suggestion rows only. An
// already-dismissed suggestion is also a no-op. Both return no changes.
func DismissSuggestion(tx *database.Tx, userID []byte, owner Owner, fieldID []byte) ([]rowchange.Change, error) {
	st, err := storeFor(owner)
	if err != nil {
		return nil, err
	}
	if tx == nil || len(fieldID) != 16 {
		return nil, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return nil, err
	}

	if err := requireOwner(tx.Tx, st, owner); err != nil {
		return nil, err
	}
	if _, err := loadFields(tx.Tx, [][]byte{fieldID}); err != nil {
		return nil, err
	}

	_, err = getPairTx(tx.Tx, st, owner, fieldID)
	switch {
	case err == nil:
		return nil, nil
	case !errors.Is(err, sql.ErrNoRows):
		return nil, err
	}

	prev, err := getLayoutTx(tx.Tx, st, owner, fieldID)
	had := true
	if errors.Is(err, sql.ErrNoRows) {
		had = false
	} else if err != nil {
		return nil, err
	}
	if had && prev.Dismissed {
		return nil, nil
	}

	if _, err := tx.Exec(st.layoutDismiss, owner.ID, fieldID, owner.ID); err != nil {
		return nil, mapConstraint(err)
	}
	action := rowchange.ActionCreate
	var oldDismissed any
	if had {
		action = rowchange.ActionUpdate
		oldDismissed = false
	}
	return []rowchange.Change{layoutChange(st, owner, fieldID, action, map[string]rowchange.FieldDiff{
		"dismissed": {Old: oldDismissed, New: true},
	})}, nil
}

// Reorder rewrites the owner's field order as 0..n-1 over fieldIDs. The caller
// records the returned changes. Fields left out of the list keep their rows and
// sort after the listed ones in ListWorkspace. Dismissed flags are preserved.
// An order that already matches returns no changes.
func Reorder(tx *database.Tx, userID []byte, owner Owner, fieldIDs [][]byte) ([]rowchange.Change, error) {
	st, err := storeFor(owner)
	if err != nil {
		return nil, err
	}
	if tx == nil || len(fieldIDs) == 0 {
		return nil, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return nil, err
	}
	if _, err := checkFieldIDs(fieldIDs); err != nil {
		return nil, err
	}

	if err := requireOwner(tx.Tx, st, owner); err != nil {
		return nil, err
	}
	if _, err := loadFields(tx.Tx, fieldIDs); err != nil {
		return nil, err
	}

	var changes []rowchange.Change
	for i, fieldID := range fieldIDs {
		prev, err := getLayoutTx(tx.Tx, st, owner, fieldID)
		had := true
		if errors.Is(err, sql.ErrNoRows) {
			had = false
		} else if err != nil {
			return nil, err
		}
		if had && prev.SortOrder == i {
			continue
		}
		if _, err := tx.Exec(st.layoutOrder, owner.ID, fieldID, i); err != nil {
			return nil, mapConstraint(err)
		}
		action := rowchange.ActionCreate
		var oldOrder any
		if had {
			action = rowchange.ActionUpdate
			oldOrder = prev.SortOrder
		}
		changes = append(changes, layoutChange(st, owner, fieldID, action, map[string]rowchange.FieldDiff{
			"sort_order": {Old: oldOrder, New: i},
		}))
	}
	return changes, nil
}

// stored is what an owner holds, with the row ids and old text the change rows
// need: values by field id, layout by sort_order then field id.
type stored struct {
	values []Row
	layout []LayoutRow
}

// readOwnerTx reads what the owner holds. It reads rows, not the owner, so a
// missing owner holds nothing.
func readOwnerTx(tx *sql.Tx, st *store, owner Owner) (stored, error) {
	owners := []Owner{owner}
	values, err := listByOwners(tx, owners)
	if err != nil {
		return stored{}, err
	}
	layout, err := listLayoutByOwners(tx, owners)
	if err != nil {
		return stored{}, err
	}
	return stored{values: values, layout: layout[owner.key()]}, nil
}

// ReadDraftTx returns what the owner holds as a Draft: its values in field-id
// order and its layout rows by sort_order then field id. An owner with no row
// holds nothing, so the Draft is empty and there is no error.
func ReadDraftTx(tx *sql.Tx, owner Owner) (Draft, error) {
	st, err := storeFor(owner)
	if err != nil {
		return Draft{}, err
	}
	if tx == nil {
		return Draft{}, ErrInvalid
	}
	have, err := readOwnerTx(tx, st, owner)
	if err != nil {
		return Draft{}, err
	}
	var d Draft
	for _, v := range have.values {
		d.Values = append(d.Values, Input{FieldID: v.FieldID, ValueText: v.ValueText})
	}
	d.Layout = have.layout
	return d, nil
}

// ListByOwnersTx returns the stored values of many owners on an open
// transaction, one query per level: each owner's rows by field id, the owners
// in the order of their ids within a level, each row carrying its Owner. A
// repeated owner counts once. No owners returns nil without a query. An owner
// at a level with no store is ErrInvalid before any query.
func ListByOwnersTx(tx *sql.Tx, owners []Owner) ([]Row, error) {
	for _, o := range owners {
		if _, err := storeFor(o); err != nil {
			return nil, err
		}
	}
	if tx == nil {
		return nil, ErrInvalid
	}
	if len(owners) == 0 {
		return nil, nil
	}
	return listByOwners(tx, owners)
}

// listByOwners reads the values of the owners, one query per level, each
// owner's rows by field id. The owners must already have passed storeFor.
func listByOwners(q database.Querier, owners []Owner) ([]Row, error) {
	var out []Row
	err := eachLevel(owners, func(st *store, ids [][]byte) error {
		return sqlbatch.Query(q, st.listIn, st.listSuffix, nil, ids, func(rows *sql.Rows) error {
			r, err := scanRow(rows, st)
			if err != nil {
				return err
			}
			out = append(out, r)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// listLayoutByOwners reads the layout rows of the owners, one query per level
// and batch, keyed by Owner.key and ordered by sort_order then field id. The
// owners must already have passed storeFor.
func listLayoutByOwners(q database.Querier, owners []Owner) (map[string][]LayoutRow, error) {
	out := make(map[string][]LayoutRow)
	err := eachLevel(owners, func(st *store, ids [][]byte) error {
		return sqlbatch.Query(q, st.layoutListIn, st.layoutListSuffix, nil, ids, func(rows *sql.Rows) error {
			var ownerID []byte
			var l LayoutRow
			var dismissed int
			if err := rows.Scan(&ownerID, &l.FieldID, &l.SortOrder, &dismissed); err != nil {
				return err
			}
			l.Dismissed = dismissed == 1
			k := Owner{Level: st.level, ID: ownerID}.key()
			out[k] = append(out[k], l)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// eachLevel calls fn for each level that keeps metadata and has owners, with
// the level's store and its distinct owner ids in byte order. SQLite orders
// BLOBs by memcmp, so a query that orders by owner id, run batch by batch over
// these ids, returns rows in one order across the batches.
func eachLevel(owners []Owner, fn func(st *store, ids [][]byte) error) error {
	for _, level := range vocabkey.MetadataLevels() {
		st := stores[level]
		if st == nil {
			continue
		}
		ids := idsAt(owners, level)
		if len(ids) == 0 {
			continue
		}
		sort.Slice(ids, func(i, j int) bool { return bytes.Compare(ids[i], ids[j]) < 0 })
		if err := fn(st, ids); err != nil {
			return err
		}
	}
	return nil
}

// idsAt is the distinct ids of the owners at one level.
func idsAt(owners []Owner, level vocabkey.Level) [][]byte {
	var ids [][]byte
	for _, o := range owners {
		if o.Level == level {
			ids = append(ids, o.ID)
		}
	}
	return database.UniqueBlobIDs(ids)
}

// ListWorkspace returns suggested fields for the owner's type plus any extra
// stored values, filtered and ordered by the owner's layout. A missing owner is
// ErrInvalid.
//
// Dismissed suggestions are omitted while they hold no value. Once the owner
// has any layout row the entries come back in layout order, with fields that
// have no row yet sorting after them in suggestion-then-extra order; an owner
// with no layout at all keeps the original suggestion-then-extra order.
func ListWorkspace(c *database.Catalog, owner Owner) ([]WorkspaceEntry, error) {
	if _, err := storeFor(owner); err != nil {
		return nil, err
	}
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	byOwner, err := listWorkspaces(tx, []Owner{owner})
	if err != nil {
		return nil, err
	}
	return byOwner[owner.key()], nil
}

// listWorkspaces is ListWorkspace over many owners, keyed by Owner.key: each
// level's types in one query, the suggestions once per distinct (type, level),
// then the values, the layout and every extra value's field in one read each.
func listWorkspaces(tx *sql.Tx, owners []Owner) (map[string][]WorkspaceEntry, error) {
	var uniq []Owner
	seenOwner := make(map[string]struct{}, len(owners))
	for _, o := range owners {
		if _, err := storeFor(o); err != nil {
			return nil, err
		}
		if _, dup := seenOwner[o.key()]; dup {
			continue
		}
		seenOwner[o.key()] = struct{}{}
		uniq = append(uniq, o)
	}

	typeOf := make(map[string][]byte, len(uniq))
	err := eachLevel(uniq, func(st *store, ids [][]byte) error {
		return sqlbatch.Query(tx, st.typeIn, st.typeSuffix, nil, ids, func(rows *sql.Rows) error {
			var id, typeID []byte
			if err := rows.Scan(&id, &typeID); err != nil {
				return err
			}
			typeOf[Owner{Level: st.level, ID: id}.key()] = typeID
			return nil
		})
	})
	if err != nil {
		return nil, err
	}

	type typeLevel struct {
		typeID string
		level  vocabkey.Level
	}
	suggested := make(map[typeLevel][]sourcevocab.Suggestion)
	for _, o := range uniq {
		typeID, ok := typeOf[o.key()]
		if !ok {
			return nil, ErrInvalid
		}
		k := typeLevel{string(typeID), o.Level}
		if _, ok := suggested[k]; ok {
			continue
		}
		list, err := sourcevocab.ListSuggestionsTx(tx, typeID, o.Level)
		if err != nil {
			return nil, err
		}
		suggested[k] = list
	}

	valueRows, err := listByOwners(tx, uniq)
	if err != nil {
		return nil, err
	}
	valuesOf := make(map[string][]Row, len(uniq))
	for _, v := range valueRows {
		valuesOf[v.Owner.key()] = append(valuesOf[v.Owner.key()], v)
	}
	layoutOf, err := listLayoutByOwners(tx, uniq)
	if err != nil {
		return nil, err
	}

	var extraIDs [][]byte
	for _, o := range uniq {
		inSuggestions := make(map[string]struct{})
		for _, s := range suggested[typeLevel{string(typeOf[o.key()]), o.Level}] {
			inSuggestions[string(s.Field.ID)] = struct{}{}
		}
		for _, v := range valuesOf[o.key()] {
			if _, ok := inSuggestions[string(v.FieldID)]; !ok {
				extraIDs = append(extraIDs, v.FieldID)
			}
		}
	}
	var extras map[string]metadatafields.Field
	if len(extraIDs) > 0 {
		if extras, err = loadFields(tx, extraIDs); err != nil {
			return nil, err
		}
	}

	out := make(map[string][]WorkspaceEntry, len(uniq))
	for _, o := range uniq {
		suggestions := suggested[typeLevel{string(typeOf[o.key()]), o.Level}]
		values := valuesOf[o.key()]
		layoutRows := make(map[string]LayoutRow, len(layoutOf[o.key()]))
		maxOrder := -1
		for _, l := range layoutOf[o.key()] {
			layoutRows[string(l.FieldID)] = l
			if l.SortOrder > maxOrder {
				maxOrder = l.SortOrder
			}
		}
		byField := make(map[string]Row, len(values))
		for _, v := range values {
			byField[string(v.FieldID)] = v
		}

		entries := make([]WorkspaceEntry, 0, len(suggestions)+len(values))
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
			entries = append(entries, entry)
		}
		for _, v := range values {
			if _, ok := seen[string(v.FieldID)]; ok {
				continue
			}
			vv := v
			entries = append(entries, WorkspaceEntry{
				Field:     extras[string(v.FieldID)],
				Value:     &vv,
				Suggested: false,
				SortOrder: 0,
			})
		}

		if len(layoutRows) > 0 {
			fallback := maxOrder + 1
			for i := range entries {
				if l, ok := layoutRows[string(entries[i].Field.ID)]; ok {
					entries[i].SortOrder = l.SortOrder
					continue
				}
				entries[i].SortOrder = fallback + i
			}
			sort.SliceStable(entries, func(i, j int) bool { return entries[i].SortOrder < entries[j].SortOrder })
		}
		out[o.key()] = entries
	}
	return out, nil
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

func scanRow(row rowScanner, st *store) (Row, error) {
	var r Row
	var ownerID []byte
	if err := row.Scan(&r.ID, &ownerID, &r.FieldID, &r.ValueText); err != nil {
		return Row{}, err
	}
	r.Owner = Owner{Level: st.level, ID: ownerID}
	return r, nil
}

func getPairTx(tx *sql.Tx, st *store, owner Owner, fieldID []byte) (Row, error) {
	return scanRow(tx.QueryRow(st.getPair, owner.ID, fieldID), st)
}

func scanLayoutRow(row rowScanner) (LayoutRow, error) {
	var l LayoutRow
	var dismissed int
	if err := row.Scan(&l.FieldID, &l.SortOrder, &dismissed); err != nil {
		return LayoutRow{}, err
	}
	l.Dismissed = dismissed == 1
	return l, nil
}

func getLayoutTx(tx *sql.Tx, st *store, owner Owner, fieldID []byte) (LayoutRow, error) {
	return scanLayoutRow(tx.QueryRow(st.layoutGet, owner.ID, fieldID))
}

// loadFields reads the fields with the given ids in one query, keyed by id; an
// id with no field is ErrInvalid.
func loadFields(tx *sql.Tx, ids [][]byte) (map[string]metadatafields.Field, error) {
	found, err := metadatafields.ListByIDsTx(tx, ids)
	if err != nil {
		return nil, err
	}
	out := make(map[string]metadatafields.Field, len(found))
	for _, f := range found {
		out[string(f.ID)] = f
	}
	for _, id := range ids {
		if _, ok := out[string(id)]; !ok {
			return nil, ErrInvalid
		}
	}
	return out, nil
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func mapConstraint(err error) error {
	if database.IsConstraintViolation(err) {
		return ErrInvalid
	}
	return err
}
