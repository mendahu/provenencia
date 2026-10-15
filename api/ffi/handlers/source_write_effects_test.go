package handlers

import (
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
	"google.golang.org/protobuf/proto"
)

// filedSubject creates a place Subject on a Source and promotes it, so the
// Source has a Subject filed to a handle. With observed, the Subject also has
// an Observation, which is what makes the Source's credibility part of the
// handle's value (and keeps the Subject from being deleted). It returns the
// Source, the Subject, and the handle (canonical entity) ids.
func filedSubject(t *testing.T, observed bool) (dir, userID, sourceID, subjectID, handleID string) {
	t.Helper()
	dir, userID, sourceID, artifactID, subjectID, propID := citationFixture(t)
	if observed {
		if _, err := CreateCitationWithObservations(marshalProto(t, &engine.CreateCitationWithObservationsRequest{
			ProjectDir: dir, UserId: userID, ArtifactId: artifactID, LocatorJson: validLocatorJSON,
			Observations: []*engine.ObservationDraft{{SubjectId: subjectID, PropertyId: propID, ValueText: "Boston"}},
		})); err != nil {
			t.Fatal(err)
		}
	}
	promoted := promoteOK(t, &engine.PromoteSubjectRequest{ProjectDir: dir, UserId: userID, SubjectId: subjectID})
	return dir, userID, sourceID, subjectID, promoted.GetEntity().GetId()
}

// TestSourceWritesReturnEffects: a Source-page or graph write tells the app
// what it touched, so no blanket trigger has to guess (S10-22 D-22).
func TestSourceWritesReturnEffects(t *testing.T) {
	tests := []struct {
		name        string
		run         func(t *testing.T) (effects *engine.WriteEffects, sourceID, handleID string)
		wantHandles bool // the filed handle is in effects.handles
	}{
		{
			name: "a note on a Source reports the Source and no handle",
			run: func(t *testing.T) (*engine.WriteEffects, string, string) {
				dir, userID, sourceID, _, handleID := filedSubject(t, false)
				out, err := AddSourceNote(marshalProto(t, &engine.AddSourceNoteRequest{
					ProjectDir: dir, UserId: userID, SourceId: sourceID, Body: "seen at the archive",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var resp engine.AddSourceNoteResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				return resp.GetEffects(), sourceID, handleID
			},
		},
		{
			name: "a credibility change reports the handles filed from the Source",
			run: func(t *testing.T) (*engine.WriteEffects, string, string) {
				dir, userID, sourceID, _, handleID := filedSubject(t, true)
				gout, err := ListSourceCredibilityGrades(marshalProto(t, &engine.ListSourceCredibilityGradesRequest{ProjectDir: dir}))
				if err != nil {
					t.Fatal(err)
				}
				var grades engine.ListSourceCredibilityGradesResponse
				if err := proto.Unmarshal(gout, &grades); err != nil {
					t.Fatal(err)
				}
				out, err := UpsertSourceCredibilityAssessment(marshalProto(t, &engine.UpsertSourceCredibilityAssessmentRequest{
					ProjectDir: dir, UserId: userID, SourceId: sourceID, GradeId: grades.Grades[2].GetId(), Argument: "firsthand",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var resp engine.UpsertSourceCredibilityAssessmentResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				return resp.GetEffects(), sourceID, handleID
			},
			wantHandles: true,
		},
		{
			name: "deleting a filed Subject reports its handle and Source",
			run: func(t *testing.T) (*engine.WriteEffects, string, string) {
				dir, userID, sourceID, subjectID, handleID := filedSubject(t, false)
				out, err := DeleteSubject(marshalProto(t, &engine.DeleteSubjectRequest{
					ProjectDir: dir, UserId: userID, SubjectId: subjectID,
				}))
				if err != nil {
					t.Fatal(err)
				}
				var resp engine.DeleteSubjectResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				return resp.GetEffects(), sourceID, handleID
			},
			wantHandles: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eff, sourceID, handleID := tt.run(t)
			if eff.GetRevision() <= 0 || eff.GetRevisionAt() == "" {
				t.Fatalf("activity = %d %q", eff.GetRevision(), eff.GetRevisionAt())
			}
			if !contains(eff.GetSources(), sourceID) {
				t.Fatalf("sources = %v, want %s", eff.GetSources(), sourceID)
			}
			if got := contains(eff.GetHandles(), handleID); got != tt.wantHandles {
				t.Fatalf("handles = %v, handle %s in handles = %v, want %v", eff.GetHandles(), handleID, got, tt.wantHandles)
			}
		})
	}
}

func contains(ids []string, id string) bool {
	for _, got := range ids {
		if got == id {
			return true
		}
	}
	return false
}
