package handlers

import (
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/conclusiondetails"
	"github.com/mendahu/provenencia/core/database/conclusionheaders"
	"github.com/mendahu/provenencia/core/valuecodec"
	"google.golang.org/protobuf/proto"
)

func ListPersonHeaders(in []byte) ([]byte, error) {
	var req engine.ListPersonHeadersRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("list_person_headers", err)
	}
	out := &engine.ListPersonHeadersResponse{}
	err := withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		db, err := c.DB()
		if err != nil {
			return err
		}
		headers, err := conclusionheaders.ListPersons(db)
		if err != nil {
			return err
		}
		for _, h := range headers {
			out.Headers = append(out.Headers, personHeaderProto(h))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func personHeaderProto(h conclusionheaders.PersonHeader) *engine.PersonHeader {
	ph := &engine.PersonHeader{
		Entity:         canonicalEntityProto(h.Entity),
		NameValueCount: int32(h.NameValueCount),
	}
	if h.Name != nil {
		ph.Name = valuecodec.NameToProto(*h.Name)
	}
	ph.Birth = lifeFactsProto(h.Birth)
	ph.Death = lifeFactsProto(h.Death)
	return ph
}

func lifeFactsProto(life conclusionheaders.LifeFacts) *engine.LifeFacts {
	out := &engine.LifeFacts{DateCount: int32(life.DateCount)}
	if life.Date != nil {
		out.Date = valuecodec.DateToProto(*life.Date)
	}
	for _, p := range life.Places {
		out.Places = append(out.Places, headerPlaceProto(p))
	}
	return out
}

func headerPlaceProto(p conclusionheaders.HeaderPlace) *engine.HeaderPlace {
	return &engine.HeaderPlace{Names: append([]string(nil), p.Names...), NameCount: int32(p.Count)}
}

func GetPersonHeader(in []byte) ([]byte, error) {
	var req engine.GetPersonHeaderRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("get_person_header", err)
	}
	entityID, err := parseID(req.GetEntityId())
	if err != nil {
		return nil, conclusiondetails.ErrNotFound
	}
	out := &engine.GetPersonHeaderResponse{}
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		db, err := c.DB()
		if err != nil {
			return err
		}
		headers, err := conclusionheaders.PersonsByIDs(db, [][]byte{entityID})
		if err != nil {
			return err
		}
		if len(headers) != 1 {
			return conclusiondetails.ErrNotFound
		}
		out.Header = personHeaderProto(headers[0])
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func ListEventHeaders(in []byte) ([]byte, error) {
	var req engine.ListEventHeadersRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("list_event_headers", err)
	}
	out := &engine.ListEventHeadersResponse{}
	err := withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		db, err := c.DB()
		if err != nil {
			return err
		}
		headers, err := conclusionheaders.ListEvents(db)
		if err != nil {
			return err
		}
		for _, h := range headers {
			out.Headers = append(out.Headers, eventHeaderProto(h))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func GetEventHeader(in []byte) ([]byte, error) {
	var req engine.GetEventHeaderRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("get_event_header", err)
	}
	entityID, err := parseID(req.GetEntityId())
	if err != nil {
		return nil, conclusiondetails.ErrNotFound
	}
	out := &engine.GetEventHeaderResponse{}
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		db, err := c.DB()
		if err != nil {
			return err
		}
		headers, err := conclusionheaders.EventsByIDs(db, [][]byte{entityID})
		if err != nil {
			return err
		}
		if len(headers) != 1 {
			return conclusiondetails.ErrNotFound
		}
		out.Header = eventHeaderProto(headers[0])
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func eventHeaderProto(h conclusionheaders.EventHeader) *engine.EventHeader {
	eh := &engine.EventHeader{
		Entity:         canonicalEntityProto(h.Entity),
		EventName:      h.EventName,
		EventNameCount: int32(h.EventNameCount),
		EventTypeCount: int32(h.EventTypeCount),
		DateCount:      int32(h.DateCount),
		StartDateCount: int32(h.StartDateCount),
		EndDateCount:   int32(h.EndDateCount),
	}
	if h.EventType != nil {
		eh.EventType = &engine.ConclusionTerm{
			Id:    uuidString(h.EventType.ID),
			Key:   h.EventType.Key,
			Label: h.EventType.Label,
		}
	}
	if h.Date != nil {
		eh.Date = valuecodec.DateToProto(*h.Date)
	}
	if h.StartDate != nil {
		eh.StartDate = valuecodec.DateToProto(*h.StartDate)
	}
	if h.EndDate != nil {
		eh.EndDate = valuecodec.DateToProto(*h.EndDate)
	}
	for _, s := range h.Subjects {
		sub := &engine.EventSubject{
			Entity:         canonicalEntityProto(s.Entity),
			NameValueCount: int32(s.NameValueCount),
		}
		if s.Name != nil {
			sub.Name = valuecodec.NameToProto(*s.Name)
		}
		eh.Subjects = append(eh.Subjects, sub)
	}
	for _, p := range h.Places {
		eh.Places = append(eh.Places, headerPlaceProto(p))
	}
	return eh
}

func ListPlaceHeaders(in []byte) ([]byte, error) {
	var req engine.ListPlaceHeadersRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("list_place_headers", err)
	}
	out := &engine.ListPlaceHeadersResponse{}
	err := withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		db, err := c.DB()
		if err != nil {
			return err
		}
		headers, err := conclusionheaders.ListPlaces(db)
		if err != nil {
			return err
		}
		for _, h := range headers {
			out.Headers = append(out.Headers, placeHeaderProto(h))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func GetPlaceHeader(in []byte) ([]byte, error) {
	var req engine.GetPlaceHeaderRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("get_place_header", err)
	}
	entityID, err := parseID(req.GetEntityId())
	if err != nil {
		return nil, conclusiondetails.ErrNotFound
	}
	out := &engine.GetPlaceHeaderResponse{}
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		db, err := c.DB()
		if err != nil {
			return err
		}
		headers, err := conclusionheaders.PlacesByIDs(db, [][]byte{entityID})
		if err != nil {
			return err
		}
		if len(headers) != 1 {
			return conclusiondetails.ErrNotFound
		}
		out.Header = placeHeaderProto(headers[0])
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func placeHeaderProto(h conclusionheaders.PlaceHeader) *engine.PlaceHeader {
	ph := &engine.PlaceHeader{
		Entity:  canonicalEntityProto(h.Entity),
		Names:   h.Names,
		Kind:    h.Kind,
		Parents: h.Parents,
	}
	if h.StartDate != nil {
		ph.StartDate = valuecodec.DateToProto(*h.StartDate)
	}
	if h.EndDate != nil {
		ph.EndDate = valuecodec.DateToProto(*h.EndDate)
	}
	return ph
}
