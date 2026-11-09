package recordlinks_test

import (
	"bytes"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/mendahu/provenencia/core/datevalue/datevaluetest"
	"github.com/mendahu/provenencia/core/kinship"
	"github.com/mendahu/provenencia/core/kinship/kinshiptest"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/namevalue/namevaluetest"
	"github.com/mendahu/provenencia/core/propertycompare"
	pt "github.com/mendahu/provenencia/core/propertycompare/propertycomparetest"
	"github.com/mendahu/provenencia/core/recordlinks"
	"github.com/mendahu/provenencia/core/segmentcompare/segmentcomparetest"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/subjectcompare/subjectcomparetest"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

const (
	name     = vocabkey.PropertyName
	sex      = vocabkey.PropertySexAtBirth
	gender   = vocabkey.PropertyGender
	religion = vocabkey.PropertyKey("religion")
	birth    = vocabkey.EventBirth
	marriage = vocabkey.EventMarriage
)

// Entities: the page's person, then the people and events the records share.
func ent(n byte) []byte { return []byte{n} }

var (
	entP = ent(1) // the page's own person
	entF = ent(2) // the shared person
	entM = ent(3)
	entS = ent(4)
	entE = ent(20) // an event
)

// Member ids: graph A's start at 0xA0, graph B's at 0xB0; events at +8.
func idA(n byte) []byte { return []byte{0xA0 + n} }
func idB(n byte) []byte { return []byte{0xB0 + n} }
func evA(n byte) []byte { return []byte{0xA8 + n} }
func evB(n byte) []byte { return []byte{0xB8 + n} }
func person(id, entity []byte, v ...subjectcompare.Values) recordlinks.Member {
	m := recordlinks.Member{Subject: subjectcompare.Subject{ID: id, Kind: vocabkey.Person}, Entity: entity}
	if len(v) > 0 {
		m.Values = v[0]
	}
	return m
}
func event(id, entity []byte) recordlinks.Member {
	return recordlinks.Member{Subject: subjectcompare.Subject{ID: id, Kind: vocabkey.Event}, Entity: entity}
}

type vals = subjectcompare.Values

// part is a participation edge with the person at A, recorded by one
// association.
func part(person, event []byte, role vocabkey.BridgeTerm, eventType vocabkey.EventType, assoc byte) kinship.Edge {
	return kinshiptest.Part(person, event, role, eventType, []byte{assoc})
}

// rel is a relationship edge with the person end at A, recorded by one
// association.
func rel(person, relatedTo []byte, term vocabkey.BridgeTerm, assoc byte) kinship.Edge {
	return kinshiptest.Rel(person, relatedTo, term, []byte{assoc})
}

// build indexes a graph the way the loader's caller does: the edges in a
// LinkSet, the implied links added, then NewGraph.
func build(members []recordlinks.Member, edges ...kinship.Edge) recordlinks.Graph {
	ls := kinship.NewLinkSet()
	for _, e := range edges {
		ls.Add(kinship.Link{Edge: e})
	}
	ls.WithImplied(segmentcomparetest.Terms(), matchregistry.Default().Kinship)
	return recordlinks.NewGraph(members, ls.Links())
}

type want struct {
	entity   []byte
	facts    int
	strength recordlinks.Strength
}

func lines(w ...want) []recordlinks.Line {
	var out []recordlinks.Line
	for _, x := range w {
		out = append(out, recordlinks.Line{Entity: x.entity, Facts: x.facts, Strength: x.strength})
	}
	return out
}

const (
	weak   = recordlinks.StrengthWeak
	strong = recordlinks.StrengthStrong
)

func TestLines(t *testing.T) {
	john, mary := pt.Name("John Smith"), pt.Name("Mary Smith")
	male, female := pt.Term("male"), pt.Term("female")
	pageA, pageB := person(idA(1), entP), person(idB(1), entP)

	cases := []struct {
		name     string
		a, b     recordlinks.Graph
		strongAt int                  // 0 keeps the shipped registry
		metas    subjectcompare.Metas // nil keeps the seeded metas
		want     []recordlinks.Line
	}{
		{"AC-1.1",
			build([]recordlinks.Member{pageA, person(idA(2), entF), person(idA(3), entM)}),
			build([]recordlinks.Member{pageB, person(idB(2), entF), person(idB(3), entS)}),
			0, nil, lines(want{entF, 0, weak})},
		{"AC-1.2",
			build([]recordlinks.Member{pageA, person(idA(2), entF, vals{name: {john}, sex: {male}})}),
			build([]recordlinks.Member{pageB,
				person(idB(2), entF, vals{name: {john}}),
				person(idB(3), entF, vals{name: {john}, sex: {male}})}),
			0, nil, lines(want{entF, 2, strong})},
		{"AC-1.3",
			build([]recordlinks.Member{pageA, person(idA(2), entF, vals{name: {john}})}),
			build([]recordlinks.Member{pageB, person(idB(2), entF, vals{name: {mary}})}),
			0, nil, lines(want{entF, 0, weak})},
		{"AC-1.4 unfiled members",
			build([]recordlinks.Member{pageA, person(idA(2), nil, vals{name: {john}})}),
			build([]recordlinks.Member{pageB, person(idB(2), nil, vals{name: {john}})}),
			0, nil, nil},
		{"AC-1.4 unfiled far end",
			build([]recordlinks.Member{pageA, person(idA(2), entF), event(evA(1), nil)},
				part(idA(2), evA(1), "subject", birth, 1)),
			build([]recordlinks.Member{pageB, person(idB(2), entF), event(evB(1), nil)},
				part(idB(2), evB(1), "subject", birth, 1)),
			0, nil, lines(want{entF, 0, weak})},
		{"AC-1.5",
			build([]recordlinks.Member{pageA, person(idA(2), entF)}),
			build([]recordlinks.Member{pageB, event(idB(2), entF)}),
			0, nil, nil},
		{"AC-2.1",
			build([]recordlinks.Member{pageA, person(idA(2), entF, vals{name: {john}})}),
			build([]recordlinks.Member{pageB, person(idB(2), entF, vals{name: {john}})}),
			0, nil, lines(want{entF, 1, weak})},
		{"AC-2.2",
			build([]recordlinks.Member{pageA, person(idA(2), entF, vals{name: {pt.Given("Amy")}})}),
			build([]recordlinks.Member{pageB, person(idB(2), entF, vals{name: {pt.Name("Amy Lee Anne Long")}})}),
			0, nil, lines(want{entF, 1, weak})},
		{"AC-2.3",
			build([]recordlinks.Member{pageA, person(idA(2), entF, vals{name: {john}})}),
			build([]recordlinks.Member{pageB, person(idB(2), entF, vals{name: {mary}})}),
			0, nil, lines(want{entF, 0, weak})},
		{"AC-2.4",
			build([]recordlinks.Member{pageA, person(idA(2), entF, vals{gender: {pt.Term("man")}})}),
			build([]recordlinks.Member{pageB, person(idB(2), entF, vals{sex: {male}})}),
			0, nil, lines(want{entF, 1, weak})},
		{"AC-2.5",
			build([]recordlinks.Member{pageA, person(idA(2), entF, vals{gender: {pt.Term("man")}, sex: {male}})}),
			build([]recordlinks.Member{pageB, person(idB(2), entF, vals{gender: {pt.Term("man")}, sex: {male}})}),
			0, nil, lines(want{entF, 1, weak})},
		{"AC-2.6",
			build([]recordlinks.Member{pageA, person(idA(2), entF, vals{sex: {male}})}),
			build([]recordlinks.Member{pageB, person(idB(2), entF, vals{sex: {female}})}),
			0, nil, lines(want{entF, 0, weak})},
		{"AC-2.7",
			build([]recordlinks.Member{pageA, person(idA(2), entF, vals{religion: {pt.Term("catholic")}})}),
			build([]recordlinks.Member{pageB, person(idB(2), entF, vals{religion: {pt.Term("catholic"), pt.Term("anglican")}})}),
			0, nil, lines(want{entF, 1, weak})},
		{"AC-2.8",
			build([]recordlinks.Member{pageA, person(idA(2), entF, vals{
				vocabkey.PropertyLifespan: {pt.Date(datevaluetest.YMD(1900, 0, 0))}, vocabkey.PropertyLifespanSpan: {pt.Date(datevaluetest.YMD(1900, 0, 0))}})}),
			build([]recordlinks.Member{pageB, person(idB(2), entF, vals{
				vocabkey.PropertyLifespan: {pt.Date(datevaluetest.YMD(1900, 0, 0))}, vocabkey.PropertyLifespanSpan: {pt.Date(datevaluetest.YMD(1900, 0, 0))}})}),
			0, withDerivedMetas(), lines(want{entF, 0, weak})},
		{"AC-2.9",
			build([]recordlinks.Member{pageA, person(idA(2), entF), event(evA(1), entE)}, part(idA(2), evA(1), "subject", birth, 1)),
			build([]recordlinks.Member{pageB, person(idB(2), entF), event(evB(1), entE)}, part(idB(2), evB(1), "subject", birth, 1)),
			0, nil, lines(want{entF, 1, weak})},
		{"AC-2.10",
			build([]recordlinks.Member{pageA, person(idA(2), entF), event(evA(1), entE)}, part(idA(2), evA(1), "subject", birth, 1)),
			build([]recordlinks.Member{pageB, person(idB(2), entF), event(evB(1), entE)}, part(idB(2), evB(1), "witness", birth, 1)),
			0, nil, lines(want{entF, 0, weak})},
		{"AC-2.11 two entities",
			build([]recordlinks.Member{pageA, person(idA(2), entF), event(evA(1), entE)}, part(idA(2), evA(1), "subject", birth, 1)),
			build([]recordlinks.Member{pageB, person(idB(2), entF), event(evB(1), ent(21))}, part(idB(2), evB(1), "subject", birth, 1)),
			0, nil, lines(want{entF, 0, weak})},
		{"AC-2.11 unfiled",
			build([]recordlinks.Member{pageA, person(idA(2), entF), event(evA(1), entE)}, part(idA(2), evA(1), "subject", birth, 1)),
			build([]recordlinks.Member{pageB, person(idB(2), entF), event(evB(1), nil)}, part(idB(2), evB(1), "subject", birth, 1)),
			0, nil, lines(want{entF, 0, weak})},
		{"AC-2.12",
			build([]recordlinks.Member{pageA, person(idA(2), entF)}, rel(idA(2), idA(1), "parent", 1)),
			build([]recordlinks.Member{pageB, person(idB(2), entF)}, rel(idB(2), idB(1), "parent", 1)),
			0, nil, lines(want{entF, 1, weak})},
		{"AC-2.13",
			build([]recordlinks.Member{pageA, person(idA(2), entF)}, rel(idA(2), idA(1), "spouse", 1)),
			build([]recordlinks.Member{pageB, person(idB(2), entF), event(evB(1), entM)},
				part(idB(2), evB(1), "subject", marriage, 1), part(idB(1), evB(1), "subject", marriage, 2)),
			0, nil, lines(want{entF, 1, weak})},
		{"AC-2.14",
			build([]recordlinks.Member{pageA, person(idA(2), entF), event(evA(1), entM)},
				part(idA(2), evA(1), "subject", marriage, 1), part(idA(1), evA(1), "subject", marriage, 2)),
			build([]recordlinks.Member{pageB, person(idB(2), entF), event(evB(1), entM)},
				part(idB(2), evB(1), "subject", marriage, 1), part(idB(1), evB(1), "subject", marriage, 2)),
			0, nil, lines(want{entF, 1, weak})},
		{"AC-2.15",
			build([]recordlinks.Member{pageA, person(idA(2), entF), event(evA(1), entM)},
				part(idA(2), evA(1), "subject", marriage, 1), part(idA(1), evA(1), "subject", marriage, 2), rel(idA(2), idA(1), "spouse", 3)),
			build([]recordlinks.Member{pageB, person(idB(2), entF), event(evB(1), entM)},
				part(idB(2), evB(1), "subject", marriage, 1), part(idB(1), evB(1), "subject", marriage, 2), rel(idB(2), idB(1), "spouse", 3)),
			0, nil, lines(want{entF, 1, weak})},
		{"AC-2.16",
			build([]recordlinks.Member{pageA, person(idA(2), entF), event(evA(1), entE)},
				part(idA(2), evA(1), "father", birth, 1), part(idA(1), evA(1), "subject", birth, 2)),
			build([]recordlinks.Member{pageB, person(idB(2), entF), event(evB(1), entE)},
				part(idB(2), evB(1), "father", birth, 1), part(idB(1), evB(1), "subject", birth, 2)),
			0, nil, lines(want{entF, 1, weak})},
		{"AC-2.17",
			build([]recordlinks.Member{pageA, person(idA(2), entF, vals{name: {john}}), event(evA(1), entE)}, part(idA(2), evA(1), "subject", birth, 1)),
			build([]recordlinks.Member{pageB, person(idB(2), entF, vals{name: {john}}), event(evB(1), entE)}, part(idB(2), evB(1), "subject", birth, 1)),
			0, nil, lines(want{entF, 2, strong})},
		{"AC-2.18",
			build([]recordlinks.Member{pageA, person(idA(2), entF), event(evA(1), nil)},
				part(idA(2), evA(1), "subject", marriage, 1), part(idA(1), evA(1), "subject", marriage, 2)),
			build([]recordlinks.Member{pageB, person(idB(2), entF), event(evB(1), entM)},
				part(idB(2), evB(1), "subject", marriage, 1), part(idB(1), evB(1), "subject", marriage, 2)),
			0, nil, lines(want{entF, 1, weak})},
		{"AC-2.19",
			build([]recordlinks.Member{pageA, person(idA(2), entF), event(evA(1), entM), event(evA(2), nil)},
				part(idA(2), evA(1), "subject", marriage, 1), part(idA(1), evA(1), "subject", marriage, 2),
				part(idA(2), evA(2), "subject", marriage, 3), part(idA(1), evA(2), "subject", marriage, 4)),
			build([]recordlinks.Member{pageB, person(idB(2), entF), event(evB(1), entM)},
				part(idB(2), evB(1), "subject", marriage, 1), part(idB(1), evB(1), "subject", marriage, 2)),
			0, nil, lines(want{entF, 1, weak})},
		{"D-3 first member wins",
			build([]recordlinks.Member{pageA, person(idA(2), entF, vals{name: {john}}), person(idA(2), entS, vals{name: {john}})}),
			build([]recordlinks.Member{pageB, person(idB(2), entF, vals{name: {john}}), person(idB(3), entS, vals{name: {john}})}),
			0, nil, lines(want{entF, 1, weak})},
		{"D-7 both pivots unfiled",
			build([]recordlinks.Member{pageA, person(idA(2), entF), event(evA(1), nil)},
				part(idA(2), evA(1), "subject", marriage, 1), part(idA(1), evA(1), "subject", marriage, 2)),
			build([]recordlinks.Member{pageB, person(idB(2), entF), event(evB(1), nil)},
				part(idB(2), evB(1), "subject", marriage, 1), part(idB(1), evB(1), "subject", marriage, 2)),
			0, nil, lines(want{entF, 1, weak})},
		{"D-7 two links one key one entity",
			build([]recordlinks.Member{pageA, person(idA(2), entF), event(evA(1), entE), event(evA(2), entE)},
				part(idA(2), evA(1), "subject", birth, 1), part(idA(2), evA(2), "subject", birth, 2)),
			build([]recordlinks.Member{pageB, person(idB(2), entF), event(evB(1), entE)},
				part(idB(2), evB(1), "subject", birth, 1)),
			0, nil, lines(want{entF, 1, weak})},
		{"D-7 any pair not skipped",
			build([]recordlinks.Member{pageA, person(idA(3), entP), person(idA(2), entF), event(evA(1), entM), event(evA(2), nil)},
				part(idA(2), evA(1), "subject", marriage, 1), part(idA(1), evA(1), "subject", marriage, 2),
				part(idA(2), evA(2), "subject", marriage, 3), part(idA(3), evA(2), "subject", marriage, 4)),
			build([]recordlinks.Member{pageB, person(idB(2), entF), event(evB(1), entM)},
				part(idB(2), evB(1), "subject", marriage, 1), part(idB(1), evB(1), "subject", marriage, 2)),
			0, nil, lines(want{entF, 2, strong})},
		{"empty graphs", recordlinks.Graph{}, recordlinks.Graph{}, 0, nil, nil},
		{"decided: a gender no pooled term maps counts nothing",
			build([]recordlinks.Member{pageA, person(idA(2), entF, vals{gender: {pt.Term("non_binary")}})}),
			build([]recordlinks.Member{pageB, person(idB(2), entF, vals{gender: {pt.Term("non_binary")}})}),
			0, nil, lines(want{entF, 0, weak})},
		{"pivot on one entity, roles differ",
			build([]recordlinks.Member{pageA, person(idA(2), entF), event(evA(1), entE)},
				part(idA(2), evA(1), "father", birth, 1), part(idA(1), evA(1), "subject", birth, 2)),
			build([]recordlinks.Member{pageB, person(idB(2), entF), event(evB(1), entE)},
				part(idB(2), evB(1), "mother", birth, 1), part(idB(1), evB(1), "subject", birth, 2)),
			0, nil, lines(want{entF, 1, weak})},
		{"each pivot's entity is counted on its own",
			build([]recordlinks.Member{pageA, person(idA(2), entF), event(evA(1), entM), event(evA(2), entE)},
				part(idA(2), evA(1), "subject", marriage, 1), part(idA(1), evA(1), "subject", marriage, 2),
				part(idA(2), evA(2), "father", birth, 3), part(idA(1), evA(2), "subject", birth, 4)),
			build([]recordlinks.Member{pageB, person(idB(2), entF), event(evB(1), entM), event(evB(2), entE)},
				part(idB(2), evB(1), "subject", marriage, 1), part(idB(1), evB(1), "subject", marriage, 2),
				part(idB(2), evB(2), "mother", birth, 3), part(idB(1), evB(2), "subject", birth, 4)),
			0, nil, lines(want{entF, 2, strong})},
		{"pivot on one entity, roles differ, parent also written down",
			build([]recordlinks.Member{pageA, person(idA(2), entF), event(evA(1), entE)},
				part(idA(2), evA(1), "father", birth, 1), part(idA(1), evA(1), "subject", birth, 2), rel(idA(2), idA(1), "parent", 3)),
			build([]recordlinks.Member{pageB, person(idB(2), entF), event(evB(1), entE)},
				part(idB(2), evB(1), "mother", birth, 1), part(idB(1), evB(1), "subject", birth, 2), rel(idB(2), idB(1), "parent", 3)),
			0, nil, lines(want{entF, 1, weak})},
		{"chain pivot filed on one entity, parent links correspond",
			build([]recordlinks.Member{pageA, person(idA(2), entF), person(idA(3), entS)},
				rel(idA(2), idA(3), "parent", 1), rel(idA(3), idA(1), "parent", 2)),
			build([]recordlinks.Member{pageB, person(idB(2), entF), person(idB(3), entS)},
				rel(idB(2), idB(3), "parent", 1), rel(idB(3), idB(1), "parent", 2)),
			0, nil, lines(want{entF, 1, weak}, want{entS, 2, strong})},
		{"chain pivot filed on one entity, parent links do not correspond",
			build([]recordlinks.Member{pageA, person(idA(2), entF), person(idA(3), entS)},
				rel(idA(2), idA(3), "parent", 1), rel(idA(3), idA(1), "parent", 2)),
			recordlinks.NewGraph([]recordlinks.Member{pageB, person(idB(2), entF), person(idB(3), entS)}, []kinship.Link{{
				Edge:      kinshiptest.Rel(idB(2), idB(1), "grandparent", []byte{0xC1}),
				Recording: kinship.Recording{Shapes: []kinship.Shape{{Pattern: kinship.PatternParentOfParent, Pivot: idB(3)}}, OnlyImplied: true},
			}}),
			0, nil, lines(want{entF, 1, weak}, want{entS, 0, weak})},
		{"AC-3.1",
			build([]recordlinks.Member{pageA,
				person(idA(2), ent(10)),
				person(idA(3), ent(11), vals{name: {john}}),
				person(idA(4), ent(12), vals{name: {john}, sex: {male}}),
				person(idA(5), ent(13), vals{name: {john}, sex: {male}, religion: {pt.Term("catholic")}})}),
			build([]recordlinks.Member{pageB,
				person(idB(2), ent(10)),
				person(idB(3), ent(11), vals{name: {john}}),
				person(idB(4), ent(12), vals{name: {john}, sex: {male}}),
				person(idB(5), ent(13), vals{name: {john}, sex: {male}, religion: {pt.Term("catholic")}})}),
			0, nil, lines(want{ent(10), 0, weak}, want{ent(11), 1, weak}, want{ent(12), 2, strong}, want{ent(13), 3, strong})},
		{"AC-3.2",
			build([]recordlinks.Member{pageA, person(idA(2), entF, vals{name: {john}, sex: {male}})}),
			build([]recordlinks.Member{pageB, person(idB(2), entF, vals{name: {john}, sex: {male}})}),
			3, nil, lines(want{entF, 2, weak})},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			metas := c.metas
			if metas == nil {
				metas = subjectcomparetest.Metas()
			}
			cfg := matchregistry.Default().RecordLinksConfig()
			if c.strongAt != 0 {
				cfg.RecordLinks.StrongAt = c.strongAt
			}
			got := recordlinks.Lines(entP, c.a, c.b, metas, cfg)
			if !equalLines(got, c.want) {
				t.Errorf("Lines = %s; want %s", show(got), show(c.want))
			}
		})
	}
}

