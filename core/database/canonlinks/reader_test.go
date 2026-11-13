package canonlinks

import (
	"bytes"
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/graphcache/graphcachetest"
	"github.com/mendahu/provenencia/core/database/namealiases"
	"github.com/mendahu/provenencia/core/database/nameformats"
	"github.com/mendahu/provenencia/core/kinship"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/segmentcompare"
	"github.com/mendahu/provenencia/core/valuecompare"
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
	r, err := NewReader(c.Graph())
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

func (w *world) person(name string) []byte {
	id := w.k.Person()
	w.note(id, name)
	return id
}

func (w *world) event(name string, eventType vocabkey.EventType) []byte {
	id := w.k.Event(eventType)
	w.note(id, name)
	return id
}

func (w *world) join(name string, person, event []byte, role string) []byte {
	id := w.k.Join(person, event, role)
	w.note(id, name)
	return id
}

// relationship files a relationship of the term from a to b.
func (w *world) relationship(name string, a, b []byte, term string) []byte {
	id := w.k.Relationship(a, b, term)
	w.note(id, name)
	return id
}

// parent files parent as the parent of child.
func (w *world) parent(name string, parent, child []byte) []byte {
	id := w.k.Parent(parent, child)
	w.note(id, name)
	return id
}

// name_ records a test name for id.
func (w *world) note(id []byte, name string) {
	w.names[string(id)] = name
	w.all = append(w.all, id)
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
	m, joins := w.k.Marriage(people...)
	w.note(m, name)
	for i, p := range people {
		w.note(joins[i], name+"/"+w.name(p))
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

// cloneLinks copies links so a later edit of the originals cannot change them.
func cloneLinks(in []kinship.Link) []kinship.Link {
	var out []kinship.Link
	for _, l := range in {
		out = append(out, kinship.Link{Edge: l.Edge.Clone(), Recording: l.Recording.Clone()})
	}
	return out
}

// cloneImplied copies implied links, bytes included.
func cloneImplied(in []kinship.Implied) []kinship.Implied {
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
	links, err := w.r.Relationships([][]byte{id}, Options{})
	if err != nil {
		w.t.Fatal(err)
	}
	return links[string(id)]
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

// hasPivot reports whether a kinship result is stored for an id.
func (w *world) hasPivot(id []byte) bool {
	_, ok := pivots.Peek(w.g, id)
	return ok
}

// A new name format or alias index empties what read it, and no kinship
// result: a result reads nodes and the vocabulary only.
func TestPivotsSurviveNameFormatAndAliasChanges(t *testing.T) {
	w := newWorld(t)
	ann, bob := w.person("Ann"), w.person("Bob")
	m1 := w.marriage("M1", ann, bob)
	w.for_(ann)
	if !w.hasPivot(m1) {
		t.Fatal("setup: no stored result")
	}
	nameformats.Default.Set(w.g, nameformats.Loaded{})
	namealiases.Index.Set(w.g, valuecompare.AliasIndex{})
	if !w.hasPivot(m1) {
		t.Fatal("a name format or alias index change dropped a kinship result")
	}
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
		if graphcache.VerifyMode() {
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
		if !w.hasPivot(m1) {
			t.Fatal("setup: the event result was not stored")
		}
		w.k.CommitVocabulary()
		if w.hasPivot(m1) {
			t.Fatal("a vocabulary commit kept a result")
		}
		if len(w.for_(ann)) != 1 || !w.hasPivot(m1) {
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
		if !w.hasPivot(m1) {
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
	pivots.Set(w.g, m1, pivotResult{kind: pivotEvent, rules: shipped, links: []kinship.Implied{{
		Edge: kinship.Edge{A: bob, B: ann}, Pattern: kinship.PatternMarriageSpouses, Event: m1,
	}}})
	if err := w.g.Verify(); err == nil {
		t.Fatal("Verify accepted a stale result")
	}
}

// sibling files a sibling relationship from a to b.
func (w *world) sibling(name string, a, b []byte) []byte {
	return w.relationship(name, a, b, "sibling")
}

// siblings are the sibling-key links of id's relationships, siblings included.
func (w *world) siblings(id []byte) []kinship.Link {
	w.t.Helper()
	ka, kb := w.keys(vocabkey.BridgeRelationship, vocabkey.TermSibling)
	links, err := w.r.Relationships([][]byte{id}, Options{Siblings: true})
	if err != nil {
		w.t.Fatal(err)
	}
	var out []kinship.Link
	for _, l := range links[string(id)] {
		if l.KeyA == ka && l.KeyB == kb {
			out = append(out, l)
		}
	}
	return out
}

// sharedParent is the shape string of a sibling implied through a person.
func sharedParent(name string) string { return string(kinship.PatternSharedParent) + "/" + name }

func TestRelationshipsOptionsSiblings(t *testing.T) {
	// want builds a sibling link as relShapes reads it.
	want := func(w *world, a, b string, assoc []string, shapes ...string) relShape {
		ka, kb := w.keys(vocabkey.BridgeRelationship, vocabkey.TermSibling)
		return relShape{A: a, B: b, KeyA: ka, KeyB: kb, Assoc: assoc, Shapes: shapes, OnlyImplied: len(shapes) > 0}
	}
	read := func(w *world, id []byte) []relShape { return lesserFirst(w.relShapes(w.siblings(id))) }
	check := func(t *testing.T, got []relShape, want ...relShape) {
		t.Helper()
		if len(got) == 0 && len(want) == 0 {
			return
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got  %+v\nwant %+v", got, want)
		}
	}
	t.Run("AC-1.1", func(t *testing.T) {
		w := newWorld(t)
		bea, cal := w.person("Bea"), w.person("Cal")
		w.sibling("BC", bea, cal)
		check(t, read(w, bea), want(w, "Bea", "Cal", []string{"BC"}))
		check(t, read(w, cal), want(w, "Bea", "Cal", []string{"BC"}))
	})
	t.Run("AC-1.2", func(t *testing.T) {
		w := newWorld(t)
		ann, bea, cal := w.person("Ann"), w.person("Bea"), w.person("Cal")
		w.parent("AB", ann, bea)
		w.parent("AC", ann, cal)
		check(t, read(w, bea), want(w, "Bea", "Cal", []string{"AB", "AC"}, sharedParent("Ann")))
		check(t, read(w, cal), want(w, "Bea", "Cal", []string{"AB", "AC"}, sharedParent("Ann")))
	})
	t.Run("AC-1.3", func(t *testing.T) {
		w := newWorld(t)
		ann, bea, cal := w.person("Ann"), w.person("Bea"), w.person("Cal")
		b1, b2 := w.event("B1", vocabkey.EventBirth), w.event("B2", vocabkey.EventBirth)
		w.join("B1/Ann", ann, b1, "father")
		w.join("B1/Bea", bea, b1, "subject")
		w.join("B2/Ann", ann, b2, "father")
		w.join("B2/Cal", cal, b2, "subject")
		check(t, read(w, bea), want(w, "Bea", "Cal", []string{"B1/Ann", "B1/Bea", "B2/Ann", "B2/Cal"}, sharedParent("Ann")))
	})
	t.Run("AC-1.4", func(t *testing.T) {
		w := newWorld(t)
		ann, bea, cal := w.person("Ann"), w.person("Bea"), w.person("Cal")
		w.parent("AB", ann, bea)
		w.parent("AC", ann, cal)
		w.sibling("BC", bea, cal)
		got := w.relShapes(w.siblings(bea))
		wantLink := want(w, "Bea", "Cal", []string{"AB", "AC", "BC"}, sharedParent("Ann"))
		wantLink.OnlyImplied = false
		check(t, got, wantLink)
	})
	t.Run("AC-1.5", func(t *testing.T) {
		w := newWorld(t)
		ann, bea, cal, dan := w.person("Ann"), w.person("Bea"), w.person("Cal"), w.person("Dan")
		w.parent("AB", ann, bea)
		w.parent("AC", ann, cal)
		w.parent("DB", dan, bea)
		w.parent("DC", dan, cal)
		check(t, read(w, bea), want(w, "Bea", "Cal", []string{"AB", "AC", "DB", "DC"}, sharedParent("Ann"), sharedParent("Dan")))
	})
	t.Run("AC-1.6", func(t *testing.T) {
		w := newWorld(t)
		ann, bea, cal, dan := w.person("Ann"), w.person("Bea"), w.person("Cal"), w.person("Dan")
		w.parent("AB", ann, bea)
		w.parent("AC", ann, cal)
		w.parent("DC", dan, cal)
		for _, id := range [][]byte{bea, cal} {
			check(t, read(w, id), want(w, "Bea", "Cal", []string{"AB", "AC"}, sharedParent("Ann")))
		}
	})
	t.Run("AC-1.7", func(t *testing.T) {
		w := newWorld(t)
		ann, bea, cal, dan, eve := w.person("Ann"), w.person("Bea"), w.person("Cal"), w.person("Dan"), w.person("Eve")
		w.parent("AB", ann, bea)
		w.relationship("AD", ann, dan, "spouse")
		w.parent("DE", dan, eve)
		w.parent("BC2", bea, cal)
		for _, id := range [][]byte{bea, cal, eve} {
			check(t, read(w, id))
		}
	})
	t.Run("AC-1.8", func(t *testing.T) {
		w := newWorld(t)
		ann, bea := w.person("Ann"), w.person("Bea")
		w.parent("AB", ann, bea)
		var wantLinks []relShape
		var kids [][]byte
		for i := 1; i <= 14; i++ {
			name := fmt.Sprintf("C%d", i)
			kid := w.person(name)
			kids = append(kids, kid)
			w.parent(fmt.Sprintf("AC%d", i), ann, kid)
			wantLinks = append(wantLinks, want(w, "Bea", name, []string{"AB", fmt.Sprintf("AC%d", i)}, sharedParent("Ann")))
		}
		got := w.siblings(bea)
		// The order is that of Ann's links, which the graph sorts by the link's
		// ref as text: not always the order the children were filed in.
		n, err := w.g.Node(ann)
		if err != nil {
			t.Fatal(err)
		}
		var order [][]byte
		for _, l := range n.Links {
			for _, k := range kids {
				if bytes.Equal(l.Neighbor, k) {
					order = append(order, k)
				}
			}
		}
		if len(got) != len(kids) || len(order) != len(kids) {
			t.Fatalf("got %d links, %d children among Ann's links, want %d", len(got), len(order), len(kids))
		}
		var wantInOrder []relShape
		for i, l := range got {
			other := l.A
			if bytes.Equal(l.A, bea) {
				other = l.B
			}
			if !bytes.Equal(other, order[i]) {
				t.Fatalf("link %d is to %s, want %s", i, w.name(other), w.name(order[i]))
			}
			wantInOrder = append(wantInOrder, wantLinks[slices.IndexFunc(kids, func(k []byte) bool { return bytes.Equal(k, order[i]) })])
		}
		check(t, lesserFirst(w.relShapes(got)), lesserFirst(wantInOrder)...)
	})
	t.Run("AC-1.9", func(t *testing.T) {
		w := newWorld(t)
		check(t, read(w, w.person("Ann")))
		check(t, read(w, make([]byte, 16)))
		check(t, read(w, []byte{1, 2, 3}))
		got, err := w.r.Relationships([][]byte{{1, 2, 3}, make([]byte, 16)}, Options{Siblings: true})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 0 {
			t.Fatalf("a wrong-length id and an unknown id gave %d entries", len(got))
		}
		bea := w.person("Bea")
		w.sibling("BB", bea, bea)
		check(t, read(w, bea), want(w, "Bea", "Bea", []string{"BB"}))
	})
	t.Run("AC-1.10", func(t *testing.T) {
		w := newWorld(t)
		ann, bea, cal := w.person("Ann"), w.person("Bea"), w.person("Cal")
		w.parent("AB", ann, bea)
		w.parent("AC", ann, cal)
		w.parent("CB", cal, bea)
		check(t, read(w, bea), want(w, "Bea", "Cal", []string{"AB", "AC"}, sharedParent("Ann")))
	})
	t.Run("an edge in both lists is read once", func(t *testing.T) {
		// fromNames reads the names of the associations of Bea's sibling
		// link's From edges, from the derivation itself.
		fromNames := func(w *world, id []byte) []string {
			steps, err := w.r.d.loadSteps([][]byte{id}, w.r.event)
			if err != nil {
				t.Fatal(err)
			}
			answers, err := w.r.d.siblingsOf(steps, w.r.event)
			if err != nil {
				t.Fatal(err)
			}
			var out []string
			for _, imp := range answers[string(id)] {
				for _, f := range imp.From {
					for _, a := range f.Associations {
						out = append(out, w.name(a))
					}
				}
			}
			slices.Sort(out)
			return out
		}
		w := newWorld(t)
		ann, bea, cal := w.person("Ann"), w.person("Bea"), w.person("Cal")
		w.parent("AB", ann, bea)
		w.parent("AC", ann, cal)
		w.siblings(bea)
		if got := fromNames(w, bea); !reflect.DeepEqual(got, []string{"AB", "AC"}) {
			t.Fatalf("From holds %v, want AB once and AC once", got)
		}
		w = newWorld(t)
		ann, bea, cal = w.person("Ann"), w.person("Bea"), w.person("Cal")
		b1, b2 := w.event("B1", vocabkey.EventBirth), w.event("B2", vocabkey.EventBirth)
		w.join("B1/Ann", ann, b1, "father")
		w.join("B1/Bea", bea, b1, "subject")
		w.join("B2/Ann", ann, b2, "father")
		w.join("B2/Cal", cal, b2, "subject")
		w.siblings(bea)
		if got := fromNames(w, bea); !reflect.DeepEqual(got, []string{"B1/Ann", "B1/Bea", "B2/Ann", "B2/Cal"}) {
			t.Fatalf("From holds %v, want each birth once", got)
		}
	})
	t.Run("AC-1.11", func(t *testing.T) {
		w := newWorld(t)
		ann, bea, cal := w.person("Ann"), w.person("Bea"), w.person("Cal")
		w.parent("AB", ann, bea)
		w.parent("AC", ann, cal)
		wantLink := want(w, "Bea", "Cal", []string{"AB", "AC"}, sharedParent("Ann"))
		got := w.siblings(bea)
		check(t, lesserFirst(w.relShapes(got)), wantLink)
		got[0].A[0] ^= 0xff
		got[0].B[0] ^= 0xff
		for _, a := range got[0].Associations {
			a[0] ^= 0xff
		}
		for _, sh := range got[0].Shapes {
			sh.Pivot[0] ^= 0xff
		}
		check(t, read(w, bea), wantLink)
	})
}

func TestSiblingsFollowCommits(t *testing.T) {
	w2 := func(t *testing.T) (w *world, ann, bea, cal []byte) {
		w = newWorld(t)
		ann, bea, cal = w.person("Ann"), w.person("Bea"), w.person("Cal")
		w.parent("AB", ann, bea)
		return
	}
	t.Run("AC-2.3", func(t *testing.T) {
		w, ann, bea, cal := w2(t)
		if got := w.siblings(bea); len(got) != 0 {
			t.Fatalf("setup: %+v", w.relShapes(got))
		}
		w.parent("AC", ann, cal)
		w.k.Commit(w.all[len(w.all)-1])
		got := lesserFirst(w.relShapes(w.siblings(bea)))
		if len(got) != 1 || got[0].A != "Bea" || got[0].B != "Cal" || !reflect.DeepEqual(got[0].Shapes, []string{sharedParent("Ann")}) {
			t.Fatalf("got %+v", got)
		}
	})
	filed := func(t *testing.T) (w *world, ann, bea []byte) {
		var cal []byte
		w, ann, bea, cal = w2(t)
		w.parent("AC", ann, cal)
		if len(w.siblings(bea)) != 1 {
			t.Fatal("setup: want one sibling")
		}
		return
	}
	// fresh reads Bea's siblings on a new world with the same links filed.
	fresh := func(t *testing.T) []relShape {
		w, ann, bea, cal := w2(t)
		w.parent("AC", ann, cal)
		return w.relShapes(w.siblings(bea))
	}
	t.Run("AC-2.4", func(t *testing.T) {
		w, ann, bea := filed(t)
		w.k.Commit(ann)
		if got := w.relShapes(w.siblings(bea)); len(got) != 1 || !reflect.DeepEqual(got, fresh(t)) {
			t.Fatalf("got %+v, want the fresh read %+v", got, fresh(t))
		}
	})
	t.Run("AC-2.5", func(t *testing.T) {
		w, _, bea := filed(t)
		w.k.CommitVocabulary()
		if got := w.relShapes(w.siblings(bea)); len(got) != 1 || !reflect.DeepEqual(got, fresh(t)) {
			t.Fatalf("got %+v, want the fresh read %+v", got, fresh(t))
		}
	})
	t.Run("AC-2.7", func(t *testing.T) {
		w, ann, bea := filed(t)
		for _, id := range slices.Clone(w.all) {
			if w.name(id) == "AB" || w.name(id) == "AC" {
				w.k.Remove(id)
				w.k.Commit(id)
			}
		}
		w.k.Remove(ann)
		w.k.Commit(ann)
		if got := w.siblings(bea); len(got) != 0 {
			t.Fatalf("got %+v, want none", w.relShapes(got))
		}
		if err := w.g.Verify(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestRelationshipsQueryCount(t *testing.T) {
	// family files a parent of a first child and of more, each also the subject
	// of a birth the parent is the father of; it returns the parent and the
	// first child.
	family := func(w *world, tag string, others int) (parent, first []byte) {
		parent = w.person(tag + "Ann")
		first = w.person(tag + "Bea")
		kids := [][]byte{first}
		for i := 0; i < others; i++ {
			kids = append(kids, w.person(fmt.Sprintf("%sK%d", tag, i)))
		}
		for i, kid := range kids {
			w.parent(fmt.Sprintf("%sAK%d", tag, i), parent, kid)
			b := w.event(fmt.Sprintf("%sB%d", tag, i), vocabkey.EventBirth)
			w.join(fmt.Sprintf("%sB%d/Ann", tag, i), parent, b, "father")
			w.join(fmt.Sprintf("%sB%d/K", tag, i), kid, b, "subject")
		}
		return parent, first
	}
	t.Run("warm", func(t *testing.T) {
		w := newWorld(t)
		parent, first := family(w, "", 3)
		ids := [][]byte{parent, first}
		got, err := w.r.Relationships(ids, Options{Siblings: true})
		if err != nil {
			t.Fatal(err)
		}
		var queries int
		w.g.SetQueryCounterForTest(func() { queries++ })
		again, err := w.r.Relationships(ids, Options{Siblings: true})
		if err != nil {
			t.Fatal(err)
		}
		if queries != 0 {
			t.Fatalf("a second read issued %d queries", queries)
		}
		if !reflect.DeepEqual(got, again) {
			t.Fatalf("second read differs:\n%+v\n%+v", got, again)
		}
	})
	t.Run("cold", func(t *testing.T) {
		count := func(build func(w *world) [][]byte) int {
			w := newWorld(t)
			ids := build(w)
			var queries int
			w.g.SetQueryCounterForTest(func() { queries++ })
			if _, err := w.r.Relationships(ids, Options{Siblings: true}); err != nil {
				t.Fatal(err)
			}
			return queries
		}
		a := count(func(w *world) [][]byte { p, f := family(w, "", 1); return [][]byte{p, f} })
		b := count(func(w *world) [][]byte { p, f := family(w, "", 7); return [][]byte{p, f} })
		c := count(func(w *world) [][]byte {
			x, xf := family(w, "x", 1)
			y, yf := family(w, "y", 1)
			z, zf := family(w, "z", 1)
			return [][]byte{x, xf, y, yf, z, zf}
		})
		if a == 0 || a != b || a != c {
			t.Fatalf("queries: one child %d, seven with births %d, three families %d", a, b, c)
		}
	})
}

func TestSiblingsLeaveMatchingAlone(t *testing.T) {
	// AC-3.2
	w := newWorld(t)
	ann, bea, cal := w.person("Ann"), w.person("Bea"), w.person("Cal")
	w.parent("AB", ann, bea)
	w.parent("AC", ann, cal)
	wantFor, wantRel := cloneImplied(w.for_(bea)), cloneLinks(w.relationships(bea))
	if len(w.siblings(bea)) != 1 {
		t.Fatal("setup: want one sibling link")
	}
	if got := w.for_(bea); !reflect.DeepEqual(got, wantFor) {
		t.Fatalf("For changed:\n%+v\n%+v", got, wantFor)
	}
	if got := w.relationships(bea); !reflect.DeepEqual(got, wantRel) {
		t.Fatalf("Relationships changed:\n%+v\n%+v", got, wantRel)
	}
	for _, l := range w.for_(bea) {
		if l.Pattern == kinship.PatternSharedParent {
			t.Fatalf("For holds %+v", l)
		}
	}
	for _, l := range w.relationships(bea) {
		for _, sh := range l.Shapes {
			if sh.Pattern == kinship.PatternSharedParent {
				t.Fatalf("Relationships holds %+v", l)
			}
		}
	}
}

func TestPivotKindMismatch(t *testing.T) {
	// Ann stands in Bea's birth and is also filed as the parent of the birth's
	// handle, so her chain reads the birth as a person. The handle's event
	// result is stored first; the chain must not be served it.
	w := newWorld(t)
	ann, bea := w.person("Ann"), w.person("Bea")
	b1 := w.event("B1", vocabkey.EventBirth)
	w.join("B1/Ann", ann, b1, "mother")
	w.join("B1/Bea", bea, b1, "subject")
	w.parent("AB1", ann, b1)
	got := w.for_(ann)
	seen := map[string]bool{}
	for _, l := range got {
		key := fmt.Sprintf("%x|%x|%s|%x|%x", l.A, l.B, l.Pattern, l.Event, l.Through)
		if seen[key] {
			t.Fatalf("For(Ann) repeats %+v in %+v", l, got)
		}
		seen[key] = true
	}
	if len(got) == 0 {
		t.Fatal("the fixture implies nothing for Ann")
	}
}

func TestComposedAnswersReusePivots(t *testing.T) {
	// Ann and Bob married (M1); Bea born of Ann and Bob (B1); Cal born of Bea
	// (B2).
	build := func(t *testing.T) (w *world, people []*[]byte, ann, cal []byte) {
		w = newWorld(t)
		ann, bob, bea := w.person("Ann"), w.person("Bob"), w.person("Bea")
		cal = w.person("Cal")
		w.marriage("M1", ann, bob)
		b1, b2 := w.event("B1", vocabkey.EventBirth), w.event("B2", vocabkey.EventBirth)
		w.join("B1/Ann", ann, b1, "father")
		w.join("B1/Bob", bob, b1, "mother")
		w.join("B1/Bea", bea, b1, "subject")
		w.join("B2/Bea", bea, b2, "mother")
		w.join("B2/Cal", cal, b2, "subject")
		return w, []*[]byte{&ann, &bob, &bea, &cal}, ann, cal
	}
	heldOf := func(w *world) map[string]bool {
		out := map[string]bool{}
		for _, id := range w.all {
			if w.hasPivot(id) {
				out[string(id)] = true
			}
		}
		return out
	}
	readAll := func(w *world, people []*[]byte) {
		for _, p := range people {
			w.for_(*p)
		}
	}
	t.Run("commit ann", func(t *testing.T) {
		w, people, ann, _ := build(t)
		readAll(w, people)
		held := heldOf(w)

		n, ok := w.g.Cached(ann)
		if !ok {
			t.Fatal("Ann's node is not cached")
		}
		expected := map[string]bool{string(ann): true}
		for _, l := range n.Links {
			expected[string(l.Neighbor)] = true
		}
		var names []string
		for id := range expected {
			names = append(names, w.name([]byte(id)))
		}
		sort.Strings(names)
		if !reflect.DeepEqual(names, []string{"Ann", "B1", "M1"}) {
			t.Fatalf("the ids a commit on Ann names: %v", names)
		}

		w.k.Commit(ann)
		after := heldOf(w)
		dropped := map[string]bool{}
		for id := range held {
			if !after[id] {
				dropped[id] = true
			}
		}
		for id := range after {
			if !held[id] {
				t.Fatalf("%s is held after a commit and was not before", w.name([]byte(id)))
			}
		}
		for id := range expected {
			if held[id] != dropped[id] {
				t.Fatalf("%s: held before %v, dropped %v", w.name([]byte(id)), held[id], dropped[id])
			}
		}
		if len(dropped) > len(expected) {
			t.Fatalf("dropped %d pivots, want at most the %d ids the commit names", len(dropped), len(expected))
		}

		readAll(w, people)
		derived := map[string]bool{}
		for id := range heldOf(w) {
			if !after[id] {
				derived[id] = true
			}
		}
		if !reflect.DeepEqual(derived, dropped) {
			t.Fatalf("derived %d pivots, dropped %d", len(derived), len(dropped))
		}
	})
	t.Run("commit cal", func(t *testing.T) {
		w, people, ann, cal := build(t)
		readAll(w, people)
		before := w.shapes(w.for_(ann))
		held := heldOf(w)
		w.k.Commit(cal)
		held2 := heldOf(w)
		got := w.shapes(w.for_(ann))
		for id := range heldOf(w) {
			if !held2[id] {
				t.Fatalf("For(Ann) derived the pivot of %s", w.name([]byte(id)))
			}
		}
		if !reflect.DeepEqual(got, before) {
			t.Fatalf("For(Ann) changed:\n%+v\n%+v", got, before)
		}
		if len(held2) >= len(held) {
			t.Fatalf("the commit dropped no pivot: held %d before, %d after", len(held), len(held2))
		}
	})
}

func TestEdgeKey(t *testing.T) {
	e := func(a, b string, assoc ...string) kinship.Edge {
		edge := kinship.Edge{A: []byte(a), B: []byte(b)}
		for _, x := range assoc {
			edge.Associations = append(edge.Associations, []byte(x))
		}
		return edge
	}
	for _, c := range []struct {
		name      string
		x, y      kinship.Edge
		wantEqual bool
	}{
		{"ends split differently", e("ab", "c"), e("a", "bc"), false},
		{"associations split differently", e("a", "b", "xy", "z"), e("a", "b", "x", "yz"), false},
		{"same edge", e("a", "b", "x", "y"), e("a", "b", "x", "y"), true},
		{"ends swapped with an empty end", e("a", ""), e("", "a"), false},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := edgeKey(c.x) == edgeKey(c.y); got != c.wantEqual {
				t.Fatalf("keys equal = %v, want %v", got, c.wantEqual)
			}
		})
	}
}
