package promotealign

import (
	"database/sql"
	"errors"

	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database/canonicalgraph"
	"github.com/mendahu/provenencia/core/database/matching"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/match"
)

// canonDiameter is how many hops to expand from seed handles (design §7).
const canonDiameter = 5

// seedCandidateLimit is top-k property-only candidates per unpromoted Subject.
const seedCandidateLimit = 5

func loadCanon(q Querier, primary map[string]primarySubject, fixed []graphalign.Fixed) (graphalign.Canon, error) {
	seeds := map[string][]byte{} // handleID → copy
	addSeed := func(id []byte) {
		if len(id) != 16 {
			return
		}
		seeds[string(id)] = append([]byte(nil), id...)
	}
	for _, f := range fixed {
		addSeed(f.HandleID)
	}

	// Top-k property-only candidates for unpromoted Subjects.
	fixedSubjects := map[string]bool{}
	for _, f := range fixed {
		fixedSubjects[string(f.SubjectID)] = true
	}
	for _, ps := range primary {
		if fixedSubjects[string(ps.id)] {
			continue
		}
		res, err := matching.ForSubject(q, ps.id, matching.Options{Limit: seedCandidateLimit})
		if err != nil {
			return graphalign.Canon{}, err
		}
		for _, m := range res.Matches {
			addSeed(m.EntityID)
		}
	}

	frontier := make([][]byte, 0, len(seeds))
	for _, id := range seeds {
		frontier = append(frontier, id)
	}

	handles := map[string]graphalign.Handle{}
	var edges []graphalign.CanonEdge
	edgeSeen := map[string]bool{}

	loadHandles := func(ids [][]byte) error {
		for _, id := range ids {
			if _, ok := handles[string(id)]; ok {
				continue
			}
			var (
				ref, kind, origin string
			)
			err := q.QueryRow(`SELECT e.ref, st.key, st.origin FROM canonical_entities e
				JOIN subject_types st ON st.id = e.subject_type_id
				WHERE e.id = ? AND e.merged_into_id IS NULL`, id).Scan(&ref, &kind, &origin)
			if err != nil {
				return err
			}
			if !promotePrimary(kind, origin) {
				continue
			}
			profile, ok := match.DefaultProfile(kind)
			if !ok {
				continue
			}
			vals, _, err := matching.LoadEntityValues(q, id, profile)
			if err != nil {
				return err
			}
			handles[string(id)] = graphalign.Handle{
				ID: append([]byte(nil), id...), Ref: ref, Kind: kind, Values: vals,
			}
		}
		return nil
	}
	if err := loadHandles(frontier); err != nil {
		return graphalign.Canon{}, err
	}

	hops := []struct {
		hop          canonicalgraph.Hop
		bridgeType   string
		roleFallback string // used when filter term is fixed on the hop
		toKind       string
	}{
		{hopPersonToEvent, "participation", "", "event"},
		{canonicalgraph.SubjectsOfEvent, "participation", "subject", "person"},
		{hopEventToPlace, "location", "", "place"},
		{canonicalgraph.EventsAtPlace, "location", "", "event"},
		{hopPersonRelated, "relationship", "", "person"},
		{hopPlaceParent, "place_relationship", "part_of", "place"},
		{canonicalgraph.PartsOfPlace, "place_relationship", "part_of", "place"},
		{hopPlaceSuccessor, "place_relationship", "succeeded_by", "place"},
		{canonicalgraph.PredecessorsOfPlace, "place_relationship", "succeeded_by", "place"},
	}

	for depth := 0; depth < canonDiameter && len(frontier) > 0; depth++ {
		var next [][]byte
		nextSeen := map[string]bool{}
		for _, step := range hops {
			walked, err := canonicalgraph.Walk(q, step.hop, frontier)
			if err != nil {
				return graphalign.Canon{}, err
			}
			for _, e := range walked {
				sig, err := canonEdgeSignature(q, e, step.bridgeType, step.roleFallback, step.toKind)
				if err != nil {
					return graphalign.Canon{}, err
				}
				ek := string(e.From) + "|" + string(e.To) + "|" + sig.Key()
				if !edgeSeen[ek] {
					edgeSeen[ek] = true
					edges = append(edges, graphalign.CanonEdge{
						From: append([]byte(nil), e.From...),
						To:   append([]byte(nil), e.To...),
						Signature: sig,
					})
				}
				if _, ok := handles[string(e.To)]; !ok && !nextSeen[string(e.To)] {
					nextSeen[string(e.To)] = true
					next = append(next, append([]byte(nil), e.To...))
				}
			}
		}
		if err := loadHandles(next); err != nil {
			return graphalign.Canon{}, err
		}
		// Drop non-primary that loadHandles skipped.
		kept := next[:0]
		for _, id := range next {
			if _, ok := handles[string(id)]; ok {
				kept = append(kept, id)
			}
		}
		frontier = kept
	}

	out := graphalign.Canon{Edges: edges}
	for _, h := range handles {
		out.Handles = append(out.Handles, h)
	}
	return out, nil
}

func promotePrimary(kind, origin string) bool {
	return origin == "provenencia" && (kind == "person" || kind == "event" || kind == "place")
}

func canonEdgeSignature(q Querier, e canonicalgraph.Edge, bridgeType, roleFallback, toKind string) (graphalign.EdgeSignature, error) {
	sig := graphalign.EdgeSignature{
		BridgeType:   bridgeType,
		RoleOrType:   roleFallback,
		NeighborKind: toKind,
	}
	switch bridgeType {
	case "participation":
		if roleFallback == "" {
			role, err := termOnEntity(q, e.Association, connectrules.DisambiguationRole)
			if err != nil {
				return sig, err
			}
			sig.RoleOrType = role
		}
		typeTerm, err := termOnEntity(q, e.To, "event_type")
		if err != nil {
			return sig, err
		}
		if toKind == "event" {
			sig.NeighborTypeTerm = typeTerm
		} else {
			// Walking event → person: type term is on the from (event) side for
			// matching the layer's person→event signature neighbor type.
			fromType, err := termOnEntity(q, e.From, "event_type")
			if err != nil {
				return sig, err
			}
			sig.NeighborKind = "person"
			sig.NeighborTypeTerm = ""
			_ = fromType
			// Align matches signatures by full Key(); layer uses NeighborKind=event
			// with the event's type. Canon edges are stored undirected in Align's
			// adjacency, so store the person→event orientation's signature.
			sig = graphalign.EdgeSignature{
				BridgeType: "participation", RoleOrType: sig.RoleOrType,
				NeighborKind: "event", NeighborTypeTerm: fromType,
			}
		}
	case "relationship":
		rel, err := termOnEntity(q, e.Association, connectrules.DisambiguationRelationshipType)
		if err != nil {
			return sig, err
		}
		sig.RoleOrType = rel
		sig.NeighborKind = "person"
	case "place_relationship":
		sig.NeighborKind = "place"
	case "location":
		sig.NeighborKind = toKind
	}
	return sig, nil
}

func termOnEntity(q Querier, entityID []byte, propertyKey string) (string, error) {
	var term string
	err := q.QueryRow(`SELECT t.key FROM auto_reconciler_values r
		JOIN properties p ON p.id = r.property_id AND p.key = ? AND p.origin = 'provenencia'
		JOIN property_terms t ON t.id = r.value_term_id
		WHERE r.entity_id = ? AND r.rank = 1 AND r.reason = 'kept'
		LIMIT 1`, propertyKey, entityID).Scan(&term)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return term, err
}
