package sourcecredibilitygrades

import (
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// DefaultKey is the provenencia grade a Source with no assessment reads as:
// every other grade weighs relative to it.
const DefaultKey = "standard"

// RelativeOrderColumn is the select-list column that reads a Source's grade as
// its sort_order relative to the provenencia default grade (DefaultKey). No
// assessment, no grade, or no vocabulary reads as 0. It binds the alias sg
// (source_credibility_grades), which the query joins itself.
const RelativeOrderColumn = `COALESCE(sg.sort_order - (SELECT sort_order FROM source_credibility_grades
				WHERE key = '` + DefaultKey + `' AND origin = '` + vocabkey.OriginProvenencia + `'), 0)`

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
