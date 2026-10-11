// Package ingest copies opaque local files into a project’s content-addressed
// object store. Checksum reuse returns the existing File, records a reuse_file
// audit revision, and surfaces IngestedAs so the UI can offer keep/overwrite
// when the picked name differs. Ingest does not decode or execute file content.
package ingest

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/files"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/ingest/mediatypes"
	"github.com/mendahu/provenencia/core/objectstore"
	"github.com/mendahu/provenencia/core/writes"
)

var (
	ErrInvalid               = apperr.New(apperr.CodeIngestInvalid, apperr.KindUser)
	ErrPermissionDenied      = apperr.New(apperr.CodeIngestPermissionDenied, apperr.KindUser)
	ErrUnsupportedOffice     = apperr.New(apperr.CodeIngestUnsupportedOffice, apperr.KindUser)
	ErrUnsupportedArchive    = apperr.New(apperr.CodeIngestUnsupportedArchive, apperr.KindUser)
	ErrUnsupportedExecutable = apperr.New(apperr.CodeIngestUnsupportedExecutable, apperr.KindUser)
	ErrUnsupportedType       = apperr.New(apperr.CodeIngestUnsupportedType, apperr.KindUser)
	ErrUnidentified          = apperr.New(apperr.CodeIngestUnidentified, apperr.KindUser)
	ErrEmpty                 = apperr.New(apperr.CodeIngestEmpty, apperr.KindUser)
	ErrTooLarge              = apperr.New(apperr.CodeIngestTooLarge, apperr.KindUser)
	ErrNotAFile              = apperr.New(apperr.CodeIngestNotAFile, apperr.KindUser)
	ErrSymlink               = apperr.New(apperr.CodeIngestSymlink, apperr.KindUser)
	ErrMissing               = apperr.New(apperr.CodeIngestMissing, apperr.KindUser)
)

// MaxBytes is the largest source file ingest accepts (512 MiB).
const MaxBytes int64 = 512 << 20

// maxBytes is overridden in tests.
var maxBytes = MaxBytes

// errChecksumTaken is returned from the create_file closure when the insert
// loses a checksum race. File handles it after that Run returns.
var errChecksumTaken = errors.New("ingest: file checksum already stored")

// Result is an ingested or reused File plus its project-relative object path.
type Result struct {
	File       files.File
	RelPath    string
	Reused     bool   // checksum matched an existing File
	IngestedAs string // sanitized base name of the path the user picked
}

