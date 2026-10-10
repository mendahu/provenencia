package handlers

import (
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/writes"
	"google.golang.org/protobuf/proto"
)

func ListProperties(in []byte) ([]byte, error) {
	var req engine.ListPropertiesRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("list_properties", err)
	}
	var out *engine.ListPropertiesResponse
	err := withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		rows, err := properties.List(c)
		if err != nil {
			return err
		}
		out = &engine.ListPropertiesResponse{}
		for _, p := range rows {
			out.Properties = append(out.Properties, propertyProto(p))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func CreateProperty(in []byte) ([]byte, error) {
	var req engine.CreatePropertyRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("create_property", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	var out *engine.CreatePropertyResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		got, _, err := writes.Run(c, writes.Op{Action: "create_property", UserID: userID}, func(tx *database.Tx) (properties.Property, []rowchange.Change, error) {
			return properties.Create(tx, userID, req.GetLabel(), req.GetValueType(), req.GetDescription(), req.GetCardinality())
		})
		if err != nil {
			return err
		}
		out = &engine.CreatePropertyResponse{Property: propertyProto(got)}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func UpdateProperty(in []byte) ([]byte, error) {
	var req engine.UpdatePropertyRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("update_property", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	propertyID, err := parseID(req.GetPropertyId())
	if err != nil {
		return nil, err
	}
	var out *engine.UpdatePropertyResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		got, _, err := writes.Run(c, writes.Op{Action: "update_property", UserID: userID}, func(tx *database.Tx) (properties.Property, []rowchange.Change, error) {
			return properties.Update(tx, userID, propertyID, req.GetLabel(), req.GetValueType(), req.GetDescription(), req.GetCardinality())
		})
		if err != nil {
			return err
		}
		out = &engine.UpdatePropertyResponse{Property: propertyProto(got)}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func DeleteProperty(in []byte) ([]byte, error) {
	var req engine.DeletePropertyRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("delete_property", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	propertyID, err := parseID(req.GetPropertyId())
	if err != nil {
		return nil, err
	}
	var out *engine.DeletePropertyResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		if _, _, err := writes.Run(c, writes.Op{Action: "delete_property", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := properties.Delete(tx, userID, propertyID)
			return struct{}{}, changes, err
		}); err != nil {
			return err
		}
		out = &engine.DeletePropertyResponse{}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func ListSubjectTypeProperties(in []byte) ([]byte, error) {
	var req engine.ListSubjectTypePropertiesRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("list_subject_type_properties", err)
	}
	typeID, err := parseID(req.GetSubjectTypeId())
	if err != nil {
		return nil, err
	}
	var out *engine.ListSubjectTypePropertiesResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		rows, err := subjectvocab.ListBindings(c, typeID)
		if err != nil {
			return err
		}
		out = &engine.ListSubjectTypePropertiesResponse{}
		for _, b := range rows {
			out.Properties = append(out.Properties, &engine.SubjectTypeProperty{
				Property:  propertyProto(b.Property),
				SortOrder: int32(b.SortOrder),
				Locked:    b.Locked,
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func AssignSubjectTypeProperty(in []byte) ([]byte, error) {
	var req engine.AssignSubjectTypePropertyRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("assign_subject_type_property", err)
	}
	if _, err := parseUserID(req.GetUserId()); err != nil {
		return nil, err
	}
	typeID, err := parseID(req.GetSubjectTypeId())
	if err != nil {
		return nil, err
	}
	propertyID, err := parseID(req.GetPropertyId())
	if err != nil {
		return nil, err
	}
	var out *engine.AssignSubjectTypePropertyResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		if err := subjectvocab.AppendBinding(c, typeID, propertyID); err != nil {
			return err
		}
		out = &engine.AssignSubjectTypePropertyResponse{}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func RemoveSubjectTypeProperty(in []byte) ([]byte, error) {
	var req engine.RemoveSubjectTypePropertyRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("remove_subject_type_property", err)
	}
	if _, err := parseUserID(req.GetUserId()); err != nil {
		return nil, err
	}
	typeID, err := parseID(req.GetSubjectTypeId())
	if err != nil {
		return nil, err
	}
	propertyID, err := parseID(req.GetPropertyId())
	if err != nil {
		return nil, err
	}
	var out *engine.RemoveSubjectTypePropertyResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		if err := subjectvocab.DeleteBinding(c, typeID, propertyID); err != nil {
			return err
		}
		out = &engine.RemoveSubjectTypePropertyResponse{}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func ListPlaceableSubjectTypes(in []byte) ([]byte, error) {
	var req engine.ListPlaceableSubjectTypesRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("list_placeable_subject_types", err)
	}
	out := &engine.ListPlaceableSubjectTypesResponse{}
	for _, t := range subjectvocab.PlaceableTypes() {
		out.Types = append(out.Types, subjectTypePresentationProto(t))
	}
	return proto.Marshal(out)
}

func GetSubjectTypePresentation(in []byte) ([]byte, error) {
	var req engine.GetSubjectTypePresentationRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("get_subject_type_presentation", err)
	}
	info, ok := subjectvocab.TypeByKey(req.GetTypeKey())
	if !ok {
		return nil, apperr.New(apperr.CodeSubjectVocabInvalid, apperr.KindUser)
	}
	out := &engine.GetSubjectTypePresentationResponse{
		Presentation: subjectTypePresentationProto(info),
	}
	return proto.Marshal(out)
}

func ListConnectRules(in []byte) ([]byte, error) {
	var req engine.ListConnectRulesRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("list_connect_rules", err)
	}
	out := &engine.ListConnectRulesResponse{}
	for _, r := range subjectvocab.ListConnectRules() {
		rule := &engine.ConnectRule{
			FromTypeKey:      r.FromTypeKey,
			ToTypeKey:        r.ToTypeKey,
			BridgeTypeKey:    r.BridgeTypeKey,
			EdgePropertyKeys: append([]string(nil), r.EdgePropertyKeys...),
			Disambiguation:   r.Disambiguation,
			Refuse:           r.Refuse,
		}
		for _, e := range r.Edges {
			rule.Edges = append(rule.Edges, &engine.ConnectEdge{
				PropertyKey:     e.PropertyKey,
				EndpointTypeKey: e.EndpointTypeKey,
			})
		}
		out.Rules = append(out.Rules, rule)
	}
	return proto.Marshal(out)
}

func GetPropertiesWorkspace(in []byte) ([]byte, error) {
	var req engine.GetPropertiesWorkspaceRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("get_properties_workspace", err)
	}
	var out *engine.GetPropertiesWorkspaceResponse
	err := withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		props, err := properties.List(c)
		if err != nil {
			return err
		}
		types, err := subjecttypes.List(c)
		if err != nil {
			return err
		}
		out = &engine.GetPropertiesWorkspaceResponse{}
		for _, p := range props {
			out.Properties = append(out.Properties, propertyProto(p))
		}
		for _, typ := range types {
			out.Types = append(out.Types, subjectTypeProto(typ))
			bindings, err := subjectvocab.ListBindings(c, typ.ID)
			if err != nil {
				return err
			}
			group := &engine.SubjectTypePropertiesGroup{SubjectTypeId: uuidString(typ.ID)}
			for _, b := range bindings {
				group.Properties = append(group.Properties, &engine.SubjectTypeProperty{
					Property:  propertyProto(b.Property),
					SortOrder: int32(b.SortOrder),
					Locked:    b.Locked,
				})
			}
			if info, ok := subjectvocab.TypeByKey(typ.Key); ok {
				group.Presentation = subjectTypePresentationProto(info)
			}
			out.Groups = append(out.Groups, group)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func propertyProto(p properties.Property) *engine.Property {
	return &engine.Property{
		Id:          uuidString(p.ID),
		Key:         p.Key,
		Origin:      p.Origin,
		Label:       p.Label,
		Description: p.Description,
		ValueType:   p.ValueType,
		UsedBy:      int32(p.UsedBy),
		Cardinality: p.Cardinality,
	}
}

func subjectTypePresentationProto(t subjectvocab.TypeInfo) *engine.SubjectTypePresentation {
	return &engine.SubjectTypePresentation{
		TypeKey:                  t.Key,
		L10NKey:                  t.Presentation.L10nKey,
		IconSymbol:               t.Presentation.IconSymbol,
		InkToken:                 t.Presentation.InkToken,
		TintToken:                t.Presentation.TintToken,
		ChipToken:                t.Presentation.ChipToken,
		LineToken:                t.Presentation.LineToken,
		EdgeFromToken:            t.Presentation.EdgeFromToken,
		EdgeToToken:              t.Presentation.EdgeToToken,
		Role:                     t.Role,
		Placeable:                t.Placeable,
		PaletteSort:              int32(t.PaletteSort),
		RequiresCitationAtCreate: t.RequiresCitationAtCreate,
		Label:                    t.Label,
	}
}
