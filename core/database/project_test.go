package database

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateOpen(t *testing.T) {
	tests := []struct {
		name    string
		run     func(t *testing.T, parent string)
		wantErr error
	}{
		{
			name: "create layout application_id and empty users",
			run: func(t *testing.T, parent string) {
				p, err := Create(parent, "Robins Family.provenance")
				if err != nil {
					t.Fatal(err)
				}
				defer p.Close()
				dir := p.Dir()
				for _, rel := range []string{catalogFile, objectsDir, derivativesDir} {
					if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
						t.Fatalf("%s: %v", rel, err)
					}
				}
				var appID, ver, n int
				if err := p.db.QueryRow(`PRAGMA application_id`).Scan(&appID); err != nil {
					t.Fatal(err)
				}
				if appID != ApplicationID {
					t.Fatalf("application_id %d want %d", appID, ApplicationID)
				}
				if err := p.db.QueryRow(`PRAGMA user_version`).Scan(&ver); err != nil {
					t.Fatal(err)
				}
				if ver != formatVersion {
					t.Fatalf("user_version %d want %d", ver, formatVersion)
				}
				if err := p.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&n); err != nil {
					t.Fatal(err)
				}
				if n != 0 {
					t.Fatalf("users count %d", n)
				}
				if _, err := p.db.Exec(`INSERT INTO users (id, display_name) VALUES (?, ?)`, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, "Jake"); err != nil {
					t.Fatal(err)
				}
				var name string
				if err := p.db.QueryRow(`SELECT display_name FROM users`).Scan(&name); err != nil {
					t.Fatal(err)
				}
				if name != "Jake" {
					t.Fatalf("got %q", name)
				}
			},
		},
		{
			name: "create fails if exists",
			run: func(t *testing.T, parent string) {
				p, err := Create(parent, "Dup.provenance")
				if err != nil {
					t.Fatal(err)
				}
				p.Close()
				_, err = Create(parent, "Dup.provenance")
				if err != ErrAlreadyExists {
					t.Fatalf("got %v want %v", err, ErrAlreadyExists)
				}
			},
		},
		{
			name: "create requires provenance suffix",
			run: func(t *testing.T, parent string) {
				_, err := Create(parent, "Robins Family")
				if err != ErrInvalidFolderName {
					t.Fatalf("got %v want %v", err, ErrInvalidFolderName)
				}
			},
		},
		{
			name: "open refuses missing sqlite",
			run: func(t *testing.T, parent string) {
				dir := filepath.Join(parent, "empty")
				if err := os.Mkdir(dir, 0o755); err != nil {
					t.Fatal(err)
				}
				_, err := Open(dir)
				if err != ErrNotAProject {
					t.Fatalf("got %v want %v", err, ErrNotAProject)
				}
			},
		},
		{
			name: "open refuses wrong application_id",
			run: func(t *testing.T, parent string) {
				dir := filepath.Join(parent, "foreign.provenance")
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatal(err)
				}
				db, err := openDB(catalogPath(dir))
				if err != nil {
					t.Fatal(err)
				}
				if _, err := db.Exec(`PRAGMA application_id = 1`); err != nil {
					db.Close()
					t.Fatal(err)
				}
				db.Close()
				_, err = Open(dir)
				if err != ErrNotAProject {
					t.Fatalf("got %v want %v", err, ErrNotAProject)
				}
			},
		},
		{
			name: "open refuses unsupported user_version",
			run: func(t *testing.T, parent string) {
				p, err := Create(parent, "Future.provenance")
				if err != nil {
					t.Fatal(err)
				}
				dir := p.Dir()
				if _, err := p.db.Exec(`PRAGMA user_version = 99`); err != nil {
					p.Close()
					t.Fatal(err)
				}
				p.Close()
				_, err = Open(dir)
				if err == nil || !errors.Is(err, ErrUnsupportedVersion) {
					t.Fatalf("got %v want %v", err, ErrUnsupportedVersion)
				}
			},
		},
		{
			name: "second open fails then reopen after close",
			run: func(t *testing.T, parent string) {
				p, err := Create(parent, "Lock.provenance")
				if err != nil {
					t.Fatal(err)
				}
				_, err = Open(p.Dir())
				if err != ErrAlreadyOpen {
					p.Close()
					t.Fatalf("got %v want %v", err, ErrAlreadyOpen)
				}
				if err := p.Close(); err != nil {
					t.Fatal(err)
				}
				p2, err := Open(p.Dir())
				if err != nil {
					t.Fatal(err)
				}
				p2.Close()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parent := t.TempDir()
			tt.run(t, parent)
		})
	}
}

func TestUpsertUser(t *testing.T) {
	id := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	tests := []struct {
		name string
		run  func(t *testing.T, p *Project)
	}{
		{
			name: "insert then update name",
			run: func(t *testing.T, p *Project) {
				if err := p.UpsertUser(id, "Jake"); err != nil {
					t.Fatal(err)
				}
				got, err := p.LookupUser(id)
				if err != nil {
					t.Fatal(err)
				}
				if got != "Jake" {
					t.Fatalf("got %q", got)
				}
				if err := p.UpsertUser(id, "Jake R."); err != nil {
					t.Fatal(err)
				}
				got, err = p.LookupUser(id)
				if err != nil {
					t.Fatal(err)
				}
				if got != "Jake R." {
					t.Fatalf("got %q", got)
				}
			},
		},
		{
			name: "rejects short id",
			run: func(t *testing.T, p *Project) {
				if err := p.UpsertUser([]byte{1}, "Jake"); err != ErrInvalidUser {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "rejects blank name",
			run: func(t *testing.T, p *Project) {
				if err := p.UpsertUser(id, "  "); err != ErrInvalidUser {
					t.Fatalf("got %v", err)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := Create(t.TempDir(), "T.provenance")
			if err != nil {
				t.Fatal(err)
			}
			defer p.Close()
			tt.run(t, p)
		})
	}
}
