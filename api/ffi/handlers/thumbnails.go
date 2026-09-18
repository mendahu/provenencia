package handlers

import (
	"database/sql"
	"errors"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/filederivatives"
	"github.com/mendahu/provenencia/core/database/files"
	"github.com/mendahu/provenencia/core/database/sources"
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

// existingThumbnailRelPath returns objects/… when a thumbnail derivative already
// exists. Lookup-only — never calls EnsureThumbnail (ListSources / Source cover).
func existingThumbnailRelPath(c *database.Catalog, sourceFileID []byte) (string, error) {
	if len(sourceFileID) != 16 {
		return "", nil
	}
	link, err := filederivatives.Lookup(c, sourceFileID, filederivatives.TypeThumbnail)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if len(link.DerivedFileID) != 16 {
		return "", nil
	}
	f, err := files.Lookup(c, link.DerivedFileID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return files.StorageRelPath(f.ChecksumSHA256, f.MediaType)
}

// sourceCoverThumb is the ListSources / Source proto cover payload.
type sourceCoverThumb struct {
	RelPath string
}

// sourceCoverThumbnail resolves cover paint fields from persisted cover mode
// without generating derivatives. artifact → existing raster only; type_icon
// (or missing derivative) → empty so the client paints the Source type icon.
func sourceCoverThumbnail(c *database.Catalog, s sources.Source) (sourceCoverThumb, error) {
	if s.CoverMode != sources.CoverModeArtifact || len(s.PrimaryArtifactID) != 16 {
		return sourceCoverThumb{}, nil
	}
	a, err := artifacts.Get(c, s.PrimaryArtifactID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sourceCoverThumb{}, nil
		}
		return sourceCoverThumb{}, err
	}
	if len(a.FileID) != 16 {
		return sourceCoverThumb{}, nil
	}
	rel, err := existingThumbnailRelPath(c, a.FileID)
	if err != nil {
		return sourceCoverThumb{}, err
	}
	return sourceCoverThumb{RelPath: rel}, nil
}

// listSourceProto fills ListSources rows: list SQL already has HasArtifact and
// UpdatedRevision; cover thumbs are lookup-only (no EnsureThumbnail).
func listSourceProto(c *database.Catalog, s sources.Source) (*engine.Source, error) {
	sp := sourceProto(s)
	cover, err := sourceCoverThumbnail(c, s)
	if err != nil {
		return nil, err
	}
	sp.ThumbnailRelPath = cover.RelPath
	sp.HasArtifact = s.HasArtifact
	return sp, nil
}

// enrichSourceProto fills identity + cover mode + resolved thumbnail fields +
// artifact gate for Create/Update/Get workspace responses. Cover thumbs stay
// lookup-only; generation belongs on EnsureFileThumbnail (and SetCover's
// raster check).
func enrichSourceProto(c *database.Catalog, s sources.Source) (*engine.Source, error) {
	if err := sources.AttachLatestRevision(c, &s); err != nil {
		return nil, err
	}
	sp := sourceProto(s)
	cover, err := sourceCoverThumbnail(c, s)
	if err != nil {
		return nil, err
	}
	sp.ThumbnailRelPath = cover.RelPath
	has, err := artifacts.HasAnyForSource(c, s.ID)
	if err != nil {
		return nil, err
	}
	sp.HasArtifact = has
	return sp, nil
}
