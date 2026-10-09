package derivatives

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"hash/crc32"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/filederivatives"
	"github.com/mendahu/provenencia/core/database/files"
	"github.com/mendahu/provenencia/core/database/searchindex"
	"github.com/mendahu/provenencia/core/derivatives/raster"
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

	binID := installSourceFile(t, c, []byte{0, 1, 2}, "application/octet-stream", "blob.bin")
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

// bombPNG returns a PNG signature + valid IHDR declaring w×h pixels; enough
// for DecodeConfig, tiny on disk — a header-declared decompression bomb.
func bombPNG(t *testing.T, w, h uint32) []byte {
	t.Helper()
	var buf bytes.Buffer
	buf.Write([]byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'})
	data := make([]byte, 13)
	binary.BigEndian.PutUint32(data[0:], w)
	binary.BigEndian.PutUint32(data[4:], h)
	data[8] = 8 // bit depth
	data[9] = 6 // color type RGBA
	var length [4]byte
	binary.BigEndian.PutUint32(length[:], 13)
	buf.Write(length[:])
	buf.WriteString("IHDR")
	buf.Write(data)
	crc := crc32.NewIEEE()
	crc.Write([]byte("IHDR"))
	crc.Write(data)
	var sum [4]byte
	binary.BigEndian.PutUint32(sum[:], crc.Sum32())
	buf.Write(sum[:])
	return buf.Bytes()
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
				return installFileRow(t, c, hex.EncodeToString(tinySum[:]), "image/png", maxSourceBytes+1)
			},
			wantIs: ErrUnprocessable,
		},
		{
			name: "object bytes do not match catalog checksum",
			srcID: func(t *testing.T) []byte {
				id := installSourceFile(t, c, bombPNG(t, 4, 4), "image/png", "swap.png")
				sum := sha256.Sum256(bombPNG(t, 4, 4))
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
			name: "pixel bomb png over decode budget",
			srcID: func(t *testing.T) []byte {
				return installSourceFile(t, c, bombPNG(t, 100000, 100000), "image/png", "bomb.png")
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

func TestEnsureThumbnailCommitsUnauditedAndRefreshesSearch(t *testing.T) {
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
