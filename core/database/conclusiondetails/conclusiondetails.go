// Package conclusiondetails composes the detail of one canonical handle — a
// Person now, Events and Places later — from the auto-reconciler cache
// (Spike 9 R6). For every Property bound to the handle's kind it returns the
// state, every auto-reconciled value (displayed or not, with its reason), and
// the auto-reconciler's outcome for each Observation it considered, joined to
// the record's Source: the "Why" a detail page shows.
//
// Composed at read time, never stored, in a fixed number of queries whatever
// the number of Properties. Go returns structures; the app formats text.
// Reconciliation Claims are never in these tables: when they ship, a
// composer lays an accepted claim over the auto-reconciled value here.
package conclusiondetails

import (
	"database/sql"
	"sort"

	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/valuecodec"
)

// ErrNotFound is returned for an unknown or merged handle.
var ErrNotFound = apperr.New(apperr.CodeConclusionDetailsNotFound, apperr.KindUser)

// Querier is *sql.Tx or *sql.DB.
type Querier interface {
	Query(query string, args ...any) (*sql.Rows, error)
}

// Value is one Property value. Exactly the field for the value type is set.
type Value struct {
	Text       string
	HasText    bool
	Integer    int64
	HasInteger bool
	TermID     []byte
	TermKey    string
	TermLabel  string
	Date       *datevalues.Value
	Name       *namevalues.Value
}

// ReconciledValue is one auto-reconciled value of a Property.
type ReconciledValue struct {
	Rank    int // 1 = first displayed
	Value   Value
	Support int    // distinct Sources
	Against int    // negative records that match it
	Reason  string // "kept" when displayed, else why not
}

// Displayed reports whether the value is one of the displayed values.
func (v ReconciledValue) Displayed() bool { return v.Reason == string(autoreconcile.ReasonKept) }

// Outcome is what the auto-reconciler did with one Observation, with the
// evidence it weighed.
type Outcome struct {
	ObservationID      []byte
	ObservationRef     string
	Reason             string
	ValueRank          int                // the value it went into; 0 for none
	DeniedBy           []byte             // the negative Observation, for "denied"
	Vote               autoreconcile.Vote // the majority that beat it, for "outvoted"
	Recorded           Value              // what the record said
	SubjectID          []byte
	SubjectRef         string
	CitationID         []byte
	ArtifactID         []byte
	SourceID           []byte
	SourceTitle        string
	CredibilityKey     string // "" = no assessment (standard)
	ClaimConfidenceKey string // "" = no grade (moderate)
	// Provenance is the evidence as the auto-reconciler weighed it, relative
	// to the default grades; Provenance.Weak says whether it is weak, and
	// each part says why.
	Provenance autoreconcile.Provenance
}

// Field is one Property of the handle.
type Field struct {
	PropertyID  []byte
	PropertyKey string
	Label       string
	ValueType   string
	State       autoreconcile.State // read off the displayed values
	Values      []ReconciledValue   // rank order
	Outcomes    []Outcome           // by value rank (none last), then Observation id
}

// Detail is one handle's detail.
type Detail struct {
	Entity      canonicalentities.Entity
	MemberCount int // accepted members
	Fields      []Field
}

