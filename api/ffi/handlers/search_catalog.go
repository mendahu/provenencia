package handlers

import (
	"context"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/search"
	"google.golang.org/protobuf/proto"
)

func SearchCatalog(in []byte) ([]byte, error) {
	var req engine.SearchCatalogRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("search_catalog", err)
	}
	q := search.Query{
		Text:     req.GetQuery(),
		Location: locationFromProto(req.GetLocation()),
		Limit:    search.DefaultHitLimit,
	}
	var out *engine.SearchCatalogResponse
	err := withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		hits, err := search.DefaultEngine().Search(context.Background(), c, q)
		if err != nil {
			return err
		}
		out = &engine.SearchCatalogResponse{Hits: make([]*engine.SearchHit, 0, len(hits))}
		for _, h := range hits {
			out.Hits = append(out.Hits, hitToProto(h))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func locationFromProto(loc *engine.WorkspaceLocation) search.WorkspaceLocation {
	if loc == nil {
		return search.WorkspaceLocation{}
	}
	return search.WorkspaceLocation{
		Section:  loc.GetSection(),
		SourceID: loc.GetSourceId(),
		FieldID:  loc.GetFieldId(),
		TypeID:   loc.GetTypeId(),
		Ref:      loc.GetRef(),
		Title:    loc.GetTitle(),
	}
}

func locationToProto(loc search.WorkspaceLocation) *engine.WorkspaceLocation {
	return &engine.WorkspaceLocation{
		Section:  loc.Section,
		SourceId: loc.SourceID,
		FieldId:  loc.FieldID,
		TypeId:   loc.TypeID,
		Ref:      loc.Ref,
		Title:    loc.Title,
	}
}

func hitToProto(h search.Hit) *engine.SearchHit {
	return &engine.SearchHit{
		Kind:        h.Kind,
		Id:          h.ID,
		Ref:         h.Ref,
		Title:       h.Title,
		Subtitle:    h.Subtitle,
		MatchReason: h.MatchReason,
		Location:    locationToProto(h.Location),
	}
}
