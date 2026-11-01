package handlers

import (
	"bytes"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/graphcache"
	_ "github.com/mendahu/provenencia/core/database/graphfill"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/depictionvalue"
	"github.com/mendahu/provenencia/core/namevalue"
	"github.com/mendahu/provenencia/core/valuecodec"
	"github.com/mendahu/provenencia/core/vocabkey"
)

type propertyInfo struct {
	Key       string
	Label     string
	ValueType string
	Origin    string
}

func loadPropertyInfo(c *database.Catalog) (map[string]propertyInfo, error) {
	if c == nil || c.Graph() == nil {
		return map[string]propertyInfo{}, nil
	}
	v, err := c.Graph().Vocabulary()
	if err != nil {
		return nil, err
	}
	out := map[string]propertyInfo{}
	for _, p := range v.Properties {
		out[string(p.ID)] = propertyInfo{Key: p.Key, Label: p.Label, ValueType: p.ValueType, Origin: p.Origin}
	}
	return out, nil
}

// listedContext is the vocabulary a response maps its Observations with, read
// once per response.
type listedContext struct {
	props map[string]propertyInfo
	terms map[string]graphcache.VocabTerm
}

func newListedContext(c *database.Catalog) (listedContext, error) {
	props, err := loadPropertyInfo(c)
	if err != nil {
		return listedContext{}, err
	}
	terms := map[string]graphcache.VocabTerm{}
	if c != nil && c.Graph() != nil {
		v, err := c.Graph().Vocabulary()
		if err != nil {
			return listedContext{}, err
		}
		terms = v.TermsByID()
	}
	return listedContext{props: props, terms: terms}, nil
}

// rows projects the graph's stored Observations that keep accepts, in the
// graph's ref order. Subject labels, for Observations whose value is another
// Subject, come from the graph's Subjects.
func (lc listedContext) rows(eg graphcache.EvidenceGraph, keep func(graphcache.EvidenceObservation) bool) ([]observations.Listed, error) {
	subjectLabels := map[string]string{}
	for _, s := range eg.Subjects {
		subjectLabels[string(s.ID)] = s.Label
	}
	var rows []observations.Listed
	for _, o := range eg.Observations {
		if !keep(o) {
			continue
		}
		listed, err := listedFromSource(o, lc.props, lc.terms, subjectLabels)
		if err != nil {
			return nil, err
		}
		rows = append(rows, listed)
	}
	return rows, nil
}

// listedProtosFromGraph projects stored observations. A citation id keeps
// that citation's rows, in the graph's ref order. Nil keeps every row.
func listedProtosFromGraph(c *database.Catalog, eg graphcache.EvidenceGraph, citationID []byte) ([]*engine.Observation, error) {
	lc, err := newListedContext(c)
	if err != nil {
		return nil, err
	}
	rows, err := lc.rows(eg, func(o graphcache.EvidenceObservation) bool {
		return citationID == nil || bytes.Equal(o.CitationID, citationID)
	})
	if err != nil {
		return nil, err
	}
	return listedProtos(c, rows)
}

func listedFromSource(o graphcache.EvidenceObservation, props map[string]propertyInfo, terms map[string]graphcache.VocabTerm, subjectLabels map[string]string) (observations.Listed, error) {
	info := props[string(o.PropertyID)]
	l := observations.Listed{
		Observation: observations.Observation{
			ID: o.ID, Ref: o.Ref, CitationID: o.CitationID, SubjectID: o.SubjectID,
			PropertyID: o.PropertyID, Polarity: o.Polarity,
			ValueText: o.Text, HasText: o.HasText,
			ValueInteger: o.Integer, HasInteger: o.HasInteger,
			ValueDateID: o.DateID, ValueNameID: o.NameID,
			ValueSubjectID: o.ValueSubjectID, ValueTermID: o.TermID,
			ValueDepictionID: o.DepictionID,
		},
		PropertyKey: info.Key, PropertyLabel: info.Label, PropertyValueType: info.ValueType,
	}
	if term, ok := terms[string(o.TermID)]; ok {
		l.ValueTermLabel = term.Label
		l.ValueTermKey = term.Key
		l.ValueTermOrigin = term.Origin
	}
	if label, ok := subjectLabels[string(o.ValueSubjectID)]; ok {
		l.ValueSubjectLabel = label
	}
	// Matches depictionvalues.FromJoin's non-nil result: depiction_values.image_file_id is NOT NULL.
	if len(o.DepictionID) > 0 {
		l.Depiction = &depictionvalue.Value{ID: o.DepictionID, RegionJSON: o.DepictionRegion, ImageFileID: o.DepictionImageFileID}
	}
	if len(o.Date) > 0 {
		d, err := valuecodec.UnmarshalDate(o.Date)
		if err != nil {
			return observations.Listed{}, err
		}
		l.Date = &d
	}
	if len(o.Name) > 0 {
		n, err := valuecodec.UnmarshalName(o.Name)
		if err != nil {
			return observations.Listed{}, err
		}
		l.Name = &n
		l.ValueNameForm = n.Form
	}
	observations.FillListedDisplayText(&l)
	return l, nil
}

func membershipsFromGraph(g *graphcache.Graph, eg graphcache.EvidenceGraph, nameProperty []byte) ([]identityclaims.Membership, error) {
	ids := make([][]byte, 0, len(eg.Members))
	for _, m := range eg.Members {
		ids = append(ids, m.EntityID)
	}
	if err := g.Nodes(ids); err != nil {
		return nil, err
	}
	var out []identityclaims.Membership
	for _, m := range eg.Members {
		n, err := g.Node(m.EntityID)
		if err != nil {
			return nil, err
		}
		if n == nil {
			continue
		}
		out = append(out, identityclaims.Membership{
			SubjectID: m.SubjectID,
			ClaimID:   m.ClaimID,
			Kind:      m.Kind,
			Entity: canonicalentities.Entity{
				ID: append([]byte(nil), n.ID...), SubjectTypeID: append([]byte(nil), n.SubjectTypeID...),
				Ref: n.Ref, Argument: n.Argument, Label: n.Label,
			},
			Name: keptName(n, nameProperty),
		})
	}
	return out, nil
}

func keptName(n *graphcache.Node, propertyID []byte) *namevalue.Value {
	if n == nil || len(propertyID) != 16 {
		return nil
	}
	for _, v := range n.Values[string(propertyID)] {
		if v.Rank == 1 && v.Reason == "kept" && len(v.Name) > 0 {
			decoded, err := valuecodec.UnmarshalName(v.Name)
			if err != nil {
				return nil
			}
			return &decoded
		}
	}
	return nil
}

func namePropertyID(props map[string]propertyInfo) []byte {
	for id, info := range props {
		if info.Key == vocabkey.PropertyName && info.Origin == vocabkey.OriginProvenencia {
			return []byte(id)
		}
	}
	return nil
}
