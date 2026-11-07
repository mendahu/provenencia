package recordlinks_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/datevalue/datevaluetest"
	"github.com/mendahu/provenencia/core/kinship"
	"github.com/mendahu/provenencia/core/kinship/kinshiptest"
	"github.com/mendahu/provenencia/core/matchregistry"
	pt "github.com/mendahu/provenencia/core/propertycompare/propertycomparetest"
	"github.com/mendahu/provenencia/core/recordlinks"
	"github.com/mendahu/provenencia/core/segmentcompare/segmentcomparetest"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/subjectcompare/subjectcomparetest"
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
	m := recordlinks.Member{ID: id, Entity: entity, Kind: vocabkey.Person}
	if len(v) > 0 {
		m.Values = v[0]
	}
	return m
}
func event(id, entity []byte) recordlinks.Member {
	return recordlinks.Member{ID: id, Entity: entity, Kind: vocabkey.Event}
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
