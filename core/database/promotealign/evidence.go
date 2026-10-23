package promotealign

import (
	"bytes"
	"slices"

	"github.com/mendahu/provenencia/core/claimevidence"
	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/graphalign"
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
	SourceID, SubjectID, HandleID []byte
	Neighbors                     []graphalign.Fixed
}

// Evidence computes one row's claim evidence: a line per accepted member of
// the target entity for each of the row Subject's values, and a relationship
// for each link to a mapped neighbor that corresponds to a link between the
// entity's members and the neighbor entity's. It only reads the catalog and
// the graph cache, and runs no alignment; the decisions are
// claimevidence.Build's. An unknown Source, a row Subject that is not a primary
// Subject on it, an empty row handle, or a row or neighbor handle that is
// missing, merged or of another kind, or more neighbors than the Source has
// primary Subjects, is promote.ErrInvalid.
func Evidence(c *database.Catalog, req EvidenceRequest) (claimevidence.Evidence, error) {
	in, err := loadInputs(c, req.SourceID)
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
// Subject the Source already filed, on its handle: the row's own Subject is
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
	rowEG, err := g.EvidenceGraph(in.graphID)
	if err != nil {
		return claimevidence.Evidence{}, err
	}
	observationsOf := map[string][][]byte{}
	for _, o := range rowEG.Observations {
		observationsOf[string(o.SubjectID)] = append(observationsOf[string(o.SubjectID)], o.ID)
	}

	// Layer links of the row Subject, by layer neighbor.
	layerLinks := map[string][]claimevidence.LayerLink{}
	for _, b := range in.layer.Bridges {
		other, key, ok := b.From(row.SubjectID)
		if !ok {
			continue
		}
		link := claimevidence.LayerLink{Key: key}
		for _, a := range b.Associations {
			link.Observations = append(link.Observations, observationsOf[string(a)]...)
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
	var neighbors []claimevidence.Neighbor
	for _, n := range wanted {
		hid := placed[string(n)]
		hn := nodeOf[string(n)]
		if hn == nil {
			continue
		}
		canon, err := q.canonLinks(handleNode, hn)
		if err != nil {
			return claimevidence.Evidence{}, err
		}
		all, err := q.accepted(hid)
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
			Order: kindOrder,
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
	memberOf map[string]graphcache.Member // by Subject id, from the members read so far
}

func newEvidenceReader(g *graphcache.Graph, in inputs) (*evidenceReader, error) {
	prof, err := newProfiler(g, in.layer.Metas, in.cfg.Subject, in.namePattern)
	if err != nil {
		return nil, err
	}
	return &evidenceReader{g: g, in: in, prof: prof, memberOf: map[string]graphcache.Member{}}, nil
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
			eg, err := r.g.EvidenceGraph([]byte(graphID))
			if err != nil {
				return nil, err
			}
			cands, err := layerValues(eg, missing, r.prof.props, r.prof.terms)
			if err != nil {
				return nil, err
			}
			if memo == nil {
				memo = map[string]subjectcompare.Values{}
			}
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

// canonLinks are the filed links from one entity to another, each keyed from
// the first entity's end, with the accepted bridge Subjects filed on its
// association.
func (r *evidenceReader) canonLinks(from, to *graphcache.Node) ([]claimevidence.CanonLink, error) {
	var out []claimevidence.CanonLink
	fromKind := typeKey(r.prof.types, from.SubjectTypeID)
	for _, l := range from.Links {
		if !bytes.Equal(l.Neighbor, to.ID) {
			continue
		}
		a, _, keyA, keyB, ok := canonKeys(from.ID, fromKind, from, l, to, r.prof.props, r.in.terms)
		if !ok {
			continue
		}
		key := keyB
		if bytes.Equal(a, from.ID) {
			key = keyA
		}
		bridges, err := r.bridges(l)
		if err != nil {
			return nil, err
		}
		out = append(out, claimevidence.CanonLink{Key: key, Bridges: bridges})
	}
	return out, nil
}

// bridges are the accepted bridge Subjects on a link's association, each with
// the two Subjects it joins (read from its endpoint Observations, in the
// order of the link's end) and every Observation on it. An end is read from a
// positive Observation only; an end that names two different Subjects is
// unresolved, and a bridge that does not record both ends is left out.
func (r *evidenceReader) bridges(l graphcache.Link) ([]claimevidence.Bridge, error) {
	spec, ok := connectrules.LookupBridge(l.BridgeTypeKey)
	if !ok || len(spec.Endpoints) != 2 {
		return nil, nil
	}
	subjectIDs, err := r.accepted(l.Association)
	if err != nil {
		return nil, err
	}
	memberEnd, neighborEnd := spec.Endpoints[0].PropertyKey, spec.Endpoints[1].PropertyKey
	if !l.FromEnd {
		memberEnd, neighborEnd = neighborEnd, memberEnd
	}
	obsOf := map[string][]graphcache.EvidenceObservation{} // by bridge Subject
	read := map[string]bool{}                              // graphs walked
	for _, id := range subjectIDs {
		m, ok := r.memberOf[string(id)]
		if !ok || read[string(m.GraphID)] {
			continue
		}
		read[string(m.GraphID)] = true
		eg, err := r.g.EvidenceGraph(m.GraphID)
		if err != nil {
			return nil, err
		}
		for _, o := range eg.Observations {
			obsOf[string(o.SubjectID)] = append(obsOf[string(o.SubjectID)], o)
		}
	}
	var out []claimevidence.Bridge
	for _, id := range subjectIDs {
		m, ok := r.memberOf[string(id)]
		if !ok {
			continue
		}
		b := claimevidence.Bridge{Subject: bytes.Clone(id), Source: bytes.Clone(m.SourceID)}
		ambiguous := false
		end := func(into *[]byte, v []byte) {
			if len(*into) > 0 && !bytes.Equal(*into, v) {
				ambiguous = true
			}
			*into = bytes.Clone(v)
		}
		for _, o := range obsOf[string(id)] {
			b.Observations = append(b.Observations, bytes.Clone(o.ID))
			if len(o.ValueSubjectID) == 0 || o.Polarity != observations.PolarityPositive {
				continue
			}
			switch r.prof.props[string(o.PropertyID)].Key {
			case memberEnd:
				end(&b.Member, o.ValueSubjectID)
			case neighborEnd:
				end(&b.Neighbor, o.ValueSubjectID)
			}
		}
		if ambiguous || len(b.Member) == 0 || len(b.Neighbor) == 0 {
			continue
		}
		out = append(out, b)
	}
	return out, nil
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
		eg, err := r.g.EvidenceGraph(m.GraphID)
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
			add(b.Source, b.Bridge)
		}
		sections(rel.Sections)
	}
	return out, failed
}
