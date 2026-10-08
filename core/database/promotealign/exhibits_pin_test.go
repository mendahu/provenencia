package promotealign

import (
	"testing"

	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/match"
)

func TestOverlappingBoundsAreSimilarAndUnpinned(t *testing.T) {
	prop := match.Property{Key: "start_date", Origin: "provenencia"}
	incoming := exhibitObs{
		id: []byte("in"), prop: prop, valueType: properties.ValueTypeDate,
		value: match.Value{Date: befDate(2001, 3, 31)},
	}
	member := exhibitObs{
		id: []byte("mem"), prop: prop, valueType: properties.ValueTypeDate,
		value: match.Value{Date: befDate(1985, 2, 1)},
	}
	lines := pairExhibits([]exhibitObs{incoming}, nil, []exhibitObs{member}, graphalign.DefaultConfig(), graphalign.Stats{})
	if len(lines) != 1 {
		t.Fatalf("lines %+v", lines)
	}
	if lines[0].Outcome != match.OutcomePartial || lines[0].Pinned || lines[0].Weight <= 0 {
		t.Fatalf("overlap %+v, want an unpinned resemblance", lines[0])
	}
	if lines[0].IncomingDate == nil || lines[0].MemberDate == nil ||
		lines[0].IncomingDate.Qualifier != datevalues.QualifierBEF ||
		lines[0].MemberDate.StartYear == nil || *lines[0].MemberDate.StartYear != 1985 {
		t.Fatalf("dates %+v %+v", lines[0].IncomingDate, lines[0].MemberDate)
	}
}

func TestDisjointDatesConflict(t *testing.T) {
	prop := match.Property{Key: "start_date", Origin: "provenencia"}
	incoming := exhibitObs{
		id: []byte("in"), prop: prop, valueType: properties.ValueTypeDate,
		value: match.Value{Date: befDate(1985, 2, 1)},
	}
	year := 2001
	member := exhibitObs{
		id: []byte("mem"), prop: prop, valueType: properties.ValueTypeDate,
		value: match.Value{Date: &datevalues.Value{Kind: datevalues.KindPoint, StartYear: &year}},
	}
	lines := pairExhibits([]exhibitObs{incoming}, nil, []exhibitObs{member}, graphalign.DefaultConfig(), graphalign.Stats{})
	if len(lines) != 1 || lines[0].Outcome != match.OutcomeConflict || lines[0].Pinned {
		t.Fatalf("disjoint %+v, want an unpinned conflict", lines)
	}
}

func befDate(year, month, day int) *datevalues.Value {
	y, m, d := year, month, day
	return &datevalues.Value{
		Kind: datevalues.KindPoint, Qualifier: datevalues.QualifierBEF,
		StartYear: &y, StartMonth: &m, StartDay: &d,
	}
}
