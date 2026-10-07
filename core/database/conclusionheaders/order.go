package conclusionheaders

import (
	"sort"
	"strings"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

// sortByTitle orders Person and Place rows by the title the row shows: its
// name (or toponym), else its working label, ignoring case and diacritics
// (Montréal beside Montreal), by the Unicode root collation so the order is
// the same in every locale. A row titled only by its ref sorts last, by ref.
// Ties break by ref. This is the one owner of list order; the app shows rows
// as given.
func sortByTitle[T any](rows []T, title func(T) string, ref func(T) string) {
	c := collate.New(language.Und, collate.IgnoreCase, collate.IgnoreDiacritics)
	sort.SliceStable(rows, func(i, j int) bool {
		ti, tj := strings.TrimSpace(title(rows[i])), strings.TrimSpace(title(rows[j]))
		if (ti == "") != (tj == "") {
			return tj == ""
		}
		if ti != "" {
			if o := c.CompareString(ti, tj); o != 0 {
				return o < 0
			}
		}
		return strings.ToLower(ref(rows[i])) < strings.ToLower(ref(rows[j]))
	})
}

func personTitle(h PersonHeader) string {
	if h.Name != nil && strings.TrimSpace(h.Name.Form) != "" {
		return h.Name.Form
	}
	return h.Entity.Label
}

func placeTitle(h PlaceHeader) string {
	if len(h.Names) > 0 {
		return h.Names[0]
	}
	return h.Entity.Label
}
