package handlers

import (
	"errors"
	"slices"
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/propertymeta"
	"google.golang.org/protobuf/proto"
)

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
				if created.Property.GetKey() != "custom-fact" || created.Property.GetOrigin() != "user" || created.Property.GetCardinality() != "single" {
					t.Fatalf("%+v", created.Property)
				}
				cr := req.(*engine.CreatePropertyRequest)
				severalOut, err := CreateProperty(marshalProto(t, &engine.CreatePropertyRequest{
					ProjectDir: cr.ProjectDir, UserId: cr.UserId,
					Label: "Languages Spoken", ValueType: "text", Cardinality: "multiple",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var several engine.CreatePropertyResponse
				if err := proto.Unmarshal(severalOut, &several); err != nil {
					t.Fatal(err)
				}
				if several.Property.GetCardinality() != "multiple" {
					t.Fatalf("%+v", several.Property)
				}
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
				if updated.Property.GetLabel() != "Custom Fact 2" || updated.Property.GetCardinality() != "single" {
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

func TestGetPropertiesWorkspace(t *testing.T) {
	runRPC(t, GetPropertiesWorkspace, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "loads properties types and groups",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, _, _ := subjectFixture(t)
				return &engine.GetPropertiesWorkspaceRequest{ProjectDir: dir}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var got engine.GetPropertiesWorkspaceResponse
				if err := proto.Unmarshal(out, &got); err != nil {
					t.Fatal(err)
				}
				if len(got.Properties) == 0 || len(got.Types) == 0 || len(got.Groups) == 0 {
					t.Fatalf("empty workspace props=%d types=%d groups=%d", len(got.Properties), len(got.Types), len(got.Groups))
				}
				if len(got.Groups) != len(got.Types) {
					t.Fatalf("groups %d types %d", len(got.Groups), len(got.Types))
				}
			},
		},
		{
			name: "lists seeded properties",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, _ := sourceFixture(t)
				return &engine.GetPropertiesWorkspaceRequest{ProjectDir: dir}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var resp engine.GetPropertiesWorkspaceResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				foundName, foundToponym, foundGender, foundInterests, foundLikeness := false, false, false, false, false
				for _, p := range resp.Properties {
					if p.GetKey() == "name" && p.GetValueType() == "name" && p.GetCardinality() == "single" {
						foundName = true
					}
					if p.GetKey() == "toponym" && p.GetCardinality() == "multiple" {
						foundToponym = true
					}
					if p.GetKey() == "gender" && p.GetValueType() == "term" && p.GetCardinality() == "single" {
						foundGender = true
					}
					if p.GetKey() == "interests" && p.GetValueType() == "text" && p.GetCardinality() == "multiple" {
						foundInterests = true
					}
					if p.GetKey() == "likeness" && p.GetValueType() == "depiction" && p.GetCardinality() == "multiple" {
						foundLikeness = true
					}
				}
				if !foundName || !foundToponym || !foundGender || !foundInterests || !foundLikeness {
					t.Fatalf("name=%v toponym=%v gender=%v interests=%v likeness=%v", foundName, foundToponym, foundGender, foundInterests, foundLikeness)
				}
			},
		},
		{
			name: "locked bindings and connect rules",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, _ := sourceFixture(t)
				return &engine.GetPropertiesWorkspaceRequest{ProjectDir: dir}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var resp engine.GetPropertiesWorkspaceResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				var participationID string
				for _, typ := range resp.Types {
					if typ.GetKey() == "participation" {
						participationID = typ.GetId()
						break
					}
				}
				if participationID == "" {
					t.Fatal("missing participation")
				}
				locked := 0
				for _, g := range resp.Groups {
					if g.GetSubjectTypeId() != participationID {
						continue
					}
					for _, f := range g.Properties {
						if f.GetLocked() {
							locked++
						}
					}
				}
				if locked != 3 {
					t.Fatalf("locked=%d want 3", locked)
				}
				placeable := 0
				for _, g := range resp.Groups {
					if g.GetPresentation().GetPlaceable() {
						placeable++
					}
				}
				if placeable != 3 {
					t.Fatalf("placeable=%d", placeable)
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
				hasEdges := false
				for _, r := range rules.Rules {
					if !r.GetRefuse() && len(r.GetEdges()) == 2 {
						hasEdges = true
						break
					}
				}
				if !hasEdges {
					t.Fatal("connect rules missing edges")
				}
			},
		},
	})
}

