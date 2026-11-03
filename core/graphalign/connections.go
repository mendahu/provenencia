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
	canonByKey := canonLinks(canon)
	links := map[string]*connectrules.Links{}
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
		key := linkKey(string(b.TypeKey), connectrules.Term(b.Terms))
		filed := func() (*connectrules.Links, error) {
			if links[key] == nil {
				links[key] = connectrules.NewLinks(canonByKey[key])
			}
			return links[key], nil
		}
		res, _ := connectrules.Decide(rule, b.Terms, values, filed)
		c := Connection{BridgeID: b.ID}
		switch res.Decision {
		case connectrules.DecisionFiles:
			c.Outcome = ConnectionFiles
			if res.Link != nil {
				links[key].Add(*res.Link)
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

// linkKey names one type and term's set of filed links.
func linkKey(typeKey string, term vocabkey.BridgeTerm) string {
	return typeKey + "|" + string(term)
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

// canonLinks reads the canon's links of registered bridge types in one pass,
// by type and term, as directed links between records, from the end of the
// rule's first endpoint.
func canonLinks(canon []CanonEdge) map[string][]connectrules.Link {
	rules := map[string]connectrules.Bridge{}
	out := map[string][]connectrules.Link{}
	for _, e := range canon {
		typeKey := string(e.KeyFrom.Type)
		rule, seen := rules[typeKey]
		if !seen {
			rule, _ = connectrules.LookupBridge(typeKey)
			rules[typeKey] = rule
		}
		if len(rule.Endpoints) == 0 {
			continue
		}
		from, to := "h:"+string(e.From), "h:"+string(e.To)
		if string(e.KeyFrom.End) != rule.Endpoints[0].PropertyKey {
			from, to = to, from
		}
		key := linkKey(typeKey, e.KeyFrom.Term)
		out[key] = append(out[key], connectrules.Link{From: from, To: to})
	}
	return out
}
