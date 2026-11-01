package conclusionsummaries

import (
	"sort"
	"strings"

	"github.com/mendahu/provenencia/core/database/nameformats"
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

// titleCollator is built once. Constructing it sorts the whole collation
// table, which is most of a list's time when every call makes a new one.
var titleCollator = collate.New(language.Und, collate.IgnoreCase, collate.IgnoreDiacritics)

// SortByTitle orders any titled row by the title it shows (a Person or Place
// row by its name or toponym, else its working label; a Source group by its
// title), ignoring case and diacritics (Montréal beside Montreal), by the
// Unicode root collation so the order is the same in every locale. A row
// titled only by its ref sorts last, by ref. Ties break by ref. This is the
// one owner of title order; the app shows rows as given.
func SortByTitle[T any](rows []T, title func(T) string, ref func(T) string) {
	sort.SliceStable(rows, func(i, j int) bool {
		ti, tj := strings.TrimSpace(title(rows[i])), strings.TrimSpace(title(rows[j]))
		if (ti == "") != (tj == "") {
			return tj == ""
		}
		if ti != "" {
			if o := titleCollator.CompareString(ti, tj); o != 0 {
				return o < 0
			}
		}
		return strings.ToLower(ref(rows[i])) < strings.ToLower(ref(rows[j]))
	})
}

// personSortKey is the People list order. A name with parts sorts by the
// profile's sorted order, undecorated; the row may add quotes and parentheses. A catalog with no
// default profile joins those parts in stored order. form is the transcription
// and is not a sort input: a name with no parts sorts by its working label.
func personSortKey(h PersonSummary, order []nameformats.PartOrder) string {
	if h.Name != nil && len(h.Name.Parts) > 0 {
		if key := strings.TrimSpace(nameformats.Format(*h.Name, order)); key != "" {
			return key
		}
	}
	return strings.TrimSpace(h.Entity.Label)
}

func placeTitle(h PlaceSummary) string {
	if len(h.Names) > 0 {
		return h.Names[0]
	}
	return h.Entity.Label
}
