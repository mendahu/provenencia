package search

import (
	"database/sql"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database/filederivatives"
	"github.com/mendahu/provenencia/core/database/files"
	"github.com/mendahu/provenencia/core/database/sources"
)

// attachLeadStubs fills ThumbnailRelPath / IconKey on the ranked hit list so
// the Mac client can render omnibar leads without listing every Source/type.
// Thumbnails are lookup-only (no EnsureThumbnail) so search stays cheap.
func attachLeadStubs(db *sql.DB, hits []Hit) error {
	if len(hits) == 0 {
		return nil
	}

	sourceIDs := make([][]byte, 0)
	typeIDs := make([][]byte, 0)
	sourceIdx := map[string][]int{}
	typeIdx := map[string][]int{}
	for i, h := range hits {
		parsed, err := uuid.Parse(h.ID)
		if err != nil {
			continue
		}
		bid := make([]byte, 16)
		copy(bid, parsed[:])
		switch h.Kind {
		case KindSource:
			sourceIDs = append(sourceIDs, bid)
			sourceIdx[h.ID] = append(sourceIdx[h.ID], i)
		case KindSourceType:
			typeIDs = append(typeIDs, bid)
			typeIdx[h.ID] = append(typeIdx[h.ID], i)
		}
	}

	if err := fillTypeIcons(db, typeIDs, typeIdx, hits); err != nil {
		return err
	}
	return fillSourceLeads(db, sourceIDs, sourceIdx, hits)
}

func fillTypeIcons(db *sql.DB, typeIDs [][]byte, typeIdx map[string][]int, hits []Hit) error {
	for _, id := range typeIDs {
		var iconKey string
		err := db.QueryRow(
			`SELECT icon_key FROM source_types WHERE id = ?`, id,
		).Scan(&iconKey)
		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			return err
		}
		key := idUUIDString(id)
		for _, i := range typeIdx[key] {
			hits[i].IconKey = strings.TrimSpace(iconKey)
		}
	}
	return nil
}

func fillSourceLeads(db *sql.DB, sourceIDs [][]byte, sourceIdx map[string][]int, hits []Hit) error {
	for _, id := range sourceIDs {
		var (
			coverMode    string
			primaryArtID []byte
			typeIcon     string
		)
		err := db.QueryRow(`
			SELECT s.cover_mode, s.primary_artifact_id, COALESCE(t.icon_key, '')
			FROM sources s
			LEFT JOIN source_types t ON t.id = s.source_type_id
			WHERE s.id = ?
		`, id).Scan(&coverMode, &primaryArtID, &typeIcon)
		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			return err
		}
		key := idUUIDString(id)
		thumb := ""
		if coverMode == sources.CoverModeArtifact && len(primaryArtID) == 16 {
			if rel, err := existingCoverThumbnailRel(db, primaryArtID); err == nil {
				thumb = rel
			}
		}
		for _, i := range sourceIdx[key] {
			hits[i].IconKey = strings.TrimSpace(typeIcon)
			hits[i].ThumbnailRelPath = thumb
		}
	}
	return nil
}

func idUUIDString(id []byte) string {
	if len(id) != 16 {
		return ""
	}
	var u uuid.UUID
	copy(u[:], id)
	return u.String()
}

// existingCoverThumbnailRel returns the objects/… path when a thumbnail
// derivative already exists for the cover artifact's primary file.
func existingCoverThumbnailRel(db *sql.DB, artifactID []byte) (string, error) {
	var fileID []byte
	err := db.QueryRow(`SELECT file_id FROM artifacts WHERE id = ?`, artifactID).Scan(&fileID)
	if err != nil {
		return "", err
	}
	if len(fileID) != 16 {
		return "", nil
	}
	var derivedID []byte
	err = db.QueryRow(
		`SELECT derived_file_id FROM file_derivatives
		 WHERE source_file_id = ? AND derivative_type = ?`,
		fileID, filederivatives.TypeThumbnail,
	).Scan(&derivedID)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if len(derivedID) != 16 {
		return "", nil
	}
	var checksum, mediaType string
	err = db.QueryRow(
		`SELECT checksum_sha256, media_type FROM files WHERE id = ?`, derivedID,
	).Scan(&checksum, &mediaType)
	if err != nil {
		return "", err
	}
	return files.StorageRelPath(checksum, mediaType)
}
