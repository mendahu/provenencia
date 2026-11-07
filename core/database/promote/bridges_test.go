package promote_test

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/autoreconciler/autoreconcilertest"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/connect"
	"github.com/mendahu/provenencia/core/database/evidencelayer/evidencelayertest"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/promote/promotetest"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjects/subjectstest"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
)

type bridgeWorld struct {
	*evidencelayertest.Records
	graph                    []byte // the Source's one Evidence graph
	personProp, eventProp    properties.Property
	placeProp, relatedProp   properties.Property
	roleProp, relTypeProp    properties.Property
	subjectRole, witnessRole propertyterms.Term
	spouse, parent           propertyterms.Term
}

func newBridgeWorld(t *testing.T) *bridgeWorld {
	t.Helper()
	r := evidencelayertest.New(t)
	w := &bridgeWorld{Records: r, graph: r.Graph(r.Source)}
	w.personProp = w.Prop("person")
	w.eventProp = w.Prop("event")
	w.placeProp = w.Prop("place")
	w.relatedProp = w.Prop("related_to")
	w.roleProp = w.Prop("role")
	w.relTypeProp = w.Prop("relationship_type")
	w.subjectRole = w.Term("role", "subject")
	w.witnessRole = w.Term("role", "witness")
	w.spouse = w.Term("relationship_type", "spouse")
	w.parent = w.Term("relationship_type", "parent")
	if w.spouse.Directed || !w.parent.Directed {
		t.Fatalf("spouse directed=%v parent directed=%v", w.spouse.Directed, w.parent.Directed)
	}
	return w
}

func (w *bridgeWorld) node(kind, label string, x, y int64) subjects.Subject {
	w.T.Helper()
	st, err := subjecttypes.Lookup(w.C, kind, vocabkey.OriginProvenencia)
	if err != nil {
		w.T.Fatal(err)
	}
	return subjectstest.At(w.T, w.C, userID, w.graph, st.ID, label, subjects.Placement{GridX: x, GridY: y})
}

func (w *bridgeWorld) participation(person, event subjects.Subject, role propertyterms.Term) connect.Result {
	w.T.Helper()
	return w.bridge("participation", person, event, []observations.Input{
		{PropertyID: w.personProp.ID, ValueSubjectID: person.ID},
		{PropertyID: w.eventProp.ID, ValueSubjectID: event.ID},
		{PropertyID: w.roleProp.ID, ValueTermID: role.ID},
	})
}

func (w *bridgeWorld) relationship(from, to subjects.Subject, kind propertyterms.Term) connect.Result {
	w.T.Helper()
	return w.bridge("relationship", from, to, []observations.Input{
		{PropertyID: w.personProp.ID, ValueSubjectID: from.ID},
		{PropertyID: w.relatedProp.ID, ValueSubjectID: to.ID},
		{PropertyID: w.relTypeProp.ID, ValueTermID: kind.ID},
	})
}

func (w *bridgeWorld) location(event, place subjects.Subject) connect.Result {
	w.T.Helper()
	return w.bridge("location", event, place, []observations.Input{
		{PropertyID: w.eventProp.ID, ValueSubjectID: event.ID},
		{PropertyID: w.placeProp.ID, ValueSubjectID: place.ID},
	})
}

func (w *bridgeWorld) placeRelationship(from, to subjects.Subject, kind propertyterms.Term) connect.Result {
	w.T.Helper()
	fromProp := w.Prop("from")
	toProp := w.Prop("to")
	typeProp := w.Prop("place_relationship_type")
	return w.bridge("place_relationship", from, to, []observations.Input{
		{PropertyID: fromProp.ID, ValueSubjectID: from.ID},
		{PropertyID: toProp.ID, ValueSubjectID: to.ID},
		{PropertyID: typeProp.ID, ValueTermID: kind.ID},
	})
}

func (w *bridgeWorld) bridge(kind string, from, to subjects.Subject, obs []observations.Input) connect.Result {
	w.T.Helper()
	res, err := writes.Call(w.C, writes.Op{Action: "create_cited_bridge", UserID: userID}, func(tx *database.Tx) (connect.Result, []rowchange.Change, error) {
		return connect.CreateCitedBridge(tx, userID, connect.CreateInput{
			FromSubjectID: from.ID, ToSubjectID: to.ID, BridgeTypeKey: kind,
			Citation:     citations.CreateInput{ArtifactID: w.Artifact.ID, LocatorJSON: evidencelayertest.Locator},
			Observations: obs,
		})
	})
	if err != nil {
		w.T.Fatal(err)
	}
	return res
}

