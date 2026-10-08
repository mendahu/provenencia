package promotealign

import (
	"database/sql"

	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database/canonicalgraph"
	"github.com/mendahu/provenencia/core/database/matching"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/match"
)

type primarySubject struct {
	id   []byte
	ref  string
	kind string
}

// loadLayer builds the Evidence layer: primary Subjects, values, bridges.
func loadLayer(q Querier, sourceID []byte) (graphalign.Layer, map[string]primarySubject, error) {
	rows, err := q.Query(`SELECT s.id, s.ref, st.key, st.origin
		FROM subjects s
		JOIN subject_types st ON st.id = s.subject_type_id
		WHERE s.source_id = ?
		ORDER BY s.ref COLLATE NOCASE`, sourceID)
	if err != nil {
		return graphalign.Layer{}, nil, err
	}
	defer rows.Close()

	// Collect ids first: Catalog uses MaxOpenConns(1), so we must not hold
	// this rows cursor open while the value reads run.
	type row struct {
		id           []byte
		ref, key, or string
	}
	var listed []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.id, &r.ref, &r.key, &r.or); err != nil {
			return graphalign.Layer{}, nil, err
		}
		listed = append(listed, r)
	}
	if err := rows.Err(); err != nil {
		return graphalign.Layer{}, nil, err
	}
	_ = rows.Close()

	primary := map[string]primarySubject{}
	var subjects []graphalign.Subject
	byKind := map[string][][]byte{}
	for _, r := range listed {
		if !promote.PrimaryKind(r.key, r.or) {
			continue
		}
		if _, ok := match.DefaultProfile(r.key); !ok {
			continue
		}
		primary[string(r.id)] = primarySubject{id: append([]byte(nil), r.id...), ref: r.ref, kind: r.key}
		byKind[r.key] = append(byKind[r.key], append([]byte(nil), r.id...))
	}
	values := map[string]match.Values{}
	for kind, ids := range byKind {
		profile, _ := match.DefaultProfile(kind)
		vals, err := matching.SubjectsValues(q, ids, profile)
		if err != nil {
			return graphalign.Layer{}, nil, err
		}
		for id, v := range vals {
			values[id] = v
		}
	}
	for _, r := range listed {
		ps, ok := primary[string(r.id)]
		if !ok {
			continue
		}
		vals := values[string(r.id)]
		if vals == nil {
			vals = match.Values{}
		}
		subjects = append(subjects, graphalign.Subject{
			ID: ps.id, Ref: ps.ref, Kind: ps.kind, Values: vals, Provenance: 1,
		})
	}

	metas, err := loadMetas(q)
	if err != nil {
		return graphalign.Layer{}, nil, err
	}

	bridges, err := loadLayerBridges(q, sourceID, primary, byKind)
	if err != nil {
		return graphalign.Layer{}, nil, err
	}

	return graphalign.Layer{Subjects: subjects, Bridges: bridges, Metas: metas}, primary, nil
}

func loadMetas(q Querier) ([]match.PropertyMeta, error) {
	seen := map[match.Property]bool{}
	var out []match.PropertyMeta
	for _, kind := range []string{"person", "event", "place"} {
		profile, ok := match.DefaultProfile(kind)
		if !ok {
			continue
		}
		for _, prop := range profile.Properties() {
			if seen[prop] {
				continue
			}
			seen[prop] = true
			var valueType, card string
			err := q.QueryRow(`SELECT value_type, cardinality FROM properties WHERE key = ? AND origin = ?`,
				prop.Key, prop.Origin).Scan(&valueType, &card)
			if err == sql.ErrNoRows {
				continue
			}
			if err != nil {
				return nil, err
			}
			out = append(out, match.PropertyMeta{
				Property: prop, ValueType: valueType, Cardinality: card,
			})
		}
	}
	return out, nil
}

// neighborTypeProperty is the Property whose term types a bridge's neighbor
// in an edge signature (an event's type). Kinds without one sign by kind only.
var neighborTypeProperty = map[string]string{"event": "event_type"}

