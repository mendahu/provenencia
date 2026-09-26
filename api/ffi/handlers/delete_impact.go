package handlers

import (
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/deleteimpact"
	"google.golang.org/protobuf/proto"
)

func GetDeleteImpact(in []byte) ([]byte, error) {
	var req engine.GetDeleteImpactRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("get_delete_impact", err)
	}
	kind, err := deleteimpact.ParseKind(req.GetKind())
	if err != nil {
		return nil, err
	}
	id, err := parseID(req.GetId())
	if err != nil {
		return nil, deleteimpact.ErrInvalid
	}
	var out *engine.GetDeleteImpactResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		db, err := c.DB()
		if err != nil {
			return err
		}
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback() }()
		report, err := deleteimpact.Impact(tx, kind, id)
		if err != nil {
			return err
		}
		out = reportToProto(report)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func reportToProto(r deleteimpact.Report) *engine.GetDeleteImpactResponse {
	out := &engine.GetDeleteImpactResponse{
		Allowed: r.Allowed,
		Gate:    gateToProto(r.Gate),
	}
	for _, g := range r.Groups {
		pg := &engine.DeleteImpactGroup{
			Via:   g.Via,
			Kind:  string(g.Kind),
			Total: int32(g.Total),
		}
		for _, item := range g.Listed {
			pg.Listed = append(pg.Listed, &engine.DeleteImpactListed{
				Id:       uuidString(item.ID),
				Ref:      item.Ref,
				Title:    item.Title,
				Location: locationFromImpact(item.Location),
			})
		}
		out.Groups = append(out.Groups, pg)
	}
	return out
}

func gateToProto(g deleteimpact.Gate) engine.DeleteImpactGate {
	switch g {
	case deleteimpact.GateOK:
		return engine.DeleteImpactGate_DELETE_IMPACT_GATE_OK
	case deleteimpact.GateInbound:
		return engine.DeleteImpactGate_DELETE_IMPACT_GATE_INBOUND
	case deleteimpact.GateNotFound:
		return engine.DeleteImpactGate_DELETE_IMPACT_GATE_NOT_FOUND
	case deleteimpact.GateEdgeLocked:
		return engine.DeleteImpactGate_DELETE_IMPACT_GATE_EDGE_LOCKED
	case deleteimpact.GateInfra:
		return engine.DeleteImpactGate_DELETE_IMPACT_GATE_INFRA
	case deleteimpact.GateOriginLocked:
		return engine.DeleteImpactGate_DELETE_IMPACT_GATE_ORIGIN_LOCKED
	default:
		return engine.DeleteImpactGate_DELETE_IMPACT_GATE_UNSPECIFIED
	}
}

func locationFromImpact(loc deleteimpact.Location) *engine.WorkspaceLocation {
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
		SourceSurface:        loc.SourceSurface,
		Ref:                  loc.Ref,
		Title:                loc.Title,
		SourceTitle:          loc.SourceTitle,
	}
}
