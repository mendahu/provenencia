// Package match scores how much one thing resembles each of a set of
// canonical handles. The thing being matched is a Probe: an Interpretation
// Subject (Promote's target suggestions) or another handle (merge hints).
// Candidates are handles of the same Subject type.
//
// The algorithm is data. A Profile lists, for one Subject type, the Features
// that count: which Property, how its values are compared (a Comparer), how
// much a resemblance adds (Weight), and how much a clear disagreement takes
// away (Contradiction). Scores are additive, so the weights read as points:
// "a matching name is worth 10, a different sex at birth costs 8". Default
// profiles live in profiles.go; callers may pass their own.
//
// Pure: no catalog access. core/database/matching loads probes and
// candidates from the catalog and calls Rank.
package match

import (
	"bytes"
	"sort"

	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/namevalues"
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
	// TermID is the catalog term id when known. Pairwise evaluation
	// (Compatible) prefers it; if empty, Evaluate falls back to Term bytes.
	TermID     []byte
	Date       *datevalues.Value
	Name       *namevalues.Value
	// NamePattern is the name pattern key for Name (its Person's name
	// format, once loaders supply it); empty uses the comparer's pattern.
	NamePattern string
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
	// Similarity is the best resemblance between any probe value and any
	// candidate value, 0…1.
	Similarity float64
	// Contribution is what the Feature added to the score: Weight ×
	// Similarity, or −Contradiction when the values clearly disagree.
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
		sim, ok := best(f.Comparer, probe[f.Property], c.Values[f.Property])
		if !ok {
			continue
		}
		r := Reason{Property: f.Property, Similarity: sim, Contribution: f.Weight * sim}
		if sim == 0 {
			r.Contribution = -f.Contradiction
		}
		m.Score += r.Contribution
		m.Reasons = append(m.Reasons, r)
	}
	return m
}

// best is the highest similarity over every comparable pair; ok is false
// when no pair is comparable (a side is empty, or only neutral values).
func best(cmp Comparer, a, b []Value) (float64, bool) {
	var (
		top float64
		ok  bool
	)
	for _, x := range a {
		for _, y := range b {
			s, comparable := cmp.Compare(x, y)
			if !comparable {
				continue
			}
			if !ok || s > top {
				top, ok = s, true
			}
		}
	}
	return top, ok
}
