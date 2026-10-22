package promotealign

import (
	"github.com/mendahu/provenencia/core/claimevidence"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/propertycompare"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// EvidenceValue is one value on a sheet line, typed and not yet worded: the
// client words it in the researcher's locale and name order. TermLabel is a
// term's catalog label, empty for a value that is not a term or a term with no
// label.
type EvidenceValue struct {
	Type      string // propertymeta.ValueType*
	Value     propertycompare.Value
	TermLabel string
}

// EvidenceValues types and labels the values of a claim evidence, from the
// product vocabulary.
type EvidenceValues struct {
	labels map[string]string // "property|term" → label
	types  map[vocabkey.PropertyKey]string
	pools  map[vocabkey.SubjectType]map[vocabkey.PropertyKey][]vocabkey.PropertyKey
}

// NewEvidenceValues reads the graph's product vocabulary: each property's value
// type and each term's label (provenencia origin only: lines compare product
// properties only).
func NewEvidenceValues(g *graphcache.Graph) (EvidenceValues, error) {
	vocab, err := g.Vocabulary()
	if err != nil {
		return EvidenceValues{}, err
	}
	out := EvidenceValues{
		labels: map[string]string{},
		types:  map[vocabkey.PropertyKey]string{},
		pools:  map[vocabkey.SubjectType]map[vocabkey.PropertyKey][]vocabkey.PropertyKey{},
	}
	for _, p := range vocab.Properties {
		if p.Origin == vocabkey.OriginProvenencia {
			out.types[vocabkey.PropertyKey(p.Key)] = p.ValueType
		}
	}
	for _, t := range vocab.Terms {
		if t.Origin == vocabkey.OriginProvenencia {
			out.labels[t.PropertyKey+"|"+t.Key] = t.Label
		}
	}
	for kind, leads := range matchregistry.Default().Subject.Pools {
		out.pools[kind] = map[vocabkey.PropertyKey][]vocabkey.PropertyKey{}
		for lead, pool := range leads {
			out.pools[kind][lead] = pool.With
		}
	}
	return out, nil
}

// Incoming is the line's incoming value. kind is the kind of the Subject the
// section belongs to: the evidence's, or the relationship's under one. A term
// shows the label found on the incoming property, then the section's lead,
// then each property pooled under that lead.
func (v EvidenceValues) Incoming(kind vocabkey.SubjectType, s claimevidence.Section, l claimevidence.Line) EvidenceValue {
	return v.value(l.Property, l.Incoming, v.search(kind, s, l.Property, true))
}

// Member is one pin's member value, labeled from the section's lead, then its
// pool, then the incoming property.
func (v EvidenceValues) Member(kind vocabkey.SubjectType, s claimevidence.Section, l claimevidence.Line, p claimevidence.Pin) EvidenceValue {
	return v.value(l.Property, p.Value, v.search(kind, s, l.Property, false))
}

// search lists the properties a term's label is looked up under.
func (v EvidenceValues) search(kind vocabkey.SubjectType, s claimevidence.Section, property vocabkey.PropertyKey, propertyFirst bool) []vocabkey.PropertyKey {
	rest := append([]vocabkey.PropertyKey{s.Key}, v.pools[kind][s.Key]...)
	if propertyFirst {
		return append([]vocabkey.PropertyKey{property}, rest...)
	}
	return append(rest, property)
}

func (v EvidenceValues) value(property vocabkey.PropertyKey, val propertycompare.Value, search []vocabkey.PropertyKey) EvidenceValue {
	typ, ok := v.types[property]
	if !ok {
		typ = val.Type
	}
	out := EvidenceValue{Type: typ, Value: val}
	if typ == propertymeta.ValueTypeTerm {
		for _, p := range search {
			if label, ok := v.labels[string(p)+"|"+val.Term]; ok {
				out.TermLabel = label
				break
			}
		}
	}
	return out
}
