// Package persondetail reads the sections of a Person's page that are not the
// reconciled fields: one query per section, each from the graph cache.
package persondetail

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/conclusiondetails"
	"github.com/mendahu/provenencia/core/database/conclusionsummaries"
	"github.com/mendahu/provenencia/core/database/evidencegraphs"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/graphfill"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// ClaimGroup is the accepted claims of one Person that come from one Source.
type ClaimGroup struct {
	Source sources.Record
	Claims []Claim
}

// Claim is one accepted identity claim: the claimed Subject, its claim
// (Accepted.ClaimID) and its Observations, on the Evidence graph that holds
// them. Graph is the graph's shell and Records its Subject records, shared and
// read-only.
type Claim struct {
	Graph evidencegraphs.Graph
	graphfill.SubjectRecord
	Records *graphfill.GraphRecords
}

// Claims returns a Person's accepted claims grouped by Source. Groups are
// ordered by Source title and claims within a group by graph name, each with
// the ref as the tie-break. An unknown or merged handle, or one that is no
// Person, is conclusiondetails.ErrNotFound; a handle with no accepted claim has
// no groups. A Person's records in one Evidence graph are capped at the
// registry's recordlinks MaxEntityMembers (see boundedMembers).
func Claims(c *database.Catalog, entityID []byte) ([]ClaimGroup, error) {
	n, err := personNode(c, entityID)
	if err != nil {
		return nil, err
	}
	return claims(c, n)
}

// personNode is the handle's node when it is a Person, the one place that
// decides whether an id is a Person page's. An unknown or merged handle, or one
// that is no Person, is conclusiondetails.ErrNotFound.
func personNode(c *database.Catalog, entityID []byte) (*graphcache.Node, error) {
	n, err := conclusiondetails.Handle(c, entityID)
	if err != nil {
		return nil, err
	}
	v, err := c.Graph().Vocabulary()
	if err != nil {
		return nil, err
	}
	if !n.Is(v.TypeIDsByKey()[vocabkey.Person]) {
		return nil, conclusiondetails.ErrNotFound
	}
	return n, nil
}

// claims is Claims for a handle already read. It loads no Evidence graph: each
// accepted member is looked up in the graph index (the graph's shell), the
// Source records and the graph's Subject records. A member whose graph left the
// index, or whose Subject or claim is missing from the records or names
// another entity, is skipped: it names a graph dropped mid-read, and the next
// read is right.
func claims(c *database.Catalog, n *graphcache.Node) ([]ClaimGroup, error) {
	g := c.Graph()
	members := boundedMembers(n)
	idx, err := evidencegraphs.IndexOf(g)
	if err != nil {
		return nil, err
	}
	recs, err := sources.RecordsOf(g)
	if err != nil {
		return nil, err
	}
	var graphIDs [][]byte
	asked := map[string]bool{}
	for _, m := range members {
		if _, ok := idx.Graph(m.GraphID); ok && !asked[string(m.GraphID)] {
			asked[string(m.GraphID)] = true
			graphIDs = append(graphIDs, m.GraphID)
		}
	}
	records, err := graphfill.RecordsOf(g, graphIDs)
	if err != nil {
		return nil, err
	}
	var groups []ClaimGroup
	bySource := map[string]int{}
	for _, m := range members {
		shell, ok := idx.Graph(m.GraphID)
		if !ok {
			continue
		}
		r := records[string(m.GraphID)]
		rec, ok := r.Subject(m.SubjectID)
		if !ok || rec.Accepted.ClaimID == nil || !bytes.Equal(rec.Accepted.EntityID, n.ID) {
			continue
		}
		i, ok := bySource[string(shell.SourceID)]
		if !ok {
			source, found := recs.Source(shell.SourceID)
			if !found {
				return nil, fmt.Errorf("persondetail: graph %x names Source %x, which has no record", shell.ID, shell.SourceID)
			}
			i = len(groups)
			bySource[string(shell.SourceID)] = i
			groups = append(groups, ClaimGroup{Source: source})
		}
		groups[i].Claims = append(groups[i].Claims, Claim{Graph: shell, SubjectRecord: rec, Records: r})
	}
	conclusionsummaries.SortByTitle(groups, func(g ClaimGroup) string { return g.Source.Title }, func(g ClaimGroup) string { return g.Source.Ref })
	for i := range groups {
		claims := groups[i].Claims
		sort.SliceStable(claims, func(a, b int) bool {
			return strings.ToLower(claims[a].Subject.Ref) < strings.ToLower(claims[b].Subject.Ref)
		})
		conclusionsummaries.SortByTitle(claims, func(cl Claim) string { return cl.Graph.Name }, func(cl Claim) string { return cl.Graph.Ref })
	}
	return groups, nil
}

// boundedMembers is n.AcceptedMembers keeping at most the registry's
// recordlinks MaxEntityMembers per Evidence graph, in member order; every
// accepted member counts, the person's first record included. More is crafted
// data: a real person files one or two records in a graph, and the claims read
// returns each record it keeps.
func boundedMembers(n *graphcache.Node) []graphcache.Member {
	most := matchregistry.Default().RecordLinks.MaxEntityMembers
	perGraph := map[string]int{}
	var kept []graphcache.Member
	for _, m := range n.AcceptedMembers() {
		if perGraph[string(m.GraphID)] >= most {
			continue
		}
		perGraph[string(m.GraphID)]++
		kept = append(kept, m)
	}
	return kept
}
