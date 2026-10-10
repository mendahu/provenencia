package observations

import (
	"database/sql"
	"errors"
	"github.com/mendahu/provenencia/core/database/catalogmodel"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/deleteimpact"
)

// Update rewrites one Observation. Edge rows are always locked, including no-ops.
// An unchanged row returns no changes. The caller records the returned changes.
func Update(tx *database.Tx, userID []byte, in Input) (Listed, []rowchange.Change, error) {
	if tx == nil || len(in.ID) != 16 {
		return Listed{}, nil, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return Listed{}, nil, err
	}

	prev, err := getRowTx(tx.Tx, in.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return Listed{}, nil, ErrInvalid
	}
	if err != nil {
		return Listed{}, nil, err
	}
	storedEdge, err := isEdgeTx(tx.Tx, prev.SubjectID, prev.PropertyID)
	if err != nil {
		return Listed{}, nil, err
	}
	if storedEdge {
		return Listed{}, nil, ErrEdgeLocked
	}
	requestedEdge, err := isEdgeTx(tx.Tx, in.SubjectID, in.PropertyID)
	if err != nil {
		return Listed{}, nil, err
	}
	if requestedEdge {
		return Listed{}, nil, ErrEdgeLocked
	}

	_, changes, err := updateOne(tx.Tx, prev.CitationID, prev, in)
	if err != nil {
		return Listed{}, nil, err
	}
	listed, err := getListedTx(tx.Tx, in.ID)
	if err != nil {
		return Listed{}, nil, err
	}
	return listed, changes, nil
}

// Delete removes one Observation. It never deletes the Citation. Edge rows are locked.
// The caller records the returned changes.
func Delete(tx *database.Tx, userID, id []byte) ([]rowchange.Change, error) {
	if tx == nil || len(id) != 16 {
		return nil, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return nil, err
	}

	prev, err := getRowTx(tx.Tx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInvalid
	}
	if err != nil {
		return nil, err
	}
	edge, err := isEdgeTx(tx.Tx, prev.SubjectID, prev.PropertyID)
	if err != nil {
		return nil, err
	}
	if edge {
		return nil, ErrEdgeLocked
	}

	report, err := deleteimpact.Impact(tx.Tx, catalogmodel.KindObservation, id)
	if err != nil {
		return nil, err
	}
	if err := deleteimpact.Refuse(report, deleteimpact.Codes{
		InUse: ErrInUse, EdgeLocked: ErrEdgeLocked, NotFound: ErrInvalid,
	}); err != nil {
		return nil, err
	}

	// Notes and pins are released first; claims that pinned this Observation
	// stay, weaker (model §5.2). The caller records the returned changes.
	released, err := deleteimpact.ReleaseFacets(tx.Tx, catalogmodel.KindObservation, id)
	if err != nil {
		return nil, err
	}
	snap, err := deleteimpact.SnapshotOwned(tx.Tx, catalogmodel.KindObservation, id)
	if err != nil {
		return nil, err
	}
	if _, err := tx.Exec(sqlDeleteObservationByID, id); err != nil {
		return nil, err
	}
	if err := deleteimpact.ReleaseSnapshot(tx.Tx, snap); err != nil {
		return nil, err
	}
	return append(released.Changes, rowchange.Change{
		EntityType: "observation",
		EntityID:   append([]byte(nil), id...),
		Action:     rowchange.ActionDelete,
		Fields:     rowchange.DeletedRow(observationRowMap(prev)),
	}), nil
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
	_, ok := connectrules.Edge(typeKey, prop.Key)
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

func replaceNotes(tx *sql.Tx, observationID []byte, notes []string) ([]rowchange.Change, error) {
	prev, err := listNotesTx(tx, observationID)
	if err != nil {
		return nil, err
	}
	var changes []rowchange.Change
	for _, note := range prev {
		if _, err := tx.Exec(sqlDeleteNote, note.id); err != nil {
			return nil, err
		}
		changes = append(changes, rowchange.Change{
			EntityType: "observation_note",
			EntityID:   note.id,
			Action:     rowchange.ActionDelete,
			Fields: rowchange.DeletedRow(map[string]any{
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
		changes = append(changes, rowchange.Change{
			EntityType: "observation_note",
			EntityID:   noteID[:],
			Action:     rowchange.ActionCreate,
			Fields: rowchange.FullRow(map[string]any{
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
