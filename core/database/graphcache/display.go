package graphcache

// DepsFunc lists handles whose finished row reads one of ids. The header
// registry registers it. The graph does not import that package.
type DepsFunc func(q Querier, ids [][]byte) ([][]byte, error)

// DisplayCheck rebuilds one stored row and compares. conclusionheaders registers it.
type DisplayCheck func(g *Graph, id []byte, stored any) error

var (
	depsFn    DepsFunc
	displayFn DisplayCheck
)

// SetDisplayDeps registers the reverse of the header reads.
func SetDisplayDeps(fn DepsFunc) { depsFn = fn }

// SetDisplayCheck registers the stored-row comparison Verify runs.
func SetDisplayCheck(fn DisplayCheck) { displayFn = fn }

// Display returns the stored row for id.
func (g *Graph) Display(id []byte) (any, bool) {
	if g == nil || len(id) != 16 {
		return nil, false
	}
	v, ok := g.displays[string(id)]
	return v, ok
}

// SetDisplay stores a finished row.
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

// forgetDisplays drops stored rows for roots and for every handle the header
// rules say reads them. A save calls this after the nodes have been reloaded.
func (g *Graph) forgetDisplays(roots [][]byte) error {
	g.dropDisplays(roots)
	if depsFn == nil || len(roots) == 0 {
		return nil
	}
	deps, err := depsFn(g, roots)
	if err != nil {
		return err
	}
	g.dropDisplays(deps)
	return nil
}
