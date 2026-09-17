package ref

import (
	"crypto/rand"
	"regexp"
	"strings"
	"unicode"

	"github.com/mendahu/provenencia/core/apperr"
)

// Crockford base32 alphabet without I, L, O, U (ambiguous with 1/0).
const alphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

const tokenLen = 5

// Reserved catalog / contributor prefixes (not used as a Node Type prefix).
const (
	PrefixUser        = "USR"
	PrefixSource      = "SRC"
	PrefixArtifact    = "ART"
	PrefixCitation    = "CIT"
	PrefixObservation = "OBS"
)

var reservedPrefixes = map[string]struct{}{
	PrefixUser:        {},
	PrefixSource:      {},
	PrefixArtifact:    {},
	PrefixCitation:    {},
	PrefixObservation: {},
}

var (
	ErrInvalidPrefix = apperr.New(apperr.CodeRefInvalidPrefix, apperr.KindInternal)
	ErrInvalid       = apperr.New(apperr.CodeRefInvalid, apperr.KindInternal)
	// ErrReservedPrefix is a user error: a Node Type may not claim a catalog prefix.
	ErrReservedPrefix = apperr.New(apperr.CodeRefReservedPrefix, apperr.KindUser)
)

var validRef = regexp.MustCompile(`^[A-Z]{3}-[0-9A-HJKMNP-TV-Z]{5}$`)

// Mint returns PREFIX-TOKEN (e.g. USR-F4N2P). prefix must be three ASCII letters.
func Mint(prefix string) (string, error) {
	p, err := normalizePrefix(prefix)
	if err != nil {
		return "", err
	}
	token, err := randomToken()
	if err != nil {
		return "", err
	}
	return p + "-" + token, nil
}

// Valid reports whether s matches {PREFIX}-{token}.
func Valid(s string) bool {
	return validRef.MatchString(strings.TrimSpace(s))
}

// Validate returns ErrInvalid when s is not a well-formed ref.
func Validate(s string) error {
	if !Valid(s) {
		return ErrInvalid
	}
	return nil
}

// ValidatePrefix normalizes a Node Type prefix — either node_types.ref_prefix
// (canonical entities) or node_types.candidate_ref_prefix (Interpretation
// Nodes) — and rejects the reserved catalog prefixes. Mint itself still accepts
// them, because they are how catalog rows are minted; only Node Type vocabulary
// is constrained.
//
// Uniqueness across the two columns is an application rule, not a format rule:
// nothing here distinguishes PER from CPR.
func ValidatePrefix(prefix string) (string, error) {
	p, err := normalizePrefix(prefix)
	if err != nil {
		return "", err
	}
	if _, ok := reservedPrefixes[p]; ok {
		return "", ErrReservedPrefix
	}
	return p, nil
}

func normalizePrefix(prefix string) (string, error) {
	p := strings.TrimSpace(prefix)
	if len(p) != 3 {
		return "", ErrInvalidPrefix
	}
	var b strings.Builder
	for _, r := range p {
		if r > unicode.MaxASCII || !unicode.IsLetter(r) {
			return "", ErrInvalidPrefix
		}
		b.WriteRune(unicode.ToUpper(r))
	}
	return b.String(), nil
}

func randomToken() (string, error) {
	buf := make([]byte, tokenLen)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	var b strings.Builder
	b.Grow(tokenLen)
	for _, by := range buf {
		b.WriteByte(alphabet[int(by)%len(alphabet)])
	}
	return b.String(), nil
}
