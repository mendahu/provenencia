package graphcache

import (
	"fmt"
	"slices"

	"github.com/mendahu/provenencia/core/database/effects"
)

// The registry of derived caches. An owner registers one typed handle per
// cache, from init, and reads and writes only through it. The graph never
// imports the owner. A cache is either keyed (one value per node id, filled by
// the reader) or whole (one value, built lazily or stored by the owner). Each
// entry says what a commit does to it (Apply), what it is derived from
// (DerivedFrom), how long it lives (Lifetime), how Verify checks it (Check)
// and, for a keyed entry, how many keys it may hold (Cap).

// Lifetime says how long an entry's value lives.
type Lifetime int

const (
	// LifeGraph entries live with the graph's nodes: Drop empties them.
	LifeGraph Lifetime = iota
	// LifeCatalog entries live with the catalog connection: Drop leaves them.
	LifeCatalog
)

// Source is what an entry can be derived from: the vocabulary, or a Whole
// handle registered earlier.
type Source interface{ source() int }

// vocabularySource is the id Vocabulary reports as a Source.
const vocabularySource = -1

// source makes Vocabulary{} a Source: an entry derived from it is emptied by a
// vocabulary commit.
func (Vocabulary) source() int { return vocabularySource }

// KeyedEntry registers a cache with one value of type T per 16-byte id (a
// handle's, or a graph's for entries whose Apply is DropNamedGraphs).
type KeyedEntry[T any] struct {
	Name string
	// Apply runs after a commit's reload when the entry holds data. roots are
	// the handles the commit reloaded or touched (nil when it names none).
	// Required.
	Apply func(g *Graph, k Keyed[T], roots [][]byte, set effects.Set) error
	// DerivedFrom lists what a change empties this entry for.
	DerivedFrom []Source
	// Check rebuilds one stored value and compares it. Nil: Verify skips.
	Check func(g *Graph, id []byte, stored T) error
	// Cap is the most keys held; past it the least recently used leaves.
	// Required, > 0.
	Cap int
}

// WholeEntry registers a cache with one value of type T.
type WholeEntry[T any] struct {
	Name string
	// Build makes the value on the first Get. Nil: the owner calls Set.
	Build func(g *Graph) (T, error)
	// Apply runs after a commit's reload when the entry holds a value. Nil:
	// commits leave it.
	Apply       func(g *Graph, w Whole[T], roots [][]byte, set effects.Set) error
	DerivedFrom []Source
	Lifetime    Lifetime
	// Check rebuilds the stored value and compares it. Nil: Verify skips.
	Check func(g *Graph, stored T) error
}

// Keyed is the handle of a registered keyed entry.
type Keyed[T any] struct{ id int }

// Whole is the handle of a registered whole entry.
type Whole[T any] struct{ id int }

func (w Whole[T]) source() int { return w.id }

type entry struct {
	name  string
	whole bool
	life  Lifetime
	from  []int
	cap   int
	build any
	apply func(g *Graph, roots [][]byte, set effects.Set) error
	check func(g *Graph) error
	held  func(g *Graph) bool
	count func(g *Graph) int
}

// registry holds every entry in registration order. An entry's id is its
// index plus one, so a zero handle names nothing.
var registry []*entry

func register(e *entry, from []Source) int {
	if e.name == "" {
		panic("graphcache: an entry needs a Name")
	}
	for _, r := range registry {
		if r.name == e.name {
			panic("graphcache: entry " + e.name + " is registered twice")
		}
	}
	for _, s := range from {
		if s == nil {
			panic("graphcache: entry " + e.name + " is derived from nil")
		}
		id := s.source()
		if id == vocabularySource {
			if e.life == LifeCatalog {
				panic("graphcache: catalog entry " + e.name + " cannot derive from the vocabulary")
			}
			e.from = append(e.from, id)
			continue
		}
		if id < 1 || id > len(registry) {
			panic("graphcache: entry " + e.name + " is derived from an unregistered entry")
		}
		if e.life == LifeCatalog && registry[id-1].life == LifeGraph {
			panic("graphcache: catalog entry " + e.name + " cannot derive from graph entry " + registry[id-1].name)
		}
		e.from = append(e.from, id)
	}
	registry = append(registry, e)
	return len(registry)
}

