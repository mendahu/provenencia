package handlers

import (
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/onboarding"
	"google.golang.org/protobuf/proto"
)

func TestGetNameFormat(t *testing.T) {
	runRPC(t, GetNameFormat, []rpcTest{
		{name: "bad proto", raw: []byte{0xff, 0xff, 0xff, 0xff}, wantErr: true},
		{
			name: "new catalog",
			reqFn: func(t *testing.T) proto.Message {
				t.Helper()
				res, err := onboarding.Complete(t.TempDir(), t.TempDir(), "Ada", "Lovelace")
				if err != nil {
					t.Fatal(err)
				}
				return &engine.GetNameFormatRequest{ProjectDir: res.ProjectDir}
			},
			exact: true,
			want: &engine.GetNameFormatResponse{
				Natural: []*engine.NameFormatPart{
					{PartType: "prefix"},
					{PartType: "given"},
					{PartType: "nick"},
					{PartType: "surname_prefix"},
					{PartType: "surname"},
					{PartType: "suffix"},
				},
				Sorted: []*engine.NameFormatPart{
					{PartType: "surname_prefix"},
					{PartType: "surname", TrailingSeparator: ", "},
					{PartType: "prefix"},
					{PartType: "given"},
					{PartType: "nick"},
					{PartType: "suffix"},
				},
			},
		},
	})
}
