package matchregistry

import (
	"fmt"
	"os"
	"reflect"
	"slices"
	"sort"
	"strings"
	"testing"
)

// notYetBuilt are the prefixes of inventory rows whose registry fields land
// with their layer. The layer's PR deletes its prefixes.
var notYetBuilt = []string{
	"evidence.",
}

// catalogData are registry fields that hold catalog data passed in at run
// time, never tunables. They get no inventory row; each is named under "Not
// knobs". The one list the guard keeps beside notYetBuilt.
var catalogData = []string{"value.Names.Aliases", "value.Names.Patterns"}

// TestKnobInventoryMatchesRegistry keeps docs/matching-model.md's knob
// inventory in step with Default(): every registry field has a row, and every
// row is a field or under a not-yet-built prefix.
func TestKnobInventoryMatchesRegistry(t *testing.T) {
	rows := inventoryKeys(t)
	fields := registryKeys(reflect.ValueOf(Default()))
	notKnobs := notKnobsSection(t)

	for _, c := range catalogData {
		if !typeHasPath(reflect.TypeOf(Default()), strings.Split(c, ".")) {
			t.Errorf("catalogData entry %s names no field of the registry type", c)
		}
		if slices.ContainsFunc(rows, func(r string) bool { return strings.HasPrefix(r, c) }) {
			t.Errorf("catalogData entry %s has an inventory row: catalog data is not a knob", c)
		}
		if !strings.Contains(notKnobs, "`"+c+"`") {
			t.Errorf("catalogData entry %s is not named in backticks under ## Not knobs", c)
		}
	}

	covered := func(key string) bool {
		return slices.ContainsFunc(notYetBuilt, func(p string) bool { return strings.HasPrefix(key, p) })
	}
	for _, f := range fields {
		if !slices.Contains(rows, f) {
			t.Errorf("registry field %s has no row in the knob inventory", f)
		}
	}
	for _, r := range rows {
		if !slices.Contains(fields, r) && !covered(r) {
			t.Errorf("inventory row %s names no registry field and is under no notYetBuilt prefix", r)
		}
	}
	for _, p := range notYetBuilt {
		if slices.ContainsFunc(fields, func(f string) bool { return strings.HasPrefix(f, p) }) {
			t.Errorf("notYetBuilt prefix %q matches a registry field: delete the prefix", p)
		}
		if !slices.ContainsFunc(rows, func(r string) bool { return strings.HasPrefix(r, p) }) {
			t.Errorf("notYetBuilt prefix %q matches no inventory row", p)
		}
	}
}

// typeHasPath reports whether the dotted path (the first element the
// lowercased top-level field name) names a field of t.
func typeHasPath(t reflect.Type, path []string) bool {
	for i, name := range path {
		for t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
		if t.Kind() != reflect.Struct {
			return false
		}
		var found bool
		for j := range t.NumField() {
			f := t.Field(j)
			if f.Name == name || (i == 0 && strings.ToLower(f.Name) == name) {
				t, found = f.Type, true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// notKnobsSection is the body of the "Not knobs" section.
func notKnobsSection(t *testing.T) string {
	body, err := os.ReadFile("../../docs/matching-model.md")
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	in := false
	for _, line := range strings.Split(string(body), "\n") {
		if strings.HasPrefix(line, "## ") {
			in = strings.TrimSpace(strings.TrimPrefix(line, "## ")) == "Not knobs"
			continue
		}
		if in {
			out = append(out, line)
		}
	}
	if len(out) == 0 {
		t.Fatal("no ## Not knobs section found")
	}
	return strings.Join(out, "\n")
}

// inventoryKeys are the first backticked cell of each row of the "Knob
// inventory" section.
func inventoryKeys(t *testing.T) []string {
	body, err := os.ReadFile("../../docs/matching-model.md")
	if err != nil {
		t.Fatal(err)
	}
	var keys []string
	in := false
	for _, line := range strings.Split(string(body), "\n") {
		if strings.HasPrefix(line, "## ") {
			in = strings.TrimSpace(strings.TrimPrefix(line, "## ")) == "Knob inventory"
			continue
		}
		if !in || !strings.HasPrefix(line, "| `") {
			continue
		}
		cell := strings.TrimPrefix(line, "| `")
		if end := strings.Index(cell, "`"); end >= 0 {
			keys = append(keys, cell[:end])
		}
	}
	if len(keys) == 0 {
		t.Fatal("no rows found under ## Knob inventory")
	}
	return keys
}

// registryKeys are the dotted paths of v's leaves: the top-level field name
// lowercased, then struct field names as written. A map whose element is a
// struct or a map is walked by its sorted keys; any other field is a leaf.
func registryKeys(v reflect.Value) []string {
	var out []string
	var walk func(path string, v reflect.Value)
	walk = func(path string, v reflect.Value) {
		if slices.Contains(catalogData, path) {
			return
		}
		switch v.Kind() {
		case reflect.Struct:
			for i := range v.NumField() {
				name := v.Type().Field(i).Name
				if path == "" {
					name = strings.ToLower(name)
				} else {
					name = path + "." + name
				}
				walk(name, v.Field(i))
			}
		case reflect.Map:
			if k := v.Type().Elem().Kind(); k == reflect.Struct || k == reflect.Map {
				keys := v.MapKeys()
				sort.Slice(keys, func(i, j int) bool { return fmt.Sprint(keys[i]) < fmt.Sprint(keys[j]) })
				for _, key := range keys {
					walk(path+"."+fmt.Sprint(key), v.MapIndex(key))
				}
				return
			}
			out = append(out, path)
		default:
			out = append(out, path)
		}
	}
	walk("", v)
	return out
}
