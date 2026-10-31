package conclusiondetails

import (
	"database/sql"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/graphcache"
)

// FieldsOf composes the fields of several handles at once, keyed by entity
// id, with no gallery leads.
func FieldsOf(c *database.Catalog, ids [][]byte) (map[string][]Field, error) {
	var nodes []*graphcache.Node
	for _, id := range ids {
		n, err := c.Graph().Node(id)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
	}
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	fields, _, err := fieldsOf(c.Graph(), db, nodes, nil, false)
	return fields, err
}

type countingQuerier struct {
	Querier
	n *int
}

func (c countingQuerier) Query(query string, args ...any) (*sql.Rows, error) {
	*c.n++
	return c.Querier.Query(query, args...)
}

// CountFieldsOfQueries composes the fields of several handles and returns how
// many SQL queries that took.
func CountFieldsOfQueries(c *database.Catalog, ids [][]byte, subjectValued bool) (int, error) {
	var nodes []*graphcache.Node
	for _, id := range ids {
		n, err := c.Graph().Node(id)
		if err != nil {
			return 0, err
		}
		nodes = append(nodes, n)
	}
	db, err := c.DB()
	if err != nil {
		return 0, err
	}
	var n int
	_, _, err = fieldsOf(c.Graph(), countingQuerier{db, &n}, nodes, nil, subjectValued)
	return n, err
}
