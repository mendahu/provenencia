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
	regionPages, err := regionDocPages(regionJSON)
	if err != nil {
		return err
	}
	citation, err := decode(citationLocatorJSON)
	if err != nil {
		return err
	}
	if err := validateDoc(citation); err != nil {
		return err
	}
	page, _, ok, err := citationPage(citation)
	if err != nil {
		return err
	}
	if ok != (len(regionPages) == 1) {
		return ErrInvalid
	}
	if ok && regionPages[0] != page {
		return ErrInvalid
	}
	return nil
}

// Page returns the artifact_page and the trimmed page_label of a Citation
// locator's page selector. It returns 0, "" when the document is empty,
// invalid, longer than MaxDocumentBytes, has two page selectors, or has none.
func Page(locatorJSON string) (artifactPage int, pageLabel string) {
	doc, err := decode(locatorJSON)
	if err != nil {
		return 0, ""
	}
	if err := validateDoc(doc); err != nil {
		return 0, ""
	}
	page, label, ok, err := citationPage(doc)
	if err != nil || !ok {
		return 0, ""
	}
	return page, label
}

// citationPage reads the page selector of a validated Citation locator. A
// Citation locator has at most one; ok is false when it has none.
func citationPage(doc document) (page int, label string, ok bool, err error) {
	for _, sel := range doc.Selectors {
		if strings.TrimSpace(sel.Type) != TypePage {
			continue
		}
		if ok {
			return 0, "", false, ErrInvalid
		}
		ok = true
		page = *sel.ArtifactPage
		if sel.PageLabel != nil {
			label = strings.TrimSpace(*sel.PageLabel)
		}
	}
	return page, label, ok, nil
}

// regionDocPages validates a region document and returns the artifact_page of
// its page selector, if any.
func regionDocPages(regionJSON string) ([]int, error) {
	region, err := decode(regionJSON)
	if err != nil {
		return nil, err
	}
	if err := validateDoc(region); err != nil {
		return nil, err
	}
	regions := 0
	var pages []int
	for _, sel := range region.Selectors {
		switch strings.TrimSpace(sel.Type) {
		case TypeArtifact:
		case TypePage:
			pages = append(pages, *sel.ArtifactPage)
		case TypeRegion:
			regions++
		default:
			return nil, ErrInvalid
		}
	}
	if regions != 1 || len(pages) > 1 {
		return nil, ErrInvalid
	}
	return pages, nil
}
