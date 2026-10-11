package identityclaims

import (
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/namevalues"
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
