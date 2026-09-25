package handlers

import (
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/subjectpositions"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"google.golang.org/protobuf/proto"
)

func ListSubjectTypes(in []byte) ([]byte, error) {
	var req engine.ListSubjectTypesRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("list_subject_types", err)
	}
	var out *engine.ListSubjectTypesResponse
	err := withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		rows, err := subjecttypes.List(c)
		if err != nil {
			return err
		}
		out = &engine.ListSubjectTypesResponse{}
		for _, t := range rows {
			out.Types = append(out.Types, subjectTypeProto(t))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func CreateSubject(in []byte) ([]byte, error) {
	var req engine.CreateSubjectRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("create_subject", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	sourceID, err := parseID(req.GetSourceId())
	if err != nil {
		return nil, err
	}
	typeID, err := parseID(req.GetSubjectTypeId())
	if err != nil {
		return nil, err
	}
	var out *engine.CreateSubjectResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		s, err := subjects.Create(c, userID, subjects.CreateInput{
			SourceID:      sourceID,
			SubjectTypeID: typeID,
			Label:         req.GetLabel(),
			Description:   req.GetDescription(),
		}, nil)
		if err != nil {
			return err
		}
		out = &engine.CreateSubjectResponse{Subject: subjectProto(s)}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func UpdateSubject(in []byte) ([]byte, error) {
	var req engine.UpdateSubjectRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("update_subject", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	subjectID, err := parseID(req.GetSubjectId())
	if err != nil {
		return nil, err
	}
	var out *engine.UpdateSubjectResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		if err := subjects.Update(c, userID, subjectID, req.GetLabel(), req.GetDescription()); err != nil {
			return err
		}
		s, err := subjects.Get(c, subjectID)
		if err != nil {
			return err
		}
		out = &engine.UpdateSubjectResponse{Subject: subjectProto(s)}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func DeleteSubject(in []byte) ([]byte, error) {
	var req engine.DeleteSubjectRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("delete_subject", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	subjectID, err := parseID(req.GetSubjectId())
	if err != nil {
		return nil, err
	}
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		return subjects.Delete(c, userID, subjectID)
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(&engine.DeleteSubjectResponse{})
}

func ListSubjects(in []byte) ([]byte, error) {
	var req engine.ListSubjectsRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("list_subjects", err)
	}
	sourceID, err := parseID(req.GetSourceId())
	if err != nil {
		return nil, err
	}
	var out *engine.ListSubjectsResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		rows, err := subjects.ListBySource(c, sourceID)
		if err != nil {
			return err
		}
		out = &engine.ListSubjectsResponse{}
		for _, s := range rows {
			out.Subjects = append(out.Subjects, subjectProto(s))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func SetSubjectPosition(in []byte) ([]byte, error) {
	var req engine.SetSubjectPositionRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("set_subject_position", err)
	}
	subjectID, err := parseID(req.GetSubjectId())
	if err != nil {
		return nil, err
	}
	var out *engine.SetSubjectPositionResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		p, err := subjectpositions.Set(c, subjectID, req.GetGridX(), req.GetGridY())
		if err != nil {
			return err
		}
		out = &engine.SetSubjectPositionResponse{Position: subjectPositionProto(p)}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func ClearSubjectPosition(in []byte) ([]byte, error) {
	var req engine.ClearSubjectPositionRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("clear_subject_position", err)
	}
	subjectID, err := parseID(req.GetSubjectId())
	if err != nil {
		return nil, err
	}
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		return subjectpositions.Clear(c, subjectID)
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(&engine.ClearSubjectPositionResponse{})
}

func ListSubjectPositions(in []byte) ([]byte, error) {
	var req engine.ListSubjectPositionsRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("list_subject_positions", err)
	}
	sourceID, err := parseID(req.GetSourceId())
	if err != nil {
		return nil, err
	}
	var out *engine.ListSubjectPositionsResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		rows, err := subjectpositions.ListBySource(c, sourceID)
		if err != nil {
			return err
		}
		out = &engine.ListSubjectPositionsResponse{}
		for _, p := range rows {
			out.Positions = append(out.Positions, subjectPositionProto(p))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func subjectTypeProto(t subjecttypes.Type) *engine.SubjectType {
	return &engine.SubjectType{
		Id:                 uuidString(t.ID),
		Key:                t.Key,
		Origin:             t.Origin,
		Label:              t.Label,
		Description:        t.Description,
		RefPrefix:          t.RefPrefix,
		CandidateRefPrefix: t.CandidateRefPrefix,
	}
}

func subjectProto(s subjects.Subject) *engine.Subject {
	return &engine.Subject{
		Id:            uuidString(s.ID),
		Ref:           s.Ref,
		SourceId:      uuidString(s.SourceID),
		SubjectTypeId: uuidString(s.SubjectTypeID),
		Label:         s.Label,
		Description:   s.Description,
	}
}

func subjectPositionProto(p subjectpositions.Position) *engine.SubjectPosition {
	return &engine.SubjectPosition{
		SubjectId: uuidString(p.SubjectID),
		GridX:     p.GridX,
		GridY:     p.GridY,
	}
}
