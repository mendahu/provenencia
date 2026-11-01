package handlers

import (
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/conclusionsummaries"
	"github.com/mendahu/provenencia/core/database/files"
	"github.com/mendahu/provenencia/core/valuecodec"
	"github.com/mendahu/provenencia/core/vocabkey"
	"google.golang.org/protobuf/proto"
)

func ListPersonSummaries(in []byte) ([]byte, error) {
	var req engine.ListPersonSummariesRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("list_person_summaries", err)
	}
	out := &engine.ListPersonSummariesResponse{}
	err := withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		summaries, err := conclusionsummaries.ListPersons(c)
		if err != nil {
			return err
		}
		var imageIDs [][]byte
		for _, h := range summaries {
			if len(h.PictureImageFileID) > 0 {
				imageIDs = append(imageIDs, h.PictureImageFileID)
			}
		}
		paths, err := files.RelPaths(c, imageIDs)
		if err != nil {
			return err
		}
		for _, h := range summaries {
			ph := personSummaryProto(h)
			ph.PictureImageRelPath = paths[string(h.PictureImageFileID)]
			out.Summaries = append(out.Summaries, ph)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

// conclusionSummaryProto wraps the summary of the handle id of the given kind,
// nil when that id was not read. paths maps a lead picture's image File to its
// objects/… path (files.RelPaths); a nil map leaves the path empty.
func conclusionSummaryProto(s conclusionsummaries.Summaries, kind string, id []byte, paths map[string]string) *engine.ConclusionSummary {
	switch kind {
	case vocabkey.Person:
		if p, ok := s.Person(id); ok {
			ps := personSummaryProto(p)
			ps.PictureImageRelPath = paths[string(p.PictureImageFileID)]
			return &engine.ConclusionSummary{Kind: &engine.ConclusionSummary_Person{Person: ps}}
		}
	case vocabkey.Event:
		if e, ok := s.Event(id); ok {
			es := eventSummaryProto(e)
			es.PictureImageRelPath = paths[string(e.PictureImageFileID)]
			return &engine.ConclusionSummary{Kind: &engine.ConclusionSummary_Event{Event: es}}
		}
	case vocabkey.Place:
		if p, ok := s.Place(id); ok {
			ps := placeSummaryProto(p)
			ps.PictureImageRelPath = paths[string(p.PictureImageFileID)]
			return &engine.ConclusionSummary{Kind: &engine.ConclusionSummary_Place{Place: ps}}
		}
	}
	return nil
}

func personSummaryProto(h conclusionsummaries.PersonSummary) *engine.PersonSummary {
	ph := &engine.PersonSummary{
		Entity:          canonicalEntityProto(h.Entity),
		NameValueCount:  int32(h.NameValueCount),
		UpdatedRevision: h.Updated.Revision,
		UpdatedAt:       h.Updated.At,
	}
	if h.Name != nil {
		ph.Name = valuecodec.NameToProto(*h.Name)
	}
	ph.Birth = lifeFactsProto(h.Birth)
	ph.Death = lifeFactsProto(h.Death)
	ph.Gender = termFactsProto(h.Gender)
	ph.SexAtBirth = termFactsProto(h.SexAtBirth)
	return ph
}

// termProto is the one Go mapper of a summary term onto the wire; nil gives nil.
func termProto(t *conclusionsummaries.Term) *engine.ConclusionTerm {
	if t == nil {
		return nil
	}
	return &engine.ConclusionTerm{
		Id:     database.UUIDString(t.ID),
		Key:    t.Key,
		Origin: t.Origin,
		Label:  t.Label,
	}
}

func termFactsProto(f conclusionsummaries.TermFacts) *engine.TermFacts {
	return &engine.TermFacts{Term: termProto(f.Term), Count: int32(f.Count), PropertyLabel: f.PropertyLabel}
}

func lifeFactsProto(life conclusionsummaries.LifeFacts) *engine.LifeFacts {
	out := &engine.LifeFacts{DateCount: int32(life.DateCount), EventCount: int32(life.EventCount)}
	if life.Event != nil {
		out.Event = canonicalEntityProto(*life.Event)
	}
	if life.Date != nil {
		out.Date = valuecodec.DateToProto(*life.Date)
	}
	for _, p := range life.Places {
		out.Places = append(out.Places, chainedPlaceProto(p))
	}
	return out
}

func chainedPlaceProto(p conclusionsummaries.ChainedPlace) *engine.ChainedPlace {
	return &engine.ChainedPlace{
		Entity:               canonicalEntityProto(p.Entity),
		Names:                append([]string(nil), p.Names...),
		Parents:              append([]string(nil), p.Parents...),
		ParentsAreCandidates: p.ParentsAreCandidates,
	}
}

func ListEventSummaries(in []byte) ([]byte, error) {
	var req engine.ListEventSummariesRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("list_event_summaries", err)
	}
	out := &engine.ListEventSummariesResponse{}
	err := withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		summaries, err := conclusionsummaries.ListEvents(c)
		if err != nil {
			return err
		}
		var imageIDs [][]byte
		for _, h := range summaries {
			if len(h.PictureImageFileID) > 0 {
				imageIDs = append(imageIDs, h.PictureImageFileID)
			}
		}
		paths, err := files.RelPaths(c, imageIDs)
		if err != nil {
			return err
		}
		for _, h := range summaries {
			eh := eventSummaryProto(h)
			eh.PictureImageRelPath = paths[string(h.PictureImageFileID)]
			out.Summaries = append(out.Summaries, eh)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func eventSummaryProto(h conclusionsummaries.EventSummary) *engine.EventSummary {
	eh := &engine.EventSummary{
		Entity:          canonicalEntityProto(h.Entity),
		EventName:       h.EventName,
		EventNameCount:  int32(h.EventNameCount),
		EventTypeCount:  int32(h.Type.Count),
		DateCount:       int32(h.DateCount),
		StartDateCount:  int32(h.StartDateCount),
		EndDateCount:    int32(h.EndDateCount),
		UpdatedRevision: h.Updated.Revision,
		UpdatedAt:       h.Updated.At,
	}
	eh.EventType = termProto(h.Type.Term)
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
		eh.Places = append(eh.Places, chainedPlaceProto(p))
	}
	eh.Title = eventTitleProto(h.Title)
	return eh
}

func ListPlaceSummaries(in []byte) ([]byte, error) {
	var req engine.ListPlaceSummariesRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("list_place_summaries", err)
	}
	out := &engine.ListPlaceSummariesResponse{}
	err := withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		summaries, err := conclusionsummaries.ListPlaces(c)
		if err != nil {
			return err
		}
		var imageIDs [][]byte
		for _, h := range summaries {
			if len(h.PictureImageFileID) > 0 {
				imageIDs = append(imageIDs, h.PictureImageFileID)
			}
		}
		paths, err := files.RelPaths(c, imageIDs)
		if err != nil {
			return err
		}
		for _, h := range summaries {
			ph := placeSummaryProto(h)
			ph.PictureImageRelPath = paths[string(h.PictureImageFileID)]
			out.Summaries = append(out.Summaries, ph)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func placeSummaryProto(h conclusionsummaries.PlaceSummary) *engine.PlaceSummary {
	ph := &engine.PlaceSummary{
		Entity:               canonicalEntityProto(h.Entity),
		Names:                h.Names,
		Kind:                 h.Kind,
		Parents:              h.Parents,
		ParentsAreCandidates: h.ParentsAreCandidates,
		UpdatedRevision:      h.Updated.Revision,
		UpdatedAt:            h.Updated.At,
	}
	if h.StartDate != nil {
		ph.StartDate = valuecodec.DateToProto(*h.StartDate)
	}
	if h.EndDate != nil {
		ph.EndDate = valuecodec.DateToProto(*h.EndDate)
	}
	for _, r := range h.PartOf {
		ph.PartOf = append(ph.PartOf, placeRelationshipProto(r))
	}
	for _, r := range h.Contains {
		ph.Contains = append(ph.Contains, placeRelationshipProto(r))
	}
	for _, r := range h.Predecessors {
		ph.Predecessors = append(ph.Predecessors, placeRelationshipProto(r))
	}
	for _, r := range h.Successors {
		ph.Successors = append(ph.Successors, placeRelationshipProto(r))
	}
	return ph
}

func placeRelationshipProto(r conclusionsummaries.PlaceRelationship) *engine.PlaceRelationship {
	out := &engine.PlaceRelationship{
		Entity: canonicalEntityProto(r.Entity),
		Title:  r.Title,
		Kind:   r.Kind,
	}
	if r.StartDate != nil {
		out.StartDate = valuecodec.DateToProto(*r.StartDate)
	}
	if r.EndDate != nil {
		out.EndDate = valuecodec.DateToProto(*r.EndDate)
	}
	return out
}
