package canonlinks

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/graphcache/graphcachetest"
	"github.com/mendahu/provenencia/core/kinship"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/segmentcompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// shipped is the registry's bound.
var shipped = matchregistry.Default().Kinship

// world is one catalog with its canon builder and a reader over it.
type world struct {
	t     *testing.T
	c     *database.Catalog
	k     *graphcachetest.Canon
	g     *graphcache.Graph
	r     *Reader
	names map[string]string
	all   [][]byte
}

func newWorld(t *testing.T) *world {
	t.Helper()
	c := graphcachetest.OpenCatalog(t)
	r, err := NewReader(c.Graph(), shipped)
	if err != nil {
		t.Fatal(err)
	}
	return &world{t: t, c: c, k: graphcachetest.NewCanon(t, c), g: c.Graph(), r: r, names: map[string]string{}}
}

func (w *world) name(id []byte) string {
	if n, ok := w.names[string(id)]; ok {
		return n
	}
	return fmt.Sprintf("%x", id)
}

func (w *world) entity(kind, name string) []byte {
	id := w.k.Entity(kind)
	w.names[string(id)] = name
	w.all = append(w.all, id)
	return id
}

func (w *world) person(name string) []byte { return w.entity("person", name) }

func (w *world) event(name string, eventType vocabkey.EventType) []byte {
	id := w.entity("event", name)
	w.k.Term(id, vocabkey.PropertyEventType, string(eventType))
	return id
}

func (w *world) join(name string, person, event []byte, role string) []byte {
	id := w.k.Link("participation", map[string][]byte{"person": person, "event": event}, "role", role)
	w.names[string(id)] = name
	w.all = append(w.all, id)
	return id
}

// relationship files a relationship of the term from a to b.
func (w *world) relationship(name string, a, b []byte, term string) []byte {
	id := w.k.Link("relationship", map[string][]byte{"person": a, "related_to": b}, "relationship_type", term)
	w.names[string(id)] = name
	w.all = append(w.all, id)
	return id
}

// parent files parent as the parent of child.
func (w *world) parent(name string, parent, child []byte) []byte {
	return w.relationship(name, parent, child, "parent")
}

// shape is a link as a test reads it: ids by name, associations as a set.
type shape struct {
	Pattern kinship.Pattern
	Event   string
	Through string
	A, B    string
	Assoc   []string
}

func (w *world) shapes(links []kinship.Implied) []shape {
	var out []shape
	for _, l := range links {
		s := shape{Pattern: l.Pattern, A: w.name(l.A), B: w.name(l.B)}
		if l.Event != nil {
			s.Event = w.name(l.Event)
		}
		if l.Through != nil {
			s.Through = w.name(l.Through)
		}
		for _, a := range l.Associations {
			s.Assoc = append(s.Assoc, w.name(a))
		}
		sort.Strings(s.Assoc)
		out = append(out, s)
	}
	return out
}

func (w *world) for_(id []byte) []kinship.Implied {
	w.t.Helper()
	links, err := w.r.For(id)
	if err != nil {
		w.t.Fatal(err)
	}
	return links
}

func (w *world) expect(id []byte, want ...shape) {
	w.t.Helper()
	got := w.shapes(w.for_(id))
	for i := range want {
		sort.Strings(want[i].Assoc)
	}
	if len(got) == 0 && len(want) == 0 {
		return
	}
	if !reflect.DeepEqual(got, want) {
		w.t.Fatalf("For(%s):\n got %+v\nwant %+v", w.name(id), got, want)
	}
}

// keys are the keys a link of one term has, as the patterns write them.
func (w *world) keys(bridge string, term vocabkey.BridgeTerm) (vocabkey.BridgeKey, vocabkey.BridgeKey) {
	w.t.Helper()
	b, ok := connectrules.LookupBridge(bridge)
	if !ok {
		w.t.Fatalf("bridge %s", bridge)
	}
	terms, err := LoadTerms(w.g)
	if err != nil {
		w.t.Fatal(err)
	}
	return segmentcompare.Keys(b, term, "", terms)
}

// marriage records a marriage of the people, each its subject.
func (w *world) marriage(name string, people ...[]byte) []byte {
	m := w.event(name, vocabkey.EventMarriage)
	for _, p := range people {
		w.join(name+"/"+w.name(p), p, m, "subject")
	}
	return m
}

