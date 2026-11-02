package handlers

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"

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
		name           string
		res            writes.Result
		handles        []string
		summaryHandles []string
		handleScopes   []string
		revision       int64
		revisionAt     string
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
			name:           "handles and summary deps stay apart",
			res:            writes.Result{Effects: effects.Set{Handles: [][]byte{a}, SummaryDeps: [][]byte{b}}},
			handles:        []string{aStr},
			summaryHandles: []string{bStr},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := writeEffectsProto(tt.res)
			if !reflect.DeepEqual(got.GetHandles(), tt.handles) {
				t.Fatalf("Handles = %v, want %v", got.GetHandles(), tt.handles)
			}
			if !reflect.DeepEqual(got.GetSummaryHandles(), tt.summaryHandles) {
				t.Fatalf("SummaryHandles = %v, want %v", got.GetSummaryHandles(), tt.summaryHandles)
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
	applyDesc := (&engine.ApplyPromoteGraphAlignmentResponse{}).ProtoReflect().Descriptor()
	for _, n := range []protoreflect.Name{"revision", "written"} {
		if !applyDesc.ReservedNames().Has(n) || applyDesc.Fields().ByName(n) != nil {
			t.Errorf("ApplyPromoteGraphAlignmentResponse.%s is not reserved", n)
		}
	}
	for _, n := range []protoreflect.FieldNumber{1, 2} {
		if !applyDesc.ReservedRanges().Has(n) || applyDesc.Fields().ByNumber(n) != nil {
			t.Errorf("ApplyPromoteGraphAlignmentResponse field %d is not reserved", n)
		}
	}
	methodDesc := engine.Method(0).Descriptor()
	if !methodDesc.ReservedRanges().Has(85) || !methodDesc.ReservedNames().Has("METHOD_PROMOTE_SUBJECT") ||
		methodDesc.Values().ByNumber(85) != nil {
		t.Error("METHOD_PROMOTE_SUBJECT (85) is not reserved")
	}
}

// effectsPending is the audited writes that still return no WriteEffects, by
// request name, with the PR that converts each. S14-17 converts them and
// deletes the entries; when the map is empty it deletes the map.
var effectsPending = map[string]string{
	"CreateSourceTypeRequest":          "S14-17",
	"UpdateSourceTypeRequest":          "S14-17",
	"DeleteSourceTypeRequest":          "S14-17",
	"CreateMetadataFieldRequest":       "S14-17",
	"UpdateMetadataFieldRequest":       "S14-17",
	"DeleteMetadataFieldRequest":       "S14-17",
	"AssignTypeFieldRequest":           "S14-17",
	"RemoveTypeFieldRequest":           "S14-17",
	"CreatePropertyRequest":            "S14-17",
	"UpdatePropertyRequest":            "S14-17",
	"DeletePropertyRequest":            "S14-17",
	"CreatePropertyTermRequest":        "S14-17",
	"UpdatePropertyTermRequest":        "S14-17",
	"DeletePropertyTermRequest":        "S14-17",
	"AssignSubjectTypePropertyRequest": "S14-17",
	"RemoveSubjectTypePropertyRequest": "S14-17",
}

// missingEffects returns one message per audited write, a message named
// ...Request with a user_id field, whose ...Response has no WriteEffects
// field and is not in pending, and one per pending entry that names no such
// request or whose response already has the field.
func missingEffects(file protoreflect.FileDescriptor, pending map[string]string) []string {
	messages := map[string]protoreflect.MessageDescriptor{}
	for i := 0; i < file.Messages().Len(); i++ {
		m := file.Messages().Get(i)
		messages[string(m.Name())] = m
	}
	effectsName := (&engine.WriteEffects{}).ProtoReflect().Descriptor().FullName()
	hasEffects := func(m protoreflect.MessageDescriptor) bool {
		for i := 0; i < m.Fields().Len(); i++ {
			if f := m.Fields().Get(i); f.Message() != nil && f.Message().FullName() == effectsName {
				return true
			}
		}
		return false
	}
	var out []string
	audited := map[string]bool{}
	for name, m := range messages {
		if !strings.HasSuffix(name, "Request") || m.Fields().ByName("user_id") == nil {
			continue
		}
		audited[name] = true
		resp, ok := messages[strings.TrimSuffix(name, "Request")+"Response"]
		switch {
		case !ok:
			out = append(out, name+" has no response message")
		case hasEffects(resp):
			if _, listed := pending[name]; listed {
				out = append(out, name+" returns effects but is still in effectsPending")
			}
		default:
			if _, listed := pending[name]; !listed {
				out = append(out, name+" returns no WriteEffects")
			}
		}
	}
	for name := range pending {
		if !audited[name] {
			out = append(out, name+" is in effectsPending but is not an audited request")
		}
	}
	slices.Sort(out)
	return out
}

