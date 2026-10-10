package promotealign

import (
	"sort"

	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/match"
	"github.com/mendahu/provenencia/core/valuecodec"
)

// seedCandidateLimit is the number of property-only candidates kept per
// unfixed Subject. The walk itself has no hop limit.
const seedCandidateLimit = 5

// loadCanon reads the open catalog's graph: fixed anchors, the top property
// matches for each unfixed kind, then every link those handles have. Link
// order is association ref, then the neighbor's ref, which is the order the
// walk used to break ties.
func loadCanon(g *graphcache.Graph, q Querier, layer graphalign.Layer, fixed []graphalign.Fixed) (graphalign.Canon, error) {
	props, err := propertyByID(q)
	if err != nil {
		return graphalign.Canon{}, err
	}
	types, err := typeByKey(q)
	if err != nil {
		return graphalign.Canon{}, err
	}
	kin, err := loadKinship(q)
	if err != nil {
		return graphalign.Canon{}, err
	}

	handles := map[string]graphalign.Handle{}
	var seeds [][]byte
	add := func(id []byte) error {
		if len(id) != 16 {
			return nil
		}
		if _, ok := handles[string(id)]; ok {
			seeds = append(seeds, id)
			return nil
		}
		n, err := g.Node(id)
		if err != nil || n == nil || n.Merged {
			return err
		}
		kind := typeKey(types, n.SubjectTypeID)
		if !promotePrimary(kind, "provenencia") {
			return nil
		}
		vals, err := handleValues(props, n)
		if err != nil {
			return err
		}
		handles[string(id)] = graphalign.Handle{ID: append([]byte(nil), n.ID...), Ref: n.Ref, Kind: kind, Values: vals}
		seeds = append(seeds, id)
		return nil
	}

	for _, f := range fixed {
		if len(f.HandleID) == 16 {
			if err := add(f.HandleID); err != nil {
				return graphalign.Canon{}, err
			}
		}
	}

	fixedSubjects := map[string]bool{}
	for _, f := range fixed {
		fixedSubjects[string(f.SubjectID)] = true
	}
	ranked := map[string][]match.Candidate{}
	for _, s := range layer.Subjects {
		if fixedSubjects[string(s.ID)] || len(s.Values) == 0 {
			continue
		}
		profile, ok := match.DefaultProfile(s.Kind)
		if !ok {
			continue
		}
		cands, ok := ranked[s.Kind]
		if !ok {
			typeID := types[s.Kind]
			if len(typeID) != 16 {
				ranked[s.Kind] = nil
				continue
			}
			nodes, err := g.Kind(typeID)
			if err != nil {
				return graphalign.Canon{}, err
			}
			for _, n := range nodes {
				vals, err := handleValues(props, n)
				if err != nil {
					return graphalign.Canon{}, err
				}
				cands = append(cands, match.Candidate{EntityID: n.ID, Ref: n.Ref, Values: vals})
			}
			ranked[s.Kind] = cands
		}
		for _, m := range match.Rank(profile, s.Values, cands, seedCandidateLimit) {
			if err := add(m.EntityID); err != nil {
				return graphalign.Canon{}, err
			}
		}
	}

	sort.SliceStable(seeds, func(i, j int) bool {
		return handles[string(seeds[i])].Ref < handles[string(seeds[j])].Ref
	})

	var edges []graphalign.CanonEdge
	edgeSeen := map[string]bool{}
	walked := map[string]bool{}
	queue := append([][]byte(nil), seeds...)
	for len(queue) > 0 {
		wave := queue
		queue = nil
		type reached struct {
			id []byte
			n  *graphcache.Node
		}
		var ready []reached
		var prefetch [][]byte
		for _, id := range wave {
			if walked[string(id)] {
				continue
			}
			walked[string(id)] = true
			n, err := g.Node(id)
			if err != nil {
				return graphalign.Canon{}, err
			}
			if n == nil {
				continue
			}
			ready = append(ready, reached{id, n})
			for _, l := range n.Links {
				prefetch = append(prefetch, l.Neighbor)
			}
		}
		if err := g.Nodes(prefetch); err != nil {
			return graphalign.Canon{}, err
		}
		for _, item := range ready {
			id, n := item.id, item.n
			selfKind := typeKey(types, n.SubjectTypeID)
			for _, l := range n.Links {
				neighbor, err := g.Node(l.Neighbor)
				if err != nil {
					return graphalign.Canon{}, err
				}
				if neighbor == nil || neighbor.Merged {
					continue
				}
				nKind := typeKey(types, neighbor.SubjectTypeID)
				if !promotePrimary(nKind, "provenencia") {
					continue
				}
				from, to, sig := canonSignature(id, selfKind, n, l, neighbor, props, kin)
				ek := string(from) + "|" + string(to) + "|" + sig.Key()
				if !edgeSeen[ek] {
					edgeSeen[ek] = true
					edges = append(edges, graphalign.CanonEdge{
						From: append([]byte(nil), from...), To: append([]byte(nil), to...), Signature: sig,
					})
				}
				if _, ok := handles[string(neighbor.ID)]; !ok {
					vals, err := handleValues(props, neighbor)
					if err != nil {
						return graphalign.Canon{}, err
					}
					handles[string(neighbor.ID)] = graphalign.Handle{
						ID: append([]byte(nil), neighbor.ID...), Ref: neighbor.Ref, Kind: nKind, Values: vals,
					}
					queue = append(queue, neighbor.ID)
				}
			}
		}
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

func canonSignature(self []byte, selfKind string, n *graphcache.Node, l graphcache.Link, neighbor *graphcache.Node, props map[string]match.Property, kin kinship) (from, to []byte, sig graphalign.EdgeSignature) {
	if l.FromEnd {
		from, to = self, l.Neighbor
	} else {
		from, to = l.Neighbor, self
	}
	b, _ := connectrules.LookupBridge(l.BridgeTypeKey)
	neighborKind := ""
	if len(b.Endpoints) == 2 {
		neighborKind = b.Endpoints[1].TypeKey
	}
	role := l.TermKey
	directed := false
	switch l.BridgeTypeKey {
	case "relationship":
		rel, flip := kin.canonical(connectrules.DisambiguationRelationshipType, role)
		if flip {
			from, to = to, from
			role = rel
		}
		directed = kin.directed[connectrules.DisambiguationRelationshipType+"|"+role]
	case "place_relationship":
		directed = kin.directed[connectrules.DisambiguationPlaceRelationshipType+"|"+role]
	}
	var typeTerm string
	if l.BridgeTypeKey == "participation" {
		event := neighbor
		if selfKind == "event" {
			event = n
		}
		if l.NeighborTypeKey == "event" {
			event = neighbor
		}
		typeTerm = keptTerm(event, props, neighborTypeProperty["event"])
	}
	sig = graphalign.EdgeSignature{
		BridgeType: l.BridgeTypeKey, RoleOrType: role,
		NeighborKind: neighborKind, NeighborTypeTerm: typeTerm, Directed: directed,
	}
	return from, to, sig
}

func keptTerm(n *graphcache.Node, props map[string]match.Property, key string) string {
	if n == nil || key == "" {
		return ""
	}
	for propID, rows := range n.Values {
		if props[propID].Key != key {
			continue
		}
		for _, row := range rows {
			if row.Rank == 1 && row.Reason == "kept" {
				return row.TermKey
			}
		}
	}
	return ""
}

func handleValues(props map[string]match.Property, n *graphcache.Node) (match.Values, error) {
	out := match.Values{}
	if n == nil {
		return out, nil
	}
	for propID, rows := range n.Values {
		prop, ok := props[propID]
		if !ok {
			continue
		}
		vals := make([]match.Value, 0, len(rows))
		for _, row := range rows {
			v := match.Value{
				Text: row.Text, HasText: row.HasText,
				Integer: row.Integer, HasInteger: row.HasInteger,
				Term: row.TermKey, TermID: append([]byte(nil), row.TermID...),
			}
			if len(row.Date) > 0 {
				d, err := valuecodec.UnmarshalDate(row.Date)
				if err != nil {
					return nil, err
				}
				v.Date = &d
			}
			if len(row.Name) > 0 {
				name, err := valuecodec.UnmarshalName(row.Name)
				if err != nil {
					return nil, err
				}
				v.Name = &name
			}
			vals = append(vals, v)
		}
		out[prop] = vals
	}
	return out, nil
}

func propertyByID(q Querier) (map[string]match.Property, error) {
	rows, err := q.Query(`SELECT id, key, origin FROM properties`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]match.Property{}
	for rows.Next() {
		var id []byte
		var p match.Property
		if err := rows.Scan(&id, &p.Key, &p.Origin); err != nil {
			return nil, err
		}
		out[string(id)] = p
	}
	return out, rows.Err()
}

func typeByKey(q Querier) (map[string][]byte, error) {
	rows, err := q.Query(`SELECT key, id FROM subject_types WHERE origin = 'provenencia'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string][]byte{}
	for rows.Next() {
		var key string
		var id []byte
		if err := rows.Scan(&key, &id); err != nil {
			return nil, err
		}
		out[key] = append([]byte(nil), id...)
	}
	return out, rows.Err()
}

func typeKey(types map[string][]byte, id []byte) string {
	for key, typeID := range types {
		if string(typeID) == string(id) {
			return key
		}
	}
	return ""
}
