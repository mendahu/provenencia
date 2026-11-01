package handlers

import (
	"errors"
	"path/filepath"
	"slices"
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/catalogsession"
	"github.com/mendahu/provenencia/core/database/sources"
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

func TestSetCoverMakesThumbnailFirst(t *testing.T) {
	t.Cleanup(func() { _ = catalogsession.CloseAll() })
	setCover := func(s thumbSource, mode, artifactID string) (*engine.SetSourceCoverResponse, error) {
		var resp engine.SetSourceCoverResponse
		err := rpc(t, SetSourceCover, &engine.SetSourceCoverRequest{
			ProjectDir: s.dir, UserId: s.userID, SourceId: s.sourceID, CoverMode: mode, PrimaryArtifactId: artifactID,
		}, &resp)
		return &resp, err
	}

	t.Run("a raster Artifact with no derivative gets its thumbnail and the cover", func(t *testing.T) {
		s := newThumbSource(t)
		p := filepath.Join(t.TempDir(), "p.png")
		writePNG(t, p)
		ingested := s.addArtifact(t, "Page", p)
		damageCatalog(t, s.dir, "DELETE FROM file_derivatives")
		resp, err := setCover(s, "artifact", ingested.Artifact.Id)
		if err != nil {
			t.Fatal(err)
		}
		if resp.GetSource().GetCoverMode() != "artifact" || resp.GetSource().GetThumbnailRelPath() == "" {
			t.Fatalf("source %v", resp.GetSource())
		}
		if n := s.count(t, derivativeCount); n != 1 {
			t.Fatalf("derivative rows %d, want 1", n)
		}
	})

	refused := []struct {
		name string
		art  func(t *testing.T, s thumbSource) string
	}{
		{"a PDF Artifact", func(t *testing.T, s thumbSource) string {
			p := filepath.Join(t.TempDir(), "deed.pdf")
			writePDF(t, p)
			return s.addArtifact(t, "Deed", p).Artifact.Id
		}},
		{"a late-failing JPEG", func(t *testing.T, s thumbSource) string {
			return s.addArtifact(t, "Late", writeLateFailingJPEG(t)).Artifact.Id
		}},
		{"an Artifact of another Source", func(t *testing.T, s thumbSource) string {
			other := thumbSource{dir: s.dir, userID: s.userID}
			var created engine.CreateSourceResponse
			var types engine.ListSourceTypesResponse
			mustRPC(t, ListSourceTypes, &engine.ListSourceTypesRequest{ProjectDir: s.dir}, &types)
			mustRPC(t, CreateSource, &engine.CreateSourceRequest{
				ProjectDir: s.dir, UserId: s.userID, SourceTypeId: types.Types[0].Id, Title: "Other",
			}, &created)
			other.sourceID = created.Source.Id
			p := filepath.Join(t.TempDir(), "p.png")
			writePNG(t, p)
			return other.addArtifact(t, "Page", p).Artifact.Id
		}},
	}
	for _, tt := range refused {
		t.Run(tt.name+" is invalid", func(t *testing.T) {
			s := newThumbSource(t)
			id := tt.art(t, s)
			_, err := setCover(s, "artifact", id)
			if !errors.Is(err, sources.ErrInvalid) {
				t.Fatalf("err = %v, want sources.ErrInvalid", err)
			}
			if temps := objectTemps(t, s.dir, "."); len(temps) != 0 {
				t.Fatalf("temps left: %v", temps)
			}
		})
	}

	t.Run("type icon mode makes no thumbnail and stages nothing", func(t *testing.T) {
		s := newThumbSource(t)
		p := filepath.Join(t.TempDir(), "p.png")
		writePNG(t, p)
		s.addArtifact(t, "Page", p)
		damageCatalog(t, s.dir, "DELETE FROM file_derivatives")
		before := objectRelPaths(t, s.dir)
		resp, err := setCover(s, "type_icon", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.GetSource().GetCoverMode() != "type_icon" {
			t.Fatalf("source %v", resp.GetSource())
		}
		if got := objectRelPaths(t, s.dir); !slices.Equal(got, before) {
			t.Fatalf("objects %v, want %v", got, before)
		}
		if temps := objectTemps(t, s.dir, "."); len(temps) != 0 {
			t.Fatalf("temps left: %v", temps)
		}
		if n := s.count(t, derivativeCount); n != 0 {
			t.Fatalf("derivative rows %d, want 0", n)
		}
	})
}
