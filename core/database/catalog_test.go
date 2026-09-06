package database

import (
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
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
				c, err := Create(parent, "Robins Family.provenencia")
				if err != nil {
					t.Fatal(err)
				}
				defer c.Close()
				dir := c.Dir()
				for _, rel := range []string{catalogFile, objectsDir, derivativesDir} {
					if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
						t.Fatalf("%s: %v", rel, err)
					}
				}
				var appID, ver, n int
				if err := c.db.QueryRow(`PRAGMA application_id`).Scan(&appID); err != nil {
					t.Fatal(err)
				}
				if appID != ApplicationID {
					t.Fatalf("application_id %d want %d", appID, ApplicationID)
				}
				if err := c.db.QueryRow(`PRAGMA user_version`).Scan(&ver); err != nil {
					t.Fatal(err)
				}
				if ver != formatVersion {
					t.Fatalf("user_version %d want %d", ver, formatVersion)
				}
				if err := c.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&n); err != nil {
					t.Fatal(err)
				}
				if n != 0 {
					t.Fatalf("users count %d", n)
				}
			},
		},
		{
			name: "create fails if exists",
			run: func(t *testing.T, parent string) {
				c, err := Create(parent, "Dup.provenencia")
				if err != nil {
					t.Fatal(err)
				}
				c.Close()
				_, err = Create(parent, "Dup.provenencia")
				if err != ErrAlreadyExists {
					t.Fatalf("got %v want %v", err, ErrAlreadyExists)
				}
			},
		},
		{
			name: "create requires provenencia suffix",
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
				dir := filepath.Join(parent, "foreign.provenencia")
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
				c, err := Create(parent, "Future.provenencia")
				if err != nil {
					t.Fatal(err)
				}
				dir := c.Dir()
				if _, err := c.db.Exec(`PRAGMA user_version = 99`); err != nil {
					c.Close()
					t.Fatal(err)
				}
				c.Close()
				_, err = Open(dir)
				if err == nil || !errors.Is(err, ErrUnsupportedVersion) {
					t.Fatalf("got %v want %v", err, ErrUnsupportedVersion)
				}
			},
		},
		{
			name: "create and reopen folder with query special chars",
			run: func(t *testing.T, parent string) {
				c, err := Create(parent, "Weird?Hash# Name.provenencia")
				if err != nil {
					t.Fatal(err)
				}
				dir := c.Dir()
				if err := c.Close(); err != nil {
					t.Fatal(err)
				}
				c2, err := Open(dir)
				if err != nil {
					t.Fatal(err)
				}
				defer c2.Close()
				var n int
				if err := c2.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&n); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "second open fails then reopen after close",
			run: func(t *testing.T, parent string) {
				c, err := Create(parent, "Lock.provenencia")
				if err != nil {
					t.Fatal(err)
				}
				_, err = Open(c.Dir())
				if err != ErrAlreadyOpen {
					c.Close()
					t.Fatalf("got %v want %v", err, ErrAlreadyOpen)
				}
				if err := c.Close(); err != nil {
					t.Fatal(err)
				}
				c2, err := Open(c.Dir())
				if err != nil {
					t.Fatal(err)
				}
				c2.Close()
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

func TestDSN(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{name: "question mark", path: "/tmp/foo?bar/provenencia.sqlite"},
		{name: "hash", path: "/tmp/foo#bar/provenencia.sqlite"},
		{name: "space", path: "/tmp/foo bar/provenencia.sqlite"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dsn(tt.path)
			u, err := url.Parse(got)
			if err != nil {
				t.Fatal(err)
			}
			if u.Scheme != "file" {
				t.Fatalf("scheme %q", u.Scheme)
			}
			if u.Path != tt.path {
				t.Fatalf("path %q want %q", u.Path, tt.path)
			}
			q := u.Query()
			if q.Get("_busy_timeout") != "100" || q.Get("_foreign_keys") != "1" || q.Get("_journal_mode") != "WAL" {
				t.Fatalf("query %v", q)
			}
			cut := strings.Index(got, "?")
			if cut < 0 {
				t.Fatal("missing query")
			}
			if strings.Contains(got[:cut], "?") {
				t.Fatalf("unencoded ? in path %s", got)
			}
		})
	}
}
