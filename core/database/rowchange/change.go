// Package rowchange is one row-level mutation: the entity, the action, and the
// old and new field values. audit.Record stores these. effects reads them.
// Neither package imports the other.
package rowchange

const (
	ActionCreate = "create"
	ActionUpdate = "update"
	ActionDelete = "delete"
)

// FieldDiff is one field's previous and resulting value in changes_json.
type FieldDiff struct {
	Old any `json:"old"`
	New any `json:"new"`
}

// Change is one row-level mutation under a revision.
type Change struct {
	EntityType string
	EntityID   []byte
	Action     string
	Fields     map[string]FieldDiff
}

// FullRow is a create change: every key is present with old = nil and new = value.
// Unset columns should be passed as nil so JSON encodes them as null.
func FullRow(fields map[string]any) map[string]FieldDiff {
	out := make(map[string]FieldDiff, len(fields))
	for k, v := range fields {
		out[k] = FieldDiff{Old: nil, New: v}
	}
	return out
}

// DeletedRow is a delete change: every key is present with old = value and new = nil.
// Unset columns should be passed as nil so JSON encodes them as null.
func DeletedRow(fields map[string]any) map[string]FieldDiff {
	out := make(map[string]FieldDiff, len(fields))
	for k, v := range fields {
		out[k] = FieldDiff{Old: v, New: nil}
	}
	return out
}
