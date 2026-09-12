package handlers

import (
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/catalogsession"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/sourcefields"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"google.golang.org/protobuf/proto"
)

func TestSourceDefs(t *testing.T) {
	runRPC(t, CreateSourceType, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "create user type and field",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _ := sourceFixture(t)
				return &engine.CreateSourceTypeRequest{
					ProjectDir: dir, UserId: userID, Label: "Deed",
				}
			},
			want: nil,
			after: func(t *testing.T, out []byte, req proto.Message) {
				var typ engine.CreateSourceTypeResponse
				if err := proto.Unmarshal(out, &typ); err != nil {
					t.Fatal(err)
				}
				if typ.Type.GetKey() != "deed" || typ.Type.GetOrigin() != "user" || typ.Type.GetLabel() != "Deed" {
					t.Fatalf("%+v", typ.Type)
				}
				if typ.Type.GetIconKey() != sourcetypes.DefaultIconKey {
					t.Fatalf("icon_key %+v", typ.Type)
				}
				cr := req.(*engine.CreateSourceTypeRequest)
				fout, err := CreateMetadataField(marshalProto(t, &engine.CreateMetadataFieldRequest{
					ProjectDir: cr.ProjectDir, UserId: cr.UserId, Label: "Folio", DataType: "text",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var field engine.CreateMetadataFieldResponse
				if err := proto.Unmarshal(fout, &field); err != nil {
					t.Fatal(err)
				}
				if field.Field.GetKey() != "folio" || field.Field.GetOrigin() != "user" {
					t.Fatalf("%+v", field.Field)
				}
			},
		},
	})
}

func TestCreateSourceTypeMintsKeyFromLabel(t *testing.T) {
	runRPC(t, CreateSourceType, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "mints slug key",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _ := sourceFixture(t)
				return &engine.CreateSourceTypeRequest{
					ProjectDir: dir, UserId: userID, Label: "Grandma's scrapbook",
				}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var resp engine.CreateSourceTypeResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				if resp.Type.GetKey() != "grandmas-scrapbook" || resp.Type.GetOrigin() != "user" {
					t.Fatalf("%+v", resp.Type)
				}
			},
		},
		{
			name: "rejects unslugifiable label",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _ := sourceFixture(t)
				return &engine.CreateSourceTypeRequest{ProjectDir: dir, UserId: userID, Label: "..."}
			},
			wantErr: true,
		},
		{
			name: "rejects unknown icon_key",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _ := sourceFixture(t)
				return &engine.CreateSourceTypeRequest{
					ProjectDir: dir, UserId: userID, Label: "Weird", IconKey: "not_a_key",
				}
			},
			wantErr:   true,
			wantErrIs: sourcetypes.ErrInvalid,
		},
		{
			name: "defaults empty icon_key to type_evidence",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _ := sourceFixture(t)
				return &engine.CreateSourceTypeRequest{
					ProjectDir: dir, UserId: userID, Label: "Loose scrap",
				}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var resp engine.CreateSourceTypeResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				if resp.Type.GetIconKey() != sourcetypes.DefaultIconKey {
					t.Fatalf("%+v", resp.Type)
				}
			},
		},
	})
}

func TestCreateMetadataFieldMintsKeyFromLabel(t *testing.T) {
	runRPC(t, CreateMetadataField, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "mints slug key",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _ := sourceFixture(t)
				return &engine.CreateMetadataFieldRequest{
					ProjectDir: dir, UserId: userID, Label: "Grandma's album code", DataType: "text",
				}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var resp engine.CreateMetadataFieldResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				if resp.Field.GetKey() != "grandmas-album-code" || resp.Field.GetOrigin() != "user" {
					t.Fatalf("%+v", resp.Field)
				}
			},
		},
		{
			name: "rejects unslugifiable label",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _ := sourceFixture(t)
				return &engine.CreateMetadataFieldRequest{ProjectDir: dir, UserId: userID, Label: "...", DataType: "text"}
			},
			wantErr: true,
		},
		{
			name: "rejects duplicate key under user origin",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _ := sourceFixture(t)
				return &engine.CreateMetadataFieldRequest{ProjectDir: dir, UserId: userID, Label: "Folio", DataType: "text"}
			},
			calls:     2,
			wantErr:   true,
			wantErrIs: sourcefields.ErrDuplicateKey,
		},
	})
}

