package graphcache

import (
	"sort"

	"github.com/mendahu/provenencia/core/vocabkey"
)

// Vocabulary is the catalog's properties, terms, subject types, and the
// properties each type binds, in binding order.
// A vocabulary write drops it. Structure drops the whole graph, which
// drops this with it.
type Vocabulary struct {
	Properties []VocabProperty
	Terms      []VocabTerm
	Types      []VocabType
	Bindings   []VocabBinding
}

// VocabBinding is one subject_type_properties row: a property bound to a
// subject type at SortOrder.
type VocabBinding struct {
	TypeID, PropertyID []byte
	SortOrder          int
}

// BindingOrder returns, per provenencia subject type key, its bound property
// keys in binding order. Types of other origins are left out: a key is only
// unique within an origin.
func (v Vocabulary) BindingOrder() map[vocabkey.SubjectType][]vocabkey.PropertyKey {
	out := map[vocabkey.SubjectType][]vocabkey.PropertyKey{}
	for _, t := range v.Types {
		if t.Origin != vocabkey.OriginProvenencia {
			continue
		}
		if order := v.TypeOrder(t.ID); len(order) > 0 {
			out[vocabkey.SubjectType(t.Key)] = order
		}
	}
	return out
}

// TypeOrder returns the property keys bound to one subject type, by SortOrder,
// then key.
func (v Vocabulary) TypeOrder(typeID []byte) []vocabkey.PropertyKey {
	props := v.PropertiesByID()
	type bound struct {
		key   vocabkey.PropertyKey
		order int
	}
	var bs []bound
	for _, b := range v.Bindings {
		if string(b.TypeID) != string(typeID) {
			continue
		}
		if p, ok := props[string(b.PropertyID)]; ok {
			bs = append(bs, bound{vocabkey.PropertyKey(p.Key), b.SortOrder})
		}
	}
	sort.SliceStable(bs, func(i, j int) bool {
		if bs[i].order != bs[j].order {
			return bs[i].order < bs[j].order
		}
		return bs[i].key < bs[j].key
	})
	keys := make([]vocabkey.PropertyKey, len(bs))
	for i, b := range bs {
		keys[i] = b.key
	}
	return keys
}

// VocabProperty is one properties row.
type VocabProperty struct {
	ID          []byte
	Key         string
	Label       string
	Origin      string
	ValueType   string
	Cardinality string
}

// VocabTerm is one property term, with the property key it belongs to.
type VocabTerm struct {
	ID          []byte
	PropertyKey string
	Key         string
	Origin      string
	Label       string
	Directed    bool
	InverseKey  string
	Neutral     bool // no evidence either way (property_terms.neutral)
}

// TermsByID maps each term's id (as a string) to the term.
func (v Vocabulary) TermsByID() map[string]VocabTerm {
	out := make(map[string]VocabTerm, len(v.Terms))
	for _, t := range v.Terms {
		out[string(t.ID)] = t
	}
	return out
}

// PropertiesByID maps each property's id (as a string) to the property.
func (v Vocabulary) PropertiesByID() map[string]VocabProperty {
	out := make(map[string]VocabProperty, len(v.Properties))
	for _, p := range v.Properties {
		out[string(p.ID)] = p
	}
	return out
}

// TypeKeysByID maps each provenencia subject type's id (as a string) to its
// key. Types of other origins are left out: a key is only unique within an
// origin.
func (v Vocabulary) TypeKeysByID() map[string]string {
	out := make(map[string]string, len(v.Types))
	for _, t := range v.Types {
		if t.Origin == vocabkey.OriginProvenencia {
			out[string(t.ID)] = t.Key
		}
	}
	return out
}

// TypeIDsByKey maps each provenencia subject type's key to its id, the other
// direction of TypeKeysByID.
func (v Vocabulary) TypeIDsByKey() map[string][]byte {
	out := make(map[string][]byte, len(v.Types))
	for _, t := range v.Types {
		if t.Origin == vocabkey.OriginProvenencia {
			out[t.Key] = t.ID
		}
	}
	return out
}

