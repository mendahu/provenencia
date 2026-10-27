package handlers

import (
	"slices"
	"strings"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/claimevidence"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/conclusiondetails"
	"github.com/mendahu/provenencia/core/database/conclusionheaders"
	"github.com/mendahu/provenencia/core/database/promotealign"
	"github.com/mendahu/provenencia/core/kinship"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/vocabkey"
	"google.golang.org/protobuf/proto"
)

// GetPromoteClaimEvidence reads one Promote row's claim evidence against its
// target (S11-25): the lines, the relationships, the pins on offer and the
// Sources they name. The neighbors are the targets the page holds for its
// other rows.
func GetPromoteClaimEvidence(in []byte) ([]byte, error) {
	var req engine.GetPromoteClaimEvidenceRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("get_promote_claim_evidence", err)
	}
	graphID, err := parseID(req.GetGraphId())
	if err != nil {
		return nil, err
	}
	subjectID, err := parseID(req.GetSubjectId())
	if err != nil {
		return nil, err
	}
	handleID, err := optionalID(req.GetHandleId())
	if err != nil {
		return nil, err
	}
	neighbors, err := fixedFromProto(req.GetNeighbors())
	if err != nil {
		return nil, err
	}

	var out *engine.GetPromoteClaimEvidenceResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		e, err := promotealign.Evidence(c, promotealign.EvidenceRequest{
			GraphID: graphID, SubjectID: subjectID, HandleID: handleID, Neighbors: neighbors,
		})
		if err != nil {
			return err
		}
		values, err := promotealign.NewEvidenceValues(c.Graph())
		if err != nil {
			return err
		}
		if err := values.ResolvePaths(c, e); err != nil {
			return err
		}
		out, err = claimEvidenceProto(c, e, values)
		return err
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func claimEvidenceProto(c *database.Catalog, e claimevidence.Evidence, values promotealign.EvidenceValues) (*engine.GetPromoteClaimEvidenceResponse, error) {
	out := &engine.GetPromoteClaimEvidenceResponse{Sections: evidenceSectionsProto(e.Kind, e.Sections, values)}
	for _, rel := range e.Relationships {
		r := &engine.PromoteClaimEvidenceRelationship{
			NeighborSubjectId:         database.UUIDString(rel.Neighbor),
			Key:                       rel.Key.String(),
			BridgeType:                string(rel.Key.Type),
			Outcome:                   string(rel.Outcome),
			Sections:                  evidenceSectionsProto(rel.Kind, rel.Sections, values),
			NeighborEnd:               string(rel.NeighborEnd),
			NeighborTermNamesNeighbor: rel.NeighborTermNamesNeighbor,
			Layer:                     evidenceShapeProto(rel.Layer),
			Canon:                     evidenceShapeProto(rel.Canon),
		}
		if rel.NeighborTerm != "" {
			r.NeighborTerm = evidenceTermProto(values, string(rel.NeighborTermProperty), string(rel.NeighborTerm))
		}
		if rel.Key.EventType != "" {
			r.EventType = evidenceTermProto(values, vocabkey.PropertyEventType, string(rel.Key.EventType))
		}
		for _, b := range rel.Members {
			r.Bundles = append(r.Bundles, &engine.PromoteClaimEvidenceBundle{Id: promotealign.BundlePinID(rel, b)})
		}
		out.Relationships = append(out.Relationships, r)
	}
	for _, o := range promotealign.PinOffers(e) {
		out.Offers = append(out.Offers, &engine.PromoteClaimEvidenceOffer{Id: o.ID, Checked: o.Checked})
	}
	for id, title := range e.Sources {
		out.Sources = append(out.Sources, &engine.PromoteClaimEvidenceSource{Id: database.UUIDString([]byte(id)), Title: title})
	}
	slices.SortFunc(out.Sources, func(a, b *engine.PromoteClaimEvidenceSource) int {
		return strings.Compare(a.GetId(), b.GetId())
	})
	if err := canonPivotHeaders(c, out.Relationships, e.Relationships); err != nil {
		return nil, err
	}
	return out, nil
}

// evidenceShapeProto is one side's shape of a relationship line, without the
// pivot's header.
func evidenceShapeProto(s kinship.Shape) *engine.PromoteClaimEvidenceShape {
	out := &engine.PromoteClaimEvidenceShape{Pattern: string(s.Pattern)}
	if len(s.Pivot) == 16 {
		out.PivotId = database.UUIDString(s.Pivot)
	}
	return out
}

