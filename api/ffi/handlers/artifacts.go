package handlers

import (
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/files"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/ingest"
	"google.golang.org/protobuf/proto"
)

func CreateArtifact(in []byte) ([]byte, error) {
	var req engine.CreateArtifactRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("create_artifact", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	sourceID, err := parseID(req.GetSourceId())
	if err != nil {
		return nil, err
	}
	var fileID []byte
	if req.GetFileId() != "" {
		fileID, err = parseID(req.GetFileId())
		if err != nil {
			return nil, err
		}
	}
	var out *engine.CreateArtifactResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		a, err := artifacts.Create(c, userID, artifacts.CreateInput{
			SourceID:    sourceID,
			FileID:      fileID,
			Label:       req.GetLabel(),
			Description: req.GetDescription(),
		})
		if err != nil {
			return err
		}
		if len(a.FileID) == 16 {
			if _, _, err := sources.MaybePinFirstFileCover(c, userID, a.SourceID, a.ID); err != nil {
				return err
			}
		}
		ap, err := artifactProto(c, a)
		if err != nil {
			return err
		}
		out = &engine.CreateArtifactResponse{Artifact: ap}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func UpdateArtifact(in []byte) ([]byte, error) {
	var req engine.UpdateArtifactRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("update_artifact", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	artifactID, err := parseID(req.GetArtifactId())
	if err != nil {
		return nil, err
	}
	var out *engine.UpdateArtifactResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		prev, err := artifacts.Get(c, artifactID)
		if err != nil {
			return err
		}
		updated := artifacts.Artifact{
			ID:          prev.ID,
			Ref:         prev.Ref,
			SourceID:    prev.SourceID,
			FileID:      prev.FileID,
			Label:       req.GetLabel(),
			Description: req.GetDescription(),
		}
		if err := artifacts.Update(c, userID, updated); err != nil {
			return err
		}
		got, err := artifacts.Get(c, artifactID)
		if err != nil {
			return err
		}
		ap, err := artifactProto(c, got)
		if err != nil {
			return err
		}
		out = &engine.UpdateArtifactResponse{Artifact: ap}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func IngestArtifactFile(in []byte) ([]byte, error) {
	var req engine.IngestArtifactFileRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("ingest_artifact_file", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	artifactID, err := parseID(req.GetArtifactId())
	if err != nil {
		return nil, err
	}
	var out *engine.IngestArtifactFileResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		prev, err := artifacts.Get(c, artifactID)
		if err != nil {
			return err
		}
		if len(prev.FileID) == 16 {
			return artifacts.ErrFileAlreadyAttached
		}
		res, err := ingest.File(c, req.GetPath(), userID)
		if err != nil {
			return err
		}
		updated := artifacts.Artifact{
			ID:          prev.ID,
			Ref:         prev.Ref,
			SourceID:    prev.SourceID,
			FileID:      res.File.ID,
			Label:       prev.Label,
			Description: prev.Description,
		}
		if err := artifacts.Update(c, userID, updated); err != nil {
			return err
		}
		if _, _, err := sources.MaybePinFirstFileCover(c, userID, prev.SourceID, artifactID); err != nil {
			return err
		}
		got, err := artifacts.Get(c, artifactID)
		if err != nil {
			return err
		}
		ap, err := artifactProto(c, got)
		if err != nil {
			return err
		}
		out = &engine.IngestArtifactFileResponse{
			Artifact: ap,
			File:     fileRefProto(res.File, res.RelPath),
			Reused:   res.Reused,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func listArtifactsProto(c *database.Catalog, sourceID []byte) ([]*engine.Artifact, error) {
	rows, err := artifacts.ListBySource(c, sourceID)
	if err != nil {
		return nil, err
	}
	out := make([]*engine.Artifact, 0, len(rows))
	for _, a := range rows {
		ap, err := artifactProto(c, a)
		if err != nil {
			return nil, err
		}
		out = append(out, ap)
	}
	return out, nil
}

func artifactProto(c *database.Catalog, a artifacts.Artifact) (*engine.Artifact, error) {
	out := &engine.Artifact{
		Id:          uuidString(a.ID),
		Ref:         a.Ref,
		SourceId:    uuidString(a.SourceID),
		FileId:      uuidString(a.FileID),
		Label:       a.Label,
		Description: a.Description,
	}
	if len(a.FileID) == 16 {
		f, err := files.Lookup(c, a.FileID)
		if err != nil {
			return nil, err
		}
		rel, err := files.StorageRelPath(f.ChecksumSHA256, f.MediaType)
		if err != nil {
			return nil, err
		}
		out.File = fileRefProto(f, rel)
		thumb, _, err := thumbnailRelPath(c, a.FileID)
		if err != nil {
			return nil, err
		}
		out.ThumbnailRelPath = thumb
	}
	return out, nil
}
