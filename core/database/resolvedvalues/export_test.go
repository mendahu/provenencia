package resolvedvalues

import "database/sql"

// countingQuerier counts the statements the loader sends.
type countingQuerier struct {
	Querier
	n int
}

func (c *countingQuerier) Query(query string, args ...any) (*sql.Rows, error) {
	c.n++
	return c.Querier.Query(query, args...)
}

func (c *countingQuerier) QueryRow(query string, args ...any) *sql.Row {
	c.n++
	return c.Querier.QueryRow(query, args...)
}

// LoadQueryCount loads the given handles and reports how many queries it took.
func LoadQueryCount(q Querier, ids [][]byte) (int, error) {
	cq := &countingQuerier{Querier: q}
	_, err := load(cq, ids)
	return cq.n, err
}
