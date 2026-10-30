package dbtest

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/mendahu/provenencia/core/database/rowchange"
)

// RequireUpdateBeforeDelete finds an update change for updateEntityID before the
// deleteEntityType delete in changes and deep-equals its Fields to wantFields.
func RequireUpdateBeforeDelete(
	t testing.TB,
	changes []rowchange.Change,
	deleteEntityType, updateEntityType string,
	updateEntityID []byte,
	wantFields map[string]rowchange.FieldDiff,
) {
	t.Helper()
	deleteIdx := -1
	for i, ch := range changes {
		if ch.EntityType == deleteEntityType && ch.Action == rowchange.ActionDelete {
			deleteIdx = i
		}
	}
	if deleteIdx < 0 {
		t.Fatalf("no %s delete change", deleteEntityType)
	}
	var entityUpdate *rowchange.Change
	for i, ch := range changes {
		if ch.EntityType != updateEntityType || ch.Action != rowchange.ActionUpdate || !bytes.Equal(ch.EntityID, updateEntityID) {
			continue
		}
		if entityUpdate != nil {
			t.Fatalf("duplicate %s update for entity", updateEntityType)
		}
		c := ch
		entityUpdate = &c
		if i >= deleteIdx {
			t.Fatalf("%s update after %s delete at %d", updateEntityType, deleteEntityType, i)
		}
	}
	if entityUpdate == nil {
		t.Fatalf("no %s update", updateEntityType)
	}
	if len(entityUpdate.Fields) != len(wantFields) {
		t.Fatalf("fields %+v want keys from %+v", entityUpdate.Fields, wantFields)
	}
	for k, w := range wantFields {
		g, ok := entityUpdate.Fields[k]
		if !ok {
			t.Fatalf("missing field %q in %+v", k, entityUpdate.Fields)
		}
		if !reflect.DeepEqual(g.New, w.New) {
			t.Fatalf("field %q New %+v want %+v", k, g.New, w.New)
		}
	}
}
