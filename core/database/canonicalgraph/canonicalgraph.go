// Package canonicalgraph walks the canonical graph: handles joined through
// association handles (a participation, a location) whose endpoint values
// the auto-reconciler kept.
//
// A Hop names one bridge from the connectrules registry and the two endpoint
// Properties it crosses, checked against that registry when the Hop is made,
// so the bridge shape lives in one place. Walk follows a Hop for a set of
// handles in a fixed number of queries, each served by the edge index.
// Every header, dependency, and later tree walk is built from Hops rather
// than its own join chain.
package canonicalgraph

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database"
)

// batch bounds an IN list. A walk is still a fixed number of queries.
const batch = 500

// Querier is *sql.Tx or *sql.DB.
type Querier interface {
	Query(query string, args ...any) (*sql.Rows, error)
}

// TermFilter keeps an association only when its kept rank-1 value of the
// bridge's disambiguation Property is the given term (role = subject).
type TermFilter struct {
	PropertyKey string
	TermKey     string
}

// Hop crosses one bridge from one endpoint Property to the other.
type Hop struct {
	origin, bridge string
	from, to       string
	filter         *TermFilter
}

// NewHop checks a hop against the registered bridge: both Properties must be
// its endpoints, and a filter must be on its disambiguation Property.
func NewHop(bridgeTypeKey, fromProperty, toProperty string, filter *TermFilter) (Hop, error) {
	b, ok := connectrules.LookupBridge(bridgeTypeKey)
	if !ok {
		return Hop{}, fmt.Errorf("canonicalgraph: no bridge %q", bridgeTypeKey)
	}
	if fromProperty == toProperty || !isEndpoint(b, fromProperty) || !isEndpoint(b, toProperty) {
		return Hop{}, fmt.Errorf("canonicalgraph: %s does not join %s to %s", bridgeTypeKey, fromProperty, toProperty)
	}
	if filter != nil && (b.Disambiguation != filter.PropertyKey || filter.TermKey == "") {
		return Hop{}, fmt.Errorf("canonicalgraph: %s does not disambiguate by %s", bridgeTypeKey, filter.PropertyKey)
	}
	origin := strings.TrimSpace(b.Origin)
	if origin == "" {
		origin = connectrules.OriginProvenencia
	}
	h := Hop{origin: origin, bridge: b.BridgeTypeKey, from: fromProperty, to: toProperty}
	if filter != nil {
		f := *filter
		h.filter = &f
	}
	return h, nil
}

// MustHop is NewHop for package-level hops over product bridges.
func MustHop(bridgeTypeKey, fromProperty, toProperty string, filter *TermFilter) Hop {
	h, err := NewHop(bridgeTypeKey, fromProperty, toProperty, filter)
	if err != nil {
		panic(err)
	}
	return h
}

// Reverse is the same hop walked the other way.
func (h Hop) Reverse() Hop {
	h.from, h.to = h.to, h.from
	return h
}

func isEndpoint(b connectrules.Bridge, propertyKey string) bool {
	for _, e := range b.Endpoints {
		if e.PropertyKey == propertyKey {
			return true
		}
	}
	return false
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

// query is the walk up to its IN list, and the arguments before the ids.
func (h Hop) query() (string, []any) {
	query := sqlWalkSelect
	args := []any{h.bridge, h.origin, h.to, h.origin}
	if h.filter != nil {
		query += sqlWalkFilter
		args = append(args, h.filter.PropertyKey, h.origin, h.filter.TermKey, h.origin)
	}
	query += sqlWalkWhere
	args = append(args, h.from, h.origin)
	return query, args
}

// Walk follows h from each handle, ordered by from-handle, then association
// ref, then to-handle ref.
func Walk(q Querier, h Hop, from [][]byte) ([]Edge, error) {
	if h.bridge == "" {
		return nil, fmt.Errorf("canonicalgraph: zero Hop")
	}
	from = database.UniqueBlobIDs(from)
	query, args := h.query()
	var out []Edge
	for start := 0; start < len(from); start += batch {
		ids := from[start:min(start+batch, len(from))]
		rows, err := q.Query(query+database.SQLInPlaceholders(len(ids))+`)`, append(args, database.BlobArgs(ids)...)...)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var e Edge
			if err := rows.Scan(&e.From, &e.Association, &e.AssociationRef, &e.To, &e.ToRef); err != nil {
				_ = rows.Close()
				return nil, err
			}
			e.From = append([]byte(nil), e.From...)
			e.Association = append([]byte(nil), e.Association...)
			e.To = append([]byte(nil), e.To...)
			out = append(out, e)
		}
		err = rows.Err()
		_ = rows.Close()
		if err != nil {
			return nil, err
		}
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
