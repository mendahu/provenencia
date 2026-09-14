package search

import (
	"context"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/sourcefields"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
)

// NaiveScanner lists live catalog tables and scores in memory (S3-07 bridge).
type NaiveScanner struct{}

// NewNaiveScanner returns the default table-scan Searcher.
func NewNaiveScanner() *NaiveScanner {
	return &NaiveScanner{}
}

// Search implements Searcher.
func (n *NaiveScanner) Search(ctx context.Context, c *database.Catalog, q Query) ([]Hit, error) {
	_ = ctx
	tokens := tokenize(q.Text)
	if len(tokens) == 0 {
		return nil, nil
	}
	limit := q.Limit
	if limit <= 0 {
		limit = DefaultHitLimit
	}

	var hits []Hit

	srcRows, err := sources.List(c)
	if err != nil {
		return nil, err
	}
	typeByID := map[string]sourcetypes.Type{}
	typeRows, err := sourcetypes.List(c)
	if err != nil {
		return nil, err
	}
	for _, t := range typeRows {
		typeByID[uuidString(t.ID)] = t
	}

	spec, _ := kindSpec(KindSource)
	for _, s := range srcRows {
		id := uuidString(s.ID)
		typeLabel := ""
		if t, ok := typeByID[uuidString(s.SourceTypeID)]; ok {
			typeLabel = t.Label
		}
		values := map[string]string{
			"title":       s.Title,
			"ref":         s.Ref,
			"description": s.Description,
		}
		// Type label participates at description weight for naïve recall.
		if typeLabel != "" {
			values["description"] = strings.TrimSpace(s.Description + " " + typeLabel)
		}
		score, reason := scoreFields(spec, values, tokens)
		if score <= 0 {
			continue
		}
		score *= contextMultiplier(spec, q.Location.Section)
		subtitle := typeLabel
		hits = append(hits, Hit{
			Kind:        KindSource,
			ID:          id,
			Ref:         s.Ref,
			Title:       s.Title,
			Subtitle:    subtitle,
			MatchReason: reason,
			Location: WorkspaceLocation{
				Section:  SectionSources,
				SourceID: id,
				Ref:      s.Ref,
				Title:    s.Title,
			},
			Score: score,
		})
	}

	spec, _ = kindSpec(KindSourceType)
	for _, t := range typeRows {
		id := uuidString(t.ID)
		values := map[string]string{
			"label":       t.Label,
			"key":         t.Key,
			"description": t.Description,
		}
		score, reason := scoreFields(spec, values, tokens)
		if score <= 0 {
			continue
		}
		score *= contextMultiplier(spec, q.Location.Section)
		hits = append(hits, Hit{
			Kind:        KindSourceType,
			ID:          id,
			Title:       t.Label,
			Subtitle:    t.Key,
			MatchReason: reason,
			Location: WorkspaceLocation{
				Section: SectionSourceTypes,
				TypeID:  id,
				Title:   t.Label,
			},
			Score: score,
		})
	}

	fieldRows, err := sourcefields.List(c)
	if err != nil {
		return nil, err
	}
	spec, _ = kindSpec(KindSourceField)
	for _, f := range fieldRows {
		id := uuidString(f.ID)
		values := map[string]string{
			"label":       f.Label,
			"key":         f.Key,
			"description": f.Description,
		}
		score, reason := scoreFields(spec, values, tokens)
		if score <= 0 {
			continue
		}
		score *= contextMultiplier(spec, q.Location.Section)
		hits = append(hits, Hit{
			Kind:        KindSourceField,
			ID:          id,
			Title:       f.Label,
			Subtitle:    f.Key,
			MatchReason: reason,
			Location: WorkspaceLocation{
				Section: SectionSourceFields,
				FieldID: id,
				Title:   f.Label,
			},
			Score: score,
		})
	}

	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].Score != hits[j].Score {
			return hits[i].Score > hits[j].Score
		}
		if hits[i].Kind != hits[j].Kind {
			return hits[i].Kind < hits[j].Kind
		}
		return hits[i].Title < hits[j].Title
	})
	if len(hits) > limit {
		hits = hits[:limit]
	}
	return hits, nil
}

func uuidString(id []byte) string {
	if len(id) != 16 {
		return ""
	}
	var u uuid.UUID
	copy(u[:], id)
	return u.String()
}
