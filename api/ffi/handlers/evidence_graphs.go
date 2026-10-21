package handlers

import (
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/evidencegraphs"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/writes"
	"google.golang.org/protobuf/proto"
)

func evidenceGraphProto(g evidencegraphs.Graph) *engine.EvidenceGraph {
	out := &engine.EvidenceGraph{
		Id:          database.UUIDString(g.ID),
		Ref:         g.Ref,
		SourceId:    database.UUIDString(g.SourceID),
		Name:        g.Name,
		Description: g.Description,
	}
	for _, id := range g.ArtifactIDs {
		out.ArtifactIds = append(out.ArtifactIds, database.UUIDString(id))
	}
	return out
}

// ListEvidenceGraphs returns every graph of the project, by name.
func ListEvidenceGraphs(in []byte) ([]byte, error) {
	var req engine.ListEvidenceGraphsRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("list_evidence_graphs", err)
	}
	var out *engine.ListEvidenceGraphsResponse
	err := withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		graphs, err := evidencegraphs.List(c, nil)
		if err != nil {
			return err
		}
		out = &engine.ListEvidenceGraphsResponse{}
		for _, g := range graphs {
			out.Graphs = append(out.Graphs, evidenceGraphProto(g))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func CreateEvidenceGraph(in []byte) ([]byte, error) {
	var req engine.CreateEvidenceGraphRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("create_evidence_graph", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	sourceID, err := parseID(req.GetSourceId())
	if err != nil {
		return nil, err
	}
	artifactIDs, err := parseIDs(req.GetArtifactIds())
	if err != nil {
		return nil, err
	}
	var out *engine.CreateEvidenceGraphResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		g, wrote, err := writes.Run(c, writes.Op{Action: "create_evidence_graph", UserID: userID},
			func(tx *database.Tx) (evidencegraphs.Graph, []rowchange.Change, error) {
				return evidencegraphs.Create(tx, userID, evidencegraphs.CreateInput{
					SourceID:    sourceID,
					Name:        req.GetName(),
					Description: req.GetDescription(),
					ArtifactIDs: artifactIDs,
				})
			})
		if err != nil {
			return err
		}
		out = &engine.CreateEvidenceGraphResponse{Graph: evidenceGraphProto(g), Effects: writeEffectsProto(wrote)}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func UpdateEvidenceGraph(in []byte) ([]byte, error) {
	var req engine.UpdateEvidenceGraphRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("update_evidence_graph", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	graphID, err := parseID(req.GetGraphId())
	if err != nil {
		return nil, err
	}
	artifactIDs, err := parseIDs(req.GetArtifactIds())
	if err != nil {
		return nil, err
	}
	var out *engine.UpdateEvidenceGraphResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		_, wrote, err := writes.Run(c, writes.Op{Action: "update_evidence_graph", UserID: userID},
			func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
				changes, err := evidencegraphs.Update(tx, userID, graphID, evidencegraphs.UpdateInput{
					Name:        req.GetName(),
					Description: req.GetDescription(),
					ArtifactIDs: artifactIDs,
				})
				return struct{}{}, changes, err
			})
		if err != nil {
			return err
		}
		out = &engine.UpdateEvidenceGraphResponse{Effects: writeEffectsProto(wrote)}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func DeleteEvidenceGraph(in []byte) ([]byte, error) {
	var req engine.DeleteEvidenceGraphRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("delete_evidence_graph", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	graphID, err := parseID(req.GetGraphId())
	if err != nil {
		return nil, err
	}
	var out *engine.DeleteEvidenceGraphResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		_, wrote, err := writes.Run(c, writes.Op{Action: "delete_evidence_graph", UserID: userID},
			func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
				changes, err := evidencegraphs.Delete(tx, userID, graphID)
				return struct{}{}, changes, err
			})
		if err != nil {
			return err
		}
		out = &engine.DeleteEvidenceGraphResponse{Effects: writeEffectsProto(wrote)}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func LinkEvidenceGraphArtifact(in []byte) ([]byte, error) {
	var req engine.LinkEvidenceGraphArtifactRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("link_evidence_graph_artifact", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	graphID, err := parseID(req.GetGraphId())
	if err != nil {
		return nil, err
	}
	artifactID, err := parseID(req.GetArtifactId())
	if err != nil {
		return nil, err
	}
	var out *engine.LinkEvidenceGraphArtifactResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		_, wrote, err := writes.Run(c, writes.Op{Action: "link_evidence_graph_artifact", UserID: userID},
			func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
				changes, err := evidencegraphs.Link(tx, userID, graphID, artifactID)
				return struct{}{}, changes, err
			})
		if err != nil {
			return err
		}
		out = &engine.LinkEvidenceGraphArtifactResponse{Effects: writeEffectsProto(wrote)}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func UnlinkEvidenceGraphArtifact(in []byte) ([]byte, error) {
	var req engine.UnlinkEvidenceGraphArtifactRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("unlink_evidence_graph_artifact", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	graphID, err := parseID(req.GetGraphId())
	if err != nil {
		return nil, err
	}
	artifactID, err := parseID(req.GetArtifactId())
	if err != nil {
		return nil, err
	}
	var out *engine.UnlinkEvidenceGraphArtifactResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		_, wrote, err := writes.Run(c, writes.Op{Action: "unlink_evidence_graph_artifact", UserID: userID},
			func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
				changes, err := evidencegraphs.Unlink(tx, userID, graphID, artifactID)
				return struct{}{}, changes, err
			})
		if err != nil {
			return err
		}
		out = &engine.UnlinkEvidenceGraphArtifactResponse{Effects: writeEffectsProto(wrote)}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

// GetEvidenceGraphUnlinkImpact is a read: which of the Artifacts a graph edit
// removes are used by the graph's Citations.
func GetEvidenceGraphUnlinkImpact(in []byte) ([]byte, error) {
	var req engine.GetEvidenceGraphUnlinkImpactRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("get_evidence_graph_unlink_impact", err)
	}
	graphID, err := parseID(req.GetGraphId())
	if err != nil {
		return nil, err
	}
	artifactIDs, err := parseIDs(req.GetArtifactIds())
	if err != nil {
		return nil, err
	}
	var out *engine.GetEvidenceGraphUnlinkImpactResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		blocks, err := evidencegraphs.UnlinkImpact(c, graphID, artifactIDs)
		if err != nil {
			return err
		}
		out = &engine.GetEvidenceGraphUnlinkImpactResponse{}
		for _, b := range blocks {
			out.Blocked = append(out.Blocked, &engine.EvidenceGraphUnlinkBlock{
				ArtifactId: database.UUIDString(b.ArtifactID),
				Citations:  groupToProto(b.Citations),
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}
