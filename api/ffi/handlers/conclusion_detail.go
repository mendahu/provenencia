package handlers

import (
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/conclusiondetails"
	"github.com/mendahu/provenencia/core/database/conclusionheaders"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/valuecodec"
	"google.golang.org/protobuf/proto"
)

func GetConclusionDetail(in []byte) ([]byte, error) {
	var req engine.GetConclusionDetailRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("get_conclusion_detail", err)
	}
	entityID, err := parseID(req.GetEntityId())
	if err != nil {
		return nil, conclusiondetails.ErrNotFound
	}
	out := &engine.ConclusionDetail{}
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		db, err := c.DB()
		if err != nil {
			return err
		}
		d, err := conclusiondetails.ForEntity(db, entityID)
		if err != nil {
			return err
		}
		out = conclusionDetailProto(d)
		return attachHeader(c, db, d, out)
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

func conclusionDetailProto(d conclusiondetails.Detail) *engine.ConclusionDetail {
	out := &engine.ConclusionDetail{Entity: canonicalEntityProto(d.Entity), MemberCount: int32(d.MemberCount)}
	for _, f := range d.Fields {
		pf := &engine.ConclusionField{
			PropertyId:  uuidString(f.PropertyID),
			PropertyKey: f.PropertyKey,
			Label:       f.Label,
			ValueType:   f.ValueType,
			State:       string(f.State),
		}
		for _, v := range f.Values {
			pf.Values = append(pf.Values, &engine.ReconciledValueDetail{
				Rank:    int32(v.Rank),
				Reason:  v.Reason,
				Support: int32(v.Support),
				Against: int32(v.Against),
				Value:   conclusionValueProto(v.Value),
			})
		}
		for _, o := range f.Outcomes {
			po := &engine.ReconcilerOutcomeDetail{
				ObservationId:          uuidString(o.ObservationID),
				ObservationRef:         o.ObservationRef,
				Reason:                 o.Reason,
				ValueRank:              int32(o.ValueRank),
				Recorded:               conclusionValueProto(o.Recorded),
				SubjectId:              uuidString(o.SubjectID),
				SubjectRef:             o.SubjectRef,
				CitationId:             uuidString(o.CitationID),
				ArtifactId:             uuidString(o.ArtifactID),
				VoteSupport:            int32(o.Vote.Support),
				VoteTotal:              int32(o.Vote.Of),
				SourceId:               uuidString(o.SourceID),
				SourceTitle:            o.SourceTitle,
				CredibilityKey:         o.CredibilityKey,
				TranscriptionUncertain: o.Provenance.Uncertain,
				ClaimConfidenceKey:     o.ClaimConfidenceKey,
				CredibilityOffset:      int32(o.Provenance.Credibility),
				ClaimConfidenceOffset:  int32(o.Provenance.ClaimConfidence),
			}
			if len(o.DeniedBy) > 0 {
				po.DeniedByObservationId = uuidString(o.DeniedBy)
			}
			pf.Outcomes = append(pf.Outcomes, po)
		}
		out.Fields = append(out.Fields, pf)
	}
	return out
}

func conclusionValueProto(v conclusiondetails.Value) *engine.ConclusionValue {
	switch {
	case v.Name != nil:
		return &engine.ConclusionValue{Kind: &engine.ConclusionValue_Name{Name: valuecodec.NameToProto(*v.Name)}}
	case v.Date != nil:
		return &engine.ConclusionValue{Kind: &engine.ConclusionValue_Date{Date: valuecodec.DateToProto(*v.Date)}}
	case len(v.TermID) > 0:
		return &engine.ConclusionValue{Kind: &engine.ConclusionValue_Term{Term: &engine.ConclusionTerm{
			Id: uuidString(v.TermID), Key: v.TermKey, Label: v.TermLabel,
		}}}
	case v.HasInteger:
		return &engine.ConclusionValue{Kind: &engine.ConclusionValue_Integer{Integer: v.Integer}}
	case v.HasText:
		return &engine.ConclusionValue{Kind: &engine.ConclusionValue_Text{Text: v.Text}}
	}
	return &engine.ConclusionValue{}
}

// attachHeader reads the handle's header for its kind beside the detail.
func attachHeader(c *database.Catalog, db conclusionheaders.Querier, d conclusiondetails.Detail, out *engine.ConclusionDetail) error {
	st, err := subjecttypes.GetByID(c, d.Entity.SubjectTypeID)
	if err != nil {
		return err
	}
	if st.Origin != subjecttypes.OriginProvenencia {
		return nil
	}
	ids := [][]byte{d.Entity.ID}
	switch st.Key {
	case "person":
		h, err := conclusionheaders.PersonsByIDs(db, ids)
		if err != nil || len(h) != 1 {
			return err
		}
		out.Header = &engine.ConclusionDetail_Person{Person: personHeaderProto(h[0])}
	case "event":
		h, err := conclusionheaders.EventsByIDs(db, ids)
		if err != nil || len(h) != 1 {
			return err
		}
		out.Header = &engine.ConclusionDetail_Event{Event: eventHeaderProto(h[0])}
	case "place":
		h, found, err := conclusionheaders.PlaceDetail(db, d.Entity.ID)
		if err != nil || !found {
			return err
		}
		out.Header = &engine.ConclusionDetail_Place{Place: placeHeaderProto(h)}
	}
	return nil
}
