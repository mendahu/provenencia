// Package conclusiondetails composes the detail of one canonical handle from
// the catalog graph. For every Property its members' records speak to, it
// returns the state, every auto-reconciled value (displayed or not, with its
// reason), and the auto-reconciler's outcome for each Observation it
// considered, joined to the record's Source: the "Why" a detail page shows.
//
// Field values are remembered on the graph and dropped with the header rows.
// The field list, term labels, and the Why are read when the page opens, in
// a fixed number of queries whatever the number of Properties. A Property no
// record mentions isn't returned. Which empty fields a page still shows is
// the page's choice. Go returns structures; the app formats text.
// Reconciliation Claims are never in these tables: when they ship, a
// composer lays an accepted claim over the auto-reconciled value here.
package conclusiondetails

import (
	"database/sql"
	"sort"

	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/database/properties"
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
	cardinality string
	Values      []ReconciledValue // rank order
	Outcomes    []Outcome         // by value rank (none last), then Observation id
}

// Detail is one handle's detail.
type Detail struct {
	Entity      canonicalentities.Entity
	MemberCount int // accepted members
	Fields      []Field
}

const (
	// The Properties this handle's records speak to (every considered
	// Observation has an outcome row), in the kind's binding order.
	sqlProperties = `SELECT p.id, p.key, p.label, p.value_type, p.cardinality
		FROM (SELECT DISTINCT property_id FROM auto_reconciler_outcomes WHERE entity_id = ?) ao
		JOIN properties p ON p.id = ao.property_id
		LEFT JOIN subject_type_properties stp ON stp.property_id = p.id AND stp.subject_type_id = ?
		WHERE p.value_type <> 'subject'
		ORDER BY stp.sort_order IS NULL, stp.sort_order, p.key`

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
func ForEntity(c *database.Catalog, entityID []byte) (Detail, error) {
	if c == nil || c.Graph() == nil || len(entityID) != 16 {
		return Detail{}, ErrNotFound
	}
	n, err := c.Graph().Node(entityID)
	if err != nil {
		return Detail{}, err
	}
	if n == nil || n.Merged {
		return Detail{}, ErrNotFound
	}
	db, err := c.DB()
	if err != nil {
		return Detail{}, err
	}
	d := Detail{Entity: entityOf(n), MemberCount: acceptedMembers(n)}
	fields, byProp, err := loadProperties(db, entityID, n.SubjectTypeID)
	if err != nil {
		return Detail{}, err
	}
	if err := attachValues(db, valueMemoOf(c.Graph(), n), byProp); err != nil {
		return Detail{}, err
	}
	if err := loadOutcomes(db, entityID, byProp); err != nil {
		return Detail{}, err
	}
	for _, f := range fields {
		f.State = state(f.Values, f.cardinality)
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

func entityOf(n *graphcache.Node) canonicalentities.Entity {
	return canonicalentities.Entity{
		ID:            append([]byte(nil), n.ID...),
		SubjectTypeID: append([]byte(nil), n.SubjectTypeID...),
		Ref:           n.Ref,
		Argument:      n.Argument,
		Label:         n.Label,
	}
}

func acceptedMembers(n *graphcache.Node) int {
	var nAccepted int
	for _, m := range n.Members {
		if m.Accepted {
			nAccepted++
		}
	}
	return nAccepted
}

// state reads a field's state off its displayed values, by the same rule as
// autoreconcile.Result.State. Several displayed values on a multiple Property
// are all true, so the state is multiple rather than mixed.
func state(values []ReconciledValue, cardinality string) autoreconcile.State {
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
		if cardinality == properties.CardinalityMultiple {
			return autoreconcile.StateMultiple
		}
		return autoreconcile.StateMixed
	case shown[0].Support > 1:
		return autoreconcile.StateMerged
	default:
		return autoreconcile.StateSingle
	}
}

func loadProperties(q Querier, entityID, subjectTypeID []byte) ([]*Field, map[string]*Field, error) {
	rows, err := q.Query(sqlProperties, entityID, subjectTypeID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	var fields []*Field
	byProp := map[string]*Field{}
	for rows.Next() {
		f := &Field{}
		if err := rows.Scan(&f.PropertyID, &f.PropertyKey, &f.Label, &f.ValueType, &f.cardinality); err != nil {
			return nil, nil, err
		}
		fields = append(fields, f)
		byProp[string(f.PropertyID)] = f
	}
	return fields, byProp, rows.Err()
}

type termName struct {
	key   string
	label string
}

func attachValues(q Querier, memo valueMemo, byProp map[string]*Field) error {
	labels, err := termNames(q, memo.termIDs())
	if err != nil {
		return err
	}
	for prop, ranks := range memo.ranks {
		f, ok := byProp[prop]
		if !ok {
			continue // a subject-valued Property, or one the outcomes no longer name
		}
		for _, rank := range ranks {
			v := ReconciledValue{
				Rank: rank.Rank, Reason: rank.Reason,
				Support: rank.Support, Against: rank.Against,
			}
			v.Value.Text, v.Value.HasText = rank.Text, rank.HasText
			v.Value.Integer, v.Value.HasInteger = rank.Integer, rank.HasInteger
			v.Value.TermID = append([]byte(nil), rank.TermID...)
			if name, ok := labels[string(rank.TermID)]; ok {
				v.Value.TermKey, v.Value.TermLabel = name.key, name.label
			}
			if len(rank.Date) > 0 {
				d, err := valuecodec.UnmarshalDate(rank.Date)
				if err != nil {
					return err
				}
				v.Value.Date = &d
			}
			if len(rank.Name) > 0 {
				decoded, err := valuecodec.UnmarshalName(rank.Name)
				if err != nil {
					return err
				}
				v.Value.Name = &decoded
			}
			f.Values = append(f.Values, v)
		}
	}
	return nil
}

func termNames(q Querier, ids [][]byte) (map[string]termName, error) {
	ids = database.UniqueBlobIDs(ids)
	out := map[string]termName{}
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := q.Query(`SELECT id, key, label FROM property_terms WHERE id IN (`+database.SQLInPlaceholders(len(ids))+`)`, database.BlobArgs(ids)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id []byte
		var name termName
		if err := rows.Scan(&id, &name.key, &name.label); err != nil {
			return nil, err
		}
		out[string(id)] = name
	}
	return out, rows.Err()
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