func equalLines(a, b []recordlinks.Line) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !bytes.Equal(a[i].Entity, b[i].Entity) || a[i].Facts != b[i].Facts || a[i].Strength != b[i].Strength {
			return false
		}
	}
	return true
}

func show(ls []recordlinks.Line) string {
	var out []string
	for _, l := range ls {
		out = append(out, fmt.Sprintf("{entity %x, %d facts, %s}", l.Entity, l.Facts, l.Strength))
	}
	return "[" + strings.Join(out, " ") + "]"
}

// withDerivedMetas is the seeded metas with the derived lifespan keys given a
// date meta, so only the rule that skips them keeps them from counting.
func withDerivedMetas() subjectcompare.Metas {
	metas := subjectcomparetest.Metas()
	date := metas[vocabkey.PropertyDate]
	metas[vocabkey.PropertyLifespan] = date
	metas[vocabkey.PropertyLifespanSpan] = date
	return metas
}

// TestLinesBounds holds the two registry bounds: members of any kind per
// entity and values per property. Each bound is the largest size that still
// compares as usual; one more counts nothing.
func TestLinesBounds(t *testing.T) {
	cfg := matchregistry.Default().RecordLinksConfig()
	maxMembers, maxValues := cfg.RecordLinks.MaxEntityMembers, cfg.RecordLinks.MaxUnitValues
	john := pt.Name("John Smith")
	male := pt.Term("male")
	pageA, pageB := person(idA(1), entP), person(idB(1), entP)
	repeat := func(v propertycompare.Value, n int) []propertycompare.Value {
		out := make([]propertycompare.Value, n)
		for i := range out {
			out[i] = v
		}
		return out
	}
	// persons are n people on entF named John, with ids from 2 up in one graph.
	persons := func(id func(byte) []byte, n int) []recordlinks.Member {
		var out []recordlinks.Member
		for i := range n {
			out = append(out, person(id(byte(2+i)), entF, vals{name: {john}}))
		}
		return out
	}
	// events are n events on entity e, with ids from 1 up in one graph.
	events := func(id func(byte) []byte, e []byte, n int) []recordlinks.Member {
		var out []recordlinks.Member
		for i := range n {
			out = append(out, event(id(byte(1+i)), e))
		}
		return out
	}
	with := func(a []recordlinks.Member, rest ...[]recordlinks.Member) []recordlinks.Member {
		for _, r := range rest {
			a = append(a, r...)
		}
		return a
	}
	one := func(m ...recordlinks.Member) []recordlinks.Member { return m }

	cases := []struct {
		name string
		a, b recordlinks.Graph
		want []recordlinks.Line
	}{
		{"members at the bound in A",
			build(with(one(pageA), persons(idA, maxMembers))), build(with(one(pageB), persons(idB, 1))),
			lines(want{entF, 1, weak})},
		{"members over the bound in A",
			build(with(one(pageA), persons(idA, maxMembers+1))), build(with(one(pageB), persons(idB, 1))),
			lines(want{entF, 0, weak})},
		{"members over the bound in B only",
			build(with(one(pageA), persons(idA, 1))), build(with(one(pageB), persons(idB, 1), events(evB, entF, maxMembers))),
			lines(want{entF, 0, weak})},
		{"members at the bound in B, events counted",
			build(with(one(pageA), persons(idA, 1))), build(with(one(pageB), persons(idB, 1), events(evB, entF, maxMembers-1))),
			lines(want{entF, 1, weak})},
		{"far entity at the bound counts its link",
			build(with(one(pageA, person(idA(2), entF, vals{name: {john}})), events(evA, entE, maxMembers)),
				part(idA(2), evA(1), "subject", birth, 1)),
			build(one(pageB, person(idB(2), entF, vals{name: {john}}), event(evB(1), entE)),
				part(idB(2), evB(1), "subject", birth, 1)),
			lines(want{entF, 2, strong})},
		{"far entity over the bound counts no link",
			build(with(one(pageA, person(idA(2), entF, vals{name: {john}})), events(evA, entE, maxMembers+1)),
				part(idA(2), evA(1), "subject", birth, 1)),
			build(one(pageB, person(idB(2), entF, vals{name: {john}}), event(evB(1), entE)),
				part(idB(2), evB(1), "subject", birth, 1)),
			lines(want{entF, 1, weak})},
		{"pivot entity over the bound counts its link and does not skip the pivoted one",
			build(with(one(pageA, person(idA(2), entF)), events(evA, entM, maxMembers+1)),
				part(idA(2), evA(1), "subject", marriage, 1), part(idA(1), evA(1), "subject", marriage, 2)),
			build(one(pageB, person(idB(2), entF), event(evB(1), entM)),
				part(idB(2), evB(1), "subject", marriage, 1), part(idB(1), evB(1), "subject", marriage, 2)),
			lines(want{entF, 1, weak})},
		{"values at the bound in A",
			build(one(pageA, person(idA(2), entF, vals{name: repeat(john, maxValues), sex: {male}}))),
			build(one(pageB, person(idB(2), entF, vals{name: {john}, sex: {male}}))),
			lines(want{entF, 2, strong})},
		{"values over the bound in A, other properties still count",
			build(one(pageA, person(idA(2), entF, vals{name: repeat(john, maxValues+1), sex: {male}}))),
			build(one(pageB, person(idB(2), entF, vals{name: {john}, sex: {male}}))),
			lines(want{entF, 1, weak})},
		{"values over the bound in B",
			build(one(pageA, person(idA(2), entF, vals{name: {john}, sex: {male}}))),
			build(one(pageB, person(idB(2), entF, vals{name: repeat(john, maxValues+1), sex: {male}}))),
			lines(want{entF, 1, weak})},
		{"pooled values add up at the bound",
			build(one(pageA, person(idA(2), entF, vals{gender: repeat(pt.Term("man"), maxValues-maxValues/2), sex: repeat(male, maxValues/2)}))),
			build(one(pageB, person(idB(2), entF, vals{sex: {male}}))),
			lines(want{entF, 1, weak})},
		{"pooled values add up over the bound",
			build(one(pageA, person(idA(2), entF, vals{gender: repeat(pt.Term("man"), maxValues-maxValues/2+1), sex: repeat(male, maxValues/2)}))),
			build(one(pageB, person(idB(2), entF, vals{sex: {male}}))),
			lines(want{entF, 0, weak})},
	}
	metas := subjectcomparetest.Metas()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			for _, swap := range []bool{false, true} {
				a, b := c.a, c.b
				if swap {
					a, b = b, a
				}
				if got := recordlinks.Lines(entP, a, b, metas, cfg); !equalLines(got, c.want) {
					t.Errorf("Lines (swapped %v) = %s; want %s", swap, show(got), show(c.want))
				}
			}
		})
	}
}

