package handlers

import (
	"strings"
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/locator"
	"google.golang.org/protobuf/proto"
)

const validLocatorJSON = `{"version":1,"selectors":[{"type":"page","artifact_page":12,"page_label":"10"}]}`

func citationFixture(t *testing.T) (projectDir, userID, sourceID, artifactID, placeSubjectID, toponymPropertyID string) {
	t.Helper()
	dir, userID, sourceID, _ := subjectFixture(t)
	aout, err := CreateArtifact(marshalProto(t, &engine.CreateArtifactRequest{
		ProjectDir: dir,
		UserId:     userID,
		SourceId:   sourceID,
		Label:      "Scan",
	}))
	if err != nil {
		t.Fatal(err)
	}
	var art engine.CreateArtifactResponse
	if err := proto.Unmarshal(aout, &art); err != nil {
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
	placeTypeID := ""
	for _, typ := range types.Types {
		if typ.GetKey() == "place" {
			placeTypeID = typ.GetId()
			break
		}
	}
	if placeTypeID == "" {
		t.Fatal("place subject type missing")
	}
	sout, err := CreateSubject(marshalProto(t, &engine.CreateSubjectRequest{
		ProjectDir:    dir,
		UserId:        userID,
		SourceId:      sourceID,
		SubjectTypeId: placeTypeID,
		Label:         "Boston",
	}))
	if err != nil {
		t.Fatal(err)
	}
	var subj engine.CreateSubjectResponse
	if err := proto.Unmarshal(sout, &subj); err != nil {
		t.Fatal(err)
	}
	propsOut, err := ListProperties(marshalProto(t, &engine.ListPropertiesRequest{ProjectDir: dir}))
	if err != nil {
		t.Fatal(err)
	}
	var props engine.ListPropertiesResponse
	if err := proto.Unmarshal(propsOut, &props); err != nil {
		t.Fatal(err)
	}
	toponymID := ""
	for _, p := range props.Properties {
		if p.GetKey() == "toponym" {
			toponymID = p.GetId()
			break
		}
	}
	if toponymID == "" {
		t.Fatal("toponym property missing")
	}
	return dir, userID, sourceID, art.Artifact.GetId(), subj.Subject.GetId(), toponymID
}

func TestCreateCitationWithObservations(t *testing.T) {
	runRPC(t, CreateCitationWithObservations, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "creates citation with zero observations",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _, artifactID, _, _ := citationFixture(t)
				return &engine.CreateCitationWithObservationsRequest{
					ProjectDir:    dir,
					UserId:        userID,
					ArtifactId:    artifactID,
					LocatorJson:   validLocatorJSON,
					Transcription: "transcribe first",
				}
			},
			after: func(t *testing.T, out []byte, req proto.Message) {
				var created engine.CreateCitationWithObservationsResponse
				if err := proto.Unmarshal(out, &created); err != nil {
					t.Fatal(err)
				}
				if !strings.HasPrefix(created.Citation.GetRef(), "CIT-") {
					t.Fatalf("citation ref %q", created.Citation.GetRef())
				}
				if len(created.Observations) != 0 {
					t.Fatalf("observations %+v", created.Observations)
				}
				if created.Citation.GetTranscription() != "transcribe first" {
					t.Fatalf("transcription %q", created.Citation.GetTranscription())
				}
				cr := req.(*engine.CreateCitationWithObservationsRequest)
				assertLatestAuditAction(t, cr.ProjectDir, "create_citation_with_observations")
			},
		},
		{
			name: "bad locator",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _, artifactID, placeID, propID := citationFixture(t)
				return &engine.CreateCitationWithObservationsRequest{
					ProjectDir:  dir,
					UserId:      userID,
					ArtifactId:  artifactID,
					LocatorJson: `{}`,
					Observations: []*engine.ObservationDraft{{
						SubjectId:  placeID,
						PropertyId: propID,
						ValueText:  "Boston",
					}},
				}
			},
			wantErr:   true,
			wantErrIs: locator.ErrInvalid,
		},
		{
			name: "creates citation with text observation",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _, artifactID, placeID, propID := citationFixture(t)
				return &engine.CreateCitationWithObservationsRequest{
					ProjectDir:    dir,
					UserId:        userID,
					ArtifactId:    artifactID,
					LocatorJson:   validLocatorJSON,
					Transcription: "Boston",
					Observations: []*engine.ObservationDraft{{
						SubjectId:  placeID,
						PropertyId: propID,
						ValueText:  "Boston",
					}},
				}
			},
			after: func(t *testing.T, out []byte, req proto.Message) {
				var created engine.CreateCitationWithObservationsResponse
				if err := proto.Unmarshal(out, &created); err != nil {
					t.Fatal(err)
				}
				if !strings.HasPrefix(created.Citation.GetRef(), "CIT-") {
					t.Fatalf("citation ref %q", created.Citation.GetRef())
				}
				if len(created.Observations) != 1 {
					t.Fatalf("observations %+v", created.Observations)
				}
				if !strings.HasPrefix(created.Observations[0].GetRef(), "OBS-") {
					t.Fatalf("obs ref %q", created.Observations[0].GetRef())
				}
				if created.Observations[0].GetValueText() != "Boston" {
					t.Fatalf("value %+v", created.Observations[0])
				}
				cr := req.(*engine.CreateCitationWithObservationsRequest)
				assertLatestAuditAction(t, cr.ProjectDir, "create_citation_with_observations")
			},
		},
	})
}

