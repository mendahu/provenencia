package handlers

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/mendahu/provenencia/api/proto/engine"
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

func TestWriteEffectsProtoLists(t *testing.T) {
	got := writeEffectsProto(writes.Result{Effects: effects.Set{
		Lists: []effects.List{effects.ListVocabulary, effects.ListPropertyUsage},
	}})
	if want := []string{"vocabulary", "property_usage"}; !reflect.DeepEqual(got.GetLists(), want) {
		t.Fatalf("Lists = %v, want %v", got.GetLists(), want)
	}
	if got := writeEffectsProto(writes.Result{}); got.GetLists() != nil {
		t.Fatalf("empty set Lists = %v", got.GetLists())
	}
}

func TestWriteEffectsProtoReads(t *testing.T) {
	g1 := uuid.Must(uuid.NewV7())
	g2 := uuid.Must(uuid.NewV7())
	got := writeEffectsProto(writes.Result{Effects: effects.Set{
		EvidenceGraphs: [][]byte{g1[:]},
		Reads: []effects.ReadIDs{
			{Read: effects.ReadCitations, IDs: [][]byte{g1[:]}},
			{Read: effects.ReadEvidenceGraphProgress, IDs: [][]byte{g1[:], g2[:]}},
		},
	}})
	if len(got.GetReads()) != 2 {
		t.Fatalf("Reads = %v", got.GetReads())
	}
	if r := got.GetReads()[0]; r.GetRead() != "citations" || !reflect.DeepEqual(r.GetIds(), []string{g1.String()}) {
		t.Fatalf("Reads[0] = %v", r)
	}
	if r := got.GetReads()[1]; r.GetRead() != "evidence_graph_progress" || !reflect.DeepEqual(r.GetIds(), []string{g1.String(), g2.String()}) {
		t.Fatalf("Reads[1] = %v", r)
	}
	if want := []string{g1.String()}; !reflect.DeepEqual(got.GetEvidenceGraphs(), want) {
		t.Fatalf("EvidenceGraphs = %v, want %v", got.GetEvidenceGraphs(), want)
	}
	if got := writeEffectsProto(writes.Result{}); got.GetReads() != nil {
		t.Fatalf("empty set Reads = %v", got.GetReads())
	}
}

// TestRetiredFieldsStayReserved pins the wire fields this PR retired: an old
// client's bytes must not decode into a field with another meaning.
func TestRetiredFieldsStayReserved(t *testing.T) {
	effectsDesc := (&engine.WriteEffects{}).ProtoReflect().Descriptor()
	if !effectsDesc.ReservedNames().Has("citation_sources") || !effectsDesc.ReservedRanges().Has(11) {
		t.Error("WriteEffects.citation_sources (11) is not reserved")
	}
	if effectsDesc.Fields().ByName("citation_sources") != nil || effectsDesc.Fields().ByNumber(11) != nil {
		t.Error("WriteEffects still declares field 11")
	}
	progressDesc := (&engine.EvidenceGraphProgress{}).ProtoReflect().Descriptor()
	if !progressDesc.ReservedNames().Has("source_id") || !progressDesc.ReservedRanges().Has(1) {
		t.Error("EvidenceGraphProgress.source_id (1) is not reserved")
	}
	if progressDesc.Fields().ByName("source_id") != nil {
		t.Error("EvidenceGraphProgress still declares source_id")
	}
}