const interests = vocabkey.PropertyKey("interests")

// words returns n distinct words, tagged so no two tags share one.
func words(tag string, n int) string {
	w := make([]string, n)
	for i := range w {
		w[i] = fmt.Sprintf("%s%d", tag, i)
	}
	return strings.Join(w, " ")
}

// axis is one way a pair of graphs can be costly, at a size n. The graphs
// differ in every long value, so the long values compare at full cost, and
// every shared person agrees on sex, so a pair that compares counts a fact for
// each.
type axis struct {
	name  string
	most  int // the largest n that stays inside the other bounds
	build func(n int) (a, b recordlinks.Graph)
}

// pair builds both graphs from one function of the graph's tag.
func pair(graph func(tag string, id func(byte) []byte, ev func(byte) []byte) recordlinks.Graph) (a, b recordlinks.Graph) {
	return graph("a", idA, evA), graph("b", idB, evB)
}

// soloGraph is one graph with one shared person on entF, holding n values of
// the given words for key; the tag keeps one graph's words from another's.
func soloGraph(key vocabkey.PropertyKey, text func(string) propertycompare.Value, wordsEach, n int) func(tag string, id, ev func(byte) []byte) recordlinks.Graph {
	return func(tag string, id, _ func(byte) []byte) recordlinks.Graph {
		v := vals{sex: {pt.Term("male")}}
		for j := range n {
			v[key] = append(v[key], text(words(fmt.Sprintf("%s%d_", tag, j), wordsEach)))
		}
		return build([]recordlinks.Member{person(id(1), entP), person(id(2), entF, v)})
	}
}

