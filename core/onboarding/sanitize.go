package onboarding

import (
	"errors"
	"strings"
	"unicode"

	"github.com/mendahu/provenencia/core/database"
)

var ErrInvalidFamilyName = errors.New("invalid family name")

// FolderName turns a project/family label into a kebab-case *.provenencia basename.
// Example: "Robins Family" → "robins-family.provenencia".
func FolderName(familyName string) (string, error) {
	s := strings.TrimSpace(familyName)
	if strings.HasSuffix(strings.ToLower(s), database.Suffix) {
		s = strings.TrimSpace(s[:len(s)-len(database.Suffix)])
	}
	var b strings.Builder
	prevHyphen := false
	for _, r := range s {
		r = unicode.ToLower(r)
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			prevHyphen = false
			continue
		}
		if b.Len() > 0 && !prevHyphen {
			b.WriteByte('-')
			prevHyphen = true
		}
	}
	s = strings.Trim(b.String(), "-")
	if s == "" {
		return "", ErrInvalidFamilyName
	}
	return s + database.Suffix, nil
}
