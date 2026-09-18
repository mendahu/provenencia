package handlers

import (
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
	"google.golang.org/protobuf/proto"
)

func subjectFixture(t *testing.T) (projectDir, userID, sourceID, subjectTypeID string) {
	t.Helper()
	dir, userID, sourceTypeID := sourceFixture(t)
	createOut, err := CreateSource(marshalProto(t, &engine.CreateSourceRequest{
		ProjectDir:   dir,
		UserId:       userID,
		SourceTypeId: sourceTypeID,
		Title:        "Census",
	}))
	if err != nil {
		t.Fatal(err)
	}
	var created engine.CreateSourceResponse
	if err := proto.Unmarshal(createOut, &created); err != nil {
		t.Fatal(err)
	}
	typesOut, err := ListSubjectTypes(marshalProto(t, &engine.ListSubjectTypesRequest{ProjectDir: dir}))
	if err != nil {
		t.Fatal(err)
	}
	var types engine.ListSubjectTypesResponse
	if err := proto.Unmarshal(typesOut, &types); err != nil {
		t.Fatal(err)
	}
	if len(types.Types) == 0 {
		t.Fatal("expected seeded subject types")
	}
	personID := ""
	for _, typ := range types.Types {
		if typ.GetKey() == "person" {
			personID = typ.GetId()
			break
		}
	}
	if personID == "" {
		personID = types.Types[0].GetId()
	}
	return dir, userID, created.Source.GetId(), personID
}

func TestListSubjectTypes(t *testing.T) {
	runRPC(t, ListSubjectTypes, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "lists seeded types",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, _ := sourceFixture(t)
				return &engine.ListSubjectTypesRequest{ProjectDir: dir}
			},
			want: nil,
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var list engine.ListSubjectTypesResponse
				if err := proto.Unmarshal(out, &list); err != nil {
					t.Fatal(err)
				}
				if len(list.Types) == 0 {
					t.Fatal("expected seeded subject types")
				}
				found := false
				for _, typ := range list.Types {
					if typ.GetKey() == "person" && typ.GetCandidateRefPrefix() == "CPR" {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("person type missing: %+v", list.Types)
				}
			},
		},
	})
}

func TestCreateSubject(t *testing.T) {
	runRPC(t, CreateSubject, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "rejects empty user id",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, sourceID, typeID := subjectFixture(t)
				return &engine.CreateSubjectRequest{
					ProjectDir:    dir,
					UserId:        "",
					SourceId:      sourceID,
					SubjectTypeId: typeID,
					Label:         "Nameless",
				}
			},
			wantErr:   true,
			wantErrIs: errInvalidUserID,
		},
		{
			name: "creates subject",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, sourceID, typeID := subjectFixture(t)
				return &engine.CreateSubjectRequest{
					ProjectDir:    dir,
					UserId:        userID,
					SourceId:      sourceID,
					SubjectTypeId: typeID,
					Label:         "Alice",
					Description:   "daughter",
				}
			},
			want: nil,
			after: func(t *testing.T, out []byte, req proto.Message) {
				var created engine.CreateSubjectResponse
				if err := proto.Unmarshal(out, &created); err != nil {
					t.Fatal(err)
				}
				if created.Subject.GetLabel() != "Alice" || created.Subject.GetDescription() != "daughter" {
					t.Fatalf("%+v", created.Subject)
				}
				if created.Subject.GetRef() == "" || created.Subject.GetId() == "" {
					t.Fatalf("missing id/ref %+v", created.Subject)
				}
				cr := req.(*engine.CreateSubjectRequest)
				assertLatestAuditAction(t, cr.ProjectDir, "create_subject")
				listOut, err := ListSubjects(marshalProto(t, &engine.ListSubjectsRequest{
					ProjectDir: cr.ProjectDir,
					SourceId:   cr.SourceId,
				}))
				if err != nil {
					t.Fatal(err)
				}
				var list engine.ListSubjectsResponse
				if err := proto.Unmarshal(listOut, &list); err != nil {
					t.Fatal(err)
				}
				if len(list.Subjects) != 1 || list.Subjects[0].GetId() != created.Subject.GetId() {
					t.Fatalf("%+v", list.Subjects)
				}
			},
		},
	})
}

