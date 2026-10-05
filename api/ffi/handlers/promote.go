package handlers

import (
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/claimconfidencegrades"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/promotecompare"
	"github.com/mendahu/provenencia/core/database/promotetargets"
	"github.com/mendahu/provenencia/core/valuecodec"
	"google.golang.org/protobuf/proto"
)

func PromoteSubject(in []byte) ([]byte, error) {
	var req engine.PromoteSubjectRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("promote_subject", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	subjectID, err := parseID(req.GetSubjectId())
	if err != nil {
		return nil, err
	}
	entityID, err := optionalID(req.GetEntityId())
	if err != nil {
		return nil, err
	}
	gradeID, err := optionalID(req.GetConfidenceGradeId())
	if err != nil {
		return nil, err
	}
	var pairs []promote.Pair
	for _, p := range req.GetPairs() {
		incoming, err := parseID(p.GetIncomingObservationId())
		if err != nil {
			return nil, err
		}
		member, err := parseID(p.GetMemberObservationId())
		if err != nil {
			return nil, err
		}
		pairs = append(pairs, promote.Pair{IncomingObservationID: incoming, MemberObservationID: member})
	}
	var out *engine.PromoteSubjectResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		res, err := promote.Save(c, userID, promote.Input{
			SubjectID:         subjectID,
			EntityID:          entityID,
			ConfidenceGradeID: gradeID,
			Argument:          req.GetArgument(),
			Pairs:             pairs,
		})
		if err != nil {
			return err
		}
		out = &engine.PromoteSubjectResponse{
			Entity:   canonicalEntityProto(res.Entity),
			Claim:    identityClaimProto(res.Claim),
			PinCount: int32(res.Pins),
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func ListPromoteComparison(in []byte) ([]byte, error) {
	var req engine.ListPromoteComparisonRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("list_promote_comparison", err)
	}
	subjectID, err := parseID(req.GetSubjectId())
	if err != nil {
		return nil, err
	}
	entityID, err := parseID(req.GetEntityId())
	if err != nil {
		return nil, err
	}
	out := &engine.ListPromoteComparisonResponse{}
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		db, err := c.DB()
		if err != nil {
			return err
		}
		got, err := promotecompare.Compare(db, subjectID, entityID)
		if err != nil {
			return err
		}
		out.MemberCount = int32(got.Members)
		for _, p := range got.Properties {
			pp := &engine.PromoteComparisonProperty{
				PropertyId:  uuidString(p.PropertyID),
				PropertyKey: p.Key,
				Label:       p.Label,
				ValueType:   p.ValueType,
			}
			for _, inc := range p.Incoming {
				pi := &engine.PromoteComparisonIncoming{Record: comparisonRecordProto(inc.Record)}
				for _, pr := range inc.Pairs {
					pi.Pairs = append(pi.Pairs, &engine.PromoteComparisonPair{
						Member:     comparisonRecordProto(pr.Member),
						Compatible: pr.Compatible,
					})
				}
				pp.Incoming = append(pp.Incoming, pi)
			}
			out.Properties = append(out.Properties, pp)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func comparisonRecordProto(r promotecompare.Record) *engine.PromoteComparisonRecord {
	return &engine.PromoteComparisonRecord{
		ObservationId:  uuidString(r.ObservationID),
		ObservationRef: r.ObservationRef,
		SubjectId:      uuidString(r.SubjectID),
		SubjectRef:     r.SubjectRef,
		SubjectLabel:   r.SubjectLabel,
		ClaimId:        uuidString(r.ClaimID),
		CitationId:     uuidString(r.CitationID),
		ArtifactId:     uuidString(r.ArtifactID),
		SourceId:       uuidString(r.SourceID),
		SourceTitle:    r.SourceTitle,
		Negative:       r.Negative,
		Value:          conclusionValueProto(r.Value),
	}
}

func ListPromoteTargetSuggestions(in []byte) ([]byte, error) {
	var req engine.ListPromoteTargetSuggestionsRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("list_promote_target_suggestions", err)
	}
	subjectID, err := parseID(req.GetSubjectId())
	if err != nil {
		return nil, err
	}
	out := &engine.ListPromoteTargetSuggestionsResponse{}
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		db, err := c.DB()
		if err != nil {
			return err
		}
		got, err := promotetargets.Suggest(db, subjectID, int(req.GetLimit()))
		if err != nil {
			return err
		}
		for _, sg := range got {
			ps := &engine.PromoteTargetSuggestion{
				Entity:      canonicalEntityProto(sg.Entity),
				Score:       sg.Score,
				MemberCount: int32(sg.MemberCount),
			}
			for _, r := range sg.Reasons {
				ps.Reasons = append(ps.Reasons, &engine.MatchReason{
					PropertyKey:    r.Property.Key,
					PropertyOrigin: r.Property.Origin,
					Similarity:     r.Similarity,
					Contribution:   r.Contribution,
				})
			}
			if sg.Person != nil {
				ps.Person = personHeaderProto(*sg.Person)
			}
			out.Suggestions = append(out.Suggestions, ps)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func ListClaimConfidenceGrades(in []byte) ([]byte, error) {
	var req engine.ListClaimConfidenceGradesRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("list_claim_confidence_grades", err)
	}
	out := &engine.ListClaimConfidenceGradesResponse{}
	err := withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		rows, err := claimconfidencegrades.List(c)
		if err != nil {
			return err
		}
		for _, g := range rows {
			out.Grades = append(out.Grades, &engine.ClaimConfidenceGrade{
				Id:        uuidString(g.ID),
				Key:       g.Key,
				Origin:    g.Origin,
				Label:     g.Label,
				SortOrder: int32(g.SortOrder),
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func ListSubjectMemberships(in []byte) ([]byte, error) {
	var req engine.ListSubjectMembershipsRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("list_subject_memberships", err)
	}
	sourceID, err := parseID(req.GetSourceId())
	if err != nil {
		return nil, err
	}
	var out *engine.ListSubjectMembershipsResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		rows, err := identityclaims.MembershipsBySource(c, sourceID)
		if err != nil {
			return err
		}
		out = &engine.ListSubjectMembershipsResponse{}
		for _, m := range rows {
			sm := &engine.SubjectMembership{
				SubjectId: uuidString(m.SubjectID),
				ClaimId:   uuidString(m.ClaimID),
				Entity:    canonicalEntityProto(m.Entity),
				Kind:      m.Kind,
			}
			if m.Name != nil {
				sm.Name = valuecodec.NameToProto(*m.Name)
			}
			out.Memberships = append(out.Memberships, sm)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func canonicalEntityProto(e canonicalentities.Entity) *engine.CanonicalEntity {
	return &engine.CanonicalEntity{
		Id:            uuidString(e.ID),
		Ref:           e.Ref,
		SubjectTypeId: uuidString(e.SubjectTypeID),
		Label:         e.Label,
	}
}

func identityClaimProto(c identityclaims.Claim) *engine.IdentityClaim {
	return &engine.IdentityClaim{
		Id:                uuidString(c.ID),
		SubjectId:         uuidString(c.SubjectID),
		EntityId:          uuidString(c.EntityID),
		Status:            c.Status,
		ConfidenceGradeId: uuidString(c.ConfidenceGradeID),
		Argument:          c.Argument,
	}
}
