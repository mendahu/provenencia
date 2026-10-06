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
	return ph
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
	return eh
}