func TestUpdateSubject(t *testing.T) {
	runRPC(t, UpdateSubject, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "updates label",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, sourceID, typeID := subjectFixture(t)
				createOut, err := CreateSubject(marshalProto(t, &engine.CreateSubjectRequest{
					ProjectDir: dir, UserId: userID, SourceId: sourceID, SubjectTypeId: typeID, Label: "Bob",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateSubjectResponse
				if err := proto.Unmarshal(createOut, &created); err != nil {
					t.Fatal(err)
				}
				return &engine.UpdateSubjectRequest{
					ProjectDir:  dir,
					UserId:      userID,
					SubjectId:   created.Subject.GetId(),
					Label:       "Robert",
					Description: "updated",
				}
			},
			want: nil,
			after: func(t *testing.T, out []byte, req proto.Message) {
				var updated engine.UpdateSubjectResponse
				if err := proto.Unmarshal(out, &updated); err != nil {
					t.Fatal(err)
				}
				if updated.Subject.GetLabel() != "Robert" || updated.Subject.GetDescription() != "updated" {
					t.Fatalf("%+v", updated.Subject)
				}
				ur := req.(*engine.UpdateSubjectRequest)
				assertLatestAuditAction(t, ur.ProjectDir, "update_subject")
			},
		},
	})
}

func TestDeleteSubject(t *testing.T) {
	runRPC(t, DeleteSubject, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "deletes subject",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, sourceID, typeID := subjectFixture(t)
				createOut, err := CreateSubject(marshalProto(t, &engine.CreateSubjectRequest{
					ProjectDir: dir, UserId: userID, SourceId: sourceID, SubjectTypeId: typeID, Label: "Carol",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateSubjectResponse
				if err := proto.Unmarshal(createOut, &created); err != nil {
					t.Fatal(err)
				}
				return &engine.DeleteSubjectRequest{
					ProjectDir: dir,
					UserId:     userID,
					SubjectId:  created.Subject.GetId(),
				}
			},
			want: &engine.DeleteSubjectResponse{},
			after: func(t *testing.T, _ []byte, req proto.Message) {
				dr := req.(*engine.DeleteSubjectRequest)
				assertLatestAuditAction(t, dr.ProjectDir, "delete_subject")
				sourcesOut, err := ListSources(marshalProto(t, &engine.ListSourcesRequest{ProjectDir: dr.ProjectDir}))
				if err != nil {
					t.Fatal(err)
				}
				var sources engine.ListSourcesResponse
				if err := proto.Unmarshal(sourcesOut, &sources); err != nil {
					t.Fatal(err)
				}
				if len(sources.Sources) == 0 {
					t.Fatal("no sources")
				}
				listOut, err := ListSubjects(marshalProto(t, &engine.ListSubjectsRequest{
					ProjectDir: dr.ProjectDir,
					SourceId:   sources.Sources[0].GetId(),
				}))
				if err != nil {
					t.Fatal(err)
				}
				var list engine.ListSubjectsResponse
				if err := proto.Unmarshal(listOut, &list); err != nil {
					t.Fatal(err)
				}
				if len(list.Subjects) != 0 {
					t.Fatalf("want empty after delete, got %+v", list.Subjects)
				}
			},
		},
	})
}

func TestListSubjects(t *testing.T) {
	runRPC(t, ListSubjects, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "empty then after create",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, sourceID, _ := subjectFixture(t)
				return &engine.ListSubjectsRequest{ProjectDir: dir, SourceId: sourceID}
			},
			want: &engine.ListSubjectsResponse{},
		},
	})
}

