package handlers

import (
	"strings"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/files"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/valuecodec"
	"github.com/mendahu/provenencia/core/writes"
	"google.golang.org/protobuf/proto"
)

func ListObservationsByEvidenceGraph(in []byte) ([]byte, error) {
	var req engine.ListObservationsByEvidenceGraphRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("list_observations_by_evidence_graph", err)
	}
	graphID, err := parseID(req.GetGraphId())
	if err != nil {
		return nil, err
	}
	var out *engine.ListObservationsByEvidenceGraphResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		eg, err := c.Graph().EvidenceGraph(graphID)
		if err != nil {
			return err
		}
		rows, err := listedProtosFromGraph(c, eg, nil)
		if err != nil {
			return err
		}
		out = &engine.ListObservationsByEvidenceGraphResponse{Observations: rows}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func UpdateObservation(in []byte) ([]byte, error) {
	var req engine.UpdateObservationRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("update_observation", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	input, err := observationToInput(req.GetObservation())
	if err != nil {
		return nil, err
	}
	if len(input.ID) != 16 {
		return nil, observations.ErrInvalid
	}
	input.Depiction = depictionInput(req.GetDepictionInput())
	var out *engine.UpdateObservationResponse
	err = withStagedDepictions(req.GetProjectDir(), []observations.Input{input}, func(c *database.Catalog) error {
		row, wrote, err := writes.Run(c, writes.Op{Action: "update_observation", UserID: userID},
			func(tx *database.Tx) (observations.Listed, []rowchange.Change, error) {
				return observations.Update(tx, userID, input)
			})
		if err != nil {
			return err
		}
		protos, err := listedProtos(c, []observations.Listed{row})
		if err != nil {
			return err
		}
		out = &engine.UpdateObservationResponse{
			Observation: protos[0],
			Effects:     writeEffectsProto(wrote),
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func DeleteObservation(in []byte) ([]byte, error) {
	var req engine.DeleteObservationRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("delete_observation", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	id, err := parseID(req.GetObservationId())
	if err != nil {
		return nil, err
	}
	var out *engine.DeleteObservationResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		_, wrote, err := writes.Run(c, writes.Op{Action: "delete_observation", UserID: userID},
			func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
				changes, err := observations.Delete(tx, userID, id)
				return struct{}{}, changes, err
			})
		if err != nil {
			return err
		}
		out = &engine.DeleteObservationResponse{Effects: writeEffectsProto(wrote)}
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
	err = withStagedDepictions(req.GetProjectDir(), inputs, func(c *database.Catalog) error {
		rows, wrote, err := writes.Run(c, writes.Op{Action: "add_observations", UserID: userID},
			func(tx *database.Tx) ([]observations.Observation, []rowchange.Change, error) {
				return observations.AddToCitation(tx, userID, citationID, inputs)
			})
		if err != nil {
			return err
		}
		listed, err := listedWritten(c, citationID, rows)
		if err != nil {
			return err
		}
		out = &engine.AddObservationsToCitationResponse{Effects: writeEffectsProto(wrote), Observations: listed}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func observationToInput(o *engine.Observation) (observations.Input, error) {
	if o == nil || strings.TrimSpace(o.GetSubjectId()) == "" {
		return observations.Input{}, observations.ErrInvalid
	}
	subjectID, err := parseID(o.GetSubjectId())
	if err != nil {
		return observations.Input{}, err
	}
	propertyID, err := parseID(o.GetPropertyId())
	if err != nil {
		return observations.Input{}, err
	}
	in := observations.Input{
		SubjectID:  subjectID,
		PropertyID: propertyID,
		Polarity:   o.GetPolarity(),
	}
	if id := strings.TrimSpace(o.GetId()); id != "" {
		parsed, err := parseID(id)
		if err != nil {
			return observations.Input{}, err
		}
		in.ID = parsed
	}
	if err := valueInput(&in, o.GetValueText(), o.ValueInteger, o.GetDate(), o.GetValueDateId(), o.GetName(), o.GetValueNameId(), o.GetValueSubjectId(), o.GetValueTermId()); err != nil {
		return observations.Input{}, err
	}
	depictionID, err := optionalID(o.GetValueDepictionId())
	if err != nil {
		return observations.Input{}, err
	}
	in.ValueDepictionID = depictionID
	return in, nil
}

func observationDraftsToInputs(drafts []*engine.ObservationDraft) ([]observations.Input, error) {
	if len(drafts) == 0 {
		return nil, nil
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
	if err := valueInput(&in, d.GetValueText(), d.ValueInteger, d.GetDate(), d.GetValueDateId(), d.GetName(), d.GetValueNameId(), d.GetValueSubjectId(), d.GetValueTermId()); err != nil {
		return observations.Input{}, err
	}
	in.Depiction = depictionInput(d.GetDepiction())
	return in, nil
}

// depictionInput maps the client's depiction (a region and the path of the
// image it rendered) for the write. It checks nothing: the write and staging
// refuse every wrong mix.
func depictionInput(d *engine.DepictionValueInput) *observations.DepictionInput {
	if d == nil {
		return nil
	}
	return &observations.DepictionInput{
		RegionJSON: strings.TrimSpace(d.GetRegionJson()),
		ImagePath:  strings.TrimSpace(d.GetImagePath()),
	}
}

func valueInput(
	in *observations.Input,
	valueText string,
	valueInteger *int64,
	date *engine.DateValueInput,
	valueDateID string,
	name *engine.NameValueInput,
	valueNameID string,
	valueSubjectID string,
	valueTermID string,
) error {
	text := strings.TrimSpace(valueText)
	if text != "" {
		in.ValueText = text
		in.HasText = true
	}
	if valueInteger != nil {
		in.ValueInteger = *valueInteger
		in.HasInteger = true
	}
	if date != nil && strings.TrimSpace(date.GetKind()) != "" {
		v := valuecodec.DateFromProto(date)
		in.Date = &v
	}
	if dateID, err := optionalID(valueDateID); err != nil {
		return err
	} else if dateID != nil {
		in.ValueDateID = dateID
	}
	if name != nil && strings.TrimSpace(name.GetForm()) != "" {
		v := valuecodec.NameFromProto(name)
		in.Name = &v
	}
	if nameID, err := optionalID(valueNameID); err != nil {
		return err
	} else if nameID != nil {
		in.ValueNameID = nameID
	}
	if subjectValID, err := optionalID(valueSubjectID); err != nil {
		return err
	} else if subjectValID != nil {
		in.ValueSubjectID = subjectValID
	}
	if termID, err := optionalID(valueTermID); err != nil {
		return err
	} else if termID != nil {
		in.ValueTermID = termID
	}
	return nil
}

func observationProto(o observations.Observation) *engine.Observation {
	return &engine.Observation{
		Id:               database.UUIDString(o.ID),
		Ref:              o.Ref,
		CitationId:       database.UUIDString(o.CitationID),
		SubjectId:        database.UUIDString(o.SubjectID),
		PropertyId:       database.UUIDString(o.PropertyID),
		Polarity:         o.Polarity,
		ValueText:        o.ValueText,
		ValueInteger:     optionalInt64(o.HasInteger, o.ValueInteger),
		ValueDateId:      database.UUIDString(o.ValueDateID),
		ValueNameId:      database.UUIDString(o.ValueNameID),
		ValueSubjectId:   database.UUIDString(o.ValueSubjectID),
		ValueTermId:      database.UUIDString(o.ValueTermID),
		ValueDepictionId: database.UUIDString(o.ValueDepictionID),
	}
}

// listedObservationProto projects one listed row; call through listedProtos,
// which resolves the depiction image paths it needs.
func listedObservationProto(l observations.Listed, paths map[string]string) *engine.Observation {
	o := observationProto(l.Observation)
	o.PropertyKey = l.PropertyKey
	o.PropertyLabel = l.PropertyLabel
	o.PropertyValueType = l.PropertyValueType
	o.ValueTermKey = l.ValueTermKey
	o.ValueTermOrigin = l.ValueTermOrigin
	if l.Date != nil {
		o.Date = valuecodec.DateToProto(*l.Date)
	}
	if l.Name != nil {
		o.Name = valuecodec.NameToProto(*l.Name)
	} else if l.ValueNameForm != "" {
		o.Name = &engine.NameValueInput{Form: l.ValueNameForm}
	}
	if l.Depiction != nil {
		o.Depiction = &engine.DepictionValue{
			RegionJson:   l.Depiction.RegionJSON,
			ImageRelPath: paths[string(l.Depiction.ImageFileID)],
		}
	}
	return o
}

// listedProtos projects listed rows to response protos, in order, resolving
// every depiction image path in one query.
func listedProtos(c *database.Catalog, rows []observations.Listed) ([]*engine.Observation, error) {
	paths, err := depictionPaths(c, rows)
	if err != nil {
		return nil, err
	}
	out := make([]*engine.Observation, 0, len(rows))
	for _, l := range rows {
		out = append(out, listedObservationProto(l, paths))
	}
	return out, nil
}

// depictionPaths resolves the objects/… path of each listed row's depiction
// image in one query. An image Readable emptied is not looked up, and an id
// with no derivable path reads as an empty path: the region still shows.
func depictionPaths(c *database.Catalog, rows []observations.Listed) (map[string]string, error) {
	var ids [][]byte
	for _, l := range rows {
		if l.Depiction != nil && len(l.Depiction.ImageFileID) > 0 {
			ids = append(ids, l.Depiction.ImageFileID)
		}
	}
	return files.RelPaths(c, ids)
}

// listedWritten returns the Observations just written to a Citation as listed
// rows, in the order they were written.
func listedWritten(c *database.Catalog, citationID []byte, written []observations.Observation) ([]*engine.Observation, error) {
	listed, err := observations.ListByCitation(c, citationID)
	if err != nil {
		return nil, err
	}
	byID := make(map[string]observations.Listed, len(listed))
	for _, l := range listed {
		byID[database.UUIDString(l.ID)] = l
	}
	rows := make([]observations.Listed, 0, len(written))
	for _, o := range written {
		if l, ok := byID[database.UUIDString(o.ID)]; ok {
			rows = append(rows, l)
		}
	}
	return listedProtos(c, rows)
}

func optionalInt64(has bool, v int64) *int64 {
	if !has {
		return nil
	}
	return &v
}
