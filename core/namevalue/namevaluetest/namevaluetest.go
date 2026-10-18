// Package namevaluetest builds structured names for test fixtures.
package namevaluetest

import (
	"strings"

	"github.com/mendahu/provenencia/core/namevalue"
)

// Western reads a form as given names then one surname, the way a fixture
// writes "James K. Robins": given James, given K., surname Robins. A one-word
// form is a surname. Fixtures only: the product never derives parts from
// form, and names with no parts drop out of resolution (S9-13).
func Western(form string) *namevalue.Value {
	v := &namevalue.Value{Form: form}
	words := strings.Fields(form)
	for i, w := range words {
		typ := namevalue.PartTypeGiven
		if i == len(words)-1 {
			typ = namevalue.PartTypeSurname
		}
		v.Parts = append(v.Parts, namevalue.Part{Idx: i, Value: w, Type: typ})
	}
	return v
}