func TestAddObservationsToCitation(t *testing.T) {
	runRPC(t, AddObservationsToCitation, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "appends observation",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _, artifactID, placeID, propID := citationFixture(t)
				createOut, err := CreateCitationWithObservations(marshalProto(t, &engine.CreateCitationWithObservationsRequest{
					ProjectDir:  dir,
					UserId:      userID,
					ArtifactId:  artifactID,
					LocatorJson: validLocatorJSON,
					Observations: []*engine.ObservationDraft{{
						SubjectId:  placeID,
						PropertyId: propID,
						ValueText:  "Cambridge",
					}},
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateCitationWithObservationsResponse
				if err := proto.Unmarshal(createOut, &created); err != nil {
					t.Fatal(err)
				}
				return &engine.AddObservationsToCitationRequest{
					ProjectDir: dir,
					UserId:     userID,
					CitationId: created.Citation.GetId(),
					Observations: []*engine.ObservationDraft{{
						SubjectId:  placeID,
						PropertyId: propID,
						ValueText:  "MA",
					}},
				}
			},
			after: func(t *testing.T, out []byte, req proto.Message) {
				var added engine.AddObservationsToCitationResponse
				if err := proto.Unmarshal(out, &added); err != nil {
					t.Fatal(err)
				}
				if len(added.Observations) != 1 || added.Observations[0].GetValueText() != "MA" {
					t.Fatalf("%+v", added.Observations)
				}
				ar := req.(*engine.AddObservationsToCitationRequest)
				assertLatestAuditAction(t, ar.ProjectDir, "add_observations")
			},
		},
	})
}

func TestListObservationsBySource(t *testing.T) {
	runRPC(t, ListObservationsBySource, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "lists by source",
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
				return &engine.ListObservationsBySourceRequest{
					ProjectDir: dir,
					SourceId:   sourceID,
				}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var list engine.ListObservationsBySourceResponse
				if err := proto.Unmarshal(out, &list); err != nil {
					t.Fatal(err)
				}
				if len(list.Observations) != 1 {
					t.Fatalf("%+v", list.Observations)
				}
				if list.Observations[0].GetPropertyKey() != "toponym" {
					t.Fatalf("property summary %+v", list.Observations[0])
				}
				if list.Observations[0].GetValueText() != "Boston" {
					t.Fatalf("value %+v", list.Observations[0])
				}
			},
		},
	})
}

