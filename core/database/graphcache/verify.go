package graphcache

import (
	"bytes"
	"fmt"
	"sort"
)

// TruthRow is one value a fresh reconcile would write for an entity.
type TruthRow struct {
	PropertyID []byte
	Rank       int
	Reason     string
	Text       string
	HasText    bool
	Integer    int64
	HasInteger bool
	TermID     []byte
	EntityID   []byte
	Date       []byte
	Name       []byte
}

// TruthFunc recomputes values from observations. autoreconciler registers it.
// Repeating the auto-reconciler read would hide a short recompute set.
type TruthFunc func(q Querier, ids [][]byte) (map[string][]TruthRow, error)

var truthFn TruthFunc

// SetTruth registers the observation recompute Verify compares against.
func SetTruth(fn TruthFunc) { truthFn = fn }

// Verify rebuilds each loaded node from the catalog. Values come from the
// registered truth recompute. Identity, members, and link endpoints are read
// back from the committed rows. A loaded kind index must name every unmerged
// handle of that type. Verify does not repair the store.
func (g *Graph) Verify() error {
	if g == nil {
		return nil
	}
	if truthFn == nil {
		return fmt.Errorf("graphcache: truth recompute is not registered")
	}
	ids := make([][]byte, 0, len(g.nodes))
	for _, n := range g.nodes {
		ids = append(ids, n.ID)
	}
	if err := g.verifyIDs(ids); err != nil {
		return err
	}
	for typeID := range g.kindOn {
		if err := g.verifyKind([]byte(typeID)); err != nil {
			return err
		}
	}
	if displayFn == nil {
		return nil
	}
	for key, stored := range g.displays {
		if err := displayFn(g, []byte(key), stored); err != nil {
			return err
		}
	}
	return nil
}

func (g *Graph) verifyIDs(ids [][]byte) error {
	if len(ids) == 0 {
		return nil
	}
	truth, err := truthFn(g.db, ids)
	if err != nil {
		return err
	}
	rows, err := g.entityRows(ids)
	if err != nil {
		return err
	}
	members, err := g.loadMembers(ids)
	if err != nil {
		return err
	}
	var assocIDs [][]byte
	for _, id := range ids {
		n := g.nodes[string(id)]
		if n == nil {
			continue
		}
		row, ok := rows[string(id)]
		if !ok {
			return fmt.Errorf("graphcache: %x is loaded but has no entity row", id)
		}
		if n.Ref != row.ref || !sameID(n.SubjectTypeID, row.typeID) || n.Merged != row.merged {
			return fmt.Errorf("graphcache: %x identity does not match the catalog", id)
		}
		if err := sameValues(n.Values, truth[string(id)]); err != nil {
			return fmt.Errorf("graphcache: %x values: %w", id, err)
		}
		if err := sameMembers(n.Members, members[string(id)]); err != nil {
			return fmt.Errorf("graphcache: %x members: %w", id, err)
		}
		for _, l := range n.Links {
			assocIDs = append(assocIDs, l.Association)
		}
	}
	if len(assocIDs) == 0 {
		return nil
	}
	assocTruth, err := truthFn(g.db, uniqueIDs(assocIDs))
	if err != nil {
		return err
	}
	for _, id := range ids {
		n := g.nodes[string(id)]
		if n == nil {
			continue
		}
		for _, l := range n.Links {
			if !linkEndpoints(id, l, assocTruth[string(l.Association)]) {
				return fmt.Errorf("graphcache: %x link to %x does not match reconciled endpoints", id, l.Neighbor)
			}
		}
	}
	return nil
}

func (g *Graph) verifyKind(typeID []byte) error {
	want, err := g.unmergedIDs(typeID)
	if err != nil {
		return err
	}
	got := g.kinds[string(typeID)]
	if len(want) != len(got) {
		return fmt.Errorf("graphcache: kind index has %d ids, catalog has %d", len(got), len(want))
	}
	set := map[string]bool{}
	for _, id := range got {
		set[string(id)] = true
		if _, ok := g.nodes[string(id)]; !ok {
			return fmt.Errorf("graphcache: kind index lists %x but it is not loaded", id)
		}
	}
	for _, id := range want {
		if !set[string(id)] {
			return fmt.Errorf("graphcache: kind index is missing %x", id)
		}
	}
	return nil
}

func sameValues(got map[string][]Value, want []TruthRow) error {
	byProp := map[string][]TruthRow{}
	for _, row := range want {
		byProp[string(row.PropertyID)] = append(byProp[string(row.PropertyID)], row)
	}
	if len(byProp) != len(got) {
		return fmt.Errorf("property count %d, truth %d", len(got), len(byProp))
	}
	for prop, rows := range got {
		truth := byProp[prop]
		sort.Slice(truth, func(i, j int) bool { return truth[i].Rank < truth[j].Rank })
		if len(rows) != len(truth) {
			return fmt.Errorf("property %x rank count %d, truth %d", prop, len(rows), len(truth))
		}
		for i := range rows {
			if !valueEqual(rows[i], truth[i]) {
				return fmt.Errorf("property %x rank %d differs", prop, rows[i].Rank)
			}
		}
	}
	return nil
}

func valueEqual(v Value, t TruthRow) bool {
	return v.Rank == t.Rank && v.Reason == t.Reason &&
		v.HasText == t.HasText && v.Text == t.Text &&
		v.HasInteger == t.HasInteger && v.Integer == t.Integer &&
		bytes.Equal(v.TermID, t.TermID) &&
		bytes.Equal(v.EntityID, t.EntityID) &&
		bytes.Equal(v.Date, t.Date) &&
		bytes.Equal(v.Name, t.Name)
}

func sameMembers(got, want []Member) error {
	if len(got) != len(want) {
		return fmt.Errorf("count %d, catalog %d", len(got), len(want))
	}
	for i := range got {
		if !sameID(got[i].SubjectID, want[i].SubjectID) || got[i].Accepted != want[i].Accepted {
			return fmt.Errorf("claim %d differs", i)
		}
	}
	return nil
}

func linkEndpoints(self []byte, l Link, rows []TruthRow) bool {
	var ids [][]byte
	for _, row := range rows {
		if len(row.EntityID) == 16 {
			ids = append(ids, row.EntityID)
		}
	}
	return hasID(ids, self) && hasID(ids, l.Neighbor)
}

func hasID(ids [][]byte, id []byte) bool {
	for _, cur := range ids {
		if sameID(cur, id) {
			return true
		}
	}
	return false
}
