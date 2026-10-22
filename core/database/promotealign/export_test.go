package promotealign

import (
	"slices"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/evidencegraphs"
	"github.com/mendahu/provenencia/core/database/valuerules"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/subjectcompare"
)

// LayerBridges exposes the bridges the loader reads for one Source's one
// Evidence graph.
func LayerBridges(c *database.Catalog, sourceID []byte) ([]graphalign.Bridge, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	rules, namePattern, err := valuerules.Load(c.Graph(), db)
	if err != nil {
		return nil, err
	}
	cfg := matchregistry.Default().AlignConfig()
	cfg.Value = rules
	terms, err := loadTerms(c.Graph())
	if err != nil {
		return nil, err
	}
	graphID, err := evidencegraphs.SoleGraph(c, sourceID)
	if err != nil {
		return nil, err
	}
	layer, _, err := loadLayer(c.Graph(), graphID, cfg, namePattern, terms)
	if err != nil {
		return nil, err
	}
	return layer.Bridges, nil
}

// ProfiledRefsForTest proposes for the Source with no fixed rows and returns,
// sorted, the Ref of every handle whose full profile was built.
func ProfiledRefsForTest(c *database.Catalog, sourceID []byte) ([]string, error) {
	_, _, prof, err := propose(c, sourceID, nil)
	if err != nil || prof == nil {
		return nil, err
	}
	var refs []string
	for id := range prof.done {
		n, err := c.Graph().Node([]byte(id))
		if err != nil {
			return nil, err
		}
		if n != nil {
			refs = append(refs, n.Ref)
		}
	}
	slices.Sort(refs)
	return refs, nil
}

// BridgedOnly exposes the filter that limits a neighbor entity's read members.
var BridgedOnly = bridgedOnly

// MemoSnapshot is what the graph holds as promotealign's Evidence graph memos.
type MemoSnapshot struct {
	Layer     graphalign.Layer
	LayerKind map[string]string // primary Subject id -> kind
	LayerOK   bool
	Members   map[string]subjectcompare.Values
}

// MemoSnapshotForTest reads the memoized layer of graphID and the memoized
// member values of memberGraphID, as stored. Nothing is built.
func MemoSnapshotForTest(c *database.Catalog, graphID, memberGraphID []byte) MemoSnapshot {
	db, err := c.DB()
	if err != nil {
		return MemoSnapshot{}
	}
	_, namePattern, err := valuerules.Load(c.Graph(), db)
	if err != nil {
		return MemoSnapshot{}
	}
	var out MemoSnapshot
	if v, ok := c.Graph().EvidenceGraphMemo(graphID, layerMemoKey+namePattern); ok {
		m := v.(layerMemo)
		out.Layer, out.LayerOK = m.layer, true
		out.LayerKind = map[string]string{}
		for id, p := range m.primary {
			out.LayerKind[id] = p.kind
		}
	}
	if v, ok := c.Graph().EvidenceGraphMemo(memberGraphID, membersMemoKey+namePattern); ok {
		out.Members = v.(map[string]subjectcompare.Values)
	}
	return out
}

// ProfilesKeptForTest is how many handle profiles the graph's match index
// holds, and false when the graph has no match index.
func ProfilesKeptForTest(c *database.Catalog) (int, bool) {
	if _, ok := c.Graph().MatchIndex(); !ok {
		return 0, false
	}
	return len(matchIndexOf(c.Graph()).profiles), true
}
