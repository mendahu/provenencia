package deleteimpact

import (
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database"
)

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

	registered := map[string]fkSpec{}
	for _, fk := range foreignKeys {
		registered[fk.FromTable+"."+fk.FromCol] = fk
	}
	resourceVias := resourceInboundVias()
	owned := ownedColumns()

	tables, err := db.Query(`SELECT name FROM sqlite_schema
		WHERE type = 'table' AND name NOT LIKE 'sqlite_%' AND name NOT LIKE 'catalog_search_fts%'
		ORDER BY name`)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for tables.Next() {
		var name string
		if err := tables.Scan(&name); err != nil {
			tables.Close()
			t.Fatal(err)
		}
		names = append(names, name)
	}
	if err := tables.Err(); err != nil {
		tables.Close()
		t.Fatal(err)
	}
	tables.Close()
	seen := map[string]bool{}
	for _, name := range names {
		fks, err := db.Query(`PRAGMA foreign_key_list(` + name + `)`)
		if err != nil {
			t.Fatal(err)
		}
		for fks.Next() {
			var id, seq int
			var toTable, fromCol, toCol, onUpdate, onDelete, match string
			if err := fks.Scan(&id, &seq, &toTable, &fromCol, &toCol, &onUpdate, &onDelete, &match); err != nil {
				t.Fatal(err)
			}
			key := name + "." + fromCol
			seen[key] = true
			spec, ok := registered[key]
			if !ok {
				t.Errorf("unregistered FK %s -> %s", key, toTable)
				continue
			}
			gotAction := normalizeOnDelete(onDelete)
			wantAction := normalizeOnDelete(spec.OnDelete)
			if gotAction != wantAction {
				t.Errorf("%s on_delete=%s registered=%s", key, gotAction, wantAction)
			}
			if spec.ToTable != toTable {
				t.Errorf("%s to=%s registered=%s", key, toTable, spec.ToTable)
			}
			switch spec.Bucket {
			case BucketResource:
				if !resourceVias[key] {
					t.Errorf("resource FK %s has no list probe", key)
				}
			case BucketOwnedOutbound:
				if !owned[key] {
					t.Errorf("owned-outbound FK %s missing from release list", key)
				}
			}
		}
		fks.Close()
	}
	for key := range registered {
		if !seen[key] {
			t.Errorf("registered FK %s is not live", key)
		}
	}
}

func normalizeOnDelete(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	if s == "" || s == "RESTRICT" {
		return "NO ACTION"
	}
	return s
}
