package ingest

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"io"
	"path/filepath"
	"strings"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/files"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/depictionvalue"
	"github.com/mendahu/provenencia/core/ingest/mediatypes"
	"github.com/mendahu/provenencia/core/objectstore"
	"github.com/mendahu/provenencia/core/raster"
)

// imageMaxBytes is the largest image file StageImage reads (32 MiB).
const imageMaxBytes int64 = 32 << 20

// maxImageBytes is overridden in tests.
var maxImageBytes = imageMaxBytes

// stageMaxEdge is the longest edge, in pixels, of a stored staged image.
const stageMaxEdge = 2048

// stageFormats is the decode allowlist for staged images. It must stay inside
// mediatypes.Allowed.
var stageFormats = []string{"png", "jpeg"}

// Temp name prefixes in the objects folder. Purge builds paths from checksums,
// so it never touches either.
const (
	importTempPrefix = ".ingest-"
	stageTempPrefix  = ".stage-"
)

// Staged is an object whose bytes are already on disk in a temp file in the
// project's objects folder, checked and ready to install, but not yet stored:
// no catalog row names it and no object sits at its storage path. The three
// constructors do all the file work (copy, hash, decode, encode) and need no
// catalog session; install and purge run only inside catalogsession.Do, so a
// purge and an import of the same bytes never interleave.
//
//   - StageFile: a file the researcher picked, copied as is (CommitFile).
//   - StageImage: an image the app rendered, re-encoded as a PNG (InsertTx).
//   - StageBytes: bytes Go made, such as a thumbnail (InsertTx).
//
// A stage serves one write: after that write fails, call Settle (inside the
// session) or Discard (outside it) and stage again. Defer Discard outside the
// session and Settle first inside it; withStaged in api/ffi/handlers does both.
type Staged struct {
	dir       string
	tmp       *objectstore.Temp
	checksum  string
	mediaType string
	byteSize  int64
	rel       string
	filename  string // set by StageFile only
	installed bool
	wrote     bool
	file      files.File
	haveFile  bool
}

// StageFile applies the import path checks to absPath, copies it to a temp
// file in the project's objects folder while it hashes and sniffs it, and
// resolves its media type. It reads no catalog row. Never call it inside a
// catalog session.
func StageFile(projectDir, absPath string) (*Staged, error) {
	if err := database.RequireProject(projectDir); err != nil {
		return nil, err
	}
	absPath = strings.TrimSpace(absPath)
	src, err := openChecked(absPath, maxBytes)
	if err != nil {
		return nil, err
	}
	defer src.Close()

	tmp, checksum, byteSize, sniffed, err := streamSource(src, projectDir)
	if err != nil {
		return nil, mapObjectsErr(err)
	}
	filename := sanitizeFilename(filepath.Base(absPath))
	mediaType, reason, ok := mediatypes.Resolve(sniffed, filename)
	if !ok {
		tmp.Discard()
		return nil, errForReject(reason, sniffed)
	}
	rel, err := files.StorageRelPath(checksum, mediaType)
	if err != nil {
		tmp.Discard()
		return nil, err
	}
	return &Staged{dir: projectDir, tmp: tmp, checksum: checksum, mediaType: mediaType, byteSize: byteSize, rel: rel, filename: filename}, nil
}

