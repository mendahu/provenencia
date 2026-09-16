package handlers

import (
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
	"google.golang.org/protobuf/proto"
)

func TestListSources(t *testing.T) {
	runRPC(t, ListSources, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "empty then create",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, _ := sourceFixture(t)
				return &engine.ListSourcesRequest{ProjectDir: dir}
			},
			want: &engine.ListSourcesResponse{},
			after: func(t *testing.T, out []byte, req proto.Message) {
				var list engine.ListSourcesResponse
				if err := proto.Unmarshal(out, &list); err != nil {
					t.Fatal(err)
				}
				if len(list.Sources) != 0 {
					t.Fatalf("want empty got %d", len(list.Sources))
				}
			},
		},
	})
}

func TestCreateSourceAndWorkspace(t *testing.T) {
	runRPC(t, CreateSource, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "creates source",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, typeID := sourceFixture(t)
				return &engine.CreateSourceRequest{
					ProjectDir:   dir,
					UserId:       userID,
					SourceTypeId: typeID,
					Title:        "Passport",
					Description:  "1948",
				}
			},
			want: nil,
			after: func(t *testing.T, out []byte, req proto.Message) {
				var created engine.CreateSourceResponse
				if err := proto.Unmarshal(out, &created); err != nil {
					t.Fatal(err)
				}
				if created.Source.GetTitle() != "Passport" || created.Source.GetDescription() != "1948" {
					t.Fatalf("%+v", created.Source)
				}
				cr := req.(*engine.CreateSourceRequest)
				assertLatestAuditAction(t, cr.ProjectDir, "create_source")
				wsOut, err := GetSourceWorkspace(marshalProto(t, &engine.GetSourceWorkspaceRequest{
					ProjectDir: cr.ProjectDir,
					SourceId:   created.Source.Id,
				}))
				if err != nil {
					t.Fatal(err)
				}
				var ws engine.GetSourceWorkspaceResponse
				if err := proto.Unmarshal(wsOut, &ws); err != nil {
					t.Fatal(err)
				}
				if ws.Source.GetTitle() != "Passport" {
					t.Fatalf("%+v", ws.Source)
				}
			},
		},
	})
}

