package handlers

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
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
			name: "list and workspace carry thumbnail after png ingest",
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
			},
		},
		{
			name: "pdf-only source has empty raster and pdf media type",
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
				if s.GetThumbnailRelPath() != "" {
					t.Fatalf("want empty raster for pdf, got %q", s.GetThumbnailRelPath())
				}
				if s.GetThumbnailMediaType() != "application/pdf" {
					t.Fatalf("media type %q", s.GetThumbnailMediaType())
				}
				if s.GetThumbnailOriginalFilename() != "deed.pdf" {
					t.Fatalf("filename %q", s.GetThumbnailOriginalFilename())
				}
			},
		},
		{
			name: "pdf then png keeps first file pin as cover",
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
				for _, label := range []string{"PDF", "PNG"} {
					aout, err := CreateArtifact(marshalProto(t, &engine.CreateArtifactRequest{
						ProjectDir: dir, UserId: userID, SourceId: created.Source.Id, Label: label,
					}))
					if err != nil {
						t.Fatal(err)
					}
					var art engine.CreateArtifactResponse
					if err := proto.Unmarshal(aout, &art); err != nil {
						t.Fatal(err)
					}
					var path string
					if label == "PDF" {
						path = filepath.Join(t.TempDir(), "first.pdf")
						writePDF(t, path)
					} else {
						path = filepath.Join(t.TempDir(), "second.png")
						writePNG(t, path)
					}
					if _, err := IngestArtifactFile(marshalProto(t, &engine.IngestArtifactFileRequest{
						ProjectDir: dir, UserId: userID, ArtifactId: art.Artifact.Id, Path: path,
					})); err != nil {
						t.Fatal(err)
					}
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
				if s.GetCoverMode() != "artifact" {
					t.Fatalf("cover_mode %q", s.GetCoverMode())
				}
				if s.GetThumbnailRelPath() != "" {
					t.Fatalf("first pin is pdf, want empty raster got %q", s.GetThumbnailRelPath())
				}
				if s.GetThumbnailMediaType() != "application/pdf" {
					t.Fatalf("media type %q", s.GetThumbnailMediaType())
				}
			},
		},
	})
}
