package promotealign

import (
	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/segmentcompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// loadTerms reads the catalog's facts about every term of each registered
// bridge's disambiguation Property: whether it keeps direction and which term
// reads it from the other end. It is the one source of both.
func loadTerms(g *graphcache.Graph) (segmentcompare.Terms, error) {
	v, err := g.Vocabulary()
	if err != nil {
		return nil, err
	}
	out := segmentcompare.Terms{}
	for _, b := range connectrules.Bridges() {
		if !connectrules.HasDisambiguation(b.Disambiguation) {
			continue
		}
		typ := vocabkey.SubjectType(b.BridgeTypeKey)
		for _, t := range v.Terms {
			if t.PropertyKey != b.Disambiguation {
				continue
			}
			if out[typ] == nil {
				out[typ] = map[vocabkey.BridgeTerm]segmentcompare.Term{}
			}
			out[typ][vocabkey.BridgeTerm(t.Key)] = segmentcompare.Term{
				Directed: t.Directed, Inverse: vocabkey.BridgeTerm(t.InverseKey),
			}
		}
	}
	return out, nil
}

// linkOf is a bridge as recorded, read from its first endpoint's end or its
// second. It is the one builder of a link for both the layer and the canon,
// so the two sides cannot drift: a bridge has two endpoints, and only a
// participation carries an event type.
func linkOf(b connectrules.Bridge, term, eventType string, firstEnd bool) segmentcompare.Link {
	first, second := b.Endpoints[0], b.Endpoints[1]
	l := segmentcompare.Link{
		Type:     vocabkey.SubjectType(b.BridgeTypeKey),
		Term:     vocabkey.BridgeTerm(term),
		End:      vocabkey.PropertyKey(first.PropertyKey),
		Other:    vocabkey.PropertyKey(second.PropertyKey),
		SameType: first.TypeKey == second.TypeKey,
	}
	if !firstEnd {
		l.End, l.Other = l.Other, l.End
	}
	if b.BridgeTypeKey == vocabkey.BridgeParticipation {
		l.EventType = vocabkey.EventType(eventType)
	}
	return l
}

// eventEndpoint says which end of a bridge is an event: first when it is the
// bridge's first endpoint, and ok false when neither is. It is the one rule
// for where a link reads its event type, for the layer and the canon alike.
func eventEndpoint(b connectrules.Bridge) (first, ok bool) {
	switch {
	case b.Endpoints[0].TypeKey == vocabkey.Event:
		return true, true
	case b.Endpoints[1].TypeKey == vocabkey.Event:
		return false, true
	}
	return false, false
}

// seenEdge identifies one edge, whichever end listed it, by its ends and the
// key it reads from each.
type seenEdge struct {
	a, b       string
	keyA, keyB vocabkey.BridgeKey
}

// edgeSeenKey is the edge a and b share, with the ends in a fixed order.
func edgeSeenKey(a, b []byte, keyA, keyB vocabkey.BridgeKey) seenEdge {
	if string(a) > string(b) {
		a, b, keyA, keyB = b, a, keyB, keyA
	}
	return seenEdge{a: string(a), b: string(b), keyA: keyA, keyB: keyB}
}

// observedTerms maps each Subject to the key of a positive Observation's term
// on the Property, among those the Evidence graph records for the listed
// Subjects. The first key in order wins.
func observedTerms(eg graphcache.EvidenceGraph, props map[string]graphcache.VocabProperty, terms map[string]graphcache.VocabTerm, subjectIDs [][]byte, propertyKey string) map[string]string {
	out := map[string]string{}
	if propertyKey == "" {
		return out
	}
	want := make(map[string]bool, len(subjectIDs))
	for _, id := range subjectIDs {
		want[string(id)] = true
	}
	for _, o := range eg.Observations {
		if o.Polarity != observations.PolarityPositive || !want[string(o.SubjectID)] || len(o.TermID) == 0 {
			continue
		}
		if p, ok := props[string(o.PropertyID)]; !ok || p.Key != propertyKey || p.Origin != vocabkey.OriginProvenencia {
			continue
		}
		t, ok := terms[string(o.TermID)]
		if !ok {
			continue
		}
		if cur, ok := out[string(o.SubjectID)]; !ok || t.Key < cur {
			out[string(o.SubjectID)] = t.Key
		}
	}
	return out
}
