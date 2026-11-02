package handlers

import (
	"strings"
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
	"google.golang.org/protobuf/proto"
)

func TestApplyPromoteGraphAlignment(t *testing.T) {
	runRPC(t, ApplyPromoteGraphAlignment, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "bad graph id",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _, _ := subjectFixture(t)
				return &engine.ApplyPromoteGraphAlignmentRequest{
					ProjectDir: dir, UserId: userID, GraphId: "nope",
				}
			},
			wantErr: true,
		},
	})

	t.Run("mints a person in one revision", func(t *testing.T) {
		dir, userID, sourceID, typeID := subjectFixture(t)
		out, err := CreateSubject(marshalProto(t, &engine.CreateSubjectRequest{
			ProjectDir: dir, UserId: userID, GraphId: soleGraphID(t, dir, sourceID), SubjectTypeId: typeID, Label: "Ada",
		}))
		if err != nil {
			t.Fatal(err)
		}
		var created engine.CreateSubjectResponse
		if err := proto.Unmarshal(out, &created); err != nil {
			t.Fatal(err)
		}
		proposed, err := ProposePromoteGraphAlignment(marshalProto(t, &engine.ProposePromoteGraphAlignmentRequest{
			ProjectDir: dir, GraphId: soleGraphID(t, dir, sourceID),
		}))
		if err != nil {
			t.Fatal(err)
		}
		var proposal engine.ProposePromoteGraphAlignmentResponse
		if err := proto.Unmarshal(proposed, &proposal); err != nil {
			t.Fatal(err)
		}
		if proposal.GetRevision() == 0 {
			t.Fatal("proposal revision")
		}
		applied, err := ApplyPromoteGraphAlignment(marshalProto(t, &engine.ApplyPromoteGraphAlignmentRequest{
			ProjectDir: dir, UserId: userID, GraphId: soleGraphID(t, dir, sourceID), SeenRevision: proposal.GetRevision(),
			Rows: []*engine.ApplyPromoteGraphAlignmentRow{{
				SubjectId: created.Subject.GetId(), Target: "new",
			}},
		}))
		if err != nil {
			t.Fatal(err)
		}
		var res engine.ApplyPromoteGraphAlignmentResponse
		if err := proto.Unmarshal(applied, &res); err != nil {
			t.Fatal(err)
		}
		eff := res.GetEffects()
		if eff.GetRevision() <= proposal.GetRevision() {
			t.Fatalf("revision %d, proposal %d", eff.GetRevision(), proposal.GetRevision())
		}
		if len(eff.GetHandles()) != 1 {
			t.Fatalf("handles = %v, want the one new handle", eff.GetHandles())
		}
		var detail engine.ConclusionDetail
		mustRPC(t, GetConclusionDetail, &engine.GetConclusionDetailRequest{ProjectDir: dir, EntityId: eff.GetHandles()[0]}, &detail)
		if !strings.HasPrefix(detail.GetEntity().GetRef(), "PER-") {
			t.Fatalf("ref = %q", detail.GetEntity().GetRef())
		}
		graphID := soleGraphID(t, dir, sourceID)
		if !contains(eff.GetEvidenceGraphs(), graphID) {
			t.Fatalf("evidence_graphs = %v, want %s", eff.GetEvidenceGraphs(), graphID)
		}
		if !readHas(eff, "evidence_graph_progress", graphID) {
			t.Fatalf("reads = %v, want progress for %s", eff.GetReads(), graphID)
		}
		if !contains(eff.GetLists(), "counts") {
			t.Fatalf("lists = %v, want counts", eff.GetLists())
		}
	})

	t.Run("files a found record", func(t *testing.T) {
		f := newFoundRecordFixture(t)
		var proposal engine.ProposePromoteGraphAlignmentResponse
		mustRPC(t, ProposePromoteGraphAlignment, &engine.ProposePromoteGraphAlignmentRequest{
			ProjectDir: f.dir, GraphId: f.graphID,
			Fixed: []*engine.PromoteGraphAlignmentFixed{{SubjectId: f.incoming, HandleId: f.handle}},
		}, &proposal)
		var res engine.ApplyPromoteGraphAlignmentResponse
		err := rpc(t, ApplyPromoteGraphAlignment, &engine.ApplyPromoteGraphAlignmentRequest{
			ProjectDir: f.dir, UserId: f.userID, GraphId: f.graphID, SeenRevision: proposal.GetRevision(),
			Rows: []*engine.ApplyPromoteGraphAlignmentRow{{SubjectId: f.incoming, Target: "handle", EntityId: f.handle}},
		}, &res)
		if err != nil {
			t.Fatal(err)
		}
		if !contains(res.GetEffects().GetHandles(), f.handle) {
			t.Fatalf("handles = %v, want %s", res.GetEffects().GetHandles(), f.handle)
		}
		if contains(res.GetEffects().GetLists(), "counts") {
			t.Fatalf("a join reported counts: %v", res.GetEffects().GetLists())
		}
	})
}

// A Done whose connection has both ends filed mints an association handle
// beside the two it files, and reports all three.
func TestApplyPromoteGraphAlignmentReportsABridgeMint(t *testing.T) {
	dir, userID, sourceID, artifactID, personID, eventID, personProp, eventProp, roleProp, roleTerm := connectFixture(t)
	var bridge engine.CreateCitedBridgeResponse
	mustRPC(t, CreateCitedBridge, &engine.CreateCitedBridgeRequest{
		ProjectDir: dir, UserId: userID, FromSubjectId: personID, ToSubjectId: eventID,
		BridgeTypeKey: "participation", ArtifactId: artifactID, LocatorJson: validLocatorJSON,
		Observations: []*engine.ObservationDraft{
			{PropertyId: personProp, ValueSubjectId: personID},
			{PropertyId: eventProp, ValueSubjectId: eventID},
			{PropertyId: roleProp, ValueTermId: roleTerm},
		},
	}, &bridge)
	graphID := soleGraphID(t, dir, sourceID)
	var proposal engine.ProposePromoteGraphAlignmentResponse
	mustRPC(t, ProposePromoteGraphAlignment, &engine.ProposePromoteGraphAlignmentRequest{ProjectDir: dir, GraphId: graphID}, &proposal)
	var res engine.ApplyPromoteGraphAlignmentResponse
	mustRPC(t, ApplyPromoteGraphAlignment, &engine.ApplyPromoteGraphAlignmentRequest{
		ProjectDir: dir, UserId: userID, GraphId: graphID, SeenRevision: proposal.GetRevision(),
		Rows: []*engine.ApplyPromoteGraphAlignmentRow{
			{SubjectId: personID, Target: "new"},
			{SubjectId: eventID, Target: "new"},
		},
	}, &res)
	if !contains(res.GetEffects().GetLists(), "counts") {
		t.Fatalf("lists = %v, want counts", res.GetEffects().GetLists())
	}
	// The person, the event and the association their connection minted.
	if got := len(res.GetEffects().GetHandles()); got != 3 {
		t.Fatalf("handles = %v, want the person, the event and the association", res.GetEffects().GetHandles())
	}
}

// readHas reports whether the effects name a keyed read for the id.
func readHas(eff *engine.WriteEffects, read, id string) bool {
	for _, r := range eff.GetReads() {
		if r.GetRead() == read && contains(r.GetIds(), id) {
			return true
		}
	}
	return false
}
