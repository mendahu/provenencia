package handlers

import (
	"errors"
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/catalogsession"
	"github.com/mendahu/provenencia/core/database"
	"google.golang.org/protobuf/proto"
)

// TestListPersonClaims: a Person with a claim on each of two Evidence graphs
// of one Source (a likeness on each, a Subject-valued Observation on the
// first) reads as one group of two claims, the second read touches no graph
// query, and an id that is not a Person handle reads as not found.
func TestListPersonClaims(t *testing.T) {
	t.Cleanup(func() { _ = catalogsession.CloseAll() })
	dir, userID, sourceID, artifactID, adaID, likenessID := likenessFixture(t)
	personPropID := propertyIDByKey(t, dir, "person")
	baptisms := soleGraphID(t, dir, sourceID)

	// Ada is on the Source's sole graph. A second graph, "Burials", holds Bea.
	var created engine.CreateEvidenceGraphResponse
	mustRPC(t, CreateEvidenceGraph, &engine.CreateEvidenceGraphRequest{
		ProjectDir: dir, UserId: userID, SourceId: sourceID, Name: "Burials", ArtifactIds: []string{artifactID},
	}, &created)
	burials := created.GetGraph().GetId()
	var subjTypes engine.ListSubjectTypesResponse
	mustRPC(t, ListSubjectTypes, &engine.ListSubjectTypesRequest{ProjectDir: dir}, &subjTypes)
	personTypeID := ""
	for _, typ := range subjTypes.Types {
		if typ.GetKey() == "person" {
			personTypeID = typ.GetId()
		}
	}
	// A Person Subject carries no Subject-valued Property until one is assigned.
	mustRPC(t, AssignSubjectTypeProperty, &engine.AssignSubjectTypePropertyRequest{
		ProjectDir: dir, UserId: userID, SubjectTypeId: personTypeID, PropertyId: personPropID,
	}, &engine.AssignSubjectTypePropertyResponse{})
	newSubject := func(graphID, label string) string {
		var s engine.CreateSubjectResponse
		mustRPC(t, CreateSubject, &engine.CreateSubjectRequest{
			ProjectDir: dir, UserId: userID, GraphId: graphID, SubjectTypeId: personTypeID, Label: label,
		}, &s)
		return s.Subject.GetId()
	}
	beaID := newSubject(burials, "Bea")
	cyID := newSubject(baptisms, "Cy")

	cite := func(graphID string, drafts ...*engine.ObservationDraft) {
		var out engine.CreateCitationWithObservationsResponse
		mustRPC(t, CreateCitationWithObservations, &engine.CreateCitationWithObservationsRequest{
			ProjectDir: dir, UserId: userID, ArtifactId: artifactID, GraphId: graphID, LocatorJson: likenessLocatorJSON, Observations: drafts,
		}, &out)
	}
	cite(baptisms,
		likenessDraft(adaID, likenessID, likenessRegionJSON, writeCrop(t, 300, 200, 128)),
		&engine.ObservationDraft{SubjectId: adaID, PropertyId: personPropID, ValueSubjectId: cyID})
	cite(burials, likenessDraft(beaID, likenessID, likenessRegion2JSON, writeCrop(t, 300, 200, 64)))

	var ada engine.PromoteSubjectResponse
	mustRPC(t, PromoteSubject, &engine.PromoteSubjectRequest{ProjectDir: dir, UserId: userID, SubjectId: adaID}, &ada)
	mustRPC(t, PromoteSubject, &engine.PromoteSubjectRequest{ProjectDir: dir, UserId: userID, SubjectId: beaID, EntityId: ada.Entity.GetId()}, &engine.PromoteSubjectResponse{})

	req := &engine.ListPersonClaimsRequest{ProjectDir: dir, EntityId: ada.Entity.GetId()}
	var got engine.ListPersonClaimsResponse
	mustRPC(t, ListPersonClaims, req, &got)
	if len(got.Groups) != 1 {
		t.Fatalf("%d groups, want 1", len(got.Groups))
	}
	g := got.Groups[0]
	if g.GetSourceId() != sourceID || g.GetSourceRef() == "" || g.GetSourceTitle() == "" || g.GetSourceTypeId() == "" {
		t.Fatalf("group %+v", g)
	}
	if len(g.Claims) != 2 {
		t.Fatalf("%d claims, want 2", len(g.Claims))
	}
	byGraph := map[string]*engine.PersonClaim{}
	for _, cl := range g.Claims {
		byGraph[cl.GetGraphId()] = cl
	}
	first, second := byGraph[baptisms], byGraph[burials]
	if first == nil || second == nil || g.Claims[0].GetGraphName() >= g.Claims[1].GetGraphName() {
		t.Fatalf("claims %+v", g.Claims)
	}
	if first.GetClaimId() == "" || first.GetGraphRef() == "" || first.GetSubject().GetId() != adaID || first.GetSubject().GetLabel() != "Ada" {
		t.Fatalf("first claim %+v", first)
	}
	if second.GetSubject().GetId() != beaID || len(second.Observations) != 1 {
		t.Fatalf("second claim %+v", second)
	}
	var sawValueSubject bool
	var likenessRows int
	for _, o := range first.Observations {
		if o.GetSubjectId() != adaID {
			t.Fatalf("Observation of another Subject: %+v", o)
		}
		if o.GetValueSubjectId() == cyID {
			sawValueSubject = true
			if o.GetValueText() != "Cy" {
				t.Fatalf("Subject-valued Observation reads %q, want the other Subject's label", o.GetValueText())
			}
		}
		if o.GetPropertyValueType() == "depiction" {
			likenessRows++
			if o.GetDepiction().GetImageRelPath() == "" {
				t.Fatalf("depiction without a path: %+v", o)
			}
		}
	}
	if !sawValueSubject || likenessRows != 1 || len(first.Observations) != 2 {
		t.Fatalf("first claim Observations %+v", first.Observations)
	}
	if rows := second.Observations[0]; rows.GetPropertyValueType() != "depiction" || rows.GetDepiction().GetImageRelPath() == "" {
		t.Fatalf("second claim Observation %+v", rows)
	}

	var n int
	if err := catalogsession.Do(dir, func(c *database.Catalog) error {
		c.Graph().SetQueryCounterForTest(func() { n++ })
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	var again engine.ListPersonClaimsResponse
	mustRPC(t, ListPersonClaims, req, &again)
	if n != 0 {
		t.Fatalf("second read issued %d graph queries", n)
	}
	if !proto.Equal(&got, &again) {
		t.Fatalf("second read differs:\n%+v\n%+v", &got, &again)
	}

	for _, id := range []string{"not-an-id", "00000000-0000-7000-8000-000000000000"} {
		err := rpc(t, ListPersonClaims, &engine.ListPersonClaimsRequest{ProjectDir: dir, EntityId: id}, &engine.ListPersonClaimsResponse{})
		var ae *apperr.Error
		if !errors.As(err, &ae) || ae.Code() != apperr.CodeConclusionDetailsNotFound {
			t.Fatalf("id %q: err = %v, want %s", id, err, apperr.CodeConclusionDetailsNotFound)
		}
	}
	if _, err := ListPersonClaims([]byte{0xff}); err == nil {
		t.Fatal("a bad proto was accepted")
	}
}
