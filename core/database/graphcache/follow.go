package graphcache

import (
	"fmt"

	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/hops"
)

// Follow returns this node's links that match hop. The node is loaded on
// first read. Link order is the node's order: association ref, then neighbor ref.
func (g *Graph) Follow(id []byte, hop hops.Hop) ([]Link, error) {
	n, err := g.Node(id)
	if err != nil || n == nil {
		return nil, err
	}
	b, ok := connectrules.LookupBridge(hop.Bridge())
	if !ok || len(b.Endpoints) == 0 {
		return nil, fmt.Errorf("graphcache: hop bridge %q is not registered", hop.Bridge())
	}
	fromEnd := hop.From() == b.Endpoints[0].PropertyKey
	_, term, filtered := hop.Filter()
	var out []Link
	for _, l := range n.Links {
		if l.BridgeTypeKey != hop.Bridge() || l.FromEnd != fromEnd {
			continue
		}
		if filtered && l.TermKey != term {
			continue
		}
		out = append(out, l)
	}
	return out, nil
}
