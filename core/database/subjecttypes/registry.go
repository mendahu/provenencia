package subjecttypes

import (
	"github.com/mendahu/provenencia/core/database"
)

type seedType struct {
	Key, Label, Description, RefPrefix, CandidateRefPrefix string
}

// Declarative provenencia seed (create-time only). Keys and prefixes from
// docs/seeded-vocabulary.md §3.1.
var seedTypes = []seedType{
	{
		Key: "person", Label: "Person",
		Description: "A person represented by interpreted evidence.",
		RefPrefix:   "PER", CandidateRefPrefix: "CPR",
	},
	{
		Key: "event", Label: "Event",
		Description: "An occurrence represented by interpreted evidence.",
		RefPrefix:   "EVT", CandidateRefPrefix: "CEV",
	},
	{
		Key: "place", Label: "Place",
		Description: "A geographic feature at one grain (town, township, colony, farm, …).",
		RefPrefix:   "PLC", CandidateRefPrefix: "CPL",
	},
	{
		Key: "relationship", Label: "Relationship",
		Description: "A general association when evidence is not a more specific event/context structure.",
		RefPrefix:   "REL", CandidateRefPrefix: "CRL",
	},
	{
		Key: "participation", Label: "Participation",
		Description: "Association between a person and an event, including role.",
		RefPrefix:   "PTN", CandidateRefPrefix: "CPA",
	},
	{
		Key: "location", Label: "Location",
		Description: "Association between an event and a place.",
		RefPrefix:   "LOC", CandidateRefPrefix: "CLO",
	},
	{
		Key: "source", Label: "Source",
		Description: "A Source reified so other evidence can refer to or comment on it.",
		RefPrefix:   "SRN", CandidateRefPrefix: "CSR",
	},
}

// Install writes the provenencia Subject type registry into a new catalog.
// Call only at create time (onboarding.createCatalog). Does not heal deleted
// rows on open.
func Install(c *database.Catalog) error {
	if _, err := c.DB(); err != nil {
		return err
	}
	for _, s := range seedTypes {
		if _, err := Upsert(c, Type{
			Key:                s.Key,
			Origin:             OriginProvenencia,
			Label:              s.Label,
			Description:        s.Description,
			RefPrefix:          s.RefPrefix,
			CandidateRefPrefix: s.CandidateRefPrefix,
		}); err != nil {
			return err
		}
	}
	return nil
}
