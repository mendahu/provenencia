// Package connect creates cited Interpretation bridges atomically.
package connect

import (
	"bytes"
	"database/sql"
	"strings"

	"github.com/mendahu/provenencia/core/database/rowchange"

	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/connectrules"
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
// on tx. Nothing persists if citation insert fails. The caller records the
// returned changes. The position is one of them; Run stores it and does not
// record it.
func CreateCitedBridge(tx *database.Tx, userID []byte, in CreateInput) (Result, []rowchange.Change, error) {
	if tx == nil {
		return Result{}, nil, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return Result{}, nil, err
	}

	q := tx.Tx
	from, err := subjects.GetTx(q, in.FromSubjectID)
	if err != nil {
		return Result{}, nil, ErrInvalid
	}
	to, err := subjects.GetTx(q, in.ToSubjectID)
	if err != nil {
		return Result{}, nil, ErrInvalid
	}
	if bytes.Equal(from.ID, to.ID) {
		return Result{}, nil, ErrInvalid
	}
	if !bytes.Equal(from.SourceID, in.SourceID) || !bytes.Equal(to.SourceID, in.SourceID) {
		return Result{}, nil, ErrInvalid
	}
	fromType, err := subjecttypes.GetByIDTx(q, from.SubjectTypeID)
	if err != nil {
		return Result{}, nil, ErrInvalid
	}
	toType, err := subjecttypes.GetByIDTx(q, to.SubjectTypeID)
	if err != nil {
		return Result{}, nil, ErrInvalid
	}
	rule := subjectvocab.Connect(fromType.Key, toType.Key)
	if rule.Refuse {
		return Result{}, nil, ErrRefused
	}
	if in.BridgeTypeKey != "" && in.BridgeTypeKey != rule.BridgeTypeKey {
		return Result{}, nil, ErrInvalid
	}
	bound, err := bindEndpoints(rule, from, to, fromType.Key, toType.Key)
	if err != nil {
		return Result{}, nil, err
	}
	keyed, err := exactObservationInputs(q, rule, in.Observations)
	if err != nil {
		return Result{}, nil, err
	}
	for _, edge := range rule.Edges {
		row := keyed[edge.PropertyKey]
		if !bytes.Equal(row.ValueSubjectID, bound[edge.PropertyKey]) {
			return Result{}, nil, ErrInvalid
		}
		polarity := strings.TrimSpace(row.Polarity)
		if polarity != "" && polarity != observations.PolarityPositive {
			return Result{}, nil, ErrInvalid
		}
	}
	if needsDisambiguation(rule) {
		if len(keyed[rule.Disambiguation].ValueTermID) != 16 {
			return Result{}, nil, ErrInvalid
		}
	}

	if len(in.CitationID) != 0 {
		if !citationFieldsZero(in.Citation) {
			return Result{}, nil, ErrInvalid
		}
		if _, err := citations.GetTx(q, in.CitationID); err != nil {
			return Result{}, nil, ErrInvalid
		}
		sourceID, err := citations.SourceIDTx(q, in.CitationID)
		if err != nil || !bytes.Equal(sourceID, in.SourceID) {
			return Result{}, nil, ErrInvalid
		}
	} else if err := citations.NormalizeCreateInput(&in.Citation); err != nil {
		return Result{}, nil, err
	}

	fromPos, err := subjectpositions.GetTx(q, from.ID)
	if err != nil {
		return Result{}, nil, ErrInvalid
	}
	toPos, err := subjectpositions.GetTx(q, to.ID)
	if err != nil {
		return Result{}, nil, ErrInvalid
	}

	bridgeType, err := subjecttypes.LookupTx(q, rule.BridgeTypeKey, subjecttypes.OriginProvenencia)
	if err != nil {
		return Result{}, nil, ErrInvalid
	}
	bridge, subjectChange, err := subjects.InsertTx(q, subjects.CreateInput{
		SourceID:      in.SourceID,
		SubjectTypeID: bridgeType.ID,
		Description:   in.Description,
	})
	if err != nil {
		return Result{}, nil, err
	}
	gridX := floorDiv(fromPos.GridX+toPos.GridX+1, 2)
	gridY := floorDiv(fromPos.GridY+toPos.GridY+1, 2)
	_, posChanges, err := subjectpositions.Set(tx, bridge.ID, gridX, gridY)
	if err != nil {
		return Result{}, nil, err
	}

	obs := make([]observations.Input, len(in.Observations))
	copy(obs, in.Observations)
	for i := range obs {
		obs[i].SubjectID = append([]byte(nil), bridge.ID...)
	}

	var (
		citation citations.Citation
		written  []observations.Observation
		changes  = append([]rowchange.Change{subjectChange}, posChanges...)
	)
	if len(in.CitationID) != 0 {
		var obsChanges []rowchange.Change
		written, obsChanges, err = observations.InsertManyTx(q, in.CitationID, obs, observations.InsertOptions{AllowEdgeRows: true})
		if err != nil {
			return Result{}, nil, err
		}
		cit, err := citations.GetTx(q, in.CitationID)
		if err != nil {
			return Result{}, nil, err
		}
		citation = cit
		changes = append(changes, obsChanges...)
	} else {
		cited, citChanges, err := citations.InsertWithObservationsTx(q, in.Citation, obs, observations.InsertOptions{AllowEdgeRows: true})
		if err != nil {
			return Result{}, nil, err
		}
		citation = cited.Citation
		written = cited.Observations
		changes = append(changes, citChanges...)
	}
	return Result{
		Subject:      bridge,
		Citation:     citation,
		Observations: written,
	}, changes, nil
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
	return connectrules.HasDisambiguation(rule.Disambiguation)
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
