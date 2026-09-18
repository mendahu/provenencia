package handlers

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/catalogsession"
	"github.com/mendahu/provenencia/core/database"
	"google.golang.org/protobuf/proto"
)

func writePNG(t *testing.T, path string) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writePDF(t *testing.T, path string) {
	t.Helper()
	// Minimal bytes that http.DetectContentType reports as application/pdf.
	const body = "%PDF-1.1\n1 0 obj<<>>endobj\ntrailer<<>>\n%%EOF\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureFileThumbnail(t *testing.T) {
	runRPC(t, EnsureFileThumbnail, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "file not attached to artifact",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, _ := sourceFixture(t)
				return &engine.EnsureFileThumbnailRequest{
					ProjectDir: dir,
					FileId:     "00000000-0000-7000-8000-000000000099",
				}
			},
			wantErr: true,
		},
		{
			name: "png ingest then ensure",
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
				path := filepath.Join(t.TempDir(), "tiny.png")
				writePNG(t, path)
				iout, err := IngestArtifactFile(marshalProto(t, &engine.IngestArtifactFileRequest{
					ProjectDir: dir, UserId: userID, ArtifactId: art.Artifact.Id, Path: path,
				}))
				if err != nil {
					t.Fatal(err)
				}
				var ingested engine.IngestArtifactFileResponse
				if err := proto.Unmarshal(iout, &ingested); err != nil {
					t.Fatal(err)
				}
				return &engine.EnsureFileThumbnailRequest{
					ProjectDir: dir, FileId: ingested.File.Id,
				}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var resp engine.EnsureFileThumbnailResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				if resp.GetSkipped() || resp.GetRelPath() == "" {
					t.Fatalf("want thumb path, got %+v", &resp)
				}
				if len(resp.GetRelPath()) < len("objects/") || resp.GetRelPath()[:8] != "objects/" {
					t.Fatalf("rel_path %q", resp.GetRelPath())
				}
			},
		},
		{
			name: "non-image skipped",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, typeID := sourceFixture(t)
				cout, err := CreateSource(marshalProto(t, &engine.CreateSourceRequest{
					ProjectDir: dir, UserId: userID, SourceTypeId: typeID, Title: "PDF",
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
				path := filepath.Join(t.TempDir(), "note.txt")
				if err := os.WriteFile(path, []byte("plain text"), 0o644); err != nil {
					t.Fatal(err)
				}
				iout, err := IngestArtifactFile(marshalProto(t, &engine.IngestArtifactFileRequest{
					ProjectDir: dir, UserId: userID, ArtifactId: art.Artifact.Id, Path: path,
				}))
				if err != nil {
					t.Fatal(err)
				}
				var ingested engine.IngestArtifactFileResponse
				if err := proto.Unmarshal(iout, &ingested); err != nil {
					t.Fatal(err)
				}
				return &engine.EnsureFileThumbnailRequest{
					ProjectDir: dir, FileId: ingested.File.Id,
				}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var resp engine.EnsureFileThumbnailResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				if !resp.GetSkipped() || resp.GetRelPath() != "" {
					t.Fatalf("want skipped empty, got %+v", &resp)
				}
			},
		},
	})
}

