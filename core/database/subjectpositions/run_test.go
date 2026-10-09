package subjectpositions

import (
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/writes"
)

func runSet(c *database.Catalog, subjectID []byte, gridX, gridY int64) (Position, error) {
	p, _, err := writes.Run(c, writes.Op{},
		func(tx *database.Tx) (Position, []rowchange.Change, error) {
			return Set(tx, subjectID, gridX, gridY)
		})
	return p, err
}

func runClear(c *database.Catalog, subjectID []byte) error {
	_, _, err := writes.Run(c, writes.Op{},
		func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := Clear(tx, subjectID)
			return struct{}{}, changes, err
		})
	return err
}
