package handlers

import (
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/metadatafields"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"google.golang.org/protobuf/proto"
)

func GetWorkspaceNavCounts(in []byte) ([]byte, error) {
	var req engine.GetWorkspaceNavCountsRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("get_workspace_nav_counts", err)
	}
	var out *engine.GetWorkspaceNavCountsResponse
	err := withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		sourceCount, err := sources.Count(c)
		if err != nil {
			return err
		}
		types, err := sourcetypes.CountByOrigin(c)
		if err != nil {
			return err
		}
		fields, err := metadatafields.CountByOrigin(c)
		if err != nil {
			return err
		}
		persons, err := canonicalentities.CountByTypeKey(c, "person")
		if err != nil {
			return err
		}
		events, err := canonicalentities.CountByTypeKey(c, "event")
		if err != nil {
			return err
		}
		places, err := canonicalentities.CountByTypeKey(c, "place")
		if err != nil {
			return err
		}
		out = &engine.GetWorkspaceNavCountsResponse{
			Sources:        int32(sourceCount),
			SourceTypes:    typeOriginProto(types),
			MetadataFields: fieldOriginProto(fields),
			Persons:        int32(persons),
			Events:         int32(events),
			Places:         int32(places),
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func typeOriginProto(c sourcetypes.OriginCounts) *engine.VocabularyOriginCounts {
	return &engine.VocabularyOriginCounts{
		Total:  int32(c.Total),
		Seeded: int32(c.Seeded),
		User:   int32(c.User),
		Plugin: int32(c.Plugin),
	}
}

func fieldOriginProto(c metadatafields.OriginCounts) *engine.VocabularyOriginCounts {
	return &engine.VocabularyOriginCounts{
		Total:  int32(c.Total),
		Seeded: int32(c.Seeded),
		User:   int32(c.User),
		Plugin: int32(c.Plugin),
	}
}
