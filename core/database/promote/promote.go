// Package promote is the Identity Claim write behind Promote: one transaction
// per promoted Subject. A step either mints a new handle of the Subject's type
// or joins an existing one, and files an accepted claim. A batch row may carry
// pin sets: each pins its Observations on the new claim and backfills them onto
// the named member's claim (conclusion-layer-data-model §5.1).
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
	SubjectID         []byte
	EntityID          []byte
	ConfidenceGradeID []byte // nil = no grade
	Argument          string
}

// PinSet is evidence the researcher confirmed: Observations pinned on the new
// claim and backfilled onto Member's accepted claim.
type PinSet struct {
	Member       []byte
	Observations [][]byte
}

// Result is the handle and the claim one step wrote. On a join, Entity is
// the existing handle.
type Result struct {
	Entity canonicalentities.Entity
	Claim  identityclaims.Claim
}

// Save files an accepted Identity Claim for the Subject in one transaction,
// recorded as promote_subject: onto a newly minted handle of its type, or onto
// in.EntityID. It pins nothing: SaveBatch is the one pin write. The caller
// records the returned changes.
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
	st, err := subjecttypes.GetByIDTx(q, subject.SubjectTypeID)
	if err != nil {
		return Result{}, nil, err
	}
	if !vocabkey.IsPrimaryType(st.Key, st.Origin) {
		return Result{}, nil, ErrUnsupportedType
	}
	if _, err := identityclaims.AcceptedEntityForSubjectTx(q, subject.ID); err == nil {
		return Result{}, nil, identityclaims.ErrAlreadyMember
	} else if !errors.Is(err, sql.ErrNoRows) {
		return Result{}, nil, err
	}

	var (
		entity  canonicalentities.Entity
		changes []rowchange.Change
	)
	if in.EntityID == nil {
		minted, change, err := canonicalentities.InsertTx(q, canonicalentities.CreateInput{
			SubjectTypeID: subject.SubjectTypeID,
		})
		if err != nil {
			return Result{}, nil, err
		}
		entity, changes = minted, append(changes, change)
	} else {
		entity, err = joinTarget(q, in.EntityID, subject.SubjectTypeID)
		if err != nil {
			return Result{}, nil, err
		}
	}
	claim, claimChange, err := identityclaims.InsertTx(q, identityclaims.CreateInput{
		SubjectID:         subject.ID,
		EntityID:          entity.ID,
		Status:            identityclaims.StatusAccepted,
		ConfidenceGradeID: in.ConfidenceGradeID,
		Argument:          in.Argument,
	})
	if err != nil {
		return Result{}, nil, err
	}
	changes = append(changes, claimChange)
	_, changes, err = identityclaims.AppendFiling(q, subject.ID, changes)
	if err != nil {
		return Result{}, nil, err
	}
	return Result{Entity: entity, Claim: claim}, changes, nil
}

// pinSets pins every Observation of every set on the new claim and on the
// set's Member's accepted claim (backfill). Member and every Observation id
// must be 16 bytes, and Member must have an accepted claim. Pins a claim
// already carries are skipped; it returns how many the new claim carries and a
// change per new pin.
func pinSets(tx *sql.Tx, claim identityclaims.Claim, sets []PinSet) (int, []rowchange.Change, error) {
	var (
		changes []rowchange.Change
		pins    int
	)
	for _, set := range sets {
		if len(set.Member) != 16 {
			return 0, nil, ErrInvalid
		}
		for _, obsID := range set.Observations {
			if len(obsID) != 16 {
				return 0, nil, ErrInvalid
			}
		}
		member, err := identityclaims.AcceptedEntityForSubjectTx(tx, set.Member)
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil, ErrInvalid
		}
		if err != nil {
			return 0, nil, err
		}
		for _, claimID := range [][]byte{claim.ID, member.ID} {
			for _, obsID := range set.Observations {
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