// File reads absPath into the catalog object store. Same bytes reuse the existing
// File row, verify or repair the object, and record a reuse_file audit event.
func File(c *database.Catalog, absPath string, userID []byte) (Result, error) {
	if _, err := c.DB(); err != nil {
		return Result{}, err
	}
	absPath = strings.TrimSpace(absPath)
	if absPath == "" || !filepath.IsAbs(absPath) {
		return Result{}, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return Result{}, err
	}

	info, err := os.Lstat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return Result{}, ErrMissing
		}
		return Result{}, mapOpenErr(err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return Result{}, ErrSymlink
	}
	if !info.Mode().IsRegular() {
		return Result{}, ErrNotAFile
	}
	if info.Size() == 0 {
		return Result{}, ErrEmpty
	}
	if info.Size() > maxBytes {
		return Result{}, ErrTooLarge.WithParams(formatByteSize(info.Size()))
	}

	src, err := openSource(absPath)
	if err != nil {
		return Result{}, mapOpenErr(err)
	}
	defer src.Close()

	st, err := src.Stat()
	if err != nil {
		return Result{}, err
	}
	if !st.Mode().IsRegular() {
		return Result{}, ErrNotAFile
	}
	if st.Size() == 0 {
		return Result{}, ErrEmpty
	}
	if st.Size() > maxBytes {
		return Result{}, ErrTooLarge.WithParams(formatByteSize(st.Size()))
	}

	objectsDir := filepath.Join(c.Dir(), "objects")
	tmpPath, checksum, byteSize, sniffed, err := streamSource(src, objectsDir)
	if err != nil {
		return Result{}, err
	}

	filename := sanitizeFilename(filepath.Base(absPath))
	mediaType, reason, ok := mediatypes.Resolve(sniffed, filename)
	if !ok {
		_ = os.Remove(tmpPath)
		return Result{}, errForReject(reason, sniffed)
	}

	relPath, err := files.StorageRelPath(checksum, mediaType)
	if err != nil {
		_ = os.Remove(tmpPath)
		return Result{}, err
	}
	objPath := filepath.Join(c.Dir(), filepath.FromSlash(relPath))

	// LookupByChecksum uses another pool connection and deadlocks while a Run
	// transaction is open. Reuse is decided here, before create_file's Run.
	if existing, err := files.LookupByChecksum(c, checksum); err == nil {
		if _, err := installObject(tmpPath, objPath, checksum); err != nil {
			return Result{}, err
		}
		return reuseFile(c, existing, relPath, filename, userID)
	} else if !errors.Is(err, sql.ErrNoRows) {
		_ = os.Remove(tmpPath)
		return Result{}, err
	}

	wrote, err := installObject(tmpPath, objPath, checksum)
	if err != nil {
		return Result{}, err
	}

	id, err := files.NewID()
	if err != nil {
		if wrote {
			_ = os.Remove(objPath)
		}
		return Result{}, err
	}
	row := files.File{
		ID:               id,
		ChecksumSHA256:   checksum,
		OriginalFilename: filename,
		MediaType:        mediaType,
		ByteSize:         byteSize,
	}

	res, _, err := writes.Run(c, writes.Op{
		Action: "create_file",
		UserID: userID,
	}, func(tx *database.Tx) (Result, []rowchange.Change, error) {
		if err := files.Insert(tx.Tx, row); err != nil {
			if files.IsUniqueConflict(err) {
				return Result{}, nil, errChecksumTaken
			}
			return Result{}, nil, err
		}
		uid, err := uuid.FromBytes(id)
		if err != nil {
			return Result{}, nil, err
		}
		fields := map[string]rowchange.FieldDiff{
			"id":              {Old: nil, New: uid.String()},
			"checksum_sha256": {Old: nil, New: checksum},
			"byte_size":       {Old: nil, New: row.ByteSize},
		}
		if filename != "" {
			fields["original_filename"] = rowchange.FieldDiff{Old: nil, New: filename}
		}
		if mediaType != "" {
			fields["media_type"] = rowchange.FieldDiff{Old: nil, New: mediaType}
		}
		return Result{File: row, RelPath: relPath, Reused: false, IngestedAs: filename}, []rowchange.Change{{
			EntityType: "file",
			EntityID:   id,
			Action:     rowchange.ActionCreate,
			Fields:     fields,
		}}, nil
	})
	if err != nil {
		if errors.Is(err, errChecksumTaken) {
			// LookupByChecksum uses another pool connection and deadlocks
			// while a Run transaction is open. This lookup is after Run.
			existing, lookupErr := files.LookupByChecksum(c, checksum)
			if lookupErr != nil {
				return Result{}, lookupErr
			}
			return reuseFile(c, existing, relPath, filename, userID)
		}
		if wrote {
			_ = os.Remove(objPath)
		}
		return Result{}, err
	}
	return res, nil
}

// SetFilename updates a File’s original_filename after a reuse keep/overwrite
// choice. The caller records the returned changes. An unchanged name returns
// no changes.
func SetFilename(tx *database.Tx, fileID []byte, name string, userID []byte) ([]rowchange.Change, error) {
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return nil, err
	}
	name = sanitizeFilename(name)
	existing, err := lookupFile(tx, fileID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalid
		}
		return nil, err
	}
	if existing.OriginalFilename == name {
		return nil, nil
	}
	if err := files.UpdateOriginalFilename(tx.Tx, fileID, name); err != nil {
		return nil, err
	}
	return []rowchange.Change{{
		EntityType: "file",
		EntityID:   append([]byte(nil), fileID...),
		Action:     rowchange.ActionUpdate,
		Fields: map[string]rowchange.FieldDiff{
			"original_filename": {Old: existing.OriginalFilename, New: name},
		},
	}}, nil
}

func lookupFile(tx *database.Tx, id []byte) (files.File, error) {
	if tx == nil || len(id) != 16 {
		return files.File{}, ErrInvalid
	}
	var f files.File
	err := tx.QueryRow(
		`SELECT id, checksum_sha256, COALESCE(original_filename, ''), COALESCE(media_type, ''), byte_size
		 FROM files WHERE id = ?`,
		id,
	).Scan(&f.ID, &f.ChecksumSHA256, &f.OriginalFilename, &f.MediaType, &f.ByteSize)
	return f, err
}

// reuseFile records reuse_file for an existing checksum. Call it only when no
// Run is open: it starts its own.
func reuseFile(c *database.Catalog, existing files.File, relPath, filename string, userID []byte) (Result, error) {
	res, _, err := writes.Run(c, writes.Op{
		Action: "reuse_file",
		UserID: userID,
	}, func(tx *database.Tx) (Result, []rowchange.Change, error) {
		return recordReuse(tx, existing, relPath, filename)
	})
	return res, err
}

