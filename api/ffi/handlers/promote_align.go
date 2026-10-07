package handlers

import (
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/conclusionheaders"
	"github.com/mendahu/provenencia/core/database/promotealign"
	"github.com/mendahu/provenencia/core/graphalign"
	"google.golang.org/protobuf/proto"
)

// ProposePromoteGraphAlignment loads one Source's Evidence layer and returns
// Align's proposal (S9-42).
func ProposePromoteGraphAlignment(in []byte) ([]byte, error) {
	var req engine.ProposePromoteGraphAlignmentRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("propose_promote_graph_alignment", err)
	}
	sourceID, err := parseID(req.GetSourceId())
	if err != nil {
		return nil, err
	}
	fixed := make([]graphalign.Fixed, 0, len(req.GetFixed()))
	for _, f := range req.GetFixed() {
		sid, err := parseID(f.GetSubjectId())
		if err != nil {
			return nil, err
		}
		hid, err := parseID(f.GetHandleId())
		if err != nil {
			return nil, err
		}
		fixed = append(fixed, graphalign.Fixed{SubjectID: sid, HandleID: hid})
	}

	var out *engine.ProposePromoteGraphAlignmentResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		db, err := c.DB()
		if err != nil {
			return err
		}
		prop, err := promotealign.Propose(db, sourceID, fixed)
		if err != nil {
			return err
		}
		out, err = proposalProto(db, prop)
		return err
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func proposalProto(q conclusionheaders.Querier, prop graphalign.Proposal) (*engine.ProposePromoteGraphAlignmentResponse, error) {
	out := &engine.ProposePromoteGraphAlignmentResponse{}
	var personIDs, eventIDs, placeIDs [][]byte
	for _, r := range prop.Rows {
		if r.Target != graphalign.TargetHandle || len(r.HandleID) == 0 {
			continue
		}
		switch r.Kind {
		case "person":
			personIDs = append(personIDs, r.HandleID)
		case "event":
			eventIDs = append(eventIDs, r.HandleID)
		case "place":
			placeIDs = append(placeIDs, r.HandleID)
		}
	}
	persons, err := conclusionheaders.PersonsByIDs(q, personIDs)
	if err != nil {
		return nil, err
	}
	events, err := conclusionheaders.EventsByIDs(q, eventIDs)
	if err != nil {
		return nil, err
	}
	places, err := conclusionheaders.PlacesByIDs(q, placeIDs)
	if err != nil {
		return nil, err
	}
	personByID := map[string]conclusionheaders.PersonHeader{}
	for _, h := range persons {
		personByID[string(h.Entity.ID)] = h
	}
	eventByID := map[string]conclusionheaders.EventHeader{}
	for _, h := range events {
		eventByID[string(h.Entity.ID)] = h
	}
	placeByID := map[string]conclusionheaders.PlaceHeader{}
	for _, h := range places {
		placeByID[string(h.Entity.ID)] = h
	}

	for _, r := range prop.Rows {
		row := &engine.PromoteGraphAlignmentRow{
			SubjectId:         uuidString(r.SubjectID),
			Kind:              r.Kind,
			Target:            string(r.Target),
			HandleId:          uuidString(r.HandleID),
			HandleRef:         r.HandleRef,
			Score:             r.Score,
			Assessment:        string(r.Assessment),
			Reasons:           append([]string(nil), r.Reasons...),
			ConflictWithFixed: r.Flags.ConflictWithFixed,
			PossibleDuplicate: r.Flags.PossibleDuplicate,
		}
		for _, c := range r.Comparisons {
			row.Comparisons = append(row.Comparisons, &engine.PromoteGraphAlignmentComparison{
				PropertyKey:    c.Property.Key,
				PropertyOrigin: c.Property.Origin,
				Outcome:        string(c.Outcome),
				ValueType:      c.ValueType,
				Pinned:         c.Pinned,
			})
		}
		for _, a := range r.Alternatives {
			row.Alternatives = append(row.Alternatives, &engine.PromoteGraphAlignmentAlternative{
				HandleId:  uuidString(a.HandleID),
				HandleRef: a.Ref,
				Score:     a.Score,
			})
		}
		if r.Target == graphalign.TargetHandle && len(r.HandleID) == 16 {
			switch r.Kind {
			case "person":
				if h, ok := personByID[string(r.HandleID)]; ok {
					row.Header = &engine.PromoteGraphAlignmentRow_Person{Person: personHeaderProto(h)}
				}
			case "event":
				if h, ok := eventByID[string(r.HandleID)]; ok {
					row.Header = &engine.PromoteGraphAlignmentRow_Event{Event: eventHeaderProto(h)}
				}
			case "place":
				if h, ok := placeByID[string(r.HandleID)]; ok {
					row.Header = &engine.PromoteGraphAlignmentRow_Place{Place: placeHeaderProto(h)}
				}
			}
		}
		out.Rows = append(out.Rows, row)
	}
	return out, nil
}
