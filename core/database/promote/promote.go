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
	"github.com/mendahu/provenencia/core/connectrules"
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
	// ErrStale is a lost race: the catalog changed since the proposal was read.
	ErrStale = apperr.New(apperr.CodePromoteStale, apperr.KindConflict)
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
	pins, pinChanges, err := pinPairs(tx, claim, in.Pairs, nil)
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

// pinTargets is what each Subject in one write is being filed on: the handle
// for a join, nil for New or Skip. A Subject outside the write is on the
// handle its accepted claim names, if any.
type pinTargets map[string][]byte

func (t pinTargets) handleOf(tx *sql.Tx, subjectID []byte) ([]byte, bool, error) {
	if h, ok := t[string(subjectID)]; ok {
		return h, h != nil, nil
	}
	cl, err := identityclaims.AcceptedEntityForSubjectTx(tx, subjectID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return cl.EntityID, true, nil
}

// sqlOneHop is a row when the two Subjects are the ends of one bridge.
func sqlOneHop(bridgeKeys int) string {
	return `SELECT 1 FROM observations e1
	JOIN observations e2 ON e2.subject_id = e1.subject_id AND e2.id != e1.id
		AND e2.polarity = 'positive' AND e2.value_subject_id = ?
	JOIN subjects bs ON bs.id = e1.subject_id
	JOIN subject_types st ON st.id = bs.subject_type_id AND st.origin = 'provenencia'
		AND st.key IN (` + database.SQLInPlaceholders(bridgeKeys) + `)
	WHERE e1.value_subject_id = ? AND e1.polarity = 'positive'
	LIMIT 1`
}

// sqlNearHandle is a row when the Subject is one bridge from an accepted
// member of the handle.
func sqlNearHandle(bridgeKeys int) string {
	return `SELECT 1 FROM identity_claims m
	JOIN observations e1 ON e1.value_subject_id = m.subject_id AND e1.polarity = 'positive'
	JOIN observations e2 ON e2.subject_id = e1.subject_id AND e2.id != e1.id
		AND e2.polarity = 'positive' AND e2.value_subject_id = ?
	JOIN subjects bs ON bs.id = e1.subject_id
	JOIN subject_types st ON st.id = bs.subject_type_id AND st.origin = 'provenencia'
		AND st.key IN (` + database.SQLInPlaceholders(bridgeKeys) + `)
	WHERE m.entity_id = ? AND m.status = 'accepted'
	LIMIT 1`
}

func bridgeRow(tx *sql.Tx, query string, first, second []byte) (bool, error) {
	args := []any{first}
	for _, k := range connectrules.BridgeTypeKeys() {
		args = append(args, k)
	}
	args = append(args, second)
	var one int
	err := tx.QueryRow(query, args...).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

// pairClaim checks one pair and returns the accepted claim that owns the
// member Observation. Both Observations share a Property. Either the
// incoming one is the Subject's own and the member one sits on the claim's
// handle, or the incoming one sits on a neighbor one bridge from the Subject
// and the member one sits on the handle that neighbor is being filed on (or
// already is), on a member one bridge from the claim's handle. A neighbor
// filed New or skipped pins nothing: its records match no existing handle.
func pairClaim(tx *sql.Tx, claim identityclaims.Claim, p Pair, targets pinTargets) ([]byte, error) {
	if len(p.IncomingObservationID) != 16 || len(p.MemberObservationID) != 16 {
		return nil, ErrInvalid
	}
	var inSubject, inProp, memSubject, memProp []byte
	read := func(id []byte, subject, prop *[]byte) error {
		err := tx.QueryRow(`SELECT subject_id, property_id FROM observations WHERE id = ?`, id).Scan(subject, prop)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalid
		}
		return err
	}
	if err := read(p.IncomingObservationID, &inSubject, &inProp); err != nil {
		return nil, err
	}
	if err := read(p.MemberObservationID, &memSubject, &memProp); err != nil {
		return nil, err
	}
	if !bytes.Equal(inProp, memProp) {
		return nil, ErrInvalid
	}
	member, err := identityclaims.AcceptedEntityForSubjectTx(tx, memSubject)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInvalid
	}
	if err != nil {
		return nil, err
	}
	if bytes.Equal(inSubject, claim.SubjectID) {
		if !bytes.Equal(member.EntityID, claim.EntityID) {
			return nil, ErrInvalid
		}
		return member.ID, nil
	}
	keys := len(connectrules.BridgeTypeKeys())
	hop, err := bridgeRow(tx, sqlOneHop(keys), inSubject, claim.SubjectID)
	if err != nil {
		return nil, err
	}
	neighborHandle, filed, err := targets.handleOf(tx, inSubject)
	if err != nil {
		return nil, err
	}
	if !hop || !filed || !bytes.Equal(member.EntityID, neighborHandle) {
		return nil, ErrInvalid
	}
	near, err := bridgeRow(tx, sqlNearHandle(keys), memSubject, claim.EntityID)
	if err != nil {
		return nil, err
	}
	if !near {
		return nil, ErrInvalid
	}
	return member.ID, nil
}

// pinPairs pins both Observations of every pair on the new claim and on the
// member's claim (backfill). Pins a claim already carries are skipped; it
// returns how many the new claim carries and a change per new pin.
func pinPairs(tx *sql.Tx, claim identityclaims.Claim, pairs []Pair, targets pinTargets) (int, []audit.Change, error) {
	var (
		changes []audit.Change
		pins    int
	)
	for _, p := range pairs {
		memberClaimID, err := pairClaim(tx, claim, p, targets)
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
