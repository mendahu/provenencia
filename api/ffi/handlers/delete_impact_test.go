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
	})
}
