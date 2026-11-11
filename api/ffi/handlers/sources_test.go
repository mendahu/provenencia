package handlers

import (
	"cmp"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/catalogsession"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/dbtest"
	"github.com/mendahu/provenencia/core/database/sourcemetadata"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/vocabkey"
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
		{
			name: "date and sort key",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, typeID := sourceFixture(t)
				dates := []struct {
					title string
					date  *engine.DateValueInput
				}{
					{"1881", pointDate(1881, 0, 0)},
					{"aft. 29 Mar 2011", qualifiedDate(datevalue.QualifierAFT, 2011, 3, 29)},
					{"undated", nil},
					{"1879", pointDate(1879, 0, 0)},
					{"bef. 1880", qualifiedDate(datevalue.QualifierBEF, 1880, 0, 0)},
					{"phrase", &engine.DateValueInput{Kind: datevalue.KindPoint, Phrase: "in the reign of Victoria"}},
					{"29 Mar 2011", pointDate(2011, 3, 29)},
					{"1880", pointDate(1880, 0, 0)},
					{"bef. 29 Mar 2011", qualifiedDate(datevalue.QualifierBEF, 2011, 3, 29)},
				}
				for _, d := range dates {
					var created engine.CreateSourceResponse
					mustRPC(t, CreateSource, &engine.CreateSourceRequest{
						ProjectDir: dir, UserId: userID, SourceTypeId: typeID, Title: d.title,
					}, &created)
					if d.date != nil {
						mustRPC(t, SetSourceDate, &engine.SetSourceDateRequest{
							ProjectDir: dir, UserId: userID, SourceId: created.Source.Id, Date: d.date,
						}, &engine.SetSourceDateResponse{})
					}
				}
				return &engine.ListSourcesRequest{ProjectDir: dir}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var list engine.ListSourcesResponse
				if err := proto.Unmarshal(out, &list); err != nil {
					t.Fatal(err)
				}
				byKey := slices.Clone(list.Sources)
				slices.SortStableFunc(byKey, func(a, b *engine.Source) int {
					switch {
					case a.DateSortKey == nil && b.DateSortKey == nil:
						return 0
					case a.DateSortKey == nil:
						return 1
					case b.DateSortKey == nil:
						return -1
					}
					return cmp.Compare(*a.DateSortKey, *b.DateSortKey)
				})
				var titles []string
				for _, src := range byKey {
					titles = append(titles, src.GetTitle())
					if src.GetTitle() == "undated" && (src.GetDate() != nil || src.DateSortKey != nil) {
						t.Fatalf("undated source %+v", src)
					}
					if src.GetTitle() == "phrase" && (src.GetDate().GetPhrase() == "" || src.DateSortKey != nil) {
						t.Fatalf("phrase-only source %+v", src)
					}
					if !src.GetIsPublic() {
						t.Fatalf("%s is not public", src.GetTitle())
					}
				}
				want := []string{
					"1879", "bef. 1880", "1880", "1881", "bef. 29 Mar 2011", "29 Mar 2011", "aft. 29 Mar 2011",
				}
				if !slices.Equal(titles[:len(want)], want) {
					t.Fatalf("date order %v, want %v first", titles, want)
				}
			},
		},
		{
			name: "cover thumbnail with a damaged type has no path",
			reqFn: func(t *testing.T) proto.Message {
				f := sourceWithPNG(t)
				mustRPC(t, SetSourceCover, &engine.SetSourceCoverRequest{
					ProjectDir: f.dir, UserId: f.userID, SourceId: f.sourceID,
					CoverMode: "artifact", PrimaryArtifactId: f.artifactID,
				}, &engine.SetSourceCoverResponse{})
				var before engine.ListSourcesResponse
				mustRPC(t, ListSources, &engine.ListSourcesRequest{ProjectDir: f.dir}, &before)
				if len(before.Sources) != 1 || before.Sources[0].GetThumbnailRelPath() == "" {
					t.Fatalf("want a cover thumbnail before the damage, got %+v", before.Sources)
				}
				damageCatalog(t, f.dir, `UPDATE files SET media_type = 'application/x-sh'
					WHERE id IN (SELECT derived_file_id FROM file_derivatives)`)
				return &engine.ListSourcesRequest{ProjectDir: f.dir}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var list engine.ListSourcesResponse
				if err := proto.Unmarshal(out, &list); err != nil {
					t.Fatal(err)
				}
				if len(list.Sources) != 1 || list.Sources[0].GetThumbnailRelPath() != "" {
					t.Fatalf("want an empty thumbnail_rel_path, got %+v", list.Sources)
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
// sourceOwnerProto is the wire owner of a Source's metadata board.
func sourceOwnerProto(id string) *engine.MetadataOwner {
	return &engine.MetadataOwner{Level: string(vocabkey.LevelSource), Id: id}
}

func metadataSetFixture(t *testing.T, dataType string) *engine.SetMetadataRequest {
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
	return &engine.SetMetadataRequest{
		ProjectDir: dir, UserId: userID, Owner: sourceOwnerProto(created.Source.Id), FieldId: fieldID,
	}
}

// urlMetadataSetFixture is metadataSetFixture aimed at a new url-typed field
// (the seeded fields are all text).
func urlMetadataSetFixture(t *testing.T) *engine.SetMetadataRequest {
	t.Helper()
	req := metadataSetFixture(t, "text")
	var field engine.CreateMetadataFieldResponse
	mustRPC(t, CreateMetadataField, &engine.CreateMetadataFieldRequest{
		ProjectDir: req.ProjectDir, UserId: req.UserId, Label: "Catalogue URL", DataType: "url",
	}, &field)
	req.FieldId = field.Field.GetId()
	return req
}

func unmarshalSetMetadataEntry(t *testing.T, raw []byte) *engine.MetadataWorkspaceEntry {
	t.Helper()
	var resp engine.SetMetadataResponse
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
			name: "unknown source is not found",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, _ := sourceFixture(t)
				return &engine.GetSourceWorkspaceRequest{ProjectDir: dir, SourceId: "01890000-0000-7000-8000-000000000001"}
			},
			wantErr:   true,
			wantErrIs: sources.ErrNotFound,
		},
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
		{
			name: "a stored type off the allowlist keeps the file and drops its path",
			reqFn: func(t *testing.T) proto.Message {
				f := sourceWithPNG(t)
				damageCatalog(t, f.dir, `UPDATE files SET media_type = 'application/x-sh'
					WHERE id IN (SELECT file_id FROM artifacts)`)
				return &engine.GetSourceWorkspaceRequest{ProjectDir: f.dir, SourceId: f.sourceID}
			},
			after: func(t *testing.T, raw []byte, _ proto.Message) {
				var resp engine.GetSourceWorkspaceResponse
				if err := proto.Unmarshal(raw, &resp); err != nil {
					t.Fatal(err)
				}
				if len(resp.Artifacts) != 1 {
					t.Fatalf("artifacts %+v", resp.Artifacts)
				}
				file := resp.Artifacts[0].GetFile()
				if file.GetRelPath() != "" {
					t.Fatalf("rel_path %q want empty", file.GetRelPath())
				}
				if file.GetOriginalFilename() != "page.png" || file.GetMediaType() != "application/x-sh" {
					t.Fatalf("file %+v", file)
				}
			},
		},
		{
			name: "cover thumbnail with a damaged type has no path",
			reqFn: func(t *testing.T) proto.Message {
				f := sourceWithPNG(t)
				mustRPC(t, SetSourceCover, &engine.SetSourceCoverRequest{
					ProjectDir: f.dir, UserId: f.userID, SourceId: f.sourceID,
					CoverMode: "artifact", PrimaryArtifactId: f.artifactID,
				}, &engine.SetSourceCoverResponse{})
				var before engine.GetSourceWorkspaceResponse
				mustRPC(t, GetSourceWorkspace, &engine.GetSourceWorkspaceRequest{ProjectDir: f.dir, SourceId: f.sourceID}, &before)
				if before.Source.GetThumbnailRelPath() == "" {
					t.Fatalf("want a cover thumbnail before the damage, got %+v", before.Source)
				}
				damageCatalog(t, f.dir, `UPDATE files SET media_type = 'application/x-sh'
					WHERE id IN (SELECT derived_file_id FROM file_derivatives)`)
				return &engine.GetSourceWorkspaceRequest{ProjectDir: f.dir, SourceId: f.sourceID}
			},
			after: func(t *testing.T, raw []byte, _ proto.Message) {
				var resp engine.GetSourceWorkspaceResponse
				if err := proto.Unmarshal(raw, &resp); err != nil {
					t.Fatal(err)
				}
				if resp.Source.GetThumbnailRelPath() != "" {
					t.Fatalf("thumbnail_rel_path %q want empty", resp.Source.GetThumbnailRelPath())
				}
			},
		},
		{
			name: "a raster type off the allowlist keeps the file and drops its path",
			reqFn: func(t *testing.T) proto.Message {
				f := sourceWithPNG(t)
				damageCatalog(t, f.dir, `DELETE FROM file_derivatives`)
				damageCatalog(t, f.dir, `UPDATE files SET media_type = 'image/jpg'
					WHERE id IN (SELECT file_id FROM artifacts)`)
				return &engine.GetSourceWorkspaceRequest{ProjectDir: f.dir, SourceId: f.sourceID}
			},
			after: func(t *testing.T, raw []byte, _ proto.Message) {
				var resp engine.GetSourceWorkspaceResponse
				if err := proto.Unmarshal(raw, &resp); err != nil {
					t.Fatal(err)
				}
				if len(resp.Artifacts) != 1 {
					t.Fatalf("artifacts %+v", resp.Artifacts)
				}
				file := resp.Artifacts[0].GetFile()
				if file.GetRelPath() != "" {
					t.Fatalf("rel_path %q want empty", file.GetRelPath())
				}
				if file.GetOriginalFilename() != "page.png" || file.GetMediaType() != "image/jpg" {
					t.Fatalf("file %+v", file)
				}
				if resp.Artifacts[0].GetThumbnailRelPath() != "" {
					t.Fatalf("thumbnail_rel_path %q want empty", resp.Artifacts[0].GetThumbnailRelPath())
				}
			},
		},
		{
			name: "url metadata carries the web link Go computes",
			reqFn: func(t *testing.T) proto.Message {
				req := urlMetadataSetFixture(t)
				req.ValueText = "www.example.org"
				mustRPC(t, SetMetadata, req, &engine.SetMetadataResponse{})
				return &engine.GetSourceWorkspaceRequest{ProjectDir: req.ProjectDir, SourceId: req.Owner.Id}
			},
			after: func(t *testing.T, raw []byte, _ proto.Message) {
				entry := onlyValuedMetadata(t, raw)
				if entry.GetValueText() != "www.example.org" || entry.GetOpenHref() != "https://www.example.org" {
					t.Fatalf("entry %+v", entry)
				}
			},
		},
		{
			name: "text metadata has no web link",
			reqFn: func(t *testing.T) proto.Message {
				req := metadataSetFixture(t, "text")
				req.ValueText = "www.example.org"
				mustRPC(t, SetMetadata, req, &engine.SetMetadataResponse{})
				return &engine.GetSourceWorkspaceRequest{ProjectDir: req.ProjectDir, SourceId: req.Owner.Id}
			},
			after: func(t *testing.T, raw []byte, _ proto.Message) {
				entry := onlyValuedMetadata(t, raw)
				if entry.GetValueText() != "www.example.org" || entry.GetOpenHref() != "" {
					t.Fatalf("entry %+v", entry)
				}
			},
		},
		{
			name: "a url value damaged to a file scheme has no web link",
			reqFn: func(t *testing.T) proto.Message {
				req := urlMetadataSetFixture(t)
				req.ValueText = "www.example.org"
				mustRPC(t, SetMetadata, req, &engine.SetMetadataResponse{})
				damageCatalog(t, req.ProjectDir, `UPDATE source_metadata SET value_text = 'file:///System/Applications/Calculator.app'`)
				return &engine.GetSourceWorkspaceRequest{ProjectDir: req.ProjectDir, SourceId: req.Owner.Id}
			},
			after: func(t *testing.T, raw []byte, _ proto.Message) {
				entry := onlyValuedMetadata(t, raw)
				if entry.GetValueText() != "file:///System/Applications/Calculator.app" || entry.GetOpenHref() != "" {
					t.Fatalf("entry %+v", entry)
				}
			},
		},
	})
}