func TestFor(t *testing.T) {
	t.Run("AC-1.1", func(t *testing.T) {
		w := newWorld(t)
		ann, bob := w.person("Ann"), w.person("Bob")
		w.marriage("M1", ann, bob)
		w.expect(ann, shape{Pattern: kinship.PatternMarriageSpouses, Event: "M1", A: "Ann", B: "Bob", Assoc: []string{"M1/Ann", "M1/Bob"}})
		links := w.for_(ann)
		first, second := w.keys(vocabkey.BridgeRelationship, vocabkey.TermSpouse)
		if links[0].KeyA != first || links[0].KeyB != second {
			t.Fatalf("keys %v %v, want %v %v", links[0].KeyA, links[0].KeyB, first, second)
		}
		w.expect(bob, shape{Pattern: kinship.PatternMarriageSpouses, Event: "M1", A: "Ann", B: "Bob", Assoc: []string{"M1/Ann", "M1/Bob"}})
	})
	t.Run("AC-1.2", func(t *testing.T) {
		w := newWorld(t)
		ann, bea := w.person("Ann"), w.person("Bea")
		b1 := w.event("B1", vocabkey.EventBirth)
		w.join("B1/Ann", ann, b1, "father")
		w.join("B1/Bea", bea, b1, "subject")
		w.expect(bea, shape{Pattern: kinship.PatternBirthParents, Event: "B1", A: "Ann", B: "Bea", Assoc: []string{"B1/Ann", "B1/Bea"}})
		first, second := w.keys(vocabkey.BridgeRelationship, vocabkey.TermParent)
		if l := w.for_(bea)[0]; l.KeyA != first || l.KeyB != second {
			t.Fatalf("keys %v %v, want %v %v", l.KeyA, l.KeyB, first, second)
		}
	})
	t.Run("AC-1.3", func(t *testing.T) {
		w := newWorld(t)
		ann, bea, cal := w.person("Ann"), w.person("Bea"), w.person("Cal")
		w.parent("AB", ann, bea)
		w.parent("BC", bea, cal)
		want := shape{Pattern: kinship.PatternParentOfParent, Through: "Bea", A: "Ann", B: "Cal", Assoc: []string{"AB", "BC"}}
		w.expect(ann, want)
		w.expect(cal, want)
		first, second := w.keys(vocabkey.BridgeRelationship, vocabkey.TermGrandparent)
		if l := w.for_(ann)[0]; l.KeyA != first || l.KeyB != second {
			t.Fatalf("keys %v %v, want %v %v", l.KeyA, l.KeyB, first, second)
		}
	})
	t.Run("AC-1.4", func(t *testing.T) {
		w := newWorld(t)
		ann, bea, cal := w.person("Ann"), w.person("Bea"), w.person("Cal")
		b1, b2 := w.event("B1", vocabkey.EventBirth), w.event("B2", vocabkey.EventBirth)
		w.join("B1/Ann", ann, b1, "father")
		w.join("B1/Bea", bea, b1, "subject")
		w.join("B2/Bea", bea, b2, "mother")
		w.join("B2/Cal", cal, b2, "subject")
		w.expect(cal,
			shape{Pattern: kinship.PatternBirthParents, Event: "B2", A: "Bea", B: "Cal", Assoc: []string{"B2/Bea", "B2/Cal"}},
			shape{Pattern: kinship.PatternParentOfParent, Through: "Bea", A: "Ann", B: "Cal", Assoc: []string{"B1/Ann", "B1/Bea", "B2/Bea", "B2/Cal"}},
		)
		w.expect(ann, shape{Pattern: kinship.PatternBirthParents, Event: "B1", A: "Ann", B: "Bea", Assoc: []string{"B1/Ann", "B1/Bea"}},
			shape{Pattern: kinship.PatternParentOfParent, Through: "Bea", A: "Ann", B: "Cal", Assoc: []string{"B1/Ann", "B1/Bea", "B2/Bea", "B2/Cal"}})
	})
	t.Run("AC-1.5", func(t *testing.T) {
		w := newWorld(t)
		var eleven, ten [][]byte
		for i := 0; i < 11; i++ {
			eleven = append(eleven, w.person(fmt.Sprintf("E%d", i)))
		}
		for i := 0; i < 10; i++ {
			ten = append(ten, w.person(fmt.Sprintf("T%d", i)))
		}
		w.marriage("M2", eleven...)
		w.marriage("M3", ten...)
		w.expect(eleven[0])
		if got := len(w.for_(ten[3])); got != 9 {
			t.Fatalf("a subject of a marriage of 10 has %d links, want 9", got)
		}
	})
	t.Run("AC-1.6", func(t *testing.T) {
		w := newWorld(t)
		ann, bob, cy, di := w.person("Ann"), w.person("Bob"), w.person("Cy"), w.person("Di")
		m1 := w.event("M1", vocabkey.EventMarriage)
		w.join("M1/Ann", ann, m1, "subject")
		w.join("M1/Bob", bob, m1, "witness")
		// A catalog's own event type implies nothing, whatever its roles.
		db, err := w.c.DB()
		if err != nil {
			t.Fatal(err)
		}
		var propID []byte
		if err := db.QueryRow(`SELECT id FROM properties WHERE key = 'event_type' AND origin = 'provenencia'`).Scan(&propID); err != nil {
			t.Fatal(err)
		}
		termID := []byte{7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7}
		if _, err := db.Exec(`INSERT INTO property_terms (id, property_id, key, origin, label) VALUES (?, ?, 'wedding', 'user', 'Wedding')`, termID, propID); err != nil {
			t.Fatal(err)
		}
		wd := w.entity("event", "W1")
		w.k.Term(wd, vocabkey.PropertyEventType, "wedding")
		w.join("W1/Cy", cy, wd, "subject")
		w.join("W1/Di", di, wd, "subject")
		for _, id := range [][]byte{ann, bob, cy, di} {
			w.expect(id)
		}
	})
	t.Run("AC-1.7", func(t *testing.T) {
		w := newWorld(t)
		w.expect(w.person("Ann"))
		ann := w.person("Zed")
		if _, err := w.r.For(ann); err != nil {
			t.Fatal(err)
		}
	})
}

