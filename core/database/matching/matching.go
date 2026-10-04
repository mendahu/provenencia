// Package matching runs core/match against the catalog: it reads a probe
// (an Interpretation Subject's Observations, or a handle's resolved values)
// and the candidate handles of the same Subject type (their resolved values
// from the cache, every rank), then ranks them with the type's Profile.
//
// Consumers shape the result for their surface: Promote's target
// suggestions (promotetargets) from ForSubject; merge hints from ForEntity.
//
// Candidates are every unmerged handle of the type that has a cached value
// for a profile Property, read in one query. When catalogs outgrow a scan,
// a blocking step (the R8 search index narrowing by name / date / toponym)
// belongs in loadCandidates; the profile and the scoring do not change.
package matching

import (
	"bytes"
	"database/sql"
	"errors"

	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/match"
	"github.com/mendahu/provenencia/core/valuecodec"
)

// Querier is *sql.Tx or *sql.DB.
type Querier interface {
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

// Options tune one run.
type Options struct {
	// Limit caps the matches; <= 0 returns every match.
	Limit int
	// Profile replaces the type's default profile (match.DefaultProfile).
	Profile *match.Profile
}

// Result is a ranked run. Kind is the probe's Subject type key and Origin
// its origin; Profiled is false when that type has no profile, so nothing
// was matched.
type Result struct {
	Kind     string
	Origin   string
	Profiled bool
	Matches  []match.Match
}

const (
	sqlSubjectType = `SELECT st.id, st.key, st.origin
		FROM subjects s JOIN subject_types st ON st.id = s.subject_type_id
		WHERE s.id = ?`
	sqlEntityType = `SELECT st.id, st.key, st.origin
		FROM canonical_entities e JOIN subject_types st ON st.id = e.subject_type_id
		WHERE e.id = ?`
	sqlOwnHandle = `SELECT entity_id FROM identity_claims
		WHERE subject_id = ? AND status = 'accepted'`

	// A Subject's positive Observations, terms by key.
	sqlSubjectValues = `SELECT p.key, p.origin, o.value_text, o.value_integer,
			o.value_date_id, o.value_name_id, t.key
		FROM observations o
		JOIN properties p ON p.id = o.property_id
		LEFT JOIN property_terms t ON t.id = o.value_term_id
		WHERE o.subject_id = ? AND o.polarity = 'positive'`

	// Cached resolved values (every rank) of unmerged handles, terms by key.
	sqlEntityValues = `SELECT r.entity_id, e.ref, p.key, p.origin, r.value_text, r.value_integer,
			r.value_date, r.value_name, t.key
		FROM conclusion_resolved_values r
		JOIN canonical_entities e ON e.id = r.entity_id
		JOIN properties p ON p.id = r.property_id
		LEFT JOIN property_terms t ON t.id = r.value_term_id
		WHERE e.merged_into_id IS NULL`
	sqlEntityValuesOfType = sqlEntityValues + ` AND e.subject_type_id = ? ORDER BY e.ref, r.entity_id`
	sqlEntityValuesOfOne  = sqlEntityValues + ` AND e.id = ?`
)

// ForSubject ranks the handles an Interpretation Subject could be filed
// onto, never the handle it already belongs to. An unknown Subject is
// sql.ErrNoRows.
func ForSubject(q Querier, subjectID []byte, opts Options) (Result, error) {
	typeID, res, profile, err := start(q, sqlSubjectType, subjectID, opts)
	if err != nil || !res.Profiled {
		return res, err
	}
	probe, err := loadSubjectValues(q, subjectID, profile)
	if err != nil {
		return Result{}, err
	}
	var exclude [][]byte
	var own []byte
	switch err := q.QueryRow(sqlOwnHandle, subjectID).Scan(&own); {
	case err == nil:
		exclude = append(exclude, own)
	case !errors.Is(err, sql.ErrNoRows):
		return Result{}, err
	}
	return rank(q, res, typeID, probe, profile, exclude, opts)
}

// ForEntity ranks the other handles of a handle's type that resemble it:
// the basis for merge hints. An unknown handle is sql.ErrNoRows.
func ForEntity(q Querier, entityID []byte, opts Options) (Result, error) {
	typeID, res, profile, err := start(q, sqlEntityType, entityID, opts)
	if err != nil || !res.Profiled {
		return res, err
	}
	rows, err := q.Query(sqlEntityValuesOfOne, entityID)
	if err != nil {
		return Result{}, err
	}
	probes, err := scanEntityValues(rows, profile)
	if err != nil {
		return Result{}, err
	}
	var probe match.Values
	if len(probes) == 1 {
		probe = probes[0].Values
	}
	return rank(q, res, typeID, probe, profile, [][]byte{entityID}, opts)
}

// start reads the probe's Subject type and picks the profile.
func start(q Querier, typeSQL string, id []byte, opts Options) ([]byte, Result, match.Profile, error) {
	if len(id) != 16 {
		return nil, Result{}, match.Profile{}, sql.ErrNoRows
	}
	var (
		typeID []byte
		res    Result
	)
	if err := q.QueryRow(typeSQL, id).Scan(&typeID, &res.Kind, &res.Origin); err != nil {
		return nil, Result{}, match.Profile{}, err
	}
	if opts.Profile != nil {
		res.Profiled = true
		return typeID, res, *opts.Profile, nil
	}
	if res.Origin != "provenencia" {
		return typeID, res, match.Profile{}, nil
	}
	p, ok := match.DefaultProfile(res.Kind)
	res.Profiled = ok
	return typeID, res, p, nil
}

func rank(q Querier, res Result, typeID []byte, probe match.Values, profile match.Profile, exclude [][]byte, opts Options) (Result, error) {
	if len(probe) == 0 {
		return res, nil
	}
	candidates, err := loadCandidates(q, typeID, profile)
	if err != nil {
		return Result{}, err
	}
	kept := candidates[:0]
	for _, c := range candidates {
		if !containsID(exclude, c.EntityID) {
			kept = append(kept, c)
		}
	}
	res.Matches = match.Rank(profile, probe, kept, opts.Limit)
	return res, nil
}

// loadCandidates reads every unmerged handle of the type with a cached value
// for a profile Property: one query. The blocking seam — narrow it here.
func loadCandidates(q Querier, typeID []byte, profile match.Profile) ([]match.Candidate, error) {
	rows, err := q.Query(sqlEntityValuesOfType, typeID)
	if err != nil {
		return nil, err
	}
	return scanEntityValues(rows, profile)
}

// loadSubjectValues reads the Subject's positive Observations for the
// profile's Properties: one query, plus the date and name lookups.
func loadSubjectValues(q Querier, subjectID []byte, profile match.Profile) (match.Values, error) {
	wanted := propertySet(profile)
	rows, err := q.Query(sqlSubjectValues, subjectID)
	if err != nil {
		return nil, err
	}
	type pending struct {
		prop           match.Property
		v              match.Value
		dateID, nameID []byte
	}
	var (
		all              []pending
		dateIDs, nameIDs [][]byte
	)
	for rows.Next() {
		var (
			p             pending
			text, termKey sql.NullString
			integer       sql.NullInt64
		)
		if err := rows.Scan(&p.prop.Key, &p.prop.Origin, &text, &integer, &p.dateID, &p.nameID, &termKey); err != nil {
			_ = rows.Close()
			return nil, err
		}
		if !wanted[p.prop] {
			continue
		}
		p.v = match.Value{Text: text.String, HasText: text.Valid, Integer: integer.Int64, HasInteger: integer.Valid, Term: termKey.String}
		all = append(all, p)
		if len(p.dateID) > 0 {
			dateIDs = append(dateIDs, p.dateID)
		}
		if len(p.nameID) > 0 {
			nameIDs = append(nameIDs, p.nameID)
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
	out := match.Values{}
	for _, p := range all {
		if d, ok := dates[string(p.dateID)]; ok {
			p.v.Date = &d
		}
		if n, ok := names[string(p.nameID)]; ok {
			p.v.Name = &n
		}
		out[p.prop] = append(out[p.prop], p.v)
	}
	return out, nil
}

// scanEntityValues groups cache rows into one Candidate per handle, in row
// order, keeping only the profile's Properties. It closes rows.
func scanEntityValues(rows *sql.Rows, profile match.Profile) ([]match.Candidate, error) {
	defer rows.Close()
	wanted := propertySet(profile)
	var out []match.Candidate
	for rows.Next() {
		var (
			entityID           []byte
			ref                string
			prop               match.Property
			text, termKey      sql.NullString
			integer            sql.NullInt64
			dateBlob, nameBlob []byte
		)
		if err := rows.Scan(&entityID, &ref, &prop.Key, &prop.Origin, &text, &integer, &dateBlob, &nameBlob, &termKey); err != nil {
			return nil, err
		}
		if !wanted[prop] {
			continue
		}
		v := match.Value{Text: text.String, HasText: text.Valid, Integer: integer.Int64, HasInteger: integer.Valid, Term: termKey.String}
		if dateBlob != nil {
			d, err := valuecodec.UnmarshalDate(dateBlob)
			if err != nil {
				return nil, err
			}
			v.Date = &d
		}
		if nameBlob != nil {
			n, err := valuecodec.UnmarshalName(nameBlob)
			if err != nil {
				return nil, err
			}
			v.Name = &n
		}
		if len(out) == 0 || !bytes.Equal(out[len(out)-1].EntityID, entityID) {
			out = append(out, match.Candidate{EntityID: entityID, Ref: ref, Values: match.Values{}})
		}
		c := &out[len(out)-1]
		c.Values[prop] = append(c.Values[prop], v)
	}
	return out, rows.Err()
}

func propertySet(p match.Profile) map[match.Property]bool {
	out := map[match.Property]bool{}
	for _, prop := range p.Properties() {
		out[prop] = true
	}
	return out
}

func containsID(ids [][]byte, id []byte) bool {
	for _, x := range ids {
		if bytes.Equal(x, id) {
			return true
		}
	}
	return false
}
