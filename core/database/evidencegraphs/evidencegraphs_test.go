package evidencegraphs_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/evidencegraphs"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
)

var userID = []byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}

type fixture struct {
	c      *database.Catalog
	typeID []byte
}

func newFixture(t *testing.T) fixture {
	t.Helper()
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	r, err := ref.Mint(ref.PrefixUser)
	if err != nil {
		t.Fatal(err)
	}
	if err := users.Upsert(c, userID, "Jake", r); err != nil {
		t.Fatal(err)
	}
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
		Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book",
	})
	if err != nil {
		t.Fatal(err)
	}
	return fixture{c: c, typeID: typeID}
}

// createSource creates a Source and returns it with the changes of its create.
func (f fixture) createSource(t *testing.T, title string) (sources.Source, []rowchange.Change) {
	t.Helper()
	var changes []rowchange.Change
	src, err := writes.Call(f.c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		s, ch, err := sources.Create(tx, userID, sources.CreateInput{SourceTypeID: f.typeID, Title: title})
		changes = ch
		return s, ch, err
	})
	if err != nil {
		t.Fatal(err)
	}
	return src, changes
}

type graphRow struct {
	id          []byte
	ref         string
	name        string
	description string
}

func graphRows(t *testing.T, c *database.Catalog, sourceID []byte) []graphRow {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	rows, err := db.Query(`SELECT id, COALESCE(ref, ''), name, COALESCE(description, '') FROM evidence_graphs WHERE source_id = ? ORDER BY id`, sourceID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out []graphRow
	for rows.Next() {
		var g graphRow
		if err := rows.Scan(&g.id, &g.ref, &g.name, &g.description); err != nil {
			t.Fatal(err)
		}
		out = append(out, g)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestCreateDefault(t *testing.T) {
	f := newFixture(t)
	src, changes := f.createSource(t, "Census 1881")

	graphs := graphRows(t, f.c, src.ID)
	if len(graphs) != 1 {
		t.Fatalf("graphs %d", len(graphs))
	}
	g := graphs[0]
	if g.name != evidencegraphs.DefaultName {
		t.Fatalf("name %q", g.name)
	}
	if !ref.Valid(g.ref) || !strings.HasPrefix(g.ref, "GPH-") {
		t.Fatalf("ref %q", g.ref)
	}
	if bytes.Equal(g.id, src.ID) {
		t.Fatal("new Default graph reuses the Source id")
	}
	if g.description != "" {
		t.Fatalf("description %q", g.description)
	}
	if len(changes) != 2 || changes[0].EntityType != "source" || changes[1].EntityType != "evidence_graph" {
		t.Fatalf("changes %+v", changes)
	}
	if got := changes[1].Fields["source_id"].New; got != uuid.UUID(src.ID).String() {
		t.Fatalf("graph change source_id %v", got)
	}
	if !bytes.Equal(changes[1].EntityID, g.id) {
		t.Fatal("graph change names another graph")
	}
}

func TestSoleGraphTx(t *testing.T) {
	f := newFixture(t)
	src, _ := f.createSource(t, "Census 1881")
	want := graphRows(t, f.c, src.ID)[0].id

	unknown := make([]byte, 16)
	unknown[0] = 9
	tests := []struct {
		name     string
		sourceID []byte
		want     []byte
		wantErr  error
	}{
		{name: "the Default graph", sourceID: src.ID, want: want},
		{name: "unknown Source", sourceID: unknown, wantErr: evidencegraphs.ErrInvalid},
		{name: "short id", sourceID: unknown[:15], wantErr: evidencegraphs.ErrInvalid},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := f.c.DB()
			if err != nil {
				t.Fatal(err)
			}
			tx, err := db.Begin()
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback()
			got, err := evidencegraphs.SoleGraphTx(tx, tt.sourceID)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("got %v want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Fatalf("got %x want %x", got, tt.want)
			}
		})
	}

	t.Run("two graphs is not ErrInvalid", func(t *testing.T) {
		db, err := f.c.DB()
		if err != nil {
			t.Fatal(err)
		}
		second, err := uuid.NewV7()
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`INSERT INTO evidence_graphs (id, source_id, name) VALUES (?, ?, 'Second')`, second[:], src.ID); err != nil {
			t.Fatal(err)
		}
		tx, err := db.Begin()
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback()
		_, err = evidencegraphs.SoleGraphTx(tx, src.ID)
		if err == nil || errors.Is(err, evidencegraphs.ErrInvalid) {
			t.Fatalf("got %v", err)
		}
	})
}
