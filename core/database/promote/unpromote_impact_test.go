package promote_test

import (
	"bytes"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/deleteimpact"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/writes"
)

func TestUnpromoteImpactListsBridgeHandles(t *testing.T) {
	w := newBridgeWorld(t)
	h := newHousehold(w)
	report, err := promote.UnpromoteImpact(w.C, h.james.ID)
	if err != nil {
		t.Fatal(err)
	}
	if group := report.Deletes; group.Via != deleteimpact.ViaLastMember || group.Total < 2 {
		t.Fatalf("want primary and association deletes, got %+v", group)
	}
}

func TestUnpromoteImpactEmptyWhenHandleStays(t *testing.T) {
	w := newBridgeWorld(t)
	james := w.node("person", "James", 0, 0)
	john := w.node("person", "John", 2, 0)
	entity := w.promote(james).Entity.ID
	w.join(john, entity)
	report, err := promote.UnpromoteImpact(w.C, james.ID)
	if err != nil {
		t.Fatal(err)
	}
	if group := report.Deletes; group.Via != "" {
		t.Fatalf("handle stays: %+v", group)
	}
}

// bridgeFixture is a promoted person P with two filed bridges, a relationship
// R then a participation Q, and an unfiled relationship U (its other end is
// not promoted).
type bridgeFixture struct {
	p, r, q, u subjects.Subject
}

func newBridgeFixture(w *bridgeWorld) bridgeFixture {
	p := w.node("person", "James", 0, 0)
	john := w.node("person", "John", 2, 0)
	ev := w.node("event", "1881 census", 2, 2)
	waiting := w.node("person", "Mary", 4, 0)
	f := bridgeFixture{p: p}
	f.r = w.relationship(p, john, w.parent).Subject
	f.q = w.participation(p, ev, w.subjectRole).Subject
	f.u = w.relationship(p, waiting, w.parent).Subject
	w.promote(john)
	w.promote(ev)
	w.promote(p)
	return f
}

func TestFiledBridgesOf(t *testing.T) {
	tests := []struct {
		name string
		run  func(w *bridgeWorld, f bridgeFixture) (subjectID []byte, want []subjects.Subject)
	}{
		{"filed bridges in bridge Subject id order, without the unfiled one", func(w *bridgeWorld, f bridgeFixture) ([]byte, []subjects.Subject) {
			return f.p.ID, []subjects.Subject{f.r, f.q}
		}},
		{"a person with no bridges", func(w *bridgeWorld, f bridgeFixture) ([]byte, []subjects.Subject) {
			return w.node("person", "Alone", 8, 0).ID, nil
		}},
		{"a bridge Subject's own id is no end", func(w *bridgeWorld, f bridgeFixture) ([]byte, []subjects.Subject) {
			return f.r.ID, nil
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newBridgeWorld(t)
			f := newBridgeFixture(w)
			id, want := tt.run(w, f)
			db, err := w.C.DB()
			if err != nil {
				t.Fatal(err)
			}
			tx, err := db.Begin()
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback() }()
			got, err := identityclaims.FiledBridgesOf(tx, id)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != len(want) {
				t.Fatalf("got %d filed bridges, want %d", len(got), len(want))
			}
			for i, b := range want {
				claim, err := identityclaims.AcceptedEntityForSubjectTx(tx, b.ID)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(got[i].SubjectID, b.ID) || !bytes.Equal(got[i].ClaimID, claim.ID) || !bytes.Equal(got[i].EntityID, claim.EntityID) {
					t.Fatalf("bridge %d: got %+v, want subject %x claim %x entity %x", i, got[i], b.ID, claim.ID, claim.EntityID)
				}
			}
		})
	}
}

