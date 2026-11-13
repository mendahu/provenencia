package evidencegraphs_test

import (
	"bytes"
	"errors"
	"slices"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/artifacts/artifactstest"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/evidencegraphs"
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
	"github.com/mendahu/provenencia/core/database/evidencelayer/evidencelayertest"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjects/subjectstest"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/namevalue/namevaluetest"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"

	_ "github.com/mendahu/provenencia/core/database/autoreconciler" // registers Verify's truth recompute
)

// derivedFromIndex is an entry another PR would register: it is emptied by
// any commit that reloads a graph or its counts.
var derivedFromIndex = graphcache.RegisterWhole(graphcache.WholeEntry[string]{
	Name:        "evidencegraphs_test.derived",
	DerivedFrom: []graphcache.Source{evidencegraphs.Index},
})

type indexFix struct {
	r                 *evidencelayertest.Records
	ada               subjects.Subject
	def, jones, apple evidencegraphs.Graph
	page2             artifacts.Artifact
	second            sources.Source
}

// indexFixture: a Source with graphs Default, jones (two Artifacts) and Apple,
// a person with one Observation on Default, and a second Source.
func indexFixture(t *testing.T) indexFix {
	t.Helper()
	r := evidencelayertest.New(t)
	f := indexFix{r: r}
	// Person reads SoleGraph, so it runs while the Source has one graph.
	f.ada = r.Person(r.Source, r.Artifact, "Ada")
	f.def = evidencegraphstest.SoleGraph(t, r.C, r.Source.ID)
	var err error
	f.page2, err = artifactstest.Create(r.C, evidencelayertest.UserID, artifacts.CreateInput{SourceID: r.Source.ID, Label: "Page 2"})
	evidencelayertest.Must(t, err)
	f.jones = evidencegraphstest.Create(t, r.C, evidencelayertest.UserID, r.Source.ID, "jones", r.Artifact.ID, f.page2.ID)
	f.apple = evidencegraphstest.Create(t, r.C, evidencelayertest.UserID, r.Source.ID, "Apple", r.Artifact.ID)
	f.second, _ = r.NewSource("Second")
	return f
}

func graphIndex(t *testing.T, r *evidencelayertest.Records) *evidencegraphs.GraphIndex {
	t.Helper()
	idx, err := evidencegraphs.IndexOf(r.C.Graph())
	evidencelayertest.Must(t, err)
	return idx
}

func names(gs []evidencegraphs.Graph) []string {
	out := make([]string, len(gs))
	for i, g := range gs {
		out[i] = g.Name
	}
	return out
}

// TestGraphIndexBuild: a Source's graphs are in list order, a graph carries
// its row and links, counts are held, and the project lists Sources by id.
func TestGraphIndexBuild(t *testing.T) {
	f := indexFixture(t)
	idx := graphIndex(t, f.r)

	if got := names(idx.OfSource(f.r.Source.ID)); !slices.Equal(got, []string{"Apple", "Default", "jones"}) {
		t.Fatalf("OfSource = %v", got)
	}
	got, ok := idx.Graph(f.jones.ID)
	if !ok {
		t.Fatal("no jones")
	}
	links := []artifacts.Artifact{f.r.Artifact, f.page2}
	slices.SortFunc(links, func(a, b artifacts.Artifact) int {
		if c := bytes.Compare([]byte(lower(a.Ref)), []byte(lower(b.Ref))); c != 0 {
			return c
		}
		return bytes.Compare(a.ID, b.ID)
	})
	want := evidencegraphs.Graph{
		ID: f.jones.ID, Ref: f.jones.Ref, SourceID: f.r.Source.ID, Name: "jones",
		ArtifactIDs: [][]byte{links[0].ID, links[1].ID},
	}
	if !graphEqual(got, want) {
		t.Fatalf("Graph(jones) = %+v, want %+v", got, want)
	}
	if p := idx.Progress(f.def.ID); p.SubjectCount != 1 || p.FiledCount != 0 || p.ObservationCount != 1 {
		t.Fatalf("Progress(default) = %+v", p)
	}
	if p := idx.Progress(f.jones.ID); p.SubjectCount != 0 || p.ObservationCount != 0 || len(p.GraphID) != 16 {
		t.Fatalf("Progress(jones) = %+v", p)
	}
	if n := len(idx.AllProgress()); n != 1 {
		t.Fatalf("AllProgress has %d rows, want 1 (Default; the other graphs have no counts)", n)
	}

	lo, hi := f.r.Source, f.second
	if bytes.Compare(hi.ID, lo.ID) < 0 {
		lo, hi = hi, lo
	}
	all := idx.All()
	if len(all) != 4 {
		t.Fatalf("All has %d graphs, want 4", len(all))
	}
	if !bytes.Equal(all[0].SourceID, lo.ID) || !bytes.Equal(all[len(all)-1].SourceID, hi.ID) {
		t.Fatalf("All starts with Source %x and ends with %x, want %x then %x", all[0].SourceID, all[3].SourceID, lo.ID, hi.ID)
	}
}

