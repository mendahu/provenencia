package match

import (
	"reflect"
	"testing"

	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/namevalues"
)

// Every comparer setting has a registry value, so nothing falls back to a
// number written anywhere else. A new setting fails here until registry.go
// gives it one.
func TestRegistrySetsEverySetting(t *testing.T) {
	for name, v := range map[string]any{
		"DefaultWordRules": DefaultWordRules,
		"DefaultNames":     DefaultNames,
		"DefaultDates":     DefaultDates,
		"DefaultText":      DefaultText,
		"DefaultIntegers":  DefaultIntegers,
	} {
		checkSet(t, name, reflect.ValueOf(v))
	}
}

func checkSet(t *testing.T, path string, v reflect.Value) {
	t.Helper()
	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			checkSet(t, path+"."+v.Type().Field(i).Name, v.Field(i))
		}
	case reflect.Pointer, reflect.Interface, reflect.Map:
		if v.IsNil() {
			t.Errorf("%s has no registry value", path)
		}
	}
}

// A bare comparer and the registry's comparer score every case alike.
func TestBareComparersMatchTheRegistry(t *testing.T) {
	for _, tt := range nameCases {
		if !reflect.DeepEqual(tt.cmp, NameComparer{}) {
			continue
		}
		bare, _ := NameComparer{}.Compare(tt.a, tt.b)
		reg, _ := DefaultNames.Compare(tt.a, tt.b)
		if bare != reg {
			t.Errorf("name %s/%s: bare %.4f, registry %.4f", tt.group, tt.name, bare, reg)
		}
	}
	for _, pair := range [][2]Value{
		{date(1817, ip(5), ip(14)), date(1817, ip(5), nil)},
		{span(ip(1815), ip(1820)), date(1822, nil, nil)},
		{bound("BEF", 1820), date(1815, nil, nil)},
		{Value{Date: &datevalues.Value{StartYear: ip(1817), Qualifier: "ABT"}}, date(1820, nil, nil)},
	} {
		bare, _ := DateComparer{}.Compare(pair[0], pair[1])
		reg, _ := DefaultDates.Compare(pair[0], pair[1])
		if bare != reg {
			t.Errorf("date %+v: bare %.4f, registry %.4f", pair, bare, reg)
		}
	}
	bare, _ := TextComparer{}.Compare(text("York, Upper Canada"), text("York"))
	reg, _ := DefaultText.Compare(text("York, Upper Canada"), text("York"))
	if bare != reg {
		t.Errorf("text: bare %.4f, registry %.4f", bare, reg)
	}
}

// The default profiles use the registry's comparers, not copies.
func TestDefaultProfilesUseRegistryComparers(t *testing.T) {
	want := map[string]Comparer{
		"name": DefaultNames, "date": DefaultDates, "start_date": DefaultDates,
		"end_date": DefaultDates, "toponym": DefaultText,
	}
	for _, kind := range []string{"person", "event", "place"} {
		p, ok := DefaultProfile(kind)
		if !ok {
			t.Fatalf("no %s profile", kind)
		}
		for _, f := range p.Features {
			if w, ok := want[f.Property.Key]; ok && !reflect.DeepEqual(f.Comparer, w) {
				t.Errorf("%s.%s does not use the registry comparer", kind, f.Property.Key)
			}
		}
	}
}

// The Western pattern gives every product part type a role, so a new part
// type forces a decision here.
func TestWesternPatternCoversEveryPartType(t *testing.T) {
	for _, pt := range namevalues.PartTypes() {
		if _, ok := WesternNamePattern.PartRoles[pt.Key]; !ok {
			t.Errorf("part type %q has no role in the Western pattern", pt.Key)
		}
	}
	if _, ok := WesternNamePattern.PartRoles[""]; !ok {
		t.Error("untyped parts have no role in the Western pattern")
	}
	for key, p := range BuiltinNamePatterns {
		if p.Key != key {
			t.Errorf("pattern %q is registered as %q", p.Key, key)
		}
	}
	if _, ok := BuiltinNamePatterns.NamePattern(*DefaultNames.Pattern); !ok {
		t.Errorf("the default pattern %q is not built in", *DefaultNames.Pattern)
	}
}
