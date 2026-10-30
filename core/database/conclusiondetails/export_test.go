package conclusiondetails

import (
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
	fields, _, err := fieldsOf(c.Graph(), db, nodes, nil)
	return fields, err
}
