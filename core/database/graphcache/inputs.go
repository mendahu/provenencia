package graphcache

import (
	"bytes"
	"sort"

	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/hops"
)

// InputKey is one thing a row was composed from. Bridge "" is the handle ID's
// own values (an association's included). Otherwise Bridge is the bridge type
// key and Endpoint the endpoint Property key where ID sits: the links ID has at
// that end of that bridge. The key carries no bridge origin; only product
// bridges are recorded.
type InputKey struct {
	ID               []byte
	Bridge, Endpoint string
}

// Key is a string that names the key and nothing else, for a map or a set.
func (k InputKey) Key() string { return string(k.ID) + "\x00" + k.Bridge + "\x00" + k.Endpoint }

// Equal reports whether o is the same key.
func (k InputKey) Equal(o InputKey) bool {
	return bytes.Equal(k.ID, o.ID) && k.Bridge == o.Bridge && k.Endpoint == o.Endpoint
}

// NodeKey is the key a read of id's values records.
func NodeKey(id []byte) InputKey { return InputKey{ID: id} }

// FollowKeys are the keys a Follow of hop from id that returned links
// records: the bridge end, which covers the links a filter drops, and each
// returned association's values.
func FollowKeys(id []byte, hop hops.Hop, links []Link) []InputKey {
	keys := make([]InputKey, 0, 1+len(links))
	keys = append(keys, InputKey{ID: id, Bridge: hop.Bridge(), Endpoint: hop.From()})
	for _, l := range links {
		keys = append(keys, NodeKey(l.Association))
	}
	return keys
}

// InputScope reads the graph on behalf of one dependent row and records what
// it read. A composer reads other handles only through its scope, so a key a
// row holds was recorded by a call the row made.
type InputScope struct {
	g    *Graph
	self []byte
	off  bool
	keys map[string]InputKey
}

// Inputs returns a scope that records the keys read for dependent and leaves
// out dependent's own values. Inputs(nil) records every key.
func (g *Graph) Inputs(dependent []byte) *InputScope {
	if g.scopeCounter != nil {
		g.scopeCounter()
	}
	return &InputScope{g: g, self: cloneID(dependent), keys: map[string]InputKey{}}
}

// NoInputs returns a scope that reads and records nothing, for batch phases
// that only collect ids.
func (g *Graph) NoInputs() *InputScope {
	return &InputScope{g: g, off: true}
}

// Node is Graph.Node and records the node's values, a missing node included.
func (s *InputScope) Node(id []byte) (*Node, error) {
	n, err := s.g.Node(id)
	if err == nil {
		s.add(NodeKey(id))
	}
	return n, err
}

// Nodes is Graph.Nodes. It records nothing: it only loads.
func (s *InputScope) Nodes(ids [][]byte) error {
	return s.g.Nodes(ids)
}

// Follow is Graph.Follow and records the hop's bridge end at id and each
// returned association's values, not id's own values.
func (s *InputScope) Follow(id []byte, hop hops.Hop) ([]Link, error) {
	links, err := s.g.Follow(id, hop)
	if err == nil {
		s.add(FollowKeys(id, hop, links)...)
	}
	return links, err
}

// Add records keys, leaving out the dependent's own values. A graph with a test
// recorder reports each key to it first, so a key handed in from a registry
// entry is heard as a read is.
func (s *InputScope) Add(keys ...InputKey) {
	if s != nil && s.g != nil && s.g.inputRecorder != nil {
		for _, k := range keys {
			s.g.inputRecorder(k)
		}
	}
	s.add(keys...)
}

// add records keys without reporting them: Node and Follow read through the
// graph, which reports.
func (s *InputScope) add(keys ...InputKey) {
	if s == nil || s.off {
		return
	}
	for _, k := range keys {
		if s.self != nil && k.Bridge == "" && bytes.Equal(k.ID, s.self) {
			continue
		}
		mk := k.Key()
		if _, ok := s.keys[mk]; ok {
			continue
		}
		k.ID = cloneID(k.ID)
		s.keys[mk] = k
	}
}

// Keys returns the recorded keys once each, sorted by ID, Bridge, Endpoint.
func (s *InputScope) Keys() []InputKey {
	if s == nil {
		return nil
	}
	out := make([]InputKey, 0, len(s.keys))
	for _, k := range s.keys {
		out = append(out, k)
	}
	SortInputKeys(out)
	return out
}

// SortInputKeys orders keys by ID, Bridge, Endpoint.
func SortInputKeys(keys []InputKey) {
	sort.Slice(keys, func(i, j int) bool {
		if c := bytes.Compare(keys[i].ID, keys[j].ID); c != 0 {
			return c < 0
		}
		if keys[i].Bridge != keys[j].Bridge {
			return keys[i].Bridge < keys[j].Bridge
		}
		return keys[i].Endpoint < keys[j].Endpoint
	})
}

// AssociationInputs returns, for each id that is an unmerged association of a
// registered bridge, the keys of the handles at its ends: its rank-1 kept
// value of each endpoint Property, as {end, bridge, endpoint}. An id that is
// missing, merged or not a bridge has no entry. Keyed by raw id.
func (g *Graph) AssociationInputs(ids [][]byte) (map[string][]InputKey, error) {
	out := map[string][]InputKey{}
	if g == nil || len(ids) == 0 {
		return out, nil
	}
	v, err := g.Vocabulary()
	if err != nil {
		return nil, err
	}
	types := v.TypeKeysByID()
	for _, id := range ids {
		n, err := g.node(id)
		if err != nil {
			return nil, err
		}
		if n == nil || n.Merged {
			continue
		}
		b, ok := connectrules.LookupBridge(types[string(n.SubjectTypeID)])
		if !ok {
			continue
		}
		origin := b.OriginKey()
		var keys []InputKey
		for _, e := range b.Endpoints {
			propID := v.PropertyID(e.PropertyKey, origin)
			if propID == nil {
				continue
			}
			row, ok := Rank1(n.Kept(propID))
			if !ok || len(row.EntityID) == 0 {
				continue
			}
			keys = append(keys, InputKey{ID: cloneID(row.EntityID), Bridge: b.BridgeTypeKey, Endpoint: e.PropertyKey})
		}
		if len(keys) == 0 {
			continue
		}
		SortInputKeys(keys)
		out[string(id)] = keys
	}
	return out, nil
}

// SetInputRecorderForTest reports every key Node and Follow read on this graph
// and every key handed to InputScope.Add, with no self rule. Tests only.
func (g *Graph) SetInputRecorderForTest(fn func(InputKey)) {
	if g == nil {
		return
	}
	g.inputRecorder = fn
}

// SetScopeCounterForTest reports every recording scope Inputs creates on this
// graph. Tests only.
func (g *Graph) SetScopeCounterForTest(fn func()) {
	if g == nil {
		return
	}
	g.scopeCounter = fn
}
