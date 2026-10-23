package handlers

import (
	"strings"
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database/observations"
	"google.golang.org/protobuf/proto"
)

func TestApplyPromoteGraphAlignment(t *testing.T) {
	runRPC(t, ApplyPromoteGraphAlignment, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "bad source id",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _, _ := subjectFixture(t)
				return &engine.ApplyPromoteGraphAlignmentRequest{
					ProjectDir: dir, UserId: userID, SourceId: "nope",
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
			ProjectDir: dir, SourceId: sourceID,
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
			ProjectDir: dir, UserId: userID, SourceId: sourceID, SeenRevision: proposal.GetRevision(),
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
		if len(res.Written) != 1 || !strings.HasPrefix(res.Written[0].Entity.GetRef(), "PER-") {
			t.Fatalf("%+v", res.Written)
		}
		if res.GetRevision() <= proposal.GetRevision() {
			t.Fatalf("revision %d, proposal %d", res.GetRevision(), proposal.GetRevision())
		}
	})

	// T-8: Done pins what the evidence offers, whether or not the sheet was opened.
	t.Run("pins the evidence's offers", func(t *testing.T) {
		apply := func(t *testing.T, f evidenceFixture, row *engine.ApplyPromoteGraphAlignmentRow) *engine.ApplyPromoteGraphAlignmentResponse {
			t.Helper()
			var proposal engine.ProposePromoteGraphAlignmentResponse
			mustRPC(t, ProposePromoteGraphAlignment, &engine.ProposePromoteGraphAlignmentRequest{ProjectDir: f.dir, SourceId: f.sourceID}, &proposal)
			row.SubjectId, row.Target, row.EntityId = f.incoming, "handle", f.handle
			var res engine.ApplyPromoteGraphAlignmentResponse
			err := rpc(t, ApplyPromoteGraphAlignment, &engine.ApplyPromoteGraphAlignmentRequest{
				ProjectDir: f.dir, UserId: f.userID, SourceId: f.sourceID, SeenRevision: proposal.GetRevision(),
				Rows: []*engine.ApplyPromoteGraphAlignmentRow{row},
			}, &res)
			if err != nil {
				t.Fatal(err)
			}
			return &res
		}
		// requireSexPins: the incoming Observation and each member's own are on
		// the new claim, and each member's claim gained the incoming one.
		requireSexPins := func(t *testing.T, f evidenceFixture, line *engine.PromoteClaimEvidenceLine, res *engine.ApplyPromoteGraphAlignmentResponse) {
			t.Helper()
			created := pinnedOn(t, f.dir, res.GetWritten()[0].GetClaim().GetId())
			if !created[line.GetIncomingObservationId()] {
				t.Errorf("the new claim lacks the incoming Observation")
			}
			for i, p := range line.GetMembers() {
				if !created[memberObservation(p)] {
					t.Errorf("the new claim lacks member %d's Observation", i)
				}
				member := pinnedOn(t, f.dir, f.claims[i])
				if !member[line.GetIncomingObservationId()] || !member[memberObservation(p)] {
					t.Errorf("member %d's claim pins %v", i, member)
				}
			}
		}

		t.Run("a row never opened pins its defaults", func(t *testing.T) {
			f := newEvidenceFixture(t, observations.PolarityPositive)
			line := sectionOf(f.evidence(t), "sex_at_birth").GetLines()[0]
			requireSexPins(t, f, line, apply(t, f, &engine.ApplyPromoteGraphAlignmentRow{DefaultPins: true}))
		})
		t.Run("a row that kept the offered ids pins the same", func(t *testing.T) {
			f := newEvidenceFixture(t, observations.PolarityPositive)
			before := f.evidence(t)
			var ids []string
			for _, o := range before.GetOffers() {
				ids = append(ids, o.GetId())
			}
			requireSexPins(t, f, sectionOf(before, "sex_at_birth").GetLines()[0], apply(t, f, &engine.ApplyPromoteGraphAlignmentRow{PinIds: ids}))
		})
		t.Run("default pins with ids is refused", func(t *testing.T) {
			f := newEvidenceFixture(t, observations.PolarityPositive)
			ids := []string{f.evidence(t).GetOffers()[0].GetId()}
			var proposal engine.ProposePromoteGraphAlignmentResponse
			mustRPC(t, ProposePromoteGraphAlignment, &engine.ProposePromoteGraphAlignmentRequest{ProjectDir: f.dir, SourceId: f.sourceID}, &proposal)
			err := rpc(t, ApplyPromoteGraphAlignment, &engine.ApplyPromoteGraphAlignmentRequest{
				ProjectDir: f.dir, UserId: f.userID, SourceId: f.sourceID, SeenRevision: proposal.GetRevision(),
				Rows: []*engine.ApplyPromoteGraphAlignmentRow{{SubjectId: f.incoming, Target: "handle", EntityId: f.handle, DefaultPins: true, PinIds: ids}},
			}, &engine.ApplyPromoteGraphAlignmentResponse{})
			if err == nil {
				t.Fatal("expected error")
			}
		})
		t.Run("a negative sex at birth pins nothing for the conflict", func(t *testing.T) {
			f := newEvidenceFixture(t, observations.PolarityNegative)
			line := sectionOf(f.evidence(t), "sex_at_birth").GetLines()[0]
			res := apply(t, f, &engine.ApplyPromoteGraphAlignmentRow{DefaultPins: true})
			created := pinnedOn(t, f.dir, res.GetWritten()[0].GetClaim().GetId())
			if created[line.GetIncomingObservationId()] {
				t.Errorf("the conflict's incoming Observation was pinned")
			}
			for _, p := range line.GetMembers() {
				if created[memberObservation(p)] {
					t.Errorf("a conflict member Observation was pinned")
				}
			}
		})
	})
}
