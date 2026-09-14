// Package derivatives generates disposable File derivatives (thumbnails, etc.).
package derivatives

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/filederivatives"
	"github.com/mendahu/provenencia/core/database/files"
	"github.com/mendahu/provenencia/core/database/searchindex"
	"github.com/mendahu/provenencia/core/derivatives/raster"
)

var ErrInvalid = apperr.New(apperr.CodeFileDerivativesInvalid, apperr.KindUser)

// ErrUnprocessable means the source image cannot be decoded safely
// (byte size or declared pixel dimensions over budget, or bytes that are
// not an allowlisted raster format). The File itself remains valid.
var ErrUnprocessable = apperr.New(apperr.CodeFileDerivativesUnprocessable, apperr.KindUser)

// ErrCorruptObject means the object-store bytes do not match the checksum
// recorded in the catalog, so they were refused before decode.
var ErrCorruptObject = apperr.New(apperr.CodeFileDerivativesCorruptObject, apperr.KindInternal)

// maxSourceBytes caps how much of a source object Ensure will read and
// decode. Larger Files are valid catalog content but are refused for
// derivative generation with ErrUnprocessable.
const maxSourceBytes = 128 << 20 // 128 MiB

// decodeSem bounds concurrent read+decode+encode work so parallel Ensure
// calls cannot multiply peak image-decode memory.
var decodeSem = make(chan struct{}, decodeSlots())

func decodeSlots() int {
	n := runtime.GOMAXPROCS(0)
	if n > 4 {
		n = 4
	}
	if n < 1 {
		n = 1
	}
	return n
}

// Spec describes one derivative to ensure. Type is the file_derivatives.derivative_type key
// (unique per source). Raster holds encode parameters; add Spec helpers (Medium, …) as needed.
type Spec struct {
	Type   string
	Raster raster.Options
}

// ThumbnailSpec is the default small preview: JPEG, longest edge ≤ 256.
func ThumbnailSpec() Spec {
	return Spec{
		Type: filederivatives.TypeThumbnail,
		Raster: raster.Options{
			MaxEdge: 256,
			Quality: raster.DefaultQuality,
		},
	}
}

// Result is the outcome of Ensure / EnsureThumbnail.
type Result struct {
	Link    filederivatives.Link
	Skipped bool // true when media type is not a supported raster image
}

// EnsureThumbnail creates or returns the default thumbnail derivative.
// See Ensure for the authorization contract on sourceFileID.
func EnsureThumbnail(c *database.Catalog, sourceFileID []byte) (Result, error) {
	return Ensure(c, sourceFileID, ThumbnailSpec())
}

// Ensure creates or returns the derivative described by spec.
// Non-supported media types return Skipped without error. Generation is not audited.
//
// Authorization: Ensure trusts sourceFileID. Callers — especially future FFI
// handlers — must first verify the requester may access that File (e.g. via
// its artifact/source chain) before invoking derivative generation.
func Ensure(c *database.Catalog, sourceFileID []byte, spec Spec) (Result, error) {
	spec.Type = strings.TrimSpace(spec.Type)
	if len(sourceFileID) != 16 || spec.Type == "" ||
		spec.Raster.MaxEdge < 1 || spec.Raster.MaxEdge > raster.MaxEdgeLimit {
		return Result{}, ErrInvalid
	}
	src, err := files.Lookup(c, sourceFileID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Result{}, ErrInvalid
		}
		return Result{}, err
	}
	if !raster.Supported(src.MediaType) {
		return Result{Skipped: true}, nil
	}

	if existing, err := filederivatives.Lookup(c, sourceFileID, spec.Type); err == nil {
		return Result{Link: existing}, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return Result{}, err
	}

	if src.ByteSize > maxSourceBytes {
		return Result{}, ErrUnprocessable
	}

	rel, err := files.StorageRelPath(src.ChecksumSHA256, src.MediaType)
	if err != nil {
		return Result{}, err
	}
	objPath := filepath.Join(c.Dir(), filepath.FromSlash(rel))
	derivedJPEG, err := generate(objPath, src.ChecksumSHA256, spec.Raster)
	if err != nil {
		return Result{}, err
	}

	sum := sha256.Sum256(derivedJPEG)
	checksum := hex.EncodeToString(sum[:])
	derivedRel, err := files.StorageRelPath(checksum, "image/jpeg")
	if err != nil {
		return Result{}, err
	}
	derivedPath := filepath.Join(c.Dir(), filepath.FromSlash(derivedRel))
	if err := writeObjectIfAbsent(derivedPath, derivedJPEG); err != nil {
		return Result{}, err
	}

	// Resolve derived File id before opening a write tx — LookupByChecksum uses
	// another pool connection and would deadlock against SQLite while tx is open.
	var derivedID []byte
	if existing, err := files.LookupByChecksum(c, checksum); err == nil {
		derivedID = append([]byte(nil), existing.ID...)
	} else if !errors.Is(err, sql.ErrNoRows) {
		return Result{}, err
	}

	db, err := c.DB()
	if err != nil {
		return Result{}, err
	}
	tx, err := db.Begin()
	if err != nil {
		return Result{}, err
	}
	defer func() { _ = tx.Rollback() }()

	if derivedID == nil {
		id, err := files.NewID()
		if err != nil {
			return Result{}, err
		}
		row := files.File{
			ID:             id,
			ChecksumSHA256: checksum,
			MediaType:      "image/jpeg",
			ByteSize:       int64(len(derivedJPEG)),
		}
		if err := files.Insert(tx, row); err != nil {
			if files.IsUniqueConflict(err) {
				_ = tx.Rollback()
				existing, lookupErr := files.LookupByChecksum(c, checksum)
				if lookupErr != nil {
					return Result{}, lookupErr
				}
				derivedID = append([]byte(nil), existing.ID...)
				return insertLinkOnly(c, sourceFileID, derivedID, spec.Type)
			}
			return Result{}, err
		}
		derivedID = append([]byte(nil), id...)
	}

	linkID, err := filederivatives.NewID()
	if err != nil {
		return Result{}, err
	}
	link := filederivatives.Link{
		ID:             linkID,
		SourceFileID:   append([]byte(nil), sourceFileID...),
		DerivedFileID:  derivedID,
		DerivativeType: spec.Type,
	}
	if err := filederivatives.Insert(tx, link); err != nil {
		if filederivatives.IsUniqueConflict(err) {
			_ = tx.Rollback()
			existing, lookupErr := filederivatives.Lookup(c, sourceFileID, spec.Type)
			if lookupErr != nil {
				return Result{}, lookupErr
			}
			return Result{Link: existing}, nil
		}
		return Result{}, err
	}
	if err := tx.Commit(); err != nil {
		return Result{}, err
	}
	if err := reprojectSourcesForNewThumbnail(c, sourceFileID, spec.Type); err != nil {
		return Result{}, err
	}
	return Result{Link: link}, nil
}

