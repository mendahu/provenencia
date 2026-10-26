package promotealign

import (
	"bytes"
	"maps"
	"sort"

	"github.com/mendahu/provenencia/core/database/canonlinks"
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
// shortlisted ones together, the ones the walk reaches after it. Recorded
// links come from the walk and implied ones from canonlinks.Reader.For, which
// answers from the graph cache's results for each walked person; both go
// through one kinship.LinkSet, so a link read from either end or recorded both ways
// merges, and the implied links follow the recorded ones in walk order.
func loadCanon(g *graphcache.Graph, graphID []byte, prof *profiler, layer graphalign.Layer, fixed []graphalign.Fixed, cfg graphalign.Config, terms segmentcompare.Terms) (graphalign.Canon, error) {
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
		kind := string(prof.kind(n))
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

	reader, err := canonlinks.NewReader(g, matchregistry.Default().Kinship)
	if err != nil {
		return graphalign.Canon{}, err
	}
	set := kinship.NewLinkSet()
	var walkedPeople [][]byte
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
			if kinds[string(id)] == vocabkey.Person {
				walkedPeople = append(walkedPeople, id)
			}
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
			for _, l := range n.Links {
				neighbor, err := g.Node(l.Neighbor)
				if err != nil {
					return graphalign.Canon{}, err
				}
				if neighbor == nil || neighbor.Merged {
					continue
				}
				nKind := string(prof.kind(neighbor))
				if !vocabkey.IsPrimaryKind(nKind) {
					continue
				}
				if from, to, keyFrom, keyTo, ok := canonlinks.EdgeOf(id, n, l, neighbor, prof.eventTypeProp, terms); ok {
					set.Add(kinship.Link{Edge: kinship.Edge{A: from, B: to, KeyA: keyFrom, KeyB: keyTo}})
				}
				if _, ok := handles[string(neighbor.ID)]; !ok {
					handles[string(neighbor.ID)] = neighbor
					kinds[string(neighbor.ID)] = nKind
					queue = append(queue, neighbor.ID)
				}
			}
		}
	}

	// The implied links are collected before any edge is built. Every link is
	// walked from both ends and each end's call returns it, so only the call
	// of its A end adds it.
	var implied []kinship.Implied
	for _, id := range walkedPeople {
		links, err := reader.For(id)
		if err != nil {
			return graphalign.Canon{}, err
		}
		for _, imp := range links {
			if !bytes.Equal(imp.A, id) {
				continue
			}
			_, aOK := handles[string(imp.A)]
			_, bOK := handles[string(imp.B)]
			if aOK && bOK {
				implied = append(implied, imp)
			}
		}
	}
	// Walked links first, in walk order. An implied link that shares ends and
	// keys with a walked or earlier implied one merges into it; the rest follow
	// in order. The implied links are the cache's, so an edge copies what it keeps.
	links := set.Links()
	edges := make([]graphalign.CanonEdge, len(links), len(links)+len(implied))
	for i, l := range links {
		edges[i] = graphalign.CanonEdge{
			From: l.A, To: l.B, KeyFrom: l.KeyA, KeyTo: l.KeyB, Recording: l.Recording,
		}
	}
	for i, at := range set.PlaceImplied(implied) {
		imp := implied[i]
		rec := imp.Recording()
		if at < len(edges) {
			edges[at].Recording.Merge(rec)
			continue
		}
		edges = append(edges, graphalign.CanonEdge{
			From: bytes.Clone(imp.A), To: bytes.Clone(imp.B), KeyFrom: imp.KeyA, KeyTo: imp.KeyB,
			Recording: rec.Clone(),
		})
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
