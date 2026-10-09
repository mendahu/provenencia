package handlers

import (
	"context"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/conclusionheaders"
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
		db, err := c.DB()
		if err != nil {
			return err
		}
		return fillHitHeaders(db, out.Hits)
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
		SourceSurface:        loc.GetSourceSurface(),
		Ref:                  loc.GetRef(),
		Title:                loc.GetTitle(),
		SourceTitle:          loc.GetSourceTitle(),
	}
}

func locationToProto(loc search.WorkspaceLocation) *engine.WorkspaceLocation {
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
		SourceSurface:        loc.SourceSurface,
		Ref:                  loc.Ref,
		Title:                loc.Title,
		SourceTitle:          loc.SourceTitle,
	}
}

// fillHitHeaders attaches the list header for each person, event, and place hit.
func fillHitHeaders(db conclusionheaders.Querier, hits []*engine.SearchHit) error {
	var persons, events, places [][]byte
	for _, h := range hits {
		id, err := parseID(h.GetId())
		if err != nil {
			continue
		}
		switch h.GetKind() {
		case search.KindPerson:
			persons = append(persons, id)
		case search.KindEvent:
			events = append(events, id)
		case search.KindPlace:
			places = append(places, id)
		}
	}
	listedPersons, err := conclusionheaders.PersonsByIDs(db, persons)
	if err != nil {
		return err
	}
	listedEvents, err := conclusionheaders.EventsByIDs(db, events)
	if err != nil {
		return err
	}
	listedPlaces, err := conclusionheaders.PlacesByIDs(db, places)
	if err != nil {
		return err
	}
	personByID := map[string]conclusionheaders.PersonHeader{}
	for _, p := range listedPersons {
		personByID[uuidString(p.Entity.ID)] = p
	}
	eventByID := map[string]conclusionheaders.EventHeader{}
	for _, e := range listedEvents {
		eventByID[uuidString(e.Entity.ID)] = e
	}
	placeByID := map[string]conclusionheaders.PlaceHeader{}
	for _, p := range listedPlaces {
		placeByID[uuidString(p.Entity.ID)] = p
	}
	for _, h := range hits {
		switch h.GetKind() {
		case search.KindPerson:
			if p, ok := personByID[h.GetId()]; ok {
				h.Header = &engine.SearchHit_Person{Person: personHeaderProto(p)}
			}
		case search.KindEvent:
			if e, ok := eventByID[h.GetId()]; ok {
				h.Header = &engine.SearchHit_Event{Event: eventHeaderProto(e)}
			}
		case search.KindPlace:
			if p, ok := placeByID[h.GetId()]; ok {
				h.Header = &engine.SearchHit_Place{Place: placeHeaderProto(p)}
			}
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
