package depictionvalues

import (
	"bytes"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/files"
	"github.com/mendahu/provenencia/core/depictionvalue"
	"github.com/mendahu/provenencia/core/depictionvalue/depictionvaluetest"
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
	const region = depictionvaluetest.Region

	t.Run("insert and lookup", func(t *testing.T) {
		id, err := InsertTx(tx, depictionvalue.Value{RegionJSON: region, ImageFileID: fileID})
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
		if _, err := InsertTx(tx, depictionvalue.Value{RegionJSON: "  ", ImageFileID: fileID}); !errors.Is(err, depictionvalue.ErrInvalid) {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("short image id", func(t *testing.T) {
		if _, err := InsertTx(tx, depictionvalue.Value{RegionJSON: region, ImageFileID: fileID[:15]}); !errors.Is(err, depictionvalue.ErrInvalid) {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("missing row", func(t *testing.T) {
		if _, err := LookupTx(tx, make([]byte, 16)); !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("got %v", err)
		}
	})
}

func TestFromJoin(t *testing.T) {
	const locatorJSON = depictionvaluetest.Locator
	const region = depictionvaluetest.Region
	offPage := strings.Replace(region, `"artifact_page":1`, `"artifact_page":9`, 1)
	id := []byte("0123456789abcdef")
	img := []byte("fedcba9876543210")
	ns := func(s string) sql.NullString { return sql.NullString{String: s, Valid: true} }
	tests := []struct {
		name       string
		image      []byte
		region     string
		media      string
		wantNil    bool
		wantRegion string
		wantImage  bool
	}{
		{name: "no image is no depiction", image: nil, region: region, media: "image/png", wantNil: true},
		{name: "off-page region", image: img, region: offPage, media: "image/png", wantRegion: "", wantImage: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FromJoin(id, ns(tt.region), tt.image, ns(locatorJSON), ns(tt.media))
			if tt.wantNil {
				if got != nil {
					t.Fatalf("got %+v", got)
				}
				return
			}
			if got == nil {
				t.Fatal("nil")
			}
			if got.RegionJSON != tt.wantRegion || (len(got.ImageFileID) > 0) != tt.wantImage || !bytes.Equal(got.ID, id) {
				t.Fatalf("%+v", got)
			}
		})
	}
	t.Run("does not alias inputs", func(t *testing.T) {
		in := append([]byte(nil), img...)
		inID := append([]byte(nil), id...)
		got := FromJoin(inID, ns(region), in, ns(locatorJSON), ns("image/png"))
		in[0], inID[0] = 'X', 'X'
		if !bytes.Equal(got.ImageFileID, img) || !bytes.Equal(got.ID, id) {
			t.Fatalf("aliased: %+v", got)
		}
	})
}
