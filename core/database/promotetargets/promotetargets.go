// Package promotetargets suggests existing handles a Subject could join in
// Promote's choose-target step (Spike 9 R7). Suggestions are same-type,
// unmerged handles that resemble the Subject, ranked by resemblance; the
// picker's search over every handle is the list read, not this one.
//
// v1 resemblance is the resolved name (cache sort_key) of Person handles:
// an exact match ranks before a shared word. Event and Place suggestions wait
// for their header composers and resemblance keys (S9-20 / S9-21, S9-22 /
// S9-25); related-first ordering during a walk is S9-29.
package promotetargets

import (
	"database/sql"
	"errors"
	"sort"
	"strings"

	"github.com/mendahu/provenencia/core/database/conclusionheaders"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/resolve"
)

// DefaultLimit caps suggestions when the caller passes no limit.
const DefaultLimit = 10

// Querier is *sql.Tx or *sql.DB.
type Querier interface {
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

// Suggestions are the ranked handles one Subject could join, as headers of
// the Subject's kind. Only the slice for that kind is filled.
type Suggestions struct {
	Persons []conclusionheaders.PersonHeader
}

const (
	sqlSubjectType = `SELECT st.id, st.key, st.origin
		FROM subjects s JOIN subject_types st ON st.id = s.subject_type_id
		WHERE s.id = ?`

	// The Subject's own positive name Observations.
	sqlSubjectNames = `SELECT o.value_name_id
		FROM observations o JOIN properties p ON p.id = o.property_id
		WHERE o.subject_id = ? AND p.key = 'name' AND p.origin = 'provenencia'
		  AND o.polarity = 'positive' AND o.value_name_id IS NOT NULL`

	// Every resolved name cluster of an unmerged handle of the type, except
	// the handle the Subject already belongs to.
	sqlHandleNames = `SELECT r.entity_id, r.sort_key, r.support
		FROM conclusion_resolved_values r
		JOIN properties np ON np.id = r.property_id
		JOIN canonical_entities e ON e.id = r.entity_id
		WHERE np.key = 'name' AND np.origin = 'provenencia'
		  AND e.subject_type_id = ? AND e.merged_into_id IS NULL
		  AND r.sort_key IS NOT NULL
		  AND e.id NOT IN (SELECT entity_id FROM identity_claims
			WHERE subject_id = ? AND status = 'accepted')`
)

// Resemblance tiers, best first.
const (
	tierExact = iota
	tierSharedWord
)

type match struct {
	tier    int
	support int
}

// Suggest returns up to limit handles the Subject could join, best first:
// an exact resolved-name match, then a shared name word; within a tier, the
// better-supported name first, then list order. A Subject with no name, or of
// a kind without suggestions yet, gets none. Non-primary kinds are
// promote.ErrUnsupportedType, as Promote refuses them.
func Suggest(q Querier, subjectID []byte, limit int) (Suggestions, error) {
	if len(subjectID) != 16 {
		return Suggestions{}, promote.ErrInvalid
	}
	if limit <= 0 {
		limit = DefaultLimit
	}
	var typeID []byte
	var key, origin string
	err := q.QueryRow(sqlSubjectType, subjectID).Scan(&typeID, &key, &origin)
	if errors.Is(err, sql.ErrNoRows) {
		return Suggestions{}, promote.ErrInvalid
	}
	if err != nil {
		return Suggestions{}, err
	}
	if !promote.PrimaryKind(key, origin) {
		return Suggestions{}, promote.ErrUnsupportedType
	}
	if key != "person" {
		return Suggestions{}, nil
	}

	exact, words, err := subjectNameKeys(q, subjectID)
	if err != nil || len(exact) == 0 {
		return Suggestions{}, err
	}
	matches, err := matchHandles(q, typeID, subjectID, exact, words)
	if err != nil || len(matches) == 0 {
		return Suggestions{}, err
	}
	ids := make([][]byte, 0, len(matches))
	for id := range matches {
		ids = append(ids, []byte(id))
	}
	headers, err := conclusionheaders.PersonsByIDs(q, ids)
	if err != nil {
		return Suggestions{}, err
	}
	// headers are in list order; the stable sort keeps it within a rank.
	sort.SliceStable(headers, func(i, j int) bool {
		a, b := matches[string(headers[i].Entity.ID)], matches[string(headers[j].Entity.ID)]
		if a.tier != b.tier {
			return a.tier < b.tier
		}
		return a.support > b.support
	})
	if len(headers) > limit {
		headers = headers[:limit]
	}
	return Suggestions{Persons: headers}, nil
}

// subjectNameKeys returns the sort keys of the Subject's names and the words
// in them worth matching (initials are too weak to suggest on their own).
func subjectNameKeys(q Querier, subjectID []byte) (exact, words map[string]bool, err error) {
	rows, err := q.Query(sqlSubjectNames, subjectID)
	if err != nil {
		return nil, nil, err
	}
	var ids [][]byte
	for rows.Next() {
		var id []byte
		if err := rows.Scan(&id); err != nil {
			_ = rows.Close()
			return nil, nil, err
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	_ = rows.Close()
	if err != nil {
		return nil, nil, err
	}
	names, err := namevalues.LookupManyTx(q, ids)
	if err != nil {
		return nil, nil, err
	}
	exact, words = map[string]bool{}, map[string]bool{}
	for _, n := range names {
		n := n
		k, ok := resolve.SortKey(properties.ValueTypeName, resolve.Value{Name: &n})
		if !ok || k == "" {
			continue
		}
		exact[k] = true
		for _, w := range nameWords(k) {
			words[w] = true
		}
	}
	return exact, words, nil
}

// matchHandles scores every same-type handle with a resembling name cluster,
// keeping each handle's best tier and, within it, best support.
func matchHandles(q Querier, typeID, subjectID []byte, exact, words map[string]bool) (map[string]match, error) {
	rows, err := q.Query(sqlHandleNames, typeID, subjectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]match{}
	for rows.Next() {
		var (
			entityID []byte
			sortKey  string
			support  int
		)
		if err := rows.Scan(&entityID, &sortKey, &support); err != nil {
			return nil, err
		}
		m := match{support: support}
		switch {
		case exact[sortKey]:
			m.tier = tierExact
		case sharesWord(sortKey, words):
			m.tier = tierSharedWord
		default:
			continue
		}
		if prev, ok := out[string(entityID)]; ok &&
			(prev.tier < m.tier || (prev.tier == m.tier && prev.support >= m.support)) {
			continue
		}
		out[string(entityID)] = m
	}
	return out, rows.Err()
}

// nameWords splits a normalized name form (single-spaced, lower case, no
// punctuation) into words of two or more letters.
func nameWords(key string) []string {
	var out []string
	for _, w := range strings.Split(key, " ") {
		if len([]rune(w)) >= 2 {
			out = append(out, w)
		}
	}
	return out
}

func sharesWord(key string, words map[string]bool) bool {
	for _, w := range nameWords(key) {
		if words[w] {
			return true
		}
	}
	return false
}
