package recordlinks_test

import (
	"bytes"
	"math/rand"
	"reflect"
	"slices"
	"testing"

	"github.com/mendahu/provenencia/core/kinship"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/propertycompare"
	pt "github.com/mendahu/provenencia/core/propertycompare/propertycomparetest"
	"github.com/mendahu/provenencia/core/recordlinks"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/subjectcompare/subjectcomparetest"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// A random model of one Evidence graph: who is in it, what they record and
// what links them. Persons are referred to by index.
type personSpec struct {
	entity []byte
	values subjectcompare.Values
}

type partSpec struct {
	person int
	role   vocabkey.BridgeTerm
}

type eventSpec struct {
	entity    []byte
	eventType vocabkey.EventType
	parts     []partSpec
}

type relSpec struct {
	a, b int
	term vocabkey.BridgeTerm
}

type graphSpec struct {
	persons []personSpec
	events  []eventSpec
	rels    []relSpec
}

var (
	invNames  = []string{"John Smith", "Mary Smith", "Amy Lee Long", "Amy", "Robert Brown"}
	invRoles  = []vocabkey.BridgeTerm{"subject", "father", "mother", "witness"}
	invEvents = []vocabkey.EventType{birth, marriage}
)

func genGraph(r *rand.Rand, page []byte, people, events [][]byte) graphSpec {
	var g graphSpec
	for i := range 3 + r.Intn(4) {
		var p personSpec
		switch {
		case i == 0:
			p.entity = page
		case r.Intn(5) == 0:
			p.entity = nil
		default:
			p.entity = people[r.Intn(len(people))]
		}
		p.values = subjectcompare.Values{name: {pt.Name(invNames[r.Intn(len(invNames))])}}
		p.values[sex] = []propertycompare.Value{pt.Term([]string{"male", "female"}[r.Intn(2)])}
		switch r.Intn(3) {
		case 0:
			p.values[gender] = append(p.values[gender], pt.Term("man"))
		case 1:
			p.values[gender] = append(p.values[gender], pt.Term("woman"))
		}
		g.persons = append(g.persons, p)
	}
	for range 1 + r.Intn(3) {
		g.events = append(g.events, genEvent(r, len(g.persons), events))
	}
	for range r.Intn(3) {
		a, b := r.Intn(len(g.persons)), r.Intn(len(g.persons))
		if a == b {
			continue
		}
		g.rels = append(g.rels, relSpec{a, b, []vocabkey.BridgeTerm{"parent", "spouse"}[r.Intn(2)]})
	}
	return g
}

func genEvent(r *rand.Rand, persons int, events [][]byte) eventSpec {
	e := eventSpec{eventType: invEvents[r.Intn(len(invEvents))]}
	if r.Intn(5) != 0 {
		e.entity = events[r.Intn(len(events))]
	}
	for _, p := range r.Perm(persons)[:min(persons, 1+r.Intn(3))] {
		e.parts = append(e.parts, partSpec{p, invRoles[r.Intn(len(invRoles))]})
	}
	return e
}

// copyEvents gives b a's events and participations on the same entities, onto
// b's own persons.
func copyEvents(a graphSpec, b *graphSpec) {
	b.events = nil
	for _, e := range a.events {
		c := eventSpec{entity: e.entity, eventType: e.eventType}
		for _, p := range e.parts {
			c.parts = append(c.parts, partSpec{p.person % len(b.persons), p.role})
		}
		seen := map[int]bool{}
		c.parts = slices.DeleteFunc(c.parts, func(p partSpec) bool {
			dup := seen[p.person]
			seen[p.person] = true
			return dup
		})
		b.events = append(b.events, c)
	}
}

// materialize turns a spec into members and edges with ids from prefix.
func materialize(g graphSpec, prefix, assocPrefix byte) ([]recordlinks.Member, []kinship.Edge) {
	var members []recordlinks.Member
	var edges []kinship.Edge
	personID := func(i int) []byte { return []byte{prefix + byte(i)} }
	var assoc byte
	next := func() byte { assoc++; return assocPrefix + assoc }
	for i, p := range g.persons {
		members = append(members, person(personID(i), p.entity, p.values))
	}
	for i, e := range g.events {
		id := []byte{prefix + 8 + byte(i)}
		members = append(members, event(id, e.entity))
		for _, p := range e.parts {
			edges = append(edges, part(personID(p.person), id, p.role, e.eventType, next()))
		}
	}
	for _, x := range g.rels {
		edges = append(edges, rel(personID(x.a), personID(x.b), x.term, next()))
	}
	return members, edges
}

func TestLinesInvariants(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	cfg := matchregistry.Default().RecordLinksConfig()
	metas := subjectcomparetest.Metas()
	people := [][]byte{ent(1), ent(2), ent(3), ent(4), ent(5), ent(6)}
	events := [][]byte{ent(20), ent(21), ent(22)}
	p := people[0]

	for round := range 200 {
		a := genGraph(r, p, people, events)
		b := genGraph(r, p, people, events)
		if round%2 == 0 {
			copyEvents(a, &b)
		}
		ma, ea := materialize(a, 0xA0, 0xC0)
		mb, eb := materialize(b, 0xB0, 0xD0)
		ga, gb := build(ma, ea...), build(mb, eb...)
		got := recordlinks.Lines(p, ga, gb, metas, cfg)

		// (a) The two graphs are interchangeable.
		if back := recordlinks.Lines(p, gb, ga, metas, cfg); !reflect.DeepEqual(got, back) {
			t.Fatalf("round %d: Lines(a, b) = %s, Lines(b, a) = %s", round, show(got), show(back))
		}

		// (b) Input order does not matter.
		r.Shuffle(len(ma), func(i, j int) { ma[i], ma[j] = ma[j], ma[i] })
		r.Shuffle(len(ea), func(i, j int) { ea[i], ea[j] = ea[j], ea[i] })
		r.Shuffle(len(mb), func(i, j int) { mb[i], mb[j] = mb[j], mb[i] })
		r.Shuffle(len(eb), func(i, j int) { eb[i], eb[j] = eb[j], eb[i] })
		if shuffled := recordlinks.Lines(p, build(ma, ea...), build(mb, eb...), metas, cfg); !reflect.DeepEqual(got, shuffled) {
			t.Fatalf("round %d: shuffled inputs give %s, want %s", round, show(shuffled), show(got))
		}

		// (c) A line for exactly each entity with a person member in both
		// graphs, never the page's or nil.
		var wantEntities [][]byte
		for _, x := range a.persons {
			if x.entity == nil || string(x.entity) == string(p) {
				continue
			}
			if slices.ContainsFunc(b.persons, func(y personSpec) bool { return string(y.entity) == string(x.entity) }) &&
				!slices.ContainsFunc(wantEntities, func(e []byte) bool { return string(e) == string(x.entity) }) {
				wantEntities = append(wantEntities, x.entity)
			}
		}
		slices.SortFunc(wantEntities, func(x, y []byte) int { return int(x[0]) - int(y[0]) })
		var gotEntities [][]byte
		for _, l := range got {
			gotEntities = append(gotEntities, l.Entity)
		}
		if !reflect.DeepEqual(gotEntities, wantEntities) {
			t.Fatalf("round %d: line entities %x, want %x", round, gotEntities, wantEntities)
		}

		// (d) Strength follows the registry.
		for _, l := range got {
			if (l.Strength == recordlinks.StrengthStrong) != (l.Facts >= cfg.RecordLinks.StrongAt) {
				t.Fatalf("round %d: %d facts weigh %s", round, l.Facts, l.Strength)
			}
		}

		// (e) A graph against itself: every person has a name, so every line
		// holds at least that fact.
		for _, l := range recordlinks.Lines(p, ga, ga, metas, cfg) {
			if l.Facts < 1 {
				t.Fatalf("round %d: Lines(a, a) gives entity %x %d facts", round, l.Entity, l.Facts)
			}
		}
	}
}

// TestPairsInvariants holds Pairs to the one-pair count: each pair of a read
// equals the lines over the same two graphs, and a pair is present exactly when
// those lines are not empty.
func TestPairsInvariants(t *testing.T) {
	r := rand.New(rand.NewSource(2))
	base := matchregistry.Default().RecordLinksConfig()
	metas := subjectcomparetest.Metas()
	people := [][]byte{ent(1), ent(2), ent(3), ent(4), ent(5), ent(6)}
	events := [][]byte{ent(20), ent(21), ent(22)}
	p := people[0]
	prefixes := []byte{0xA0, 0xB0, 0xC0, 0xD0}

	for round := range 300 {
		var graphs []recordlinks.Graph
		for i, prefix := range prefixes {
			ms, es := materialize(genGraph(r, p, people, events), prefix, 0x10*byte(i+1))
			graphs = append(graphs, build(ms, es...))
		}
		// Each round sets the read's cap anywhere from nothing to just over
		// its work, and one round in three the pair's cap below forty, so the
		// caps are held over generated input.
		cfg := base
		if r.Intn(3) == 0 {
			cfg.RecordLinks.MaxPairWork = r.Intn(40)
		}
		work := recordlinks.ReadWork(p, graphs, cfg)
		cfg.RecordLinks.MaxReadWork = r.Intn(work + 3)

		got := recordlinks.Pairs(p, graphs, metas, cfg)
		if back, want := recordlinks.Pairs(p, reversed(graphs), metas, cfg), mirror(got, len(graphs)); !reflect.DeepEqual(back, want) && (len(back) != 0 || len(want) != 0) {
			t.Fatalf("round %d: the reversed read = %+v, want the mirrored read %+v", round, back, want)
		}
		byPair := map[[2]int][]recordlinks.Line{}
		for _, pair := range got {
			if pair.A >= pair.B {
				t.Fatalf("round %d: pair %d-%d is not ascending", round, pair.A, pair.B)
			}
			byPair[[2]int{pair.A, pair.B}] = pair.Lines
		}
		var lines int
		for a := range graphs {
			for b := a + 1; b < len(graphs); b++ {
				lines += len(recordlinks.Lines(p, graphs[a], graphs[b], metas, cfg))
			}
		}
		if lines > cfg.RecordLinks.MaxReadWork {
			if got != nil {
				t.Fatalf("round %d: %d lines pass the cap %d but Pairs = %+v", round, lines, cfg.RecordLinks.MaxReadWork, got)
			}
			continue
		}
		for a := range graphs {
			for b := a + 1; b < len(graphs); b++ {
				want := recordlinks.Lines(p, graphs[a], graphs[b], metas, cfg)
				have, present := byPair[[2]int{a, b}]
				if present != (len(want) != 0) {
					t.Fatalf("round %d pair %d-%d: Pairs = %s (present %v), Lines = %s", round, a, b, show(have), present, show(want))
				}
				if work <= cfg.RecordLinks.MaxReadWork {
					if present && !reflect.DeepEqual(have, want) {
						t.Fatalf("round %d pair %d-%d: Pairs = %s, Lines = %s", round, a, b, show(have), show(want))
					}
					continue
				}
				for i, l := range have {
					if !bytes.Equal(l.Entity, want[i].Entity) || l.Facts != 0 || l.Strength != recordlinks.StrengthWeak {
						t.Fatalf("round %d pair %d-%d: past the read's cap line %d = %+v; want %x at 0 facts, weak", round, a, b, i, l, want[i].Entity)
					}
				}
			}
		}
	}
}