func lower(s string) string { return string(bytes.ToLower([]byte(s))) }

func graphEqual(a, b evidencegraphs.Graph) bool {
	return bytes.Equal(a.ID, b.ID) && a.Ref == b.Ref && bytes.Equal(a.SourceID, b.SourceID) &&
		a.Name == b.Name && a.Description == b.Description && slices.EqualFunc(a.ArtifactIDs, b.ArtifactIDs, bytes.Equal)
}

// TestGraphIndexApply: every write that changes a graph, its links or its
// counts leaves the held index equal to a fresh build, and empties what is
// derived from it; a Source rename does neither.
func TestGraphIndexApply(t *testing.T) {
	f := indexFixture(t)
	r := f.r
	g := r.C.Graph()
	graphIndex(t, r)
	uid := evidencelayertest.UserID
	var smith evidencegraphs.Graph
	var hull subjects.Subject

	check := func(t *testing.T, derivedEmptied bool, write func()) {
		t.Helper()
		derivedFromIndex.Set(g, "held")
		write()
		if err := g.Verify(); err != nil {
			t.Fatalf("Verify after the write: %v", err)
		}
		_, held, _ := derivedFromIndex.Get(g)
		if held == derivedEmptied {
			t.Fatalf("derived entry held = %v after the write, want %v", held, !derivedEmptied)
		}
	}
	has := func(t *testing.T, sourceID []byte, name string) bool {
		t.Helper()
		return slices.Contains(names(graphIndex(t, r).OfSource(sourceID)), name)
	}
	run := func(action string, fn func(tx *database.Tx) ([]rowchange.Change, error)) {
		r.Run(action, fn)
	}

	t.Run("graph create", func(t *testing.T) {
		check(t, true, func() {
			smith = evidencegraphstest.Create(t, r.C, uid, r.Source.ID, "Smith", r.Artifact.ID)
		})
		if !has(t, r.Source.ID, "Smith") {
			t.Fatal("Smith not listed")
		}
	})
	t.Run("graph rename", func(t *testing.T) {
		check(t, true, func() {
			run("update_evidence_graph", func(tx *database.Tx) ([]rowchange.Change, error) {
				return evidencegraphs.Update(tx, uid, f.jones.ID, evidencegraphs.UpdateInput{Name: "Zed"})
			})
		})
		ns := names(graphIndex(t, r).OfSource(r.Source.ID))
		if ns[len(ns)-1] != "Zed" {
			t.Fatalf("order after rename %v", ns)
		}
	})
	t.Run("link", func(t *testing.T) {
		check(t, true, func() {
			run("link_artifact", func(tx *database.Tx) ([]rowchange.Change, error) {
				return evidencegraphs.Link(tx, uid, smith.ID, f.page2.ID)
			})
		})
		got, _ := graphIndex(t, r).Graph(smith.ID)
		if len(got.ArtifactIDs) != 2 {
			t.Fatalf("links %d", len(got.ArtifactIDs))
		}
	})
	t.Run("unlink", func(t *testing.T) {
		check(t, true, func() {
			run("unlink_artifact", func(tx *database.Tx) ([]rowchange.Change, error) {
				return evidencegraphs.Unlink(tx, uid, smith.ID, f.page2.ID)
			})
		})
		got, _ := graphIndex(t, r).Graph(smith.ID)
		if len(got.ArtifactIDs) != 1 {
			t.Fatalf("links %d", len(got.ArtifactIDs))
		}
	})
	t.Run("subject create", func(t *testing.T) {
		check(t, true, func() { hull = subjectstest.On(t, r.C, uid, smith.ID, "place", "Hull") })
		if p := graphIndex(t, r).Progress(smith.ID); p.SubjectCount != 1 {
			t.Fatalf("Smith counts %+v", p)
		}
	})
	t.Run("observation create", func(t *testing.T) {
		check(t, true, func() {
			r.Cite(r.Artifact, f.ada, observations.Input{PropertyID: r.Prop("name").ID, Name: namevaluetest.Western("Ada Lee")})
		})
		if p := graphIndex(t, r).Progress(f.def.ID); p.ObservationCount != 2 {
			t.Fatalf("Default counts %+v", p)
		}
	})
	t.Run("claim accept", func(t *testing.T) {
		person, err := subjecttypes.Lookup(r.C, "person", vocabkey.OriginProvenencia)
		evidencelayertest.Must(t, err)
		db, err := r.C.DB()
		evidencelayertest.Must(t, err)
		entity := []byte{9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 2}
		_, err = db.Exec(`INSERT INTO canonical_entities (id, subject_type_id, ref) VALUES (?, ?, ?)`, entity, person.ID, "CNE-TEST2")
		evidencelayertest.Must(t, err)
		check(t, true, func() {
			run("create_identity_claim", func(tx *database.Tx) ([]rowchange.Change, error) {
				_, ch, err := identityclaims.Create(tx, uid, identityclaims.CreateInput{SubjectID: f.ada.ID, EntityID: entity, Status: identityclaims.StatusAccepted})
				return ch, err
			})
		})
		if p := graphIndex(t, r).Progress(f.def.ID); p.FiledCount != 1 {
			t.Fatalf("Default counts %+v", p)
		}
	})
	t.Run("subject delete", func(t *testing.T) {
		check(t, true, func() {
			run("delete_subject", func(tx *database.Tx) ([]rowchange.Change, error) {
				return subjects.Delete(tx, uid, hull.ID)
			})
		})
		if p := graphIndex(t, r).Progress(smith.ID); p.SubjectCount != 0 || p.ObservationCount != 0 {
			t.Fatalf("Smith counts %+v", p)
		}
	})
	t.Run("graph delete", func(t *testing.T) {
		check(t, true, func() {
			run("delete_evidence_graph", func(tx *database.Tx) ([]rowchange.Change, error) {
				return evidencegraphs.Delete(tx, uid, smith.ID)
			})
		})
		idx := graphIndex(t, r)
		if _, ok := idx.Graph(smith.ID); ok || has(t, r.Source.ID, "Smith") {
			t.Fatal("Smith still listed")
		}
		if p := idx.Progress(smith.ID); p.SubjectCount != 0 {
			t.Fatalf("Smith counts %+v", p)
		}
	})
	t.Run("source rename", func(t *testing.T) {
		src := r.Source
		src.Title = "Parish register"
		check(t, false, func() {
			run("update_source", func(tx *database.Tx) ([]rowchange.Change, error) {
				return sources.Update(tx, uid, src)
			})
		})
	})
	t.Run("source delete", func(t *testing.T) {
		typeID := r.Source.SourceTypeID
		third, _, err := writes.Run(r.C, writes.Op{Action: "create_source", UserID: uid}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
			return sources.Create(tx, uid, sources.CreateInput{SourceTypeID: typeID, Title: "Third"})
		})
		evidencelayertest.Must(t, err)
		if len(graphIndex(t, r).OfSource(third.ID)) != 1 {
			t.Fatal("the new Source has no Default graph")
		}
		check(t, true, func() {
			run("delete_source", func(tx *database.Tx) ([]rowchange.Change, error) {
				return sources.Delete(tx, uid, third.ID)
			})
		})
		if graphIndex(t, r).OfSource(third.ID) != nil {
			t.Fatal("the deleted Source still has graphs")
		}
	})
}

