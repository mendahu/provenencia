package handlers

import (
	"errors"
	"slices"
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database/sourcemetadata"
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
			if set.GetEntry().GetField().GetId() != m.suggested || set.GetEntry().GetValueText() != "FamilySearch" || !set.GetEntry().GetSuggested() {
				t.Fatalf("entry %+v", set.GetEntry())
			}
			if !slices.Contains(set.GetEffects().GetSourceScopes(), m.sourceID) {
				t.Fatalf("source_scopes %v, want %s", set.GetEffects().GetSourceScopes(), m.sourceID)
			}

			var dismissed engine.DismissMetadataSuggestionResponse
			mustRPC(t, SetMetadata, &engine.SetMetadataRequest{
				ProjectDir: m.dir, UserId: m.userID, Owner: owner, FieldId: m.extra, ValueText: "14",
			}, &engine.SetMetadataResponse{})
			var reordered engine.ReorderMetadataResponse
			mustRPC(t, ReorderMetadata, &engine.ReorderMetadataRequest{
				ProjectDir: m.dir, UserId: m.userID, Owner: owner, FieldIds: []string{m.extra, m.suggested},
			}, &reordered)
			if len(reordered.GetMetadata()) != 2 || reordered.GetMetadata()[0].GetField().GetId() != m.extra ||
				!slices.Contains(reordered.GetEffects().GetSourceScopes(), m.sourceID) {
				t.Fatalf("reordered %+v", &reordered)
			}

			mustRPC(t, ClearMetadata, &engine.ClearMetadataRequest{
				ProjectDir: m.dir, UserId: m.userID, Owner: owner, FieldId: m.suggested,
			}, &engine.ClearMetadataResponse{})
			var cleared engine.ClearMetadataResponse
			mustRPC(t, ClearMetadata, &engine.ClearMetadataRequest{
				ProjectDir: m.dir, UserId: m.userID, Owner: owner, FieldId: m.extra,
			}, &cleared)
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
