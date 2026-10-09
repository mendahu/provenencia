package promote_test

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database/rowchange"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/audit"
	"github.com/mendahu/provenencia/core/database/autoreconciler"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/connect"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/evrun"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/project"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/writes"
)

const bridgeLocator = `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`

type bridgeWorld struct {
	t                        *testing.T
	c                        *database.Catalog
	source                   sources.Source
	artifact                 artifacts.Artifact
	personProp, eventProp    properties.Property
	placeProp, relatedProp   properties.Property
	roleProp, relTypeProp    properties.Property
	subjectRole, witnessRole propertyterms.Term
	spouse, parent           propertyterms.Term
}

func newBridgeWorld(t *testing.T) *bridgeWorld {
	t.Helper()
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	r, err := ref.Mint(ref.PrefixUser)
	if err != nil {
		t.Fatal(err)
	}
	if err := users.Upsert(c, userID, "Tester", r); err != nil {
		t.Fatal(err)
	}
	if err := subjectvocab.Install(c); err != nil {
		t.Fatal(err)
	}
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{Key: "book", Origin: sourcetypes.OriginProvenencia, Label: "Book"})
	if err != nil {
		t.Fatal(err)
	}
	src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Register"})
	})
	if err != nil {
		t.Fatal(err)
	}
	art, err := runArtifactCreate(c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
	if err != nil {
		t.Fatal(err)
	}
	w := &bridgeWorld{t: t, c: c, source: src, artifact: art}
	w.personProp = w.prop("person")
	w.eventProp = w.prop("event")
	w.placeProp = w.prop("place")
	w.relatedProp = w.prop("related_to")
	w.roleProp = w.prop("role")
	w.relTypeProp = w.prop("relationship_type")
	w.subjectRole = w.term(w.roleProp, "subject")
	w.witnessRole = w.term(w.roleProp, "witness")
	w.spouse = w.term(w.relTypeProp, "spouse")
	w.parent = w.term(w.relTypeProp, "parent")
	if w.spouse.Directed || !w.parent.Directed {
		t.Fatalf("spouse directed=%v parent directed=%v", w.spouse.Directed, w.parent.Directed)
	}
	return w
}

func (w *bridgeWorld) prop(key string) properties.Property {
	w.t.Helper()
	p, err := properties.Lookup(w.c, key, properties.OriginProvenencia)
	if err != nil {
		w.t.Fatal(err)
	}
	return p
}

func (w *bridgeWorld) term(p properties.Property, key string) propertyterms.Term {
	w.t.Helper()
	term, err := propertyterms.Lookup(w.c, p.ID, key, propertyterms.OriginProvenencia)
	if err != nil {
		w.t.Fatal(err)
	}
	return term
}

func (w *bridgeWorld) node(kind, label string, x, y int64) subjects.Subject {
	w.t.Helper()
	st, err := subjecttypes.Lookup(w.c, kind, subjecttypes.OriginProvenencia)
	if err != nil {
		w.t.Fatal(err)
	}
	s, err := evrun.CreateSubject(w.c, userID, subjects.CreateInput{
		SourceID: w.source.ID, SubjectTypeID: st.ID, Label: label,
	}, &subjects.Placement{GridX: x, GridY: y})
	if err != nil {
		w.t.Fatal(err)
	}
	return s
}

func (w *bridgeWorld) participation(person, event subjects.Subject, role propertyterms.Term) connect.Result {
	w.t.Helper()
	return w.bridge("participation", person, event, []observations.Input{
		{PropertyID: w.personProp.ID, ValueSubjectID: person.ID},
		{PropertyID: w.eventProp.ID, ValueSubjectID: event.ID},
		{PropertyID: w.roleProp.ID, ValueTermID: role.ID},
	})
}

func (w *bridgeWorld) relationship(from, to subjects.Subject, kind propertyterms.Term) connect.Result {
	w.t.Helper()
	return w.bridge("relationship", from, to, []observations.Input{
		{PropertyID: w.personProp.ID, ValueSubjectID: from.ID},
		{PropertyID: w.relatedProp.ID, ValueSubjectID: to.ID},
		{PropertyID: w.relTypeProp.ID, ValueTermID: kind.ID},
	})
}

func (w *bridgeWorld) location(event, place subjects.Subject) connect.Result {
	w.t.Helper()
	return w.bridge("location", event, place, []observations.Input{
		{PropertyID: w.eventProp.ID, ValueSubjectID: event.ID},
		{PropertyID: w.placeProp.ID, ValueSubjectID: place.ID},
	})
}

func (w *bridgeWorld) placeRelationship(from, to subjects.Subject, kind propertyterms.Term) connect.Result {
	w.t.Helper()
	fromProp := w.prop("from")
	toProp := w.prop("to")
	typeProp := w.prop("place_relationship_type")
	return w.bridge("place_relationship", from, to, []observations.Input{
		{PropertyID: fromProp.ID, ValueSubjectID: from.ID},
		{PropertyID: toProp.ID, ValueSubjectID: to.ID},
		{PropertyID: typeProp.ID, ValueTermID: kind.ID},
	})
}

