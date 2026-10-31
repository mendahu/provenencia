package namevalue

import (
	"slices"
	"testing"
)

func TestPartTypesRegistry(t *testing.T) {
	types := PartTypes()
	want := []string{"prefix", "given", "nick", "surname_prefix", "surname", "former_surname", "suffix", "undetermined"}
	var keys []string
	for _, info := range types {
		keys = append(keys, info.Key)
	}
	if !slices.Equal(keys, want) {
		t.Fatalf("keys=%v, want %v", keys, want)
	}
	if err := Validate(Value{Form: "Mary (Smith) Jones", Parts: []Part{{Idx: 0, Type: PartTypeFormerSurname, Value: "Smith"}}}); err != nil {
		t.Fatalf("former_surname rejected: %v", err)
	}
	for i, info := range types {
		if info.Key == "" || info.L10nKey == "" {
			t.Fatalf("empty fields at %d: %+v", i, info)
		}
		if !KnownPartType(info.Key) {
			t.Fatalf("KnownPartType(%q)=false", info.Key)
		}
		if i > 0 && types[i-1].Sort > info.Sort {
			t.Fatalf("sort order broken at %d", i)
		}
	}
	if KnownPartType("") || KnownPartType("first_name") || KnownPartType("initial") {
		t.Fatal("unexpected known types")
	}
}
