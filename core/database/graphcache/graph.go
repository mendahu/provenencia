// Package graphcache is the in-memory canonical graph for one open catalog.
// A read fills it. A commit replaces or inserts only what a read has already
// filled, except an association commit, which loads that association's new
// endpoints on purpose. Unused, the maps stay empty.
package graphcache

import (
	"database/sql"

	"github.com/mendahu/provenencia/core/database/catalogmodel"
)

const batch = 500

// Graph is one catalog's canonical graph. Nodes are immutable: a commit
// replaces the pointer in the map. catalogsession.Do serializes the process,
// so Graph has no lock of its own.
type Graph struct {
	db *sql.DB

	nodes   map[string]*Node
	holders map[string][][]byte // association id -> loaded endpoint ids
	kinds   map[string][][]byte // subject type id -> unmerged ids, after Kind
	kindOn  map[string]bool
	kindSet map[string]map[string]bool

	stats    any
	statsRev int64
	statsOK  bool

	// displays are finished header rows keyed by handle id. details are the
	// conclusion-detail value memos for the same ids; a drop clears both.
	// sources are loaded Evidence graphs keyed by source id.
	// labels is the vocabulary map the rows resolve at read time. alt, when
	// set, is the querier a one-shot graph reads instead of db (a write transaction).
	displays map[string]any
	details  map[string]any
	sources  map[string]SourceGraph
	labels   any
	labelsOK bool
	alt      Querier

	entityTable string
	claimTable  string
	valueTable  string

	count func()
}

// Node is one canonical entity. Kind is not stored: callers look the type
// id up. Values keeps every auto-reconciler rank. Members are accepted and
// provisional subjects in claim id order. Argument and Label are the
// entity row's own text; a header uses them when it has no reconciled name.
type Node struct {
	ID            []byte
	Ref           string
	SubjectTypeID []byte
	Argument      string
	Label         string
	Merged        bool
	Values        map[string][]Value // property id -> ranks
	Links         []Link
	Members       []Member
}

// Value is one auto_reconciler_values row. Date and Name are the stored blobs.
type Value struct {
	Rank       int
	Reason     string
	Text       string
	HasText    bool
	Integer    int64
	HasInteger bool
	TermID     []byte
	TermKey    string
	EntityID   []byte
	Date       []byte
	Name       []byte
	Support    int
	Against    int
}

// Member is one accepted or provisional identity claim.
type Member struct {
	SubjectID []byte
	Accepted  bool
}

// Link is one association seen from this node. AssociationRef is the sort
// key, then the neighbor's ref. FromEnd is true when this node is the
// association's first endpoint. Directed and InverseKey are copied off the
// term row at load; the edge signature is built later, at the feature edge.
type Link struct {
	Neighbor                 []byte
	NeighborRef              string
	Association              []byte
	AssociationRef           string
	FromEnd                  bool
	EndpointPropertyID       []byte
	DisambiguationPropertyID []byte
	TermID                   []byte
	TermKey                  string
	InverseKey               string
	Directed                 bool
	NeighborTypeID           []byte
	BridgeTypeKey            string
	NeighborTypeKey          string
}

