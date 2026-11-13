// Package promotealign loads catalog inputs for Promote graph alignment and
// calls pure graphalign.Align (S9-42). It does not score pairs or walk policy.
package promotealign

import (
	"bytes"
	"sort"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonlinks"
	"github.com/mendahu/provenencia/core/database/evidencelayer"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/graphfill"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/valuerules"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/segmentcompare"
)

// Propose loads the Evidence layer of graphID, walks the catalog's canonical
// graph from the fixed anchors and each unfixed kind's top property matches,
// and runs graphalign.Align. The int64 is MAX(audit_transactions.revision) at
// load time, the stamp Done sends back. only, when set, is the one Subject
// the page is open on: every other unpromoted Subject is off the page, as
// Done will not send it, and the proposal's connections wait for it. An
// unknown or empty graphID is promote.ErrInvalid. A graph read error is
// returned instead of a proposal.
func Propose(c *database.Catalog, graphID []byte, fixed []graphalign.Fixed, only []byte) (graphalign.Proposal, int64, error) {
	prop, rev, _, err := propose(c, graphID, fixed, only)
	return prop, rev, err
}

// propose is Propose, and the profiler it used, for a test that reads which
// handles were profiled. The profiler is nil when the graph has no Subject to
// place.
func propose(c *database.Catalog, graphID []byte, fixed []graphalign.Fixed, only []byte) (graphalign.Proposal, int64, *profiler, error) {
	if c == nil || len(graphID) != 16 {
		return graphalign.Proposal{}, 0, nil, promote.ErrInvalid
	}
	db, err := c.DB()
	if err != nil {
		return graphalign.Proposal{}, 0, nil, err
	}
	var rev int64
	if err := db.QueryRow(`SELECT COALESCE(MAX(revision), 0) FROM audit_transactions`).Scan(&rev); err != nil {
		return graphalign.Proposal{}, 0, nil, err
	}
	in, err := loadInputs(c, graphID)
	if err != nil {
		return graphalign.Proposal{}, 0, nil, err
	}
	cfg, namePattern, terms, layer := in.cfg, in.namePattern, in.terms, in.layer
	if len(layer.Subjects) == 0 {
		return graphalign.Proposal{Rows: nil}, rev, nil, nil
	}

	anchors, promoted, err := mergeFixed(c.Graph(), in.graphID, in.evidence, fixed)
	if err != nil {
		return graphalign.Proposal{}, 0, nil, err
	}

	prof, err := newProfiler(c.Graph(), layer.Metas, cfg.Subject, namePattern)
	if err != nil {
		return graphalign.Proposal{}, 0, nil, err
	}
	canon, err := loadCanon(c.Graph(), in.graphID, prof, layer, anchors, cfg, terms)
	if err != nil {
		return graphalign.Proposal{}, 0, nil, err
	}

	prop := graphalign.Align(layer, canon, anchors, cfg)
	eg, err := graphfill.EvidenceGraph(c.Graph(), in.graphID)
	if err != nil {
		return graphalign.Proposal{}, 0, nil, err
	}
	vocab, err := c.Graph().Vocabulary()
	if err != nil {
		return graphalign.Proposal{}, 0, nil, err
	}
	prop.Connections = graphalign.Connections(pageBridges(eg, vocab), pageEnds(promoted, in.evidence, fixed, prop.Rows, only), canon.Edges)
	return prop, rev, prof, nil
}

// inputs is what a proposal reads first: the Evidence graph's layer and the
// rules and vocabulary it was built with.
type inputs struct {
	graphID     []byte
	cfg         graphalign.Config
	namePattern string
	terms       segmentcompare.Terms
	evidence    evidencelayer.Layer
	layer       graphalign.Layer
}

// loadConfig is the alignment config a proposal ranks with and the shortlist
// index is built with: the registry's, with the catalog's value rules. The
// string is the name pattern those rules came from.
func loadConfig(g *graphcache.Graph) (graphalign.Config, string, error) {
	rules, namePattern, err := valuerules.Load(g)
	if err != nil {
		return graphalign.Config{}, "", err
	}
	cfg := matchregistry.Default().AlignConfig()
	cfg.Value = rules
	return cfg, namePattern, nil
}

