package workspacelocation_test

import (
	"testing"

	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/workspacelocation"
)

func TestConclusion(t *testing.T) {
	tests := []struct {
		key    string
		want   string
		ok     bool
		entity string
		ref    string
		title  string
	}{
		{key: vocabkey.Person, want: workspacelocation.Persons, ok: true, entity: "e1", ref: "PER-A", title: "Ann"},
		{key: vocabkey.Event, want: workspacelocation.Events, ok: true, entity: "e2", ref: "EVT-B", title: "Birth"},
		{key: vocabkey.Place, want: workspacelocation.Places, ok: true, entity: "e3", ref: "PLC-C", title: "York"},
		{key: "relationship", ok: false},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			loc, ok := workspacelocation.Conclusion(tt.key, tt.entity, tt.ref, tt.title)
			if ok != tt.ok {
				t.Fatalf("ok %v want %v", ok, tt.ok)
			}
			if !tt.ok {
				return
			}
			if loc.Section != tt.want || loc.EntityID != tt.entity || loc.Ref != tt.ref || loc.Title != tt.title {
				t.Fatalf("%+v", loc)
			}
		})
	}
}
