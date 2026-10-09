package promotealign

import (
	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalgraph"
	"github.com/mendahu/provenencia/core/database/matching"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/match"
)

// canonDiameter is how many hops to expand from seed handles (design §7).
const canonDiameter = 5

// seedCandidateLimit is top-k property-only candidates per unpromoted Subject.
const seedCandidateLimit = 5

// canonStep is one hop the expansion follows, and how its edges sign.
// reversed hops walk from the bridge's second endpoint; their edges are
// flipped so CanonEdge.From is always the first, as on the layer.
type canonStep struct {
	hop          canonicalgraph.Hop
	bridgeType   string
	roleFallback string // the hop's fixed term, when it filters on one
	toKind       string
	reversed     bool
}

var canonSteps = []canonStep{
	{hopPersonToEvent, "participation", "", "event", false},
	{canonicalgraph.SubjectsOfEvent, "participation", "subject", "person", true},
	{hopEventToPlace, "location", "", "place", false},
	{canonicalgraph.EventsAtPlace, "location", "", "event", true},
	{hopPersonRelated, "relationship", "", "person", false},
	// A relationship is reached from either end: a handle on related_to
	// (a child under "parent") walks back to the person end too.
	{hopPersonRelated.Reverse(), "relationship", "", "person", true},
	{hopPlaceParent, "place_relationship", "part_of", "place", false},
	{canonicalgraph.PartsOfPlace, "place_relationship", "part_of", "place", true},
	{hopPlaceSuccessor, "place_relationship", "succeeded_by", "place", false},
	{canonicalgraph.PredecessorsOfPlace, "place_relationship", "succeeded_by", "place", true},
}

// loadCanon gathers a bounded piece of the canonical graph around the layer:
// the fixed handles and each unfixed Subject's top property-only candidates,
// expanded breadth-first. Every read is batched — candidates once per kind,
// one walk per step per hop, terms and handle values per hop — so the query
// count follows the hop count, not the size of the layer or the catalog.
func loadCanon(q Querier, layer graphalign.Layer, fixed []graphalign.Fixed) (graphalign.Canon, error) {
	handles := map[string]graphalign.Handle{}
	var seeds [][]byte
	for _, f := range fixed {
		if len(f.HandleID) == 16 {
			seeds = append(seeds, f.HandleID)
		}
	}

	// Top-k property-only candidates for unfixed Subjects: each kind's
	// candidates load once, and every Subject of that kind ranks against them.
	fixedSubjects := map[string]bool{}
	for _, f := range fixed {
		fixedSubjects[string(f.SubjectID)] = true
	}
	candidates := map[string][]match.Candidate{}
	for _, s := range layer.Subjects {
		if fixedSubjects[string(s.ID)] || len(s.Values) == 0 {
			continue
		}
		profile, ok := match.DefaultProfile(s.Kind)
		if !ok {
			continue
		}
		cands, loaded := candidates[s.Kind]
		if !loaded {
			var err error
			cands, err = matching.CandidatesOfType(q, s.Kind, "provenencia", profile)
			if err != nil {
				return graphalign.Canon{}, err
			}
			candidates[s.Kind] = cands
		}
		for _, m := range match.Rank(profile, s.Values, cands, seedCandidateLimit) {
			seeds = append(seeds, m.EntityID)
		}
	}

	frontier, err := loadHandles(q, database.UniqueBlobIDs(seeds), handles)
	if err != nil {
		return graphalign.Canon{}, err
	}

	kin, err := loadKinship(q)
	if err != nil {
		return graphalign.Canon{}, err
	}
	var edges []graphalign.CanonEdge
	edgeSeen := map[string]bool{}
	for depth := 0; depth < canonDiameter && len(frontier) > 0; depth++ {
		type walked struct {
			step  canonStep
			edges []canonicalgraph.Edge
		}
		var all []walked
		var assocRoles, assocRels, events [][]byte
		for _, step := range canonSteps {
			es, err := canonicalgraph.Walk(q, step.hop, frontier)
			if err != nil {
				return graphalign.Canon{}, err
			}
			all = append(all, walked{step, es})
			for _, e := range es {
				switch step.bridgeType {
				case "participation":
					assocRoles = append(assocRoles, e.Association)
					events = append(events, e.From, e.To)
				case "relationship":
					assocRels = append(assocRels, e.Association)
				}
			}
		}
		roles, err := keptTerms(q, assocRoles, connectrules.DisambiguationRole)
		if err != nil {
			return graphalign.Canon{}, err
		}
		rels, err := keptTerms(q, assocRels, connectrules.DisambiguationRelationshipType)
		if err != nil {
			return graphalign.Canon{}, err
		}
		eventTypes, err := keptTerms(q, events, neighborTypeProperty["event"])
		if err != nil {
			return graphalign.Canon{}, err
		}

		var next [][]byte
		for _, w := range all {
			for _, e := range w.edges {
				sig := canonEdgeSignature(e, w.step, roles, rels, eventTypes, kin.directed)
				from, to := e.From, e.To
				if w.step.reversed {
					from, to = to, from
				}
				if w.step.bridgeType == "relationship" {
					// Same reading as the layer: a term and its inverse share
					// one key, with the ends swapped.
					rel, flip := kin.canonical(connectrules.DisambiguationRelationshipType, sig.RoleOrType)
					if flip {
						from, to = to, from
						sig.RoleOrType = rel
						sig.Directed = kin.directed[connectrules.DisambiguationRelationshipType+"|"+rel]
					}
				}
				ek := string(from) + "|" + string(to) + "|" + sig.Key()
				if !edgeSeen[ek] {
					edgeSeen[ek] = true
					edges = append(edges, graphalign.CanonEdge{
						From:      append([]byte(nil), from...),
						To:        append([]byte(nil), to...),
						Signature: sig,
					})
				}
				if _, ok := handles[string(e.To)]; !ok {
					next = append(next, e.To)
				}
			}
		}
		frontier, err = loadHandles(q, database.UniqueBlobIDs(next), handles)
		if err != nil {
			return graphalign.Canon{}, err
		}
	}

	out := graphalign.Canon{Edges: edges}
	for _, h := range handles {
		out.Handles = append(out.Handles, h)
	}
	return out, nil
}

