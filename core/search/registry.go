package search

import "strings"

// FieldWeight relative importance for FTS / registry scoring.
type FieldWeight struct {
	Name   string
	Weight float64
}

// FTSBM25Weights are global bm25 column weights for catalog_search_fts, in
// table column order: title, ref, secondary, body. Shared across kinds;
// per-kind field importance still comes from KindSpec.Fields via scoreFields.
var FTSBM25Weights = struct {
	Title     float64
	Ref       float64
	Secondary float64
	Body      float64
}{
	Title:     10,
	Ref:       12,
	Secondary: 4,
	Body:      1,
}

// RefBoostWeights multiply a hit's score when the query matches display_ref
// via the ref fast path (exact ≫ prefix ≫ FTS-only / no ref match).
// Section/kind boosts stay on KindSpec.ContextBoost.
var RefBoostWeights = struct {
	Exact  float64
	Prefix float64
	None   float64
}{
	Exact:  8,
	Prefix: 3,
	None:   1,
}

// FuzzyWeights tune the S3-11 typo shortlist pass (trigram OR + Jaro–Winkler).
// Never full-catalog fuzzy scan — only a capped trigram candidate set.
var FuzzyWeights = struct {
	// MinTokenRunes skips short tokens (trigrams need ≥3; typos need room).
	MinTokenRunes int
	// CandidateCap limits rows merged from the trigram index per Search.
	CandidateCap int
	// JaroWinklerMin is the similarity gate against title/ref/label/key.
	JaroWinklerMin float64
	// ScoreScale multiplies fuzzy-only hits so exact FTS still ranks higher.
	ScoreScale float64
}{
	MinTokenRunes:  4,
	CandidateCap:   150,
	JaroWinklerMin: 0.86,
	ScoreScale:     0.55,
}

// KindSpec is one searchable navigable root in the registry.
type KindSpec struct {
	Kind                string
	DefaultInEverything bool
	// ContextSections that boost this kind when Query.Location.Section matches.
	ContextSections []string
	ContextBoost    float64
	Fields          []FieldWeight
}

// Registry is the declarative searchable-kind table for Spike 3+.
var Registry = []KindSpec{
	{
		Kind:                KindSource,
		DefaultInEverything: true,
		ContextSections:     []string{SectionSources},
		ContextBoost:        2.0,
		Fields: []FieldWeight{
			{Name: "title", Weight: 10},
			{Name: "ref", Weight: 12},
			{Name: "description", Weight: 3},
			{Name: "notes", Weight: 1.5},
			{Name: "metadata", Weight: 1.5},
			{Name: "filename", Weight: 1},
		},
	},
	{
		Kind:                KindSourceType,
		DefaultInEverything: true,
		ContextSections:     []string{SectionSourceTypes},
		ContextBoost:        2.0,
		Fields: []FieldWeight{
			{Name: "label", Weight: 10},
			{Name: "key", Weight: 8},
			{Name: "description", Weight: 3},
		},
	},
	{
		Kind:                KindSourceField,
		DefaultInEverything: true,
		ContextSections:     []string{SectionSourceFields},
		ContextBoost:        2.0,
		Fields: []FieldWeight{
			{Name: "label", Weight: 10},
			{Name: "key", Weight: 8},
			{Name: "description", Weight: 3},
		},
	},
}

func kindSpec(kind string) (KindSpec, bool) {
	for _, s := range Registry {
		if s.Kind == kind {
			return s, true
		}
	}
	return KindSpec{}, false
}

func contextMultiplier(spec KindSpec, section string) float64 {
	section = strings.TrimSpace(section)
	if section == "" || spec.ContextBoost <= 1 {
		return 1
	}
	for _, s := range spec.ContextSections {
		if s == section {
			return spec.ContextBoost
		}
	}
	return 1
}

// scoreFields returns a weighted score and the best-matching field reason.
// For rolled-up Source body fields, reason is a cheap snippet (note: …).
func scoreFields(spec KindSpec, values map[string]string, tokens []string) (score float64, reason string) {
	if len(tokens) == 0 {
		return 0, ""
	}
	var bestField string
	var bestWeight float64
	matchedTokens := 0
	for _, tok := range tokens {
		tokHit := false
		for _, f := range spec.Fields {
			v := strings.ToLower(values[f.Name])
			if v == "" {
				continue
			}
			if strings.Contains(v, tok) {
				tokHit = true
				score += f.Weight
				if f.Weight > bestWeight || (f.Weight == bestWeight && bestField == "") {
					bestWeight = f.Weight
					bestField = f.Name
				}
			}
		}
		if tokHit {
			matchedTokens++
		}
	}
	if matchedTokens == 0 {
		return 0, ""
	}
	// Prefer fuller term coverage.
	score *= float64(matchedTokens) / float64(len(tokens))
	if bestField != "" {
		reason = matchReasonForField(bestField, values[bestField], tokens)
	}
	return score, reason
}

func matchReasonForField(field, value string, tokens []string) string {
	switch field {
	case "notes":
		return "note: " + snippetAround(value, tokens, 48)
	case "metadata":
		return "metadata: " + snippetAround(value, tokens, 48)
	case "filename":
		return "filename: " + snippetAround(value, tokens, 48)
	default:
		return field
	}
}

func snippetAround(value string, tokens []string, maxLen int) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	lower := strings.ToLower(value)
	idx := -1
	tokLen := 0
	for _, tok := range tokens {
		if i := strings.Index(lower, tok); i >= 0 {
			idx = i
			tokLen = len(tok)
			break
		}
	}
	if idx < 0 {
		if len(value) <= maxLen {
			return value
		}
		return value[:maxLen] + "…"
	}
	start := idx - 8
	if start < 0 {
		start = 0
	}
	end := idx + tokLen + 24
	if end > len(value) {
		end = len(value)
	}
	out := value[start:end]
	if start > 0 {
		out = "…" + out
	}
	if end < len(value) {
		out += "…"
	}
	if len(out) > maxLen+1 {
		out = out[:maxLen] + "…"
	}
	return out
}

func tokenize(q string) []string {
	q = strings.ToLower(strings.TrimSpace(q))
	if q == "" {
		return nil
	}
	parts := strings.Fields(q)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.Trim(p, `"'.,;:!?()[]{}`)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
