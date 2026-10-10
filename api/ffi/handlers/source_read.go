package handlers

import (
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/valuecodec"
)

type propertyInfo struct {
	Key       string
	Label     string
	ValueType string
	Origin    string
}

type termInfo struct {
	Key   string
	Label string
}

func loadPropertyInfo(c *database.Catalog) (map[string]propertyInfo, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(`SELECT id, key, label, value_type, origin FROM properties`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]propertyInfo{}
	for rows.Next() {
		var id []byte
		var info propertyInfo
		if err := rows.Scan(&id, &info.Key, &info.Label, &info.ValueType, &info.Origin); err != nil {
			return nil, err
		}
		out[string(id)] = info
	}
	return out, rows.Err()
}

func loadTermInfo(c *database.Catalog) (map[string]termInfo, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(`SELECT id, key, label FROM property_terms`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]termInfo{}
	for rows.Next() {
		var id []byte
		var info termInfo
		if err := rows.Scan(&id, &info.Key, &info.Label); err != nil {
			return nil, err
		}
		out[string(id)] = info
	}
	return out, rows.Err()
}

func listedFromSource(o graphcache.SourceObservation, props map[string]propertyInfo) observations.Listed {
	info := props[string(o.PropertyID)]
	l := observations.Listed{
		Observation: observations.Observation{
			ID: o.ID, Ref: o.Ref, CitationID: o.CitationID, SubjectID: o.SubjectID,
			PropertyID: o.PropertyID, Polarity: o.Polarity,
			ValueText: o.Text, HasText: o.HasText,
			ValueInteger: o.Integer, HasInteger: o.HasInteger,
			ValueDateID: o.DateID, ValueNameID: o.NameID,
			ValueSubjectID: o.ValueSubjectID, ValueTermID: o.TermID,
		},
		PropertyKey: info.Key, PropertyLabel: info.Label, PropertyValueType: info.ValueType,
	}
	if len(o.Date) > 0 {
		if d, err := valuecodec.UnmarshalDate(o.Date); err == nil {
			l.Date = &d
		}
	}
	if len(o.Name) > 0 {
		if n, err := valuecodec.UnmarshalName(o.Name); err == nil {
			l.Name = &n
			l.ValueNameForm = n.Form
		}
	}
	return l
}

func membershipsFromSource(g *graphcache.Graph, sg graphcache.SourceGraph, nameProperty []byte) ([]identityclaims.Membership, error) {
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

func keptName(n *graphcache.Node, propertyID []byte) *namevalues.Value {
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
		if info.Key == "name" && info.Origin == "provenencia" {
			return []byte(id)
		}
	}
	return nil
}
