package promotealign

import (
	"database/sql"
	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/sqlbatch"
)

// kinship is the directed and inverse marks of the catalog's terms, keyed
// "propertyKey|termKey".
type kinship struct {
	directed map[string]bool
	inverse  map[string]string
}

// loadKinship reads which terms keep direction and which name an inverse.
func loadKinship(g *graphcache.Graph) (kinship, error) {
	v, err := g.Vocabulary()
	if err != nil {
		return kinship{}, err
	}
	k := kinship{directed: map[string]bool{}, inverse: map[string]string{}}
	for _, t := range v.Terms {
		if t.Directed {
			k.directed[t.PropertyKey+"|"+t.Key] = true
		}
		if t.InverseKey != "" {
			k.inverse[t.PropertyKey+"|"+t.Key] = t.InverseKey
		}
	}
	return k, nil
}

// edgeStamp is the role, direction, and neighbor kind one bridge signs with.
// A term and its inverse share the key that sorts first; flip says the ends swap.
func edgeStamp(bridgeKey, role string, kin kinship) (roleOut string, directed bool, neighborKind string, flip bool) {
	b, ok := connectrules.LookupBridge(bridgeKey)
	if ok && len(b.Endpoints) == 2 {
		neighborKind = b.Endpoints[1].TypeKey
	}
	roleOut = role
	if ok && connectrules.HasDisambiguation(b.Disambiguation) {
		roleOut, flip = kin.canonical(b.Disambiguation, role)
		directed = kin.directed[b.Disambiguation+"|"+roleOut]
	}
	return roleOut, directed, neighborKind, flip
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
