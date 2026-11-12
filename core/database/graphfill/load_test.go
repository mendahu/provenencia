package graphfill_test

import (
	"bytes"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/artifacts/artifactstest"
	"github.com/mendahu/provenencia/core/database/evidencegraphs"
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/observations/observationstest"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjects/subjectstest"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"

	_ "github.com/mendahu/provenencia/core/database/autoreconciler"
	_ "github.com/mendahu/provenencia/core/database/graphfill"
)

var userID = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

// newSingleSource is a catalog with one Source ("Register") and a second
// source type, for the tests that edit the Source of a loaded graph.
func newSingleSource(t *testing.T) (c *database.Catalog, src sources.Source, book, letter []byte) {
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
	book, err = sourcetypes.Upsert(c, sourcetypes.Type{Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book"})
	if err != nil {
		t.Fatal(err)
	}
	letter, err = sourcetypes.Upsert(c, sourcetypes.Type{Key: "letter", Origin: vocabkey.OriginProvenencia, Label: "Letter"})
	if err != nil {
		t.Fatal(err)
	}
	src, _, err = writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: book, Title: "Register"})
	})
	if err != nil {
		t.Fatal(err)
	}
	return c, src, book, letter
}

func updateSource(t *testing.T, c *database.Catalog, s sources.Source) {
	t.Helper()
	if _, _, err := writes.Run(c, writes.Op{Action: "update_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		ch, err := sources.Update(tx, userID, s)
		return s, ch, err
	}); err != nil {
		t.Fatal(err)
	}
}

// TestRetitleKeepsEvidenceGraph: an Evidence graph holds no Source words, so a
// rename leaves every loaded graph of the Source loaded with its memo. An id no
// graph has reads as not found.
func TestRetitleKeepsEvidenceGraph(t *testing.T) {
	f := newTwoGraphs(t)
	g := f.c.Graph()
	for _, id := range [][]byte{f.def.ID, f.jones.ID} {
		f.read(t, id)
		g.SetEvidenceGraphMemo(id, "k", "memo")
	}
	f.src.Title = "Parish register"
	updateSource(t, f.c, f.src)
	for _, id := range [][]byte{f.def.ID, f.jones.ID} {
		if !g.LoadedEvidenceGraphForTest(id) {
			t.Fatalf("a rename dropped graph %x", id)
		}
		if _, ok := g.EvidenceGraphMemo(id, "k"); !ok {
			t.Fatalf("a rename dropped the memo of graph %x", id)
		}
		if eg := f.read(t, id); !bytes.Equal(eg.SourceID, f.src.ID) {
			t.Fatalf("source %x after the rename", eg.SourceID)
		}
	}
	if eg := f.read(t, make([]byte, 16)); eg.Found || eg.SourceID != nil {
		t.Fatalf("unknown id found %v, source %x", eg.Found, eg.SourceID)
	}
}

// TestSourceEditKeepsEvidenceGraph: a description edit leaves a loaded graph
// loaded, and a Subject create in one graph drops that graph only.
func TestSourceEditKeepsEvidenceGraph(t *testing.T) {
	f := newTwoGraphs(t)
	g := f.c.Graph()
	f.read(t, f.def.ID)
	f.read(t, f.jones.ID)
	f.src.Description = "Baptisms 1790 to 1820"
	updateSource(t, f.c, f.src)
	if !g.LoadedEvidenceGraphForTest(f.def.ID) || !g.LoadedEvidenceGraphForTest(f.jones.ID) {
		t.Fatal("a description edit dropped an evidence graph")
	}
	subjectstest.On(t, f.c, userID, f.def.ID, "person", "Cal")
	if g.LoadedEvidenceGraphForTest(f.def.ID) {
		t.Fatal("a Subject create left its graph loaded")
	}
	if !g.LoadedEvidenceGraphForTest(f.jones.ID) {
		t.Fatal("a Subject create in one graph dropped the other")
	}
}

// TestSourceTypeChangeKeepsEvidenceGraph: a type change leaves the graph
// loaded. The graph's name and ref read back, and a graph with no ref reads "".
func TestSourceTypeChangeKeepsEvidenceGraph(t *testing.T) {
	c, src, _, letter := newSingleSource(t)
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
	if eg := read(); !bytes.Equal(eg.SourceID, src.ID) || eg.GraphName != sole.Name || eg.GraphRef != sole.Ref {
		t.Fatalf("graph head source %x name %q graph ref %q", eg.SourceID, eg.GraphName, eg.GraphRef)
	}
	src.SourceTypeID = letter
	updateSource(t, c, src)
	if !g.LoadedEvidenceGraphForTest(sole.ID) {
		t.Fatal("a type change dropped the evidence graph")
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
	if _, _, err := observationstest.Cite(c, userID, f.jones.ID, page2.ID, pageLocator, []observations.Input{{SubjectID: f.bea.ID, PropertyID: toponym.ID, Scalar: propertymeta.TextScalar("Bea")}}...); err != nil {
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
// own Subjects and Observations, and both name the Source.
func TestTwoEvidenceGraphsLoadApart(t *testing.T) {
	f := newTwoGraphs(t)
	def, jones := f.read(t, f.def.ID), f.read(t, f.jones.ID)
	for name, eg := range map[string]graphcache.EvidenceGraph{"Default": def, "Jones": jones} {
		if !eg.Found || !bytes.Equal(eg.SourceID, f.src.ID) {
			t.Fatalf("%s: found %v, source %x", name, eg.Found, eg.SourceID)
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