const (
	sqlProperties = `SELECT p.id, p.key, p.label, p.value_type
		FROM subject_type_properties stp
		JOIN properties p ON p.id = stp.property_id
		WHERE stp.subject_type_id = ? AND p.value_type <> 'subject'
		ORDER BY stp.sort_order, p.key`

	sqlValues = `SELECT v.property_id, v.rank, v.value_text, v.value_integer, v.value_term_id,
			COALESCE(t.key, ''), COALESCE(t.label, ''), v.value_date, v.value_name,
			v.support, v.against, v.reason
		FROM auto_reconciler_values v
		LEFT JOIN property_terms t ON t.id = v.value_term_id
		WHERE v.entity_id = ?
		ORDER BY v.property_id, v.rank`

	sqlMemberCount = `SELECT COUNT(*) FROM identity_claims WHERE entity_id = ? AND status = 'accepted'`

	sqlOutcomes = `SELECT ao.property_id, ao.observation_id, o.ref, ao.reason, ao.value_rank, ao.denied_by,
			COALESCE(ao.vote_support, 0), COALESCE(ao.vote_total, 0),
			o.value_text, o.value_integer, o.value_term_id, COALESCE(t.key, ''), COALESCE(t.label, ''),
			o.value_date_id, o.value_name_id,
			s.id, s.ref, c.id, a.id, src.id, COALESCE(src.title, ''),
			COALESCE(sg.key, ''), COALESCE(cg.key, ''),
			COALESCE(sg.sort_order - (SELECT sort_order FROM source_credibility_grades
				WHERE key = 'standard' AND origin = 'provenencia'), 0),
			c.transcription_uncertain,
			COALESCE(cg.sort_order - (SELECT sort_order FROM claim_confidence_grades
				WHERE key = 'moderate' AND origin = 'provenencia'), 0)
		FROM auto_reconciler_outcomes ao
		JOIN observations o ON o.id = ao.observation_id
		JOIN subjects s ON s.id = o.subject_id
		JOIN citations c ON c.id = o.citation_id
		JOIN artifacts a ON a.id = c.artifact_id
		JOIN sources src ON src.id = a.source_id
		LEFT JOIN property_terms t ON t.id = o.value_term_id
		LEFT JOIN source_credibility_assessments sca ON sca.source_id = src.id
		LEFT JOIN source_credibility_grades sg ON sg.id = sca.credibility_grade_id
		LEFT JOIN identity_claims ic ON ic.subject_id = o.subject_id AND ic.entity_id = ao.entity_id
		LEFT JOIN claim_confidence_grades cg ON cg.id = ic.confidence_grade_id
		WHERE ao.entity_id = ?`
)

// ForEntity composes the detail of one unmerged handle.
func ForEntity(q Querier, entityID []byte) (Detail, error) {
	if len(entityID) != 16 {
		return Detail{}, ErrNotFound
	}
	entities, err := canonicalentities.GetManyTx(q, [][]byte{entityID})
	if err != nil {
		return Detail{}, err
	}
	e, ok := entities[string(entityID)]
	if !ok || len(e.MergedIntoID) > 0 {
		return Detail{}, ErrNotFound
	}
	d := Detail{Entity: e}
	if d.MemberCount, err = memberCount(q, entityID); err != nil {
		return Detail{}, err
	}

	fields, byProp, err := loadProperties(q, e.SubjectTypeID)
	if err != nil {
		return Detail{}, err
	}
	if err := loadValues(q, entityID, byProp); err != nil {
		return Detail{}, err
	}
	if err := loadOutcomes(q, entityID, byProp); err != nil {
		return Detail{}, err
	}
	for _, f := range fields {
		f.State = state(f.Values)
		sort.SliceStable(f.Outcomes, func(i, j int) bool {
			ri, rj := f.Outcomes[i].ValueRank, f.Outcomes[j].ValueRank
			if (ri == 0) != (rj == 0) {
				return rj == 0
			}
			if ri != rj {
				return ri < rj
			}
			return string(f.Outcomes[i].ObservationID) < string(f.Outcomes[j].ObservationID)
		})
		d.Fields = append(d.Fields, *f)
	}
	return d, nil
}

