package handlers

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestParseIDRequiresV7(t *testing.T) {
	v7 := uuid.Must(uuid.NewV7())
	got, err := parseID(v7.String())
	if err != nil || len(got) != 16 {
		t.Fatalf("v7: %v len=%d", err, len(got))
	}
	v4 := uuid.Must(uuid.NewRandom())
	if v4.Version() == 7 {
		t.Fatal("expected non-v7")
	}
	if _, err := parseID(v4.String()); err == nil {
		t.Fatal("want reject non-v7")
	}
	if _, err := parseID("not-a-uuid"); err == nil {
		t.Fatal("want reject garbage")
	}
}

func TestParseIDs(t *testing.T) {
	a, b := uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7())
	got, err := parseIDs([]string{a.String(), b.String()})
	if err != nil || len(got) != 2 || string(got[0]) != string(a[:]) || string(got[1]) != string(b[:]) {
		t.Fatalf("two ids: %x %v", got, err)
	}
	if got, err := parseIDs(nil); got != nil || err != nil {
		t.Fatalf("empty: %x %v", got, err)
	}
	if _, err := parseIDs([]string{a.String(), "nope"}); !errors.Is(err, errInvalidID) {
		t.Fatalf("one bad: %v", err)
	}
}