// loadHandles adds the primary handles among ids to handles (one header read
// and one values read per kind, per IN batch) and returns the ids it added.
func loadHandles(q Querier, ids [][]byte, handles map[string]graphalign.Handle) ([][]byte, error) {
	var fresh [][]byte
	for _, id := range ids {
		if _, ok := handles[string(id)]; !ok && len(id) == 16 {
			fresh = append(fresh, id)
		}
	}
	byKind := map[string][][]byte{}
	refs := map[string]string{}
	for start := 0; start < len(fresh); start += inBatch {
		chunk := fresh[start:min(start+inBatch, len(fresh))]
		rows, err := q.Query(`SELECT e.id, e.ref, st.key, st.origin FROM canonical_entities e
			JOIN subject_types st ON st.id = e.subject_type_id
			WHERE e.merged_into_id IS NULL AND e.id IN (`+database.SQLInPlaceholders(len(chunk))+`)`,
			database.BlobArgs(chunk)...)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var id []byte
			var ref, kind, origin string
			if err := rows.Scan(&id, &ref, &kind, &origin); err != nil {
				_ = rows.Close()
				return nil, err
			}
			if !promotePrimary(kind, origin) {
				continue
			}
			refs[string(id)] = ref
			byKind[kind] = append(byKind[kind], id)
		}
		err = rows.Err()
		_ = rows.Close()
		if err != nil {
			return nil, err
		}
	}
	var added [][]byte
	for _, kind := range []string{"person", "event", "place"} {
		ids := byKind[kind]
		if len(ids) == 0 {
			continue
		}
		profile, _ := match.DefaultProfile(kind)
		values, err := matching.EntitiesValues(q, ids, profile)
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			vals := values[string(id)]
			if vals == nil {
				vals = match.Values{}
			}
			handles[string(id)] = graphalign.Handle{ID: id, Ref: refs[string(id)], Kind: kind, Values: vals}
			added = append(added, id)
		}
	}
	return added, nil
}

func promotePrimary(kind, origin string) bool {
	return origin == "provenencia" && (kind == "person" || kind == "event" || kind == "place")
}

// canonEdgeSignature signs a walked edge the way the layer signs its
// bridges, whichever way the walk crossed it: a participation is person →
// event carrying the event's type, a location is event → place.
func canonEdgeSignature(e canonicalgraph.Edge, step canonStep, roles, rels, eventTypes map[string]string, directed map[string]bool) graphalign.EdgeSignature {
	switch step.bridgeType {
	case "participation":
		role := step.roleFallback
		if role == "" {
			role = roles[string(e.Association)]
		}
		event := e.To
		if step.toKind != "event" {
			event = e.From
		}
		return graphalign.EdgeSignature{
			BridgeType: "participation", RoleOrType: role,
			NeighborKind: "event", NeighborTypeTerm: eventTypes[string(event)],
		}
	case "relationship":
		rel := rels[string(e.Association)]
		return graphalign.EdgeSignature{
			BridgeType: "relationship", RoleOrType: rel, NeighborKind: "person",
			Directed: directed[connectrules.DisambiguationRelationshipType+"|"+rel],
		}
	case "location":
		return graphalign.EdgeSignature{BridgeType: "location", NeighborKind: "place"}
	default:
		return graphalign.EdgeSignature{
			BridgeType: step.bridgeType, RoleOrType: step.roleFallback, NeighborKind: "place",
			Directed: directed[connectrules.DisambiguationPlaceRelationshipType+"|"+step.roleFallback],
		}
	}
}