// TestGraphIndexApplyMove: a graph that changes Source ends under its new
// Source, whichever order the commit names the two in.
func TestGraphIndexApplyMove(t *testing.T) {
	for _, tc := range []struct {
		name  string
		order func(a, b []byte) [][]byte
	}{
		{"first then second", func(a, b []byte) [][]byte { return [][]byte{a, b} }},
		{"second then first", func(a, b []byte) [][]byte { return [][]byte{b, a} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := indexFixture(t)
			r := f.r
			g := r.C.Graph()
			a, b := r.Source, f.second
			graphIndex(t, r)
			db, err := r.C.DB()
			evidencelayertest.Must(t, err)
			// A direct-SQL write reports no effects; the commit is announced by hand.
			_, err = db.Exec(`UPDATE evidence_graphs SET source_id = ? WHERE id = ?`, b.ID, f.apple.ID)
			evidencelayertest.Must(t, err)
			evidencelayertest.Must(t, g.OnCommit(effects.Commit{Effects: effects.Set{
				Reads: []effects.ReadIDs{{Read: effects.ReadEvidenceGraphs, IDs: tc.order(a.ID, b.ID)}},
			}}))
			idx := graphIndex(t, r)
			if slices.Contains(names(idx.OfSource(a.ID)), "Apple") {
				t.Fatal("Apple still under the first Source")
			}
			if got := names(idx.OfSource(b.ID)); !slices.Equal(got, []string{"Apple", "Default"}) {
				t.Fatalf("second Source's graphs %v", got)
			}
			if gr, _ := idx.Graph(f.apple.ID); !bytes.Equal(gr.SourceID, b.ID) {
				t.Fatalf("Graph(Apple).SourceID = %x", gr.SourceID)
			}
			evidencelayertest.Must(t, g.Verify())
		})
	}
}

// TestIndexOf: a nil graph is a closed catalog; a write's graph builds none.
func TestIndexOf(t *testing.T) {
	r := evidencelayertest.New(t)
	if _, err := evidencegraphs.IndexOf(nil); !errors.Is(err, database.ErrClosed) {
		t.Fatalf("IndexOf(nil): %v", err)
	}
	r.Run("probe", func(tx *database.Tx) ([]rowchange.Change, error) {
		queries := 0
		over := graphcache.Over(tx.Tx)
		over.SetQueryCounterForTest(func() { queries++ })
		if _, err := evidencegraphs.IndexOf(over); !errors.Is(err, database.ErrWholeOnWrite) {
			t.Fatalf("IndexOf on a write's graph: %v", err)
		}
		if queries != 0 {
			t.Fatalf("a write's graph ran %d queries", queries)
		}
		return nil, nil
	})
}
