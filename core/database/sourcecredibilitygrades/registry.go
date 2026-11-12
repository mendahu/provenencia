package sourcecredibilitygrades

import (
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// DefaultKey is the provenencia grade a Source with no assessment reads as:
// every other grade weighs relative to it.
const DefaultKey = "standard"

// orderBound clamps a stored sort_order before it is subtracted, so the
// difference is always an INTEGER. Real grades are small; the bound only stops
// a crafted catalog from overflowing the column to REAL.
const orderBound = "1000000"

// RelativeOrderColumn is the select-list column that reads a Source's grade as
// its sort_order relative to the provenencia default grade (DefaultKey). No
// assessment, no grade, or no vocabulary reads as 0. It binds the alias sg
// (source_credibility_grades); RelativeOrderJoins puts it in scope.
const RelativeOrderColumn = `COALESCE(MAX(-` + orderBound + `, MIN(` + orderBound + `, sg.sort_order)) - MAX(-` + orderBound + `, MIN(` + orderBound + `,
	(SELECT sort_order FROM source_credibility_grades
				WHERE key = '` + DefaultKey + `' AND origin = '` + vocabkey.OriginProvenencia + `'))), 0)`

// RelativeOrderJoins is the join pair that binds sg (the Source's grade) and
// sca (its assessment) for RelativeOrderColumn. sourceIDExpr is the Source's id
// in the query, such as "s.id".
func RelativeOrderJoins(sourceIDExpr string) string {
	return `LEFT JOIN source_credibility_assessments sca ON sca.source_id = ` + sourceIDExpr + `
		LEFT JOIN source_credibility_grades sg ON sg.id = sca.credibility_grade_id`
}

type seedGrade struct {
	Key, Label string
	SortOrder  int
}

// Declarative provenencia seed (create-time only).
var seedGrades = []seedGrade{
	{Key: "low_trust", Label: "Low trust", SortOrder: 1},
	{Key: DefaultKey, Label: "Standard", SortOrder: 2},
	{Key: "high_trust", Label: "High trust", SortOrder: 3},
}

// Install writes the provenencia credibility grade registry into a new catalog.
// Call only at create time (onboarding.createCatalog).
func Install(c *database.Catalog) error {
	if _, err := c.DB(); err != nil {
		return err
	}
	for _, g := range seedGrades {
		if _, err := Upsert(c, Grade{
			Key:       g.Key,
			Origin:    vocabkey.OriginProvenencia,
			Label:     g.Label,
			SortOrder: g.SortOrder,
		}); err != nil {
			return err
		}
	}
	return nil
}
