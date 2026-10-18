package handlers

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/writes"
)

func TestWriteEffectsProto(t *testing.T) {
	newID := func() ([]byte, string) {
		u, err := uuid.NewV7()
		if err != nil {
			t.Fatal(err)
		}
		return u[:], u.String()
	}
	a, aStr := newID()
	b, bStr := newID()

	tests := []struct {
		name          string
		res           writes.Result
		handles       []string
		headerHandles []string
		handleScopes  []string
		revision      int64
		revisionAt    string
	}{
		{name: "empty set"},
		{
			name:         "the write's activity and its handle scopes",
			res:          writes.Result{Revision: 42, At: "2026-10-15T16:00:00Z", Effects: effects.Set{HandleScopes: [][]byte{a}}},
			handleScopes: []string{aStr},
			revision:     42,
			revisionAt:   "2026-10-15T16:00:00Z",
		},
		{
			name:          "handles and header deps stay apart",
			res:           writes.Result{Effects: effects.Set{Handles: [][]byte{a}, HeaderDeps: [][]byte{b}}},
			handles:       []string{aStr},
			headerHandles: []string{bStr},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := writeEffectsProto(tt.res)
			if !reflect.DeepEqual(got.GetHandles(), tt.handles) {
				t.Fatalf("Handles = %v, want %v", got.GetHandles(), tt.handles)
			}
			if !reflect.DeepEqual(got.GetHeaderHandles(), tt.headerHandles) {
				t.Fatalf("HeaderHandles = %v, want %v", got.GetHeaderHandles(), tt.headerHandles)
			}
			if !reflect.DeepEqual(got.GetHandleScopes(), tt.handleScopes) {
				t.Fatalf("HandleScopes = %v, want %v", got.GetHandleScopes(), tt.handleScopes)
			}
			if got.GetRevision() != tt.revision || got.GetRevisionAt() != tt.revisionAt {
				t.Fatalf("Revision = %d %q, want %d %q", got.GetRevision(), got.GetRevisionAt(), tt.revision, tt.revisionAt)
			}
		})
	}
}
