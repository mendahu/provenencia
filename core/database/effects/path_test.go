package effects

import (
	"bytes"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mendahu/provenencia/core/database/rowchange"
)

// countingQuerier counts the single-row lookups a walk issues.
type countingQuerier struct {
	querier
	rows int
}

func (c *countingQuerier) QueryRow(query string, args ...any) *sql.Row {
	c.rows++
	return c.querier.QueryRow(query, args...)
}

// TestReadRowsColumnLooksUpEachIDOnce pins that a path step looks each
// distinct id up once: three Observations of one Subject read that Subject's
// graph once, and the output keeps first-sighting order.
func TestReadRowsColumnLooksUpEachIDOnce(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	db.SetMaxOpenConns(1)
	id := func(b byte) []byte { return bytes.Repeat([]byte{b}, 16) }
	if _, err := db.Exec(`CREATE TABLE subjects (id BLOB PRIMARY KEY, graph_id BLOB)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO subjects (id, graph_id) VALUES (?, ?), (?, ?)`,
		id(1), id(9), id(2), id(8)); err != nil {
		t.Fatal(err)
	}
	cq := &countingQuerier{querier: db}
	w := &walk{q: cq, table: "subjects"}
	got, err := readRowsColumn(w, "subjects", [][]byte{id(1), id(1), id(2), id(1)}, "graph_id")
	if err != nil {
		t.Fatal(err)
	}
	if cq.rows != 2 {
		t.Errorf("lookups = %d, want 2", cq.rows)
	}
	if len(got) != 2 || !bytes.Equal(got[0], id(9)) || !bytes.Equal(got[1], id(8)) {
		t.Errorf("got %x, want 9s then 8s", got)
	}
}

// TestScopeThroughDeletedParent resolves a released row's Source scope when the
// parent it hangs off is gone in the same revision: the second hop reads the
// parent's column from its delete change.
func TestScopeThroughDeletedParent(t *testing.T) {
	id := func(b byte) []byte { return bytes.Repeat([]byte{b}, 16) }
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	db.SetMaxOpenConns(1)
	// The parents are deleted; only the table of the final hop holds a row.
	for _, ddl := range []string{
		`CREATE TABLE artifacts (id BLOB PRIMARY KEY, source_id BLOB)`,
		`CREATE TABLE evidence_graphs (id BLOB PRIMARY KEY, source_id BLOB)`,
		`CREATE TABLE citations (id BLOB PRIMARY KEY, artifact_id BLOB)`,
		`CREATE TABLE artifact_metadata (id BLOB PRIMARY KEY, artifact_id BLOB)`,
	} {
		if _, err := db.Exec(ddl); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`INSERT INTO artifacts (id, source_id) VALUES (?, ?)`, id(2), id(7)); err != nil {
		t.Fatal(err)
	}
	del := func(entity string, entityID []byte, old map[string]any) rowchange.Change {
		return rowchange.Change{EntityType: entity, EntityID: entityID, Action: rowchange.ActionDelete, Fields: rowchange.DeletedRow(old)}
	}
	uid := func(b []byte) string { return uuid.UUID(b).String() }
	tests := []struct {
		name    string
		table   string
		change  rowchange.Change
		parents []rowchange.Change
	}{
		{
			name:   "artifact metadata value",
			table:  "artifact_metadata",
			change: del("artifact_metadata", id(10), map[string]any{"id": uid(id(10)), "artifact_id": uid(id(1)), "field_id": uid(id(5))}),
			parents: []rowchange.Change{
				del("artifact", id(1), map[string]any{"id": uid(id(1)), "source_id": uid(id(7))}),
			},
		},
		{
			name:   "artifact metadata layout",
			table:  "artifact_metadata_layout",
			change: del("artifact_metadata_layout", id(5), map[string]any{"artifact_id": uid(id(1)), "field_id": uid(id(5))}),
			parents: []rowchange.Change{
				del("artifact", id(1), map[string]any{"id": uid(id(1)), "source_id": uid(id(7))}),
			},
		},
		{
			name:   "graph metadata value",
			table:  "evidence_graph_metadata",
			change: del("evidence_graph_metadata", id(10), map[string]any{"id": uid(id(10)), "graph_id": uid(id(3)), "field_id": uid(id(5))}),
			parents: []rowchange.Change{
				del("evidence_graph", id(3), map[string]any{"id": uid(id(3)), "source_id": uid(id(7))}),
			},
		},
		{
			name:   "graph metadata layout",
			table:  "evidence_graph_metadata_layout",
			change: del("evidence_graph_metadata_layout", id(5), map[string]any{"graph_id": uid(id(3)), "field_id": uid(id(5))}),
			parents: []rowchange.Change{
				del("evidence_graph", id(3), map[string]any{"id": uid(id(3)), "source_id": uid(id(7))}),
			},
		},
		{
			name:   "citation note under a Citation delete",
			table:  "citation_notes",
			change: del("citation_note", id(11), map[string]any{"id": uid(id(11)), "citation_id": uid(id(4)), "body": "x"}),
			parents: []rowchange.Change{
				del("citation", id(4), map[string]any{"id": uid(id(4)), "artifact_id": uid(id(2))}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SourceScopeOfChange(db, append([]rowchange.Change{tt.change}, tt.parents...), 0)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != 1 || !bytes.Equal(got[0], id(7)) {
				t.Fatalf("scope = %x, want the Source %x", got, id(7))
			}
			// Without the parent's delete change nothing resolves.
			if got, err := SourceScopeOfChange(db, []rowchange.Change{tt.change}, 0); err != nil || len(got) != 0 {
				t.Fatalf("scope without the parent change = %x, %v; want none", got, err)
			}
		})
	}
}
