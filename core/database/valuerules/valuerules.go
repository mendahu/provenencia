// Package valuerules builds a catalog's value-compare Rules: the registry's
// value section with the catalog's default name pattern and alias tables in
// it. Every caller that compares values against a catalog gets its rules here,
// so a new input to Rules is loaded in one place.
package valuerules

import (
	"database/sql"

	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/namealiases"
	"github.com/mendahu/provenencia/core/database/nameformats"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/valuecompare"
)

// Querier is *sql.DB or *sql.Tx.
type Querier interface {
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

// Load returns the catalog's rules and the name pattern key its name values
// carry. With no stored default name format, Names.Patterns stays empty and the
// key is western. g may be nil: the alias tables are then read from q.
func Load(g *graphcache.Graph, q Querier) (rules valuecompare.Rules, namePattern string, err error) {
	loaded, err := nameformats.LoadFor(g, q)
	if err != nil {
		return valuecompare.Rules{}, "", err
	}
	rules, err = namealiases.Rules(g, q, loaded.Rules(matchregistry.Default().Value))
	if err != nil {
		return valuecompare.Rules{}, "", err
	}
	return rules, loaded.Key, nil
}

// Product returns the rules the reconciler cache reads: the shipped registry
// and the product alias seed, never a catalog's own rows or name formats. The
// cache then changes only when autoreconciler.CacheVersion does.
func Product() valuecompare.Rules {
	rules := matchregistry.Default().Value
	rules.Names.Aliases = namealiases.ProductIndex()
	return rules
}
