package handlers

import (
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/observations"
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

func CitationCountsBySource(in []byte) ([]byte, error) {
	var req engine.CitationCountsBySourceRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("citation_counts_by_source", err)
	}
	sourceID, err := parseID(req.GetSourceId())
	if err != nil {
		return nil, err
	}
	var out *engine.CitationCountsBySourceResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		counts, err := citations.CountBySource(c, sourceID)
		if err != nil {
			return err
		}
		out = &engine.CitationCountsBySourceResponse{}
		for artifactID, n := range counts {
			out.Counts = append(out.Counts, &engine.ArtifactCitationCount{
				ArtifactId: artifactID,
				Count:      n,
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func GetCitation(in []byte) ([]byte, error) {
	var req engine.GetCitationRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("get_citation", err)
	}
	citationID, err := parseID(req.GetCitationId())
	if err != nil {
		return nil, err
	}
	var out *engine.GetCitationResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		cit, err := citations.Get(c, citationID)
		if err != nil {
			return err
		}
		notes, err := citations.ListNotes(c, citationID)
		if err != nil {
			return err
		}
		listed, err := observations.ListByCitation(c, citationID)
		if err != nil {
			return err
		}
		out = &engine.GetCitationResponse{
			Citation: citationProto(cit),
			Notes:    notes,
		}
		for _, row := range listed {
			out.Observations = append(out.Observations, listedObservationProto(row))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func UpdateCitation(in []byte) ([]byte, error) {
	var req engine.UpdateCitationRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("update_citation", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	citationID, err := parseID(req.GetCitationId())
	if err != nil {
		return nil, err
	}
	var out *engine.UpdateCitationResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		cit, err := citations.Update(c, userID, citationID, citations.CitationFieldsInput{
			LocatorJSON:            req.GetLocatorJson(),
			Transcription:          req.GetTranscription(),
			Description:            req.GetDescription(),
			TranscriptionUncertain: req.GetTranscriptionUncertain(),
			TranscriptionNote:      req.GetTranscriptionNote(),
		})
		if err != nil {
			return err
		}
		out = &engine.UpdateCitationResponse{Citation: citationProto(cit)}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func ListCitationsByArtifact(in []byte) ([]byte, error) {
	var req engine.ListCitationsByArtifactRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("list_citations_by_artifact", err)
	}
	artifactID, err := parseID(req.GetArtifactId())
	if err != nil {
		return nil, err
	}
	var out *engine.ListCitationsByArtifactResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		list, err := citations.ListByArtifact(c, artifactID)
		if err != nil {
			return err
		}
		out = &engine.ListCitationsByArtifactResponse{}
		for _, row := range list {
			out.Citations = append(out.Citations, &engine.ListedCitation{
				Citation:         citationProto(row.Citation),
				ObservationCount: row.ObservationCount,
			})
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
