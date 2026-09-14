package search

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database/filederivatives"
	"github.com/mendahu/provenencia/core/database/files"
	"github.com/mendahu/provenencia/core/database/sources"
)

// attachLeadStubs fills ThumbnailRelPath / IconKey on the ranked hit list so
// the Mac client can render omnibar leads without listing every Source/type.
// Thumbnails are lookup-only (no EnsureThumbnail) so search stays cheap.
//
// Uses at most two batched SELECTs (types, sources+cover) keyed by the hit
// IDs — never one QueryRow per hit.
func attachLeadStubs(db *sql.DB, hits []Hit) error {
	if len(hits) == 0 {
		return nil
	}

	sourceIDs := make([][]byte, 0)
	typeIDs := make([][]byte, 0)
	sourceSeen := map[string]struct{}{}
	typeSeen := map[string]struct{}{}
	sourceIdx := map[string][]int{}
	typeIdx := map[string][]int{}
	for i, h := range hits {
		parsed, err := uuid.Parse(h.ID)
		if err != nil {
			continue
		}
		switch h.Kind {
		case KindSource:
			sourceIdx[h.ID] = append(sourceIdx[h.ID], i)
			if _, ok := sourceSeen[h.ID]; ok {
				continue
			}
			sourceSeen[h.ID] = struct{}{}
			bid := make([]byte, 16)
			copy(bid, parsed[:])
			sourceIDs = append(sourceIDs, bid)
		case KindSourceType:
			typeIdx[h.ID] = append(typeIdx[h.ID], i)
			if _, ok := typeSeen[h.ID]; ok {
				continue
			}
			typeSeen[h.ID] = struct{}{}
			bid := make([]byte, 16)
			copy(bid, parsed[:])
			typeIDs = append(typeIDs, bid)
		}
	}

	if err := fillTypeIcons(db, typeIDs, typeIdx, hits); err != nil {
		return err
	}
	return fillSourceLeads(db, sourceIDs, sourceIdx, hits)
}

func fillTypeIcons(db *sql.DB, typeIDs [][]byte, typeIdx map[string][]int, hits []Hit) error {
	if len(typeIDs) == 0 {
		return nil
	}
	q := fmt.Sprintf(
		`SELECT id, icon_key FROM source_types WHERE id IN (%s)`,
		placeholders(len(typeIDs)),
	)
	rows, err := db.Query(q, blobArgs(typeIDs)...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id []byte
		var iconKey string
		if err := rows.Scan(&id, &iconKey); err != nil {
			return err
		}
		key := idUUIDString(id)
		for _, i := range typeIdx[key] {
			hits[i].IconKey = strings.TrimSpace(iconKey)
		}
	}
	return rows.Err()
}

func fillSourceLeads(db *sql.DB, sourceIDs [][]byte, sourceIdx map[string][]int, hits []Hit) error {
	if len(sourceIDs) == 0 {
		return nil
	}
	// One query: type icon + existing cover thumbnail (when already derived).
	q := fmt.Sprintf(`
		SELECT s.id,
			COALESCE(t.icon_key, ''),
			COALESCE(f.checksum_sha256, ''),
			COALESCE(f.media_type, '')
		FROM sources s
		LEFT JOIN source_types t ON t.id = s.source_type_id
		LEFT JOIN artifacts a
			ON s.cover_mode = ?
			AND a.id = s.primary_artifact_id
			AND length(a.file_id) = 16
		LEFT JOIN file_derivatives d
			ON d.source_file_id = a.file_id
			AND d.derivative_type = ?
		LEFT JOIN files f ON f.id = d.derived_file_id
		WHERE s.id IN (%s)
	`, placeholders(len(sourceIDs)))

	args := make([]any, 0, 2+len(sourceIDs))
	args = append(args, sources.CoverModeArtifact, filederivatives.TypeThumbnail)
	args = append(args, blobArgs(sourceIDs)...)

	rows, err := db.Query(q, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id                  []byte
			typeIcon            string
			checksum, mediaType string
		)
		if err := rows.Scan(&id, &typeIcon, &checksum, &mediaType); err != nil {
			return err
		}
		thumb := ""
		if checksum != "" {
			if rel, err := files.StorageRelPath(checksum, mediaType); err == nil {
				thumb = rel
			}
		}
		key := idUUIDString(id)
		for _, i := range sourceIdx[key] {
			hits[i].IconKey = strings.TrimSpace(typeIcon)
			hits[i].ThumbnailRelPath = thumb
		}
	}
	return rows.Err()
}

func placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	b := strings.Builder{}
	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteByte('?')
	}
	return b.String()
}

func blobArgs(ids [][]byte) []any {
	out := make([]any, len(ids))
	for i, id := range ids {
		out[i] = id
	}
	return out
}

func idUUIDString(id []byte) string {
	if len(id) != 16 {
		return ""
	}
	var u uuid.UUID
	copy(u[:], id)
	return u.String()
}
