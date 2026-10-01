package handlers

import (
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/promote"
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
	var out *engine.PromoteSubjectResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		res, err := promote.Save(c, userID, promote.Input{SubjectID: subjectID})
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
			out.Memberships = append(out.Memberships, &engine.SubjectMembership{
				SubjectId: uuidString(m.SubjectID),
				Entity:    canonicalEntityProto(m.Entity),
				Kind:      m.Kind,
			})
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
		Id:        uuidString(c.ID),
		SubjectId: uuidString(c.SubjectID),
		EntityId:  uuidString(c.EntityID),
		Status:    c.Status,
	}
}
