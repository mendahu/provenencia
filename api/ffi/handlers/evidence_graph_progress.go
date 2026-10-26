package handlers

import (
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/graphprogress"
	"google.golang.org/protobuf/proto"
)

func ListEvidenceGraphProgress(in []byte) ([]byte, error) {
	var req engine.ListEvidenceGraphProgressRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("list_evidence_graph_progress", err)
	}
	var out *engine.ListEvidenceGraphProgressResponse
	err := withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		rows, err := graphprogress.List(c)
		if err != nil {
			return err
		}
		out = &engine.ListEvidenceGraphProgressResponse{}
		for _, row := range rows {
			out.Rows = append(out.Rows, protoProgress(row))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func GetEvidenceGraphProgress(in []byte) ([]byte, error) {
	var req engine.GetEvidenceGraphProgressRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("get_evidence_graph_progress", err)
	}
	graphID, err := parseID(req.GetGraphId())
	if err != nil {
		return nil, err
	}
	var out *engine.GetEvidenceGraphProgressResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		row, err := graphprogress.Get(c, graphID)
		if err != nil {
			return err
		}
		out = &engine.GetEvidenceGraphProgressResponse{Progress: protoProgress(row)}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func protoProgress(row graphprogress.Progress) *engine.EvidenceGraphProgress {
	return &engine.EvidenceGraphProgress{
		GraphId:          database.UUIDString(row.GraphID),
		SourceId:         database.UUIDString(row.SourceID),
		SubjectCount:     row.SubjectCount,
		ObservationCount: row.ObservationCount,
		FiledCount:       row.FiledCount,
	}
}