// solo is a pair of soloGraphs.
func solo(key vocabkey.PropertyKey, text func(string) propertycompare.Value, wordsEach int) func(n int) (recordlinks.Graph, recordlinks.Graph) {
	return func(n int) (recordlinks.Graph, recordlinks.Graph) {
		return pair(soloGraph(key, text, wordsEach, n))
	}
}

// crowd is people shared people, each on its own entity, with members members
// in each graph, each with one name of the longest exact pairing.
func crowd(cfg recordlinks.Config, people, members int) (a, b recordlinks.Graph) {
	return pair(func(tag string, id, _ func(byte) []byte) recordlinks.Graph {
		ms := []recordlinks.Member{person(id(1), entP)}
		for i := range people {
			for j := range members {
				ms = append(ms, person(append(id(byte(2+j)), byte(i), byte(i>>8)), []byte{100, byte(i), byte(i >> 8)}, vals{
					sex:  {pt.Term("male")},
					name: {pt.Name(words(fmt.Sprintf("%s%d_%d_", tag, i, j), cfg.Value.Words.MaxExactPairing))}}))
			}
		}
		return build(ms)
	})
}

// marriages is two people married to each other n times, each marriage an event
// on its own entity.
func marriages(n int) (a, b recordlinks.Graph) {
	return pair(func(_ string, id, ev func(byte) []byte) recordlinks.Graph {
		ms := []recordlinks.Member{person(id(1), entP), person(id(2), entF), person(id(3), entS)}
		var edges []kinship.Edge
		for i := range n {
			e := append(ev(0), byte(i), byte(i>>8))
			ms = append(ms, event(e, []byte{120, byte(i), byte(i >> 8)}))
			edges = append(edges, part(id(2), e, "subject", marriage, 1), part(id(3), e, "subject", marriage, 2))
		}
		return build(ms, edges...)
	})
}