func (w *bridgeWorld) promote(s subjects.Subject) promote.Result {
	w.T.Helper()
	return promotetest.Save(w.T, w.C, userID, promote.Input{SubjectID: s.ID})
}

func (w *bridgeWorld) join(s subjects.Subject, entityID []byte) {
	w.T.Helper()
	promotetest.Save(w.T, w.C, userID, promote.Input{SubjectID: s.ID, EntityID: entityID})
}

func (w *bridgeWorld) handles(typeKey string) [][]byte {
	w.T.Helper()
	db, err := w.C.DB()
	if err != nil {
		w.T.Fatal(err)
	}
	rows, err := db.Query(`SELECT e.id FROM canonical_entities e
		JOIN subject_types st ON st.id = e.subject_type_id
		WHERE st.key = ? AND st.origin = 'provenencia'
		ORDER BY e.ref`, typeKey)
	if err != nil {
		w.T.Fatal(err)
	}
	defer rows.Close()
	var out [][]byte
	for rows.Next() {
		var id []byte
		if err := rows.Scan(&id); err != nil {
			w.T.Fatal(err)
		}
		out = append(out, append([]byte(nil), id...))
	}
	if err := rows.Err(); err != nil {
		w.T.Fatal(err)
	}
	return out
}

func (w *bridgeWorld) members(entityID []byte) int {
	w.T.Helper()
	got, err := identityclaims.AcceptedMembers(w.C, entityID)
	if err != nil {
		w.T.Fatal(err)
	}
	return len(got)
}

func (w *bridgeWorld) cachedEntity(assocID []byte, propKey string) []byte {
	w.T.Helper()
	db, err := w.C.DB()
	if err != nil {
		w.T.Fatal(err)
	}
	var id []byte
	err = db.QueryRow(`SELECT r.value_entity_id FROM auto_reconciler_values r
		JOIN properties p ON p.id = r.property_id
		WHERE r.entity_id = ? AND p.key = ? AND r.rank = 1 AND r.reason = 'kept'`, assocID, propKey).Scan(&id)
	if err != nil {
		w.T.Fatal(err)
	}
	return id
}

