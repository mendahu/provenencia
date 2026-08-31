package database

import (
	"strings"
	"testing"
	"testing/fstest"
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
				if err == nil || !strings.Contains(err.Error(), tt.errSub) {
					t.Fatalf("err %v want substring %q", err, tt.errSub)
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

func TestShippedMigrations(t *testing.T) {
	tests := []struct {
		name  string
		check func(t *testing.T)
	}{
		{
			name: "embed parses and matches init",
			check: func(t *testing.T) {
				steps, ver, err := parseMigrations(migrationFS, "migrations")
				if err != nil {
					t.Fatal(err)
				}
				if ver != formatVersion || len(steps) != ver {
					t.Fatalf("ver %d formatVersion %d steps %d", ver, formatVersion, len(steps))
				}
				if !strings.Contains(steps[0].sql, "CREATE TABLE users") {
					t.Fatalf("000001.sql missing users: %s", steps[0].sql)
				}
				if ver < 2 || !strings.Contains(steps[1].sql, "CREATE TABLE project") {
					t.Fatalf("000002.sql missing project: ver=%d sql=%s", ver, steps[1].sql)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t)
		})
	}
}
