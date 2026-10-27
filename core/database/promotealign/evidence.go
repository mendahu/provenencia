package promotealign

import (
	"bytes"
	"maps"
	"slices"

	"github.com/mendahu/provenencia/core/claimevidence"
	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonlinks"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/kinship"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/segmentcompare"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/subjectprofile"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// EvidenceRequest names one Promote row's target and the targets the page
// holds for its neighbors: the inputs of the claim evidence query. A neighbor
// whose Target is New or Skip, or whose Subject is not a layer neighbor of the
// row's Subject, contributes nothing.
type EvidenceRequest struct {
	GraphID, SubjectID, HandleID []byte
	Neighbors                    []graphalign.Fixed
}

// Evidence computes one row's claim evidence: a line per accepted member of
// the target entity for each of the row Subject's values, and a relationship
// for each link to a mapped neighbor that corresponds to a link between the
// entity's members and the neighbor entity's. It only reads the catalog and
// the graph cache, and runs no alignment; the decisions are
// claimevidence.Build's. An unknown graph, a row Subject that is not a primary
// Subject on it, an empty row handle, or a row or neighbor handle that is
// missing, merged or of another kind, or more neighbors than the graph has
// primary Subjects, is promote.ErrInvalid.
func Evidence(c *database.Catalog, req EvidenceRequest) (claimevidence.Evidence, error) {
	in, err := loadInputs(c, req.GraphID)
	if err != nil {
		return claimevidence.Evidence{}, err
	}
	g := c.Graph()
	r, err := newEvidenceReader(g, in)
	if err != nil {
		return claimevidence.Evidence{}, err
	}
	return evidence(g, in, r, req)
}

// evidence is Evidence over inputs and a reader the caller loaded, so a page
// of rows loads once. The neighbors are the request's, de-duplicated (a
// Subject named twice keeps its last target at its first position), plus every
// Subject the graph already filed, on its handle: the row's own Subject is
// never its own neighbor.
func evidence(g *graphcache.Graph, in inputs, q *evidenceReader, req EvidenceRequest) (claimevidence.Evidence, error) {
	row, err := checkFixed(g, in.primary, graphalign.Fixed{SubjectID: req.SubjectID, HandleID: req.HandleID})
	if err != nil {
		return claimevidence.Evidence{}, err
	}
	if len(req.Neighbors) > len(in.primary) {
		return claimevidence.Evidence{}, promote.ErrInvalid
	}
	var named []graphalign.Fixed
	at := map[string]int{}
	for _, f := range req.Neighbors {
		if i, dup := at[string(f.SubjectID)]; dup {
			named[i] = f
			continue
		}
		at[string(f.SubjectID)] = len(named)
		named = append(named, f)
	}
	merged, err := mergeFixed(g, in.graphID, in.primary, named)
	if err != nil {
		return claimevidence.Evidence{}, err
	}
	placed := map[string][]byte{} // layer neighbor → its handle
	var order [][]byte
	for _, nf := range merged {
		if len(nf.HandleID) != 16 || bytes.Equal(nf.SubjectID, row.SubjectID) {
			continue
		}
		order = append(order, nf.SubjectID)
		placed[string(nf.SubjectID)] = nf.HandleID
	}

	valuesOf := map[string]subjectcompare.Values{}
	for _, s := range in.layer.Subjects {
		valuesOf[string(s.ID)] = s.Values
	}
	rowIx, err := q.graphIndex(in.graphID)
	if err != nil {
		return claimevidence.Evidence{}, err
	}

	// Layer links of the row Subject, by layer neighbor.
	layerLinks := map[string][]claimevidence.LayerLink{}
	for _, b := range in.layer.Bridges {
		other, key, ok := b.From(row.SubjectID)
		if !ok {
			continue
		}
		link := claimevidence.LayerLink{Key: key, Recording: b.Recording}
		for _, sh := range b.Recording.Shapes {
			if info, ok := in.primary[string(sh.Pivot)]; ok {
				link.Pivots = append(link.Pivots, claimevidence.Pivot{
					Subject: sh.Pivot, Kind: vocabkey.SubjectType(info.kind), Values: valuesOf[string(sh.Pivot)],
				})
			}
		}
		for _, a := range b.Associations {
			for _, o := range rowIx.obs[string(a)] {
				link.Observations = append(link.Observations, o.ID)
			}
		}
		layerLinks[string(other)] = append(layerLinks[string(other)], link)
	}

	handleNode, err := g.Node(row.HandleID)
	if err != nil {
		return claimevidence.Evidence{}, err
	}
	if handleNode == nil {
		return claimevidence.Evidence{Sources: map[string]string{}}, nil
	}
	nodes := []*graphcache.Node{handleNode}
	var wanted [][]byte
	nodeOf := map[string]*graphcache.Node{} // neighbor's handle node, by layer neighbor
	for _, n := range order {
		if len(layerLinks[string(n)]) == 0 {
			continue
		}
		wanted = append(wanted, n)
		hn, err := g.Node(placed[string(n)])
		if err != nil {
			return claimevidence.Evidence{}, err
		}
		nodes = append(nodes, hn)
		nodeOf[string(n)] = hn
	}
	if err := q.prof.profile(nodes); err != nil {
		return claimevidence.Evidence{}, err
	}
	q.prof.keep()

	members, err := q.members(row.HandleID)
	if err != nil {
		return claimevidence.Evidence{}, err
	}
	// Kinship links have only person ends, so an event or place row reads none.
	var rels []kinship.Link
	if in.primary[string(row.SubjectID)].kind == string(vocabkey.Person) {
		if rels, err = q.kin.Relationships(row.HandleID); err != nil {
			return claimevidence.Evidence{}, err
		}
	}
	rowSet := map[string]bool{}
	for _, m := range members {
		rowSet[string(m.Subject)] = true
	}
	var neighbors []claimevidence.Neighbor
	for _, n := range wanted {
		hid := placed[string(n)]
		hn := nodeOf[string(n)]
		if hn == nil {
			continue
		}
		all, err := q.accepted(hid)
		if err != nil {
			return claimevidence.Evidence{}, err
		}
		nbrSet := make(map[string]bool, len(all))
		for _, id := range all {
			nbrSet[string(id)] = true
		}
		canon, err := q.canonLinks(handleNode, hn, rels, rowSet, nbrSet)
		if err != nil {
			return claimevidence.Evidence{}, err
		}
		nm, err := q.memberValues(bridgedOnly(all, canon))
		if err != nil {
			return claimevidence.Evidence{}, err
		}
		neighbors = append(neighbors, claimevidence.Neighbor{
			Subject: n,
			Kind:    vocabkey.SubjectType(in.primary[string(n)].kind),
			Values:  valuesOf[string(n)],
			Handle:  q.prof.values(hid),
			Members: nm,
			Layer:   layerLinks[string(n)],
			Canon:   canon,
		})
	}

	kindOrder, err := bindingOrder(g)
	if err != nil {
		return claimevidence.Evidence{}, err
	}
	out := claimevidence.Build(vocabkey.SubjectType(in.primary[string(row.SubjectID)].kind), valuesOf[string(row.SubjectID)],
		q.prof.values(row.HandleID), members, neighbors, in.layer.Metas, matchregistry.Default().Evidence,
		in.cfg.Segment, in.cfg.Subject, in.cfg.Value, claimevidence.Inputs{
			Order:  kindOrder,
			Placed: placed,
		})
	for i := range out.Relationships {
		readFromNeighbor(&out.Relationships[i], in.terms)
	}
	if out.Sources, err = q.titles(out); err != nil {
		return claimevidence.Evidence{}, err
	}
	return out, nil
}

// readFromNeighbor fills a relationship's neighbor-side reading from its key,
// with the bridge type's endpoints in connect-rule order. A bridge type the
// rules do not know is left unread.
func readFromNeighbor(r *claimevidence.Relationship, terms segmentcompare.Terms) {
	b, ok := connectrules.LookupBridge(string(r.Key.Type))
	if !ok || len(b.Endpoints) != 2 {
		return
	}
	r.NeighborEnd, r.NeighborTerm, r.NeighborTermNamesNeighbor = segmentcompare.ReadFrom(
		r.Key, vocabkey.PropertyKey(b.Endpoints[0].PropertyKey), vocabkey.PropertyKey(b.Endpoints[1].PropertyKey), terms,
	)
	r.NeighborTermProperty = vocabkey.PropertyKey(b.Disambiguation)
}

// bindingOrder is each kind's bound properties in the catalog's order, from
// the cached vocabulary.
func bindingOrder(g *graphcache.Graph) (map[vocabkey.SubjectType][]vocabkey.PropertyKey, error) {
	v, err := g.Vocabulary()
	if err != nil {
		return nil, err
	}
	return v.BindingOrder(), nil
}

// acceptedClaims is a handle's accepted members, each with its Source and
// Evidence graph, in claim id order. A handle with no node has none.
func acceptedClaims(g *graphcache.Graph, entityID []byte) ([]graphcache.Member, error) {
	n, err := g.Node(entityID)
	if err != nil || n == nil {
		return nil, err
	}
	var out []graphcache.Member
	for _, m := range n.Members {
		if m.Accepted {
			out = append(out, m)
		}
	}
	return out, nil
}

// evidenceReader reads what the query needs beyond the layer: members and
// their values, the Source and graph each member Subject is on, and the filed
// links between entities. Every read is cached by the graph or kept here.
type evidenceReader struct {
	g        *graphcache.Graph
	in       inputs
	prof     *profiler
	kin      *canonlinks.Reader                  // the entities' recorded and implied relationship links
	memberOf map[string]graphcache.Member        // by Subject id, from the members read so far
	graphs   map[string]graphcache.EvidenceGraph // by graph id, the graphs this request has read
	indexes  map[string]*graphIndex              // by graph id, the graphs the request has read
}

func newEvidenceReader(g *graphcache.Graph, in inputs) (*evidenceReader, error) {
	prof, err := newProfiler(g, in.layer.Metas, in.cfg.Subject, in.namePattern)
	if err != nil {
		return nil, err
	}
	kin, err := canonlinks.NewReader(g, matchregistry.Default().Kinship)
	if err != nil {
		return nil, err
	}
	return &evidenceReader{g: g, in: in, prof: prof, kin: kin, memberOf: map[string]graphcache.Member{}, graphs: map[string]graphcache.EvidenceGraph{}, indexes: map[string]*graphIndex{}}, nil
}

// graph is the Evidence graph with id. The request keeps each one it reads, so
// a graph the cache evicts mid-request is not read again.
func (r *evidenceReader) graph(id []byte) (graphcache.EvidenceGraph, error) {
	if eg, ok := r.graphs[string(id)]; ok {
		return eg, nil
	}
	eg, err := r.g.EvidenceGraph(id)
	if err != nil {
		return graphcache.EvidenceGraph{}, err
	}
	r.graphs[string(id)] = eg
	return eg, nil
}

// accepted are the accepted member Subjects of an entity, in claim order. It
// remembers each one's Source and graph, which a node's members carry.
func (r *evidenceReader) accepted(entity []byte) ([][]byte, error) {
	ms, err := acceptedClaims(r.g, entity)
	if err != nil {
		return nil, err
	}
	ids := make([][]byte, len(ms))
	for i, m := range ms {
		r.memberOf[string(m.SubjectID)] = m
		ids[i] = m.SubjectID
	}
	return ids, nil
}

// bridgedOnly keeps, in order, the Subjects some link's bridge reaches as its
// neighbor end. The evidence reads only those members of a neighbor entity.
func bridgedOnly(ids [][]byte, links []claimevidence.CanonLink) [][]byte {
	var out [][]byte
	for _, id := range ids {
		if slices.ContainsFunc(links, func(l claimevidence.CanonLink) bool {
			return slices.ContainsFunc(l.Bridges, func(b claimevidence.Bridge) bool { return bytes.Equal(b.Neighbor, id) })
		}) {
			out = append(out, id)
		}
	}
	return out
}

// members are the accepted members of an entity, in claim order, each with its
// values as its own Source records them.
func (r *evidenceReader) members(entity []byte) ([]claimevidence.Member, error) {
	ids, err := r.accepted(entity)
	if err != nil {
		return nil, err
	}
	return r.memberValues(ids)
}

// membersMemoKey is the Evidence graph memo key of the member values already
// built for a graph's Subjects, with the name pattern they were built under after it.
const membersMemoKey = "promotealign.members\x00"

// memberMemo is an Evidence graph's memoized member values by Subject id,
// filled as Subjects are read. The values are shared and never changed.
func (r *evidenceReader) memberMemo(graphID []byte, key string) (map[string]subjectcompare.Values, bool) {
	v, ok := r.g.EvidenceGraphMemo(graphID, key)
	if !ok {
		return nil, false
	}
	m, ok := v.(map[string]subjectcompare.Values)
	return m, ok
}

// memberValues reads the listed Subjects as members, in the order given. Each
// Evidence graph is read once, for the Subjects of it not memoized yet.
func (r *evidenceReader) memberValues(subjectIDs [][]byte) ([]claimevidence.Member, error) {
	byGraph := map[string]map[string]bool{}
	for _, id := range subjectIDs {
		if m, ok := r.memberOf[string(id)]; ok {
			if byGraph[string(m.GraphID)] == nil {
				byGraph[string(m.GraphID)] = map[string]bool{}
			}
			byGraph[string(m.GraphID)][string(id)] = true
		}
	}
	key := membersMemoKey + r.in.namePattern
	valuesOf := map[string]subjectcompare.Values{}
	for graphID, set := range byGraph {
		memo, _ := r.memberMemo([]byte(graphID), key)
		var missing map[string]bool
		for id := range set {
			if _, ok := memo[id]; !ok {
				if missing == nil {
					missing = map[string]bool{}
				}
				missing[id] = true
			}
		}
		if len(missing) > 0 {
			eg, err := r.graph([]byte(graphID))
			if err != nil {
				return nil, err
			}
			cands, err := layerValues(eg, missing, r.prof.props, r.prof.terms)
			if err != nil {
				return nil, err
			}
			stored := make(map[string]subjectcompare.Values, len(memo)+len(missing))
			maps.Copy(stored, memo)
			memo = stored
			for id := range missing {
				memo[id] = subjectprofile.ForSubject(subjectprofile.Subject{
					Values: cands[id], NamePattern: r.in.namePattern,
				}, nil)
			}
			r.g.SetEvidenceGraphMemo([]byte(graphID), key, memo)
		}
		for id := range set {
			valuesOf[id] = memo[id]
		}
	}
	var out []claimevidence.Member
	for _, id := range subjectIDs {
		m, ok := r.memberOf[string(id)]
		if !ok {
			continue
		}
		out = append(out, claimevidence.Member{
			Subject: bytes.Clone(id),
			Source:  bytes.Clone(m.SourceID),
			Values:  valuesOf[string(id)],
		})
	}
	return out, nil
}

// canonLinks are the links from one entity to another, each keyed from the
// first entity's end, with the accepted bridge Subjects behind it. row and nbr
// are the Subject ids of the two entities' accepted members.
func (r *evidenceReader) canonLinks(from, to *graphcache.Node, rels []kinship.Link, row, nbr map[string]bool) ([]claimevidence.CanonLink, error) {
	var out []claimevidence.CanonLink
	for _, l := range from.Links {
		if !bytes.Equal(l.Neighbor, to.ID) || l.BridgeTypeKey == vocabkey.BridgeRelationship {
			continue
		}
		a, _, keyA, keyB, ok := canonlinks.EdgeOf(from.ID, from, l, to, r.prof.eventTypeProp, r.in.terms)
		if !ok {
			continue
		}
		key := keyB
		if bytes.Equal(a, from.ID) {
			key = keyA
		}
		bridges, err := r.bridgesOf([][]byte{l.Association}, row, nbr, nil)
		if err != nil {
			return nil, err
		}
		out = append(out, claimevidence.CanonLink{Key: key, Bridges: bridges})
	}
	for _, l := range rels {
		other, key, ok := l.From(from.ID)
		if !ok || !bytes.Equal(other, to.ID) {
			continue
		}
		pivot := map[string]bool{}
		for _, sh := range l.Recording.Shapes {
			ids, err := r.accepted(sh.Pivot)
			if err != nil {
				return nil, err
			}
			for _, id := range ids {
				pivot[string(id)] = true
			}
		}
		bridges, err := r.bridgesOf(l.Associations, row, nbr, pivot)
		if err != nil {
			return nil, err
		}
		pivots, err := r.pivotsOf(bridges)
		if err != nil {
			return nil, err
		}
		out = append(out, claimevidence.CanonLink{Key: key, Bridges: bridges, Recording: l.Recording, Pivots: pivots})
	}
	return out, nil
}

// pivotsOf are the pivot members the bridges reach, each once, with their kind
// and values.
func (r *evidenceReader) pivotsOf(bridges []claimevidence.Bridge) ([]claimevidence.Pivot, error) {
	var ids [][]byte
	for _, b := range bridges {
		if b.Pivot != nil && !slices.ContainsFunc(ids, func(x []byte) bool { return bytes.Equal(x, b.Pivot) }) {
			ids = append(ids, b.Pivot)
		}
	}
	if len(ids) == 0 {
		return nil, nil
	}
	ms, err := r.memberValues(ids)
	if err != nil {
		return nil, err
	}
	out := make([]claimevidence.Pivot, 0, len(ms))
	for _, m := range ms {
		ix, err := r.graphIndex(r.memberOf[string(m.Subject)].GraphID)
		if err != nil {
			return nil, err
		}
		out = append(out, claimevidence.Pivot{Subject: m.Subject, Kind: vocabkey.SubjectType(ix.types[string(m.Subject)]), Values: m.Values})
	}
	return out, nil
}

// bridgesOf are the accepted bridge Subjects on the associations, each with the
// Subjects it joins and every Observation on it. A bridge's two ends are read
// through its own bridge type's endpoint properties from positive Observations
// only; an end that names two different Subjects is unresolved and the bridge
// is left out. An end that is a Subject in row is the bridge's Member, else in
// nbr its Neighbor, else in pivot its Pivot; a bridge that does not reach two of
// the three is left out.
func (r *evidenceReader) bridgesOf(assocs [][]byte, row, nbr, pivot map[string]bool) ([]claimevidence.Bridge, error) {
	var out []claimevidence.Bridge
	for _, assoc := range assocs {
		subjectIDs, err := r.accepted(assoc)
		if err != nil {
			return nil, err
		}
		for _, id := range subjectIDs {
			m, ok := r.memberOf[string(id)]
			if !ok {
				continue
			}
			ix, err := r.graphIndex(m.GraphID)
			if err != nil {
				return nil, err
			}
			spec, ok := connectrules.LookupBridge(ix.types[string(id)])
			if !ok || len(spec.Endpoints) != 2 {
				continue
			}
			b := claimevidence.Bridge{Subject: bytes.Clone(id), Source: bytes.Clone(m.SourceID)}
			var ends [2][]byte
			ambiguous := false
			for _, o := range ix.obs[string(id)] {
				b.Observations = append(b.Observations, bytes.Clone(o.ID))
				if len(o.ValueSubjectID) == 0 || o.Polarity != observations.PolarityPositive {
					continue
				}
				key := r.prof.props[string(o.PropertyID)].Key
				for i, e := range spec.Endpoints {
					if key != e.PropertyKey {
						continue
					}
					if len(ends[i]) > 0 && !bytes.Equal(ends[i], o.ValueSubjectID) {
						ambiguous = true
					}
					ends[i] = o.ValueSubjectID
				}
			}
			if ambiguous || len(ends[0]) == 0 || len(ends[1]) == 0 {
				continue
			}
			reached := 0
			for _, end := range ends {
				var into *[]byte
				switch {
				case row[string(end)]:
					into = &b.Member
				case nbr[string(end)]:
					into = &b.Neighbor
				case pivot[string(end)]:
					into = &b.Pivot
				default:
					continue
				}
				if *into != nil {
					ambiguous = true
				}
				*into = bytes.Clone(end)
				reached++
			}
			if ambiguous || reached < 2 {
				continue
			}
			out = append(out, b)
		}
	}
	return out, nil
}

// graphIndex is a graph's Subjects by type and Observations by Subject, indexed
// once per request.
type graphIndex struct {
	types map[string]string
	obs   map[string][]graphcache.EvidenceObservation
}

// graphIndex indexes the Evidence graph with id once per request.
func (r *evidenceReader) graphIndex(graphID []byte) (*graphIndex, error) {
	if ix, ok := r.indexes[string(graphID)]; ok {
		return ix, nil
	}
	eg, err := r.graph(graphID)
	if err != nil {
		return nil, err
	}
	ix := &graphIndex{
		types: make(map[string]string, len(eg.Subjects)),
		obs:   make(map[string][]graphcache.EvidenceObservation),
	}
	for _, s := range eg.Subjects {
		ix.types[string(s.ID)] = s.TypeKey
	}
	for _, o := range eg.Observations {
		ix.obs[string(o.SubjectID)] = append(ix.obs[string(o.SubjectID)], o)
	}
	r.indexes[string(graphID)] = ix
	return ix, nil
}

// titles maps the id of every Source a pin or bundle in the evidence names to
// its title, from the cached graph of the member that names it.
func (r *evidenceReader) titles(e claimevidence.Evidence) (map[string]string, error) {
	out := map[string]string{}
	var failed error
	add := func(source, subject []byte) {
		if _, done := out[string(source)]; done || failed != nil {
			return
		}
		m, ok := r.memberOf[string(subject)]
		if !ok {
			return
		}
		eg, err := r.graph(m.GraphID)
		if err != nil {
			failed = err
			return
		}
		if eg.Found {
			out[string(source)] = eg.SourceTitle
		}
	}
	var sections func(ss []claimevidence.Section)
	sections = func(ss []claimevidence.Section) {
		for _, s := range ss {
			for _, l := range s.Lines {
				for _, p := range l.Members {
					add(p.Source, p.Subject)
				}
			}
		}
	}
	sections(e.Sections)
	for _, rel := range e.Relationships {
		for _, b := range rel.Members {
			add(b.Source, b.Bridges[0])
		}
		sections(rel.Sections)
	}
	return out, failed
}
