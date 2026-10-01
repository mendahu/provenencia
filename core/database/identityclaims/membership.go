package identityclaims

import (
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
)

// Membership is one Subject's accepted handle: the canonical entity it belongs
// to and that entity's kind (the subject type key: person, event, place, …).
type Membership struct {
	SubjectID []byte
	Entity    canonicalentities.Entity
	Kind      string
}

// sqlMembershipsBySource joins the Source's Subjects to their accepted claim and
// handle. Unpromoted Subjects have no row.
const sqlMembershipsBySource = `SELECT s.id, e.id, e.subject_type_id, e.ref,
		COALESCE(e.argument, ''), COALESCE(e.label, ''), e.merged_into_id, st.key
	FROM subjects s
	JOIN identity_claims ic ON ic.subject_id = s.id AND ic.status = 'accepted'
	JOIN canonical_entities e ON e.id = ic.entity_id
	JOIN subject_types st ON st.id = e.subject_type_id
	WHERE s.source_id = ?
	ORDER BY s.ref COLLATE NOCASE`

// MembershipsBySource returns the accepted handle of every promoted Subject
// homed to sourceID, in one query. Subjects with no row are unpromoted.
func MembershipsBySource(c *database.Catalog, sourceID []byte) ([]Membership, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	if len(sourceID) != 16 {
		return nil, ErrInvalid
	}
	rows, err := db.Query(sqlMembershipsBySource, sourceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Membership
	for rows.Next() {
		var m Membership
		e := &m.Entity
		if err := rows.Scan(&m.SubjectID, &e.ID, &e.SubjectTypeID, &e.Ref,
			&e.Argument, &e.Label, &e.MergedIntoID, &m.Kind); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
