package handlers

import (
	"context"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/conclusionsummaries"
	"github.com/mendahu/provenencia/core/search"
	"github.com/mendahu/provenencia/core/workspacelocation"
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
		Kinds:    req.GetKinds(),
	}
	var out *engine.SearchCatalogResponse
	err := withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		// FFI dispatch has no call-scoped context yet; search honors ctx for future wiring.
		hits, err := search.DefaultEngine().Search(context.Background(), c, q)
		if err != nil {
			return err
		}
		out = &engine.SearchCatalogResponse{Hits: make([]*engine.SearchHit, 0, len(hits))}
		for _, h := range hits {
			out.Hits = append(out.Hits, hitToProto(h))
		}
		return fillHitSummaries(c, out.Hits)
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func locationFromProto(loc *engine.WorkspaceLocation) workspacelocation.Location {
	if loc == nil {
		return workspacelocation.Location{}
	}
	return workspacelocation.Location{
		Section:              loc.GetSection(),
		SourceID:             loc.GetSourceId(),
		FieldID:              loc.GetFieldId(),
		TypeID:               loc.GetTypeId(),
		SubjectID:            loc.GetSubjectId(),
		CitationID:           loc.GetCitationId(),
		ArtifactID:           loc.GetArtifactId(),
		ObservationID:        loc.GetObservationId(),
		ConnectFromSubjectID: loc.GetConnectFromSubjectId(),
		ConnectToSubjectID:   loc.GetConnectToSubjectId(),
		ConnectBridgeTypeKey: loc.GetConnectBridgeTypeKey(),
		SubjectTypeKey:       loc.GetSubjectTypeKey(),
		PropertyID:           loc.GetPropertyId(),
		EntityID:             loc.GetEntityId(),
		GraphID:              loc.GetGraphId(),
		SourceSurface:        loc.GetSourceSurface(),
		Ref:                  loc.GetRef(),
		Title:                loc.GetTitle(),
		SourceTitle:          loc.GetSourceTitle(),
	}
}

func locationToProto(loc workspacelocation.Location) *engine.WorkspaceLocation {
	return &engine.WorkspaceLocation{
		Section:              loc.Section,
		SourceId:             loc.SourceID,
		FieldId:              loc.FieldID,
		TypeId:               loc.TypeID,
		SubjectId:            loc.SubjectID,
		CitationId:           loc.CitationID,
		ArtifactId:           loc.ArtifactID,
		ObservationId:        loc.ObservationID,
		ConnectFromSubjectId: loc.ConnectFromSubjectID,
		ConnectToSubjectId:   loc.ConnectToSubjectID,
		ConnectBridgeTypeKey: loc.ConnectBridgeTypeKey,
		SubjectTypeKey:       loc.SubjectTypeKey,
		PropertyId:           loc.PropertyID,
		EntityId:             loc.EntityID,
		GraphId:              loc.GraphID,
		SourceSurface:        loc.SourceSurface,
		Ref:                  loc.Ref,
		Title:                loc.Title,
		SourceTitle:          loc.SourceTitle,
	}
}

// fillHitSummaries attaches the list summary for each person, event, and place hit.
func fillHitSummaries(c *database.Catalog, hits []*engine.SearchHit) error {
	var w conclusionsummaries.Want
	for _, h := range hits {
		if id, err := parseID(h.GetId()); err == nil {
			w.Add(h.GetKind(), id)
		}
	}
	summaries, err := conclusionsummaries.ByIDs(c, w)
	if err != nil {
		return err
	}
	for _, h := range hits {
		if id, err := parseID(h.GetId()); err == nil {
			h.Summary = conclusionSummaryProto(summaries, h.GetKind(), id)
		}
	}
	return nil
}

func hitToProto(h search.Hit) *engine.SearchHit {
	return &engine.SearchHit{
		Kind:             h.Kind,
		Id:               h.ID,
		Ref:              h.Ref,
		Title:            h.Title,
		Subtitle:         h.Subtitle,
		MatchReason:      h.MatchReason,
		MatchSnippet:     h.MatchSnippet,
		Location:         locationToProto(h.Location),
		ThumbnailRelPath: h.ThumbnailRelPath,
		IconKey:          h.IconKey,
		MemberCount:      int32(h.MemberCount),
	}
}
