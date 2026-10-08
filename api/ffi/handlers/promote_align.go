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
		hid, err := optionalID(f.GetHandleId())
		if err != nil {
			return nil, err
		}
		fixed = append(fixed, graphalign.Fixed{SubjectID: sid, HandleID: hid, Target: graphalign.Target(f.GetTarget())})
	}

	var out *engine.ProposePromoteGraphAlignmentResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		db, err := c.DB()
		if err != nil {
			return err
		}
		prop, rev, err := promotealign.Propose(db, sourceID, fixed)
		if err != nil {
			return err
		}
		out, err = proposalProto(db, prop)
		if err != nil {
			return err
		}
		out.Revision = rev
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func proposalProto(q conclusionheaders.Querier, prop graphalign.Proposal) (*engine.ProposePromoteGraphAlignmentResponse, error) {
	out := &engine.ProposePromoteGraphAlignmentResponse{}
	var personIDs, eventIDs, placeIDs [][]byte
	addHeaderID := func(kind string, id []byte) {
		if len(id) != 16 {
			return
		}
		switch kind {
		case "person":
			personIDs = append(personIDs, id)
		case "event":
			eventIDs = append(eventIDs, id)
		case "place":
			placeIDs = append(placeIDs, id)
		}
	}
	for _, r := range prop.Rows {
		if r.Target == graphalign.TargetHandle {
			addHeaderID(r.Kind, r.HandleID)
		}
		for _, a := range r.Alternatives {
			addHeaderID(r.Kind, a.HandleID)
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
			SubjectId:            uuidString(r.SubjectID),
			Kind:                 r.Kind,
			Target:               string(r.Target),
			HandleId:             uuidString(r.HandleID),
			HandleRef:            r.HandleRef,
			Score:                r.Score,
			Assessment:           string(r.Assessment),
			Reasons:              append([]string(nil), r.Reasons...),
			ConflictWithFixed:    r.Flags.ConflictWithFixed,
			PossibleDuplicate:    r.Flags.PossibleDuplicate,
			DuplicateOfSubjectId: uuidString(r.Flags.DuplicateOf),
		}
		if r.Via != nil {
			row.ViaNeighborSubjectId = uuidString(r.Via.NeighborSubjectID)
			row.ViaBridgeType = r.Via.Signature.BridgeType
			row.ViaRole = r.Via.Signature.RoleOrType
		}
		if len(r.Exhibits) > 0 {
			for _, ex := range r.Exhibits {
				row.Comparisons = append(row.Comparisons, &engine.PromoteGraphAlignmentComparison{
					PropertyKey:           ex.Property.Key,
					PropertyOrigin:        ex.Property.Origin,
					Outcome:               string(ex.Outcome),
					ValueType:             ex.ValueType,
					Pinned:                ex.Pinned,
					Weight:                ex.Weight,
					GroupLabel:            ex.GroupLabel,
					IncomingObservationId: uuidString(ex.IncomingObservationID),
					IncomingDisplay:       ex.IncomingDisplay,
					IncomingSource:        ex.IncomingSource,
					MemberObservationId:   uuidString(ex.MemberObservationID),
					MemberDisplay:         ex.MemberDisplay,
					MemberSource:          ex.MemberSource,
				})
			}
		} else {
			for _, c := range r.Comparisons {
				row.Comparisons = append(row.Comparisons, &engine.PromoteGraphAlignmentComparison{
					PropertyKey:    c.Property.Key,
					PropertyOrigin: c.Property.Origin,
					Outcome:        string(c.Outcome),
					ValueType:      c.ValueType,
					Pinned:         c.Pinned,
					Weight:         c.Weight,
				})
			}
		}
		for _, a := range r.Alternatives {
			alt := &engine.PromoteGraphAlignmentAlternative{
				HandleId:  uuidString(a.HandleID),
				HandleRef: a.Ref,
				Score:     a.Score,
			}
			switch r.Kind {
			case "person":
				if h, ok := personByID[string(a.HandleID)]; ok {
					alt.Header = &engine.PromoteGraphAlignmentAlternative_Person{Person: personHeaderProto(h)}
				}
			case "event":
				if h, ok := eventByID[string(a.HandleID)]; ok {
					alt.Header = &engine.PromoteGraphAlignmentAlternative_Event{Event: eventHeaderProto(h)}
				}
			case "place":
				if h, ok := placeByID[string(a.HandleID)]; ok {
					alt.Header = &engine.PromoteGraphAlignmentAlternative_Place{Place: placeHeaderProto(h)}
				}
			}
			row.Alternatives = append(row.Alternatives, alt)
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
