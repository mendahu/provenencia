package claimevidence_test

import (
	"bytes"
	"math"
	"slices"
	"testing"

	"github.com/mendahu/provenencia/core/claimevidence"
	"github.com/mendahu/provenencia/core/datevalue/datevaluetest"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/propertycompare"
	"github.com/mendahu/provenencia/core/propertycompare/propertycomparetest"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/subjectcompare/subjectcomparetest"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

type (
	vals = subjectcompare.Values
	pv   = propertycompare.Value
)

var (
	name = propertycomparetest.Name
	term = propertycomparetest.Term
	ref  = propertycomparetest.Ref
)

func date(y int) pv { return propertycomparetest.Date(datevaluetest.YMD(y, 0, 0)) }

// member is a member whose Subject and Source are one byte each, and whose
// values carry the given ref.
func member(id byte, v vals) claimevidence.Member {
	return claimevidence.Member{Subject: []byte{id}, Source: []byte{0xA0 + id}, Values: v}
}

// seedOrder is the seeded binding order of each primary kind.
func seedOrder() map[vocabkey.SubjectType][]vocabkey.PropertyKey {
	return map[vocabkey.SubjectType][]vocabkey.PropertyKey{
		vocabkey.Person: {"name", "likeness", "sex_at_birth", "gender", "interests"},
		vocabkey.Event:  {"event_type", "event_name", "date"},
	}
}

func sections(kind vocabkey.SubjectType, in vals, ms ...claimevidence.Member) []claimevidence.Section {
	return sectionsOrdered(kind, seedOrder()[kind], in, ms...)
}

func sectionsOrdered(kind vocabkey.SubjectType, order []vocabkey.PropertyKey, in vals, ms ...claimevidence.Member) []claimevidence.Section {
	reg := matchregistry.Default()
	return claimevidence.Sections(kind, order, in, ms, subjectcomparetest.Metas(), reg.Evidence, reg.Subject, reg.Value)
}

func depiction(id byte, r byte) pv { return ref(propertycomparetest.Depiction(id), r) }

func keysOf(got []claimevidence.Section) []vocabkey.PropertyKey {
	var out []vocabkey.PropertyKey
	for _, s := range got {
		out = append(out, s.Key)
	}
	return out
}

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

// pins lists each pin's member Subject byte.
func pins(l claimevidence.Line) []byte {
	var out []byte
	for _, p := range l.Members {
		out = append(out, p.Subject[0])
	}
	return out
}

// only returns the one line of the one section with key, or fails.
func only(t *testing.T, got []claimevidence.Section, key vocabkey.PropertyKey) claimevidence.Line {
	t.Helper()
	if len(got) != 1 || got[0].Key != key || len(got[0].Lines) != 1 {
		t.Fatalf("sections = %+v, want one %s section with one line", got, key)
	}
	return got[0].Lines[0]
}

func TestSections(t *testing.T) {
	reg := matchregistry.Default()
	nameW := reg.Subject.Weights[vocabkey.Person][vocabkey.PropertyName]
	sexW := reg.Subject.Weights[vocabkey.Person][vocabkey.PropertySexAtBirth]
	dateW := reg.Subject.Weights[vocabkey.Event][vocabkey.PropertyDate]
	same, equivalent, contains, different := valuecompare.Same, valuecompare.Equivalent, valuecompare.Contains, valuecompare.Different

	t.Run("AC-1.1 five members with one value are one checked line", func(t *testing.T) {
		var ms []claimevidence.Member
		for i := byte(1); i <= 5; i++ {
			ms = append(ms, member(i, vals{"name": {ref(name("John Smith"), 0x10+i)}}))
		}
		l := only(t, sections(vocabkey.Person, vals{"name": {ref(name("John Smith"), 1)}}, ms...), "name")
		if l.Relation != same || l.Outcome != propertycompare.OutcomeAgree || !l.Checked || !l.Pinnable || !near(l.Points, nameW.Agree) {
			t.Errorf("line = %+v", l)
		}
		if !bytes.Equal(pins(l), []byte{1, 2, 3, 4, 5}) {
			t.Errorf("members = %v, want 1..5 in order", pins(l))
		}
		for i, p := range l.Members {
			if !bytes.Equal(p.Observation, []byte{0x10 + byte(i) + 1}) || !bytes.Equal(p.Source, []byte{0xA0 + byte(i) + 1}) {
				t.Errorf("pin %d = %+v", i, p)
			}
			if !bytes.Equal(p.Value.Ref, p.Observation) || p.Value.Value.Type != "name" {
				t.Errorf("pin %d value = %+v, want the member's name value", i, p.Value)
			}
		}
	})

	t.Run("AC-1.2 a near spelling is its own unchecked line", func(t *testing.T) {
		got := sections(vocabkey.Person, vals{"name": {ref(name("John Smith"), 1)}},
			member(1, vals{"name": {ref(name("John Smith"), 0x11)}}),
			member(2, vals{"name": {ref(name("John Smith"), 0x12)}}),
			member(3, vals{"name": {ref(name("Jon Smith"), 0x13)}}))
		if len(got) != 1 || len(got[0].Lines) != 2 {
			t.Fatalf("sections = %+v", got)
		}
		a, b := got[0].Lines[0], got[0].Lines[1]
		if a.Relation != same || !a.Checked || !bytes.Equal(pins(a), []byte{1, 2}) {
			t.Errorf("first line = %+v", a)
		}
		if b.Relation != valuecompare.Near || b.Cause != valuecompare.CauseSpelling || !b.Pinnable || b.Checked || !bytes.Equal(pins(b), []byte{3}) {
			t.Errorf("second line = %+v", b)
		}
	})

	t.Run("AC-1.3 a folded name is equivalent and checked", func(t *testing.T) {
		l := only(t, sections(vocabkey.Person, vals{"name": {ref(name("José Smith"), 1)}},
			member(1, vals{"name": {ref(name("Jose Smith"), 0x11)}})), "name")
		if l.Relation != equivalent || l.Cause != valuecompare.CauseFold || !l.Checked || !l.Pinnable {
			t.Errorf("line = %+v", l)
		}
	})

	t.Run("AC-1.4 an initial is contained, pinnable and unchecked", func(t *testing.T) {
		in, theirs := ref(name("J. Smith"), 1), ref(name("James Smith"), 0x11)
		l := only(t, sections(vocabkey.Person, vals{"name": {in}}, member(1, vals{"name": {theirs}})), "name")
		sim := propertycompare.Compare([]pv{in}, []pv{theirs}, subjectcomparetest.Metas()["name"], reg.Value).Similarity()
		if l.Relation != contains || l.Cause != valuecompare.CauseInitial || !l.Pinnable || l.Checked || !near(l.Points, nameW.Agree*sim) {
			t.Errorf("line = %+v, similarity %v", l, sim)
		}
	})

	t.Run("AC-1.5 a different date is shown, not pinnable", func(t *testing.T) {
		l := only(t, sections(vocabkey.Event, vals{"date": {ref(date(1817), 1)}},
			member(1, vals{"date": {ref(date(1830), 0x11)}})), "date")
		if l.Outcome != propertycompare.OutcomeConflict || l.Relation != different || l.Pinnable || l.Checked || !near(l.Points, -dateW.Conflict) {
			t.Errorf("line = %+v", l)
		}
	})

	t.Run("AC-1.6 a member with nothing to compare has no line", func(t *testing.T) {
		l := only(t, sections(vocabkey.Person, vals{"name": {ref(name("John Smith"), 1)}},
			member(1, vals{"name": {ref(name("John Smith"), 0x11)}}),
			member(2, vals{"sex_at_birth": {ref(term("male"), 0x12)}})), "name")
		if !bytes.Equal(pins(l), []byte{1}) {
			t.Errorf("members = %v, want [1]", pins(l))
		}
	})

	t.Run("AC-1.7 a denied name conflicts", func(t *testing.T) {
		l := only(t, sections(vocabkey.Person, vals{"name": {propertycomparetest.Neg(ref(name("John Smith"), 1))}},
			member(1, vals{"name": {ref(name("John Smith"), 0x11)}})), "name")
		if l.Outcome != propertycompare.OutcomeConflict || l.PropertyCause != propertycompare.CauseNegative || l.Pinnable {
			t.Errorf("line = %+v", l)
		}
	})

	t.Run("AC-1.8 a gender lands in the sex at birth section", func(t *testing.T) {
		l := only(t, sections(vocabkey.Person, vals{"gender": {ref(term("man"), 1)}},
			member(1, vals{"sex_at_birth": {ref(term("male"), 0x11)}})), "sex_at_birth")
		if l.Property != "gender" || l.Outcome != propertycompare.OutcomeAgree || !l.Checked || !near(l.Points, sexW.Agree) {
			t.Errorf("line = %+v", l)
		}
	})

	t.Run("AC-1.9 a start date meets the member's start date, not its date", func(t *testing.T) {
		l := only(t, sections(vocabkey.Event, vals{"start_date": {ref(date(1817), 1)}},
			member(1, vals{"start_date": {ref(date(1817), 0x11)}, "date": {ref(date(1817), 0x12)}})), "date")
		if l.Property != "start_date" || l.Relation != same || !l.Checked || !near(l.Points, dateW.Agree) || !bytes.Equal(l.Members[0].Observation, []byte{0x11}) {
			t.Errorf("line = %+v", l)
		}
	})

	t.Run("AC-1.10 a property with no lead has its own section and no points", func(t *testing.T) {
		l := only(t, sections(vocabkey.Event, vals{"event_name": {ref(propertycomparetest.Text("Baptism of John"), 1)}},
			member(1, vals{"event_name": {ref(propertycomparetest.Text("Baptism of John"), 0x11)}})), "event_name")
		if !l.Checked || l.Points != 0 {
			t.Errorf("line = %+v", l)
		}
	})

	t.Run("AC-1.12 a period built from start and end names no Observation", func(t *testing.T) {
		got := sections(vocabkey.Event, vals{"date": {ref(date(1817), 1)}},
			member(1, vals{"start_date": {ref(date(1817), 0x11)}, "end_date": {ref(date(1817), 0x12)}}))
		if len(got) != 0 {
			t.Errorf("sections = %+v, want none", got)
		}
	})

	t.Run("AC-1.11 a value with no ref gives no line", func(t *testing.T) {
		got := sections(vocabkey.Person, vals{"name": {name("John Smith")}},
			member(1, vals{"name": {ref(name("John Smith"), 0x11)}}))
		if len(got) != 0 {
			t.Errorf("sections = %+v, want none", got)
		}
	})

	t.Run("AC-1.1 likenesses: the same image is checked, another is judged by eye", func(t *testing.T) {
		got := sections(vocabkey.Person, vals{"likeness": {depiction(40, 1)}},
			member(1, vals{"likeness": {depiction(40, 0x11)}}),
			member(2, vals{"likeness": {depiction(200, 0x12)}}))
		if len(got) != 1 || got[0].Key != "likeness" || len(got[0].Lines) != 2 {
			t.Fatalf("sections = %+v", got)
		}
		a, b := got[0].Lines[0], got[0].Lines[1]
		if a.Relation != same || a.Cause != valuecompare.CauseExact || a.Outcome != propertycompare.OutcomeAgree || !a.Checked || !a.Pinnable {
			t.Errorf("same-image line = %+v", a)
		}
		if b.Relation != valuecompare.Unknown || b.Cause != valuecompare.CauseUnjudged || b.Outcome != propertycompare.OutcomeUnknown || !b.Pinnable || b.Checked {
			t.Errorf("other-image line = %+v", b)
		}
	})

	t.Run("AC-1.4 three members with one image are one collapsed line", func(t *testing.T) {
		l := only(t, sections(vocabkey.Person, vals{"likeness": {depiction(40, 1)}},
			member(1, vals{"likeness": {depiction(40, 0x11)}}),
			member(2, vals{"likeness": {depiction(40, 0x12)}}),
			member(3, vals{"likeness": {depiction(40, 0x13)}})), "likeness")
		if !bytes.Equal(pins(l), []byte{1, 2, 3}) {
			t.Errorf("members = %v", pins(l))
		}
	})

	t.Run("an unknown line the researcher does not judge is dropped", func(t *testing.T) {
		in := vals{"likeness": {ref(propertycomparetest.Depiction(40), 1)}}
		got := sections(vocabkey.Person, in, member(1, vals{"likeness": {{Ref: []byte{0x11}}}}))
		if len(got) != 0 {
			t.Errorf("sections = %+v, want none", got)
		}
	})

	t.Run("AC-3.1 person sections follow the binding order", func(t *testing.T) {
		got := sections(vocabkey.Person, vals{
			"interests": {ref(propertycomparetest.Text("sailing"), 1)},
			"likeness":  {depiction(40, 2)},
			"gender":    {ref(term("man"), 3)},
			"name":      {ref(name("John Smith"), 4)},
		}, member(1, vals{
			"interests":    {ref(propertycomparetest.Text("sailing"), 0x11)},
			"likeness":     {depiction(40, 0x12)},
			"sex_at_birth": {ref(term("male"), 0x13)},
			"name":         {ref(name("John Smith"), 0x14)},
		}))
		want := []vocabkey.PropertyKey{"name", "likeness", "sex_at_birth", "interests"}
		if k := keysOf(got); !equalKeys(k, want) {
			t.Errorf("order = %v, want %v", k, want)
		}
	})

	t.Run("AC-3.2 event sections follow the binding order", func(t *testing.T) {
		got := sections(vocabkey.Event, vals{
			"date":       {ref(date(1817), 1)},
			"event_type": {ref(term("birth"), 2)},
		}, member(1, vals{
			"date":       {ref(date(1817), 0x11)},
			"event_type": {ref(term("birth"), 0x12)},
		}))
		if k := keysOf(got); !equalKeys(k, []vocabkey.PropertyKey{"event_type", "date"}) {
			t.Errorf("order = %v", k)
		}
	})

	t.Run("AC-3.4 a section with no binding sorts after the bound ones, by key", func(t *testing.T) {
		in := vals{
			"date":       {ref(date(1817), 1)},
			"event_name": {ref(propertycomparetest.Text("Baptism"), 2)},
		}
		m := member(1, vals{
			"date":       {ref(date(1817), 0x11)},
			"event_name": {ref(propertycomparetest.Text("Baptism"), 0x12)},
		})
		got := sectionsOrdered(vocabkey.Event, []vocabkey.PropertyKey{"date"}, in, m)
		if k := keysOf(got); !equalKeys(k, []vocabkey.PropertyKey{"date", "event_name"}) {
			t.Errorf("order = %v", k)
		}
		got = sectionsOrdered(vocabkey.Event, nil, in, m)
		if k := keysOf(got); !equalKeys(k, []vocabkey.PropertyKey{"date", "event_name"}) {
			t.Errorf("unordered = %v", k)
		}
	})

	t.Run("a reordered kind lists the moved section first", func(t *testing.T) {
		in := vals{"name": {ref(name("John Smith"), 1)}, "interests": {ref(propertycomparetest.Text("sailing"), 2)}}
		m := member(1, vals{"name": {ref(name("John Smith"), 0x11)}, "interests": {ref(propertycomparetest.Text("sailing"), 0x12)}})
		got := sectionsOrdered(vocabkey.Person, []vocabkey.PropertyKey{"interests", "name"}, in, m)
		if k := keysOf(got); !equalKeys(k, []vocabkey.PropertyKey{"interests", "name"}) {
			t.Errorf("order = %v", k)
		}
	})
}

func equalKeys(a, b []vocabkey.PropertyKey) bool { return slices.Equal(a, b) }
