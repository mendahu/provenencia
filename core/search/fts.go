package search

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/searchindex"
)

// FTSSearcher retrieves hits from catalog_search_fts + catalog_search_docs.
type FTSSearcher struct{}

// NewFTSSearcher returns the default FTS-backed Searcher.
func NewFTSSearcher() *FTSSearcher {
	return &FTSSearcher{}
}

type docRow struct {
	kind, entityID, displayRef, displayTitle, displaySubtitle string
	displayIconKey, displayThumbnailRelPath                   string
	title, ref, secondary, body                               string
	ftsRank                                                   float64
	hasFTS                                                    bool
	fromFuzzy                                                 bool
	fuzzySim                                                  float64
	fuzzyReason                                               string
	refMatch                                                  refMatchTier
}

func candidateKey(kind, entityID string) string {
	return kind + "\x00" + entityID
}

func errIfCancelled(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}

// Search implements Searcher.
func (f *FTSSearcher) Search(ctx context.Context, c *database.Catalog, q Query) ([]Hit, error) {
	if err := errIfCancelled(ctx); err != nil {
		return nil, err
	}
	raw := strings.TrimSpace(q.Text)
	tokens := tokenize(raw)
	limit := q.Limit
	if limit <= 0 {
		limit = DefaultHitLimit
	}

	db, err := c.DB()
	if err != nil {
		return nil, err
	}

	kinds := effectiveKinds(q.Kinds)
	if len(kinds) == 0 {
		return nil, nil
	}
	byKey := make(map[string]docRow)

	// Ref fast path: exact / prefix on projected docs (does not depend on FTS tokenization).
	if key, exact := classifyRefQuery(raw); key != "" {
		refDocs, err := lookUpRefDocs(ctx, db, key, exact, limit, kinds)
		if err != nil {
			return nil, err
		}
		for _, d := range refDocs {
			byKey[candidateKey(d.kind, d.entityID)] = d
		}
	}

	if len(tokens) > 0 {
		// Ref-shaped queries skip FTS: hyphens are FTS operators and break MATCH.
		if refKey, _ := classifyRefQuery(raw); refKey == "" {
			andMatch := buildMatchQuery(tokens, false)
			if andMatch != "" {
				if err := mergeFTS(ctx, db, byKey, andMatch, kinds); err != nil {
					return nil, err
				}
			}
			// Controlled OR for multi-token queries so partial term coverage can retrieve.
			if len(tokens) >= 2 {
				orMatch := buildMatchQuery(tokens, true)
				if orMatch != "" {
					if err := mergeFTS(ctx, db, byKey, orMatch, kinds); err != nil {
						return nil, err
					}
				}
			}
			// Typo shortlist: trigram OR expansion + Jaro–Winkler gate (never full scan).
			if len(byKey) < limit {
				if err := mergeFuzzyShortlist(ctx, db, byKey, tokens, limit, kinds); err != nil {
					return nil, err
				}
			}
		}
	}

	if len(byKey) == 0 {
		return nil, nil
	}

	// Score tokens: for pure ref queries, still tokenize the raw string so
	// scoreFields can credit the ref field when present.
	scoreTokens := tokens
	if len(scoreTokens) == 0 {
		scoreTokens = tokenize(strings.ToLower(raw))
	}

	var hits []Hit
	for _, d := range byKey {
		if err := errIfCancelled(ctx); err != nil {
			return nil, err
		}
		spec, ok := kindSpec(d.kind)
		if !ok || !containsKind(kinds, d.kind) {
			continue
		}
		values := fieldValuesForKind(d.kind, d.title, d.ref, d.secondary, d.body)
		score, reason, snippet := scoreFields(spec, values, scoreTokens)
		if d.refMatch == refMatchExact || d.refMatch == refMatchPrefix {
			reason = "ref"
			snippet = ""
			if score <= 0 {
				score = 12 // ref field weight floor
			}
		} else if field, line := alternateMatch(d.secondary, scoreTokens); reason == "other" && field != "" {
			reason = field
			snippet = line
		} else if score <= 0 && d.fromFuzzy && d.fuzzySim >= FuzzyWeights.JaroWinklerMin {
			// Substring scorer missed (typo); keep JW-gated fuzzy hit below exact FTS.
			score = d.fuzzySim * 12 * FuzzyWeights.ScoreScale
			reason = d.fuzzyReason
			snippet = ""
			if reason == "" {
				reason = "fuzzy"
			}
		} else if score <= 0 {
			// FTS matched (e.g. prefix) but scoreFields saw no substring — keep a floor.
			score = 0.1
			reason = "title"
			snippet = ""
		} else if d.fromFuzzy && !d.hasFTS {
			score *= FuzzyWeights.ScoreScale
		}
		if d.hasFTS {
			// bm25 is negative, and a more negative value is a stronger match.
			// Scale the field score up by that strength. Dividing by it ranked
			// the tightest hit last.
			score *= 1 + absFloat(d.ftsRank)
		}
		score *= contextMultiplier(spec, q.Location.Section)
		score *= refBoostFor(d.refMatch)
		score = applyScoreMix(score, d.kind)

		hits = append(hits, Hit{
			Kind:             d.kind,
			ID:               d.entityID,
			Ref:              d.displayRef,
			Title:            d.displayTitle,
			Subtitle:         d.displaySubtitle,
			MatchReason:      reason,
			MatchSnippet:     snippet,
			Location:         locationFor(d.kind, d.entityID, d.displayRef, d.displayTitle),
			ThumbnailRelPath: d.displayThumbnailRelPath,
			IconKey:          d.displayIconKey,
			Score:            score,
		})
	}

	sortHits(hits)
	if len(hits) > limit {
		hits = hits[:limit]
	}
	if err := fillMemberCounts(ctx, db, hits); err != nil {
		return nil, err
	}
	return hits, nil
}

