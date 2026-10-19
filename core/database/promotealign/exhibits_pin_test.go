package promotealign

import (
	"math"
	"testing"

	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/propertycompare"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/subjectprofile"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// obs is a positive Observation of a product property.
func obs(id, key, valueType string, auto autoreconcile.Value) exhibitObs {
	o := exhibitObs{
		id: []byte(id), key: key, origin: vocabkey.OriginProvenencia, valueType: valueType,
		auto: auto, value: subjectprofile.Value(auto, false, ""),
	}
	o.value.Type = valueType
	return o
}

func dateObs(id, key string, d *datevalue.Value) exhibitObs {
	return obs(id, key, propertymeta.ValueTypeDate, autoreconcile.Value{Date: d})
}

func TestOverlappingBoundsAreSimilarAndUnpinned(t *testing.T) {
	incoming := dateObs("in", vocabkey.PropertyStartDate, befDate(2001, 3, 31))
	member := dateObs("mem", vocabkey.PropertyStartDate, befDate(1985, 2, 1))
	lines := pairExhibits([]exhibitObs{incoming}, nil, vocabkey.Event, []exhibitObs{member}, matchregistry.Default().AlignConfig())
	if len(lines) != 1 {
		t.Fatalf("lines %+v", lines)
	}
	if lines[0].Outcome != propertycompare.OutcomePartial || lines[0].Pinned || lines[0].Weight <= 0 || lines[0].Weight > 2 {
		t.Fatalf("overlap %+v, want an unpinned resemblance weighing at most the date's agree weight", lines[0])
	}
	if lines[0].IncomingDate == nil || lines[0].MemberDate == nil ||
		lines[0].IncomingDate.Qualifier != datevalue.QualifierBEF ||
		lines[0].MemberDate.StartYear == nil || *lines[0].MemberDate.StartYear != 1985 {
		t.Fatalf("dates %+v %+v", lines[0].IncomingDate, lines[0].MemberDate)
	}
}

func TestDisjointDatesConflict(t *testing.T) {
	incoming := dateObs("in", vocabkey.PropertyStartDate, befDate(1985, 2, 1))
	year := 2001
	member := dateObs("mem", vocabkey.PropertyStartDate, &datevalue.Value{Kind: datevalue.KindPoint, StartYear: &year})
	lines := pairExhibits([]exhibitObs{incoming}, nil, vocabkey.Event, []exhibitObs{member}, matchregistry.Default().AlignConfig())
	if len(lines) != 1 || lines[0].Outcome != propertycompare.OutcomeConflict || lines[0].Pinned || lines[0].Weight != -3 {
		t.Fatalf("disjoint %+v, want an unpinned conflict weighing -3", lines)
	}
}

// A line weighs what its property counts in the Subject's score: its lead's
// weight, or nothing for a property the subject section does not weigh.
func TestExhibitWeightIsTheLeadWeight(t *testing.T) {
	cfg := matchregistry.Default().AlignConfig()
	t.Run("a property the subject section does not weigh", func(t *testing.T) {
		text := autoreconcile.Value{Text: "Great Fire", HasText: true}
		in := obs("in", "event_name", propertymeta.ValueTypeText, text)
		mem := obs("mem", "event_name", propertymeta.ValueTypeText, text)
		lines := pairExhibits([]exhibitObs{in}, nil, vocabkey.Event, []exhibitObs{mem}, cfg)
		if len(lines) != 1 || lines[0].Outcome != propertycompare.OutcomeAgree || !lines[0].Pinned || lines[0].Weight != 0 {
			t.Fatalf("event_name %+v, want a pinned agreement weighing 0", lines)
		}
	})
	t.Run("a pooled property weighs its lead", func(t *testing.T) {
		female := autoreconcile.Value{TermID: []byte("t-female"), TermKey: "female"}
		in := obs("in", vocabkey.PropertyGender, propertymeta.ValueTypeTerm, female)
		mem := obs("mem", vocabkey.PropertyGender, propertymeta.ValueTypeTerm, female)
		lines := pairExhibits([]exhibitObs{in}, nil, vocabkey.Person, []exhibitObs{mem}, cfg)
		want := cfg.Subject.Weights[vocabkey.Person][vocabkey.PropertySexAtBirth].Agree
		if len(lines) != 1 || lines[0].Outcome != propertycompare.OutcomeAgree || !lines[0].Pinned || math.Abs(lines[0].Weight-want) > 1e-9 {
			t.Fatalf("gender %+v, want a pinned agreement weighing %v", lines, want)
		}
	})
}

func befDate(year, month, day int) *datevalue.Value {
	y, m, d := year, month, day
	return &datevalue.Value{
		Kind: datevalue.KindPoint, Qualifier: datevalue.QualifierBEF,
		StartYear: &y, StartMonth: &m, StartDay: &d,
	}
}
