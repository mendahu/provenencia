// Package promotetargets is Promote's choose-target read (Spike 9 R7): the
// existing handles a Subject could join, best first, each with its score,
// the reasons for it, and the row header for its kind.
//
// Ranking is core/match via core/database/matching (ForSubject), with the
// Subject type's default profile; this package only shapes the result for
// the picker. Person suggestions carry PersonHeaders; Events and Places carry
// the handle alone until their header composers land (S9-22, S9-25).
// Related-first ordering during a walk is S9-29.
package promotetargets

import (
	"database/sql"
	"errors"

	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/conclusionheaders"
	"github.com/mendahu/provenencia/core/database/matching"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/match"
)

// DefaultLimit caps suggestions when the caller passes no limit.
const DefaultLimit = 10

// Suggestion is one handle the Subject could join.
type Suggestion struct {
	Entity  canonicalentities.Entity
	Score   float64
	Reasons []match.Reason
	// Person is the list header when the handle is a Person.
	Person *conclusionheaders.PersonHeader
}

// Suggest returns up to limit handles the Subject could join, best first.
// An unknown Subject is promote.ErrInvalid; a kind Promote refuses is
// promote.ErrUnsupportedType.
func Suggest(q matching.Querier, subjectID []byte, limit int) ([]Suggestion, error) {
	if len(subjectID) != 16 {
		return nil, promote.ErrInvalid
	}
	if limit <= 0 {
		limit = DefaultLimit
	}
	res, err := matching.ForSubject(q, subjectID, matching.Options{Limit: limit})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, promote.ErrInvalid
	}
	if err != nil {
		return nil, err
	}
	if !promote.PrimaryKind(res.Kind, res.Origin) {
		return nil, promote.ErrUnsupportedType
	}
	if len(res.Matches) == 0 {
		return nil, nil
	}

	ids := make([][]byte, len(res.Matches))
	for i, m := range res.Matches {
		ids[i] = m.EntityID
	}
	entities, err := canonicalentities.GetManyTx(q, ids)
	if err != nil {
		return nil, err
	}
	persons := map[string]conclusionheaders.PersonHeader{}
	if res.Kind == "person" {
		headers, err := conclusionheaders.PersonsByIDs(q, ids)
		if err != nil {
			return nil, err
		}
		for _, h := range headers {
			persons[string(h.Entity.ID)] = h
		}
	}

	out := make([]Suggestion, 0, len(res.Matches))
	for _, m := range res.Matches {
		e, ok := entities[string(m.EntityID)]
		if !ok {
			continue
		}
		s := Suggestion{Entity: e, Score: m.Score, Reasons: m.Reasons}
		if h, ok := persons[string(m.EntityID)]; ok {
			s.Person = &h
		}
		out = append(out, s)
	}
	return out, nil
}