// RegisterKeyed registers a keyed entry. Owners call it from init.
func RegisterKeyed[T any](e KeyedEntry[T]) Keyed[T] {
	if e.Cap <= 0 {
		panic("graphcache: keyed entry " + e.Name + " needs a Cap")
	}
	if e.Apply == nil {
		panic("graphcache: keyed entry " + e.Name + " needs an Apply")
	}
	info := &entry{name: e.Name, life: LifeGraph, cap: e.Cap}
	k := Keyed[T]{id: register(info, e.DerivedFrom)}
	info.apply = func(g *Graph, roots [][]byte, set effects.Set) error {
		return e.Apply(g, k, roots, set)
	}
	info.held = func(g *Graph) bool { return k.data(g).len() > 0 }
	info.count = func(g *Graph) int { return k.data(g).len() }
	if e.Check != nil {
		// A Check reads the stored values without counting a use.
		info.check = func(g *Graph) error {
			kv := k.data(g)
			if kv == nil {
				return nil
			}
			for key, it := range kv.items {
				if err := e.Check(g, []byte(key), it.v); err != nil {
					return err
				}
			}
			return nil
		}
	}
	return k
}

// RegisterWhole registers a whole entry. Owners call it from init.
func RegisterWhole[T any](e WholeEntry[T]) Whole[T] {
	info := &entry{name: e.Name, whole: true, life: e.Lifetime}
	w := Whole[T]{id: register(info, e.DerivedFrom)}
	if e.Build != nil {
		info.build = e.Build
	}
	if e.Apply != nil {
		info.apply = func(g *Graph, roots [][]byte, set effects.Set) error {
			return e.Apply(g, w, roots, set)
		}
	}
	info.held = func(g *Graph) bool { _, ok := w.held(g); return ok }
	if e.Check != nil {
		info.check = func(g *Graph) error {
			if v, ok := w.held(g); ok {
				return e.Check(g, v)
			}
			return nil
		}
	}
	return w
}

// EntryNames returns every registered entry's Name in registration order.
func EntryNames() []string {
	out := make([]string, len(registry))
	for i, e := range registry {
		out[i] = e.name
	}
	return out
}

// KeyCounts returns the keys each keyed entry holds, by Name.
func (g *Graph) KeyCounts() map[string]int {
	out := map[string]int{}
	if g == nil {
		return out
	}
	for _, e := range registry {
		if !e.whole {
			out[e.name] = e.count(g)
		}
	}
	return out
}

// --- keyed handle ---

type keyedValues[T any] struct {
	items       map[string]*keyedItem[T]
	front, back *keyedItem[T]
}

type keyedItem[T any] struct {
	key        string
	v          T
	prev, next *keyedItem[T]
}

func (kv *keyedValues[T]) len() int {
	if kv == nil {
		return 0
	}
	return len(kv.items)
}

func (kv *keyedValues[T]) unlink(it *keyedItem[T]) {
	if it.prev != nil {
		it.prev.next = it.next
	} else {
		kv.front = it.next
	}
	if it.next != nil {
		it.next.prev = it.prev
	} else {
		kv.back = it.prev
	}
	it.prev, it.next = nil, nil
}

func (kv *keyedValues[T]) pushFront(it *keyedItem[T]) {
	it.next = kv.front
	if kv.front != nil {
		kv.front.prev = it
	}
	kv.front = it
	if kv.back == nil {
		kv.back = it
	}
}

func (kv *keyedValues[T]) touch(it *keyedItem[T]) {
	if kv.front == it {
		return
	}
	kv.unlink(it)
	kv.pushFront(it)
}

func (k Keyed[T]) data(g *Graph) *keyedValues[T] {
	kv, _ := g.entries[k.id].(*keyedValues[T])
	return kv
}

// capOf is the most keys the entry holds on g: its Cap, or the test override.
func (k Keyed[T]) capOf(g *Graph) int {
	if n, ok := g.capsForTest[k.id]; ok {
		return n
	}
	return registry[k.id-1].cap
}

