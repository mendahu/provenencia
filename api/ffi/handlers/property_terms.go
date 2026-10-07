package handlers

import (
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"google.golang.org/protobuf/proto"
)

func ListPropertyTerms(in []byte) ([]byte, error) {
	var req engine.ListPropertyTermsRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("list_property_terms", err)
	}
	propertyID, err := parseID(req.GetPropertyId())
	if err != nil {
		return nil, err
	}
	var out *engine.ListPropertyTermsResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		rows, err := propertyterms.ListByProperty(c, propertyID)
		if err != nil {
			return err
		}
		out = &engine.ListPropertyTermsResponse{}
		for _, t := range rows {
			out.Terms = append(out.Terms, propertyTermProto(t))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func CreatePropertyTerm(in []byte) ([]byte, error) {
	var req engine.CreatePropertyTermRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("create_property_term", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	propertyID, err := parseID(req.GetPropertyId())
	if err != nil {
		return nil, err
	}
	var out *engine.CreatePropertyTermResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		got, err := propertyterms.Create(c, userID, propertyID, req.GetLabel(), req.GetDescription())
		if err != nil {
			return err
		}
		out = &engine.CreatePropertyTermResponse{Term: propertyTermProto(got)}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func UpdatePropertyTerm(in []byte) ([]byte, error) {
	var req engine.UpdatePropertyTermRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("update_property_term", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	termID, err := parseID(req.GetTermId())
	if err != nil {
		return nil, err
	}
	var out *engine.UpdatePropertyTermResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		got, err := propertyterms.Update(c, userID, termID, req.GetLabel(), req.GetDescription())
		if err != nil {
			return err
		}
		out = &engine.UpdatePropertyTermResponse{Term: propertyTermProto(got)}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func DeletePropertyTerm(in []byte) ([]byte, error) {
	var req engine.DeletePropertyTermRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("delete_property_term", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	termID, err := parseID(req.GetTermId())
	if err != nil {
		return nil, err
	}
	var out *engine.DeletePropertyTermResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		if err := propertyterms.Delete(c, userID, termID); err != nil {
			return err
		}
		out = &engine.DeletePropertyTermResponse{}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func propertyTermProto(t propertyterms.Term) *engine.PropertyTerm {
	return &engine.PropertyTerm{
		Id:          uuidString(t.ID),
		PropertyId:  uuidString(t.PropertyID),
		Key:         t.Key,
		Origin:      t.Origin,
		Label:       t.Label,
		Description: t.Description,
	}
}
