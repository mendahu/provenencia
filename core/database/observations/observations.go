// Package observations stores Interpretation Observation rows and notes.
package observations

import (
	"bytes"
	"database/sql"
	"errors"
	"strings"

	"github.com/mendahu/provenencia/core/database/rowchange"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/depictionvalues"
	"github.com/mendahu/provenencia/core/database/evidencegraphs"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/depictionvalue"
	"github.com/mendahu/provenencia/core/namevalue"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/vocabkey"
)

var (
	ErrInvalid    = apperr.New(apperr.CodeObservationsInvalid, apperr.KindUser)
	ErrEdgeLocked = apperr.New(apperr.CodeObservationsEdgeLocked, apperr.KindConflict)
	ErrInUse      = apperr.New(apperr.CodeObservationsInUse, apperr.KindConflict)
	ErrOtherGraph = apperr.New(apperr.CodeObservationsOtherGraph, apperr.KindUser)
)

// InsertOptions controls engine-locked edge Observation writes.
type InsertOptions struct {
	AllowEdgeRows bool
}

const (
	PolarityPositive = "positive"
	PolarityNegative = "negative"

	sqlInsert = `INSERT INTO observations (
		id, ref, citation_id, subject_id, property_id, polarity,
		value_text, value_integer, value_date_id, value_name_id, value_subject_id, value_term_id,
		value_depiction_id
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	sqlInsertNote = `INSERT INTO observation_notes (id, observation_id, body) VALUES (?, ?, ?)`

	sqlCitationExists = `SELECT 1 FROM citations WHERE id = ?`
	sqlSubjectType    = `SELECT subject_type_id FROM subjects WHERE id = ?`
	sqlSubjectTypeKey = `SELECT t.key, t.origin FROM subjects s
		JOIN subject_types t ON t.id = s.subject_type_id
		WHERE s.id = ?`
	sqlGetRow = `SELECT id, ref, citation_id, subject_id, property_id, polarity,
		value_text, value_integer, value_date_id, value_name_id, value_subject_id, value_term_id,
		value_depiction_id
		FROM observations WHERE id = ?`
	sqlListNotesByObservation = `SELECT id, body FROM observation_notes WHERE observation_id = ? ORDER BY rowid`
	sqlDeleteNote             = `DELETE FROM observation_notes WHERE id = ?`
	sqlDeleteObservationByID  = `DELETE FROM observations WHERE id = ?`
	sqlCitationGraph          = `SELECT graph_id FROM citations WHERE id = ?`
	sqlBindingExists          = `SELECT 1 FROM subject_type_properties WHERE subject_type_id = ? AND property_id = ?`
	sqlTermOnProperty         = `SELECT 1 FROM property_terms WHERE id = ? AND property_id = ?`
	sqlPropertyGet            = `SELECT id, key, origin, label, COALESCE(description, ''), value_type
		FROM properties WHERE id = ?`

	// List SELECTs denormalize structured value display into value_text (term
	// label, name form, subject label, compact date) so graph cards / clients
	// that only read value_text stay correct for every Property value_type.
	sqlListSelect = `SELECT o.id, o.ref, o.citation_id, o.subject_id, o.property_id, o.polarity,
		o.value_text, o.value_integer, o.value_date_id, o.value_name_id, o.value_subject_id, o.value_term_id,
		o.value_depiction_id,
		p.key, p.label, p.value_type,
		COALESCE(pt.label, ''),
		COALESCE(pt.key, ''),
		COALESCE(pt.origin, ''),
		COALESCE(nv.form, ''),
		COALESCE(vs.label, ''),
		dv.kind, dv.qualifier, dv.calendar,
		dv.start_year, dv.start_month, dv.start_day,
		dv.start_hour, dv.start_minute, dv.start_second, dv.start_millisecond, dv.start_tz,
		dv.end_year, dv.end_month, dv.end_day,
		dv.end_hour, dv.end_minute, dv.end_second, dv.end_millisecond, dv.end_tz,
		dv.phrase,
		dpv.region_json, dpv.image_file_id, dc.locator_json, dpf.media_type
		FROM observations o
		JOIN properties p ON p.id = o.property_id
		LEFT JOIN property_terms pt ON pt.id = o.value_term_id
		LEFT JOIN name_values nv ON nv.id = o.value_name_id
		LEFT JOIN subjects vs ON vs.id = o.value_subject_id
		LEFT JOIN date_values dv ON dv.id = o.value_date_id
		LEFT JOIN depiction_values dpv ON dpv.id = o.value_depiction_id
		LEFT JOIN citations dc ON dc.id = o.citation_id
		LEFT JOIN files dpf ON dpf.id = dpv.image_file_id`

	sqlListBySubject = sqlListSelect + `
		WHERE o.subject_id = ?
		ORDER BY o.ref COLLATE NOCASE`

	sqlListByCitation = sqlListSelect + `
		WHERE o.citation_id = ?
		ORDER BY o.ref COLLATE NOCASE`

	sqlCountBySubjectOrValue = `SELECT COUNT(*) FROM observations
		WHERE subject_id = ? OR value_subject_id = ?`
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
	// ValueDepictionID is depiction_values.id for a depiction Property.
	ValueDepictionID []byte
}

// Listed is an Observation plus Property summary for graph / card rows.
type Listed struct {
	Observation
	PropertyKey       string
	PropertyLabel     string
	PropertyValueType string
	// ValueTermLabel is the property_terms.label for value_term_id (empty when unset).
	ValueTermLabel string
	// ValueTermKey is the property_terms.key for value_term_id (empty when unset).
	ValueTermKey string
	// ValueTermOrigin is the property_terms.origin for value_term_id (empty when unset).
	ValueTermOrigin string
	// ValueNameForm is name_values.form for value_name_id (empty when unset).
	ValueNameForm string
	// Name is the joined name_values row plus parts when value_name_id is set.
	Name *namevalue.Value
	// ValueSubjectLabel is subjects.label for value_subject_id (empty when unset).
	ValueSubjectLabel string
	// Date is the joined date_values row when value_date_id is set.
	Date *datevalue.Value
	// Depiction is the joined depiction_values row when value_depiction_id is set.
	// An empty RegionJSON means the stored region is not a valid region of the
	// Citation's page; an empty ImageFileID means the stored image is not a
	// Go-stored PNG. Every later reader that returns a depiction (the source
	// graph's hydration) goes through depictionvalue.Readable.
	Depiction *depictionvalue.Value
}

// Input is one Observation write. ID empty inserts. ID set updates that row.
type Input struct {
	ID             []byte
	SubjectID      []byte
	PropertyID     []byte
	Polarity       string // "" → positive
	ValueText      string
	HasText        bool
	ValueInteger   int64
	HasInteger     bool
	Date           *datevalue.Value // when set, InsertTx inside the caller's tx
	ValueDateID    []byte           // existing date_values.id when Date is nil
	Name           *namevalue.Value
	ValueNameID    []byte
	ValueSubjectID []byte
	ValueTermID    []byte
	Depiction      *DepictionInput // staged by ingest.StageDepictions before the write
	// ValueDepictionID is the Observation's own depiction_values.id, kept on an
	// update that sets no Depiction. A value row has one owner.
	ValueDepictionID []byte
	Notes            []string
}

// AddToCitation appends ≥1 Observations to an existing Citation on tx.
// The caller records the returned changes.
func AddToCitation(tx *database.Tx, userID, citationID []byte, inputs []Input) ([]Observation, []rowchange.Change, error) {
	if tx == nil || len(citationID) != 16 || len(inputs) == 0 {
		return nil, nil, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return nil, nil, err
	}

	var one int
	if err := tx.QueryRow(sqlCitationExists, citationID).Scan(&one); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, ErrInvalid
		}
		return nil, nil, err
	}
	return InsertManyTx(tx.Tx, citationID, inputs, InsertOptions{AllowEdgeRows: false})
}

// InsertManyTx inserts Observations for citationID inside an existing transaction.
// Returns rows and changes. The caller records the revision and recomputes.
func InsertManyTx(tx *sql.Tx, citationID []byte, inputs []Input, opts InsertOptions) ([]Observation, []rowchange.Change, error) {
	if tx == nil || len(citationID) != 16 || len(inputs) == 0 {
		return nil, nil, ErrInvalid
	}
	out := make([]Observation, 0, len(inputs))
	changes := make([]rowchange.Change, 0, len(inputs))
	for _, in := range inputs {
		obs, chs, err := insertOne(tx, citationID, in, opts)
		if err != nil {
			return nil, nil, err
		}
		out = append(out, obs)
		changes = append(changes, chs...)
	}
	return out, changes, nil
}

type resolvedValue struct {
	Text              string
	Integer           *int64
	DateID, NameID    []byte
	SubjectID, TermID []byte
	DepictionID       []byte
	// ValueChanges are the audit changes for the value rows this resolve wrote
	// (a date or name create or update; a depiction's image file create, when
	// new, then its depiction_value create), in write order.
	ValueChanges []rowchange.Change
}

// requireSameGraph refuses an Observation whose Subject or Subject value is
// not in the Evidence graph its Citation belongs to.
// A product reification (a source Subject) may be the value of a non-edge
// row in any graph.
func requireSameGraph(tx *sql.Tx, citationID, subjectID, valueSubjectID []byte, edge bool) error {
	graph, err := evidencegraphs.SubjectGraphTx(tx, subjectID)
	if err != nil {
		return mapMissing(err)
	}
	var citationGraph []byte
	if err := tx.QueryRow(sqlCitationGraph, citationID).Scan(&citationGraph); err != nil {
		return mapMissing(err)
	}
	if !bytes.Equal(citationGraph, graph) {
		return ErrOtherGraph
	}
	if len(valueSubjectID) == 0 {
		return nil
	}
	valueGraph, err := evidencegraphs.SubjectGraphTx(tx, valueSubjectID)
	if err != nil {
		return mapMissing(err)
	}
	if bytes.Equal(valueGraph, graph) {
		return nil
	}
	if !edge {
		key, origin, err := subjectTypeKeyTx(tx, valueSubjectID)
		if err != nil {
			return err
		}
		if origin == vocabkey.OriginProvenencia {
			if info, ok := subjectvocab.TypeByKey(key); ok && info.Role == subjectvocab.RoleReification {
				return nil
			}
		}
	}
	return ErrOtherGraph
}

// subjectTypeKeyTx reads a Subject's type key and origin; a missing Subject is
// ErrInvalid.
func subjectTypeKeyTx(tx *sql.Tx, subjectID []byte) (key, origin string, err error) {
	if err := tx.QueryRow(sqlSubjectTypeKey, subjectID).Scan(&key, &origin); err != nil {
		return "", "", mapMissing(err)
	}
	return key, origin, nil
}

// mapMissing turns a missing row into ErrInvalid.
func mapMissing(err error) error {
	if errors.Is(err, sql.ErrNoRows) || errors.Is(err, evidencegraphs.ErrInvalid) {
		return ErrInvalid
	}
	return err
}

func insertOne(tx *sql.Tx, citationID []byte, in Input, opts InsertOptions) (Observation, []rowchange.Change, error) {
	if len(in.ID) != 0 {
		return Observation{}, nil, ErrInvalid
	}
	polarity := strings.TrimSpace(in.Polarity)
	if polarity == "" {
		polarity = PolarityPositive
	}
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
	typeKey, _, err := subjectTypeKeyTx(tx, in.SubjectID)
	if err != nil {
		return Observation{}, nil, err
	}
	_, edge := connectrules.Edge(typeKey, prop.Key)
	if edge && !opts.AllowEdgeRows {
		return Observation{}, nil, ErrEdgeLocked
	}

	resolved, err := resolveValue(tx, citationID, prop.ValueType, in, nil)
	if err != nil {
		return Observation{}, nil, err
	}
	if err := requireSameGraph(tx, citationID, in.SubjectID, resolved.SubjectID, edge); err != nil {
		return Observation{}, nil, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return Observation{}, nil, err
	}
	idBytes := id[:]

	obsRef, err := database.InsertWithRef(database.Minter(ref.PrefixObservation), func(storedRef string) error {
		_, err := tx.Exec(
			sqlInsert,
			idBytes,
			storedRef,
			citationID,
			in.SubjectID,
			in.PropertyID,
			polarity,
			database.NullString(resolved.Text),
			nullInt64(resolved.Integer),
			database.NullBlob(resolved.DateID),
			database.NullBlob(resolved.NameID),
			database.NullBlob(resolved.SubjectID),
			database.NullBlob(resolved.TermID),
			database.NullBlob(resolved.DepictionID),
		)
		return err
	})
	if errors.Is(err, database.ErrRefsExhausted) {
		return Observation{}, nil, ErrInvalid
	}
	if err != nil {
		return Observation{}, nil, mapConstraint(err)
	}

	changes := []rowchange.Change{{
		EntityType: "observation",
		EntityID:   idBytes,
		Action:     rowchange.ActionCreate,
		Fields: rowchange.FullRow(map[string]any{
			"id":                 id.String(),
			"ref":                obsRef,
			"citation_id":        database.UUIDJSON(citationID),
			"subject_id":         database.UUIDJSON(in.SubjectID),
			"property_id":        database.UUIDJSON(in.PropertyID),
			"polarity":           polarity,
			"value_text":         database.NullString(resolved.Text),
			"value_integer":      nullInt64(resolved.Integer),
			"value_date_id":      database.UUIDJSON(resolved.DateID),
			"value_name_id":      database.UUIDJSON(resolved.NameID),
			"value_subject_id":   database.UUIDJSON(resolved.SubjectID),
			"value_term_id":      database.UUIDJSON(resolved.TermID),
			"value_depiction_id": database.UUIDJSON(resolved.DepictionID),
		}),
	}}
	changes = append(changes, resolved.ValueChanges...)

	for _, body := range in.Notes {
		body = strings.TrimSpace(body)
		if body == "" {
			continue
		}
		noteID, err := uuid.NewV7()
		if err != nil {
			return Observation{}, nil, err
		}
		if _, err := tx.Exec(sqlInsertNote, noteID[:], idBytes, body); err != nil {
			return Observation{}, nil, err
		}
		changes = append(changes, rowchange.Change{
			EntityType: "observation_note",
			EntityID:   noteID[:],
			Action:     rowchange.ActionCreate,
			Fields: rowchange.FullRow(map[string]any{
				"id":             noteID.String(),
				"observation_id": id.String(),
				"body":           body,
			}),
		})
	}

	obs := Observation{
		ID:         append([]byte(nil), idBytes...),
		Ref:        obsRef,
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
	return obs, changes, nil
}

func resolveValue(tx *sql.Tx, citationID []byte, valueType string, in Input, existing *Observation) (resolvedValue, error) {
	switch valueType {
	case propertymeta.ValueTypeText:
		text := strings.TrimSpace(in.ValueText)
		if !in.HasText || text == "" || hasOtherScalars(in, "text") {
			return resolvedValue{}, ErrInvalid
		}
		return resolvedValue{Text: text}, nil
	case propertymeta.ValueTypeInteger:
		if !in.HasInteger || hasOtherScalars(in, "integer") {
			return resolvedValue{}, ErrInvalid
		}
		v := in.ValueInteger
		return resolvedValue{Integer: &v}, nil
	case propertymeta.ValueTypeDate:
		if hasOtherScalars(in, "date") {
			return resolvedValue{}, ErrInvalid
		}
		if in.Date != nil {
			if existing != nil && len(existing.ValueDateID) == 16 {
				prev, err := datevalues.LookupTx(tx, existing.ValueDateID)
				if err != nil {
					return resolvedValue{}, err
				}
				if err := datevalues.UpdateTx(tx, existing.ValueDateID, *in.Date); err != nil {
					return resolvedValue{}, err
				}
				next := *in.Date
				next.ID = append([]byte(nil), existing.ValueDateID...)
				resolved := resolvedValue{DateID: append([]byte(nil), existing.ValueDateID...)}
				if ch := dateUpdateChange(prev, next); ch != nil {
					resolved.ValueChanges = []rowchange.Change{*ch}
				}
				return resolved, nil
			}
			dateID, err := datevalues.InsertTx(tx, *in.Date)
			if err != nil {
				return resolvedValue{}, err
			}
			ch := dateCreateChange(dateID, *in.Date)
			return resolvedValue{DateID: dateID, ValueChanges: []rowchange.Change{ch}}, nil
		}
		if len(in.ValueDateID) != 16 {
			return resolvedValue{}, ErrInvalid
		}
		return resolvedValue{DateID: in.ValueDateID}, nil
	case propertymeta.ValueTypeName:
		if hasOtherScalars(in, "name") {
			return resolvedValue{}, ErrInvalid
		}
		if in.Name != nil {
			if existing != nil && len(existing.ValueNameID) == 16 {
				prev, err := namevalues.LookupTx(tx, existing.ValueNameID)
				if err != nil {
					return resolvedValue{}, err
				}
				if err := namevalues.UpdateTx(tx, existing.ValueNameID, *in.Name); err != nil {
					return resolvedValue{}, err
				}
				next := *in.Name
				next.ID = append([]byte(nil), existing.ValueNameID...)
				resolved := resolvedValue{NameID: append([]byte(nil), existing.ValueNameID...)}
				if ch := nameUpdateChange(prev, next); ch != nil {
					resolved.ValueChanges = []rowchange.Change{*ch}
				}
				return resolved, nil
			}
			nameID, err := namevalues.InsertTx(tx, *in.Name)
			if err != nil {
				return resolvedValue{}, err
			}
			ch := nameCreateChange(nameID, *in.Name)
			return resolvedValue{NameID: nameID, ValueChanges: []rowchange.Change{ch}}, nil
		}
		if len(in.ValueNameID) != 16 {
			return resolvedValue{}, ErrInvalid
		}
		return resolvedValue{NameID: in.ValueNameID}, nil
	case propertymeta.ValueTypeSubject:
		if len(in.ValueSubjectID) != 16 || hasOtherScalars(in, "subject") {
			return resolvedValue{}, ErrInvalid
		}
		return resolvedValue{SubjectID: in.ValueSubjectID}, nil
	case propertymeta.ValueTypeTerm:
		if len(in.ValueTermID) != 16 || hasOtherScalars(in, "term") {
			return resolvedValue{}, ErrInvalid
		}
		var one int
		if err := tx.QueryRow(sqlTermOnProperty, in.ValueTermID, in.PropertyID).Scan(&one); err != nil {
			if err == sql.ErrNoRows {
				return resolvedValue{}, ErrInvalid
			}
			return resolvedValue{}, err
		}
		return resolvedValue{TermID: in.ValueTermID}, nil
	case propertymeta.ValueTypeDepiction:
		if hasOtherScalars(in, "depiction") {
			return resolvedValue{}, ErrInvalid
		}
		if in.Depiction != nil {
			return resolveDepiction(tx, citationID, in.Depiction)
		}
		// No new picture: only the Observation's own value row may be kept.
		if existing != nil && len(existing.ValueDepictionID) == 16 && bytes.Equal(in.ValueDepictionID, existing.ValueDepictionID) {
			return resolvedValue{DepictionID: append([]byte(nil), existing.ValueDepictionID...)}, nil
		}
		return resolvedValue{}, ErrInvalid
	default:
		return resolvedValue{}, ErrInvalid
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
	if keep != "depiction" && (in.Depiction != nil || len(in.ValueDepictionID) > 0) {
		return true
	}
	return false
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
	listed, err := scanListed(rows)
	if err != nil {
		return nil, err
	}
	return hydrateNameValues(c, listed)
}

// ListByCitation returns Observations for one citation.
func ListByCitation(c *database.Catalog, citationID []byte) ([]Listed, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	if len(citationID) != 16 {
		return nil, ErrInvalid
	}
	rows, err := db.Query(sqlListByCitation, citationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	listed, err := scanListed(rows)
	if err != nil {
		return nil, err
	}
	return hydrateNameValues(c, listed)
}

// CountReferencingSubject returns Observations that cite subjectID as subject or value.
func CountReferencingSubject(c *database.Catalog, subjectID []byte) (int, error) {
	db, err := c.DB()
	if err != nil {
		return 0, err
	}
	if len(subjectID) != 16 {
		return 0, ErrInvalid
	}
	var n int
	if err := db.QueryRow(sqlCountBySubjectOrValue, subjectID, subjectID).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func scanListed(rows *sql.Rows) ([]Listed, error) {
	var out []Listed
	for rows.Next() {
		var (
			l                                                         Listed
			valueText                                                 sql.NullString
			valueInt                                                  sql.NullInt64
			dateID, nameID, subjectID, termID, depictionID            []byte
			depictionRegion                                           sql.NullString
			depictionImage                                            []byte
			citationLocator, imageMediaType                           sql.NullString
			termLabel, termKey, termOrigin, nameForm, subjectLabel    string
			kind, qual, cal, phrase, startTZ, endTZ                   sql.NullString
			startY, startM, startD, startH, startMin, startS, startMs sql.NullInt64
			endY, endM, endD, endH, endMin, endS, endMs               sql.NullInt64
		)
		if err := rows.Scan(
			&l.ID, &l.Ref, &l.CitationID, &l.SubjectID, &l.PropertyID, &l.Polarity,
			&valueText, &valueInt, &dateID, &nameID, &subjectID, &termID,
			&depictionID,
			&l.PropertyKey, &l.PropertyLabel, &l.PropertyValueType,
			&termLabel, &termKey, &termOrigin, &nameForm, &subjectLabel,
			&kind, &qual, &cal,
			&startY, &startM, &startD,
			&startH, &startMin, &startS, &startMs, &startTZ,
			&endY, &endM, &endD,
			&endH, &endMin, &endS, &endMs, &endTZ,
			&phrase,
			&depictionRegion, &depictionImage, &citationLocator, &imageMediaType,
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
		l.ValueDepictionID = append([]byte(nil), depictionID...)
		l.Depiction = depictionvalues.FromJoin(depictionID, depictionRegion, depictionImage, citationLocator, imageMediaType)
		l.ValueTermLabel = termLabel
		l.ValueTermKey = termKey
		l.ValueTermOrigin = termOrigin
		l.ValueNameForm = nameForm
		l.ValueSubjectLabel = subjectLabel
		if kind.Valid && kind.String != "" {
			dv := datevalue.Value{
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
		FillListedDisplayText(&l)
		out = append(out, l)
	}
	return out, rows.Err()
}

func hydrateNameValues(c *database.Catalog, listed []Listed) ([]Listed, error) {
	ids := make([][]byte, 0, len(listed))
	for i := range listed {
		if len(listed[i].ValueNameID) == 16 {
			ids = append(ids, listed[i].ValueNameID)
		}
	}
	byID, err := namevalues.LookupMany(c, ids)
	if err != nil {
		return nil, err
	}
	for i := range listed {
		v, ok := byID[string(listed[i].ValueNameID)]
		if !ok {
			continue
		}
		nv := v
		listed[i].Name = &nv
		listed[i].ValueNameForm = nv.Form
	}
	return listed, nil
}

// FillListedDisplayText copies a term label, name form, subject label, or
// compact date into ValueText when the row has no text of its own. Card rows
// that only read value_text stay correct for every property value type.
func FillListedDisplayText(l *Listed) {
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
		l.ValueText = datevalue.CompactDisplay(*l.Date)
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

func nullInt64(p *int64) any {
	if p == nil {
		return nil
	}
	return *p
}

func mapConstraint(err error) error {
	if database.IsConstraintViolation(err) {
		return ErrInvalid
	}
	return err
}

func dateUpdateChange(prev, next datevalue.Value) *rowchange.Change {
	fields := map[string]rowchange.FieldDiff{}
	add := func(key string, old, new any) {
		if old != new {
			fields[key] = rowchange.FieldDiff{Old: old, New: new}
		}
	}
	add("kind", database.NullString(prev.Kind), database.NullString(next.Kind))
	add("qualifier", database.NullString(prev.Qualifier), database.NullString(next.Qualifier))
	add("calendar", database.NullString(prev.Calendar), database.NullString(next.Calendar))
	add("start_year", intPtrAny(prev.StartYear), intPtrAny(next.StartYear))
	add("start_month", intPtrAny(prev.StartMonth), intPtrAny(next.StartMonth))
	add("start_day", intPtrAny(prev.StartDay), intPtrAny(next.StartDay))
	add("start_hour", intPtrAny(prev.StartHour), intPtrAny(next.StartHour))
	add("start_minute", intPtrAny(prev.StartMinute), intPtrAny(next.StartMinute))
	add("start_second", intPtrAny(prev.StartSecond), intPtrAny(next.StartSecond))
	add("start_millisecond", intPtrAny(prev.StartMillisecond), intPtrAny(next.StartMillisecond))
	add("start_tz", database.NullString(prev.StartTZ), database.NullString(next.StartTZ))
	add("end_year", intPtrAny(prev.EndYear), intPtrAny(next.EndYear))
	add("end_month", intPtrAny(prev.EndMonth), intPtrAny(next.EndMonth))
	add("end_day", intPtrAny(prev.EndDay), intPtrAny(next.EndDay))
	add("end_hour", intPtrAny(prev.EndHour), intPtrAny(next.EndHour))
	add("end_minute", intPtrAny(prev.EndMinute), intPtrAny(next.EndMinute))
	add("end_second", intPtrAny(prev.EndSecond), intPtrAny(next.EndSecond))
	add("end_millisecond", intPtrAny(prev.EndMillisecond), intPtrAny(next.EndMillisecond))
	add("end_tz", database.NullString(prev.EndTZ), database.NullString(next.EndTZ))
	add("phrase", database.NullString(prev.Phrase), database.NullString(next.Phrase))
	if len(fields) == 0 {
		return nil
	}
	return &rowchange.Change{
		EntityType: "date_value",
		EntityID:   append([]byte(nil), prev.ID...),
		Action:     rowchange.ActionUpdate,
		Fields:     fields,
	}
}

func nameUpdateChange(prev, next namevalue.Value) *rowchange.Change {
	fields := map[string]rowchange.FieldDiff{}
	if prev.Form != next.Form {
		fields["form"] = rowchange.FieldDiff{Old: prev.Form, New: next.Form}
	}
	if !namePartsEqual(prev.Parts, next.Parts) {
		fields["parts"] = rowchange.FieldDiff{Old: namePartsJSON(prev.Parts), New: namePartsJSON(next.Parts)}
	}
	if len(fields) == 0 {
		return nil
	}
	id := prev.ID
	if len(id) != 16 {
		id = next.ID
	}
	return &rowchange.Change{
		EntityType: "name_value",
		EntityID:   append([]byte(nil), id...),
		Action:     rowchange.ActionUpdate,
		Fields:     fields,
	}
}

func namePartsEqual(a, b []namevalue.Part) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Idx != b[i].Idx || a[i].Value != b[i].Value || a[i].Type != b[i].Type {
			return false
		}
	}
	return true
}

func namePartsJSON(parts []namevalue.Part) []map[string]any {
	out := make([]map[string]any, len(parts))
	for i, p := range parts {
		out[i] = map[string]any{
			"idx":   p.Idx,
			"value": p.Value,
			"type":  database.NullString(p.Type),
		}
	}
	return out
}

func dateValueMap(id []byte, v datevalue.Value) map[string]any {
	return map[string]any{
		"id":                database.UUIDJSON(id),
		"kind":              database.NullString(v.Kind),
		"qualifier":         database.NullString(v.Qualifier),
		"calendar":          database.NullString(v.Calendar),
		"start_year":        intPtrAny(v.StartYear),
		"start_month":       intPtrAny(v.StartMonth),
		"start_day":         intPtrAny(v.StartDay),
		"start_hour":        intPtrAny(v.StartHour),
		"start_minute":      intPtrAny(v.StartMinute),
		"start_second":      intPtrAny(v.StartSecond),
		"start_millisecond": intPtrAny(v.StartMillisecond),
		"start_tz":          database.NullString(v.StartTZ),
		"end_year":          intPtrAny(v.EndYear),
		"end_month":         intPtrAny(v.EndMonth),
		"end_day":           intPtrAny(v.EndDay),
		"end_hour":          intPtrAny(v.EndHour),
		"end_minute":        intPtrAny(v.EndMinute),
		"end_second":        intPtrAny(v.EndSecond),
		"end_millisecond":   intPtrAny(v.EndMillisecond),
		"end_tz":            database.NullString(v.EndTZ),
		"phrase":            database.NullString(v.Phrase),
	}
}

func nameValueMap(id []byte, v namevalue.Value) map[string]any {
	return map[string]any{
		"id":    database.UUIDJSON(id),
		"form":  v.Form,
		"parts": namePartsJSON(v.Parts),
	}
}

func dateCreateChange(id []byte, v datevalue.Value) rowchange.Change {
	return rowchange.Change{
		EntityType: "date_value",
		EntityID:   append([]byte(nil), id...),
		Action:     rowchange.ActionCreate,
		Fields: rowchange.FullRow(map[string]any{
			"id":                database.UUIDJSON(id),
			"kind":              database.NullString(v.Kind),
			"qualifier":         database.NullString(v.Qualifier),
			"calendar":          database.NullString(v.Calendar),
			"start_year":        intPtrAny(v.StartYear),
			"start_month":       intPtrAny(v.StartMonth),
			"start_day":         intPtrAny(v.StartDay),
			"start_hour":        intPtrAny(v.StartHour),
			"start_minute":      intPtrAny(v.StartMinute),
			"start_second":      intPtrAny(v.StartSecond),
			"start_millisecond": intPtrAny(v.StartMillisecond),
			"start_tz":          database.NullString(v.StartTZ),
			"end_year":          intPtrAny(v.EndYear),
			"end_month":         intPtrAny(v.EndMonth),
			"end_day":           intPtrAny(v.EndDay),
			"end_hour":          intPtrAny(v.EndHour),
			"end_minute":        intPtrAny(v.EndMinute),
			"end_second":        intPtrAny(v.EndSecond),
			"end_millisecond":   intPtrAny(v.EndMillisecond),
			"end_tz":            database.NullString(v.EndTZ),
			"phrase":            database.NullString(v.Phrase),
		}),
	}
}

func nameCreateChange(id []byte, v namevalue.Value) rowchange.Change {
	parts := make([]map[string]any, len(v.Parts))
	for i, p := range v.Parts {
		parts[i] = map[string]any{
			"idx":   p.Idx,
			"value": p.Value,
			"type":  database.NullString(p.Type),
		}
	}
	return rowchange.Change{
		EntityType: "name_value",
		EntityID:   append([]byte(nil), id...),
		Action:     rowchange.ActionCreate,
		Fields: rowchange.FullRow(map[string]any{
			"id":    database.UUIDJSON(id),
			"form":  v.Form,
			"parts": parts,
		}),
	}
}

func intPtrAny(p *int) any {
	if p == nil {
		return nil
	}
	return *p
}
