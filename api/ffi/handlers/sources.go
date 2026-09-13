package handlers

import (
	"bytes"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/files"
	"github.com/mendahu/provenencia/core/database/sourcecredibilitygrades"
	"github.com/mendahu/provenencia/core/database/sourcefields"
	"github.com/mendahu/provenencia/core/database/sourcemetadata"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"google.golang.org/protobuf/proto"
)

func ListSources(in []byte) ([]byte, error) {
	var req engine.ListSourcesRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("list_sources", err)
	}
	var out *engine.ListSourcesResponse
	err := withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		rows, err := sources.List(c)
		if err != nil {
			return err
		}
		out = &engine.ListSourcesResponse{}
		for _, s := range rows {
			sp, err := enrichSourceProto(c, s)
			if err != nil {
				return err
			}
			out.Sources = append(out.Sources, sp)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func GetSourceWorkspace(in []byte) ([]byte, error) {
	var req engine.GetSourceWorkspaceRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("get_source_workspace", err)
	}
	sourceID, err := parseID(req.GetSourceId())
	if err != nil {
		return nil, err
	}
	var out *engine.GetSourceWorkspaceResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		s, err := sources.Get(c, sourceID)
		if err != nil {
			return err
		}
		notes, err := sources.ListNotes(c, sourceID)
		if err != nil {
			return err
		}
		meta, err := sourcemetadata.ListWorkspace(c, sourceID)
		if err != nil {
			return err
		}
		arts, err := listArtifactsProto(c, sourceID)
		if err != nil {
			return err
		}
		cred, err := credibilityForSource(c, sourceID)
		if err != nil {
			return err
		}
		types, err := sourcetypes.List(c)
		if err != nil {
			return err
		}
		grades, err := sourcecredibilitygrades.List(c)
		if err != nil {
			return err
		}
		fields, err := sourcefields.List(c)
		if err != nil {
			return err
		}

		sp, err := enrichSourceProto(c, s)
		if err != nil {
			return err
		}
		out = &engine.GetSourceWorkspaceResponse{Source: sp}
		for _, n := range notes {
			out.Notes = append(out.Notes, noteProto(n))
		}
		for _, e := range meta {
			out.Metadata = append(out.Metadata, metadataEntryProto(c, e))
		}
		out.Artifacts = arts
		out.Credibility = cred
		for _, t := range types {
			out.Types = append(out.Types, sourceTypeProto(t))
		}
		for _, g := range grades {
			out.Grades = append(out.Grades, credibilityGradeProto(g))
		}
		for _, f := range fields {
			out.Fields = append(out.Fields, metadataFieldProto(f))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func CreateSource(in []byte) ([]byte, error) {
	var req engine.CreateSourceRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("create_source", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	typeID, err := parseID(req.GetSourceTypeId())
	if err != nil {
		return nil, err
	}
	var out *engine.CreateSourceResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		s, err := sources.Create(c, userID, sources.CreateInput{
			SourceTypeID: typeID,
			Title:        req.GetTitle(),
			Description:  req.GetDescription(),
		})
		if err != nil {
			return err
		}
		sp, err := enrichSourceProto(c, s)
		if err != nil {
			return err
		}
		out = &engine.CreateSourceResponse{Source: sp}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func UpdateSource(in []byte) ([]byte, error) {
	var req engine.UpdateSourceRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("update_source", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	id, err := parseID(req.GetSourceId())
	if err != nil {
		return nil, err
	}
	typeID, err := parseID(req.GetSourceTypeId())
	if err != nil {
		return nil, err
	}
	var out *engine.UpdateSourceResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		prev, err := sources.Get(c, id)
		if err != nil {
			return err
		}
		s := sources.Source{
			ID:           id,
			Ref:          prev.Ref,
			SourceTypeID: typeID,
			Title:        req.GetTitle(),
			Description:  req.GetDescription(),
		}
		if err := sources.Update(c, userID, s); err != nil {
			return err
		}
		got, err := sources.Get(c, id)
		if err != nil {
			return err
		}
		sp, err := enrichSourceProto(c, got)
		if err != nil {
			return err
		}
		out = &engine.UpdateSourceResponse{Source: sp}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func SetSourceCover(in []byte) ([]byte, error) {
	var req engine.SetSourceCoverRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("set_source_cover", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	sourceID, err := parseID(req.GetSourceId())
	if err != nil {
		return nil, err
	}
	var primaryID []byte
	if req.GetPrimaryArtifactId() != "" {
		primaryID, err = parseID(req.GetPrimaryArtifactId())
		if err != nil {
			return nil, err
		}
	}
	var out *engine.SetSourceCoverResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		got, err := sources.SetCover(c, userID, sourceID, req.GetCoverMode(), primaryID)
		if err != nil {
			return err
		}
		sp, err := enrichSourceProto(c, got)
		if err != nil {
			return err
		}
		out = &engine.SetSourceCoverResponse{Source: sp}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func AddSourceNote(in []byte) ([]byte, error) {
	var req engine.AddSourceNoteRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("add_source_note", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	sourceID, err := parseID(req.GetSourceId())
	if err != nil {
		return nil, err
	}
	var out *engine.AddSourceNoteResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		n, err := sources.AddNote(c, userID, sourceID, req.GetBody())
		if err != nil {
			return err
		}
		out = &engine.AddSourceNoteResponse{Note: noteProto(n)}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func UpdateSourceNote(in []byte) ([]byte, error) {
	var req engine.UpdateSourceNoteRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("update_source_note", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	noteID, err := parseID(req.GetNoteId())
	if err != nil {
		return nil, err
	}
	var out *engine.UpdateSourceNoteResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		if err := sources.UpdateNote(c, userID, noteID, req.GetBody()); err != nil {
			return err
		}
		n, err := sources.GetNote(c, noteID)
		if err != nil {
			return err
		}
		out = &engine.UpdateSourceNoteResponse{Note: noteProto(n)}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func DeleteSourceNote(in []byte) ([]byte, error) {
	var req engine.DeleteSourceNoteRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("delete_source_note", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	noteID, err := parseID(req.GetNoteId())
	if err != nil {
		return nil, err
	}
	var out *engine.DeleteSourceNoteResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		if err := sources.DeleteNote(c, userID, noteID); err != nil {
			return err
		}
		out = &engine.DeleteSourceNoteResponse{}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func SetSourceMetadata(in []byte) ([]byte, error) {
	var req engine.SetSourceMetadataRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("set_source_metadata", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	sourceID, err := parseID(req.GetSourceId())
	if err != nil {
		return nil, err
	}
	fieldID, err := parseID(req.GetFieldId())
	if err != nil {
		return nil, err
	}
	var out *engine.SetSourceMetadataResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		var dateID []byte
		if d := req.GetDate(); d != nil && d.GetKind() != "" {
			v := dateValueFromProto(d)
			var err error
			dateID, err = datevalues.Insert(c, v)
			if err != nil {
				return err
			}
		} else {
			// Text-only updates keep any existing structured DateValue.
			existing, listErr := sourcemetadata.ListBySource(c, sourceID)
			if listErr != nil {
				return listErr
			}
			for _, row := range existing {
				if bytes.Equal(row.FieldID, fieldID) {
					dateID = row.DateValueID
					break
				}
			}
		}
		if _, err := sourcemetadata.Set(c, userID, sourcemetadata.Input{
			SourceID:    sourceID,
			FieldID:     fieldID,
			ValueText:   req.GetValueText(),
			DateValueID: dateID,
		}); err != nil {
			return err
		}
		entries, err := sourcemetadata.ListWorkspace(c, sourceID)
		if err != nil {
			return err
		}
		for _, e := range entries {
			if bytes.Equal(e.Field.ID, fieldID) {
				out = &engine.SetSourceMetadataResponse{
					Entry: metadataEntryProto(c, e),
				}
				return nil
			}
		}
		// Set succeeded, so the row is always in the workspace list.
		return apperr.New(apperr.CodeInternalUnknown, apperr.KindInternal)
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func ClearSourceMetadata(in []byte) ([]byte, error) {
	var req engine.ClearSourceMetadataRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("clear_source_metadata", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	sourceID, err := parseID(req.GetSourceId())
	if err != nil {
		return nil, err
	}
	fieldID, err := parseID(req.GetFieldId())
	if err != nil {
		return nil, err
	}
	var out *engine.ClearSourceMetadataResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		if err := sourcemetadata.Clear(c, userID, sourceID, fieldID); err != nil {
			return err
		}
		out = &engine.ClearSourceMetadataResponse{}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func DismissSourceMetadataSuggestion(in []byte) ([]byte, error) {
	var req engine.DismissSourceMetadataSuggestionRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("dismiss_source_metadata_suggestion", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	sourceID, err := parseID(req.GetSourceId())
	if err != nil {
		return nil, err
	}
	fieldID, err := parseID(req.GetFieldId())
	if err != nil {
		return nil, err
	}
	var out *engine.DismissSourceMetadataSuggestionResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		if err := sourcemetadata.DismissSuggestion(c, userID, sourceID, fieldID); err != nil {
			return err
		}
		entries, err := sourcemetadata.ListWorkspace(c, sourceID)
		if err != nil {
			return err
		}
		out = &engine.DismissSourceMetadataSuggestionResponse{}
		for _, e := range entries {
			out.Metadata = append(out.Metadata, metadataEntryProto(c, e))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func ReorderSourceMetadata(in []byte) ([]byte, error) {
	var req engine.ReorderSourceMetadataRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("reorder_source_metadata", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	sourceID, err := parseID(req.GetSourceId())
	if err != nil {
		return nil, err
	}
	fieldIDs := make([][]byte, 0, len(req.GetFieldIds()))
	for _, id := range req.GetFieldIds() {
		fid, err := parseID(id)
		if err != nil {
			return nil, err
		}
		fieldIDs = append(fieldIDs, fid)
	}
	var out *engine.ReorderSourceMetadataResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		if err := sourcemetadata.Reorder(c, userID, sourceID, fieldIDs); err != nil {
			return err
		}
		entries, err := sourcemetadata.ListWorkspace(c, sourceID)
		if err != nil {
			return err
		}
		out = &engine.ReorderSourceMetadataResponse{}
		for _, e := range entries {
			out.Metadata = append(out.Metadata, metadataEntryProto(c, e))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func sourceProto(s sources.Source) *engine.Source {
	return &engine.Source{
		Id:                uuidString(s.ID),
		Ref:               s.Ref,
		SourceTypeId:      uuidString(s.SourceTypeID),
		Title:             s.Title,
		Description:       s.Description,
		CoverMode:         s.CoverMode,
		PrimaryArtifactId: uuidString(s.PrimaryArtifactID),
	}
}

func noteProto(n sources.Note) *engine.SourceNote {
	return &engine.SourceNote{
		Id:                uuidString(n.ID),
		SourceId:          uuidString(n.SourceID),
		Body:              n.Body,
		AuthorDisplayName: n.AuthorDisplayName,
		CreatedAt:         n.CreatedAt,
	}
}

func metadataEntryProto(c *database.Catalog, e sourcemetadata.WorkspaceEntry) *engine.MetadataWorkspaceEntry {
	out := &engine.MetadataWorkspaceEntry{
		Field: &engine.MetadataField{
			Id:          uuidString(e.Field.ID),
			Key:         e.Field.Key,
			Origin:      e.Field.Origin,
			Label:       e.Field.Label,
			DataType:    e.Field.DataType,
			Description: e.Field.Description,
		},
		Suggested: e.Suggested,
		SortOrder: int32(e.SortOrder),
	}
	if e.Value != nil {
		out.HasValue = true
		out.ValueText = e.Value.ValueText
		out.DateValueId = uuidString(e.Value.DateValueID)
		if len(e.Value.DateValueID) == 16 {
			if dv, err := datevalues.Lookup(c, e.Value.DateValueID); err == nil {
				out.Date = dateValueProto(dv)
			}
		}
	}
	return out
}

func fileRefProto(f files.File, relPath string) *engine.SourceFileRef {
	return &engine.SourceFileRef{
		Id:               uuidString(f.ID),
		RelPath:          relPath,
		OriginalFilename: f.OriginalFilename,
		MediaType:        f.MediaType,
		ByteSize:         f.ByteSize,
	}
}

// dateValueProto is the inverse of dateValueFromProto: it carries a stored
// DateValue's components back to the client (MetadataWorkspaceEntry.date).
func dateValueProto(v datevalues.Value) *engine.DateValueInput {
	d := &engine.DateValueInput{
		Kind:      v.Kind,
		Qualifier: v.Qualifier,
		Calendar:  v.Calendar,
		StartTz:   v.StartTZ,
		EndTz:     v.EndTZ,
		Phrase:    v.Phrase,
	}
	toInt32 := func(p *int) *int32 {
		if p == nil {
			return nil
		}
		n := int32(*p)
		return &n
	}
	d.StartYear = toInt32(v.StartYear)
	d.StartMonth = toInt32(v.StartMonth)
	d.StartDay = toInt32(v.StartDay)
	d.StartHour = toInt32(v.StartHour)
	d.StartMinute = toInt32(v.StartMinute)
	d.StartSecond = toInt32(v.StartSecond)
	d.StartMillisecond = toInt32(v.StartMillisecond)
	d.EndYear = toInt32(v.EndYear)
	d.EndMonth = toInt32(v.EndMonth)
	d.EndDay = toInt32(v.EndDay)
	d.EndHour = toInt32(v.EndHour)
	d.EndMinute = toInt32(v.EndMinute)
	d.EndSecond = toInt32(v.EndSecond)
	d.EndMillisecond = toInt32(v.EndMillisecond)
	return d
}

func dateValueFromProto(d *engine.DateValueInput) datevalues.Value {
	v := datevalues.Value{
		Kind:      d.GetKind(),
		Qualifier: d.GetQualifier(),
		Calendar:  d.GetCalendar(),
		StartTZ:   d.GetStartTz(),
		EndTZ:     d.GetEndTz(),
		Phrase:    d.GetPhrase(),
	}
	if d.StartYear != nil {
		y := int(d.GetStartYear())
		v.StartYear = &y
	}
	if d.StartMonth != nil {
		m := int(d.GetStartMonth())
		v.StartMonth = &m
	}
	if d.StartDay != nil {
		day := int(d.GetStartDay())
		v.StartDay = &day
	}
	if d.StartHour != nil {
		h := int(d.GetStartHour())
		v.StartHour = &h
	}
	if d.StartMinute != nil {
		m := int(d.GetStartMinute())
		v.StartMinute = &m
	}
	if d.StartSecond != nil {
		s := int(d.GetStartSecond())
		v.StartSecond = &s
	}
	if d.StartMillisecond != nil {
		ms := int(d.GetStartMillisecond())
		v.StartMillisecond = &ms
	}
	if d.EndYear != nil {
		y := int(d.GetEndYear())
		v.EndYear = &y
	}
	if d.EndMonth != nil {
		m := int(d.GetEndMonth())
		v.EndMonth = &m
	}
	if d.EndDay != nil {
		day := int(d.GetEndDay())
		v.EndDay = &day
	}
	if d.EndHour != nil {
		h := int(d.GetEndHour())
		v.EndHour = &h
	}
	if d.EndMinute != nil {
		m := int(d.GetEndMinute())
		v.EndMinute = &m
	}
	if d.EndSecond != nil {
		s := int(d.GetEndSecond())
		v.EndSecond = &s
	}
	if d.EndMillisecond != nil {
		ms := int(d.GetEndMillisecond())
		v.EndMillisecond = &ms
	}
	return v
}
