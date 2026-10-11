package conclusionheaders

import "github.com/mendahu/provenencia/core/database"

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

// SortPlaces applies the list order to Place rows.
func SortPlaces(rows []PlaceHeader) {
	sortByTitle(rows, placeTitle, func(h PlaceHeader) string { return h.Entity.Ref })
}
