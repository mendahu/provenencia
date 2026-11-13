package conclusionsummaries

import (
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/graphcache"
)

func countList(c *database.Catalog, list func(*database.Catalog) error) (int, error) {
	n := 0
	c.Graph().SetQueryCounterForTest(func() { n++ })
	defer c.Graph().SetQueryCounterForTest(nil)
	err := list(c)
	return n, err
}

// ListPersonsQueryCount lists Persons and reports how many graph queries it took.
func ListPersonsQueryCount(c *database.Catalog) (int, error) {
	return countList(c, func(c *database.Catalog) error {
		_, err := ListPersons(c)
		return err
	})
}

// ListEventsQueryCount lists Events and reports how many graph queries it took.
func ListEventsQueryCount(c *database.Catalog) (int, error) {
	return countList(c, func(c *database.Catalog) error {
		_, err := ListEvents(c)
		return err
	})
}

// ListPlacesQueryCount lists Places and reports how many graph queries it took.
func ListPlacesQueryCount(c *database.Catalog) (int, error) {
	return countList(c, func(c *database.Catalog) error {
		_, err := ListPlaces(c)
		return err
	})
}

// SortPersons and SortPlaces apply the list orders to rows.
var (
	SortPersons = sortPersons
	SortPlaces  = sortPlaces
)

// SetInputKeyCapForTest lowers the per-call input key cap and returns the
// function that restores it.
func SetInputKeyCapForTest(n int) (restore func()) {
	old := inputKeyCap
	inputKeyCap = n
	return func() { inputKeyCap = old }
}

// HeldPlaceEntryForTest reports whether g holds an entry for the place.
func HeldPlaceEntryForTest(g *graphcache.Graph, id []byte) bool {
	_, ok := placeEntries.Peek(g, id)
	return ok
}

// PlaceEntryForTest returns whether the chain entry for the place is held and
// found, and the keys it is dropped by (the keys a summary row records for it).
func PlaceEntryForTest(g *graphcache.Graph, id []byte) (held, found bool, inputs []graphcache.InputKey) {
	e, ok := placeEntries.Peek(g, id)
	return ok, e.found, e.inputs
}

// PlacePageForTest returns whether the page entry for the place is held and the
// keys it is dropped by.
func PlacePageForTest(g *graphcache.Graph, id []byte) (held bool, inputs []graphcache.InputKey) {
	p, ok := placePages.Peek(g, id)
	return ok, p.inputs
}

// SetPlaceEntryCapForTest lowers the place entries g holds to n.
func SetPlaceEntryCapForTest(g *graphcache.Graph, n int) { placeEntries.SetCapForTest(g, n) }

// CorruptPlaceEntryForTest replaces the held entry's names so it no longer
// matches a fresh load.
func CorruptPlaceEntryForTest(g *graphcache.Graph, id []byte) {
	e, ok := placeEntries.Peek(g, id)
	if !ok {
		return
	}
	e.shell.Names = []string{"corrupt"}
	placeEntries.Set(g, id, e)
}

// FillPlaceEntriesForTest holds the entries of ids on g, as a chain read does.
func FillPlaceEntriesForTest(g *graphcache.Graph, ids ...[]byte) error {
	v, err := vocabulary(g)
	if err != nil {
		return err
	}
	_, err = placeEntriesOf(g, v, ids)
	return err
}

// FillPlacePagesForTest holds the pages of ids on g, as the place page does.
func FillPlacePagesForTest(g *graphcache.Graph, ids ...[]byte) error {
	v, err := vocabulary(g)
	if err != nil {
		return err
	}
	_, err = placePagesOf(g, v, ids)
	return err
}
