package promotealign

import (
	"github.com/mendahu/provenencia/core/claimevidence"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/files"
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
	// ImageRelPath is a likeness's image File path under the project, set
	// from ResolvePaths; empty for any other value.
	ImageRelPath string
}

// EvidenceValues types and labels the values of a claim evidence, from the
// product vocabulary.
type EvidenceValues struct {
	labels map[string]string // "property|term" → label
	types  map[vocabkey.PropertyKey]string
	pools  map[vocabkey.SubjectType]map[vocabkey.PropertyKey][]vocabkey.PropertyKey
	paths  map[string]string // image File id → objects/… path, from ResolvePaths
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

// ResolvePaths reads the image path of every likeness the evidence shows, on
// lines and pins, in its sections and its relationships' sections, with one
// query. Incoming and Member then fill ImageRelPath.
func (v *EvidenceValues) ResolvePaths(c *database.Catalog, e claimevidence.Evidence) error {
	var ids [][]byte
	seen := map[string]bool{}
	add := func(x propertycompare.Value) {
		if x.Depiction == nil || len(x.Depiction.ImageFileID) == 0 || seen[string(x.Depiction.ImageFileID)] {
			return
		}
		seen[string(x.Depiction.ImageFileID)] = true
		ids = append(ids, x.Depiction.ImageFileID)
	}
	walk := func(sections []claimevidence.Section) {
		for _, s := range sections {
			for _, l := range s.Lines {
				add(l.Incoming)
				for _, p := range l.Members {
					add(p.Value)
				}
			}
		}
	}
	walk(e.Sections)
	for _, r := range e.Relationships {
		walk(r.Sections)
	}
	if len(ids) == 0 {
		return nil
	}
	paths, err := files.RelPaths(c, ids)
	if err != nil {
		return err
	}
	v.paths = paths
	return nil
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
	if val.Depiction != nil {
		out.ImageRelPath = v.paths[string(val.Depiction.ImageFileID)]
	}
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
