package graphcache

var displayFn MemoCheck

// SetDisplayCheck registers the stored-row comparison Verify runs.
func SetDisplayCheck(fn MemoCheck) { displayFn = fn }

// Display returns the stored row for id.
func (g *Graph) Display(id []byte) (any, bool) {
	if g == nil || len(id) != 16 {
		return nil, false
	}
	v, ok := g.displays[string(id)]
	return v, ok
}

// SetDisplay stores a composed summary row. It is not finished: its owner
// fills read-time labels when it reads the row.
func (g *Graph) SetDisplay(id []byte, v any) {
	if g == nil || len(id) != 16 || v == nil {
		return
	}
	g.displays[string(id)] = v
}

// Labels returns the vocabulary map stored for this catalog.
func (g *Graph) Labels() (any, bool) {
	if g == nil || !g.labelsOK {
		return nil, false
	}
	return g.labels, true
}

// SetLabels stores the vocabulary map.
func (g *Graph) SetLabels(v any) {
	if g == nil {
		return
	}
	g.labels = v
	g.labelsOK = true
}

// DropLabels forgets the vocabulary map. The next read loads it again.
func (g *Graph) DropLabels() {
	if g == nil {
		return
	}
	g.labelsOK = false
	g.labels = nil
}

func (g *Graph) dropDisplay(id []byte) {
	if g == nil || len(id) != 16 {
		return
	}
	delete(g.displays, string(id))
}

func (g *Graph) dropDisplays(ids [][]byte) {
	for _, id := range ids {
		g.dropDisplay(id)
	}
}