// sortHits orders by the mixed score, then kind name, then title.
func sortHits(hits []Hit) {
	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].Score != hits[j].Score {
			return hits[i].Score > hits[j].Score
		}
		if hits[i].Kind != hits[j].Kind {
			return hits[i].Kind < hits[j].Kind
		}
		return hits[i].Title < hits[j].Title
	})
}

// applyScoreMix keeps ScoreMix.Text of the text score and adds up to
// ScoreMix.Priority for the kind's registry priority.
func applyScoreMix(text float64, kind string) float64 {
	fraction := 0.0
	if spec, ok := kindSpec(kind); ok {
		if max := maxKindPriority(); max > 0 {
			fraction = float64(spec.Priority) / float64(max)
		}
	}
	return text * (ScoreMix.Text + ScoreMix.Priority*fraction)
}

func maxKindPriority() int {
	max := 0
	for _, spec := range Registry {
		if spec.Priority > max {
			max = spec.Priority
		}
	}
	return max
}

func kindSQL(kinds []string) string {
	clause, _ := kindClause("d.kind", kinds)
	return clause
}

func kindArgs(kinds []string) []any {
	_, args := kindClause("d.kind", kinds)
	return args
}

func containsKind(kinds []string, kind string) bool {
	for _, k := range kinds {
		if k == kind {
			return true
		}
	}
	return false
}

func isHandleKind(kind string) bool {
	return kind == KindPerson || kind == KindEvent || kind == KindPlace
}

// fillMemberCounts sets MemberCount on handle hits: one query for the page.
func fillMemberCounts(ctx context.Context, db *sql.DB, hits []Hit) error {
	var ids []any
	index := map[string][]int{}
	for i, h := range hits {
		if !isHandleKind(h.Kind) {
			continue
		}
		u, err := uuid.Parse(h.ID)
		if err != nil {
			continue
		}
		if _, seen := index[string(u[:])]; !seen {
			ids = append(ids, u[:])
		}
		index[string(u[:])] = append(index[string(u[:])], i)
	}
	if len(ids) == 0 {
		return nil
	}
	marks := strings.TrimSuffix(strings.Repeat("?, ", len(ids)), ", ")
	rows, err := db.QueryContext(ctx, `SELECT entity_id, COUNT(*) FROM identity_claims
		WHERE status = 'accepted' AND entity_id IN (`+marks+`) GROUP BY entity_id`, ids...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			id []byte
			n  int
		)
		if err := rows.Scan(&id, &n); err != nil {
			return err
		}
		for _, i := range index[string(id)] {
			hits[i].MemberCount = n
		}
	}
	return rows.Err()
}

