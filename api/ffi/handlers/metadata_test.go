package handlers

import (
	"errors"
	"slices"
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/catalogsession"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/dbtest"
	"github.com/mendahu/provenencia/core/database/sourcemetadata"
	"google.golang.org/protobuf/proto"
)

// metadataOwners is the graphFixture's Source with its two metadata fields
// suggested for its type at the artifact and graph levels, and the owners
// whose boards the cases edit.
type metadataOwners struct {
	graphFixture
	suggested, extra string // field ids: suggested at artifact and graph, not suggested
	artifact, graph  *engine.MetadataOwner
}

func newMetadataOwners(t *testing.T) metadataOwners {
	t.Helper()
	f := newGraphFixture(t)
	var ws engine.GetSourceWorkspaceResponse
	mustRPC(t, GetSourceWorkspace, &engine.GetSourceWorkspaceRequest{ProjectDir: f.dir, SourceId: f.sourceID}, &ws)
	sourceType := ws.GetSource().GetSourceTypeId()
	m := metadataOwners{
		graphFixture: f,
		suggested:    newUserField(t, f.dir, f.userID, "Provider"),
		extra:        newUserField(t, f.dir, f.userID, "Line"),
		artifact:     &engine.MetadataOwner{Level: "artifact", Id: f.page1},
		graph:        &engine.MetadataOwner{Level: "graph", Id: f.defaultGraphID},
	}
	for _, level := range []string{"artifact", "graph"} {
		mustRPC(t, AssignTypeField, &engine.AssignTypeFieldRequest{
			ProjectDir: f.dir, TypeId: sourceType, FieldId: m.suggested, Level: level,
		}, &engine.AssignTypeFieldResponse{})
	}
	return m
}

func TestMetadataOwners(t *testing.T) {
	m := newMetadataOwners(t)
	for name, owner := range map[string]*engine.MetadataOwner{"artifact": m.artifact, "graph": m.graph} {
		t.Run(name, func(t *testing.T) {
			var set engine.SetMetadataResponse
			mustRPC(t, SetMetadata, &engine.SetMetadataRequest{
				ProjectDir: m.dir, UserId: m.userID, Owner: owner, FieldId: m.suggested, ValueText: "FamilySearch",
			}, &set)
			if got := metadataByField(set.GetMetadata(), m.suggested); got == nil || got.GetValueText() != "FamilySearch" || !got.GetSuggested() {
				t.Fatalf("set metadata %+v", set.GetMetadata())
			}
			if !slices.Contains(set.GetEffects().GetSourceScopes(), m.sourceID) {
				t.Fatalf("source_scopes %v, want %s", set.GetEffects().GetSourceScopes(), m.sourceID)
			}

			var dismissed engine.DismissMetadataSuggestionResponse
			var setExtra engine.SetMetadataResponse
			mustRPC(t, SetMetadata, &engine.SetMetadataRequest{
				ProjectDir: m.dir, UserId: m.userID, Owner: owner, FieldId: m.extra, ValueText: "14",
			}, &setExtra)
			if ids := metadataFieldIDs(setExtra.GetMetadata()); !slices.Equal(ids, []string{m.suggested, m.extra}) {
				t.Fatalf("set extra order %v, want suggested then extra", ids)
			}
			var reordered engine.ReorderMetadataResponse
			mustRPC(t, ReorderMetadata, &engine.ReorderMetadataRequest{
				ProjectDir: m.dir, UserId: m.userID, Owner: owner, FieldIds: []string{m.extra, m.suggested},
			}, &reordered)
			if len(reordered.GetMetadata()) != 2 || reordered.GetMetadata()[0].GetField().GetId() != m.extra ||
				!slices.Contains(reordered.GetEffects().GetSourceScopes(), m.sourceID) {
				t.Fatalf("reordered %+v", &reordered)
			}

			var clearedSuggested engine.ClearMetadataResponse
			mustRPC(t, ClearMetadata, &engine.ClearMetadataRequest{
				ProjectDir: m.dir, UserId: m.userID, Owner: owner, FieldId: m.suggested,
			}, &clearedSuggested)
			if got := metadataByField(clearedSuggested.GetMetadata(), m.suggested); got == nil || got.GetHasValue() || !got.GetSuggested() {
				t.Fatalf("cleared suggested field %+v, want an empty suggestion", clearedSuggested.GetMetadata())
			}
			var cleared engine.ClearMetadataResponse
			mustRPC(t, ClearMetadata, &engine.ClearMetadataRequest{
				ProjectDir: m.dir, UserId: m.userID, Owner: owner, FieldId: m.extra,
			}, &cleared)
			if metadataByField(cleared.GetMetadata(), m.extra) != nil {
				t.Fatalf("cleared extra field still listed: %+v", cleared.GetMetadata())
			}
			if !slices.Contains(cleared.GetEffects().GetSourceScopes(), m.sourceID) {
				t.Fatalf("clear source_scopes %v", cleared.GetEffects().GetSourceScopes())
			}

			mustRPC(t, DismissMetadataSuggestion, &engine.DismissMetadataSuggestionRequest{
				ProjectDir: m.dir, UserId: m.userID, Owner: owner, FieldId: m.suggested,
			}, &dismissed)
			for _, e := range dismissed.GetMetadata() {
				if e.GetField().GetId() == m.suggested {
					t.Fatalf("dismissed suggestion still listed: %+v", e)
				}
			}
			if !slices.Contains(dismissed.GetEffects().GetSourceScopes(), m.sourceID) {
				t.Fatalf("dismiss source_scopes %v", dismissed.GetEffects().GetSourceScopes())
			}
		})
	}

	t.Run("a refused url value names its field", func(t *testing.T) {
		link := newField(t, m.dir, m.userID, "Link", "url")
		err := rpc(t, SetMetadata, &engine.SetMetadataRequest{
			ProjectDir: m.dir, UserId: m.userID, Owner: m.artifact, FieldId: link, ValueText: "not a link",
		}, &engine.SetMetadataResponse{})
		if !errors.Is(err, sourcemetadata.ErrInvalid) || !slices.Equal(errParams(t, err), []string{link}) {
			t.Fatalf("got %v, params %v, want params [%s]", err, errParams(t, err), link)
		}
	})

	// Owners the wire refuses: an unwired level and a missing owner are
	// sourcemetadata.invalid; a malformed id is request.invalid_id.
	bad := map[string]struct {
		owner *engine.MetadataOwner
		want  error
	}{
		"citation level": {&engine.MetadataOwner{Level: "citation", Id: m.page1}, sourcemetadata.ErrInvalid},
		"empty level":    {&engine.MetadataOwner{Level: "", Id: m.page1}, sourcemetadata.ErrInvalid},
		"missing owner":  {nil, sourcemetadata.ErrInvalid},
		"unknown artifact": {&engine.MetadataOwner{Level: "artifact", Id: "01890000-0000-7000-8000-000000000001"},
			sourcemetadata.ErrInvalid},
		"malformed id": {&engine.MetadataOwner{Level: "artifact", Id: "nope"}, errInvalidID},
	}
	for name, tt := range bad {
		t.Run(name, func(t *testing.T) {
			err := rpc(t, SetMetadata, &engine.SetMetadataRequest{
				ProjectDir: m.dir, UserId: m.userID, Owner: tt.owner, FieldId: m.suggested, ValueText: "x",
			}, &engine.SetMetadataResponse{})
			if !errors.Is(err, tt.want) {
				t.Fatalf("got %v, want %v", err, tt.want)
			}
		})
	}
}

