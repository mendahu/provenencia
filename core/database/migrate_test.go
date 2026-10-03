package database

import (
	"errors"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/mendahu/provenencia/core/apperr"
)

func TestParseMigrations(t *testing.T) {
	tests := []struct {
		name    string
		files   fstest.MapFS
		wantN   int
		wantVer int
		errSub  string
	}{
		{
			name: "single 000001",
			files: fstest.MapFS{
				"migrations/000001.sql": {Data: []byte("SELECT 1;")},
			},
			wantN:   1,
			wantVer: 1,
		},
		{
			name: "consecutive 000001 and 000002",
			files: fstest.MapFS{
				"migrations/000001.sql": {Data: []byte("SELECT 1;")},
				"migrations/000002.sql": {Data: []byte("SELECT 2;")},
			},
			wantN:   2,
			wantVer: 2,
		},
		{
			name: "gap skips 000002",
			files: fstest.MapFS{
				"migrations/000001.sql": {Data: []byte("SELECT 1;")},
				"migrations/000003.sql": {Data: []byte("SELECT 3;")},
			},
			errSub: "missing step 2",
		},
		{
			name: "short name",
			files: fstest.MapFS{
				"migrations/1.sql": {Data: []byte("SELECT 1;")},
			},
			errSub: "NNNNNN.sql",
		},
		{
			name: "empty sql",
			files: fstest.MapFS{
				"migrations/000001.sql": {Data: []byte(" \n\t")},
			},
			errSub: "empty",
		},
		{
			name: "no sql files",
			files: fstest.MapFS{
				"migrations/readme.txt": {Data: []byte("x")},
			},
			errSub: "none",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			steps, ver, err := parseMigrations(tt.files, "migrations")
			if tt.errSub != "" {
				if err == nil || !errors.Is(err, apperr.New(apperr.CodeInternalMigrations, apperr.KindInternal)) {
					t.Fatalf("err %v want internal.migrations", err)
				}
				var ae *apperr.Error
				if !errors.As(err, &ae) {
					t.Fatalf("err %T want *apperr.Error", err)
				}
				joined := strings.Join(ae.Params(), " ")
				if !strings.Contains(joined, tt.errSub) {
					t.Fatalf("params %v want substring %q", ae.Params(), tt.errSub)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(steps) != tt.wantN || ver != tt.wantVer {
				t.Fatalf("got %d steps ver %d want %d steps ver %d", len(steps), ver, tt.wantN, tt.wantVer)
			}
		})
	}
}

// TestMigration35RenamesPropertiesAndRewritesAudit upgrades a catalog that
// stopped at version 34 — with a binding row and an audited metadata-field
// delete in the old vocabulary — and checks S9-07b's rename (000035).
func TestMigration35RenamesPropertiesAndRewritesAudit(t *testing.T) {
	db, err := openDBURI(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, step := range migrations[:34] {
		if _, err := db.Exec(step.sql); err != nil {
			t.Fatalf("step %d: %v", step.to, err)
		}
	}
	if _, err := db.Exec(`PRAGMA user_version = 34`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`PRAGMA foreign_keys = OFF`); err != nil {
		t.Fatal(err)
	}
	seed := []string{
		`INSERT INTO subject_type_fields (subject_type_id, property_id, sort_order) VALUES (x'01', x'02', 0)`,
		`INSERT INTO audit_transactions (id, revision, action_type, created_at) VALUES (x'10', 1, 'delete_source_field', 'now')`,
		`INSERT INTO audit_transactions (id, revision, action_type, created_at) VALUES (x'11', 2, 'update_source_metadata', 'now')`,
		`INSERT INTO audit_changes (id, audit_transaction_id, entity_type, entity_id, action, changes_json)
			VALUES (x'20', x'10', 'source_field', x'30', 'delete', '{}')`,
		`INSERT INTO audit_changes (id, audit_transaction_id, entity_type, entity_id, action, changes_json)
			VALUES (x'21', x'10', 'source_metadata_layout', x'31', 'delete', '{}')`,
	}
	for _, q := range seed {
		if _, err := db.Exec(q); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}

	if err := migrate(db); err != nil {
		t.Fatal(err)
	}

	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM subject_type_properties`).Scan(&n); err != nil || n != 1 {
		t.Fatalf("binding rows %d %v", n, err)
	}
	var names []string
	rows, err := db.Query(`SELECT name FROM sqlite_schema WHERE name LIKE 'subject_type_%' ORDER BY name`)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		names = append(names, name)
	}
	_ = rows.Close()
	for _, name := range names {
		if strings.Contains(name, "subject_type_fields") {
			t.Fatalf("old name left in schema: %v", names)
		}
	}
	if got := strings.Join(names, ","); !strings.Contains(got, "subject_type_properties_type_id_idx") ||
		!strings.Contains(got, "subject_type_properties_property_id_idx") {
		t.Fatalf("indexes not renamed: %s", got)
	}

	var action string
	if err := db.QueryRow(`SELECT action_type FROM audit_transactions WHERE revision = 1`).Scan(&action); err != nil || action != "delete_metadata_field" {
		t.Fatalf("action %q %v", action, err)
	}
	if err := db.QueryRow(`SELECT action_type FROM audit_transactions WHERE revision = 2`).Scan(&action); err != nil || action != "update_source_metadata" {
		t.Fatalf("unrelated action rewritten: %q %v", action, err)
	}
	var types []string
	rows, err = db.Query(`SELECT entity_type FROM audit_changes ORDER BY id`)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var typ string
		if err := rows.Scan(&typ); err != nil {
			t.Fatal(err)
		}
		types = append(types, typ)
	}
	_ = rows.Close()
	if strings.Join(types, ",") != "metadata_field,source_metadata_layout" {
		t.Fatalf("entity types %v", types)
	}

	digest, err := schemaDigest(db)
	if err != nil {
		t.Fatal(err)
	}
	if digest != expectedSchemaHash {
		t.Fatal("upgraded schema differs from a fresh catalog's")
	}
}
