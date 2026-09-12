package sourcetypes

import "strings"

// DefaultIconKey is written for new user types and migration backfill.
const DefaultIconKey = "type_evidence"

// ValidIconKeys is the closed v1 set persisted on source_types.icon_key.
var ValidIconKeys = map[string]struct{}{
	"type_certificate":    {},
	"type_book":           {},
	"type_document":       {},
	"type_scroll":         {},
	"type_photograph":     {},
	"type_newspaper":      {},
	"type_map":            {},
	"type_microfilm":      {},
	"type_cassette":       {},
	"type_oral_history":   {},
	"type_video":          {},
	"type_website":        {},
	"type_census":         {},
	"type_dna":            {},
	"type_gedcom":         {},
	"type_grave":          {},
	"type_scrapbook":      {},
	"type_evidence":       {},
	"type_folder_archive": {},
	"type_email":          {},
	"type_postcard":       {},
}

// NormalizeIconKey trims and validates icon_key. Empty becomes DefaultIconKey.
// Unknown keys return ErrInvalid.
func NormalizeIconKey(raw string) (string, error) {
	key := strings.TrimSpace(raw)
	if key == "" {
		return DefaultIconKey, nil
	}
	if _, ok := ValidIconKeys[key]; !ok {
		return "", ErrInvalid
	}
	return key, nil
}
