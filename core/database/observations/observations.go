// Package observations stores Interpretation Observation rows and notes.
package observations

import (
	"database/sql"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/audit"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/database/project"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/ref"
)

var ErrInvalid = apperr.New(apperr.CodeObservationsInvalid, apperr.KindUser)

const (
	PolarityPositive = "positive"
	PolarityNegative = "negative"

	sqlInsert = `INSERT INTO observations (
		id, ref, citation_id, subject_id, property_id, polarity,
		value_text, value_integer, value_date_id, value_name_id, value_subject_id, value_term_id
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	sqlInsertNote = `INSERT INTO observation_notes (id, observation_id, body) VALUES (?, ?, ?)`

	sqlCitationExists = `SELECT 1 FROM citations WHERE id = ?`
	sqlSubjectType    = `SELECT subject_type_id FROM subjects WHERE id = ?`
	sqlSubjectExists  = `SELECT 1 FROM subjects WHERE id = ?`
	sqlBindingExists  = `SELECT 1 FROM subject_type_fields WHERE subject_type_id = ? AND property_id = ?`
	sqlTermOnProperty = `SELECT 1 FROM property_terms WHERE id = ? AND property_id = ?`
	sqlPropertyGet    = `SELECT id, key, origin, label, COALESCE(description, ''), value_type
		FROM properties WHERE id = ?`

	// List SELECTs denormalize structured value display into value_text (term
	// label, name form, subject label, compact date) so graph cards / clients
	// that only read value_text stay correct for every Property value_type.
	sqlListSelect = `SELECT o.id, o.ref, o.citation_id, o.subject_id, o.property_id, o.polarity,
		o.value_text, o.value_integer, o.value_date_id, o.value_name_id, o.value_subject_id, o.value_term_id,
		p.key, p.label, p.value_type,
		COALESCE(pt.label, ''),
		COALESCE(nv.form, ''),
		COALESCE(vs.label, ''),
		dv.kind, dv.qualifier, dv.calendar,
		dv.start_year, dv.start_month, dv.start_day,
		dv.start_hour, dv.start_minute, dv.start_second, dv.start_millisecond, dv.start_tz,
		dv.end_year, dv.end_month, dv.end_day,
		dv.end_hour, dv.end_minute, dv.end_second, dv.end_millisecond, dv.end_tz,
		dv.phrase
		FROM observations o
		JOIN properties p ON p.id = o.property_id
		LEFT JOIN property_terms pt ON pt.id = o.value_term_id
		LEFT JOIN name_values nv ON nv.id = o.value_name_id
		LEFT JOIN subjects vs ON vs.id = o.value_subject_id
		LEFT JOIN date_values dv ON dv.id = o.value_date_id`

	sqlListBySource = sqlListSelect + `
		JOIN subjects s ON s.id = o.subject_id
		WHERE s.source_id = ?
		ORDER BY o.ref COLLATE NOCASE`

	sqlListBySubject = sqlListSelect + `
		WHERE o.subject_id = ?
		ORDER BY o.ref COLLATE NOCASE`

	maxRefRetries = 8
)

// Observation is one observations row.
type Observation struct {
	ID             []byte
	Ref            string
	CitationID     []byte
	SubjectID      []byte
	PropertyID     []byte
	Polarity       string
	ValueText      string
	HasText        bool
	ValueInteger   int64
	HasInteger     bool
	ValueDateID    []byte
	ValueNameID    []byte
	ValueSubjectID []byte
	ValueTermID    []byte
}

// Listed is an Observation plus Property summary for graph / card rows.
type Listed struct {
	Observation
	PropertyKey       string
	PropertyLabel     string
	PropertyValueType string
	// ValueTermLabel is the property_terms.label for value_term_id (empty when unset).
	ValueTermLabel string
	// ValueNameForm is name_values.form for value_name_id (empty when unset).
	ValueNameForm string
	// ValueSubjectLabel is subjects.label for value_subject_id (empty when unset).
	ValueSubjectLabel string
	// Date is the joined date_values row when value_date_id is set.
	Date *datevalues.Value
}

// Input is one Observation draft for insert (create or append).
type Input struct {
	SubjectID      []byte
	PropertyID     []byte
	Polarity       string // "" → positive
	ValueText      string
	HasText        bool
	ValueInteger   int64
	HasInteger     bool
	Date           *datevalues.Value // when set, InsertTx inside the caller's tx
	ValueDateID    []byte            // existing date_values.id when Date is nil
	Name           *namevalues.Value
	ValueNameID    []byte
	ValueSubjectID []byte
	ValueTermID    []byte
	Notes          []string
}

// AddToCitation appends ≥1 Observations to an existing Citation in one audited transaction.
func AddToCitation(c *database.Catalog, userID, citationID []byte, inputs []Input) ([]Observation, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	if len(citationID) != 16 || len(inputs) == 0 {
		return nil, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return nil, err
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var one int
	if err := tx.QueryRow(sqlCitationExists, citationID).Scan(&one); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrInvalid
		}
		return nil, err
	}

	out := make([]Observation, 0, len(inputs))
	changes := make([]audit.Change, 0, len(inputs))
	for _, in := range inputs {
		obs, ch, err := insertOne(tx, citationID, in)
		if err != nil {
			return nil, err
		}
		out = append(out, obs)
		changes = append(changes, ch)
	}
	if _, err := audit.Record(tx, audit.Revision{
		UserID:     userID,
		ActionType: "add_observations",
		CreatedAt:  project.NowUTC(),
		Changes:    changes,
	}); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return out, nil
}

// InsertManyTx inserts Observations for citationID inside an existing transaction.
// Returns rows and audit Changes (caller records the revision).
func InsertManyTx(tx *sql.Tx, citationID []byte, inputs []Input) ([]Observation, []audit.Change, error) {
	if tx == nil || len(citationID) != 16 || len(inputs) == 0 {
		return nil, nil, ErrInvalid
	}
	out := make([]Observation, 0, len(inputs))
	changes := make([]audit.Change, 0, len(inputs))
	for _, in := range inputs {
		obs, ch, err := insertOne(tx, citationID, in)
		if err != nil {
			return nil, nil, err
		}
		out = append(out, obs)
		changes = append(changes, ch)
	}
	return out, changes, nil
}

func insertOne(tx *sql.Tx, citationID []byte, in Input) (Observation, audit.Change, error) {
	polarity := strings.TrimSpace(in.Polarity)
	if polarity == "" {
		polarity = PolarityPositive
	}
	if polarity != PolarityPositive && polarity != PolarityNegative {
		return Observation{}, audit.Change{}, ErrInvalid
	}
	if len(in.SubjectID) != 16 || len(in.PropertyID) != 16 {
		return Observation{}, audit.Change{}, ErrInvalid
	}

	var subjectTypeID []byte
	if err := tx.QueryRow(sqlSubjectType, in.SubjectID).Scan(&subjectTypeID); err != nil {
		if err == sql.ErrNoRows {
			return Observation{}, audit.Change{}, ErrInvalid
		}
		return Observation{}, audit.Change{}, err
	}
	var one int
	if err := tx.QueryRow(sqlBindingExists, subjectTypeID, in.PropertyID).Scan(&one); err != nil {
		if err == sql.ErrNoRows {
			return Observation{}, audit.Change{}, ErrInvalid
		}
		return Observation{}, audit.Change{}, err
	}

	prop, err := getPropertyTx(tx, in.PropertyID)
	if err != nil {
		return Observation{}, audit.Change{}, err
	}

	valueText, valueInt, dateID, nameID, subjectID, termID, err := resolveValue(tx, prop.ValueType, in)
	if err != nil {
		return Observation{}, audit.Change{}, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return Observation{}, audit.Change{}, err
	}
	idBytes := id[:]

	var obsRef string
	for attempt := 0; attempt < maxRefRetries; attempt++ {
		obsRef, err = ref.Mint(ref.PrefixObservation)
		if err != nil {
			return Observation{}, audit.Change{}, err
		}
		_, err = tx.Exec(
			sqlInsert,
			idBytes,
			obsRef,
			citationID,
			in.SubjectID,
			in.PropertyID,
			polarity,
			nullIfEmpty(valueText),
			nullInt64(valueInt),
			nullBlob(dateID),
			nullBlob(nameID),
			nullBlob(subjectID),
			nullBlob(termID),
		)
		if err == nil {
			break
		}
		if !database.IsUniqueConflict(err) {
			return Observation{}, audit.Change{}, ErrInvalid
		}
	}
	if err != nil {
		return Observation{}, audit.Change{}, ErrInvalid
	}

	for _, body := range in.Notes {
		body = strings.TrimSpace(body)
		if body == "" {
			continue
		}
		noteID, err := uuid.NewV7()
		if err != nil {
			return Observation{}, audit.Change{}, err
		}
		if _, err := tx.Exec(sqlInsertNote, noteID[:], idBytes, body); err != nil {
			return Observation{}, audit.Change{}, err
		}
	}

	obs := Observation{
		ID:         append([]byte(nil), idBytes...),
		Ref:        obsRef,
		CitationID: append([]byte(nil), citationID...),
		SubjectID:  append([]byte(nil), in.SubjectID...),
		PropertyID: append([]byte(nil), in.PropertyID...),
		Polarity:   polarity,
	}
	if valueText != "" {
		obs.ValueText = valueText
		obs.HasText = true
	}
	if valueInt != nil {
		obs.ValueInteger = *valueInt
		obs.HasInteger = true
	}
	obs.ValueDateID = append([]byte(nil), dateID...)
	obs.ValueNameID = append([]byte(nil), nameID...)
	obs.ValueSubjectID = append([]byte(nil), subjectID...)
	obs.ValueTermID = append([]byte(nil), termID...)

	fields := map[string]audit.FieldDiff{
		"id":          {Old: nil, New: id.String()},
		"ref":         {Old: nil, New: obsRef},
		"citation_id": {Old: nil, New: uuidString(citationID)},
		"subject_id":  {Old: nil, New: uuidString(in.SubjectID)},
		"property_id": {Old: nil, New: uuidString(in.PropertyID)},
		"polarity":    {Old: nil, New: polarity},
	}
	return obs, audit.Change{
		EntityType: "observation",
		EntityID:   idBytes,
		Action:     audit.ActionCreate,
		Fields:     fields,
	}, nil
}

func resolveValue(tx *sql.Tx, valueType string, in Input) (
	text string, integer *int64, dateID, nameID, subjectID, termID []byte, err error,
) {
	switch valueType {
	case properties.ValueTypeText:
		text = strings.TrimSpace(in.ValueText)
		if !in.HasText || text == "" || hasOtherScalars(in, "text") {
			return "", nil, nil, nil, nil, nil, ErrInvalid
		}
		return text, nil, nil, nil, nil, nil, nil
	case properties.ValueTypeInteger:
		if !in.HasInteger || hasOtherScalars(in, "integer") {
			return "", nil, nil, nil, nil, nil, ErrInvalid
		}
		v := in.ValueInteger
		return "", &v, nil, nil, nil, nil, nil
	case properties.ValueTypeDate:
		if hasOtherScalars(in, "date") {
			return "", nil, nil, nil, nil, nil, ErrInvalid
		}
		if in.Date != nil {
			dateID, err = datevalues.InsertTx(tx, *in.Date)
			return "", nil, dateID, nil, nil, nil, err
		}
		if len(in.ValueDateID) != 16 {
			return "", nil, nil, nil, nil, nil, ErrInvalid
		}
		return "", nil, in.ValueDateID, nil, nil, nil, nil
	case properties.ValueTypeName:
		if hasOtherScalars(in, "name") {
			return "", nil, nil, nil, nil, nil, ErrInvalid
		}
		if in.Name != nil {
			nameID, err = namevalues.InsertTx(tx, *in.Name)
			return "", nil, nil, nameID, nil, nil, err
		}
		if len(in.ValueNameID) != 16 {
			return "", nil, nil, nil, nil, nil, ErrInvalid
		}
		return "", nil, nil, in.ValueNameID, nil, nil, nil
	case properties.ValueTypeSubject:
		if len(in.ValueSubjectID) != 16 || hasOtherScalars(in, "subject") {
			return "", nil, nil, nil, nil, nil, ErrInvalid
		}
		var one int
		if err := tx.QueryRow(sqlSubjectExists, in.ValueSubjectID).Scan(&one); err != nil {
			if err == sql.ErrNoRows {
				return "", nil, nil, nil, nil, nil, ErrInvalid
			}
			return "", nil, nil, nil, nil, nil, err
		}
		return "", nil, nil, nil, in.ValueSubjectID, nil, nil
	case properties.ValueTypeTerm:
		if len(in.ValueTermID) != 16 || hasOtherScalars(in, "term") {
			return "", nil, nil, nil, nil, nil, ErrInvalid
		}
		var one int
		if err := tx.QueryRow(sqlTermOnProperty, in.ValueTermID, in.PropertyID).Scan(&one); err != nil {
			if err == sql.ErrNoRows {
				return "", nil, nil, nil, nil, nil, ErrInvalid
			}
			return "", nil, nil, nil, nil, nil, err
		}
		return "", nil, nil, nil, nil, in.ValueTermID, nil
	default:
		return "", nil, nil, nil, nil, nil, ErrInvalid
	}
}

func hasOtherScalars(in Input, keep string) bool {
	if keep != "text" && in.HasText {
		return true
	}
	if keep != "integer" && in.HasInteger {
		return true
	}
	if keep != "date" && (in.Date != nil || len(in.ValueDateID) > 0) {
		return true
	}
	if keep != "name" && (in.Name != nil || len(in.ValueNameID) > 0) {
		return true
	}
	if keep != "subject" && len(in.ValueSubjectID) > 0 {
		return true
	}
	if keep != "term" && len(in.ValueTermID) > 0 {
		return true
	}
	return false
}

// ListBySource returns Observations whose subject is home to sourceID.
func ListBySource(c *database.Catalog, sourceID []byte) ([]Listed, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	if len(sourceID) != 16 {
		return nil, ErrInvalid
	}
	rows, err := db.Query(sqlListBySource, sourceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanListed(rows)
}

// ListBySubject returns Observations for one subject.
func ListBySubject(c *database.Catalog, subjectID []byte) ([]Listed, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	if len(subjectID) != 16 {
		return nil, ErrInvalid
	}
	rows, err := db.Query(sqlListBySubject, subjectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanListed(rows)
}

func scanListed(rows *sql.Rows) ([]Listed, error) {
	var out []Listed
	for rows.Next() {
		var (
			l                                                          Listed
			valueText                                                  sql.NullString
			valueInt                                                   sql.NullInt64
			dateID, nameID, subjectID, termID                          []byte
			termLabel, nameForm, subjectLabel                          string
			kind, qual, cal, phrase, startTZ, endTZ                    sql.NullString
			startY, startM, startD, startH, startMin, startS, startMs  sql.NullInt64
			endY, endM, endD, endH, endMin, endS, endMs                sql.NullInt64
		)
		if err := rows.Scan(
			&l.ID, &l.Ref, &l.CitationID, &l.SubjectID, &l.PropertyID, &l.Polarity,
			&valueText, &valueInt, &dateID, &nameID, &subjectID, &termID,
			&l.PropertyKey, &l.PropertyLabel, &l.PropertyValueType,
			&termLabel, &nameForm, &subjectLabel,
			&kind, &qual, &cal,
			&startY, &startM, &startD,
			&startH, &startMin, &startS, &startMs, &startTZ,
			&endY, &endM, &endD,
			&endH, &endMin, &endS, &endMs, &endTZ,
			&phrase,
		); err != nil {
			return nil, err
		}
		if valueText.Valid {
			l.ValueText = valueText.String
			l.HasText = true
		}
		if valueInt.Valid {
			l.ValueInteger = valueInt.Int64
			l.HasInteger = true
		}
		l.ValueDateID = append([]byte(nil), dateID...)
		l.ValueNameID = append([]byte(nil), nameID...)
		l.ValueSubjectID = append([]byte(nil), subjectID...)
		l.ValueTermID = append([]byte(nil), termID...)
		l.ValueTermLabel = termLabel
		l.ValueNameForm = nameForm
		l.ValueSubjectLabel = subjectLabel
		if kind.Valid && kind.String != "" {
			dv := datevalues.Value{
				ID:               append([]byte(nil), dateID...),
				Kind:             kind.String,
				Qualifier:        qual.String,
				Calendar:         cal.String,
				StartYear:        nullIntPtr(startY),
				StartMonth:       nullIntPtr(startM),
				StartDay:         nullIntPtr(startD),
				StartHour:        nullIntPtr(startH),
				StartMinute:      nullIntPtr(startMin),
				StartSecond:      nullIntPtr(startS),
				StartMillisecond: nullIntPtr(startMs),
				StartTZ:          startTZ.String,
				EndYear:          nullIntPtr(endY),
				EndMonth:         nullIntPtr(endM),
				EndDay:           nullIntPtr(endD),
				EndHour:          nullIntPtr(endH),
				EndMinute:        nullIntPtr(endMin),
				EndSecond:        nullIntPtr(endS),
				EndMillisecond:   nullIntPtr(endMs),
				EndTZ:            endTZ.String,
				Phrase:           phrase.String,
			}
			l.Date = &dv
		}
		fillListedDisplayText(&l)
		out = append(out, l)
	}
	return out, rows.Err()
}

// fillListedDisplayText denormalizes structured value types into ValueText so
// card rows that only read value_text work for term / name / date / subject.
func fillListedDisplayText(l *Listed) {
	if l.HasText {
		return
	}
	switch {
	case l.ValueTermLabel != "":
		l.ValueText = l.ValueTermLabel
	case l.ValueNameForm != "":
		l.ValueText = l.ValueNameForm
	case l.ValueSubjectLabel != "":
		l.ValueText = l.ValueSubjectLabel
	case l.Date != nil:
		l.ValueText = datevalues.CompactDisplay(*l.Date)
	default:
		return
	}
	if l.ValueText != "" {
		l.HasText = true
	}
}

func nullIntPtr(n sql.NullInt64) *int {
	if !n.Valid {
		return nil
	}
	v := int(n.Int64)
	return &v
}

func getPropertyTx(tx *sql.Tx, id []byte) (properties.Property, error) {
	var p properties.Property
	err := tx.QueryRow(sqlPropertyGet, id).Scan(
		&p.ID, &p.Key, &p.Origin, &p.Label, &p.Description, &p.ValueType,
	)
	if err == sql.ErrNoRows {
		return properties.Property{}, ErrInvalid
	}
	return p, err
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullInt64(p *int64) any {
	if p == nil {
		return nil
	}
	return *p
}

func nullBlob(b []byte) any {
	if len(b) == 0 {
		return nil
	}
	return b
}

func uuidString(id []byte) string {
	u, err := uuid.FromBytes(id)
	if err != nil {
		return ""
	}
	return u.String()
}
