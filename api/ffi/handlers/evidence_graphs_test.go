package handlers

import (
	"errors"
	"slices"
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database/evidencegraphs"
)

// graphFixture is citationFixture plus a second Artifact on the same Source, a
// Citation on the first Artifact with one Observation on the place Subject, and
// the Source's Default graph.
type graphFixture struct {
	dir, userID, typeID                string
	sourceID, page1, page2             string
	subjectID, citationID, citationRef string
	defaultGraphID                     string
}

func newGraphFixture(t *testing.T) graphFixture {
	t.Helper()
	dir, userID, sourceID, page1, subjectID, toponymID := citationFixture(t)
	f := graphFixture{dir: dir, userID: userID, sourceID: sourceID, page1: page1, subjectID: subjectID}
	f.page2 = f.createArtifact(t, sourceID, "Page 2", nil)
	var cit engine.CreateCitationWithObservationsResponse
	mustRPC(t, CreateCitationWithObservations, &engine.CreateCitationWithObservationsRequest{
		ProjectDir: dir, UserId: userID, ArtifactId: page1, GraphId: graphLinking(t, dir, page1), LocatorJson: validLocatorJSON,
		Observations: []*engine.ObservationDraft{{SubjectId: subjectID, PropertyId: toponymID, ValueText: "Boston"}},
	}, &cit)
	f.citationID = cit.GetCitation().GetId()
	f.citationRef = cit.GetCitation().GetRef()
	graphs := f.graphsOf(t, sourceID)
	if len(graphs) != 1 {
		t.Fatalf("graphs %+v", graphs)
	}
	f.defaultGraphID = graphs[0].GetId()
	var types engine.ListSourceTypesResponse
	mustRPC(t, ListSourceTypes, &engine.ListSourceTypesRequest{ProjectDir: dir}, &types)
	f.typeID = types.Types[0].GetId()
	return f
}

// createArtifact sends links as given: nil is absent, the default (every graph).
func (f graphFixture) createArtifact(t *testing.T, sourceID, label string, links *engine.GraphLinks) string {
	t.Helper()
	return f.createArtifactResp(t, sourceID, label, links).GetArtifact().GetId()
}

func (f graphFixture) createArtifactResp(t *testing.T, sourceID, label string, links *engine.GraphLinks) *engine.CreateArtifactResponse {
	t.Helper()
	var art engine.CreateArtifactResponse
	mustRPC(t, CreateArtifact, &engine.CreateArtifactRequest{
		ProjectDir: f.dir, UserId: f.userID, SourceId: sourceID, Label: label, GraphLinks: links,
	}, &art)
	return &art
}

// otherSource makes a second Source with one Artifact and returns both ids.
func (f graphFixture) otherSource(t *testing.T) (sourceID, artifactID string) {
	t.Helper()
	var created engine.CreateSourceResponse
	mustRPC(t, CreateSource, &engine.CreateSourceRequest{
		ProjectDir: f.dir, UserId: f.userID, SourceTypeId: f.typeID, Title: "Parish register",
	}, &created)
	sourceID = created.GetSource().GetId()
	return sourceID, f.createArtifact(t, sourceID, "Page 9", nil)
}

func (f graphFixture) list(t *testing.T) []*engine.EvidenceGraph {
	t.Helper()
	var resp engine.ListEvidenceGraphsResponse
	mustRPC(t, ListEvidenceGraphs, &engine.ListEvidenceGraphsRequest{ProjectDir: f.dir}, &resp)
	return resp.GetGraphs()
}

func (f graphFixture) graphsOf(t *testing.T, sourceID string) []*engine.EvidenceGraph {
	t.Helper()
	var out []*engine.EvidenceGraph
	for _, g := range f.list(t) {
		if g.GetSourceId() == sourceID {
			out = append(out, g)
		}
	}
	return out
}

func (f graphFixture) graph(t *testing.T, id string) *engine.EvidenceGraph {
	t.Helper()
	for _, g := range f.list(t) {
		if g.GetId() == id {
			return g
		}
	}
	return nil
}

func (f graphFixture) createGraph(t *testing.T, name string, artifactIDs ...string) *engine.CreateEvidenceGraphResponse {
	t.Helper()
	var resp engine.CreateEvidenceGraphResponse
	mustRPC(t, CreateEvidenceGraph, &engine.CreateEvidenceGraphRequest{
		ProjectDir: f.dir, UserId: f.userID, SourceId: f.sourceID, Name: name, ArtifactIds: artifactIDs,
	}, &resp)
	return &resp
}

