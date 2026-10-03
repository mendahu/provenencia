package conclusionheaders

import "database/sql"

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