func TestUpdateMetadataField(t *testing.T) {
	runRPC(t, UpdateMetadataField, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "updates label and description, key and data type unchanged",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _ := sourceFixture(t)
				out, err := CreateMetadataField(marshalProto(t, &engine.CreateMetadataFieldRequest{
					ProjectDir: dir, UserId: userID, Label: "Album code", DataType: "text",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateMetadataFieldResponse
				if err := proto.Unmarshal(out, &created); err != nil {
					t.Fatal(err)
				}
				return &engine.UpdateMetadataFieldRequest{
					ProjectDir: dir, UserId: userID, FieldId: created.Field.GetId(),
					Label: "Album Code", DataType: "text", Description: "updated",
				}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var resp engine.UpdateMetadataFieldResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				if resp.Field.GetKey() != "album-code" || resp.Field.GetLabel() != "Album Code" ||
					resp.Field.GetDataType() != "text" || resp.Field.GetDescription() != "updated" {
					t.Fatalf("%+v", resp.Field)
				}
			},
		},
		{
			name: "rejects data type change",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _ := sourceFixture(t)
				out, err := CreateMetadataField(marshalProto(t, &engine.CreateMetadataFieldRequest{
					ProjectDir: dir, UserId: userID, Label: "Album code", DataType: "text",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateMetadataFieldResponse
				if err := proto.Unmarshal(out, &created); err != nil {
					t.Fatal(err)
				}
				return &engine.UpdateMetadataFieldRequest{
					ProjectDir: dir, UserId: userID, FieldId: created.Field.GetId(),
					Label: "Album Code", DataType: "date", Description: "updated",
				}
			},
			wantErr: true,
		},
		{
			name: "updates a seeded field",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _ := sourceFixture(t)
				listOut, err := ListMetadataFields(marshalProto(t, &engine.ListMetadataFieldsRequest{ProjectDir: dir}))
				if err != nil {
					t.Fatal(err)
				}
				var list engine.ListMetadataFieldsResponse
				if err := proto.Unmarshal(listOut, &list); err != nil {
					t.Fatal(err)
				}
				var seededID string
				for _, f := range list.Fields {
					if f.GetOrigin() == "provenencia" {
						seededID = f.GetId()
						break
					}
				}
				if seededID == "" {
					t.Fatal("expected a seeded field")
				}
				return &engine.UpdateMetadataFieldRequest{
					ProjectDir: dir, UserId: userID, FieldId: seededID,
					Label: "Renamed starter", DataType: "text", Description: "edited",
				}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var resp engine.UpdateMetadataFieldResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				if resp.Field.GetOrigin() != "provenencia" || resp.Field.GetLabel() != "Renamed starter" ||
					resp.Field.GetDescription() != "edited" || resp.Field.GetKey() == "" {
					t.Fatalf("%+v", resp.Field)
				}
			},
		},
		{
			name: "rejects editing a plugin field",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _ := sourceFixture(t)
				var id []byte
				err := catalogsession.Do(dir, func(c *database.Catalog) error {
					var err error
					id, err = sourcefields.Upsert(c, sourcefields.Field{
						Key: "memorial_id", Origin: "plugin:findagrave", Label: "Memorial id", DataType: sourcefields.DataTypeText,
					})
					return err
				})
				if err != nil {
					t.Fatal(err)
				}
				return &engine.UpdateMetadataFieldRequest{
					ProjectDir: dir, UserId: userID, FieldId: uuidString(id),
					Label: "Changed", DataType: "text",
				}
			},
			wantErr: true,
		},
	})
}

func TestDeleteSourceType(t *testing.T) {
	runRPC(t, DeleteSourceType, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "deletes unused user type",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _ := sourceFixture(t)
				out, err := CreateSourceType(marshalProto(t, &engine.CreateSourceTypeRequest{
					ProjectDir: dir, UserId: userID, Label: "Deed",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateSourceTypeResponse
				if err := proto.Unmarshal(out, &created); err != nil {
					t.Fatal(err)
				}
				return &engine.DeleteSourceTypeRequest{
					ProjectDir: dir, UserId: userID, TypeId: created.Type.GetId(),
				}
			},
			want: &engine.DeleteSourceTypeResponse{},
		},
		{
			name: "refuses type in use",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, typeID := sourceFixture(t)
				if _, err := CreateSource(marshalProto(t, &engine.CreateSourceRequest{
					ProjectDir: dir, UserId: userID, SourceTypeId: typeID, Title: "One",
				})); err != nil {
					t.Fatal(err)
				}
				return &engine.DeleteSourceTypeRequest{
					ProjectDir: dir, UserId: userID, TypeId: typeID,
				}
			},
			wantErr: true,
		},
		{
			name: "deletes unused seeded type",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, typeID := sourceFixture(t)
				return &engine.DeleteSourceTypeRequest{
					ProjectDir: dir, UserId: userID, TypeId: typeID,
				}
			},
			want: &engine.DeleteSourceTypeResponse{},
		},
	})
}

