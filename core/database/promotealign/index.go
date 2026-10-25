package promotealign

import (
	"bytes"
	"cmp"
	"slices"
	"sort"
	"strings"
	"sync/atomic"

	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/shortlist"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// matchIndex is what Promote derives from the whole catalog and keeps in the
// graph's match index slot. A nil field is not built yet. profiles are the
// handle profiles built so far, by handle id, under namePattern: a profile
// reads the handle's values and the events it takes part in, so it lives as
// long as the index does, and every handle commit drops both.
type matchIndex struct {
	shortlist       *shortlist.Index
	gen             uint64
	profiles        map[string]profile
	profilesPattern string
}

// indexGen mints the generation of each shortlist index loadShortlist builds,
// so a memo ranked against one index can tell it is not the one now stored.
var indexGen atomic.Uint64

// matchIndexOf is the stored match index, or the zero value when none is.
func matchIndexOf(g *graphcache.Graph) matchIndex {
	if v, ok := g.MatchIndex(); ok {
		if m, ok := v.(matchIndex); ok {
			return m
		}
	}
	return matchIndex{}
}

// loadShortlist is the catalog's shortlist index: the stored one, or one built
// from every unmerged handle of the kinds the rules key, and stored. Keys need
// no life, so no life event is loaded or built. It is dropped by a handle
// commit, as any match index is; the graph's commit path serializes with it.
// The uint64 is the index's generation, new for every index built here.
func loadShortlist(g *graphcache.Graph, p *profiler, cfg graphalign.Config) (*shortlist.Index, uint64, error) {
	m := matchIndexOf(g)
	if m.shortlist != nil {
		return m.shortlist, m.gen, nil
	}
	ix := shortlist.NewIndex(cfg.Alignment.Shortlist, cfg.Subject, cfg.Value, p.metas)
	kinds := map[vocabkey.SubjectType]bool{}
	for kind := range cfg.Alignment.Shortlist.Words {
		kinds[kind] = true
	}
	for kind := range cfg.Alignment.Shortlist.Dated {
		kinds[kind] = true
	}
	sorted := make([]string, 0, len(kinds))
	for kind := range kinds {
		sorted = append(sorted, string(kind))
	}
	sort.Strings(sorted)
	for _, kind := range sorted {
		typeID := p.types[kind]
		if len(typeID) != 16 {
			continue
		}
		nodes, err := g.Kind(typeID)
		if err != nil {
			return nil, 0, err
		}
		for _, n := range nodes {
			if n.Merged {
				continue
			}
			v, err := p.keyValues(n)
			if err != nil {
				return nil, 0, err
			}
			ix.Add(vocabkey.SubjectType(kind), n.ID, v)
		}
	}
	m.shortlist = ix
	m.gen = indexGen.Add(1)
	g.SetMatchIndex(m)
	p.keep()
	return ix, m.gen, nil
}

// seedHandles loads and profiles the handles the index offers each Subject, in
// one node load and one profile batch however many Subjects there are. Entry i
// is subject i's handles, in Ref then ID order. An id with no node, or a merged
// node, is skipped.
func seedHandles(g *graphcache.Graph, p *profiler, ix *shortlist.Index, subjects []graphalign.Subject) ([][]graphalign.Handle, error) {
	found := make([][][]byte, len(subjects))
	var union [][]byte
	for i, s := range subjects {
		found[i] = ix.Lookup(vocabkey.SubjectType(s.Kind), s.Values)
		union = append(union, found[i]...)
	}
	if err := g.Nodes(union); err != nil {
		return nil, err
	}
	nodes := map[string]*graphcache.Node{}
	var present []*graphcache.Node
	for _, id := range union {
		if _, ok := nodes[string(id)]; ok {
			continue
		}
		n, err := g.Node(id)
		if err != nil {
			return nil, err
		}
		if n == nil || n.Merged {
			continue
		}
		nodes[string(id)] = n
		present = append(present, n)
	}
	if err := p.profile(present); err != nil {
		return nil, err
	}
	out := make([][]graphalign.Handle, len(subjects))
	for i, ids := range found {
		for _, id := range ids {
			if n, ok := nodes[string(id)]; ok {
				out[i] = append(out[i], graphalign.Handle{ID: n.ID, Ref: n.Ref, Kind: subjects[i].Kind, Values: p.values(n.ID)})
			}
		}
		slices.SortFunc(out[i], func(a, b graphalign.Handle) int {
			return cmp.Or(strings.Compare(a.Ref, b.Ref), bytes.Compare(a.ID, b.ID))
		})
	}
	return out, nil
}
