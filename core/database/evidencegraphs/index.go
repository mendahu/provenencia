package evidencegraphs

import (
	"errors"
	"reflect"
	"slices"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/graphcache"
)

// ErrIndexOnWrite is returned by IndexOf on a write's graph: the index is
// built once per open catalog, never once per write.
var ErrIndexOnWrite = errors.New("evidencegraphs: the graph index is not read on a write's graph")

// GraphIndex holds every Evidence graph's row and counts, and each Source's
// graphs in list order. A reader treats it, and the ArtifactIDs of the graphs
// it returns, as read-only.
type GraphIndex struct {
	graphs   map[string]Graph
	progress map[string]Progress
	// bySource maps a Source id to its graph ids in list order (name,
	// case-insensitive, then ref, then id).
	bySource map[string][][]byte
}

// Graph returns the graph with this id.
func (x *GraphIndex) Graph(id []byte) (Graph, bool) {
	g, ok := x.graphs[string(id)]
	return g, ok
}

// OfSource returns a Source's graphs in list order; nil for none.
func (x *GraphIndex) OfSource(sourceID []byte) []Graph {
	return x.of(string(sourceID))
}

func (x *GraphIndex) of(sourceID string) []Graph {
	ids := x.bySource[sourceID]
	if len(ids) == 0 {
		return nil
	}
	out := make([]Graph, len(ids))
	for i, id := range ids {
		out[i] = x.graphs[string(id)]
	}
	return out
}

// All returns every graph: Sources by id ascending, each Source's graphs in
// list order.
func (x *GraphIndex) All() []Graph {
	sourceIDs := make([]string, 0, len(x.bySource))
	for sid := range x.bySource {
		sourceIDs = append(sourceIDs, sid)
	}
	slices.Sort(sourceIDs)
	out := make([]Graph, 0, len(x.graphs))
	for _, sid := range sourceIDs {
		out = append(out, x.of(sid)...)
	}
	return out
}

// Progress returns a graph's counts, zeros for a graph with none.
func (x *GraphIndex) Progress(id []byte) Progress {
	if p, ok := x.progress[string(id)]; ok {
		return p
	}
	return Progress{GraphID: append([]byte(nil), id...)}
}

// AllProgress returns the counts of every graph that has any, in All order.
func (x *GraphIndex) AllProgress() []Progress {
	out := make([]Progress, 0, len(x.progress))
	for _, g := range x.All() {
		if p, ok := x.progress[string(g.ID)]; ok {
			out = append(out, p)
		}
	}
	return out
}

// Index is the Evidence-graph entry: every graph with its linked Artifacts
// and counts, built on the first read of an open catalog and patched from the
// reads of effects.Set: evidence_graphs re-lists the named Sources and
// evidence_graph_progress recounts the named graphs.
var Index graphcache.Whole[*GraphIndex]

func init() {
	Index = graphcache.RegisterWhole(graphcache.WholeEntry[*GraphIndex]{
		Name:     "evidencegraphs.index",
		Lifetime: graphcache.LifeGraph,
		Build: func(g *graphcache.Graph) (*GraphIndex, error) {
			return buildIndex(g)
		},
		Apply: func(g *graphcache.Graph, w graphcache.Whole[*GraphIndex], _ [][]byte, set effects.Set) error {
			var sourceIDs, graphIDs [][]byte
			for _, r := range set.Reads {
				switch r.Read {
				case effects.ReadEvidenceGraphs:
					sourceIDs = append(sourceIDs, r.IDs...)
				case effects.ReadEvidenceGraphProgress:
					graphIDs = append(graphIDs, r.IDs...)
				}
			}
			if len(sourceIDs) == 0 && len(graphIDs) == 0 {
				return nil
			}
			idx, ok, err := w.Get(g)
			if err != nil || !ok {
				return err
			}
			// Everything is read before anything changes.
			fresh := make([][]Graph, len(sourceIDs))
			for i, sid := range sourceIDs {
				if fresh[i], err = list(g, sid); err != nil {
					return err
				}
			}
			// A nil id list would count every graph.
			var counts map[string]Progress
			if len(graphIDs) > 0 {
				if counts, err = loadProgress(g, graphIDs); err != nil {
					return err
				}
			}
			// Every named Source leaves before any returns, so a graph that
			// moved between two named Sources keeps its new place.
			var removed []string
			for _, sid := range sourceIDs {
				for _, id := range idx.bySource[string(sid)] {
					delete(idx.graphs, string(id))
					removed = append(removed, string(id))
				}
				delete(idx.bySource, string(sid))
			}
			for i, sid := range sourceIDs {
				if len(fresh[i]) == 0 {
					continue
				}
				ids := make([][]byte, len(fresh[i]))
				for j, gr := range fresh[i] {
					idx.graphs[string(gr.ID)] = gr
					ids[j] = gr.ID
				}
				idx.bySource[string(sid)] = ids
			}
			for _, id := range removed {
				if _, ok := idx.graphs[id]; !ok {
					delete(idx.progress, id)
				}
			}
			for _, id := range graphIDs {
				delete(idx.progress, string(id))
			}
			for id, p := range counts {
				if _, ok := idx.graphs[id]; ok {
					idx.progress[id] = p
				}
			}
			w.Set(g, idx)
			return nil
		},
		Check: func(g *graphcache.Graph, stored *GraphIndex) error {
			want, err := buildIndex(g)
			if err != nil {
				return err
			}
			if !reflect.DeepEqual(stored.graphs, want.graphs) ||
				!reflect.DeepEqual(stored.progress, want.progress) ||
				!reflect.DeepEqual(stored.bySource, want.bySource) {
				return errors.New("evidencegraphs: the stored graph index does not match the catalog tables")
			}
			return nil
		},
	})
}

// IndexOf returns the graph index of an open catalog's graph, building it on
// the first read. A write's graph is ErrIndexOnWrite; a nil graph is
// database.ErrClosed.
func IndexOf(g *graphcache.Graph) (*GraphIndex, error) {
	if g == nil {
		return nil, database.ErrClosed
	}
	idx, ok, err := Index.Get(g)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrIndexOnWrite
	}
	return idx, nil
}

func buildIndex(q database.Querier) (*GraphIndex, error) {
	graphs, err := list(q, nil)
	if err != nil {
		return nil, err
	}
	counts, err := loadProgress(q, nil)
	if err != nil {
		return nil, err
	}
	return newGraphIndex(graphs, counts), nil
}

// newGraphIndex groups graphs, already in list order, by Source.
func newGraphIndex(graphs []Graph, counts map[string]Progress) *GraphIndex {
	x := &GraphIndex{
		graphs:   make(map[string]Graph, len(graphs)),
		progress: counts,
		bySource: map[string][][]byte{},
	}
	for _, gr := range graphs {
		x.graphs[string(gr.ID)] = gr
		x.bySource[string(gr.SourceID)] = append(x.bySource[string(gr.SourceID)], gr.ID)
	}
	return x
}