// loadInputs checks the catalog and graph id (promote.ErrInvalid for a nil
// catalog, an id that is not 16 bytes, or a graph that does not exist) and
// loads the graph's layer.
func loadInputs(c *database.Catalog, graphID []byte) (inputs, error) {
	if c == nil || len(graphID) != 16 {
		return inputs{}, promote.ErrInvalid
	}
	if _, err := c.DB(); err != nil {
		return inputs{}, err
	}
	eg, err := graphfill.EvidenceGraph(c.Graph(), graphID)
	if err != nil {
		return inputs{}, err
	}
	if !eg.Found {
		return inputs{}, promote.ErrInvalid
	}

	cfg, namePattern, err := loadConfig(c.Graph())
	if err != nil {
		return inputs{}, err
	}

	terms, err := canonlinks.LoadTerms(c.Graph())
	if err != nil {
		return inputs{}, err
	}
	evidence, err := evidencelayer.Load(c.Graph(), graphID)
	if err != nil {
		return inputs{}, err
	}
	layer, err := alignLayer(c.Graph(), evidence)
	if err != nil {
		return inputs{}, err
	}
	return inputs{graphID: graphID, cfg: cfg, namePattern: namePattern, terms: terms, evidence: evidence, layer: layer}, nil
}

// alignLayer is the evidence layer with the value metas graphalign compares by.
func alignLayer(g *graphcache.Graph, ev evidencelayer.Layer) (graphalign.Layer, error) {
	metas, err := valuerules.Metas(g)
	if err != nil {
		return graphalign.Layer{}, err
	}
	return graphalign.Layer{Subjects: ev.Subjects, Bridges: ev.Links, Metas: metas}, nil
}

// mergeFixed validates caller fixed pairs and adds already-promoted Subjects
// as free anchors, read from the Evidence graph's accepted members. Caller fixed wins
// when both name the same Subject. promoted is the handle of each Subject
// added from the members, by Subject id.
func mergeFixed(g *graphcache.Graph, graphID []byte, layer evidencelayer.Layer, caller []graphalign.Fixed) (anchors []graphalign.Fixed, promoted map[string][]byte, err error) {
	out := make([]graphalign.Fixed, 0, len(caller)+len(layer.Subjects))
	seen := map[string]bool{}
	promoted = map[string][]byte{}

	for _, f := range caller {
		fixed, err := checkFixed(g, layer, f)
		if err != nil {
			return nil, nil, err
		}
		out = append(out, fixed)
		seen[string(f.SubjectID)] = true
	}

	eg, err := graphfill.EvidenceGraph(g, graphID)
	if err != nil {
		return nil, nil, err
	}
	for _, m := range eg.Members {
		sk := string(m.SubjectID)
		if seen[sk] {
			continue
		}
		if _, ok := layer.Subject(m.SubjectID); !ok {
			continue
		}
		out = append(out, graphalign.Fixed{
			SubjectID: bytes.Clone(m.SubjectID),
			HandleID:  bytes.Clone(m.EntityID),
		})
		promoted[sk] = m.EntityID
		seen[sk] = true
	}

	// Stable order: by subject id bytes (each Subject appears once).
	sort.Slice(out, func(i, j int) bool { return bytes.Compare(out[i].SubjectID, out[j].SubjectID) < 0 })
	return out, promoted, nil
}

// checkFixed validates one caller pair and returns it with its ids copied: the
// Subject must be a primary Subject of the graph; New and Skip carry no
// handle; any other pair names a handle of 16 bytes that the graph cache has
// unmerged and of the Subject's kind. Anything else is promote.ErrInvalid.
func checkFixed(g *graphcache.Graph, layer evidencelayer.Layer, f graphalign.Fixed) (graphalign.Fixed, error) {
	if len(f.SubjectID) != 16 {
		return graphalign.Fixed{}, promote.ErrInvalid
	}
	ps, ok := layer.Subject(f.SubjectID)
	if !ok {
		return graphalign.Fixed{}, promote.ErrInvalid
	}
	switch f.Target {
	case graphalign.TargetNew, graphalign.TargetSkip:
		if len(f.HandleID) != 0 {
			return graphalign.Fixed{}, promote.ErrInvalid
		}
		return graphalign.Fixed{SubjectID: append([]byte(nil), f.SubjectID...), Target: f.Target}, nil
	case "", graphalign.TargetHandle:
	default:
		return graphalign.Fixed{}, promote.ErrInvalid
	}
	if len(f.HandleID) != 16 {
		return graphalign.Fixed{}, promote.ErrInvalid
	}
	n, err := g.Node(f.HandleID)
	if err != nil {
		return graphalign.Fixed{}, err
	}
	if n == nil || n.Merged {
		return graphalign.Fixed{}, promote.ErrInvalid
	}
	v, err := g.Vocabulary()
	if err != nil {
		return graphalign.Fixed{}, err
	}
	if v.TypeKeysByID()[string(n.SubjectTypeID)] != ps.Kind {
		return graphalign.Fixed{}, promote.ErrInvalid
	}
	return graphalign.Fixed{
		SubjectID: append([]byte(nil), f.SubjectID...),
		HandleID:  append([]byte(nil), f.HandleID...),
	}, nil
}