// longest is one shared person with the most members, each holding n one-word
// names of the most bytes MaxValueBytes allows (a name's form and its part both
// count), and agreeing on sex.
func longest(cfg recordlinks.Config) func(n int) (recordlinks.Graph, recordlinks.Graph) {
	return func(n int) (recordlinks.Graph, recordlinks.Graph) {
		return pair(func(tag string, id, _ func(byte) []byte) recordlinks.Graph {
			ms := []recordlinks.Member{person(id(1), entP)}
			for j := range cfg.RecordLinks.MaxEntityMembers {
				v := vals{sex: {pt.Term("male")}}
				for k := range n {
					v[name] = append(v[name], pt.Name(sized(fmt.Sprintf("%s%d_%d_", tag, j, k), cfg.RecordLinks.MaxValueBytes/2)))
				}
				ms = append(ms, person(id(byte(2+j)), entF, v))
			}
			return build(ms)
		})
	}
}

// compact returns n distinct three-letter words, the first letter the tag's,
// so a text of Text.MaxWords of them stays within MaxValueBytes; k shifts the
// words so two values differ.
func compact(tag byte, k, n int) string {
	w := make([]string, n)
	for i := range w {
		x := (i + k*n) % 676
		w[i] = string([]byte{tag, 'a' + byte(x/26), 'a' + byte(x%26)})
	}
	return strings.Join(w, " ")
}

// sized is prefix padded with letters to exactly n bytes.
func sized(prefix string, n int) string {
	return prefix + strings.Repeat("z", n-len(prefix))
}

// unsharedLinks is one graph with one shared person on entF who is a witness on
// k events only this graph holds, each event on its own entity: k links to
// entities the other graph lacks. The tag keeps one graph's entities from
// another's.
func unsharedLinks(k int) func(tag string, id, ev func(byte) []byte) recordlinks.Graph {
	return func(tag string, id, ev func(byte) []byte) recordlinks.Graph {
		ms := []recordlinks.Member{person(id(1), entP), person(id(2), entF, vals{sex: {pt.Term("male")}})}
		var edges []kinship.Edge
		for j := range k {
			at := []byte{byte(j), byte(j >> 8)}
			e := append(ev(0), at...)
			ms = append(ms, event(e, append(append([]byte{0xEE}, tag...), at...)))
			edges = append(edges, kinshiptest.Part(id(2), e, "witness", birth, at))
		}
		return build(ms, edges...)
	}
}

func axes(cfg recordlinks.Config) []axis {
	exact, text := cfg.Value.Words.MaxExactPairing, cfg.Value.Text.MaxWords
	return []axis{
		{"names", cfg.RecordLinks.MaxUnitValues, solo(name, pt.Name, exact)},
		{"interests", cfg.RecordLinks.MaxUnitValues, func(n int) (recordlinks.Graph, recordlinks.Graph) {
			return pair(func(tag string, id, _ func(byte) []byte) recordlinks.Graph {
				v := vals{sex: {pt.Term("male")}}
				for k := range n {
					v[interests] = append(v[interests], pt.Text(compact(tag[0], k, text)))
				}
				return build([]recordlinks.Member{person(id(1), entP), person(id(2), entF, v)})
			})
		}},
		{"shared people", 1000, func(n int) (recordlinks.Graph, recordlinks.Graph) { return crowd(cfg, n, 1) }},
		{"shapes", 2000, marriages},
		{"unshared links", 40000, func(n int) (recordlinks.Graph, recordlinks.Graph) { return pair(unsharedLinks(n)) }},
		{"longest values", cfg.RecordLinks.MaxUnitValues, longest(cfg)},
	}
}

// underCap is the largest n on the axis whose pair does not pass the cap.
func underCap(t testing.TB, ax axis, cfg recordlinks.Config) int {
	t.Helper()
	n := 0
	for lo, hi := 1, ax.most; lo <= hi; {
		mid := (lo + hi) / 2
		a, b := ax.build(mid)
		if recordlinks.PairWork(entP, a, b, cfg) <= cfg.RecordLinks.MaxPairWork {
			n, lo = mid, mid+1
		} else {
			hi = mid - 1
		}
	}
	// longest values is stopped by MaxUnitValues, below the cap.
	if n == 0 || n == ax.most && ax.name != "longest values" {
		t.Fatalf("axis %s: the cap sits at %d, outside 1 to %d", ax.name, n, ax.most)
	}
	return n
}

// obituary is the two graphs of a 41-Subject obituary written twice: the
// deceased and forty relatives, each with a three-word name, a sex and a
// religion, linked to the deceased and to the relative before them.
func obituary() (a, b recordlinks.Graph) {
	graph := func(tag byte) recordlinks.Graph {
		id := func(i int) []byte { return []byte{tag, byte(i)} }
		ms := []recordlinks.Member{person(id(100), entP)}
		var edges []kinship.Edge
		terms := []vocabkey.BridgeTerm{"parent", "child", "spouse", "sibling"}
		for i := range 41 {
			ms = append(ms, person(id(i), ent(byte(30+i)), vals{
				name:     {pt.Name(fmt.Sprintf("Given%d Middle%d Family", i, i))},
				sex:      {pt.Term("male")},
				religion: {pt.Term(fmt.Sprintf("faith%d", i%3))},
			}))
			if i > 0 {
				edges = append(edges, kinshiptest.Rel(id(i), id(0), terms[i%4], []byte{byte(i)}))
				if i > 1 {
					edges = append(edges, kinshiptest.Rel(id(i), id(i-1), "parent", []byte{byte(100 + i)}))
				}
			}
		}
		return build(ms, edges...)
	}
	return graph('a'), graph('b')
}

// TestLinesWorkCap holds the cap on the work of one graph pair: a pair over it
// on any axis alone gives every shared person a line with 0 facts, in both
// argument orders and fast; a pair just under it counts as usual, and so does a
// 41-Subject obituary with room to spare.
func TestLinesWorkCap(t *testing.T) {
	cfg := matchregistry.Default().RecordLinksConfig()
	metas := subjectcomparetest.Metas()
	check := func(t *testing.T, a, b recordlinks.Graph, people int, counts bool) {
		t.Helper()
		for _, swap := range []bool{false, true} {
			x, y := a, b
			if swap {
				x, y = y, x
			}
			start := time.Now()
			got := recordlinks.Lines(entP, x, y, metas, cfg)
			if took := time.Since(start); !counts && took > time.Second {
				t.Errorf("Lines (swapped %v) took %v over the cap", swap, took)
			}
			if len(got) != people {
				t.Fatalf("Lines (swapped %v) = %d lines; want %d", swap, len(got), people)
			}
			for _, l := range got {
				if counts != (l.Facts > 0) {
					t.Errorf("Lines (swapped %v): %x has %d facts; counts %v", swap, l.Entity, l.Facts, counts)
				}
			}
			if counts {
				return // an order is enough; TestLinesInvariants holds the symmetry
			}
		}
	}
	for _, ax := range axes(cfg) {
		t.Run(ax.name, func(t *testing.T) {
			n := underCap(t, ax, cfg)
			people := map[string]int{"shared people": n, "shapes": 2}[ax.name]
			people = max(people, 1)
			a, b := ax.build(n)
			check(t, a, b, people, true)
			if ax.name == "longest values" {
				return // MaxUnitValues stops it below the cap, and its own test holds that
			}
			a, b = ax.build(n + 1)
			if ax.name == "shared people" {
				people = n + 1
			}
			check(t, a, b, people, false)
		})
	}
	t.Run("reviewer's inputs", func(t *testing.T) {
		most := cfg.RecordLinks.MaxEntityMembers
		a, b := pair(func(tag string, id, _ func(byte) []byte) recordlinks.Graph {
			ms := []recordlinks.Member{person(id(1), entP)}
			for j := range most {
				v := vals{sex: {pt.Term("male")}}
				for k := range cfg.RecordLinks.MaxUnitValues {
					v[interests] = append(v[interests], pt.Text(compact(tag[0], j*cfg.RecordLinks.MaxUnitValues+k, cfg.Value.Text.MaxWords)))
				}
				ms = append(ms, person(id(byte(2+j)), entF, v))
			}
			return build(ms)
		})
		t.Run("20 long interests at 10 by 10 members", func(t *testing.T) { check(t, a, b, 1, false) })
		a, b = crowd(cfg, 64, most)
		t.Run("64 shared people at 10 by 10 members", func(t *testing.T) { check(t, a, b, 64, false) })
		a, b = marriages(800)
		t.Run("800 marriages", func(t *testing.T) { check(t, a, b, 2, false) })
	})
	t.Run("obituary", func(t *testing.T) {
		a, b := obituary()
		if work, cap := recordlinks.PairWork(entP, a, b, cfg), cfg.RecordLinks.MaxPairWork; work*20 > cap {
			t.Errorf("a 41-Subject obituary does %d work; the cap %d is under 20 times that", work, cap)
		}
		got := recordlinks.Lines(entP, a, b, metas, cfg)
		if len(got) != 41 || got[0].Facts == 0 {
			t.Errorf("Lines = %d lines, first %v; want 41 counted", len(got), got[:1])
		}
	})
}

