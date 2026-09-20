package handlers

import (
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/citations"
	"google.golang.org/protobuf/proto"
)

func CreateCitationWithObservations(in []byte) ([]byte, error) {
	var req engine.CreateCitationWithObservationsRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("create_citation_with_observations", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	artifactID, err := parseID(req.GetArtifactId())
	if err != nil {
		return nil, err
	}
	inputs, err := observationDraftsToInputs(req.GetObservations())
	if err != nil {
		return nil, err
	}
	var out *engine.CreateCitationWithObservationsResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		res, err := citations.CreateWithObservations(c, userID, citations.CreateInput{
			ArtifactID:             artifactID,
			LocatorJSON:            req.GetLocatorJson(),
			Transcription:          req.GetTranscription(),
			Description:            req.GetDescription(),
			TranscriptionUncertain: req.GetTranscriptionUncertain(),
			TranscriptionNote:      req.GetTranscriptionNote(),
			Notes:                  req.GetCitationNotes(),
		}, inputs)
		if err != nil {
			return err
		}
		out = &engine.CreateCitationWithObservationsResponse{
			Citation: citationProto(res.Citation),
		}
		for _, o := range res.Observations {
			out.Observations = append(out.Observations, observationProto(o))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func citationProto(c citations.Citation) *engine.Citation {
	return &engine.Citation{
		Id:                     uuidString(c.ID),
		Ref:                    c.Ref,
		ArtifactId:             uuidString(c.ArtifactID),
		LocatorJson:            c.LocatorJSON,
		Transcription:          c.Transcription,
		Description:            c.Description,
		TranscriptionUncertain: c.TranscriptionUncertain,
		TranscriptionNote:      c.TranscriptionNote,
	}
}