// PropertyID returns the id of the property with this key and origin, or nil
// when the catalog has none.
func (v Vocabulary) PropertyID(key, origin string) []byte {
	for _, p := range v.Properties {
		if p.Key == key && p.Origin == origin {
			return p.ID
		}
	}
	return nil
}

// VocabType is one subject type.
type VocabType struct {
	ID     []byte
	Key    string
	Origin string
}

// Vocabulary returns the cached vocabulary, loading it on the first read.
func (g *Graph) Vocabulary() (Vocabulary, error) {
	if g == nil {
		return Vocabulary{}, nil
	}
	if g.vocabOK {
		return g.vocab, nil
	}
	v, err := g.loadVocabulary()
	if err != nil {
		return Vocabulary{}, err
	}
	g.vocab = v
	g.vocabOK = true
	return v, nil
}

func (g *Graph) loadVocabulary() (Vocabulary, error) {
	var v Vocabulary
	props, err := g.query(`SELECT id, key, label, origin, value_type, cardinality FROM properties`)
	if err != nil {
		return Vocabulary{}, err
	}
	for props.Next() {
		var p VocabProperty
		if err := props.Scan(&p.ID, &p.Key, &p.Label, &p.Origin, &p.ValueType, &p.Cardinality); err != nil {
			_ = props.Close()
			return Vocabulary{}, err
		}
		p.ID = cloneID(p.ID)
		v.Properties = append(v.Properties, p)
	}
	if err := props.Err(); err != nil {
		_ = props.Close()
		return Vocabulary{}, err
	}
	_ = props.Close()

	types, err := g.query(`SELECT id, key, origin FROM subject_types`)
	if err != nil {
		return Vocabulary{}, err
	}
	for types.Next() {
		var t VocabType
		if err := types.Scan(&t.ID, &t.Key, &t.Origin); err != nil {
			_ = types.Close()
			return Vocabulary{}, err
		}
		t.ID = cloneID(t.ID)
		v.Types = append(v.Types, t)
	}
	if err := types.Err(); err != nil {
		_ = types.Close()
		return Vocabulary{}, err
	}
	_ = types.Close()

	terms, err := g.query(`SELECT t.id, p.key, t.key, t.origin, t.label, t.directed, COALESCE(t.inverse_key, ''), t.neutral
		FROM property_terms t
		JOIN properties p ON p.id = t.property_id`)
	if err != nil {
		return Vocabulary{}, err
	}
	for terms.Next() {
		var t VocabTerm
		var directed, neutral int
		if err := terms.Scan(&t.ID, &t.PropertyKey, &t.Key, &t.Origin, &t.Label, &directed, &t.InverseKey, &neutral); err != nil {
			_ = terms.Close()
			return Vocabulary{}, err
		}
		t.Directed = directed != 0
		t.Neutral = neutral != 0
		t.ID = cloneID(t.ID)
		v.Terms = append(v.Terms, t)
	}
	if err := terms.Err(); err != nil {
		_ = terms.Close()
		return Vocabulary{}, err
	}
	_ = terms.Close()

	bindings, err := g.query(`SELECT subject_type_id, property_id, sort_order FROM subject_type_properties ORDER BY subject_type_id, sort_order, property_id`)
	if err != nil {
		return Vocabulary{}, err
	}
	for bindings.Next() {
		var b VocabBinding
		if err := bindings.Scan(&b.TypeID, &b.PropertyID, &b.SortOrder); err != nil {
			_ = bindings.Close()
			return Vocabulary{}, err
		}
		b.TypeID, b.PropertyID = cloneID(b.TypeID), cloneID(b.PropertyID)
		v.Bindings = append(v.Bindings, b)
	}
	if err := bindings.Err(); err != nil {
		_ = bindings.Close()
		return Vocabulary{}, err
	}
	_ = bindings.Close()
	return v, nil
}
