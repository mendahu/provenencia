package promotealign

import (
	"bytes"

	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/database/autoreconciler"
	"github.com/mendahu/provenencia/core/database/evidencelayer"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/valuerules"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/hops"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/subjectprofile"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// The canon side's profile builder: each handle is built here and held by the
// match index; the Evidence side is built by core/database/evidencelayer.

// reconciled is a cached handle value as the reconciler left it.
func reconciled(row graphcache.Value, terms map[string]graphcache.VocabTerm) (autoreconcile.ReconciledValue, error) {
	v, err := evidencelayer.Decode(row.Scalar, row.TermID, row.Date, row.Name, row.ImageFileID, terms)
	if err != nil {
		return autoreconcile.ReconciledValue{}, err
	}
	return autoreconcile.ReconciledValue{
		Value: v, Support: row.Support, Against: row.Against, Reason: autoreconcile.Reason(row.Reason),
	}, nil
}

// profiler is the one place that says what a handle's profile reads, in what
// order profiles are built, and who must be rebuilt when a read input changes.
// It holds the vocabulary, the value rules and the alignment config a profile is
// built under, and stores nothing: the match index holds the profiles.
type profiler struct {
	g           *graphcache.Graph
	props       map[string]graphcache.VocabProperty
	terms       map[string]graphcache.VocabTerm
	metas       subjectcompare.Metas
	types       map[string][]byte // subject type key -> id
	typeKeys    map[string]string // subject type id -> key
	reconcile   valuecompare.Rules
	namePattern string
	cfg         graphalign.Config
}

// newProfiler reads the alignment config, the value metas and the vocabulary a
// profile is built under.
func newProfiler(g *graphcache.Graph) (*profiler, error) {
	cfg, namePattern, err := loadConfig(g)
	if err != nil {
		return nil, err
	}
	metas, err := valuerules.Metas(g)
	if err != nil {
		return nil, err
	}
	vocab, err := g.Vocabulary()
	if err != nil {
		return nil, err
	}
	return &profiler{
		g: g, props: vocab.PropertiesByID(), terms: vocab.TermsByID(), metas: metas,
		types: vocab.TypeIDsByKey(), typeKeys: vocab.TypeKeysByID(),
		reconcile: autoreconciler.ValueRules, namePattern: namePattern, cfg: cfg,
	}, nil
}

func (p *profiler) kind(n *graphcache.Node) vocabkey.SubjectType {
	return vocabkey.SubjectType(p.typeKeys[string(n.SubjectTypeID)])
}

// holds reports whether the match index holds a handle's profile: a node that
// is present, not merged, and of a primary kind.
func (p *profiler) holds(n *graphcache.Node) bool {
	return n != nil && !n.Merged && vocabkey.IsPrimaryKind(string(p.kind(n)))
}

// derivesLife reports whether a kind reads its life from the events it takes
// part in (a Lifespans rule): its profile reads other handles' profiles.
func (p *profiler) derivesLife(kind vocabkey.SubjectType) bool {
	return len(p.cfg.Subject.Lifespans[kind]) > 0
}

// subjects builds the nodes' profiles into built, overwriting a held one. A nil
// or merged node is skipped. The nodes whose kind reads no other profile are
// built first, so a kind that derives its life reads the events built now. A
// profile it reads that is not in built is skipped.
func (p *profiler) subjects(nodes []*graphcache.Node, built map[string]subjectcompare.Subject) error {
	queued := map[string]bool{}
	var plain, derived []*graphcache.Node
	for _, n := range nodes {
		if n == nil || n.Merged || queued[string(n.ID)] {
			continue
		}
		queued[string(n.ID)] = true
		if p.derivesLife(p.kind(n)) {
			derived = append(derived, n)
		} else {
			plain = append(plain, n)
		}
	}
	for _, group := range [][]*graphcache.Node{plain, derived} {
		for _, n := range group {
			s, err := p.subject(n, built)
			if err != nil {
				return err
			}
			built[string(n.ID)] = s
		}
	}
	return nil
}

// readers is the ids, deduplicated in first-seen order, whose profile reads
// the profile of one of ids: for each id whose kind derives no life, the
// neighbors across hops.ParticipantsOfEvent whose kind does. It is the inverse
// of the read subject makes.
func (p *profiler) readers(ids [][]byte) ([][]byte, error) {
	var out [][]byte
	seen := map[string]bool{}
	for _, id := range ids {
		n, err := p.g.Node(id)
		if err != nil {
			return nil, err
		}
		if n == nil || p.derivesLife(p.kind(n)) {
			continue
		}
		links, err := p.g.Follow(id, hops.ParticipantsOfEvent)
		if err != nil {
			return nil, err
		}
		for _, l := range links {
			if !p.derivesLife(vocabkey.SubjectType(l.NeighborTypeKey)) || seen[string(l.Neighbor)] {
				continue
			}
			seen[string(l.Neighbor)] = true
			out = append(out, l.Neighbor)
		}
	}
	return out, nil
}

// entity is a node's reconciled properties as the profile builder reads them,
// with no life.
func (p *profiler) entity(n *graphcache.Node) (subjectprofile.Entity, error) {
	props := map[vocabkey.PropertyKey]subjectprofile.Property{}
	for propID, rows := range n.Values {
		prop, ok := p.props[propID]
		if !ok || prop.Origin != vocabkey.OriginProvenencia {
			continue
		}
		vals := make([]autoreconcile.ReconciledValue, 0, len(rows))
		for _, row := range rows {
			rv, err := reconciled(row, p.terms)
			if err != nil {
				return subjectprofile.Entity{}, err
			}
			vals = append(vals, rv)
		}
		props[vocabkey.PropertyKey(prop.Key)] = subjectprofile.Property{Values: vals}
	}
	return subjectprofile.Entity{Properties: props, NamePattern: p.namePattern}, nil
}

// subject makes one node's profile. A kind that derives its life reads it from
// the profiles in built of the events it takes part in.
func (p *profiler) subject(n *graphcache.Node, built map[string]subjectcompare.Subject) (subjectcompare.Subject, error) {
	kind := p.kind(n)
	e, err := p.entity(n)
	if err != nil {
		return subjectcompare.Subject{}, err
	}
	if p.derivesLife(kind) {
		links, err := p.g.Follow(n.ID, hops.EventsOfParticipant)
		if err != nil {
			return subjectcompare.Subject{}, err
		}
		for _, l := range links {
			if ev, ok := built[string(l.Neighbor)]; ok {
				e.Life = append(e.Life, subjectprofile.LifeEventOf(ev.Values, vocabkey.Role(l.TermKey)))
			}
		}
	}
	vals, err := subjectprofile.Build(e, p.metas, p.reconcile, p.cfg.Subject.Lifespans[kind], subjectprofile.Options{})
	if err != nil {
		return subjectcompare.Subject{}, err
	}
	return subjectcompare.Subject{ID: bytes.Clone(n.ID), Ref: n.Ref, Kind: string(kind), Values: vals}, nil
}
