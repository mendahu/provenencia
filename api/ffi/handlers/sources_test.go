package handlers

import (
	"slices"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database/sourcemetadata"
	"github.com/mendahu/provenencia/core/database/sources"
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
		{
			name: "has_artifact false then true after create",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, typeID := sourceFixture(t)
				cout, err := CreateSource(marshalProto(t, &engine.CreateSourceRequest{
					ProjectDir: dir, UserId: userID, SourceTypeId: typeID, Title: "Bare",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateSourceResponse
				if err := proto.Unmarshal(cout, &created); err != nil {
					t.Fatal(err)
				}
				lout, err := ListSources(marshalProto(t, &engine.ListSourcesRequest{ProjectDir: dir}))
				if err != nil {
					t.Fatal(err)
				}
				var before engine.ListSourcesResponse
				if err := proto.Unmarshal(lout, &before); err != nil {
					t.Fatal(err)
				}
				if len(before.Sources) != 1 || before.Sources[0].GetHasArtifact() {
					t.Fatalf("want has_artifact false got %+v", before.Sources)
				}
				if _, err := CreateArtifact(marshalProto(t, &engine.CreateArtifactRequest{
					ProjectDir: dir, UserId: userID, SourceId: created.Source.Id, Label: "Scan",
				})); err != nil {
					t.Fatal(err)
				}
				return &engine.ListSourcesRequest{ProjectDir: dir}
			},
			after: func(t *testing.T, out []byte, req proto.Message) {
				var list engine.ListSourcesResponse
				if err := proto.Unmarshal(out, &list); err != nil {
					t.Fatal(err)
				}
				if len(list.Sources) != 1 || !list.Sources[0].GetHasArtifact() {
					t.Fatalf("want has_artifact true got %+v", list.Sources)
				}
			},
		},
	})
}

