package locator

import "strings"

// ValidateDepictionRegion checks the region locator of a depiction value
// against the locator of the Citation that holds it.
//
// Both documents must pass Validate. The region has exactly one region
// selector, at most one page selector, an optional leading artifact selector,
// and no other selector. The Citation locator has at most one page selector.
// The pages must agree: both absent, or both present with the same
// artifact_page (page_label is not compared). The region need not lie inside
// a region of the Citation's own locator.
func ValidateDepictionRegion(regionJSON, citationLocatorJSON string) error {
	region, err := decode(regionJSON)
	if err != nil {
		return err
	}
	if err := validateDoc(region); err != nil {
		return err
	}
	citation, err := decode(citationLocatorJSON)
	if err != nil {
		return err
	}
	if err := validateDoc(citation); err != nil {
		return err
	}
	regions := 0
	var regionPages []int
	for _, sel := range region.Selectors {
		switch strings.TrimSpace(sel.Type) {
		case TypeArtifact:
		case TypePage:
			regionPages = append(regionPages, *sel.ArtifactPage)
		case TypeRegion:
			regions++
		default:
			return ErrInvalid
		}
	}
	if regions != 1 || len(regionPages) > 1 {
		return ErrInvalid
	}
	var citationPages []int
	for _, sel := range citation.Selectors {
		if strings.TrimSpace(sel.Type) == TypePage {
			citationPages = append(citationPages, *sel.ArtifactPage)
		}
	}
	if len(citationPages) > 1 || len(citationPages) != len(regionPages) {
		return ErrInvalid
	}
	if len(regionPages) == 1 && regionPages[0] != citationPages[0] {
		return ErrInvalid
	}
	return nil
}
