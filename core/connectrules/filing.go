package connectrules

import "github.com/mendahu/provenencia/core/vocabkey"

// This file is the filing rule: whether a bridge Subject is filed onto an
// association, and if not, why. Done's filer (identityclaims) and Promote's
// proposal (graphalign) both call Decide, so the page says what Done does and
// a new refusal is one rung here.

// Link is one filed link of an acyclic term, from a record key to a record key.
type Link struct{ From, To string }

// Value is one Subject an endpoint names. Record is the record key it lands
// on; Wait is the caller's code for why it has none.
type Value struct{ Record, Wait string }

// Decision is the filing rule's answer for one bridge.
type Decision string

const (
	// DecisionFiles: the bridge is filed onto its association.
	DecisionFiles Decision = "files"
	// DecisionWaits: an endpoint's Subject has no record yet (Result.Wait).
	DecisionWaits Decision = "waits"
	// DecisionIncomplete: an endpoint names no Subject or Subjects on two
	// records, or a typed link has no type term.
	DecisionIncomplete Decision = "incomplete"
	// DecisionSelfLink: both ends are one record.
	DecisionSelfLink Decision = "self_link"
	// DecisionCycle: the link would close a loop among filed links of its term.
	DecisionCycle Decision = "cycle"
)

// Refusals lists the decisions that refuse a bridge whose ends are ready.
func Refusals() []Decision {
	return []Decision{DecisionSelfLink, DecisionCycle}
}

// Result is Decide's answer. With DecisionWaits, Endpoint and Value index the
// waiting Subject and Wait is its code.
type Result struct {
	Decision        Decision
	Endpoint, Value int
	Wait            string
}

// KeysOnType reports whether the bridge's disambiguation is the association's
// identity (relationship type, place relationship type). Role is a reconciled
// value, not part of the key.
func (b Bridge) KeysOnType() bool {
	return HasDisambiguation(b.Disambiguation) && b.Disambiguation != DisambiguationRole
}

// Term is the bridge's term: the lowest non-empty key, empty when none. It is
// order-free, so every caller picks the same term from the same Observations.
func Term(terms []vocabkey.BridgeTerm) vocabkey.BridgeTerm {
	var out vocabkey.BridgeTerm
	for _, t := range terms {
		if t != "" && (out == "" || t < out) {
			out = t
		}
	}
	return out
}

// IsAcyclic reports whether term's filed links may not loop.
func (b Bridge) IsAcyclic(term vocabkey.BridgeTerm) bool {
	for _, t := range b.AcyclicTerms {
		if t == term {
			return true
		}
	}
	return false
}

// ClosesCycle reports whether adding from→to to edges would close a loop,
// that is, whether from is reachable from to. It terminates on edges that
// already loop.
func ClosesCycle(edges []Link, from, to string) bool {
	adj := map[string][]string{}
	for _, e := range edges {
		adj[e.From] = append(adj[e.From], e.To)
	}
	seen := map[string]bool{to: true}
	queue := []string{to}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if cur == from {
			return true
		}
		for _, next := range adj[cur] {
			if !seen[next] {
				seen[next] = true
				queue = append(queue, next)
			}
		}
	}
	return false
}

// Decide runs the filing ladder for one bridge, in this order: for each
// endpoint in rule order, no value is incomplete, the first value with a Wait
// waits, values on more than one record are incomplete; then two endpoints on
// one record is a self-link; then a typed bridge with no term is incomplete;
// then, for an acyclic term, filed supplies the filed links and a loop is a
// cycle. filed is called only for that last rung.
func Decide(b Bridge, terms []vocabkey.BridgeTerm, ends [][]Value, filed func() ([]Link, error)) (Result, error) {
	records := make([]string, 0, len(b.Endpoints))
	for e := range b.Endpoints {
		var values []Value
		if e < len(ends) {
			values = ends[e]
		}
		if len(values) == 0 {
			return Result{Decision: DecisionIncomplete}, nil
		}
		for i, v := range values {
			if v.Wait != "" {
				return Result{Decision: DecisionWaits, Endpoint: e, Value: i, Wait: v.Wait}, nil
			}
		}
		record := values[0].Record
		for _, v := range values[1:] {
			if v.Record != record {
				return Result{Decision: DecisionIncomplete}, nil
			}
		}
		records = append(records, record)
	}
	seen := map[string]bool{}
	for _, r := range records {
		seen[r] = true
	}
	if len(seen) < len(records) {
		return Result{Decision: DecisionSelfLink}, nil
	}
	term := Term(terms)
	if b.KeysOnType() && term == "" {
		return Result{Decision: DecisionIncomplete}, nil
	}
	if len(records) == 2 && b.IsAcyclic(term) {
		edges, err := filed()
		if err != nil {
			return Result{}, err
		}
		if ClosesCycle(edges, records[0], records[1]) {
			return Result{Decision: DecisionCycle}, nil
		}
	}
	return Result{Decision: DecisionFiles}, nil
}
