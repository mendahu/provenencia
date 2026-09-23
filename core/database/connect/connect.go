// Package connect creates cited Interpretation bridges atomically.
package connect

import (
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/subjectpositions"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
)

var (
	ErrInvalid = apperr.New(apperr.CodeConnectInvalid, apperr.KindUser)
	ErrRefused = apperr.New(apperr.CodeConnectRefused, apperr.KindUser)
)

// CreateInput is one durable Connect submit.
type CreateInput struct {
	SourceID      []byte
	FromSubjectID []byte
	ToSubjectID   []byte
	BridgeTypeKey string
	Label         string
	Description   string
	GridX         int64
	GridY         int64
	Citation      citations.CreateInput
	Observations  []observations.Input
}

// Result is the cited bridge written in one transaction.
type Result struct {
	Subject      subjects.Subject
	Citation     citations.Citation
	Observations []observations.Observation
}

// CreateCitedBridge writes a bridge Subject, position, Citation, and Observations
// in one SQLite transaction. Nothing persists if citation insert fails.
func CreateCitedBridge(c *database.Catalog, userID []byte, in CreateInput) (Result, error) {
	from, err := subjects.Get(c, in.FromSubjectID)
	if err != nil {
		return Result{}, ErrInvalid
	}
	to, err := subjects.Get(c, in.ToSubjectID)
	if err != nil {
		return Result{}, ErrInvalid
	}
	if !sameSource(from.SourceID, in.SourceID) || !sameSource(to.SourceID, in.SourceID) {
		return Result{}, ErrInvalid
	}
	fromType, err := subjecttypes.GetByID(c, from.SubjectTypeID)
	if err != nil {
		return Result{}, ErrInvalid
	}
	toType, err := subjecttypes.GetByID(c, to.SubjectTypeID)
	if err != nil {
		return Result{}, ErrInvalid
	}
	rule := subjectvocab.Connect(fromType.Key, toType.Key)
	if rule.Refuse {
		return Result{}, ErrRefused
	}
	if in.BridgeTypeKey != "" && in.BridgeTypeKey != rule.BridgeTypeKey {
		return Result{}, ErrInvalid
	}
	bridgeType, err := subjecttypes.Lookup(c, rule.BridgeTypeKey, subjecttypes.OriginProvenencia)
	if err != nil {
		return Result{}, ErrInvalid
	}
	if err := requireEdgeObservations(c, rule, in.Observations); err != nil {
		return Result{}, err
	}

	db, err := c.DB()
	if err != nil {
		return Result{}, err
	}
	tx, err := db.Begin()
	if err != nil {
		return Result{}, err
	}
	defer func() { _ = tx.Rollback() }()

	bridge, err := subjects.InsertTx(tx, userID, subjects.CreateInput{
		SourceID:      in.SourceID,
		SubjectTypeID: bridgeType.ID,
		Label:         in.Label,
		Description:   in.Description,
	})
	if err != nil {
		return Result{}, err
	}
	if _, err := subjectpositions.SetTx(tx, bridge.ID, in.GridX, in.GridY); err != nil {
		return Result{}, err
	}

	obs := make([]observations.Input, len(in.Observations))
	copy(obs, in.Observations)
	for i := range obs {
		if len(obs[i].SubjectID) == 0 {
			obs[i].SubjectID = append([]byte(nil), bridge.ID...)
		}
	}

	cited, err := citations.InsertWithObservationsTx(tx, userID, in.Citation, obs)
	if err != nil {
		return Result{}, err
	}
	if err := tx.Commit(); err != nil {
		return Result{}, err
	}
	return Result{
		Subject:      bridge,
		Citation:     cited.Citation,
		Observations: cited.Observations,
	}, nil
}

func requireEdgeObservations(c *database.Catalog, rule subjectvocab.ConnectRule, inputs []observations.Input) error {
	keys := make(map[string]bool, len(inputs))
	for _, in := range inputs {
		prop, err := properties.GetByID(c, in.PropertyID)
		if err != nil {
			return ErrInvalid
		}
		keys[prop.Key] = true
	}
	for _, edge := range rule.EdgePropertyKeys {
		if !keys[edge] {
			return ErrInvalid
		}
	}
	if rule.Disambiguation != "" && rule.Disambiguation != "none" && !keys[rule.Disambiguation] {
		return ErrInvalid
	}
	return nil
}

func sameSource(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