func (w *bridgeWorld) roleRows(assocID []byte) int {
	w.T.Helper()
	db, err := w.C.DB()
	if err != nil {
		w.T.Fatal(err)
	}
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM auto_reconciler_values r
		JOIN properties p ON p.id = r.property_id
		WHERE r.entity_id = ? AND p.key = 'role' AND r.reason = 'kept'`, assocID).Scan(&n); err != nil {
		w.T.Fatal(err)
	}
	return n
}

func (w *bridgeWorld) relationships(term propertyterms.Term) int {
	w.T.Helper()
	db, err := w.C.DB()
	if err != nil {
		w.T.Fatal(err)
	}
	var n int
	err = db.QueryRow(`SELECT COUNT(DISTINCT ic.entity_id)
		FROM identity_claims ic
		JOIN subjects b ON b.id = ic.subject_id
		JOIN subject_types st ON st.id = b.subject_type_id AND st.key = 'relationship'
		JOIN observations o ON o.subject_id = b.id AND o.value_term_id = ?
		WHERE ic.status = 'accepted'`, term.ID).Scan(&n)
	if err != nil {
		w.T.Fatal(err)
	}
	return n
}

func TestBridgeFiling(t *testing.T) {
	t.Run("participation files when both ends are promoted and a second record joins", func(t *testing.T) {
		w := newBridgeWorld(t)
		james := w.node("person", "James", 0, 0)
		birth := w.node("event", "Birth", 2, 0)
		first := w.participation(james, birth, w.subjectRole)
		person := w.promote(james)
		if got := w.handles("participation"); len(got) != 0 {
			t.Fatalf("filed before the event was a handle: %d", len(got))
		}
		event := w.promote(birth)
		got := w.handles("participation")
		if len(got) != 1 {
			t.Fatalf("associations %d", len(got))
		}
		if w.members(got[0]) != 1 {
			t.Fatalf("members %d", w.members(got[0]))
		}
		if !bytes.Equal(w.cachedEntity(got[0], "person"), person.Entity.ID) {
			t.Fatal("person end is not the person handle")
		}
		if !bytes.Equal(w.cachedEntity(got[0], "event"), event.Entity.ID) {
			t.Fatal("event end is not the event handle")
		}
		claim, err := identityclaims.AcceptedEntityForSubject(w.C, first.Subject.ID)
		if err != nil || !bytes.Equal(claim.EntityID, got[0]) {
			t.Fatalf("bridge claim %v %+v", err, claim.EntityID)
		}

		james2 := w.node("person", "James (baptism)", 0, 2)
		birth2 := w.node("event", "Birth (baptism)", 2, 2)
		w.participation(james2, birth2, w.witnessRole)
		w.join(james2, person.Entity.ID)
		w.join(birth2, event.Entity.ID)
		if again := w.handles("participation"); len(again) != 1 || !bytes.Equal(again[0], got[0]) {
			t.Fatalf("second record minted %d associations", len(again))
		}
		if w.members(got[0]) != 2 {
			t.Fatalf("members %d, want 2", w.members(got[0]))
		}
		if w.roleRows(got[0]) != 2 {
			t.Fatalf("role rows %d, want the two roles kept apart from the key", w.roleRows(got[0]))
		}
		autoreconcilertest.AssertCacheMatchesRebuild(t, w.C)
	})

	t.Run("spouse is one relationship and parent keeps direction", func(t *testing.T) {
		w := newBridgeWorld(t)
		alice := w.node("person", "Alice", 0, 0)
		bob := w.node("person", "Bob", 4, 0)
		w.relationship(alice, bob, w.spouse)
		w.relationship(bob, alice, w.spouse)
		w.relationship(alice, bob, w.parent)
		w.relationship(bob, alice, w.parent)
		w.promote(alice)
		w.promote(bob)
		if w.relationships(w.spouse) != 1 {
			t.Fatalf("spouse associations %d, want 1", w.relationships(w.spouse))
		}
		if w.relationships(w.parent) != 2 {
			t.Fatalf("parent associations %d, want 2", w.relationships(w.parent))
		}
		autoreconcilertest.AssertCacheMatchesRebuild(t, w.C)
	})

	t.Run("a self-link is not filed and the claim stands", func(t *testing.T) {
		w := newBridgeWorld(t)
		alice := w.node("person", "Alice", 0, 0)
		dup := w.node("person", "Alice again", 4, 0)
		w.relationship(alice, dup, w.spouse)
		first := w.promote(alice)
		w.join(dup, first.Entity.ID)
		if got := w.handles("relationship"); len(got) != 0 {
			t.Fatalf("self-link filed %d associations", len(got))
		}
		if _, err := identityclaims.AcceptedEntityForSubject(w.C, alice.ID); err != nil {
			t.Fatal(err)
		}
		if _, err := identityclaims.AcceptedEntityForSubject(w.C, dup.ID); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("location files by its ends", func(t *testing.T) {
		w := newBridgeWorld(t)
		birth := w.node("event", "Birth", 0, 0)
		york := w.node("place", "York", 2, 0)
		w.location(birth, york)
		w.promote(birth)
		place := w.promote(york)
		got := w.handles("location")
		if len(got) != 1 {
			t.Fatalf("associations %d", len(got))
		}
		if !bytes.Equal(w.cachedEntity(got[0], "place"), place.Entity.ID) {
			t.Fatal("place end is not the place handle")
		}
	})

	t.Run("an accepted claim create files the bridge", func(t *testing.T) {
		w := newBridgeWorld(t)
		james := w.node("person", "James", 0, 0)
		birth := w.node("event", "Birth", 2, 0)
		w.participation(james, birth, w.subjectRole)
		w.promote(james)
		eventType, err := subjecttypes.Lookup(w.C, "event", vocabkey.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		entity, err := writes.Call(w.C, writes.Op{Action: "create_canonical_entity", UserID: userID}, func(tx *database.Tx) (canonicalentities.Entity, []rowchange.Change, error) {
			return canonicalentities.Create(tx, userID, canonicalentities.CreateInput{SubjectTypeID: eventType.ID})
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := writes.Call(w.C, writes.Op{Action: "create_identity_claim", UserID: userID}, func(tx *database.Tx) (identityclaims.Claim, []rowchange.Change, error) {
			return identityclaims.Create(tx, userID, identityclaims.CreateInput{
				SubjectID: birth.ID, EntityID: entity.ID, Status: identityclaims.StatusAccepted,
			})
		}); err != nil {
			t.Fatal(err)
		}
		if got := w.handles("participation"); len(got) != 1 || w.members(got[0]) != 1 {
			t.Fatalf("associations %d", len(got))
		}
	})

	t.Run("place relationship files; cycles refuse; split and amalgamation ok", func(t *testing.T) {
		w := newBridgeWorld(t)
		partOf := w.Term("place_relationship_type", "part_of")
		succeededBy := w.Term("place_relationship_type", "succeeded_by")
		if !partOf.Directed || !succeededBy.Directed {
			t.Fatalf("directed part_of=%v succeeded_by=%v", partOf.Directed, succeededBy.Directed)
		}

		york := w.node("place", "York", 0, 0)
		uc := w.node("place", "Upper Canada", 2, 0)
		canada := w.node("place", "Canada", 4, 0)
		link := w.placeRelationship(york, uc, partOf)
		w.placeRelationship(uc, canada, partOf)
		w.placeRelationship(canada, york, partOf) // would cycle after all three file

		year := 1791
		startProp := w.Prop("start_date")
		endProp := w.Prop("end_date")
		if _, err := writes.Call(w.C, writes.Op{Action: "add_observations", UserID: userID}, func(tx *database.Tx) ([]observations.Observation, []rowchange.Change, error) {
			return observations.AddToCitation(tx, userID, link.Citation.ID, []observations.Input{
				{SubjectID: link.Subject.ID, PropertyID: startProp.ID, Date: &datevalue.Value{
					Kind: datevalue.KindPoint, StartYear: &year,
				}},
				{SubjectID: link.Subject.ID, PropertyID: endProp.ID, Date: &datevalue.Value{
					Kind: datevalue.KindPoint, StartYear: &year,
				}},
			})
		}); err != nil {
			t.Fatal(err)
		}

		w.promote(york)
		w.promote(uc)
		w.promote(canada)
		got := w.handles("place_relationship")
		if len(got) != 2 {
			t.Fatalf("hierarchical associations %d, want 2 (cycle refused)", len(got))
		}
		listed, err := observations.ListBySubject(w.C, link.Subject.ID)
		if err != nil {
			t.Fatal(err)
		}
		var sawStart, sawEnd bool
		for _, o := range listed {
			switch {
			case bytes.Equal(o.PropertyID, startProp.ID) && len(o.ValueDateID) == 16:
				sawStart = true
			case bytes.Equal(o.PropertyID, endProp.ID) && len(o.ValueDateID) == 16:
				sawEnd = true
			}
		}
		if !sawStart || !sawEnd {
			t.Fatalf("hierarchical link lost start/end (start=%v end=%v)", sawStart, sawEnd)
		}
		if _, err := identityclaims.AcceptedEntityForSubject(w.C, link.Subject.ID); err != nil {
			t.Fatalf("filed link claim: %v", err)
		}

		// Split: one → two successors. Amalgamation: two → one.
		oldA := w.node("place", "Old A", 0, 2)
		oldB := w.node("place", "Old B", 2, 2)
		newOne := w.node("place", "New", 4, 2)
		left := w.node("place", "Left", 0, 4)
		right := w.node("place", "Right", 2, 4)
		w.placeRelationship(oldA, newOne, succeededBy)
		w.placeRelationship(oldB, newOne, succeededBy)
		w.placeRelationship(newOne, left, succeededBy)
		w.placeRelationship(newOne, right, succeededBy)
		// Temporal loop: left → oldA would close after the chain files.
		w.placeRelationship(left, oldA, succeededBy)
		w.promote(oldA)
		w.promote(oldB)
		w.promote(newOne)
		w.promote(left)
		w.promote(right)
		// Hierarchical 2 + amalgamation 2 + split 2 = 6; temporal loop soft-refused.
		if got := w.handles("place_relationship"); len(got) != 6 {
			t.Fatalf("place_relationship associations %d, want 6", len(got))
		}
		autoreconcilertest.AssertCacheMatchesRebuild(t, w.C)
	})
}

func TestPlaceRelationshipTypeCreateRefused(t *testing.T) {
	w := newBridgeWorld(t)
	prop := w.Prop("place_relationship_type")
	_, _, err := writes.Run(w.C, writes.Op{Action: "create_property_term", UserID: userID},
		func(tx *database.Tx) (propertyterms.Term, []rowchange.Change, error) {
			return propertyterms.Create(tx, userID, prop.ID, "Adjacent", "")
		})
	if !errors.Is(err, propertyterms.ErrLocked) {
		t.Fatalf("Create under place_relationship_type: %v", err)
	}
}

func TestBridgeSubjectDeleteKeepsTheEndClaims(t *testing.T) {
	w := newBridgeWorld(t)
	james := w.node("person", "James", 0, 0)
	birth := w.node("event", "Birth", 2, 0)
	bridge := w.participation(james, birth, w.subjectRole)
	w.promote(james)
	w.promote(birth)

	if _, err := writes.Call(w.C, writes.Op{Action: "delete_subject", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
		changes, err := subjects.Delete(tx, userID, bridge.Subject.ID)
		return struct{}{}, changes, err
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := identityclaims.AcceptedEntityForSubject(w.C, james.ID); err != nil {
		t.Fatalf("person claim: %v", err)
	}
}

// An association Subject is not a primary type: saving one that is already
// filed by its bridge is unsupported, not "already a member".
func TestSaveRefusesAnAssociationAlreadyFiledByItsBridge(t *testing.T) {
	w := newBridgeWorld(t)
	james := w.node("person", "James", 0, 0)
	birth := w.node("event", "Birth", 2, 0)
	bridge := w.participation(james, birth, w.subjectRole)
	w.promote(james)
	w.promote(birth)
	_, err := writes.Call(w.C, writes.Op{Action: "promote_subject", UserID: userID}, func(tx *database.Tx) (promote.Result, []rowchange.Change, error) {
		return promote.Save(tx, userID, promote.Input{SubjectID: bridge.Subject.ID})
	})
	if !errors.Is(err, promote.ErrUnsupportedType) {
		t.Fatalf("err %v, want promote.ErrUnsupportedType", err)
	}
}

// TestBridgeFilingReadsDoNotGrowWithTheBridges: one Done that files N place
// relationships of one term, with one more waiting on an unpromoted place,
// reads the term's filed links and the term rows a number of times that does
// not depend on N.
func TestBridgeFilingReadsDoNotGrowWithTheBridges(t *testing.T) {
	reads := func(n int) int {
		w := newBridgeWorld(t)
		partOf := w.Term("place_relationship_type", "part_of")
		canada := w.node("place", "Canada", 0, 0)
		for i := 0; i < n; i++ {
			child := w.node("place", fmt.Sprintf("Child %d", i), int64(2+2*i), 0)
			w.placeRelationship(child, canada, partOf)
			w.promote(child)
		}
		w.placeRelationship(canada, w.node("place", "Elsewhere", 0, 2), partOf)
		var count int
		identityclaims.SetReadCounterForTest(func() { count++ })
		defer identityclaims.SetReadCounterForTest(nil)
		w.promote(canada)
		if got := w.handles("place_relationship"); len(got) != n {
			t.Fatalf("%d bridges filed %d associations", n, len(got))
		}
		return count
	}
	few, many := reads(3), reads(12)
	if few != many {
		t.Fatalf("3 bridges read %d times and 12 read %d: the reads grow with the bridges", few, many)
	}
}

// A bridge drawn after both of its ends are handles files in the same write.
func TestBridgeFilesWhenDrawnBetweenHandles(t *testing.T) {
	t.Run("relationship between two handles files", func(t *testing.T) {
		w := newBridgeWorld(t)
		alice := w.node("person", "Alice", 0, 0)
		bob := w.node("person", "Bob", 4, 0)
		a := w.promote(alice)
		b := w.promote(bob)
		res := w.relationship(alice, bob, w.spouse)
		got := w.handles("relationship")
		if len(got) != 1 {
			t.Fatalf("relationship handles %d, want 1", len(got))
		}
		if w.members(got[0]) != 1 {
			t.Fatalf("members %d, want 1", w.members(got[0]))
		}
		claim, err := identityclaims.AcceptedEntityForSubject(w.C, res.Subject.ID)
		if err != nil || !bytes.Equal(claim.EntityID, got[0]) {
			t.Fatalf("bridge claim %v %x", err, claim.EntityID)
		}
		if !bytes.Equal(w.cachedEntity(got[0], "person"), a.Entity.ID) {
			t.Fatal("person end is not Alice's handle")
		}
		if !bytes.Equal(w.cachedEntity(got[0], "related_to"), b.Entity.ID) {
			t.Fatal("related_to end is not Bob's handle")
		}
		autoreconcilertest.AssertCacheMatchesRebuild(t, w.C)
	})

	t.Run("participation, location and place relationship file", func(t *testing.T) {
		w := newBridgeWorld(t)
		james := w.node("person", "James", 0, 0)
		birth := w.node("event", "Birth", 2, 0)
		w.promote(james)
		w.promote(birth)
		w.participation(james, birth, w.subjectRole)
		if got := w.handles("participation"); len(got) != 1 {
			t.Fatalf("participation handles %d, want 1", len(got))
		}

		birth2 := w.node("event", "Birth again", 2, 2)
		york := w.node("place", "York", 4, 2)
		w.promote(birth2)
		w.promote(york)
		w.location(birth2, york)
		if got := w.handles("location"); len(got) != 1 {
			t.Fatalf("location handles %d, want 1", len(got))
		}

		partOf := w.Term("place_relationship_type", "part_of")
		uc := w.node("place", "Upper Canada", 6, 2)
		w.promote(uc)
		w.placeRelationship(york, uc, partOf)
		if got := w.handles("place_relationship"); len(got) != 1 {
			t.Fatalf("place_relationship handles %d, want 1", len(got))
		}
	})

	t.Run("a second record joins the association", func(t *testing.T) {
		w := newBridgeWorld(t)
		alice := w.node("person", "Alice", 0, 0)
		bob := w.node("person", "Bob", 4, 0)
		a := w.promote(alice)
		b := w.promote(bob)
		w.relationship(alice, bob, w.spouse)

		alice2 := w.node("person", "Alice (baptism)", 0, 2)
		bob2 := w.node("person", "Bob (baptism)", 4, 2)
		w.join(alice2, a.Entity.ID)
		w.join(bob2, b.Entity.ID)
		w.relationship(alice2, bob2, w.spouse)
		got := w.handles("relationship")
		if len(got) != 1 {
			t.Fatalf("relationship handles %d, want 1", len(got))
		}
		if w.members(got[0]) != 2 {
			t.Fatalf("members %d, want 2", w.members(got[0]))
		}
		autoreconcilertest.AssertCacheMatchesRebuild(t, w.C)
	})

	t.Run("one end unpromoted waits", func(t *testing.T) {
		w := newBridgeWorld(t)
		alice := w.node("person", "Alice", 0, 0)
		bob := w.node("person", "Bob", 4, 0)
		w.promote(alice)
		res := w.relationship(alice, bob, w.spouse)
		if w.filed(res.Subject) {
			t.Fatal("bridge filed with an unpromoted end")
		}
		if got := w.handles("relationship"); len(got) != 0 {
			t.Fatalf("relationship handles %d, want 0", len(got))
		}
		w.promote(bob)
		got := w.handles("relationship")
		if len(got) != 1 || w.members(got[0]) != 1 {
			t.Fatalf("after promoting Bob: %d handles", len(got))
		}
	})

	t.Run("a refused filing does not fail the connect", func(t *testing.T) {
		w := newBridgeWorld(t)
		alice := w.node("person", "Alice", 0, 0)
		dup := w.node("person", "Alice again", 4, 0)
		a := w.promote(alice)
		w.join(dup, a.Entity.ID)
		res := w.relationship(alice, dup, w.spouse)
		if w.filed(res.Subject) {
			t.Fatal("a self-link filed")
		}
		if got := w.handles("relationship"); len(got) != 0 {
			t.Fatalf("relationship handles %d, want 0", len(got))
		}
		if !w.filed(alice) || !w.filed(dup) {
			t.Fatal("an endpoint claim was lost")
		}
	})

	t.Run("a place loop does not fail the connect", func(t *testing.T) {
		w := newBridgeWorld(t)
		partOf := w.Term("place_relationship_type", "part_of")
		york := w.node("place", "York", 0, 0)
		uc := w.node("place", "Upper Canada", 2, 0)
		canada := w.node("place", "Canada", 4, 0)
		w.promote(york)
		w.promote(uc)
		w.promote(canada)
		w.placeRelationship(york, uc, partOf)
		w.placeRelationship(uc, canada, partOf)
		loop := w.placeRelationship(canada, york, partOf)
		if w.filed(loop.Subject) {
			t.Fatal("a loop filed")
		}
		if got := w.handles("place_relationship"); len(got) != 2 {
			t.Fatalf("place_relationship handles %d, want 2", len(got))
		}
	})
}