func TestCitationCountsBySource(t *testing.T) {
	runRPC(t, CitationCountsBySource, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "counts citations per artifact",
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
				if _, err := CreateCitationWithObservations(marshalProto(t, &engine.CreateCitationWithObservationsRequest{
					ProjectDir:  dir,
					UserId:      userID,
					ArtifactId:  artifactID,
					LocatorJson: validLocatorJSON,
					Observations: []*engine.ObservationDraft{{
						SubjectId:  placeID,
						PropertyId: propID,
						ValueText:  "Boston again",
					}},
				})); err != nil {
					t.Fatal(err)
				}
				return &engine.CitationCountsBySourceRequest{
					ProjectDir: dir,
					SourceId:   sourceID,
				}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var list engine.CitationCountsBySourceResponse
				if err := proto.Unmarshal(out, &list); err != nil {
					t.Fatal(err)
				}
				if len(list.Counts) != 1 || list.Counts[0].GetCount() != 2 {
					t.Fatalf("%+v", list.Counts)
				}
			},
		},
	})
}

func TestListCitationsByArtifact(t *testing.T) {
	runRPC(t, ListCitationsByArtifact, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "lists citations with observation counts",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _, artifactID, placeID, propID := citationFixture(t)
				if _, err := CreateCitationWithObservations(marshalProto(t, &engine.CreateCitationWithObservationsRequest{
					ProjectDir:    dir,
					UserId:        userID,
					ArtifactId:    artifactID,
					LocatorJson:   validLocatorJSON,
					Transcription: "empty reading",
				})); err != nil {
					t.Fatal(err)
				}
				if _, err := CreateCitationWithObservations(marshalProto(t, &engine.CreateCitationWithObservationsRequest{
					ProjectDir:    dir,
					UserId:        userID,
					ArtifactId:    artifactID,
					LocatorJson:   validLocatorJSON,
					Transcription: "Boston",
					Observations: []*engine.ObservationDraft{{
						SubjectId:  placeID,
						PropertyId: propID,
						ValueText:  "Boston",
					}},
				})); err != nil {
					t.Fatal(err)
				}
				return &engine.ListCitationsByArtifactRequest{
					ProjectDir: dir,
					ArtifactId: artifactID,
				}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var list engine.ListCitationsByArtifactResponse
				if err := proto.Unmarshal(out, &list); err != nil {
					t.Fatal(err)
				}
				if len(list.Citations) != 2 {
					t.Fatalf("%+v", list.Citations)
				}
				counts := map[int32]int{}
				for _, row := range list.Citations {
					counts[row.GetObservationCount()]++
					if !strings.HasPrefix(row.GetCitation().GetRef(), "CIT-") {
						t.Fatalf("ref %q", row.GetCitation().GetRef())
					}
				}
				if counts[0] != 1 || counts[1] != 1 {
					t.Fatalf("counts %+v", counts)
				}
			},
		},
	})
}

func TestGetCitation(t *testing.T) {
	runRPC(t, GetCitation, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "loads citation with observations",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _, artifactID, placeID, propID := citationFixture(t)
				createOut, err := CreateCitationWithObservations(marshalProto(t, &engine.CreateCitationWithObservationsRequest{
					ProjectDir:    dir,
					UserId:        userID,
					ArtifactId:    artifactID,
					LocatorJson:   validLocatorJSON,
					Transcription: "Boston",
					CitationNotes: []string{"note A"},
					Observations: []*engine.ObservationDraft{{
						SubjectId:  placeID,
						PropertyId: propID,
						ValueText:  "Boston",
					}},
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateCitationWithObservationsResponse
				if err := proto.Unmarshal(createOut, &created); err != nil {
					t.Fatal(err)
				}
				return &engine.GetCitationRequest{
					ProjectDir: dir,
					CitationId: created.Citation.GetId(),
				}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var got engine.GetCitationResponse
				if err := proto.Unmarshal(out, &got); err != nil {
					t.Fatal(err)
				}
				if got.Citation.GetTranscription() != "Boston" {
					t.Fatalf("citation %+v", got.Citation)
				}
				if len(got.Notes) != 1 || got.Notes[0] != "note A" {
					t.Fatalf("notes %+v", got.Notes)
				}
				if len(got.Observations) != 1 || got.Observations[0].GetValueText() != "Boston" {
					t.Fatalf("observations %+v", got.Observations)
				}
			},
		},
	})
}

