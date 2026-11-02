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
		if len(res.Written) != 1 || !strings.HasPrefix(res.Written[0].Entity.GetRef(), "PER-") {
			t.Fatalf("%+v", res.Written)
		}
		if res.GetRevision() <= proposal.GetRevision() {
			t.Fatalf("revision %d, proposal %d", res.GetRevision(), proposal.GetRevision())
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
		if len(res.GetWritten()) != 1 || res.GetWritten()[0].GetEntity().GetId() != f.handle {
			t.Fatalf("written = %+v, want one row on %s", res.GetWritten(), f.handle)
		}
	})
}
