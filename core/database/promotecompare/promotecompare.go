// Package promotecompare is the comparison read behind Promote's compare step
// (S9-17 / S9-19): an incoming Subject's Observations lined up, Property by
// Property, against each accepted member of the handle it would join. Each
// pair says whether the values are compatible by the same module test the
// auto-reconciler uses (autoreconcile.Compatible), which is what the step
// pre-checks. Confirmed pairs go back through promote.Input.Pairs.
package promotecompare

import (
	"bytes"
	"database/sql"

	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/database/conclusiondetails"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/database/promote"
)

// Querier is a *sql.DB or *sql.Tx.
type Querier interface {
	Query(query string, args ...any) (*sql.Rows, error)
}

// Record is one Observation in the comparison and where it comes from.
type Record struct {
	ObservationID  []byte
	ObservationRef string
	SubjectID      []byte
	SubjectRef     string
	SubjectLabel   string
	ClaimID        []byte // the member's accepted claim; nil for the incoming Subject
	CitationID     []byte
	ArtifactID     []byte
	SourceID       []byte
	SourceTitle    string
	Negative       bool
	Value          conclusiondetails.Value
}

// Pair is one member Observation against an incoming one.
type Pair struct {
	Member     Record
	Compatible bool // pre-checked: the same value, or one folds into the other
}

// Incoming is one of the incoming Subject's Observations with the member
// Observations on the same Property beneath it.
type Incoming struct {
	Record Record
	Pairs  []Pair // by member (join order), then Observation id
}

// Property is one Property the incoming Subject speaks to.
type Property struct {
	PropertyID []byte
	Key        string
	Label      string
	ValueType  string
	Incoming   []Incoming // by Observation id
}

// Comparison is the whole compare step for one Subject and one handle.
type Comparison struct {
	Members    int        // accepted members compared against
	Properties []Property // in the kind's binding order
}

const (
	sqlTarget = `SELECT e.merged_into_id IS NULL, e.subject_type_id = s.subject_type_id
		FROM canonical_entities e, subjects s
		WHERE e.id = ? AND s.id = ?`

	sqlMembers = `SELECT COUNT(*) FROM identity_claims
		WHERE entity_id = ? AND status = 'accepted' AND subject_id <> ?`

	// The incoming Subject's Observations and every accepted member's.
	// Subject-valued Properties wait for the subject module (S9-28).
	sqlRecords = `SELECT o.id, o.ref, o.subject_id, s.ref, COALESCE(s.label, ''), ic.id,
			o.property_id, p.key, p.label, p.value_type,
			o.polarity = 'negative',
			o.value_text, o.value_integer, o.value_term_id, COALESCE(t.key, ''), COALESCE(t.label, ''),
			o.value_date_id, o.value_name_id,
			c.id, a.id, src.id, COALESCE(src.title, '')
		FROM observations o
		JOIN subjects s ON s.id = o.subject_id
		JOIN properties p ON p.id = o.property_id
		JOIN citations c ON c.id = o.citation_id
		JOIN artifacts a ON a.id = c.artifact_id
		JOIN sources src ON src.id = a.source_id
		LEFT JOIN property_terms t ON t.id = o.value_term_id
		LEFT JOIN identity_claims ic ON ic.subject_id = o.subject_id
			AND ic.entity_id = ? AND ic.status = 'accepted'
		LEFT JOIN subject_type_properties stp
			ON stp.property_id = p.id AND stp.subject_type_id = s.subject_type_id
		WHERE p.value_type <> 'subject'
		  AND (o.subject_id = ? OR (ic.id IS NOT NULL AND o.subject_id <> ?))
		ORDER BY stp.sort_order IS NULL, stp.sort_order, p.key, ic.id IS NOT NULL, ic.id, o.id`
)

