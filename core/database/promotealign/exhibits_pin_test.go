package promotealign

import (
	"testing"

	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/match"
)

func TestOverlappingBoundStaysUnpinnedWhenItConflicts(t *testing.T) {
	prop := match.Property{Key: "start_date", Origin: "provenencia"}
	incoming := exhibitObs{
		id: []byte("in"), prop: prop, valueType: properties.ValueTypeDate,
		value: match.Value{Date: befDate(1990, 12, 27)},
	}
	member := exhibitObs{
		id: []byte("mem"), prop: prop, valueType: properties.ValueTypeDate,
		value: match.Value{Date: befDate(1985, 2, 1)},
	}
	lines := pairExhibits([]exhibitObs{incoming}, nil, []exhibitObs{member}, graphalign.DefaultConfig(), graphalign.Stats{})
	if len(lines) != 1 {
		t.Fatalf("lines %+v", lines)
	}
	if lines[0].Outcome != match.OutcomeConflict || lines[0].Pinned {
		t.Fatalf("overlap %+v, want an unpinned conflict", lines[0])
	}
}

func befDate(year, month, day int) *datevalues.Value {
	y, m, d := year, month, day
	return &datevalues.Value{
		Kind: datevalues.KindPoint, Qualifier: datevalues.QualifierBEF,
		StartYear: &y, StartMonth: &m, StartDay: &d,
	}
}
