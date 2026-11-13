package graphcache

import (
	"fmt"

	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/hops"
)

// Follow returns this node's links that match hop. The node is loaded on
// first read. Link order is the node's order: association ref, then neighbor ref.
func (g *Graph) Follow(id []byte, hop hops.Hop) ([]Link, error) {
	out, err := g.follow(id, hop)
	if err == nil && g != nil && g.inputRecorder != nil {
		for _, k := range FollowKeys(id, hop, out) {
			g.inputRecorder(k)
		}
	}
	return out, err
}

// follow is Follow without the test recorder; it reads the node through node,
// so it reports no node values.
func (g *Graph) follow(id []byte, hop hops.Hop) ([]Link, error) {
	n, err := g.node(id)
	if err != nil || n == nil {
		return nil, err
	}
	return n.Follow(hop)
}

// Follow returns the node's links that match hop, in the node's order.
func (n *Node) Follow(hop hops.Hop) ([]Link, error) {
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