func (w *bridgeWorld) bridge(kind string, from, to subjects.Subject, obs []observations.Input) connect.Result {
	w.t.Helper()
	res, err := evrun.CreateBridge(w.c, userID, connect.CreateInput{
		SourceID: w.source.ID, FromSubjectID: from.ID, ToSubjectID: to.ID, BridgeTypeKey: kind,
		Citation:     citations.CreateInput{ArtifactID: w.artifact.ID, LocatorJSON: bridgeLocator},
		Observations: obs,
	})
	if err != nil {
		w.t.Fatal(err)
	}
	return res
}

func (w *bridgeWorld) promote(s subjects.Subject) promote.Result {
	w.t.Helper()
	res, err := promote.Save(w.c, userID, promote.Input{SubjectID: s.ID})
	if err != nil {
		w.t.Fatal(err)
	}
	return res
}

func (w *bridgeWorld) join(s subjects.Subject, entityID []byte) {
	w.t.Helper()
	if _, err := promote.Save(w.c, userID, promote.Input{SubjectID: s.ID, EntityID: entityID}); err != nil {
		w.t.Fatal(err)
	}
}

func (w *bridgeWorld) handles(typeKey string) [][]byte {
	w.t.Helper()
	db, err := w.c.DB()
	if err != nil {
		w.t.Fatal(err)
	}
	rows, err := db.Query(`SELECT e.id FROM canonical_entities e
		JOIN subject_types st ON st.id = e.subject_type_id
		WHERE st.key = ? AND st.origin = 'provenencia'
		ORDER BY e.ref`, typeKey)
	if err != nil {
		w.t.Fatal(err)
	}
	defer rows.Close()
	var out [][]byte
	for rows.Next() {
		var id []byte
		if err := rows.Scan(&id); err != nil {
			w.t.Fatal(err)
		}
		out = append(out, append([]byte(nil), id...))
	}
	if err := rows.Err(); err != nil {
		w.t.Fatal(err)
	}
	return out
}

func (w *bridgeWorld) members(entityID []byte) int {
	w.t.Helper()
	got, err := identityclaims.AcceptedMembers(w.c, entityID)
	if err != nil {
		w.t.Fatal(err)
	}
	return len(got)
}

func (w *bridgeWorld) cachedEntity(assocID []byte, propKey string) []byte {
	w.t.Helper()
	db, err := w.c.DB()
	if err != nil {
		w.t.Fatal(err)
	}
	var id []byte
	err = db.QueryRow(`SELECT r.value_entity_id FROM auto_reconciler_values r
		JOIN properties p ON p.id = r.property_id
		WHERE r.entity_id = ? AND p.key = ? AND r.rank = 1 AND r.reason = 'kept'`, assocID, propKey).Scan(&id)
	if err != nil {
		w.t.Fatal(err)
	}
	return id
}

