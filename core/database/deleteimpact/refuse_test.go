package deleteimpact

import (
	"errors"
	"testing"
)

func TestRefuse(t *testing.T) {
	inUse := errors.New("in_use")
	origin := errors.New("origin")
	edge := errors.New("edge")
	codes := Codes{InUse: inUse, OriginLocked: origin, EdgeLocked: edge, NotFound: ErrInvalid}

	if err := Refuse(Report{Allowed: true, Gate: GateOK}, codes); err != nil {
		t.Fatalf("allowed %v", err)
	}
	if err := Refuse(Report{Gate: GateInbound}, codes); !errors.Is(err, inUse) {
		t.Fatalf("inbound %v", err)
	}
	if err := Refuse(Report{Gate: GateOriginLocked}, codes); !errors.Is(err, origin) {
		t.Fatalf("origin %v", err)
	}
	if err := Refuse(Report{Gate: GateEdgeLocked}, codes); !errors.Is(err, edge) {
		t.Fatalf("edge %v", err)
	}
	if err := Refuse(Report{Gate: GateNotFound}, codes); !errors.Is(err, ErrInvalid) {
		t.Fatalf("not_found %v", err)
	}
	if err := Refuse(Report{Gate: GateOriginLocked}, Codes{InUse: inUse}); !errors.Is(err, inUse) {
		t.Fatalf("origin fallback %v", err)
	}
}
