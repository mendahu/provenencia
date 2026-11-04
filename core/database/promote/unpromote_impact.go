package promote

import (
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/deleteimpact"
)

// UnpromoteImpact reports the handle an Unpromote would delete when the Subject
// is its last accepted member. It is a read snapshot like deleteimpact.Snapshot.
func UnpromoteImpact(c *database.Catalog, subjectID []byte) (deleteimpact.Group, error) {
	if c == nil || len(subjectID) != 16 {
		return deleteimpact.Group{}, ErrInvalid
	}
	db, err := c.DB()
	if err != nil {
		return deleteimpact.Group{}, err
	}
	tx, err := db.Begin()
	if err != nil {
		return deleteimpact.Group{}, err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := loadPrimary(tx, subjectID); err != nil {
		return deleteimpact.Group{}, err
	}
	return deleteimpact.LastMemberGroup(tx, subjectID)
}
