package search

import (
	"context"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/artifacts/artifactstest"
	"github.com/mendahu/provenencia/core/database/catalogmodel"
	"github.com/mendahu/provenencia/core/database/deleteimpact"
	"github.com/mendahu/provenencia/core/database/evidencegraphs"
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/searchindex"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/writes"
)

type graphSearchFixture struct {
	*handleFixture
	jones, smith evidencegraphs.Graph
	sourceRef    string
}

const jonesDescription = "Thomas Jones and family, page 2"

// newGraphSearchFixture is the handle fixture's Source "Register" (Artifact
// "Scan", one Default graph) with an Artifact "Page 2" and two more graphs.
func newGraphSearchFixture(t *testing.T) *graphSearchFixture {
	t.Helper()
	f := newHandleFixture(t)
	page2, err := artifactstest.Create(f.c, f.user, artifacts.CreateInput{SourceID: f.source, Label: "Page 2"})
	if err != nil {
		t.Fatal(err)
	}
	jones, err := writes.Call(f.c, writes.Op{Action: "create_evidence_graph", UserID: f.user}, func(tx *database.Tx) (evidencegraphs.Graph, []rowchange.Change, error) {
		return evidencegraphs.Create(tx, f.user, evidencegraphs.CreateInput{
			SourceID: f.source, Name: "Jones household", Description: jonesDescription,
			ArtifactIDs: [][]byte{page2.ID},
		})
	})
	if err != nil {
		t.Fatal(err)
	}
	smith := evidencegraphstest.Create(t, f.c, f.user, f.source, "Smith household", f.artifact)
	src, err := sources.Get(f.c, f.source)
	if err != nil {
		t.Fatal(err)
	}
	return &graphSearchFixture{handleFixture: f, jones: jones, smith: smith, sourceRef: src.Ref}
}

func (f *graphSearchFixture) search(text string, kinds ...string) []Hit {
	f.t.Helper()
	hits, err := DefaultEngine().Search(context.Background(), f.c, Query{Text: text, Kinds: kinds})
	if err != nil {
		f.t.Fatal(err)
	}
	return hits
}

// graphHits are the evidence_graph hits of a default-kinds search.
func (f *graphSearchFixture) graphHits(text string) []Hit {
	f.t.Helper()
	var out []Hit
	for _, h := range f.search(text) {
		if h.Kind == KindEvidenceGraph {
			out = append(out, h)
		}
	}
	return out
}

// graphHit is the evidence_graph hit for the graph, if the search returns one.
func (f *graphSearchFixture) graphHit(text string, g evidencegraphs.Graph) (Hit, bool) {
	f.t.Helper()
	for _, h := range f.graphHits(text) {
		if h.ID == uuidString(g.ID) {
			return h, true
		}
	}
	return Hit{}, false
}

func (f *graphSearchFixture) jonesLocation() WorkspaceLocation {
	return WorkspaceLocation{
		Section: SectionSources, SourceID: uuidString(f.source), GraphID: uuidString(f.jones.ID),
		SourceSurface: SurfaceGraph, Ref: f.sourceRef, Title: "Register", SourceTitle: "Register",
	}
}

