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

// Handles resolves handle ids.
func Handles(tx *sql.Tx, changes []rowchange.Change) ([][]byte, error) {
	return resolve(tx, changes, func(e Effect) Path { return e.Handles })
}

// Resolve is the whole effect of a batch: sources, handles, search documents,
// and whether vocabulary labels or structure changed.
func Resolve(tx *sql.Tx, changes []rowchange.Change) (Set, error) {
	if tx == nil {
		return Set{}, errf("effects: nil tx")
	}
	ghosts := ghostsFrom(changes)
	var set Set
	search := map[string][][]byte{}
	var searchOrder []string
	for _, ch := range changes {
		table, ok := entityTable[ch.EntityType]
		if !ok {
			return Set{}, errf("effects: no entry for %s", ch.EntityType)
		}
		eff := registry[table]
		if eff.Vocabulary {
			set.Vocabulary = true
		}
		w := &walk{q: tx, ch: ch, ghosts: ghosts, table: table}
		var err error
		set.Sources, err = appendPath(w, table, eff.Source, set.Sources)
		if err != nil {
			return Set{}, err
		}
		set.Handles, err = appendPath(w, table, eff.Handles, set.Handles)
		if err != nil {
			return Set{}, err
		}
		if !eff.Structure.zero() {
			ids, err := runPath(w, table, eff.Structure)
			if err != nil {
				return Set{}, err
			}
			if len(ids) > 0 {
				set.Structure = true
			}
		}
		for _, doc := range eff.Search {
			ids, err := runPath(w, table, doc.Path)
			if err != nil {
				return Set{}, err
			}
			if len(ids) == 0 {
				continue
			}
			if _, ok := search[doc.Kind]; !ok {
				searchOrder = append(searchOrder, doc.Kind)
			}
			search[doc.Kind] = appendIDs(search[doc.Kind], ids)
		}
	}
	for _, kind := range searchOrder {
		set.Search = append(set.Search, SearchIDs{Kind: kind, IDs: search[kind]})
	}
	return set, nil
}

func ghostsFrom(changes []rowchange.Change) ghostMap {
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
	return ghosts
}

func runPath(w *walk, table string, path Path) ([][]byte, error) {
	if path.zero() {
		return nil, nil
	}
	w.table = table
	return path.run(w, nil, true)
}

func appendPath(w *walk, table string, path Path, out [][]byte) ([][]byte, error) {
	ids, err := runPath(w, table, path)
	if err != nil {
		return nil, err
	}
	return appendIDs(out, ids), nil
}

func appendIDs(out, ids [][]byte) [][]byte {
	seen := map[string]struct{}{}
	for _, id := range out {
		if len(id) == 16 {
			seen[string(id)] = struct{}{}
		}
	}
	for _, id := range ids {
		if len(id) != 16 {
			continue
		}
		if _, ok := seen[string(id)]; ok {
			continue
		}
		seen[string(id)] = struct{}{}
		out = append(out, append([]byte(nil), id...))
	}
	return out
}

func resolve(tx *sql.Tx, changes []rowchange.Change, pick func(Effect) Path) ([][]byte, error) {
	if tx == nil {
		return nil, errf("effects: nil tx")
	}
	ghosts := ghostsFrom(changes)
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
