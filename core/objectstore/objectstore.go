// Package objectstore verifies content-addressed object files on disk.
package objectstore

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

// MatchesChecksum reports whether objPath is a regular file whose SHA-256
// equals wantChecksum (lowercase hex). Missing paths return (false, nil).
func MatchesChecksum(objPath, wantChecksum string) (bool, error) {
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
