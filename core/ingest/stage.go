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

// StagedImage is an image the app rendered, checked and re-encoded by Go as a
// PNG but not yet stored. StageImage writes nothing; InsertTx writes the row and
// its object inside the caller's write. A stage serves one write: after that
// write fails, call Discard and stage again.
type StagedImage struct {
	c        *database.Catalog
	png      []byte
	checksum string
	rel      string
	wrote    bool
	file     files.File
	haveFile bool
}

// StageImage applies the import path checks to absPath, decodes it as PNG or
// JPEG within the pixel budget, and re-encodes it as a PNG of at most 2048
// pixels on the long edge. It reads no catalog row and writes nothing. Never
// call it inside writes.Run. Defer Discard before the write that inserts it.
func StageImage(c *database.Catalog, absPath string) (*StagedImage, error) {
	if _, err := c.DB(); err != nil {
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
	sum := sha256.Sum256(out)
	checksum := hex.EncodeToString(sum[:])
	rel, err := files.StorageRelPath(checksum, "image/png")
	if err != nil {
		return nil, err
	}
	return &StagedImage{c: c, png: out, checksum: checksum, rel: rel}, nil
}

// InsertTx makes the staged PNG a File inside the caller's write: it reuses a
// files row with the same checksum (no change) or inserts one with its audited
// create change, then writes the object if it is missing. Calling it again
// returns the same File and no change.
func (s *StagedImage) InsertTx(tx *sql.Tx) (files.File, []rowchange.Change, error) {
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
		file = files.File{ID: id, ChecksumSHA256: s.checksum, MediaType: "image/png", ByteSize: int64(len(s.png))}
		ch, err := insertFile(tx, file)
		if err != nil {
			return files.File{}, nil, err
		}
		changes = []rowchange.Change{ch}
	default:
		return files.File{}, nil, err
	}
	wrote, err := objectstore.WriteIfAbsent(filepath.Join(s.c.Dir(), filepath.FromSlash(s.rel)), s.checksum, s.png)
	s.wrote = s.wrote || wrote
	if err != nil {
		return files.File{}, nil, err
	}
	s.file, s.haveFile = file, true
	return file, changes, nil
}

// Discard removes the staged object when this stage wrote it and no files row
// names its checksum. Call it (deferred) after the write returns, never inside
// the write function, and inside catalogsession.Do. Safe to call more than once
// and after a commit.
func (s *StagedImage) Discard() {
	if s.wrote {
		files.PurgeUnreferenced(s.c, []files.Object{{ChecksumSHA256: s.checksum, MediaType: "image/png"}})
		s.wrote = false
	}
	s.png = nil
}
