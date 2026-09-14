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

// KindSpec is one searchable navigable root in the registry.
type KindSpec struct {
	Kind               string
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
		ContextBoost:        1.35,
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
		ContextBoost:        1.35,
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
		ContextBoost:        1.35,
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

// scoreFields returns a weighted score and the best-matching field name.
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
		reason = bestField
	}
	return score, reason
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
