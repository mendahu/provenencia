package onboarding

import (
	"errors"
	"strings"

	"github.com/mendahu/provenance/core/database"
)

var ErrInvalidFamilyName = errors.New("invalid family name")

// FolderName turns a project/family name into a *.provenance directory basename.
func FolderName(familyName string) (string, error) {
	s := strings.TrimSpace(familyName)
	var b strings.Builder
	for _, r := range s {
		switch r {
		case 0, '/', '\\', ':', '*', '?', '"', '<', '>', '|':
			b.WriteByte('-')
		default:
			b.WriteRune(r)
		}
	}
	s = strings.TrimSpace(b.String())
	s = strings.TrimLeft(s, ".")
	s = strings.TrimSpace(s)
	if s == "" {
		return "", ErrInvalidFamilyName
	}
	if strings.HasSuffix(s, database.Suffix) {
		return s, nil
	}
	return s + database.Suffix, nil
}