func mergeFTS(ctx context.Context, db *sql.DB, byKey map[string]docRow, match string, kinds []string) error {
	if err := errIfCancelled(ctx); err != nil {
		return err
	}
	w := FTSBM25Weights
	rows, err := db.QueryContext(ctx, fmt.Sprintf(`
		SELECT d.kind, d.entity_id, d.display_ref, d.display_title, d.display_subtitle,
			d.display_icon_key, d.display_thumbnail_rel_path,
			d.title, d.ref, d.secondary, d.body,
			bm25(catalog_search_fts, %f, %f, %f, %f) AS rank
		FROM catalog_search_fts
		JOIN catalog_search_docs d ON d.rowid = catalog_search_fts.rowid
		WHERE catalog_search_fts MATCH ?%s
	`, w.Title, w.Ref, w.Secondary, w.Body, kindSQL(kinds)), append([]any{match}, kindArgs(kinds)...)...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		if err := errIfCancelled(ctx); err != nil {
			return err
		}
		var d docRow
		if err := rows.Scan(
			&d.kind, &d.entityID, &d.displayRef, &d.displayTitle, &d.displaySubtitle,
			&d.displayIconKey, &d.displayThumbnailRelPath,
			&d.title, &d.ref, &d.secondary, &d.body, &d.ftsRank,
		); err != nil {
			return err
		}
		d.hasFTS = true
		key := candidateKey(d.kind, d.entityID)
		if prev, ok := byKey[key]; ok {
			d.refMatch = betterRefMatch(prev.refMatch, d.refMatch)
			// bm25: more negative is stronger — keep the better rank across AND/OR passes.
			if prev.hasFTS && prev.ftsRank < d.ftsRank {
				d.ftsRank = prev.ftsRank
			}
		}
		byKey[key] = d
	}
	return rows.Err()
}

// mergeFuzzyShortlist expands candidates via trigram OR MATCH, then keeps only
// rows that pass a Jaro–Winkler gate against identity fields.
func mergeFuzzyShortlist(ctx context.Context, db *sql.DB, byKey map[string]docRow, tokens []string, limit int, kinds []string) error {
	remaining := FuzzyWeights.CandidateCap
	if remaining <= 0 {
		return nil
	}
	for _, tok := range tokens {
		if err := errIfCancelled(ctx); err != nil {
			return err
		}
		if len(byKey) >= limit || remaining <= 0 {
			break
		}
		match := buildTrigramORMatch(tok)
		if match == "" {
			continue
		}
		added, err := mergeTrigramCandidates(ctx, db, byKey, match, remaining, tokens, kinds)
		if err != nil {
			return err
		}
		remaining -= added
	}
	return nil
}

func mergeTrigramCandidates(ctx context.Context, db *sql.DB, byKey map[string]docRow, match string, capN int, tokens []string, kinds []string) (added int, err error) {
	if err := errIfCancelled(ctx); err != nil {
		return 0, err
	}
	rows, err := db.QueryContext(ctx, `
		SELECT d.kind, d.entity_id, d.display_ref, d.display_title, d.display_subtitle,
			d.display_icon_key, d.display_thumbnail_rel_path,
			d.title, d.ref, d.secondary, d.body
		FROM catalog_search_fts_trigram
		JOIN catalog_search_docs d ON d.rowid = catalog_search_fts_trigram.rowid
		WHERE catalog_search_fts_trigram MATCH ?`+kindSQL(kinds)+`
		LIMIT ?
	`, append(append([]any{match}, kindArgs(kinds)...), capN)...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := errIfCancelled(ctx); err != nil {
			return added, err
		}
		var d docRow
		if err := rows.Scan(
			&d.kind, &d.entityID, &d.displayRef, &d.displayTitle, &d.displaySubtitle,
			&d.displayIconKey, &d.displayThumbnailRelPath,
			&d.title, &d.ref, &d.secondary, &d.body,
		); err != nil {
			return added, err
		}
		key := candidateKey(d.kind, d.entityID)
		if prev, ok := byKey[key]; ok {
			// Already retrieved via unicode FTS or ref — keep prior row.
			_ = prev
			continue
		}
		values := fieldValuesForKind(d.kind, d.title, d.ref, d.secondary, d.body)
		sim, reason := bestFuzzyFieldScore(values, tokens)
		if sim < FuzzyWeights.JaroWinklerMin {
			continue
		}
		d.fromFuzzy = true
		d.fuzzySim = sim
		d.fuzzyReason = reason
		byKey[key] = d
		added++
		if added >= capN {
			break
		}
	}
	return added, rows.Err()
}

