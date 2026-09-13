package ingest

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/files"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
)

func TestFile(t *testing.T) {
	userID := []byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}

	mustUser := func(t *testing.T, c *database.Catalog) {
		t.Helper()
		r, err := ref.Mint(ref.PrefixUser)
		if err != nil {
			t.Fatal(err)
		}
		if err := users.Upsert(c, userID, "Jake", r); err != nil {
			t.Fatal(err)
		}
	}
	writeTemp := func(t *testing.T, dir, name string, data []byte) string {
		t.Helper()
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatal(err)
		}
		return path
	}
	auditCount := func(t *testing.T, c *database.Catalog, actionType string) int {
		t.Helper()
		db, err := c.DB()
		if err != nil {
			t.Fatal(err)
		}
		var n int
		if err := db.QueryRow(`SELECT COUNT(*) FROM audit_transactions WHERE action_type = ?`, actionType).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}

	t.Run("ingest new file", func(t *testing.T) {
		c, err := database.Create(t.TempDir(), "t.provenencia")
		if err != nil {
			t.Fatal(err)
		}
		defer c.Close()
		mustUser(t, c)
		srcDir := t.TempDir()
		data := []byte("hello evidence")
		path := writeTemp(t, srcDir, "scan.jpg", data)

		res, err := File(c, path, userID)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(data)
		wantSum := hex.EncodeToString(sum[:])
		if res.File.ChecksumSHA256 != wantSum || res.File.OriginalFilename != "scan.jpg" {
			t.Fatalf("%+v", res.File)
		}
		if res.Reused || res.IngestedAs != "scan.jpg" {
			t.Fatalf("result %+v", res)
		}
		wantRel, err := files.StorageRelPath(wantSum, res.File.MediaType)
		if err != nil || res.RelPath != wantRel {
			t.Fatalf("rel %q want %q err %v", res.RelPath, wantRel, err)
		}
		obj := filepath.Join(c.Dir(), filepath.FromSlash(res.RelPath))
		got, err := os.ReadFile(obj)
		if err != nil || string(got) != string(data) {
			t.Fatalf("object %v %q", err, got)
		}
		if filepath.Base(obj) == "scan.jpg" {
			t.Fatal("object path must not use original filename")
		}
		if auditCount(t, c, "create_file") != 1 {
			t.Fatalf("create_file audit %d", auditCount(t, c, "create_file"))
		}
	})

	t.Run("dedup same bytes records reuse_file", func(t *testing.T) {
		c, err := database.Create(t.TempDir(), "t.provenencia")
		if err != nil {
			t.Fatal(err)
		}
		defer c.Close()
		mustUser(t, c)
		srcDir := t.TempDir()
		data := []byte("same bytes")
		a := writeTemp(t, srcDir, "a.bin", data)
		b := writeTemp(t, srcDir, "b.bin", data)
		first, err := File(c, a, userID)
		if err != nil {
			t.Fatal(err)
		}
		second, err := File(c, b, userID)
		if err != nil {
			t.Fatal(err)
		}
		if string(first.File.ID) != string(second.File.ID) {
			t.Fatal("expected same file id")
		}
		if !second.Reused || second.IngestedAs != "b.bin" {
			t.Fatalf("second %+v", second)
		}
		if second.File.OriginalFilename != "a.bin" {
			t.Fatalf("stored name %q", second.File.OriginalFilename)
		}
		if auditCount(t, c, "create_file") != 1 {
			t.Fatalf("create_file %d", auditCount(t, c, "create_file"))
		}
		if auditCount(t, c, "reuse_file") != 1 {
			t.Fatalf("reuse_file %d", auditCount(t, c, "reuse_file"))
		}
	})

	t.Run("SetFilename overwrites and audits", func(t *testing.T) {
		c, err := database.Create(t.TempDir(), "t.provenencia")
		if err != nil {
			t.Fatal(err)
		}
		defer c.Close()
		mustUser(t, c)
		path := writeTemp(t, t.TempDir(), "old.pdf", []byte("doc"))
		res, err := File(c, path, userID)
		if err != nil {
			t.Fatal(err)
		}
		if err := SetFilename(c, res.File.ID, "grandpas_will.pdf", userID); err != nil {
			t.Fatal(err)
		}
		got, err := files.Lookup(c, res.File.ID)
		if err != nil || got.OriginalFilename != "grandpas_will.pdf" {
			t.Fatalf("%v %+v", err, got)
		}
		if auditCount(t, c, "update_file") != 1 {
			t.Fatalf("update_file %d", auditCount(t, c, "update_file"))
		}
	})

	t.Run("SetFilename no-op same name", func(t *testing.T) {
		c, err := database.Create(t.TempDir(), "t.provenencia")
		if err != nil {
			t.Fatal(err)
		}
		defer c.Close()
		mustUser(t, c)
		path := writeTemp(t, t.TempDir(), "keep.pdf", []byte("same"))
		res, err := File(c, path, userID)
		if err != nil {
			t.Fatal(err)
		}
		if err := SetFilename(c, res.File.ID, "keep.pdf", userID); err != nil {
			t.Fatal(err)
		}
		if auditCount(t, c, "update_file") != 0 {
			t.Fatalf("update_file %d", auditCount(t, c, "update_file"))
		}
	})

	t.Run("SetFilename rejects bad userID", func(t *testing.T) {
		c, err := database.Create(t.TempDir(), "t.provenencia")
		if err != nil {
			t.Fatal(err)
		}
		defer c.Close()
		mustUser(t, c)
		path := writeTemp(t, t.TempDir(), "x.pdf", []byte("x"))
		res, err := File(c, path, userID)
		if err != nil {
			t.Fatal(err)
		}
		if err := SetFilename(c, res.File.ID, "y.pdf", nil); !errors.Is(err, ErrInvalid) {
			t.Fatalf("got %v", err)
		}
		if err := SetFilename(c, res.File.ID, "y.pdf", []byte{1, 2, 3}); !errors.Is(err, ErrInvalid) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("SetFilename rejects missing file", func(t *testing.T) {
		c, err := database.Create(t.TempDir(), "t.provenencia")
		if err != nil {
			t.Fatal(err)
		}
		defer c.Close()
		mustUser(t, c)
		missing := make([]byte, 16)
		missing[15] = 1
		if err := SetFilename(c, missing, "nope.pdf", userID); !errors.Is(err, ErrInvalid) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("SetFilename rejects closed catalog", func(t *testing.T) {
		c, err := database.Create(t.TempDir(), "t.provenencia")
		if err != nil {
			t.Fatal(err)
		}
		mustUser(t, c)
		path := writeTemp(t, t.TempDir(), "z.pdf", []byte("z"))
		res, err := File(c, path, userID)
		if err != nil {
			t.Fatal(err)
		}
		_ = c.Close()
		if err := SetFilename(c, res.File.ID, "renamed.pdf", userID); !errors.Is(err, database.ErrClosed) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("rewrites corrupted object on re-ingest", func(t *testing.T) {
		c, err := database.Create(t.TempDir(), "t.provenencia")
		if err != nil {
			t.Fatal(err)
		}
		defer c.Close()
		mustUser(t, c)
		data := []byte("repair me")
		path := writeTemp(t, t.TempDir(), "x.bin", data)
		res, err := File(c, path, userID)
		if err != nil {
			t.Fatal(err)
		}
		obj := filepath.Join(c.Dir(), filepath.FromSlash(res.RelPath))
		if err := os.WriteFile(obj, []byte("CORRUPT"), 0o644); err != nil {
			t.Fatal(err)
		}
		again := writeTemp(t, t.TempDir(), "y.bin", data)
		if _, err := File(c, again, userID); err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(obj)
		if err != nil || string(got) != string(data) {
			t.Fatalf("object %v %q", err, got)
		}
	})

	t.Run("rejects symlink", func(t *testing.T) {
		c, err := database.Create(t.TempDir(), "t.provenencia")
		if err != nil {
			t.Fatal(err)
		}
		defer c.Close()
		srcDir := t.TempDir()
		target := writeTemp(t, srcDir, "real.txt", []byte("x"))
		link := filepath.Join(srcDir, "link.txt")
		if err := os.Symlink(target, link); err != nil {
			t.Fatal(err)
		}
		if _, err := File(c, link, userID); !errors.Is(err, ErrSymlink) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("rejects directory", func(t *testing.T) {
		c, err := database.Create(t.TempDir(), "t.provenencia")
		if err != nil {
			t.Fatal(err)
		}
		defer c.Close()
		if _, err := File(c, t.TempDir(), userID); !errors.Is(err, ErrNotAFile) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("rejects oversize", func(t *testing.T) {
		c, err := database.Create(t.TempDir(), "t.provenencia")
		if err != nil {
			t.Fatal(err)
		}
		defer c.Close()
		prev := maxBytes
		maxBytes = 4
		defer func() { maxBytes = prev }()
		path := writeTemp(t, t.TempDir(), "big.bin", []byte("12345"))
		_, err = File(c, path, userID)
		if !errors.Is(err, ErrTooLarge) {
			t.Fatalf("got %v", err)
		}
		if ae := apperr.From(err); len(ae.Params()) == 0 {
			t.Fatal("expected size param")
		}
	})

	t.Run("rejects empty", func(t *testing.T) {
		c, err := database.Create(t.TempDir(), "t.provenencia")
		if err != nil {
			t.Fatal(err)
		}
		defer c.Close()
		mustUser(t, c)
		path := writeTemp(t, t.TempDir(), "empty.txt", nil)
		if _, err := File(c, path, userID); !errors.Is(err, ErrEmpty) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("accepts docx zip package", func(t *testing.T) {
		c, err := database.Create(t.TempDir(), "t.provenencia")
		if err != nil {
			t.Fatal(err)
		}
		defer c.Close()
		mustUser(t, c)
		// Minimal ZIP local-file header — DetectContentType → application/zip.
		data := []byte("PK\x03\x04" + strings.Repeat("x", 64))
		path := writeTemp(t, t.TempDir(), "notes.docx", data)
		res, err := File(c, path, userID)
		if err != nil {
			t.Fatal(err)
		}
		want := "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
		if res.File.MediaType != want {
			t.Fatalf("media %q want %q", res.File.MediaType, want)
		}
	})

	t.Run("rejects xlsx zip package", func(t *testing.T) {
		c, err := database.Create(t.TempDir(), "t.provenencia")
		if err != nil {
			t.Fatal(err)
		}
		defer c.Close()
		mustUser(t, c)
		data := []byte("PK\x03\x04" + strings.Repeat("x", 64))
		path := writeTemp(t, t.TempDir(), "table.xlsx", data)
		if _, err := File(c, path, userID); !errors.Is(err, ErrUnsupportedOffice) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("rejects html", func(t *testing.T) {
		c, err := database.Create(t.TempDir(), "t.provenencia")
		if err != nil {
			t.Fatal(err)
		}
		defer c.Close()
		mustUser(t, c)
		path := writeTemp(t, t.TempDir(), "page.html", []byte("<!DOCTYPE html><html></html>"))
		if _, err := File(c, path, userID); !errors.Is(err, ErrUnsupportedType) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("accepts pdf magic", func(t *testing.T) {
		c, err := database.Create(t.TempDir(), "t.provenencia")
		if err != nil {
			t.Fatal(err)
		}
		defer c.Close()
		mustUser(t, c)
		path := writeTemp(t, t.TempDir(), "deed.pdf", []byte("%PDF-1.4\n%\xe2\xe3\xcf\xd3\n"))
		res, err := File(c, path, userID)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(res.File.MediaType, "application/pdf") {
			t.Fatalf("media %q", res.File.MediaType)
		}
	})

	t.Run("rejects relative path", func(t *testing.T) {
		c, err := database.Create(t.TempDir(), "t.provenencia")
		if err != nil {
			t.Fatal(err)
		}
		defer c.Close()
		if _, err := File(c, "relative.bin", userID); !errors.Is(err, ErrInvalid) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("rejects nil userID", func(t *testing.T) {
		c, err := database.Create(t.TempDir(), "t.provenencia")
		if err != nil {
			t.Fatal(err)
		}
		defer c.Close()
		path := writeTemp(t, t.TempDir(), "x.bin", []byte("x"))
		if _, err := File(c, path, nil); !errors.Is(err, ErrInvalid) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("closed catalog", func(t *testing.T) {
		c, err := database.Create(t.TempDir(), "t.provenencia")
		if err != nil {
			t.Fatal(err)
		}
		path := writeTemp(t, t.TempDir(), "x.bin", []byte("x"))
		_ = c.Close()
		if _, err := File(c, path, userID); !errors.Is(err, database.ErrClosed) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("missing path", func(t *testing.T) {
		c, err := database.Create(t.TempDir(), "t.provenencia")
		if err != nil {
			t.Fatal(err)
		}
		defer c.Close()
		if _, err := File(c, filepath.Join(t.TempDir(), "nope"), userID); !errors.Is(err, ErrMissing) {
			t.Fatalf("got %v", err)
		}
	})
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "plain", in: "scan.jpg", want: "scan.jpg"},
		{name: "strips control", in: "a\x00b\nc", want: "abc"},
		{name: "strips c1", in: "a\u007fb\u009fc", want: "abc"},
		{name: "strips bidi override", in: "gpj.\u202Eexe", want: "gpj.exe"},
		{name: "strips bidi isolate", in: "a\u2066b\u2069c", want: "abc"},
		{name: "caps length", in: strings.Repeat("a", 300), want: strings.Repeat("a", 255)},
		{name: "caps on rune boundary", in: strings.Repeat("a", 254) + "é", want: strings.Repeat("a", 254)},
		{name: "base only", in: "/tmp/foo/bar.txt", want: "bar.txt"},
		{name: "dot only", in: ".", want: ""},
		{name: "slash only", in: "/", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeFilename(tt.in); got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}

func TestMapOpenErr(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		want    error
		wantNil bool
	}{
		{name: "nil", err: nil, wantNil: true},
		{name: "not exist", err: os.ErrNotExist, want: ErrMissing},
		{name: "permission", err: os.ErrPermission, want: ErrPermissionDenied},
		{name: "other", err: errors.New("boom"), want: errors.New("boom")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapOpenErr(tt.err)
			if tt.wantNil {
				if got != nil {
					t.Fatalf("got %v", got)
				}
				return
			}
			if tt.name == "other" {
				if got.Error() != "boom" {
					t.Fatalf("got %v", got)
				}
				return
			}
			if !errors.Is(got, tt.want) {
				t.Fatalf("got %v want %v", got, tt.want)
			}
			if tt.want == ErrPermissionDenied {
				if ae := apperr.From(got); ae.Code() != apperr.CodeIngestPermissionDenied {
					t.Fatalf("code %s", ae.Code())
				}
			}
		})
	}
}
