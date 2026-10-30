package handlers

import (
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/conclusiondetails"
	"github.com/mendahu/provenencia/core/database/conclusionheaders"
	"github.com/mendahu/provenencia/core/database/files"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/valuecodec"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
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
		d, err := conclusiondetails.ForEntity(c, entityID)
		if err != nil {
			return err
		}
		out = conclusionDetailProto(d)
		lead, err := attachHeader(c, d, out)
		if err != nil {
			return err
		}
		return attachPictures(c, d, lead, out)
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}

// SetPreferredLikeness chooses a handle's gallery portrait or thumbnail
// pointer among its gallery pictures, or clears the choice (S10-39).
func SetPreferredLikeness(in []byte) ([]byte, error) {
	var req engine.SetPreferredLikenessRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("set_preferred_likeness", err)
	}
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	entityID, err := parseID(req.GetEntityId())
	if err != nil {
		return nil, err
	}
	var observationID []byte
	if req.GetObservationId() != "" {
		if observationID, err = parseID(req.GetObservationId()); err != nil {
			return nil, err
		}
	}
	var out *engine.SetPreferredLikenessResponse
	err = withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		_, wrote, err := writes.Run(c, writes.Op{Action: "set_preferred_likeness", UserID: userID},
			func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
				changes, err := canonicalentities.SetPreferredLikeness(tx, userID, entityID, observationID)
				return struct{}{}, changes, err
			})
		if err != nil {
			return err
		}
		out = &engine.SetPreferredLikenessResponse{Effects: writeEffectsProto(wrote)}
		return nil
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
			PropertyId:  database.UUIDString(f.PropertyID),
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
				Value:   conclusionValueProto(v.Value, nil),
			})
		}
		for _, o := range f.Outcomes {
			pf.Outcomes = append(pf.Outcomes, reconcilerOutcomeProto(o))
		}
		out.Fields = append(out.Fields, pf)
	}
	return out
}

func reconcilerOutcomeProto(o conclusiondetails.Outcome) *engine.ReconcilerOutcomeDetail {
	po := &engine.ReconcilerOutcomeDetail{
		ObservationId:          database.UUIDString(o.ObservationID),
		ObservationRef:         o.ObservationRef,
		Reason:                 o.Reason,
		ValueRank:              int32(o.ValueRank),
		Recorded:               conclusionValueProto(o.Recorded, nil),
		SubjectId:              database.UUIDString(o.SubjectID),
		SubjectRef:             o.SubjectRef,
		GraphId:                database.UUIDString(o.GraphID),
		CitationId:             database.UUIDString(o.CitationID),
		ArtifactId:             database.UUIDString(o.ArtifactID),
		VoteSupport:            int32(o.Vote.Support),
		VoteTotal:              int32(o.Vote.Of),
		SourceId:               database.UUIDString(o.SourceID),
		SourceTitle:            o.SourceTitle,
		CredibilityKey:         o.CredibilityKey,
		TranscriptionUncertain: o.Provenance.Uncertain,
		ClaimConfidenceKey:     o.ClaimConfidenceKey,
		CredibilityOffset:      int32(o.Provenance.Credibility),
		ClaimConfidenceOffset:  int32(o.Provenance.ClaimConfidence),
	}
	if len(o.DeniedBy) > 0 {
		po.DeniedByObservationId = database.UUIDString(o.DeniedBy)
	}
	return po
}