// StageImage applies the import path checks to absPath, decodes it as PNG or
// JPEG within the byte limit and pixel budget, re-encodes it as a PNG of at
// most 2048 pixels on the long edge, and writes that PNG to a temp file in the
// objects folder. It reads no catalog row. Never call it inside a catalog
// session.
func StageImage(projectDir, absPath string) (*Staged, error) {
	if err := database.RequireProject(projectDir); err != nil {
		return nil, err
	}
	absPath = strings.TrimSpace(absPath)
	f, err := openChecked(absPath, maxImageBytes)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	release := raster.Slot()
	defer release()

	raw, err := io.ReadAll(io.LimitReader(f, maxImageBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(raw)) > maxImageBytes {
		return nil, ErrTooLarge.WithParams(formatByteSize(int64(len(raw))))
	}
	out, err := raster.EncodePNG(raw, raster.Options{MaxEdge: stageMaxEdge, Formats: stageFormats})
	if err != nil {
		return nil, ErrUnsupportedImage
	}
	return stageData(projectDir, out, depictionvalue.ImageMediaType)
}

// StageBytes writes data, which Go made, to a temp file in the objects folder.
// It reads no catalog row. Never call it inside a catalog session.
func StageBytes(projectDir string, data []byte, mediaType string) (*Staged, error) {
	if err := database.RequireProject(projectDir); err != nil {
		return nil, err
	}
	return stageData(projectDir, data, mediaType)
}

func stageData(projectDir string, data []byte, mediaType string) (*Staged, error) {
	sum := sha256.Sum256(data)
	checksum := hex.EncodeToString(sum[:])
	rel, err := files.StorageRelPath(checksum, mediaType)
	if err != nil {
		return nil, err
	}
	tmp, err := objectstore.CreateTemp(projectDir, stageTempPrefix)
	if err != nil {
		return nil, mapObjectsErr(err)
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Discard()
		return nil, err
	}
	if err := tmp.Close(); err != nil {
		tmp.Discard()
		return nil, err
	}
	return &Staged{dir: projectDir, tmp: tmp, checksum: checksum, mediaType: mediaType, byteSize: int64(len(data)), rel: rel}, nil
}

// install moves the temp to the object's storage path unless a matching object
// is already there, and records whether this stage left new bytes. It runs
// inside the catalog session, once.
func (s *Staged) install() error {
	if s.installed {
		return nil
	}
	if err := objectstore.CheckObjects(s.dir); err != nil {
		return mapObjectsErr(err)
	}
	wrote, err := s.tmp.Install(s.rel, s.checksum)
	s.wrote = s.wrote || wrote
	if err != nil {
		return mapObjectsErr(err)
	}
	s.installed = true
	return nil
}

// InsertTx makes the staged object a File inside the caller's write: it reuses
// a files row with the same checksum (no change) or inserts one with its
// audited create change, then installs the temp if the object is missing.
// Calling it again returns the same File and no change.
func (s *Staged) InsertTx(tx *sql.Tx) (files.File, []rowchange.Change, error) {
	if s.haveFile {
		return s.file, nil, nil
	}
	var (
		file    files.File
		changes []rowchange.Change
	)
	found, err := files.LookupByChecksumTx(tx, s.checksum)
	switch {
	case err == nil:
		file = found
	case errors.Is(err, sql.ErrNoRows):
		id, err := files.NewID()
		if err != nil {
			return files.File{}, nil, err
		}
		file = files.File{ID: id, ChecksumSHA256: s.checksum, OriginalFilename: s.filename, MediaType: s.mediaType, ByteSize: s.byteSize}
		ch, err := insertFile(tx, file)
		if err != nil {
			return files.File{}, nil, err
		}
		changes = []rowchange.Change{ch}
	default:
		return files.File{}, nil, err
	}
	if err := s.install(); err != nil {
		return files.File{}, nil, err
	}
	s.file, s.haveFile = file, true
	return file, changes, nil
}

// Discard removes the temp file when it was never installed. It runs outside
// the session, never touches an installed object, and is safe to call more
// than once and after a commit.
func (s *Staged) Discard() {
	if s != nil && s.tmp != nil {
		s.tmp.Discard()
	}
}

// Settle runs inside the catalog session after the write returns, whether it
// committed or not: when this stage installed its object and no files row
// names the checksum, it purges the object; then it removes the temp. Safe to
// call more than once and after a commit.
func (s *Staged) Settle(c *database.Catalog) {
	if s == nil {
		return
	}
	if s.wrote {
		files.PurgeUnreferenced(c, []files.Object{{ChecksumSHA256: s.checksum, MediaType: s.mediaType}})
		s.wrote = false
	}
	s.Discard()
}

// Stages is the stages one request made.
type Stages []*Staged

// Discard discards every stage. Outside the session.
func (ss Stages) Discard() {
	for _, s := range ss {
		s.Discard()
	}
}

// Settle settles every stage. Inside the session, after the write.
func (ss Stages) Settle(c *database.Catalog) {
	for _, s := range ss {
		s.Settle(c)
	}
}
