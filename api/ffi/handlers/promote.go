package handlers

import (
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/claimconfidencegrades"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/valuecodec"
	"github.com/mendahu/provenencia/core/writes"
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
	var out *engine.PromoteSubjectResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		res, _, err := writes.Run(c, writes.Op{Action: "promote_subject", UserID: userID}, func(tx *database.Tx) (promote.Result, []rowchange.Change, error) {
			return promote.Save(tx, userID, promote.Input{
				SubjectID:         subjectID,
				EntityID:          entityID,
				ConfidenceGradeID: gradeID,
				Argument:          req.GetArgument(),
			})
		})
		if err != nil {
			return err
		}
		out = &engine.PromoteSubjectResponse{
			Entity: canonicalEntityProto(res.Entity),
			Claim:  identityClaimProto(res.Claim),
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func UnpromoteSubject(in []byte) ([]byte, error) {
	var req engine.UnpromoteSubjectRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("unpromote_subject", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	subjectID, err := parseID(req.GetSubjectId())
	if err != nil {
		return nil, err
	}
	var out *engine.UnpromoteSubjectResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		_, wrote, err := writes.Run(c, writes.Op{Action: "unpromote_subject", UserID: userID},
			func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
				changes, err := promote.Unpromote(tx, userID, subjectID)
				return struct{}{}, changes, err
			})
		if err != nil {
			return err
		}
		out = &engine.UnpromoteSubjectResponse{Effects: writeEffectsProto(wrote)}
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
				Id:        database.UUIDString(g.ID),
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
	graphID, err := parseID(req.GetGraphId())
	if err != nil {
		return nil, err
	}
	var out *engine.ListSubjectMembershipsResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		eg, err := c.Graph().EvidenceGraph(graphID)
		if err != nil {
			return err
		}
		props, err := loadPropertyInfo(c)
		if err != nil {
			return err
		}
		rows, err := membershipsFromGraph(c.Graph(), eg, namePropertyID(props))
		if err != nil {
			return err
		}
		out = &engine.ListSubjectMembershipsResponse{}
		for _, m := range rows {
			sm := &engine.SubjectMembership{
				SubjectId: database.UUIDString(m.SubjectID),
				ClaimId:   database.UUIDString(m.ClaimID),
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
		Id:            database.UUIDString(e.ID),
		Ref:           e.Ref,
		SubjectTypeId: database.UUIDString(e.SubjectTypeID),
		Label:         e.Label,
	}
}

func identityClaimProto(c identityclaims.Claim) *engine.IdentityClaim {
	return &engine.IdentityClaim{
		Id:                database.UUIDString(c.ID),
		SubjectId:         database.UUIDString(c.SubjectID),
		EntityId:          database.UUIDString(c.EntityID),
		Status:            c.Status,
		ConfidenceGradeId: database.UUIDString(c.ConfidenceGradeID),
		Argument:          c.Argument,
	}
}
