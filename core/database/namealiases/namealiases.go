// Package namealiases stores abbreviation and nickname tables. The product
// lists are in registry.go. Install writes them once when a catalog is
// created. AddProductNicknames copies the current product nickname sets into
// a catalog that already exists. Matching reads these rows and no other list;
// the reconciler cache reads the product seed through valuerules.Product.
// An open project keeps the alias index built from those rows on the catalog
// graph until the project closes; Invalidate drops it and the next read rebuilds it.
package namealiases

import (
	"database/sql"
	"errors"
	"reflect"
	"sort"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/valuecompare"
)

// Querier is *sql.DB or *sql.Tx.
type Querier interface {
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

// Install writes the registered alias tables into a new catalog. Call only
// from catalog create. A table that already exists is left alone, including
// a deleted token. ON CONFLICT DO NOTHING.
func Install(c *database.Catalog) error {
	db, err := c.DB()
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	for _, table := range AliasTables() {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		res, err := tx.Exec(`INSERT INTO name_alias_tables (id, key, origin, label, kind)
			VALUES (?, ?, ?, ?, ?)
			ON CONFLICT(key, origin) DO NOTHING`,
			id[:], table.Key, table.Origin, table.Label, string(table.Kind))
		if err != nil {
			return err
		}
		n, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if n == 0 {
			continue
		}
		var stored []byte
		if err := tx.QueryRow(`SELECT id FROM name_alias_tables WHERE key = ? AND origin = ?`,
			table.Key, table.Origin).Scan(&stored); err != nil {
			return err
		}
		for i, set := range table.Sets {
			if _, err := insertSet(tx, stored, i, set); err != nil {
				return err
			}
		}
		for _, role := range table.Roles {
			if _, err := tx.Exec(`INSERT INTO name_alias_roles (table_id, role) VALUES (?, ?)`,
				stored, string(role)); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

// Load reads every alias table. No rows returns a nil slice. A table's Sets
// are the stored sets in ascending set_idx order; a position is not a set_idx,
// so a stored index far above the rest costs nothing.
func Load(q Querier) ([]valuecompare.AliasTable, error) {
	rows, err := q.Query(`SELECT id, key, origin, label, kind FROM name_alias_tables ORDER BY key, origin`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []valuecompare.AliasTable
	ids := map[string]int{}
	for rows.Next() {
		var id []byte
		var table valuecompare.AliasTable
		var kind string
		if err := rows.Scan(&id, &table.Key, &table.Origin, &table.Label, &kind); err != nil {
			return nil, err
		}
		table.Kind = valuecompare.AliasKind(kind)
		ids[string(id)] = len(tables)
		tables = append(tables, table)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(tables) == 0 {
		return nil, nil
	}

	members, err := q.Query(`SELECT table_id, set_idx, token FROM name_alias_members ORDER BY table_id, set_idx, token`)
	if err != nil {
		return nil, err
	}
	defer members.Close()
	sets := map[int]map[int][]string{}
	for members.Next() {
		var id []byte
		var idx int
		var token string
		if err := members.Scan(&id, &idx, &token); err != nil {
			return nil, err
		}
		at, ok := ids[string(id)]
		if !ok {
			continue
		}
		if sets[at] == nil {
			sets[at] = map[int][]string{}
		}
		sets[at][idx] = append(sets[at][idx], token)
	}
	if err := members.Err(); err != nil {
		return nil, err
	}

	roles, err := q.Query(`SELECT table_id, role FROM name_alias_roles ORDER BY table_id, role`)
	if err != nil {
		return nil, err
	}
	defer roles.Close()
	for roles.Next() {
		var id []byte
		var role string
		if err := roles.Scan(&id, &role); err != nil {
			return nil, err
		}
		at, ok := ids[string(id)]
		if !ok {
			continue
		}
		tables[at].Roles = append(tables[at].Roles, valuecompare.NameRole(role))
	}
	if err := roles.Err(); err != nil {
		return nil, err
	}

	for i := range tables {
		byIdx := sets[i]
		if len(byIdx) == 0 {
			continue
		}
		idxs := make([]int, 0, len(byIdx))
		for idx := range byIdx {
			idxs = append(idxs, idx)
		}
		sort.Ints(idxs)
		tables[i].Sets = make([][]string, 0, len(idxs))
		for _, idx := range idxs {
			tables[i].Sets = append(tables[i].Sets, byIdx[idx])
		}
	}
	return tables, nil
}

// Index is the alias index built from every alias table, one per catalog
// connection. An open project builds it once and keeps it until the project
// closes; a write's graph builds it once from its transaction. A stored index
// is shared, so readers treat it as read-only.
var Index graphcache.Whole[valuecompare.AliasIndex]

func init() {
	Index = graphcache.RegisterWhole(graphcache.WholeEntry[valuecompare.AliasIndex]{
		Name:     "namealiases.index",
		Lifetime: graphcache.LifeCatalog,
		Build: func(g *graphcache.Graph) (valuecompare.AliasIndex, error) {
			tables, err := Load(g)
			if err != nil {
				return valuecompare.AliasIndex{}, err
			}
			return buildIndex(tables), nil
		},
		Check: func(g *graphcache.Graph, stored valuecompare.AliasIndex) error {
			tables, err := Load(g)
			if err != nil {
				return err
			}
			if !reflect.DeepEqual(stored, buildIndex(tables)) {
				return errors.New("namealiases: the stored alias index does not match the alias tables")
			}
			return nil
		},
	})
}

// Warm builds the alias index onto the catalog graph. catalogsession calls
// this when a project opens. Closing the catalog drops it with the graph.
func Warm(c *database.Catalog) error {
	if _, err := c.DB(); err != nil {
		return err
	}
	g := c.Graph()
	if g == nil {
		return database.ErrClosed
	}
	_, _, err := Index.Get(g)
	return err
}

// Invalidate drops the alias index; the next read rebuilds it. A future editor
// calls this after inserting, changing, or deleting a table.
func Invalidate(c *database.Catalog) error {
	if _, err := c.DB(); err != nil {
		return err
	}
	g := c.Graph()
	if g == nil {
		return database.ErrClosed
	}
	Index.Drop(g)
	return nil
}

// Rules is base with the catalog's alias index in Names.Aliases. The graph
// supplies its built index, which is shared and read-only, including an index
// of no tables. A nil graph is database.ErrClosed. The index replaces whatever
// base had.
func Rules(g *graphcache.Graph, base valuecompare.Rules) (valuecompare.Rules, error) {
	if g == nil {
		return valuecompare.Rules{}, database.ErrClosed
	}
	index, _, err := Index.Get(g)
	if err != nil {
		return valuecompare.Rules{}, err
	}
	base.Names.Aliases = index
	return base, nil
}

// ProductIndex is the alias index of the product seed, for the reconciler cache.
func ProductIndex() valuecompare.AliasIndex { return buildIndex(ProductTables()) }

// buildIndex keys the tables with valueWords, the word rules that already key
// every stored alias token, so storage, the warm index and the cold index use
// one rule.
func buildIndex(tables []valuecompare.AliasTable) valuecompare.AliasIndex {
	return valuecompare.NewAliasIndex(tables, valueWords)
}
