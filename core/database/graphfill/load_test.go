package graphfill_test

import (
	"bytes"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/artifacts/artifactstest"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/evidencegraphs"
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjects/subjectstest"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"

	_ "github.com/mendahu/provenencia/core/database/autoreconciler"
	_ "github.com/mendahu/provenencia/core/database/graphfill"
)

var userID = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

// TestRetitleDropsEvidenceGraph: an Evidence graph knows its Source's title, so
// a title edit drops it and the next read has the new title. An id no graph has
// reads as not found. A retitle drops every graph of its Source.
func TestRetitleDropsEvidenceGraph(t *testing.T) {
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
	graphID := evidencegraphstest.SoleGraph(t, c, src.ID).ID
	read := func(id []byte) graphcache.EvidenceGraph {
		t.Helper()
		eg, err := g.EvidenceGraph(id)
		if err != nil {
			t.Fatal(err)
		}
		return eg
	}
	if eg := read(graphID); !eg.Found || eg.SourceTitle != "Register" || !bytes.Equal(eg.SourceID, src.ID) {
		t.Fatalf("evidence graph found %v, title %q, source %x", eg.Found, eg.SourceTitle, eg.SourceID)
	}

	src.Title = "Parish register"
	if _, _, err := writes.Run(c, writes.Op{Action: "update_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		ch, err := sources.Update(tx, userID, src)
		return src, ch, err
	}); err != nil {
		t.Fatal(err)
	}
	if eg := read(graphID); eg.SourceTitle != "Parish register" {
		t.Fatalf("title after the edit %q, want the new one", eg.SourceTitle)
	}

	if eg := read(make([]byte, 16)); eg.Found || eg.SourceTitle != "" {
		t.Fatalf("unknown id found %v, title %q", eg.Found, eg.SourceTitle)
	}

	t.Run("a Source with two graphs", func(t *testing.T) {
		f := newTwoGraphs(t)
		f.read(t, f.def.ID)
		f.read(t, f.jones.ID)
		f.src.Title = "Parish register"
		if _, _, err := writes.Run(f.c, writes.Op{Action: "update_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
			ch, err := sources.Update(tx, userID, f.src)
			return f.src, ch, err
		}); err != nil {
			t.Fatal(err)
		}
		g := f.c.Graph()
		if g.LoadedEvidenceGraphForTest(f.def.ID) || g.LoadedEvidenceGraphForTest(f.jones.ID) {
			t.Fatal("a title edit left a graph of the Source loaded")
		}
		for _, id := range [][]byte{f.def.ID, f.jones.ID} {
			if eg := f.read(t, id); eg.SourceTitle != "Parish register" {
				t.Fatalf("title after the edit %q", eg.SourceTitle)
			}
		}
	})
}

