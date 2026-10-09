package sources

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/mendahu/provenencia/core/database/rowchange"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/project"
)

const (
	sqlInsertNote = `INSERT INTO source_notes (id, source_id, body) VALUES (?, ?, ?)`
	sqlUpdateNote = `UPDATE source_notes SET body = ? WHERE id = ?`
	sqlDeleteNote = `DELETE FROM source_notes WHERE id = ?`
	sqlGetNote    = `SELECT id, source_id, body FROM source_notes WHERE id = ?`
	// Create attribution lives in audit, not on source_notes (see source-layer-data-model §4.1).
	sqlListNotes = `SELECT n.id, n.source_id, n.body,
		COALESCE((
			SELECT u.display_name
			FROM audit_changes c
			JOIN audit_transactions t ON t.id = c.audit_transaction_id
			LEFT JOIN users u ON u.id = t.user_id
			WHERE c.entity_type = 'source_note' AND c.entity_id = n.id AND c.action = 'create'
			ORDER BY t.revision ASC
			LIMIT 1
		), ''),
		COALESCE((
			SELECT t.created_at
			FROM audit_changes c
			JOIN audit_transactions t ON t.id = c.audit_transaction_id
			WHERE c.entity_type = 'source_note' AND c.entity_id = n.id AND c.action = 'create'
			ORDER BY t.revision ASC
			LIMIT 1
		), '')
		FROM source_notes n
		WHERE n.source_id = ?
		ORDER BY n.id`
	sqlGetNoteAttributed = `SELECT n.id, n.source_id, n.body,
		COALESCE((
			SELECT u.display_name
			FROM audit_changes c
			JOIN audit_transactions t ON t.id = c.audit_transaction_id
			LEFT JOIN users u ON u.id = t.user_id
			WHERE c.entity_type = 'source_note' AND c.entity_id = n.id AND c.action = 'create'
			ORDER BY t.revision ASC
			LIMIT 1
		), ''),
		COALESCE((
			SELECT t.created_at
			FROM audit_changes c
			JOIN audit_transactions t ON t.id = c.audit_transaction_id
			WHERE c.entity_type = 'source_note' AND c.entity_id = n.id AND c.action = 'create'
			ORDER BY t.revision ASC
			LIMIT 1
		), '')
		FROM source_notes n
		WHERE n.id = ?`
	sqlSourceExists = `SELECT 1 FROM sources WHERE id = ?`
)

// Note is one source_notes row plus create attribution from audit for UI.
type Note struct {
	ID                []byte
	SourceID          []byte
	Body              string
	AuthorDisplayName string // users.display_name from create_source_note
	CreatedAt         string // RFC3339 UTC from create audit transaction
}

// AddNote inserts a note. The caller records the returned changes.
// AuthorDisplayName stays empty; the caller fills it from the user after Run.
func AddNote(tx *database.Tx, userID, sourceID []byte, body string) (Note, []rowchange.Change, error) {
	body = strings.TrimSpace(body)
	if tx == nil || len(sourceID) != 16 || body == "" {
		return Note{}, nil, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return Note{}, nil, err
	}
	if err := requireSource(tx.Tx, sourceID); err != nil {
		return Note{}, nil, err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return Note{}, nil, err
	}
	idBytes := id[:]
	if _, err := tx.Exec(sqlInsertNote, idBytes, sourceID, body); err != nil {
		return Note{}, nil, mapConstraint(err)
	}
	return Note{
			ID:        append([]byte(nil), idBytes...),
			SourceID:  append([]byte(nil), sourceID...),
			Body:      body,
			CreatedAt: project.NowUTC(),
		}, []rowchange.Change{{
			EntityType: "source_note",
			EntityID:   append([]byte(nil), idBytes...),
			Action:     rowchange.ActionCreate,
			Fields: map[string]rowchange.FieldDiff{
				"id":        {Old: nil, New: id.String()},
				"source_id": {Old: nil, New: uuidString(sourceID)},
				"body":      {Old: nil, New: body},
			},
		}}, nil
}

// UpdateNote changes body. An unchanged body returns no changes.
func UpdateNote(tx *database.Tx, userID, noteID []byte, body string) ([]rowchange.Change, error) {
	body = strings.TrimSpace(body)
	if tx == nil || len(noteID) != 16 || body == "" {
		return nil, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return nil, err
	}
	prev, err := getNoteTx(tx.Tx, noteID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInvalid
	}
	if err != nil {
		return nil, err
	}
	if prev.Body == body {
		return nil, nil
	}
	if _, err := tx.Exec(sqlUpdateNote, body, noteID); err != nil {
		return nil, err
	}
	return []rowchange.Change{{
		EntityType: "source_note",
		EntityID:   append([]byte(nil), noteID...),
		Action:     rowchange.ActionUpdate,
		Fields: map[string]rowchange.FieldDiff{
			"body": {Old: prev.Body, New: body},
		},
	}}, nil
}

// DeleteNote removes a note. The caller records the returned changes.
func DeleteNote(tx *database.Tx, userID, noteID []byte) ([]rowchange.Change, error) {
	if tx == nil || len(noteID) != 16 {
		return nil, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return nil, err
	}
	prev, err := getNoteTx(tx.Tx, noteID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInvalid
	}
	if err != nil {
		return nil, err
	}
	if _, err := tx.Exec(sqlDeleteNote, noteID); err != nil {
		return nil, err
	}
	return []rowchange.Change{{
		EntityType: "source_note",
		EntityID:   append([]byte(nil), noteID...),
		Action:     rowchange.ActionDelete,
		Fields: map[string]rowchange.FieldDiff{
			"id":        {Old: uuidString(noteID), New: nil},
			"source_id": {Old: uuidString(prev.SourceID), New: nil},
			"body":      {Old: prev.Body, New: nil},
		},
	}}, nil
}

// ListNotes returns notes for a Source with create attribution, ordered by id.
func ListNotes(c *database.Catalog, sourceID []byte) ([]Note, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	if len(sourceID) != 16 {
		return nil, ErrInvalid
	}
	rows, err := db.Query(sqlListNotes, sourceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Note
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.SourceID, &n.Body, &n.AuthorDisplayName, &n.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

// GetNote returns one note with create attribution.
func GetNote(c *database.Catalog, noteID []byte) (Note, error) {
	db, err := c.DB()
	if err != nil {
		return Note{}, err
	}
	if len(noteID) != 16 {
		return Note{}, ErrInvalid
	}
	var n Note
	err = db.QueryRow(sqlGetNoteAttributed, noteID).Scan(
		&n.ID, &n.SourceID, &n.Body, &n.AuthorDisplayName, &n.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Note{}, ErrInvalid
	}
	if err != nil {
		return Note{}, err
	}
	return n, nil
}

func getNoteTx(tx *sql.Tx, id []byte) (Note, error) {
	var n Note
	err := tx.QueryRow(sqlGetNote, id).Scan(&n.ID, &n.SourceID, &n.Body)
	if err != nil {
		return Note{}, err
	}
	return n, nil
}

func requireSource(tx *sql.Tx, sourceID []byte) error {
	var one int
	err := tx.QueryRow(sqlSourceExists, sourceID).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrInvalid
	}
	return err
}
