package graphfill_test

import (
	"bytes"
	"crypto/rand"
	"reflect"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/evidencegraphs"
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
	"github.com/mendahu/provenencia/core/database/evidencelayer/evidencelayertest"
	"github.com/mendahu/provenencia/core/database/graphfill"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/promote/promotetest"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/namevalue/namevaluetest"
	"github.com/mendahu/provenencia/core/writes"
)

// family is a Source whose one graph holds Ann (claimed on an entity), Bob
// (unclaimed, named by the Relationship's Observations) and that Relationship.
type family struct {
	r             *evidencelayertest.Records
	graph         []byte
	ann, bob, rel subjects.Subject
	entity        []byte
}

func newFamily(t *testing.T) family {
	t.Helper()
	r := evidencelayertest.New(t)
	f := family{r: r, graph: r.Graph(r.Source)}
	f.ann = r.Labelled(r.Source, "person", "Ann Lee")
	f.bob = r.Labelled(r.Source, "person", "Bob Lee")
	f.rel = r.Relationship(r.Source, r.Artifact, f.ann, f.bob, "parent")
	f.entity = promotetest.Save(t, r.C, evidencelayertest.UserID, promote.Input{SubjectID: f.ann.ID}).Entity.ID
	return f
}

func (f family) records(t *testing.T) *graphfill.GraphRecords {
	t.Helper()
	got, err := graphfill.RecordsOf(f.r.C.Graph(), [][]byte{f.graph})
	if err != nil {
		t.Fatal(err)
	}
	return got[string(f.graph)]
}

// TestGraphRecordsFill: a graph's records are the same whether the loaded
// graph or the rows fill them, and a graph just viewed costs no query.
func TestGraphRecordsFill(t *testing.T) {
	t.Run("loaded and not loaded fill the same records", func(t *testing.T) {
		f := newFamily(t)
		g := f.r.C.Graph()
		g.Drop()
		if _, err := graphfill.EvidenceGraph(g, f.graph); err != nil {
			t.Fatal(err)
		}
		queries := 0
		g.SetQueryCounterForTest(func() { queries++ })
		fromGraph := f.records(t)
		if queries != 0 {
			t.Fatalf("a fill from the loaded graph issued %d queries", queries)
		}
		g.SetQueryCounterForTest(nil)
		graphfill.Records.Drop(g, [][]byte{f.graph})
		g.Drop()
		fromRows := f.records(t)
		if !reflect.DeepEqual(fromGraph, fromRows) {
			t.Fatalf("records differ:\nloaded %+v\nrows   %+v", fromGraph, fromRows)
		}
	})

	t.Run("claimed and unclaimed Subjects", func(t *testing.T) {
		f := newFamily(t)
		rec := f.records(t)
		ann, ok := rec.Subject(f.ann.ID)
		if !ok || !bytes.Equal(ann.Accepted.EntityID, f.entity) || len(ann.Accepted.ClaimID) != 16 {
			t.Fatalf("Ann: %+v, %v", ann.Accepted, ok)
		}
		bob, ok := rec.Subject(f.bob.ID)
		if !ok || bob.Accepted.ClaimID != nil || bob.Accepted.EntityID != nil {
			t.Fatalf("Bob has a claim: %+v, %v", bob.Accepted, ok)
		}
		if got := rec.Labels()[string(f.bob.ID)]; got != "Bob Lee" {
			t.Errorf("Bob's label = %q", got)
		}
		if got := rec.Labels()[string(f.ann.ID)]; got != "Ann Lee" {
			t.Errorf("Ann's label = %q", got)
		}
		if !bytes.Equal(rec.Entity(f.ann.ID), f.entity) {
			t.Errorf("Entity(Ann) = %x, want %x", rec.Entity(f.ann.ID), f.entity)
		}
		if rec.Entity(f.bob.ID) != nil {
			t.Errorf("Entity(Bob) = %x, want nil", rec.Entity(f.bob.ID))
		}
		rel, ok := rec.Subject(f.rel.ID)
		if !ok || len(rel.Observations) == 0 {
			t.Fatalf("the Relationship has no Observations: %+v", rel)
		}
		named := false
		for _, o := range rel.Observations {
			named = named || bytes.Equal(o.ValueSubjectID, f.bob.ID)
		}
		if !named {
			t.Error("no Observation of the Relationship names Bob")
		}
		if _, ok := rec.Subject(make([]byte, 16)); ok {
			t.Error("a Subject that is not on the graph has a record")
		}
	})

	t.Run("an id with no graph row stores empty records", func(t *testing.T) {
		f := newFamily(t)
		id := make([]byte, 16)
		if _, err := rand.Read(id); err != nil {
			t.Fatal(err)
		}
		got, err := graphfill.RecordsOf(f.r.C.Graph(), [][]byte{id})
		if err != nil {
			t.Fatal(err)
		}
		rec := got[string(id)]
		if rec == nil || len(rec.Labels()) != 0 {
			t.Fatalf("records of an unknown graph = %+v", rec)
		}
		if _, ok := rec.Subject(f.ann.ID); ok {
			t.Error("an unknown graph has a Subject")
		}
		if _, ok := graphfill.Records.Peek(f.r.C.Graph(), id); !ok {
			t.Error("empty records were not stored")
		}
	})

	t.Run("a nil graph is closed", func(t *testing.T) {
		if _, err := graphfill.RecordsOf(nil, [][]byte{make([]byte, 16)}); err != database.ErrClosed {
			t.Errorf("err = %v, want database.ErrClosed", err)
		}
	})
}