// TestLinesValueSize holds the limit on one value: a value over MaxValueBytes
// makes its property count nothing, whatever its words, and fast; a value at the
// limit compares as usual. The first three inputs are the ones that got past
// the cap's word count.
func TestLinesValueSize(t *testing.T) {
	cfg := matchregistry.Default().RecordLinksConfig()
	metas := subjectcomparetest.Metas()
	most, limit := cfg.RecordLinks.MaxEntityMembers, cfg.RecordLinks.MaxValueBytes
	male := pt.Term("male")
	// graph is the page's person and, on entF, members members each holding
	// the values per(tag, j) gives.
	graph := func(members int, per func(tag string, j int) vals) func(tag string, id, _ func(byte) []byte) recordlinks.Graph {
		return func(tag string, id, _ func(byte) []byte) recordlinks.Graph {
			ms := []recordlinks.Member{person(id(1), entP)}
			for j := range members {
				ms = append(ms, person(id(byte(2+j)), entF, per(tag, j)))
			}
			return build(ms)
		}
	}
	// twoLetter is 64 distinct two-letter words from eight letters from first.
	twoLetter := func(first byte) string {
		var w []string
		for i := range 8 {
			for j := range 8 {
				w = append(w, string([]byte{first + byte(i), first + byte(j)}))
			}
		}
		return strings.Join(w, " ")
	}
	both := func(tag string) propertycompare.Value {
		first := byte('a')
		if tag == "b" {
			first = 'i'
		}
		return propertycompare.Value{Value: valuecompare.Value{Name: namevaluetest.Western("x"), Text: twoLetter(first)}}
	}
	repeatN := func(n int, f func(k int) propertycompare.Value) []propertycompare.Value {
		var out []propertycompare.Value
		for k := range n {
			out = append(out, f(k))
		}
		return out
	}
	one := func(e []byte, v vals) func(tag string, id, _ func(byte) []byte) recordlinks.Graph {
		return func(tag string, id, _ func(byte) []byte) recordlinks.Graph {
			return build([]recordlinks.Member{person(id(1), entP), person(id(2), e, v)})
		}
	}
	// twoPeople is entF holding vF and entS holding sex, in both graphs.
	twoPeople := func(vF vals) func(tag string, id, _ func(byte) []byte) recordlinks.Graph {
		return func(tag string, id, _ func(byte) []byte) recordlinks.Graph {
			return build([]recordlinks.Member{person(id(1), entP), person(id(2), entF, vF), person(id(3), entS, vals{sex: {male}})})
		}
	}
	cases := []struct {
		name  string
		build func(tag string, id, _ func(byte) []byte) recordlinks.Graph
		want  []recordlinks.Line
	}{
		{"15 one-word 64 KB names at 10 by 10 members",
			graph(most, func(tag string, j int) vals {
				return vals{name: repeatN(15, func(k int) propertycompare.Value {
					return pt.Name(sized(fmt.Sprintf("%s%d_%d_", tag, j, k), 64<<10))
				})}
			}),
			lines(want{entF, 0, weak})},
		{"21 two-word interests of long words at 10 by 10 members",
			graph(most, func(tag string, j int) vals {
				return vals{interests: repeatN(cfg.RecordLinks.MaxUnitValues+1, func(k int) propertycompare.Value {
					p := fmt.Sprintf("%s%d_%d_", tag, j, k)
					return pt.Text(sized(p+"a", 32<<10) + " " + sized(p+"b", 32<<10))
				})}
			}),
			lines(want{entF, 0, weak})},
		{"15 interests of name and 64 words at 10 by 10 members, cap trips",
			graph(most, func(tag string, j int) vals {
				return vals{sex: {male}, interests: repeatN(15, func(int) propertycompare.Value { return both(tag) })}
			}),
			lines(want{entF, 0, weak})},
		{"one 159-word interests on one person, other lines count",
			twoPeople(vals{sex: {male}, interests: {pt.Text(words("w", 159))}}),
			lines(want{entF, 1, weak}, want{entS, 1, weak})},
		{"a text of exactly the limit compares",
			one(entF, vals{interests: {pt.Text(strings.Repeat("a", limit))}}),
			lines(want{entF, 1, weak})},
		{"a text one byte over the limit counts nothing",
			one(entF, vals{interests: {pt.Text(strings.Repeat("a", limit+1))}}),
			lines(want{entF, 0, weak})},
		{"a name of exactly the limit, form and part together, compares",
			one(entF, vals{name: {pt.Name(strings.Repeat("n", limit/2))}}),
			lines(want{entF, 1, weak})},
		{"a name one byte over the limit counts nothing",
			one(entF, vals{name: {pt.Name(strings.Repeat("n", limit/2+1))}}),
			lines(want{entF, 0, weak})},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			a, b := pair(c.build)
			for _, swap := range []bool{false, true} {
				x, y := a, b
				if swap {
					x, y = y, x
				}
				start := time.Now()
				got := recordlinks.Lines(entP, x, y, metas, cfg)
				if took := time.Since(start); took > time.Second {
					t.Errorf("Lines (swapped %v) took %v", swap, took)
				}
				if !equalLines(got, c.want) {
					t.Errorf("Lines (swapped %v) = %s; want %s", swap, show(got), show(c.want))
				}
			}
		})
	}
}

// BenchmarkLinesAtCap is the slowest pair the cap lets through, one
// sub-benchmark for each axis of cost, each at the largest size under the cap.
func BenchmarkLinesAtCap(b *testing.B) {
	cfg := matchregistry.Default().RecordLinksConfig()
	metas := subjectcomparetest.Metas()
	for _, ax := range axes(cfg) {
		b.Run(ax.name, func(b *testing.B) {
			ga, gb := ax.build(underCap(b, ax, cfg))
			b.ResetTimer()
			for range b.N {
				recordlinks.Lines(entP, ga, gb, metas, cfg)
			}
		})
	}
}

// householdGraphs is n household graphs in memory, the same shape as
// persondetail's households fixture, which asserts the same numbers: each
// holds the son (entP, the page), the father (entF), the mother (entM) and the
// sister (entS), named with their names only, and four parent links (father and
// mother to the son, father and mother to the sister). Each pair of graphs then
// has three strong lines (father, mother, sister): 1,225 pairs at n = 50.
func householdGraphs(n int) []recordlinks.Graph {
	out := make([]recordlinks.Graph, n)
	for i := range out {
		id := func(k byte) []byte { return []byte{byte(i), byte(i >> 8), k} }
		out[i] = build([]recordlinks.Member{
			person(id(1), entP, vals{name: {pt.Name("Thomas Smith")}}),
			person(id(2), entF, vals{name: {pt.Name("William Smith")}}),
			person(id(3), entM, vals{name: {pt.Name("Mary Ann Jones")}}),
			person(id(4), entS, vals{name: {pt.Name("Ann Smith")}}),
		},
			rel(id(2), id(1), "parent", 1), rel(id(3), id(1), "parent", 2),
			rel(id(2), id(4), "parent", 3), rel(id(3), id(4), "parent", 4))
	}
	return out
}

