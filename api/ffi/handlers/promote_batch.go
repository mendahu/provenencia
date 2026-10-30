package handlers

import (
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/writes"
	"google.golang.org/protobuf/proto"
)

// ApplyPromoteGraphAlignment files one Done for an Evidence graph (S9-43).
func ApplyPromoteGraphAlignment(in []byte) ([]byte, error) {
	var req engine.ApplyPromoteGraphAlignmentRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("apply_promote_graph_alignment", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	graphID, err := parseID(req.GetGraphId())
	if err != nil {
		return nil, err
	}
	rows := make([]promote.BatchRow, 0, len(req.GetRows()))
	for _, row := range req.GetRows() {
		subjectID, err := parseID(row.GetSubjectId())
		if err != nil {
			return nil, err
		}
		entityID, err := optionalID(row.GetEntityId())
		if err != nil {
			return nil, err
		}
		rows = append(rows, promote.BatchRow{SubjectID: subjectID, Target: row.GetTarget(), EntityID: entityID})
	}

	var out *engine.ApplyPromoteGraphAlignmentResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		res, result, err := writes.Run(c, writes.Op{Action: "promote_batch", UserID: userID}, func(tx *database.Tx) (promote.BatchResult, []rowchange.Change, error) {
			return promote.SaveBatch(tx, userID, promote.Batch{
				GraphID:      graphID,
				SeenRevision: req.GetSeenRevision(),
				Rows:         rows,
			})
		})
		if err != nil {
			return err
		}
		if result.Revision != 0 {
			res.Revision = result.Revision
		}
		out = &engine.ApplyPromoteGraphAlignmentResponse{Revision: res.Revision}
		for _, w := range res.Written {
			out.Written = append(out.Written, &engine.ApplyPromoteGraphAlignmentWritten{
				Entity: canonicalEntityProto(w.Entity),
				Claim:  identityClaimProto(w.Claim),
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}
