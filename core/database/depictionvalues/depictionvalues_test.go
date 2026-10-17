package depictionvalues

import (
	"bytes"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/files"
)

func TestDepictionValues(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()

	fileID, err := files.NewID()
	if err != nil {
		t.Fatal(err)
	}
	if err := files.Insert(tx, files.File{ID: fileID, ChecksumSHA256: strings.Repeat("a", 64), MediaType: "image/png", ByteSize: 1}); err != nil {
		t.Fatal(err)
	}
	const region = `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`

	t.Run("insert and lookup", func(t *testing.T) {
		id, err := InsertTx(tx, Value{RegionJSON: region, ImageFileID: fileID})
		if err != nil {
			t.Fatal(err)
		}
		got, err := LookupTx(tx, id)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got.ID, id) || got.RegionJSON != region || !bytes.Equal(got.ImageFileID, fileID) {
			t.Fatalf("got %+v", got)
		}
	})
	t.Run("empty region", func(t *testing.T) {
		if _, err := InsertTx(tx, Value{RegionJSON: "  ", ImageFileID: fileID}); !errors.Is(err, ErrInvalid) {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("short image id", func(t *testing.T) {
		if _, err := InsertTx(tx, Value{RegionJSON: region, ImageFileID: fileID[:15]}); !errors.Is(err, ErrInvalid) {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("missing row", func(t *testing.T) {
		if _, err := LookupTx(tx, make([]byte, 16)); !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("got %v", err)
		}
	})
}
