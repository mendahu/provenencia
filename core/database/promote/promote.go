// Package promote is the Identity Claim write behind Promote: one transaction
// per promoted Subject. A step either mints a new handle of the Subject's type
// or joins an existing one, and files an accepted claim. A join may carry
// confirmed pairs: each pins both Observations on the new claim and backfills
// them onto the member's claim (conclusion-layer-data-model §5.1).
package promote

import (
	"bytes"
	"database/sql"
	"errors"

	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/audit"
	"github.com/mendahu/provenencia/core/database/autoreconciler"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/project"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
)

var (
	ErrInvalid         = apperr.New(apperr.CodePromoteInvalid, apperr.KindUser)
	ErrUnsupportedType = apperr.New(apperr.CodePromoteUnsupportedType, apperr.KindUser)
)

// primaryKinds are the Subject types Promote starts from. Bridge kinds
// (participation, location, relationship) are filed by the walk (S9-28).
var primaryKinds = map[string]bool{"person": true, "event": true, "place": true}

// PrimaryKind reports whether Promote starts from Subjects of this type.
func PrimaryKind(key, origin string) bool {
	return origin == subjecttypes.OriginProvenencia && primaryKinds[key]
}

// Input is one Promote step. A nil EntityID mints a new handle; otherwise
// the Subject joins that handle, which must be unmerged and of its type.
type Input struct {
	SubjectID         []byte
	EntityID          []byte
	ConfidenceGradeID []byte // nil = no grade
	Argument          string
	// Pairs are the comparisons the researcher confirmed. Join only.
	Pairs []Pair
}

// Pair is one confirmed comparison: an Observation of the incoming Subject
// and one of an accepted member of the target, on the same Property.
type Pair struct {
	IncomingObservationID []byte
	MemberObservationID   []byte
}

// Result is the handle and the claim one step wrote. On a join, Entity is
// the existing handle.
type Result struct {
	Entity canonicalentities.Entity
	Claim  identityclaims.Claim
	// Pins is the number of Observations pinned on the new claim.
	Pins int
}

// Save files an accepted Identity Claim for the Subject in one transaction,
// recorded as promote_subject: onto a newly minted handle of its type, or onto
// in.EntityID. A join pins each confirmed pair's two Observations on the new
// claim and on the member's claim; the member's argument is not touched.
func Save(c *database.Catalog, userID []byte, in Input) (Result, error) {
	db, err := c.DB()
	if err != nil {
		return Result{}, err
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return Result{}, err
	}
	if len(in.SubjectID) != 16 || (in.EntityID != nil && len(in.EntityID) != 16) {
		return Result{}, ErrInvalid
	}
	if in.EntityID == nil && len(in.Pairs) > 0 {
		return Result{}, ErrInvalid
	}

	tx, err := db.Begin()
	if err != nil {
		return Result{}, err
	}
	defer func() { _ = tx.Rollback() }()

	subject, err := subjects.GetTx(tx, in.SubjectID)
	if errors.Is(err, sql.ErrNoRows) {
		return Result{}, ErrInvalid
	}
	if err != nil {
		return Result{}, err
	}
	st, err := subjecttypes.GetByIDTx(tx, subject.SubjectTypeID)
	if err != nil {
		return Result{}, err
	}
	if !PrimaryKind(st.Key, st.Origin) {
		return Result{}, ErrUnsupportedType
	}
	if _, err := identityclaims.AcceptedEntityForSubjectTx(tx, subject.ID); err == nil {
		return Result{}, identityclaims.ErrAlreadyMember
	} else if !errors.Is(err, sql.ErrNoRows) {
		return Result{}, err
	}

	var (
		entity  canonicalentities.Entity
		changes []audit.Change
	)
	if in.EntityID == nil {
		minted, change, err := canonicalentities.InsertTx(tx, canonicalentities.CreateInput{
			SubjectTypeID: subject.SubjectTypeID,
		})
		if err != nil {
			return Result{}, err
		}
		entity, changes = minted, append(changes, change)
	} else {
		entity, err = joinTarget(tx, in.EntityID, subject.SubjectTypeID)
		if err != nil {
			return Result{}, err
		}
	}
	claim, claimChange, err := identityclaims.InsertTx(tx, identityclaims.CreateInput{
		SubjectID:         subject.ID,
		EntityID:          entity.ID,
		Status:            identityclaims.StatusAccepted,
		ConfidenceGradeID: in.ConfidenceGradeID,
		Argument:          in.Argument,
	})
	if err != nil {
		return Result{}, err
	}
	changes = append(changes, claimChange)
	pins, pinChanges, err := pinPairs(tx, claim, in.Pairs)
	if err != nil {
		return Result{}, err
	}
	changes = append(changes, pinChanges...)
	assocs, changes, err := identityclaims.AppendFiling(tx, subject.ID, changes)
	if err != nil {
		return Result{}, err
	}
	if _, err := audit.Record(tx, audit.Revision{
		UserID:     userID,
		ActionType: "promote_subject",
		CreatedAt:  project.NowUTC(),
		Changes:    changes,
	}); err != nil {
		return Result{}, err
	}
	if err := autoreconciler.RecomputeTouchingTx(tx, append([][]byte{entity.ID}, assocs...), subject.ID); err != nil {
		return Result{}, err
	}
	if err := tx.Commit(); err != nil {
		return Result{}, err
	}
	return Result{Entity: entity, Claim: claim, Pins: pins}, nil
}

