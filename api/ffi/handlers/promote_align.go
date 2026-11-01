package handlers

import (
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/conclusionsummaries"
	"github.com/mendahu/provenencia/core/database/promotealign"
	"github.com/mendahu/provenencia/core/graphalign"
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
	var w conclusionsummaries.Want
	for _, r := range prop.Rows {
		w.Add(r.Kind, r.Start().HandleID)
		for _, a := range r.Alternatives {
			w.Add(r.Kind, a.HandleID)
		}
	}
	summaries, err := conclusionsummaries.ByIDs(c, w)
	if err != nil {
		return nil, err
	}

	for _, r := range prop.Rows {
		start := r.Start()
		row := &engine.PromoteGraphAlignmentRow{
			SubjectId:            database.UUIDString(r.SubjectID),
			Kind:                 r.Kind,
			Target:               string(start.Target),
			HandleId:             database.UUIDString(start.HandleID),
			HandleRef:            start.Ref,
			ConflictWithFixed:    r.Flags.ConflictWithFixed,
			PossibleDuplicate:    r.Flags.PossibleDuplicate,
			DuplicateOfSubjectId: database.UUIDString(r.Flags.DuplicateOf),
		}
		if r.Via != nil {
			row.ViaNeighborSubjectId = database.UUIDString(r.Via.NeighborSubjectID)
		}
		if r.Move != nil {
			row.MoveReason = string(r.Move.Reason)
			row.MoveNeighborSubjectId = database.UUIDString(r.Move.NeighborSubjectID)
		}
		for _, a := range r.Alternatives {
			alt := &engine.PromoteGraphAlignmentAlternative{
				HandleId:  database.UUIDString(a.HandleID),
				HandleRef: a.Ref,
			}
			alt.Summary = conclusionSummaryProto(summaries, r.Kind, a.HandleID)
			row.Alternatives = append(row.Alternatives, alt)
		}
		if start.Target == graphalign.TargetHandle && len(start.HandleID) == 16 {
			row.Summary = conclusionSummaryProto(summaries, r.Kind, start.HandleID)
		}
		out.Rows = append(out.Rows, row)
	}
	return out, nil
}
