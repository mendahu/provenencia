package handlers

import (
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/sourcegraphprogress"
	"google.golang.org/protobuf/proto"
)

func ListSourceGraphProgress(in []byte) ([]byte, error) {
	var req engine.ListSourceGraphProgressRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("list_source_graph_progress", err)
	}
	var out *engine.ListSourceGraphProgressResponse
	err := withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		rows, err := sourcegraphprogress.List(c)
		if err != nil {
			return err
		}
		out = &engine.ListSourceGraphProgressResponse{}
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

func GetSourceGraphProgress(in []byte) ([]byte, error) {
	var req engine.GetSourceGraphProgressRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("get_source_graph_progress", err)
	}
	sourceID, err := parseID(req.GetSourceId())
	if err != nil {
		return nil, err
	}
	var out *engine.GetSourceGraphProgressResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		row, err := sourcegraphprogress.Get(c, sourceID)
		if err != nil {
			return err
		}
		out = &engine.GetSourceGraphProgressResponse{Progress: protoProgress(row)}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func protoProgress(row sourcegraphprogress.Progress) *engine.SourceGraphProgress {
	return &engine.SourceGraphProgress{
		SourceId:         uuidString(row.SourceID),
		SubjectCount:     row.SubjectCount,
		ObservationCount: row.ObservationCount,
	}
}
