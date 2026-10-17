// Package namevalue is the pure NameValue type: the value, its parts, the
// compiled part-type registry and validation. It imports no database package;
// queries live in core/database/namevalues.
//
// A NameValue always has a full-form normalized reading (`form`). Ordered
// parts are optional. Part `type` is a product-known key from the compiled
// registry (see registry.go) when set; empty type means an untyped segment.
// Validate rejects unknown non-empty types. User-minted part-type vocabulary
// (DB rows) is deferred.
//
// NameValues are value objects (UUID for persistence only). Format profiles
// and project defaults live elsewhere (structured-name-model §4).
package namevalue

import "github.com/mendahu/provenencia/core/apperr"

var ErrInvalid = apperr.New(apperr.CodeNameValuesInvalid, apperr.KindUser)

// Product part-type keys (seeded-vocabulary §4.1). Same strings as PartTypes().
const (
	PartTypePrefix        = "prefix"
	PartTypeGiven         = "given"
	PartTypeNick          = "nick"
	PartTypeSurnamePrefix = "surname_prefix"
	PartTypeSurname       = "surname"
	PartTypeSuffix        = "suffix"
	PartTypeUndetermined  = "undetermined"
)

// Part is one ordered segment of a NameValue.
type Part struct {
	ID    []byte
	Idx   int
	Value string
	Type  string // empty or a KnownPartType key
}

// Value is one name_values row plus optional ordered parts.
type Value struct {
	ID    []byte
	Form  string
	Parts []Part // ordered by idx on Lookup
}

// Validate reports ErrInvalid when v is not a storable NameValue. Callers trim
// string fields first.
func Validate(v Value) error {
	if v.Form == "" {
		return ErrInvalid
	}
	seen := make(map[int]struct{}, len(v.Parts))
	for _, p := range v.Parts {
		if p.Idx < 0 {
			return ErrInvalid
		}
		if p.Value == "" {
			return ErrInvalid
		}
		if p.Type != "" && !KnownPartType(p.Type) {
			return ErrInvalid
		}
		if _, dup := seen[p.Idx]; dup {
			return ErrInvalid
		}
		seen[p.Idx] = struct{}{}
	}
	return nil
}
