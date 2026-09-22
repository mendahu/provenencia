package namevalues

import "strings"

// Product part-type registry (compiled, not catalog rows).
// Keys are stable identifiers for UI, format profiles, and reconcile.
// User-minted part types / DB vocabulary are a later PR — do not add
// name_part_types tables here until extensibility is scheduled.

// PartTypeInfo is one product-known name part type.
type PartTypeInfo struct {
	Key string
	// L10nKey is the String Catalog key for the localized label (S7-02b).
	L10nKey string
	// Sort is display/picker order (lower first).
	Sort int
}

// PartTypes returns the product part-type registry in picker order.
func PartTypes() []PartTypeInfo {
	out := make([]PartTypeInfo, len(partTypeRegistry))
	copy(out, partTypeRegistry)
	return out
}

// KnownPartType reports whether key is in the product registry.
// Empty string is not known — callers treat empty as “untyped part”.
// Keys are matched after TrimSpace.
func KnownPartType(key string) bool {
	_, ok := partTypeByKey[strings.TrimSpace(key)]
	return ok
}

var partTypeRegistry = []PartTypeInfo{
	{Key: PartTypePrefix, L10nKey: "namePartTypes.prefix", Sort: 0},
	{Key: PartTypeGiven, L10nKey: "namePartTypes.given", Sort: 1},
	{Key: PartTypeInitial, L10nKey: "namePartTypes.initial", Sort: 2},
	{Key: PartTypeNick, L10nKey: "namePartTypes.nick", Sort: 3},
	{Key: PartTypeSurnamePrefix, L10nKey: "namePartTypes.surnamePrefix", Sort: 4},
	{Key: PartTypeSurname, L10nKey: "namePartTypes.surname", Sort: 5},
	{Key: PartTypeSuffix, L10nKey: "namePartTypes.suffix", Sort: 6},
	{Key: PartTypeUndetermined, L10nKey: "namePartTypes.undetermined", Sort: 7},
}

var partTypeByKey map[string]PartTypeInfo

func init() {
	partTypeByKey = make(map[string]PartTypeInfo, len(partTypeRegistry))
	for _, info := range partTypeRegistry {
		partTypeByKey[info.Key] = info
	}
}
