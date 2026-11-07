// Package observationrow is one observations row as the database holds it: the
// column list, the scan, and the audit fields. Writing an Observation and
// erasing one with a Connection facet both read and audit the whole row, so
// they share this instead of each keeping a column list.
package observationrow

import (
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/scalarcols"
	"github.com/mendahu/provenencia/core/propertymeta"
)

// Columns is the SELECT list of Scan, in scan order.
const Columns = `id, ref, citation_id, subject_id, property_id, polarity,
		value_text, value_integer, value_boolean, value_date_id, value_name_id, value_subject_id, value_term_id,
		value_depiction_id`

// Row is one observations row.
type Row struct {
	ID         []byte
	Ref        string
	CitationID []byte
	SubjectID  []byte
	PropertyID []byte
	Polarity   string
	propertymeta.Scalar
	ValueDateID      []byte
	ValueNameID      []byte
	ValueSubjectID   []byte
	ValueTermID      []byte
	ValueDepictionID []byte
}

// Scanner is *sql.Row or *sql.Rows.
type Scanner interface {
	Scan(dest ...any) error
}

// Scan reads one row selected with Columns. Its byte slices are copies.
func Scan(s Scanner) (Row, error) {
	var (
		r                                         Row
		scalar                                    scalarcols.Columns
		dateID, nameID, subjectID, termID, depict []byte
	)
	if err := s.Scan(
		&r.ID, &r.Ref, &r.CitationID, &r.SubjectID, &r.PropertyID, &r.Polarity,
		&scalar.Text, &scalar.Integer, &scalar.Boolean, &dateID, &nameID, &subjectID, &termID, &depict,
	); err != nil {
		return Row{}, err
	}
	r.Scalar = scalar.Scalar()
	r.ValueDateID = append([]byte(nil), dateID...)
	r.ValueNameID = append([]byte(nil), nameID...)
	r.ValueSubjectID = append([]byte(nil), subjectID...)
	r.ValueTermID = append([]byte(nil), termID...)
	r.ValueDepictionID = append([]byte(nil), depict...)
	return r, nil
}

// AuditFields is the row as an audit change records it, one key per column.
func (r Row) AuditFields() map[string]any {
	var integer, boolean any
	if r.HasInteger {
		integer = r.Integer
	}
	if r.HasBoolean {
		boolean = r.Boolean
	}
	return map[string]any{
		"id":                 database.UUIDJSON(r.ID),
		"ref":                r.Ref,
		"citation_id":        database.UUIDJSON(r.CitationID),
		"subject_id":         database.UUIDJSON(r.SubjectID),
		"property_id":        database.UUIDJSON(r.PropertyID),
		"polarity":           r.Polarity,
		"value_text":         database.NullString(r.Text),
		"value_integer":      integer,
		"value_boolean":      boolean,
		"value_date_id":      database.UUIDJSON(r.ValueDateID),
		"value_name_id":      database.UUIDJSON(r.ValueNameID),
		"value_subject_id":   database.UUIDJSON(r.ValueSubjectID),
		"value_term_id":      database.UUIDJSON(r.ValueTermID),
		"value_depiction_id": database.UUIDJSON(r.ValueDepictionID),
	}
}
