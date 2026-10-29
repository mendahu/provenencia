package handlers

import (
	"errors"
	"math"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/catalogsession"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/subjects"
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
					GraphId:       soleGraphID(t, dir, sourceID),
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
					GraphId:       soleGraphID(t, dir, sourceID),
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
					GraphId:    cr.GraphId,
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
		{
			name: "creates subject with placement",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, sourceID, typeID := subjectFixture(t)
				return &engine.CreateSubjectRequest{
					ProjectDir:    dir,
					UserId:        userID,
					GraphId:       soleGraphID(t, dir, sourceID),
					SubjectTypeId: typeID,
					Label:         "Placed",
					HasPlacement:  true,
					GridX:         -2,
					GridY:         5,
				}
			},
			after: func(t *testing.T, out []byte, req proto.Message) {
				var created engine.CreateSubjectResponse
				if err := proto.Unmarshal(out, &created); err != nil {
					t.Fatal(err)
				}
				cr := req.(*engine.CreateSubjectRequest)
				listOut, err := ListSubjectPositions(marshalProto(t, &engine.ListSubjectPositionsRequest{
					ProjectDir: cr.ProjectDir,
					GraphId:    cr.GraphId,
				}))
				if err != nil {
					t.Fatal(err)
				}
				var list engine.ListSubjectPositionsResponse
				if err := proto.Unmarshal(listOut, &list); err != nil {
					t.Fatal(err)
				}
				if len(list.Positions) != 1 || list.Positions[0].GetGridX() != -2 || list.Positions[0].GetGridY() != 5 {
					t.Fatalf("%+v", list.Positions)
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
					ProjectDir: dir, UserId: userID, GraphId: soleGraphID(t, dir, sourceID), SubjectTypeId: typeID, Label: "Bob",
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
					ProjectDir: dir, UserId: userID, GraphId: soleGraphID(t, dir, sourceID), SubjectTypeId: typeID, Label: "Carol",
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
					GraphId:    soleGraphID(t, dr.ProjectDir, sources.Sources[0].GetId()),
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
		{
			name: "refuses when observations reference subject",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _, artifactID, placeID, propID := citationFixture(t)
				if _, err := CreateCitationWithObservations(marshalProto(t, &engine.CreateCitationWithObservationsRequest{
					ProjectDir:  dir,
					UserId:      userID,
					ArtifactId:  artifactID,
					LocatorJson: validLocatorJSON,
					Observations: []*engine.ObservationDraft{{
						SubjectId:  placeID,
						PropertyId: propID,
						ValueText:  "Boston",
					}},
				})); err != nil {
					t.Fatal(err)
				}
				return &engine.DeleteSubjectRequest{
					ProjectDir: dir,
					UserId:     userID,
					SubjectId:  placeID,
				}
			},
			wantErr:   true,
			wantErrIs: subjects.ErrInUse,
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
				return &engine.ListSubjectsRequest{ProjectDir: dir, GraphId: soleGraphID(t, dir, sourceID)}
			},
			want: &engine.ListSubjectsResponse{},
		},
		{
			name: "a graph no catalog row names reads empty",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, _, _ := subjectFixture(t)
				return &engine.ListSubjectsRequest{ProjectDir: dir, GraphId: uuid.Must(uuid.NewV7()).String()}
			},
			want: &engine.ListSubjectsResponse{},
			after: func(t *testing.T, _ []byte, req proto.Message) {
				r := req.(*engine.ListSubjectsRequest)
				out, err := ListSubjectPositions(marshalProto(t, &engine.ListSubjectPositionsRequest{ProjectDir: r.GetProjectDir(), GraphId: r.GetGraphId()}))
				if err != nil {
					t.Fatal(err)
				}
				var got engine.ListSubjectPositionsResponse
				if err := proto.Unmarshal(out, &got); err != nil {
					t.Fatal(err)
				}
				if len(got.Positions) != 0 {
					t.Fatalf("positions %+v", got.Positions)
				}
			},
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
					ProjectDir: dir, UserId: userID, GraphId: soleGraphID(t, dir, sourceID), SubjectTypeId: typeID, Label: "Dave",
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
					GraphId:    soleGraphID(t, sr.ProjectDir, sourceID),
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
					GraphId:    soleGraphID(t, sr.ProjectDir, sourceID),
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

// A position outside the grid's range is refused on write; a stored row the
// app did not write reads back held to the edge.
func TestSubjectPositionsAreBounded(t *testing.T) {
	t.Cleanup(func() { _ = catalogsession.CloseAll() })
	f := newGraphFixture(t)
	set := func(x, y int64) error {
		var resp engine.SetSubjectPositionResponse
		return rpc(t, SetSubjectPosition, &engine.SetSubjectPositionRequest{
			ProjectDir: f.dir, SubjectId: f.subjectID, GridX: x, GridY: y,
		}, &resp)
	}
	list := func() *engine.SubjectPosition {
		var resp engine.ListSubjectPositionsResponse
		mustRPC(t, ListSubjectPositions, &engine.ListSubjectPositionsRequest{ProjectDir: f.dir, GraphId: f.defaultGraphID}, &resp)
		for _, p := range resp.GetPositions() {
			if p.GetSubjectId() == f.subjectID {
				return p
			}
		}
		t.Fatalf("no position for %s in %+v", f.subjectID, resp.GetPositions())
		return nil
	}

	if err := set(1000000, -1000000); err != nil {
		t.Fatal(err)
	}
	if p := list(); p.GetGridX() != 1000000 || p.GetGridY() != -1000000 {
		t.Fatalf("%+v", p)
	}
	err := set(1000001, 0)
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Code() != apperr.CodeSubjectPositionsInvalid {
		t.Fatalf("got %v", err)
	}
	if p := list(); p.GetGridX() != 1000000 || p.GetGridY() != -1000000 {
		t.Fatalf("row changed: %+v", p)
	}

	if err := catalogsession.Do(f.dir, func(c *database.Catalog) error {
		db, err := c.DB()
		if err != nil {
			return err
		}
		_, err = db.Exec(`UPDATE subject_positions SET grid_x = ?, grid_y = ?`, int64(math.MaxInt64), int64(math.MinInt64))
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if p := list(); p.GetGridX() != 1000000 || p.GetGridY() != -1000000 {
		t.Fatalf("not clamped: %+v", p)
	}
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

// onSplitSource is a Source with Ada on its Default graph and Bea, placed at
// (2, 3), on a second graph named Jones.
func onSplitSource(t *testing.T) (f graphFixture, jones string, bea *engine.Subject) {
	t.Helper()
	f = newGraphFixture(t)
	jones = f.createGraph(t, "Jones", f.page2).GetGraph().GetId()
	var subs engine.ListSubjectsResponse
	mustRPC(t, ListSubjects, &engine.ListSubjectsRequest{ProjectDir: f.dir, GraphId: f.defaultGraphID}, &subs)
	var created engine.CreateSubjectResponse
	mustRPC(t, CreateSubject, &engine.CreateSubjectRequest{
		ProjectDir: f.dir, UserId: f.userID, GraphId: jones, SubjectTypeId: subs.Subjects[0].GetSubjectTypeId(),
		Label: "Bea", HasPlacement: true, GridX: 2, GridY: 3,
	}, &created)
	return f, jones, created.GetSubject()
}

// A graph reads only its own Subjects and positions; an id no graph has reads empty.
func TestListSubjectsOfASplitSource(t *testing.T) {
	f, jones, bea := onSplitSource(t)
	var subs engine.ListSubjectsResponse
	mustRPC(t, ListSubjects, &engine.ListSubjectsRequest{ProjectDir: f.dir, GraphId: jones}, &subs)
	if len(subs.GetSubjects()) != 1 || subs.GetSubjects()[0].GetId() != bea.GetId() {
		t.Fatalf("Jones subjects %+v", subs.GetSubjects())
	}
	var pos engine.ListSubjectPositionsResponse
	mustRPC(t, ListSubjectPositions, &engine.ListSubjectPositionsRequest{ProjectDir: f.dir, GraphId: jones}, &pos)
	if len(pos.GetPositions()) != 1 || pos.GetPositions()[0].GetSubjectId() != bea.GetId() {
		t.Fatalf("Jones positions %+v", pos.GetPositions())
	}
	unknown := uuid.Must(uuid.NewV7()).String()
	var none engine.ListSubjectsResponse
	mustRPC(t, ListSubjects, &engine.ListSubjectsRequest{ProjectDir: f.dir, GraphId: unknown}, &none)
	var nonePos engine.ListSubjectPositionsResponse
	mustRPC(t, ListSubjectPositions, &engine.ListSubjectPositionsRequest{ProjectDir: f.dir, GraphId: unknown}, &nonePos)
	if len(none.GetSubjects()) != 0 || len(nonePos.GetPositions()) != 0 {
		t.Fatalf("unknown graph read %+v %+v", none.GetSubjects(), nonePos.GetPositions())
	}
}

// A Subject created on a graph takes that graph's Source and reports only that
// graph; an unknown graph is refused.
func TestCreateSubjectOnGraph(t *testing.T) {
	f, jones, _ := onSplitSource(t)
	var subs engine.ListSubjectsResponse
	mustRPC(t, ListSubjects, &engine.ListSubjectsRequest{ProjectDir: f.dir, GraphId: f.defaultGraphID}, &subs)
	typeID := subs.Subjects[0].GetSubjectTypeId()
	var created engine.CreateSubjectResponse
	mustRPC(t, CreateSubject, &engine.CreateSubjectRequest{
		ProjectDir: f.dir, UserId: f.userID, GraphId: jones, SubjectTypeId: typeID, Label: "Cal",
	}, &created)
	if created.GetSubject().GetSourceId() != f.sourceID {
		t.Fatalf("source %s, want %s", created.GetSubject().GetSourceId(), f.sourceID)
	}
	if got := created.GetEffects().GetEvidenceGraphs(); !slices.Equal(got, []string{jones}) {
		t.Fatalf("evidence_graphs %v, want [%s]", got, jones)
	}
	var onJones engine.ListSubjectsResponse
	mustRPC(t, ListSubjects, &engine.ListSubjectsRequest{ProjectDir: f.dir, GraphId: jones}, &onJones)
	found := false
	for _, s := range onJones.GetSubjects() {
		found = found || s.GetId() == created.GetSubject().GetId()
	}
	if !found {
		t.Fatalf("Cal is not on Jones: %+v", onJones.GetSubjects())
	}
	var refused engine.CreateSubjectResponse
	err := rpc(t, CreateSubject, &engine.CreateSubjectRequest{
		ProjectDir: f.dir, UserId: f.userID, GraphId: uuid.Must(uuid.NewV7()).String(), SubjectTypeId: typeID, Label: "Dee",
	}, &refused)
	if !errors.Is(err, subjects.ErrInvalid) {
		t.Fatalf("unknown graph: got %v, want subjects.invalid", err)
	}
}
