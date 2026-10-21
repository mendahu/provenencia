// Package canonicalgraph walks the graph of people, events, and places:
// nodes joined through bridges (a participation, a location).
//
// A hop is declared in core/hops. Walk follows it through the canonical
// graph (handles, through association handles whose endpoint values the
// auto-reconciler kept) in a fixed number of indexed queries. A Source's
// Evidence graph is walked in memory by graphcache.Graph.WalkSource.
package canonicalgraph

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/sqlbatch"
	"github.com/mendahu/provenencia/core/hops"
)

// Querier is *sql.Tx or *sql.DB.
type Querier interface {
	Query(query string, args ...any) (*sql.Rows, error)
}

// Edge is one association a hop crossed.
type Edge struct {
	From           []byte
	Association    []byte
	AssociationRef string
	To             []byte
	ToRef          string
}

// The association's from-endpoint is one of the given handles; its
// to-endpoint is an unmerged handle. Both are kept rank-1 values, as is the
// filter term when there is one. a reads the edge index
// (property_id, value_entity_id); b and f read the primary key.
const (
	sqlWalkSelect = `SELECT a.value_entity_id, a.entity_id, assoc.ref, b.value_entity_id, dst.ref
FROM auto_reconciler_values a
JOIN canonical_entities assoc ON assoc.id = a.entity_id AND assoc.merged_into_id IS NULL
	AND assoc.subject_type_id = (SELECT id FROM subject_types WHERE key = ? AND origin = ?)
JOIN auto_reconciler_values b ON b.entity_id = a.entity_id AND b.rank = 1 AND b.reason = 'kept'
	AND b.property_id = (SELECT id FROM properties WHERE key = ? AND origin = ?)
JOIN canonical_entities dst ON dst.id = b.value_entity_id AND dst.merged_into_id IS NULL`
	sqlWalkFilter = `
JOIN auto_reconciler_values f ON f.entity_id = a.entity_id AND f.rank = 1 AND f.reason = 'kept'
	AND f.property_id = (SELECT id FROM properties WHERE key = ? AND origin = ?)
JOIN property_terms ft ON ft.id = f.value_term_id AND ft.key = ? AND ft.origin = ?`
	sqlWalkWhere = `
WHERE a.rank = 1 AND a.reason = 'kept'
	AND a.property_id = (SELECT id FROM properties WHERE key = ? AND origin = ?)
	AND a.value_entity_id IN (`
)

// hopQuery is the canonical walk up to its IN list, and the arguments before
// the ids.
func hopQuery(h hops.Hop) (string, []any) {
	query := sqlWalkSelect
	args := []any{h.Bridge(), h.Origin(), h.To(), h.Origin()}
	if key, term, ok := h.Filter(); ok {
		query += sqlWalkFilter
		args = append(args, key, h.Origin(), term, h.Origin())
	}
	query += sqlWalkWhere
	args = append(args, h.From(), h.Origin())
	return query, args
}

// Walk follows h through the canonical graph from each handle, ordered by
// from-handle, then association ref, then to-handle ref.
func Walk(q Querier, h hops.Hop, from [][]byte) ([]Edge, error) {
	if h.Bridge() == "" {
		return nil, fmt.Errorf("canonicalgraph: zero Hop")
	}
	query, args := hopQuery(h)
	return walk(q, query, args, from)
}

func walk(q Querier, query string, args []any, from [][]byte) ([]Edge, error) {
	from = database.UniqueBlobIDs(from)
	var out []Edge
	err := sqlbatch.Query(q, query, `)`, args, from, func(rows *sql.Rows) error {
		var e Edge
		if err := rows.Scan(&e.From, &e.Association, &e.AssociationRef, &e.To, &e.ToRef); err != nil {
			return err
		}
		e.From = append([]byte(nil), e.From...)
		e.Association = append([]byte(nil), e.Association...)
		e.To = append([]byte(nil), e.To...)
		out = append(out, e)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if c := strings.Compare(string(a.From), string(b.From)); c != 0 {
			return c < 0
		}
		if a.AssociationRef != b.AssociationRef {
			return a.AssociationRef < b.AssociationRef
		}
		return a.ToRef < b.ToRef
	})
	return out, nil
}

// Targets returns the distinct to-handles of edges, in edge order.
func Targets(edges []Edge) [][]byte {
	seen := map[string]bool{}
	var out [][]byte
	for _, e := range edges {
		if !seen[string(e.To)] {
			seen[string(e.To)] = true
			out = append(out, e.To)
		}
	}
	return out
}
