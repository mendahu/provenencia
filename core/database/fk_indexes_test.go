package database

import "testing"

func TestMigrationCreatesFKIndexes(t *testing.T) {
	c, err := Create(t.TempDir(), "t.provenencia")
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
	if version < 22 {
		t.Fatalf("user_version %d, want >= 22", version)
	}

	want := []struct {
		name  string
		table string
	}{
		{name: "artifacts_source_id_idx", table: "artifacts"},
		{name: "artifacts_file_id_idx", table: "artifacts"},
		{name: "source_notes_source_id_idx", table: "source_notes"},
		{name: "subjects_source_id_idx", table: "subjects"},
		{name: "sources_source_type_id_idx", table: "sources"},
	}
	for _, tt := range want {
		t.Run(tt.name, func(t *testing.T) {
			var gotName, gotTable string
			err := db.QueryRow(
				`SELECT name, tbl_name FROM sqlite_master WHERE type = 'index' AND name = ?`,
				tt.name,
			).Scan(&gotName, &gotTable)
			if err != nil {
				t.Fatalf("index %s: %v", tt.name, err)
			}
			if gotTable != tt.table {
				t.Fatalf("index %s on %q want %q", tt.name, gotTable, tt.table)
			}
		})
	}
}
