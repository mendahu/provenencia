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
	if set.Structure {
		g.Drop()
		return g.verifyIfSet()
	}
	ids := uniqueIDs(set.Handles)
	if len(ids) == 0 {
		return g.verifyIfSet()
	}
	rows, err := g.entityRows(ids)
	if err != nil {
		return err
	}
	var entities, assocs []entityRow
	for _, id := range ids {
		row, ok := rows[string(id)]
		if !ok {
			if err := g.onGone(id); err != nil {
				return err
			}
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
		if err := g.onAssociation(row.id); err != nil {
			return err
		}
	}
	return g.verifyIfSet()
}

func (g *Graph) onGone(id []byte) error {
	if _, ok := g.holders[string(id)]; ok {
		return g.onAssociation(id)
	}
	g.removeNode(id)
	return nil
}

func (g *Graph) onEntity(row entityRow) error {
	_, loaded := g.nodes[string(row.id)]
	if !loaded && !g.kindOn[string(row.typeID)] {
		return nil
	}
	return g.loadIDs([][]byte{row.id}, true)
}

func (g *Graph) onAssociation(id []byte) error {
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
		return err
	}
	fresh, err := g.endpointIDs(id)
	if err != nil {
		return err
	}
	return g.loadIDs(fresh, false)
}

func (g *Graph) verifyIfSet() error {
	if os.Getenv("PROVENENCIA_GRAPH_VERIFY") != "1" {
		return nil
	}
	return g.Verify()
}