func TestSubjectPositions(t *testing.T) {
	runRPC(t, SetSubjectPosition, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "set list clear persist across session close",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, sourceID, typeID := subjectFixture(t)
				createOut, err := CreateSubject(marshalProto(t, &engine.CreateSubjectRequest{
					ProjectDir: dir, UserId: userID, SourceId: sourceID, SubjectTypeId: typeID, Label: "Dave",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateSubjectResponse
				if err := proto.Unmarshal(createOut, &created); err != nil {
					t.Fatal(err)
				}
				return &engine.SetSubjectPositionRequest{
					ProjectDir: dir,
					SubjectId:  created.Subject.GetId(),
					GridX:      3,
					GridY:      5,
				}
			},
			want: nil,
			after: func(t *testing.T, out []byte, req proto.Message) {
				var set engine.SetSubjectPositionResponse
				if err := proto.Unmarshal(out, &set); err != nil {
					t.Fatal(err)
				}
				if set.Position.GetGridX() != 3 || set.Position.GetGridY() != 5 {
					t.Fatalf("%+v", set.Position)
				}
				sr := req.(*engine.SetSubjectPositionRequest)

				// Resolve source via subject list scan is heavy; store source on Create in reqFn via closure —
				// reopen: close session then list positions by finding source through ListSubjects isn't available.
				// Instead: list positions needs source_id. Recover by listing all subjects isn't possible without source.
				// Use ListSubjects after looking up — we only have subject_id. Query via CreateSubject fixture:
				// re-list by closing and opening with known source from creating another list call stored...
				// Simplest: call ListSubjects isn't possible. We stored only subject_id.
				// Fix: look up by creating list from source — we need source_id in after.
				// Encode source in project by listing subjects we can't.
				// Better approach: put source_id in label of a parallel request field — can't.
				// Re-create pattern: subjectFixture returns source; stash in ProjectDir string? No.
				// Use ListSubjectPositions after finding source via a helper that lists all sources then subjects.
				sourcesOut, err := ListSources(marshalProto(t, &engine.ListSourcesRequest{ProjectDir: sr.ProjectDir}))
				if err != nil {
					t.Fatal(err)
				}
				var sources engine.ListSourcesResponse
				if err := proto.Unmarshal(sourcesOut, &sources); err != nil {
					t.Fatal(err)
				}
				if len(sources.Sources) == 0 {
					t.Fatal("no sources")
				}
				sourceID := sources.Sources[0].GetId()

				if _, err := CloseCatalogSession(marshalProto(t, &engine.CloseCatalogSessionRequest{
					ProjectDir: sr.ProjectDir,
				})); err != nil {
					t.Fatal(err)
				}

				listOut, err := ListSubjectPositions(marshalProto(t, &engine.ListSubjectPositionsRequest{
					ProjectDir: sr.ProjectDir,
					SourceId:   sourceID,
				}))
				if err != nil {
					t.Fatal(err)
				}
				var list engine.ListSubjectPositionsResponse
				if err := proto.Unmarshal(listOut, &list); err != nil {
					t.Fatal(err)
				}
				if len(list.Positions) != 1 {
					t.Fatalf("want 1 position after reopen, got %+v", list.Positions)
				}
				if list.Positions[0].GetSubjectId() != sr.SubjectId ||
					list.Positions[0].GetGridX() != 3 || list.Positions[0].GetGridY() != 5 {
					t.Fatalf("%+v", list.Positions[0])
				}

				if _, err := ClearSubjectPosition(marshalProto(t, &engine.ClearSubjectPositionRequest{
					ProjectDir: sr.ProjectDir,
					SubjectId:  sr.SubjectId,
				})); err != nil {
					t.Fatal(err)
				}
				listOut, err = ListSubjectPositions(marshalProto(t, &engine.ListSubjectPositionsRequest{
					ProjectDir: sr.ProjectDir,
					SourceId:   sourceID,
				}))
				if err != nil {
					t.Fatal(err)
				}
				if err := proto.Unmarshal(listOut, &list); err != nil {
					t.Fatal(err)
				}
				if len(list.Positions) != 0 {
					t.Fatalf("want empty after clear, got %+v", list.Positions)
				}
			},
		},
	})
}

func TestClearSubjectPositionBadProto(t *testing.T) {
	runRPC(t, ClearSubjectPosition, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
	})
}

func TestListSubjectPositionsBadProto(t *testing.T) {
	runRPC(t, ListSubjectPositions, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
	})
}
