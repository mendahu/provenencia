package matching

import "database/sql"

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

// ForSubjectQueryCount runs ForSubject and reports how many queries it took.
func ForSubjectQueryCount(q Querier, subjectID []byte) (int, error) {
	cq := &countingQuerier{Querier: q}
	_, err := ForSubject(cq, subjectID, Options{})
	return cq.n, err
}
