// Package persondetail reads the sections of a Person's page that are not the
// reconciled fields: one query per section, each from the graph cache.
package persondetail

import (
	"bytes"
	"sort"
	"strings"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/conclusiondetails"
	"github.com/mendahu/provenencia/core/database/conclusionheaders"
	"github.com/mendahu/provenencia/core/database/graphcache"
)

// ClaimGroup is the accepted claims of one Person that come from one Source.
type ClaimGroup struct {
	SourceID     []byte
	SourceRef    string
	Title        string
	SourceTypeID []byte
	Claims       []Claim
}

// Claim is one accepted identity claim: the claimed Subject and its
// Observations, on the Evidence graph that holds them. Graph is the loaded
// graph, shared and read-only.
type Claim struct {
	ClaimID      []byte
	Graph        graphcache.EvidenceGraph
	Subject      graphcache.EvidenceSubject
	Observations []graphcache.EvidenceObservation
}

// Claims returns a Person's accepted claims grouped by Source. Groups are
// ordered by Source title and claims within a group by graph name, each with
// the ref as the tie-break. An unknown or merged handle is
// conclusiondetails.ErrNotFound; a handle with no accepted claim has no
// groups.
func Claims(c *database.Catalog, entityID []byte) ([]ClaimGroup, error) {
	n, err := conclusiondetails.Handle(c, entityID)
	if err != nil {
		return nil, err
	}
	members, err := memberGraphs(c, n)
	if err != nil {
		return nil, err
	}
	var groups []ClaimGroup
	bySource := map[string]int{}
	for _, m := range members {
		claim := Claim{ClaimID: m.member.ClaimID, Graph: m.graph, Subject: m.subject}
		for _, o := range m.graph.Observations {
			if bytes.Equal(o.SubjectID, m.subject.ID) {
				claim.Observations = append(claim.Observations, o)
			}
		}
		i, ok := bySource[string(m.graph.SourceID)]
		if !ok {
			i = len(groups)
			bySource[string(m.graph.SourceID)] = i
			groups = append(groups, ClaimGroup{
				SourceID:     m.graph.SourceID,
				SourceRef:    m.graph.SourceRef,
				Title:        m.graph.SourceTitle,
				SourceTypeID: m.graph.SourceTypeID,
			})
		}
		groups[i].Claims = append(groups[i].Claims, claim)
	}
	conclusionheaders.SortByTitle(groups, func(g ClaimGroup) string { return g.Title }, func(g ClaimGroup) string { return g.SourceRef })
	for i := range groups {
		claims := groups[i].Claims
		sort.SliceStable(claims, func(a, b int) bool {
			return strings.ToLower(claims[a].Subject.Ref) < strings.ToLower(claims[b].Subject.Ref)
		})
		conclusionheaders.SortByTitle(claims, func(cl Claim) string { return cl.Graph.GraphName }, func(cl Claim) string { return cl.Graph.GraphRef })
	}
	return groups, nil
}

// memberGraph is one accepted member of a handle with what its Evidence graph
// says about it.
type memberGraph struct {
	member  graphcache.EvidenceMember
	graph   graphcache.EvidenceGraph
	subject graphcache.EvidenceSubject
}

// memberGraphs returns each accepted member of n with its loaded Evidence
// graph, claim and Subject, in member order. A member whose graph is not
// found, or whose Subject or claim is missing from the loaded graph, is
// skipped: it names a graph dropped mid-read, and the next read is right.
func memberGraphs(c *database.Catalog, n *graphcache.Node) ([]memberGraph, error) {
	graphs := map[string]graphcache.EvidenceGraph{}
	var out []memberGraph
	for _, m := range n.AcceptedMembers() {
		eg, ok := graphs[string(m.GraphID)]
		if !ok {
			var err error
			if eg, err = c.Graph().EvidenceGraph(m.GraphID); err != nil {
				return nil, err
			}
			graphs[string(m.GraphID)] = eg
		}
		if !eg.Found {
			continue
		}
		var member *graphcache.EvidenceMember
		for i := range eg.Members {
			if bytes.Equal(eg.Members[i].SubjectID, m.SubjectID) && bytes.Equal(eg.Members[i].EntityID, n.ID) {
				member = &eg.Members[i]
				break
			}
		}
		var subject *graphcache.EvidenceSubject
		for i := range eg.Subjects {
			if bytes.Equal(eg.Subjects[i].ID, m.SubjectID) {
				subject = &eg.Subjects[i]
				break
			}
		}
		if member == nil || member.ClaimID == nil || subject == nil {
			continue
		}
		out = append(out, memberGraph{member: *member, graph: eg, subject: *subject})
	}
	return out, nil
}
