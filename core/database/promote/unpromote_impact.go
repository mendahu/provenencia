package promote

import (
	"database/sql"
	"sort"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/catalogmodel"
	"github.com/mendahu/provenencia/core/database/deleteimpact"
	"github.com/mendahu/provenencia/core/database/identityclaims"
)

// UnpromoteImpact reports every handle an Unpromote would delete: the primary
// Subject's handle when it is the last accepted member, and association handles
// whose last bridge member the unfiling removes. It is a read snapshot like
// deleteimpact.Snapshot.
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
	rows, err := unpromoteDeletedHandleRows(tx, subjectID)
	if err != nil {
		return deleteimpact.Group{}, err
	}
	return deleteimpact.GroupFromProbeRows(tx, deleteimpact.ViaLastMember, catalogmodel.KindCanonicalEntity, rows)
}

func unpromoteDeletedHandleRows(tx *sql.Tx, subjectID []byte) ([]deleteimpact.ProbeRow, error) {
	primary, err := lastMemberRowsForUnpromote(tx, subjectID)
	if err != nil {
		return nil, err
	}
	bridges, err := identityclaims.UnpromoteBridgeHandlesDeleted(tx, subjectID)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var rows []deleteimpact.ProbeRow
	for _, r := range primary {
		key := string(r.ID)
		if seen[key] {
			continue
		}
		seen[key] = true
		rows = append(rows, r)
	}
	for _, b := range bridges {
		key := string(b.ID)
		if seen[key] {
			continue
		}
		seen[key] = true
		rows = append(rows, deleteimpact.ProbeRow{ID: b.ID, Ref: b.Ref})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Ref < rows[j].Ref })
	return rows, nil
}

func lastMemberRowsForUnpromote(tx *sql.Tx, subjectID []byte) ([]deleteimpact.ProbeRow, error) {
	group, err := deleteimpact.LastMemberGroup(tx, subjectID)
	if err != nil || group.Via == "" {
		return nil, err
	}
	out := make([]deleteimpact.ProbeRow, len(group.Listed))
	for i, l := range group.Listed {
		out[i] = deleteimpact.ProbeRow{ID: l.ID, Ref: l.Ref}
	}
	return out, nil
}
