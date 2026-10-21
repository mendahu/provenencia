package handlers

import (
	"errors"
	"slices"
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database/evidencegraphs"
	"google.golang.org/protobuf/proto"
)

// rpc calls fn with req and decodes the reply into resp.
func rpc(t *testing.T, fn func([]byte) ([]byte, error), req, resp proto.Message) error {
	t.Helper()
	out, err := fn(marshalProto(t, req))
	if err != nil {
		return err
	}
	if err := proto.Unmarshal(out, resp); err != nil {
		t.Fatal(err)
	}
	return nil
}

func mustRPC(t *testing.T, fn func([]byte) ([]byte, error), req, resp proto.Message) {
	t.Helper()
	if err := rpc(t, fn, req, resp); err != nil {
		t.Fatal(err)
	}
}

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
		ProjectDir: dir, UserId: userID, ArtifactId: page1, LocatorJson: validLocatorJSON,
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

func (f graphFixture) createArtifact(t *testing.T, sourceID, label string, graphIDs []string) string {
	t.Helper()
	return f.createArtifactResp(t, sourceID, label, graphIDs).GetArtifact().GetId()
}

func (f graphFixture) createArtifactResp(t *testing.T, sourceID, label string, graphIDs []string) *engine.CreateArtifactResponse {
	t.Helper()
	var art engine.CreateArtifactResponse
	mustRPC(t, CreateArtifact, &engine.CreateArtifactRequest{
		ProjectDir: f.dir, UserId: f.userID, SourceId: sourceID, Label: label, GraphIds: graphIDs,
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

func TestLinkUnlinkEvidenceGraphArtifact(t *testing.T) {
	f := newGraphFixture(t)
	id := f.createGraph(t, "Jones", f.page1).GetGraph().GetId()

	var linked engine.LinkEvidenceGraphArtifactResponse
	mustRPC(t, LinkEvidenceGraphArtifact, &engine.LinkEvidenceGraphArtifactRequest{
		ProjectDir: f.dir, UserId: f.userID, GraphId: id, ArtifactId: f.page2,
	}, &linked)
	assertGraphEffects(t, linked.GetEffects())
	if got := f.graph(t, id).GetArtifactIds(); len(got) != 2 || !slices.Contains(got, f.page2) {
		t.Fatalf("after link %v", got)
	}

	var unlinked engine.UnlinkEvidenceGraphArtifactResponse
	mustRPC(t, UnlinkEvidenceGraphArtifact, &engine.UnlinkEvidenceGraphArtifactRequest{
		ProjectDir: f.dir, UserId: f.userID, GraphId: id, ArtifactId: f.page2,
	}, &unlinked)
	assertGraphEffects(t, unlinked.GetEffects())
	if got := f.graph(t, id).GetArtifactIds(); !slices.Equal(got, []string{f.page1}) {
		t.Fatalf("after unlink %v", got)
	}

	err := rpc(t, UnlinkEvidenceGraphArtifact, &engine.UnlinkEvidenceGraphArtifactRequest{
		ProjectDir: f.dir, UserId: f.userID, GraphId: id, ArtifactId: f.page1,
	}, &engine.UnlinkEvidenceGraphArtifactResponse{})
	if !errors.Is(err, evidencegraphs.ErrLastArtifact) {
		t.Fatalf("last link: %v", err)
	}
}

func TestGetEvidenceGraphUnlinkImpact(t *testing.T) {
	f := newGraphFixture(t)
	var resp engine.GetEvidenceGraphUnlinkImpactResponse
	mustRPC(t, GetEvidenceGraphUnlinkImpact, &engine.GetEvidenceGraphUnlinkImpactRequest{
		ProjectDir: f.dir, GraphId: f.defaultGraphID, ArtifactIds: []string{f.page2, f.page1},
	}, &resp)
	if len(resp.GetBlocked()) != 1 {
		t.Fatalf("blocked %+v", resp.GetBlocked())
	}
	b := resp.GetBlocked()[0]
	g := b.GetCitations()
	if b.GetArtifactId() != f.page1 || g.GetVia() != evidencegraphs.UnlinkVia || g.GetKind() != "citation" ||
		g.GetTotal() != 1 || len(g.GetListed()) != 1 {
		t.Fatalf("block %+v", b)
	}
	if l := g.GetListed()[0]; l.GetId() != f.citationID || l.GetRef() != f.citationRef || l.GetLocation().GetSourceSurface() != "citationComposer" {
		t.Fatalf("listed %+v", l)
	}

	t.Run("malformed graph id", func(t *testing.T) {
		err := rpc(t, GetEvidenceGraphUnlinkImpact, &engine.GetEvidenceGraphUnlinkImpactRequest{
			ProjectDir: f.dir, GraphId: "nope", ArtifactIds: []string{f.page1},
		}, &engine.GetEvidenceGraphUnlinkImpactResponse{})
		if !errors.Is(err, errInvalidID) {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("unknown graph", func(t *testing.T) {
		err := rpc(t, GetEvidenceGraphUnlinkImpact, &engine.GetEvidenceGraphUnlinkImpactRequest{
			ProjectDir: f.dir, GraphId: f.page1, ArtifactIds: []string{f.page1},
		}, &engine.GetEvidenceGraphUnlinkImpactResponse{})
		if !errors.Is(err, evidencegraphs.ErrInvalid) {
			t.Fatalf("got %v", err)
		}
	})
}
