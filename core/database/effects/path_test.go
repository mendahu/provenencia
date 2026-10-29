package effects

import (
	"bytes"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
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
