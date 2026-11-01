package handlers

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/catalogsession"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/raster/rastertest"
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

// writeCrop writes a w x h JPEG filled with opaque gray shade to a temp folder
// as crop.jpg and returns its absolute path. A JPEG, so the PNG Go stores
// cannot be the crop's own bytes.
func writeCrop(t *testing.T, w, h int, shade uint8) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.Gray{Y: shade}), image.Point{}, draw.Src)
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "crop.jpg")
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
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
				f := sourceWithPNG(t)
				return &engine.EnsureFileThumbnailRequest{
					ProjectDir: f.dir, FileId: f.fileID,
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
			name: "pdf skipped",
			reqFn: func(t *testing.T) proto.Message {
				s := newThumbSource(t)
				p := filepath.Join(t.TempDir(), "deed.pdf")
				writePDF(t, p)
				ingested := s.addArtifact(t, "Deed", p)
				return &engine.EnsureFileThumbnailRequest{ProjectDir: s.dir, FileId: ingested.File.Id}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var resp engine.EnsureFileThumbnailResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				if !resp.GetSkipped() || resp.GetRelPath() != "" {
					t.Fatalf("want skipped empty, got %v", &resp)
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
				f := sourceWithPNG(t)
				if _, err := SetSourceCover(marshalProto(t, &engine.SetSourceCoverRequest{
					ProjectDir: f.dir, UserId: f.userID, SourceId: f.sourceID,
					CoverMode: "artifact", PrimaryArtifactId: f.artifactID,
				})); err != nil {
					t.Fatal(err)
				}
				return &engine.ListSourcesRequest{ProjectDir: f.dir}
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
				f := sourceWithPNG(t)
				if _, err := SetSourceCover(marshalProto(t, &engine.SetSourceCoverRequest{
					ProjectDir: f.dir, UserId: f.userID, SourceId: f.sourceID,
					CoverMode: "artifact", PrimaryArtifactId: f.artifactID,
				})); err != nil {
					t.Fatal(err)
				}
				// Drop the derivative SetCover created so ListSources must not
				// regenerate it while holding the catalog session.
				damageCatalog(t, f.dir, "DELETE FROM file_derivatives")
				return &engine.ListSourcesRequest{ProjectDir: f.dir}
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
				f := sourceWithPNG(t)
				return &engine.ListSourcesRequest{ProjectDir: f.dir}
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

// thumbSource is a Source for the thumbnail pass tests; addArtifact puts
// Artifacts on it.
type thumbSource struct {
	dir, userID, sourceID string
}

func newThumbSource(t *testing.T) thumbSource {
	t.Helper()
	dir, userID, typeID := sourceFixture(t)
	var created engine.CreateSourceResponse
	mustRPC(t, CreateSource, &engine.CreateSourceRequest{
		ProjectDir: dir, UserId: userID, SourceTypeId: typeID, Title: "Scans",
	}, &created)
	return thumbSource{dir: dir, userID: userID, sourceID: created.Source.Id}
}

// addArtifact creates an Artifact holding the file at path and returns the
// ingest response.
func (s thumbSource) addArtifact(t *testing.T, label, path string) *engine.IngestArtifactFileResponse {
	t.Helper()
	var art engine.CreateArtifactResponse
	mustRPC(t, CreateArtifact, &engine.CreateArtifactRequest{
		ProjectDir: s.dir, UserId: s.userID, SourceId: s.sourceID, Label: label,
	}, &art)
	var ingested engine.IngestArtifactFileResponse
	mustRPC(t, IngestArtifactFile, &engine.IngestArtifactFileRequest{
		ProjectDir: s.dir, UserId: s.userID, ArtifactId: art.Artifact.Id, Path: path,
	}, &ingested)
	return &ingested
}

// addArtifactOnFile creates an Artifact on a File an Artifact already holds.
func (s thumbSource) addArtifactOnFile(t *testing.T, label, fileID string) {
	t.Helper()
	var art engine.CreateArtifactResponse
	mustRPC(t, CreateArtifact, &engine.CreateArtifactRequest{
		ProjectDir: s.dir, UserId: s.userID, SourceId: s.sourceID, Label: label, FileId: fileID,
	}, &art)
}

func (s thumbSource) workspace(t *testing.T) *engine.GetSourceWorkspaceResponse {
	t.Helper()
	var ws engine.GetSourceWorkspaceResponse
	mustRPC(t, GetSourceWorkspace, &engine.GetSourceWorkspaceRequest{ProjectDir: s.dir, SourceId: s.sourceID}, &ws)
	return &ws
}

// count runs a COUNT(*) query on the project's catalog.
func (s thumbSource) count(t *testing.T, query string) int {
	t.Helper()
	var n int
	err := catalogsession.Do(s.dir, func(c *database.Catalog) error {
		db, err := c.DB()
		if err != nil {
			return err
		}
		return db.QueryRow(query).Scan(&n)
	})
	if err != nil {
		t.Fatal(err)
	}
	return n
}

// writeShadedPNG writes a w x h PNG whose pixels are shade, so each shade is a
// File of its own.
func writeShadedPNG(t *testing.T, path string, shade uint8) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.Gray{Y: shade}), image.Point{}, draw.Src)
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}