// conclusionValueProto is a typed value as a ConclusionValue. depiction, when
// not nil, is the result: a likeness has no field on conclusiondetails.Value.
func conclusionValueProto(v conclusiondetails.Value, depiction *engine.DepictionValue) *engine.ConclusionValue {
	if depiction != nil {
		return &engine.ConclusionValue{Kind: &engine.ConclusionValue_Depiction{Depiction: depiction}}
	}
	switch {
	case v.Name != nil:
		return &engine.ConclusionValue{Kind: &engine.ConclusionValue_Name{Name: valuecodec.NameToProto(*v.Name)}}
	case v.Date != nil:
		return &engine.ConclusionValue{Kind: &engine.ConclusionValue_Date{Date: valuecodec.DateToProto(*v.Date)}}
	case len(v.TermID) > 0 || v.TermKey != "":
		term := &engine.ConclusionTerm{Key: v.TermKey, Label: v.TermLabel, Origin: v.TermOrigin}
		if len(v.TermID) > 0 {
			term.Id = database.UUIDString(v.TermID)
		}
		return &engine.ConclusionValue{Kind: &engine.ConclusionValue_Term{Term: term}}
	case v.HasInteger:
		return &engine.ConclusionValue{Kind: &engine.ConclusionValue_Integer{Integer: v.Integer}}
	case v.HasText:
		return &engine.ConclusionValue{Kind: &engine.ConclusionValue_Text{Text: v.Text}}
	}
	return &engine.ConclusionValue{}
}

// attachHeader reads the handle's header for its kind beside the detail. It
// returns the header's picture lead File id (nil when none).
func attachHeader(c *database.Catalog, d conclusiondetails.Detail, out *engine.ConclusionDetail) ([]byte, error) {
	st, err := subjecttypes.GetByID(c, d.Entity.SubjectTypeID)
	if err != nil {
		return nil, err
	}
	if st.Origin != vocabkey.OriginProvenencia {
		return nil, nil
	}
	ids := [][]byte{d.Entity.ID}
	switch st.Key {
	case vocabkey.Person:
		h, err := conclusionheaders.PersonsByIDs(c, ids)
		if err != nil || len(h) != 1 {
			return nil, err
		}
		out.Header = &engine.ConclusionDetail_Person{Person: personHeaderProto(h[0])}
		return h[0].PictureImageFileID, nil
	case vocabkey.Event:
		h, err := conclusionheaders.EventsByIDs(c, ids)
		if err != nil || len(h) != 1 {
			return nil, err
		}
		out.Header = &engine.ConclusionDetail_Event{Event: eventHeaderProto(h[0])}
		return h[0].PictureImageFileID, nil
	case vocabkey.Place:
		h, err := conclusionheaders.PlacesByIDs(c, ids)
		if err != nil || len(h) != 1 {
			return nil, err
		}
		if err := conclusionheaders.AttachPlaceRelationships(c, &h[0]); err != nil {
			return nil, err
		}
		out.Header = &engine.ConclusionDetail_Place{Place: placeHeaderProto(h[0])}
		return h[0].PictureImageFileID, nil
	}
	return nil, nil
}

// attachPictures lists the detail's pictures with their image paths, and sets
// the header's lead path, in one RelPaths call. A picture whose path cannot
// be derived is still listed, with an empty path.
func attachPictures(c *database.Catalog, d conclusiondetails.Detail, lead []byte, out *engine.ConclusionDetail) error {
	var imageIDs [][]byte
	if len(lead) > 0 {
		imageIDs = append(imageIDs, lead)
	}
	for _, l := range d.Pictures {
		imageIDs = append(imageIDs, l.ImageFileID)
	}
	paths, err := files.RelPaths(c, imageIDs)
	if err != nil {
		return err
	}
	path := paths[string(lead)]
	if person := out.GetPerson(); person != nil {
		person.PictureImageRelPath = path
	}
	if event := out.GetEvent(); event != nil {
		event.PictureImageRelPath = path
	}
	if place := out.GetPlace(); place != nil {
		place.PictureImageRelPath = path
	}
	for _, l := range d.Pictures {
		out.Pictures = append(out.Pictures, &engine.PictureDetail{
			ImageRelPath: paths[string(l.ImageFileID)],
			Record:       reconcilerOutcomeProto(l.Record),
			Preferred:    l.Preferred,
			Leads:        l.Leads,
		})
	}
	return nil
}