// canonPivotHeaders sets the header of each entity-side pivot on the mapped
// relationships (parallel to rels): the pivot is a handle, an event or a
// person by its pattern, and every header is read in one call.
func canonPivotHeaders(c *database.Catalog, mapped []*engine.PromoteClaimEvidenceRelationship, rels []claimevidence.Relationship) error {
	var persons, events [][]byte
	for _, rel := range rels {
		switch rel.Canon.Pattern.PivotKind() {
		case vocabkey.Person:
			persons = append(persons, rel.Canon.Pivot)
		case vocabkey.Event:
			events = append(events, rel.Canon.Pivot)
		}
	}
	if len(persons)+len(events) == 0 {
		return nil
	}
	headers, err := conclusionheaders.ByIDs(c, persons, events, nil)
	if err != nil {
		return err
	}
	for i, rel := range rels {
		switch rel.Canon.Pattern.PivotKind() {
		case vocabkey.Person:
			if h, ok := headers.Person(rel.Canon.Pivot); ok {
				mapped[i].Canon.PivotHeader = &engine.PromoteClaimEvidenceShape_Person{Person: personHeaderProto(h)}
			}
		case vocabkey.Event:
			if h, ok := headers.Event(rel.Canon.Pivot); ok {
				mapped[i].Canon.PivotHeader = &engine.PromoteClaimEvidenceShape_Event{Event: eventHeaderProto(h)}
			}
		}
	}
	return nil
}

// evidenceSectionsProto maps sections whose values belong to a Subject of
// kind.
func evidenceSectionsProto(kind vocabkey.SubjectType, sections []claimevidence.Section, values promotealign.EvidenceValues) []*engine.PromoteClaimEvidenceSection {
	var out []*engine.PromoteClaimEvidenceSection
	for _, s := range sections {
		sec := &engine.PromoteClaimEvidenceSection{PropertyKey: string(s.Key)}
		for _, l := range s.Lines {
			line := &engine.PromoteClaimEvidenceLine{
				PropertyKey:           string(l.Property),
				Outcome:               string(l.Outcome),
				PropertyCause:         string(l.PropertyCause),
				Relation:              string(l.Relation),
				Cause:                 string(l.Cause),
				IncomingObservationId: database.UUIDString(l.Incoming.Ref),
				Incoming:              evidenceValueProto(values.Incoming(kind, s, l)),
				IncomingNegative:      l.Incoming.Negative,
			}
			for _, p := range l.Members {
				line.Members = append(line.Members, &engine.PromoteClaimEvidencePin{
					Id:       promotealign.LinePinID(l, p),
					SourceId: database.UUIDString(p.Source),
					Value:    evidenceValueProto(values.Member(kind, s, l, p)),
				})
			}
			sec.Lines = append(sec.Lines, line)
		}
		out = append(out, sec)
	}
	return out
}

// evidenceTermProto is a term of a property with the label and origin the
// catalog holds, so the client words a researcher's term by its label and a
// product term by its key.
func evidenceTermProto(values promotealign.EvidenceValues, property, key string) *engine.ConclusionTerm {
	label, origin := values.Term(property, key)
	return &engine.ConclusionTerm{Key: key, Label: label, Origin: origin}
}

// evidenceValueProto is a sheet value as a ConclusionValue, the one wire shape
// for a typed value. A term has no id here: its key and label name it.
func evidenceValueProto(v promotealign.EvidenceValue) *engine.ConclusionValue {
	var d conclusiondetails.Value
	var depiction *engine.DepictionValue
	switch v.Type {
	case propertymeta.ValueTypeDepiction:
		depiction = &engine.DepictionValue{ImageRelPath: v.ImageRelPath}
	case propertymeta.ValueTypeName:
		d.Name = v.Value.Name
	case propertymeta.ValueTypeDate:
		d.Date = v.Value.Date
	case propertymeta.ValueTypeTerm:
		d.TermKey, d.TermLabel = v.Value.Term, v.TermLabel
	case propertymeta.ValueTypeInteger:
		d.HasInteger, d.Integer = true, v.Value.Integer
	case propertymeta.ValueTypeText:
		d.HasText, d.Text = true, v.Value.Text
	}
	return conclusionValueProto(d, depiction)
}