// relShape is a relationship link as a test reads it: ends by name, keys,
// associations as sorted names, shapes as (pattern, pivot name).
type relShape struct {
	A, B        string
	KeyA, KeyB  vocabkey.BridgeKey
	Assoc       []string
	Shapes      []string
	OnlyImplied bool
}

func (w *world) relShapes(links []kinship.Link) []relShape {
	var out []relShape
	for _, l := range links {
		s := relShape{A: w.name(l.A), B: w.name(l.B), KeyA: l.KeyA, KeyB: l.KeyB, OnlyImplied: l.OnlyImplied}
		for _, a := range l.Associations {
			s.Assoc = append(s.Assoc, w.name(a))
		}
		sort.Strings(s.Assoc)
		for _, sh := range l.Shapes {
			s.Shapes = append(s.Shapes, fmt.Sprintf("%s/%s", sh.Pattern, w.name(sh.Pivot)))
		}
		out = append(out, s)
	}
	return out
}

func (w *world) relationships(id []byte) []kinship.Link {
	w.t.Helper()
	links, err := w.r.Relationships(id)
	if err != nil {
		w.t.Fatal(err)
	}
	return links
}

// lesserFirst turns each link so the lesser end's name is A, keys with it.
func lesserFirst(links []relShape) []relShape {
	out := slices.Clone(links)
	for i, l := range out {
		if l.A > l.B {
			out[i].A, out[i].B, out[i].KeyA, out[i].KeyB = l.B, l.A, l.KeyB, l.KeyA
		}
	}
	return out
}

