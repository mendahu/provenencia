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

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/filederivatives"
	"github.com/mendahu/provenencia/core/database/files"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/searchindex"
	"github.com/mendahu/provenencia/core/derivatives/raster"
	"github.com/mendahu/provenencia/core/objectstore"
	"github.com/mendahu/provenencia/core/writes"
)

// errDerivedFileExists is returned from the insert closure when the derived
// file checksum loses a race. Ensure starts a second Run after that returns.
// errDerivativeLinkExists is the same for a file_derivatives unique conflict.
var (
	errDerivedFileExists    = errors.New("derivatives: derived file already stored")
	errDerivativeLinkExists = errors.New("derivatives: derivative link already stored")
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

	// LookupByChecksum uses another pool connection and deadlocks while a Run
	// transaction is open. Resolve the derived File id before Run.
	var derivedID []byte
	if existing, err := files.LookupByChecksum(c, checksum); err == nil {
		derivedID = append([]byte(nil), existing.ID...)
	} else if !errors.Is(err, sql.ErrNoRows) {
		return Result{}, err
	}

	var afterErr error
	link, _, err := writes.Run(c, writes.Op{}, func(tx *database.Tx) (filederivatives.Link, []rowchange.Change, error) {
		fileID := derivedID
		if fileID == nil {
			id, err := files.NewID()
			if err != nil {
				return filederivatives.Link{}, nil, err
			}
			row := files.File{
				ID:             id,
				ChecksumSHA256: checksum,
				MediaType:      "image/jpeg",
				ByteSize:       int64(len(derivedJPEG)),
			}
			if err := files.Insert(tx.Tx, row); err != nil {
				if files.IsUniqueConflict(err) {
					return filederivatives.Link{}, nil, errDerivedFileExists
				}
				return filederivatives.Link{}, nil, err
			}
			fileID = append([]byte(nil), id...)
		}
		link, changes, err := insertDerivativeLink(tx, c, sourceFileID, fileID, spec.Type, &afterErr)
		if err != nil {
			return filederivatives.Link{}, nil, err
		}
		return link, changes, nil
	})
	if errors.Is(err, errDerivedFileExists) {
		existing, lookupErr := files.LookupByChecksum(c, checksum)
		if lookupErr != nil {
			return Result{}, lookupErr
		}
		return insertLinkOnly(c, sourceFileID, existing.ID, spec.Type)
	}
	if errors.Is(err, errDerivativeLinkExists) {
		existing, lookupErr := filederivatives.Lookup(c, sourceFileID, spec.Type)
		if lookupErr != nil {
			return Result{}, lookupErr
		}
		return Result{Link: existing}, nil
	}
	if err != nil {
		return Result{}, err
	}
	if afterErr != nil {
		return Result{}, afterErr
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

// insertLinkOnly inserts the derivative link for a File that already exists.
// Call it only after the previous Run has returned.
func insertLinkOnly(c *database.Catalog, sourceFileID, derivedID []byte, derivativeType string) (Result, error) {
	if existing, err := filederivatives.Lookup(c, sourceFileID, derivativeType); err == nil {
		return Result{Link: existing}, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return Result{}, err
	}
	var afterErr error
	link, _, err := writes.Run(c, writes.Op{}, func(tx *database.Tx) (filederivatives.Link, []rowchange.Change, error) {
		link, changes, err := insertDerivativeLink(tx, c, sourceFileID, derivedID, derivativeType, &afterErr)
		if err != nil {
			return filederivatives.Link{}, nil, err
		}
		return link, changes, nil
	})
	if errors.Is(err, errDerivativeLinkExists) {
		existing, lookupErr := filederivatives.Lookup(c, sourceFileID, derivativeType)
		if lookupErr != nil {
			return Result{}, lookupErr
		}
		return Result{Link: existing}, nil
	}
	if err != nil {
		return Result{}, err
	}
	if afterErr != nil {
		return Result{}, afterErr
	}
	return Result{Link: link}, nil
}

// insertDerivativeLink writes one file_derivatives row and registers search
// reprojection for after commit. file_derivatives stays unaudited.
func insertDerivativeLink(tx *database.Tx, c *database.Catalog, sourceFileID, derivedID []byte, derivativeType string, afterErr *error) (filederivatives.Link, []rowchange.Change, error) {
	linkID, err := filederivatives.NewID()
	if err != nil {
		return filederivatives.Link{}, nil, err
	}
	link := filederivatives.Link{
		ID:             linkID,
		SourceFileID:   append([]byte(nil), sourceFileID...),
		DerivedFileID:  append([]byte(nil), derivedID...),
		DerivativeType: derivativeType,
	}
	if err := filederivatives.Insert(tx.Tx, link); err != nil {
		if filederivatives.IsUniqueConflict(err) {
			return filederivatives.Link{}, nil, errDerivativeLinkExists
		}
		return filederivatives.Link{}, nil, err
	}
	tx.AfterCommit(func() {
		*afterErr = reprojectSourcesForNewThumbnail(c, sourceFileID, derivativeType)
	})
	ch, err := fileDerivativeChange(link)
	if err != nil {
		return filederivatives.Link{}, nil, err
	}
	return link, []rowchange.Change{ch}, nil
}

func fileDerivativeChange(link filederivatives.Link) (rowchange.Change, error) {
	id, err := uuid.FromBytes(link.ID)
	if err != nil {
		return rowchange.Change{}, err
	}
	sourceID, err := uuid.FromBytes(link.SourceFileID)
	if err != nil {
		return rowchange.Change{}, err
	}
	derivedID, err := uuid.FromBytes(link.DerivedFileID)
	if err != nil {
		return rowchange.Change{}, err
	}
	return rowchange.Change{
		EntityType: "file_derivative",
		EntityID:   append([]byte(nil), link.ID...),
		Action:     rowchange.ActionCreate,
		Fields: map[string]rowchange.FieldDiff{
			"id":              {Old: nil, New: id.String()},
			"source_file_id":  {Old: nil, New: sourceID.String()},
			"derived_file_id": {Old: nil, New: derivedID.String()},
			"derivative_type": {Old: nil, New: link.DerivativeType},
		},
	}, nil
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
	if match, err := objectstore.MatchesChecksum(objPath, sha256Hex(data)); err != nil {
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
		if match, statErr := objectstore.MatchesChecksum(objPath, sha256Hex(data)); statErr == nil && match {
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
