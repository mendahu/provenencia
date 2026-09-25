package observations

import (
	"bytes"
	"database/sql"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database/audit"
)

const (
	sqlUpdateObservation = `UPDATE observations SET
		subject_id = ?, property_id = ?, polarity = ?,
		value_text = ?, value_integer = ?,
		value_date_id = ?, value_name_id = ?, value_subject_id = ?, value_term_id = ?
		WHERE id = ? AND citation_id = ?`

	sqlDeleteObservation = `DELETE FROM observations WHERE id = ? AND citation_id = ?`

	sqlDeleteObservationNotes = `DELETE FROM observation_notes WHERE observation_id = ?`

	sqlListRowsByCitation = `SELECT id, ref, citation_id, subject_id, property_id, polarity,
		value_text, value_integer, value_date_id, value_name_id, value_subject_id, value_term_id
		FROM observations WHERE citation_id = ?`

	sqlCountDateRefs = `SELECT
		(SELECT COUNT(*) FROM observations WHERE value_date_id = ?) +
		(SELECT COUNT(*) FROM source_metadata WHERE date_value_id = ?)`

	sqlCountNameRefs = `SELECT COUNT(*) FROM observations WHERE value_name_id = ?`

	sqlDeleteDateValue = `DELETE FROM date_values WHERE id = ?`

	sqlDeleteNameValue = `DELETE FROM name_values WHERE id = ?`
)

// ApplyForCitationTx updates each input that already belongs to citationID,
// inserts inputs with no id, and deletes stored rows missing from inputs.
func ApplyForCitationTx(tx *sql.Tx, citationID []byte, inputs []Input) ([]Observation, []audit.Change, error) {
	if tx == nil || len(citationID) != 16 {
		return nil, nil, ErrInvalid
	}
	existing, err := listRowsByCitationTx(tx, citationID)
	if err != nil {
		return nil, nil, err
	}
	byID := make(map[string]Observation, len(existing))
	for _, row := range existing {
		byID[uuidString(row.ID)] = row
	}

	out := make([]Observation, 0, len(inputs))
	changes := make([]audit.Change, 0, len(inputs))
	seen := make(map[string]struct{}, len(inputs))
	for _, in := range inputs {
		switch {
		case len(in.ID) == 0:
			obs, chs, err := insertOne(tx, citationID, in)
			if err != nil {
				return nil, nil, err
			}
			out = append(out, obs)
			changes = append(changes, chs...)
		case len(in.ID) != 16:
			return nil, nil, ErrInvalid
		default:
			key := uuidString(in.ID)
			prev, ok := byID[key]
			if key == "" || !ok {
				return nil, nil, ErrInvalid
			}
			if _, dup := seen[key]; dup {
				return nil, nil, ErrInvalid
			}
			seen[key] = struct{}{}
			obs, ch, err := updateOne(tx, citationID, prev, in)
			if err != nil {
				return nil, nil, err
			}
			out = append(out, obs)
			if ch != nil {
				changes = append(changes, *ch)
			}
		}
	}

	for _, prev := range existing {
		key := uuidString(prev.ID)
		if _, ok := seen[key]; ok {
			continue
		}
		if _, err := tx.Exec(sqlDeleteObservation, prev.ID, citationID); err != nil {
			return nil, nil, err
		}
		changes = append(changes, audit.Change{
			EntityType: "observation",
			EntityID:   prev.ID,
			Action:     audit.ActionDelete,
			Fields: map[string]audit.FieldDiff{
				"id":  {Old: uuidString(prev.ID), New: nil},
				"ref": {Old: prev.Ref, New: nil},
			},
		})
	}
	return out, changes, nil
}

func updateOne(tx *sql.Tx, citationID []byte, prev Observation, in Input) (Observation, *audit.Change, error) {
	polarity := stringsTrimPolarity(in.Polarity)
	if polarity != PolarityPositive && polarity != PolarityNegative {
		return Observation{}, nil, ErrInvalid
	}
	if len(in.SubjectID) != 16 || len(in.PropertyID) != 16 {
		return Observation{}, nil, ErrInvalid
	}

	var subjectTypeID []byte
	if err := tx.QueryRow(sqlSubjectType, in.SubjectID).Scan(&subjectTypeID); err != nil {
		if err == sql.ErrNoRows {
			return Observation{}, nil, ErrInvalid
		}
		return Observation{}, nil, err
	}
	var one int
	if err := tx.QueryRow(sqlBindingExists, subjectTypeID, in.PropertyID).Scan(&one); err != nil {
		if err == sql.ErrNoRows {
			return Observation{}, nil, ErrInvalid
		}
		return Observation{}, nil, err
	}
	prop, err := getPropertyTx(tx, in.PropertyID)
	if err != nil {
		return Observation{}, nil, err
	}
	resolved, err := resolveValue(tx, prop.ValueType, in, &prev)
	if err != nil {
		return Observation{}, nil, err
	}

	if _, err := tx.Exec(
		sqlUpdateObservation,
		in.SubjectID,
		in.PropertyID,
		polarity,
		nullIfEmpty(resolved.Text),
		nullInt64(resolved.Integer),
		nullBlob(resolved.DateID),
		nullBlob(resolved.NameID),
		nullBlob(resolved.SubjectID),
		nullBlob(resolved.TermID),
		prev.ID,
		citationID,
	); err != nil {
		return Observation{}, nil, err
	}
	if in.Notes != nil {
		if _, err := tx.Exec(sqlDeleteObservationNotes, prev.ID); err != nil {
			return Observation{}, nil, err
		}
		for _, body := range in.Notes {
			body = strings.TrimSpace(body)
			if body == "" {
				continue
			}
			if err := insertObservationNote(tx, prev.ID, body); err != nil {
				return Observation{}, nil, err
			}
		}
	}
	if err := releaseDateValue(tx, prev.ValueDateID, resolved.DateID); err != nil {
		return Observation{}, nil, err
	}
	if err := releaseNameValue(tx, prev.ValueNameID, resolved.NameID); err != nil {
		return Observation{}, nil, err
	}

	obs := Observation{
		ID:         append([]byte(nil), prev.ID...),
		Ref:        prev.Ref,
		CitationID: append([]byte(nil), citationID...),
		SubjectID:  append([]byte(nil), in.SubjectID...),
		PropertyID: append([]byte(nil), in.PropertyID...),
		Polarity:   polarity,
	}
	if resolved.Text != "" {
		obs.ValueText = resolved.Text
		obs.HasText = true
	}
	if resolved.Integer != nil {
		obs.ValueInteger = *resolved.Integer
		obs.HasInteger = true
	}
	obs.ValueDateID = append([]byte(nil), resolved.DateID...)
	obs.ValueNameID = append([]byte(nil), resolved.NameID...)
	obs.ValueSubjectID = append([]byte(nil), resolved.SubjectID...)
	obs.ValueTermID = append([]byte(nil), resolved.TermID...)
	return obs, observationUpdateChange(prev, obs), nil
}