func TestUnpromoteImpactListsUnpromotedBridges(t *testing.T) {
	t.Run("the preflight lists the bridges the write takes back", func(t *testing.T) {
		w := newBridgeWorld(t)
		f := newBridgeFixture(w)
		wantClaims := map[string]bool{}
		for _, b := range []subjects.Subject{f.r, f.q, f.p} {
			claim, err := identityclaims.AcceptedEntityForSubject(w.C, b.ID)
			if err != nil {
				t.Fatal(err)
			}
			wantClaims[string(claim.ID)] = true
		}
		report, err := promote.UnpromoteImpact(w.C, f.p.ID)
		if err != nil {
			t.Fatal(err)
		}
		got := report.UnpromotedBridgeSubjects
		if len(got) != 2 || !bytes.Equal(got[0], f.r.ID) || !bytes.Equal(got[1], f.q.ID) {
			t.Fatalf("unpromoted bridges %x, want [%x %x]", got, f.r.ID, f.q.ID)
		}
		changes, err := writes.Call(w.C, writes.Op{Action: "unpromote_subject", UserID: userID}, func(tx *database.Tx) ([]rowchange.Change, []rowchange.Change, error) {
			ch, err := promote.Unpromote(tx, userID, f.p.ID)
			return ch, ch, err
		})
		if err != nil {
			t.Fatal(err)
		}
		released := map[string]bool{}
		for _, ch := range changes {
			if ch.EntityType == "identity_claim" && ch.Action == rowchange.ActionDelete {
				released[string(ch.EntityID)] = true
			}
		}
		if len(released) != len(wantClaims) {
			t.Fatalf("released %d claims, want %d", len(released), len(wantClaims))
		}
		for id := range wantClaims {
			if !released[id] {
				t.Fatalf("claim %x was not released", id)
			}
		}
	})

	t.Run("a Subject with no accepted claim reports nothing", func(t *testing.T) {
		w := newBridgeWorld(t)
		f := newBridgeFixture(w)
		claim, err := identityclaims.AcceptedEntityForSubject(w.C, f.p.ID)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := writes.Call(w.C, writes.Op{Action: "release_claim", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			ch, err := deleteimpact.ReleaseClaim(tx, claim.ID)
			return struct{}{}, ch, err
		}); err != nil {
			t.Fatal(err)
		}
		if !w.filed(f.r) || !w.filed(f.q) {
			t.Fatal("the bridges should stay filed with only P's claim released")
		}
		report, err := promote.UnpromoteImpact(w.C, f.p.ID)
		if err != nil {
			t.Fatal(err)
		}
		if len(report.UnpromotedBridgeSubjects) != 0 || report.Deletes.Via != "" {
			t.Fatalf("want an empty report, got %+v", report)
		}
	})
}

func TestUnpromoteImpactListsSharedBridgeHandle(t *testing.T) {
	w := newBridgeWorld(t)
	james := w.node("person", "James", 0, 0)
	john := w.node("person", "John", 2, 0)
	johnny := w.node("person", "Johnny", 4, 0)
	w.relationship(james, john, w.spouse)
	w.relationship(james, johnny, w.spouse)
	w.join(johnny, w.promote(john).Entity.ID)
	w.promote(james)
	report, err := promote.UnpromoteImpact(w.C, james.ID)
	if err != nil {
		t.Fatal(err)
	}
	_, res, err := writes.Run(w.C, writes.Op{Action: "unpromote_subject", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
		ch, err := promote.Unpromote(tx, userID, james.ID)
		return struct{}{}, ch, err
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Deletes.Total != 2 || len(report.Deletes.Listed) != 2 || len(deletedHandleIDs(res.Effects)) != 2 {
		t.Fatalf("forecast %+v, write deleted %x, want 2 each", report.Deletes, deletedHandleIDs(res.Effects))
	}
	for _, l := range report.Deletes.Listed {
		found := false
		for _, id := range deletedHandleIDs(res.Effects) {
			found = found || bytes.Equal(id, l.ID)
		}
		if !found {
			t.Fatalf("forecast lists %x, which the write did not delete (%x)", l.ID, deletedHandleIDs(res.Effects))
		}
	}
}
