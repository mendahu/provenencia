package promotealign

import (
	"database/sql"
	"sort"

	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database/canonicalgraph"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/matching"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/hops"
	"github.com/mendahu/provenencia/core/match"
)

type primarySubject struct {
	id   []byte
	ref  string
	kind string
}

// loadLayer builds the Evidence layer from a loaded Source. Property metas
// stay a catalog read; they are not part of one Source.
func loadLayer(g *graphcache.Graph, q Querier, sourceID []byte) (graphalign.Layer, map[string]primarySubject, error) {
	if g == nil {
		return graphalign.Layer{}, nil, promote.ErrInvalid
	}
	sg, err := g.Source(sourceID)
	if err != nil {
		return graphalign.Layer{}, nil, err
	}
	primary := map[string]primarySubject{}
	var subjects []graphalign.Subject
	byKind := map[string][][]byte{}
	for _, s := range sg.Subjects {
		if !promote.PrimaryKind(s.TypeKey, s.TypeOrigin) {
			continue
		}
		if _, ok := match.DefaultProfile(s.TypeKey); !ok {
			continue
		}
		primary[string(s.ID)] = primarySubject{id: append([]byte(nil), s.ID...), ref: s.Ref, kind: s.TypeKey}
		byKind[s.TypeKey] = append(byKind[s.TypeKey], append([]byte(nil), s.ID...))
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
	for _, s := range sg.Subjects {
		ps, ok := primary[string(s.ID)]
		if !ok {
			continue
		}
		vals := values[string(s.ID)]
		if vals == nil {
			vals = match.Values{}
		}
		subjects = append(subjects, graphalign.Subject{
			ID: ps.id, Ref: ps.ref, Kind: ps.kind, Values: vals, Provenance: 1,
		})
	}
	sort.Slice(subjects, func(i, j int) bool { return subjects[i].Ref < subjects[j].Ref })

	metas, err := loadMetas(q)
	if err != nil {
		return graphalign.Layer{}, nil, err
	}
	bridges := make([]graphalign.Bridge, 0, len(sg.Bridges))
	for _, b := range sg.Bridges {
		bridges = append(bridges, graphalign.Bridge{
			A: append([]byte(nil), b.A...), B: append([]byte(nil), b.B...),
			Signature: graphalign.EdgeSignature{
				BridgeType: b.BridgeType, RoleOrType: b.Term, NeighborKind: b.NeighborKind,
				NeighborTypeTerm: b.NeighborTypeTerm, Directed: b.Directed,
			},
		})
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
	hopPersonToEvent  = hops.MustHop("participation", "person", "event", nil)
	hopEventToPlace   = hops.PlacesOfEvent
	hopPersonRelated  = hops.MustHop("relationship", "person", "related_to", nil)
	hopPlaceParent    = hops.ParentsOfPlace
	hopPlaceSuccessor = hops.SuccessorsOfPlace
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
	kin, err := loadKinship(q)
	if err != nil {
		return nil, err
	}
	walk := func(kind string, hop hops.Hop) ([]canonicalgraph.Edge, error) {
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
		// "John child of Mary" reads as "Mary parent of John" (or the reverse)
		// so one relationship has one signature whichever end recorded it.
		rel, flip := kin.canonical(connectrules.DisambiguationRelationshipType, relTypes[string(e.Association)])
		from, to := e.From, e.To
		if flip {
			from, to = to, from
		}
		add(from, to, graphalign.EdgeSignature{
			BridgeType: "relationship", RoleOrType: rel, NeighborKind: "person",
			Directed: kin.directed[connectrules.DisambiguationRelationshipType+"|"+rel],
		})
	}

	// Place ↔ place (part_of / succeeded_by).
	for _, step := range []struct {
		hop  hops.Hop
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
				Directed: kin.directed[connectrules.DisambiguationPlaceRelationshipType+"|"+step.term],
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
