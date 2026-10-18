package deleteimpact

import (
	"database/sql"
	"errors"
	"github.com/mendahu/provenencia/core/database/catalogmodel"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/filederivatives"
	"github.com/mendahu/provenencia/core/database/files"
)

func TestCountFilePointersIfUnused(t *testing.T) {
	c := testCatalogInternal(t)
	typeID := insertSourceType(t, c, "book", "Book")
	srcID := insertSource(t, c, typeID, "SRC-AAAAA", "Deed")
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}

	fileA := newFileID(t)
	const sum = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	insertFile(t, tx, fileA, sum, "a.jpg", 4)
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

	fileB := newFileID(t)
	const sumB = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	deriv := newFileID(t)
	const sumD = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	linkID := newFileID(t)
	tx, err = db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	insertFile(t, tx, fileB, sumB, "b.jpg", 4)
	insertFile(t, tx, deriv, sumD, "b-thumb.jpg", 2)
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
	if _, err := releaseFile(tx, productionPlan(), fileB); err != nil {
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

	// A File is a pointer target through derived_file_id, not source_file_id.
	other := newFileID(t)
	const sumOther = "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"
	otherDeriv := newFileID(t)
	const sumOtherDeriv = "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
	tx, err = db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	insertFile(t, tx, other, sumOther, "o.jpg", 4)
	insertFile(t, tx, otherDeriv, sumOtherDeriv, "o-thumb.jpg", 2)
	if err := filederivatives.Insert(tx, filederivatives.Link{
		ID: newFileID(t), SourceFileID: other, DerivedFileID: otherDeriv, DerivativeType: filederivatives.TypeThumbnail,
	}); err != nil {
		t.Fatal(err)
	}
	if n, err := CountFilePointers(tx, otherDeriv); err != nil || n != 1 {
		t.Fatalf("derived end %d %v", n, err)
	}
	if n, err := CountFilePointers(tx, other); err != nil || n != 0 {
		t.Fatalf("source end %d %v", n, err)
	}

	// A table the test declares as pointing at files counts too.
	if _, err := tx.Exec(`CREATE TABLE test_values (id BLOB PRIMARY KEY, file_id BLOB)`); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`INSERT INTO test_values (id, file_id) VALUES (?, ?)`, newFileID(t), other); err != nil {
		t.Fatal(err)
	}
	testFKs := append(append([]catalogmodel.FK(nil), catalogmodel.FKs...),
		catalogmodel.FK{From: "test_values", Column: "file_id", To: "files", OnDelete: "NO ACTION", Bucket: catalogmodel.BucketOwnedOutbound})
	if n, err := countFilePointers(tx, ownedPlanFrom(testFKs, catalogmodel.Tables), other); err != nil || n != 1 {
		t.Fatalf("test table pointer %d %v", n, err)
	}
	if n, err := countFilePointers(tx, ownedPlanFrom(catalogmodel.FKs, catalogmodel.Tables), other); err != nil || n != 0 {
		t.Fatalf("production plan ignores the test table %d %v", n, err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
}

func TestListedSnippet(t *testing.T) {
	if got := listedSnippet("  John\nHartley  ", 80); got != "John Hartley" {
		t.Fatalf("collapse %q", got)
	}
	if got := listedSnippet("", 80); got != "" {
		t.Fatalf("empty %q", got)
	}
	long := strings.Repeat("a", 90)
	got := listedSnippet(long, 80)
	if got != strings.Repeat("a", 80)+"…" {
		t.Fatalf("trunc %q", got)
	}
}

// fileFixture inserts a File with the given checksum and returns its id.
func fileFixture(t *testing.T, tx *sql.Tx, sum string) []byte {
	t.Helper()
	id := newFileID(t)
	insertFile(t, tx, id, sum, sum[:4]+".jpg", 4)
	return id
}

func linkThumbnail(t *testing.T, tx *sql.Tx, source, derived []byte) {
	t.Helper()
	if err := filederivatives.Insert(tx, filederivatives.Link{
		ID: newFileID(t), SourceFileID: source, DerivedFileID: derived, DerivativeType: filederivatives.TypeThumbnail,
	}); err != nil {
		t.Fatal(err)
	}
}

// deleteArtifact deletes an Artifact row the way artifacts.Delete does:
// snapshot, DELETE, release.
func deleteArtifact(t *testing.T, tx *sql.Tx, artID []byte) []files.Object {
	t.Helper()
	snap, err := SnapshotOwned(tx, catalogmodel.KindArtifact, artID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`DELETE FROM artifacts WHERE id = ?`, artID); err != nil {
		t.Fatal(err)
	}
	objs, err := releaseSnapshot(tx, snap)
	if err != nil {
		t.Fatal(err)
	}
	return objs
}

func countRows(t *testing.T, tx *sql.Tx, query string, args ...any) int {
	t.Helper()
	var n int
	if err := tx.QueryRow(query, args...).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestReleaseKeepsSharedDerived(t *testing.T) {
	c := testCatalogInternal(t)
	typeID := insertSourceType(t, c, "book", "Book")
	srcID := insertSource(t, c, typeID, "SRC-CCCCC", "Deed")
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	sumA := strings.Repeat("a1", 32)
	fileA := fileFixture(t, tx, sumA)
	fileB := fileFixture(t, tx, strings.Repeat("b1", 32))
	shared := fileFixture(t, tx, strings.Repeat("d1", 32))
	linkThumbnail(t, tx, fileA, shared)
	linkThumbnail(t, tx, fileB, shared)
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	artA := insertArtifact(t, c, srcID, fileA, "ART-AAAAA", "A")
	insertArtifact(t, c, srcID, fileB, "ART-BBBBB", "B")

	tx, err = db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	objs := deleteArtifact(t, tx, artA)
	if len(objs) != 1 || objs[0].ChecksumSHA256 != sumA {
		t.Fatalf("objects %+v", objs)
	}
	if n := countRows(t, tx, `SELECT COUNT(*) FROM files WHERE id = ?`, fileA); n != 0 {
		t.Fatalf("A left %d", n)
	}
	if n := countRows(t, tx, `SELECT COUNT(*) FROM file_derivatives WHERE source_file_id = ?`, fileA); n != 0 {
		t.Fatalf("A link left %d", n)
	}
	if n := countRows(t, tx, `SELECT COUNT(*) FROM files WHERE id = ?`, shared); n != 1 {
		t.Fatalf("shared thumbnail deleted")
	}
	if n := countRows(t, tx, `SELECT COUNT(*) FROM file_derivatives WHERE source_file_id = ? AND derived_file_id = ?`, fileB, shared); n != 1 {
		t.Fatalf("B link gone")
	}
}

func TestReleaseReportsReleasedObjects(t *testing.T) {
	sumA := strings.Repeat("a2", 32)
	sumT := strings.Repeat("c2", 32)
	tests := []struct {
		name      string
		withFile  bool
		withThumb bool
		extra     bool // a second Artifact on the same File
		want      []string
	}{
		{name: "lone artifact with thumbnail", withFile: true, withThumb: true, want: []string{sumA, sumT}},
		{name: "second artifact keeps the file", withFile: true, withThumb: true, extra: true, want: nil},
		{name: "fileless artifact", want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := testCatalogInternal(t)
			typeID := insertSourceType(t, c, "book", "Book")
			srcID := insertSource(t, c, typeID, "SRC-DDDDD", "Deed")
			db, err := c.DB()
			if err != nil {
				t.Fatal(err)
			}
			tx, err := db.Begin()
			if err != nil {
				t.Fatal(err)
			}
			var fileID []byte
			if tt.withFile {
				fileID = fileFixture(t, tx, sumA)
				if tt.withThumb {
					linkThumbnail(t, tx, fileID, fileFixture(t, tx, sumT))
				}
			}
			if err := tx.Commit(); err != nil {
				t.Fatal(err)
			}
			art := insertArtifact(t, c, srcID, fileID, "ART-AAAAA", "A")
			if tt.extra {
				insertArtifact(t, c, srcID, fileID, "ART-BBBBB", "B")
			}
			tx, err = db.Begin()
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback() }()
			got := deleteArtifact(t, tx, art)
			var sums []string
			for _, o := range got {
				sums = append(sums, o.ChecksumSHA256)
			}
			if strings.Join(sums, ",") != strings.Join(tt.want, ",") {
				t.Fatalf("objects %v want %v", sums, tt.want)
			}
			if tt.extra {
				if n := countRows(t, tx, `SELECT COUNT(*) FROM files`); n != 2 {
					t.Fatalf("files changed: %d", n)
				}
				if n := countRows(t, tx, `SELECT COUNT(*) FROM file_derivatives`); n != 1 {
					t.Fatalf("file_derivatives changed: %d", n)
				}
			}
		})
	}
}