// SetCapForTest overrides the entry's Cap for g, evicting at once down to it.
// The override survives Drop and Clear and applies to data created later; n <= 0
// restores the Cap. Tests only.
func (k Keyed[T]) SetCapForTest(g *Graph, n int) {
	if g == nil {
		return
	}
	if n <= 0 {
		delete(g.capsForTest, k.id)
		return
	}
	if g.capsForTest == nil {
		g.capsForTest = map[int]int{}
	}
	g.capsForTest[k.id] = n
	if kv := k.data(g); kv != nil {
		for len(kv.items) > n {
			kv.evict()
		}
	}
}

// evict removes the least recently used key.
func (kv *keyedValues[T]) evict() {
	it := kv.back
	kv.unlink(it)
	delete(kv.items, it.key)
}

// hit returns the value held for id and counts a use. A key that is not 16
// bytes is never held.
func (k Keyed[T]) hit(g *Graph, id []byte) (T, bool) {
	var zero T
	kv := k.data(g)
	if kv == nil || len(id) != 16 {
		return zero, false
	}
	it, ok := kv.items[string(id)]
	if !ok {
		return zero, false
	}
	kv.touch(it)
	return it.v, true
}

// store keeps v under id as the most recently used key, evicting the least
// recently used one when a new key arrives at the Cap.
func (k Keyed[T]) store(g *Graph, id []byte, v T) {
	if g == nil || len(id) != 16 {
		return
	}
	kv := k.data(g)
	if kv == nil {
		kv = &keyedValues[T]{items: map[string]*keyedItem[T]{}}
		if g.entries == nil {
			g.entries = map[int]any{}
		}
		g.entries[k.id] = kv
	}
	if it, ok := kv.items[string(id)]; ok {
		it.v = v
		kv.touch(it)
		return
	}
	for len(kv.items) >= k.capOf(g) {
		kv.evict()
	}
	key := string(id) // one copy, shared by the map and the item
	it := &keyedItem[T]{key: key, v: v}
	kv.items[key] = it
	kv.pushFront(it)
}

// Get returns the value held for id, else calls fill, stores its result and
// returns it. A hit and a store count as a use of the key; a new key at the
// Cap evicts the least recently used one. A fill error is returned and nothing
// is stored. A nil graph returns the zero value.
func (k Keyed[T]) Get(g *Graph, id []byte, fill func() (T, error)) (T, error) {
	var zero T
	if g == nil {
		return zero, nil
	}
	if v, ok := k.hit(g, id); ok {
		return v, nil
	}
	v, err := fill()
	if err != nil {
		return zero, err
	}
	k.store(g, id, v)
	return v, nil
}

// GetMany returns the values for ids, calling fill once with every id not held
// and storing what it returns for those ids. Held ids count as a use first, in
// ids order, then the fill's values are stored in the order of the missing ids;
// a key the fill returns that was not asked for is neither stored nor
// returned. The map holds each id that has a value, even one a later store
// evicted. A fill error is returned and nothing is stored.
func (k Keyed[T]) GetMany(g *Graph, ids [][]byte, fill func(missing [][]byte) (map[string]T, error)) (map[string]T, error) {
	out := map[string]T{}
	if g == nil {
		return out, nil
	}
	var missing [][]byte
	asked := map[string]bool{}
	for _, id := range ids {
		if v, ok := k.hit(g, id); ok {
			out[string(id)] = v
			continue
		}
		if !asked[string(id)] {
			asked[string(id)] = true
			missing = append(missing, id)
		}
	}
	if len(missing) == 0 {
		return out, nil
	}
	filled, err := fill(missing)
	if err != nil {
		return nil, err
	}
	for _, id := range missing {
		v, ok := filled[string(id)]
		if !ok {
			continue
		}
		k.store(g, id, v)
		out[string(id)] = v
	}
	return out, nil
}

// Peek returns the value held for id and never fills. It counts as a use.
func (k Keyed[T]) Peek(g *Graph, id []byte) (T, bool) {
	if g == nil {
		var zero T
		return zero, false
	}
	return k.hit(g, id)
}

// Set stores v for id as the most recently used key. An id that is not 16
// bytes stores nothing.
func (k Keyed[T]) Set(g *Graph, id []byte, v T) { k.store(g, id, v) }