func assertGraphEffects(t *testing.T, effects *engine.WriteEffects) {
	t.Helper()
	if !slices.Contains(effects.GetLists(), "evidence_graphs") || effects.GetRevision() <= 0 {
		t.Fatalf("effects %+v", effects)
	}
}

func TestListEvidenceGraphs(t *testing.T) {
	t.Run("bad proto", func(t *testing.T) {
		if _, err := ListEvidenceGraphs([]byte{0xff}); err == nil {
			t.Fatal("expected error")
		}
	})
	f := newGraphFixture(t)
	f.createGraph(t, "Smith", f.page1)
	srcB, pageB := f.otherSource(t)

	graphs := f.list(t)
	var names []string
	for _, g := range graphs {
		names = append(names, g.GetName())
		if g.GetRef() == "" || g.GetSourceId() == "" || len(g.GetArtifactIds()) == 0 {
			t.Fatalf("graph %+v", g)
		}
	}
	if !slices.Equal(names, []string{"Default", "Default", "Smith"}) {
		t.Fatalf("names %v", names)
	}
	if got := f.graphsOf(t, srcB); len(got) != 1 || !slices.Equal(got[0].GetArtifactIds(), []string{pageB}) {
		t.Fatalf("source B graphs %+v", got)
	}
	a := f.graphsOf(t, f.sourceID)
	if len(a) != 2 || a[0].GetName() != "Default" || a[1].GetName() != "Smith" ||
		!slices.Equal(a[1].GetArtifactIds(), []string{f.page1}) || len(a[0].GetArtifactIds()) != 2 {
		t.Fatalf("source A graphs %+v", a)
	}
}

