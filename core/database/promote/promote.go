// Package promote is the Identity Claim write behind Promote's Done. A step
// either mints a new handle of the Subject's type or joins an existing one,
// and files an accepted claim. SaveBatch is the production write; Save is the
// single-Subject form tests seed with, and both file a Subject through
// fileSubject.
package promote

import (
	"bytes"
	"database/sql"
	"errors"

	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/vocabkey"

	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
)

var (
	ErrInvalid         = apperr.New(apperr.CodePromoteInvalid, apperr.KindUser)
	ErrUnsupportedType = apperr.New(apperr.CodePromoteUnsupportedType, apperr.KindUser)
	// ErrStale is a lost race: the catalog changed since the proposal was read.
	ErrStale = apperr.New(apperr.CodePromoteStale, apperr.KindConflict)
)

// Input is one Promote step. A nil EntityID mints a new handle; otherwise
// the Subject joins that handle, which must be unmerged and of its type.
type Input struct {
	SubjectID []byte
	EntityID  []byte
}

// Result is the handle and the claim one step wrote. On a join, Entity is
// the existing handle.
type Result struct {
	Entity canonicalentities.Entity
	Claim  identityclaims.Claim
}

// Save files an accepted Identity Claim for one Subject, then its bridges, in
// one transaction: onto a newly minted handle of its type, or onto in.EntityID.
// It is the single-Subject form used to seed tests, by the same filing rule as
// SaveBatch. The caller records the returned changes.
func Save(tx *database.Tx, userID []byte, in Input) (Result, []rowchange.Change, error) {
	if tx == nil {
		return Result{}, nil, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return Result{}, nil, err
	}
	if len(in.SubjectID) != 16 || (in.EntityID != nil && len(in.EntityID) != 16) {
		return Result{}, nil, ErrInvalid
	}

	q := tx.Tx
	subject, err := subjects.GetTx(q, in.SubjectID)
	if errors.Is(err, sql.ErrNoRows) {
		return Result{}, nil, ErrInvalid
	}
	if err != nil {
		return Result{}, nil, err
	}
	if err := requirePrimary(q, subject); err != nil {
		return Result{}, nil, err
	}
	if _, err := identityclaims.AcceptedEntityForSubjectTx(q, subject.ID); err == nil {
		return Result{}, nil, identityclaims.ErrAlreadyMember
	} else if !errors.Is(err, sql.ErrNoRows) {
		return Result{}, nil, err
	}

	entity, claim, changes, err := fileSubject(q, subject, in.EntityID)
	if err != nil {
		return Result{}, nil, err
	}
	_, changes, err = identityclaims.AppendFiling(q, subject.ID, changes)
	if err != nil {
		return Result{}, nil, err
	}
	return Result{Entity: entity, Claim: claim}, changes, nil
}

// requirePrimary refuses a Subject whose type is not a primary type (an
// association is filed by its bridge, never promoted on its own). It runs
// before any membership check, so the refusal does not depend on filing state.
func requirePrimary(tx *sql.Tx, subject subjects.Subject) error {
	st, err := subjecttypes.GetByIDTx(tx, subject.SubjectTypeID)
	if err != nil {
		return err
	}
	if !vocabkey.IsPrimaryType(st.Key, st.Origin) {
		return ErrUnsupportedType
	}
	return nil
}

// fileSubject files one Subject that has no accepted claim: a new handle of
// its type (entityID nil) or a join onto entityID, and the accepted claim.
// Callers check the primary type (requirePrimary) and membership first and
// keep their own follow-up (Save appends the filing of bridges; SaveBatch
// files them once for the whole graph).
func fileSubject(tx *sql.Tx, subject subjects.Subject, entityID []byte) (canonicalentities.Entity, identityclaims.Claim, []rowchange.Change, error) {
	var (
		entity  canonicalentities.Entity
		changes []rowchange.Change
		err     error
	)
	if entityID == nil {
		minted, change, err := canonicalentities.InsertTx(tx, canonicalentities.CreateInput{
			SubjectTypeID: subject.SubjectTypeID,
		})
		if err != nil {
			return canonicalentities.Entity{}, identityclaims.Claim{}, nil, err
		}
		entity, changes = minted, append(changes, change)
	} else {
		entity, err = joinTarget(tx, entityID, subject.SubjectTypeID)
		if err != nil {
			return canonicalentities.Entity{}, identityclaims.Claim{}, nil, err
		}
	}
	claim, claimChange, err := identityclaims.InsertTx(tx, identityclaims.CreateInput{
		SubjectID: subject.ID,
		EntityID:  entity.ID,
		Status:    identityclaims.StatusAccepted,
	})
	if err != nil {
		return canonicalentities.Entity{}, identityclaims.Claim{}, nil, err
	}
	return entity, claim, append(changes, claimChange), nil
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
