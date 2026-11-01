package ingest

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"

	"github.com/mendahu/provenencia/core/catalogsession/catalogsessiontest"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/files"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ingest/mediatypes"
	"github.com/mendahu/provenencia/core/objectstore"
	"github.com/mendahu/provenencia/core/raster/rastertest"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/writes"
)

var stageUserID = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

// stageUser upserts a user row and returns its id.
func stageUser(t *testing.T, c *database.Catalog) []byte {
	t.Helper()
	r, err := ref.Mint(ref.PrefixUser)
	if err != nil {
		t.Fatal(err)
	}
	if err := users.Upsert(c, stageUserID, "Jake", r); err != nil {
		t.Fatal(err)
	}
	return stageUserID
}

// stageWrite writes data to dir/name and returns its absolute path.
func stageWrite(t *testing.T, dir, name string, data []byte) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func stageCatalog(t *testing.T) *database.Catalog {
	t.Helper()
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}

func stagePixels(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: 9, A: 255})
		}
	}
	return img
}

func stageEncode(t *testing.T, enc func(io.Writer) error) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := enc(&buf); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func stagePNG(t *testing.T, w, h int) []byte {
	return stageEncode(t, func(wr io.Writer) error { return png.Encode(wr, stagePixels(w, h)) })
}

func stageJPEG(t *testing.T, w, h int) []byte {
	return stageEncode(t, func(wr io.Writer) error { return jpeg.Encode(wr, stagePixels(w, h), nil) })
}

// isStageTemp reports whether name is a staged temp, not a stored object.
func isStageTemp(name string) bool {
	return strings.HasPrefix(name, importTempPrefix) || strings.HasPrefix(name, stageTempPrefix)
}

// stageTemps lists the staged temps at the top of the project's objects folder.
func stageTemps(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(dir, "objects"))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		t.Fatal(err)
	}
	var out []string
	for _, e := range entries {
		if isStageTemp(e.Name()) {
			out = append(out, e.Name())
		}
	}
	return out
}

