// Package likenesses reads the likenesses a handle's gallery shows: its
// cached likeness values whose reason a gallery lists
// (autoreconcile.Reason.InGallery), earliest recorded first. Lead picks the
// likeness a handle leads with.
package likenesses

import (
	"bytes"
	"database/sql"
	"sort"

	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/depictionvalues"
	"github.com/mendahu/provenencia/core/database/sqlbatch"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// Likeness is one likeness value a gallery shows.
type Likeness struct {
	Rank           int // the value's rank in auto_reconciler_values
	ImageFileID    []byte
	ObservationIDs [][]byte // the records that went into it, ascending; [0] is the first recorded
	// Preferred is true for the likeness the handle's preferred_likeness_observation_id
	// names (S10-39); at most one.
	Preferred bool
}

const sqlGallery = `SELECT v.entity_id, v.rank, v.reason, v.value_image_file_id,
		COALESCE(f.media_type, ''), ao.observation_id, e.preferred_likeness_observation_id
	FROM auto_reconciler_values v
	JOIN canonical_entities e ON e.id = v.entity_id
	JOIN properties p ON p.id = v.property_id
	JOIN auto_reconciler_outcomes ao ON ao.entity_id = v.entity_id
		AND ao.property_id = v.property_id AND ao.value_rank = v.rank
	LEFT JOIN files f ON f.id = v.value_image_file_id
	WHERE p.key = ? AND p.origin = ? AND v.entity_id IN (`

// ForEntities returns each handle's gallery, keyed by string(entity id),
// earliest first: by ObservationIDs[0], then Rank. A handle with none is absent.
func ForEntities(q sqlbatch.Querier, entityIDs [][]byte) (map[string][]Likeness, error) {
	ids := database.UniqueBlobIDs(entityIDs)
	if len(ids) == 0 {
		return nil, nil
	}
	type key struct {
		entity string
		rank   int
	}
	byValue := map[key]*Likeness{}
	var order []key
	pointers := map[string][]byte{}
	err := sqlbatch.Query(q, sqlGallery, ")", []any{vocabkey.PropertyLikeness, vocabkey.OriginProvenencia}, ids, func(rows *sql.Rows) error {
		var (
			entity, image, obs, pointer []byte
			rank                        int
			reason, mediaType           string
		)
		if err := rows.Scan(&entity, &rank, &reason, &image, &mediaType, &obs, &pointer); err != nil {
			return err
		}
		if len(pointer) > 0 {
			pointers[string(entity)] = append([]byte(nil), pointer...)
		}
		if !autoreconcile.Reason(reason).InGallery() {
			return nil
		}
		image = depictionvalues.ReadableImage(image, mediaType)
		if image == nil {
			return nil
		}
		k := key{string(entity), rank}
		l, ok := byValue[k]
		if !ok {
			l = &Likeness{Rank: rank, ImageFileID: append([]byte(nil), image...)}
			byValue[k] = l
			order = append(order, k)
		}
		l.ObservationIDs = append(l.ObservationIDs, append([]byte(nil), obs...))
		return nil
	})
	if err != nil {
		return nil, err
	}
	out := map[string][]Likeness{}
	for _, k := range order {
		l := byValue[k]
		sort.Slice(l.ObservationIDs, func(i, j int) bool { return bytes.Compare(l.ObservationIDs[i], l.ObservationIDs[j]) < 0 })
		out[k.entity] = append(out[k.entity], *l)
	}
	for entity, ls := range out {
		sort.SliceStable(ls, func(i, j int) bool {
			if c := bytes.Compare(ls[i].ObservationIDs[0], ls[j].ObservationIDs[0]); c != 0 {
				return c < 0
			}
			return ls[i].Rank < ls[j].Rank
		})
		if p, ok := pointers[entity]; ok {
			if i, ok := ByObservation(ls, p); ok {
				ls[i].Preferred = true
			}
		}
	}
	return out, nil
}

// ByObservation returns the index of the gallery likeness an Observation went
// into: the one rule for which likeness a record names. -1, false when none.
func ByObservation(gallery []Likeness, observationID []byte) (i int, ok bool) {
	for i := range gallery {
		for _, id := range gallery[i].ObservationIDs {
			if bytes.Equal(id, observationID) {
				return i, true
			}
		}
	}
	return -1, false
}

// Lead returns the likeness a handle leads with: its preferred portrait when
// that likeness is in the gallery, else the earliest (working rule "A preferred
// portrait is a display pointer, not a claim"). ok is false when the gallery
// is empty.
func Lead(gallery []Likeness) (l Likeness, ok bool) {
	for _, g := range gallery {
		if g.Preferred {
			return g, true
		}
	}
	if len(gallery) == 0 {
		return Likeness{}, false
	}
	return gallery[0], true
}
