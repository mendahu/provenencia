// Package promote is the Identity Claim write behind Promote: one transaction
// per promoted Subject. v1 mints a new handle of the Subject's type and files
// an accepted claim with zero pins (a grounding claim).
package promote

import (
	"database/sql"
	"errors"

	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/audit"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/project"
	"github.com/mendahu/provenencia/core/database/resolvedvalues"
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

// Input is one Promote step.
type Input struct {
	SubjectID []byte
}

// Result is the handle and the claim one step wrote.
type Result struct {
	Entity canonicalentities.Entity
	Claim  identityclaims.Claim
}

// Save mints a handle of the Subject's type and files an accepted Identity
// Claim onto it in one transaction, recorded as promote_subject.
func Save(c *database.Catalog, userID []byte, in Input) (Result, error) {
	db, err := c.DB()
	if err != nil {
		return Result{}, err
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return Result{}, err
	}
	if len(in.SubjectID) != 16 {
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
	if st.Origin != subjecttypes.OriginProvenencia || !primaryKinds[st.Key] {
		return Result{}, ErrUnsupportedType
	}
	if _, err := identityclaims.AcceptedEntityForSubjectTx(tx, subject.ID); err == nil {
		return Result{}, identityclaims.ErrAlreadyMember
	} else if !errors.Is(err, sql.ErrNoRows) {
		return Result{}, err
	}

	entity, entityChange, err := canonicalentities.InsertTx(tx, canonicalentities.CreateInput{
		SubjectTypeID: subject.SubjectTypeID,
	})
	if err != nil {
		return Result{}, err
	}
	claim, claimChange, err := identityclaims.InsertTx(tx, identityclaims.CreateInput{
		SubjectID: subject.ID,
		EntityID:  entity.ID,
		Status:    identityclaims.StatusAccepted,
	})
	if err != nil {
		return Result{}, err
	}
	if _, err := audit.Record(tx, audit.Revision{
		UserID:     userID,
		ActionType: "promote_subject",
		CreatedAt:  project.NowUTC(),
		Changes:    []audit.Change{entityChange, claimChange},
	}); err != nil {
		return Result{}, err
	}
	if err := resolvedvalues.RecomputeTx(tx, [][]byte{entity.ID}); err != nil {
		return Result{}, err
	}
	if err := tx.Commit(); err != nil {
		return Result{}, err
	}
	return Result{Entity: entity, Claim: claim}, nil
}
