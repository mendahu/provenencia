// Package graphcache is the in-memory canonical graph for one open catalog.
// A read fills it. A commit replaces or inserts only what a read has already
// filled, except an association commit, which loads that association's new
// endpoints on purpose. Unused, the maps stay empty.
package graphcache

import (
	"database/sql"

	"github.com/mendahu/provenencia/core/database/catalogmodel"
	"github.com/mendahu/provenencia/core/database/effects"
)

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

	// matchIndex is what Promote derives from the whole catalog.
	matchIndex   any
	matchIndexOK bool

	vocab   Vocabulary
	vocabOK bool

	// displays are finished header rows keyed by handle id. details are the
	// conclusion-detail value memos for the same ids; a drop clears both.
	// evidenceGraphs are loaded Evidence graphs keyed by graph id.
	// labels is the vocabulary map the rows resolve at read time. alt, when
	// set, is the querier a one-shot graph reads instead of db (a write transaction).
	displays       map[string]any
	details        map[string]any
	evidenceGraphs map[string]EvidenceGraph
	// memos are what a feature derives from one loaded Evidence graph, by
	// graph id and then its own key. A graph and its memos go together.
	memos    map[string]map[string]any
	labels   any
	labelsOK bool
	alt      Querier

	// activity is last activity per scope type, then per scope id, loaded by
	// the first Activity read of that type and patched by OnCommit.
	activity map[string]map[string]effects.Activity

	// aliasIndex is the built abbreviation and nickname index for this open
	// catalog. It is not a graph node: Drop leaves it. Closing the catalog
	// drops it with this graph. namealiases owns the concrete value.
	aliasIndex   any
	aliasIndexOK bool

	// nameFormat is the catalog's default name format, warmed the same way.
	// nameformats owns the concrete value.
	nameFormat   any
	nameFormatOK bool

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
	Rank        int
	Reason      string
	Text        string
	HasText     bool
	Integer     int64
	HasInteger  bool
	TermID      []byte
	TermKey     string
	EntityID    []byte
	Date        []byte
	Name        []byte
	ImageFileID []byte
	Support     int
	Against     int
}

// Member is one accepted or provisional identity claim. SourceID is the Source
// and GraphID the Evidence graph the claimed Subject is on. A Subject never
// changes Source or graph, and a commit that names the handle reloads the node.
type Member struct {
	SubjectID []byte
	SourceID  []byte
	GraphID   []byte
	Accepted  bool
}

// Link is one association seen from this node. AssociationRef is the sort
// key, then the neighbor's ref. FromEnd is true when this node is the
// association's first endpoint.
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

// dropDerived empties what features derive from the graph and its vocabulary
// and alias index: the Evidence graph memos and the match index.
func (g *Graph) dropDerived() {
	g.memos = map[string]map[string]any{}
	g.matchIndexOK = false
	g.matchIndex = nil
}

func (g *Graph) clear() {
	g.nodes = map[string]*Node{}
	g.holders = map[string][][]byte{}
	g.kinds = map[string][][]byte{}
	g.kindOn = map[string]bool{}
	g.kindSet = map[string]map[string]bool{}
	g.dropDerived()
	g.vocabOK = false
	g.vocab = Vocabulary{}
	g.displays = map[string]any{}
	g.details = map[string]any{}
	g.evidenceGraphs = map[string]EvidenceGraph{}
	g.labelsOK = false
	g.labels = nil
	g.activity = map[string]map[string]effects.Activity{}
	// The alias index and the name format stay. They follow the catalog connection,
	// not a rebuild.
}

// Drop empties nodes, holders, kind indexes, and the match index. The next read fills
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

// MatchIndex returns the match index: what Promote derives from the whole
// catalog, stored in this slot by a feature. A commit that names handles, a
// vocabulary commit, a structure drop and SetAliasIndex clear it: its keys
// read term keys and metas as well as handles.
func (g *Graph) MatchIndex() (any, bool) {
	if g == nil || !g.matchIndexOK {
		return nil, false
	}
	return g.matchIndex, true
}

// AliasIndex returns the abbreviation and nickname index namealiases.Warm
// built for this open catalog. The bool is false until Warm stores it,
// including when the catalog has no alias rows.
func (g *Graph) AliasIndex() (any, bool) {
	if g == nil || !g.aliasIndexOK {
		return nil, false
	}
	return g.aliasIndex, true
}

// SetAliasIndex stores the built alias index until the catalog closes. An
// index built from no tables is a warm catalog with no rows. Drop does not
// clear this. The match index and Evidence graph memos read the alias sets,
// so a new index drops them. A stored index is shared, so readers treat it as
// read-only.
func (g *Graph) SetAliasIndex(index any) {
	if g == nil {
		return
	}
	g.dropDerived()
	g.aliasIndex = index
	g.aliasIndexOK = true
}

// NameFormat returns the catalog's default name format warmed for this open
// catalog. The bool is false until nameformats.Warm stores it.
func (g *Graph) NameFormat() (any, bool) {
	if g == nil || !g.nameFormatOK {
		return nil, false
	}
	return g.nameFormat, true
}

// SetNameFormat stores the catalog's default name format until the catalog
// closes. Drop does not clear it. The match index and Evidence graph memos read
// the name format's part roles, so a new one drops them.
func (g *Graph) SetNameFormat(v any) {
	if g == nil {
		return
	}
	g.dropDerived()
	g.nameFormat = v
	g.nameFormatOK = true
}

// SetMatchIndex stores the match index until the next handle commit,
// vocabulary commit, alias index load or drop.
func (g *Graph) SetMatchIndex(index any) {
	if g == nil {
		return
	}
	g.matchIndex = index
	g.matchIndexOK = true
}

// ResetMatchIndexForTest drops the match index.
func (g *Graph) ResetMatchIndexForTest() {
	if g == nil {
		return
	}
	g.matchIndexOK = false
	g.matchIndex = nil
}

// SetQueryCounterForTest counts queries this graph issues. Tests only.
func (g *Graph) SetQueryCounterForTest(fn func()) {
	if g == nil {
		return
	}
	g.count = fn
}

// LoadedEvidenceGraphForTest reports whether an Evidence graph is loaded.
func (g *Graph) LoadedEvidenceGraphForTest(graphID []byte) bool {
	_, ok := g.evidenceGraphs[string(graphID)]
	return ok
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

// CorruptMemberSourceForTest replaces the Source of a loaded node's first
// member so Verify can fail a test.
func (g *Graph) CorruptMemberSourceForTest(id []byte) {
	n := g.nodes[string(id)]
	if n == nil || len(n.Members) == 0 {
		return
	}
	cp := *n
	cp.Members = append([]Member(nil), n.Members...)
	cp.Members[0].SourceID = make([]byte, 16)
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
