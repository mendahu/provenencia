package handlers

import (
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/sourcefields"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/sourcevocab"
	"google.golang.org/protobuf/proto"
)

func ListSourceTypes(in []byte) ([]byte, error) {
	var req engine.ListSourceTypesRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("list_source_types", err)
	}
	var out *engine.ListSourceTypesResponse
	err := withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		rows, err := sourcetypes.List(c)
		if err != nil {
			return err
		}
		out = &engine.ListSourceTypesResponse{}
		for _, t := range rows {
			out.Types = append(out.Types, sourceTypeProto(t))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func CreateSourceType(in []byte) ([]byte, error) {
	var req engine.CreateSourceTypeRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("create_source_type", err)
	}
	if _, err := parseUserID(req.GetUserId()); err != nil {
		return nil, err
	}
	var out *engine.CreateSourceTypeResponse
	err := withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		got, err := sourcetypes.Create(c, req.GetLabel(), req.GetDescription(), req.GetIconKey())
		if err != nil {
			return err
		}
		out = &engine.CreateSourceTypeResponse{Type: sourceTypeProto(got)}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func UpdateSourceType(in []byte) ([]byte, error) {
	var req engine.UpdateSourceTypeRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("update_source_type", err)
	}
	if _, err := parseUserID(req.GetUserId()); err != nil {
		return nil, err
	}
	typeID, err := parseID(req.GetTypeId())
	if err != nil {
		return nil, err
	}
	var out *engine.UpdateSourceTypeResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		got, err := sourcetypes.Update(c, typeID, req.GetLabel(), req.GetDescription(), req.GetIconKey())
		if err != nil {
			return err
		}
		if got.UsedBy, err = sourcetypes.UsedBy(c, typeID); err != nil {
			return err
		}
		if got.SuggestedFields, err = sourcevocab.CountSuggestions(c, typeID); err != nil {
			return err
		}
		out = &engine.UpdateSourceTypeResponse{Type: sourceTypeProto(got)}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func ListTypeSuggestions(in []byte) ([]byte, error) {
	var req engine.ListTypeSuggestionsRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("list_type_suggestions", err)
	}
	typeID, err := parseID(req.GetTypeId())
	if err != nil {
		return nil, err
	}
	var out *engine.ListTypeSuggestionsResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		suggestions, err := suggestionsProto(c, typeID)
		if err != nil {
			return err
		}
		out = &engine.ListTypeSuggestionsResponse{Suggestions: suggestions}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func AssignTypeField(in []byte) ([]byte, error) {
	var req engine.AssignTypeFieldRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("assign_type_field", err)
	}
	typeID, fieldID, err := parseSuggestionJoin(req.GetUserId(), req.GetTypeId(), req.GetFieldId())
	if err != nil {
		return nil, err
	}
	var out *engine.AssignTypeFieldResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		if err := sourcevocab.AppendSuggestion(c, typeID, fieldID); err != nil {
			return err
		}
		suggestions, err := suggestionsProto(c, typeID)
		if err != nil {
			return err
		}
		out = &engine.AssignTypeFieldResponse{Suggestions: suggestions}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func RemoveTypeField(in []byte) ([]byte, error) {
	var req engine.RemoveTypeFieldRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("remove_type_field", err)
	}
	typeID, fieldID, err := parseSuggestionJoin(req.GetUserId(), req.GetTypeId(), req.GetFieldId())
	if err != nil {
		return nil, err
	}
	var out *engine.RemoveTypeFieldResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		if err := sourcevocab.DeleteSuggestion(c, typeID, fieldID); err != nil {
			return err
		}
		suggestions, err := suggestionsProto(c, typeID)
		if err != nil {
			return err
		}
		out = &engine.RemoveTypeFieldResponse{Suggestions: suggestions}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

// parseSuggestionJoin validates the user and both join sides — the shared
// preamble of assign and remove.
func parseSuggestionJoin(userID, typeID, fieldID string) ([]byte, []byte, error) {
	if _, err := parseUserID(userID); err != nil {
		return nil, nil, err
	}
	tid, err := parseID(typeID)
	if err != nil {
		return nil, nil, err
	}
	fid, err := parseID(fieldID)
	if err != nil {
		return nil, nil, err
	}
	return tid, fid, nil
}

func suggestionsProto(c *database.Catalog, typeID []byte) ([]*engine.TypeSuggestion, error) {
	rows, err := sourcevocab.ListSuggestions(c, typeID)
	if err != nil {
		return nil, err
	}
	out := make([]*engine.TypeSuggestion, 0, len(rows))
	for _, s := range rows {
		out = append(out, &engine.TypeSuggestion{
			Field:     metadataFieldProto(s.Field),
			SortOrder: int32(s.SortOrder),
		})
	}
	return out, nil
}

func ListMetadataFields(in []byte) ([]byte, error) {
	var req engine.ListMetadataFieldsRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("list_metadata_fields", err)
	}
	var out *engine.ListMetadataFieldsResponse
	err := withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		rows, err := sourcefields.List(c)
		if err != nil {
			return err
		}
		out = &engine.ListMetadataFieldsResponse{}
		for _, f := range rows {
			out.Fields = append(out.Fields, metadataFieldProto(f))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func CreateMetadataField(in []byte) ([]byte, error) {
	var req engine.CreateMetadataFieldRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("create_metadata_field", err)
	}
	if _, err := parseUserID(req.GetUserId()); err != nil {
		return nil, err
	}
	var out *engine.CreateMetadataFieldResponse
	err := withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		got, err := sourcefields.Create(c, req.GetLabel(), req.GetDataType(), req.GetDescription())
		if err != nil {
			return err
		}
		out = &engine.CreateMetadataFieldResponse{Field: metadataFieldProto(got)}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func UpdateMetadataField(in []byte) ([]byte, error) {
	var req engine.UpdateMetadataFieldRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("update_metadata_field", err)
	}
	if _, err := parseUserID(req.GetUserId()); err != nil {
		return nil, err
	}
	fieldID, err := parseID(req.GetFieldId())
	if err != nil {
		return nil, err
	}
	var out *engine.UpdateMetadataFieldResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		got, err := sourcefields.Update(c, fieldID, req.GetLabel(), req.GetDataType(), req.GetDescription())
		if err != nil {
			return err
		}
		if got.UsedBy, err = sourcefields.UsedBy(c, fieldID); err != nil {
			return err
		}
		out = &engine.UpdateMetadataFieldResponse{Field: metadataFieldProto(got)}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func DeleteSourceType(in []byte) ([]byte, error) {
	var req engine.DeleteSourceTypeRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("delete_source_type", err)
	}
	if _, err := parseUserID(req.GetUserId()); err != nil {
		return nil, err
	}
	typeID, err := parseID(req.GetTypeId())
	if err != nil {
		return nil, err
	}
	var out *engine.DeleteSourceTypeResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		if err := sourcetypes.Delete(c, typeID); err != nil {
			return err
		}
		out = &engine.DeleteSourceTypeResponse{}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func DeleteMetadataField(in []byte) ([]byte, error) {
	var req engine.DeleteMetadataFieldRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("delete_metadata_field", err)
	}
	if _, err := parseUserID(req.GetUserId()); err != nil {
		return nil, err
	}
	fieldID, err := parseID(req.GetFieldId())
	if err != nil {
		return nil, err
	}
	var out *engine.DeleteMetadataFieldResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		if err := sourcefields.Delete(c, fieldID); err != nil {
			return err
		}
		out = &engine.DeleteMetadataFieldResponse{}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func sourceTypeProto(t sourcetypes.Type) *engine.SourceType {
	return &engine.SourceType{
		Id:                  uuidString(t.ID),
		Key:                 t.Key,
		Origin:              t.Origin,
		Label:               t.Label,
		Description:         t.Description,
		UsedBy:              int32(t.UsedBy),
		SuggestedFieldCount: int32(t.SuggestedFields),
		IconKey:             t.IconKey,
	}
}

func metadataFieldProto(f sourcefields.Field) *engine.MetadataField {
	return &engine.MetadataField{
		Id:          uuidString(f.ID),
		Key:         f.Key,
		Origin:      f.Origin,
		Label:       f.Label,
		DataType:    f.DataType,
		Description: f.Description,
		UsedBy:      int32(f.UsedBy),
	}
}
