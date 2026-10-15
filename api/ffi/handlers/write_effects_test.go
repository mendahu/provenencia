package handlers

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/mendahu/provenencia/core/database/effects"
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
		set           effects.Set
		handles       []string
		headerHandles []string
		handleScopes  []string
	}{
		{name: "empty set"},
		{name: "handle scopes only", set: effects.Set{HandleScopes: [][]byte{a}}, handleScopes: []string{aStr}},
		{
			name:          "handles and header deps stay apart",
			set:           effects.Set{Handles: [][]byte{a}, HeaderDeps: [][]byte{b}},
			handles:       []string{aStr},
			headerHandles: []string{bStr},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := writeEffectsProto(tt.set)
			if !reflect.DeepEqual(got.GetHandles(), tt.handles) {
				t.Fatalf("Handles = %v, want %v", got.GetHandles(), tt.handles)
			}
			if !reflect.DeepEqual(got.GetHeaderHandles(), tt.headerHandles) {
				t.Fatalf("HeaderHandles = %v, want %v", got.GetHeaderHandles(), tt.headerHandles)
			}
			if !reflect.DeepEqual(got.GetHandleScopes(), tt.handleScopes) {
				t.Fatalf("HandleScopes = %v, want %v", got.GetHandleScopes(), tt.handleScopes)
			}
		})
	}
}
