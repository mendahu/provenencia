package subjects

import (
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/writes"
)

func runCreate(c *database.Catalog, userID []byte, in CreateInput, placement *Placement) (Subject, error) {
	s, _, err := writes.Run(c, writes.Op{Action: "create_subject", UserID: userID},
		func(tx *database.Tx) (Subject, []rowchange.Change, error) {
			return Create(tx, userID, in, placement)
		})
	return s, err
}

func runUpdate(c *database.Catalog, userID, id []byte, label, description string) error {
	_, _, err := writes.Run(c, writes.Op{Action: "update_subject", UserID: userID},
		func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := Update(tx, userID, id, label, description)
			return struct{}{}, changes, err
		})
	return err
}

func runDelete(c *database.Catalog, userID, id []byte) error {
	_, _, err := writes.Run(c, writes.Op{Action: "delete_subject", UserID: userID},
		func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := Delete(tx, userID, id)
			return struct{}{}, changes, err
		})
	return err
}
