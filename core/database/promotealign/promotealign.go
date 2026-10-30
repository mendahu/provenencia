// Package promotealign loads catalog inputs for Promote graph alignment and
// calls pure graphalign.Align (S9-42). It does not score pairs or walk policy.
package promotealign

import (
	"bytes"
	"sort"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonlinks"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/valuerules"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/segmentcompare"
)

// Propose loads the Evidence layer of graphID, walks the catalog's canonical
// graph from the fixed anchors and each unfixed kind's top property matches,
// and runs graphalign.Align. The int64 is MAX(audit_transactions.revision) at
// load time, the stamp Done sends back. An unknown or empty graphID is
// promote.ErrInvalid. A graph read error is returned instead of a proposal.
func Propose(c *database.Catalog, graphID []byte, fixed []graphalign.Fixed) (graphalign.Proposal, int64, error) {
	prop, rev, _, err := propose(c, graphID, fixed)
	return prop, rev, err
}

// propose is Propose, and the profiler it used, for a test that reads which
// handles were profiled. The profiler is nil when the graph has no Subject to
// place.
func propose(c *database.Catalog, graphID []byte, fixed []graphalign.Fixed) (graphalign.Proposal, int64, *profiler, error) {
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
	cfg, namePattern, terms, layer, primary := in.cfg, in.namePattern, in.terms, in.layer, in.primary
	if len(layer.Subjects) == 0 {
		return graphalign.Proposal{Rows: nil}, rev, nil, nil
	}

	anchors, err := mergeFixed(c.Graph(), in.graphID, primary, fixed)
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
	return prop, rev, prof, nil
}

// inputs is what a proposal reads first: the Evidence graph's layer and the
// rules and vocabulary it was built with.
type inputs struct {
	graphID     []byte
	cfg         graphalign.Config
	namePattern string
	terms       segmentcompare.Terms
	layer       graphalign.Layer
	primary     map[string]primarySubject
}

// loadConfig is the alignment config a proposal ranks with and the shortlist
// index is built with: the registry's, with the catalog's value rules. The
// string is the name pattern those rules came from.
func loadConfig(g *graphcache.Graph, q database.Querier) (graphalign.Config, string, error) {
	rules, namePattern, err := valuerules.Load(g, q)
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
	db, err := c.DB()
	if err != nil {
		return inputs{}, err
	}
	var q database.Querier = db
	eg, err := c.Graph().EvidenceGraph(graphID)
	if err != nil {
		return inputs{}, err
	}
	if !eg.Found {
		return inputs{}, promote.ErrInvalid
	}

	cfg, namePattern, err := loadConfig(c.Graph(), q)
	if err != nil {
		return inputs{}, err
	}

	terms, err := canonlinks.LoadTerms(c.Graph())
	if err != nil {
		return inputs{}, err
	}
	layer, primary, err := loadLayerMemo(c.Graph(), graphID, cfg, namePattern, terms)
	if err != nil {
		return inputs{}, err
	}
	return inputs{graphID: graphID, cfg: cfg, namePattern: namePattern, terms: terms, layer: layer, primary: primary}, nil
}

// mergeFixed validates caller fixed pairs and adds already-promoted Subjects
// as free anchors, read from the Evidence graph's accepted members. Caller fixed wins
// when both name the same Subject.
func mergeFixed(g *graphcache.Graph, graphID []byte, primary map[string]primarySubject, caller []graphalign.Fixed) ([]graphalign.Fixed, error) {
	out := make([]graphalign.Fixed, 0, len(caller)+len(primary))
	seen := map[string]bool{}

	for _, f := range caller {
		fixed, err := checkFixed(g, primary, f)
		if err != nil {
			return nil, err
		}
		out = append(out, fixed)
		seen[string(f.SubjectID)] = true
	}

	eg, err := g.EvidenceGraph(graphID)
	if err != nil {
		return nil, err
	}
	for _, m := range eg.Members {
		sk := string(m.SubjectID)
		if seen[sk] {
			continue
		}
		if _, ok := primary[sk]; !ok {
			continue
		}
		out = append(out, graphalign.Fixed{
			SubjectID: bytes.Clone(m.SubjectID),
			HandleID:  bytes.Clone(m.EntityID),
		})
		seen[sk] = true
	}

	// Stable order: by subject id bytes (each Subject appears once).
	sort.Slice(out, func(i, j int) bool { return bytes.Compare(out[i].SubjectID, out[j].SubjectID) < 0 })
	return out, nil
}

// checkFixed validates one caller pair and returns it with its ids copied: the
// Subject must be a primary Subject of the graph; New and Skip carry no
// handle; any other pair names a handle of 16 bytes that the graph cache has
// unmerged and of the Subject's kind. Anything else is promote.ErrInvalid.
func checkFixed(g *graphcache.Graph, primary map[string]primarySubject, f graphalign.Fixed) (graphalign.Fixed, error) {
	if len(f.SubjectID) != 16 {
		return graphalign.Fixed{}, promote.ErrInvalid
	}
	ps, ok := primary[string(f.SubjectID)]
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
	if v.TypeKeysByID()[string(n.SubjectTypeID)] != ps.kind {
		return graphalign.Fixed{}, promote.ErrInvalid
	}
	return graphalign.Fixed{
		SubjectID: append([]byte(nil), f.SubjectID...),
		HandleID:  append([]byte(nil), f.HandleID...),
	}, nil
}