func TestDeleteMetadataField(t *testing.T) {
	runRPC(t, DeleteMetadataField, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "deletes unused user field",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _ := sourceFixture(t)
				out, err := CreateMetadataField(marshalProto(t, &engine.CreateMetadataFieldRequest{
					ProjectDir: dir, UserId: userID, Label: "Folio", DataType: "text",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateMetadataFieldResponse
				if err := proto.Unmarshal(out, &created); err != nil {
					t.Fatal(err)
				}
				return &engine.DeleteMetadataFieldRequest{
					ProjectDir: dir, UserId: userID, FieldId: created.Field.GetId(),
				}
			},
			want: &engine.DeleteMetadataFieldResponse{},
		},
		{
			name: "refuses field in use",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, typeID := sourceFixture(t)
				fout, err := CreateMetadataField(marshalProto(t, &engine.CreateMetadataFieldRequest{
					ProjectDir: dir, UserId: userID, Label: "Folio", DataType: "text",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var field engine.CreateMetadataFieldResponse
				if err := proto.Unmarshal(fout, &field); err != nil {
					t.Fatal(err)
				}
				sout, err := CreateSource(marshalProto(t, &engine.CreateSourceRequest{
					ProjectDir: dir, UserId: userID, SourceTypeId: typeID, Title: "One",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var src engine.CreateSourceResponse
				if err := proto.Unmarshal(sout, &src); err != nil {
					t.Fatal(err)
				}
				if _, err := SetSourceMetadata(marshalProto(t, &engine.SetSourceMetadataRequest{
					ProjectDir: dir, UserId: userID, SourceId: src.Source.GetId(),
					FieldId: field.Field.GetId(), ValueText: "12",
				})); err != nil {
					t.Fatal(err)
				}
				return &engine.DeleteMetadataFieldRequest{
					ProjectDir: dir, UserId: userID, FieldId: field.Field.GetId(),
				}
			},
			wantErr: true,
		},
	})
}

// newUserType creates a user-origin type through the RPC and returns its id.
func newUserType(t *testing.T, dir, userID, label string) string {
	t.Helper()
	out, err := CreateSourceType(marshalProto(t, &engine.CreateSourceTypeRequest{
		ProjectDir: dir, UserId: userID, Label: label,
	}))
	if err != nil {
		t.Fatal(err)
	}
	var created engine.CreateSourceTypeResponse
	if err := proto.Unmarshal(out, &created); err != nil {
		t.Fatal(err)
	}
	return created.Type.GetId()
}

// newUserField creates a user-origin metadata field through the RPC and
// returns its id.
func newUserField(t *testing.T, dir, userID, label string) string {
	t.Helper()
	out, err := CreateMetadataField(marshalProto(t, &engine.CreateMetadataFieldRequest{
		ProjectDir: dir, UserId: userID, Label: label, DataType: "text",
	}))
	if err != nil {
		t.Fatal(err)
	}
	var created engine.CreateMetadataFieldResponse
	if err := proto.Unmarshal(out, &created); err != nil {
		t.Fatal(err)
	}
	return created.Field.GetId()
}

func TestUpdateSourceType(t *testing.T) {
	runRPC(t, UpdateSourceType, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "edits a seeded type and reports its use count",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, typeID := sourceFixture(t)
				if _, err := CreateSource(marshalProto(t, &engine.CreateSourceRequest{
					ProjectDir: dir, UserId: userID, SourceTypeId: typeID, Title: "One",
				})); err != nil {
					t.Fatal(err)
				}
				return &engine.UpdateSourceTypeRequest{
					ProjectDir: dir, UserId: userID, TypeId: typeID,
					Label: "Renamed starter", Description: "edited",
				}
			},
			after: func(t *testing.T, out []byte, req proto.Message) {
				var resp engine.UpdateSourceTypeResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				if resp.Type.GetOrigin() != "provenencia" || resp.Type.GetLabel() != "Renamed starter" ||
					resp.Type.GetDescription() != "edited" || resp.Type.GetUsedBy() != 1 {
					t.Fatalf("%+v", resp.Type)
				}
				// The update reply also refreshes the suggested-field count
				// — the seeded starter ships with suggestions, so it must
				// agree with what ListTypeSuggestions reads.
				ur := req.(*engine.UpdateSourceTypeRequest)
				lout, err := ListTypeSuggestions(marshalProto(t, &engine.ListTypeSuggestionsRequest{
					ProjectDir: ur.ProjectDir, TypeId: ur.TypeId,
				}))
				if err != nil {
					t.Fatal(err)
				}
				var list engine.ListTypeSuggestionsResponse
				if err := proto.Unmarshal(lout, &list); err != nil {
					t.Fatal(err)
				}
				if len(list.Suggestions) == 0 || int(resp.Type.GetSuggestedFieldCount()) != len(list.Suggestions) {
					t.Fatalf("suggested_field_count %d, suggestions %d", resp.Type.GetSuggestedFieldCount(), len(list.Suggestions))
				}
			},
		},
		{
			name: "rejects editing a plugin type",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _ := sourceFixture(t)
				var id []byte
				err := catalogsession.Do(dir, func(c *database.Catalog) error {
					var err error
					id, err = sourcetypes.Upsert(c, sourcetypes.Type{
						Key: "grave_memorial", Origin: "plugin:findagrave", Label: "Grave memorial",
					})
					return err
				})
				if err != nil {
					t.Fatal(err)
				}
				return &engine.UpdateSourceTypeRequest{
					ProjectDir: dir, UserId: userID, TypeId: uuidString(id), Label: "Changed",
				}
			},
			wantErr: true,
		},
	})
}

func TestAssignAndRemoveTypeField(t *testing.T) {
	runRPC(t, AssignTypeField, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "assigns in order, is idempotent, and detaches without deleting the field",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _ := sourceFixture(t)
				typeID := newUserType(t, dir, userID, "Parish register")
				second := newUserField(t, dir, userID, "Folio")
				first := newUserField(t, dir, userID, "Officiant")
				// Assign Folio first so stored order and label order disagree.
				for _, fieldID := range []string{second, first} {
					if _, err := AssignTypeField(marshalProto(t, &engine.AssignTypeFieldRequest{
						ProjectDir: dir, UserId: userID, TypeId: typeID, FieldId: fieldID,
					})); err != nil {
						t.Fatal(err)
					}
				}
				// Re-assigning must not reorder.
				return &engine.AssignTypeFieldRequest{
					ProjectDir: dir, UserId: userID, TypeId: typeID, FieldId: second,
				}
			},
			after: func(t *testing.T, out []byte, req proto.Message) {
				var resp engine.AssignTypeFieldResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				if len(resp.Suggestions) != 2 ||
					resp.Suggestions[0].Field.GetLabel() != "Folio" ||
					resp.Suggestions[1].Field.GetLabel() != "Officiant" {
					t.Fatalf("%+v", resp.Suggestions)
				}
				ar := req.(*engine.AssignTypeFieldRequest)
				rout, err := RemoveTypeField(marshalProto(t, &engine.RemoveTypeFieldRequest{
					ProjectDir: ar.ProjectDir, UserId: ar.UserId, TypeId: ar.TypeId, FieldId: ar.FieldId,
				}))
				if err != nil {
					t.Fatal(err)
				}
				var removed engine.RemoveTypeFieldResponse
				if err := proto.Unmarshal(rout, &removed); err != nil {
					t.Fatal(err)
				}
				if len(removed.Suggestions) != 1 || removed.Suggestions[0].Field.GetLabel() != "Officiant" {
					t.Fatalf("%+v", removed.Suggestions)
				}
				// The detached field stays in the vocabulary.
				lout, err := ListMetadataFields(marshalProto(t, &engine.ListMetadataFieldsRequest{ProjectDir: ar.ProjectDir}))
				if err != nil {
					t.Fatal(err)
				}
				var list engine.ListMetadataFieldsResponse
				if err := proto.Unmarshal(lout, &list); err != nil {
					t.Fatal(err)
				}
				var found bool
				for _, f := range list.Fields {
					found = found || f.GetId() == ar.FieldId
				}
				if !found {
					t.Fatal("removing a suggestion must not delete the field")
				}
			},
		},
	})
}

func TestListTypeSuggestions(t *testing.T) {
	runRPC(t, ListTypeSuggestions, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "reads the seeded starter's suggestions",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, typeID := sourceFixture(t)
				return &engine.ListTypeSuggestionsRequest{ProjectDir: dir, TypeId: typeID}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var resp engine.ListTypeSuggestionsResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				if len(resp.Suggestions) == 0 {
					t.Fatal("expected the seeded starter to suggest fields")
				}
				for _, s := range resp.Suggestions {
					if s.Field.GetKey() == "" || s.Field.GetDataType() == "" {
						t.Fatalf("%+v", s)
					}
				}
			},
		},
		{
			name: "a type with no suggestions reads empty",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _ := sourceFixture(t)
				return &engine.ListTypeSuggestionsRequest{
					ProjectDir: dir, TypeId: newUserType(t, dir, userID, "Letter"),
				}
			},
			want: &engine.ListTypeSuggestionsResponse{},
		},
	})
}
