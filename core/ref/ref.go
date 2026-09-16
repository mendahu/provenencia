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

// candidateMarker separates prefix from token in Interpretation Node refs.
// Fixed application marker, never a ref_prefix.
const candidateMarker = "C"

// Reserved catalog / contributor prefixes (not used as node_types.ref_prefix).
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

// Canonical and candidate forms are disjoint by length (9 vs 11), so a token
// beginning with C (e.g. PER-C4N2P) never reads as a candidate ref.
var (
	validRef          = regexp.MustCompile(`^[A-Z]{3}-[0-9A-HJKMNP-TV-Z]{5}$`)
	validCandidateRef = regexp.MustCompile(`^[A-Z]{3}-C-[0-9A-HJKMNP-TV-Z]{5}$`)
	// partialRef matches a non-empty leading fragment of either form, short of
	// a complete ref: PER-7, PER-C, PER-C-, PER-C-7KD4. Keep in step with the
	// two forms above.
	partialRef = regexp.MustCompile(`^[A-Z]{3}-(?:[0-9A-HJKMNP-TV-Z]{1,4}|C-[0-9A-HJKMNP-TV-Z]{0,4})$`)
)

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

// MintCandidate returns PREFIX-C-TOKEN (e.g. PER-C-7KD45) for an Interpretation
// Node. prefix is the Node Type's ref_prefix.
func MintCandidate(prefix string) (string, error) {
	p, err := normalizePrefix(prefix)
	if err != nil {
		return "", err
	}
	token, err := randomToken()
	if err != nil {
		return "", err
	}
	return p + "-" + candidateMarker + "-" + token, nil
}

// Valid reports whether s matches {PREFIX}-{token}. Candidate refs are not
// valid here; use ValidCandidate.
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

// ValidCandidate reports whether s matches {PREFIX}-C-{token}.
func ValidCandidate(s string) bool {
	return validCandidateRef.MatchString(strings.TrimSpace(s))
}

// ValidateCandidate returns ErrInvalid when s is not a well-formed candidate ref.
func ValidateCandidate(s string) error {
	if !ValidCandidate(s) {
		return ErrInvalid
	}
	return nil
}

// ValidAny reports whether s is a complete ref in either form. Callers that
// resolve a ref the researcher typed or spoke want this; callers that assert a
// specific layer want Valid or ValidCandidate.
func ValidAny(s string) bool {
	t := strings.TrimSpace(s)
	return Valid(t) || ValidCandidate(t)
}

// ValidPartial reports whether s is a leading fragment of some complete ref —
// what a researcher has typed so far. Complete refs are not partial: test
// ValidAny first.
func ValidPartial(s string) bool {
	return partialRef.MatchString(strings.TrimSpace(s))
}

// ValidatePrefix normalizes a node_types.ref_prefix and rejects the reserved
// catalog prefixes. Mint itself still accepts them — they are how catalog rows
// are minted; only Node Types are constrained.
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
