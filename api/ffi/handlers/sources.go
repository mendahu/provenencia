package handlers

import (
	"bytes"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/files"
	"github.com/mendahu/provenencia/core/database/sourcemetadata"
	"github.com/mendahu/provenencia/core/database/sources"
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
		covers, err := sourceCoverThumbnails(c, rows)
		if err != nil {
			return err
		}
		for _, s := range rows {
			sp := sourceProto(s)
			sp.ThumbnailRelPath = covers[string(s.ID)]
			sp.HasArtifact = s.HasArtifact
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

func DeleteSource(in []byte) ([]byte, error) {
	var req engine.DeleteSourceRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("delete_source", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	sourceID, err := parseID(req.GetSourceId())
	if err != nil {
		return nil, err
	}
	var out *engine.DeleteSourceResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		if err := sources.Delete(c, userID, sourceID); err != nil {
			return err
		}
		out = &engine.DeleteSourceResponse{}
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
		if _, err := sourcemetadata.Set(c, userID, sourcemetadata.Input{
			SourceID:  sourceID,
			FieldID:   fieldID,
			ValueText: req.GetValueText(),
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
		UpdatedRevision:   s.UpdatedRevision,
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

func metadataEntryProto(_ *database.Catalog, e sourcemetadata.WorkspaceEntry) *engine.MetadataWorkspaceEntry {
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