// TestOnlyATitleChangeDropsEvidenceGraph: an Evidence graph holds the title and
// the type off the sources row and nothing else, so a description edit leaves it
// loaded and a title edit drops it.
func TestOnlyATitleChangeDropsEvidenceGraph(t *testing.T) {
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
	graphID := evidencegraphstest.SoleGraph(t, c, src.ID).ID
	load := func() {
		t.Helper()
		if _, err := g.EvidenceGraph(graphID); err != nil {
			t.Fatal(err)
		}
		if !g.LoadedEvidenceGraphForTest(graphID) {
			t.Fatal("evidence graph not loaded")
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
	if !g.LoadedEvidenceGraphForTest(graphID) {
		t.Fatal("a description edit dropped the evidence graph")
	}

	src.Title = "Parish register"
	update(src)
	if g.LoadedEvidenceGraphForTest(graphID) {
		t.Fatal("a title edit left the evidence graph loaded")
	}
}

// TestSourceTypeChangeDropsEvidenceGraph: an Evidence graph knows its Source's
// type, so a type change drops it and the next read has the new type id. The
// graph's name and its Source's ref read back, and a graph with no ref reads "".
func TestSourceTypeChangeDropsEvidenceGraph(t *testing.T) {
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
	book, err := sourcetypes.Upsert(c, sourcetypes.Type{Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book"})
	if err != nil {
		t.Fatal(err)
	}
	letter, err := sourcetypes.Upsert(c, sourcetypes.Type{Key: "letter", Origin: vocabkey.OriginProvenencia, Label: "Letter"})
	if err != nil {
		t.Fatal(err)
	}
	src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: book, Title: "Register"})
	})
	if err != nil {
		t.Fatal(err)
	}
	g := c.Graph()
	sole := evidencegraphstest.SoleGraph(t, c, src.ID)
	read := func() graphcache.EvidenceGraph {
		t.Helper()
		eg, err := g.EvidenceGraph(sole.ID)
		if err != nil {
			t.Fatal(err)
		}
		return eg
	}
	eg := read()
	if !bytes.Equal(eg.SourceTypeID, book) || eg.SourceRef != src.Ref || eg.GraphName != sole.Name || eg.GraphRef != sole.Ref {
		t.Fatalf("graph head type %x ref %q name %q graph ref %q", eg.SourceTypeID, eg.SourceRef, eg.GraphName, eg.GraphRef)
	}

	src.SourceTypeID = letter
	if _, _, err := writes.Run(c, writes.Op{Action: "update_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		ch, err := sources.Update(tx, userID, src)
		return src, ch, err
	}); err != nil {
		t.Fatal(err)
	}
	if g.LoadedEvidenceGraphForTest(sole.ID) {
		t.Fatal("a type change left the evidence graph loaded")
	}
	if eg := read(); !bytes.Equal(eg.SourceTypeID, letter) {
		t.Fatalf("type after the change %x, want the new one", eg.SourceTypeID)
	}

	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE evidence_graphs SET ref = NULL WHERE id = ?`, sole.ID); err != nil {
		t.Fatal(err)
	}
	g.Drop()
	if eg := read(); eg.GraphRef != "" || !eg.Found {
		t.Fatalf("a ref-less graph found %v, ref %q", eg.Found, eg.GraphRef)
	}
}

const pageLocator = `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`

// twoGraphs is one Source ("Register") split into a Default graph holding
// Ada (a person) and a Jones graph holding Bea (a place) and one Observation on her, plus a second
// Source with one graph holding one Subject.
type twoGraphs struct {
	c          *database.Catalog
	src, other sources.Source
	def, jones evidencegraphs.Graph
	ada, bea   subjects.Subject
	away       subjects.Subject
	page1      artifacts.Artifact
	typeID     []byte
}

func newTwoGraphs(t *testing.T) twoGraphs {
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
	if err := users.Upsert(c, userID, "Tester", r); err != nil {
		t.Fatal(err)
	}
	if err := subjectvocab.Install(c); err != nil {
		t.Fatal(err)
	}
	f := twoGraphs{c: c}
	f.typeID, err = sourcetypes.Upsert(c, sourcetypes.Type{Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book"})
	if err != nil {
		t.Fatal(err)
	}
	newSource := func(title string) sources.Source {
		s, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
			return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: f.typeID, Title: title})
		})
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
	person, err := subjecttypes.Lookup(c, "person", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	newSubject := func(src sources.Source, graphID []byte, typeID []byte, label string) subjects.Subject {
		if graphID == nil {
			graphID = evidencegraphstest.SoleGraph(t, c, src.ID).ID
		}
		s := subjectstest.OfType(t, c, userID, graphID, typeID, label)
		return s
	}
	f.src = newSource("Register")
	f.page1, err = artifactstest.Create(c, userID, artifacts.CreateInput{SourceID: f.src.ID, Label: "Page 1"})
	if err != nil {
		t.Fatal(err)
	}
	var page2 artifacts.Artifact
	f.def, f.jones, page2 = artifactstest.Split(t, c, userID, f.src.ID, "Jones")
	f.ada = subjectstest.On(t, c, userID, f.def.ID, "person", "Ada")
	f.bea = subjectstest.On(t, c, userID, f.jones.ID, "place", "Bea")
	toponym, err := properties.Lookup(c, "toponym", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writes.Call(c, writes.Op{Action: "create_citation_with_observations", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
		return citations.CreateWithObservations(tx, userID, citations.CreateInput{GraphID: f.jones.ID, ArtifactID: page2.ID, LocatorJSON: pageLocator},
			[]observations.Input{{SubjectID: f.bea.ID, PropertyID: toponym.ID, ValueText: "Bea", HasText: true}})
	}); err != nil {
		t.Fatal(err)
	}
	f.other = newSource("Elsewhere")
	f.away = newSubject(f.other, nil, person.ID, "Cy")
	return f
}

func (f twoGraphs) read(t *testing.T, graphID []byte) graphcache.EvidenceGraph {
	t.Helper()
	eg, err := f.c.Graph().EvidenceGraph(graphID)
	if err != nil {
		t.Fatal(err)
	}
	return eg
}

func labels(eg graphcache.EvidenceGraph) []string {
	var out []string
	for _, s := range eg.Subjects {
		out = append(out, s.Label)
	}
	return out
}

// TestTwoEvidenceGraphsLoadApart: two graphs of one Source each hold only their
// own Subjects and Observations, and both name the Source and its title.
func TestTwoEvidenceGraphsLoadApart(t *testing.T) {
	f := newTwoGraphs(t)
	def, jones := f.read(t, f.def.ID), f.read(t, f.jones.ID)
	for name, eg := range map[string]graphcache.EvidenceGraph{"Default": def, "Jones": jones} {
		if !eg.Found || eg.SourceTitle != "Register" || !bytes.Equal(eg.SourceID, f.src.ID) {
			t.Fatalf("%s: found %v, title %q, source %x", name, eg.Found, eg.SourceTitle, eg.SourceID)
		}
	}
	if !bytes.Equal(def.GraphID, f.def.ID) || !bytes.Equal(jones.GraphID, f.jones.ID) {
		t.Fatalf("graph ids %x, %x", def.GraphID, jones.GraphID)
	}
	if got := labels(def); len(got) != 1 || got[0] != "Ada" {
		t.Fatalf("Default subjects %v", got)
	}
	if got := labels(jones); len(got) != 1 || got[0] != "Bea" {
		t.Fatalf("Jones subjects %v", got)
	}
	if len(def.Observations) != 0 {
		t.Fatalf("Default observations %+v", def.Observations)
	}
	if len(jones.Observations) != 1 || !bytes.Equal(jones.Observations[0].SubjectID, f.bea.ID) {
		t.Fatalf("Jones observations %+v", jones.Observations)
	}
}

// TestGraphWriteDropsOnlyThatGraph: a write that names an Evidence graph drops
// that loaded graph with its memo, and leaves its Source's other graph and
// another Source's graph.
func TestGraphWriteDropsOnlyThatGraph(t *testing.T) {
	f := newTwoGraphs(t)
	g := f.c.Graph()
	otherGraph := evidencegraphstest.SoleGraph(t, f.c, f.other.ID).ID
	for _, id := range [][]byte{f.def.ID, f.jones.ID, otherGraph} {
		f.read(t, id)
		g.SetEvidenceGraphMemo(id, "k", "memo")
		if _, ok := g.EvidenceGraphMemo(id, "k"); !ok {
			t.Fatalf("memo for %x not held", id)
		}
	}
	subjectstest.On(t, f.c, userID, f.jones.ID, "person", "Cal")
	if g.LoadedEvidenceGraphForTest(f.jones.ID) {
		t.Fatal("the written graph is still loaded")
	}
	if _, ok := g.EvidenceGraphMemo(f.jones.ID, "k"); ok {
		t.Fatal("the memo of the written graph outlived it")
	}
	for _, id := range [][]byte{f.def.ID, otherGraph} {
		if !g.LoadedEvidenceGraphForTest(id) {
			t.Fatalf("a write on Jones dropped graph %x", id)
		}
		if _, ok := g.EvidenceGraphMemo(id, "k"); !ok {
			t.Fatalf("a write on Jones dropped the memo of graph %x", id)
		}
	}
	if got := labels(f.read(t, f.jones.ID)); len(got) != 2 {
		t.Fatalf("Jones after the write %v", got)
	}
}

// TestDeletedGraphIsNotFound: deleting a graph drops it, and the next read of
// its id finds nothing.
func TestDeletedGraphIsNotFound(t *testing.T) {
	f := newTwoGraphs(t)
	empty := evidencegraphstest.Create(t, f.c, userID, f.src.ID, "Empty", f.page1.ID)
	if eg := f.read(t, empty.ID); !eg.Found || len(eg.Subjects) != 0 {
		t.Fatalf("empty graph found %v, subjects %d", eg.Found, len(eg.Subjects))
	}
	g := f.c.Graph()
	if !g.LoadedEvidenceGraphForTest(empty.ID) {
		t.Fatal("graph not loaded")
	}
	if _, _, err := writes.Run(f.c, writes.Op{Action: "delete_evidence_graph", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
		ch, err := evidencegraphs.Delete(tx, userID, empty.ID)
		return struct{}{}, ch, err
	}); err != nil {
		t.Fatal(err)
	}
	if g.LoadedEvidenceGraphForTest(empty.ID) {
		t.Fatal("a deleted graph is still loaded")
	}
	if eg := f.read(t, empty.ID); eg.Found {
		t.Fatalf("a deleted graph reads as found: %+v", eg)
	}
}
