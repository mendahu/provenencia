package namealiases

import (
	"database/sql"
	"sync"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/match"
)

// cachedTables is one open catalog's alias tables. A present map entry is
// warm even when tables is nil (the catalog has no rows).
type cachedTables struct {
	tables []match.AliasTable
}

var (
	cacheMu sync.Mutex
	cache   = map[*sql.DB]cachedTables{}
)

// Warm loads every alias table and keeps it for this catalog's connection.
// catalogsession calls this when a project opens. A later Warm replaces the
// entry. Close does not drop it; Forget does.
func Warm(c *database.Catalog) error {
	db, err := c.DB()
	if err != nil {
		return err
	}
	tables, err := Load(db)
	if err != nil {
		return err
	}
	remember(db, tables)
	return nil
}

// Invalidate reloads the catalog's alias tables. A future editor calls this
// after inserting, changing, or deleting a table. It leaves the project warm.
func Invalidate(c *database.Catalog) error {
	return Warm(c)
}

// Forget drops the catalog's alias tables. catalogsession calls this when
// the project closes, before the connection is closed.
func Forget(c *database.Catalog) {
	if c == nil {
		return
	}
	db, err := c.DB()
	if err != nil {
		return
	}
	cacheMu.Lock()
	delete(cache, db)
	cacheMu.Unlock()
}

func remember(db *sql.DB, tables []match.AliasTable) {
	copied := append([]match.AliasTable(nil), tables...)
	cacheMu.Lock()
	cache[db] = cachedTables{tables: copied}
	cacheMu.Unlock()
}

func recall(db *sql.DB) ([]match.AliasTable, bool) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	entry, ok := cache[db]
	if !ok {
		return nil, false
	}
	return append([]match.AliasTable(nil), entry.tables...), true
}
