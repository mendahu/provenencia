package observations

import (
	"bytes"
	"strings"

	"github.com/mendahu/provenencia/core/database/graphcache"
)

// TextRecord is the part of an Observation a first-text pick reads.
type TextRecord struct {
	Ref      string
	Polarity string
	Text     string
}

// FirstText is the one rule for "a Subject reads as its first text" (a Place
// as its first toponym): the positive record whose text is not blank, lowest
// ref first, refs compared without case as the Evidence graph loads them. It
// returns that text trimmed, or "" when no record qualifies.
func FirstText(records []TextRecord) string {
	best, found := TextRecord{}, false
	for _, r := range records {
		text := strings.TrimSpace(r.Text)
		if r.Polarity != PolarityPositive || text == "" {
			continue
		}
		r.Text = text
		if !found || refBefore(r.Ref, best.Ref) {
			best, found = r, true
		}
	}
	return best.Text
}

func refBefore(a, b string) bool {
	if la, lb := strings.ToLower(a), strings.ToLower(b); la != lb {
		return la < lb
	}
	return a < b
}

// FirstTextOf is FirstText over the cached Observations of one Property: a
// Subject's first text of propertyID. It returns "" when none qualifies.
func FirstTextOf(obs []graphcache.EvidenceObservation, propertyID []byte) string {
	var records []TextRecord
	for _, o := range obs {
		if bytes.Equal(o.PropertyID, propertyID) {
			records = append(records, TextRecord{Ref: o.Ref, Polarity: o.Polarity, Text: o.Text})
		}
	}
	return FirstText(records)
}
