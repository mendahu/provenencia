package conclusionheaders

import (
	"database/sql"
	"sort"
)

type countingQuerier struct {
	Querier
	n int
}

func (c *countingQuerier) Query(query string, args ...any) (*sql.Rows, error) {
	c.n++
	return c.Querier.Query(query, args...)
}

// ListPersonsQueryCount lists Persons and reports how many queries it took.
func ListPersonsQueryCount(q Querier) (int, error) {
	cq := &countingQuerier{Querier: q}
	_, err := ListPersons(cq)
	return cq.n, err
}

// ListEventsQueryCount lists Events and reports how many queries it took.
func ListEventsQueryCount(q Querier) (int, error) {
	cq := &countingQuerier{Querier: q}
	_, err := ListEvents(cq)
	return cq.n, err
}

// ListPlacesQueryCount lists Places and reports how many queries it took.
func ListPlacesQueryCount(q Querier) (int, error) {
	cq := &countingQuerier{Querier: q}
	_, err := ListPlaces(cq)
	return cq.n, err
}

// SortPlaces applies the list order to Place rows.
func SortPlaces(rows []PlaceHeader) {
	sortByTitle(rows, placeTitle, func(h PlaceHeader) string { return h.Entity.Ref })
}

// PlaceGraphRefsForTest lists the titles of the Places a graph load reads:
// ancestors only, or the detail reach.
func PlaceGraphRefsForTest(q Querier, seed []byte, detail bool) ([]string, error) {
	reach := reachAncestors
	if detail {
		reach = reachDetail
	}
	g, err := loadPlaceGraph(q, [][]byte{seed}, reach)
	if err != nil {
		return nil, err
	}
	var refs []string
	for _, h := range g.byID {
		refs = append(refs, placeTitle(h))
	}
	sort.Strings(refs)
	return refs, nil
}
