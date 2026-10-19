package promotealign

import (
	"slices"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/valuerules"
	"github.com/mendahu/provenencia/core/graphalign"
)

// LayerBridges exposes the bridges the loader reads for one Source.
func LayerBridges(c *database.Catalog, sourceID []byte) ([]graphalign.Bridge, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	rules, namePattern, err := valuerules.Load(c.Graph(), db)
	if err != nil {
		return nil, err
	}
	cfg := graphalign.DefaultConfig()
	cfg.Value = rules
	terms, err := loadTerms(c.Graph())
	if err != nil {
		return nil, err
	}
	layer, _, err := loadLayer(c.Graph(), db, sourceID, cfg, namePattern, terms)
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