// Product hops used to discover undirected bridges on the layer.
var (
	hopPersonToEvent  = canonicalgraph.MustHop("participation", "person", "event", nil)
	hopEventToPlace   = canonicalgraph.PlacesOfEvent
	hopPersonRelated  = canonicalgraph.MustHop("relationship", "person", "related_to", nil)
	hopPlaceParent    = canonicalgraph.ParentsOfPlace
	hopPlaceSuccessor = canonicalgraph.SuccessorsOfPlace
)

func loadLayerBridges(q Querier, sourceID []byte, primary map[string]primarySubject, byKind map[string][][]byte) ([]graphalign.Bridge, error) {
	var out []graphalign.Bridge
	seen := map[string]bool{}
	add := func(a, b []byte, sig graphalign.EdgeSignature) {
		if _, ok := primary[string(a)]; !ok {
			return
		}
		if _, ok := primary[string(b)]; !ok {
			return
		}
		key := bridgeKey(a, b, sig)
		if seen[key] {
			return
		}
		seen[key] = true
		out = append(out, graphalign.Bridge{
			A: append([]byte(nil), a...), B: append([]byte(nil), b...), Signature: sig,
		})
	}
	walk := func(kind string, hop canonicalgraph.Hop) ([]canonicalgraph.Edge, error) {
		if len(byKind[kind]) == 0 {
			return nil, nil
		}
		return canonicalgraph.WalkSource(q, sourceID, hop, byKind[kind])
	}
	assocs := func(edges []canonicalgraph.Edge) [][]byte {
		ids := make([][]byte, len(edges))
		for i, e := range edges {
			ids[i] = e.Association
		}
		return ids
	}

	// Person ↔ event (participation), any role.
	participations, err := walk("person", hopPersonToEvent)
	if err != nil {
		return nil, err
	}
	roles, err := observedTerms(q, assocs(participations), connectrules.DisambiguationRole)
	if err != nil {
		return nil, err
	}
	var events [][]byte
	for _, e := range participations {
		events = append(events, e.To)
	}
	eventTypes, err := observedTerms(q, events, neighborTypeProperty["event"])
	if err != nil {
		return nil, err
	}
	for _, e := range participations {
		add(e.From, e.To, graphalign.EdgeSignature{
			BridgeType: "participation", RoleOrType: roles[string(e.Association)],
			NeighborKind: "event", NeighborTypeTerm: eventTypes[string(e.To)],
		})
	}

	// Event ↔ place (location).
	locations, err := walk("event", hopEventToPlace)
	if err != nil {
		return nil, err
	}
	for _, e := range locations {
		add(e.From, e.To, graphalign.EdgeSignature{BridgeType: "location", NeighborKind: "place"})
	}

	// Person ↔ person (relationship).
	relationships, err := walk("person", hopPersonRelated)
	if err != nil {
		return nil, err
	}
	relTypes, err := observedTerms(q, assocs(relationships), connectrules.DisambiguationRelationshipType)
	if err != nil {
		return nil, err
	}
	for _, e := range relationships {
		add(e.From, e.To, graphalign.EdgeSignature{
			BridgeType: "relationship", RoleOrType: relTypes[string(e.Association)], NeighborKind: "person",
		})
	}

	// Place ↔ place (part_of / succeeded_by).
	for _, step := range []struct {
		hop  canonicalgraph.Hop
		term string
	}{
		{hopPlaceParent, "part_of"},
		{hopPlaceSuccessor, "succeeded_by"},
	} {
		edges, err := walk("place", step.hop)
		if err != nil {
			return nil, err
		}
		for _, e := range edges {
			add(e.From, e.To, graphalign.EdgeSignature{
				BridgeType: "place_relationship", RoleOrType: step.term, NeighborKind: "place",
			})
		}
	}

	return out, nil
}

func bridgeKey(a, b []byte, sig graphalign.EdgeSignature) string {
	if string(a) > string(b) {
		a, b = b, a
	}
	return string(a) + "|" + string(b) + "|" + sig.Key()
}
