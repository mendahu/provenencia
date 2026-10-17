package observations

import (
	"bytes"
	"database/sql"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/deleteimpact"
	"github.com/mendahu/provenencia/core/database/depictionvalues"
	"github.com/mendahu/provenencia/core/database/namevalues"
)

const (
	sqlUpdateObservation = `UPDATE observations SET
		subject_id = ?, property_id = ?, polarity = ?,
		value_text = ?, value_integer = ?,
		value_date_id = ?, value_name_id = ?, value_subject_id = ?, value_term_id = ?,
		value_depiction_id = ?
		WHERE id = ? AND citation_id = ?`
)

func updateOne(tx *database.Tx, citationID []byte, prev Observation, in Input) (Observation, []rowchange.Change, error) {
	polarity := stringsTrimPolarity(in.Polarity)
	if polarity != PolarityPositive && polarity != PolarityNegative {
		return Observation{}, nil, ErrInvalid
	}
	if len(in.SubjectID) != 16 || len(in.PropertyID) != 16 {
		return Observation{}, nil, ErrInvalid
	}
	if err := requireSubjectSource(tx.Tx, citationID, in.SubjectID); err != nil {
		return Observation{}, nil, err
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
	prop, err := getPropertyTx(tx.Tx, in.PropertyID)
	if err != nil {
		return Observation{}, nil, err
	}
	resolved, err := resolveValue(tx.Tx, citationID, prop.ValueType, in, &prev)
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
		nullBlob(resolved.DepictionID),
		prev.ID,
		citationID,
	); err != nil {
		return Observation{}, nil, err
	}
	var extra []rowchange.Change
	if in.Notes != nil {
		noteChanges, err := replaceNotes(tx.Tx, prev.ID, in.Notes)
		if err != nil {
			return Observation{}, nil, err
		}
		extra = append(extra, noteChanges...)
	}
	for _, rel := range []struct {
		column, entity string
		prevID, newID  []byte
		rowMap         func() (map[string]any, error)
	}{
		{"observations.value_date_id", "date_value", prev.ValueDateID, resolved.DateID, func() (map[string]any, error) {
			v, err := datevalues.LookupTx(tx.Tx, prev.ValueDateID)
			return dateValueMap(prev.ValueDateID, v), err
		}},
		{"observations.value_name_id", "name_value", prev.ValueNameID, resolved.NameID, func() (map[string]any, error) {
			v, err := namevalues.LookupTx(tx.Tx, prev.ValueNameID)
			return nameValueMap(prev.ValueNameID, v), err
		}},
		{"observations.value_depiction_id", "depiction_value", prev.ValueDepictionID, resolved.DepictionID, func() (map[string]any, error) {
			v, err := depictionvalues.LookupTx(tx.Tx, prev.ValueDepictionID)
			return depictionValueMap(prev.ValueDepictionID, v), err
		}},
	} {
		ch, err := releaseReplacedValue(tx, rel.column, rel.entity, rel.prevID, rel.newID, rel.rowMap)
		if err != nil {
			return Observation{}, nil, err
		}
		if ch != nil {
			extra = append(extra, *ch)
		}
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
	obs.ValueDepictionID = append([]byte(nil), resolved.DepictionID...)
	changes := extra
	if resolved.DateChange != nil {
		changes = append(changes, *resolved.DateChange)
	}
	if resolved.NameChange != nil {
		changes = append(changes, *resolved.NameChange)
	}
	changes = append(changes, resolved.DepictionChanges...)
	if ch := observationUpdateChange(prev, obs); ch != nil {
		changes = append([]rowchange.Change{*ch}, changes...)
	}
	return obs, changes, nil
}

func observationUpdateChange(prev, next Observation) *rowchange.Change {
	fields := map[string]rowchange.FieldDiff{}
	if !bytes.Equal(prev.SubjectID, next.SubjectID) {
		fields["subject_id"] = rowchange.FieldDiff{Old: uuidString(prev.SubjectID), New: uuidString(next.SubjectID)}
	}
	if !bytes.Equal(prev.PropertyID, next.PropertyID) {
		fields["property_id"] = rowchange.FieldDiff{Old: uuidString(prev.PropertyID), New: uuidString(next.PropertyID)}
	}
	if prev.Polarity != next.Polarity {
		fields["polarity"] = rowchange.FieldDiff{Old: prev.Polarity, New: next.Polarity}
	}
	if prev.ValueText != next.ValueText {
		fields["value_text"] = rowchange.FieldDiff{Old: emptyAsNil(prev.ValueText), New: emptyAsNil(next.ValueText)}
	}
	if prev.HasInteger != next.HasInteger || (next.HasInteger && prev.ValueInteger != next.ValueInteger) {
		var old, new any
		if prev.HasInteger {
			old = prev.ValueInteger
		}
		if next.HasInteger {
			new = next.ValueInteger
		}
		fields["value_integer"] = rowchange.FieldDiff{Old: old, New: new}
	}
	if !bytes.Equal(prev.ValueDateID, next.ValueDateID) {
		fields["value_date_id"] = rowchange.FieldDiff{Old: uuidString(prev.ValueDateID), New: uuidString(next.ValueDateID)}
	}
	if !bytes.Equal(prev.ValueNameID, next.ValueNameID) {
		fields["value_name_id"] = rowchange.FieldDiff{Old: uuidString(prev.ValueNameID), New: uuidString(next.ValueNameID)}
	}
	if !bytes.Equal(prev.ValueSubjectID, next.ValueSubjectID) {
		fields["value_subject_id"] = rowchange.FieldDiff{Old: uuidString(prev.ValueSubjectID), New: uuidString(next.ValueSubjectID)}
	}
	if !bytes.Equal(prev.ValueTermID, next.ValueTermID) {
		fields["value_term_id"] = rowchange.FieldDiff{Old: uuidString(prev.ValueTermID), New: uuidString(next.ValueTermID)}
	}
	if !bytes.Equal(prev.ValueDepictionID, next.ValueDepictionID) {
		fields["value_depiction_id"] = rowchange.FieldDiff{Old: uuidString(prev.ValueDepictionID), New: uuidString(next.ValueDepictionID)}
	}
	if len(fields) == 0 {
		return nil
	}
	return &rowchange.Change{
		EntityType: "observation",
		EntityID:   next.ID,
		Action:     rowchange.ActionUpdate,
		Fields:     fields,
	}
}

// releaseReplacedValue releases the value row an update replaced: the old
// child of an owned FK, with its own owned children. It does nothing when
// prevID is empty or still the Observation's value. The delete change is built
// from rowMap before the release. Call it after the Observation holds the new
// value.
func releaseReplacedValue(tx *database.Tx, column, entity string, prevID, newID []byte, rowMap func() (map[string]any, error)) (*rowchange.Change, error) {
	if len(prevID) != 16 || bytes.Equal(prevID, newID) {
		return nil, nil
	}
	row, err := rowMap()
	if err != nil {
		return nil, err
	}
	if err := deleteimpact.ReleaseReplaced(tx, column, prevID); err != nil {
		return nil, err
	}
	return &rowchange.Change{
		EntityType: entity,
		EntityID:   append([]byte(nil), prevID...),
		Action:     rowchange.ActionDelete,
		Fields:     rowchange.DeletedRow(row),
	}, nil
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
