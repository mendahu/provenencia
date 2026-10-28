// Package gallery reads the pictures a handle's gallery shows: its cached
// depiction values with their gallery records, each record whose own outcome
// reason a gallery lists (autoreconcile.Reason.InGallery), earliest recorded
// first. Lead picks the picture a handle leads with.
package gallery

import (
	"bytes"
	"database/sql"
	"sort"

	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/sqlbatch"
	"github.com/mendahu/provenencia/core/depictionvalue"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// Picture is one depiction value a gallery shows.
type Picture struct {
	Rank           int // the value's rank in auto_reconciler_values
	ImageFileID    []byte
	ObservationIDs [][]byte // the value's gallery records, ascending; [0] is the first recorded; a negative, denied, or provisional record of the same picture is not one
	// Preferred is true for the picture the handle's preferred_likeness_observation_id
	// names (S10-39); at most one.
	Preferred bool
	// Leads is the picture the handle leads with: the preferred one when it is
	// in the gallery, else the earliest; exactly one in a non-empty gallery.
	Leads bool
}

var sqlGallery = `SELECT v.entity_id, v.rank, ao.reason, dv.image_file_id,
		COALESCE(f.media_type, ''), ao.observation_id, e.preferred_likeness_observation_id
	FROM auto_reconciler_values v
	JOIN canonical_entities e ON e.id = v.entity_id
	JOIN properties p ON p.id = v.property_id
	JOIN auto_reconciler_outcomes ao ON ao.entity_id = v.entity_id
		AND ao.property_id = v.property_id AND ao.value_rank = v.rank
	JOIN observations o ON o.id = ao.observation_id
	JOIN depiction_values dv ON dv.id = o.value_depiction_id
	LEFT JOIN files f ON f.id = dv.image_file_id
	WHERE p.origin = ? AND p.key IN (` + database.SQLInPlaceholders(len(vocabkey.GalleryKeys())) + `) AND v.entity_id IN (`

func galleryPrefixArgs() []any {
	keys := vocabkey.GalleryKeys()
	args := make([]any, 0, 1+len(keys))
	args = append(args, vocabkey.OriginProvenencia)
	for _, k := range keys {
		args = append(args, k)
	}
	return args
}

// ForEntities returns each handle's gallery, keyed by string(entity id),
// earliest first: by ObservationIDs[0], then Rank. A handle with none is absent.
func ForEntities(q sqlbatch.Querier, entityIDs [][]byte) (map[string][]Picture, error) {
	ids := database.UniqueBlobIDs(entityIDs)
	if len(ids) == 0 {
		return nil, nil
	}
	type key struct {
		entity string
		rank   int
	}
	byValue := map[key]*Picture{}
	var order []key
	pointers := map[string][]byte{}
	err := sqlbatch.Query(q, sqlGallery, ")", galleryPrefixArgs(), ids, func(rows *sql.Rows) error {
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
		image = depictionvalue.ReadableImage(image, mediaType)
		if image == nil {
			return nil
		}
		k := key{string(entity), rank}
		l, ok := byValue[k]
		if !ok {
			l = &Picture{Rank: rank, ImageFileID: append([]byte(nil), image...)}
			byValue[k] = l
			order = append(order, k)
		}
		l.ObservationIDs = append(l.ObservationIDs, append([]byte(nil), obs...))
		return nil
	})
	if err != nil {
		return nil, err
	}
	out := map[string][]Picture{}
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
		lead := 0
		if p, ok := pointers[entity]; ok {
			if i, ok := ByObservation(ls, p); ok {
				ls[i].Preferred = true
				lead = i
			}
		}
		ls[lead].Leads = true
	}
	return out, nil
}

// ByObservation returns the index of the gallery picture whose gallery records
// include the Observation: the one rule for which picture a record names. -1,
// false when none.
func ByObservation(pictures []Picture, observationID []byte) (i int, ok bool) {
	for i := range pictures {
		for _, id := range pictures[i].ObservationIDs {
			if bytes.Equal(id, observationID) {
				return i, true
			}
		}
	}
	return -1, false
}

// Lead returns the picture a handle leads with: its preferred portrait when
// that picture is in the gallery, else the earliest (working rule "A preferred
// portrait is a display pointer, not a claim"). ok is false when the gallery
// is empty.
func Lead(pictures []Picture) (l Picture, ok bool) {
	for _, g := range pictures {
		if g.Leads {
			return g, true
		}
	}
	return Picture{}, false
}
