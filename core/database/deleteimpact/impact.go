package deleteimpact

import (
	"database/sql"
	"strings"
)

// ParseKind maps a GetDeleteImpact kind string to a registered Kind.
func ParseKind(s string) (Kind, error) {
	k := Kind(strings.TrimSpace(s))
	if _, ok := tableByKind(k); !ok {
		return "", ErrInvalid
	}
	return k, nil
}

// Impact reports whether id of kind can be erased and, if not, the inbound list.
func Impact(tx *sql.Tx, kind Kind, id []byte) (Report, error) {
	if tx == nil {
		return Report{}, ErrInvalid
	}
	spec, ok := tableByKind(kind)
	if !ok {
		return Report{}, ErrInvalid
	}
	if spec.Bucket == BucketInfra || spec.Bucket == BucketSkip {
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

	if kind == KindObservation {
		edge, err := isEdgeObservation(tx, id)
		if err != nil {
			return Report{}, err
		}
		if edge {
			return Report{Allowed: false, Gate: GateEdgeLocked}, nil
		}
	}

	if gate, err := originGate(tx, kind, id); err != nil {
		return Report{}, err
	} else if gate != GateOK {
		return Report{Allowed: false, Gate: gate}, nil
	}

	var groups []Group
	for _, edge := range inboundFor(kind) {
		if edge.Count == nil {
			continue
		}
		total, err := edge.Count(tx, id)
		if err != nil {
			return Report{}, err
		}
		if total == 0 {
			continue
		}
		g := Group{Via: edge.Via, Kind: edge.Child, Total: total}
		if edge.List != nil {
			rows, err := edge.List(tx, id, listedCap)
			if err != nil {
				return Report{}, err
			}
			for _, row := range rows {
				listed, err := projectListed(tx, edge.Child, row)
				if err != nil {
					return Report{}, err
				}
				g.Listed = append(g.Listed, listed)
			}
		}
		groups = append(groups, g)
	}
	if len(groups) > 0 {
		return Report{Allowed: false, Gate: GateInbound, Groups: groups}, nil
	}
	return Report{Allowed: true, Gate: GateOK}, nil
}

func originGate(tx *sql.Tx, kind Kind, id []byte) (Gate, error) {
	rule, ok := originRuleFor(kind)
	if !ok {
		return GateOK, nil
	}
	var origin string
	err := tx.QueryRow(rule.OriginSQL, id).Scan(&origin)
	if err == sql.ErrNoRows {
		return GateOK, nil
	}
	if err != nil {
		return "", err
	}
	origin = strings.TrimSpace(origin)
	if rule.PluginNever && strings.HasPrefix(origin, originPluginPref) && len(origin) > len(originPluginPref) {
		return GateOriginLocked, nil
	}
	if rule.SeededLocked && origin == originProvenencia {
		return GateOriginLocked, nil
	}
	return GateOK, nil
}
