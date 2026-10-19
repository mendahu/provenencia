package database

import (
	"database/sql"
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

func TestParseMigrationsForeignKeys(t *testing.T) {
	for _, body := range []string{"PRAGMA foreign_keys=OFF;", "pragma foreign_key_check;"} {
		fsys := fstest.MapFS{"migrations/000001.sql": {Data: []byte(body)}}
		_, _, err := parseMigrations(fsys, "migrations")
		if err == nil || !errors.Is(err, apperr.New(apperr.CodeInternalMigrations, apperr.KindInternal)) {
			t.Fatalf("%q: err %v want internal.migrations", body, err)
		}
	}
	refused := []string{
		"-- migrate: foreign_keys off \nSELECT 1;",
		"-- migrate:  foreign_keys off\nSELECT 1;",
		"-- Migrate: foreign_keys off\nSELECT 1;",
		"-- migrate: foreign_keys on\nSELECT 1;",
		"SELECT 1;\nCOMMIT;",
		"BEGIN TRANSACTION;\nSELECT 1;",
		"SELECT 1;\nROLLBACK;",
		"SELECT 1;\n  begin;\nSELECT 2;",
	}
	for _, body := range refused {
		fsys := fstest.MapFS{"migrations/000001.sql": {Data: []byte(body)}}
		_, _, err := parseMigrations(fsys, "migrations")
		if err == nil || !errors.Is(err, apperr.New(apperr.CodeInternalMigrations, apperr.KindInternal)) {
			t.Fatalf("%q: err %v want internal.migrations", body, err)
		}
	}
	trigger := "CREATE TABLE t (id INTEGER);\nCREATE TRIGGER t_ai AFTER INSERT ON t\nBEGIN\n\tUPDATE t SET id = id;\nEND;"
	fsys := fstest.MapFS{
		"migrations/000001.sql": {Data: []byte("SELECT 1;")},
		"migrations/000002.sql": {Data: []byte("-- migrate: foreign_keys off\nSELECT 2;")},
		"migrations/000003.sql": {Data: []byte(trigger)},
	}
	steps, _, err := parseMigrations(fsys, "migrations")
	if err != nil {
		t.Fatal(err)
	}
	if steps[0].foreignKeysOff || !steps[1].foreignKeysOff {
		t.Fatalf("foreignKeysOff = %v, %v want false, true", steps[0].foreignKeysOff, steps[1].foreignKeysOff)
	}
}

// useMigrations swaps the package's step list for a fake one and returns a
// fresh in-memory database.
func useMigrations(t *testing.T, steps []migration) *sql.DB {
	t.Helper()
	oldSteps, oldVersion := migrations, formatVersion
	migrations, formatVersion = steps, steps[len(steps)-1].to
	t.Cleanup(func() { migrations, formatVersion = oldSteps, oldVersion })
	db, err := openDBURI(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func pragmaInt(t *testing.T, db *sql.DB, pragma string) int {
	t.Helper()
	var n int
	if err := db.QueryRow("PRAGMA " + pragma).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func tableExists(t *testing.T, db *sql.DB, name string) bool {
	t.Helper()
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_schema WHERE type = 'table' AND name = ?`, name).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n == 1
}

func TestMigrateRollsBackAFailedStep(t *testing.T) {
	db := useMigrations(t, []migration{
		{to: 1, sql: `CREATE TABLE a (id INTEGER PRIMARY KEY);`},
		{to: 2, sql: `CREATE TABLE b (id INTEGER PRIMARY KEY); THIS IS NOT SQL;`},
	})
	if err := migrate(db); err == nil {
		t.Fatal("migrate succeeded; want an error")
	}
	if v := pragmaInt(t, db, "user_version"); v != 1 {
		t.Fatalf("user_version = %d want 1", v)
	}
	if tableExists(t, db, "b") {
		t.Fatal("table b exists after a failed step")
	}
}

const parentChildSQL = `
CREATE TABLE parent (id INTEGER PRIMARY KEY) STRICT;
CREATE TABLE child (id INTEGER PRIMARY KEY, parent_id INTEGER NOT NULL REFERENCES parent(id) ON DELETE CASCADE) STRICT;
`

func TestMigrateRefusesBrokenForeignKeys(t *testing.T) {
	db := useMigrations(t, []migration{
		{to: 1, sql: parentChildSQL},
		{to: 2, foreignKeysOff: true, sql: `INSERT INTO child (id, parent_id) VALUES (1, 99);`},
	})
	err := migrate(db)
	if err == nil || !errors.Is(err, apperr.New(apperr.CodeInternalMigrations, apperr.KindInternal)) {
		t.Fatalf("err %v want internal.migrations", err)
	}
	if v := pragmaInt(t, db, "user_version"); v != 1 {
		t.Fatalf("user_version = %d want 1", v)
	}
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM child`).Scan(&n); err != nil || n != 0 {
		t.Fatalf("child rows = %d (err %v) want 0", n, err)
	}
	if fk := pragmaInt(t, db, "foreign_keys"); fk != 1 {
		t.Fatalf("foreign_keys = %d want 1", fk)
	}
}

func TestMigrateForeignKeysOffStepRebuilds(t *testing.T) {
	db := useMigrations(t, []migration{
		{to: 1, sql: parentChildSQL + `
INSERT INTO parent (id) VALUES (1), (2);
INSERT INTO child (id, parent_id) VALUES (10, 1), (20, 2);`},
		{to: 2, foreignKeysOff: true, sql: `
CREATE TABLE parent__new (id INTEGER PRIMARY KEY, note TEXT) STRICT;
INSERT INTO parent__new (id) SELECT id FROM parent;
DROP TABLE parent;
ALTER TABLE parent__new RENAME TO parent;`},
	})
	if err := migrate(db); err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM child`).Scan(&n); err != nil || n != 2 {
		t.Fatalf("child rows = %d (err %v) want 2", n, err)
	}
	if fk := pragmaInt(t, db, "foreign_keys"); fk != 1 {
		t.Fatalf("foreign_keys = %d want 1", fk)
	}
	if v := pragmaInt(t, db, "user_version"); v != 2 {
		t.Fatalf("user_version = %d want 2", v)
	}
}
