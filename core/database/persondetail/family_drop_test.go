package persondetail

import (
	"testing"

	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/conclusionsummaries"
	"github.com/mendahu/provenencia/core/kinship"
)

// TestDropUnread covers D-10: a relative whose summary was not read is dropped
// with the unit it leaves empty. A summary missing mid-read cannot be built
// from the catalog, so the summaries come from the rows given.
func TestDropUnread(t *testing.T) {
	id := func(n byte) []byte { return []byte{n} }
	rel := func(n byte) kinship.Relative { return kinship.Relative{ID: id(n)} }
	row := func(n byte) conclusionsummaries.PersonSummary {
		return conclusionsummaries.PersonSummary{Entity: canonicalentities.Entity{ID: id(n)}}
	}
	family := Family{Person: id(1), Units: []kinship.Unit{
		{Kind: kinship.UnitParents, Partners: []kinship.Relative{rel(2), rel(3)}, Children: []kinship.Relative{rel(4)}},
		{Kind: kinship.UnitPartner, Partners: []kinship.Relative{rel(5)}, Children: []kinship.Relative{rel(6)}},
		{Kind: kinship.UnitOneParent, Children: []kinship.Relative{rel(7)}},
		{Kind: kinship.UnitSiblings, Children: []kinship.Relative{rel(8)}},
	}}
	// 3, 4 and 6 and 5 are read; 2, 7 and 8 are not. The parents unit keeps 3; the
	// partner unit keeps both; the one-parent and siblings units are empty.
	got := dropUnread(family, conclusionsummaries.NewSummaries([]conclusionsummaries.PersonSummary{row(3), row(4), row(5), row(6)}, nil, nil))
	if len(got.Units) != 2 {
		t.Fatalf("%d units, want 2: %+v", len(got.Units), got.Units)
	}
	if u := got.Units[0]; len(u.Partners) != 1 || u.Partners[0].ID[0] != 3 || len(u.Children) != 1 {
		t.Fatalf("parents unit %+v", u)
	}
	// A partner no one read empties its unit although it has a child.
	got = dropUnread(family, conclusionsummaries.NewSummaries([]conclusionsummaries.PersonSummary{row(3), row(4), row(6)}, nil, nil))
	if len(got.Units) != 1 || got.Units[0].Kind != kinship.UnitParents {
		t.Fatalf("units %+v", got.Units)
	}
}
