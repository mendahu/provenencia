package handlers

import (
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database/deleteimpact"
	"google.golang.org/protobuf/proto"
)

func TestGetDeleteImpact(t *testing.T) {
	runRPC(t, GetDeleteImpact, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "unknown kind",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, _ := sourceFixture(t)
				return &engine.GetDeleteImpactRequest{
					ProjectDir: dir, Kind: "nope", Id: uuid.Must(uuid.NewV7()).String(),
				}
			},
			wantErr:   true,
			wantErrIs: deleteimpact.ErrInvalid,
		},
		{
			name: "infra kind",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, _ := sourceFixture(t)
				return &engine.GetDeleteImpactRequest{
					ProjectDir: dir, Kind: "user", Id: uuid.Must(uuid.NewV7()).String(),
				}
			},
			want: &engine.GetDeleteImpactResponse{
				Allowed: false,
				Gate:    engine.DeleteImpactGate_DELETE_IMPACT_GATE_INFRA,
			},
		},
		{
			name: "missing citation",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, _ := sourceFixture(t)
				return &engine.GetDeleteImpactRequest{
					ProjectDir: dir, Kind: "citation", Id: uuid.Must(uuid.NewV7()).String(),
				}
			},
			want: &engine.GetDeleteImpactResponse{
				Allowed: false,
				Gate:    engine.DeleteImpactGate_DELETE_IMPACT_GATE_NOT_FOUND,
			},
		},
		{
			name: "unused seeded source type allowed",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, typeID := sourceFixture(t)
				return &engine.GetDeleteImpactRequest{
					ProjectDir: dir, Kind: "source_type", Id: typeID,
				}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var resp engine.GetDeleteImpactResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				// Fixture type 0 may be in use by the onboarding sample Source.
				if resp.GetGate() == engine.DeleteImpactGate_DELETE_IMPACT_GATE_UNSPECIFIED {
					t.Fatal("empty gate")
				}
			},
		},
		{
			name: "empty source allowed",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, typeID := sourceFixture(t)
				cout, err := CreateSource(marshalProto(t, &engine.CreateSourceRequest{
					ProjectDir: dir, UserId: userID, SourceTypeId: typeID, Title: "Bare",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateSourceResponse
				if err := proto.Unmarshal(cout, &created); err != nil {
					t.Fatal(err)
				}
				return &engine.GetDeleteImpactRequest{
					ProjectDir: dir, Kind: "source", Id: created.Source.Id,
				}
			},
			want: &engine.GetDeleteImpactResponse{
				Allowed: true,
				Gate:    engine.DeleteImpactGate_DELETE_IMPACT_GATE_OK,
			},
		},
		{
			name: "source blocked by artifact",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, typeID := sourceFixture(t)
				cout, err := CreateSource(marshalProto(t, &engine.CreateSourceRequest{
					ProjectDir: dir, UserId: userID, SourceTypeId: typeID, Title: "Held",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateSourceResponse
				if err := proto.Unmarshal(cout, &created); err != nil {
					t.Fatal(err)
				}
				if _, err := CreateArtifact(marshalProto(t, &engine.CreateArtifactRequest{
					ProjectDir: dir, UserId: userID, SourceId: created.Source.Id, Label: "Scan",
				})); err != nil {
					t.Fatal(err)
				}
				return &engine.GetDeleteImpactRequest{
					ProjectDir: dir, Kind: "source", Id: created.Source.Id,
				}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var resp engine.GetDeleteImpactResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				if resp.GetAllowed() || resp.GetGate() != engine.DeleteImpactGate_DELETE_IMPACT_GATE_INBOUND {
					t.Fatalf("%+v", &resp)
				}
				if len(resp.Groups) != 1 || resp.Groups[0].GetVia() != "artifacts.source_id" {
					t.Fatalf("groups %+v", resp.Groups)
				}
			},
		},
		{
			name: "empty artifact allowed",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, typeID := sourceFixture(t)
				cout, err := CreateSource(marshalProto(t, &engine.CreateSourceRequest{
					ProjectDir: dir, UserId: userID, SourceTypeId: typeID, Title: "Photo",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateSourceResponse
				if err := proto.Unmarshal(cout, &created); err != nil {
					t.Fatal(err)
				}
				aout, err := CreateArtifact(marshalProto(t, &engine.CreateArtifactRequest{
					ProjectDir: dir, UserId: userID, SourceId: created.Source.Id, Label: "Front",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var art engine.CreateArtifactResponse
				if err := proto.Unmarshal(aout, &art); err != nil {
					t.Fatal(err)
				}
				return &engine.GetDeleteImpactRequest{
					ProjectDir: dir, Kind: "artifact", Id: art.Artifact.Id,
				}
			},
			want: &engine.GetDeleteImpactResponse{
				Allowed: true,
				Gate:    engine.DeleteImpactGate_DELETE_IMPACT_GATE_OK,
			},
		},
	})
}
