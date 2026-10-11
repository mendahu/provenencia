package promotealign

import (
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/graphalign"
)

// LoadStatsForTest exposes the catalog's cached stats read.
func LoadStatsForTest(c *database.Catalog) (graphalign.Stats, error) {
	return loadStats(c)
}
