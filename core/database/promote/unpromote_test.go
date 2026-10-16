package promote_test

import (
	"bytes"
	"database/sql"
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/writes"
)

func (w *bridgeWorld) unpromote(s subjects.Subject) error {
	w.t.Helper()
	_, err := writes.Call(w.c, writes.Op{Action: "unpromote_subject", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
		ch, err := promote.Unpromote(tx, userID, s.ID)
		return struct{}{}, ch, err
	})
	return err
}

func (w *bridgeWorld) filed(s subjects.Subject) bool {
	w.t.Helper()
	_, err := identityclaims.AcceptedEntityForSubject(w.c, s.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return false
	}
	if err != nil {
		w.t.Fatal(err)
	}
	return true
}

func (w *bridgeWorld) revisions() int {
	w.t.Helper()
	db, err := w.c.DB()
	if err != nil {
		w.t.Fatal(err)
	}
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM audit_transactions`).Scan(&n); err != nil {
		w.t.Fatal(err)
	}
	return n
}

type household struct {
	james, john, mary, event subjects.Subject
	sonOf, part, spouse      subjects.Subject
	person, johnH            []byte
}

func newHousehold(w *bridgeWorld) household {
	h := household{
		james: w.node("person", "James", 0, 0),
		john:  w.node("person", "John", 2, 0),
		mary:  w.node("person", "Mary", 4, 0),
		event: w.node("event", "1881 census", 2, 2),
	}
	h.sonOf = w.relationship(h.james, h.john, w.parent).Subject
	h.part = w.participation(h.james, h.event, w.subjectRole).Subject
	h.spouse = w.relationship(h.john, h.mary, w.spouse).Subject
	h.person = w.promote(h.james).Entity.ID
	h.johnH = w.promote(h.john).Entity.ID
	w.promote(h.mary)
	w.promote(h.event)
	return h
}

func TestUnpromote(t *testing.T) {
	t.Run("takes the subject's claim and the bridges it ends off, and nothing else", func(t *testing.T) {
		w := newBridgeWorld(t)
		h := newHousehold(w)
		for _, b := range []subjects.Subject{h.sonOf, h.part, h.spouse} {
			if !w.filed(b) {
				t.Fatal("household bridge not filed after promoting its ends")
			}
		}
		before := w.revisions()
		personHandles, personAssoc := len(w.handles("person")), len(w.handles("relationship"))+len(w.handles("participation"))
		if err := w.unpromote(h.james); err != nil {
			t.Fatal(err)
		}
		if w.revisions() != before+1 {
			t.Fatalf("revisions %d, want %d", w.revisions(), before+1)
		}
		if got := lastRevisionChanges(t, w.c); got[:len("unpromote_subject:")] != "unpromote_subject:" {
			t.Fatalf("revision %s", got)
		}
		for name, s := range map[string]subjects.Subject{"james": h.james, "son-of": h.sonOf, "participation": h.part} {
			if w.filed(s) {
				t.Fatalf("%s still filed", name)
			}
		}
		for name, s := range map[string]subjects.Subject{"john": h.john, "mary": h.mary, "event": h.event, "spouse": h.spouse} {
			if !w.filed(s) {
				t.Fatalf("%s lost its filing", name)
			}
		}
		if len(w.handles("person")) != personHandles || len(w.handles("relationship"))+len(w.handles("participation")) != personAssoc {
			t.Fatal("a handle was deleted")
		}
		if w.members(h.person) != 0 {
			t.Fatalf("James's handle has %d members", w.members(h.person))
		}
		w.cacheMatchesRebuild()
	})

	t.Run("a Subject with no filed bridges loses only its own claim", func(t *testing.T) {
		w := newBridgeWorld(t)
		james := w.node("person", "James", 0, 0)
		john := w.node("person", "John", 2, 0)
		w.relationship(james, john, w.parent)
		w.promote(james)
		if err := w.unpromote(james); err != nil {
			t.Fatal(err)
		}
		if w.filed(james) {
			t.Fatal("still filed")
		}
	})

	t.Run("a relationship waiting on an unpromoted end is untouched", func(t *testing.T) {
		w := newBridgeWorld(t)
		james := w.node("person", "James", 0, 0)
		john := w.node("person", "John", 2, 0)
		rel := w.relationship(james, john, w.parent).Subject
		w.promote(james)
		if w.filed(rel) {
			t.Fatal("filed with an unpromoted end")
		}
		if err := w.unpromote(james); err != nil {
			t.Fatal(err)
		}
		if w.filed(rel) {
			t.Fatal("rel changed")
		}
	})

	t.Run("a bridge on another Source stays filed", func(t *testing.T) {
		w := newBridgeWorld(t)
		h := newHousehold(w)
		other := make([]byte, 16)
		other[0] = 0x7f
		_, err := writes.Call(w.c, writes.Op{Action: "unpromote_subject", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			ch, err := identityclaims.UnfileBridgesTx(tx.Tx, other, h.james.ID)
			if len(ch) != 0 {
				t.Fatalf("changes %d", len(ch))
			}
			return struct{}{}, ch, err
		})
		if err != nil {
			t.Fatal(err)
		}
		if !w.filed(h.sonOf) {
			t.Fatal("son-of unfiled")
		}
	})

	t.Run("no accepted claim writes nothing and is not an error", func(t *testing.T) {
		w := newBridgeWorld(t)
		james := w.node("person", "James", 0, 0)
		before := w.revisions()
		if err := w.unpromote(james); err != nil {
			t.Fatal(err)
		}
		if w.revisions() != before {
			t.Fatal("a revision was written")
		}
	})

	t.Run("a relationship is unsupported", func(t *testing.T) {
		w := newBridgeWorld(t)
		h := newHousehold(w)
		before := w.revisions()
		if err := w.unpromote(h.sonOf); !errors.Is(err, promote.ErrUnsupportedType) {
			t.Fatalf("err %v", err)
		}
		if w.revisions() != before {
			t.Fatal("a revision was written")
		}
	})

	t.Run("an unknown id and no user are invalid", func(t *testing.T) {
		w := newBridgeWorld(t)
		bad := subjects.Subject{ID: bytes.Repeat([]byte{9}, 16)}
		if err := w.unpromote(bad); !errors.Is(err, promote.ErrInvalid) {
			t.Fatalf("unknown: %v", err)
		}
		_, err := writes.Call(w.c, writes.Op{Action: "unpromote_subject", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			ch, err := promote.Unpromote(tx, nil, bad.ID)
			return struct{}{}, ch, err
		})
		if !errors.Is(err, promote.ErrInvalid) {
			t.Fatalf("nil user: %v", err)
		}
	})
}

func TestUnpromoteKeepsOtherMembersPins(t *testing.T) {
	f := newPinFixture(t)
	birth, birthObs := f.person1("James")
	first, err := writes.Call(f.c, writes.Op{Action: "promote_subject", UserID: userID}, func(tx *database.Tx) (promote.Result, []rowchange.Change, error) {
		return promote.Save(tx, userID, promote.Input{SubjectID: birth.ID})
	})
	f.must(err)
	census, censusObs := f.person1("James")
	joined, err := writes.Call(f.c, writes.Op{Action: "promote_subject", UserID: userID}, func(tx *database.Tx) (promote.Result, []rowchange.Change, error) {
		return promote.Save(tx, userID, promote.Input{SubjectID: census.ID, EntityID: first.Entity.ID,
			Pairs: []promote.Pair{{IncomingObservationID: censusObs[0].ID, MemberObservationID: birthObs[0].ID}}})
	})
	f.must(err)
	_, err = writes.Call(f.c, writes.Op{Action: "unpromote_subject", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
		ch, err := promote.Unpromote(tx, userID, census.ID)
		return struct{}{}, ch, err
	})
	f.must(err)
	if _, err := identityclaims.Get(f.c, joined.Claim.ID); err == nil {
		t.Fatal("James's claim survived")
	}
	if got, want := f.pinned(joined.Claim.ID), ""; got != want {
		t.Fatalf("James's pins survived: %q", got)
	}
	if got, want := f.pinned(first.Claim.ID), idsOf(birthObs[0], censusObs[0]); got != want {
		t.Fatalf("birth claim pins = %q, want both", got)
	}
	kept, err := identityclaims.Get(f.c, first.Claim.ID)
	f.must(err)
	if kept.Status != identityclaims.StatusAccepted {
		t.Fatalf("birth claim status %s", kept.Status)
	}
}