func TestRelationships(t *testing.T) {
	check := func(t *testing.T, got, want []relShape) {
		t.Helper()
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got  %+v\nwant %+v", got, want)
		}
	}
	t.Run("AC-1.1", func(t *testing.T) {
		w := newWorld(t)
		ann, bob := w.person("Ann"), w.person("Bob")
		w.relationship("AB", ann, bob, "spouse")
		w.marriage("M1", ann, bob)
		ka, kb := w.keys(vocabkey.BridgeRelationship, vocabkey.TermSpouse)
		want := []relShape{{A: "Ann", B: "Bob", KeyA: ka, KeyB: kb, Assoc: []string{"AB", "M1/Ann", "M1/Bob"},
			Shapes: []string{string(kinship.PatternMarriageSpouses) + "/M1"}}}
		check(t, lesserFirst(w.relShapes(w.relationships(ann))), want)
		check(t, lesserFirst(w.relShapes(w.relationships(bob))), want)
	})
	t.Run("AC-1.2", func(t *testing.T) {
		w := newWorld(t)
		ann, bea := w.person("Ann"), w.person("Bea")
		w.parent("AB", ann, bea)
		b1 := w.event("B1", vocabkey.EventBirth)
		w.join("B1/Ann", ann, b1, "father")
		w.join("B1/Bea", bea, b1, "subject")
		ka, kb := w.keys(vocabkey.BridgeRelationship, vocabkey.TermParent)
		check(t, w.relShapes(w.relationships(bea)), []relShape{{A: "Ann", B: "Bea", KeyA: ka, KeyB: kb,
			Assoc: []string{"AB", "B1/Ann", "B1/Bea"}, Shapes: []string{string(kinship.PatternBirthParents) + "/B1"}}})
	})
	t.Run("AC-1.3", func(t *testing.T) {
		w := newWorld(t)
		ann, bea, cal := w.person("Ann"), w.person("Bea"), w.person("Cal")
		w.parent("AB", ann, bea)
		w.parent("BC", bea, cal)
		pa, pb := w.keys(vocabkey.BridgeRelationship, vocabkey.TermParent)
		ga, gb := w.keys(vocabkey.BridgeRelationship, vocabkey.TermGrandparent)
		check(t, w.relShapes(w.relationships(ann)), []relShape{
			{A: "Ann", B: "Bea", KeyA: pa, KeyB: pb, Assoc: []string{"AB"}},
			{A: "Ann", B: "Cal", KeyA: ga, KeyB: gb, Assoc: []string{"AB", "BC"},
				Shapes: []string{string(kinship.PatternParentOfParent) + "/Bea"}, OnlyImplied: true},
		})
	})
	t.Run("AC-1.4", func(t *testing.T) {
		w := newWorld(t)
		ann, bob := w.person("Ann"), w.person("Bob")
		w.marriage("M1", ann, bob)
		links := w.relationships(ann)
		if len(links) != 1 || !links[0].OnlyImplied {
			t.Fatalf("links %+v, want one implied spouse link", w.relShapes(links))
		}
		for _, l := range links {
			if l.KeyA.Type != vocabkey.BridgeRelationship {
				t.Fatalf("link %+v is not a relationship", l)
			}
		}
	})
	t.Run("AC-1.5", func(t *testing.T) {
		w := newWorld(t)
		if got := w.relationships(w.person("Ann")); len(got) != 0 {
			t.Fatalf("got %+v, want none", got)
		}
		if got := w.relationships(make([]byte, 16)); len(got) != 0 {
			t.Fatalf("got %+v for an unknown id, want none", got)
		}
	})
	t.Run("AC-1.6", func(t *testing.T) {
		w := newWorld(t)
		ann, bob := w.person("Ann"), w.person("Bob")
		w.relationship("AB", ann, bob, "spouse")
		w.marriage("M1", ann, bob)
		cloneLinks := func(in []kinship.Link) []kinship.Link {
			var out []kinship.Link
			for _, l := range in {
				out = append(out, kinship.Link{Edge: l.Edge.Clone(), Recording: l.Recording.Clone()})
			}
			return out
		}
		cloneImplied := func(in []kinship.Implied) []kinship.Implied {
			var out []kinship.Implied
			for _, l := range in {
				c := l
				c.Edge = l.Edge.Clone()
				c.Event, c.Through = bytes.Clone(l.Event), bytes.Clone(l.Through)
				c.From = nil
				for _, f := range l.From {
					c.From = append(c.From, f.Clone())
				}
				out = append(out, c)
			}
			return out
		}
		wantRel, wantFor := cloneLinks(w.relationships(ann)), cloneImplied(w.for_(ann))
		got := w.relationships(ann)
		got[0].A[0] ^= 0xff
		for _, a := range got[0].Associations {
			a[0] ^= 0xff
		}
		if !reflect.DeepEqual(w.relationships(ann), wantRel) {
			t.Fatalf("Relationships changed after the caller edited a result")
		}
		if !reflect.DeepEqual(w.for_(ann), wantFor) {
			t.Fatalf("For changed after the caller edited a result")
		}
	})
}

// hasMemo reports whether a result is stored for an id.
func (w *world) hasMemo(slot graphcache.MemoSlot, id []byte) bool {
	_, ok := w.g.Memo(slot, id)
	return ok
}