func TestEvidenceGraphSearch(t *testing.T) {
	f := newGraphSearchFixture(t)
	if f.jones.Ref == "" {
		t.Fatal("Jones household has no ref")
	}
	lower := f.jones.Ref[:6]
	tests := []struct {
		name       string
		query      string
		kinds      []string
		wantFirst  bool // the Jones hit must be the first hit
		wantReason string
		wantNone   bool
	}{
		{name: "name", query: "jones hou", wantReason: "title"},
		{name: "description", query: "thomas", wantReason: "description"},
		{name: "exact ref", query: f.jones.Ref, wantFirst: true, wantReason: "ref"},
		{name: "prefix ref lower case", query: lowerASCII(lower), wantReason: "ref"},
		{name: "source title only", query: "register", wantNone: true},
		{name: "other kind only", query: "jones", kinds: []string{KindPerson}, wantNone: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hits := f.search(tc.query, tc.kinds...)
			var got *Hit
			for i := range hits {
				if hits[i].Kind == KindEvidenceGraph && hits[i].ID == uuidString(f.jones.ID) {
					got = &hits[i]
					break
				}
			}
			if tc.wantNone {
				for _, h := range hits {
					if h.Kind == KindEvidenceGraph {
						t.Fatalf("got graph hit %+v", h)
					}
				}
				return
			}
			if got == nil {
				t.Fatalf("no Jones hit in %+v", hits)
			}
			if tc.wantFirst && hits[0].ID != got.ID {
				t.Fatalf("first hit is %+v", hits[0])
			}
			if got.Title != "Jones household" || got.Subtitle != "Register" || got.Ref != f.jones.Ref {
				t.Fatalf("hit %+v", *got)
			}
			if got.MatchReason != tc.wantReason {
				t.Fatalf("reason %q, want %q", got.MatchReason, tc.wantReason)
			}
			if want := f.jonesLocation(); got.Location != want {
				t.Fatalf("location %+v, want %+v", got.Location, want)
			}
		})
	}

	t.Run("default graph per source", func(t *testing.T) {
		hits := f.graphHits("default")
		if len(hits) != 1 || hits[0].Title != "Default" || hits[0].Subtitle != "Register" {
			t.Fatalf("hits %+v", hits)
		}
	})
}

func lowerASCII(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 'a' - 'A'
		}
	}
	return string(b)
}

func TestEvidenceGraphSearchFollowsWrites(t *testing.T) {
	update := func(f *graphSearchFixture, g evidencegraphs.Graph, name, description string) {
		f.t.Helper()
		_, err := writes.Call(f.c, writes.Op{Action: "update_evidence_graph", UserID: f.user}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			ch, err := evidencegraphs.Update(tx, f.user, g.ID, evidencegraphs.UpdateInput{Name: name, Description: description})
			return struct{}{}, ch, err
		})
		if err != nil {
			f.t.Fatal(err)
		}
	}
	tests := []struct {
		name  string
		write func(f *graphSearchFixture)
		check func(t *testing.T, f *graphSearchFixture)
	}{
		{
			name:  "rename a graph",
			write: func(f *graphSearchFixture) { update(f, f.jones, "Jones family", jonesDescription) },
			check: func(t *testing.T, f *graphSearchFixture) {
				// Multi-token search retrieves on any token ("jones", "hou"), so the
				// renamed graph may still be a partial hit: it must carry the new name.
				if h, ok := f.graphHit("jones fam", f.jones); !ok || h.Title != "Jones family" {
					t.Fatalf("jones fam: %+v", h)
				}
				for _, h := range f.graphHits("jones hou") {
					if h.Title == "Jones household" {
						t.Fatalf("stale name: %+v", h)
					}
				}
			},
		},
		{
			name:  "edit the description",
			write: func(f *graphSearchFixture) { update(f, f.jones, "Jones household", "Mary Jones") },
			check: func(t *testing.T, f *graphSearchFixture) {
				if h, ok := f.graphHit("thomas", f.jones); ok {
					t.Fatalf("thomas: %+v", h)
				}
				if _, ok := f.graphHit("mary", f.jones); !ok {
					t.Fatal("mary: no hit")
				}
			},
		},
		{
			name: "rename the Source",
			write: func(f *graphSearchFixture) {
				s, err := sources.Get(f.c, f.source)
				if err != nil {
					t.Fatal(err)
				}
				s.Title = "Register (Toronto)"
				_, err = writes.Call(f.c, writes.Op{Action: "update_source", UserID: f.user}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
					ch, err := sources.Update(tx, f.user, s)
					return struct{}{}, ch, err
				})
				if err != nil {
					t.Fatal(err)
				}
			},
			check: func(t *testing.T, f *graphSearchFixture) {
				h := f.graphHits("jones")
				if len(h) != 1 || h[0].Subtitle != "Register (Toronto)" || h[0].Location.Title != "Register (Toronto)" {
					t.Fatalf("jones: %+v", h)
				}
			},
		},
		{
			name: "delete a graph",
			write: func(f *graphSearchFixture) {
				_, err := writes.Call(f.c, writes.Op{Action: "delete_evidence_graph", UserID: f.user}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
					ch, err := evidencegraphs.Delete(tx, f.user, f.smith.ID)
					return struct{}{}, ch, err
				})
				if err != nil {
					t.Fatal(err)
				}
			},
			check: func(t *testing.T, f *graphSearchFixture) {
				// "hou" still prefixes Jones household, so look at the one graph.
				if h, ok := f.graphHit("smith hou", f.smith); ok {
					t.Fatalf("smith hou: %+v", h)
				}
			},
		},
		{
			name: "delete a Source with no Artifact",
			write: func(f *graphSearchFixture) {
				typeID := seedType(t, f.c, "Book", "")
				empty, _, err := writes.Run(f.c, writes.Op{Action: "create_source", UserID: f.user}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
					return sources.Create(tx, f.user, sources.CreateInput{SourceTypeID: typeID, Title: "Empty deed"})
				})
				if err != nil {
					t.Fatal(err)
				}
				g := evidencegraphstest.SoleGraph(t, f.c, empty.ID)
				if g.Ref == "" {
					t.Fatal("Default graph has no ref")
				}
				if h := f.graphHits(g.Ref); len(h) != 1 {
					t.Fatalf("before delete: %+v", h)
				}
				if _, err := writes.Call(f.c, writes.Op{Action: "delete_source", UserID: f.user}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
					ch, err := sources.Delete(tx, f.user, empty.ID)
					return struct{}{}, ch, err
				}); err != nil {
					t.Fatal(err)
				}
				f.smith.Ref = g.Ref
			},
			check: func(t *testing.T, f *graphSearchFixture) {
				if h := f.graphHits(f.smith.Ref); len(h) != 0 {
					t.Fatalf("after delete: %+v", h)
				}
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := newGraphSearchFixture(t)
			tc.write(f)
			tc.check(t, f)
		})
	}
}

