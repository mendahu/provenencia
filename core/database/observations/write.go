package observations

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/audit"
	"github.com/mendahu/provenencia/core/database/project"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
)

// Update rewrites one Observation. Edge rows are always locked, including no-ops.
func Update(c *database.Catalog, userID []byte, in Input) (Listed, error) {
	db, err := c.DB()
	if err != nil {
		return Listed{}, err
	}
	if len(in.ID) != 16 {
		return Listed{}, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return Listed{}, err
	}

	tx, err := db.Begin()
	if err != nil {
		return Listed{}, err
	}
	defer func() { _ = tx.Rollback() }()

	prev, err := getRowTx(tx, in.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return Listed{}, ErrInvalid
	}
	if err != nil {
		return Listed{}, err
	}
	storedEdge, err := isEdgeTx(tx, prev.SubjectID, prev.PropertyID)
	if err != nil {
		return Listed{}, err
	}
	if storedEdge {
		return Listed{}, ErrEdgeLocked
	}
	requestedEdge, err := isEdgeTx(tx, in.SubjectID, in.PropertyID)
	if err != nil {
		return Listed{}, err
	}
	if requestedEdge {
		return Listed{}, ErrEdgeLocked
	}

	_, changes, err := updateOne(tx, prev.CitationID, prev, in)
	if err != nil {
		return Listed{}, err
	}
	if len(changes) > 0 {
		if _, err := audit.Record(tx, audit.Revision{
			UserID:     userID,
			ActionType: "update_observation",
			CreatedAt:  project.NowUTC(),
			Changes:    changes,
		}); err != nil {
			return Listed{}, err
		}
	}
	listed, err := getListedTx(tx, in.ID)
	if err != nil {
		return Listed{}, err
	}
	if err := tx.Commit(); err != nil {
		return Listed{}, err
	}
	return listed, nil
}

// Delete removes one Observation. It never deletes the Citation. Edge rows are locked.
func Delete(c *database.Catalog, userID, id []byte) error {
	db, err := c.DB()
	if err != nil {
		return err
	}
	if len(id) != 16 {
		return ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	prev, err := getRowTx(tx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrInvalid
	}
	if err != nil {
		return err
	}
	edge, err := isEdgeTx(tx, prev.SubjectID, prev.PropertyID)
	if err != nil {
		return err
	}
	if edge {
		return ErrEdgeLocked
	}

	notes, err := listNotesTx(tx, id)
	if err != nil {
		return err
	}
	changes := make([]audit.Change, 0, 2+len(notes))
	for _, note := range notes {
		if _, err := tx.Exec(sqlDeleteNote, note.id); err != nil {
			return err
		}
		changes = append(changes, audit.Change{
			EntityType: "observation_note",
			EntityID:   note.id,
			Action:     audit.ActionDelete,
			Fields: audit.DeletedRow(map[string]any{
				"id":             uuidJSON(note.id),
				"observation_id": uuidJSON(id),
				"body":           note.body,
			}),
		})
	}
	if _, err := tx.Exec(sqlDeleteObservationByID, id); err != nil {
		return err
	}
	changes = append(changes, audit.Change{
		EntityType: "observation",
		EntityID:   id,
		Action:     audit.ActionDelete,
		Fields:     audit.DeletedRow(observationRowMap(prev)),
	})
	if deleted, err := releaseDateValue(tx, prev.ValueDateID, nil); err != nil {
		return err
	} else if deleted != nil {
		changes = append(changes, audit.Change{
			EntityType: "date_value",
			EntityID:   append([]byte(nil), prev.ValueDateID...),
			Action:     audit.ActionDelete,
			Fields:     audit.DeletedRow(deleted),
		})
	}
	if deleted, err := releaseNameValue(tx, prev.ValueNameID, nil); err != nil {
		return err
	} else if deleted != nil {
		changes = append(changes, audit.Change{
			EntityType: "name_value",
			EntityID:   append([]byte(nil), prev.ValueNameID...),
			Action:     audit.ActionDelete,
			Fields:     audit.DeletedRow(deleted),
		})
	}
	if _, err := audit.Record(tx, audit.Revision{
		UserID:     userID,
		ActionType: "delete_observation",
		CreatedAt:  project.NowUTC(),
		Changes:    changes,
	}); err != nil {
		return err
	}
	return tx.Commit()
}

func isEdgeTx(tx *sql.Tx, subjectID, propertyID []byte) (bool, error) {
	var typeKey string
	if err := tx.QueryRow(sqlSubjectTypeKey, subjectID).Scan(&typeKey); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, ErrInvalid
		}
		return false, err
	}
	prop, err := getPropertyTx(tx, propertyID)
	if err != nil {
		return false, err
	}
	_, ok := subjectvocab.EdgeEndpoint(typeKey, prop.Key)
	return ok, nil
}

