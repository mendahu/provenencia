package evidencelayer

import (
	"bytes"
	"slices"

	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/depictionvalue"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/valuecodec"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// Decode is the one row decoder for both sides: a Source Observation and a
// cached handle value carry the same value columns.
func Decode(scalar propertymeta.Scalar, termID, date, name, image []byte, terms map[string]graphcache.VocabTerm) (autoreconcile.Value, error) {
	v := autoreconcile.Value{Scalar: scalar}
	if len(termID) > 0 {
		t := terms[string(termID)]
		v.TermID = bytes.Clone(termID)
		v.TermKey = t.Key
		v.Neutral = t.Neutral
	}
	if len(date) > 0 {
		d, err := valuecodec.UnmarshalDate(date)
		if err != nil {
			return autoreconcile.Value{}, err
		}
		v.Date = &d
	}
	if len(name) > 0 {
		n, err := valuecodec.UnmarshalName(name)
		if err != nil {
			return autoreconcile.Value{}, err
		}
		v.Name = &n
	}
	if len(image) > 0 {
		v.Depiction = &depictionvalue.Value{ImageFileID: bytes.Clone(image)}
	}
	return v, nil
}

// layerValues is each listed Subject's own product Observations, both
// polarities, by property key and in Observation id order. subjects is the set
// of Subject ids to read.
func layerValues(eg graphcache.EvidenceGraph, subjects map[string]bool, props map[string]graphcache.VocabProperty, terms map[string]graphcache.VocabTerm) (map[string]map[vocabkey.PropertyKey][]autoreconcile.Candidate, error) {
	obs := slices.Clone(eg.Observations)
	slices.SortStableFunc(obs, func(a, b graphcache.EvidenceObservation) int { return bytes.Compare(a.ID, b.ID) })
	out := map[string]map[vocabkey.PropertyKey][]autoreconcile.Candidate{}
	for _, o := range obs {
		if !subjects[string(o.SubjectID)] || len(o.ValueSubjectID) > 0 {
			continue
		}
		prop, ok := props[string(o.PropertyID)]
		if !ok || prop.Origin != vocabkey.OriginProvenencia {
			continue
		}
		v, err := Decode(o.Scalar, o.TermID, o.Date, o.Name, o.DepictionImageFileID, terms)
		if err != nil {
			return nil, err
		}
		key := vocabkey.PropertyKey(prop.Key)
		if out[string(o.SubjectID)] == nil {
			out[string(o.SubjectID)] = map[vocabkey.PropertyKey][]autoreconcile.Candidate{}
		}
		out[string(o.SubjectID)][key] = append(out[string(o.SubjectID)][key], autoreconcile.Candidate{
			ObservationID: bytes.Clone(o.ID),
			Value:         v,
			Negative:      o.Polarity == observations.PolarityNegative,
		})
	}
	return out, nil
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
