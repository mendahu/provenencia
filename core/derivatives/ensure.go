// Package derivatives generates disposable File derivatives (thumbnails, etc.).
package derivatives

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/filederivatives"
	"github.com/mendahu/provenencia/core/database/files"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/searchindex"
	"github.com/mendahu/provenencia/core/ingest"
	"github.com/mendahu/provenencia/core/objectstore"
	"github.com/mendahu/provenencia/core/raster"
	"github.com/mendahu/provenencia/core/writes"
)

// errDerivativeLinkExists is returned from the insert closure on a
// file_derivatives unique conflict. Commit looks the link up after that returns.
var errDerivativeLinkExists = errors.New("derivatives: derivative link already stored")

var ErrInvalid = apperr.New(apperr.CodeFileDerivativesInvalid, apperr.KindUser)

// ErrUnprocessable means the source image cannot be decoded safely
// (byte size or declared pixel dimensions over budget, or bytes that are
// not an allowlisted raster format). The File itself remains valid.
var ErrUnprocessable = apperr.New(apperr.CodeFileDerivativesUnprocessable, apperr.KindUser)

// ErrCorruptObject means the object-store bytes do not match the checksum
// recorded in the catalog, so they were refused before decode.
var ErrCorruptObject = apperr.New(apperr.CodeFileDerivativesCorruptObject, apperr.KindInternal)

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

// Job is the file work of one derivative that Plan found missing: the source
// File and its object, and the spec to make. Job.Stage does the decode and
// encode outside the catalog session; Commit stores the result inside it.
type Job struct {
	sourceFileID []byte
	spec         Spec
	rel          string // the source object's path
	checksum     string // the source File's checksum when planned
	staged       *ingest.Staged
}

// Plan is Ensure up to the decode. It returns a final Result and a nil Job
// when there is nothing to decode (a non-raster File is Skipped; an existing
// link is returned), or a Job whose Stage and Commit make the derivative.
// Non-supported media types return Skipped without error.
//
// Authorization: Plan trusts sourceFileID. Callers must first verify the
// requester may access that File (e.g. via its artifact/source chain).
func Plan(c *database.Catalog, sourceFileID []byte, spec Spec) (Result, *Job, error) {
	spec.Type = strings.TrimSpace(spec.Type)
	if len(sourceFileID) != 16 || spec.Type == "" ||
		spec.Raster.MaxEdge < 1 || spec.Raster.MaxEdge > raster.MaxEdgeLimit {
		return Result{}, nil, ErrInvalid
	}
	src, err := files.Lookup(c, sourceFileID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Result{}, nil, ErrInvalid
		}
		return Result{}, nil, err
	}
	if !raster.Supported(src.MediaType) {
		return Result{Skipped: true}, nil, nil
	}

	if existing, err := filederivatives.Lookup(c, sourceFileID, spec.Type); err == nil {
		return Result{Link: existing}, nil, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return Result{}, nil, err
	}

	if src.ByteSize > raster.MaxDecodeBytes {
		return Result{}, nil, ErrUnprocessable
	}

	rel, err := files.ReadRelPath(src)
	if err != nil {
		return Result{}, nil, err
	}
	if rel == "" {
		// A stored type off the allowlist has no object path to read.
		return Result{}, nil, ErrUnprocessable
	}
	return Result{}, &Job{
		sourceFileID: append([]byte(nil), sourceFileID...),
		spec:         spec,
		rel:          rel,
		checksum:     src.ChecksumSHA256,
	}, nil
}

// Stage reads the source object (bounded and checksum-verified), decodes and
// encodes the derivative, and writes it to a temp file in the objects folder.
// It reads no catalog row, so it runs outside the catalog session. The stage
// is kept on the Job for Commit and returned so the caller can add it to its
// ingest.Stages.
func (j *Job) Stage(projectDir string) (*ingest.Staged, error) {
	derivedJPEG, err := generate(projectDir, j.rel, j.checksum, j.spec.Raster)
	if err != nil {
		return nil, err
	}
	s, err := ingest.StageBytes(projectDir, derivedJPEG, "image/jpeg")
	if err != nil {
		return nil, err
	}
	j.staged = s
	return s, nil
}