func TestPairs(t *testing.T) {
	john, mary := pt.Name("John Smith"), pt.Name("Mary Smith")
	cfg := matchregistry.Default().RecordLinksConfig()
	metas := subjectcomparetest.Metas()
	g := func(id func(byte) []byte, others ...recordlinks.Member) recordlinks.Graph {
		return build(append([]recordlinks.Member{person(id(1), entP)}, others...))
	}
	ida := func(n byte) []byte { return []byte{0xC0 + n} }
	cases := []struct {
		name   string
		graphs []recordlinks.Graph
		want   []recordlinks.Pair
	}{
		{"AC-1.1 a shared person",
			[]recordlinks.Graph{
				g(idA, person(idA(2), entF, vals{name: {john}})),
				g(idB, person(idB(2), entF, vals{name: {john}}))},
			[]recordlinks.Pair{{A: 0, B: 1, Lines: lines(want{entF, 1, weak})}}},
		{"two pairs, ordered",
			[]recordlinks.Graph{
				g(idA, person(idA(2), entF)),
				g(idB, person(idB(2), entM)),
				g(ida, person(ida(2), entF), person(ida(3), entM))},
			[]recordlinks.Pair{
				{A: 0, B: 2, Lines: lines(want{entF, 0, weak})},
				{A: 1, B: 2, Lines: lines(want{entM, 0, weak})}}},
		{"AC-1.2 the page is no line",
			[]recordlinks.Graph{g(idA), g(idB), g(ida)}, nil},
		{"AC-1.7 one graph", []recordlinks.Graph{g(idA, person(idA(2), entF))}, nil},
		{"no graph", nil, nil},
		{"AC-1.7 a graph that shares no one",
			[]recordlinks.Graph{
				g(idA, person(idA(2), entF, vals{name: {mary}})),
				g(idB, person(idB(2), entF, vals{name: {mary}})),
				g(ida, person(ida(2), entS))},
			[]recordlinks.Pair{{A: 0, B: 1, Lines: lines(want{entF, 1, weak})}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, _ := recordlinks.Pairs(entP, c.graphs, metas, cfg)
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("Pairs = %+v, want %+v", got, c.want)
			}
		})
	}
}

// reversed is graphs in the opposite order.
func reversed(graphs []recordlinks.Graph) []recordlinks.Graph {
	out := slices.Clone(graphs)
	slices.Reverse(out)
	return out
}

// mirror maps the pairs of a read to the pairs of the reversed read.
func mirror(pairs []recordlinks.Pair, n int) []recordlinks.Pair {
	var out []recordlinks.Pair
	for _, p := range pairs {
		out = append(out, recordlinks.Pair{A: n - 1 - p.B, B: n - 1 - p.A, Lines: p.Lines})
	}
	slices.SortFunc(out, func(x, y recordlinks.Pair) int {
		if x.A != y.A {
			return x.A - y.A
		}
		return x.B - y.B
	})
	return out
}

func TestPairsReadCap(t *testing.T) {
	cfg := matchregistry.Default().RecordLinksConfig()
	metas := subjectcomparetest.Metas()
	john := pt.Name("John Smith")
	named := func(id func(byte) []byte, others ...recordlinks.Member) recordlinks.Graph {
		return build(append([]recordlinks.Member{person(id(1), entP)}, others...))
	}
	ids := func(base byte) func(byte) []byte { return func(n byte) []byte { return []byte{base + n} } }
	two := []recordlinks.Graph{
		named(idA, person(idA(2), entF, vals{name: {john}, sex: {pt.Term("male")}})),
		named(idB, person(idB(2), entF, vals{name: {john}, sex: {pt.Term("male")}}))}
	three := []recordlinks.Graph{
		named(idA, person(idA(2), entF)), named(idB, person(idB(2), entF)), named(ids(0xC0), person(ids(0xC0)(2), entF))}
	big1, big2 := solo(name, pt.Name, cfg.Value.Words.MaxExactPairing)(underCap(t, axes(cfg)[0], cfg) + 1)
	mixed := []recordlinks.Graph{big1, big2,
		named(ids(0xC0), person(ids(0xC0)(2), entM, vals{name: {john}})),
		named(ids(0xD0), person(ids(0xD0)(2), entM, vals{name: {john}}))}
	if w := recordlinks.PairWork(entP, big1, big2, cfg); w <= cfg.RecordLinks.MaxPairWork {
		t.Fatalf("the big pair's work %d is not over MaxPairWork", w)
	}

	// Sixty-four graphs, each holding the shared person as a witness on 5,000
	// events only that graph holds: 2,016 pairs, none of them over MaxPairWork,
	// whose work is only the links each compare would walk.
	var crafted []recordlinks.Graph
	for i := range 64 {
		tag := fmt.Sprintf("g%d", i)
		id := func(k byte) []byte { return []byte{byte(i), byte(i >> 8), k} }
		ev := func(k byte) []byte { return []byte{0xEE, byte(i), byte(i >> 8), k} }
		crafted = append(crafted, unsharedLinks(5000)(tag, id, ev))
	}

	cases := []struct {
		name   string
		graphs []recordlinks.Graph
		cap    func(work int) int
		check  func(t *testing.T, got []recordlinks.Pair)
	}{
		{"AC-2.5 cap equals the work: counted", two, func(w int) int { return w },
			func(t *testing.T, got []recordlinks.Pair) {
				if want := lines(want{entF, 2, strong}); len(got) != 1 || !equalLines(got[0].Lines, want) {
					t.Fatalf("got %+v", got)
				}
			}},
		{"AC-2.3 cap one short: every pair listed at 0 facts", two, func(w int) int { return w - 1 },
			func(t *testing.T, got []recordlinks.Pair) {
				if want := lines(want{entF, 0, weak}); len(got) != 1 || !equalLines(got[0].Lines, want) {
					t.Fatalf("got %+v", got)
				}
			}},
		{"AC-2.4 lines alone pass the cap: no pair", three, func(int) int { return 2 },
			func(t *testing.T, got []recordlinks.Pair) {
				if got != nil {
					t.Fatalf("got %+v", got)
				}
			}},
		{"AC-2.5 a pair over MaxPairWork beside a small pair", mixed, func(w int) int { return w },
			func(t *testing.T, got []recordlinks.Pair) {
				if len(got) != 2 || got[0].A != 0 || got[0].B != 1 || got[1].A != 2 || got[1].B != 3 ||
					!equalLines(got[0].Lines, lines(want{entF, 0, weak})) ||
					!equalLines(got[1].Lines, lines(want{entM, 1, weak})) {
					t.Fatalf("got %+v", got)
				}
			}},
		{"unshared links: the links each compare walks are work", crafted, func(int) int { return cfg.RecordLinks.MaxReadWork },
			func(t *testing.T, got []recordlinks.Pair) {
				if len(got) != 2016 {
					t.Fatalf("pairs = %d, want 2016", len(got))
				}
				for _, p := range got {
					if len(p.Lines) != 1 || p.Lines[0].Facts != 0 || p.Lines[0].Strength != weak {
						t.Fatalf("pair %d-%d lines = %s; want one line at 0 facts", p.A, p.B, show(p.Lines))
					}
				}
			}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			run := cfg
			run.RecordLinks.MaxReadWork = c.cap(recordlinks.ReadWork(entP, c.graphs, cfg))
			start := time.Now()
			got, _ := recordlinks.Pairs(entP, c.graphs, metas, run)
			if took := time.Since(start); took > time.Second {
				t.Errorf("Pairs took %v", took)
			}
			c.check(t, got)
			back, _ := recordlinks.Pairs(entP, reversed(c.graphs), metas, run)
			if want := mirror(got, len(c.graphs)); !reflect.DeepEqual(back, want) && (len(back) != 0 || len(want) != 0) {
				t.Fatalf("reversed read = %+v, want %+v", back, want)
			}
		})
	}
}

