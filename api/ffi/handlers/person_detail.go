package handlers

import (
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/conclusiondetails"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/persondetail"
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
					ClaimId:   database.UUIDString(cl.ClaimID),
					GraphId:   database.UUIDString(cl.Graph.GraphID),
					GraphRef:  cl.Graph.GraphRef,
					GraphName: cl.Graph.GraphName,
					Subject: evidenceSubjectProto(cl.Subject),
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
