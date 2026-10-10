package catalogmodel_test

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/catalogmodel"
)

// TestPragmaHonesty keeps the hand-written catalog model in step with a
// freshly created catalog. Tables and FKs are declared in this package; a
// migration that adds, changes, or drops a table or foreign key fails here
// until the model is updated in the same change.
//
// The check is four-way. Every catalog table except sqlite_% and FTS shadow
// tables is in Tables. Every live foreign key, keyed by its leading column,
// matches the registered To, OnDelete, and FromCols, and no registered FK is
// missing from PRAGMA foreign_key_list. The leading column of each FK has a
// covering index (explicit, UNIQUE, or primary key leftmost).
//
// Probes, releases, and projectors are checked in deleteimpact. This package
// stays below that policy and does not import it.
func TestPragmaHonesty(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}

	registered := map[string]catalogmodel.FK{}
	for _, fk := range catalogmodel.FKs {
		registered[fk.From+"."+fk.Column] = fk
	}
	registeredTables := map[string]bool{}
	for _, spec := range catalogmodel.Tables {
		if spec.Name != "" {
			registeredTables[spec.Name] = true
		}
	}

	tableRows, err := db.Query(`SELECT name FROM sqlite_schema
		WHERE type = 'table' AND name NOT LIKE 'sqlite_%' AND name NOT LIKE 'catalog_search_fts%'
		ORDER BY name`)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for tableRows.Next() {
		var name string
		if err := tableRows.Scan(&name); err != nil {
			tableRows.Close()
			t.Fatal(err)
		}
		names = append(names, name)
	}
	if err := tableRows.Err(); err != nil {
		tableRows.Close()
		t.Fatal(err)
	}
	tableRows.Close()
	seen := map[string]bool{}
	var fkCols [][2]string
	for _, name := range names {
		if !registeredTables[name] {
			t.Errorf("unregistered table %s", name)
		}
		fks, err := db.Query(`PRAGMA foreign_key_list(` + name + `)`)
		if err != nil {
			t.Fatal(err)
		}
		// PRAGMA returns one row per column; a composite FK is keyed by its
		// leading (seq 0) column and registers the rest in FromCols.
		type liveFK struct {
			toTable, onDelete string
			cols              []string
		}
		var order []int
		live := map[int]*liveFK{}
		for fks.Next() {
			var id, seq int
			var toTable, fromCol, toCol, onUpdate, onDelete, match string
			if err := fks.Scan(&id, &seq, &toTable, &fromCol, &toCol, &onUpdate, &onDelete, &match); err != nil {
				t.Fatal(err)
			}
			fk, ok := live[id]
			if !ok {
				fk = &liveFK{toTable: toTable, onDelete: onDelete}
				live[id] = fk
				order = append(order, id)
			}
			for len(fk.cols) <= seq {
				fk.cols = append(fk.cols, "")
			}
			fk.cols[seq] = fromCol
		}
		fks.Close()
		for _, id := range order {
			fk := live[id]
			fromCol, toTable, onDelete := fk.cols[0], fk.toTable, fk.onDelete
			key := name + "." + fromCol
			seen[key] = true
			fkCols = append(fkCols, [2]string{name, fromCol})
			spec, ok := registered[key]
			if !ok {
				t.Errorf("unregistered FK %s -> %s", key, toTable)
				continue
			}
			wantCols := spec.FromCols
			if len(wantCols) == 0 {
				wantCols = []string{spec.Column}
			}
			if strings.Join(wantCols, ",") != strings.Join(fk.cols, ",") {
				t.Errorf("%s cols=%v registered=%v", key, fk.cols, wantCols)
			}
			gotAction := normalizeOnDelete(onDelete)
			wantAction := normalizeOnDelete(spec.OnDelete)
			if gotAction != wantAction {
				t.Errorf("%s on_delete=%s registered=%s", key, gotAction, wantAction)
			}
			if spec.To != toTable {
				t.Errorf("%s to=%s registered=%s", key, toTable, spec.To)
			}
		}
	}
	for _, pair := range fkCols {
		name, fromCol := pair[0], pair[1]
		if !leftmostIndexed(t, db, name, fromCol) {
			t.Errorf("FK %s.%s has no covering index", name, fromCol)
		}
	}
	for key := range registered {
		if !seen[key] {
			t.Errorf("registered FK %s is not live", key)
		}
	}
}

func leftmostIndexed(t *testing.T, db *sql.DB, table, col string) bool {
	t.Helper()
	indexes, err := db.Query(`PRAGMA index_list(` + table + `)`)
	if err != nil {
		t.Fatal(err)
	}
	names, err := pragmaColumn(indexes, 1)
	indexes.Close()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		info, err := db.Query(`PRAGMA index_info("` + strings.ReplaceAll(name, `"`, `""`) + `")`)
		if err != nil {
			t.Fatal(err)
		}
		cols, err := pragmaColumn(info, 2)
		info.Close()
		if err != nil {
			t.Fatal(err)
		}
		if len(cols) > 0 && cols[0] == col {
			return true
		}
	}
	return false
}

func pragmaColumn(rows *sql.Rows, idx int) ([]string, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	if idx < 0 || idx >= len(cols) {
		return nil, fmt.Errorf("pragma column %d out of range", idx)
	}
	var out []string
	for rows.Next() {
		raw := make([]any, len(cols))
		dest := make([]any, len(cols))
		for i := range raw {
			dest[i] = &raw[i]
		}
		if err := rows.Scan(dest...); err != nil {
			return nil, err
		}
		out = append(out, asString(raw[idx]))
	}
	return out, rows.Err()
}

func asString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	default:
		return ""
	}
}

func normalizeOnDelete(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	if s == "" || s == "RESTRICT" {
		return "NO ACTION"
	}
	return s
}
