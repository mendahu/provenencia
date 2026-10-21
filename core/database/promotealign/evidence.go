package promotealign

import (
	"bytes"

	"github.com/mendahu/provenencia/core/claimevidence"
	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/matchregistry"
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
// missing, merged or of another kind is promote.ErrInvalid.
func Evidence(c *database.Catalog, req EvidenceRequest) (claimevidence.Evidence, error) {
	in, err := loadInputs(c, req.SourceID)
	if err != nil {
		return claimevidence.Evidence{}, err
	}
	row, err := checkFixed(in.q, in.primary, graphalign.Fixed{SubjectID: req.SubjectID, HandleID: req.HandleID})
	if err != nil {
		return claimevidence.Evidence{}, err
	}
	placed := map[string][]byte{} // layer neighbor → its handle
	var order [][]byte
	for _, f := range req.Neighbors {
		nf, err := checkFixed(in.q, in.primary, f)
		if err != nil {
			return claimevidence.Evidence{}, err
		}
		if len(nf.HandleID) != 16 {
			continue
		}
		if _, dup := placed[string(nf.SubjectID)]; !dup {
			order = append(order, nf.SubjectID)
		}
		placed[string(nf.SubjectID)] = nf.HandleID
	}

	g := c.Graph()
	q, err := newEvidenceReader(g, in)
	if err != nil {
		return claimevidence.Evidence{}, err
	}
	valuesOf := map[string]subjectcompare.Values{}
	for _, s := range in.layer.Subjects {
		valuesOf[string(s.ID)] = s.Values
	}
	rowSG, err := g.Source(req.SourceID)
	if err != nil {
		return claimevidence.Evidence{}, err
	}
	observationsOf := map[string][][]byte{}
	for _, o := range rowSG.Observations {
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
	}
	if err := q.prof.profile(nodes); err != nil {
		return claimevidence.Evidence{}, err
	}

	members, err := q.members(row.HandleID)
	if err != nil {
		return claimevidence.Evidence{}, err
	}
	var neighbors []claimevidence.Neighbor
	for _, n := range wanted {
		hid := placed[string(n)]
		hn, err := g.Node(hid)
		if err != nil {
			return claimevidence.Evidence{}, err
		}
		if hn == nil {
			continue
		}
		nm, err := q.members(hid)
		if err != nil {
			return claimevidence.Evidence{}, err
		}
		canon, err := q.canonLinks(handleNode, hn)
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

	out := claimevidence.Build(vocabkey.SubjectType(in.primary[string(row.SubjectID)].kind), valuesOf[string(row.SubjectID)],
		q.prof.values(row.HandleID), members, neighbors, in.layer.Metas, matchregistry.Default().Evidence,
		in.cfg.Segment, in.cfg.Subject, in.cfg.Value)
	out.Sources = q.titles(out)
	return out, nil
}

// evidenceReader reads what the query needs beyond the layer: members and
// their values, the Source each Subject is on, and the filed links between
// entities. Every read is cached by the graph or kept here.
type evidenceReader struct {
	g       *graphcache.Graph
	in      inputs
	prof    *profiler
	sources map[string]subjects.SourceOf // by Subject id
	titleOf map[string]string            // by Source id
}

func newEvidenceReader(g *graphcache.Graph, in inputs) (*evidenceReader, error) {
	prof, err := newProfiler(g, in.layer.Metas, in.cfg.Subject, in.namePattern)
	if err != nil {
		return nil, err
	}
	return &evidenceReader{g: g, in: in, prof: prof, sources: map[string]subjects.SourceOf{}, titleOf: map[string]string{}}, nil
}

// sourcesOf fills the Source of every listed Subject not read yet.
func (r *evidenceReader) sourcesOf(ids [][]byte) error {
	var missing [][]byte
	for _, id := range ids {
		if _, ok := r.sources[string(id)]; !ok {
			missing = append(missing, id)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	got, err := subjects.SourcesOf(r.in.q, missing)
	if err != nil {
		return err
	}
	for id, s := range got {
		r.sources[id] = s
		r.titleOf[string(s.SourceID)] = s.Title
	}
	return nil
}

// members are the accepted members of an entity, in claim order, each with its
// values as its own Source records them.
func (r *evidenceReader) members(entity []byte) ([]claimevidence.Member, error) {
	ids, err := acceptedMembers(r.g, [][]byte{entity})
	if err != nil {
		return nil, err
	}
	subjectIDs := ids[string(entity)]
	if err := r.sourcesOf(subjectIDs); err != nil {
		return nil, err
	}
	var out []claimevidence.Member
	for _, id := range subjectIDs {
		src, ok := r.sources[string(id)]
		if !ok {
			continue
		}
		sg, err := r.g.Source(src.SourceID)
		if err != nil {
			return nil, err
		}
		cands, err := layerValues(sg, map[string]bool{string(id): true}, r.prof.props, r.prof.terms)
		if err != nil {
			return nil, err
		}
		out = append(out, claimevidence.Member{
			Subject: bytes.Clone(id),
			Source:  bytes.Clone(src.SourceID),
			Values: subjectprofile.ForSubject(subjectprofile.Subject{
				Values: cands[string(id)], NamePattern: r.in.namePattern,
			}, nil),
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
// order of the link's end) and every Observation on it. A bridge that does not
// record both ends is left out.
func (r *evidenceReader) bridges(l graphcache.Link) ([]claimevidence.Bridge, error) {
	spec, ok := connectrules.LookupBridge(l.BridgeTypeKey)
	if !ok || len(spec.Endpoints) != 2 {
		return nil, nil
	}
	ids, err := acceptedMembers(r.g, [][]byte{l.Association})
	if err != nil {
		return nil, err
	}
	subjectIDs := ids[string(l.Association)]
	if err := r.sourcesOf(subjectIDs); err != nil {
		return nil, err
	}
	memberEnd, neighborEnd := spec.Endpoints[0].PropertyKey, spec.Endpoints[1].PropertyKey
	if !l.FromEnd {
		memberEnd, neighborEnd = neighborEnd, memberEnd
	}
	var out []claimevidence.Bridge
	for _, id := range subjectIDs {
		src, ok := r.sources[string(id)]
		if !ok {
			continue
		}
		sg, err := r.g.Source(src.SourceID)
		if err != nil {
			return nil, err
		}
		b := claimevidence.Bridge{Subject: bytes.Clone(id), Source: bytes.Clone(src.SourceID)}
		for _, o := range sg.Observations {
			if !bytes.Equal(o.SubjectID, id) {
				continue
			}
			b.Observations = append(b.Observations, bytes.Clone(o.ID))
			if len(o.ValueSubjectID) == 0 {
				continue
			}
			switch r.prof.props[string(o.PropertyID)].Key {
			case memberEnd:
				b.Member = bytes.Clone(o.ValueSubjectID)
			case neighborEnd:
				b.Neighbor = bytes.Clone(o.ValueSubjectID)
			}
		}
		if len(b.Member) == 0 || len(b.Neighbor) == 0 {
			continue
		}
		out = append(out, b)
	}
	return out, nil
}

// titles maps the id of every Source a pin or bundle in the evidence names to
// its title.
func (r *evidenceReader) titles(e claimevidence.Evidence) map[string]string {
	out := map[string]string{}
	add := func(id []byte) {
		if t, ok := r.titleOf[string(id)]; ok {
			out[string(id)] = t
		}
	}
	var sections func(ss []claimevidence.Section)
	sections = func(ss []claimevidence.Section) {
		for _, s := range ss {
			for _, l := range s.Lines {
				for _, p := range l.Members {
					add(p.Source)
				}
			}
		}
	}
	sections(e.Sections)
	for _, rel := range e.Relationships {
		for _, b := range rel.Members {
			add(b.Source)
		}
		sections(rel.Sections)
	}
	return out
}
