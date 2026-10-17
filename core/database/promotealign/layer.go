package promotealign

import (
	"sort"

	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database/canonicalgraph"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/matching"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/hops"
	"github.com/mendahu/provenencia/core/match"
	"github.com/mendahu/provenencia/core/vocabkey"
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
		if !vocabkey.IsPrimaryType(s.TypeKey, s.TypeOrigin) {
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

	metas, err := loadMetas(g)
	if err != nil {
		return graphalign.Layer{}, nil, err
	}
	bridges, err := loadLayerBridges(g, q, sourceID, primary, byKind)
	if err != nil {
		return graphalign.Layer{}, nil, err
	}
	return graphalign.Layer{Subjects: subjects, Bridges: bridges, Metas: metas}, primary, nil
}

func loadMetas(g *graphcache.Graph) ([]match.PropertyMeta, error) {
	v, err := g.Vocabulary()
	if err != nil {
		return nil, err
	}
	byKey := map[string]graphcache.VocabProperty{}
	for _, p := range v.Properties {
		byKey[p.Origin+"\x00"+p.Key] = p
	}
	seen := map[match.Property]bool{}
	var out []match.PropertyMeta
	for _, kind := range vocabkey.PrimaryKinds() {
		profile, ok := match.DefaultProfile(kind)
		if !ok {
			continue
		}
		for _, prop := range profile.Properties() {
			if seen[prop] {
				continue
			}
			seen[prop] = true
			row, ok := byKey[prop.Origin+"\x00"+prop.Key]
			if !ok {
				continue
			}
			out = append(out, match.PropertyMeta{
				Property: prop, ValueType: row.ValueType, Cardinality: row.Cardinality,
			})
		}
	}
	return out, nil
}

// neighborTypeProperty is the Property whose term types a bridge's neighbor
// in an edge signature (an event's type). Kinds without one sign by kind only.
var neighborTypeProperty = map[string]string{vocabkey.Event: vocabkey.PropertyEventType}

// Product hops used to discover undirected bridges on the layer.
var (
	hopPersonToEvent  = hops.MustHop("participation", vocabkey.Person, vocabkey.Event, nil)
	hopEventToPlace   = hops.PlacesOfEvent
	hopPersonRelated  = hops.MustHop("relationship", vocabkey.Person, "related_to", nil)
	hopPlaceParent    = hops.ParentsOfPlace
	hopPlaceSuccessor = hops.SuccessorsOfPlace
)

func loadLayerBridges(g *graphcache.Graph, q Querier, sourceID []byte, primary map[string]primarySubject, byKind map[string][][]byte) ([]graphalign.Bridge, error) {
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
	kin, err := loadKinship(g)
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
	participations, err := walk(vocabkey.Person, hopPersonToEvent)
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
	eventTypes, err := observedTerms(q, events, neighborTypeProperty[vocabkey.Event])
	if err != nil {
		return nil, err
	}
	for _, e := range participations {
		role, _, neighborKind, _ := edgeStamp("participation", roles[string(e.Association)], kin)
		add(e.From, e.To, graphalign.EdgeSignature{
			BridgeType: "participation", RoleOrType: role,
			NeighborKind: neighborKind, NeighborTypeTerm: eventTypes[string(e.To)],
		})
	}

	// Event ↔ place (location).
	locations, err := walk(vocabkey.Event, hopEventToPlace)
	if err != nil {
		return nil, err
	}
	for _, e := range locations {
		_, _, neighborKind, _ := edgeStamp("location", "", kin)
		add(e.From, e.To, graphalign.EdgeSignature{BridgeType: "location", NeighborKind: neighborKind})
	}

	// Person ↔ person (relationship).
	relationships, err := walk(vocabkey.Person, hopPersonRelated)
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
		rel, directed, neighborKind, flip := edgeStamp("relationship", relTypes[string(e.Association)], kin)
		from, to := e.From, e.To
		if flip {
			from, to = to, from
		}
		add(from, to, graphalign.EdgeSignature{
			BridgeType: "relationship", RoleOrType: rel, NeighborKind: neighborKind,
			Directed: directed,
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
		edges, err := walk(vocabkey.Place, step.hop)
		if err != nil {
			return nil, err
		}
		for _, e := range edges {
			role, directed, neighborKind, _ := edgeStamp("place_relationship", step.term, kin)
			add(e.From, e.To, graphalign.EdgeSignature{
				BridgeType: "place_relationship", RoleOrType: role, NeighborKind: neighborKind,
				Directed: directed,
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