func fieldValuesForKind(kind, title, refCol, secondary, body string) map[string]string {
	switch kind {
	case KindSource:
		notes, metadata, filename := splitTaggedBody(body)
		return map[string]string{
			"title":       title,
			"ref":         refCol,
			"description": secondary,
			"notes":       notes,
			"metadata":    metadata,
			"filename":    filename,
		}
	case KindSourceType, KindMetadataField:
		return map[string]string{
			"label":       title,
			"key":         secondary,
			"description": secondary,
		}
	case KindPerson, KindEvent, KindPlace:
		return map[string]string{
			"title": title,
			"ref":   refCol,
			"other": secondary,
		}
	default:
		return map[string]string{
			"title": title,
			"ref":   refCol,
			"body":  body,
		}
	}
}

func splitTaggedBody(body string) (notes, metadata, filename string) {
	var noteParts, metaParts, fileParts []string
	for _, line := range strings.Split(body, "\n") {
		switch {
		case strings.HasPrefix(line, searchindex.BodyTagNote):
			noteParts = append(noteParts, strings.TrimPrefix(line, searchindex.BodyTagNote))
		case strings.HasPrefix(line, searchindex.BodyTagMetadata):
			metaParts = append(metaParts, strings.TrimPrefix(line, searchindex.BodyTagMetadata))
		case strings.HasPrefix(line, searchindex.BodyTagFilename):
			fileParts = append(fileParts, strings.TrimPrefix(line, searchindex.BodyTagFilename))
		default:
			// Legacy untagged body (pre-ProjectionVersion 2): treat as notes.
			if t := strings.TrimSpace(line); t != "" {
				noteParts = append(noteParts, t)
			}
		}
	}
	return strings.Join(noteParts, "\n"), strings.Join(metaParts, "\n"), strings.Join(fileParts, "\n")
}

func locationFor(kind, id, refCol, title string) WorkspaceLocation {
	switch kind {
	case KindSource:
		return WorkspaceLocation{Section: SectionSources, SourceID: id, Ref: refCol, Title: title}
	case KindSourceType:
		return WorkspaceLocation{Section: SectionSourceTypes, TypeID: id, Title: title}
	case KindMetadataField:
		return WorkspaceLocation{Section: SectionMetadata, FieldID: id, Title: title}
	case KindPerson:
		return WorkspaceLocation{Section: SectionPersons, EntityID: id, Ref: refCol, Title: title}
	case KindEvent:
		return WorkspaceLocation{Section: SectionEvents, EntityID: id, Ref: refCol, Title: title}
	case KindPlace:
		return WorkspaceLocation{Section: SectionPlaces, EntityID: id, Ref: refCol, Title: title}
	default:
		return WorkspaceLocation{}
	}
}

// buildMatchQuery joins tokens with AND or OR; last token gets a prefix star for typeahead.
func buildMatchQuery(tokens []string, orJoin bool) string {
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
			parts = append(parts, `"`+tok+`"*`)
		} else {
			parts = append(parts, `"`+tok+`"`)
		}
	}
	if len(parts) == 0 {
		return ""
	}
	sep := " AND "
	if orJoin {
		sep = " OR "
	}
	return strings.Join(parts, sep)
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
