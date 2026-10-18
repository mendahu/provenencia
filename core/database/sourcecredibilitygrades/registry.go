package sourcecredibilitygrades

import (
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/vocabkey"
)

type seedGrade struct {
	Key, Label string
	SortOrder  int
}

// Declarative provenencia seed (create-time only).
var seedGrades = []seedGrade{
	{Key: "low_trust", Label: "Low trust", SortOrder: 1},
	{Key: "standard", Label: "Standard", SortOrder: 2},
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