// Querier is the read surface Verify's truth check and tests use.
type Querier interface {
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

// New returns an empty graph bound to db. Open and Create attach it and Listen.
func New(db *sql.DB) *Graph {
	g := &Graph{db: db}
	g.entityTable = mustTable(catalogmodel.KindCanonicalEntity)
	g.claimTable = mustTable(catalogmodel.KindIdentityClaim)
	g.valueTable = mustTableName("auto_reconciler_values")
	g.clear()
	return g
}

func mustTable(kind catalogmodel.Kind) string {
	for _, t := range catalogmodel.Tables {
		if t.Kind == kind {
			return t.Name
		}
	}
	panic("graphcache: catalog model missing " + string(kind))
}

func mustTableName(name string) string {
	for _, t := range catalogmodel.Tables {
		if t.Name == name {
			return t.Name
		}
	}
	panic("graphcache: catalog model missing " + name)
}

func (g *Graph) clear() {
	g.nodes = map[string]*Node{}
	g.holders = map[string][][]byte{}
	g.kinds = map[string][][]byte{}
	g.kindOn = map[string]bool{}
	g.kindSet = map[string]map[string]bool{}
	g.statsOK = false
	g.stats = nil
	g.statsRev = 0
	g.displays = map[string]any{}
	g.details = map[string]any{}
	g.sources = map[string]SourceGraph{}
	g.labelsOK = false
	g.labels = nil
}

// Drop empties nodes, holders, kind indexes, and stats. The next read fills
// them. EnsureCatalog calls this after a rebuild, which does not go through Run.
func (g *Graph) Drop() {
	if g == nil {
		return
	}
	g.clear()
}

// Nodes loads ids that are not already cached, in batches.
func (g *Graph) Nodes(ids [][]byte) error {
	if g == nil {
		return nil
	}
	return g.loadIDs(ids, false)
}

// Node returns the cached node, loading it on first read. A missing entity
// is (nil, nil).
func (g *Graph) Node(id []byte) (*Node, error) {
	if g == nil || len(id) != 16 {
		return nil, nil
	}
	if n, ok := g.nodes[string(id)]; ok {
		return n, nil
	}
	if err := g.loadIDs([][]byte{id}, false); err != nil {
		return nil, err
	}
	return g.nodes[string(id)], nil
}

// Kind returns every unmerged node of a subject type. The first read loads
// them and stores the id list. Later commits insert and remove on that list.
func (g *Graph) Kind(typeID []byte) ([]*Node, error) {
	if g == nil || len(typeID) != 16 {
		return nil, nil
	}
	k := string(typeID)
	if !g.kindOn[k] {
		ids, err := g.unmergedIDs(typeID)
		if err != nil {
			return nil, err
		}
		g.kindOn[k] = true
		g.kinds[k] = ids
		g.kindSet[k] = map[string]bool{}
		for _, id := range ids {
			g.kindSet[k][string(id)] = true
		}
	}
	if err := g.loadIDs(g.kinds[k], false); err != nil {
		return nil, err
	}
	out := make([]*Node, 0, len(g.kinds[k]))
	for _, id := range g.kinds[k] {
		if n, ok := g.nodes[string(id)]; ok {
			out = append(out, n)
		}
	}
	return out, nil
}

// Stats returns the cached scan when rev is the revision it was built at.
// The value is the graphalign.Stats the caller stored. The graph does not
// import that package.
func (g *Graph) Stats(rev int64) (any, bool) {
	if g == nil || !g.statsOK || g.statsRev != rev {
		return nil, false
	}
	return g.stats, true
}

// SetStats stores a scan for rev.
func (g *Graph) SetStats(rev int64, stats any) {
	if g == nil {
		return
	}
	g.stats = stats
	g.statsRev = rev
	g.statsOK = true
}

// ResetStatsForTest drops the cached scan.
func (g *Graph) ResetStatsForTest() {
	if g == nil {
		return
	}
	g.statsOK = false
}

// SetQueryCounterForTest counts queries this graph issues. Tests only.
func (g *Graph) SetQueryCounterForTest(fn func()) {
	if g == nil {
		return
	}
	g.count = fn
}

// LoadedForTest reports whether id is in the node map.
func (g *Graph) LoadedForTest(id []byte) bool {
	if g == nil {
		return false
	}
	_, ok := g.nodes[string(id)]
	return ok
}

// CorruptRefForTest replaces a loaded node's ref so Verify can fail a test.
func (g *Graph) CorruptRefForTest(id []byte) {
	n := g.nodes[string(id)]
	if n == nil {
		return
	}
	cp := *n
	cp.Ref = n.Ref + "-corrupt"
	g.nodes[string(id)] = &cp
}

func (g *Graph) note() {
	if g.count != nil {
		g.count()
	}
}

// Over is a graph that reads q and stores nothing the caller keeps. A save
// composes rows from its open transaction this way, because the catalog
// graph is still on the previous commit and the catalog has one connection.
func Over(q Querier) *Graph {
	g := New(nil)
	g.alt = q
	return g
}

// Query runs SQL through this graph's connection, counting it when a test asked.
func (g *Graph) Query(query string, args ...any) (*sql.Rows, error) {
	return g.query(query, args...)
}

// QueryRow runs SQL through this graph's connection.
func (g *Graph) QueryRow(query string, args ...any) *sql.Row {
	return g.queryRow(query, args...)
}

func (g *Graph) query(query string, args ...any) (*sql.Rows, error) {
	g.note()
	if g.alt != nil {
		return g.alt.Query(query, args...)
	}
	return g.db.Query(query, args...)
}

func (g *Graph) queryRow(query string, args ...any) *sql.Row {
	g.note()
	if g.alt != nil {
		return g.alt.QueryRow(query, args...)
	}
	return g.db.QueryRow(query, args...)
}
