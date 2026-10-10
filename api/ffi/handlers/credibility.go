package handlers

import (
	"database/sql"
	"errors"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sourcecredibility"
	"github.com/mendahu/provenencia/core/database/sourcecredibilitygrades"
	"github.com/mendahu/provenencia/core/writes"
	"google.golang.org/protobuf/proto"
)

func ListSourceCredibilityGrades(in []byte) ([]byte, error) {
	var req engine.ListSourceCredibilityGradesRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("list_source_credibility_grades", err)
	}
	var out *engine.ListSourceCredibilityGradesResponse
	err := withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		rows, err := sourcecredibilitygrades.List(c)
		if err != nil {
			return err
		}
		out = &engine.ListSourceCredibilityGradesResponse{}
		for _, g := range rows {
			out.Grades = append(out.Grades, credibilityGradeProto(g))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func UpsertSourceCredibilityAssessment(in []byte) ([]byte, error) {
	var req engine.UpsertSourceCredibilityAssessmentRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("upsert_source_credibility_assessment", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	sourceID, err := parseID(req.GetSourceId())
	if err != nil {
		return nil, err
	}
	var out *engine.UpsertSourceCredibilityAssessmentResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		gradeID, err := resolveCredibilityGradeID(c, &req)
		if err != nil {
			return err
		}
		action := "create_source_credibility_assessment"
		if _, err := sourcecredibility.GetBySource(c, sourceID); err == nil {
			action = "update_source_credibility_assessment"
		} else if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		a, _, err := writes.Run(c, writes.Op{Action: action, UserID: userID}, func(tx *database.Tx) (sourcecredibility.Assessment, []rowchange.Change, error) {
			return sourcecredibility.Upsert(tx, userID, sourcecredibility.UpsertInput{
				SourceID:           sourceID,
				CredibilityGradeID: gradeID,
				Argument:           req.GetArgument(),
			})
		})
		if err != nil {
			return err
		}
		ap, err := credibilityAssessmentProto(c, a)
		if err != nil {
			return err
		}
		out = &engine.UpsertSourceCredibilityAssessmentResponse{Assessment: ap}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func resolveCredibilityGradeID(c *database.Catalog, req *engine.UpsertSourceCredibilityAssessmentRequest) ([]byte, error) {
	if req.GetGradeId() != "" {
		return parseID(req.GetGradeId())
	}
	key := req.GetGradeKey()
	origin := req.GetGradeOrigin()
	if key == "" || origin == "" {
		return nil, sourcecredibility.ErrInvalid
	}
	g, err := sourcecredibilitygrades.Lookup(c, key, origin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sourcecredibility.ErrInvalid
		}
		return nil, err
	}
	return g.ID, nil
}

func credibilityForSource(c *database.Catalog, sourceID []byte) (*engine.SourceCredibilityAssessment, error) {
	a, err := sourcecredibility.GetBySource(c, sourceID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return credibilityAssessmentProto(c, a)
}

func credibilityAssessmentProto(c *database.Catalog, a sourcecredibility.Assessment) (*engine.SourceCredibilityAssessment, error) {
	g, err := sourcecredibilitygrades.GetByID(c, a.CredibilityGradeID)
	if err != nil {
		return nil, err
	}
	return &engine.SourceCredibilityAssessment{
		Id:         uuidString(a.ID),
		SourceId:   uuidString(a.SourceID),
		GradeId:    uuidString(a.CredibilityGradeID),
		GradeKey:   g.Key,
		GradeLabel: g.Label,
		Argument:   a.Argument,
	}, nil
}

func credibilityGradeProto(g sourcecredibilitygrades.Grade) *engine.SourceCredibilityGrade {
	return &engine.SourceCredibilityGrade{
		Id:        uuidString(g.ID),
		Key:       g.Key,
		Origin:    g.Origin,
		Label:     g.Label,
		SortOrder: int32(g.SortOrder),
	}
}