func memberCount(q Querier, entityID []byte) (int, error) {
	rows, err := q.Query(sqlMemberCount, entityID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var n int
	if rows.Next() {
		if err := rows.Scan(&n); err != nil {
			return 0, err
		}
	}
	return n, rows.Err()
}

// state reads a field's state off its displayed values, by the same rule as
// autoreconcile.Result.State.
func state(values []ReconciledValue) autoreconcile.State {
	var shown []ReconciledValue
	for _, v := range values {
		if v.Displayed() {
			shown = append(shown, v)
		}
	}
	switch {
	case len(shown) == 0:
		return autoreconcile.StateEmpty
	case len(shown) > 1:
		return autoreconcile.StateMixed
	case shown[0].Support > 1:
		return autoreconcile.StateMerged
	default:
		return autoreconcile.StateSingle
	}
}

func loadProperties(q Querier, subjectTypeID []byte) ([]*Field, map[string]*Field, error) {
	rows, err := q.Query(sqlProperties, subjectTypeID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	var fields []*Field
	byProp := map[string]*Field{}
	for rows.Next() {
		f := &Field{}
		if err := rows.Scan(&f.PropertyID, &f.PropertyKey, &f.Label, &f.ValueType); err != nil {
			return nil, nil, err
		}
		fields = append(fields, f)
		byProp[string(f.PropertyID)] = f
	}
	return fields, byProp, rows.Err()
}

func loadValues(q Querier, entityID []byte, byProp map[string]*Field) error {
	rows, err := q.Query(sqlValues, entityID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			propertyID, termID, dateBlob, nameBlob []byte
			text                                   sql.NullString
			integer                                sql.NullInt64
			v                                      ReconciledValue
		)
		if err := rows.Scan(&propertyID, &v.Rank, &text, &integer, &termID, &v.Value.TermKey, &v.Value.TermLabel,
			&dateBlob, &nameBlob, &v.Support, &v.Against, &v.Reason); err != nil {
			return err
		}
		f, ok := byProp[string(propertyID)]
		if !ok {
			continue // a Property no longer bound to the kind
		}
		v.Value.Text, v.Value.HasText = text.String, text.Valid
		v.Value.Integer, v.Value.HasInteger = integer.Int64, integer.Valid
		v.Value.TermID = termID
		if len(dateBlob) > 0 {
			d, err := valuecodec.UnmarshalDate(dateBlob)
			if err != nil {
				return err
			}
			v.Value.Date = &d
		}
		if len(nameBlob) > 0 {
			n, err := valuecodec.UnmarshalName(nameBlob)
			if err != nil {
				return err
			}
			v.Value.Name = &n
		}
		f.Values = append(f.Values, v)
	}
	return rows.Err()
}

func loadOutcomes(q Querier, entityID []byte, byProp map[string]*Field) error {
	rows, err := q.Query(sqlOutcomes, entityID)
	if err != nil {
		return err
	}
	type pending struct {
		f              *Field
		o              Outcome
		dateID, nameID []byte
	}
	var all []pending
	var dateIDs, nameIDs [][]byte
	for rows.Next() {
		var (
			propertyID, termID, dateID, nameID []byte
			rank                               sql.NullInt64
			text                               sql.NullString
			integer                            sql.NullInt64
			o                                  Outcome
		)
		if err := rows.Scan(&propertyID, &o.ObservationID, &o.ObservationRef, &o.Reason, &rank, &o.DeniedBy,
			&o.Vote.Support, &o.Vote.Of,
			&text, &integer, &termID, &o.Recorded.TermKey, &o.Recorded.TermLabel, &dateID, &nameID,
			&o.SubjectID, &o.SubjectRef, &o.CitationID, &o.ArtifactID, &o.SourceID, &o.SourceTitle,
			&o.CredibilityKey, &o.ClaimConfidenceKey,
			&o.Provenance.Credibility, &o.Provenance.Uncertain, &o.Provenance.ClaimConfidence); err != nil {
			_ = rows.Close()
			return err
		}
		f, ok := byProp[string(propertyID)]
		if !ok {
			continue
		}
		o.ValueRank = int(rank.Int64)
		o.Recorded.Text, o.Recorded.HasText = text.String, text.Valid
		o.Recorded.Integer, o.Recorded.HasInteger = integer.Int64, integer.Valid
		o.Recorded.TermID = termID
		all = append(all, pending{f: f, o: o, dateID: dateID, nameID: nameID})
		if len(dateID) > 0 {
			dateIDs = append(dateIDs, dateID)
		}
		if len(nameID) > 0 {
			nameIDs = append(nameIDs, nameID)
		}
	}
	err = rows.Err()
	_ = rows.Close()
	if err != nil {
		return err
	}
	dates, err := datevalues.LookupManyTx(q, dateIDs)
	if err != nil {
		return err
	}
	names, err := namevalues.LookupManyTx(q, nameIDs)
	if err != nil {
		return err
	}
	for _, p := range all {
		if d, ok := dates[string(p.dateID)]; ok {
			p.o.Recorded.Date = &d
		}
		if n, ok := names[string(p.nameID)]; ok {
			p.o.Recorded.Name = &n
		}
		p.f.Outcomes = append(p.f.Outcomes, p.o)
	}
	return nil
}