func getRowTx(tx *sql.Tx, id []byte) (Observation, error) {
	return scanRawObservation(tx.QueryRow(sqlGetRow, id))
}

func scanRawObservation(row rowScanner) (Observation, error) {
	var (
		obs               Observation
		valueText         sql.NullString
		valueInt          sql.NullInt64
		dateID, nameID    []byte
		subjectID, termID []byte
	)
	if err := row.Scan(
		&obs.ID, &obs.Ref, &obs.CitationID, &obs.SubjectID, &obs.PropertyID, &obs.Polarity,
		&valueText, &valueInt, &dateID, &nameID, &subjectID, &termID,
	); err != nil {
		return Observation{}, err
	}
	if valueText.Valid {
		obs.ValueText = valueText.String
		obs.HasText = obs.ValueText != ""
	}
	if valueInt.Valid {
		obs.ValueInteger = valueInt.Int64
		obs.HasInteger = true
	}
	obs.ValueDateID = append([]byte(nil), dateID...)
	obs.ValueNameID = append([]byte(nil), nameID...)
	obs.ValueSubjectID = append([]byte(nil), subjectID...)
	obs.ValueTermID = append([]byte(nil), termID...)
	return obs, nil
}

type observationNote struct {
	id   []byte
	body string
}

func listNotesTx(tx *sql.Tx, observationID []byte) ([]observationNote, error) {
	rows, err := tx.Query(sqlListNotesByObservation, observationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []observationNote
	for rows.Next() {
		var n observationNote
		if err := rows.Scan(&n.id, &n.body); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func replaceNotes(tx *sql.Tx, observationID []byte, notes []string) ([]audit.Change, error) {
	prev, err := listNotesTx(tx, observationID)
	if err != nil {
		return nil, err
	}
	var changes []audit.Change
	for _, note := range prev {
		if _, err := tx.Exec(sqlDeleteNote, note.id); err != nil {
			return nil, err
		}
		changes = append(changes, audit.Change{
			EntityType: "observation_note",
			EntityID:   note.id,
			Action:     audit.ActionDelete,
			Fields: audit.DeletedRow(map[string]any{
				"id":             uuidJSON(note.id),
				"observation_id": uuidJSON(observationID),
				"body":           note.body,
			}),
		})
	}
	for _, body := range notes {
		body = strings.TrimSpace(body)
		if body == "" {
			continue
		}
		noteID, err := uuid.NewV7()
		if err != nil {
			return nil, err
		}
		if _, err := tx.Exec(sqlInsertNote, noteID[:], observationID, body); err != nil {
			return nil, err
		}
		changes = append(changes, audit.Change{
			EntityType: "observation_note",
			EntityID:   noteID[:],
			Action:     audit.ActionCreate,
			Fields: audit.FullRow(map[string]any{
				"id":             noteID.String(),
				"observation_id": uuidJSON(observationID),
				"body":           body,
			}),
		})
	}
	return changes, nil
}

func getListedTx(tx *sql.Tx, id []byte) (Listed, error) {
	rows, err := tx.Query(sqlListSelect+" WHERE o.id = ?", id)
	if err != nil {
		return Listed{}, err
	}
	defer rows.Close()
	listed, err := scanListed(rows)
	if err != nil {
		return Listed{}, err
	}
	if len(listed) != 1 {
		return Listed{}, ErrInvalid
	}
	return listed[0], nil
}

func observationRowMap(o Observation) map[string]any {
	var integer any
	if o.HasInteger {
		integer = o.ValueInteger
	}
	return map[string]any{
		"id":               uuidJSON(o.ID),
		"ref":              o.Ref,
		"citation_id":      uuidJSON(o.CitationID),
		"subject_id":       uuidJSON(o.SubjectID),
		"property_id":      uuidJSON(o.PropertyID),
		"polarity":         o.Polarity,
		"value_text":       emptyAsNil(o.ValueText),
		"value_integer":    integer,
		"value_date_id":    uuidJSON(o.ValueDateID),
		"value_name_id":    uuidJSON(o.ValueNameID),
		"value_subject_id": uuidJSON(o.ValueSubjectID),
		"value_term_id":    uuidJSON(o.ValueTermID),
	}
}

type rowScanner interface {
	Scan(dest ...any) error
}
