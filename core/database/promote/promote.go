// Package promote is the Identity Claim write behind Promote: one transaction
// per promoted Subject. A step either mints a new handle of the Subject's type
// or joins an existing one, and files an accepted claim with zero pins.
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
}

// Result is the handle and the claim one step wrote. On a join, Entity is
// the existing handle.
type Result struct {
	Entity canonicalentities.Entity
	Claim  identityclaims.Claim
}

// Save files an accepted Identity Claim for the Subject in one transaction,
// recorded as promote_subject: onto a newly minted handle of its type, or onto
// in.EntityID. Joining writes the claim only.
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
	if _, err := audit.Record(tx, audit.Revision{
		UserID:     userID,
		ActionType: "promote_subject",
		CreatedAt:  project.NowUTC(),
		Changes:    changes,
	}); err != nil {
		return Result{}, err
	}
	if err := autoreconciler.RecomputeTx(tx, [][]byte{entity.ID}); err != nil {
		return Result{}, err
	}
	if err := tx.Commit(); err != nil {
		return Result{}, err
	}
	return Result{Entity: entity, Claim: claim}, nil
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
