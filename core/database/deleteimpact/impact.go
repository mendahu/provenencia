package deleteimpact

import (
	"database/sql"
	"strings"

	"github.com/mendahu/provenencia/core/database/catalogmodel"
)

// ParseKind maps a GetDeleteImpact kind string to a registered kind.
func ParseKind(s string) (catalogmodel.Kind, error) {
	k := catalogmodel.Kind(strings.TrimSpace(s))
	if _, ok := tableByKind(k); !ok {
		return "", ErrInvalid
	}
	return k, nil
}

// Snapshot reports Impact inside a read transaction and rolls it back.
// GetDeleteImpact is not a write, so it does not go through writes.Run.
func Snapshot(db *sql.DB, kind catalogmodel.Kind, id []byte) (Report, error) {
	if db == nil {
		return Report{}, ErrInvalid
	}
	tx, err := db.Begin()
	if err != nil {
		return Report{}, err
	}
	defer func() { _ = tx.Rollback() }()
	return Impact(tx, kind, id)
}

// Impact reports whether id of kind can be erased and, if not, the inbound list.
func Impact(tx *sql.Tx, kind catalogmodel.Kind, id []byte) (Report, error) {
	if tx == nil {
		return Report{}, ErrInvalid
	}
	spec, ok := tableByKind(kind)
	if !ok {
		return Report{}, ErrInvalid
	}
	if spec.Bucket == catalogmodel.BucketInfra || spec.Bucket == catalogmodel.BucketSkip {
		return Report{Allowed: false, Gate: GateInfra}, nil
	}
	if len(id) != 16 {
		return Report{}, ErrInvalid
	}

	var one int
	err := tx.QueryRow(spec.Exists, id).Scan(&one)
	if err == sql.ErrNoRows {
		return Report{Allowed: false, Gate: GateNotFound}, nil
	}
	if err != nil {
		return Report{}, err
	}

	for _, e := range extraGatesFor(kind) {
		held, err := e.Holds(tx, id)
		if err != nil {
			return Report{}, err
		}
		if held {
			return Report{Allowed: false, Gate: e.Gate}, nil
		}
	}

	groups, err := collectGroups(tx, inboundFor(kind), id)
	if err != nil {
		return Report{}, err
	}
	cascades, err := collectGroups(tx, cascadesFor(kind), id)
	if err != nil {
		return Report{}, err
	}
	if len(groups) > 0 {
		return Report{Allowed: false, Gate: GateInbound, Groups: groups, Cascades: cascades}, nil
	}
	return Report{Allowed: true, Gate: GateOK, Cascades: cascades}, nil
}

// collectGroups counts and lists each edge with rows for id; empty edges are skipped.
func collectGroups(tx *sql.Tx, edges []inboundEdge, id []byte) ([]Group, error) {
	var groups []Group
	for _, edge := range edges {
		if edge.Count == nil {
			continue
		}
		total, err := edge.Count(tx, id)
		if err != nil {
			return nil, err
		}
		if total == 0 {
			continue
		}
		g := Group{Via: edge.Via, Kind: edge.Child, Total: total}
		if edge.List != nil {
			rows, err := edge.List(tx, id, listedCap)
			if err != nil {
				return nil, err
			}
			for _, row := range rows {
				listed, err := projectListed(tx, edge.Child, row)
				if err != nil {
					return nil, err
				}
				g.Listed = append(g.Listed, listed)
			}
		}
		groups = append(groups, g)
	}
	return groups, nil
}

// ListedGroup shapes rows a domain read found into one report group, projected
// and capped as Impact lists inbound rows. Callers pass rows with ID and Ref
// set, from the query that found them.
func ListedGroup(tx *sql.Tx, via string, kind catalogmodel.Kind, rows []Listed) (Group, error) {
	if tx == nil {
		return Group{}, ErrInvalid
	}
	g := Group{Via: via, Kind: kind, Total: len(rows)}
	for i, r := range rows {
		if i == listedCap {
			break
		}
		listed, err := projectListed(tx, kind, probeRow{ID: r.ID, Ref: r.Ref})
		if err != nil {
			return Group{}, err
		}
		g.Listed = append(g.Listed, listed)
	}
	return g, nil
}
