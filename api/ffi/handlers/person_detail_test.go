package handlers

import (
	"errors"
	"slices"
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/catalogsession"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/graphcache/graphcachetest"
	"github.com/mendahu/provenencia/core/kinship"
	"github.com/mendahu/provenencia/core/onboarding"
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

// TestGetPersonFamily: two marriages and a child of each, written with the
// Canon builders inside the project's session, read over the wire.
func TestGetPersonFamily(t *testing.T) {
	t.Cleanup(func() { _ = catalogsession.CloseAll() })
	created, err := onboarding.Complete(t.TempDir(), t.TempDir(), "Jake", "Robins")
	if err != nil {
		t.Fatal(err)
	}
	dir := created.ProjectDir
	var pat, sam, tess, uma, wren, xan, m1, patUma []byte
	err = catalogsession.Do(dir, func(c *database.Catalog) error {
		k := graphcachetest.NewCanon(t, c)
		pat, sam, tess, uma, wren, xan = k.Person(), k.Person(), k.Person(), k.Person(), k.Person(), k.Person()
		var joins [][]byte
		m1, joins = k.Marriage(pat, sam)
		_, more := k.Marriage(pat, tess)
		joins = append(joins, more...)
		patUma = k.Parent(pat, uma)
		samUma := k.Parent(sam, uma)
		patWren := k.Parent(pat, wren)
		xanWren := k.Parent(xan, wren)
		ids := [][]byte{pat, sam, tess, uma, wren, xan, m1, patUma, samUma, patWren, xanWren}
		k.Commit(append(ids, joins...)...)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	var got engine.GetPersonFamilyResponse
	mustRPC(t, GetPersonFamily, &engine.GetPersonFamilyRequest{ProjectDir: dir, EntityId: database.UUIDString(pat)}, &got)
	if len(got.Units) != 3 {
		t.Fatalf("%d units, want 3: %+v", len(got.Units), got.Units)
	}
	for _, u := range got.Units {
		if u.GetKind() != engine.FamilyUnitKind_FAMILY_UNIT_KIND_PARTNER || len(u.Partners) != 1 || len(u.Children) > 1 {
			t.Fatalf("unit %+v", u)
		}
	}
	byPartner := map[string]*engine.FamilyUnit{}
	for _, u := range got.Units {
		byPartner[u.Partners[0].GetPerson().GetEntity().GetId()] = u
	}
	samUnit := byPartner[database.UUIDString(sam)]
	if samUnit == nil || len(samUnit.Children) != 1 {
		t.Fatalf("Sam's unit %+v", samUnit)
	}
	samMember := samUnit.Partners[0]
	if samMember.GetPerson().GetEntity().GetRef() == "" || samMember.GetLink() == nil || samMember.GetLink().GetRecorded() ||
		len(samMember.GetLink().GetImplied()) != 1 ||
		samMember.GetLink().GetImplied()[0].GetPattern() != engine.KinshipPattern_KINSHIP_PATTERN_MARRIAGE_SPOUSES ||
		samMember.GetLink().GetImplied()[0].GetPivotId() != database.UUIDString(m1) {
		t.Fatalf("Sam's link %+v", samMember)
	}
	umaMember := samUnit.Children[0]
	if umaMember.GetPerson().GetEntity().GetId() != database.UUIDString(uma) || !umaMember.GetLink().GetRecorded() ||
		!slices.Contains(umaMember.GetLink().GetAssociationIds(), database.UUIDString(patUma)) {
		t.Fatalf("Uma's member %+v", umaMember)
	}
	xanUnit := byPartner[database.UUIDString(xan)]
	if xanUnit == nil || xanUnit.Partners[0].GetLink() != nil || len(xanUnit.Children) != 1 {
		t.Fatalf("Xan's unit %+v", xanUnit)
	}

	for _, id := range []string{"not-an-id", "00000000-0000-7000-8000-000000000000"} {
		err := rpc(t, GetPersonFamily, &engine.GetPersonFamilyRequest{ProjectDir: dir, EntityId: id}, &engine.GetPersonFamilyResponse{})
		var ae *apperr.Error
		if !errors.As(err, &ae) || ae.Code() != apperr.CodeConclusionDetailsNotFound {
			t.Fatalf("id %q: err = %v, want %s", id, err, apperr.CodeConclusionDetailsNotFound)
		}
	}

	t.Run("carries the person and the origin siblings", func(t *testing.T) {
		created, err := onboarding.Complete(t.TempDir(), t.TempDir(), "Jake", "Robins")
		if err != nil {
			t.Fatal(err)
		}
		dir := created.ProjectDir
		var person, sibling, partner []byte
		err = catalogsession.Do(dir, func(c *database.Catalog) error {
			k := graphcachetest.NewCanon(t, c)
			person, sibling, partner = k.Person(), k.Person(), k.Person()
			father, mother := k.Person(), k.Person()
			ev1, j1 := k.Birth(person, father, mother)
			ev2, j2 := k.Birth(sibling, father, mother)
			m, j3 := k.Marriage(person, partner)
			ids := [][]byte{person, sibling, partner, father, mother, ev1, ev2, m}
			k.Commit(append(append(append(ids, j1...), j2...), j3...)...)
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		var got engine.GetPersonFamilyResponse
		mustRPC(t, GetPersonFamily, &engine.GetPersonFamilyRequest{ProjectDir: dir, EntityId: database.UUIDString(person)}, &got)
		if got.GetPerson().GetEntity().GetId() != database.UUIDString(person) {
			t.Fatalf("person %+v", got.GetPerson())
		}
		if len(got.OriginSiblings) != 2 {
			t.Fatalf("%d origin siblings, want 2: %+v", len(got.OriginSiblings), got.OriginSiblings)
		}
		ids := map[string]*engine.OriginSibling{}
		for _, o := range got.OriginSiblings {
			ids[o.GetMember().GetPerson().GetEntity().GetId()] = o
			if len(o.GetParentSets()) != 1 || o.GetParentSets()[0] != 0 {
				t.Fatalf("parent_sets %v, want [0]", o.GetParentSets())
			}
		}
		if o := ids[database.UUIDString(person)]; o == nil || o.GetMember().GetLink() != nil {
			t.Fatalf("the person's member %+v: want present, with no link", o)
		}
		if o := ids[database.UUIDString(sibling)]; o == nil || o.GetMember().GetLink() == nil {
			t.Fatalf("the sibling's member %+v: want present, with a link", o)
		}
		var partnerUnits int
		for _, u := range got.Units {
			if u.GetKind() == engine.FamilyUnitKind_FAMILY_UNIT_KIND_PARTNER &&
				u.Partners[0].GetPerson().GetEntity().GetId() == database.UUIDString(partner) {
				partnerUnits++
			}
		}
		if partnerUnits != 1 {
			t.Fatalf("%d partner units for the partner, want 1", partnerUnits)
		}
	})
}

// TestKinshipPatternProto: every pattern code the engine can put on a link has
// a wire value, so a new code cannot reach the app as unspecified.
func TestKinshipPatternProto(t *testing.T) {
	for _, p := range append(kinship.Patterns(), kinship.PatternSharedParent) {
		if kinshipPatternProto(p) == engine.KinshipPattern_KINSHIP_PATTERN_UNSPECIFIED {
			t.Errorf("pattern %q maps to unspecified", p)
		}
	}
}
