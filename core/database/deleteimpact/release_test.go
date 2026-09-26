package deleteimpact

import (
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/filederivatives"
	"github.com/mendahu/provenencia/core/database/files"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
)

func TestCountFilePointersIfUnused(t *testing.T) {
	c, userID := testCatalog(t)
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
		Key: "book", Origin: sourcetypes.OriginProvenencia, Label: "Book",
	})
	if err != nil {
		t.Fatal(err)
	}
	src, err := sources.Create(c, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Deed"})
	if err != nil {
		t.Fatal(err)
	}
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

	art1, err := artifacts.Create(c, userID, artifacts.CreateInput{
		SourceID: src.ID, FileID: fileA, Label: "Scan 1",
	})
	if err != nil {
		t.Fatal(err)
	}
	art2, err := artifacts.Create(c, userID, artifacts.CreateInput{
		SourceID: src.ID, FileID: fileA, Label: "Scan 2",
	})
	if err != nil {
		t.Fatal(err)
	}

	tx, err = db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	n, err := CountFilePointers(tx, fileA)
	if err != nil || n != 2 {
		t.Fatalf("shared count %d %v", n, err)
	}
	if _, err := tx.Exec(`DELETE FROM artifacts WHERE id = ?`, art1.ID); err != nil {
		t.Fatal(err)
	}
	n, err = CountFilePointers(tx, fileA)
	if err != nil || n != 1 {
		t.Fatalf("after one delete %d %v", n, err)
	}
	if _, err := tx.Exec(`DELETE FROM artifacts WHERE id = ?`, art2.ID); err != nil {
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
	art3, err := artifacts.Create(c, userID, artifacts.CreateInput{
		SourceID: src.ID, FileID: fileB, Label: "Lone",
	})
	if err != nil {
		t.Fatal(err)
	}
	tx, err = db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	n, err = CountFilePointers(tx, fileB)
	if err != nil || n != 1 {
		t.Fatalf("lone+deriv artifacts %d %v", n, err)
	}
	if _, err := tx.Exec(`DELETE FROM artifacts WHERE id = ?`, art3.ID); err != nil {
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

func testCatalog(t *testing.T) (*database.Catalog, []byte) {
	t.Helper()
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	userID := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	r, err := ref.Mint(ref.PrefixUser)
	if err != nil {
		t.Fatal(err)
	}
	if err := users.Upsert(c, userID, "Tester", r); err != nil {
		t.Fatal(err)
	}
	return c, userID
}