// generate reads the source object (bounded and checksum-verified) and
// encodes the derivative JPEG. It holds decodeSem for the whole
// read+decode+encode section to bound peak memory across goroutines.
func generate(objPath, wantChecksum string, opts raster.Options) ([]byte, error) {
	decodeSem <- struct{}{}
	defer func() { <-decodeSem }()

	raw, err := readSourceObject(objPath, wantChecksum)
	if err != nil {
		return nil, err
	}
	derivedJPEG, err := raster.EncodeJPEG(raw, opts)
	if err != nil {
		if errors.Is(err, raster.ErrTooLarge) || errors.Is(err, raster.ErrUnsupportedFormat) {
			return nil, ErrUnprocessable
		}
		return nil, err
	}
	return derivedJPEG, nil
}

// readSourceObject reads at most maxSourceBytes from objPath and verifies
// the bytes against the catalog checksum before they reach any decoder.
func readSourceObject(objPath, wantChecksum string) ([]byte, error) {
	f, err := os.Open(objPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, maxSourceBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(raw)) > maxSourceBytes {
		return nil, ErrUnprocessable
	}
	if sha256Hex(raw) != wantChecksum {
		return nil, ErrCorruptObject
	}
	return raw, nil
}

func insertLinkOnly(c *database.Catalog, sourceFileID, derivedID []byte, derivativeType string) (Result, error) {
	if existing, err := filederivatives.Lookup(c, sourceFileID, derivativeType); err == nil {
		return Result{Link: existing}, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return Result{}, err
	}
	db, err := c.DB()
	if err != nil {
		return Result{}, err
	}
	tx, err := db.Begin()
	if err != nil {
		return Result{}, err
	}
	defer func() { _ = tx.Rollback() }()
	linkID, err := filederivatives.NewID()
	if err != nil {
		return Result{}, err
	}
	link := filederivatives.Link{
		ID:             linkID,
		SourceFileID:   append([]byte(nil), sourceFileID...),
		DerivedFileID:  append([]byte(nil), derivedID...),
		DerivativeType: derivativeType,
	}
	if err := filederivatives.Insert(tx, link); err != nil {
		if filederivatives.IsUniqueConflict(err) {
			_ = tx.Rollback()
			existing, lookupErr := filederivatives.Lookup(c, sourceFileID, derivativeType)
			if lookupErr != nil {
				return Result{}, lookupErr
			}
			return Result{Link: existing}, nil
		}
		return Result{}, err
	}
	if err := tx.Commit(); err != nil {
		return Result{}, err
	}
	if err := reprojectSourcesForNewThumbnail(c, sourceFileID, derivativeType); err != nil {
		return Result{}, err
	}
	return Result{Link: link}, nil
}

// reprojectSourcesForNewThumbnail refreshes Source search docs after a new
// thumbnail link is created so omnibar display stubs can pick up the path.
func reprojectSourcesForNewThumbnail(c *database.Catalog, sourceFileID []byte, derivativeType string) error {
	if derivativeType != filederivatives.TypeThumbnail {
		return nil
	}
	db, err := c.DB()
	if err != nil {
		return err
	}
	ids, err := searchindex.SourceIDsForFile(db, sourceFileID)
	if err != nil {
		return err
	}
	for _, id := range ids {
		if err := searchindex.ReprojectSource(db, id); err != nil {
			return err
		}
	}
	return nil
}

func writeObjectIfAbsent(objPath string, data []byte) error {
	if match, err := objectMatches(objPath, sha256Hex(data)); err != nil {
		return err
	} else if match {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(objPath), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(objPath), ".deriv-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	ok := false
	defer func() {
		if !ok {
			_ = os.Remove(tmpName)
		}
	}()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, objPath); err != nil {
		if match, statErr := objectMatches(objPath, sha256Hex(data)); statErr == nil && match {
			_ = os.Remove(tmpName)
			ok = true
			return nil
		}
		return err
	}
	_ = syncDir(filepath.Dir(objPath))
	ok = true
	return nil
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func objectMatches(objPath, wantChecksum string) (bool, error) {
	f, err := os.Open(objPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return false, err
	}
	if !st.Mode().IsRegular() {
		return false, nil
	}
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return false, err
	}
	return hex.EncodeToString(h.Sum(nil)) == wantChecksum, nil
}
