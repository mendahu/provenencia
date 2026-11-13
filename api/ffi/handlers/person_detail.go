package handlers

import (
	"fmt"
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/conclusiondetails"
	"github.com/mendahu/provenencia/core/database/conclusionsummaries"
	"github.com/mendahu/provenencia/core/database/files"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/persondetail"
	"github.com/mendahu/provenencia/core/kinship"
	"github.com/mendahu/provenencia/core/recordlinks"
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
				claimRows, err := lc.rowsOf(cl.Records.Labels(), cl.Observations)
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
				SourceId:     database.UUIDString(g.Source.ID),
				SourceRef:    g.Source.Ref,
				SourceTitle:  g.Source.Title,
				SourceTypeId: database.UUIDString(g.Source.SourceTypeID),
			}
			for _, cl := range g.Claims {
				n := counts[0]
				counts = counts[1:]
				group.Claims = append(group.Claims, &engine.PersonClaim{
					ClaimId:      database.UUIDString(cl.Accepted.ClaimID),
					GraphId:      database.UUIDString(cl.Graph.ID),
					GraphRef:     cl.Graph.Ref,
					GraphName:    cl.Graph.Name,
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
		family, siblings, summaries, err := persondetail.PersonFamily(c, entityID)
		if err != nil {
			return err
		}
		paths, err := files.RelPaths(c, summaries.PictureImageFileIDs())
		if err != nil {
			return err
		}
		members := func(rs []kinship.Relative) ([]*engine.FamilyMember, error) {
			var out []*engine.FamilyMember
			for _, r := range rs {
				m, err := familyMemberProto(r, summaries, paths)
				if err != nil {
					return nil, err
				}
				out = append(out, m)
			}
			return out, nil
		}
		for _, u := range family.Units {
			partners, err := members(u.Partners)
			if err != nil {
				return err
			}
			children, err := members(u.Children)
			if err != nil {
				return err
			}
			out.Units = append(out.Units, &engine.FamilyUnit{
				Kind:     familyUnitKindProto(u.Kind),
				Partners: partners,
				Children: children,
			})
		}
		if h, ok := summaries.Person(entityID); ok {
			out.Person = personSummaryProto(h, paths)
		}
		for _, s := range siblings {
			m, err := familyMemberProto(s.Relative, summaries, paths)
			if err != nil {
				return err
			}
			o := &engine.OriginSibling{Member: m}
			for _, k := range s.ParentSets {
				o.ParentSets = append(o.ParentSets, int32(k))
			}
			out.OriginSiblings = append(out.OriginSiblings, o)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

// ListPersonEvents returns every canonical event a canonical person took part
// in, in any role, as event summaries with the person's roles and the other
// canonical participants, and the birth and death markers.
func ListPersonEvents(in []byte) ([]byte, error) {
	var req engine.ListPersonEventsRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("list_person_events", err)
	}
	entityID, err := parseID(req.GetEntityId())
	if err != nil {
		return nil, conclusiondetails.ErrNotFound
	}
	out := &engine.ListPersonEventsResponse{}
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		timeline, summaries, err := persondetail.PersonEvents(c, entityID)
		if err != nil {
			return err
		}
		paths, err := files.RelPaths(c, summaries.PictureImageFileIDs())
		if err != nil {
			return err
		}
		for _, r := range timeline.Events {
			e, err := personEventProto(r, summaries, paths)
			if err != nil {
				return err
			}
			out.Events = append(out.Events, e)
		}
		out.Birth, out.Death = lifeMarkerProto(timeline.Birth), lifeMarkerProto(timeline.Death)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

// GetPersonRecordLinks returns the Evidence graphs that hold an accepted member
// of a canonical person and, for each pair of them, one line for each other
// person both name. The lines are computed on each read.
func GetPersonRecordLinks(in []byte) ([]byte, error) {
	var req engine.GetPersonRecordLinksRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("get_person_record_links", err)
	}
	entityID, err := parseID(req.GetEntityId())
	if err != nil {
		return nil, conclusiondetails.ErrNotFound
	}
	out := &engine.GetPersonRecordLinksResponse{}
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		links, err := persondetail.RecordLinks(c, entityID)
		if err != nil {
			return err
		}
		out = recordLinksProto(links)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

// recordLinksProto maps a record links read to the wire, with the two flags
// that say it is partial.
func recordLinksProto(links persondetail.PersonRecordLinks) *engine.GetPersonRecordLinksResponse {
	out := &engine.GetPersonRecordLinksResponse{NotCounted: links.NotCounted, LinesDropped: links.LinesDropped}
	for _, g := range links.Graphs {
		out.Graphs = append(out.Graphs, &engine.RecordLinksGraph{
			GraphId:     database.UUIDString(g.Graph.ID),
			GraphRef:    g.Graph.Ref,
			GraphName:   g.Graph.Name,
			SourceId:    database.UUIDString(g.Source.ID),
			SourceTitle: g.Source.Title,
		})
	}
	for _, p := range links.Pairs {
		pair := &engine.RecordLinksPair{
			GraphAId: database.UUIDString(links.Graphs[p.A].Graph.ID),
			GraphBId: database.UUIDString(links.Graphs[p.B].Graph.ID),
		}
		for _, l := range p.Lines {
			pair.Lines = append(pair.Lines, &engine.RecordLink{
				EntityId: database.UUIDString(l.Entity),
				Strength: recordLinkStrengthProto(l.Strength),
			})
		}
		out.Pairs = append(out.Pairs, pair)
	}
	return out
}

// recordLinkStrengthProto maps a line's weight to the wire.
func recordLinkStrengthProto(s recordlinks.Strength) engine.RecordLinkStrength {
	switch s {
	case recordlinks.StrengthWeak:
		return engine.RecordLinkStrength_RECORD_LINK_STRENGTH_WEAK
	case recordlinks.StrengthStrong:
		return engine.RecordLinkStrength_RECORD_LINK_STRENGTH_STRONG
	}
	return engine.RecordLinkStrength_RECORD_LINK_STRENGTH_UNSPECIFIED
}

// lifeMarkerProto maps a marker; an id that is not there is empty.
func lifeMarkerProto(m persondetail.LifeMarker) *engine.LifeMarker {
	out := &engine.LifeMarker{}
	if len(m.Event) > 0 {
		out.EventId = database.UUIDString(m.Event)
	}
	if m.Place != nil && len(m.Place.Row) > 0 {
		id := database.UUIDString(m.Place.Row)
		switch m.Place.Side {
		case persondetail.SideAfter:
			out.Place = &engine.LifeMarker_AfterEventId{AfterEventId: id}
		case persondetail.SideBefore:
			out.Place = &engine.LifeMarker_BeforeEventId{BeforeEventId: id}
		}
	}
	return out
}

// personEventProto maps one timeline row. PersonEvents returns a summary for
// every event and participant it names, so a missing one is a broken
// invariant, not a case to hide.
func personEventProto(r persondetail.PersonEvent, summaries conclusionsummaries.Summaries, paths map[string]string) (*engine.PersonEvent, error) {
	h, ok := summaries.Event(r.Event)
	if !ok {
		return nil, fmt.Errorf("list_person_events: event %s has no summary", database.UUIDString(r.Event))
	}
	out := &engine.PersonEvent{Event: eventSummaryProto(h, paths), Roles: termsProto(r.Roles)}
	if r.Span != nil {
		out.Span = &engine.TimelineSpan{From: r.Span.From, To: r.Span.To, OpenStart: r.Span.OpenStart, OpenEnd: r.Span.OpenEnd}
	}
	if r.Age != nil {
		out.Age = &engine.AgeRange{Lo: int32(r.Age.Lo), Hi: int32(r.Age.Hi), OpenLo: r.Age.OpenLo, OpenHi: r.Age.OpenHi}
	}
	for _, p := range r.Participants {
		ph, ok := summaries.Person(p.ID)
		if !ok {
			return nil, fmt.Errorf("list_person_events: participant %s has no summary", database.UUIDString(p.ID))
		}
		out.Participants = append(out.Participants, &engine.EventParticipant{Person: personSummaryProto(ph, paths), Roles: termsProto(p.Roles)})
	}
	return out, nil
}

// familyMemberProto maps one relative. PersonFamily drops a relative with no
// summary, so a missing one is a broken invariant, not a case to hide.
func familyMemberProto(r kinship.Relative, summaries conclusionsummaries.Summaries, paths map[string]string) (*engine.FamilyMember, error) {
	h, ok := summaries.Person(r.ID)
	if !ok {
		return nil, fmt.Errorf("get_person_family: relative %s has no summary", database.UUIDString(r.ID))
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
	return m, nil
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
