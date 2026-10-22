package handlers

import (
	"slices"
	"strings"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/claimevidence"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/conclusiondetails"
	"github.com/mendahu/provenencia/core/database/promotealign"
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
	sourceID, err := parseID(req.GetSourceId())
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
			SourceID: sourceID, SubjectID: subjectID, HandleID: handleID, Neighbors: neighbors,
		})
		if err != nil {
			return err
		}
		values, err := promotealign.NewEvidenceValues(c.Graph())
		if err != nil {
			return err
		}
		out = claimEvidenceProto(e, values)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func claimEvidenceProto(e claimevidence.Evidence, values promotealign.EvidenceValues) *engine.GetPromoteClaimEvidenceResponse {
	out := &engine.GetPromoteClaimEvidenceResponse{Sections: evidenceSectionsProto(e.Kind, e.Sections, values)}
	for _, rel := range e.Relationships {
		r := &engine.PromoteClaimEvidenceRelationship{
			NeighborSubjectId: database.UUIDString(rel.Neighbor),
			Key:               rel.Key.String(),
			BridgeType:        string(rel.Key.Type),
			Term:              string(rel.Key.Term),
			End:               string(rel.Key.End),
			EventType:         string(rel.Key.EventType),
			Outcome:           string(rel.Outcome),
			Cause:             string(rel.Cause),
			Points:            rel.Points,
			Pinnable:          rel.Pinnable,
			Checked:           rel.Checked,
			Sections:          evidenceSectionsProto(rel.Kind, rel.Sections, values),
		}
		for _, b := range rel.Members {
			r.Bundles = append(r.Bundles, &engine.PromoteClaimEvidenceBundle{
				Id:              promotealign.BundlePinID(rel, b),
				MemberSubjectId: database.UUIDString(b.Member),
				SourceId:        database.UUIDString(b.Source),
			})
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
	return out
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
				Points:                l.Points,
				Pinnable:              l.Pinnable,
				Checked:               l.Checked,
				IncomingObservationId: database.UUIDString(l.Incoming.Ref),
				Incoming:              evidenceValueProto(values.Incoming(kind, s, l)),
				IncomingNegative:      l.Incoming.Negative,
			}
			for _, p := range l.Members {
				line.Members = append(line.Members, &engine.PromoteClaimEvidencePin{
					Id:                  promotealign.LinePinID(l, p),
					MemberSubjectId:     database.UUIDString(p.Subject),
					MemberObservationId: database.UUIDString(p.Observation),
					SourceId:            database.UUIDString(p.Source),
					Value:               evidenceValueProto(values.Member(kind, s, l, p)),
				})
			}
			sec.Lines = append(sec.Lines, line)
		}
		out = append(out, sec)
	}
	return out
}

// evidenceValueProto is a sheet value as a ConclusionValue, the one wire shape
// for a typed value. A term has no id here: its key and label name it.
func evidenceValueProto(v promotealign.EvidenceValue) *engine.ConclusionValue {
	var d conclusiondetails.Value
	switch v.Type {
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
	return conclusionValueProto(d)
}
