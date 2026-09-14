package search

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/mendahu/provenencia/core/database"
)

// FTSSearcher retrieves hits from catalog_search_fts + catalog_search_docs.
type FTSSearcher struct{}

// NewFTSSearcher returns the default FTS-backed Searcher.
func NewFTSSearcher() *FTSSearcher {
	return &FTSSearcher{}
}

// Search implements Searcher.
func (f *FTSSearcher) Search(ctx context.Context, c *database.Catalog, q Query) ([]Hit, error) {
	_ = ctx
	tokens := tokenize(q.Text)
	if len(tokens) == 0 {
		return nil, nil
	}
	limit := q.Limit
	if limit <= 0 {
		limit = DefaultHitLimit
	}

	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	match := buildMatchQuery(tokens)
	if match == "" {
		return nil, nil
	}

	// bm25 column weights from FTSBM25Weights (lower bm25 = better).
	w := FTSBM25Weights
	rows, err := db.Query(fmt.Sprintf(`
		SELECT d.kind, d.entity_id, d.display_ref, d.display_title, d.display_subtitle,
			d.title, d.ref, d.secondary, d.body,
			bm25(catalog_search_fts, %f, %f, %f, %f) AS rank
		FROM catalog_search_fts
		JOIN catalog_search_docs d ON d.rowid = catalog_search_fts.rowid
		WHERE catalog_search_fts MATCH ?
	`, w.Title, w.Ref, w.Secondary, w.Body), match)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hits []Hit
	for rows.Next() {
		var kind, entityID, displayRef, displayTitle, displaySubtitle string
		var titleCol, refCol, secondaryCol, bodyCol string
		var rank float64
		if err := rows.Scan(
			&kind, &entityID, &displayRef, &displayTitle, &displaySubtitle,
			&titleCol, &refCol, &secondaryCol, &bodyCol, &rank,
		); err != nil {
			return nil, err
		}
		spec, ok := kindSpec(kind)
		if !ok || !spec.DefaultInEverything {
			continue
		}
		values := fieldValuesForKind(kind, titleCol, refCol, secondaryCol, bodyCol)
		score, reason := scoreFields(spec, values, tokens)
		if score <= 0 {
			// FTS matched (e.g. prefix) but scoreFields saw no substring — keep a floor.
			score = 0.1
			reason = "title"
		}
		// Blend registry weights with FTS rank (bm25: more negative ≈ stronger).
		score *= 1.0 / (1.0 + absFloat(rank))
		score *= contextMultiplier(spec, q.Location.Section)

		hits = append(hits, Hit{
			Kind:        kind,
			ID:          entityID,
			Ref:         displayRef,
			Title:       displayTitle,
			Subtitle:    displaySubtitle,
			MatchReason: reason,
			Location:    locationFor(kind, entityID, displayRef, displayTitle),
			Score:       score,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
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

func fieldValuesForKind(kind, title, ref, secondary, body string) map[string]string {
	switch kind {
	case KindSource:
		return map[string]string{
			"title":       title,
			"ref":         ref,
			"description": secondary,
			"notes":       body,
			"metadata":    body,
			"filename":    body,
		}
	case KindSourceType, KindSourceField:
		return map[string]string{
			"label":       title,
			"key":         secondary,
			"description": secondary,
		}
	default:
		return map[string]string{
			"title": title,
			"ref":   ref,
			"body":  body,
		}
	}
}

func locationFor(kind, id, ref, title string) WorkspaceLocation {
	switch kind {
	case KindSource:
		return WorkspaceLocation{Section: SectionSources, SourceID: id, Ref: ref, Title: title}
	case KindSourceType:
		return WorkspaceLocation{Section: SectionSourceTypes, TypeID: id, Title: title}
	case KindSourceField:
		return WorkspaceLocation{Section: SectionSourceFields, FieldID: id, Title: title}
	default:
		return WorkspaceLocation{}
	}
}

// buildMatchQuery ANDs tokens; last token gets a prefix star for typeahead.
func buildMatchQuery(tokens []string) string {
	if len(tokens) == 0 {
		return ""
	}
	parts := make([]string, 0, len(tokens))
	for i, tok := range tokens {
		tok = sanitizeFTSToken(tok)
		if tok == "" {
			continue
		}
		if i == len(tokens)-1 {
			parts = append(parts, tok+"*")
		} else {
			parts = append(parts, tok)
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " AND ")
}

func sanitizeFTSToken(tok string) string {
	tok = strings.TrimSpace(strings.ToLower(tok))
	replacer := strings.NewReplacer(
		`"`, "", `'`, "", `(`, "", `)`, "", `{`, "", `}`, "",
		`[`, "", `]`, "", `^`, "", `*`, "", `:`, "",
	)
	tok = replacer.Replace(tok)
	tok = strings.TrimSpace(tok)
	if tok == "" || tok == "and" || tok == "or" || tok == "not" {
		return ""
	}
	return tok
}

func absFloat(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
