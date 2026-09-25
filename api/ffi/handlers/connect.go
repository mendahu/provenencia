package handlers

import (
	"strings"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/connect"
	"github.com/mendahu/provenencia/core/database/observations"
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
	sourceID, err := parseID(req.GetSourceId())
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
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		res, err := connect.CreateCitedBridge(c, userID, connect.CreateInput{
			SourceID:      sourceID,
			FromSubjectID: fromID,
			ToSubjectID:   toID,
			BridgeTypeKey: req.GetBridgeTypeKey(),
			Description:   req.GetDescription(),
			CitationID:    citationID,
			Citation: citations.CreateInput{
				ArtifactID:             artifactID,
				LocatorJSON:            req.GetLocatorJson(),
				Transcription:          req.GetTranscription(),
				Description:            req.GetCitationDescription(),
				TranscriptionUncertain: req.GetTranscriptionUncertain(),
				TranscriptionNote:      req.GetTranscriptionNote(),
				Notes:                  req.GetCitationNotes(),
			},
			Observations: inputs,
		})
		if err != nil {
			return err
		}
		listed, err := observations.ListByCitation(c, res.Citation.ID)
		if err != nil {
			return err
		}
		keep := make(map[string]struct{}, len(res.Observations))
		for _, o := range res.Observations {
			keep[uuidString(o.ID)] = struct{}{}
		}
		out = &engine.CreateCitedBridgeResponse{
			Subject:  subjectProto(res.Subject),
			Citation: citationProto(res.Citation),
		}
		for _, row := range listed {
			if _, ok := keep[uuidString(row.ID)]; !ok {
				continue
			}
			out.Observations = append(out.Observations, listedObservationProto(row))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}
