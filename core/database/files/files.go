// Package files accesses the content-addressed files catalog table.
package files

import (
	"database/sql"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/searchindex"
)

var ErrInvalid = apperr.New(apperr.CodeFilesInvalid, apperr.KindUser)

const (
	sqlInsert = `INSERT INTO files (id, checksum_sha256, original_filename, media_type, byte_size)
		VALUES (?, ?, ?, ?, ?)`
	sqlLookup = `SELECT id, checksum_sha256, COALESCE(original_filename, ''), COALESCE(media_type, ''), byte_size
		FROM files WHERE id = ?`
	sqlLookupChecksum = `SELECT id, checksum_sha256, COALESCE(original_filename, ''), COALESCE(media_type, ''), byte_size
		FROM files WHERE checksum_sha256 = ?`
	sqlUpdateFilename = `UPDATE files SET original_filename = ? WHERE id = ?`
	sqlCount          = `SELECT COUNT(*) FROM files`
)

// File is one files row. Storage path is derived from ChecksumSHA256 + MediaType, not stored.
type File struct {
	ID               []byte
	ChecksumSHA256   string
	OriginalFilename string
	// MediaType is advisory content sniffing (http.DetectContentType), not a trust boundary for rendering.
	MediaType string
	ByteSize  int64
}

// ExtensionForMediaType returns a leading-dot suffix for known MIME types
// (e.g. ".jpg"), or "" when unknown / empty. Strips ";…" parameters.
func ExtensionForMediaType(mediaType string) string {
	mediaType = strings.TrimSpace(mediaType)
	if i := strings.IndexByte(mediaType, ';'); i >= 0 {
		mediaType = mediaType[:i]
	}
	mediaType = strings.ToLower(strings.TrimSpace(mediaType))
	switch mediaType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "image/bmp":
		return ".bmp"
	case "image/tiff", "image/tif":
		return ".tiff"
	case "image/heic":
		return ".heic"
	case "image/heif":
		return ".heif"
	case "application/pdf":
		return ".pdf"
	case "application/msword":
		return ".doc"
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return ".docx"
	case "text/plain":
		return ".txt"
	case "text/csv", "application/csv":
		return ".csv"
	case "text/markdown", "text/x-markdown":
		return ".md"
	case "video/mp4":
		return ".mp4"
	case "video/quicktime":
		return ".mov"
	case "video/webm":
		return ".webm"
	case "video/x-m4v":
		return ".m4v"
	case "video/x-msvideo":
		return ".avi"
	case "audio/mpeg":
		return ".mp3"
	case "audio/mp4", "audio/x-m4a":
		return ".m4a"
	case "audio/aac":
		return ".aac"
	case "audio/wav", "audio/wave", "audio/x-wav":
		return ".wav"
	case "audio/ogg":
		return ".ogg"
	case "audio/flac":
		return ".flac"
	case "audio/aiff", "audio/x-aiff":
		return ".aiff"
	default:
		return ""
	}
}

// StorageRelPath returns objects/{hh}/{hh}/{fullhex}{ext} for a 64-char
// lowercase hex checksum. ext is MIME-derived (see ExtensionForMediaType);
// unknown media types keep the bare hex basename.
func StorageRelPath(checksumHex, mediaType string) (string, error) {
	checksumHex = strings.TrimSpace(checksumHex)
	if len(checksumHex) != 64 || !isLowerHex(checksumHex) {
		return "", ErrInvalid
	}
	base := "objects/" + checksumHex[0:2] + "/" + checksumHex[2:4] + "/" + checksumHex
	return base + ExtensionForMediaType(mediaType), nil
}

// Lookup returns a File by id, or sql.ErrNoRows.
func Lookup(c *database.Catalog, id []byte) (File, error) {
	db, err := c.DB()
	if err != nil {
		return File{}, err
	}
	if len(id) != 16 {
		return File{}, ErrInvalid
	}
	return scanFile(db.QueryRow(sqlLookup, id))
}

// LookupByChecksum returns a File by SHA-256 hex, or sql.ErrNoRows.
func LookupByChecksum(c *database.Catalog, checksumHex string) (File, error) {
	db, err := c.DB()
	if err != nil {
		return File{}, err
	}
	checksumHex = strings.TrimSpace(checksumHex)
	if len(checksumHex) != 64 || !isLowerHex(checksumHex) {
		return File{}, ErrInvalid
	}
	return scanFile(db.QueryRow(sqlLookupChecksum, checksumHex))
}

// Insert writes a new files row on tx. ID must be 16 bytes; checksum lowercase hex.
func Insert(tx *sql.Tx, f File) error {
	if tx == nil {
		return ErrInvalid
	}
	f.ChecksumSHA256 = strings.TrimSpace(f.ChecksumSHA256)
	f.OriginalFilename = strings.TrimSpace(f.OriginalFilename)
	f.MediaType = strings.TrimSpace(f.MediaType)
	if len(f.ID) != 16 || len(f.ChecksumSHA256) != 64 || !isLowerHex(f.ChecksumSHA256) || f.ByteSize < 0 {
		return ErrInvalid
	}
	_, err := tx.Exec(
		sqlInsert,
		f.ID,
		f.ChecksumSHA256,
		nullStr(f.OriginalFilename),
		nullStr(f.MediaType),
		f.ByteSize,
	)
	if err != nil {
		return mapConstraint(err)
	}
	return nil
}

// UpdateOriginalFilename sets original_filename for id on tx.
func UpdateOriginalFilename(tx *sql.Tx, id []byte, name string) error {
	if tx == nil || len(id) != 16 {
		return ErrInvalid
	}
	name = strings.TrimSpace(name)
	res, err := tx.Exec(sqlUpdateFilename, nullStr(name), id)
	if err != nil {
		return mapConstraint(err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	sourceIDs, err := searchindex.SourceIDsForFile(tx, id)
	if err != nil {
		return err
	}
	for _, sourceID := range sourceIDs {
		if err := searchindex.ReprojectSource(tx, sourceID); err != nil {
			return err
		}
	}
	return nil
}

// Count returns the total number of files rows — content-addressed, so
// this is distinct files, not the (larger, per-source) artifact count.
func Count(c *database.Catalog) (int, error) {
	db, err := c.DB()
	if err != nil {
		return 0, err
	}
	var n int
	if err := db.QueryRow(sqlCount).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

// NewID mints a UUIDv7 id for a new File row.
func NewID() ([]byte, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	return id[:], nil
}

func scanFile(row interface{ Scan(dest ...any) error }) (File, error) {
	var f File
	if err := row.Scan(&f.ID, &f.ChecksumSHA256, &f.OriginalFilename, &f.MediaType, &f.ByteSize); err != nil {
		return File{}, err
	}
	return f, nil
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func isLowerHex(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= '0' && c <= '9' || c >= 'a' && c <= 'f' {
			continue
		}
		return false
	}
	return true
}

func IsUniqueConflict(err error) bool {
	return database.IsUniqueConflict(err)
}

func mapConstraint(err error) error {
	if IsUniqueConflict(err) {
		return err
	}
	if database.IsConstraint(err) {
		return ErrInvalid
	}
	return err
}
