package promotealign

import (
	"github.com/mendahu/provenencia/core/database"
)

// inBatch bounds an IN list; a batched read is still a fixed number of queries.
const inBatch = 500

// keptTerms maps each handle to the key of its kept rank-1 term on the
// Property. With several, the first key in order wins, so a signature is the
// same every run.
func keptTerms(q Querier, entityIDs [][]byte, propertyKey string) (map[string]string, error) {
	return termsBy(q, `SELECT r.entity_id, t.key FROM auto_reconciler_values r
		JOIN properties p ON p.id = r.property_id AND p.key = ? AND p.origin = 'provenencia'
		JOIN property_terms t ON t.id = r.value_term_id
		WHERE r.rank = 1 AND r.reason = 'kept' AND r.entity_id IN (`, `) ORDER BY r.entity_id, t.key`,
		entityIDs, propertyKey)
}

// directedTerms is every directed term ("propertyKey|termKey"): a term whose
// order is part of its meaning (parent of, part of), as the catalog marks it.
func directedTerms(q Querier) (map[string]bool, error) {
	rows, err := q.Query(`SELECT p.key, t.key FROM property_terms t
		JOIN properties p ON p.id = t.property_id
		WHERE t.directed = 1`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var prop, term string
		if err := rows.Scan(&prop, &term); err != nil {
			return nil, err
		}
		out[prop+"|"+term] = true
	}
	return out, rows.Err()
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
	for start := 0; start < len(ids); start += inBatch {
		chunk := ids[start:min(start+inBatch, len(ids))]
		args := append([]any{propertyKey}, database.BlobArgs(chunk)...)
		rows, err := q.Query(head+database.SQLInPlaceholders(len(chunk))+tail, args...)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var id []byte
			var term string
			if err := rows.Scan(&id, &term); err != nil {
				_ = rows.Close()
				return nil, err
			}
			if _, ok := out[string(id)]; !ok {
				out[string(id)] = term
			}
		}
		err = rows.Err()
		_ = rows.Close()
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}
