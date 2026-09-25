// Package connect creates cited Interpretation bridges atomically.
package connect

import (
	"bytes"
	"database/sql"
	"strings"

	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/audit"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/project"
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
	Description   string
	CitationID    []byte
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
	db, err := c.DB()
	if err != nil {
		return Result{}, err
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return Result{}, err
	}
	tx, err := db.Begin()
	if err != nil {
		return Result{}, err
	}
	defer func() { _ = tx.Rollback() }()

	from, err := subjects.GetTx(tx, in.FromSubjectID)
	if err != nil {
		return Result{}, ErrInvalid
	}
	to, err := subjects.GetTx(tx, in.ToSubjectID)
	if err != nil {
		return Result{}, ErrInvalid
	}
	if bytes.Equal(from.ID, to.ID) {
		return Result{}, ErrInvalid
	}
	if !bytes.Equal(from.SourceID, in.SourceID) || !bytes.Equal(to.SourceID, in.SourceID) {
		return Result{}, ErrInvalid
	}
	fromType, err := subjecttypes.GetByIDTx(tx, from.SubjectTypeID)
	if err != nil {
		return Result{}, ErrInvalid
	}
	toType, err := subjecttypes.GetByIDTx(tx, to.SubjectTypeID)
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
	bound, err := bindEndpoints(rule, from, to, fromType.Key, toType.Key)
	if err != nil {
		return Result{}, err
	}
	keyed, err := exactObservationInputs(tx, rule, in.Observations)
	if err != nil {
		return Result{}, err
	}
	for _, edge := range rule.Edges {
		row := keyed[edge.PropertyKey]
		if !bytes.Equal(row.ValueSubjectID, bound[edge.PropertyKey]) {
			return Result{}, ErrInvalid
		}
		polarity := strings.TrimSpace(row.Polarity)
		if polarity != "" && polarity != observations.PolarityPositive {
			return Result{}, ErrInvalid
		}
	}
	if needsDisambiguation(rule) {
		if len(keyed[rule.Disambiguation].ValueTermID) != 16 {
			return Result{}, ErrInvalid
		}
	}

	if len(in.CitationID) != 0 {
		if !citationFieldsZero(in.Citation) {
			return Result{}, ErrInvalid
		}
		if _, err := citations.GetTx(tx, in.CitationID); err != nil {
			return Result{}, ErrInvalid
		}
		sourceID, err := citations.SourceIDTx(tx, in.CitationID)
		if err != nil || !bytes.Equal(sourceID, in.SourceID) {
			return Result{}, ErrInvalid
		}
	} else if err := citations.NormalizeCreateInput(&in.Citation); err != nil {
		return Result{}, err
	}

	fromPos, err := subjectpositions.GetTx(tx, from.ID)
	if err != nil {
		return Result{}, ErrInvalid
	}
	toPos, err := subjectpositions.GetTx(tx, to.ID)
	if err != nil {
		return Result{}, ErrInvalid
	}

	bridgeType, err := subjecttypes.LookupTx(tx, rule.BridgeTypeKey, subjecttypes.OriginProvenencia)
	if err != nil {
		return Result{}, ErrInvalid
	}
	bridge, subjectChange, err := subjects.InsertTx(tx, subjects.CreateInput{
		SourceID:      in.SourceID,
		SubjectTypeID: bridgeType.ID,
		Description:   in.Description,
	})
	if err != nil {
		return Result{}, err
	}
	gridX := floorDiv(fromPos.GridX+toPos.GridX+1, 2)
	gridY := floorDiv(fromPos.GridY+toPos.GridY+1, 2)
	if _, err := subjectpositions.SetTx(tx, bridge.ID, gridX, gridY); err != nil {
		return Result{}, err
	}

	obs := make([]observations.Input, len(in.Observations))
	copy(obs, in.Observations)
	for i := range obs {
		obs[i].SubjectID = append([]byte(nil), bridge.ID...)
	}

	var (
		citation citations.Citation
		written  []observations.Observation
		changes  = []audit.Change{subjectChange}
	)
	if len(in.CitationID) != 0 {
		var obsChanges []audit.Change
		written, obsChanges, err = observations.InsertManyTx(tx, in.CitationID, obs, observations.InsertOptions{AllowEdgeRows: true})
		if err != nil {
			return Result{}, err
		}
		cit, err := citations.GetTx(tx, in.CitationID)
		if err != nil {
			return Result{}, err
		}
		citation = cit
		changes = append(changes, obsChanges...)
	} else {
		cited, citChanges, err := citations.InsertWithObservationsTx(tx, in.Citation, obs, observations.InsertOptions{AllowEdgeRows: true})
		if err != nil {
			return Result{}, err
		}
		citation = cited.Citation
		written = cited.Observations
		changes = append(changes, citChanges...)
	}
	if _, err := audit.Record(tx, audit.Revision{
		UserID:     userID,
		ActionType: "create_cited_bridge",
		CreatedAt:  project.NowUTC(),
		Changes:    changes,
	}); err != nil {
		return Result{}, err
	}
	if err := tx.Commit(); err != nil {
		return Result{}, err
	}
	return Result{
		Subject:      bridge,
		Citation:     citation,
		Observations: written,
	}, nil
}

