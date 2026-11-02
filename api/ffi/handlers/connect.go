package handlers

import (
	"strings"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/connect"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/writes"
	"google.golang.org/protobuf/proto"
)

func CreateCitedBridge(in []byte) ([]byte, error) {
	var req engine.CreateCitedBridgeRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("create_cited_bridge", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	fromID, err := parseID(req.GetFromSubjectId())
	if err != nil {
		return nil, err
	}
	toID, err := parseID(req.GetToSubjectId())
	if err != nil {
		return nil, err
	}
	var citationID []byte
	if s := strings.TrimSpace(req.GetCitationId()); s != "" {
		citationID, err = parseID(s)
		if err != nil {
			return nil, err
		}
	}
	var artifactID []byte
	if citationID == nil {
		artifactID, err = parseID(req.GetArtifactId())
		if err != nil {
			return nil, err
		}
	}
	inputs, err := observationDraftsToInputsAllowEmptySubject(req.GetObservations())
	if err != nil {
		return nil, err
	}
	var out *engine.CreateCitedBridgeResponse
	err = withStagedDepictions(req.GetProjectDir(), inputs, func(c *database.Catalog) error {
		res, wrote, err := writes.Run(c, writes.Op{Action: "create_cited_bridge", UserID: userID},
			func(tx *database.Tx) (connect.Result, []rowchange.Change, error) {
				return connect.CreateCitedBridge(tx, userID, connect.CreateInput{
					FromSubjectID: fromID,
					ToSubjectID:   toID,
					BridgeTypeKey: req.GetBridgeTypeKey(),
					Description:   req.GetDescription(),
					CitationID:    citationID,
					Citation: citations.CreateInput{
						ArtifactID:             artifactID,
						LocatorJSON:            req.GetLocatorJson(),
						Transcription:          req.GetTranscription(),
						Label:                  req.GetCitationLabel(),
						TranscriptionUncertain: req.GetTranscriptionUncertain(),
						TranscriptionNote:      req.GetTranscriptionNote(),
						Notes:                  req.GetCitationNotes(),
					},
					Observations: inputs,
				})
			})
		if err != nil {
			return err
		}
		listed, err := listedWritten(c, res.Citation.ID, res.Observations)
		if err != nil {
			return err
		}
		out = &engine.CreateCitedBridgeResponse{
			Subject:      subjectProto(res.Subject),
			Citation:     citationProto(res.Citation),
			Effects:      writeEffectsProto(wrote),
			Observations: listed,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}
