// Package citations stores Interpretation Citation rows and orchestrates
// first-submit create with Observations.
package citations

import (
	"database/sql"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/audit"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/project"
	"github.com/mendahu/provenencia/core/locator"
	"github.com/mendahu/provenencia/core/ref"
)

var ErrInvalid = apperr.New(apperr.CodeCitationsInvalid, apperr.KindUser)

const (
	sqlInsert = `INSERT INTO citations (
		id, ref, artifact_id, locator_json, transcription, description,
		transcription_uncertain, transcription_note
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	sqlInsertNote = `INSERT INTO citation_notes (id, citation_id, body) VALUES (?, ?, ?)`

	sqlGet = `SELECT id, ref, artifact_id, locator_json,
		COALESCE(transcription, ''), COALESCE(description, ''),
		transcription_uncertain, COALESCE(transcription_note, '')
		FROM citations WHERE id = ?`

	sqlUpdate = `UPDATE citations SET
		artifact_id = ?, locator_json = ?, transcription = ?, description = ?,
		transcription_uncertain = ?, transcription_note = ?
		WHERE id = ?`

	sqlListNotes = `SELECT body FROM citation_notes WHERE citation_id = ? ORDER BY rowid`

	sqlDeleteNotes = `DELETE FROM citation_notes WHERE citation_id = ?`

	sqlListByArtifact = `SELECT id, ref, artifact_id, locator_json,
		COALESCE(transcription, ''), COALESCE(description, ''),
		transcription_uncertain, COALESCE(transcription_note, '')
		FROM citations WHERE artifact_id = ?
		ORDER BY ref COLLATE NOCASE`

	sqlCountBySource = `SELECT c.artifact_id, COUNT(*)
		FROM citations c
		INNER JOIN artifacts a ON a.id = c.artifact_id
		WHERE a.source_id = ?
		GROUP BY c.artifact_id`

	sqlArtifactExists = `SELECT 1 FROM artifacts WHERE id = ?`

	maxRefRetries = 8
)

// Citation is one citations row.
type Citation struct {
	ID                      []byte
	Ref                     string
	ArtifactID              []byte
	LocatorJSON             string
	Transcription           string
	Description             string
	TranscriptionUncertain  bool
	TranscriptionNote       string
}

// CreateInput is the Citation side of first-submit create.
type CreateInput struct {
	ArtifactID             []byte
	LocatorJSON            string
	Transcription          string
	Description            string
	TranscriptionUncertain bool
	TranscriptionNote      string
	Notes                  []string
}

// CreateResult is the Citation plus Observations created in one transaction.
type CreateResult struct {
	Citation     Citation
	Observations []observations.Observation
}

// CreateWithObservations mints a Citation and ≥1 Observations atomically.
func CreateWithObservations(
	c *database.Catalog,
	userID []byte,
	in CreateInput,
	obsInputs []observations.Input,
) (CreateResult, error) {
	db, err := c.DB()
	if err != nil {
		return CreateResult{}, err
	}
	in.LocatorJSON = strings.TrimSpace(in.LocatorJSON)
	in.Transcription = strings.TrimSpace(in.Transcription)
	in.Description = strings.TrimSpace(in.Description)
	in.TranscriptionNote = strings.TrimSpace(in.TranscriptionNote)
	if len(in.ArtifactID) != 16 || len(obsInputs) == 0 {
		return CreateResult{}, ErrInvalid
	}
	if err := locator.Validate(in.LocatorJSON); err != nil {
		return CreateResult{}, err
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return CreateResult{}, err
	}

	tx, err := db.Begin()
	if err != nil {
		return CreateResult{}, err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := InsertWithObservationsTx(tx, userID, in, obsInputs)
	if err != nil {
		return CreateResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return CreateResult{}, err
	}
	return res, nil
}

// InsertWithObservationsTx writes a Citation + Observations on an open transaction.
func InsertWithObservationsTx(
	tx *sql.Tx,
	userID []byte,
	in CreateInput,
	obsInputs []observations.Input,
) (CreateResult, error) {
	in.LocatorJSON = strings.TrimSpace(in.LocatorJSON)
	in.Transcription = strings.TrimSpace(in.Transcription)
	in.Description = strings.TrimSpace(in.Description)
	in.TranscriptionNote = strings.TrimSpace(in.TranscriptionNote)
	if len(in.ArtifactID) != 16 || len(obsInputs) == 0 {
		return CreateResult{}, ErrInvalid
	}
	if err := locator.Validate(in.LocatorJSON); err != nil {
		return CreateResult{}, err
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return CreateResult{}, err
	}

	var one int
	if err := tx.QueryRow(sqlArtifactExists, in.ArtifactID).Scan(&one); err != nil {
		if err == sql.ErrNoRows {
			return CreateResult{}, ErrInvalid
		}
		return CreateResult{}, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return CreateResult{}, err
	}
	idBytes := id[:]

	uncertain := 0
	if in.TranscriptionUncertain {
		uncertain = 1
	}

	var citRef string
	for attempt := 0; attempt < maxRefRetries; attempt++ {
		citRef, err = ref.Mint(ref.PrefixCitation)
		if err != nil {
			return CreateResult{}, err
		}
		_, err = tx.Exec(
			sqlInsert,
			idBytes,
			citRef,
			in.ArtifactID,
			in.LocatorJSON,
			nullIfEmpty(in.Transcription),
			nullIfEmpty(in.Description),
			uncertain,
			nullIfEmpty(in.TranscriptionNote),
		)
		if err == nil {
			break
		}
		if !database.IsUniqueConflict(err) {
			return CreateResult{}, ErrInvalid
		}
	}
	if err != nil {
		return CreateResult{}, ErrInvalid
	}

	for _, body := range in.Notes {
		body = strings.TrimSpace(body)
		if body == "" {
			continue
		}
		noteID, err := uuid.NewV7()
		if err != nil {
			return CreateResult{}, err
		}
		if _, err := tx.Exec(sqlInsertNote, noteID[:], idBytes, body); err != nil {
			return CreateResult{}, err
		}
	}

	obsOut, obsChanges, err := observations.InsertManyTx(tx, idBytes, obsInputs)
	if err != nil {
		return CreateResult{}, err
	}

	citFields := map[string]audit.FieldDiff{
		"id":          {Old: nil, New: id.String()},
		"ref":         {Old: nil, New: citRef},
		"artifact_id": {Old: nil, New: uuidString(in.ArtifactID)},
	}
	changes := make([]audit.Change, 0, 1+len(obsChanges))
	changes = append(changes, audit.Change{
		EntityType: "citation",
		EntityID:   idBytes,
		Action:     audit.ActionCreate,
		Fields:     citFields,
	})
	changes = append(changes, obsChanges...)

	if _, err := audit.Record(tx, audit.Revision{
		UserID:     userID,
		ActionType: "create_citation_with_observations",
		CreatedAt:  project.NowUTC(),
		Changes:    changes,
	}); err != nil {
		return CreateResult{}, err
	}

	return CreateResult{
		Citation: Citation{
			ID:                     append([]byte(nil), idBytes...),
			Ref:                    citRef,
			ArtifactID:             append([]byte(nil), in.ArtifactID...),
			LocatorJSON:            in.LocatorJSON,
			Transcription:          in.Transcription,
			Description:            in.Description,
			TranscriptionUncertain: in.TranscriptionUncertain,
			TranscriptionNote:      in.TranscriptionNote,
		},
		Observations: obsOut,
	}, nil
}

// UpdateWithObservations replaces Citation fields and all Observations for one citation.
func UpdateWithObservations(
	c *database.Catalog,
	userID, citationID []byte,
	in CreateInput,
	obsInputs []observations.Input,
) (CreateResult, error) {
	db, err := c.DB()
	if err != nil {
		return CreateResult{}, err
	}
	in.LocatorJSON = strings.TrimSpace(in.LocatorJSON)
	in.Transcription = strings.TrimSpace(in.Transcription)
	in.Description = strings.TrimSpace(in.Description)
	in.TranscriptionNote = strings.TrimSpace(in.TranscriptionNote)
	if len(citationID) != 16 || len(in.ArtifactID) != 16 || len(obsInputs) == 0 {
		return CreateResult{}, ErrInvalid
	}
	if err := locator.Validate(in.LocatorJSON); err != nil {
		return CreateResult{}, err
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return CreateResult{}, err
	}

	tx, err := db.Begin()
	if err != nil {
		return CreateResult{}, err
	}
	defer func() { _ = tx.Rollback() }()

	prev, err := scanOne(tx.QueryRow(sqlGet, citationID))
	if err == sql.ErrNoRows {
		return CreateResult{}, ErrInvalid
	}
	if err != nil {
		return CreateResult{}, err
	}

	var one int
	if err := tx.QueryRow(sqlArtifactExists, in.ArtifactID).Scan(&one); err != nil {
		if err == sql.ErrNoRows {
			return CreateResult{}, ErrInvalid
		}
		return CreateResult{}, err
	}

	uncertain := 0
	if in.TranscriptionUncertain {
		uncertain = 1
	}
	if _, err := tx.Exec(
		sqlUpdate,
		in.ArtifactID,
		in.LocatorJSON,
		nullIfEmpty(in.Transcription),
		nullIfEmpty(in.Description),
		uncertain,
		nullIfEmpty(in.TranscriptionNote),
		citationID,
	); err != nil {
		return CreateResult{}, err
	}

	if _, err := tx.Exec(sqlDeleteNotes, citationID); err != nil {
		return CreateResult{}, err
	}
	for _, body := range in.Notes {
		body = strings.TrimSpace(body)
		if body == "" {
			continue
		}
		noteID, err := uuid.NewV7()
		if err != nil {
			return CreateResult{}, err
		}
		if _, err := tx.Exec(sqlInsertNote, noteID[:], citationID, body); err != nil {
			return CreateResult{}, err
		}
	}

	if err := observations.DeleteByCitationTx(tx, citationID); err != nil {
		return CreateResult{}, err
	}
	obsOut, obsChanges, err := observations.InsertManyTx(tx, citationID, obsInputs)
	if err != nil {
		return CreateResult{}, err
	}

	citFields := map[string]audit.FieldDiff{
		"id": {Old: uuidString(citationID), New: uuidString(citationID)},
	}
	if string(prev.ArtifactID) != string(in.ArtifactID) {
		citFields["artifact_id"] = audit.FieldDiff{
			Old: uuidString(prev.ArtifactID),
			New: uuidString(in.ArtifactID),
		}
	}
	if prev.LocatorJSON != in.LocatorJSON {
		citFields["locator_json"] = audit.FieldDiff{Old: prev.LocatorJSON, New: in.LocatorJSON}
	}
	changes := make([]audit.Change, 0, 1+len(obsChanges))
	changes = append(changes, audit.Change{
		EntityType: "citation",
		EntityID:   citationID,
		Action:     audit.ActionUpdate,
		Fields:     citFields,
	})
	changes = append(changes, obsChanges...)

	if _, err := audit.Record(tx, audit.Revision{
		UserID:     userID,
		ActionType: "update_citation_with_observations",
		CreatedAt:  project.NowUTC(),
		Changes:    changes,
	}); err != nil {
		return CreateResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return CreateResult{}, err
	}

	return CreateResult{
		Citation: Citation{
			ID:                     append([]byte(nil), citationID...),
			Ref:                    prev.Ref,
			ArtifactID:             append([]byte(nil), in.ArtifactID...),
			LocatorJSON:            in.LocatorJSON,
			Transcription:          in.Transcription,
			Description:            in.Description,
			TranscriptionUncertain: in.TranscriptionUncertain,
			TranscriptionNote:      in.TranscriptionNote,
		},
		Observations: obsOut,
	}, nil
}

// ListNotes returns citation_notes bodies for a citation.
func ListNotes(c *database.Catalog, citationID []byte) ([]string, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	if len(citationID) != 16 {
		return nil, ErrInvalid
	}
	rows, err := db.Query(sqlListNotes, citationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var body string
		if err := rows.Scan(&body); err != nil {
			return nil, err
		}
		out = append(out, body)
	}
	return out, rows.Err()
}

// Get returns a Citation by id.
func Get(c *database.Catalog, id []byte) (Citation, error) {
	db, err := c.DB()
	if err != nil {
		return Citation{}, err
	}
	if len(id) != 16 {
		return Citation{}, ErrInvalid
	}
	return scanOne(db.QueryRow(sqlGet, id))
}

// ListByArtifact returns Citations for an Artifact.
func ListByArtifact(c *database.Catalog, artifactID []byte) ([]Citation, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	if len(artifactID) != 16 {
		return nil, ErrInvalid
	}
	rows, err := db.Query(sqlListByArtifact, artifactID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Citation
	for rows.Next() {
		cit, err := scanRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, cit)
	}
	return out, rows.Err()
}

// CountBySource returns citation counts keyed by artifact UUID string.
func CountBySource(c *database.Catalog, sourceID []byte) (map[string]int32, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	if len(sourceID) != 16 {
		return nil, ErrInvalid
	}
	rows, err := db.Query(sqlCountBySource, sourceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int32{}
	for rows.Next() {
		var artifactID []byte
		var n int32
		if err := rows.Scan(&artifactID, &n); err != nil {
			return nil, err
		}
		u, err := uuid.FromBytes(artifactID)
		if err != nil {
			return nil, err
		}
		out[u.String()] = n
	}
	return out, rows.Err()
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanOne(row rowScanner) (Citation, error) {
	cit, err := scanRow(row)
	if err == sql.ErrNoRows {
		return Citation{}, err
	}
	return cit, err
}

func scanRow(row rowScanner) (Citation, error) {
	var (
		c         Citation
		uncertain int
	)
	err := row.Scan(
		&c.ID, &c.Ref, &c.ArtifactID, &c.LocatorJSON,
		&c.Transcription, &c.Description, &uncertain, &c.TranscriptionNote,
	)
	if err != nil {
		return Citation{}, err
	}
	c.TranscriptionUncertain = uncertain != 0
	return c, nil
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func uuidString(id []byte) string {
	u, err := uuid.FromBytes(id)
	if err != nil {
		return ""
	}
	return u.String()
}
