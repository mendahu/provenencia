package handlers

import (
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/writes"
	"google.golang.org/protobuf/proto"
)

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
		got, wrote, err := writes.Run(c, writes.Op{Action: "create_property", UserID: userID}, func(tx *database.Tx) (properties.Property, []rowchange.Change, error) {
			return properties.Create(tx, userID, req.GetLabel(), req.GetValueType(), req.GetDescription(), req.GetCardinality())
		})
		if err != nil {
			return err
		}
		out = &engine.CreatePropertyResponse{Property: propertyProto(got), Effects: writeEffectsProto(wrote)}
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
		got, wrote, err := writes.Run(c, writes.Op{Action: "update_property", UserID: userID}, func(tx *database.Tx) (properties.Property, []rowchange.Change, error) {
			return properties.Update(tx, userID, propertyID, req.GetLabel(), req.GetValueType(), req.GetDescription(), req.GetCardinality())
		})
		if err != nil {
			return err
		}
		out = &engine.UpdatePropertyResponse{Property: propertyProto(got), Effects: writeEffectsProto(wrote)}
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
		_, wrote, err := writes.Run(c, writes.Op{Action: "delete_property", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := properties.Delete(tx, userID, propertyID)
			return struct{}{}, changes, err
		})
		if err != nil {
			return err
		}
		out = &engine.DeletePropertyResponse{Effects: writeEffectsProto(wrote)}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func subjectTypePropertyProto(b subjectvocab.Binding) *engine.SubjectTypeProperty {
	return &engine.SubjectTypeProperty{
		Property:  propertyProto(b.Property),
		SortOrder: int32(b.SortOrder),
		Locked:    b.Locked,
	}
}

func ReorderSubjectTypeProperties(in []byte) ([]byte, error) {
	var req engine.ReorderSubjectTypePropertiesRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("reorder_subject_type_properties", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	typeID, err := parseID(req.GetSubjectTypeId())
	if err != nil {
		return nil, err
	}
	propertyIDs, err := parseIDs(req.GetPropertyIds())
	if err != nil {
		return nil, err
	}
	var out *engine.ReorderSubjectTypePropertiesResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		_, wrote, err := writes.Run(c, writes.Op{Action: subjectvocab.ActionReorder, UserID: userID},
			func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
				changes, err := subjectvocab.ReorderBindings(tx, userID, typeID, propertyIDs)
				return struct{}{}, changes, err
			})
		if err != nil {
			return err
		}
		rows, err := subjectvocab.ListBindings(c, typeID)
		if err != nil {
			return err
		}
		out = &engine.ReorderSubjectTypePropertiesResponse{Effects: writeEffectsProto(wrote)}
		for _, b := range rows {
			out.Properties = append(out.Properties, subjectTypePropertyProto(b))
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
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
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
		_, wrote, err := writes.Run(c, writes.Op{Action: subjectvocab.ActionAssign, UserID: userID},
			func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
				changes, err := subjectvocab.AppendBinding(tx, userID, typeID, propertyID)
				return struct{}{}, changes, err
			})
		if err != nil {
			return err
		}
		out = &engine.AssignSubjectTypePropertyResponse{Effects: writeEffectsProto(wrote)}
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
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
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
		_, wrote, err := writes.Run(c, writes.Op{Action: subjectvocab.ActionRemove, UserID: userID},
			func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
				changes, err := subjectvocab.DeleteBinding(tx, userID, typeID, propertyID)
				return struct{}{}, changes, err
			})
		if err != nil {
			return err
		}
		out = &engine.RemoveSubjectTypePropertyResponse{Effects: writeEffectsProto(wrote)}
		return nil
	})
	if err != nil {
		return nil, err
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
		for _, r := range propertymeta.Types {
			out.ValueTypes = append(out.ValueTypes, &engine.PropertyValueTypeRule{
				ValueType:           r.ValueType,
				ResearcherCreatable: r.ResearcherCreatable,
				SingleOnly:          r.SingleOnly,
				Negatable:           r.Negatable,
			})
		}
		for _, p := range props {
			out.Properties = append(out.Properties, propertyProto(p))
		}
		for _, typ := range types {
			out.Types = append(out.Types, subjectTypeProto(typ))
			bindings, err := subjectvocab.ListBindings(c, typ.ID)
			if err != nil {
				return err
			}
			group := &engine.SubjectTypePropertiesGroup{SubjectTypeId: database.UUIDString(typ.ID)}
			for _, b := range bindings {
				group.Properties = append(group.Properties, subjectTypePropertyProto(b))
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
		Id:          database.UUIDString(p.ID),
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