func TestSnapshotOwnedRequiresParent(t *testing.T) {
	c := testCatalogInternal(t)
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()

	missing := make([]byte, 16)
	missing[15] = 1
	if _, err := SnapshotOwned(tx, catalogmodel.KindArtifact, missing); !errors.Is(err, ErrInvalid) {
		t.Fatalf("snapshot missing parent %v", err)
	}
	if _, err := releaseOwned(tx, catalogmodel.KindArtifact, missing); !errors.Is(err, ErrInvalid) {
		t.Fatalf("release missing parent %v", err)
	}
}

func insertSourceType(t *testing.T, c *database.Catalog, key, label string) []byte {
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
		`INSERT INTO source_types (id, key, origin, label) VALUES (?, ?, 'provenencia', ?)`,
		id[:], key, label,
	); err != nil {
		t.Fatal(err)
	}
	return id[:]
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

func newFileID(t *testing.T) []byte {
	t.Helper()
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	return id[:]
}

func insertFile(t *testing.T, tx *sql.Tx, id []byte, checksum, name string, size int) {
	t.Helper()
	if _, err := tx.Exec(
		`INSERT INTO files (id, checksum_sha256, original_filename, media_type, byte_size) VALUES (?, ?, ?, 'image/jpeg', ?)`,
		id, checksum, name, size,
	); err != nil {
		t.Fatal(err)
	}
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

// ownedPlanFixture is a plan over two temp tables, test_parent.value_id ->
// test_values.file_id -> files, declared as owned-outbound on top of the
// production model. It never touches catalogmodel's own slices.
type ownedPlanFixture struct {
	t    *testing.T
	tx   *sql.Tx
	plan *ownedPlan
}

func newOwnedPlanFixture(t *testing.T) *ownedPlanFixture {
	t.Helper()
	c := testCatalogInternal(t)
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = tx.Rollback() })
	for _, ddl := range []string{
		`CREATE TABLE test_parent (id BLOB PRIMARY KEY, value_id BLOB)`,
		`CREATE TABLE test_values (id BLOB PRIMARY KEY, file_id BLOB)`,
	} {
		if _, err := tx.Exec(ddl); err != nil {
			t.Fatal(err)
		}
	}
	fks := append(append([]catalogmodel.FK(nil), catalogmodel.FKs...),
		catalogmodel.FK{From: "test_parent", Column: "value_id", To: "test_values", OnDelete: "NO ACTION", Bucket: catalogmodel.BucketOwnedOutbound},
		catalogmodel.FK{From: "test_values", Column: "file_id", To: "files", OnDelete: "NO ACTION", Bucket: catalogmodel.BucketOwnedOutbound},
	)
	tables := append(append([]catalogmodel.Table(nil), catalogmodel.Tables...),
		catalogmodel.Table{Name: "test_parent", PK: "id"},
		catalogmodel.Table{Name: "test_values", PK: "id", Bucket: catalogmodel.BucketOwnedOutbound},
	)
	return &ownedPlanFixture{t: t, tx: tx, plan: ownedPlanFrom(fks, tables)}
}

