package handlers

import (
	"errors"
	"reflect"
	"slices"
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/catalogsession"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/conclusiondetails"
	"github.com/mendahu/provenencia/core/database/evidencegraphs"
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
	"github.com/mendahu/provenencia/core/database/evidencelayer/evidencelayertest"
	"github.com/mendahu/provenencia/core/database/graphcache/graphcachetest"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/promote/promotetest"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/datevalue/datevaluetest"
	"github.com/mendahu/provenencia/core/kinship"
	"github.com/mendahu/provenencia/core/onboarding"
	"github.com/mendahu/provenencia/core/recordlinks"
	"github.com/mendahu/provenencia/core/vocabkey"
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
	personTypeID := subjectTypeIDByKey(t, dir, "person")
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

	ada := promoteInDir(t, dir, userID, adaID, "")
	promoteInDir(t, dir, userID, beaID, database.UUIDString(ada.Entity.ID))

	req := &engine.ListPersonClaimsRequest{ProjectDir: dir, EntityId: database.UUIDString(ada.Entity.ID)}
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

// TestListPersonEvents covers AC-3.1 and AC-3.2: a person's events in date
// order, each with the person's roles and the other canonical participants.
func TestListPersonEvents(t *testing.T) {
	t.Cleanup(func() { _ = catalogsession.CloseAll() })
	created, err := onboarding.Complete(t.TempDir(), t.TempDir(), "Jake", "Robins")
	if err != nil {
		t.Fatal(err)
	}
	dir := created.ProjectDir
	var pat, ned, kid, sam, b1, b2, u, mar, cs []byte
	err = catalogsession.Do(dir, func(c *database.Catalog) error {
		k := graphcachetest.NewCanon(t, c)
		pat, ned, kid, sam = k.Person(), k.Person(), k.Person(), k.Person()
		var ids [][]byte
		var j [][]byte
		b1, j = k.Birth(pat, nil, nil)
		k.Date(b1, vocabkey.PropertyDate, *datevaluetest.YMD(1820, 0, 0))
		ids = append(ids, j...)
		mar, j = k.Marriage(pat, sam)
		k.Date(mar, vocabkey.PropertyDate, *datevaluetest.YMD(1845, 0, 0))
		ids = append(ids, mar)
		ids = append(ids, j...)
		cs = k.Event(vocabkey.EventType("census"))
		k.Date(cs, vocabkey.PropertyDate, *datevaluetest.YMD(1851, 0, 0))
		ids = append(ids, cs, k.Join(pat, cs, "subject"), k.Join(kid, cs, "subject"))
		u = k.Event(vocabkey.EventBurial)
		k.Date(u, vocabkey.PropertyDate, *datevaluetest.YMD(1890, 0, 0))
		ids = append(ids, u, k.Join(ned, u, "subject"), k.Join(pat, u, "witness"))
		b2, j = k.Birth(kid, pat, sam)
		k.Date(b2, vocabkey.PropertyDate, *datevaluetest.YMD(1850, 0, 0))
		ids = append(ids, b2)
		ids = append(ids, j...)
		k.Commit(append(ids, pat, ned, kid, sam, b1)...)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	var got engine.ListPersonEventsResponse
	mustRPC(t, ListPersonEvents, &engine.ListPersonEventsRequest{ProjectDir: dir, EntityId: database.UUIDString(pat)}, &got)
	if len(got.Events) != 5 {
		t.Fatalf("%d events, want 5: %+v", len(got.Events), got.Events)
	}
	wantOrder := []string{database.UUIDString(b1), database.UUIDString(mar), database.UUIDString(b2), database.UUIDString(cs), database.UUIDString(u)}
	for i, e := range got.Events {
		if e.GetSpan() == nil || e.GetEvent().GetEntity().GetRef() == "" {
			t.Errorf("event %d: %+v", i, e)
		}
		if e.GetEvent().GetEntity().GetId() != wantOrder[i] {
			t.Errorf("event %d is %s, want %s", i, e.GetEvent().GetEntity().GetId(), wantOrder[i])
		}
	}
	last := got.Events[4]
	if len(last.Roles) != 1 || last.Roles[0].GetKey() != "witness" || last.Roles[0].GetLabel() != "Witness" {
		t.Fatalf("burial roles %+v", last.Roles)
	}
	if len(last.Participants) != 1 || last.Participants[0].GetPerson().GetEntity().GetId() != database.UUIDString(ned) ||
		len(last.Participants[0].Roles) != 1 || last.Participants[0].Roles[0].GetKey() != "subject" {
		t.Fatalf("burial participants %+v", last.Participants)
	}
	births := got.Events[2]
	if len(births.Participants) != 2 || births.Participants[0].GetPerson().GetEntity().GetId() == births.Participants[1].GetPerson().GetEntity().GetId() {
		t.Fatalf("B2 participants %+v", births.Participants)
	}
	for _, p := range births.Participants {
		id := p.GetPerson().GetEntity().GetId()
		if id != database.UUIDString(kid) && id != database.UUIDString(sam) {
			t.Fatalf("B2 participant %s", id)
		}
	}

	if got.GetBirth().GetEventId() != database.UUIDString(b1) || got.GetBirth().GetBesideEventId() != "" {
		t.Errorf("birth marker %+v", got.GetBirth())
	}
	if got.GetDeath() == nil || got.GetDeath().GetEventId() != "" || got.GetDeath().GetBesideEventId() != database.UUIDString(cs) {
		t.Errorf("death marker %+v, want a placeholder beside the census (the burial may follow death)", got.GetDeath())
	}
	if sp := got.Events[0].GetSpan(); sp.GetFrom() != 1820 || sp.GetTo() != 1821 || sp.GetOpenStart() || sp.GetOpenEnd() {
		t.Errorf("birth span %+v", sp)
	}
	if a := got.Events[0].GetAge(); a == nil || a.GetLo() != 0 || a.GetHi() != 0 {
		t.Errorf("birth age %+v", a)
	}

	t.Run("carries spans and markers", func(t *testing.T) {
		created, err := onboarding.Complete(t.TempDir(), t.TempDir(), "Jake", "Robins")
		if err != nil {
			t.Fatal(err)
		}
		dir := created.ProjectDir
		var lee, born, resided []byte
		err = catalogsession.Do(dir, func(c *database.Catalog) error {
			k := graphcachetest.NewCanon(t, c)
			lee = k.Person()
			var j []byte
			born, _ = k.Birth(lee, nil, nil)
			k.Date(born, vocabkey.PropertyDate, *datevaluetest.YMD(1800, 0, 0))
			resided = k.Event(vocabkey.EventType("residence"))
			j = k.Join(lee, resided, "subject")
			k.Commit(lee, born, resided, j)
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		var res engine.ListPersonEventsResponse
		mustRPC(t, ListPersonEvents, &engine.ListPersonEventsRequest{ProjectDir: dir, EntityId: database.UUIDString(lee)}, &res)
		if len(res.Events) != 2 || res.Events[0].GetSpan() == nil || res.Events[1].GetSpan() != nil {
			t.Fatalf("events %+v", res.Events)
		}
		if a := res.Events[0].GetAge(); a == nil || a.GetLo() != 0 || a.GetHi() != 0 || res.Events[1].GetAge() != nil {
			t.Errorf("ages %+v, %+v", res.Events[0].GetAge(), res.Events[1].GetAge())
		}
		if res.GetBirth().GetEventId() != database.UUIDString(born) || res.GetBirth().GetBesideEventId() != "" {
			t.Errorf("birth %+v", res.GetBirth())
		}
		if res.GetDeath() == nil || res.GetDeath().GetEventId() != "" || res.GetDeath().GetBesideEventId() != database.UUIDString(born) {
			t.Errorf("death %+v", res.GetDeath())
		}
	})

	for _, id := range []string{"not-an-id", "00000000-0000-7000-8000-000000000000", database.UUIDString(u)} {
		err := rpc(t, ListPersonEvents, &engine.ListPersonEventsRequest{ProjectDir: dir, EntityId: id}, &engine.ListPersonEventsResponse{})
		var ae *apperr.Error
		if !errors.As(err, &ae) || ae.Code() != apperr.CodeConclusionDetailsNotFound {
			t.Fatalf("id %q: err = %v, want %s", id, err, apperr.CodeConclusionDetailsNotFound)
		}
	}
}

// TestRecordLinkStrengthProto: every strength the engine can give a line has a
// wire value, so a new one cannot reach the app as unspecified.
func TestRecordLinkStrengthProto(t *testing.T) {
	for _, s := range []recordlinks.Strength{recordlinks.StrengthWeak, recordlinks.StrengthStrong} {
		if recordLinkStrengthProto(s) == engine.RecordLinkStrength_RECORD_LINK_STRENGTH_UNSPECIFIED {
			t.Errorf("strength %q maps to unspecified", s)
		}
	}
}

// TestGetPersonRecordLinks covers AC-1.1 and AC-1.6 through the wire: four
// registers of one son, the father named in the first two and the mother in
// the first and third, read as four graphs and two pairs.
func TestGetPersonRecordLinks(t *testing.T) {
	t.Cleanup(func() { _ = catalogsession.CloseAll() })
	created, err := onboarding.Complete(t.TempDir(), t.TempDir(), "Jake", "Robins")
	if err != nil {
		t.Fatal(err)
	}
	dir := created.ProjectDir
	type record struct {
		title string
		src   sources.Source
		graph evidencegraphs.Graph
	}
	var records []record
	var son, father, mother, birth []byte
	err = catalogsession.Do(dir, func(c *database.Catalog) error {
		r := evidencelayertest.Open(t, c)
		save := func(s subjects.Subject, onto []byte) []byte {
			return promotetest.Save(t, c, evidencelayertest.UserID, promote.Input{SubjectID: s.ID, EntityID: onto}).Entity.ID
		}
		file := func(title string, fatherName, motherName string, withBirth bool) (sonS, fatherS, motherS, birthS subjects.Subject) {
			src, art := r.NewSource(title)
			sonS = r.Person(src, art, "Thomas Smith")
			if fatherName != "" {
				fatherS = r.Person(src, art, fatherName)
			}
			if withBirth {
				birthS = r.Birth(src, art)
				r.Participation(src, art, fatherS, birthS, "subject")
			}
			if motherName != "" {
				motherS = r.Person(src, art, motherName)
			}
			records = append(records, record{title, src, evidencegraphstest.SoleGraph(t, c, src.ID)})
			return
		}
		aSon, aFather, aMother, aBirth := file("A register", "William Smith", "Mary Ann Jones", true)
		bSon, bFather, _, bBirth := file("B register", "William Smith", "", true)
		cSon, _, cMother, _ := file("C register", "", "Mary Jones", false)
		dSon, _, _, _ := file("D register", "", "", false)
		son = save(aSon, nil)
		for _, s := range []subjects.Subject{bSon, cSon, dSon} {
			save(s, son)
		}
		father = save(aFather, nil)
		save(bFather, father)
		birth = save(aBirth, nil)
		save(bBirth, birth)
		mother = save(aMother, nil)
		save(cMother, mother)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	var got engine.GetPersonRecordLinksResponse
	mustRPC(t, GetPersonRecordLinks, &engine.GetPersonRecordLinksRequest{ProjectDir: dir, EntityId: database.UUIDString(son)}, &got)
	if len(got.Graphs) != 4 {
		t.Fatalf("%d graphs, want 4: %+v", len(got.Graphs), got.Graphs)
	}
	for i, g := range got.Graphs {
		r := records[i]
		if g.GetGraphId() != database.UUIDString(r.graph.ID) || g.GetGraphRef() != r.graph.Ref || g.GetGraphName() != r.graph.Name ||
			g.GetSourceId() != database.UUIDString(r.src.ID) || g.GetSourceTitle() != r.title ||
			g.GetSourceTypeId() != database.UUIDString(r.src.SourceTypeID) {
			t.Errorf("graph %d = %+v, want the %s", i, g, r.title)
		}
	}
	a, b, c := got.Graphs[0].GetGraphId(), got.Graphs[1].GetGraphId(), got.Graphs[2].GetGraphId()
	type line struct {
		entity   string
		strength engine.RecordLinkStrength
	}
	gotPairs := map[[2]string][]line{}
	for _, p := range got.Pairs {
		for _, l := range p.Lines {
			k := [2]string{p.GetGraphAId(), p.GetGraphBId()}
			gotPairs[k] = append(gotPairs[k], line{l.GetEntityId(), l.GetStrength()})
		}
	}
	want := map[[2]string][]line{
		{a, b}: {{database.UUIDString(father), engine.RecordLinkStrength_RECORD_LINK_STRENGTH_STRONG}},
		{a, c}: {{database.UUIDString(mother), engine.RecordLinkStrength_RECORD_LINK_STRENGTH_WEAK}},
	}
	if len(got.Pairs) != 2 || !reflect.DeepEqual(gotPairs, want) {
		t.Fatalf("pairs %+v, want %+v", got.Pairs, want)
	}

	err = rpc(t, GetPersonRecordLinks, &engine.GetPersonRecordLinksRequest{ProjectDir: dir, EntityId: database.UUIDString(birth)}, &engine.GetPersonRecordLinksResponse{})
	if !errors.Is(err, conclusiondetails.ErrNotFound) {
		t.Fatalf("an Event id: error %v, want conclusiondetails.not_found", err)
	}
}
