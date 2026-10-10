package promotealign_test

import (
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/writes"
)

func TestSourceGraphSecondReadIssuesNoQueries(t *testing.T) {
	f := newFixture(t)
	s := f.person(f.source, f.artifact, "Ann Lee")
	res := f.promote(s)
	g := f.c.Graph()
	g.Drop()
	sg, err := g.Source(f.source.ID)
	must(t, err)
	if len(sg.Members) != 1 || string(sg.Members[0].EntityID) != string(res.Entity.ID) {
		t.Fatalf("members %+v", sg.Members)
	}
	if _, err := g.Node(res.Entity.ID); err != nil {
		t.Fatal(err)
	}
	var n int
	g.SetQueryCounterForTest(func() { n++ })
	again, err := g.Source(f.source.ID)
	must(t, err)
	if n != 0 {
		t.Fatalf("second read issued %d graph queries", n)
	}
	if len(again.Members) != 1 || string(again.Members[0].EntityID) != string(res.Entity.ID) {
		t.Fatalf("members %+v", again.Members)
	}
	node, err := g.Node(res.Entity.ID)
	must(t, err)
	if node == nil || node.Ref == "" {
		t.Fatalf("handle %+v", node)
	}
}

func TestPromoteDropsTheSource(t *testing.T) {
	f := newFixture(t)
	s := f.person(f.source, f.artifact, "Ann Lee")
	g := f.c.Graph()
	g.Drop()
	before, err := g.Source(f.source.ID)
	must(t, err)
	if len(before.Members) != 0 {
		t.Fatalf("members before promote %+v", before.Members)
	}
	res := f.promote(s)
	after, err := g.Source(f.source.ID)
	must(t, err)
	if len(after.Members) != 1 || string(after.Members[0].EntityID) != string(res.Entity.ID) {
		t.Fatalf("members after promote %+v", after.Members)
	}
}

func TestHandleLabelDoesNotDropTheSource(t *testing.T) {
	f := newFixture(t)
	s := f.person(f.source, f.artifact, "Ann Lee")
	res := f.promote(s)
	g := f.c.Graph()
	g.Drop()
	if _, err := g.Source(f.source.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := g.Node(res.Entity.ID); err != nil {
		t.Fatal(err)
	}
	_, err := writes.Call(f.c, writes.Op{Action: "update_canonical_entity", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
		if _, err := tx.Exec(`UPDATE canonical_entities SET label = ? WHERE id = ?`, "Renamed", res.Entity.ID); err != nil {
			return struct{}{}, nil, err
		}
		return struct{}{}, []rowchange.Change{{
			EntityType: "canonical_entity",
			EntityID:   res.Entity.ID,
			Action:     rowchange.ActionUpdate,
			Fields:     map[string]rowchange.FieldDiff{"label": {Old: "", New: "Renamed"}},
		}}, nil
	})
	must(t, err)
	var n int
	g.SetQueryCounterForTest(func() { n++ })
	if _, err := g.Source(f.source.ID); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("label edit reloaded the source in %d queries", n)
	}
	node, err := g.Node(res.Entity.ID)
	must(t, err)
	if node == nil || node.Label != "Renamed" {
		t.Fatalf("label %+v", node)
	}
}
