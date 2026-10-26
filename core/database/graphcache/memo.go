package graphcache

// MemoCheck rebuilds one stored result and compares it. Verify runs it over
// every stored display row and every stored per-node memo.
type MemoCheck func(g *Graph, id []byte, stored any) error

// MemoDeps lists the handles whose stored result reads one of roots, beyond
// the roots themselves. OnCommit calls it after the reload, so g holds the
// committed nodes.
type MemoDeps func(g *Graph, roots [][]byte) ([][]byte, error)

// NodeMemo is one kind of result an owner stores per node. The owner says how
// far a commit reaches (Deps), how Verify rebuilds a result (Check) and
// whether a vocabulary commit drops it. The graph does not import the owner.
type NodeMemo struct {
	Name             string
	Check            MemoCheck // nil: Verify skips the slot
	Deps             MemoDeps  // nil: a commit drops only the roots
	DropOnVocabulary bool
}

// MemoSlot names one registered NodeMemo.
type MemoSlot int

var memoSlots []NodeMemo

// RegisterNodeMemo registers a slot. Owners call it from init, before any
// graph is made.
func RegisterNodeMemo(m NodeMemo) MemoSlot {
	memoSlots = append(memoSlots, m)
	return MemoSlot(len(memoSlots) - 1)
}

// Memo returns the result an owner stored for a node.
func (g *Graph) Memo(slot MemoSlot, id []byte) (any, bool) {
	if g == nil || len(id) != 16 {
		return nil, false
	}
	v, ok := g.memos[slot][string(id)]
	return v, ok
}

// SetMemo stores an owner's result for a node. A commit that touches the node,
// or a node the slot's Deps names for it, drops it. A first load does not.
func (g *Graph) SetMemo(slot MemoSlot, id []byte, v any) {
	if g == nil || len(id) != 16 || v == nil || int(slot) >= len(memoSlots) {
		return
	}
	if g.memos == nil {
		g.memos = map[MemoSlot]map[string]any{}
	}
	if g.memos[slot] == nil {
		g.memos[slot] = map[string]any{}
	}
	g.memos[slot][string(id)] = v
}

// Cached returns the node when it is loaded. It never loads.
func (g *Graph) Cached(id []byte) (*Node, bool) {
	if g == nil {
		return nil, false
	}
	n, ok := g.nodes[string(id)]
	return n, ok
}

// dropMemos forgets every slot's results for roots and for what the slot's
// Deps names for them. A slot with nothing stored is not asked.
func (g *Graph) dropMemos(roots [][]byte) error {
	for slot, entries := range g.memos {
		if len(entries) == 0 {
			continue
		}
		for _, id := range roots {
			delete(entries, string(id))
		}
		m := memoSlots[slot]
		if m.Deps == nil {
			continue
		}
		deps, err := m.Deps(g, roots)
		if err != nil {
			return err
		}
		for _, id := range deps {
			delete(entries, string(id))
		}
	}
	return nil
}

// dropVocabularyMemos empties the slots that read the vocabulary.
func (g *Graph) dropVocabularyMemos() {
	for slot := range g.memos {
		if memoSlots[slot].DropOnVocabulary {
			delete(g.memos, slot)
		}
	}
}

// verifyMemos runs each slot's Check over its stored results.
func (g *Graph) verifyMemos() error {
	for slot, entries := range g.memos {
		check := memoSlots[slot].Check
		if check == nil {
			continue
		}
		for key, stored := range entries {
			if err := check(g, []byte(key), stored); err != nil {
				return err
			}
		}
	}
	return nil
}
