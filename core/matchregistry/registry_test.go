package matchregistry

import (
	"reflect"
	"testing"
)

// Default builds a fresh registry each call (D-2): one caller's edits reach no
// one else.
func TestDefaultIsFresh(t *testing.T) {
	first := Default()
	want := Default()
	first.Value.Words.Particles[0] = "changed"
	first.Value.Words.Folds['x'] = "added"
	if got := Default(); !reflect.DeepEqual(got, want) {
		t.Errorf("second Default() = %+v, want %+v", got, want)
	}
}
