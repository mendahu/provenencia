package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// TEMPORARY (Spike 1 PR3): packaging spike, not the project catalog.
// Delete this file when create/open provenance.sqlite exists (PR4).
func Probe() error {
	f, err := os.CreateTemp("", "provenance-sqlite-probe-*.sqlite")
	if err != nil {
		return err
	}
	path := f.Name()
	if err := f.Close(); err != nil {
		_ = os.Remove(path)
		return err
	}
	defer os.Remove(path)

	dsn := fmt.Sprintf("file:%s?_busy_timeout=5000&_foreign_keys=1&_journal_mode=WAL", path)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	if _, err := db.Exec(`CREATE TABLE spike (
		id INTEGER PRIMARY KEY,
		note TEXT NOT NULL
	) STRICT`); err != nil {
		return fmt.Errorf("create: %w", err)
	}
	if _, err := db.Exec(`INSERT INTO spike (id, note) VALUES (1, ?)`, "ok"); err != nil {
		return fmt.Errorf("insert: %w", err)
	}
	var note string
	if err := db.QueryRow(`SELECT note FROM spike WHERE id = 1`).Scan(&note); err != nil {
		return fmt.Errorf("select: %w", err)
	}
	if note != "ok" {
		return fmt.Errorf("got note %q", note)
	}
	return nil
}
