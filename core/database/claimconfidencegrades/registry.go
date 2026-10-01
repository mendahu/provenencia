package claimconfidencegrades

import (
	"github.com/mendahu/provenencia/core/database"
)

type seedGrade struct {
	Key, Label string
	SortOrder  int
}

// Declarative provenencia seed (create-time only).
var seedGrades = []seedGrade{
	{Key: "low_confidence", Label: "Low confidence", SortOrder: 1},
	{Key: "moderate", Label: "Moderate", SortOrder: 2},
	{Key: "high_confidence", Label: "High confidence", SortOrder: 3},
}

// Install writes the provenencia claim confidence grade registry into a new catalog.
// Call only at create time (onboarding.createCatalog).
func Install(c *database.Catalog) error {
	if _, err := c.DB(); err != nil {
		return err
	}
	for _, g := range seedGrades {
		if _, err := Upsert(c, Grade{
			Key:       g.Key,
			Origin:    OriginProvenencia,
			Label:     g.Label,
			SortOrder: g.SortOrder,
		}); err != nil {
			return err
		}
	}
	return nil
}