// objectFiles lists every stored object under the project's objects folder
// (staged temps are not objects; see stageTemps).
func objectFiles(t *testing.T, c *database.Catalog) []string {
	t.Helper()
	var out []string
	root := filepath.Join(c.Dir(), "objects")
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			return err
		}
		if !d.IsDir() && !isStageTemp(d.Name()) {
			out = append(out, p)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func fileRows(t *testing.T, c *database.Catalog) int {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM files`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// stageInsert runs one committing write that inserts every stage.
func stageInsert(t *testing.T, c *database.Catalog, stages ...*Staged) ([]files.File, []rowchange.Change) {
	t.Helper()
	type out struct {
		files   []files.File
		changes []rowchange.Change
	}
	res, _, err := writes.Run(c, writes.Op{Action: "create_file", UserID: stageUser(t, c)},
		func(tx *database.Tx) (out, []rowchange.Change, error) {
			var o out
			for _, s := range stages {
				f, ch, err := s.InsertTx(tx.Tx)
				if err != nil {
					return out{}, nil, err
				}
				o.files = append(o.files, f)
				o.changes = append(o.changes, ch...)
			}
			return o, o.changes, nil
		})
	if err != nil {
		t.Fatal(err)
	}
	return res.files, res.changes
}

func TestStageImageRefuses(t *testing.T) {
	dir := t.TempDir()
	noise := image.NewRGBA(image.Rect(0, 0, 64, 64))
	rng := rand.New(rand.NewSource(1))
	for i := range noise.Pix {
		noise.Pix[i] = uint8(rng.Intn(256))
	}
	bigPNG := stageEncode(t, func(w io.Writer) error { return png.Encode(w, noise) })
	if len(bigPNG) < 2048 {
		t.Fatalf("fixture too small: %d", len(bigPNG))
	}
	randomBytes := make([]byte, 600)
	rng.Read(randomBytes)

	real := stageWrite(t, dir, "real.png", stagePNG(t, 8, 8))
	link := filepath.Join(dir, "link.png")
	if err := os.Symlink(real, link); err != nil {
		t.Fatal(err)
	}
	subdir := filepath.Join(dir, "sub")
	if err := os.Mkdir(subdir, 0o755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name  string
		path  string
		limit int64
		want  error
	}{
		{"relative path", "rel.png", 0, ErrInvalid},
		{"missing path", filepath.Join(dir, "gone.png"), 0, ErrMissing},
		{"symlink", link, 0, ErrSymlink},
		{"directory", subdir, 0, ErrNotAFile},
		{"empty file", stageWrite(t, dir, "empty.png", nil), 0, ErrEmpty},
		{"over the cap", stageWrite(t, dir, "big.png", bigPNG), 1024, ErrTooLarge},
		{"tiff", stageWrite(t, dir, "a.tiff", stageEncode(t, func(w io.Writer) error { return tiff.Encode(w, stagePixels(8, 8), nil) })), 0, ErrUnsupportedImage},
		{"gif", stageWrite(t, dir, "a.gif", stageEncode(t, func(w io.Writer) error {
			return gif.Encode(w, image.NewPaletted(image.Rect(0, 0, 8, 8), color.Palette{color.Black, color.White}), nil)
		})), 0, ErrUnsupportedImage},
		{"bmp", stageWrite(t, dir, "a.bmp", stageEncode(t, func(w io.Writer) error { return bmp.Encode(w, stagePixels(8, 8)) })), 0, ErrUnsupportedImage},
		{"html named png", stageWrite(t, dir, "page.png", []byte("<html><body>hi</body></html>")), 0, ErrUnsupportedImage},
		{"random bytes", stageWrite(t, dir, "rand.png", randomBytes), 0, ErrUnsupportedImage},
		{"decompression bomb", stageWrite(t, dir, "bomb.png", rastertest.HeaderOnlyPNG(t, 100000, 100000)), 0, ErrUnsupportedImage},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.limit > 0 {
				old := maxImageBytes
				maxImageBytes = tt.limit
				defer func() { maxImageBytes = old }()
			}
			c := stageCatalog(t)
			s, err := StageImage(c.Dir(), tt.path)
			if !errors.Is(err, tt.want) {
				t.Fatalf("err = %v, want %v", err, tt.want)
			}
			if s != nil {
				t.Fatal("refused stage returned a value")
			}
			if got := objectFiles(t, c); len(got) != 0 {
				t.Fatalf("objects written: %v", got)
			}
		})
	}

	t.Run("valid png writes nothing", func(t *testing.T) {
		c := stageCatalog(t)
		s, err := StageImage(c.Dir(), real)
		if err != nil || s == nil {
			t.Fatalf("stage: %v", err)
		}
		if got := objectFiles(t, c); len(got) != 0 {
			t.Fatalf("objects written: %v", got)
		}
	})

	t.Run("stage formats are import-allowed", func(t *testing.T) {
		for _, f := range stageFormats {
			if !mediatypes.Allowed("image/" + f) {
				t.Fatalf("image/%s is not in the import allowlist", f)
			}
		}
	})
}

func TestStageImageReencodes(t *testing.T) {
	dir := t.TempDir()
	appended := append(stagePNG(t, 300, 200), bytes.Repeat([]byte("PK\x03\x04"), 256)...)
	tests := []struct {
		name         string
		data         []byte
		wantW, wantH int
		differs      bool
	}{
		{"png with trailing zip bytes", appended, 300, 200, true},
		{"wide jpeg scaled down", stageJPEG(t, 4000, 1000), 2048, 512, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := stageCatalog(t)
			path := stageWrite(t, dir, "in.bin", tt.data)
			s, err := StageImage(c.Dir(), path)
			if err != nil {
				t.Fatal(err)
			}
			defer s.Discard()
			fs, _ := stageInsert(t, c, s)
			obj, err := os.ReadFile(filepath.Join(c.Dir(), filepath.FromSlash(mustRel(t, fs[0]))))
			if err != nil {
				t.Fatal(err)
			}
			img, err := png.Decode(bytes.NewReader(obj))
			if err != nil {
				t.Fatal(err)
			}
			if b := img.Bounds(); b.Dx() != tt.wantW || b.Dy() != tt.wantH {
				t.Fatalf("size = %dx%d", b.Dx(), b.Dy())
			}
			if tt.differs && bytes.Equal(obj, tt.data) {
				t.Fatal("stored the caller's bytes")
			}
			iend := bytes.LastIndex(obj, []byte("IEND"))
			if iend < 0 || iend+8 != len(obj) {
				t.Fatalf("object does not end at IEND (IEND at %d of %d)", iend, len(obj))
			}
		})
	}
}

func mustRel(t *testing.T, f files.File) string {
	t.Helper()
	rel, err := files.StorageRelPath(f.ChecksumSHA256, f.MediaType)
	if err != nil {
		t.Fatal(err)
	}
	return rel
}

func TestStagedInsertTx(t *testing.T) {
	dir := t.TempDir()
	path := stageWrite(t, dir, "in.png", stagePNG(t, 40, 30))

	t.Run("one row, object, and create change", func(t *testing.T) {
		c := stageCatalog(t)
		s, err := StageImage(c.Dir(), path)
		if err != nil {
			t.Fatal(err)
		}
		defer s.Discard()
		fs, changes := stageInsert(t, c, s)
		if fileRows(t, c) != 1 {
			t.Fatalf("rows = %d", fileRows(t, c))
		}
		if fs[0].MediaType != "image/png" || fs[0].OriginalFilename != "" {
			t.Fatalf("%+v", fs[0])
		}
		if _, err := os.Stat(filepath.Join(c.Dir(), filepath.FromSlash(mustRel(t, fs[0])))); err != nil {
			t.Fatal(err)
		}
		if len(changes) != 1 || changes[0].EntityType != "file" || changes[0].Action != rowchange.ActionCreate {
			t.Fatalf("changes = %+v", changes)
		}
	})

	t.Run("same stage twice in one write", func(t *testing.T) {
		c := stageCatalog(t)
		s, err := StageImage(c.Dir(), path)
		if err != nil {
			t.Fatal(err)
		}
		defer s.Discard()
		fs, changes := stageInsert(t, c, s, s)
		if fileRows(t, c) != 1 || len(changes) != 1 || string(fs[0].ID) != string(fs[1].ID) {
			t.Fatalf("rows=%d changes=%d", fileRows(t, c), len(changes))
		}
	})

	t.Run("two stages of the same bytes in one write", func(t *testing.T) {
		c := stageCatalog(t)
		a, err := StageImage(c.Dir(), path)
		if err != nil {
			t.Fatal(err)
		}
		defer a.Discard()
		b, err := StageImage(c.Dir(), path)
		if err != nil {
			t.Fatal(err)
		}
		defer b.Discard()
		fs, changes := stageInsert(t, c, a, b)
		if fileRows(t, c) != 1 || len(changes) != 1 || string(fs[0].ID) != string(fs[1].ID) {
			t.Fatalf("rows=%d changes=%d", fileRows(t, c), len(changes))
		}
	})
}

func TestStagedSettle(t *testing.T) {
	dir := t.TempDir()
	path := stageWrite(t, dir, "in.png", stagePNG(t, 40, 30))
	errBoom := errors.New("boom")

	failingInsert := func(t *testing.T, c *database.Catalog, s *Staged) {
		t.Helper()
		_, _, err := writes.Run(c, writes.Op{Action: "create_file", UserID: stageUser(t, c)},
			func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
				if _, _, err := s.InsertTx(tx.Tx); err != nil {
					return struct{}{}, nil, err
				}
				return struct{}{}, nil, errBoom
			})
		if !errors.Is(err, errBoom) {
			t.Fatalf("err = %v", err)
		}
	}

	t.Run("failed write leaves no row and no object", func(t *testing.T) {
		c := stageCatalog(t)
		s, err := StageImage(c.Dir(), path)
		if err != nil {
			t.Fatal(err)
		}
		defer s.Discard()
		failingInsert(t, c, s)
		s.Settle(c)
		if fileRows(t, c) != 0 || len(objectFiles(t, c)) != 0 {
			t.Fatalf("rows=%d objects=%v", fileRows(t, c), objectFiles(t, c))
		}
	})

	t.Run("an existing row and object stay", func(t *testing.T) {
		c := stageCatalog(t)
		first, err := StageImage(c.Dir(), path)
		if err != nil {
			t.Fatal(err)
		}
		defer first.Discard()
		stageInsert(t, c, first)
		second, err := StageImage(c.Dir(), path)
		if err != nil {
			t.Fatal(err)
		}
		defer second.Discard()
		failingInsert(t, c, second)
		second.Settle(c)
		if fileRows(t, c) != 1 || len(objectFiles(t, c)) != 1 {
			t.Fatalf("rows=%d objects=%v", fileRows(t, c), objectFiles(t, c))
		}
	})

	t.Run("discard after commit, twice", func(t *testing.T) {
		c := stageCatalog(t)
		s, err := StageImage(c.Dir(), path)
		if err != nil {
			t.Fatal(err)
		}
		stageInsert(t, c, s)
		s.Settle(c)
		s.Settle(c)
		s.Discard()
		if fileRows(t, c) != 1 || len(objectFiles(t, c)) != 1 {
			t.Fatalf("rows=%d objects=%v", fileRows(t, c), objectFiles(t, c))
		}
	})
}

// stageProjectDir makes a project and closes it, so a test can hold it with
// catalogsessiontest.Hold.
func stageProjectDir(t *testing.T) string {
	t.Helper()
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	dir := c.Dir()
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestStageOutsideSession(t *testing.T) {
	src := stageWrite(t, t.TempDir(), "in.png", stagePNG(t, 40, 30))
	stagers := []struct {
		name  string
		stage func(dir string) (*Staged, error)
	}{
		{"StageFile", func(dir string) (*Staged, error) { return StageFile(dir, src) }},
		{"StageImage", func(dir string) (*Staged, error) { return StageImage(dir, src) }},
		{"StageBytes", func(dir string) (*Staged, error) { return StageBytes(dir, []byte("thumb"), "image/jpeg") }},
	}
	for _, st := range stagers {
		t.Run(st.name+" while the session is held", func(t *testing.T) {
			dir := stageProjectDir(t)
			release := catalogsessiontest.Hold(t, dir)
			s, err := st.stage(dir)
			if err != nil {
				t.Fatal(err)
			}
			if temps := stageTemps(t, dir); len(temps) != 1 {
				t.Fatalf("temps = %v, want one before release", temps)
			}
			release()
			s.Discard()
			if temps := stageTemps(t, dir); len(temps) != 0 {
				t.Fatalf("temps after Discard = %v", temps)
			}
		})
		t.Run(st.name+" in a folder with no catalog", func(t *testing.T) {
			dir := t.TempDir()
			s, err := st.stage(dir)
			if !errors.Is(err, database.ErrNotAProject) || s != nil {
				t.Fatalf("stage = %v, %v; want ErrNotAProject", s, err)
			}
			if _, err := os.Stat(filepath.Join(dir, "objects")); err == nil {
				t.Fatal("an objects folder was made")
			}
		})
		t.Run(st.name+" in a missing folder", func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), "gone")
			_, err := st.stage(dir)
			_, openErr := database.Open(dir)
			if !errors.Is(err, fs.ErrNotExist) || !errors.Is(openErr, fs.ErrNotExist) {
				t.Fatalf("stage err = %v, open err = %v", err, openErr)
			}
		})
	}
}

func TestStagedPurgeAndDiscard(t *testing.T) {
	path := stageWrite(t, t.TempDir(), "in.png", stagePNG(t, 40, 30))
	errBoom := errors.New("boom")

	t.Run("discard with no session leaves no temp", func(t *testing.T) {
		c := stageCatalog(t)
		s, err := StageImage(c.Dir(), path)
		if err != nil {
			t.Fatal(err)
		}
		if len(stageTemps(t, c.Dir())) != 1 {
			t.Fatalf("temps = %v", stageTemps(t, c.Dir()))
		}
		s.Discard()
		s.Discard()
		if len(stageTemps(t, c.Dir())) != 0 || len(objectFiles(t, c)) != 0 {
			t.Fatalf("temps=%v objects=%v", stageTemps(t, c.Dir()), objectFiles(t, c))
		}
	})

	t.Run("a refused write then settle leaves nothing", func(t *testing.T) {
		c := stageCatalog(t)
		s, err := StageImage(c.Dir(), path)
		if err != nil {
			t.Fatal(err)
		}
		_, _, err = writes.Run(c, writes.Op{Action: "create_file", UserID: stageUser(t, c)},
			func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
				if _, _, err := s.InsertTx(tx.Tx); err != nil {
					return struct{}{}, nil, err
				}
				return struct{}{}, nil, errBoom
			})
		if !errors.Is(err, errBoom) {
			t.Fatalf("err = %v", err)
		}
		if len(objectFiles(t, c)) != 1 {
			t.Fatalf("object not installed before settle: %v", objectFiles(t, c))
		}
		s.Settle(c)
		if len(objectFiles(t, c)) != 0 || len(stageTemps(t, c.Dir())) != 0 || fileRows(t, c) != 0 {
			t.Fatalf("objects=%v temps=%v rows=%d", objectFiles(t, c), stageTemps(t, c.Dir()), fileRows(t, c))
		}
	})

	t.Run("a committed write keeps its object", func(t *testing.T) {
		c := stageCatalog(t)
		s, err := StageImage(c.Dir(), path)
		if err != nil {
			t.Fatal(err)
		}
		stageInsert(t, c, s)
		s.Settle(c)
		s.Discard()
		if len(objectFiles(t, c)) != 1 || fileRows(t, c) != 1 {
			t.Fatalf("objects=%v rows=%d", objectFiles(t, c), fileRows(t, c))
		}
	})

	t.Run("CommitFile on a stage Go made is invalid", func(t *testing.T) {
		c := stageCatalog(t)
		s, err := StageBytes(c.Dir(), []byte("thumb"), "image/jpeg")
		if err != nil {
			t.Fatal(err)
		}
		defer s.Discard()
		if _, err := s.CommitFile(c, stageUser(t, c)); !errors.Is(err, ErrInvalid) {
			t.Fatalf("err = %v, want ErrInvalid", err)
		}
	})
}

func TestPurgeLeavesTemps(t *testing.T) {
	c := stageCatalog(t)
	path := stageWrite(t, t.TempDir(), "in.png", stagePNG(t, 40, 30))
	s, err := StageImage(c.Dir(), path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Discard()

	// An unrelated object no row names.
	data := []byte("unrelated bytes")
	sum := sha256.Sum256(data)
	checksum := hex.EncodeToString(sum[:])
	rel, err := files.StorageRelPath(checksum, "image/jpeg")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := objectstore.WriteIfAbsent(c.Dir(), rel, checksum, data); err != nil {
		t.Fatal(err)
	}
	files.PurgeUnreferenced(c, []files.Object{{ChecksumSHA256: checksum, MediaType: "image/jpeg"}})

	if len(objectFiles(t, c)) != 0 {
		t.Fatalf("unrelated object kept: %v", objectFiles(t, c))
	}
	if len(stageTemps(t, c.Dir())) != 1 {
		t.Fatalf("purge removed the staged temp: %v", stageTemps(t, c.Dir()))
	}
	stageInsert(t, c, s)
	if len(objectFiles(t, c)) != 1 || fileRows(t, c) != 1 {
		t.Fatalf("objects=%v rows=%d", objectFiles(t, c), fileRows(t, c))
	}
}
