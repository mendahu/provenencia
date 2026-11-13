package graphcache

import (
	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/sqlbatch"
)

// OnCommit applies a committed write. Structure drops the store. A handle
// already loaded is replaced. A handle whose kind index is loaded is inserted
// when it is new. Associations in the set drop their holders, then every
// endpoint that was loaded or that the committed rows name is reloaded once.
// Finished rows for those handles are dropped after the reload, and each
// registered cache entry that holds data then applies the commit (cache.go). Loaded last-activity maps move forward to the commit.
// A read does not consult the audit log beyond the first Activity load.
// A deleted handle's neighbors become roots; its writer commits the
// associations it removed. A merged handle's loaded neighbors are reloaded by
// the graph, so none keeps a link a fresh load would drop.
func (g *Graph) OnCommit(c effects.Commit) error {
	if g == nil {
		return nil
	}
	g.applyActivity(c)
	set := c.Effects
	if set.Has(effects.ListPropertyVocabulary) {
		g.DropLabels()
		g.vocabOK = false
		g.vocab = Vocabulary{}
		g.dropReaders(vocabularySource)
	}
	if set.Structure {
		g.Drop()
		return g.verifyIfSet()
	}
	ids := uniqueIDs(set.Handles)
	if len(ids) == 0 {
		if err := g.applyEntries(nil, set); err != nil {
			return err
		}
		return g.verifyIfSet()
	}
	rows, err := g.entityRows(ids)
	if err != nil {
		return err
	}
	var entities, assocs []entityRow
	var reload, merged [][]byte
	roots := append([][]byte(nil), ids...)
	for _, id := range ids {
		row, ok := rows[string(id)]
		if !ok {
			if _, held := g.holders[string(id)]; held {
				ends, touched, err := g.detachAssociation(id)
				if err != nil {
					return err
				}
				reload = append(reload, ends...)
				roots = append(roots, touched...)
				continue
			}
			// The node's neighbors read it, and the stored rows that name them
			// are dropped from the roots.
			if n, cached := g.nodes[string(id)]; cached {
				for _, l := range n.Links {
					roots = append(roots, l.Neighbor)
				}
			}
			g.removeNode(id)
			continue
		}
		if _, bridge := connectrules.LookupBridge(row.typeKey); bridge {
			assocs = append(assocs, row)
			continue
		}
		entities = append(entities, row)
		if row.merged {
			merged = append(merged, row.id)
		}
	}
	for _, row := range entities {
		if g.entityNeedsReload(row) {
			reload = append(reload, row.id)
		}
	}
	for _, row := range assocs {
		ends, touched, err := g.detachAssociation(row.id)
		if err != nil {
			return err
		}
		reload = append(reload, ends...)
		roots = append(roots, touched...)
	}
	neighbors, touched, err := g.mergedNeighbors(merged)
	if err != nil {
		return err
	}
	reload = append(reload, neighbors...)
	roots = append(roots, touched...)
	if err := g.loadIDs(reload, true); err != nil {
		return err
	}
	roots = uniqueIDs(roots)
	g.dropCommittedDisplays(roots, set)
	if err := g.applyEntries(roots, set); err != nil {
		return err
	}
	return g.verifyIfSet()
}

// dropCommittedDisplays forgets stored rows for roots and for the rows the
// write's input-index lookup found (set.SummaryDeps). An empty display map has
// nothing to forget.
func (g *Graph) dropCommittedDisplays(roots [][]byte, set effects.Set) {
	if len(g.displays) == 0 {
		return
	}
	g.dropDisplays(uniqueIDs(append(append([][]byte(nil), roots...), set.SummaryDeps...)))
}

func (g *Graph) entityNeedsReload(row entityRow) bool {
	_, loaded := g.nodes[string(row.id)]
	return loaded || g.kindOn[string(row.typeID)]
}

// detachAssociation forgets the association. reload is what must be loaded
// again: holders already in the node map, plus the endpoints the committed
// rows name. touched is every holder and endpoint, which is what a stored
// row may still name.
func (g *Graph) detachAssociation(id []byte) (reload, touched [][]byte, err error) {
	old := append([][]byte(nil), g.holders[string(id)]...)
	delete(g.holders, string(id))
	g.removeNode(id)
	for _, eid := range old {
		if _, ok := g.nodes[string(eid)]; ok {
			reload = append(reload, eid)
		}
	}
	fresh, err := g.endpointIDs(id)
	if err != nil {
		return nil, nil, err
	}
	return append(reload, fresh...), append(old, fresh...), nil
}

// mergedNeighbors finds the unmerged nodes each merged handle's own links
// name. reload is the ones already loaded, whose links to the handle a fresh
// load would drop; touched is all of them, which a stored row may still name.
func (g *Graph) mergedNeighbors(ids [][]byte) (reload, touched [][]byte, err error) {
	if len(ids) == 0 {
		return nil, nil, nil
	}
	err = sqlbatch.Each(ids, func(chunk [][]byte) error {
		raws, err := g.linkRaws(chunk)
		if err != nil {
			return err
		}
		for _, r := range raws {
			touched = append(touched, r.neighbor)
			if _, ok := g.nodes[string(r.neighbor)]; ok {
				reload = append(reload, r.neighbor)
			}
		}
		return nil
	})
	return reload, touched, err
}

func (g *Graph) verifyIfSet() error {
	if !VerifyMode() {
		return nil
	}
	return g.Verify()
}