// metadataByField is the entry for a field id, or nil.
func metadataByField(entries []*engine.MetadataWorkspaceEntry, fieldID string) *engine.MetadataWorkspaceEntry {
	for _, e := range entries {
		if e.GetField().GetId() == fieldID {
			return e
		}
	}
	return nil
}

// metadataFieldIDs lists the field ids of a board in its order.
func metadataFieldIDs(entries []*engine.MetadataWorkspaceEntry) []string {
	ids := make([]string, len(entries))
	for i, e := range entries {
		ids[i] = e.GetField().GetId()
	}
	return ids
}

func TestGetMetadataBoard(t *testing.T) {
	m := newMetadataOwners(t)
	mustRPC(t, SetMetadata, &engine.SetMetadataRequest{
		ProjectDir: m.dir, UserId: m.userID, Owner: m.graph, FieldId: m.extra, ValueText: "14",
	}, &engine.SetMetadataResponse{})
	var got engine.GetMetadataBoardResponse
	mustRPC(t, GetMetadataBoard, &engine.GetMetadataBoardRequest{ProjectDir: m.dir, Owner: m.graph}, &got)
	board := got.GetBoard()
	if board.GetOwnerId() != m.defaultGraphID || len(board.GetEntries()) != 2 {
		t.Fatalf("board %+v", board)
	}
	byField := map[string]*engine.MetadataWorkspaceEntry{}
	for _, e := range board.GetEntries() {
		byField[e.GetField().GetId()] = e
	}
	if e := byField[m.suggested]; e == nil || !e.GetSuggested() || e.GetHasValue() {
		t.Errorf("suggested entry %+v", e)
	}
	if e := byField[m.extra]; e == nil || e.GetSuggested() || e.GetValueText() != "14" {
		t.Errorf("extra entry %+v", e)
	}

	if err := rpc(t, GetMetadataBoard, &engine.GetMetadataBoardRequest{ProjectDir: m.dir}, &got); !errors.Is(err, sourcemetadata.ErrInvalid) {
		t.Errorf("no owner: %v", err)
	}
}

