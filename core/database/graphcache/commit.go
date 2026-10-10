package graphcache

import (
	"os"

	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database/effects"
)

// OnCommit applies a committed write. Structure drops the store. A handle
// already loaded is replaced. A handle whose kind index is loaded is inserted
// when it is new. An association drops its holders and loads the endpoints
// the committed rows name, which is the one insert that does not need a prior
// read of those endpoints. A read does not consult the audit log.
func (g *Graph) OnCommit(rev int64, set effects.Set) error {
	if g == nil {
		return nil
	}
	if set.Vocabulary {
		g.DropLabels()
	}
	if set.Structure {
		g.Drop()
		return g.verifyIfSet()
	}
	g.dropSources(set.Sources)
	g.dropSources(set.CacheSources)
	ids := uniqueIDs(set.Handles)
	if len(ids) == 0 {
		return g.verifyIfSet()
	}
	rows, err := g.entityRows(ids)
	if err != nil {
		return err
	}
	var entities, assocs []entityRow
	roots := append([][]byte(nil), ids...)
	for _, id := range ids {
		row, ok := rows[string(id)]
		if !ok {
			ends, err := g.onGone(id)
			if err != nil {
				return err
			}
			roots = append(roots, ends...)
			continue
		}
		if _, bridge := connectrules.LookupBridge(row.typeKey); bridge {
			assocs = append(assocs, row)
			continue
		}
		entities = append(entities, row)
	}
	for _, row := range entities {
		if err := g.onEntity(row); err != nil {
			return err
		}
	}
	for _, row := range assocs {
		ends, err := g.onAssociation(row.id)
		if err != nil {
			return err
		}
		roots = append(roots, ends...)
	}
	if err := g.forgetDisplays(uniqueIDs(roots)); err != nil {
		return err
	}
	return g.verifyIfSet()
}

func (g *Graph) onGone(id []byte) ([][]byte, error) {
	if _, ok := g.holders[string(id)]; ok {
		return g.onAssociation(id)
	}
	g.removeNode(id)
	return nil, nil
}

func (g *Graph) onEntity(row entityRow) error {
	_, loaded := g.nodes[string(row.id)]
	if !loaded && !g.kindOn[string(row.typeID)] {
		return nil
	}
	return g.loadIDs([][]byte{row.id}, true)
}

func (g *Graph) onAssociation(id []byte) ([][]byte, error) {
	old := append([][]byte(nil), g.holders[string(id)]...)
	delete(g.holders, string(id))
	g.removeNode(id)
	var reload [][]byte
	for _, eid := range old {
		if _, ok := g.nodes[string(eid)]; ok {
			reload = append(reload, eid)
		}
	}
	if err := g.loadIDs(reload, true); err != nil {
		return nil, err
	}
	fresh, err := g.endpointIDs(id)
	if err != nil {
		return nil, err
	}
	// Replace endpoints that are already loaded. A later association would
	// otherwise leave the first snapshot of their links in place.
	if err := g.loadIDs(fresh, true); err != nil {
		return nil, err
	}
	return append(old, fresh...), nil
}

func (g *Graph) verifyIfSet() error {
	if os.Getenv("PROVENENCIA_GRAPH_VERIFY") != "1" {
		return nil
	}
	return g.Verify()
}