// onlyValuedMetadata returns the one entry of a GetSourceWorkspace reply that
// holds a value (the rest are the type's unset suggestions).
func onlyValuedMetadata(t *testing.T, raw []byte) *engine.MetadataWorkspaceEntry {
	t.Helper()
	var resp engine.GetSourceWorkspaceResponse
	if err := proto.Unmarshal(raw, &resp); err != nil {
		t.Fatal(err)
	}
	var valued []*engine.MetadataWorkspaceEntry
	for _, e := range resp.Metadata {
		if e.GetHasValue() {
			valued = append(valued, e)
		}
	}
	if len(valued) != 1 {
		t.Fatalf("valued metadata %+v", valued)
	}
	return valued[0]
}

func TestSetMetadata(t *testing.T) {
	runRPC(t, SetMetadata, []rpcTest{
		{
			name: "set text metadata returns entry",
			reqFn: func(t *testing.T) proto.Message {
				req := metadataSetFixture(t, "text")
				req.ValueText = "Ada"
				return req
			},
			after: func(t *testing.T, raw []byte, req proto.Message) {
				sr := req.(*engine.SetMetadataRequest)
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
			name: "set url metadata returns the web link",
			reqFn: func(t *testing.T) proto.Message {
				req := urlMetadataSetFixture(t)
				req.ValueText = "example.org/records"
				return req
			},
			after: func(t *testing.T, raw []byte, _ proto.Message) {
				entry := unmarshalSetMetadataEntry(t, raw)
				if entry.GetValueText() != "example.org/records" || entry.GetOpenHref() != "https://example.org/records" {
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

func TestDismissAndReorderMetadata(t *testing.T) {
	runRPC(t, DismissMetadataSuggestion, []rpcTest{
		{
			name: "dismiss suggestion returns metadata without that empty field",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _, sourceID, fieldID := metadataSuggestionFixture(t)
				return &engine.DismissMetadataSuggestionRequest{
					ProjectDir: dir, UserId: userID, Owner: sourceOwnerProto(sourceID), FieldId: fieldID,
				}
			},
			after: func(t *testing.T, raw []byte, req proto.Message) {
				dismissReq := req.(*engine.DismissMetadataSuggestionRequest)
				var resp engine.DismissMetadataSuggestionResponse
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

	runRPC(t, ReorderMetadata, []rpcTest{
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
				return &engine.ReorderMetadataRequest{
					ProjectDir: dir, UserId: userID, Owner: sourceOwnerProto(sourceID), FieldIds: ids,
				}
			},
			after: func(t *testing.T, raw []byte, req proto.Message) {
				reorderReq := req.(*engine.ReorderMetadataRequest)
				var resp engine.ReorderMetadataResponse
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
			wantErrIs: sources.ErrNotFound,
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
		ProjectDir: dir, TypeId: typeID, Level: "source",
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
			name: "copy carries the date when chosen and stays public",
			reqFn: func(t *testing.T) proto.Message {
				return notPublicDatedCopyRequest(t, true)
			},
			after: func(t *testing.T, out []byte, req proto.Message) {
				var resp engine.CopySourceResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				d := resp.Source.GetDate()
				if d == nil || d.GetStartYear() != 1880 || d.GetStartMonth() != 3 || d.GetStartDay() != 12 ||
					resp.Source.DateSortKey == nil || !resp.Source.GetIsPublic() {
					t.Fatalf("copy %+v", resp.Source)
				}
				assertLatestAuditAction(t, req.(*engine.CopySourceRequest).ProjectDir, "copy_source")
			},
		},
		{
			name: "copy without the date is undated",
			reqFn: func(t *testing.T) proto.Message {
				return notPublicDatedCopyRequest(t, false)
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var resp engine.CopySourceResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				if resp.Source.GetDate() != nil || resp.Source.DateSortKey != nil || !resp.Source.GetIsPublic() {
					t.Fatalf("copy %+v", resp.Source)
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

// notPublicDatedCopyRequest dates `12 Mar 1880` and hides the original, then
// asks to copy it with every part ticked and the date as given.
func notPublicDatedCopyRequest(t *testing.T, copyDate bool) *engine.CopySourceRequest {
	t.Helper()
	dir, userID, typeID, sourceA, _, _ := copySourceFixture(t)
	mustRPC(t, SetSourceDate, &engine.SetSourceDateRequest{
		ProjectDir: dir, UserId: userID, SourceId: sourceA, Date: pointDate(1880, 3, 12),
	}, &engine.SetSourceDateResponse{})
	mustRPC(t, SetSourcePublic, &engine.SetSourcePublicRequest{
		ProjectDir: dir, UserId: userID, SourceId: sourceA, IsPublic: false,
	}, &engine.SetSourcePublicResponse{})
	return &engine.CopySourceRequest{
		ProjectDir: dir, UserId: userID, SourceId: sourceA, Title: "Copy", SourceTypeId: typeID,
		CopyDescription: true, CopyMetadata: true, CopyCredibility: true, CopyDate: copyDate,
	}
}

func pointDate(y, m, d int32) *engine.DateValueInput {
	return qualifiedDate("", y, m, d)
}

func qualifiedDate(qualifier string, y, m, d int32) *engine.DateValueInput {
	in := &engine.DateValueInput{Kind: datevalue.KindPoint, Qualifier: qualifier, StartYear: &y}
	if m != 0 {
		in.StartMonth = &m
	}
	if d != 0 {
		in.StartDay = &d
	}
	return in
}

func TestSetSourceDate(t *testing.T) {
	newSource := func(t *testing.T) (dir, userID, sourceID string) {
		dir, userID, typeID := sourceFixture(t)
		var created engine.CreateSourceResponse
		mustRPC(t, CreateSource, &engine.CreateSourceRequest{
			ProjectDir: dir, UserId: userID, SourceTypeId: typeID, Title: "Register",
		}, &created)
		return dir, userID, created.Source.Id
	}
	set := func(t *testing.T, dir, userID, sourceID string, date *engine.DateValueInput) (*engine.SetSourceDateResponse, error) {
		t.Helper()
		var resp engine.SetSourceDateResponse
		err := rpc(t, SetSourceDate, &engine.SetSourceDateRequest{
			ProjectDir: dir, UserId: userID, SourceId: sourceID, Date: date,
		}, &resp)
		return &resp, err
	}
	dateRows := func(t *testing.T, dir string) int {
		t.Helper()
		return countCatalogRows(t, dir, "date_values")
	}

	t.Run("set, edit in place, same, clear", func(t *testing.T) {
		t.Cleanup(func() { _ = catalogsession.CloseAll() })
		dir, userID, id := newSource(t)

		resp, err := set(t, dir, userID, id, pointDate(1880, 0, 0))
		if err != nil {
			t.Fatal(err)
		}
		d := resp.Source.GetDate()
		if d.GetKind() != datevalue.KindPoint || d.GetStartYear() != 1880 || resp.Source.DateSortKey == nil ||
			resp.Effects.GetRevision() <= 0 || !slices.Contains(resp.Effects.GetSourceScopes(), id) {
			t.Fatalf("set %+v effects %+v", resp.Source, resp.Effects)
		}
		var ws engine.GetSourceWorkspaceResponse
		mustRPC(t, GetSourceWorkspace, &engine.GetSourceWorkspaceRequest{ProjectDir: dir, SourceId: id}, &ws)
		if ws.Source.GetDate().GetStartYear() != 1880 || ws.Source.DateSortKey == nil {
			t.Fatalf("workspace source %+v", ws.Source)
		}
		if n := dateRows(t, dir); n != 1 {
			t.Fatalf("date rows %d", n)
		}

		if resp, err = set(t, dir, userID, id, qualifiedDate(datevalue.QualifierABT, 1880, 3, 12)); err != nil {
			t.Fatal(err)
		}
		d = resp.Source.GetDate()
		if d.GetQualifier() != datevalue.QualifierABT || d.GetStartMonth() != 3 || d.GetStartDay() != 12 || dateRows(t, dir) != 1 {
			t.Fatalf("edit %+v rows=%d", resp.Source, dateRows(t, dir))
		}

		if resp, err = set(t, dir, userID, id, qualifiedDate(datevalue.QualifierABT, 1880, 3, 12)); err != nil {
			t.Fatal(err)
		}
		if resp.Effects.GetRevision() != 0 || len(resp.Effects.GetSourceScopes()) != 0 {
			t.Fatalf("same date wrote: %+v", resp.Effects)
		}

		// A date with a blank kind clears, whatever else it holds.
		if resp, err = set(t, dir, userID, id, &engine.DateValueInput{Kind: " ", Phrase: "ignored"}); err != nil {
			t.Fatal(err)
		}
		if resp.Source.GetDate() != nil || resp.Source.DateSortKey != nil || dateRows(t, dir) != 0 {
			t.Fatalf("clear %+v rows=%d", resp.Source, dateRows(t, dir))
		}
		if resp.Effects.GetRevision() <= 0 {
			t.Fatalf("clear effects %+v", resp.Effects)
		}
	})
	t.Run("invalid date keeps the stored one", func(t *testing.T) {
		t.Cleanup(func() { _ = catalogsession.CloseAll() })
		dir, userID, id := newSource(t)
		if _, err := set(t, dir, userID, id, pointDate(1880, 0, 0)); err != nil {
			t.Fatal(err)
		}
		month := int32(13)
		_, err := set(t, dir, userID, id, &engine.DateValueInput{Kind: datevalue.KindPoint, StartMonth: &month})
		if !errors.Is(err, datevalue.ErrInvalid) {
			t.Fatalf("err %v", err)
		}
		var ws engine.GetSourceWorkspaceResponse
		mustRPC(t, GetSourceWorkspace, &engine.GetSourceWorkspaceRequest{ProjectDir: dir, SourceId: id}, &ws)
		if ws.Source.GetDate().GetStartYear() != 1880 {
			t.Fatalf("stored date %+v", ws.Source.GetDate())
		}
	})
	t.Run("phrase-only date has no sort key", func(t *testing.T) {
		t.Cleanup(func() { _ = catalogsession.CloseAll() })
		dir, userID, id := newSource(t)
		resp, err := set(t, dir, userID, id, &engine.DateValueInput{Kind: datevalue.KindPoint, Phrase: "in the reign of Victoria"})
		if err != nil {
			t.Fatal(err)
		}
		if resp.Source.GetDate().GetPhrase() == "" || resp.Source.DateSortKey != nil {
			t.Fatalf("source %+v", resp.Source)
		}
	})
	t.Run("unknown source", func(t *testing.T) {
		t.Cleanup(func() { _ = catalogsession.CloseAll() })
		dir, userID, _ := newSource(t)
		_, err := set(t, dir, userID, "00000000-0000-7000-8000-000000000000", pointDate(1880, 0, 0))
		if !errors.Is(err, sources.ErrNotFound) {
			t.Fatalf("err %v", err)
		}
	})
	runRPC(t, SetSourceDate, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
	})
}

func TestSetSourcePublic(t *testing.T) {
	t.Cleanup(func() { _ = catalogsession.CloseAll() })
	dir, userID, typeID := sourceFixture(t)
	var created engine.CreateSourceResponse
	mustRPC(t, CreateSource, &engine.CreateSourceRequest{
		ProjectDir: dir, UserId: userID, SourceTypeId: typeID, Title: "Register",
	}, &created)
	id := created.Source.Id
	if !created.Source.GetIsPublic() {
		t.Fatalf("created source is not public: %+v", created.Source)
	}

	var resp engine.SetSourcePublicResponse
	mustRPC(t, SetSourcePublic, &engine.SetSourcePublicRequest{ProjectDir: dir, UserId: userID, SourceId: id, IsPublic: false}, &resp)
	if resp.Source.GetIsPublic() || resp.Effects.GetRevision() <= 0 || !slices.Contains(resp.Effects.GetSourceScopes(), id) {
		t.Fatalf("set false %+v effects %+v", resp.Source, resp.Effects)
	}
	var ws engine.GetSourceWorkspaceResponse
	mustRPC(t, GetSourceWorkspace, &engine.GetSourceWorkspaceRequest{ProjectDir: dir, SourceId: id}, &ws)
	if ws.Source.GetIsPublic() {
		t.Fatalf("workspace source %+v", ws.Source)
	}
	var list engine.ListSourcesResponse
	mustRPC(t, ListSources, &engine.ListSourcesRequest{ProjectDir: dir}, &list)
	if len(list.Sources) != 1 || list.Sources[0].GetIsPublic() {
		t.Fatalf("list %+v", list.Sources)
	}

	var again engine.SetSourcePublicResponse
	mustRPC(t, SetSourcePublic, &engine.SetSourcePublicRequest{ProjectDir: dir, UserId: userID, SourceId: id, IsPublic: false}, &again)
	if again.Effects.GetRevision() != 0 {
		t.Fatalf("unchanged flag wrote: %+v", again.Effects)
	}
	mustRPC(t, SetSourcePublic, &engine.SetSourcePublicRequest{ProjectDir: dir, UserId: userID, SourceId: id, IsPublic: true}, &resp)
	if !resp.Source.GetIsPublic() {
		t.Fatalf("set true %+v", resp.Source)
	}
	t.Run("unknown source", func(t *testing.T) {
		_, err := SetSourcePublic(marshalProto(t, &engine.SetSourcePublicRequest{
			ProjectDir: dir, UserId: userID, SourceId: "00000000-0000-7000-8000-000000000000", IsPublic: false,
		}))
		if !errors.Is(err, sources.ErrNotFound) {
			t.Fatalf("err %v", err)
		}
	})
	runRPC(t, SetSourcePublic, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
	})
}

// countCatalogRows is how many rows table holds.
func countCatalogRows(t *testing.T, dir, table string) int {
	t.Helper()
	var n int
	err := catalogsession.Do(dir, func(c *database.Catalog) error {
		n = dbtest.Count(t, dbtest.MustDB(t, c), `SELECT COUNT(*) FROM `+table)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return n
}
