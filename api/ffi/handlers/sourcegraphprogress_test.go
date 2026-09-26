package handlers

import (
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
	"google.golang.org/protobuf/proto"
)

func TestListSourceGraphProgress(t *testing.T) {
	runRPC(t, ListSourceGraphProgress, []rpcTest{
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
				return &engine.ListSourceGraphProgressRequest{ProjectDir: dir}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var list engine.ListSourceGraphProgressResponse
				if err := proto.Unmarshal(out, &list); err != nil {
					t.Fatal(err)
				}
				if len(list.Rows) != 1 {
					t.Fatalf("rows %+v", list.Rows)
				}
				if list.Rows[0].GetSubjectCount() != 1 || list.Rows[0].GetObservationCount() != 1 {
					t.Fatalf("%+v", list.Rows[0])
				}
			},
		},
	})
}

func TestGetSourceGraphProgress(t *testing.T) {
	runRPC(t, GetSourceGraphProgress, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "zeros for empty source",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, sourceID, _ := subjectFixture(t)
				return &engine.GetSourceGraphProgressRequest{
					ProjectDir: dir,
					SourceId:   sourceID,
				}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var got engine.GetSourceGraphProgressResponse
				if err := proto.Unmarshal(out, &got); err != nil {
					t.Fatal(err)
				}
				if got.Progress.GetSubjectCount() != 0 || got.Progress.GetObservationCount() != 0 {
					t.Fatalf("%+v", got.Progress)
				}
			},
		},
		{
			name: "counts cited and uncited",
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
				return &engine.GetSourceGraphProgressRequest{
					ProjectDir: dir,
					SourceId:   sourceID,
				}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var got engine.GetSourceGraphProgressResponse
				if err := proto.Unmarshal(out, &got); err != nil {
					t.Fatal(err)
				}
				if got.Progress.GetSubjectCount() != 1 || got.Progress.GetObservationCount() != 1 || got.Progress.GetUncitedCount() != 0 {
					t.Fatalf("%+v", got.Progress)
				}
			},
		},
	})
}
