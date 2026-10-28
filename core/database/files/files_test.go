package files

import (
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/mattn/go-sqlite3"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/objectpath"
)

func TestStorageRelPath(t *testing.T) {
	const sum = "8fce3b0000000000000000000000000000000000000000000000000000000000"
	tests := []struct {
		name      string
		mediaType string
		want      string
		wantErr   error
	}{
		{name: "jpeg", mediaType: "image/jpeg", want: "objects/8f/ce/" + sum + ".jpg"},
		{name: "jpeg with charset", mediaType: "image/jpeg; charset=binary", want: "objects/8f/ce/" + sum + ".jpg"},
		{name: "png", mediaType: "image/png", want: "objects/8f/ce/" + sum + ".png"},
		{name: "pdf", mediaType: "application/pdf", want: "objects/8f/ce/" + sum + ".pdf"},
		{name: "unknown", mediaType: "application/octet-stream", wantErr: objectpath.ErrDisallowedType},
		{name: "empty", mediaType: "", wantErr: objectpath.ErrDisallowedType},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StorageRelPath(sum, tt.mediaType)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err %v want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
	if _, err := StorageRelPath("abcd", ""); !errors.Is(err, ErrInvalid) {
		t.Fatalf("short %v", err)
	}
	if _, err := StorageRelPath("8FCE3B0000000000000000000000000000000000000000000000000000000000", "image/jpeg"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("upper %v", err)
	}
}

func TestLookupInsert(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	id, err := NewID()
	if err != nil {
		t.Fatal(err)
	}
	const sum = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	if err := Insert(tx, File{
		ID: id, ChecksumSHA256: sum, OriginalFilename: "a.jpg", MediaType: "image/jpeg", ByteSize: 12,
	}); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	got, err := LookupByChecksum(c, sum)
	if err != nil {
		t.Fatal(err)
	}
	if got.OriginalFilename != "a.jpg" || got.ByteSize != 12 {
		t.Fatalf("%+v", got)
	}
	byID, err := Lookup(c, id)
	if err != nil || string(byID.ID) != string(id) {
		t.Fatalf("%v %+v", err, byID)
	}
	_, err = Lookup(c, make([]byte, 16))
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("missing %v", err)
	}

	tx2, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	if err := UpdateOriginalFilename(tx2, id, "renamed.jpg"); err != nil {
		t.Fatal(err)
	}
	if err := tx2.Commit(); err != nil {
		t.Fatal(err)
	}
	got2, err := Lookup(c, id)
	if err != nil || got2.OriginalFilename != "renamed.jpg" {
		t.Fatalf("%v %+v", err, got2)
	}
}

func TestUpdateOriginalFilenameInvalid(t *testing.T) {
	if err := UpdateOriginalFilename(nil, make([]byte, 16), "x"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("nil tx %v", err)
	}
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := UpdateOriginalFilename(tx, []byte{1}, "x"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("short id %v", err)
	}
	if err := UpdateOriginalFilename(tx, make([]byte, 16), "missing"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("missing row %v", err)
	}
}

func TestCount(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	n, err := count(c)
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("empty catalog: got %d want 0", n)
	}

	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	for i, sum := range []string{
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	} {
		id, err := NewID()
		if err != nil {
			t.Fatal(err)
		}
		if err := Insert(tx, File{ID: id, ChecksumSHA256: sum, ByteSize: int64(i)}); err != nil {
			t.Fatal(err)
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	n, err = count(c)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("after insert: got %d want 2", n)
	}
}

