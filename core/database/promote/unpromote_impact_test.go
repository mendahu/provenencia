package promote_test

import (
	"testing"

	"github.com/mendahu/provenencia/core/database/deleteimpact"
	"github.com/mendahu/provenencia/core/database/promote"
)

func TestUnpromoteImpactListsBridgeHandles(t *testing.T) {
	w := newBridgeWorld(t)
	h := newHousehold(w)
	group, err := promote.UnpromoteImpact(w.c, h.james.ID)
	if err != nil {
		t.Fatal(err)
	}
	if group.Via != deleteimpact.ViaLastMember || group.Total < 2 {
		t.Fatalf("want primary and association deletes, got %+v", group)
	}
}

func TestUnpromoteImpactEmptyWhenHandleStays(t *testing.T) {
	w := newBridgeWorld(t)
	james := w.node("person", "James", 0, 0)
	john := w.node("person", "John", 2, 0)
	entity := w.promote(james).Entity.ID
	w.join(john, entity)
	group, err := promote.UnpromoteImpact(w.c, james.ID)
	if err != nil {
		t.Fatal(err)
	}
	if group.Via != "" {
		t.Fatalf("handle stays: %+v", group)
	}
}
