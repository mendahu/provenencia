package conclusionsummaries

import (
	"fmt"
	"reflect"

	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/hops"
	"github.com/mendahu/provenencia/core/valuecompare"
)

// maxPlaceEntries caps the place entries one graph holds. It matches the other
// per-handle entries (canonlinks.maxKeys) and holds every place of a real
// catalog; only a crafted catalog reaches it.
const maxPlaceEntries = 65536

// placeEntry is what a chain read needs of one place: its shell, its period,
// and its filed part_of and succeeded_by links in both directions with their
// decoded dates. The fill reads only the place's values and its
// place_relationship links, never the events and other links a city carries.
// Values are shared and read only.
type placeEntry struct {
	found    bool
	shell    PlaceSummary
	period   valuecompare.Window
	partFrom []placeLink // this place is the part: part id → wholes
	partTo   []placeLink // this place is the whole: whole id → parts
	succFrom []placeLink // this place is the predecessor
	succTo   []placeLink // this place is the successor
	// inputs are the keys a summary row records for this place: its values, its
	// part_of links at the part end and those links' associations.
	inputs []graphcache.InputKey
}

// placeEntries holds one placeEntry per place id. A commit drops an entry
// through its Inputs, so a write that touches neither a place nor its links
// leaves it.
var placeEntries graphcache.Keyed[placeEntry]

func init() {
	placeEntries = graphcache.RegisterKeyed(graphcache.KeyedEntry[placeEntry]{
		Name:        "conclusionsummaries.places",
		DerivedFrom: []graphcache.Source{graphcache.Vocabulary{}},
		Inputs:      placeInputs,
		Check:       checkPlaceEntry,
		Cap:         maxPlaceEntries,
	})
}

// placeEntriesOf returns the entry of each place in ids, keyed by raw id,
// filling those not held in one fill a frontier at a time. An id that is not 16
// bytes has none.
func placeEntriesOf(g *graphcache.Graph, v vocab, ids [][]byte) (map[string]placeEntry, error) {
	return placeEntries.GetMany(g, ids, func(missing [][]byte) (map[string]placeEntry, error) {
		return fillPlaceEntries(g, v, missing)
	})
}

// fillPlaceEntries reads the entries of ids with a narrow load: the places'
// values and place_relationship links, then every link's association values.
// A place with no entity row gets a found-false entry whose only input is its
// own values, so a commit that creates it drops the entry.
func fillPlaceEntries(g *graphcache.Graph, v vocab, ids [][]byte) (map[string]placeEntry, error) {
	nodes, err := g.NodesOnBridges(ids, "place_relationship")
	if err != nil {
		return nil, err
	}
	type found struct {
		n     *graphcache.Node
		links [4][]graphcache.Link
	}
	byHop := [4]hops.Hop{hops.ParentsOfPlace, hops.PartsOfPlace, hops.SuccessorsOfPlace, hops.PredecessorsOfPlace}
	have := map[string]*found{}
	var assocIDs [][]byte
	for _, id := range ids {
		n := nodes[string(id)]
		if n == nil {
			continue
		}
		f := &found{n: n}
		for i, hop := range byHop {
			if f.links[i], err = n.Follow(hop); err != nil {
				return nil, err
			}
			for _, l := range f.links[i] {
				assocIDs = append(assocIDs, l.Association)
			}
		}
		have[string(id)] = f
	}
	assocs, err := g.NodesOnBridges(assocIDs)
	if err != nil {
		return nil, err
	}
	out := make(map[string]placeEntry, len(ids))
	for _, id := range ids {
		id = append([]byte(nil), id...)
		f := have[string(id)]
		if f == nil {
			out[string(id)] = placeEntry{inputs: []graphcache.InputKey{graphcache.NodeKey(id)}}
			continue
		}
		shell, err := placeShell(f.n, v)
		if err != nil {
			return nil, err
		}
		e := placeEntry{found: true, shell: shell, period: periodWindow(shell.StartDate, shell.EndDate)}
		part := map[string]bool{}
		succ := map[string]bool{}
		for i, links := range f.links {
			for _, l := range links {
				pl, err := orientPlaceLink(v, id, l, assocs)
				if err != nil {
					return nil, err
				}
				key := string(pl.from) + "\x00" + string(pl.to) + "\x00" + string(pl.association)
				if i < 2 {
					if part[key] {
						continue
					}
					part[key] = true
					e.partFrom, e.partTo = addPlaceLink(e.partFrom, e.partTo, id, pl)
					continue
				}
				if succ[key] {
					continue
				}
				succ[key] = true
				e.succFrom, e.succTo = addPlaceLink(e.succFrom, e.succTo, id, pl)
			}
		}
		keys := append([]graphcache.InputKey{graphcache.NodeKey(id)}, graphcache.FollowKeys(id, hops.ParentsOfPlace, f.links[0])...)
		e.inputs = sortedKeys(keys)
		out[string(id)] = e
	}
	return out, nil
}

