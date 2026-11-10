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
	// NotCounted is true when the read passed its work bound: it compared
	// nothing and lists no pair.
	NotCounted bool
	// LinesDropped is true when the read found more than maxRecordLinkLines
	// lines and sent that many, strong lines first.
	LinesDropped bool
}

// maxRecordLinkLines is the most lines the read sends, strong first.
// decided: S13-28 round 1 (SEC-1), the caps-for-adversarial-only rule; moved from the app's arc cap by S16-01 (user, 2026-11-08: Go owns the bound): 20 times the 3,675 lines of S13-27's 50-graph household Scale
const maxRecordLinkLines = 75_000

// RecordLinks is the Person page's record-links read: every Evidence graph
// that holds an accepted member of the person, ordered as Claims orders them
// (Source title, then graph name, each with its ref as the tie-break), each
// listed once, and for each pair of them one line per other person both name.
// Nothing is stored: the lines are computed from the graph cache and the
// Evidence layer, bounded as a whole by recordlinks.Rules.MaxReadWork. An
// unknown or merged handle, or one that is no Person, is
// conclusiondetails.ErrNotFound.
//
// It sends at most maxRecordLinkLines lines, strong first, and says in
// NotCounted and LinesDropped when it did not count or dropped lines.
func RecordLinks(c *database.Catalog, entityID []byte) (PersonRecordLinks, error) {
	return recordLinks(c, entityID, matchregistry.Default().RecordLinksConfig(), maxRecordLinkLines)
}

// recordLinks is RecordLinks over a config and a line cap, for the tests that
// reach the bounds on a few graphs.
func recordLinks(c *database.Catalog, entityID []byte, cfg recordlinks.Config, most int) (PersonRecordLinks, error) {
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
	pairs, counted, err := linksBetween(c, entityID, graphs, cfg)
	if err != nil {
		return PersonRecordLinks{}, err
	}
	pairs, dropped := recordlinks.CapLines(pairs, most)
	return PersonRecordLinks{Graphs: graphs, Pairs: pairs, NotCounted: !counted, LinesDropped: dropped}, nil
}

// linksBetween is the lines between every pair of the Evidence graphs, other
// than those naming the page's own entity. It assumes nothing about the page's
// kind. Each graph's Subjects take the entity of their accepted claim; a
// Subject with none draws no line. cfg is the caller's; its Value is set from
// the catalog's rules on a copy. counted is Pairs' report.
func linksBetween(c *database.Catalog, page []byte, graphs []graphcache.EvidenceGraph, cfg recordlinks.Config) ([]recordlinks.Pair, bool, error) {
	db, err := c.DB()
	if err != nil {
		return nil, false, err
	}
	g := c.Graph()
	rules, namePattern, err := valuerules.Load(g, db)
	if err != nil {
		return nil, false, err
	}
	metas, err := valuerules.Metas(g)
	if err != nil {
		return nil, false, err
	}
	cfg.Value = rules
	built := make([]recordlinks.Graph, len(graphs))
	for i, graph := range graphs {
		layer, err := evidencelayer.Load(g, graph.GraphID, namePattern)
		if err != nil {
			return nil, false, err
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
	pairs, counted := recordlinks.Pairs(page, built, metas, cfg)
	return pairs, counted, nil
}