func (f *ownedPlanFixture) value(file []byte) []byte {
	f.t.Helper()
	id := newFileID(f.t)
	if _, err := f.tx.Exec(`INSERT INTO test_values (id, file_id) VALUES (?, ?)`, id, file); err != nil {
		f.t.Fatal(err)
	}
	return id
}

func (f *ownedPlanFixture) parent(value []byte) []byte {
	f.t.Helper()
	id := newFileID(f.t)
	if _, err := f.tx.Exec(`INSERT INTO test_parent (id, value_id) VALUES (?, ?)`, id, value); err != nil {
		f.t.Fatal(err)
	}
	return id
}

// drop deletes a parent the way a delete does: snapshot, DELETE, release.
func (f *ownedPlanFixture) drop(parent []byte) []files.Object {
	f.t.Helper()
	snap, err := snapshotOwnedRow(f.tx, f.plan, "test_parent", parent)
	if err != nil {
		f.t.Fatal(err)
	}
	if _, err := f.tx.Exec(`DELETE FROM test_parent WHERE id = ?`, parent); err != nil {
		f.t.Fatal(err)
	}
	objs, err := releaseOwnedSnap(f.tx, f.plan, snap)
	if err != nil {
		f.t.Fatal(err)
	}
	return objs
}

func (f *ownedPlanFixture) has(table string, id []byte) bool {
	f.t.Helper()
	return countRows(f.t, f.tx, `SELECT COUNT(*) FROM `+table+` WHERE id = ?`, id) == 1
}