// TestGraphRecordsFollowCommits: a write that names a graph drops that graph's
// records and no other's, a deleted graph's leave with it, and Drop empties the
// entry; Verify passes after each.
func TestGraphRecordsFollowCommits(t *testing.T) {
	t.Run("a write that names a graph", func(t *testing.T) {
		t.Setenv("PROVENENCIA_GRAPH_VERIFY", "1")
		f := newFamily(t)
		srcB, artB := f.r.NewSource("Second")
		b := f.r.Graph(srcB)
		f.r.Person(srcB, artB, "Cy Day")
		g := f.r.C.Graph()
		got, err := graphfill.RecordsOf(g, [][]byte{f.graph, b})
		if err != nil || len(got) != 2 {
			t.Fatalf("RecordsOf = %v, %v", got, err)
		}
		before, _ := got[string(f.graph)].Subject(f.ann.ID)
		f.r.Cite(f.r.Artifact, f.ann, observations.Input{PropertyID: f.r.Prop("name").ID, Name: namevaluetest.Western("Annie Lee")})
		if _, ok := graphfill.Records.Peek(g, f.graph); ok {
			t.Error("the written graph's records are still held")
		}
		if _, ok := graphfill.Records.Peek(g, b); !ok {
			t.Error("a write on A dropped B's records")
		}
		after, _ := f.records(t).Subject(f.ann.ID)
		if len(after.Observations) != len(before.Observations)+1 {
			t.Errorf("Ann has %d Observations after the write, want %d", len(after.Observations), len(before.Observations)+1)
		}
	})

	t.Run("a deleted graph", func(t *testing.T) {
		f := newFamily(t)
		empty := evidencegraphstest.Create(t, f.r.C, evidencelayertest.UserID, f.r.Source.ID, "Empty", f.r.Artifact.ID)
		g := f.r.C.Graph()
		if _, err := graphfill.RecordsOf(g, [][]byte{empty.ID}); err != nil {
			t.Fatal(err)
		}
		if _, _, err := writes.Run(f.r.C, writes.Op{Action: "delete_evidence_graph", UserID: evidencelayertest.UserID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			ch, err := evidencegraphs.Delete(tx, evidencelayertest.UserID, empty.ID)
			return struct{}{}, ch, err
		}); err != nil {
			t.Fatal(err)
		}
		if _, ok := graphfill.Records.Peek(g, empty.ID); ok {
			t.Error("a deleted graph's records are still held")
		}
	})

	t.Run("Drop", func(t *testing.T) {
		f := newFamily(t)
		f.records(t)
		g := f.r.C.Graph()
		if n := g.KeyCounts()["graphfill.records"]; n == 0 {
			t.Fatal("no records held before Drop")
		}
		g.Drop()
		if n := g.KeyCounts()["graphfill.records"]; n != 0 {
			t.Errorf("%d records held after Drop", n)
		}
	})
}

// TestGraphRecordsOutliveLoadedGraph: records are their own entry, so the
// loaded-graph store evicting a graph leaves its records.
func TestGraphRecordsOutliveLoadedGraph(t *testing.T) {
	f := newFamily(t)
	srcB, _ := f.r.NewSource("Second")
	b := f.r.Graph(srcB)
	g := f.r.C.Graph()
	g.Drop()
	graphfill.Graphs.SetCapForTest(g, 1)
	t.Cleanup(func() { graphfill.Graphs.SetCapForTest(g, 0) })
	if _, err := graphfill.EvidenceGraph(g, f.graph); err != nil {
		t.Fatal(err)
	}
	f.records(t)
	if _, err := graphfill.EvidenceGraph(g, b); err != nil {
		t.Fatal(err)
	}
	if _, ok := graphfill.PeekEvidenceGraph(g, f.graph); ok {
		t.Fatal("graph A is still loaded")
	}
	if _, ok := graphfill.Records.Peek(g, f.graph); !ok {
		t.Error("A's records left with its loaded graph")
	}
}

