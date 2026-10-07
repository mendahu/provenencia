package handlers

import (
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
	"google.golang.org/protobuf/proto"
)

func TestProposePromoteGraphAlignment(t *testing.T) {
	runRPC(t, ProposePromoteGraphAlignment, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "bad source id",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, _, _ := subjectFixture(t)
				return &engine.ProposePromoteGraphAlignmentRequest{ProjectDir: dir, SourceId: "nope"}
			},
			wantErr: true,
		},
		{
			name: "empty source returns no rows",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, sourceID, _ := subjectFixture(t)
				return &engine.ProposePromoteGraphAlignmentRequest{ProjectDir: dir, SourceId: sourceID}
			},
			want: &engine.ProposePromoteGraphAlignmentResponse{},
		},
	})
}
