package promotealign

import (
	"bytes"
	"fmt"
	"maps"
	"reflect"
	"sort"

	"github.com/mendahu/provenencia/core/database/canonlinks"
	"github.com/mendahu/provenencia/core/database/evidencelayer"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/graphfill"
	"github.com/mendahu/provenencia/core/database/nameformats"
	"github.com/mendahu/provenencia/core/database/valuerules"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/kinship"
	"github.com/mendahu/provenencia/core/segmentcompare"
	"github.com/mendahu/provenencia/core/shortlist"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// candidatesMemo is the candidates (handle ids, in order) of each Subject the
// graph's Evidence graph has had ranked, and the generation of the match index
// they were ranked against. Candidates depend on the Subject's values and the
// vocabulary and aliases (the entry is derived from the vocabulary and the name
// format), the handles' profiles and the shortlist index (dropped together on
// every handle commit, which gives the next index a new generation). A stored
// memo is never changed.
type candidatesMemo struct {
	gen       uint64
	bySubject map[string][][]byte
}

// candidates holds each Evidence graph's candidatesMemo, keyed by graph id. It
// is derived from the vocabulary and the name format, and a commit that moves
// the graph's rows (Set.EvidenceGraphs) drops that graph's. A held memo is
// used only while its generation is the stored shortlist index's.
var candidates graphcache.Keyed[candidatesMemo]

func init() {
	candidates = graphcache.RegisterKeyed(graphcache.KeyedEntry[candidatesMemo]{
		Name:        "promotealign.candidates",
		DerivedFrom: []graphcache.Source{graphcache.Vocabulary{}, nameformats.Default},
		Apply:       graphcache.DropNamedGraphs[candidatesMemo],
		Check:       checkCandidates,
		Cap:         graphfill.MaxEvidenceGraphs,
	})
}

// loadCanon reads the open catalog's graph: fixed anchors, the candidates of
// each unfixed Subject (loadCandidatesMemo), then every link those handles
// have. Link order is association ref, then the neighbor's ref, which is the
// order the walk used to break ties. Every handle's profile is built once:
// the shortlisted ones together, the ones the walk reaches after it. Recorded
// links come from the walk and implied ones from canonlinks.Reader.For, which
// answers from the graph cache's results for each walked person; both go
// through one kinship.LinkSet, so a link read from either end or recorded
// both ways merges, and the implied links follow the recorded ones in walk
// order.
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

	reader, err := canonlinks.NewReader(g)
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
// graphalign.Candidates over the handles the shortlist index offers it, held in
// the candidates entry until the index changes. It reads the stored memo for
// the current index generation, ranks only the Subjects it lacks, and stores a
// new map, never changing the stored one. A stale generation is a miss.
func loadCandidatesMemo(g *graphcache.Graph, graphID []byte, prof *profiler, layer graphalign.Layer, todo []graphalign.Subject, cfg graphalign.Config) (map[string][][]byte, error) {
	ix, gen, err := loadShortlist(g, prof, cfg)
	if err != nil {
		return nil, err
	}
	var known map[string][][]byte
	if m, ok := candidates.Peek(g, graphID); ok && m.gen == gen {
		known = m.bySubject
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
	ranked, err := rankCandidates(g, prof, ix, layer, missing, cfg)
	if err != nil {
		return nil, err
	}
	stored := make(map[string][][]byte, len(known)+len(missing))
	maps.Copy(stored, known)
	maps.Copy(stored, ranked)
	candidates.Set(g, graphID, candidatesMemo{gen: gen, bySubject: stored})
	return stored, nil
}

// rankCandidates is the ranked candidate ids of each Subject: the handles the
// shortlist index offers it, profiled once together, ordered by
// graphalign.Candidates. Every Subject has an entry, an empty list for none.
func rankCandidates(g *graphcache.Graph, prof *profiler, ix *shortlist.Index, layer graphalign.Layer, subjects []graphalign.Subject, cfg graphalign.Config) (map[string][][]byte, error) {
	shortlisted, err := seedHandles(g, prof, ix, subjects)
	if err != nil {
		return nil, err
	}
	out := make(map[string][][]byte, len(subjects))
	for i, s := range subjects {
		ranked := [][]byte{}
		for _, h := range graphalign.Candidates(s.Kind, s.Values, shortlisted[i], layer.Metas, cfg) {
			ranked = append(ranked, h.ID)
		}
		out[string(s.ID)] = ranked
	}
	return out, nil
}

// checkCandidates ranks the held graph's Subjects again from fresh inputs,
// storing nothing, and compares. A memo ranked against an index that is no
// longer the stored one is never used, so it is not compared.
func checkCandidates(g *graphcache.Graph, graphID []byte, stored candidatesMemo) error {
	m := matchIndexOf(g)
	if m.shortlist == nil || stored.gen != m.gen {
		return nil
	}
	eg, err := graphfill.LoadEvidenceGraph(g, graphID)
	if err != nil {
		return err
	}
	if !eg.Found {
		return fmt.Errorf("promotealign: candidates are held for %x, which is no Evidence graph", graphID)
	}
	cfg, namePattern, prof, err := shortlistInputs(g, true)
	if err != nil {
		return err
	}
	ev, err := evidencelayer.Build(g, eg, namePattern)
	if err != nil {
		return err
	}
	metas, err := valuerules.Metas(g)
	if err != nil {
		return err
	}
	layer := graphalign.Layer{Subjects: ev.Subjects, Bridges: ev.Links, Metas: metas}
	var subjects []graphalign.Subject
	for id := range stored.bySubject {
		s, ok := ev.Subject([]byte(id))
		if !ok {
			return fmt.Errorf("promotealign: candidates are held for Subject %x, which is no primary Subject of %x", id, graphID)
		}
		subjects = append(subjects, s)
	}
	fresh, err := rankCandidates(g, prof, m.shortlist, layer, subjects, cfg)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(fresh, stored.bySubject) {
		return fmt.Errorf("promotealign: the stored candidates of %x do not match a fresh ranking", graphID)
	}
	return nil
}