func TestForFollowsCommits(t *testing.T) {
	t.Run("AC-2.1", func(t *testing.T) {
		w := newWorld(t)
		ann, bob := w.person("Ann"), w.person("Bob")
		m1 := w.event("M1", vocabkey.EventMarriage)
		w.join("M1/Ann", ann, m1, "subject")
		bobIn := w.join("M1/Bob", bob, m1, "subject")
		if len(w.for_(ann)) != 1 {
			t.Fatal("setup: want one link")
		}
		w.k.Remove(bobIn)
		w.k.Commit(bobIn)
		w.expect(ann)
	})
	t.Run("AC-2.2", func(t *testing.T) {
		w := newWorld(t)
		ann, bob := w.person("Ann"), w.person("Bob")
		m1 := w.marriage("M1", ann, bob)
		if len(w.for_(ann)) != 1 {
			t.Fatal("setup: want one link")
		}
		w.k.Retype(m1, vocabkey.PropertyEventType, "residence")
		w.k.Commit(m1)
		w.expect(ann)
	})
	t.Run("AC-2.3", func(t *testing.T) {
		w := newWorld(t)
		ann, bea, cal := w.person("Ann"), w.person("Bea"), w.person("Cal")
		w.parent("AB", ann, bea)
		w.expect(ann)
		bc := w.parent("BC", bea, cal)
		w.k.Commit(bc)
		w.expect(ann, shape{Pattern: kinship.PatternParentOfParent, Through: "Bea", A: "Ann", B: "Cal", Assoc: []string{"AB", "BC"}})
	})
	t.Run("AC-2.4", func(t *testing.T) {
		w := newWorld(t)
		ann, bea, cal := w.person("Ann"), w.person("Bea"), w.person("Cal")
		b1, b2 := w.event("B1", vocabkey.EventBirth), w.event("B2", vocabkey.EventBirth)
		father := w.join("B1/Ann", ann, b1, "father")
		w.join("B1/Bea", bea, b1, "subject")
		w.join("B2/Bea", bea, b2, "mother")
		w.join("B2/Cal", cal, b2, "subject")
		if len(w.for_(cal)) != 2 {
			t.Fatal("setup: want a parent and a grandparent")
		}
		w.k.Remove(father)
		w.k.Commit(father)
		w.expect(cal, shape{Pattern: kinship.PatternBirthParents, Event: "B2", A: "Bea", B: "Cal", Assoc: []string{"B2/Bea", "B2/Cal"}})
	})
	t.Run("an event deleted alone", func(t *testing.T) {
		// It commits a handle set that leaves the cache out of step with the catalog on purpose, to exercise the neighbor drop for a writer that omits an event's associations; verify mode would flag that inconsistency, not the drop.
		if os.Getenv("PROVENENCIA_GRAPH_VERIFY") == "1" {
			t.Skip("commits a deliberately inconsistent handle set")
		}
		w := newWorld(t)
		ann, bea, cal := w.person("Ann"), w.person("Bea"), w.person("Cal")
		b1, b2 := w.event("B1", vocabkey.EventBirth), w.event("B2", vocabkey.EventBirth)
		father := w.join("B1/Ann", ann, b1, "father")
		child := w.join("B1/Bea", bea, b1, "subject")
		w.join("B2/Bea", bea, b2, "mother")
		w.join("B2/Cal", cal, b2, "subject")
		if len(w.for_(cal)) != 2 {
			t.Fatal("setup: want a parent and a grandparent")
		}
		// The commit names the event alone, not its participations.
		w.k.Remove(father)
		w.k.Remove(child)
		w.k.Remove(b1)
		w.k.Commit(b1)
		w.expect(cal, shape{Pattern: kinship.PatternBirthParents, Event: "B2", A: "Bea", B: "Cal", Assoc: []string{"B2/Bea", "B2/Cal"}})
	})
	t.Run("AC-2.5", func(t *testing.T) {
		w := newWorld(t)
		ann, bob := w.person("Ann"), w.person("Bob")
		m1 := w.marriage("M1", ann, bob)
		w.for_(ann)
		if !w.hasMemo(kinshipSlot, m1) {
			t.Fatal("setup: the event result was not stored")
		}
		w.k.CommitVocabulary()
		if w.hasMemo(kinshipSlot, m1) {
			t.Fatal("a vocabulary commit kept a result")
		}
		if len(w.for_(ann)) != 1 || !w.hasMemo(kinshipSlot, m1) {
			t.Fatal("the next read did not derive again")
		}
	})
	t.Run("AC-2.7", func(t *testing.T) {
		w := newWorld(t)
		ann, bob := w.person("Ann"), w.person("Bob")
		m1 := w.marriage("M1", ann, bob)
		w.for_(ann)
		if _, ok := w.g.Cached(bob); ok {
			t.Fatal("setup: Bob was loaded")
		}
		if _, err := w.g.Node(bob); err != nil {
			t.Fatal(err)
		}
		if !w.hasMemo(kinshipSlot, m1) {
			t.Fatal("a first load dropped the result")
		}
	})
}

