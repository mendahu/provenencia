package sourcefill_test

import (
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"

	_ "github.com/mendahu/provenencia/core/database/autoreconciler"
	_ "github.com/mendahu/provenencia/core/database/sourcefill"
)

var userID = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

// TestRetitleDropsSourceGraph: a Source graph knows its title, so a title
// edit drops it and the next read has the new title. An id no Source has reads
// as not found.
func TestRetitleDropsSourceGraph(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	r, err := ref.Mint(ref.PrefixUser)
	if err != nil {
		t.Fatal(err)
	}
	if err := users.Upsert(c, userID, "Tester", r); err != nil {
		t.Fatal(err)
	}
	if err := subjectvocab.Install(c); err != nil {
		t.Fatal(err)
	}
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book"})
	if err != nil {
		t.Fatal(err)
	}
	src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Register"})
	})
	if err != nil {
		t.Fatal(err)
	}
	g := c.Graph()
	read := func(id []byte) graphcache.SourceGraph {
		t.Helper()
		sg, err := g.Source(id)
		if err != nil {
			t.Fatal(err)
		}
		return sg
	}
	if sg := read(src.ID); !sg.Found || sg.Title != "Register" {
		t.Fatalf("source graph found %v, title %q", sg.Found, sg.Title)
	}

	src.Title = "Parish register"
	if _, _, err := writes.Run(c, writes.Op{Action: "update_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		ch, err := sources.Update(tx, userID, src)
		return src, ch, err
	}); err != nil {
		t.Fatal(err)
	}
	if sg := read(src.ID); sg.Title != "Parish register" {
		t.Fatalf("title after the edit %q, want the new one", sg.Title)
	}

	if sg := read(make([]byte, 16)); sg.Found || sg.Title != "" {
		t.Fatalf("unknown id found %v, title %q", sg.Found, sg.Title)
	}
}

// TestOnlyATitleChangeDropsSourceGraph: a Source graph holds the title and
// nothing else off the sources row, so a description edit leaves it loaded and
// a title edit drops it.
func TestOnlyATitleChangeDropsSourceGraph(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	r, err := ref.Mint(ref.PrefixUser)
	if err != nil {
		t.Fatal(err)
	}
	if err := users.Upsert(c, userID, "Tester", r); err != nil {
		t.Fatal(err)
	}
	if err := subjectvocab.Install(c); err != nil {
		t.Fatal(err)
	}
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book"})
	if err != nil {
		t.Fatal(err)
	}
	src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Register"})
	})
	if err != nil {
		t.Fatal(err)
	}
	g := c.Graph()
	load := func() {
		t.Helper()
		if _, err := g.Source(src.ID); err != nil {
			t.Fatal(err)
		}
		if !g.LoadedSourceForTest(src.ID) {
			t.Fatal("source graph not loaded")
		}
	}
	update := func(s sources.Source) {
		t.Helper()
		if _, _, err := writes.Run(c, writes.Op{Action: "update_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
			ch, err := sources.Update(tx, userID, s)
			return s, ch, err
		}); err != nil {
			t.Fatal(err)
		}
	}

	load()
	src.Description = "Baptisms 1790 to 1820"
	update(src)
	if !g.LoadedSourceForTest(src.ID) {
		t.Fatal("a description edit dropped the source graph")
	}

	src.Title = "Parish register"
	update(src)
	if g.LoadedSourceForTest(src.ID) {
		t.Fatal("a title edit left the source graph loaded")
	}
}
