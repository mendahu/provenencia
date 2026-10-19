package promotealign

import (
	"bytes"
	"sort"

	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/segmentcompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// loadCanon reads the open catalog's graph: fixed anchors, the candidates of
// each unfixed Subject (graphalign.Candidates over the handles the shortlist
// index offers it), then every link those handles have. Link order is
// association ref, then the neighbor's ref, which is the order the walk used to
// break ties. Every handle's profile is built once: the shortlisted ones
// together, the ones the walk reaches after it.
func loadCanon(g *graphcache.Graph, prof *profiler, layer graphalign.Layer, fixed []graphalign.Fixed, cfg graphalign.Config, terms segmentcompare.Terms) (graphalign.Canon, error) {
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
		ix, err := loadShortlist(g, prof, cfg)
		if err != nil {
			return graphalign.Canon{}, err
		}
		shortlisted, err := seedHandles(g, prof, ix, todo)
		if err != nil {
			return graphalign.Canon{}, err
		}
		for i, s := range todo {
			for _, h := range graphalign.Candidates(s.Kind, s.Values, shortlisted[i], layer.Metas, cfg) {
				candidates[string(s.ID)] = append(candidates[string(s.ID)], h.ID)
				if err := add(h.ID); err != nil {
					return graphalign.Canon{}, err
				}
			}
		}
	}

	sort.SliceStable(seeds, func(i, j int) bool {
		return handles[string(seeds[i])].Ref < handles[string(seeds[j])].Ref
	})

	var edges []graphalign.CanonEdge
	edgeSeen := map[seenEdge]bool{}
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
				return graphalign.Canon{}, err
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
			return graphalign.Canon{}, err
		}
		for _, item := range ready {
			id, n := item.id, item.n
			selfKind := typeKey(types, n.SubjectTypeID)
			for _, l := range n.Links {
				neighbor, err := g.Node(l.Neighbor)
				if err != nil {
					return graphalign.Canon{}, err
				}
				if neighbor == nil || neighbor.Merged {
					continue
				}
				nKind := typeKey(types, neighbor.SubjectTypeID)
				if !vocabkey.IsPrimaryKind(nKind) {
					continue
				}
				if from, to, keyFrom, keyTo, ok := canonKeys(id, selfKind, n, l, neighbor, props, terms); ok {
					if ek := edgeSeenKey(from, to, keyFrom, keyTo); !edgeSeen[ek] {
						edgeSeen[ek] = true
						edges = append(edges, graphalign.CanonEdge{
							From: append([]byte(nil), from...), To: append([]byte(nil), to...),
							KeyFrom: keyFrom, KeyTo: keyTo,
						})
					}
				}
				if _, ok := handles[string(neighbor.ID)]; !ok {
					handles[string(neighbor.ID)] = neighbor
					kinds[string(neighbor.ID)] = nKind
					queue = append(queue, neighbor.ID)
				}
			}
		}
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
	// The event type comes from whichever end is an event; linkOf alone
	// decides whether the key keeps it.
	var eventType string
	if first, has := eventEndpoint(b); has {
		// The first endpoint is the edge's from end; self is that end when
		// the link was filed from it.
		if first == l.FromEnd {
			eventType = keptTerm(n, props, neighborTypeProperty[vocabkey.Event])
		} else {
			eventType = keptTerm(neighbor, props, neighborTypeProperty[vocabkey.Event])
		}
	}
	keyFrom = segmentcompare.Key(linkOf(b, l.TermKey, eventType, true), terms)
	keyTo = segmentcompare.Key(linkOf(b, l.TermKey, eventType, false), terms)
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