// orientPlaceLink makes the link l of place id a placeLink, from its part or
// predecessor to its whole or successor, with the dates of its association.
func orientPlaceLink(v vocab, id []byte, l graphcache.Link, assocs map[string]*graphcache.Node) (placeLink, error) {
	from, to := id, l.Neighbor
	if !l.FromEnd {
		from, to = l.Neighbor, id
	}
	assoc := assocs[string(l.Association)]
	if assoc == nil {
		return placeLink{}, fmt.Errorf("conclusionsummaries: association %x is missing", l.Association)
	}
	d, err := assocDatesOf(v, assoc)
	if err != nil {
		return placeLink{}, err
	}
	return placeLink{
		from: append([]byte(nil), from...), to: append([]byte(nil), to...),
		association: append([]byte(nil), l.Association...), start: d.start, end: d.end,
	}, nil
}

// addPlaceLink files pl in the list of the place it leaves and the list of the
// place it enters, as id is either end: a self-loop sits in both.
func addPlaceLink(fromList, toList []placeLink, id []byte, pl placeLink) (_, _ []placeLink) {
	if string(pl.from) == string(id) {
		fromList = append(fromList, pl)
	}
	if string(pl.to) == string(id) {
		toList = append(toList, pl)
	}
	return fromList, toList
}

func sortedKeys(keys []graphcache.InputKey) []graphcache.InputKey {
	seen := map[string]bool{}
	out := keys[:0]
	for _, k := range keys {
		if mk := k.Key(); !seen[mk] {
			seen[mk] = true
			out = append(out, k)
		}
	}
	graphcache.SortInputKeys(out)
	return out
}

// placeInputs is everything the entry for id was built from: the keys a summary
// row records (its values, its part_of links at the part end and their
// associations) and the same for the links at the other three ends. A commit
// that moves one drops the entry.
func placeInputs(id []byte, e placeEntry) []graphcache.InputKey {
	keys := []graphcache.InputKey{graphcache.NodeKey(id)}
	if !e.found {
		return keys
	}
	keys = append(keys, e.inputs...)
	for _, set := range []struct {
		hop   hops.Hop
		links []placeLink
	}{
		{hops.PartsOfPlace, e.partTo}, {hops.SuccessorsOfPlace, e.succFrom}, {hops.PredecessorsOfPlace, e.succTo},
	} {
		links := make([]graphcache.Link, len(set.links))
		for i, l := range set.links {
			links[i] = graphcache.Link{Association: l.association}
		}
		keys = append(keys, graphcache.FollowKeys(id, set.hop, links)...)
	}
	return sortedKeys(keys)
}

// checkPlaceEntry rebuilds the entry of id from the committed rows and compares.
func checkPlaceEntry(g *graphcache.Graph, id []byte, stored placeEntry) error {
	v, err := vocabulary(g)
	if err != nil {
		return err
	}
	fresh, err := fillPlaceEntries(g, v, [][]byte{id})
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(fresh[string(id)], stored) {
		return fmt.Errorf("conclusionsummaries: stored place entry %x does not match a fresh load", id)
	}
	return nil
}
