// Package objectpath builds content-addressed relative paths under a project
// directory (objects/{hh}/{hh}/{hex}{ext}). Kept free of catalog imports so
// both files and searchindex can share it without a cycle.
package objectpath

import (
	"errors"
	"strings"
)

// ErrInvalidChecksum means the checksum is not 64 lowercase hex characters.
var ErrInvalidChecksum = errors.New("objectpath: invalid checksum")

// Extension returns a leading-dot suffix for known MIME types (e.g. ".jpg"),
// or "" when unknown / empty. Strips ";…" parameters.
func Extension(mediaType string) string {
	mediaType = strings.TrimSpace(mediaType)
	if i := strings.IndexByte(mediaType, ';'); i >= 0 {
		mediaType = mediaType[:i]
	}
	mediaType = strings.ToLower(strings.TrimSpace(mediaType))
	switch mediaType {
	case "image/jpeg", "image/jpg":
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

// Rel returns objects/{hh}/{hh}/{fullhex}{ext} for a 64-char lowercase hex
// checksum. Unknown media types keep the bare hex basename.
func Rel(checksumHex, mediaType string) (string, error) {
	checksumHex = strings.TrimSpace(checksumHex)
	if len(checksumHex) != 64 || !isLowerHex(checksumHex) {
		return "", ErrInvalidChecksum
	}
	base := "objects/" + checksumHex[0:2] + "/" + checksumHex[2:4] + "/" + checksumHex
	return base + Extension(mediaType), nil
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