// latestRevisionTypes lists the entity type and action of every change in the
// project's latest revision, in write order.
func latestRevisionTypes(t *testing.T, dir string) []string {
	t.Helper()
	var out []string
	err := catalogsession.Do(dir, func(c *database.Catalog) error {
		_, changes := dbtest.LatestRevision(t, dbtest.MustDB(t, c))
		for _, ch := range changes {
			out = append(out, ch.EntityType+" "+ch.Action)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// errParams is the params of the *apperr.Error in err's chain.
func errParams(t *testing.T, err error) []string {
	t.Helper()
	var ae *apperr.Error
	if !errors.As(err, &ae) {
		t.Fatalf("not an apperr: %v", err)
	}
	return ae.Params()
}

// boardEntries is the owner's board entries as GetMetadataBoard reads them.
func boardEntries(t *testing.T, dir string, owner *engine.MetadataOwner) []*engine.MetadataWorkspaceEntry {
	t.Helper()
	var board engine.GetMetadataBoardResponse
	mustRPC(t, GetMetadataBoard, &engine.GetMetadataBoardRequest{ProjectDir: dir, Owner: owner}, &board)
	return board.GetBoard().GetEntries()
}

// assertReplyBoard fails unless got is the owner's board as GetMetadataBoard
// reads it now, entry for entry.
func assertReplyBoard(t *testing.T, dir string, owner *engine.MetadataOwner, got []*engine.MetadataWorkspaceEntry) {
	t.Helper()
	want := boardEntries(t, dir, owner)
	if len(got) != len(want) {
		t.Fatalf("reply board has %d entries, want %d: %+v vs %+v", len(got), len(want), got, want)
	}
	for i := range want {
		if !proto.Equal(got[i], want[i]) {
			t.Fatalf("reply entry %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

// draftOf is a draft with one value and its layout row at the top.
func draftOf(fieldID, text string) *engine.MetadataDraft {
	return &engine.MetadataDraft{
		Values: []*engine.MetadataValue{{FieldId: fieldID, ValueText: text}},
		Layout: []*engine.MetadataLayoutRow{{FieldId: fieldID, SortOrder: 0}},
	}
}

// TestMetadataBoardsKeepTheirLevel suggests one field at the artifact level
// only and one at the graph level only: each board lists its own level's
// suggestions and not the other's, on the board read and on the Source page.
func TestMetadataBoardsKeepTheirLevel(t *testing.T) {
	m := newMetadataOwners(t)
	artifactOnly := newUserField(t, m.dir, m.userID, "Provider URL")
	graphOnly := newUserField(t, m.dir, m.userID, "Tree")
	var ws engine.GetSourceWorkspaceResponse
	mustRPC(t, GetSourceWorkspace, &engine.GetSourceWorkspaceRequest{ProjectDir: m.dir, SourceId: m.sourceID}, &ws)
	for field, level := range map[string]string{artifactOnly: "artifact", graphOnly: "graph"} {
		mustRPC(t, AssignTypeField, &engine.AssignTypeFieldRequest{
			ProjectDir: m.dir, TypeId: ws.GetSource().GetSourceTypeId(), FieldId: field, Level: level,
		}, &engine.AssignTypeFieldResponse{})
	}
	ids := func(entries []*engine.MetadataWorkspaceEntry) []string {
		var out []string
		for _, e := range entries {
			out = append(out, e.GetField().GetId())
		}
		slices.Sort(out)
		return out
	}
	sorted := func(ids ...string) []string { slices.Sort(ids); return ids }
	for name, tt := range map[string]struct {
		owner *engine.MetadataOwner
		want  []string
	}{
		"artifact": {m.artifact, sorted(m.suggested, artifactOnly)},
		"graph":    {m.graph, sorted(m.suggested, graphOnly)},
	} {
		t.Run(name, func(t *testing.T) {
			if got := ids(boardEntries(t, m.dir, tt.owner)); !slices.Equal(got, tt.want) {
				t.Fatalf("board fields %v, want %v", got, tt.want)
			}
		})
	}
	// The Source page's Artifact boards carry the artifact-level fields only.
	var after engine.GetSourceWorkspaceResponse
	mustRPC(t, GetSourceWorkspace, &engine.GetSourceWorkspaceRequest{ProjectDir: m.dir, SourceId: m.sourceID}, &after)
	for _, board := range after.GetArtifactMetadata() {
		if got, want := ids(board.GetEntries()), sorted(m.suggested, artifactOnly); !slices.Equal(got, want) {
			t.Fatalf("Artifact board fields %v, want %v", got, want)
		}
	}
}
