package promotealign

import (
	"database/sql"
	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/sqlbatch"
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

// observedTerms maps each Subject to the key of a positive Observation's
// term on the Property (first key in order wins).
func observedTerms(q Querier, subjectIDs [][]byte, propertyKey string) (map[string]string, error) {
	return termsBy(q, `SELECT o.subject_id, t.key FROM observations o
		JOIN properties p ON p.id = o.property_id AND p.key = ? AND p.origin = 'provenencia'
		JOIN property_terms t ON t.id = o.value_term_id
		WHERE o.polarity = 'positive' AND o.subject_id IN (`, `) ORDER BY o.subject_id, t.key`,
		subjectIDs, propertyKey)
}

func termsBy(q Querier, head, tail string, ids [][]byte, propertyKey string) (map[string]string, error) {
	out := map[string]string{}
	ids = database.UniqueBlobIDs(ids)
	if propertyKey == "" {
		return out, nil
	}
	err := sqlbatch.Query(q, head, tail, []any{propertyKey}, ids, func(rows *sql.Rows) error {
		var id []byte
		var term string
		if err := rows.Scan(&id, &term); err != nil {
			return err
		}
		if _, ok := out[string(id)]; !ok {
			out[string(id)] = term
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
