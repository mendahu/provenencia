package handlers

import (
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
