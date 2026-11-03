package handlers

import (
	"path/filepath"
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
	"google.golang.org/protobuf/proto"
)

func TestGetWorkspaceNavCounts(t *testing.T) {
	runRPC(t, GetWorkspaceNavCounts, []rpcTest{
		{name: "bad proto", raw: []byte{0xff, 0xff, 0xff, 0xff}, wantErr: true},
		{
			name: "onboarded project has zero sources and no handles",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, _ := sourceFixture(t)
				return &engine.GetWorkspaceNavCountsRequest{ProjectDir: dir}
			},
			want:  &engine.GetWorkspaceNavCountsResponse{Sources: 0},
			exact: true,
		},
		{
			name: "counts sources after ingest",
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
					ProjectDir: dir, UserId: userID, SourceId: created.Source.Id, Label: "Front", Description: "front",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var art engine.CreateArtifactResponse
				if err := proto.Unmarshal(aout, &art); err != nil {
					t.Fatal(err)
				}
				path := filepath.Join(t.TempDir(), "scan.pdf")
				writePDF(t, path)
				if _, err := IngestArtifactFile(marshalProto(t, &engine.IngestArtifactFileRequest{
					ProjectDir: dir, UserId: userID, ArtifactId: art.Artifact.Id, Path: path,
				})); err != nil {
					t.Fatal(err)
				}
				return &engine.GetWorkspaceNavCountsRequest{ProjectDir: dir}
			},
			want:  &engine.GetWorkspaceNavCountsResponse{Sources: 1},
			exact: true,
		},
	})
}
