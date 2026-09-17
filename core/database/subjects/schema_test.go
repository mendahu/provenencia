package subjects

import (
	"testing"

	"github.com/mendahu/provenencia/core/database"
)

func TestMigrationCreatesSubjectTables(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	var version int
	if err := db.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil {
		t.Fatal(err)
	}
	if version < 21 {
		t.Fatalf("user_version %d, want >= 21", version)
	}

	wantCols := map[string][]string{
		"subject_types": {
			"id", "key", "origin", "label", "description",
			"ref_prefix", "candidate_ref_prefix",
		},
		"subjects": {
			"id", "ref", "source_id", "subject_type_id", "label", "description",
		},
		"subject_positions": {
			"source_id", "subject_id", "grid_x", "grid_y",
		},
	}
	for table, cols := range wantCols {
		rows, err := db.Query(`PRAGMA table_info(` + table + `)`)
		if err != nil {
			t.Fatalf("%s: %v", table, err)
		}
		got := map[string]struct{}{}
		for rows.Next() {
			var cid int
			var name, colType string
			var notNull int
			var dflt any
			var pk int
			if err := rows.Scan(&cid, &name, &colType, &notNull, &dflt, &pk); err != nil {
				rows.Close()
				t.Fatal(err)
			}
			got[name] = struct{}{}
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			t.Fatal(err)
		}
		rows.Close()
		if len(got) == 0 {
			t.Fatalf("missing table %s", table)
		}
		for _, want := range cols {
			if _, ok := got[want]; !ok {
				t.Fatalf("%s: column %q missing, got %v", table, want, got)
			}
		}
	}
}
