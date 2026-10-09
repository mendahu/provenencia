package effects

import (
	"bytes"
	"database/sql"
	"fmt"

	"github.com/mendahu/provenencia/core/database/rowchange"
)

type ghostMap map[string]map[string]any

func ghostKey(entity string, id []byte) string { return entity + ":" + string(id) }

func (g ghostMap) id(entity string, id []byte, field string) []byte {
	if g == nil {
		return nil
	}
	return uuidValue(g[ghostKey(entity, id)][field])
}

func errf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}

// Sources resolves the Source ids a batch of changes touches. Column values
// come from the diff when it has them, from the live row when it does not,
// and from a delete's old fields only after the row is gone.
func Sources(tx *sql.Tx, changes []rowchange.Change) ([][]byte, error) {
	return resolve(tx, changes, func(e Effect) Path { return e.Source })
}

// Handles resolves handle ids. Nothing in the write path calls this yet.
func Handles(tx *sql.Tx, changes []rowchange.Change) ([][]byte, error) {
	return resolve(tx, changes, func(e Effect) Path { return e.Handles })
}

func resolve(tx *sql.Tx, changes []rowchange.Change, pick func(Effect) Path) ([][]byte, error) {
	if tx == nil {
		return nil, errf("effects: nil tx")
	}
	ghosts := ghostMap{}
	for _, ch := range changes {
		if ch.Action != rowchange.ActionDelete {
			continue
		}
		old := make(map[string]any, len(ch.Fields))
		for k, v := range ch.Fields {
			old[k] = v.Old
		}
		ghosts[ghostKey(ch.EntityType, ch.EntityID)] = old
	}
	var out [][]byte
	seen := map[string]struct{}{}
	for _, ch := range changes {
		table, ok := entityTable[ch.EntityType]
		if !ok {
			return nil, errf("effects: no entry for %s", ch.EntityType)
		}
		path := pick(registry[table])
		if path.zero() {
			continue
		}
		w := &walk{q: tx, ch: ch, ghosts: ghosts, table: table}
		ids, err := path.run(w, nil, true)
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			if len(id) != 16 {
				continue
			}
			key := string(id)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, append([]byte(nil), id...))
		}
	}
	return out, nil
}

func sameIDs(a, b [][]byte) bool {
	if len(a) != len(b) {
		return false
	}
	for _, id := range a {
		found := false
		for _, other := range b {
			if bytes.Equal(id, other) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
