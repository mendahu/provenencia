package handlers

import (
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/conclusionheaders"
	"github.com/mendahu/provenencia/core/database/promotealign"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/vocabkey"
	"google.golang.org/protobuf/proto"
)

// ProposePromoteGraphAlignment loads one Evidence graph's layer and returns
// Align's proposal (S9-42).
func ProposePromoteGraphAlignment(in []byte) ([]byte, error) {
	var req engine.ProposePromoteGraphAlignmentRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("propose_promote_graph_alignment", err)
	}
	graphID, err := parseID(req.GetGraphId())
	if err != nil {
		return nil, err
	}
	fixed, err := fixedFromProto(req.GetFixed())
	if err != nil {
		return nil, err
	}

	var out *engine.ProposePromoteGraphAlignmentResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		prop, rev, err := promotealign.Propose(c, graphID, fixed)
		if err != nil {
			return err
		}
		out, err = proposalProto(c, prop)
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

// fixedFromProto reads the targets a page holds for its rows: the ids parse,
// and a target is kept as the client sent it for the core to check.
func fixedFromProto(in []*engine.PromoteGraphAlignmentFixed) ([]graphalign.Fixed, error) {
	fixed := make([]graphalign.Fixed, 0, len(in))
	for _, f := range in {
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
	return fixed, nil
}

func proposalProto(c *database.Catalog, prop graphalign.Proposal) (*engine.ProposePromoteGraphAlignmentResponse, error) {
	out := &engine.ProposePromoteGraphAlignmentResponse{}
	var personIDs, eventIDs, placeIDs [][]byte
	addHeaderID := func(kind string, id []byte) {
		if len(id) != 16 {
			return
		}
		switch kind {
		case vocabkey.Person:
			personIDs = append(personIDs, id)
		case vocabkey.Event:
			eventIDs = append(eventIDs, id)
		case vocabkey.Place:
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
	headers, err := conclusionheaders.ByIDs(c, personIDs, eventIDs, placeIDs)
	if err != nil {
		return nil, err
	}

	for _, r := range prop.Rows {
		row := &engine.PromoteGraphAlignmentRow{
			SubjectId:            database.UUIDString(r.SubjectID),
			Kind:                 r.Kind,
			Target:               string(r.Target),
			HandleId:             database.UUIDString(r.HandleID),
			HandleRef:            r.HandleRef,
			Score:                r.Score,
			Assessment:           string(r.Assessment),
			Reason:               string(r.Reason),
			ReasonPropertyKey:    string(r.ReasonProperty),
			ReasonPropertyOrigin: productOrigin(r.ReasonProperty),
			ConflictWithFixed:    r.Flags.ConflictWithFixed,
			Contested:            r.Flags.Contested,
			PossibleDuplicate:    r.Flags.PossibleDuplicate,
			DuplicateOfSubjectId: database.UUIDString(r.Flags.DuplicateOf),
		}
		if r.Via != nil {
			row.ViaNeighborSubjectId = database.UUIDString(r.Via.NeighborSubjectID)
			row.ViaBridgeType = string(r.Via.Key.Type)
			row.ViaRole = string(r.Via.Key.Term)
		}
		if r.Move != nil {
			row.MoveReason = string(r.Move.Reason)
			row.MoveNeighborSubjectId = database.UUIDString(r.Move.NeighborSubjectID)
		}
		for _, a := range r.Alternatives {
			alt := &engine.PromoteGraphAlignmentAlternative{
				HandleId:             database.UUIDString(a.HandleID),
				HandleRef:            a.Ref,
				Score:                a.Score,
				Assessment:           string(a.Assessment),
				Reason:               string(a.Reason),
				ReasonPropertyKey:    string(a.ReasonProperty),
				ReasonPropertyOrigin: productOrigin(a.ReasonProperty),
				ViaNeighborSubjectId: database.UUIDString(a.ViaNeighbor),
			}
			switch r.Kind {
			case vocabkey.Person:
				if h, ok := headers.Person(a.HandleID); ok {
					alt.Header = &engine.PromoteGraphAlignmentAlternative_Person{Person: personHeaderProto(h)}
				}
			case vocabkey.Event:
				if h, ok := headers.Event(a.HandleID); ok {
					alt.Header = &engine.PromoteGraphAlignmentAlternative_Event{Event: eventHeaderProto(h)}
				}
			case vocabkey.Place:
				if h, ok := headers.Place(a.HandleID); ok {
					alt.Header = &engine.PromoteGraphAlignmentAlternative_Place{Place: placeHeaderProto(h)}
				}
			}
			row.Alternatives = append(row.Alternatives, alt)
		}
		if r.Target == graphalign.TargetHandle && len(r.HandleID) == 16 {
			switch r.Kind {
			case vocabkey.Person:
				if h, ok := headers.Person(r.HandleID); ok {
					row.Header = &engine.PromoteGraphAlignmentRow_Person{Person: personHeaderProto(h)}
				}
			case vocabkey.Event:
				if h, ok := headers.Event(r.HandleID); ok {
					row.Header = &engine.PromoteGraphAlignmentRow_Event{Event: eventHeaderProto(h)}
				}
			case vocabkey.Place:
				if h, ok := headers.Place(r.HandleID); ok {
					row.Header = &engine.PromoteGraphAlignmentRow_Place{Place: placeHeaderProto(h)}
				}
			}
		}
		out.Rows = append(out.Rows, row)
	}
	return out, nil
}

// productOrigin is the origin a reason property's key goes out with: subject
// compare leads are product properties, and an empty key has no origin.
func productOrigin(key vocabkey.PropertyKey) string {
	if key == "" {
		return ""
	}
	return vocabkey.OriginProvenencia
}
