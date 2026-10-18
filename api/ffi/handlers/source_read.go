package handlers

import (
	"bytes"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/observations"
	_ "github.com/mendahu/provenencia/core/database/sourcefill"
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

// listedProtosFromSource projects stored observations. A citation id keeps
// that citation's rows, in the Source's ref order. Nil keeps every row.
func listedProtosFromSource(c *database.Catalog, sg graphcache.SourceGraph, citationID []byte) ([]*engine.Observation, error) {
	props, err := loadPropertyInfo(c)
	if err != nil {
		return nil, err
	}
	terms := map[string]graphcache.VocabTerm{}
	if c != nil && c.Graph() != nil {
		v, err := c.Graph().Vocabulary()
		if err != nil {
			return nil, err
		}
		terms = v.TermsByID()
	}
	subjectLabels := map[string]string{}
	for _, s := range sg.Subjects {
		subjectLabels[string(s.ID)] = s.Label
	}
	var rows []observations.Listed
	for _, o := range sg.Observations {
		if citationID != nil && !bytes.Equal(o.CitationID, citationID) {
			continue
		}
		listed, err := listedFromSource(o, props, terms, subjectLabels)
		if err != nil {
			return nil, err
		}
		rows = append(rows, listed)
	}
	return listedProtos(c, rows)
}

func listedFromSource(o graphcache.SourceObservation, props map[string]propertyInfo, terms map[string]graphcache.VocabTerm, subjectLabels map[string]string) (observations.Listed, error) {
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

func membershipsFromSource(g *graphcache.Graph, sg graphcache.SourceGraph, nameProperty []byte) ([]identityclaims.Membership, error) {
	ids := make([][]byte, 0, len(sg.Members))
	for _, m := range sg.Members {
		ids = append(ids, m.EntityID)
	}
	if err := g.Nodes(ids); err != nil {
		return nil, err
	}
	var out []identityclaims.Membership
	for _, m := range sg.Members {
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
