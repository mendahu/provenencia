// Package searchindex maintains catalog_search_docs / catalog_search_fts.
// Domain mutators call Reproject* in the same transaction; core/search
// EnsureIndex rebuilds when projection_version lags.
package searchindex

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
)

// Kind ids must match core/search constants (wire SearchHit.kind).
const (
	KindSource      = "source"
	KindSourceType  = "source_type"
	KindSourceField = "source_field"
)

// ProjectionVersion is the Go-side projector shape. Bump when rollup columns
// or document layout change so Open heals old indexes.
const ProjectionVersion = 1

// Querier is *sql.Tx or *sql.DB.
type Querier interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

// Document is one navigable search root ready to write.
type Document struct {
	Kind            string
	EntityID        string // UUID string
	DisplayRef      string
	DisplayTitle    string
	DisplaySubtitle string
	Title           string
	Ref             string
	Secondary       string
	Body            string
}

// Delete removes a root from docs + FTS. No-op if missing.
func Delete(q Querier, kind, entityID string) error {
	kind = strings.TrimSpace(kind)
	entityID = strings.TrimSpace(entityID)
	if kind == "" || entityID == "" {
		return nil
	}
	var rowid int64
	var title, ref, secondary, body string
	err := q.QueryRow(
		`SELECT rowid, title, ref, secondary, body FROM catalog_search_docs WHERE kind = ? AND entity_id = ?`,
		kind, entityID,
	).Scan(&rowid, &title, &ref, &secondary, &body)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	if _, err := q.Exec(
		`INSERT INTO catalog_search_fts(catalog_search_fts, rowid, title, ref, secondary, body) VALUES('delete', ?, ?, ?, ?, ?)`,
		rowid, title, ref, secondary, body,
	); err != nil {
		return err
	}
	_, err = q.Exec(`DELETE FROM catalog_search_docs WHERE rowid = ?`, rowid)
	return err
}