func TestUpdateCitationWithObservations(t *testing.T) {
	runRPC(t, UpdateCitationWithObservations, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "replaces fields and observations",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _, artifactID, placeID, propID := citationFixture(t)
				createOut, err := CreateCitationWithObservations(marshalProto(t, &engine.CreateCitationWithObservationsRequest{
					ProjectDir:  dir,
					UserId:      userID,
					ArtifactId:  artifactID,
					LocatorJson: validLocatorJSON,
					Observations: []*engine.ObservationDraft{{
						SubjectId:  placeID,
						PropertyId: propID,
						ValueText:  "Boston",
					}},
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateCitationWithObservationsResponse
				if err := proto.Unmarshal(createOut, &created); err != nil {
					t.Fatal(err)
				}
				if len(created.Observations) != 1 {
					t.Fatalf("created observations %+v", created.Observations)
				}
				return &engine.UpdateCitationWithObservationsRequest{
					ProjectDir:    dir,
					UserId:        userID,
					CitationId:    created.Citation.GetId(),
					ArtifactId:    artifactID,
					LocatorJson:   validLocatorJSON,
					Transcription: "Salem",
					Observations: []*engine.Observation{{
						Id:         created.Observations[0].GetId(),
						SubjectId:  placeID,
						PropertyId: propID,
						ValueText:  "Salem",
					}},
				}
			},
			after: func(t *testing.T, out []byte, req proto.Message) {
				var updated engine.UpdateCitationWithObservationsResponse
				if err := proto.Unmarshal(out, &updated); err != nil {
					t.Fatal(err)
				}
				if updated.Citation.GetTranscription() != "Salem" {
					t.Fatalf("citation %+v", updated.Citation)
				}
				ur := req.(*engine.UpdateCitationWithObservationsRequest)
				if len(updated.Observations) != 1 || updated.Observations[0].GetValueText() != "Salem" {
					t.Fatalf("observations %+v", updated.Observations)
				}
				if updated.Observations[0].GetId() != ur.Observations[0].GetId() {
					t.Fatalf("id %s want %s", updated.Observations[0].GetId(), ur.Observations[0].GetId())
				}
				assertLatestAuditAction(t, ur.ProjectDir, "update_citation_with_observations")
				getOut, err := GetCitation(marshalProto(t, &engine.GetCitationRequest{
					ProjectDir: ur.ProjectDir,
					CitationId: ur.CitationId,
				}))
				if err != nil {
					t.Fatal(err)
				}
				var got engine.GetCitationResponse
				if err := proto.Unmarshal(getOut, &got); err != nil {
					t.Fatal(err)
				}
				if len(got.Observations) != 1 || got.Observations[0].GetValueText() != "Salem" {
					t.Fatalf("persisted %+v", got.Observations)
				}
				if got.Observations[0].GetId() != ur.Observations[0].GetId() {
					t.Fatalf("persisted id %s", got.Observations[0].GetId())
				}
			},
		},
	})
}

