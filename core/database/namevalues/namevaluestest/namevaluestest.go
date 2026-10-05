// Package namevaluestest builds structured names for test fixtures.
package namevaluestest

import (
	"strings"

	"github.com/mendahu/provenencia/core/database/namevalues"
)

// Western reads a form as given names then one surname, the way a fixture
// writes "James K. Robins": given James, given K., surname Robins. A one-word
// form is a surname. Fixtures only: the product never derives parts from
// form, and names with no parts drop out of resolution (S9-13).
func Western(form string) *namevalues.Value {
	v := &namevalues.Value{Form: form}
	words := strings.Fields(form)
	for i, w := range words {
		typ := namevalues.PartTypeGiven
		if i == len(words)-1 {
			typ = namevalues.PartTypeSurname
		}
		v.Parts = append(v.Parts, namevalues.Part{Idx: i, Value: w, Type: typ})
	}
	return v
}
