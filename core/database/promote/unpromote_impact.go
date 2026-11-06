package promote

import (
	"database/sql"
	"errors"
	"sort"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/catalogmodel"
	"github.com/mendahu/provenencia/core/database/deleteimpact"
	"github.com/mendahu/provenencia/core/database/identityclaims"
)

// UnpromoteReport is what an Unpromote takes away: the handles it deletes and
// the bridge Subjects whose accepted claims it takes back, in bridge Subject
// id order.
type UnpromoteReport struct {
	Deletes                  deleteimpact.Group
	UnpromotedBridgeSubjects [][]byte
}

// UnpromoteImpact reports what an Unpromote would remove: every handle it
// would delete (the primary Subject's handle when it is the last accepted
// member, and association handles whose last bridge member the unfiling
// removes) and the filed bridges it takes back. A Subject with no accepted
// claim yields an empty report, as the write changes nothing. It is a read
// snapshot like deleteimpact.Snapshot.
func UnpromoteImpact(c *database.Catalog, subjectID []byte) (UnpromoteReport, error) {
	if c == nil || len(subjectID) != 16 {
		return UnpromoteReport{}, ErrInvalid
	}
	db, err := c.DB()
	if err != nil {
		return UnpromoteReport{}, err
	}
	tx, err := db.Begin()
	if err != nil {
		return UnpromoteReport{}, err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := loadPrimary(tx, subjectID); err != nil {
		return UnpromoteReport{}, err
	}
	if _, err := identityclaims.AcceptedEntityForSubjectTx(tx, subjectID); errors.Is(err, sql.ErrNoRows) {
		return UnpromoteReport{}, nil
	} else if err != nil {
		return UnpromoteReport{}, err
	}
	filed, err := identityclaims.FiledBridgesOf(tx, subjectID)
	if err != nil {
		return UnpromoteReport{}, err
	}
	rows, err := unpromoteDeletedHandleRows(tx, subjectID, filed)
	if err != nil {
		return UnpromoteReport{}, err
	}
	report := UnpromoteReport{}
	for _, f := range filed {
		report.UnpromotedBridgeSubjects = append(report.UnpromotedBridgeSubjects, f.SubjectID)
	}
	if len(rows) == 0 {
		return report, nil
	}
	report.Deletes, err = deleteimpact.ListedGroup(tx, deleteimpact.ViaLastMember, catalogmodel.KindCanonicalEntity, rows)
	if err != nil {
		return UnpromoteReport{}, err
	}
	return report, nil
}

func unpromoteDeletedHandleRows(tx *sql.Tx, subjectID []byte, filed []identityclaims.FiledBridge) ([]deleteimpact.Listed, error) {
	primary, err := deleteimpact.LastMemberGroup(tx, subjectID)
	if err != nil {
		return nil, err
	}
	bridges, err := identityclaims.UnpromoteBridgeHandlesDeleted(tx, filed)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var rows []deleteimpact.Listed
	if primary.Via != "" {
		for _, l := range primary.Listed {
			key := string(l.ID)
			if seen[key] {
				continue
			}
			seen[key] = true
			rows = append(rows, l)
		}
	}
	for _, b := range bridges {
		key := string(b.ID)
		if seen[key] {
			continue
		}
		seen[key] = true
		rows = append(rows, b)
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Ref < rows[j].Ref })
	return rows, nil
}