func observationUpdateChange(prev, next Observation) *audit.Change {
	fields := map[string]audit.FieldDiff{}
	if !bytes.Equal(prev.SubjectID, next.SubjectID) {
		fields["subject_id"] = audit.FieldDiff{Old: uuidString(prev.SubjectID), New: uuidString(next.SubjectID)}
	}
	if !bytes.Equal(prev.PropertyID, next.PropertyID) {
		fields["property_id"] = audit.FieldDiff{Old: uuidString(prev.PropertyID), New: uuidString(next.PropertyID)}
	}
	if prev.Polarity != next.Polarity {
		fields["polarity"] = audit.FieldDiff{Old: prev.Polarity, New: next.Polarity}
	}
	if prev.ValueText != next.ValueText {
		fields["value_text"] = audit.FieldDiff{Old: emptyAsNil(prev.ValueText), New: emptyAsNil(next.ValueText)}
	}
	if prev.HasInteger != next.HasInteger || (next.HasInteger && prev.ValueInteger != next.ValueInteger) {
		var old, new any
		if prev.HasInteger {
			old = prev.ValueInteger
		}
		if next.HasInteger {
			new = next.ValueInteger
		}
		fields["value_integer"] = audit.FieldDiff{Old: old, New: new}
	}
	if !bytes.Equal(prev.ValueDateID, next.ValueDateID) {
		fields["value_date_id"] = audit.FieldDiff{Old: uuidString(prev.ValueDateID), New: uuidString(next.ValueDateID)}
	}
	if !bytes.Equal(prev.ValueNameID, next.ValueNameID) {
		fields["value_name_id"] = audit.FieldDiff{Old: uuidString(prev.ValueNameID), New: uuidString(next.ValueNameID)}
	}
	if !bytes.Equal(prev.ValueSubjectID, next.ValueSubjectID) {
		fields["value_subject_id"] = audit.FieldDiff{Old: uuidString(prev.ValueSubjectID), New: uuidString(next.ValueSubjectID)}
	}
	if !bytes.Equal(prev.ValueTermID, next.ValueTermID) {
		fields["value_term_id"] = audit.FieldDiff{Old: uuidString(prev.ValueTermID), New: uuidString(next.ValueTermID)}
	}
	if len(fields) == 0 {
		return nil
	}
	return &audit.Change{
		EntityType: "observation",
		EntityID:   next.ID,
		Action:     audit.ActionUpdate,
		Fields:     fields,
	}
}

func listRowsByCitationTx(tx *sql.Tx, citationID []byte) ([]Observation, error) {
	rows, err := tx.Query(sqlListRowsByCitation, citationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Observation
	for rows.Next() {
		var (
			obs       Observation
			valueText sql.NullString
			valueInt  sql.NullInt64
			dateID    []byte
			nameID    []byte
			subjectID []byte
			termID    []byte
		)
		if err := rows.Scan(
			&obs.ID, &obs.Ref, &obs.CitationID, &obs.SubjectID, &obs.PropertyID, &obs.Polarity,
			&valueText, &valueInt, &dateID, &nameID, &subjectID, &termID,
		); err != nil {
			return nil, err
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
		out = append(out, obs)
	}
	return out, rows.Err()
}

func releaseDateValue(tx *sql.Tx, oldID, newID []byte) error {
	if len(oldID) != 16 || bytes.Equal(oldID, newID) {
		return nil
	}
	var n int
	if err := tx.QueryRow(sqlCountDateRefs, oldID, oldID).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	_, err := tx.Exec(sqlDeleteDateValue, oldID)
	return err
}

func releaseNameValue(tx *sql.Tx, oldID, newID []byte) error {
	if len(oldID) != 16 || bytes.Equal(oldID, newID) {
		return nil
	}
	var n int
	if err := tx.QueryRow(sqlCountNameRefs, oldID).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	_, err := tx.Exec(sqlDeleteNameValue, oldID)
	return err
}

func stringsTrimPolarity(polarity string) string {
	polarity = strings.TrimSpace(polarity)
	if polarity == "" {
		return PolarityPositive
	}
	return polarity
}

func emptyAsNil(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func insertObservationNote(tx *sql.Tx, observationID []byte, body string) error {
	noteID, err := uuid.NewV7()
	if err != nil {
		return err
	}
	_, err = tx.Exec(sqlInsertNote, noteID[:], observationID, body)
	return err
}
