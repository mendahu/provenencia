package promotealign

import (
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
