package derivatives

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/catalogsession"
	"github.com/mendahu/provenencia/core/catalogsession/catalogsessiontest"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/filederivatives"
	"github.com/mendahu/provenencia/core/database/files"
	"github.com/mendahu/provenencia/core/database/searchindex"
	"github.com/mendahu/provenencia/core/ingest"
	"github.com/mendahu/provenencia/core/raster"
	"github.com/mendahu/provenencia/core/raster/rastertest"
	"github.com/mendahu/provenencia/core/ref"
)

func installSourceFile(t *testing.T, c *database.Catalog, data []byte, mediaType, name string) []byte {
	t.Helper()
	sum := sha256.Sum256(data)
	checksum := hex.EncodeToString(sum[:])
	rel, err := files.StorageRelPath(checksum, mediaType)
	if err != nil {
		t.Fatal(err)
	}
	objPath := filepath.Join(c.Dir(), filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(objPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(objPath, data, 0o644); err != nil {
		t.Fatal(err)
	}
	id, err := files.NewID()
	if err != nil {
		t.Fatal(err)
	}
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	if err := files.Insert(tx, files.File{
		ID: id, ChecksumSHA256: checksum, OriginalFilename: name, MediaType: mediaType, ByteSize: int64(len(data)),
	}); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	return id
}

func auditTotal(t *testing.T, c *database.Catalog) int {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM audit_transactions`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestEnsureThumbnailIdempotent(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	pngBytes, err := os.ReadFile("testdata/tiny.png")
	if err != nil {
		t.Fatal(err)
	}
	srcID := installSourceFile(t, c, pngBytes, "image/png", "tiny.png")
	before := auditTotal(t, c)

	first, err := EnsureThumbnail(c, srcID)
	if err != nil || first.Skipped {
		t.Fatalf("%+v %v", first, err)
	}
	if len(first.Link.DerivedFileID) != 16 {
		t.Fatal("missing derived id")
	}
	derived, err := files.Lookup(c, first.Link.DerivedFileID)
	if err != nil {
		t.Fatal(err)
	}
	if derived.MediaType != "image/jpeg" || derived.OriginalFilename != "" {
		t.Fatalf("%+v", derived)
	}
	rel, err := files.StorageRelPath(derived.ChecksumSHA256, derived.MediaType)
	if err != nil {
		t.Fatal(err)
	}
	thumbPath := filepath.Join(c.Dir(), filepath.FromSlash(rel))
	thumbBytes, err := os.ReadFile(thumbPath)
	if err != nil || len(thumbBytes) == 0 {
		t.Fatalf("thumb object %v len=%d", err, len(thumbBytes))
	}
	if _, _, err := image.Decode(bytes.NewReader(thumbBytes)); err != nil {
		t.Fatalf("decode thumb: %v", err)
	}

	second, err := EnsureThumbnail(c, srcID)
	if err != nil || second.Skipped {
		t.Fatalf("%+v %v", second, err)
	}
	if string(second.Link.DerivedFileID) != string(first.Link.DerivedFileID) {
		t.Fatalf("derived id changed: %x vs %x", first.Link.DerivedFileID, second.Link.DerivedFileID)
	}
	if string(second.Link.ID) != string(first.Link.ID) {
		t.Fatalf("link id changed")
	}
	list, err := filederivatives.ListBySourceFile(c, srcID)
	if err != nil || len(list) != 1 {
		t.Fatalf("%v %+v", err, list)
	}
	if auditTotal(t, c) != before {
		t.Fatalf("EnsureThumbnail wrote audit: before=%d after=%d", before, auditTotal(t, c))
	}
}

func TestEnsureThumbnailOversizedJPEG(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	srcID := installSourceFile(t, c, rastertest.OversizedJPEG(t), "image/jpeg", "big.jpg")

	first, err := EnsureThumbnail(c, srcID)
	if err != nil || first.Skipped {
		t.Fatalf("%+v %v", first, err)
	}
	derived, err := files.Lookup(c, first.Link.DerivedFileID)
	if err != nil {
		t.Fatal(err)
	}
	if derived.MediaType != "image/jpeg" {
		t.Fatalf("derived media type = %q", derived.MediaType)
	}
	rel, err := files.StorageRelPath(derived.ChecksumSHA256, derived.MediaType)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(c.Dir(), filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Width != 256 || cfg.Height != 255 {
		t.Fatalf("thumb = %dx%d, want 256x255", cfg.Width, cfg.Height)
	}
	second, err := EnsureThumbnail(c, srcID)
	if err != nil {
		t.Fatal(err)
	}
	if string(second.Link.ID) != string(first.Link.ID) {
		t.Fatal("second call made a new link")
	}
}

func TestEnsureThumbnailSkipsNonImage(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	pdfID := installSourceFile(t, c, []byte("%PDF-1.4"), "application/pdf", "doc.pdf")
	res, err := EnsureThumbnail(c, pdfID)
	if err != nil || !res.Skipped {
		t.Fatalf("%+v %v", res, err)
	}
	list, err := filederivatives.ListBySourceFile(c, pdfID)
	if err != nil || len(list) != 0 {
		t.Fatalf("%v %+v", err, list)
	}

	binID := installSourceFile(t, c, []byte{0, 1, 2}, "audio/mpeg", "blob.mp3")
	res, err = EnsureThumbnail(c, binID)
	if err != nil || !res.Skipped {
		t.Fatalf("%+v %v", res, err)
	}
}

func TestEnsureThumbnailFromGeneratedPNG(t *testing.T) {
	// Extra coverage: encode a larger PNG in-memory (not only fixture).
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	img := image.NewRGBA(image.Rect(0, 0, 400, 200))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	srcID := installSourceFile(t, c, buf.Bytes(), "image/png", "wide.png")
	res, err := EnsureThumbnail(c, srcID)
	if err != nil || res.Skipped {
		t.Fatalf("%+v %v", res, err)
	}
	derived, err := files.Lookup(c, res.Link.DerivedFileID)
	if err != nil {
		t.Fatal(err)
	}
	rel, _ := files.StorageRelPath(derived.ChecksumSHA256, derived.MediaType)
	raw, err := os.ReadFile(filepath.Join(c.Dir(), filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	decoded, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	b := decoded.Bounds()
	if b.Dx() > 256 || b.Dy() > 256 {
		t.Fatalf("thumb too large: %dx%d", b.Dx(), b.Dy())
	}
	if b.Dx() != 256 || b.Dy() != 128 {
		t.Fatalf("want 256x128, got %dx%d", b.Dx(), b.Dy())
	}
}

// installFileRow inserts a files row without writing an object, so tests can
// declare arbitrary ByteSize/checksum metadata.
func installFileRow(t *testing.T, c *database.Catalog, checksum, mediaType string, byteSize int64) []byte {
	t.Helper()
	id, err := files.NewID()
	if err != nil {
		t.Fatal(err)
	}
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	if err := files.Insert(tx, files.File{
		ID: id, ChecksumSHA256: checksum, MediaType: mediaType, ByteSize: byteSize,
	}); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	return id
}

func TestEnsureThumbnailRejectsUnsafeSources(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	tinySum := sha256.Sum256([]byte("placeholder"))
	tests := []struct {
		name   string
		srcID  func(t *testing.T) []byte
		wantIs error
	}{
		{
			name: "declared ByteSize over source cap",
			srcID: func(t *testing.T) []byte {
				return installFileRow(t, c, hex.EncodeToString(tinySum[:]), "image/png", raster.MaxDecodeBytes+1)
			},
			wantIs: ErrUnprocessable,
		},
		{
			name: "object bytes do not match catalog checksum",
			srcID: func(t *testing.T) []byte {
				id := installSourceFile(t, c, rastertest.HeaderOnlyPNG(t, 4, 4), "image/png", "swap.png")
				sum := sha256.Sum256(rastertest.HeaderOnlyPNG(t, 4, 4))
				rel, err := files.StorageRelPath(hex.EncodeToString(sum[:]), "image/png")
				if err != nil {
					t.Fatal(err)
				}
				objPath := filepath.Join(c.Dir(), filepath.FromSlash(rel))
				if err := os.WriteFile(objPath, []byte("tampered bytes"), 0o644); err != nil {
					t.Fatal(err)
				}
				return id
			},
			wantIs: ErrCorruptObject,
		},
		{
			name: "object is a symlink to an outside file with the same bytes",
			srcID: func(t *testing.T) []byte {
				data := rastertest.HeaderOnlyPNG(t, 5, 5)
				id := installSourceFile(t, c, data, "image/png", "link.png")
				sum := sha256.Sum256(data)
				rel, err := files.StorageRelPath(hex.EncodeToString(sum[:]), "image/png")
				if err != nil {
					t.Fatal(err)
				}
				objPath := filepath.Join(c.Dir(), filepath.FromSlash(rel))
				outside := filepath.Join(t.TempDir(), "outside.png")
				if err := os.WriteFile(outside, data, 0o644); err != nil {
					t.Fatal(err)
				}
				if err := os.Remove(objPath); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(outside, objPath); err != nil {
					t.Fatal(err)
				}
				return id
			},
			wantIs: ErrCorruptObject,
		},
		{
			name: "source's shard folder is a symlink",
			srcID: func(t *testing.T) []byte {
				data := rastertest.HeaderOnlyPNG(t, 6, 6)
				id := installSourceFile(t, c, data, "image/png", "shard.png")
				sum := sha256.Sum256(data)
				rel, err := files.StorageRelPath(hex.EncodeToString(sum[:]), "image/png")
				if err != nil {
					t.Fatal(err)
				}
				shard := filepath.Join(c.Dir(), filepath.FromSlash(strings.Join(strings.Split(rel, "/")[:2], "/")))
				moved := filepath.Join(t.TempDir(), "moved")
				if err := os.Rename(shard, moved); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(moved, shard); err != nil {
					t.Fatal(err)
				}
				return id
			},
			wantIs: ErrCorruptObject,
		},
		{
			name: "raster type off the allowlist has no object path",
			srcID: func(t *testing.T) []byte {
				jpgSum := sha256.Sum256([]byte("image/jpg is not a stored type"))
				return installFileRow(t, c, hex.EncodeToString(jpgSum[:]), "image/jpg", 10)
			},
			wantIs: ErrUnprocessable,
		},
		{
			name: "truncated jpeg",
			srcID: func(t *testing.T) []byte {
				b := rastertest.GoFixture(t, "video-001.jpeg")
				return installSourceFile(t, c, b[:len(b)/2], "image/jpeg", "half.jpg")
			},
			wantIs: ErrUnprocessable,
		},
		{
			name: "pixel bomb png over decode budget",
			srcID: func(t *testing.T) []byte {
				return installSourceFile(t, c, rastertest.HeaderOnlyPNG(t, 100000, 100000), "image/png", "bomb.png")
			},
			wantIs: ErrUnprocessable,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srcID := tt.srcID(t)
			res, err := EnsureThumbnail(c, srcID)
			if err == nil {
				t.Fatalf("expected error, got %+v", res)
			}
			if !errors.Is(err, tt.wantIs) {
				t.Fatalf("got %v, want %v", err, tt.wantIs)
			}
			list, err := filederivatives.ListBySourceFile(c, srcID)
			if err != nil || len(list) != 0 {
				t.Fatalf("no link rows expected: %v %+v", err, list)
			}
		})
	}
}

func TestEnsureRejectsMaxEdgeOverLimit(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	pngBytes, err := os.ReadFile("testdata/tiny.png")
	if err != nil {
		t.Fatal(err)
	}
	srcID := installSourceFile(t, c, pngBytes, "image/png", "tiny.png")
	_, err = Ensure(c, srcID, Spec{
		Type:   "huge",
		Raster: raster.Options{MaxEdge: raster.MaxEdgeLimit + 1},
	})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("got %v, want ErrInvalid", err)
	}
}

func TestEnsureCustomSpecAlongsideThumbnail(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	img := image.NewRGBA(image.Rect(0, 0, 800, 400))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	srcID := installSourceFile(t, c, buf.Bytes(), "image/png", "wide.png")

	thumb, err := EnsureThumbnail(c, srcID)
	if err != nil || thumb.Skipped {
		t.Fatalf("thumb %+v %v", thumb, err)
	}
	medium, err := Ensure(c, srcID, Spec{
		Type:   "medium",
		Raster: raster.Options{MaxEdge: 512, Quality: 90},
	})
	if err != nil || medium.Skipped {
		t.Fatalf("medium %+v %v", medium, err)
	}
	if medium.Link.DerivativeType != "medium" {
		t.Fatalf("type %q", medium.Link.DerivativeType)
	}
	if string(medium.Link.DerivedFileID) == string(thumb.Link.DerivedFileID) {
		t.Fatal("medium and thumbnail should be distinct derived files")
	}
	list, err := filederivatives.ListBySourceFile(c, srcID)
	if err != nil || len(list) != 2 {
		t.Fatalf("%v %+v", err, list)
	}

	derived, err := files.Lookup(c, medium.Link.DerivedFileID)
	if err != nil {
		t.Fatal(err)
	}
	rel, _ := files.StorageRelPath(derived.ChecksumSHA256, derived.MediaType)
	raw, err := os.ReadFile(filepath.Join(c.Dir(), filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	decoded, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	b := decoded.Bounds()
	if b.Dx() != 512 || b.Dy() != 256 {
		t.Fatalf("want 512x256, got %dx%d", b.Dx(), b.Dy())
	}
}

func TestEnsureThumbnailCommitsWithoutRevisionAndRefreshesSearch(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	pngBytes, err := os.ReadFile("testdata/tiny.png")
	if err != nil {
		t.Fatal(err)
	}
	fileID := installSourceFile(t, c, pngBytes, "image/png", "tiny.png")

	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	typeID, sourceID, artID := newID(t), newID(t), newID(t)
	sourceRef, err := ref.Mint(ref.PrefixSource)
	if err != nil {
		t.Fatal(err)
	}
	artRef, err := ref.Mint(ref.PrefixArtifact)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(
		`INSERT INTO source_types (id, key, origin, label, icon_key) VALUES (?, 'letter', 'user', 'Letter', 'type_evidence')`,
		typeID,
	); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(
		`INSERT INTO sources (id, ref, source_type_id, title) VALUES (?, ?, ?, 'Letter home')`,
		sourceID, sourceRef, typeID,
	); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(
		`INSERT INTO artifacts (id, ref, source_id, file_id, label) VALUES (?, ?, ?, ?, 'Scan')`,
		artID, artRef, sourceID, fileID,
	); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(
		`UPDATE sources SET cover_mode = 'artifact', primary_artifact_id = ? WHERE id = ?`,
		artID, sourceID,
	); err != nil {
		t.Fatal(err)
	}

	before := auditTotal(t, c)
	got, err := EnsureThumbnail(c, fileID)
	if err != nil || got.Skipped {
		t.Fatalf("%+v %v", got, err)
	}
	if _, err := files.Lookup(c, got.Link.DerivedFileID); err != nil {
		t.Fatal(err)
	}
	if auditTotal(t, c) != before {
		t.Fatalf("unaudited thumbnail wrote audit: before=%d after=%d", before, auditTotal(t, c))
	}

	id, err := uuid.FromBytes(sourceID)
	if err != nil {
		t.Fatal(err)
	}
	var path string
	if err := db.QueryRow(
		`SELECT COALESCE(display_thumbnail_rel_path, '') FROM catalog_search_docs WHERE kind = ? AND entity_id = ?`,
		searchindex.KindSource, id.String(),
	).Scan(&path); err != nil {
		t.Fatal(err)
	}
	if path == "" {
		t.Fatal("thumbnail commit did not refresh the source search document")
	}
}

func newID(t *testing.T) []byte {
	t.Helper()
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	return id[:]
}

// sessionProject makes a project and closes it, so the test works through
// catalogsession.Do and can hold it with catalogsessiontest.Hold.
func sessionProject(t *testing.T) string {
	t.Helper()
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	dir := c.Dir()
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = catalogsession.CloseAll() })
	return dir
}

func tinyPNG(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile("testdata/tiny.png")
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestPlanStageCommit(t *testing.T) {
	t.Run("a File with no link is staged while the session is held, then committed", func(t *testing.T) {
		dir := sessionProject(t)
		var srcID []byte
		var job *Job
		err := catalogsession.Do(dir, func(c *database.Catalog) error {
			srcID = installSourceFile(t, c, tinyPNG(t), "image/png", "tiny.png")
			res, j, err := Plan(c, srcID, ThumbnailSpec())
			if err != nil || j == nil || res.Skipped || len(res.Link.ID) != 0 {
				t.Fatalf("Plan = %+v, %v, %v", res, j, err)
			}
			job = j
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}

		release := catalogsessiontest.Hold(t, dir)
		staged, err := job.Stage(dir)
		if err != nil {
			t.Fatal(err)
		}
		release()

		err = catalogsession.Do(dir, func(c *database.Catalog) error {
			stages := ingest.Stages{staged}
			defer stages.Settle(c)
			res, err := Commit(c, job)
			if err != nil {
				return err
			}
			if string(res.Link.SourceFileID) != string(srcID) || len(res.Link.DerivedFileID) != 16 {
				t.Fatalf("link %+v", res.Link)
			}
			derived, err := files.Lookup(c, res.Link.DerivedFileID)
			if err != nil || derived.MediaType != "image/jpeg" {
				t.Fatalf("derived %+v, %v", derived, err)
			}
			rel, err := files.StorageRelPath(derived.ChecksumSHA256, derived.MediaType)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(filepath.Join(c.Dir(), filepath.FromSlash(rel))); err != nil {
				t.Fatal(err)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("a File with a link plans no job, and a PDF is skipped", func(t *testing.T) {
		dir := sessionProject(t)
		err := catalogsession.Do(dir, func(c *database.Catalog) error {
			srcID := installSourceFile(t, c, tinyPNG(t), "image/png", "tiny.png")
			made, err := EnsureThumbnail(c, srcID)
			if err != nil {
				return err
			}
			res, job, err := Plan(c, srcID, ThumbnailSpec())
			if err != nil || job != nil || string(res.Link.ID) != string(made.Link.ID) {
				t.Fatalf("Plan = %+v, %v, %v", res, job, err)
			}
			pdfID := installSourceFile(t, c, []byte("%PDF-1.4"), "application/pdf", "doc.pdf")
			res, job, err = Plan(c, pdfID, ThumbnailSpec())
			if err != nil || job != nil || !res.Skipped {
				t.Fatalf("Plan(pdf) = %+v, %v, %v", res, job, err)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestCommitAfterChange(t *testing.T) {
	// plan makes a project with one PNG File, plans it, and stages the job.
	plan := func(t *testing.T) (dir string, srcID []byte, job *Job, staged *ingest.Staged) {
		t.Helper()
		dir = sessionProject(t)
		err := catalogsession.Do(dir, func(c *database.Catalog) error {
			srcID = installSourceFile(t, c, tinyPNG(t), "image/png", "tiny.png")
			_, j, err := Plan(c, srcID, ThumbnailSpec())
			job = j
			return err
		})
		if err != nil || job == nil {
			t.Fatalf("plan: %v, %v", job, err)
		}
		staged, err = job.Stage(dir)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(staged.Discard)
		return dir, srcID, job, staged
	}

	t.Run("a commit adds no audit revision", func(t *testing.T) {
		dir, _, job, staged := plan(t)
		err := catalogsession.Do(dir, func(c *database.Catalog) error {
			defer ingest.Stages{staged}.Settle(c)
			before := auditTotal(t, c)
			if _, err := Commit(c, job); err != nil {
				return err
			}
			if after := auditTotal(t, c); after != before {
				t.Fatalf("audit revisions %d, was %d", after, before)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("a link made meanwhile is returned and nothing unreferenced is left", func(t *testing.T) {
		dir, srcID, job, staged := plan(t)
		err := catalogsession.Do(dir, func(c *database.Catalog) error {
			made, err := EnsureThumbnail(c, srcID)
			if err != nil {
				return err
			}
			stages := ingest.Stages{staged}
			res, err := Commit(c, job)
			if err != nil {
				return err
			}
			stages.Settle(c)
			if string(res.Link.ID) != string(made.Link.ID) {
				t.Fatalf("Commit link %x, want %x", res.Link.ID, made.Link.ID)
			}
			var objects, temps int
			err = filepath.WalkDir(filepath.Join(c.Dir(), "objects"), func(p string, d os.DirEntry, err error) error {
				if err != nil || d.IsDir() {
					return err
				}
				if strings.HasPrefix(d.Name(), ".") {
					temps++
				} else {
					objects++
				}
				return nil
			})
			if err != nil {
				return err
			}
			// The source object and the one thumbnail; no temp.
			if objects != 2 || temps != 0 {
				t.Fatalf("objects = %d, temps = %d", objects, temps)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("a source File deleted meanwhile is invalid and leaves no link", func(t *testing.T) {
		dir, srcID, job, staged := plan(t)
		err := catalogsession.Do(dir, func(c *database.Catalog) error {
			defer ingest.Stages{staged}.Settle(c)
			db, err := c.DB()
			if err != nil {
				return err
			}
			if _, err := db.Exec(`DELETE FROM files WHERE id = ?`, srcID); err != nil {
				return err
			}
			if _, err := Commit(c, job); !errors.Is(err, ErrInvalid) {
				t.Fatalf("Commit err = %v, want ErrInvalid", err)
			}
			var n int
			if err := db.QueryRow(`SELECT COUNT(*) FROM file_derivatives`).Scan(&n); err != nil || n != 0 {
				t.Fatalf("links = %d, %v", n, err)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestStageAfterPurge(t *testing.T) {
	dir := sessionProject(t)
	var job *Job
	err := catalogsession.Do(dir, func(c *database.Catalog) error {
		srcID := installSourceFile(t, c, tinyPNG(t), "image/png", "tiny.png")
		_, j, err := Plan(c, srcID, ThumbnailSpec())
		job = j
		return err
	})
	if err != nil || job == nil {
		t.Fatalf("plan: %v, %v", job, err)
	}
	// What files.PurgeAfterCommit does after an Artifact delete.
	err = filepath.WalkDir(filepath.Join(dir, "objects"), func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		return os.Remove(p)
	})
	if err != nil {
		t.Fatal(err)
	}
	staged, err := job.Stage(dir)
	if err == nil {
		staged.Discard()
		t.Fatal("Stage succeeded with the source object gone")
	}
	if !Refused(err) {
		t.Fatalf("Stage err = %v, want Refused", err)
	}
}
