package handlers

import (
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/api/proto/engine"
	"google.golang.org/protobuf/proto"
)

func TestListEvidenceGraphProgress(t *testing.T) {
	runRPC(t, ListEvidenceGraphProgress, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "lists worked source and omits empty",
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
				typesOut, err := ListSourceTypes(marshalProto(t, &engine.ListSourceTypesRequest{ProjectDir: dir}))
				if err != nil {
					t.Fatal(err)
				}
				var types engine.ListSourceTypesResponse
				if err := proto.Unmarshal(typesOut, &types); err != nil {
					t.Fatal(err)
				}
				if len(types.Types) == 0 {
					t.Fatal("no source types")
				}
				if _, err := CreateSource(marshalProto(t, &engine.CreateSourceRequest{
					ProjectDir:   dir,
					UserId:       userID,
					SourceTypeId: types.Types[0].GetId(),
					Title:        "Empty",
				})); err != nil {
					t.Fatal(err)
				}
				return &engine.ListEvidenceGraphProgressRequest{ProjectDir: dir}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var list engine.ListEvidenceGraphProgressResponse
				if err := proto.Unmarshal(out, &list); err != nil {
					t.Fatal(err)
				}
				if len(list.Rows) != 1 {
					t.Fatalf("rows %+v", list.Rows)
				}
				if list.Rows[0].GetSubjectCount() != 1 || list.Rows[0].GetObservationCount() != 1 {
					t.Fatalf("%+v", list.Rows[0])
				}
				// The fixture subject is a place: primary, not yet filed.
				if list.Rows[0].GetFiledCount() != 0 {
					t.Fatalf("primary/filed %+v", list.Rows[0])
				}
			},
		},
	})
}

func TestGetEvidenceGraphProgress(t *testing.T) {
	runRPC(t, GetEvidenceGraphProgress, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "zeros for empty source",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, sourceID, _ := subjectFixture(t)
				return &engine.GetEvidenceGraphProgressRequest{
					ProjectDir: dir,
					GraphId:    soleGraphID(t, dir, sourceID),
				}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var got engine.GetEvidenceGraphProgressResponse
				if err := proto.Unmarshal(out, &got); err != nil {
					t.Fatal(err)
				}
				if got.Progress.GetSubjectCount() != 0 || got.Progress.GetObservationCount() != 0 {
					t.Fatalf("%+v", got.Progress)
				}
			},
		},
		{
			name: "zeros with the request's graph id and no Source for a graph no row names",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, _, _ := subjectFixture(t)
				return &engine.GetEvidenceGraphProgressRequest{ProjectDir: dir, GraphId: uuid.Must(uuid.NewV7()).String()}
			},
			after: func(t *testing.T, out []byte, req proto.Message) {
				var got engine.GetEvidenceGraphProgressResponse
				if err := proto.Unmarshal(out, &got); err != nil {
					t.Fatal(err)
				}
				want := &engine.EvidenceGraphProgress{GraphId: req.(*engine.GetEvidenceGraphProgressRequest).GetGraphId()}
				if !proto.Equal(got.Progress, want) {
					t.Fatalf("%+v, want %+v", got.Progress, want)
				}
			},
		},
		{
			name: "counts subjects and observations",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, sourceID, artifactID, placeID, propID := citationFixture(t)
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
				return &engine.GetEvidenceGraphProgressRequest{
					ProjectDir: dir,
					GraphId:    soleGraphID(t, dir, sourceID),
				}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var got engine.GetEvidenceGraphProgressResponse
				if err := proto.Unmarshal(out, &got); err != nil {
					t.Fatal(err)
				}
				if got.Progress.GetSubjectCount() != 1 || got.Progress.GetObservationCount() != 1 {
					t.Fatalf("%+v", got.Progress)
				}
			},
		},
	})
}

// Two graphs of one Source count apart, and each row names its graph and its Source.
func TestEvidenceGraphProgressPerGraph(t *testing.T) {
	f := newGraphFixture(t)
	jones := f.createGraph(t, "Jones", f.page2).GetGraph().GetId()
	var subs engine.ListSubjectsResponse
	mustRPC(t, ListSubjects, &engine.ListSubjectsRequest{ProjectDir: f.dir, GraphId: f.defaultGraphID}, &subs)
	var created engine.CreateSubjectResponse
	mustRPC(t, CreateSubject, &engine.CreateSubjectRequest{
		ProjectDir: f.dir, UserId: f.userID, GraphId: jones,
		SubjectTypeId: subs.Subjects[0].GetSubjectTypeId(), Label: "Bea",
	}, &created)

	var got engine.GetEvidenceGraphProgressResponse
	mustRPC(t, GetEvidenceGraphProgress, &engine.GetEvidenceGraphProgressRequest{ProjectDir: f.dir, GraphId: jones}, &got)
	if p := got.GetProgress(); p.GetGraphId() != jones || p.GetSourceId() != f.sourceID || p.GetSubjectCount() != 1 || p.GetObservationCount() != 0 {
		t.Fatalf("Jones progress %+v", p)
	}

	var list engine.ListEvidenceGraphProgressResponse
	mustRPC(t, ListEvidenceGraphProgress, &engine.ListEvidenceGraphProgressRequest{ProjectDir: f.dir}, &list)
	byGraph := map[string]*engine.EvidenceGraphProgress{}
	for _, row := range list.GetRows() {
		byGraph[row.GetGraphId()] = row
	}
	if len(byGraph) != 2 {
		t.Fatalf("rows %+v", list.GetRows())
	}
	for _, id := range []string{f.defaultGraphID, jones} {
		if row := byGraph[id]; row == nil || row.GetSourceId() != f.sourceID {
			t.Fatalf("row for %s: %+v", id, row)
		}
	}
}