// Drop forgets the values held for ids.
func (k Keyed[T]) Drop(g *Graph, ids [][]byte) {
	if g == nil {
		return
	}
	kv := k.data(g)
	if kv == nil {
		return
	}
	for _, id := range ids {
		if it, ok := kv.items[string(id)]; ok {
			kv.unlink(it)
			delete(kv.items, string(id))
		}
	}
}

// Clear forgets every value held.
func (k Keyed[T]) Clear(g *Graph) {
	if g == nil {
		return
	}
	delete(g.entries, k.id)
}

// DropNamedGraphs is the Apply of an entry keyed by Evidence graph: a commit
// drops the graphs its effects name (Set.EvidenceGraphs) and nothing else.
func DropNamedGraphs[T any](g *Graph, k Keyed[T], _ [][]byte, set effects.Set) error {
	k.Drop(g, set.EvidenceGraphs)
	return nil
}

// --- whole handle ---

func (w Whole[T]) held(g *Graph) (T, bool) {
	var zero T
	if g == nil {
		return zero, false
	}
	v, ok := g.entries[w.id].(wholeValue[T])
	if !ok {
		return zero, false
	}
	return v.v, true
}

type wholeValue[T any] struct{ v T }

// Get returns the held value, else runs the entry's Build, stores and returns
// it. With no Build it returns false and the owner stores with Set. A first
// fill by Build empties no reader. On a write's graph (Over) a catalog entry
// builds through the transaction, once per graph; a graph entry is neither
// built nor stored, so ok is false and the owner composes without it.
func (w Whole[T]) Get(g *Graph) (T, bool, error) {
	var zero T
	if g == nil {
		return zero, false, nil
	}
	if v, ok := w.held(g); ok {
		return v, true, nil
	}
	e := registry[w.id-1]
	build, _ := e.build.(func(*Graph) (T, error))
	if build == nil || (g.alt != nil && e.life == LifeGraph) {
		return zero, false, nil
	}
	v, err := build(g)
	if err != nil {
		return zero, false, err
	}
	w.put(g, v)
	return v, true, nil
}

func (w Whole[T]) put(g *Graph, v T) {
	if g.entries == nil {
		g.entries = map[int]any{}
	}
	g.entries[w.id] = wholeValue[T]{v: v}
}

// Set stores v and empties every entry derived from this one. On a write's
// graph a graph-lifetime entry stores nothing.
func (w Whole[T]) Set(g *Graph, v T) {
	if g == nil || (g.alt != nil && registry[w.id-1].life == LifeGraph) {
		return
	}
	w.put(g, v)
	g.dropReaders(w.id)
}

// Drop empties the entry and every entry derived from it.
func (w Whole[T]) Drop(g *Graph) {
	if g == nil {
		return
	}
	delete(g.entries, w.id)
	g.dropReaders(w.id)
}

// --- graph side ---

// dropReaders empties every entry derived from src, then those derived from
// them. A slot derives only from earlier ones, so it ends.
func (g *Graph) dropReaders(src int) {
	for i, e := range registry {
		if slices.Contains(e.from, src) {
			delete(g.entries, i+1)
			g.dropReaders(i + 1)
		}
	}
}

// dropGraphEntries empties every graph-lifetime entry.
func (g *Graph) dropGraphEntries() {
	for i, e := range registry {
		if e.life == LifeGraph {
			delete(g.entries, i+1)
		}
	}
}

// applyEntries runs each held entry's Apply, in registration order. The first
// error stops the loop.
func (g *Graph) applyEntries(roots [][]byte, set effects.Set) error {
	for _, e := range registry {
		if e.apply == nil || !e.held(g) {
			continue
		}
		if err := e.apply(g, roots, set); err != nil {
			return err
		}
	}
	return nil
}

// verifyEntries runs each entry's Check over its stored values.
func (g *Graph) verifyEntries() error {
	for _, e := range registry {
		if e.check == nil {
			continue
		}
		if err := e.check(g); err != nil {
			return fmt.Errorf("graphcache: entry %s: %w", e.name, err)
		}
	}
	return nil
}
