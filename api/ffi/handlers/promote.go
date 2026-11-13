package handlers

import (
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/graphfill"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/valuecodec"
	"github.com/mendahu/provenencia/core/writes"
	"google.golang.org/protobuf/proto"
)

func GetUnpromoteImpact(in []byte) ([]byte, error) {
	var req engine.GetUnpromoteImpactRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("get_unpromote_impact", err)
	}
	subjectID, err := parseID(req.GetSubjectId())
	if err != nil {
		return nil, err
	}
	var out *engine.GetUnpromoteImpactResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		report, err := promote.UnpromoteImpact(c, subjectID)
		if err != nil {
			return err
		}
		out = &engine.GetUnpromoteImpactResponse{UnpromotedBridgeSubjectIds: idStrings(report.UnpromotedBridgeSubjects)}
		if report.Deletes.Via != "" {
			out.Deletes = groupToProto(report.Deletes)
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
		eg, err := graphfill.EvidenceGraph(c.Graph(), graphID)
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