func TestSourceNotes(t *testing.T) {
	runRPC(t, AddSourceNote, []rpcTest{
		{
			name: "add update delete note",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, typeID := sourceFixture(t)
				cout, err := CreateSource(marshalProto(t, &engine.CreateSourceRequest{
					ProjectDir: dir, UserId: userID, SourceTypeId: typeID, Title: "Book",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateSourceResponse
				if err := proto.Unmarshal(cout, &created); err != nil {
					t.Fatal(err)
				}
				return &engine.AddSourceNoteRequest{
					ProjectDir: dir, UserId: userID, SourceId: created.Source.Id, Body: "First",
				}
			},
			want: nil,
			after: func(t *testing.T, out []byte, req proto.Message) {
				var added engine.AddSourceNoteResponse
				if err := proto.Unmarshal(out, &added); err != nil {
					t.Fatal(err)
				}
				if added.Note.GetBody() != "First" {
					t.Fatalf("%+v", added.Note)
				}
				ar := req.(*engine.AddSourceNoteRequest)
				uout, err := UpdateSourceNote(marshalProto(t, &engine.UpdateSourceNoteRequest{
					ProjectDir: ar.ProjectDir, UserId: ar.UserId, NoteId: added.Note.Id, Body: "Revised",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var updated engine.UpdateSourceNoteResponse
				if err := proto.Unmarshal(uout, &updated); err != nil {
					t.Fatal(err)
				}
				if updated.Note.GetBody() != "Revised" {
					t.Fatalf("%+v", updated.Note)
				}
				if _, err := DeleteSourceNote(marshalProto(t, &engine.DeleteSourceNoteRequest{
					ProjectDir: ar.ProjectDir, UserId: ar.UserId, NoteId: added.Note.Id,
				})); err != nil {
					t.Fatal(err)
				}
			},
		},
	})
}

// metadataSetFixture creates a source and returns a request skeleton targeting
// the first seeded field of the wanted data type ("text" | "date").
func metadataSetFixture(t *testing.T, dataType string) *engine.SetSourceMetadataRequest {
	t.Helper()
	dir, userID, typeID := sourceFixture(t)
	cout, err := CreateSource(marshalProto(t, &engine.CreateSourceRequest{
		ProjectDir: dir, UserId: userID, SourceTypeId: typeID, Title: "Book",
	}))
	if err != nil {
		t.Fatal(err)
	}
	var created engine.CreateSourceResponse
	if err := proto.Unmarshal(cout, &created); err != nil {
		t.Fatal(err)
	}
	fout, err := ListMetadataFields(marshalProto(t, &engine.ListMetadataFieldsRequest{ProjectDir: dir}))
	if err != nil {
		t.Fatal(err)
	}
	var fields engine.ListMetadataFieldsResponse
	if err := proto.Unmarshal(fout, &fields); err != nil {
		t.Fatal(err)
	}
	var fieldID string
	for _, f := range fields.Fields {
		if f.DataType == dataType {
			fieldID = f.Id
			break
		}
	}
	if fieldID == "" {
		t.Fatalf("no seeded %s field", dataType)
	}
	return &engine.SetSourceMetadataRequest{
		ProjectDir: dir, UserId: userID, SourceId: created.Source.Id, FieldId: fieldID,
	}
}

func unmarshalSetMetadataEntry(t *testing.T, raw []byte) *engine.MetadataWorkspaceEntry {
	t.Helper()
	var resp engine.SetSourceMetadataResponse
	if err := proto.Unmarshal(raw, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Entry == nil {
		t.Fatal("missing entry")
	}
	return resp.Entry
}

func TestGetSourceWorkspace(t *testing.T) {
	runRPC(t, GetSourceWorkspace, []rpcTest{
		{
			name: "workspace reads source-specific data only",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, typeID := sourceFixture(t)
				cout, err := CreateSource(marshalProto(t, &engine.CreateSourceRequest{
					ProjectDir: dir, UserId: userID, SourceTypeId: typeID, Title: "Book",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateSourceResponse
				if err := proto.Unmarshal(cout, &created); err != nil {
					t.Fatal(err)
				}
				return &engine.GetSourceWorkspaceRequest{ProjectDir: dir, SourceId: created.Source.Id}
			},
			after: func(t *testing.T, raw []byte, _ proto.Message) {
				var resp engine.GetSourceWorkspaceResponse
				if err := proto.Unmarshal(raw, &resp); err != nil {
					t.Fatal(err)
				}
				if resp.Source.GetTitle() != "Book" {
					t.Fatalf("source %+v", resp.Source)
				}
				// The type's seeded suggestions still arrive as empty
				// metadata rows — that join is per-source, unlike the
				// vocabulary lists, which each have their own RPC.
				if len(resp.Metadata) == 0 {
					t.Fatal("expected suggested metadata rows for the seeded type")
				}
			},
		},
	})
}

func TestSetSourceMetadata(t *testing.T) {
	runRPC(t, SetSourceMetadata, []rpcTest{
		{
			name: "set text metadata returns entry",
			reqFn: func(t *testing.T) proto.Message {
				req := metadataSetFixture(t, "text")
				req.ValueText = "Ada"
				return req
			},
			after: func(t *testing.T, raw []byte, req proto.Message) {
				sr := req.(*engine.SetSourceMetadataRequest)
				entry := unmarshalSetMetadataEntry(t, raw)
				if entry.GetValueText() != "Ada" || !entry.GetHasValue() {
					t.Fatalf("entry %+v", entry)
				}
				if entry.GetField().GetId() != sr.FieldId {
					t.Fatalf("field %s want %s", entry.GetField().GetId(), sr.FieldId)
				}
				if entry.GetDateValueId() != "" || entry.GetDate() != nil {
					t.Fatalf("unexpected date on text entry: %+v", entry)
				}
			},
		},
		{
			name: "set date metadata returns structured entry and keeps date on text-only update",
			reqFn: func(t *testing.T) proto.Message {
				req := metadataSetFixture(t, "date")
				req.ValueText = "about the year 1890"
				year := int32(1890)
				req.Date = &engine.DateValueInput{Kind: "point", Qualifier: "ABT", StartYear: &year}
				return req
			},
			after: func(t *testing.T, raw []byte, req proto.Message) {
				sr := req.(*engine.SetSourceMetadataRequest)
				entry := unmarshalSetMetadataEntry(t, raw)
				if entry.GetDateValueId() == "" {
					t.Fatal("missing date_value_id")
				}
				d := entry.GetDate()
				if d.GetKind() != "point" || d.GetQualifier() != "ABT" || d.GetStartYear() != 1890 {
					t.Fatalf("date %+v", d)
				}

				// A text-only update must keep the structured DateValue and
				// still return it on the refreshed entry.
				tout, err := SetSourceMetadata(marshalProto(t, &engine.SetSourceMetadataRequest{
					ProjectDir: sr.ProjectDir, UserId: sr.UserId, SourceId: sr.SourceId,
					FieldId: sr.FieldId, ValueText: "circa 1890",
				}))
				if err != nil {
					t.Fatal(err)
				}
				textEntry := unmarshalSetMetadataEntry(t, tout)
				if textEntry.GetValueText() != "circa 1890" {
					t.Fatalf("value %q", textEntry.GetValueText())
				}
				if textEntry.GetDateValueId() != entry.GetDateValueId() {
					t.Fatalf("date_value_id changed: %q -> %q", entry.GetDateValueId(), textEntry.GetDateValueId())
				}
				if textEntry.GetDate().GetStartYear() != 1890 {
					t.Fatalf("date lost on text-only update: %+v", textEntry.GetDate())
				}
			},
		},
	})
}

func TestDismissAndReorderSourceMetadata(t *testing.T) {
	runRPC(t, DismissSourceMetadataSuggestion, []rpcTest{
		{
			name: "dismiss suggestion returns metadata without that empty field",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _, sourceID, fieldID := metadataSuggestionFixture(t)
				return &engine.DismissSourceMetadataSuggestionRequest{
					ProjectDir: dir, UserId: userID, SourceId: sourceID, FieldId: fieldID,
				}
			},
			after: func(t *testing.T, raw []byte, req proto.Message) {
				dismissReq := req.(*engine.DismissSourceMetadataSuggestionRequest)
				var resp engine.DismissSourceMetadataSuggestionResponse
				if err := proto.Unmarshal(raw, &resp); err != nil {
					t.Fatal(err)
				}
				for _, e := range resp.Metadata {
					if e.GetField().GetId() == dismissReq.FieldId && !e.HasValue {
						t.Fatalf("dismissed empty suggestion %s still present", dismissReq.FieldId)
					}
				}
			},
		},
	})

	runRPC(t, ReorderSourceMetadata, []rpcTest{
		{
			name: "reorder swaps visible metadata order",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _, sourceID, _ := metadataSuggestionFixture(t)
				wout, err := GetSourceWorkspace(marshalProto(t, &engine.GetSourceWorkspaceRequest{
					ProjectDir: dir, SourceId: sourceID,
				}))
				if err != nil {
					t.Fatal(err)
				}
				var ws engine.GetSourceWorkspaceResponse
				if err := proto.Unmarshal(wout, &ws); err != nil {
					t.Fatal(err)
				}
				if len(ws.Metadata) < 2 {
					t.Fatalf("need ≥2 metadata rows, got %d", len(ws.Metadata))
				}
				ids := make([]string, len(ws.Metadata))
				for i, e := range ws.Metadata {
					ids[i] = e.Field.Id
				}
				// reverse
				for i, j := 0, len(ids)-1; i < j; i, j = i+1, j-1 {
					ids[i], ids[j] = ids[j], ids[i]
				}
				return &engine.ReorderSourceMetadataRequest{
					ProjectDir: dir, UserId: userID, SourceId: sourceID, FieldIds: ids,
				}
			},
			after: func(t *testing.T, raw []byte, req proto.Message) {
				reorderReq := req.(*engine.ReorderSourceMetadataRequest)
				var resp engine.ReorderSourceMetadataResponse
				if err := proto.Unmarshal(raw, &resp); err != nil {
					t.Fatal(err)
				}
				if len(resp.Metadata) != len(reorderReq.FieldIds) {
					t.Fatalf("got %d metadata want %d", len(resp.Metadata), len(reorderReq.FieldIds))
				}
				for i, e := range resp.Metadata {
					if e.Field.Id != reorderReq.FieldIds[i] {
						t.Fatalf("index %d: got %s want %s", i, e.Field.Id, reorderReq.FieldIds[i])
					}
				}
			},
		},
	})
}

func metadataSuggestionFixture(t *testing.T) (dir, userID, typeID, sourceID, fieldID string) {
	t.Helper()
	dir, userID, typeID = sourceFixture(t)
	cout, err := CreateSource(marshalProto(t, &engine.CreateSourceRequest{
		ProjectDir: dir, UserId: userID, SourceTypeId: typeID, Title: "Meta src",
	}))
	if err != nil {
		t.Fatal(err)
	}
	var created engine.CreateSourceResponse
	if err := proto.Unmarshal(cout, &created); err != nil {
		t.Fatal(err)
	}
	sourceID = created.Source.Id
	sout, err := ListTypeSuggestions(marshalProto(t, &engine.ListTypeSuggestionsRequest{
		ProjectDir: dir, TypeId: typeID,
	}))
	if err != nil {
		t.Fatal(err)
	}
	var suggestions engine.ListTypeSuggestionsResponse
	if err := proto.Unmarshal(sout, &suggestions); err != nil {
		t.Fatal(err)
	}
	if len(suggestions.Suggestions) == 0 {
		t.Fatal("expected type suggestions")
	}
	fieldID = suggestions.Suggestions[0].Field.Id
	return dir, userID, typeID, sourceID, fieldID
}
