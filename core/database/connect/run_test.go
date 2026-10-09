package connect

import (
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/subjectpositions"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/writes"
)

func runCreateSubject(c *database.Catalog, userID []byte, in subjects.CreateInput, placement *subjects.Placement) (subjects.Subject, error) {
	s, _, err := writes.Run(c, writes.Op{Action: "create_subject", UserID: userID},
		func(tx *database.Tx) (subjects.Subject, []rowchange.Change, error) {
			return subjects.Create(tx, userID, in, placement)
		})
	return s, err
}

func runCreateCitation(c *database.Catalog, userID []byte, in citations.CreateInput, obs []observations.Input) (citations.CreateResult, error) {
	res, _, err := writes.Run(c, writes.Op{Action: "create_citation_with_observations", UserID: userID},
		func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
			return citations.CreateWithObservations(tx, userID, in, obs)
		})
	return res, err
}

func runClearPosition(c *database.Catalog, subjectID []byte) error {
	_, _, err := writes.Run(c, writes.Op{},
		func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := subjectpositions.Clear(tx, subjectID)
			return struct{}{}, changes, err
		})
	return err
}

func runCreateCitedBridge(c *database.Catalog, userID []byte, in CreateInput) (Result, error) {
	res, _, err := writes.Run(c, writes.Op{Action: "create_cited_bridge", UserID: userID},
		func(tx *database.Tx) (Result, []rowchange.Change, error) {
			return CreateCitedBridge(tx, userID, in)
		})
	return res, err
}
