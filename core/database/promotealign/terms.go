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

// kinship is the directed and inverse marks of the catalog's terms, keyed
// "propertyKey|termKey".
type kinship struct {
	directed map[string]bool
	inverse  map[string]string
}

// loadKinship reads which terms keep direction and which name an inverse.
func loadKinship(q Querier) (kinship, error) {
	rows, err := q.Query(`SELECT p.key, t.key, t.directed, COALESCE(t.inverse_key, '') FROM property_terms t
		JOIN properties p ON p.id = t.property_id
		WHERE t.directed = 1 OR t.inverse_key IS NOT NULL`)
	if err != nil {
		return kinship{}, err
	}
	defer rows.Close()
	k := kinship{directed: map[string]bool{}, inverse: map[string]string{}}
	for rows.Next() {
		var prop, term, inverse string
		var directed bool
		if err := rows.Scan(&prop, &term, &directed, &inverse); err != nil {
			return kinship{}, err
		}
		if directed {
			k.directed[prop+"|"+term] = true
		}
		if inverse != "" {
			k.inverse[prop+"|"+term] = inverse
		}
	}
	return k, rows.Err()
}

// canonical is the key one relationship is matched under. A term and its
// inverse read one relationship from opposite ends ("parent" from the
// parent, "child" from the child), so the pair shares the key that sorts
// first; flip is true when the ends must swap to read that way.
func (k kinship) canonical(propertyKey, term string) (key string, flip bool) {
	if inv, ok := k.inverse[propertyKey+"|"+term]; ok && inv < term {
		return inv, true
	}
	return term, false
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
