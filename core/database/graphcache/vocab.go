package graphcache

// Vocabulary is the catalog's properties, terms, and subject types.
// A vocabulary write drops it. Structure drops the whole graph, which
// drops this with it.
type Vocabulary struct {
	Properties []VocabProperty
	Terms      []VocabTerm
	Types      []VocabType
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
	return v, nil
}
