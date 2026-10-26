package promotealign

import (
	"bytes"
	"maps"
	"slices"
	"sort"
	"strings"

	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/kinship"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/segmentcompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// candidatesMemoKey is the Evidence graph memo key of each Subject's candidates,
// with the name pattern they were ranked under after it.
const candidatesMemoKey = "promotealign.candidates\x00"

// canonWalkMemoKey is the Evidence graph memo key of the canon walk.
const canonWalkMemoKey = "promotealign.canonWalk\x00"

// canonWalk is the edges the canon walk made from the seeds, the handles it
// reached (in Ref-then-ID order, with their kinds), the seeds it started from
// (seedsKey) and the generation of the match index it ran against. A stored
// value is never mutated.
type canonWalk struct {
	gen     uint64
	seeds   string
	edges   []graphalign.CanonEdge
	handles []walkedHandle
}

type walkedHandle struct {
	id   []byte
	kind string
}

// seedsKey is the seeds as one comparable key: sorted, deduplicated and
// concatenated, so the same set of seeds in any order has the same key.
func seedsKey(seeds [][]byte) string {
	ids := make([]string, 0, len(seeds))
	for _, s := range seeds {
		ids = append(ids, string(s))
	}
	sort.Strings(ids)
	return strings.Join(slices.Compact(ids), "")
}

// candidatesMemo is the candidates (handle ids, in order) of each Subject the
// graph's Evidence graph has had ranked, and the generation of the match index
// they were ranked against. Candidates depend on the Subject's values and the
// vocabulary and aliases (the memo is dropped with the Evidence graph's other
// memos), the handles' profiles and the shortlist index (dropped together on
// every handle commit, which gives the next index a new generation), and the
// name pattern (in the key). A stored memo is never changed.
type candidatesMemo struct {
	gen       uint64
	bySubject map[string][][]byte
}

// loadCanon reads the open catalog's graph: fixed anchors, the candidates of
// each unfixed Subject (loadCandidatesMemo), then every link those handles
// have. Link order is association ref, then the neighbor's ref, which is the
// order the walk used to break ties. Every handle's profile is built once: the
// shortlisted ones together, the ones the walk reaches after it. Links go
// through one linkSet, so links with the same ends and keys merge, and the
// links the kinship patterns imply follow the recorded ones (with no
// association of their own beyond those they were made from). The walk is
// remembered by loadCanonWalk until a handle commit, a different set of seeds
// or a change to the Evidence graph's memos, as candidatesMemo is, because it
// reads only what those change.
func loadCanon(g *graphcache.Graph, graphID []byte, prof *profiler, layer graphalign.Layer, fixed []graphalign.Fixed, cfg graphalign.Config, terms segmentcompare.Terms) (graphalign.Canon, error) {
	props := prof.props
	types := prof.types

	handles := map[string]*graphcache.Node{}
	kinds := map[string]string{}
	var seeds [][]byte
	candidates := map[string][][]byte{}
	add := func(id []byte) error {
		if len(id) != 16 {
			return nil
		}
		if _, ok := handles[string(id)]; ok {
			seeds = append(seeds, id)
			return nil
		}
		n, err := g.Node(id)
		if err != nil || n == nil || n.Merged {
			return err
		}
		kind := typeKey(types, n.SubjectTypeID)
		if !vocabkey.IsPrimaryKind(kind) {
			return nil
		}
		handles[string(id)] = n
		kinds[string(id)] = kind
		seeds = append(seeds, id)
		return nil
	}

	for _, f := range fixed {
		if len(f.HandleID) == 16 {
			if err := add(f.HandleID); err != nil {
				return graphalign.Canon{}, err
			}
		}
	}

	fixedSubjects := map[string]bool{}
	for _, f := range fixed {
		fixedSubjects[string(f.SubjectID)] = true
	}
	var todo []graphalign.Subject
	for _, s := range layer.Subjects {
		if fixedSubjects[string(s.ID)] {
			continue
		}
		if _, ok := cfg.Subject.Weights[vocabkey.SubjectType(s.Kind)]; !ok {
			continue
		}
		todo = append(todo, s)
	}
	if len(todo) > 0 {
		known, err := loadCandidatesMemo(g, graphID, prof, layer, todo, cfg)
		if err != nil {
			return graphalign.Canon{}, err
		}
		for _, s := range todo {
			for _, id := range known[string(s.ID)] {
				candidates[string(s.ID)] = append(candidates[string(s.ID)], id)
				if err := add(id); err != nil {
					return graphalign.Canon{}, err
				}
			}
		}
	}

	sort.SliceStable(seeds, func(i, j int) bool {
		return handles[string(seeds[i])].Ref < handles[string(seeds[j])].Ref
	})

	edges, err := loadCanonWalk(g, graphID, seeds, handles, kinds, props, types, terms)
	if err != nil {
		return graphalign.Canon{}, err
	}

	nodes := make([]*graphcache.Node, 0, len(handles))
	for _, n := range handles {
		nodes = append(nodes, n)
	}
	if err := prof.profile(nodes); err != nil {
		return graphalign.Canon{}, err
	}
	out := graphalign.Canon{Edges: edges, Candidates: candidates}
	for id, n := range handles {
		out.Handles = append(out.Handles, graphalign.Handle{
			ID: append([]byte(nil), n.ID...), Ref: n.Ref, Kind: kinds[id], Values: prof.values(n.ID),
		})
	}
	sort.Slice(out.Handles, func(i, j int) bool {
		if out.Handles[i].Ref != out.Handles[j].Ref {
			return out.Handles[i].Ref < out.Handles[j].Ref
		}
		return bytes.Compare(out.Handles[i].ID, out.Handles[j].ID) < 0
	})
	return out, nil
}

// walkCanon follows every filed link from the seeds, wave by wave, and returns
// the edges they make: each recorded link as one edge, merged with the others
// that have its ends and keys, then the edges the kinship patterns imply. It
// adds each handle it reaches to handles and kinds.
func walkCanon(g *graphcache.Graph, seeds [][]byte, handles map[string]*graphcache.Node, kinds map[string]string, props map[string]graphcache.VocabProperty, types map[string][]byte, terms segmentcompare.Terms) ([]graphalign.CanonEdge, error) {
	set := newLinkSet()
	walked := map[string]bool{}
	queue := append([][]byte(nil), seeds...)
	for len(queue) > 0 {
		wave := queue
		queue = nil
		type reached struct {
			id []byte
			n  *graphcache.Node
		}
		var ready []reached
		var prefetch [][]byte
		for _, id := range wave {
			if walked[string(id)] {
				continue
			}
			walked[string(id)] = true
			n, err := g.Node(id)
			if err != nil {
				return nil, err
			}
			if n == nil {
				continue
			}
			ready = append(ready, reached{id, n})
			for _, l := range n.Links {
				prefetch = append(prefetch, l.Neighbor)
			}
		}
		if err := g.Nodes(prefetch); err != nil {
			return nil, err
		}
		for _, item := range ready {
			id, n := item.id, item.n
			selfKind := typeKey(types, n.SubjectTypeID)
			for _, l := range n.Links {
				neighbor, err := g.Node(l.Neighbor)
				if err != nil {
					return nil, err
				}
				if neighbor == nil || neighbor.Merged {
					continue
				}
				nKind := typeKey(types, neighbor.SubjectTypeID)
				if !vocabkey.IsPrimaryKind(nKind) {
					continue
				}
				if from, to, keyFrom, keyTo, ok := canonKeys(id, selfKind, n, l, neighbor, props, terms); ok {
					set.add(sideLink{Edge: kinship.Edge{A: from, B: to, KeyA: keyFrom, KeyB: keyTo}})
				}
				if _, ok := handles[string(neighbor.ID)]; !ok {
					handles[string(neighbor.ID)] = neighbor
					kinds[string(neighbor.ID)] = nKind
					queue = append(queue, neighbor.ID)
				}
			}
		}
	}

	set.withImplied(terms, matchregistry.Default().Kinship)
	var edges []graphalign.CanonEdge
	for _, l := range set.links {
		edges = append(edges, graphalign.CanonEdge{
			From: l.A, To: l.B, KeyFrom: l.KeyA, KeyTo: l.KeyB, Shapes: l.Shapes, OnlyImplied: l.OnlyImplied,
		})
	}
	return edges, nil
}

// canonKeys reads one filed link as an edge between its two ends, ordered as
// recorded, with the key it has from each. A link whose bridge is not
// registered is not an edge.
func canonKeys(self []byte, selfKind string, n *graphcache.Node, l graphcache.Link, neighbor *graphcache.Node, props map[string]graphcache.VocabProperty, terms segmentcompare.Terms) (from, to []byte, keyFrom, keyTo vocabkey.BridgeKey, ok bool) {
	b, ok := connectrules.LookupBridge(l.BridgeTypeKey)
	if !ok || len(b.Endpoints) != 2 {
		return nil, nil, vocabkey.BridgeKey{}, vocabkey.BridgeKey{}, false
	}
	if l.FromEnd {
		from, to = self, l.Neighbor
	} else {
		from, to = l.Neighbor, self
	}
	// The event type comes from whichever end is an event; `segmentcompare.Keys` alone
	// decides whether the key keeps it.
	var eventType string
	if first, has := segmentcompare.EventEndpoint(b); has {
		// The first endpoint is the edge's from end; self is that end when
		// the link was filed from it.
		if first == l.FromEnd {
			eventType = keptTerm(n, props, neighborTypeProperty[vocabkey.Event])
		} else {
			eventType = keptTerm(neighbor, props, neighborTypeProperty[vocabkey.Event])
		}
	}
	keyFrom, keyTo = segmentcompare.Keys(b, vocabkey.BridgeTerm(l.TermKey), vocabkey.EventType(eventType), terms)
	return from, to, keyFrom, keyTo, true
}

func keptTerm(n *graphcache.Node, props map[string]graphcache.VocabProperty, key string) string {
	if n == nil || key == "" {
		return ""
	}
	for propID, rows := range n.Values {
		if props[propID].Key != key {
			continue
		}
		for _, row := range rows {
			if row.Rank == 1 && row.Reason == "kept" {
				return row.TermKey
			}
		}
	}
	return ""
}

func typeByKey(g *graphcache.Graph) (map[string][]byte, error) {
	v, err := g.Vocabulary()
	if err != nil {
		return nil, err
	}
	out := map[string][]byte{}
	for _, t := range v.Types {
		if t.Origin != vocabkey.OriginProvenencia {
			continue
		}
		out[t.Key] = append([]byte(nil), t.ID...)
	}
	return out, nil
}

func typeKey(types map[string][]byte, id []byte) string {
	for key, typeID := range types {
		if string(typeID) == string(id) {
			return key
		}
	}
	return ""
}

// loadCandidatesMemo is the ranked candidate ids of each Subject in todo:
// graphalign.Candidates over the handles the shortlist index offers it,
// remembered on the Evidence graph's memo until the index changes. It reads the
// stored memo for the current index generation, ranks only the Subjects it
// lacks, and stores a new map, never changing the stored one.
func loadCandidatesMemo(g *graphcache.Graph, graphID []byte, prof *profiler, layer graphalign.Layer, todo []graphalign.Subject, cfg graphalign.Config) (map[string][][]byte, error) {
	ix, gen, err := loadShortlist(g, prof, cfg)
	if err != nil {
		return nil, err
	}
	key := candidatesMemoKey + prof.namePattern
	var known map[string][][]byte
	if v, ok := g.EvidenceGraphMemo(graphID, key); ok {
		if m, ok := v.(candidatesMemo); ok && m.gen == gen {
			known = m.bySubject
		}
	}
	var missing []graphalign.Subject
	for _, s := range todo {
		if _, ok := known[string(s.ID)]; !ok {
			missing = append(missing, s)
		}
	}
	if len(missing) == 0 {
		return known, nil
	}
	shortlisted, err := seedHandles(g, prof, ix, missing)
	if err != nil {
		return nil, err
	}
	stored := make(map[string][][]byte, len(known)+len(missing))
	maps.Copy(stored, known)
	for i, s := range missing {
		ranked := [][]byte{}
		for _, h := range graphalign.Candidates(s.Kind, s.Values, shortlisted[i], layer.Metas, cfg) {
			ranked = append(ranked, h.ID)
		}
		stored[string(s.ID)] = ranked
	}
	g.SetEvidenceGraphMemo(graphID, key, candidatesMemo{gen: gen, bySubject: stored})
	return stored, nil
}

// loadCanonWalk is walkCanon's edges, remembered on the Evidence graph's memo
// beside the candidates and cleared with it (Drop, a vocabulary commit,
// SetAliasIndex, SetNameFormat, a commit to that graph). The walk reads only
// handles' links, values and kinds, which change only in a handle commit, and
// that gives the next match index a new generation; so a walk made for the
// same seeds under the same generation is the walk again. A catalog with no
// match index (generation 0) stores nothing. A hit adds the stored handles to
// handles and kinds and returns the stored edges, shared and read-only; a miss
// walks and stores a new canonWalk, never changing a stored one.
func loadCanonWalk(g *graphcache.Graph, graphID []byte, seeds [][]byte, handles map[string]*graphcache.Node, kinds map[string]string, props map[string]graphcache.VocabProperty, types map[string][]byte, terms segmentcompare.Terms) ([]graphalign.CanonEdge, error) {
	gen := matchIndexOf(g).gen
	key := seedsKey(seeds)
	if gen != 0 {
		if v, ok := g.EvidenceGraphMemo(graphID, canonWalkMemoKey); ok {
			if m, ok := v.(canonWalk); ok && m.gen == gen && m.seeds == key {
				for _, h := range m.handles {
					if _, ok := handles[string(h.id)]; ok {
						continue
					}
					n, err := g.Node(h.id)
					if err != nil {
						return nil, err
					}
					handles[string(h.id)] = n
					kinds[string(h.id)] = h.kind
				}
				return m.edges, nil
			}
		}
	}
	edges, err := walkCanon(g, seeds, handles, kinds, props, types, terms)
	if err != nil {
		return nil, err
	}
	if gen != 0 {
		m := canonWalk{gen: gen, seeds: key, edges: edges}
		for id := range handles {
			m.handles = append(m.handles, walkedHandle{id: []byte(id), kind: kinds[id]})
		}
		sort.Slice(m.handles, func(i, j int) bool {
			ri, rj := handles[string(m.handles[i].id)].Ref, handles[string(m.handles[j].id)].Ref
			if ri != rj {
				return ri < rj
			}
			return bytes.Compare(m.handles[i].id, m.handles[j].id) < 0
		})
		g.SetEvidenceGraphMemo(graphID, canonWalkMemoKey, m)
	}
	return edges, nil
}
