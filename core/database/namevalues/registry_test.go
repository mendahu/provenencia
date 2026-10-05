package namevalues

import "testing"

func TestPartTypesRegistry(t *testing.T) {
	types := PartTypes()
	if len(types) != 7 {
		t.Fatalf("len=%d", len(types))
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