func TestMapConstraint(t *testing.T) {
	unique := sqlite3.Error{Code: sqlite3.ErrConstraint, ExtendedCode: sqlite3.ErrConstraintUnique}
	otherConstraint := sqlite3.Error{Code: sqlite3.ErrConstraint, ExtendedCode: sqlite3.ErrConstraintForeignKey}
	plain := errors.New("plain")

	if !IsUniqueConflict(unique) {
		t.Fatal("expected unique")
	}
	if IsUniqueConflict(otherConstraint) || IsUniqueConflict(plain) {
		t.Fatal("unexpected unique")
	}
	if err := mapConstraint(unique); !IsUniqueConflict(err) {
		t.Fatalf("unique passthrough %v", err)
	}
	if err := mapConstraint(otherConstraint); !errors.Is(err, ErrInvalid) {
		t.Fatalf("constraint %v", err)
	}
	if err := mapConstraint(plain); !errors.Is(err, plain) {
		t.Fatalf("plain %v", err)
	}
}

func TestLookupByChecksumTx(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	id, err := NewID()
	if err != nil {
		t.Fatal(err)
	}
	const sum = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	if err := Insert(tx, File{ID: id, ChecksumSHA256: sum, MediaType: "image/png", ByteSize: 3}); err != nil {
		t.Fatal(err)
	}
	got, err := LookupByChecksumTx(tx, sum)
	if err != nil || string(got.ID) != string(id) || got.MediaType != "image/png" {
		t.Fatalf("found: %v %+v", err, got)
	}
	if _, err := LookupByChecksumTx(tx, "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("unknown: %v", err)
	}
	if _, err := LookupByChecksumTx(tx, "xyz"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("bad hex: %v", err)
	}
}

func TestReadRelPath(t *testing.T) {
	const sum = "8fce3b0000000000000000000000000000000000000000000000000000000000"
	tests := []struct {
		name    string
		file    File
		want    string
		wantErr error
	}{
		{name: "png", file: File{ChecksumSHA256: sum, MediaType: "image/png"}, want: "objects/8f/ce/" + sum + ".png"},
		{name: "disallowed type has no path", file: File{ChecksumSHA256: sum, MediaType: "application/x-sh"}},
		{name: "empty type has no path", file: File{ChecksumSHA256: sum}},
		{name: "bad checksum", file: File{ChecksumSHA256: "abcd", MediaType: "image/png"}, wantErr: ErrInvalid},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadRelPath(tt.file)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err %v want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}

func TestRelPaths(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	var stored []File
	for _, f := range []File{
		{ChecksumSHA256: strings.Repeat("a", 64), MediaType: "image/png", ByteSize: 1},
		{ChecksumSHA256: strings.Repeat("b", 64), MediaType: "application/pdf", ByteSize: 2},
	} {
		id, err := NewID()
		if err != nil {
			t.Fatal(err)
		}
		f.ID = id
		if err := Insert(tx, f); err != nil {
			t.Fatal(err)
		}
		stored = append(stored, f)
	}
	// A damaged row: its type is off the allowlist, so it has no path.
	disallowedID, err := NewID()
	if err != nil {
		t.Fatal(err)
	}
	if err := Insert(tx, File{ID: disallowedID, ChecksumSHA256: strings.Repeat("c", 64), MediaType: "application/x-sh", ByteSize: 3}); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	unknown, err := NewID()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		ids  [][]byte
		want map[string]string
	}{
		{name: "no ids", ids: nil, want: map[string]string{}},
		{name: "unknown id omitted", ids: [][]byte{stored[0].ID, stored[1].ID, unknown}, want: map[string]string{}},
		{name: "disallowed type omitted", ids: [][]byte{stored[0].ID, disallowedID}, want: map[string]string{}},
	}
	for _, f := range stored {
		rel, err := StorageRelPath(f.ChecksumSHA256, f.MediaType)
		if err != nil {
			t.Fatal(err)
		}
		tests[1].want[string(f.ID)] = rel
	}
	tests[2].want[string(stored[0].ID)] = tests[1].want[string(stored[0].ID)]
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RelPaths(c, tt.ids)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %v want %v", got, tt.want)
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Fatalf("got %v want %v", got, tt.want)
				}
			}
		})
	}
}
