package graphalign

import (
	"bytes"
	"slices"

	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// ConnectionOutcome is what Done will do with one connection, as a code the
// page words. It is connectrules' filing decision, split by why a connection
// waits.
type ConnectionOutcome string

const (
	// ConnectionFiles: Done files the connection.
	ConnectionFiles ConnectionOutcome = ConnectionOutcome(connectrules.DecisionFiles)
	// ConnectionSkipped: an end is set to Skip.
	ConnectionSkipped ConnectionOutcome = "skipped"
	// ConnectionUnset: an end has no target yet.
	ConnectionUnset ConnectionOutcome = "unset"
	// ConnectionNotOnPage: an end is not on the page, or the connection is
	// not complete (no Subject at an end, two records at one, no type).
	ConnectionNotOnPage ConnectionOutcome = "not_on_page"
)

// ConnectionOutcomes lists every outcome: files, the three waits, then each
// refusal connectrules.Refusals reports.
func ConnectionOutcomes() []ConnectionOutcome {
	out := []ConnectionOutcome{ConnectionFiles, ConnectionSkipped, ConnectionUnset, ConnectionNotOnPage}
	for _, d := range connectrules.Refusals() {
		out = append(out, ConnectionOutcome(d))
	}
	return out
}

// PageBridge is one unfiled bridge Subject on the page's graph. Ends holds,
// per endpoint in rule order, the Subjects its Observations name, in graph
// order; Terms are its disambiguation term keys.
type PageBridge struct {
	ID      []byte
	TypeKey vocabkey.SubjectType
	Terms   []vocabkey.BridgeTerm
	Ends    [][][]byte
}

// EndTarget is where the page has put one Subject: a handle, TargetNew, or a
// wait (skipped, unset or not on the page).
type EndTarget struct {
	Target   Target
	HandleID []byte
	Waits    ConnectionOutcome
}

// Connection is the outcome for one bridge. Waiting is the Subject a wait
// names, nil for any other outcome.
type Connection struct {
	BridgeID []byte
	Outcome  ConnectionOutcome
	Waiting  []byte
}

// Connections decides each bridge the way Done files it: in bridge id order,
// through connectrules.Decide, with the canon's filed links and the page's
// own earlier connections as the links a loop is checked against. A Subject
// absent from ends is not on the page. A bridge of an unknown type is left out.
func Connections(bridges []PageBridge, ends map[string]EndTarget, canon []CanonEdge) []Connection {
	sorted := slices.Clone(bridges)
	slices.SortStableFunc(sorted, func(a, b PageBridge) int { return bytes.Compare(a.ID, b.ID) })
	page := map[string][]connectrules.Link{}
	var out []Connection
	for _, b := range sorted {
		rule, ok := connectrules.LookupBridge(string(b.TypeKey))
		if !ok {
			continue
		}
		values := make([][]connectrules.Value, len(b.Ends))
		for i, subjects := range b.Ends {
			for _, id := range subjects {
				values[i] = append(values[i], endValue(id, ends))
			}
		}
		term := connectrules.Term(b.Terms)
		key := string(b.TypeKey) + "|" + string(term)
		res, _ := connectrules.Decide(rule, b.Terms, values, func() ([]connectrules.Link, error) {
			return append(canonLinks(canon, rule, term), page[key]...), nil
		})
		c := Connection{BridgeID: b.ID}
		switch res.Decision {
		case connectrules.DecisionFiles:
			c.Outcome = ConnectionFiles
			if len(values) == 2 && rule.IsAcyclic(term) {
				page[key] = append(page[key], connectrules.Link{From: values[0][0].Record, To: values[1][0].Record})
			}
		case connectrules.DecisionWaits:
			c.Outcome = ConnectionOutcome(res.Wait)
			c.Waiting = b.Ends[res.Endpoint][res.Value]
		case connectrules.DecisionIncomplete:
			c.Outcome = ConnectionNotOnPage
		default:
			c.Outcome = ConnectionOutcome(res.Decision)
		}
		out = append(out, c)
	}
	return out
}

// endValue is the record one Subject lands on, or why it has none. A handle
// and a New record key apart.
func endValue(id []byte, ends map[string]EndTarget) connectrules.Value {
	e, ok := ends[string(id)]
	switch {
	case !ok:
		return connectrules.Value{Wait: string(ConnectionNotOnPage)}
	case e.Waits != "":
		return connectrules.Value{Wait: string(e.Waits)}
	case e.Target == TargetNew:
		return connectrules.Value{Record: "n:" + string(id)}
	}
	return connectrules.Value{Record: "h:" + string(e.HandleID)}
}

// canonLinks reads the canon's links of the bridge's type and term as
// directed links between records, from the end of the rule's first endpoint.
func canonLinks(canon []CanonEdge, rule connectrules.Bridge, term vocabkey.BridgeTerm) []connectrules.Link {
	var out []connectrules.Link
	for _, e := range canon {
		if string(e.KeyFrom.Type) != rule.BridgeTypeKey || e.KeyFrom.Term != term {
			continue
		}
		from, to := "h:"+string(e.From), "h:"+string(e.To)
		if string(e.KeyFrom.End) != rule.Endpoints[0].PropertyKey {
			from, to = to, from
		}
		out = append(out, connectrules.Link{From: from, To: to})
	}
	return out
}