// sqlPairClaim checks one pair and returns the member's claim: the incoming
// Observation is the Subject's, the member Observation is on the same
// Property, and its Subject is an accepted member of the claim's handle.
const sqlPairClaim = `SELECT ic.id
	FROM observations a
	JOIN observations b ON b.property_id = a.property_id
	JOIN identity_claims ic ON ic.subject_id = b.subject_id
		AND ic.entity_id = ? AND ic.status = 'accepted'
	WHERE a.id = ? AND a.subject_id = ? AND b.id = ?`

// pinPairs pins both Observations of every pair on the new claim and on the
// member's claim (backfill). Pins a claim already carries are skipped; it
// returns how many the new claim carries and a change per new pin.
func pinPairs(tx *sql.Tx, claim identityclaims.Claim, pairs []Pair) (int, []audit.Change, error) {
	var (
		changes []audit.Change
		pins    int
	)
	for _, p := range pairs {
		if len(p.IncomingObservationID) != 16 || len(p.MemberObservationID) != 16 {
			return 0, nil, ErrInvalid
		}
		var memberClaimID []byte
		err := tx.QueryRow(sqlPairClaim, claim.EntityID, p.IncomingObservationID, claim.SubjectID,
			p.MemberObservationID).Scan(&memberClaimID)
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil, ErrInvalid
		}
		if err != nil {
			return 0, nil, err
		}
		for _, claimID := range [][]byte{claim.ID, memberClaimID} {
			for _, obsID := range [][]byte{p.IncomingObservationID, p.MemberObservationID} {
				change, ok, err := identityclaims.PinTx(tx, claimID, obsID)
				if err != nil {
					return 0, nil, err
				}
				if !ok {
					continue
				}
				changes = append(changes, change)
				if bytes.Equal(claimID, claim.ID) {
					pins++
				}
			}
		}
	}
	return pins, changes, nil
}

// joinTarget loads an existing handle for a join. A missing or merged handle
// is ErrInvalid; one of another type is identityclaims.ErrTypeMismatch.
func joinTarget(tx *sql.Tx, entityID, subjectTypeID []byte) (canonicalentities.Entity, error) {
	entity, err := canonicalentities.GetTx(tx, entityID)
	if errors.Is(err, sql.ErrNoRows) {
		return canonicalentities.Entity{}, ErrInvalid
	}
	if err != nil {
		return canonicalentities.Entity{}, err
	}
	if entity.MergedIntoID != nil {
		return canonicalentities.Entity{}, ErrInvalid
	}
	if !bytes.Equal(entity.SubjectTypeID, subjectTypeID) {
		return canonicalentities.Entity{}, identityclaims.ErrTypeMismatch
	}
	return entity, nil
}