// TestGraphRecordsCarryCitationWords: every Observation in the records carries
// its Citation's Artifact, label, page and uncertainty, from the loaded graph
// and from the rows alike.
func TestGraphRecordsCarryCitationWords(t *testing.T) {
	cite := func(t *testing.T, f family, in citations.CreateInput) []byte {
		t.Helper()
		in.GraphID, in.ArtifactID = f.graph, f.r.Artifact.ID
		name := f.r.Prop("name").ID
		var res citations.CreateResult
		f.r.Run("create_citation_with_observations", func(tx *database.Tx) ([]rowchange.Change, error) {
			var ch []rowchange.Change
			var err error
			res, ch, err = citations.CreateWithObservations(tx, evidencelayertest.UserID, in, []observations.Input{{
				SubjectID: f.ann.ID, PropertyID: name, Name: namevaluetest.Western("Ada Lovelace"),
			}})
			return ch, err
		})
		if len(res.Observations) != 1 {
			t.Fatalf("created %d Observations", len(res.Observations))
		}
		return res.Observations[0].ID
	}
	fills := func(t *testing.T, f family, id []byte) (loaded, cold *graphfill.GraphRecords) {
		t.Helper()
		g := f.r.C.Graph()
		graphfill.Records.Drop(g, [][]byte{f.graph})
		g.Drop()
		if _, err := graphfill.EvidenceGraph(g, f.graph); err != nil {
			t.Fatal(err)
		}
		loaded = f.records(t)
		graphfill.Records.Drop(g, [][]byte{f.graph})
		g.Drop()
		cold = f.records(t)
		if !reflect.DeepEqual(loaded, cold) {
			t.Fatalf("the two fills differ:\nloaded %+v\ncold   %+v", loaded, cold)
		}
		return loaded, cold
	}

	t.Run("a labelled, paged, uncertain Citation", func(t *testing.T) {
		f := newFamily(t)
		id := cite(t, f, citations.CreateInput{
			Label: "Baptism of Ada", TranscriptionUncertain: true,
			LocatorJSON: `{"version":1,"selectors":[{"type":"page","artifact_page":37}]}`,
		})
		loaded, cold := fills(t, f, id)
		for _, rec := range []*graphfill.GraphRecords{loaded, cold} {
			_, o, ok := rec.Observation(id)
			if !ok {
				t.Fatal("the Observation is not in the records")
			}
			if !bytes.Equal(o.ArtifactID, f.r.Artifact.ID) || o.CitationLabel != "Baptism of Ada" || o.ArtifactPage != 37 || o.PageLabel != "" || !o.TranscriptionUncertain {
				t.Fatalf("citation words %+v", o)
			}
		}
	})

	t.Run("a Citation with no label and no page", func(t *testing.T) {
		f := newFamily(t)
		id := cite(t, f, citations.CreateInput{LocatorJSON: `{"version":1,"selectors":[{"type":"note"}]}`})
		_, cold := fills(t, f, id)
		_, o, ok := cold.Observation(id)
		if !ok {
			t.Fatal("the Observation is not in the records")
		}
		if o.CitationLabel != "" || o.ArtifactPage != 0 || o.PageLabel != "" || o.TranscriptionUncertain {
			t.Fatalf("citation words %+v", o)
		}
	})
}

// TestGraphRecordsObservation: an Observation is found by id with its
// Subject's record.
func TestGraphRecordsObservation(t *testing.T) {
	f := newFamily(t)
	rec := f.records(t)
	rel, ok := rec.Subject(f.rel.ID)
	if !ok || len(rel.Observations) == 0 {
		t.Fatalf("the Relationship has no Observations: %+v, %v", rel, ok)
	}
	want := rel.Observations[0]
	gotRec, got, ok := rec.Observation(want.ID)
	if !ok || !reflect.DeepEqual(got, want) || !bytes.Equal(gotRec.Subject.ID, f.rel.ID) {
		t.Fatalf("Observation = %+v, %+v, %v", gotRec.Subject.ID, got, ok)
	}
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		t.Fatal(err)
	}
	if r, o, ok := rec.Observation(id); ok || r.Subject.ID != nil || o.ID != nil {
		t.Errorf("an unknown id found %+v %+v", r, o)
	}
	var none *graphfill.GraphRecords
	if _, _, ok := none.Observation(want.ID); ok {
		t.Error("a nil records found an Observation")
	}
}