func TestCreateSourceAndWorkspace(t *testing.T) {
	runRPC(t, CreateSource, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "rejects empty user id",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, typeID := sourceFixture(t)
				return &engine.CreateSourceRequest{
					ProjectDir:   dir,
					UserId:       "",
					SourceTypeId: typeID,
					Title:        "Orphan",
				}
			},
			wantErr:   true,
			wantErrIs: errInvalidUserID,
		},
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
				if created.Effects.GetRevision() <= 0 || !slices.Contains(created.Effects.GetSourceScopes(), created.Source.GetId()) {
					t.Fatalf("effects %+v", created.Effects)
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
// the first seeded field of the wanted data type ("text" | "url").
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
			},
		},
		{
			name: "set url metadata accepts host-shaped values",
			reqFn: func(t *testing.T) proto.Message {
				req := metadataSetFixture(t, "text")
				fout, err := CreateMetadataField(marshalProto(t, &engine.CreateMetadataFieldRequest{
					ProjectDir: req.ProjectDir, UserId: req.UserId, Label: "Catalogue URL", DataType: "url",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var field engine.CreateMetadataFieldResponse
				if err := proto.Unmarshal(fout, &field); err != nil {
					t.Fatal(err)
				}
				req.FieldId = field.Field.GetId()
				req.ValueText = "www.url.com"
				return req
			},
			after: func(t *testing.T, raw []byte, _ proto.Message) {
				entry := unmarshalSetMetadataEntry(t, raw)
				if entry.GetValueText() != "www.url.com" || entry.GetField().GetDataType() != "url" {
					t.Fatalf("entry %+v", entry)
				}
			},
		},
		{
			name: "set url metadata rejects file scheme",
			reqFn: func(t *testing.T) proto.Message {
				req := metadataSetFixture(t, "text")
				fout, err := CreateMetadataField(marshalProto(t, &engine.CreateMetadataFieldRequest{
					ProjectDir: req.ProjectDir, UserId: req.UserId, Label: "Bad URL", DataType: "url",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var field engine.CreateMetadataFieldResponse
				if err := proto.Unmarshal(fout, &field); err != nil {
					t.Fatal(err)
				}
				req.FieldId = field.Field.GetId()
				req.ValueText = "file:/tmp"
				return req
			},
			wantErr:   true,
			wantErrIs: sourcemetadata.ErrInvalid,
		},
		{
			name: "set url metadata rejects a single label",
			reqFn: func(t *testing.T) proto.Message {
				req := metadataSetFixture(t, "text")
				fout, err := CreateMetadataField(marshalProto(t, &engine.CreateMetadataFieldRequest{
					ProjectDir: req.ProjectDir, UserId: req.UserId, Label: "Bare word URL", DataType: "url",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var field engine.CreateMetadataFieldResponse
				if err := proto.Unmarshal(fout, &field); err != nil {
					t.Fatal(err)
				}
				req.FieldId = field.Field.GetId()
				req.ValueText = "jakerobins"
				return req
			},
			wantErr:   true,
			wantErrIs: sourcemetadata.ErrInvalid,
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

func TestDeleteSource(t *testing.T) {
	runRPC(t, DeleteSource, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "empty source erases",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, typeID := sourceFixture(t)
				cout, err := CreateSource(marshalProto(t, &engine.CreateSourceRequest{
					ProjectDir: dir, UserId: userID, SourceTypeId: typeID, Title: "Bare",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateSourceResponse
				if err := proto.Unmarshal(cout, &created); err != nil {
					t.Fatal(err)
				}
				if _, err := AddSourceNote(marshalProto(t, &engine.AddSourceNoteRequest{
					ProjectDir: dir, UserId: userID, SourceId: created.Source.Id, Body: "faded",
				})); err != nil {
					t.Fatal(err)
				}
				return &engine.DeleteSourceRequest{
					ProjectDir: dir, UserId: userID, SourceId: created.Source.Id,
				}
			},
			want: &engine.DeleteSourceResponse{},
			after: func(t *testing.T, _ []byte, req proto.Message) {
				dr := req.(*engine.DeleteSourceRequest)
				_, err := GetSourceWorkspace(marshalProto(t, &engine.GetSourceWorkspaceRequest{
					ProjectDir: dr.ProjectDir, SourceId: dr.SourceId,
				}))
				if err == nil {
					t.Fatal("workspace still loads")
				}
			},
		},
		{
			name: "inbound artifact refuses",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, typeID := sourceFixture(t)
				cout, err := CreateSource(marshalProto(t, &engine.CreateSourceRequest{
					ProjectDir: dir, UserId: userID, SourceTypeId: typeID, Title: "Held",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateSourceResponse
				if err := proto.Unmarshal(cout, &created); err != nil {
					t.Fatal(err)
				}
				if _, err := CreateArtifact(marshalProto(t, &engine.CreateArtifactRequest{
					ProjectDir: dir, UserId: userID, SourceId: created.Source.Id, Label: "Scan",
				})); err != nil {
					t.Fatal(err)
				}
				return &engine.DeleteSourceRequest{
					ProjectDir: dir, UserId: userID, SourceId: created.Source.Id,
				}
			},
			wantErr:   true,
			wantErrIs: sources.ErrInUse,
		},
		{
			name: "missing refuses",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _ := sourceFixture(t)
				return &engine.DeleteSourceRequest{
					ProjectDir: dir, UserId: userID, SourceId: "01890000-0000-7000-8000-000000000001",
				}
			},
			wantErr:   true,
			wantErrIs: sources.ErrInvalid,
		},
		{
			name: "malformed id refuses as an id, not a source",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _ := sourceFixture(t)
				return &engine.DeleteSourceRequest{
					ProjectDir: dir, UserId: userID, SourceId: "00000000-0000-0000-0000-000000000001",
				}
			},
			wantErr:   true,
			wantErrIs: errInvalidID,
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

// copySourceFixture seeds Sources A and B, each with one fileless Artifact.
func copySourceFixture(t *testing.T) (dir, userID, typeID, sourceA, artifactA, artifactB string) {
	t.Helper()
	dir, userID, typeID = sourceFixture(t)
	seed := func(title string) (sourceID, artifactID string) {
		cout, err := CreateSource(marshalProto(t, &engine.CreateSourceRequest{
			ProjectDir: dir, UserId: userID, SourceTypeId: typeID, Title: title,
		}))
		if err != nil {
			t.Fatal(err)
		}
		var created engine.CreateSourceResponse
		if err := proto.Unmarshal(cout, &created); err != nil {
			t.Fatal(err)
		}
		aout, err := CreateArtifact(marshalProto(t, &engine.CreateArtifactRequest{
			ProjectDir: dir, UserId: userID, SourceId: created.Source.Id, Label: "Scan",
		}))
		if err != nil {
			t.Fatal(err)
		}
		var art engine.CreateArtifactResponse
		if err := proto.Unmarshal(aout, &art); err != nil {
			t.Fatal(err)
		}
		return created.Source.Id, art.Artifact.Id
	}
	sourceA, artifactA = seed("Obituary A")
	_, artifactB = seed("Obituary B")
	return dir, userID, typeID, sourceA, artifactA, artifactB
}

func TestCopySource(t *testing.T) {
	runRPC(t, CopySource, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "rejects a malformed source id",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, typeID, _, _, _ := copySourceFixture(t)
				return &engine.CopySourceRequest{
					ProjectDir: dir, UserId: userID, SourceId: "nope", Title: "Copy", SourceTypeId: typeID,
				}
			},
			wantErr:   true,
			wantErrIs: errInvalidID,
		},
		{
			name: "copies a source with its Artifact",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, typeID, sourceA, artifactA, _ := copySourceFixture(t)
				return &engine.CopySourceRequest{
					ProjectDir: dir, UserId: userID, SourceId: sourceA, Title: "Copy of Obituary A",
					SourceTypeId: typeID, CopyDescription: true, CopyMetadata: true, CopyCredibility: true,
					ArtifactIds: []string{artifactA},
				}
			},
			after: func(t *testing.T, out []byte, req proto.Message) {
				var resp engine.CopySourceResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				cr := req.(*engine.CopySourceRequest)
				if resp.Source.GetTitle() != "Copy of Obituary A" || resp.Source.GetSourceTypeId() != cr.SourceTypeId ||
					!strings.HasPrefix(resp.Source.GetRef(), "SRC-") || resp.Source.GetId() == cr.SourceId {
					t.Fatalf("source %+v", resp.Source)
				}
				if resp.Effects.GetRevision() <= 0 || !slices.Contains(resp.Effects.GetSourceScopes(), resp.Source.GetId()) {
					t.Fatalf("effects %+v", resp.Effects)
				}
				assertLatestAuditAction(t, cr.ProjectDir, "copy_source")
				wsOut, err := GetSourceWorkspace(marshalProto(t, &engine.GetSourceWorkspaceRequest{
					ProjectDir: cr.ProjectDir, SourceId: resp.Source.GetId(),
				}))
				if err != nil {
					t.Fatal(err)
				}
				var ws engine.GetSourceWorkspaceResponse
				if err := proto.Unmarshal(wsOut, &ws); err != nil {
					t.Fatal(err)
				}
				if len(ws.Artifacts) != 1 || ws.Artifacts[0].GetLabel() != "Scan" {
					t.Fatalf("artifacts %+v", ws.Artifacts)
				}
			},
		},
		{
			name: "rejects a malformed artifact id",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, typeID, sourceA, artifactA, _ := copySourceFixture(t)
				return &engine.CopySourceRequest{
					ProjectDir: dir, UserId: userID, SourceId: sourceA, Title: "Copy", SourceTypeId: typeID,
					ArtifactIds: []string{artifactA, "nope"},
				}
			},
			wantErr:   true,
			wantErrIs: errInvalidID,
		},
		{
			name: "rejects another source's Artifact",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, typeID, sourceA, _, artifactB := copySourceFixture(t)
				return &engine.CopySourceRequest{
					ProjectDir: dir, UserId: userID, SourceId: sourceA, Title: "Copy", SourceTypeId: typeID,
					ArtifactIds: []string{artifactB},
				}
			},
			wantErr:   true,
			wantErrIs: sources.ErrInvalid,
		},
	})
}
