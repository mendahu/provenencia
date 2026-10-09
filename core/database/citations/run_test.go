package citations

import (
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/rowchange"
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

func runCreateWithObservations(c *database.Catalog, userID []byte, in CreateInput, obs []observations.Input) (CreateResult, error) {
	res, _, err := writes.Run(c, writes.Op{Action: "create_citation_with_observations", UserID: userID},
		func(tx *database.Tx) (CreateResult, []rowchange.Change, error) {
			return CreateWithObservations(tx, userID, in, obs)
		})
	return res, err
}

func runUpdate(c *database.Catalog, userID, citationID []byte, in CitationFieldsInput) (Citation, error) {
	cit, _, err := writes.Run(c, writes.Op{Action: "update_citation", UserID: userID},
		func(tx *database.Tx) (Citation, []rowchange.Change, error) {
			return Update(tx, userID, citationID, in)
		})
	return cit, err
}

func runDelete(c *database.Catalog, userID, id []byte) error {
	_, _, err := writes.Run(c, writes.Op{Action: "delete_citation", UserID: userID},
		func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := Delete(tx, userID, id)
			return struct{}{}, changes, err
		})
	return err
}
