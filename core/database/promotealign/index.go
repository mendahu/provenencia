package promotealign

import (
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strings"
	"sync/atomic"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/namealiases"
	"github.com/mendahu/provenencia/core/database/nameformats"
	"github.com/mendahu/provenencia/core/database/valuerules"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/shortlist"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// matchIndex is what Promote derives from the whole catalog and keeps in the
// MatchIndex entry. A nil field is not built yet. profiles are the
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

// MatchIndex is the registry entry holding the match index. It is derived
// from the vocabulary, the alias index and the name format, and a handle commit
// drops it (S16-20 patches it instead). It has no Build: the profiler's shared
// profile map is stored with it by loadShortlist and keep.
var MatchIndex graphcache.Whole[matchIndex]

func init() {
	MatchIndex = graphcache.RegisterWhole(graphcache.WholeEntry[matchIndex]{
		Name:        "promotealign.matchindex",
		DerivedFrom: []graphcache.Source{graphcache.Vocabulary{}, namealiases.Index, nameformats.Default},
		Apply: func(g *graphcache.Graph, w graphcache.Whole[matchIndex], roots [][]byte, _ effects.Set) error {
			if len(roots) > 0 {
				w.Drop(g)
			}
			return nil
		},
		Check: checkMatchIndex,
	})
}

// matchIndexOf is the stored match index, or the zero value when none is.
func matchIndexOf(g *graphcache.Graph) matchIndex {
	m, _, _ := MatchIndex.Get(g)
	return m
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
	ix, err := buildShortlist(g, p, cfg)
	if err != nil {
		return nil, 0, err
	}
	m.shortlist = ix
	m.gen = indexGen.Add(1)
	MatchIndex.Set(g, m)
	p.keep()
	return ix, m.gen, nil
}

// buildShortlist makes a shortlist index from every unmerged handle of the
// kinds the rules key.
func buildShortlist(g *graphcache.Graph, p *profiler, cfg graphalign.Config) (*shortlist.Index, error) {
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
			return nil, err
		}
		for _, n := range nodes {
			if n.Merged {
				continue
			}
			v, err := p.keyValues(n)
			if err != nil {
				return nil, err
			}
			ix.Add(vocabkey.SubjectType(kind), n.ID, v)
		}
	}
	return ix, nil
}

// checkMatchIndex rebuilds the stored shortlist index and each stored profile
// with a profiler that shares nothing with the stored value, and compares.
func checkMatchIndex(g *graphcache.Graph, stored matchIndex) error {
	cfg, namePattern, p, err := shortlistInputs(g, true)
	if err != nil {
		return err
	}
	if stored.shortlist != nil {
		fresh, err := buildShortlist(g, p, cfg)
		if err != nil {
			return err
		}
		if !reflect.DeepEqual(stored.shortlist, fresh) {
			return errors.New("promotealign: the stored shortlist index does not match a fresh build")
		}
	}
	if stored.profilesPattern != namePattern {
		return nil
	}
	for id, want := range stored.profiles {
		n, err := g.Node([]byte(id))
		if err != nil {
			return err
		}
		if n == nil || n.Merged {
			return fmt.Errorf("promotealign: a stored profile is for %x, which is no unmerged handle", id)
		}
		if err := p.profile([]*graphcache.Node{n}); err != nil {
			return err
		}
		if !reflect.DeepEqual(p.done[id], want) {
			return fmt.Errorf("promotealign: the stored profile of %x does not match a fresh build", id)
		}
	}
	return nil
}

// shortlistInputs is the alignment config, the name pattern and a profiler for
// the shortlist index. The profiler shares the stored index's profiles, or, with
// fresh, nothing: a check compares against a build that shares nothing.
func shortlistInputs(g *graphcache.Graph, fresh bool) (graphalign.Config, string, *profiler, error) {
	cfg, namePattern, err := loadConfig(g)
	if err != nil {
		return graphalign.Config{}, "", nil, err
	}
	metas, err := valuerules.Metas(g)
	if err != nil {
		return graphalign.Config{}, "", nil, err
	}
	var p *profiler
	if fresh {
		p, err = newProfilerWith(g, metas, cfg.Subject, namePattern, map[string]profile{})
	} else {
		p, err = newProfiler(g, metas, cfg.Subject, namePattern)
	}
	return cfg, namePattern, p, err
}

// Warm loads the shortlist kinds onto the catalog graph and builds the
// shortlist index, so the first proposal after a project opens ranks without
// either. A handle commit drops the index again; the next proposal rebuilds it.
func Warm(c *database.Catalog) error {
	if _, err := c.DB(); err != nil {
		return err
	}
	g := c.Graph()
	cfg, _, prof, err := shortlistInputs(g, false)
	if err != nil {
		return err
	}
	_, _, err = loadShortlist(g, prof, cfg)
	return err
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