func TestOwnedPlanRelease(t *testing.T) {
	sum := func(c string) string { return strings.Repeat(c, 64) }
	tests := []struct {
		name string
		run  func(t *testing.T, f *ownedPlanFixture)
	}{
		{"releasing a parent deletes its value row and its unused File", func(t *testing.T, f *ownedPlanFixture) {
			file := fileFixture(t, f.tx, sum("1"))
			val := f.value(file)
			objs := f.drop(f.parent(val))
			if len(objs) != 1 || objs[0].ChecksumSHA256 != sum("1") {
				t.Fatalf("objects %+v", objs)
			}
			if f.has("test_values", val) || f.has("files", file) {
				t.Fatal("value row or File left")
			}
		}},
		{"a File shared through two value rows stays until both parents go", func(t *testing.T, f *ownedPlanFixture) {
			file := fileFixture(t, f.tx, sum("2"))
			v1, v2 := f.value(file), f.value(file)
			p1, p2 := f.parent(v1), f.parent(v2)
			if objs := f.drop(p1); len(objs) != 0 {
				t.Fatalf("first release reported %+v", objs)
			}
			if f.has("test_values", v1) || !f.has("files", file) || !f.has("test_values", v2) {
				t.Fatal("first release left the wrong rows")
			}
			if objs := f.drop(p2); len(objs) != 1 {
				t.Fatalf("second release reported %+v", objs)
			}
			if f.has("files", file) {
				t.Fatal("File left after both parents")
			}
		}},
		{"a replaced value row and its unused File go, the new ones stay", func(t *testing.T, f *ownedPlanFixture) {
			oldFile := fileFixture(t, f.tx, sum("3"))
			newFile := fileFixture(t, f.tx, sum("4"))
			oldVal, newVal := f.value(oldFile), f.value(newFile)
			parent := f.parent(oldVal)
			if _, err := f.tx.Exec(`UPDATE test_parent SET value_id = ? WHERE id = ?`, newVal, parent); err != nil {
				t.Fatal(err)
			}
			objs, err := releaseReplaced(f.tx, f.plan, "test_parent.value_id", oldVal)
			if err != nil {
				t.Fatal(err)
			}
			if len(objs) != 1 || objs[0].ChecksumSHA256 != sum("3") {
				t.Fatalf("objects %+v", objs)
			}
			if f.has("test_values", oldVal) || f.has("files", oldFile) {
				t.Fatal("old rows left")
			}
			if !f.has("test_values", newVal) || !f.has("files", newFile) {
				t.Fatal("new rows removed")
			}
			if objs, err := releaseReplaced(f.tx, f.plan, "test_parent.value_id", nil); err != nil || len(objs) != 0 {
				t.Fatalf("empty old id %v %v", objs, err)
			}
			if _, err := releaseReplaced(f.tx, f.plan, "test_parent.nope", oldVal); !errors.Is(err, ErrInvalid) {
				t.Fatalf("unknown column %v", err)
			}
		}},
		{"a NULL value column releases nothing", func(t *testing.T, f *ownedPlanFixture) {
			if objs := f.drop(f.parent(nil)); len(objs) != 0 {
				t.Fatalf("objects %+v", objs)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { tt.run(t, newOwnedPlanFixture(t)) })
	}
}