func TestCreateEvidenceGraph(t *testing.T) {
	f := newGraphFixture(t)
	t.Run("creates and lists a graph", func(t *testing.T) {
		var resp engine.CreateEvidenceGraphResponse
		mustRPC(t, CreateEvidenceGraph, &engine.CreateEvidenceGraphRequest{
			ProjectDir: f.dir, UserId: f.userID, SourceId: f.sourceID,
			Name: " Jones ", Description: "Household 12", ArtifactIds: []string{f.page1},
		}, &resp)
		g := resp.GetGraph()
		if len(g.GetRef()) < 4 || g.GetRef()[:4] != "GPH-" || g.GetName() != "Jones" ||
			g.GetDescription() != "Household 12" || !slices.Equal(g.GetArtifactIds(), []string{f.page1}) {
			t.Fatalf("graph %+v", g)
		}
		assertGraphEffects(t, resp.GetEffects())
		if listed := f.graph(t, g.GetId()); listed == nil || listed.GetName() != "Jones" {
			t.Fatalf("listed %+v", listed)
		}
	})
	_, foreign := f.otherSource(t)
	before := len(f.list(t))
	for name, ids := range map[string][]string{"no artifacts": nil, "another source's artifact": {foreign}} {
		t.Run(name, func(t *testing.T) {
			err := rpc(t, CreateEvidenceGraph, &engine.CreateEvidenceGraphRequest{
				ProjectDir: f.dir, UserId: f.userID, SourceId: f.sourceID, Name: "Bad", ArtifactIds: ids,
			}, &engine.CreateEvidenceGraphResponse{})
			if !errors.Is(err, evidencegraphs.ErrInvalid) {
				t.Fatalf("got %v", err)
			}
			if len(f.list(t)) != before {
				t.Fatal("a graph was added")
			}
		})
	}
	t.Run("bad user id", func(t *testing.T) {
		err := rpc(t, CreateEvidenceGraph, &engine.CreateEvidenceGraphRequest{
			ProjectDir: f.dir, UserId: "nope", SourceId: f.sourceID, Name: "X", ArtifactIds: []string{f.page1},
		}, &engine.CreateEvidenceGraphResponse{})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestUpdateEvidenceGraph(t *testing.T) {
	f := newGraphFixture(t)
	created := f.createGraph(t, "Jones", f.page1)
	id := created.GetGraph().GetId()

	t.Run("renames and relinks", func(t *testing.T) {
		var resp engine.UpdateEvidenceGraphResponse
		mustRPC(t, UpdateEvidenceGraph, &engine.UpdateEvidenceGraphRequest{
			ProjectDir: f.dir, UserId: f.userID, GraphId: id, Name: "Jones family", ArtifactIds: []string{f.page2},
		}, &resp)
		assertGraphEffects(t, resp.GetEffects())
		if resp.GetEffects().GetRevision() != created.GetEffects().GetRevision()+1 {
			t.Fatalf("revision %d after %d", resp.GetEffects().GetRevision(), created.GetEffects().GetRevision())
		}
		g := f.graph(t, id)
		if g.GetName() != "Jones family" || !slices.Equal(g.GetArtifactIds(), []string{f.page2}) {
			t.Fatalf("graph %+v", g)
		}
	})
	t.Run("a graph's Citations keep their Artifact linked", func(t *testing.T) {
		err := rpc(t, UpdateEvidenceGraph, &engine.UpdateEvidenceGraphRequest{
			ProjectDir: f.dir, UserId: f.userID, GraphId: f.defaultGraphID, Name: "Household", ArtifactIds: []string{f.page2},
		}, &engine.UpdateEvidenceGraphResponse{})
		if !errors.Is(err, evidencegraphs.ErrArtifactInUse) {
			t.Fatalf("got %v", err)
		}
		g := f.graph(t, f.defaultGraphID)
		if g.GetName() != "Default" || len(g.GetArtifactIds()) != 2 {
			t.Fatalf("graph changed: %+v", g)
		}
	})
}

func TestDeleteEvidenceGraph(t *testing.T) {
	del := func(f graphFixture, id string) (*engine.DeleteEvidenceGraphResponse, error) {
		var resp engine.DeleteEvidenceGraphResponse
		err := rpc(t, DeleteEvidenceGraph, &engine.DeleteEvidenceGraphRequest{ProjectDir: f.dir, UserId: f.userID, GraphId: id}, &resp)
		return &resp, err
	}
	t.Run("a Source's only graph stays", func(t *testing.T) {
		f := newGraphFixture(t)
		if _, err := del(f, f.defaultGraphID); !errors.Is(err, evidencegraphs.ErrLastGraph) {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("a graph with Subjects stays", func(t *testing.T) {
		f := newGraphFixture(t)
		f.createGraph(t, "Jones", f.page1)
		if _, err := del(f, f.defaultGraphID); !errors.Is(err, evidencegraphs.ErrInUse) {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("an empty graph is deleted", func(t *testing.T) {
		f := newGraphFixture(t)
		id := f.createGraph(t, "Jones", f.page1).GetGraph().GetId()
		resp, err := del(f, id)
		if err != nil {
			t.Fatal(err)
		}
		assertGraphEffects(t, resp.GetEffects())
		if f.graph(t, id) != nil {
			t.Fatal("graph still listed")
		}
	})
}

func TestGetEvidenceGraphUnlinkImpact(t *testing.T) {
	f := newGraphFixture(t)
	// page1 has a Citation with an Observation in Default (artifact_in_use).
	// solo's only Artifact is lone (last_artifact). needy has a Citation with no
	// Observation, made in Default (artifact_in_use).
	needy := f.createArtifact(t, f.sourceID, "Needy", &engine.GraphLinks{GraphIds: []string{f.defaultGraphID}})
	var needyCit engine.CreateCitationWithObservationsResponse
	mustRPC(t, CreateCitationWithObservations, &engine.CreateCitationWithObservationsRequest{
		ProjectDir: f.dir, UserId: f.userID, ArtifactId: needy, GraphId: graphLinking(t, f.dir, needy), LocatorJson: validLocatorJSON,
	}, &needyCit)
	lone := f.createArtifact(t, f.sourceID, "Lone", &engine.GraphLinks{GraphIds: []string{f.defaultGraphID}})
	solo := f.createGraph(t, "Solo", lone).GetGraph().GetId()
	pair := func(graph, artifact string) *engine.EvidenceGraphArtifactPair {
		return &engine.EvidenceGraphArtifactPair{GraphId: graph, ArtifactId: artifact}
	}

	var resp engine.GetEvidenceGraphUnlinkImpactResponse
	mustRPC(t, GetEvidenceGraphUnlinkImpact, &engine.GetEvidenceGraphUnlinkImpactRequest{
		ProjectDir: f.dir,
		Pairs: []*engine.EvidenceGraphArtifactPair{
			pair(f.defaultGraphID, f.page2), // allowed
			pair(f.defaultGraphID, f.page1),
			pair(solo, lone),
			pair(f.defaultGraphID, needy),
			pair(solo, f.page2), // not linked
		},
	}, &resp)
	if len(resp.GetBlocked()) != 3 {
		t.Fatalf("blocked %+v", resp.GetBlocked())
	}
	inUse, last, need := resp.GetBlocked()[0], resp.GetBlocked()[1], resp.GetBlocked()[2]

	g := inUse.GetCitations()
	if inUse.GetGraphId() != f.defaultGraphID || inUse.GetArtifactId() != f.page1 || inUse.GetReasonCode() != "evidencegraphs.artifact_in_use" ||
		g.GetVia() != evidencegraphs.UnlinkVia || g.GetKind() != "citation" || g.GetTotal() != 1 || len(g.GetListed()) != 1 {
		t.Fatalf("in use %+v", inUse)
	}
	if l := g.GetListed()[0]; l.GetId() != f.citationID || l.GetRef() != f.citationRef || l.GetLocation().GetSourceSurface() != "citationComposer" {
		t.Fatalf("listed %+v", l)
	}
	if last.GetGraphId() != solo || last.GetArtifactId() != lone || last.GetReasonCode() != "evidencegraphs.last_artifact" || last.GetCitations() != nil {
		t.Fatalf("last artifact %+v", last)
	}
	ng := need.GetCitations()
	if need.GetGraphId() != f.defaultGraphID || need.GetArtifactId() != needy || need.GetReasonCode() != "evidencegraphs.artifact_in_use" ||
		ng.GetVia() != evidencegraphs.UnlinkVia || ng.GetKind() != "citation" || ng.GetTotal() != 1 ||
		len(ng.GetListed()) != 1 || ng.GetListed()[0].GetId() != needyCit.GetCitation().GetId() {
		t.Fatalf("empty Citation in use %+v", need)
	}

	t.Run("bad proto", func(t *testing.T) {
		if _, err := GetEvidenceGraphUnlinkImpact([]byte{0xff}); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("malformed id", func(t *testing.T) {
		for name, p := range map[string]*engine.EvidenceGraphArtifactPair{
			"graph":    pair("nope", f.page1),
			"artifact": pair(f.defaultGraphID, "nope"),
		} {
			err := rpc(t, GetEvidenceGraphUnlinkImpact, &engine.GetEvidenceGraphUnlinkImpactRequest{
				ProjectDir: f.dir, Pairs: []*engine.EvidenceGraphArtifactPair{p},
			}, &engine.GetEvidenceGraphUnlinkImpactResponse{})
			if !errors.Is(err, errInvalidID) {
				t.Fatalf("%s: got %v", name, err)
			}
		}
	})
	t.Run("unknown graph", func(t *testing.T) {
		err := rpc(t, GetEvidenceGraphUnlinkImpact, &engine.GetEvidenceGraphUnlinkImpactRequest{
			ProjectDir: f.dir, Pairs: []*engine.EvidenceGraphArtifactPair{pair(f.page1, f.page1)},
		}, &engine.GetEvidenceGraphUnlinkImpactResponse{})
		if !errors.Is(err, evidencegraphs.ErrInvalid) {
			t.Fatalf("got %v", err)
		}
	})
}

// soleGraphID is the id of a Source's one Evidence graph, read through the
// engine as the app does; the test fails if it has none or several.
func soleGraphID(t *testing.T, dir, sourceID string) string {
	t.Helper()
	var resp engine.ListEvidenceGraphsResponse
	mustRPC(t, ListEvidenceGraphs, &engine.ListEvidenceGraphsRequest{ProjectDir: dir}, &resp)
	var ids []string
	for _, g := range resp.GetGraphs() {
		if g.GetSourceId() == sourceID {
			ids = append(ids, g.GetId())
		}
	}
	if len(ids) != 1 {
		t.Fatalf("Source %s has %d graphs, want 1", sourceID, len(ids))
	}
	return ids[0]
}

// graphLinking is the id of the Evidence graph an Artifact is linked to; the
// test fails unless exactly one graph links it.
func graphLinking(t *testing.T, dir, artifactID string) string {
	t.Helper()
	var resp engine.ListEvidenceGraphsResponse
	mustRPC(t, ListEvidenceGraphs, &engine.ListEvidenceGraphsRequest{ProjectDir: dir}, &resp)
	var ids []string
	for _, g := range resp.GetGraphs() {
		if slices.Contains(g.GetArtifactIds(), artifactID) {
			ids = append(ids, g.GetId())
		}
	}
	if len(ids) != 1 {
		t.Fatalf("Artifact %s is on %d graphs, want 1", artifactID, len(ids))
	}
	return ids[0]
}
