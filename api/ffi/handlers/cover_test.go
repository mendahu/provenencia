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
				f := sourceWithPNG(t)
				// Ingest does not auto-pin; this call is the first cover set.
				return &engine.SetSourceCoverRequest{
					ProjectDir:        f.dir,
					UserId:            f.userID,
					SourceId:          f.sourceID,
					CoverMode:         "artifact",
					PrimaryArtifactId: f.artifactID,
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

				// PDF cannot be cover (no raster).
				pout, err := CreateArtifact(marshalProto(t, &engine.CreateArtifactRequest{
					ProjectDir: cr.GetProjectDir(), UserId: cr.GetUserId(),
					SourceId: cr.GetSourceId(), Label: "PDF",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var pdfArt engine.CreateArtifactResponse
				if err := proto.Unmarshal(pout, &pdfArt); err != nil {
					t.Fatal(err)
				}
				pdfPath := filepath.Join(t.TempDir(), "no-cover.pdf")
				writePDF(t, pdfPath)
				if _, err := IngestArtifactFile(marshalProto(t, &engine.IngestArtifactFileRequest{
					ProjectDir: cr.GetProjectDir(), UserId: cr.GetUserId(),
					ArtifactId: pdfArt.Artifact.Id, Path: pdfPath,
				})); err != nil {
					t.Fatal(err)
				}
				if _, err := SetSourceCover(marshalProto(t, &engine.SetSourceCoverRequest{
					ProjectDir:        cr.GetProjectDir(),
					UserId:            cr.GetUserId(),
					SourceId:          cr.GetSourceId(),
					CoverMode:         "artifact",
					PrimaryArtifactId: pdfArt.Artifact.Id,
				})); err == nil {
					t.Fatal("want pdf pin error")
				}
			},
		},
	})
}

func TestGetSourceWorkspaceCover(t *testing.T) {
	runRPC(t, GetSourceWorkspace, []rpcTest{
		{
			name: "enriched cover on workspace source after explicit pin",
			reqFn: func(t *testing.T) proto.Message {
				f := sourceWithPNG(t)
				if _, err := SetSourceCover(marshalProto(t, &engine.SetSourceCoverRequest{
					ProjectDir: f.dir, UserId: f.userID, SourceId: f.sourceID,
					CoverMode: "artifact", PrimaryArtifactId: f.artifactID,
				})); err != nil {
					t.Fatal(err)
				}
				return &engine.GetSourceWorkspaceRequest{ProjectDir: f.dir, SourceId: f.sourceID}
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
