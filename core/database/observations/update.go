package observations

import (
	"bytes"
	"database/sql"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database/audit"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/namevalues"
)

const (
	sqlUpdateObservation = `UPDATE observations SET
		subject_id = ?, property_id = ?, polarity = ?,
		value_text = ?, value_integer = ?,
		value_date_id = ?, value_name_id = ?, value_subject_id = ?, value_term_id = ?
		WHERE id = ? AND citation_id = ?`

	sqlDeleteObservation = `DELETE FROM observations WHERE id = ? AND citation_id = ?`

	sqlDeleteObservationNotes = `DELETE FROM observation_notes WHERE observation_id = ?`

	sqlCountDateRefs = `SELECT COUNT(*) FROM observations WHERE value_date_id = ?`

	sqlCountNameRefs = `SELECT COUNT(*) FROM observations WHERE value_name_id = ?`

	sqlDeleteDateValue = `DELETE FROM date_values WHERE id = ?`

	sqlDeleteNameValue = `DELETE FROM name_values WHERE id = ?`
)

func updateOne(tx *sql.Tx, citationID []byte, prev Observation, in Input) (Observation, []audit.Change, error) {
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
	var extra []audit.Change
	if in.Notes != nil {
		noteChanges, err := replaceNotes(tx, prev.ID, in.Notes)
		if err != nil {
			return Observation{}, nil, err
		}
		extra = append(extra, noteChanges...)
	}
	if deleted, err := releaseDateValue(tx, prev.ValueDateID, resolved.DateID); err != nil {
		return Observation{}, nil, err
	} else if deleted != nil {
		extra = append(extra, audit.Change{
			EntityType: "date_value",
			EntityID:   append([]byte(nil), prev.ValueDateID...),
			Action:     audit.ActionDelete,
			Fields:     audit.DeletedRow(deleted),
		})
	}
	if deleted, err := releaseNameValue(tx, prev.ValueNameID, resolved.NameID); err != nil {
		return Observation{}, nil, err
	} else if deleted != nil {
		extra = append(extra, audit.Change{
			EntityType: "name_value",
			EntityID:   append([]byte(nil), prev.ValueNameID...),
			Action:     audit.ActionDelete,
			Fields:     audit.DeletedRow(deleted),
		})
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
	changes := extra
	if resolved.DateChange != nil {
		changes = append(changes, *resolved.DateChange)
	}
	if resolved.NameChange != nil {
		changes = append(changes, *resolved.NameChange)
	}
	if ch := observationUpdateChange(prev, obs); ch != nil {
		changes = append([]audit.Change{*ch}, changes...)
	}
	return obs, changes, nil
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

func releaseDateValue(tx *sql.Tx, oldID, newID []byte) (map[string]any, error) {
	if len(oldID) != 16 || bytes.Equal(oldID, newID) {
		return nil, nil
	}
	var n int
	if err := tx.QueryRow(sqlCountDateRefs, oldID).Scan(&n); err != nil {
		return nil, err
	}
	if n > 0 {
		return nil, nil
	}
	prev, err := datevalues.LookupTx(tx, oldID)
	if err != nil {
		return nil, err
	}
	if _, err := tx.Exec(sqlDeleteDateValue, oldID); err != nil {
		return nil, err
	}
	return dateValueMap(oldID, prev), nil
}

func releaseNameValue(tx *sql.Tx, oldID, newID []byte) (map[string]any, error) {
	if len(oldID) != 16 || bytes.Equal(oldID, newID) {
		return nil, nil
	}
	var n int
	if err := tx.QueryRow(sqlCountNameRefs, oldID).Scan(&n); err != nil {
		return nil, err
	}
	if n > 0 {
		return nil, nil
	}
	prev, err := namevalues.LookupTx(tx, oldID)
	if err != nil {
		return nil, err
	}
	if _, err := tx.Exec(sqlDeleteNameValue, oldID); err != nil {
		return nil, err
	}
	return nameValueMap(oldID, prev), nil
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
