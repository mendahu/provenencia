package handlers

import (
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/conclusiondetails"
	"github.com/mendahu/provenencia/core/database/conclusionsummaries"
	"github.com/mendahu/provenencia/core/database/files"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/persondetail"
	"github.com/mendahu/provenencia/core/kinship"
	"google.golang.org/protobuf/proto"
)

// ListPersonClaims returns a canonical person's accepted claims grouped by
// Source, each claim with its Subject and that Subject's Observations.
func ListPersonClaims(in []byte) ([]byte, error) {
	var req engine.ListPersonClaimsRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("list_person_claims", err)
	}
	entityID, err := parseID(req.GetEntityId())
	if err != nil {
		return nil, conclusiondetails.ErrNotFound
	}
	out := &engine.ListPersonClaimsResponse{}
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		groups, err := persondetail.Claims(c, entityID)
		if err != nil {
			return err
		}
		lc, err := newListedContext(c)
		if err != nil {
			return err
		}
		// Every claim's rows go through listedProtos together, so the response
		// issues one depiction path query; the result is split back by count.
		var rows []observations.Listed
		var counts []int
		for _, g := range groups {
			for _, cl := range g.Claims {
				claimRows, err := lc.rowsOf(cl.Graph, cl.Observations)
				if err != nil {
					return err
				}
				rows = append(rows, claimRows...)
				counts = append(counts, len(claimRows))
			}
		}
		protos, err := listedProtos(c, rows)
		if err != nil {
			return err
		}
		next := 0
		for _, g := range groups {
			group := &engine.PersonClaimGroup{
				SourceId:     database.UUIDString(g.SourceID),
				SourceRef:    g.SourceRef,
				SourceTitle:  g.Title,
				SourceTypeId: database.UUIDString(g.SourceTypeID),
			}
			for _, cl := range g.Claims {
				n := counts[0]
				counts = counts[1:]
				group.Claims = append(group.Claims, &engine.PersonClaim{
					ClaimId:      database.UUIDString(cl.ClaimID),
					GraphId:      database.UUIDString(cl.Graph.GraphID),
					GraphRef:     cl.Graph.GraphRef,
					GraphName:    cl.Graph.GraphName,
					Subject:      evidenceSubjectProto(cl.Subject),
					Observations: protos[next : next+n : next+n],
				})
				next += n
			}
			out.Groups = append(out.Groups, group)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

// GetPersonFamily returns a canonical person's family units, each relative as a
// person summary with the link that says why they are there.
func GetPersonFamily(in []byte) ([]byte, error) {
	var req engine.GetPersonFamilyRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("get_person_family", err)
	}
	entityID, err := parseID(req.GetEntityId())
	if err != nil {
		return nil, conclusiondetails.ErrNotFound
	}
	out := &engine.GetPersonFamilyResponse{}
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		family, summaries, err := persondetail.PersonFamily(c, entityID)
		if err != nil {
			return err
		}
		paths, err := files.RelPaths(c, summaries.PictureImageFileIDs())
		if err != nil {
			return err
		}
		members := func(rs []kinship.Relative) []*engine.FamilyMember {
			var out []*engine.FamilyMember
			for _, r := range rs {
				if m := familyMemberProto(r, summaries, paths); m != nil {
					out = append(out, m)
				}
			}
			return out
		}
		for _, u := range family.Units {
			out.Units = append(out.Units, &engine.FamilyUnit{
				Kind:     familyUnitKindProto(u.Kind),
				Partners: members(u.Partners),
				Children: members(u.Children),
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

// familyMemberProto maps one relative; nil when the summaries hold none for it.
func familyMemberProto(r kinship.Relative, summaries conclusionsummaries.Summaries, paths map[string]string) *engine.FamilyMember {
	h, ok := summaries.Person(r.ID)
	if !ok {
		return nil
	}
	m := &engine.FamilyMember{Person: personSummaryProto(h, paths)}
	if r.Link != nil {
		link := &engine.FamilyLink{Recorded: !r.Link.OnlyImplied}
		for _, a := range r.Link.Associations {
			link.AssociationIds = append(link.AssociationIds, database.UUIDString(a))
		}
		for _, s := range r.Link.Shapes {
			link.Implied = append(link.Implied, &engine.KinshipShape{
				Pattern: kinshipPatternProto(s.Pattern),
				PivotId: database.UUIDString(s.Pivot),
			})
		}
		m.Link = link
	}
	return m
}

func familyUnitKindProto(k kinship.UnitKind) engine.FamilyUnitKind {
	switch k {
	case kinship.UnitParents:
		return engine.FamilyUnitKind_FAMILY_UNIT_KIND_PARENTS
	case kinship.UnitPartner:
		return engine.FamilyUnitKind_FAMILY_UNIT_KIND_PARTNER
	case kinship.UnitOneParent:
		return engine.FamilyUnitKind_FAMILY_UNIT_KIND_ONE_PARENT
	case kinship.UnitSiblings:
		return engine.FamilyUnitKind_FAMILY_UNIT_KIND_SIBLINGS
	}
	return engine.FamilyUnitKind_FAMILY_UNIT_KIND_UNSPECIFIED
}

// kinshipPatternProto maps a pattern code onto the wire; TestKinshipPatternProto
// fails on a code that maps to unspecified.
func kinshipPatternProto(p kinship.Pattern) engine.KinshipPattern {
	switch p {
	case kinship.PatternMarriageSpouses:
		return engine.KinshipPattern_KINSHIP_PATTERN_MARRIAGE_SPOUSES
	case kinship.PatternBirthParents:
		return engine.KinshipPattern_KINSHIP_PATTERN_BIRTH_PARENTS
	case kinship.PatternSharedParent:
		return engine.KinshipPattern_KINSHIP_PATTERN_SHARED_PARENT
	case kinship.PatternParentOfParent:
		return engine.KinshipPattern_KINSHIP_PATTERN_PARENT_OF_PARENT
	}
	return engine.KinshipPattern_KINSHIP_PATTERN_UNSPECIFIED
}
