package persondetail

import (
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/evidencelayer"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/valuerules"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/recordlinks"
)

// PersonRecordLinks is a Person's Evidence graphs, in the Sources section's
// order, and the lines between each pair; Pairs index Graphs.
type PersonRecordLinks struct {
	Graphs []graphcache.EvidenceGraph
	Pairs  []recordlinks.Pair
}

// RecordLinks is the Person page's record-links read: every Evidence graph
// that holds an accepted member of the person, ordered as Claims orders them
// (Source title, then graph name, each with its ref as the tie-break), each
// listed once, and for each pair of them one line per other person both name.
// Nothing is stored: the lines are computed from the graph cache and the
// Evidence layer, bounded as a whole by recordlinks.Rules.MaxReadWork. An
// unknown or merged handle, or one that is no Person, is
// conclusiondetails.ErrNotFound.
func RecordLinks(c *database.Catalog, entityID []byte) (PersonRecordLinks, error) {
	n, err := personNode(c, entityID)
	if err != nil {
		return PersonRecordLinks{}, err
	}
	groups, err := claims(c, n)
	if err != nil {
		return PersonRecordLinks{}, err
	}
	var graphs []graphcache.EvidenceGraph
	seen := map[string]bool{}
	for _, g := range groups {
		for _, cl := range g.Claims {
			if !seen[string(cl.Graph.GraphID)] {
				seen[string(cl.Graph.GraphID)] = true
				graphs = append(graphs, cl.Graph)
			}
		}
	}
	pairs, err := linksBetween(c, entityID, graphs)
	if err != nil {
		return PersonRecordLinks{}, err
	}
	return PersonRecordLinks{Graphs: graphs, Pairs: pairs}, nil
}

// linksBetween is the lines between every pair of the Evidence graphs, other
// than those naming the page's own entity. It assumes nothing about the page's
// kind. Each graph's Subjects take the entity of their accepted claim; a
// Subject with none draws no line.
func linksBetween(c *database.Catalog, page []byte, graphs []graphcache.EvidenceGraph) ([]recordlinks.Pair, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	g := c.Graph()
	rules, namePattern, err := valuerules.Load(g, db)
	if err != nil {
		return nil, err
	}
	metas, err := valuerules.Metas(g)
	if err != nil {
		return nil, err
	}
	cfg := matchregistry.Default().RecordLinksConfig()
	cfg.Value = rules
	built := make([]recordlinks.Graph, len(graphs))
	for i, graph := range graphs {
		layer, err := evidencelayer.Load(g, graph.GraphID, namePattern)
		if err != nil {
			return nil, err
		}
		entityOf := map[string][]byte{}
		for _, m := range graph.Members {
			entityOf[string(m.SubjectID)] = m.EntityID
		}
		members := make([]recordlinks.Member, len(layer.Subjects))
		for j, s := range layer.Subjects {
			members[j] = recordlinks.Member{Subject: s, Entity: entityOf[string(s.ID)]}
		}
		built[i] = recordlinks.NewGraph(members, layer.Links)
	}
	return recordlinks.Pairs(page, built, metas, cfg), nil
}
