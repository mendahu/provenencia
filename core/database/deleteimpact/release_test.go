package deleteimpact

import (
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/filederivatives"
	"github.com/mendahu/provenencia/core/database/files"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
)

func TestCountFilePointersIfUnused(t *testing.T) {
	c := testCatalogInternal(t)
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
		Key: "book", Origin: sourcetypes.OriginProvenencia, Label: "Book",
	})
	if err != nil {
		t.Fatal(err)
	}
	srcID := insertSource(t, c, typeID, "SRC-AAAAA", "Deed")
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}

	fileA, err := files.NewID()
	if err != nil {
		t.Fatal(err)
	}
	const sum = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	if err := files.Insert(tx, files.File{
		ID: fileA, ChecksumSHA256: sum, OriginalFilename: "a.jpg", MediaType: "image/jpeg", ByteSize: 4,
	}); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	art1 := insertArtifact(t, c, srcID, fileA, "ART-AAAAA", "Scan 1")
	art2 := insertArtifact(t, c, srcID, fileA, "ART-BBBBB", "Scan 2")

	tx, err = db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	n, err := CountFilePointers(tx, fileA)
	if err != nil || n != 2 {
		t.Fatalf("shared count %d %v", n, err)
	}
	if _, err := tx.Exec(`DELETE FROM artifacts WHERE id = ?`, art1); err != nil {
		t.Fatal(err)
	}
	n, err = CountFilePointers(tx, fileA)
	if err != nil || n != 1 {
		t.Fatalf("after one delete %d %v", n, err)
	}
	if _, err := tx.Exec(`DELETE FROM artifacts WHERE id = ?`, art2); err != nil {
		t.Fatal(err)
	}
	n, err = CountFilePointers(tx, fileA)
	if err != nil || n != 0 {
		t.Fatalf("after both %d %v", n, err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}

	fileB, err := files.NewID()
	if err != nil {
		t.Fatal(err)
	}
	const sumB = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	deriv, err := files.NewID()
	if err != nil {
		t.Fatal(err)
	}
	const sumD = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	linkID, err := files.NewID()
	if err != nil {
		t.Fatal(err)
	}
	tx, err = db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	if err := files.Insert(tx, files.File{
		ID: fileB, ChecksumSHA256: sumB, OriginalFilename: "b.jpg", MediaType: "image/jpeg", ByteSize: 4,
	}); err != nil {
		t.Fatal(err)
	}
	if err := files.Insert(tx, files.File{
		ID: deriv, ChecksumSHA256: sumD, OriginalFilename: "b-thumb.jpg", MediaType: "image/jpeg", ByteSize: 2,
	}); err != nil {
		t.Fatal(err)
	}
	if err := filederivatives.Insert(tx, filederivatives.Link{
		ID: linkID, SourceFileID: fileB, DerivedFileID: deriv, DerivativeType: filederivatives.TypeThumbnail,
	}); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	art3 := insertArtifact(t, c, srcID, fileB, "ART-CCCCC", "Lone")
	tx, err = db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	n, err = CountFilePointers(tx, fileB)
	if err != nil || n != 1 {
		t.Fatalf("lone+deriv artifacts %d %v", n, err)
	}
	if _, err := tx.Exec(`DELETE FROM artifacts WHERE id = ?`, art3); err != nil {
		t.Fatal(err)
	}
	if err := releaseFileIfUnused(tx, fileB); err != nil {
		t.Fatal(err)
	}
	var left int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM files WHERE id IN (?, ?)`, fileB, deriv).Scan(&left); err != nil {
		t.Fatal(err)
	}
	if left != 0 {
		t.Fatalf("files left=%d", left)
	}
	if err := tx.QueryRow(`SELECT COUNT(*) FROM file_derivatives WHERE source_file_id = ?`, fileB).Scan(&left); err != nil {
		t.Fatal(err)
	}
	if left != 0 {
		t.Fatalf("derivatives left=%d", left)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
}

func TestCollectFileObjects(t *testing.T) {
	c := testCatalogInternal(t)
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
		Key: "book", Origin: sourcetypes.OriginProvenencia, Label: "Book",
	})
	if err != nil {
		t.Fatal(err)
	}
	srcID := insertSource(t, c, typeID, "SRC-BBBBB", "Census")
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	empty, err := CollectFileObjects(tx, OwnedSnapshot{cols: map[string][]byte{}})
	if err != nil || len(empty) != 0 {
		t.Fatalf("fileless %d %v", len(empty), err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}

	fileID, err := files.NewID()
	if err != nil {
		t.Fatal(err)
	}
	derivID, err := files.NewID()
	if err != nil {
		t.Fatal(err)
	}
	linkID, err := files.NewID()
	if err != nil {
		t.Fatal(err)
	}
	const sumPri = "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	const sumDer = "1111111111111111111111111111111111111111111111111111111111111111"
	tx, err = db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	if err := files.Insert(tx, files.File{
		ID: fileID, ChecksumSHA256: sumPri, OriginalFilename: "scan.jpg", MediaType: "image/jpeg", ByteSize: 8,
	}); err != nil {
		t.Fatal(err)
	}
	if err := files.Insert(tx, files.File{
		ID: derivID, ChecksumSHA256: sumDer, OriginalFilename: "scan-thumb.jpg", MediaType: "image/jpeg", ByteSize: 2,
	}); err != nil {
		t.Fatal(err)
	}
	if err := filederivatives.Insert(tx, filederivatives.Link{
		ID: linkID, SourceFileID: fileID, DerivedFileID: derivID, DerivativeType: filederivatives.TypeThumbnail,
	}); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	artID := insertArtifact(t, c, srcID, fileID, "ART-DDDDD", "Scan")

	tx, err = db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	snap, err := SnapshotOwned(tx, KindArtifact, artID)
	if err != nil {
		t.Fatal(err)
	}
	got, err := CollectFileObjects(tx, snap)
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("objects %d", len(got))
	}
	if got[0].ChecksumSHA256 != sumPri || got[0].MediaType != "image/jpeg" {
		t.Fatalf("primary %+v", got[0])
	}
	if got[1].ChecksumSHA256 != sumDer || got[1].MediaType != "image/jpeg" {
		t.Fatalf("derived %+v", got[1])
	}
}

func insertSource(t *testing.T, c *database.Catalog, typeID []byte, sourceRef, title string) []byte {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(
		`INSERT INTO sources (id, ref, source_type_id, title) VALUES (?, ?, ?, ?)`,
		id[:], sourceRef, typeID, title,
	); err != nil {
		t.Fatal(err)
	}
	return id[:]
}

func insertArtifact(t *testing.T, c *database.Catalog, sourceID, fileID []byte, artRef, label string) []byte {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(
		`INSERT INTO artifacts (id, ref, source_id, file_id, label) VALUES (?, ?, ?, ?, ?)`,
		id[:], artRef, sourceID, fileID, label,
	); err != nil {
		t.Fatal(err)
	}
	return id[:]
}

func testCatalogInternal(t *testing.T) *database.Catalog {
	t.Helper()
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}
