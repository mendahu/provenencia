package graphcache

import (
	"bytes"
	"fmt"
	"os"
	"slices"
	"sort"

	"github.com/mendahu/provenencia/core/propertymeta"
)

// VerifyMode reports whether PROVENENCIA_GRAPH_VERIFY=1, which rebuilds and checks the cache after every commit.
func VerifyMode() bool {
	return os.Getenv("PROVENENCIA_GRAPH_VERIFY") == "1"
}

// TruthRow is one value a fresh reconcile would write for an entity.
type TruthRow struct {
	PropertyID []byte
	Rank       int
	Reason     string
	propertymeta.Scalar
	TermID      []byte
	EntityID    []byte
	Date        []byte
	Name        []byte
	ImageFileID []byte
	Support     int
	Against     int
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
	if displayFn != nil {
		for key, stored := range g.displays {
			if err := displayFn(g, []byte(key), stored); err != nil {
				return err
			}
		}
	}
	if err := g.verifyEntries(); err != nil {
		return err
	}
	if evidenceGraphFn != nil {
		for key, e := range g.evidenceGraphs {
			if err := evidenceGraphFn(g, []byte(key), e.graph); err != nil {
				return err
			}
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
	outcomes, err := g.loadOutcomes(ids)
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
		if n.Ref != row.ref || n.Argument != row.argument || n.Label != row.label ||
			!sameID(n.SubjectTypeID, row.typeID) || n.Merged != row.merged ||
			!sameID(n.PreferredLikeness, row.likeness) {
			return fmt.Errorf("graphcache: %x identity does not match the catalog", id)
		}
		if err := sameValues(n.Values, truth[string(id)]); err != nil {
			return fmt.Errorf("graphcache: %x values: %w", id, err)
		}
		if err := sameMembers(n.Members, members[string(id)]); err != nil {
			return fmt.Errorf("graphcache: %x members: %w", id, err)
		}
		if err := sameOutcomes(n.Outcomes, outcomes[string(id)]); err != nil {
			return fmt.Errorf("graphcache: %x outcomes: %w", id, err)
		}
		for _, l := range n.Links {
			assocIDs = append(assocIDs, l.Association)
		}
	}
	fresh, err := g.loadLinks(ids)
	if err != nil {
		return err
	}
	for _, id := range ids {
		n := g.nodes[string(id)]
		if n == nil {
			continue
		}
		if err := sameLinks(n.Links, fresh[string(id)]); err != nil {
			return fmt.Errorf("graphcache: %x links: %w", id, err)
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

func sameLinks(got, want []Link) error {
	gk, wk := linkKeys(got), linkKeys(want)
	sort.Strings(gk)
	sort.Strings(wk)
	if len(gk) != len(wk) {
		return fmt.Errorf("count %d, catalog %d", len(gk), len(wk))
	}
	for i := range gk {
		if gk[i] != wk[i] {
			return fmt.Errorf("link %d differs", i)
		}
	}
	return nil
}

func linkKeys(links []Link) []string {
	out := make([]string, len(links))
	for i, l := range links {
		from := "0"
		if l.FromEnd {
			from = "1"
		}
		out[i] = string(l.Association) + "|" + string(l.Neighbor) + "|" + from + "|" + l.BridgeTypeKey + "|" + l.TermKey
	}
	return out
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
	return walkValues(got, want, valueEqual)
}

// SameDisplayRows reports whether the stored values show what the written rows
// show, per Property in rank order. Vote counts are ignored.
func SameDisplayRows(stored map[string][]Value, written []TruthRow) bool {
	return walkValues(stored, written, func(v Value, t TruthRow) bool {
		return SameDisplay(rowOf(v, t.PropertyID), t)
	}) == nil
}

// walkValues compares stored values to truth rows: the Property count, each
// Property's row count in rank order, then each pair with same.
func walkValues(got map[string][]Value, want []TruthRow, same func(Value, TruthRow) bool) error {
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
			if !same(rows[i], truth[i]) {
				return fmt.Errorf("property %x rank %d differs", prop, rows[i].Rank)
			}
		}
	}
	return nil
}

// valueEqual is v equal to t in every field, the vote counts included.
func valueEqual(v Value, t TruthRow) bool {
	return SameDisplay(rowOf(v, t.PropertyID), t) && v.Support == t.Support && v.Against == t.Against
}

// SameDisplay reports whether two rows show the same value: every field but
// the vote counts (Support, Against), which a join moves without changing what
// any reader displays.
func SameDisplay(a, b TruthRow) bool {
	return bytes.Equal(a.PropertyID, b.PropertyID) && a.Rank == b.Rank && a.Reason == b.Reason &&
		a.Scalar == b.Scalar &&
		bytes.Equal(a.TermID, b.TermID) &&
		bytes.Equal(a.EntityID, b.EntityID) &&
		bytes.Equal(a.Date, b.Date) &&
		bytes.Equal(a.Name, b.Name) &&
		bytes.Equal(a.ImageFileID, b.ImageFileID)
}

// rowOf is the stored value v of the Property as a TruthRow.
func rowOf(v Value, property []byte) TruthRow {
	return TruthRow{
		PropertyID: property, Rank: v.Rank, Reason: v.Reason,
		Scalar: v.Scalar,
		TermID: v.TermID, EntityID: v.EntityID, Date: v.Date, Name: v.Name, ImageFileID: v.ImageFileID,
		Support: v.Support, Against: v.Against,
	}
}

func sameMembers(got, want []Member) error {
	if len(got) != len(want) {
		return fmt.Errorf("count %d, catalog %d", len(got), len(want))
	}
	for i := range got {
		if !sameID(got[i].SubjectID, want[i].SubjectID) || !sameID(got[i].SourceID, want[i].SourceID) || !sameID(got[i].GraphID, want[i].GraphID) || got[i].Accepted != want[i].Accepted {
			return fmt.Errorf("claim %d differs", i)
		}
	}
	return nil
}

// sameOutcomes compares the node's outcome rows to the committed rows: the
// same Properties, and per Property the same rows in order.
func sameOutcomes(got, want map[string][]Outcome) error {
	if len(got) != len(want) {
		return fmt.Errorf("property count %d, catalog %d", len(got), len(want))
	}
	for prop, rows := range got {
		truth := want[prop]
		if len(rows) != len(truth) {
			return fmt.Errorf("property %x row count %d, catalog %d", prop, len(rows), len(truth))
		}
		for i, o := range rows {
			w := truth[i]
			if !sameID(o.ObservationID, w.ObservationID) || o.Reason != w.Reason || o.ValueRank != w.ValueRank ||
				!sameID(o.DeniedBy, w.DeniedBy) || o.VoteSupport != w.VoteSupport || o.VoteOf != w.VoteOf {
				return fmt.Errorf("property %x row %d differs", prop, i)
			}
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
	has := func(id []byte) bool {
		return slices.ContainsFunc(ids, func(cur []byte) bool { return sameID(cur, id) })
	}
	return has(self) && has(l.Neighbor)
}