func TestReorderSubjectTypeProperties(t *testing.T) {
	dir, userID, _, personID := subjectFixture(t)
	list := func() []string { return bindingIDsForType(t, dir, personID) }
	ids := list()
	if len(ids) < 3 {
		t.Fatalf("person binds %d", len(ids))
	}

	reversed := slices.Clone(ids)
	slices.Reverse(reversed)
	out, err := ReorderSubjectTypeProperties(marshalProto(t, &engine.ReorderSubjectTypePropertiesRequest{
		ProjectDir: dir, UserId: userID, SubjectTypeId: personID, PropertyIds: reversed,
	}))
	if err != nil {
		t.Fatal(err)
	}
	var resp engine.ReorderSubjectTypePropertiesResponse
	if err := proto.Unmarshal(out, &resp); err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, p := range resp.Properties {
		got = append(got, p.GetProperty().GetId())
	}
	if !slices.Equal(got, reversed) {
		t.Fatalf("response order %v, want %v", got, reversed)
	}
	assertLatestAuditAction(t, dir, "reorder_subject_type_properties")

	_, err = ReorderSubjectTypeProperties(marshalProto(t, &engine.ReorderSubjectTypePropertiesRequest{
		ProjectDir: dir, UserId: userID, SubjectTypeId: personID, PropertyIds: reversed[1:],
	}))
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Code() != apperr.CodeSubjectVocabInvalid {
		t.Fatalf("short list: %v", err)
	}
	if !slices.Equal(list(), reversed) {
		t.Fatal("a refused reorder moved the order")
	}
}

func TestAssignRemoveSubjectTypePropertyAreAudited(t *testing.T) {
	dir, userID, _, personID := subjectFixture(t)
	created, err := CreateProperty(marshalProto(t, &engine.CreatePropertyRequest{
		ProjectDir: dir, UserId: userID, Label: "Burial Ground", ValueType: "text",
	}))
	if err != nil {
		t.Fatal(err)
	}
	var prop engine.CreatePropertyResponse
	if err := proto.Unmarshal(created, &prop); err != nil {
		t.Fatal(err)
	}
	if _, err := AssignSubjectTypeProperty(marshalProto(t, &engine.AssignSubjectTypePropertyRequest{
		ProjectDir: dir, UserId: userID, SubjectTypeId: personID, PropertyId: prop.Property.GetId(),
	})); err != nil {
		t.Fatal(err)
	}
	assertLatestAuditAction(t, dir, "assign_subject_type_property")
	if _, err := RemoveSubjectTypeProperty(marshalProto(t, &engine.RemoveSubjectTypePropertyRequest{
		ProjectDir: dir, UserId: userID, SubjectTypeId: personID, PropertyId: prop.Property.GetId(),
	})); err != nil {
		t.Fatal(err)
	}
	assertLatestAuditAction(t, dir, "remove_subject_type_property")
}

// The workspace sends propertymeta.Types, field for field and in order.
func TestGetPropertiesWorkspaceValueTypes(t *testing.T) {
	dir, _, _, _ := subjectFixture(t)
	var got engine.GetPropertiesWorkspaceResponse
	mustRPC(t, GetPropertiesWorkspace, &engine.GetPropertiesWorkspaceRequest{ProjectDir: dir}, &got)
	if len(got.ValueTypes) != len(propertymeta.Types) {
		t.Fatalf("value_types %d, want %d", len(got.ValueTypes), len(propertymeta.Types))
	}
	for i, want := range propertymeta.Types {
		r := got.ValueTypes[i]
		if r.GetValueType() != want.ValueType || r.GetResearcherCreatable() != want.ResearcherCreatable || r.GetSingleOnly() != want.SingleOnly || r.GetNegatable() != want.Negatable {
			t.Errorf("value_types[%d] = %+v, want %+v", i, r, want)
		}
	}
	for _, r := range got.ValueTypes {
		switch r.GetValueType() {
		case "boolean":
			if !r.GetResearcherCreatable() || r.GetNegatable() {
				t.Errorf("boolean rule = %+v, want creatable and not negatable", r)
			}
		case "text":
			if !r.GetNegatable() {
				t.Errorf("text rule = %+v, want negatable", r)
			}
		}
	}
}
