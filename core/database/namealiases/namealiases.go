// Package namealiases stores abbreviation and nickname tables. The product
// lists are in registry.go. Install writes them once when a catalog is
// created. AddProductNicknames copies the current product nickname sets into
// a catalog that already exists. Scoring reads these rows and no other list.
// An open project keeps that read on the catalog graph until the project
// closes; Invalidate reloads it.
package namealiases

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/match"
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
			for _, token := range set {
				if _, err := tx.Exec(`INSERT INTO name_alias_members (table_id, set_idx, token)
					VALUES (?, ?, ?)`,
					stored, i, token); err != nil {
					return err
				}
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

// Load reads every alias table. No rows returns a nil slice.
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
		max := 0
		for idx := range byIdx {
			if idx > max {
				max = idx
			}
		}
		tables[i].Sets = make([][]string, max+1)
		for idx, tokens := range byIdx {
			tables[i].Sets[idx] = tokens
		}
	}
	return tables, nil
}

// Warm loads every alias table onto the catalog graph. catalogsession calls
// this when a project opens. A later Warm replaces the tables. Closing the
// catalog drops them with the graph.
func Warm(c *database.Catalog) error {
	db, err := c.DB()
	if err != nil {
		return err
	}
	tables, err := Load(db)
	if err != nil {
		return err
	}
	g := c.Graph()
	if g == nil {
		return database.ErrClosed
	}
	g.SetAliasTables(append([]valuecompare.AliasTable(nil), tables...))
	return nil
}

// Invalidate reloads the catalog's alias tables. A future editor calls this
// after inserting, changing, or deleting a table.
func Invalidate(c *database.Catalog) error {
	return Warm(c)
}

// Comparer is the name comparer catalog scoring uses. A warmed graph supplies
// its tables, including none. A nil or cold graph is read from q. The rows
// replace whatever list the base comparer had.
func Comparer(g *graphcache.Graph, q Querier, base match.NameComparer) (match.NameComparer, error) {
	if raw, ok := g.AliasTables(); ok {
		tables, ok := raw.([]valuecompare.AliasTable)
		if !ok {
			return match.NameComparer{}, fmt.Errorf("namealiases: alias cache has %T", raw)
		}
		return withAliases(base, tables), nil
	}
	tables, err := Load(q)
	if err != nil {
		return match.NameComparer{}, err
	}
	return withAliases(base, tables), nil
}

func withAliases(base match.NameComparer, tables []valuecompare.AliasTable) match.NameComparer {
	copied := append([]valuecompare.AliasTable(nil), tables...)
	base.Words.Aliases = &copied
	return base
}