func TestRebuildIndexesEvidenceGraphs(t *testing.T) {
	f := newGraphSearchFixture(t)
	db, err := f.c.DB()
	if err != nil {
		t.Fatal(err)
	}
	if err := searchindex.ClearAll(db); err != nil {
		t.Fatal(err)
	}
	if err := searchindex.SetProjectionVersion(db, 8); err != nil {
		t.Fatal(err)
	}
	if h := f.graphHits("jones hou"); len(h) != 0 {
		t.Fatalf("before rebuild: %+v", h)
	}
	if err := searchindex.EnsureCatalog(f.c); err != nil {
		t.Fatal(err)
	}
	h, ok := f.graphHit("jones hou", f.jones)
	if !ok || h.Location.SourceID != uuidString(f.source) {
		t.Fatalf("after rebuild: %+v", h)
	}
	if v, err := searchindex.ProjectionVersionStored(db); err != nil || v != 9 {
		t.Fatalf("stored version %d, %v", v, err)
	}
}

func TestEvidenceGraphHitLocationMatchesDeleteImpact(t *testing.T) {
	f := newGraphSearchFixture(t)
	hit, ok := f.graphHit("jones hou", f.jones)
	if !ok {
		t.Fatal("no Jones hit")
	}
	db, err := f.c.DB()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	g, err := deleteimpact.ListedGroup(tx, "x.via", catalogmodel.KindEvidenceGraph, []deleteimpact.Listed{{ID: f.jones.ID, Ref: f.jones.Ref}})
	if err != nil {
		t.Fatal(err)
	}
	if len(g.Listed) != 1 {
		t.Fatalf("group %+v", g)
	}
	got, want := hit.Location, g.Listed[0].Location
	pairs := []struct{ name, a, b string }{
		{"Section", got.Section, want.Section},
		{"SourceID", got.SourceID, want.SourceID},
		{"GraphID", got.GraphID, want.GraphID},
		{"SourceSurface", got.SourceSurface, want.SourceSurface},
		{"Ref", got.Ref, want.Ref},
		{"Title", got.Title, want.Title},
		{"SourceTitle", got.SourceTitle, want.SourceTitle},
	}
	for _, p := range pairs {
		if p.a != p.b {
			t.Errorf("%s: search %q, deleteimpact %q", p.name, p.a, p.b)
		}
	}
}