// writeLateFailingJPEG writes the oversized JPEG without its EOI marker: it
// decodes far enough to fail late, over the pixel budget.
func writeLateFailingJPEG(t *testing.T) string {
	t.Helper()
	raw := rastertest.OversizedJPEG(t)
	path := filepath.Join(t.TempDir(), "late.jpg")
	if err := os.WriteFile(path, raw[:len(raw)-2], 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

const (
	revisionCount   = `SELECT COUNT(*) FROM audit_transactions`
	derivativeCount = `SELECT COUNT(*) FROM file_derivatives`
)

func TestWorkspaceThumbnailPass(t *testing.T) {
	t.Cleanup(func() { _ = catalogsession.CloseAll() })

	t.Run("two Artifacts with no derivatives get both paths and no revision", func(t *testing.T) {
		s := newThumbSource(t)
		for i := range 2 {
			p := filepath.Join(t.TempDir(), "p.png")
			writeShadedPNG(t, p, uint8(10+i))
			s.addArtifact(t, "Page", p)
		}
		damageCatalog(t, s.dir, "DELETE FROM file_derivatives")
		before := s.count(t, revisionCount)
		ws := s.workspace(t)
		if len(ws.Artifacts) != 2 {
			t.Fatalf("artifacts %d", len(ws.Artifacts))
		}
		for _, a := range ws.Artifacts {
			if a.GetThumbnailRelPath() == "" {
				t.Fatalf("no thumbnail for %q", a.GetLabel())
			}
		}
		if n := s.count(t, derivativeCount); n != 2 {
			t.Fatalf("derivative rows %d, want 2", n)
		}
		if after := s.count(t, revisionCount); after != before {
			t.Fatalf("revisions %d, was %d", after, before)
		}
	})

	t.Run("two Artifacts on one File decode once", func(t *testing.T) {
		s := newThumbSource(t)
		p := filepath.Join(t.TempDir(), "p.png")
		writePNG(t, p)
		first := s.addArtifact(t, "Page", p)
		s.addArtifactOnFile(t, "Again", first.File.Id)
		damageCatalog(t, s.dir, "DELETE FROM file_derivatives")
		ws := s.workspace(t)
		if len(ws.Artifacts) != 2 || ws.Artifacts[0].GetThumbnailRelPath() == "" ||
			ws.Artifacts[0].GetThumbnailRelPath() != ws.Artifacts[1].GetThumbnailRelPath() {
			t.Fatalf("thumbnails %q, %q", ws.Artifacts[0].GetThumbnailRelPath(), ws.Artifacts[1].GetThumbnailRelPath())
		}
		if n := s.count(t, derivativeCount); n != 1 {
			t.Fatalf("derivative rows %d, want 1", n)
		}
	})

	t.Run("a late-failing JPEG shows no thumbnail and writes no link", func(t *testing.T) {
		s := newThumbSource(t)
		ingested := s.addArtifact(t, "Late", writeLateFailingJPEG(t))
		if ingested.Artifact.GetThumbnailRelPath() != "" {
			t.Fatalf("ingest thumbnail %q", ingested.Artifact.GetThumbnailRelPath())
		}
		ws := s.workspace(t)
		if len(ws.Artifacts) != 1 || ws.Artifacts[0].GetThumbnailRelPath() != "" {
			t.Fatalf("workspace %+v", ws.Artifacts)
		}
		if n := s.count(t, derivativeCount); n != 0 {
			t.Fatalf("derivative rows %d, want 0", n)
		}
	})

	t.Run("17 Artifacts fill in two batches", func(t *testing.T) {
		s := newThumbSource(t)
		for i := range thumbnailBatch + 1 {
			p := filepath.Join(t.TempDir(), "p.png")
			writeShadedPNG(t, p, uint8(i*7))
			s.addArtifact(t, "Page", p)
		}
		damageCatalog(t, s.dir, "DELETE FROM file_derivatives")
		ws := s.workspace(t)
		if len(ws.Artifacts) != thumbnailBatch+1 {
			t.Fatalf("artifacts %d", len(ws.Artifacts))
		}
		for _, a := range ws.Artifacts {
			if a.GetThumbnailRelPath() == "" {
				t.Fatalf("no thumbnail for artifact %q", a.GetId())
			}
		}
		if n := s.count(t, derivativeCount); n != thumbnailBatch+1 {
			t.Fatalf("derivative rows %d, want %d", n, thumbnailBatch+1)
		}
	})

	t.Run("every derivative present changes nothing", func(t *testing.T) {
		s := newThumbSource(t)
		p := filepath.Join(t.TempDir(), "p.png")
		writePNG(t, p)
		s.addArtifact(t, "Page", p)
		first := s.workspace(t)
		links, objects := s.count(t, derivativeCount), len(objectRelPaths(t, s.dir))
		second := s.workspace(t)
		if !proto.Equal(first, second) {
			t.Fatalf("workspace changed: %v vs %v", first, second)
		}
		if s.count(t, derivativeCount) != links || len(objectRelPaths(t, s.dir)) != objects {
			t.Fatal("a repeat read changed links or objects")
		}
		if temps := objectTemps(t, s.dir, "."); len(temps) != 0 {
			t.Fatalf("temps left: %v", temps)
		}
	})
}