// Upsert replaces the document for (kind, entity_id).
func Upsert(q Querier, doc Document) error {
	doc.Kind = strings.TrimSpace(doc.Kind)
	doc.EntityID = strings.TrimSpace(doc.EntityID)
	if doc.Kind == "" || doc.EntityID == "" || strings.TrimSpace(doc.DisplayTitle) == "" {
		return fmt.Errorf("searchindex: incomplete document")
	}
	if err := Delete(q, doc.Kind, doc.EntityID); err != nil {
		return err
	}
	res, err := q.Exec(
		`INSERT INTO catalog_search_docs (
			kind, entity_id, display_ref, display_title, display_subtitle,
			title, ref, secondary, body
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		doc.Kind,
		doc.EntityID,
		doc.DisplayRef,
		doc.DisplayTitle,
		doc.DisplaySubtitle,
		doc.Title,
		doc.Ref,
		doc.Secondary,
		doc.Body,
	)
	if err != nil {
		return err
	}
	rowid, err := res.LastInsertId()
	if err != nil {
		return err
	}
	_, err = q.Exec(
		`INSERT INTO catalog_search_fts(rowid, title, ref, secondary, body) VALUES (?, ?, ?, ?, ?)`,
		rowid, doc.Title, doc.Ref, doc.Secondary, doc.Body,
	)
	return err
}

// ClearAll wipes the projection (FTS + docs). Meta version is left unchanged.
func ClearAll(q Querier) error {
	if _, err := q.Exec(`INSERT INTO catalog_search_fts(catalog_search_fts) VALUES('delete-all')`); err != nil {
		return err
	}
	_, err := q.Exec(`DELETE FROM catalog_search_docs`)
	return err
}

// ProjectionVersionStored returns the meta row version (0 if missing).
func ProjectionVersionStored(q Querier) (int, error) {
	var v int
	err := q.QueryRow(`SELECT projection_version FROM catalog_search_meta WHERE id = 1`).Scan(&v)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return v, err
}

// SetProjectionVersion writes the meta singleton.
func SetProjectionVersion(q Querier, version int) error {
	_, err := q.Exec(
		`INSERT INTO catalog_search_meta (id, projection_version) VALUES (1, ?)
		 ON CONFLICT(id) DO UPDATE SET projection_version = excluded.projection_version`,
		version,
	)
	return err
}

// NeedsRebuild reports lagging or empty-vs-inconsistent index.
func NeedsRebuild(q Querier) (bool, error) {
	v, err := ProjectionVersionStored(q)
	if err != nil {
		return false, err
	}
	if v != ProjectionVersion {
		return true, nil
	}
	var docs, fts int
	if err := q.QueryRow(`SELECT COUNT(*) FROM catalog_search_docs`).Scan(&docs); err != nil {
		return false, err
	}
	// fts5 'COUNT(*)' on the virtual table counts index rows.
	if err := q.QueryRow(`SELECT COUNT(*) FROM catalog_search_fts`).Scan(&fts); err != nil {
		return false, err
	}
	return docs != fts, nil
}

// ReprojectSource builds the Source navigable document (own fields + rolled children).
func ReprojectSource(q Querier, sourceID []byte) error {
	if len(sourceID) != 16 {
		return nil
	}
	var ref, title, description, typeLabel string
	err := q.QueryRow(`
		SELECT s.ref, s.title, COALESCE(s.description, ''), COALESCE(st.label, '')
		FROM sources s
		LEFT JOIN source_types st ON st.id = s.source_type_id
		WHERE s.id = ?`, sourceID,
	).Scan(&ref, &title, &description, &typeLabel)
	if err == sql.ErrNoRows {
		return Delete(q, KindSource, uuidString(sourceID))
	}
	if err != nil {
		return err
	}

	var bodyParts []string
	noteRows, err := q.Query(`SELECT body FROM source_notes WHERE source_id = ?`, sourceID)
	if err != nil {
		return err
	}
	for noteRows.Next() {
		var body string
		if err := noteRows.Scan(&body); err != nil {
			noteRows.Close()
			return err
		}
		if t := strings.TrimSpace(body); t != "" {
			bodyParts = append(bodyParts, t)
		}
	}
	if err := noteRows.Err(); err != nil {
		noteRows.Close()
		return err
	}
	noteRows.Close()

	metaRows, err := q.Query(`
		SELECT COALESCE(m.value_text, ''), COALESCE(d.phrase, ''),
			COALESCE(CAST(d.start_year AS TEXT), '')
		FROM source_metadata m
		LEFT JOIN date_values d ON d.id = m.date_value_id
		WHERE m.source_id = ?`, sourceID)
	if err != nil {
		return err
	}
	for metaRows.Next() {
		var valueText, phrase, year string
		if err := metaRows.Scan(&valueText, &phrase, &year); err != nil {
			metaRows.Close()
			return err
		}
		for _, p := range []string{valueText, phrase, year} {
			if t := strings.TrimSpace(p); t != "" {
				bodyParts = append(bodyParts, t)
			}
		}
	}
	if err := metaRows.Err(); err != nil {
		metaRows.Close()
		return err
	}
	metaRows.Close()

	artRows, err := q.Query(`
		SELECT a.label, COALESCE(a.description, ''), COALESCE(f.original_filename, '')
		FROM artifacts a
		LEFT JOIN files f ON f.id = a.file_id
		WHERE a.source_id = ?`, sourceID)
	if err != nil {
		return err
	}
	for artRows.Next() {
		var label, desc, filename string
		if err := artRows.Scan(&label, &desc, &filename); err != nil {
			artRows.Close()
			return err
		}
		for _, p := range []string{label, desc, filename} {
			if t := strings.TrimSpace(p); t != "" {
				bodyParts = append(bodyParts, t)
			}
		}
	}
	if err := artRows.Err(); err != nil {
		artRows.Close()
		return err
	}
	artRows.Close()

	secondary := strings.TrimSpace(strings.TrimSpace(description) + " " + strings.TrimSpace(typeLabel))
	return Upsert(q, Document{
		Kind:            KindSource,
		EntityID:        uuidString(sourceID),
		DisplayRef:      ref,
		DisplayTitle:    title,
		DisplaySubtitle: typeLabel,
		Title:           title,
		Ref:             ref,
		Secondary:       secondary,
		Body:            strings.Join(bodyParts, "\n"),
	})
}

// ReprojectSourceType indexes one source_types root.
func ReprojectSourceType(q Querier, typeID []byte) error {
	if len(typeID) != 16 {
		return nil
	}
	var key, label, description string
	err := q.QueryRow(`
		SELECT key, label, COALESCE(description, '') FROM source_types WHERE id = ?`, typeID,
	).Scan(&key, &label, &description)
	if err == sql.ErrNoRows {
		return Delete(q, KindSourceType, uuidString(typeID))
	}
	if err != nil {
		return err
	}
	return Upsert(q, Document{
		Kind:            KindSourceType,
		EntityID:        uuidString(typeID),
		DisplayTitle:    label,
		DisplaySubtitle: key,
		Title:           label,
		Secondary:       strings.TrimSpace(key + " " + description),
	})
}

// ReprojectSourceField indexes one source_metadata_fields root.
func ReprojectSourceField(q Querier, fieldID []byte) error {
	if len(fieldID) != 16 {
		return nil
	}
	var key, label, description string
	err := q.QueryRow(`
		SELECT key, label, COALESCE(description, '') FROM source_metadata_fields WHERE id = ?`, fieldID,
	).Scan(&key, &label, &description)
	if err == sql.ErrNoRows {
		return Delete(q, KindSourceField, uuidString(fieldID))
	}
	if err != nil {
		return err
	}
	return Upsert(q, Document{
		Kind:            KindSourceField,
		EntityID:        uuidString(fieldID),
		DisplayTitle:    label,
		DisplaySubtitle: key,
		Title:           label,
		Secondary:       strings.TrimSpace(key + " " + description),
	})
}

// RebuildAll clears and reprojects every navigable root, then stamps the version.
func RebuildAll(q Querier) error {
	if err := ClearAll(q); err != nil {
		return err
	}
	typeIDs, err := listIDs(q, `SELECT id FROM source_types`)
	if err != nil {
		return err
	}
	for _, id := range typeIDs {
		if err := ReprojectSourceType(q, id); err != nil {
			return err
		}
	}
	fieldIDs, err := listIDs(q, `SELECT id FROM source_metadata_fields`)
	if err != nil {
		return err
	}
	for _, id := range fieldIDs {
		if err := ReprojectSourceField(q, id); err != nil {
			return err
		}
	}
	sourceIDs, err := listIDs(q, `SELECT id FROM sources`)
	if err != nil {
		return err
	}
	for _, id := range sourceIDs {
		if err := ReprojectSource(q, id); err != nil {
			return err
		}
	}
	return SetProjectionVersion(q, ProjectionVersion)
}

// ReprojectSourcesForType refreshes every Source document that uses typeID
// (type label rolls into Source secondary / subtitle).
func ReprojectSourcesForType(q Querier, typeID []byte) error {
	if len(typeID) != 16 {
		return nil
	}
	rows, err := q.Query(`SELECT id FROM sources WHERE source_type_id = ?`, typeID)
	if err != nil {
		return err
	}
	var ids [][]byte
	for rows.Next() {
		var id []byte
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	for _, id := range ids {
		if err := ReprojectSource(q, id); err != nil {
			return err
		}
	}
	return nil
}

// SourceIDsForFile returns owning Source ids for artifacts that reference fileID.
func SourceIDsForFile(q Querier, fileID []byte) ([][]byte, error) {
	if len(fileID) != 16 {
		return nil, nil
	}
	rows, err := q.Query(`SELECT DISTINCT source_id FROM artifacts WHERE file_id = ?`, fileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out [][]byte
	for rows.Next() {
		var id []byte
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// EnsureCatalog rebuilds when needed. Prefer calling after Open/Create ensures.
func EnsureCatalog(c *database.Catalog) error {
	db, err := c.DB()
	if err != nil {
		return err
	}
	need, err := NeedsRebuild(db)
	if err != nil {
		return err
	}
	if !need {
		return nil
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := RebuildAll(tx); err != nil {
		return err
	}
	return tx.Commit()
}

func listIDs(q Querier, query string) ([][]byte, error) {
	rows, err := q.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out [][]byte
	for rows.Next() {
		var id []byte
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func uuidString(id []byte) string {
	if len(id) != 16 {
		return ""
	}
	var u uuid.UUID
	copy(u[:], id)
	return u.String()
}