func recordReuse(tx *database.Tx, existing files.File, relPath, filename string) (Result, []rowchange.Change, error) {
	if tx == nil {
		return Result{}, nil, ErrInvalid
	}
	return Result{File: existing, RelPath: relPath, Reused: true, IngestedAs: filename}, []rowchange.Change{{
		EntityType: "file",
		EntityID:   existing.ID,
		Action:     rowchange.ActionUpdate,
		Fields: map[string]rowchange.FieldDiff{
			"ingested_as": {Old: existing.OriginalFilename, New: filename},
		},
	}}, nil
}

type sniffBuf struct {
	buf []byte
}

func (s *sniffBuf) Write(p []byte) (int, error) {
	if len(s.buf) < 512 {
		need := 512 - len(s.buf)
		if need > len(p) {
			need = len(p)
		}
		s.buf = append(s.buf, p[:need]...)
	}
	return len(p), nil
}

func streamSource(src *os.File, objectsDir string) (tmpPath, checksum string, byteSize int64, mediaType string, err error) {
	if err := os.MkdirAll(objectsDir, 0o755); err != nil {
		return "", "", 0, "", err
	}
	tmp, err := os.CreateTemp(objectsDir, ".ingest-*")
	if err != nil {
		return "", "", 0, "", err
	}
	tmpPath = tmp.Name()
	fail := func(e error) (string, string, int64, string, error) {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return "", "", 0, "", e
	}

	h := sha256.New()
	sniff := &sniffBuf{}
	mw := io.MultiWriter(tmp, h, sniff)
	n, err := io.Copy(mw, io.LimitReader(src, maxBytes+1))
	if err != nil {
		return fail(err)
	}
	if n > maxBytes {
		return fail(ErrTooLarge.WithParams(formatByteSize(n)))
	}
	if err := tmp.Sync(); err != nil {
		return fail(err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return "", "", 0, "", err
	}
	if n == 0 {
		_ = os.Remove(tmpPath)
		return "", "", 0, "", ErrEmpty
	}
	mediaType = http.DetectContentType(sniff.buf)
	checksum = hex.EncodeToString(h.Sum(nil))
	return tmpPath, checksum, n, mediaType, nil
}

// installObject places tmpPath at objPath when missing or content mismatches.
// Always consumes/removes tmpPath. wrote is true when this call left new bytes at objPath.
func installObject(tmpPath, objPath, checksum string) (wrote bool, err error) {
	defer func() { _ = os.Remove(tmpPath) }()

	match, err := objectstore.MatchesChecksum(objPath, checksum)
	if err != nil {
		return false, err
	}
	if match {
		return false, nil
	}
	if err := os.MkdirAll(filepath.Dir(objPath), 0o755); err != nil {
		return false, err
	}
	_ = os.Remove(objPath)
	if err := os.Rename(tmpPath, objPath); err != nil {
		match, matchErr := objectstore.MatchesChecksum(objPath, checksum)
		if matchErr != nil {
			return false, matchErr
		}
		if match {
			return false, nil
		}
		return false, err
	}
	if err := syncDir(filepath.Dir(objPath)); err != nil {
		return true, err
	}
	return true, nil
}

func mapOpenErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return ErrMissing
	}
	if errors.Is(err, os.ErrPermission) || isAccessDenied(err) {
		return ErrPermissionDenied
	}
	if isSymlinkLoop(err) {
		return ErrSymlink
	}
	return err
}

func errForReject(reason mediatypes.Reason, sniffed string) error {
	switch reason {
	case mediatypes.ReasonOffice:
		return ErrUnsupportedOffice
	case mediatypes.ReasonArchive:
		return ErrUnsupportedArchive
	case mediatypes.ReasonExecutable:
		return ErrUnsupportedExecutable
	case mediatypes.ReasonUnidentified:
		return ErrUnidentified
	case mediatypes.ReasonDisallowedSniff:
		return ErrUnsupportedType.WithParams(mediatypes.ShortLabel(sniffed))
	case mediatypes.ReasonGenericType:
		return ErrUnsupportedType.WithParams(mediatypes.ShortLabel(sniffed))
	default:
		return ErrUnsupportedType.WithParams(mediatypes.ShortLabel(sniffed))
	}
}

func formatByteSize(n int64) string {
	const (
		kb = 1024
		mb = 1024 * kb
		gb = 1024 * mb
	)
	switch {
	case n >= gb:
		return fmt.Sprintf("%.2f GB", float64(n)/float64(gb))
	case n >= mb:
		return fmt.Sprintf("%.1f MB", float64(n)/float64(mb))
	case n >= kb:
		return fmt.Sprintf("%.0f KB", float64(n)/float64(kb))
	default:
		return fmt.Sprintf("%d bytes", n)
	}
}