// Commit stores a staged Job inside the catalog session, outside any write:
// the derived File row (unaudited) and the link in one Run. A source File that
// is gone or whose checksum changed since Plan is ErrInvalid; a link made
// meanwhile is returned and the stage is left unused. It purges nothing: the
// caller defers ingest.Stages.Settle inside the session.
func Commit(c *database.Catalog, j *Job) (Result, error) {
	if j == nil || j.staged == nil {
		return Result{}, ErrInvalid
	}
	src, err := files.Lookup(c, j.sourceFileID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Result{}, ErrInvalid
		}
		return Result{}, err
	}
	if src.ChecksumSHA256 != j.checksum {
		return Result{}, ErrInvalid
	}
	if existing, err := filederivatives.Lookup(c, j.sourceFileID, j.spec.Type); err == nil {
		return Result{Link: existing}, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return Result{}, err
	}

	var afterErr error
	link, _, err := writes.Run(c, writes.Op{}, func(tx *database.Tx) (filederivatives.Link, []rowchange.Change, error) {
		// The derived File stays unaudited: InsertTx's file change is dropped.
		derived, _, err := j.staged.InsertTx(tx.Tx)
		if err != nil {
			return filederivatives.Link{}, nil, err
		}
		return insertDerivativeLink(tx, c, j.sourceFileID, derived.ID, j.spec.Type, &afterErr)
	})
	if errors.Is(err, errDerivativeLinkExists) {
		existing, lookupErr := filederivatives.Lookup(c, j.sourceFileID, j.spec.Type)
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

// Refused reports whether err means the File gets no derivative and the caller
// shows none: the image cannot be decoded safely, its object is corrupt, or
// the File is gone or not a valid source. It is the one "show no thumbnail"
// test.
func Refused(err error) bool {
	return errors.Is(err, ErrUnprocessable) || errors.Is(err, ErrCorruptObject) || errors.Is(err, ErrInvalid)
}

// Ensure creates or returns the derivative described by spec: Plan, Stage,
// Commit. Generation is not audited. It decodes inside the caller's catalog
// session, so it is for tests and fixtures only; a handler plans in one
// session, stages outside it, and commits in the next (thumbnailPass in
// api/ffi/handlers).
func Ensure(c *database.Catalog, sourceFileID []byte, spec Spec) (Result, error) {
	res, job, err := Plan(c, sourceFileID, spec)
	if err != nil || job == nil {
		return res, err
	}
	staged, err := job.Stage(c.Dir())
	if err != nil {
		return Result{}, err
	}
	stages := ingest.Stages{staged}
	defer stages.Settle(c)
	return Commit(c, job)
}

// generate reads the source object (bounded and checksum-verified) and
// encodes the derivative JPEG. It holds a raster.Slot for the whole
// read+decode+encode section to bound peak memory across goroutines.
func generate(projectDir, rel, wantChecksum string, opts raster.Options) ([]byte, error) {
	defer raster.Slot()()

	raw, err := readSourceObject(projectDir, rel, wantChecksum)
	if err != nil {
		return nil, err
	}
	derivedJPEG, err := raster.EncodeJPEG(raw, opts)
	if err != nil {
		if errors.Is(err, raster.ErrTooLarge) || errors.Is(err, raster.ErrUnsupportedFormat) ||
			errors.Is(err, raster.ErrUndecodable) {
			return nil, ErrUnprocessable
		}
		return nil, err
	}
	return derivedJPEG, nil
}

// readSourceObject reads at most raster.MaxDecodeBytes from the object at rel and
// verifies the bytes against the catalog checksum before they reach any
// decoder. An object that is not a regular file, or a symlinked objects
// folder, is ErrCorruptObject. An object that is gone (its File was deleted
// after Plan) is ErrInvalid.
func readSourceObject(projectDir, rel, wantChecksum string) ([]byte, error) {
	f, err := objectstore.OpenObject(projectDir, rel)
	if err != nil {
		if errors.Is(err, objectstore.ErrNotRegular) || errors.Is(err, objectstore.ErrUnsafeObjects) {
			return nil, ErrCorruptObject
		}
		if errors.Is(err, fs.ErrNotExist) {
			// The File was deleted and its object purged after Plan.
			return nil, ErrInvalid
		}
		return nil, err
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, raster.MaxDecodeBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(raw)) > raster.MaxDecodeBytes {
		return nil, ErrUnprocessable
	}
	if sha256Hex(raw) != wantChecksum {
		return nil, ErrCorruptObject
	}
	return raw, nil
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

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