func bindEndpoints(
	rule subjectvocab.ConnectRule,
	from, to subjects.Subject,
	fromType, toType string,
) (map[string][]byte, error) {
	if len(rule.Edges) != 2 {
		return nil, ErrInvalid
	}
	a, b := rule.Edges[0], rule.Edges[1]
	bound := make(map[string][]byte, 2)
	if a.EndpointTypeKey != b.EndpointTypeKey {
		for _, edge := range rule.Edges {
			switch edge.EndpointTypeKey {
			case fromType:
				bound[edge.PropertyKey] = from.ID
			case toType:
				bound[edge.PropertyKey] = to.ID
			default:
				return nil, ErrInvalid
			}
		}
		if len(bound) != 2 {
			return nil, ErrInvalid
		}
		return bound, nil
	}
	bound[a.PropertyKey] = from.ID
	bound[b.PropertyKey] = to.ID
	return bound, nil
}

func exactObservationInputs(
	tx *sql.Tx,
	rule subjectvocab.ConnectRule,
	inputs []observations.Input,
) (map[string]observations.Input, error) {
	expected := make(map[string]struct{}, len(rule.Edges)+1)
	for _, edge := range rule.Edges {
		expected[edge.PropertyKey] = struct{}{}
	}
	if needsDisambiguation(rule) {
		expected[rule.Disambiguation] = struct{}{}
	}
	if len(inputs) != len(expected) {
		return nil, ErrInvalid
	}
	keyed := make(map[string]observations.Input, len(inputs))
	for _, in := range inputs {
		if len(in.SubjectID) != 0 {
			return nil, ErrInvalid
		}
		prop, err := properties.GetByIDTx(tx, in.PropertyID)
		if err != nil {
			return nil, ErrInvalid
		}
		if _, ok := expected[prop.Key]; !ok {
			return nil, ErrInvalid
		}
		if _, dup := keyed[prop.Key]; dup {
			return nil, ErrInvalid
		}
		keyed[prop.Key] = in
	}
	if len(keyed) != len(expected) {
		return nil, ErrInvalid
	}
	return keyed, nil
}

func needsDisambiguation(rule subjectvocab.ConnectRule) bool {
	return rule.Disambiguation != "" && rule.Disambiguation != "none"
}

func citationFieldsZero(in citations.CreateInput) bool {
	return len(in.ArtifactID) == 0 &&
		in.LocatorJSON == "" &&
		in.Transcription == "" &&
		in.Description == "" &&
		in.TranscriptionNote == "" &&
		!in.TranscriptionUncertain &&
		len(in.Notes) == 0
}

func floorDiv(a, b int64) int64 {
	if b == 0 {
		return 0
	}
	q := a / b
	if (a^b) < 0 && a%b != 0 {
		q--
	}
	return q
}
