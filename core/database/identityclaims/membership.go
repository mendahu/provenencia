package identityclaims

import (
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/valuecodec"
)

// Membership is a read-only view of one Subject's accepted Identity Claim:
// the claim, the handle it files the Subject onto, and that handle's kind (the
// subject type key: person, event, place, …). It is the model's "member"
// relation (conclusion-layer-data-model §6), never a stored row of its own;
// provisional and rejected claims are not memberships.
type Membership struct {
	SubjectID []byte
	ClaimID   []byte
	Entity    canonicalentities.Entity
	Kind      string
	// Name is the handle's displayed auto-reconciled name; nil when it has
	// none (unnamed, or not a person).
	Name *namevalues.Value
}

// sqlMembershipsBySource joins the Source's Subjects to their accepted claim and
// handle. Unpromoted Subjects have no row.
const sqlMembershipsBySource = `SELECT s.id, ic.id, e.id, e.subject_type_id, e.ref,
		COALESCE(e.argument, ''), COALESCE(e.label, ''), e.merged_into_id, st.key, r.value_name
	FROM subjects s
	JOIN identity_claims ic ON ic.subject_id = s.id AND ic.status = 'accepted'
	JOIN canonical_entities e ON e.id = ic.entity_id
	JOIN subject_types st ON st.id = e.subject_type_id
	LEFT JOIN properties np ON np.key = 'name' AND np.origin = 'provenencia'
	LEFT JOIN auto_reconciler_values r
		ON r.entity_id = e.id AND r.property_id = np.id AND r.rank = 1
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
		var (
			m        Membership
			nameBlob []byte
		)
		e := &m.Entity
		if err := rows.Scan(&m.SubjectID, &m.ClaimID, &e.ID, &e.SubjectTypeID, &e.Ref,
			&e.Argument, &e.Label, &e.MergedIntoID, &m.Kind, &nameBlob); err != nil {
			return nil, err
		}
		if nameBlob != nil {
			n, err := valuecodec.UnmarshalName(nameBlob)
			if err != nil {
				return nil, err
			}
			m.Name = &n
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