// TestEveryAuditedWriteReturnsEffects holds every request with a user_id to a
// response that carries WriteEffects, so a write cannot drop what it touched
// and leave the app to guess. effectsPending is the shrinking allow-list.
func TestEveryAuditedWriteReturnsEffects(t *testing.T) {
	if got := missingEffects(engine.File_engine_proto, effectsPending); len(got) != 0 {
		t.Fatalf("audited writes without effects:\n%s", strings.Join(got, "\n"))
	}

	// The check itself, on a small file: XRequest{user_id} with a bare XResponse
	// fails, one with WriteEffects passes.
	fixture := func(t *testing.T, withEffects bool) protoreflect.FileDescriptor {
		t.Helper()
		str := descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum()
		msg := descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum()
		opt := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum()
		response := &descriptorpb.DescriptorProto{Name: proto.String("XResponse")}
		if withEffects {
			response.Field = []*descriptorpb.FieldDescriptorProto{{
				Name: proto.String("effects"), Number: proto.Int32(1), Type: msg, Label: opt,
				TypeName: proto.String("." + string(engine.File_engine_proto.Package()) + ".WriteEffects"),
			}}
		}
		fd, err := protodesc.NewFile(&descriptorpb.FileDescriptorProto{
			Name:       proto.String("fixture.proto"),
			Package:    proto.String("fixture"),
			Syntax:     proto.String("proto3"),
			Dependency: []string{engine.File_engine_proto.Path()},
			MessageType: []*descriptorpb.DescriptorProto{
				{Name: proto.String("XRequest"), Field: []*descriptorpb.FieldDescriptorProto{
					{Name: proto.String("user_id"), Number: proto.Int32(1), Type: str, Label: opt},
				}},
				response,
			},
		}, protoregistry.GlobalFiles)
		if err != nil {
			t.Fatal(err)
		}
		return fd
	}
	t.Run("a response without effects fails naming the request", func(t *testing.T) {
		got := missingEffects(fixture(t, false), nil)
		if len(got) != 1 || !strings.HasPrefix(got[0], "XRequest") {
			t.Fatalf("got %v", got)
		}
	})
	t.Run("a response with effects passes", func(t *testing.T) {
		if got := missingEffects(fixture(t, true), nil); len(got) != 0 {
			t.Fatalf("got %v", got)
		}
	})
	t.Run("a pending entry for a request with effects is stale", func(t *testing.T) {
		got := missingEffects(fixture(t, true), map[string]string{"XRequest": "S14-17"})
		if len(got) != 1 || !strings.HasPrefix(got[0], "XRequest") {
			t.Fatalf("got %v", got)
		}
	})
	t.Run("a pending entry for no request is stale", func(t *testing.T) {
		got := missingEffects(fixture(t, true), map[string]string{"GoneRequest": "S14-17"})
		if len(got) != 1 || !strings.HasPrefix(got[0], "GoneRequest") {
			t.Fatalf("got %v", got)
		}
	})
	t.Run("a pending entry allows a response without effects", func(t *testing.T) {
		if got := missingEffects(fixture(t, false), map[string]string{"XRequest": "S14-17"}); len(got) != 0 {
			t.Fatalf("got %v", got)
		}
	})
}