// Compare lines subjectID's Observations up against the accepted members of
// entityID. The handle must be unmerged (else promote.ErrInvalid) and of the
// Subject's type (else identityclaims.ErrTypeMismatch). A handle with no other
// accepted members compares to nothing: zero Members, no pairs.
func Compare(q Querier, subjectID, entityID []byte) (Comparison, error) {
	if len(subjectID) != 16 || len(entityID) != 16 {
		return Comparison{}, promote.ErrInvalid
	}
	if err := checkTarget(q, subjectID, entityID); err != nil {
		return Comparison{}, err
	}
	members, err := countMembers(q, entityID, subjectID)
	if err != nil {
		return Comparison{}, err
	}
	records, err := loadRecords(q, subjectID, entityID)
	if err != nil {
		return Comparison{}, err
	}

	out := Comparison{Members: members}
	byProp := map[string]int{}
	var memberRecords []record
	for _, r := range records {
		if r.ClaimID != nil {
			memberRecords = append(memberRecords, r)
			continue
		}
		i, ok := byProp[string(r.propertyID)]
		if !ok {
			i = len(out.Properties)
			byProp[string(r.propertyID)] = i
			out.Properties = append(out.Properties, Property{
				PropertyID: r.propertyID, Key: r.propertyKey, Label: r.propertyLabel, ValueType: r.valueType,
			})
		}
		out.Properties[i].Incoming = append(out.Properties[i].Incoming, Incoming{Record: r.Record})
	}
	for _, m := range memberRecords {
		i, ok := byProp[string(m.propertyID)]
		if !ok {
			continue // nothing incoming to compare it with
		}
		p := &out.Properties[i]
		for j := range p.Incoming {
			in := &p.Incoming[j]
			in.Pairs = append(in.Pairs, Pair{Member: m.Record, Compatible: compatible(p.ValueType, in.Record, m.Record)})
		}
	}
	return out, nil
}

// compatible: the same polarity, and the module's same value or fold.
func compatible(valueType string, a, b Record) bool {
	return a.Negative == b.Negative && autoreconcile.Compatible(valueType, reconcilable(a.Value), reconcilable(b.Value))
}

func reconcilable(v conclusiondetails.Value) autoreconcile.Value {
	return autoreconcile.Value{
		Text: v.Text, HasText: v.HasText,
		Integer: v.Integer, HasInteger: v.HasInteger,
		TermID: v.TermID, TermKey: v.TermKey,
		Date: v.Date, Name: v.Name,
	}
}

func checkTarget(q Querier, subjectID, entityID []byte) error {
	rows, err := q.Query(sqlTarget, entityID, subjectID)
	if err != nil {
		return err
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		return promote.ErrInvalid
	}
	var unmerged, sameType bool
	if err := rows.Scan(&unmerged, &sameType); err != nil {
		return err
	}
	if !unmerged {
		return promote.ErrInvalid
	}
	if !sameType {
		return identityclaims.ErrTypeMismatch
	}
	return rows.Err()
}

func countMembers(q Querier, entityID, subjectID []byte) (int, error) {
	rows, err := q.Query(sqlMembers, entityID, subjectID)
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

type record struct {
	Record
	propertyID                            []byte
	propertyKey, propertyLabel, valueType string
}

// loadRecords reads every record in a fixed number of queries: the
// Observations, then their date and name values.
func loadRecords(q Querier, subjectID, entityID []byte) ([]record, error) {
	rows, err := q.Query(sqlRecords, entityID, subjectID, subjectID)
	if err != nil {
		return nil, err
	}
	var (
		out              []record
		dateIDs, nameIDs [][]byte
		valueIDs         [][2][]byte // per record: date id, name id
	)
	for rows.Next() {
		var (
			r              record
			text           sql.NullString
			integer        sql.NullInt64
			termID         []byte
			dateID, nameID []byte
		)
		v := &r.Value
		if err := rows.Scan(&r.ObservationID, &r.ObservationRef, &r.SubjectID, &r.SubjectRef, &r.SubjectLabel, &r.ClaimID,
			&r.propertyID, &r.propertyKey, &r.propertyLabel, &r.valueType,
			&r.Negative,
			&text, &integer, &termID, &v.TermKey, &v.TermLabel,
			&dateID, &nameID,
			&r.CitationID, &r.ArtifactID, &r.SourceID, &r.SourceTitle); err != nil {
			_ = rows.Close()
			return nil, err
		}
		v.Text, v.HasText = text.String, text.Valid
		v.Integer, v.HasInteger = integer.Int64, integer.Valid
		v.TermID = termID
		out = append(out, r)
		valueIDs = append(valueIDs, [2][]byte{dateID, nameID})
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
		return nil, err
	}
	dates, err := datevalues.LookupManyTx(q, dateIDs)
	if err != nil {
		return nil, err
	}
	names, err := namevalues.LookupManyTx(q, nameIDs)
	if err != nil {
		return nil, err
	}
	for i := range out {
		if d, ok := dates[string(valueIDs[i][0])]; ok {
			out[i].Value.Date = &d
		}
		if n, ok := names[string(valueIDs[i][1])]; ok {
			out[i].Value.Name = &n
		}
		if bytes.Equal(out[i].SubjectID, subjectID) {
			out[i].ClaimID = nil
		}
	}
	return out, nil
}