func TestListSourcesAndWorkspaceThumbnails(t *testing.T) {
	runRPC(t, ListSources, []rpcTest{
		{
			name: "list source thumb after explicit pin; workspace artifact has raster",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, typeID := sourceFixture(t)
				cout, err := CreateSource(marshalProto(t, &engine.CreateSourceRequest{
					ProjectDir: dir, UserId: userID, SourceTypeId: typeID, Title: "Album",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateSourceResponse
				if err := proto.Unmarshal(cout, &created); err != nil {
					t.Fatal(err)
				}
				aout, err := CreateArtifact(marshalProto(t, &engine.CreateArtifactRequest{
					ProjectDir: dir, UserId: userID, SourceId: created.Source.Id, Label: "Page",
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
				if _, err := SetSourceCover(marshalProto(t, &engine.SetSourceCoverRequest{
					ProjectDir: dir, UserId: userID, SourceId: created.Source.Id,
					CoverMode: "artifact", PrimaryArtifactId: art.Artifact.Id,
				})); err != nil {
					t.Fatal(err)
				}
				return &engine.ListSourcesRequest{ProjectDir: dir}
			},
			after: func(t *testing.T, out []byte, req proto.Message) {
				var list engine.ListSourcesResponse
				if err := proto.Unmarshal(out, &list); err != nil {
					t.Fatal(err)
				}
				if len(list.Sources) != 1 || list.Sources[0].GetThumbnailRelPath() == "" {
					t.Fatalf("list thumbs %+v", list.Sources)
				}
				if list.Sources[0].GetCoverMode() != "artifact" {
					t.Fatalf("cover_mode %q", list.Sources[0].GetCoverMode())
				}
				if !list.Sources[0].GetHasArtifact() {
					t.Fatal("want has_artifact after pin")
				}
				lr := req.(*engine.ListSourcesRequest)
				wout, err := GetSourceWorkspace(marshalProto(t, &engine.GetSourceWorkspaceRequest{
					ProjectDir: lr.ProjectDir, SourceId: list.Sources[0].Id,
				}))
				if err != nil {
					t.Fatal(err)
				}
				var ws engine.GetSourceWorkspaceResponse
				if err := proto.Unmarshal(wout, &ws); err != nil {
					t.Fatal(err)
				}
				if len(ws.Artifacts) != 1 || ws.Artifacts[0].GetThumbnailRelPath() == "" {
					t.Fatalf("workspace thumbs %+v", ws.Artifacts)
				}
			},
		},
		{
			name: "list does not regenerate missing cover thumbnail",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, typeID := sourceFixture(t)
				cout, err := CreateSource(marshalProto(t, &engine.CreateSourceRequest{
					ProjectDir: dir, UserId: userID, SourceTypeId: typeID, Title: "Album",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateSourceResponse
				if err := proto.Unmarshal(cout, &created); err != nil {
					t.Fatal(err)
				}
				aout, err := CreateArtifact(marshalProto(t, &engine.CreateArtifactRequest{
					ProjectDir: dir, UserId: userID, SourceId: created.Source.Id, Label: "Page",
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
				if _, err := SetSourceCover(marshalProto(t, &engine.SetSourceCoverRequest{
					ProjectDir: dir, UserId: userID, SourceId: created.Source.Id,
					CoverMode: "artifact", PrimaryArtifactId: art.Artifact.Id,
				})); err != nil {
					t.Fatal(err)
				}
				// Drop the derivative SetCover created so ListSources must not
				// regenerate it while holding the catalog session.
				if err := catalogsession.Close(dir); err != nil {
					t.Fatal(err)
				}
				c, err := database.Open(dir)
				if err != nil {
					t.Fatal(err)
				}
				db, err := c.DB()
				if err != nil {
					_ = c.Close()
					t.Fatal(err)
				}
				if _, err := db.Exec(`DELETE FROM file_derivatives`); err != nil {
					_ = c.Close()
					t.Fatal(err)
				}
				if err := c.Close(); err != nil {
					t.Fatal(err)
				}
				return &engine.ListSourcesRequest{ProjectDir: dir}
			},
			after: func(t *testing.T, out []byte, req proto.Message) {
				var list engine.ListSourcesResponse
				if err := proto.Unmarshal(out, &list); err != nil {
					t.Fatal(err)
				}
				if len(list.Sources) != 1 {
					t.Fatalf("sources %+v", list.Sources)
				}
				s := list.Sources[0]
				if s.GetCoverMode() != "artifact" {
					t.Fatalf("cover_mode %q", s.GetCoverMode())
				}
				if s.GetThumbnailRelPath() != "" {
					t.Fatalf("list must not generate thumb on cache miss, got %q", s.GetThumbnailRelPath())
				}
				if !s.GetHasArtifact() {
					t.Fatal("want has_artifact even without derivative")
				}
				lr := req.(*engine.ListSourcesRequest)
				if err := catalogsession.Close(lr.ProjectDir); err != nil {
					t.Fatal(err)
				}
				c, err := database.Open(lr.ProjectDir)
				if err != nil {
					t.Fatal(err)
				}
				defer func() { _ = c.Close() }()
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				var n int
				if err := db.QueryRow(`SELECT COUNT(*) FROM file_derivatives`).Scan(&n); err != nil {
					t.Fatal(err)
				}
				if n != 0 {
					t.Fatalf("ListSources regenerated %d derivative row(s)", n)
				}
			},
		},
		{
			name: "fileless source has empty thumbnail",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, typeID := sourceFixture(t)
				if _, err := CreateSource(marshalProto(t, &engine.CreateSourceRequest{
					ProjectDir: dir, UserId: userID, SourceTypeId: typeID, Title: "Empty",
				})); err != nil {
					t.Fatal(err)
				}
				return &engine.ListSourcesRequest{ProjectDir: dir}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var list engine.ListSourcesResponse
				if err := proto.Unmarshal(out, &list); err != nil {
					t.Fatal(err)
				}
				if len(list.Sources) != 1 || list.Sources[0].GetThumbnailRelPath() != "" {
					t.Fatalf("want empty thumb %+v", list.Sources)
				}
				if list.Sources[0].GetThumbnailMediaType() != "" {
					t.Fatalf("want empty media type %+v", list.Sources[0])
				}
				if list.Sources[0].GetHasArtifact() {
					t.Fatal("fileless source must not set has_artifact")
				}
			},
		},
		{
			name: "pdf-only source stays type icon with empty thumbs",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, typeID := sourceFixture(t)
				cout, err := CreateSource(marshalProto(t, &engine.CreateSourceRequest{
					ProjectDir: dir, UserId: userID, SourceTypeId: typeID, Title: "Deed",
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
				path := filepath.Join(t.TempDir(), "deed.pdf")
				writePDF(t, path)
				if _, err := IngestArtifactFile(marshalProto(t, &engine.IngestArtifactFileRequest{
					ProjectDir: dir, UserId: userID, ArtifactId: art.Artifact.Id, Path: path,
				})); err != nil {
					t.Fatal(err)
				}
				return &engine.ListSourcesRequest{ProjectDir: dir}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var list engine.ListSourcesResponse
				if err := proto.Unmarshal(out, &list); err != nil {
					t.Fatal(err)
				}
				if len(list.Sources) != 1 {
					t.Fatalf("sources %+v", list.Sources)
				}
				s := list.Sources[0]
				if s.GetCoverMode() != "type_icon" {
					t.Fatalf("cover_mode %q", s.GetCoverMode())
				}
				if s.GetThumbnailRelPath() != "" || s.GetThumbnailMediaType() != "" {
					t.Fatalf("want empty source thumbs %+v", s)
				}
				if !s.GetHasArtifact() {
					t.Fatal("pdf artifact still counts for has_artifact")
				}
			},
		},
		{
			name: "png ingest leaves type icon until set cover",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, typeID := sourceFixture(t)
				cout, err := CreateSource(marshalProto(t, &engine.CreateSourceRequest{
					ProjectDir: dir, UserId: userID, SourceTypeId: typeID, Title: "Mixed",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateSourceResponse
				if err := proto.Unmarshal(cout, &created); err != nil {
					t.Fatal(err)
				}
				aout, err := CreateArtifact(marshalProto(t, &engine.CreateArtifactRequest{
					ProjectDir: dir, UserId: userID, SourceId: created.Source.Id, Label: "PNG",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var art engine.CreateArtifactResponse
				if err := proto.Unmarshal(aout, &art); err != nil {
					t.Fatal(err)
				}
				path := filepath.Join(t.TempDir(), "scan.png")
				writePNG(t, path)
				if _, err := IngestArtifactFile(marshalProto(t, &engine.IngestArtifactFileRequest{
					ProjectDir: dir, UserId: userID, ArtifactId: art.Artifact.Id, Path: path,
				})); err != nil {
					t.Fatal(err)
				}
				return &engine.ListSourcesRequest{ProjectDir: dir}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var list engine.ListSourcesResponse
				if err := proto.Unmarshal(out, &list); err != nil {
					t.Fatal(err)
				}
				if len(list.Sources) != 1 {
					t.Fatalf("sources %+v", list.Sources)
				}
				s := list.Sources[0]
				if s.GetCoverMode() != "type_icon" || s.GetPrimaryArtifactId() != "" {
					t.Fatalf("want type_icon without auto-pin %+v", s)
				}
				if s.GetThumbnailRelPath() != "" {
					t.Fatalf("want empty raster until pin, got %q", s.GetThumbnailRelPath())
				}
				if !s.GetHasArtifact() {
					t.Fatal("want has_artifact after ingest")
				}
			},
		},
	})
}
