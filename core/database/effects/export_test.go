package effects

import "github.com/mendahu/provenencia/core/database/rowchange"

// SourceScopeOfChange resolves the Source scope of changes[i] alone, with every
// delete in changes as a ghost. It is visible to the external test package,
// which can run the real deletes; no production code reaches it.
func SourceScopeOfChange(q querier, changes []rowchange.Change, i int) ([][]byte, error) {
	table, ok := entityTable[changes[i].EntityType]
	if !ok {
		return nil, errf("effects: no entry for %s", changes[i].EntityType)
	}
	w := &walk{q: q, ch: changes[i], ghosts: ghostsFrom(changes), table: table}
	return appendPath(w, table, registry[table].SourceScope, nil)
}
