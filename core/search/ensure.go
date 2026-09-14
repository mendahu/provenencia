package search

import (
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/searchindex"
)

// EnsureIndex rebuilds the FTS projection when meta version lags or the
// docs/FTS row counts diverge. Call after Open/Create post-open ensures.
func EnsureIndex(c *database.Catalog) error {
	return searchindex.EnsureCatalog(c)
}
