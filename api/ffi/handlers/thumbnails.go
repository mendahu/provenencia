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
	out := &engine.EnsureFileThumbnailResponse{}
	pass := &thumbnailPass{}
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		ok, err := artifacts.HasPrimaryFile(c, fileID)
		if err != nil {
			return err
		}
		if !ok {
			return artifacts.ErrInvalid
		}
		return pass.need(c, fileID, func(rel string, skipped bool) {
			out.RelPath, out.Skipped = rel, skipped
		})
	})
	if err == nil {
		err = pass.run(req.GetProjectDir(), nil)
	}
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

// derivedRelPath is the one link-to-path tail: the objects/… path of the
// link's derived File. A derived File that is gone, or whose stored type has no
// path, is empty and skipped.
func derivedRelPath(c *database.Catalog, link filederivatives.Link) (relPath string, skipped bool, err error) {
	if len(link.DerivedFileID) != 16 {
		return "", true, nil
	}
	f, err := files.Lookup(c, link.DerivedFileID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", true, nil
		}
		return "", false, err
	}
	rel, err := files.ReadRelPath(f)
	if err != nil {
		return "", false, err
	}
	if rel == "" {
		return "", true, nil
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
	rel, _, err := derivedRelPath(c, link)
	return rel, err
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

// sourceCoverThumbnails resolves lookup-only cover rasters for a ListSources
// page in two IN-clause round trips instead of Get+Lookup per row.
func sourceCoverThumbnails(c *database.Catalog, rows []sources.Source) (map[string]string, error) {
	out := make(map[string]string, len(rows))
	var artIDs [][]byte
	need := make([]sources.Source, 0, len(rows))
	for _, s := range rows {
		if s.CoverMode != sources.CoverModeArtifact || len(s.PrimaryArtifactID) != 16 {
			continue
		}
		need = append(need, s)
		artIDs = append(artIDs, s.PrimaryArtifactID)
	}
	if len(artIDs) == 0 {
		return out, nil
	}
	arts, err := artifacts.GetMany(c, artIDs)
	if err != nil {
		return nil, err
	}
	var fileIDs [][]byte
	type coverNeed struct {
		sourceID []byte
		fileID   []byte
	}
	var covers []coverNeed
	for _, s := range need {
		a, ok := arts[string(s.PrimaryArtifactID)]
		if !ok || len(a.FileID) != 16 {
			continue
		}
		covers = append(covers, coverNeed{sourceID: s.ID, fileID: a.FileID})
		fileIDs = append(fileIDs, a.FileID)
	}
	if len(fileIDs) == 0 {
		return out, nil
	}
	links, err := filederivatives.LookupMany(c, fileIDs, filederivatives.TypeThumbnail)
	if err != nil {
		return nil, err
	}
	var derived [][]byte
	for _, link := range links {
		if len(link.DerivedFileID) == 16 {
			derived = append(derived, link.DerivedFileID)
		}
	}
	paths, err := files.RelPaths(c, derived)
	if err != nil {
		return nil, err
	}
	for _, item := range covers {
		link, ok := links[string(item.fileID)]
		if !ok {
			continue
		}
		rel, ok := paths[string(link.DerivedFileID)]
		if !ok {
			continue
		}
		out[string(item.sourceID)] = rel
	}
	return out, nil
}

// enrichSourceProto fills identity + cover mode + resolved thumbnail fields +
// artifact gate for Create/Update/Get workspace responses. Cover thumbs stay
// lookup-only; generation belongs on thumbnailPass (EnsureFileThumbnail and
// SetSourceCover).
func enrichSourceProto(c *database.Catalog, s sources.Source) (*engine.Source, error) {
	if err := sources.AttachActivity(c, &s); err != nil {
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
