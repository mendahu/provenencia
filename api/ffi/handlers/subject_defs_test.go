package handlers

import (
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
	"google.golang.org/protobuf/proto"
)

func TestListProperties(t *testing.T) {
	runRPC(t, ListProperties, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "lists seeded properties",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, _ := sourceFixture(t)
				return &engine.ListPropertiesRequest{ProjectDir: dir}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var resp engine.ListPropertiesResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				if len(resp.Properties) != 15 {
					t.Fatalf("len=%d want 15", len(resp.Properties))
				}
				found := false
				for _, p := range resp.Properties {
					if p.GetKey() == "name" && p.GetValueType() == "name" {
						found = true
						break
					}
				}
				if !found {
					t.Fatal("missing name property")
				}
			},
		},
	})
}

func TestCreateUpdateDeleteProperty(t *testing.T) {
	runRPC(t, CreateProperty, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "create update delete round trip",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _ := sourceFixture(t)
				return &engine.CreatePropertyRequest{
					ProjectDir: dir, UserId: userID, Label: "Custom Fact", ValueType: "text",
				}
			},
			after: func(t *testing.T, out []byte, req proto.Message) {
				var created engine.CreatePropertyResponse
				if err := proto.Unmarshal(out, &created); err != nil {
					t.Fatal(err)
				}
				if created.Property.GetKey() != "custom-fact" || created.Property.GetOrigin() != "user" {
					t.Fatalf("%+v", created.Property)
				}
				cr := req.(*engine.CreatePropertyRequest)
				uout, err := UpdateProperty(marshalProto(t, &engine.UpdatePropertyRequest{
					ProjectDir: cr.ProjectDir, UserId: cr.UserId,
					PropertyId: created.Property.GetId(), Label: "Custom Fact 2", ValueType: "text",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var updated engine.UpdatePropertyResponse
				if err := proto.Unmarshal(uout, &updated); err != nil {
					t.Fatal(err)
				}
				if updated.Property.GetLabel() != "Custom Fact 2" {
					t.Fatalf("%+v", updated.Property)
				}
				if _, err := DeleteProperty(marshalProto(t, &engine.DeletePropertyRequest{
					ProjectDir: cr.ProjectDir, UserId: cr.UserId, PropertyId: created.Property.GetId(),
				})); err != nil {
					t.Fatal(err)
				}
			},
		},
	})
}

func TestSubjectTypeFieldsAndRegistry(t *testing.T) {
	runRPC(t, ListSubjectTypes, []rpcTest{
		{
			name: "bindings locked placeable connect",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, _ := sourceFixture(t)
				return &engine.ListSubjectTypesRequest{ProjectDir: dir}
			},
			after: func(t *testing.T, out []byte, req proto.Message) {
				var types engine.ListSubjectTypesResponse
				if err := proto.Unmarshal(out, &types); err != nil {
					t.Fatal(err)
				}
				var participationID string
				for _, typ := range types.Types {
					if typ.GetKey() == "participation" {
						participationID = typ.GetId()
						break
					}
				}
				if participationID == "" {
					t.Fatal("missing participation")
				}
				dir := req.(*engine.ListSubjectTypesRequest).ProjectDir
				bout, err := ListSubjectTypeFields(marshalProto(t, &engine.ListSubjectTypeFieldsRequest{
					ProjectDir: dir, SubjectTypeId: participationID,
				}))
				if err != nil {
					t.Fatal(err)
				}
				var fields engine.ListSubjectTypeFieldsResponse
				if err := proto.Unmarshal(bout, &fields); err != nil {
					t.Fatal(err)
				}
				locked := 0
				for _, f := range fields.Fields {
					if f.GetLocked() {
						locked++
					}
				}
				if locked != 2 {
					t.Fatalf("locked=%d", locked)
				}
				pout, err := ListPlaceableSubjectTypes(marshalProto(t, &engine.ListPlaceableSubjectTypesRequest{}))
				if err != nil {
					t.Fatal(err)
				}
				var placeable engine.ListPlaceableSubjectTypesResponse
				if err := proto.Unmarshal(pout, &placeable); err != nil {
					t.Fatal(err)
				}
				if len(placeable.Types) != 3 {
					t.Fatalf("placeable=%d", len(placeable.Types))
				}
				cout, err := ListConnectRules(marshalProto(t, &engine.ListConnectRulesRequest{}))
				if err != nil {
					t.Fatal(err)
				}
				var rules engine.ListConnectRulesResponse
				if err := proto.Unmarshal(cout, &rules); err != nil {
					t.Fatal(err)
				}
				if len(rules.Rules) < 3 {
					t.Fatalf("rules=%d", len(rules.Rules))
				}
			},
		},
	})
}
