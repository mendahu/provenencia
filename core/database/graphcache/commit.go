package graphcache

import (
	"os"

	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database/effects"
)

// OnCommit applies a committed write. Structure drops the store. A handle
// already loaded is replaced. A handle whose kind index is loaded is inserted
// when it is new. Associations in the set drop their holders, then every
// endpoint that was loaded or that the committed rows name is reloaded once.
// Finished rows for those handles, and for whatever reads them, are dropped
// after the reload. A read does not consult the audit log.
func (g *Graph) OnCommit(rev int64, set effects.Set) error {
	if g == nil {
		return nil
	}
	if set.Vocabulary {
		g.DropLabels()
		g.vocabOK = false
		g.vocab = Vocabulary{}
	}
	if set.Structure {
		g.Drop()
		return g.verifyIfSet()
	}
	ids := uniqueIDs(set.Handles)
	if len(ids) == 0 {
		return g.verifyIfSet()
	}
	g.statsOK = false
	g.stats = nil
	rows, err := g.entityRows(ids)
	if err != nil {
		return err
	}
	var entities, assocs []entityRow
	var reload [][]byte
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
			g.removeNode(id)
			continue
		}
		if _, bridge := connectrules.LookupBridge(row.typeKey); bridge {
			assocs = append(assocs, row)
			continue
		}
		entities = append(entities, row)
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
	if err := g.loadIDs(reload, true); err != nil {
		return err
	}
	if err := g.forgetDisplays(uniqueIDs(roots)); err != nil {
		return err
	}
	return g.verifyIfSet()
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

func (g *Graph) verifyIfSet() error {
	if os.Getenv("PROVENENCIA_GRAPH_VERIFY") != "1" {
		return nil
	}
	return g.Verify()
}