func TestUpdateCitation(t *testing.T) {
	runRPC(t, UpdateCitation, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "updates citation fields only",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _, artifactID, placeID, propID := citationFixture(t)
				createOut, err := CreateCitationWithObservations(marshalProto(t, &engine.CreateCitationWithObservationsRequest{
					ProjectDir:    dir,
					UserId:        userID,
					ArtifactId:    artifactID,
					LocatorJson:   validLocatorJSON,
					Transcription: "was",
					Observations: []*engine.ObservationDraft{{
						SubjectId:  placeID,
						PropertyId: propID,
						ValueText:  "Boston",
					}},
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateCitationWithObservationsResponse
				if err := proto.Unmarshal(createOut, &created); err != nil {
					t.Fatal(err)
				}
				return &engine.UpdateCitationRequest{
					ProjectDir:    dir,
					UserId:        userID,
					CitationId:    created.Citation.GetId(),
					LocatorJson:   validLocatorJSON,
					Transcription: "now",
				}
			},
			after: func(t *testing.T, out []byte, req proto.Message) {
				var updated engine.UpdateCitationResponse
				if err := proto.Unmarshal(out, &updated); err != nil {
					t.Fatal(err)
				}
				if updated.Citation.GetTranscription() != "now" {
					t.Fatalf("transcription %q", updated.Citation.GetTranscription())
				}
				ur := req.(*engine.UpdateCitationRequest)
				assertLatestAuditAction(t, ur.ProjectDir, "update_citation")
			},
		},
	})
}

func TestUpdateObservation(t *testing.T) {
	runRPC(t, UpdateObservation, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "updates one observation",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _, artifactID, placeID, propID := citationFixture(t)
				createOut, err := CreateCitationWithObservations(marshalProto(t, &engine.CreateCitationWithObservationsRequest{
					ProjectDir:  dir,
					UserId:      userID,
					ArtifactId:  artifactID,
					LocatorJson: validLocatorJSON,
					Observations: []*engine.ObservationDraft{{
						SubjectId:  placeID,
						PropertyId: propID,
						ValueText:  "Boston",
					}},
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateCitationWithObservationsResponse
				if err := proto.Unmarshal(createOut, &created); err != nil {
					t.Fatal(err)
				}
				obs := created.Observations[0]
				obs.ValueText = "Salem"
				return &engine.UpdateObservationRequest{
					ProjectDir:  dir,
					UserId:      userID,
					Observation: obs,
				}
			},
			after: func(t *testing.T, out []byte, req proto.Message) {
				var updated engine.UpdateObservationResponse
				if err := proto.Unmarshal(out, &updated); err != nil {
					t.Fatal(err)
				}
				if updated.Observation.GetValueText() != "Salem" || updated.Observation.GetPropertyKey() == "" {
					t.Fatalf("%+v", updated.Observation)
				}
				ur := req.(*engine.UpdateObservationRequest)
				assertLatestAuditAction(t, ur.ProjectDir, "update_observation")
			},
		},
	})
}

func TestDeleteObservation(t *testing.T) {
	runRPC(t, DeleteObservation, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "deletes observation and keeps citation",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _, artifactID, placeID, propID := citationFixture(t)
				createOut, err := CreateCitationWithObservations(marshalProto(t, &engine.CreateCitationWithObservationsRequest{
					ProjectDir:  dir,
					UserId:      userID,
					ArtifactId:  artifactID,
					LocatorJson: validLocatorJSON,
					Observations: []*engine.ObservationDraft{{
						SubjectId:  placeID,
						PropertyId: propID,
						ValueText:  "Boston",
					}},
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateCitationWithObservationsResponse
				if err := proto.Unmarshal(createOut, &created); err != nil {
					t.Fatal(err)
				}
				return &engine.DeleteObservationRequest{
					ProjectDir:    dir,
					UserId:        userID,
					ObservationId: created.Observations[0].GetId(),
				}
			},
			after: func(t *testing.T, out []byte, req proto.Message) {
				var deleted engine.DeleteObservationResponse
				if err := proto.Unmarshal(out, &deleted); err != nil {
					t.Fatal(err)
				}
				_ = deleted
				dr := req.(*engine.DeleteObservationRequest)
				assertLatestAuditAction(t, dr.ProjectDir, "delete_observation")
			},
		},
	})
}