func TestPairsAtScale(t *testing.T) {
	cfg := matchregistry.Default().RecordLinksConfig()
	graphs := householdGraphs(50)
	t.Logf("ReadWork at 50 households = %d (MaxReadWork %d)", recordlinks.ReadWork(entP, graphs, cfg), cfg.RecordLinks.MaxReadWork)
	got, _ := recordlinks.Pairs(entP, graphs, subjectcomparetest.Metas(), cfg)
	if len(got) != 1225 {
		t.Fatalf("pairs = %d, want 1225", len(got))
	}
	for _, p := range got {
		if len(p.Lines) != 3 || p.Lines[0].Strength != strong || p.Lines[1].Strength != strong || p.Lines[2].Strength != strong ||
			!bytes.Equal(p.Lines[0].Entity, entF) || !bytes.Equal(p.Lines[1].Entity, entM) || !bytes.Equal(p.Lines[2].Entity, entS) {
			t.Fatalf("pair %d-%d lines = %s", p.A, p.B, show(p.Lines))
		}
	}
}

// BenchmarkPairsAtCap is the slowest read the cap lets through: one
// sub-benchmark for each of the "names" and "unshared links" axes at its
// largest size under MaxPairWork, in as many graphs as stay within
// MaxReadWork.
func BenchmarkPairsAtCap(b *testing.B) {
	cfg := matchregistry.Default().RecordLinksConfig()
	metas := subjectcomparetest.Metas()
	exact := cfg.Value.Words.MaxExactPairing
	one := map[string]func(n int) func(tag string, id, ev func(byte) []byte) recordlinks.Graph{
		"names": func(n int) func(string, func(byte) []byte, func(byte) []byte) recordlinks.Graph {
			return soloGraph(name, pt.Name, exact, n)
		},
		"unshared links": unsharedLinks,
	}
	for _, ax := range axes(cfg) {
		build, ok := one[ax.name]
		if !ok {
			continue
		}
		b.Run(ax.name, func(b *testing.B) {
			n := underCap(b, ax, cfg)
			graphs := func(m int) []recordlinks.Graph {
				out := make([]recordlinks.Graph, m)
				for i := range out {
					id := func(k byte) []byte { return []byte{byte(i), byte(i >> 8), k} }
					ev := func(k byte) []byte { return []byte{0xEE, byte(i), byte(i >> 8), k} }
					out[i] = build(n)(fmt.Sprintf("g%d", i), id, ev)
				}
				return out
			}
			m := 2
			for recordlinks.ReadWork(entP, graphs(m+1), cfg) <= cfg.RecordLinks.MaxReadWork {
				m++
			}
			gs := graphs(m)
			b.ResetTimer()
			for range b.N {
				_, _ = recordlinks.Pairs(entP, gs, metas, cfg)
			}
			b.ReportMetric(float64(m), "graphs")
		})
	}
}

// TestPairsCounted: Pairs says whether the read's work stayed within MaxReadWork.
func TestPairsCounted(t *testing.T) {
	cfg := matchregistry.Default().RecordLinksConfig()
	metas := subjectcomparetest.Metas()
	graphs := householdGraphs(3) // 3 pairs of 3 strong lines: 9 lines
	work := recordlinks.ReadWork(entP, graphs, cfg)
	read := func(gs []recordlinks.Graph, limit int) ([]recordlinks.Pair, bool) {
		run := cfg
		run.RecordLinks.MaxReadWork = limit
		return recordlinks.Pairs(entP, gs, metas, run)
	}
	t.Run("AC-2.1 at the limit", func(t *testing.T) {
		pairs, counted := read(graphs, work)
		if !counted || len(pairs) != 3 || pairs[0].Lines[0].Facts == 0 {
			t.Fatalf("counted %v, %d pairs, first line %+v", counted, len(pairs), pairs[0].Lines[0])
		}
	})
	t.Run("AC-2.2 past it by pair work", func(t *testing.T) {
		pairs, counted := read(graphs, work-1)
		if counted || len(pairs) != 3 {
			t.Fatalf("counted %v, %d pairs, want not counted and every pair", counted, len(pairs))
		}
		for _, p := range pairs {
			for _, l := range p.Lines {
				if l.Facts != 0 {
					t.Fatalf("a line has %d facts, want 0", l.Facts)
				}
			}
		}
	})
	t.Run("AC-2.3 past it by the lines alone", func(t *testing.T) {
		if pairs, counted := read(graphs, 8); counted || pairs != nil {
			t.Fatalf("counted %v, pairs %v, want not counted and none", counted, pairs)
		}
	})
	t.Run("no graph", func(t *testing.T) {
		if pairs, counted := read(nil, 0); !counted || len(pairs) != 0 {
			t.Fatalf("counted %v, pairs %v", counted, pairs)
		}
	})
}

// TestCapLines: CapLines keeps strong lines first and the rest in read order.
func TestCapLines(t *testing.T) {
	line := func(id byte, s recordlinks.Strength) recordlinks.Line {
		return recordlinks.Line{Entity: []byte{id}, Strength: s}
	}
	pairs := []recordlinks.Pair{
		{A: 0, B: 1, Lines: []recordlinks.Line{line(1, weak), line(2, weak)}},
		{A: 0, B: 2, Lines: []recordlinks.Line{line(3, weak), line(4, weak)}},
		{A: 1, B: 2, Lines: []recordlinks.Line{line(5, weak), line(6, strong)}},
	}
	ids := func(ps []recordlinks.Pair) [][]byte {
		var out [][]byte
		for _, p := range ps {
			for _, l := range p.Lines {
				out = append(out, l.Entity)
			}
		}
		return out
	}
	t.Run("at the most nothing is dropped", func(t *testing.T) {
		got, dropped := recordlinks.CapLines(pairs, 6)
		if dropped || !reflect.DeepEqual(got, pairs) {
			t.Fatalf("dropped %v, %+v", dropped, got)
		}
	})
	t.Run("one fewer drops the last weak line", func(t *testing.T) {
		got, dropped := recordlinks.CapLines(pairs, 5)
		if !dropped || !reflect.DeepEqual(ids(got), [][]byte{{1}, {2}, {3}, {4}, {6}}) {
			t.Fatalf("dropped %v, lines %v", dropped, ids(got))
		}
		if got[2].A != 1 || got[2].B != 2 || len(got[2].Lines) != 1 {
			t.Fatalf("the last pair = %+v", got[2])
		}
	})
	t.Run("strong lines at the end of the read win; empty pairs go", func(t *testing.T) {
		more := append(append([]recordlinks.Pair(nil), pairs...), recordlinks.Pair{A: 2, B: 3, Lines: []recordlinks.Line{line(7, strong)}})
		got, dropped := recordlinks.CapLines(more, 2)
		if !dropped || !reflect.DeepEqual(ids(got), [][]byte{{6}, {7}}) || len(got) != 2 {
			t.Fatalf("dropped %v, pairs %+v", dropped, got)
		}
	})
	t.Run("20 times the cap", func(t *testing.T) {
		const most, pairCount, perPair = 75_000, 1_500, 1_000
		big := make([]recordlinks.Pair, pairCount)
		var strongAt int
		for i := range big {
			lines := make([]recordlinks.Line, perPair)
			for j := range lines {
				lines[j] = recordlinks.Line{Entity: []byte{byte(i), byte(i >> 8), byte(j), byte(j >> 8)}, Strength: weak}
			}
			if i%150 == 0 {
				lines[perPair/2].Strength = strong
				strongAt++
			}
			big[i] = recordlinks.Pair{A: i, B: i + 1, Lines: lines}
		}
		start := time.Now()
		got, dropped := recordlinks.CapLines(big, most)
		if took := time.Since(start); took > time.Second {
			t.Errorf("CapLines took %v, want under a second", took)
		}
		if !dropped {
			t.Fatal("nothing dropped")
		}
		// Expected: every strong line, and the first most-10 weak lines in read order.
		var want []recordlinks.Pair
		weakLeft := most - strongAt
		for _, p := range big {
			var keep []recordlinks.Line
			for _, l := range p.Lines {
				if l.Strength == strong {
					keep = append(keep, l)
				} else if weakLeft > 0 {
					weakLeft--
					keep = append(keep, l)
				}
			}
			if len(keep) > 0 {
				want = append(want, recordlinks.Pair{A: p.A, B: p.B, Lines: keep})
			}
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatal("the kept lines are not the 10 strong and the first 74,990 weak in read order, each at its place in its pair")
		}
		if len(big[0].Lines) != perPair {
			t.Fatal("the input was modified")
		}
	})
}
