package persondetail

import (
	"slices"
	"sort"

	"github.com/mendahu/provenencia/core/database/conclusionsummaries"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// participantRoleOrder is the order roles come in on an event: the people who
// mattered most to it first. A newly seeded role must be listed here
// (TestParticipantRolesCoverSeed).
var participantRoleOrder = []vocabkey.Role{
	vocabkey.RoleSubject, vocabkey.RoleFather, vocabkey.RoleMother, vocabkey.RoleChild,
	vocabkey.RoleWitness, vocabkey.RoleInformant, vocabkey.RolePallbearer,
}

// sortParticipants orders one event's participants by rank, then by index (their
// place in the list order of PersonsByIDs). A participant's rank is the best
// rank of its roles: a listed (product) role ranks by its place in
// participantRoleOrder; a role the researcher defined ranks after every listed
// one, the event's own such roles ordered by label; a participant with no role
// ranks last.
func sortParticipants(ps []conclusionsummaries.EventParticipant, index map[string]int) {
	var unlisted []conclusionsummaries.Term
	seen := map[string]bool{}
	listed := func(t conclusionsummaries.Term) int {
		key := t.ProductKey()
		if key == "" {
			return -1
		}
		return slices.Index(participantRoleOrder, vocabkey.Role(key))
	}
	for _, p := range ps {
		for _, t := range p.Roles {
			if listed(t) < 0 && !seen[string(t.ID)] {
				seen[string(t.ID)] = true
				unlisted = append(unlisted, t)
			}
		}
	}
	conclusionsummaries.SortByTitle(unlisted,
		func(t conclusionsummaries.Term) string { return t.Label },
		func(t conclusionsummaries.Term) string { return t.Key })
	unlistedRank := map[string]int{}
	for i, t := range unlisted {
		unlistedRank[string(t.ID)] = len(participantRoleOrder) + i
	}
	none := len(participantRoleOrder) + len(unlisted)
	rank := func(p conclusionsummaries.EventParticipant) int {
		best := none
		for _, t := range p.Roles {
			r, ok := unlistedRank[string(t.ID)]
			if i := listed(t); i >= 0 {
				r, ok = i, true
			}
			if ok && r < best {
				best = r
			}
		}
		return best
	}
	ranks := make(map[string]int, len(ps))
	for _, p := range ps {
		ranks[string(p.ID)] = rank(p)
	}
	sort.SliceStable(ps, func(i, j int) bool {
		ri, rj := ranks[string(ps[i].ID)], ranks[string(ps[j].ID)]
		if ri != rj {
			return ri < rj
		}
		return index[string(ps[i].ID)] < index[string(ps[j].ID)]
	})
}
