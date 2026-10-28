// Package objectpath builds content-addressed relative paths under a project
// directory (objects/{hh}/{hh}/{hex}{ext}). Kept free of catalog imports so
// both files and searchindex can share it without a cycle. The extension comes
// from the ingest allowlist; a media type off it has no path.
package objectpath

import (
	"errors"
	"strings"

	"github.com/mendahu/provenencia/core/ingest/mediatypes"
)

// ErrInvalidChecksum means the checksum is not 64 lowercase hex characters.
var ErrInvalidChecksum = errors.New("objectpath: invalid checksum")

// ErrDisallowedType means the media type is not on the ingest allowlist
// (core/ingest/mediatypes), so no object path exists for it.
var ErrDisallowedType = errors.New("objectpath: media type not allowed")

// Dir is the project folder that holds every object file.
const Dir = "objects"

// Rel returns objects/{hh}/{hh}/{fullhex}{ext} for a 64-char lowercase hex
// checksum and a media type on the allowlist. The extension is the one
// mediatypes.StoredExtension gives. A media type off the allowlist has no path
// and returns ErrDisallowedType, so every object path has an extension.
func Rel(checksumHex, mediaType string) (string, error) {
	checksumHex = strings.TrimSpace(checksumHex)
	if len(checksumHex) != 64 || !isLowerHex(checksumHex) {
		return "", ErrInvalidChecksum
	}
	ext, ok := mediatypes.StoredExtension(mediaType)
	if !ok {
		return "", ErrDisallowedType
	}
	return Dir + "/" + checksumHex[0:2] + "/" + checksumHex[2:4] + "/" + checksumHex + ext, nil
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