func (w *bridgeWorld) roleRows(assocID []byte) int {
	w.t.Helper()
	db, err := w.c.DB()
	if err != nil {
		w.t.Fatal(err)
	}
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM auto_reconciler_values r
		JOIN properties p ON p.id = r.property_id
		WHERE r.entity_id = ? AND p.key = 'role' AND r.reason = 'kept'`, assocID).Scan(&n); err != nil {
		w.t.Fatal(err)
	}
	return n
}

func (w *bridgeWorld) relationships(term propertyterms.Term) int {
	w.t.Helper()
	db, err := w.c.DB()
	if err != nil {
		w.t.Fatal(err)
	}
	var n int
	err = db.QueryRow(`SELECT COUNT(DISTINCT ic.entity_id)
		FROM identity_claims ic
		JOIN subjects b ON b.id = ic.subject_id
		JOIN subject_types st ON st.id = b.subject_type_id AND st.key = 'relationship'
		JOIN observations o ON o.subject_id = b.id AND o.value_term_id = ?
		WHERE ic.status = 'accepted'`, term.ID).Scan(&n)
	if err != nil {
		w.t.Fatal(err)
	}
	return n
}

func (w *bridgeWorld) cacheMatchesRebuild() {
	w.t.Helper()
	db, err := w.c.DB()
	if err != nil {
		w.t.Fatal(err)
	}
	before := cacheDump(w.t, db)
	tx, err := db.Begin()
	if err != nil {
		w.t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := autoreconciler.Rebuild(tx); err != nil {
		w.t.Fatal(err)
	}
	if after := cacheDump(w.t, tx); after != before {
		w.t.Fatalf("upkeep differs from a rebuild\nupkeep:\n%s\nrebuild:\n%s", before, after)
	}
}

func cacheDump(t *testing.T, q interface {
	Query(string, ...any) (*sql.Rows, error)
}) string {
	t.Helper()
	var b strings.Builder
	for _, table := range []string{"auto_reconciler_values", "auto_reconciler_outcomes"} {
		rows, err := q.Query(`SELECT * FROM ` + table + ` ORDER BY 1, 2, 3`)
		if err != nil {
			t.Fatal(err)
		}
		cols, _ := rows.Columns()
		for rows.Next() {
			vals := make([]any, len(cols))
			ptrs := make([]any, len(cols))
			for i := range vals {
				ptrs[i] = &vals[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				t.Fatal(err)
			}
			fmt.Fprintf(&b, "%s %v\n", table, vals)
		}
		_ = rows.Close()
	}
	return b.String()
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
		claim, err := identityclaims.AcceptedEntityForSubject(w.c, first.Subject.ID)
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
		w.cacheMatchesRebuild()
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
		w.cacheMatchesRebuild()
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
		if _, err := identityclaims.AcceptedEntityForSubject(w.c, alice.ID); err != nil {
			t.Fatal(err)
		}
		if _, err := identityclaims.AcceptedEntityForSubject(w.c, dup.ID); err != nil {
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
		eventType, err := subjecttypes.Lookup(w.c, "event", subjecttypes.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		entity, err := canonicalentities.Create(w.c, userID, canonicalentities.CreateInput{SubjectTypeID: eventType.ID})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := identityclaims.Create(w.c, userID, identityclaims.CreateInput{
			SubjectID: birth.ID, EntityID: entity.ID, Status: identityclaims.StatusAccepted,
		}); err != nil {
			t.Fatal(err)
		}
		if got := w.handles("participation"); len(got) != 1 || w.members(got[0]) != 1 {
			t.Fatalf("associations %d", len(got))
		}
	})

	t.Run("place relationship files; cycles refuse; split and amalgamation ok", func(t *testing.T) {
		w := newBridgeWorld(t)
		typeProp := w.prop("place_relationship_type")
		partOf := w.term(typeProp, "part_of")
		succeededBy := w.term(typeProp, "succeeded_by")
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
		startProp := w.prop("start_date")
		endProp := w.prop("end_date")
		if _, err := evrun.AddObservations(w.c, userID, link.Citation.ID, []observations.Input{
			{SubjectID: link.Subject.ID, PropertyID: startProp.ID, Date: &datevalues.Value{
				Kind: datevalues.KindPoint, StartYear: &year,
			}},
			{SubjectID: link.Subject.ID, PropertyID: endProp.ID, Date: &datevalues.Value{
				Kind: datevalues.KindPoint, StartYear: &year,
			}},
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
		listed, err := observations.ListBySubject(w.c, link.Subject.ID)
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
		if _, err := identityclaims.AcceptedEntityForSubject(w.c, link.Subject.ID); err != nil {
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
		w.cacheMatchesRebuild()
	})
}

func TestPlaceRelationshipTypeCreateRefused(t *testing.T) {
	w := newBridgeWorld(t)
	prop := w.prop("place_relationship_type")
	_, _, err := writes.Run(w.c, writes.Op{Action: "create_property_term", UserID: userID},
		func(tx *database.Tx) (propertyterms.Term, []rowchange.Change, error) {
			return propertyterms.Create(tx, userID, prop.ID, "Adjacent", "")
		})
	if !errors.Is(err, propertyterms.ErrLocked) {
		t.Fatalf("Create under place_relationship_type: %v", err)
	}
}

func TestBridgeSubjectDeleteReleasesPins(t *testing.T) {
	w := newBridgeWorld(t)
	james := w.node("person", "James", 0, 0)
	birth := w.node("event", "Birth", 2, 0)
	bridge := w.participation(james, birth, w.subjectRole)
	person := w.promote(james)
	w.promote(birth)
	var edge observations.Observation
	for _, o := range bridge.Observations {
		if bytes.Equal(o.PropertyID, w.personProp.ID) {
			edge = o
		}
	}
	if len(edge.ID) != 16 {
		t.Fatal("missing person edge")
	}
	db, err := w.c.DB()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	change, ok, err := identityclaims.PinTx(tx, person.Claim.ID, edge.ID)
	if err != nil || !ok {
		t.Fatalf("pin %v %v", err, ok)
	}
	if _, err := audit.Record(tx, audit.Revision{
		UserID: userID, ActionType: "pin_edge", CreatedAt: project.NowUTC(),
		Changes: []rowchange.Change{change},
	}); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	if err := evrun.DeleteSubject(w.c, userID, bridge.Subject.ID); err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM identity_claim_evidence
		WHERE identity_claim_id = ? AND observation_id = ?`, person.Claim.ID, edge.ID).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("pin survived")
	}
	if _, err := identityclaims.AcceptedEntityForSubject(w.c, james.ID); err != nil {
		t.Fatalf("person claim: %v", err)
	}
	rows, err := db.Query(`SELECT ch.entity_type FROM audit_changes ch
		JOIN audit_transactions t ON t.id = ch.audit_transaction_id
		WHERE t.action_type = 'delete_subject'
		ORDER BY ch.entity_type`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var types []string
	for rows.Next() {
		var entityType string
		if err := rows.Scan(&entityType); err != nil {
			t.Fatal(err)
		}
		types = append(types, entityType)
	}
	if !strings.Contains(strings.Join(types, ","), "identity_claim_evidence") {
		t.Fatalf("delete audited %v, want the pin released with the observation", types)
	}
}
