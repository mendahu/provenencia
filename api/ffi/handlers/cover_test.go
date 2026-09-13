package handlers

import (
	"path/filepath"
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
	"google.golang.org/protobuf/proto"
)

func TestSetSourceCover(t *testing.T) {
	runRPC(t, SetSourceCover, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "pin artifact cover and revert to type icon",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, typeID := sourceFixture(t)
				cout, err := CreateSource(marshalProto(t, &engine.CreateSourceRequest{
					ProjectDir: dir, UserId: userID, SourceTypeId: typeID, Title: "Cover",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateSourceResponse
				if err := proto.Unmarshal(cout, &created); err != nil {
					t.Fatal(err)
				}
				aout, err := CreateArtifact(marshalProto(t, &engine.CreateArtifactRequest{
					ProjectDir: dir, UserId: userID, SourceId: created.Source.Id, Label: "Scan",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var art engine.CreateArtifactResponse
				if err := proto.Unmarshal(aout, &art); err != nil {
					t.Fatal(err)
				}
				path := filepath.Join(t.TempDir(), "page.png")
				writePNG(t, path)
				if _, err := IngestArtifactFile(marshalProto(t, &engine.IngestArtifactFileRequest{
					ProjectDir: dir, UserId: userID, ArtifactId: art.Artifact.Id, Path: path,
				})); err != nil {
					t.Fatal(err)
				}
				// Ingest already auto-pinned; pin again is idempotent.
				return &engine.SetSourceCoverRequest{
					ProjectDir:        dir,
					UserId:            userID,
					SourceId:          created.Source.Id,
					CoverMode:         "artifact",
					PrimaryArtifactId: art.Artifact.Id,
				}
			},
			after: func(t *testing.T, out []byte, req proto.Message) {
				var resp engine.SetSourceCoverResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				s := resp.GetSource()
				if s.GetCoverMode() != "artifact" {
					t.Fatalf("cover_mode %q", s.GetCoverMode())
				}
				cr := req.(*engine.SetSourceCoverRequest)
				if s.GetPrimaryArtifactId() != cr.GetPrimaryArtifactId() {
					t.Fatalf("primary %q", s.GetPrimaryArtifactId())
				}
				if s.GetThumbnailRelPath() == "" {
					t.Fatal("want raster thumb")
				}

				rout, err := SetSourceCover(marshalProto(t, &engine.SetSourceCoverRequest{
					ProjectDir: cr.GetProjectDir(),
					UserId:     cr.GetUserId(),
					SourceId:   cr.GetSourceId(),
					CoverMode:  "type_icon",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var reverted engine.SetSourceCoverResponse
				if err := proto.Unmarshal(rout, &reverted); err != nil {
					t.Fatal(err)
				}
				rs := reverted.GetSource()
				if rs.GetCoverMode() != "type_icon" || rs.GetPrimaryArtifactId() != "" {
					t.Fatalf("reverted %+v", rs)
				}
				if rs.GetThumbnailRelPath() != "" || rs.GetThumbnailMediaType() != "" {
					t.Fatalf("type icon cover should clear thumbs %+v", rs)
				}

				// Fileless cannot be cover.
				aout, err := CreateArtifact(marshalProto(t, &engine.CreateArtifactRequest{
					ProjectDir: cr.GetProjectDir(), UserId: cr.GetUserId(),
					SourceId: cr.GetSourceId(), Label: "Note",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var fileless engine.CreateArtifactResponse
				if err := proto.Unmarshal(aout, &fileless); err != nil {
					t.Fatal(err)
				}
				if _, err := SetSourceCover(marshalProto(t, &engine.SetSourceCoverRequest{
					ProjectDir:        cr.GetProjectDir(),
					UserId:            cr.GetUserId(),
					SourceId:          cr.GetSourceId(),
					CoverMode:         "artifact",
					PrimaryArtifactId: fileless.Artifact.Id,
				})); err == nil {
					t.Fatal("want fileless pin error")
				}
			},
		},
	})
}

func TestGetSourceWorkspaceCover(t *testing.T) {
	runRPC(t, GetSourceWorkspace, []rpcTest{
		{
			name: "enriched cover on workspace source",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, typeID := sourceFixture(t)
				cout, err := CreateSource(marshalProto(t, &engine.CreateSourceRequest{
					ProjectDir: dir, UserId: userID, SourceTypeId: typeID, Title: "WS",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateSourceResponse
				if err := proto.Unmarshal(cout, &created); err != nil {
					t.Fatal(err)
				}
				aout, err := CreateArtifact(marshalProto(t, &engine.CreateArtifactRequest{
					ProjectDir: dir, UserId: userID, SourceId: created.Source.Id, Label: "A",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var art engine.CreateArtifactResponse
				if err := proto.Unmarshal(aout, &art); err != nil {
					t.Fatal(err)
				}
				path := filepath.Join(t.TempDir(), "ws.png")
				writePNG(t, path)
				if _, err := IngestArtifactFile(marshalProto(t, &engine.IngestArtifactFileRequest{
					ProjectDir: dir, UserId: userID, ArtifactId: art.Artifact.Id, Path: path,
				})); err != nil {
					t.Fatal(err)
				}
				return &engine.GetSourceWorkspaceRequest{ProjectDir: dir, SourceId: created.Source.Id}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var ws engine.GetSourceWorkspaceResponse
				if err := proto.Unmarshal(out, &ws); err != nil {
					t.Fatal(err)
				}
				s := ws.GetSource()
				if s.GetCoverMode() != "artifact" || s.GetPrimaryArtifactId() == "" {
					t.Fatalf("cover %+v", s)
				}
				if s.GetThumbnailRelPath() == "" {
					t.Fatal("want raster on workspace source")
				}
			},
		},
	})
}
