package sourcemetadata

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/mendahu/provenencia/core/vocabkey"
)

// Owner is the record metadata hangs off: a Source, an Artifact or an Evidence graph. Level names which
// store holds its values, and ID is the owner row's id.
type Owner struct {
	Level vocabkey.Level
	ID    []byte
}

// key names the owner in a map: the level strings differ and ids are 16 bytes,
// so two owners never share a key.
func (o Owner) key() string { return string(o.Level) + string(o.ID) }

// store is one level's tables and the SQL over them, built once by newStore
// from constant names. Every function that reads or writes metadata starts at
// storeFor, so a level with no entry is refused before any query.
type store struct {
	level vocabkey.Level

	// Names that also appear in change rows.
	ownerColumn string
	values      string
	layout      string

	// typeIn and typeSuffix bracket the IN list of sqlbatch.Query. They read
	// the owners' ids and their Source type ids.
	typeIn     string
	typeSuffix string

	ownerExists string
	insert      string
	update      string
	delete      string
	getPair     string

	// listIn and layoutListIn are heads for sqlbatch.Query, closed by the
	// suffix their order needs (listSuffix and layoutListSuffix).
	listIn       string
	listSuffix   string
	layoutListIn string
	// layoutListSuffix orders layout rows within an owner.
	layoutListSuffix string

	layoutAppend  string
	layoutDismiss string
	layoutOrder   string
	layoutInsert  string
	layoutUpdate  string
	layoutGet     string
}

// newStore builds a level's SQL from its constant names, as deleteimpact's
// rowFacet does. No argument comes from a caller; typeIn is passed whole, up to its IN list.
func newStore(level vocabkey.Level, ownerTable, ownerColumn, values, layout, typeIn string) *store {
	f := func(format string, a ...any) string { return fmt.Sprintf(format, a...) }
	return &store{
		level:       level,
		ownerColumn: ownerColumn,
		values:      values,
		layout:      layout,
		typeIn:      typeIn,
		typeSuffix:  `)`,
		ownerExists: f(`SELECT 1 FROM %s WHERE id = ?`, ownerTable),
		insert: f(`INSERT INTO %s (id, %s, field_id, value_text)
			VALUES (?, ?, ?, ?)`, values, ownerColumn),
		update:  f(`UPDATE %s SET value_text = ? WHERE id = ?`, values),
		delete:  f(`DELETE FROM %s WHERE %s = ? AND field_id = ?`, values, ownerColumn),
		getPair: f(`SELECT id, %[2]s, field_id, COALESCE(value_text, '') FROM %[1]s WHERE %[2]s = ? AND field_id = ?`, values, ownerColumn),
		listIn: f(`SELECT id, %[2]s, field_id, COALESCE(value_text, '')
			FROM %[1]s WHERE %[2]s IN (`, values, ownerColumn),
		listSuffix: f(`) ORDER BY %s, field_id`, ownerColumn),
		layoutListIn: f(`SELECT %[2]s, field_id, sort_order, dismissed
			FROM %[1]s WHERE %[2]s IN (`, layout, ownerColumn),
		layoutListSuffix: f(`) ORDER BY %s, sort_order, field_id`, ownerColumn),

		// Layout rows append at the end of the owner's existing order. The next
		// sort_order is computed inside the INSERT so two appends cannot read the
		// same MAX and share a slot; the WHERE clause is also what lets SQLite
		// parse INSERT…SELECT with an upsert clause.
		layoutAppend: f(`INSERT INTO %[1]s (%[2]s, field_id, sort_order, dismissed)
			SELECT ?, ?, COALESCE(MAX(sort_order), -1) + 1, 0
			FROM %[1]s WHERE %[2]s = ?
			ON CONFLICT(%[2]s, field_id) DO NOTHING`, layout, ownerColumn),
		layoutDismiss: f(`INSERT INTO %[1]s (%[2]s, field_id, sort_order, dismissed)
			SELECT ?, ?, COALESCE(MAX(sort_order), -1) + 1, 1
			FROM %[1]s WHERE %[2]s = ?
			ON CONFLICT(%[2]s, field_id) DO UPDATE SET dismissed = 1`, layout, ownerColumn),
		// Reorder must not disturb dismissed, so the upsert touches sort_order only.
		layoutOrder: f(`INSERT INTO %[1]s (%[2]s, field_id, sort_order)
			VALUES (?, ?, ?)
			ON CONFLICT(%[2]s, field_id) DO UPDATE SET sort_order = excluded.sort_order`, layout, ownerColumn),
		layoutInsert: f(`INSERT INTO %[1]s (%[2]s, field_id, sort_order, dismissed)
			VALUES (?, ?, ?, ?)`, layout, ownerColumn),
		layoutUpdate: f(`UPDATE %[1]s SET sort_order = ?, dismissed = ? WHERE %[2]s = ? AND field_id = ?`, layout, ownerColumn),
		layoutGet:    f(`SELECT field_id, sort_order, dismissed FROM %[1]s WHERE %[2]s = ? AND field_id = ?`, layout, ownerColumn),
	}
}

// stores holds one entry per level that keeps metadata. A level with no entry
// is ErrInvalid everywhere.
var stores = map[vocabkey.Level]*store{
	vocabkey.LevelSource: newStore(vocabkey.LevelSource, "sources", "source_id", "source_metadata", "source_metadata_layout",
		"SELECT id, source_type_id FROM sources WHERE id IN ("),
	vocabkey.LevelArtifact: newStore(vocabkey.LevelArtifact, "artifacts", "artifact_id", "artifact_metadata", "artifact_metadata_layout",
		"SELECT a.id, s.source_type_id FROM artifacts a JOIN sources s ON s.id = a.source_id WHERE a.id IN ("),
	vocabkey.LevelGraph: newStore(vocabkey.LevelGraph, "evidence_graphs", "graph_id", "evidence_graph_metadata", "evidence_graph_metadata_layout",
		"SELECT g.id, s.source_type_id FROM evidence_graphs g JOIN sources s ON s.id = g.source_id WHERE g.id IN ("),
}

// storeFor is the first check of every exported function: the owner's level
// has a store and its id is 16 bytes.
func storeFor(o Owner) (*store, error) {
	st, ok := stores[o.Level]
	if !ok || len(o.ID) != 16 {
		return nil, ErrInvalid
	}
	return st, nil
}

// requireOwner is ErrInvalid when the owner row does not exist.
func requireOwner(tx *sql.Tx, st *store, o Owner) error {
	var one int
	err := tx.QueryRow(st.ownerExists, o.ID).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrInvalid
	}
	return err
}
