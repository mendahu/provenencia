package handlers

import (
	"database/sql"
	"errors"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/files"
	"github.com/mendahu/provenencia/core/derivatives"
	"google.golang.org/protobuf/proto"
)

// EnsureFileThumbnail lazily creates (or returns) the default thumbnail for a
// File that is the primary File of some Artifact in the project.
func EnsureFileThumbnail(in []byte) ([]byte, error) {
	var req engine.EnsureFileThumbnailRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("ensure_file_thumbnail", err)
	}
	fileID, err := parseID(req.GetFileId())
	if err != nil {
		return nil, err
	}
	var out *engine.EnsureFileThumbnailResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		ok, err := artifacts.HasPrimaryFile(c, fileID)
		if err != nil {
			return err
		}
		if !ok {
			return artifacts.ErrInvalid
		}

		rel, skipped, err := thumbnailRelPath(c, fileID)
		if err != nil {
			return err
		}
		out = &engine.EnsureFileThumbnailResponse{
			RelPath: rel,
			Skipped: skipped,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

// thumbnailRelPath ensures the default thumbnail for sourceFileID and returns
// its objects/… path. Skipped / unprocessable / corrupt → empty path, skipped.
func thumbnailRelPath(c *database.Catalog, sourceFileID []byte) (relPath string, skipped bool, err error) {
	if len(sourceFileID) != 16 {
		return "", true, nil
	}
	res, err := derivatives.EnsureThumbnail(c, sourceFileID)
	if err != nil {
		if errors.Is(err, derivatives.ErrUnprocessable) ||
			errors.Is(err, derivatives.ErrCorruptObject) ||
			errors.Is(err, derivatives.ErrInvalid) {
			return "", true, nil
		}
		return "", false, err
	}
	if res.Skipped || len(res.Link.DerivedFileID) != 16 {
		return "", true, nil
	}
	f, err := files.Lookup(c, res.Link.DerivedFileID)
	if err != nil {
		return "", false, err
	}
	rel, err := files.StorageRelPath(f.ChecksumSHA256, f.MediaType)
	if err != nil {
		return "", false, err
	}
	return rel, false, nil
}

// sourceCoverThumb is the ListSources cover payload for a Source.
type sourceCoverThumb struct {
	RelPath          string
	MediaType        string
	OriginalFilename string
}

// sourceCoverThumbnail prefers the first Artifact (by ref) with a successful
// raster thumbnail. Otherwise it returns empty path plus MIME/filename from
// the first file-bearing Artifact so clients can render a file-type glyph.
func sourceCoverThumbnail(c *database.Catalog, sourceID []byte) (sourceCoverThumb, error) {
	arts, err := artifacts.ListBySource(c, sourceID)
	if err != nil {
		return sourceCoverThumb{}, err
	}
	var firstFile sourceCoverThumb
	for _, a := range arts {
		if len(a.FileID) != 16 {
			continue
		}
		f, err := files.Lookup(c, a.FileID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return sourceCoverThumb{}, err
		}
		if firstFile.MediaType == "" && firstFile.OriginalFilename == "" {
			firstFile = sourceCoverThumb{
				MediaType:        f.MediaType,
				OriginalFilename: f.OriginalFilename,
			}
		}
		rel, _, err := thumbnailRelPath(c, a.FileID)
		if err != nil {
			return sourceCoverThumb{}, err
		}
		if rel != "" {
			return sourceCoverThumb{RelPath: rel}, nil
		}
	}
	return firstFile, nil
}
