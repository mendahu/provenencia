package handlers

import (
	"strings"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/database/observations"
	"google.golang.org/protobuf/proto"
)

func ListObservationsBySource(in []byte) ([]byte, error) {
	var req engine.ListObservationsBySourceRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("list_observations_by_source", err)
	}
	sourceID, err := parseID(req.GetSourceId())
	if err != nil {
		return nil, err
	}
	var out *engine.ListObservationsBySourceResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		rows, err := observations.ListBySource(c, sourceID)
		if err != nil {
			return err
		}
		out = &engine.ListObservationsBySourceResponse{}
		for _, row := range rows {
			out.Observations = append(out.Observations, listedObservationProto(row))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func AddObservationsToCitation(in []byte) ([]byte, error) {
	var req engine.AddObservationsToCitationRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("add_observations_to_citation", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	citationID, err := parseID(req.GetCitationId())
	if err != nil {
		return nil, err
	}
	inputs, err := observationDraftsToInputs(req.GetObservations())
	if err != nil {
		return nil, err
	}
	var out *engine.AddObservationsToCitationResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		rows, err := observations.AddToCitation(c, userID, citationID, inputs)
		if err != nil {
			return err
		}
		out = &engine.AddObservationsToCitationResponse{}
		for _, o := range rows {
			out.Observations = append(out.Observations, observationProto(o))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func observationDraftsToInputs(drafts []*engine.ObservationDraft) ([]observations.Input, error) {
	if len(drafts) == 0 {
		return nil, observations.ErrInvalid
	}
	out := make([]observations.Input, 0, len(drafts))
	for _, d := range drafts {
		in, err := observationDraftToInput(d)
		if err != nil {
			return nil, err
		}
		out = append(out, in)
	}
	return out, nil
}

func observationDraftsToInputsAllowEmptySubject(drafts []*engine.ObservationDraft) ([]observations.Input, error) {
	if len(drafts) == 0 {
		return nil, observations.ErrInvalid
	}
	out := make([]observations.Input, 0, len(drafts))
	for _, d := range drafts {
		in, err := observationDraftToInputAllowEmptySubject(d)
		if err != nil {
			return nil, err
		}
		out = append(out, in)
	}
	return out, nil
}

func observationDraftToInput(d *engine.ObservationDraft) (observations.Input, error) {
	if d == nil || strings.TrimSpace(d.GetSubjectId()) == "" {
		return observations.Input{}, observations.ErrInvalid
	}
	return observationDraftToInputAllowEmptySubject(d)
}

func observationDraftToInputAllowEmptySubject(d *engine.ObservationDraft) (observations.Input, error) {
	if d == nil {
		return observations.Input{}, observations.ErrInvalid
	}
	var subjectID []byte
	if strings.TrimSpace(d.GetSubjectId()) != "" {
		parsed, err := parseID(d.GetSubjectId())
		if err != nil {
			return observations.Input{}, err
		}
		subjectID = parsed
	}
	propertyID, err := parseID(d.GetPropertyId())
	if err != nil {
		return observations.Input{}, err
	}
	in := observations.Input{
		SubjectID:  subjectID,
		PropertyID: propertyID,
		Polarity:   d.GetPolarity(),
		Notes:      d.GetNotes(),
	}
	text := strings.TrimSpace(d.GetValueText())
	if text != "" {
		in.ValueText = text
		in.HasText = true
	}
	if d.ValueInteger != nil {
		in.ValueInteger = d.GetValueInteger()
		in.HasInteger = true
	}
	if d.GetDate() != nil && strings.TrimSpace(d.GetDate().GetKind()) != "" {
		v := dateValueFromProto(d.GetDate())
		in.Date = &v
	}
	if dateID, err := optionalID(d.GetValueDateId()); err != nil {
		return observations.Input{}, err
	} else if dateID != nil {
		in.ValueDateID = dateID
	}
	if d.GetName() != nil && strings.TrimSpace(d.GetName().GetForm()) != "" {
		v := nameValueFromProto(d.GetName())
		in.Name = &v
	}
	if nameID, err := optionalID(d.GetValueNameId()); err != nil {
		return observations.Input{}, err
	} else if nameID != nil {
		in.ValueNameID = nameID
	}
	if subjectValID, err := optionalID(d.GetValueSubjectId()); err != nil {
		return observations.Input{}, err
	} else if subjectValID != nil {
		in.ValueSubjectID = subjectValID
	}
	if termID, err := optionalID(d.GetValueTermId()); err != nil {
		return observations.Input{}, err
	} else if termID != nil {
		in.ValueTermID = termID
	}
	return in, nil
}

func optionalID(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	return parseID(s)
}

func observationProto(o observations.Observation) *engine.Observation {
	return &engine.Observation{
		Id:         uuidString(o.ID),
		Ref:        o.Ref,
		CitationId: uuidString(o.CitationID),
		SubjectId:  uuidString(o.SubjectID),
		PropertyId: uuidString(o.PropertyID),
		Polarity:   o.Polarity,
		ValueText:  o.ValueText,
		ValueInteger: optionalInt64(o.HasInteger, o.ValueInteger),
		ValueDateId:  uuidString(o.ValueDateID),
		ValueNameId:  uuidString(o.ValueNameID),
		ValueSubjectId: uuidString(o.ValueSubjectID),
		ValueTermId:    uuidString(o.ValueTermID),
	}
}

func listedObservationProto(l observations.Listed) *engine.Observation {
	o := observationProto(l.Observation)
	o.PropertyKey = l.PropertyKey
	o.PropertyLabel = l.PropertyLabel
	o.PropertyValueType = l.PropertyValueType
	if l.Date != nil {
		o.Date = dateValueProto(*l.Date)
	}
	if l.Name != nil {
		o.Name = nameValueProto(*l.Name)
	} else if l.ValueNameForm != "" {
		o.Name = &engine.NameValueInput{Form: l.ValueNameForm}
	}
	return o
}

func optionalInt64(has bool, v int64) *int64 {
	if !has {
		return nil
	}
	return &v
}

func nameValueFromProto(n *engine.NameValueInput) namevalues.Value {
	v := namevalues.Value{Form: strings.TrimSpace(n.GetForm())}
	for i, p := range n.GetParts() {
		v.Parts = append(v.Parts, namevalues.Part{
			Idx:   i,
			Value: strings.TrimSpace(p.GetValue()),
			Type:  strings.TrimSpace(p.GetType()),
		})
	}
	return v
}

func nameValueProto(v namevalues.Value) *engine.NameValueInput {
	out := &engine.NameValueInput{Form: v.Form}
	for _, p := range v.Parts {
		out.Parts = append(out.Parts, &engine.NameValuePartInput{
			Value: p.Value,
			Type:  p.Type,
		})
	}
	return out
}