// edgesOf is every link of the canon as an edge, once each.
func (w *world) edgesOf() []kinship.Edge {
	w.t.Helper()
	terms, err := LoadTerms(w.g)
	if err != nil {
		w.t.Fatal(err)
	}
	v, err := w.g.Vocabulary()
	if err != nil {
		w.t.Fatal(err)
	}
	prop := v.PropertyID(vocabkey.PropertyEventType, vocabkey.OriginProvenencia)
	seen := map[string]bool{}
	var edges []kinship.Edge
	for _, id := range w.all {
		n, err := w.g.Node(id)
		if err != nil || n == nil {
			w.t.Fatalf("node %s: %v", w.name(id), err)
		}
		for _, l := range n.Links {
			if seen[string(l.Association)] {
				continue
			}
			seen[string(l.Association)] = true
			neighbor, err := w.g.Node(l.Neighbor)
			if err != nil {
				w.t.Fatal(err)
			}
			from, to, kf, kt, ok := EdgeOf(id, n, l, neighbor, prop, terms)
			if !ok {
				continue
			}
			edges = append(edges, kinship.Edge{A: from, B: to, KeyA: kf, KeyB: kt, Associations: [][]byte{l.Association}})
		}
	}
	return edges
}

func TestForMatchesDerive(t *testing.T) {
	w := newWorld(t)
	var p [6][]byte
	for i := range p {
		p[i] = w.person(fmt.Sprintf("P%d", i+1))
	}
	w.marriage("M1", p[0], p[1])
	w.marriage("M2", p[2], p[3])
	birth := func(name string, father, mother, child []byte) {
		b := w.event(name, vocabkey.EventBirth)
		w.join(name+"/f", father, b, "father")
		w.join(name+"/m", mother, b, "mother")
		w.join(name+"/c", child, b, "subject")
	}
	birth("B1", p[0], p[1], p[2])
	birth("B2", p[2], p[3], p[4])
	birth("B3", p[0], p[1], p[3])
	w.parent("R1", p[5], p[0])
	w.parent("R2", p[5], p[2])

	terms, err := LoadTerms(w.g)
	if err != nil {
		t.Fatal(err)
	}
	all := kinship.Derive(w.edgesOf(), terms, shipped)
	sortKey := func(l kinship.Implied) string {
		return strings.Join([]string{string(l.Pattern), w.name(l.Event), w.name(l.Through), w.name(l.A), w.name(l.B)}, "|")
	}
	normal := func(links []kinship.Implied) []shape {
		links = slices.Clone(links)
		sort.Slice(links, func(i, j int) bool { return sortKey(links[i]) < sortKey(links[j]) })
		return w.shapes(links)
	}
	checked := 0
	for _, id := range p {
		var want []kinship.Implied
		for _, l := range all {
			if bytes.Equal(l.A, id) || bytes.Equal(l.B, id) {
				want = append(want, l)
			}
		}
		got, wantShapes := normal(w.for_(id)), normal(want)
		if !reflect.DeepEqual(got, wantShapes) {
			t.Fatalf("For(%s):\n got %+v\nwant %+v", w.name(id), got, wantShapes)
		}
		checked += len(got)
	}
	if checked == 0 {
		t.Fatal("the fixture implies nothing")
	}
}

func TestVerifyCatchesAStaleResult(t *testing.T) {
	w := newWorld(t)
	ann, bob := w.person("Ann"), w.person("Bob")
	m1 := w.marriage("M1", ann, bob)
	w.for_(ann)
	if err := w.g.Verify(); err != nil {
		t.Fatalf("a fresh result failed Verify: %v", err)
	}
	w.g.SetMemo(kinshipSlot, m1, pivotResult{kind: pivotEvent, rules: shipped, links: []kinship.Implied{{
		Edge: kinship.Edge{A: bob, B: ann}, Pattern: kinship.PatternMarriageSpouses, Event: m1,
	}}})
	if err := w.g.Verify(); err == nil {
		t.Fatal("Verify accepted a stale result")
	}
}
