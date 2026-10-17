// Package match scores how much one thing resembles each of a set of
// canonical handles. The thing being matched is a Probe: an Interpretation
// Subject (Promote's target suggestions) or another handle (merge hints).
// Candidates are handles of the same Subject type.
//
// The algorithm is data. A Profile lists, for one Subject type, the Features
// that count: which Property, how its values are compared (a Comparer), how
// much a resemblance adds (Weight), and how much a clear disagreement takes
// away (Contradiction). Scores are additive, so the weights read as points:
// "a matching name is worth 10, a different sex at birth costs 2". Default
// profiles live in profiles.go; callers may pass their own.
//
// Pure: no catalog access. core/database/matching loads probes and
// candidates from the catalog and calls Rank.
package match

import (
	"bytes"
	"sort"

	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/namevalue"
)

// Property names one Property by its vocabulary identity.
type Property struct {
	Key    string
	Origin string
}

// Value is one value of a Property, in the shape comparers read. Terms are
// carried by key (sex_at_birth "male"), so profiles can name them.
type Value struct {
	Text       string
	HasText    bool
	Integer    int64
	HasInteger bool
	Term       string
	// TermLabel is the term's catalog label, for showing the value; empty
	// when the loader had none. Comparers never read it.
	TermLabel string
	// TermID is the catalog term id when known. Comparers compare terms by
	// key; Promote's exhibits pass the id to autoreconcile.Compatible, which
	// falls back to the key when it is empty.
	TermID []byte
	// Neutral is the term's catalog flag: no evidence either way, so a
	// comparison with it is not comparable. Loaders copy it from the term.
	Neutral bool
	Date    *datevalue.Value
	Name    *namevalue.Value
	// NamePattern is the name pattern key for Name (its Person's name
	// format, once loaders supply it); empty uses the comparer's pattern.
	NamePattern string

	// Counted is true on values a cache loader filled with the evidence
	// below; every other value weighs 1.
	Counted bool
	// Support is the distinct Sources behind a cached value; Against the
	// negative Observations that counted against it; Rank its order on the
	// handle; Reason the auto-reconciler's verdict (kept, outvoted, weak,
	// denied, provisional), not a match.Reason. Denied is Reason == denied. Set with SetEvidence.
	Support, Against, Rank int
	Reason                 autoreconcile.Reason
	Denied                 bool
	// Negative marks a negative Observation: it never speaks for a match, it
	// counts against a value it is compatible with. ObservationID is the
	// Observation a probe value was loaded from.
	Negative      bool
	ObservationID []byte
}

// Values are a thing's values, every Property it has, each possibly several
// (a Subject's Observations, a handle's resolved clusters at every rank).
type Values map[Property][]Value

// Candidate is one handle that could match.
type Candidate struct {
	EntityID []byte
	Ref      string
	Values   Values
}

// Reason is one Feature's part in a score, for explaining a match ("same
// name, born the same year") and for tuning.
type Reason struct {
	Property Property
	// Similarity is the best resemblance between a probe value and a
	// candidate value, times both values' weights (a value is weighed by its
	// support against the leading value), 0…1; 0 is a contradiction.
	Similarity float64
	// Discount is the share of points the matched values' frequency in the
	// project cost, 0…1; 0 when the values clearly disagree.
	Discount float64
	// Probe and Candidate are the two values that produced the Similarity,
	// or the contradiction.
	Probe, Candidate ValueRef
	// Contribution is what the Feature added to the score: Weight ×
	// Similarity × (1 − Discount), or −Contradiction when the values clearly
	// disagree.
	Contribution float64
}

// Match is one candidate that scored at least the profile's MinScore.
type Match struct {
	EntityID []byte
	Ref      string
	Score    float64
	Reasons  []Reason // profile order; only Features both sides could compare
}

// Rank scores every candidate against the probe and returns those reaching
// p.MinScore, best first (score, then ref). limit <= 0 returns them all.
func Rank(p Profile, probe Values, candidates []Candidate, limit int) []Match {
	var out []Match
	for _, c := range candidates {
		m := Score(p, probe, c)
		if m.Score >= p.MinScore && len(m.Reasons) > 0 {
			out = append(out, m)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		if out[i].Ref != out[j].Ref {
			return out[i].Ref < out[j].Ref
		}
		return bytes.Compare(out[i].EntityID, out[j].EntityID) < 0
	})
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}

// Score is one candidate's score and its reasons, whatever the threshold.
// A Feature counts only when both sides have a value the comparer can judge.
func Score(p Profile, probe Values, c Candidate) Match {
	m := Match{EntityID: c.EntityID, Ref: c.Ref}
	for _, f := range p.Features {
		a, b := poolPeriods(f.Property, probe, c.Values)
		pr := pairValues(f.Comparer, f.Cardinality, a, b)
		if !pr.Comparable {
			continue
		}
		sim := pr.Similarity
		r := Reason{Property: f.Property, Similarity: sim, Discount: pr.Discount, Probe: pr.Probe, Candidate: pr.Candidate, Contribution: f.Weight * sim * (1 - pr.Discount)}
		if sim == 0 {
			r.Contribution = -f.Contradiction
		}
		m.Score += r.Contribution
		m.Reasons = append(m.Reasons, r)
	}
	return m
}
