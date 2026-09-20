// Package locator validates Citation locator_json documents.
//
// Version 1 locators are an ordered selector chain. Known types validated in
// Spike 7: page, region, text_quote. Unknown selector types are preserved
// losslessly (Validate does not fail solely for an unrecognized type).
package locator

import (
	"encoding/json"
	"math"
	"strings"

	"github.com/mendahu/provenencia/core/apperr"
)

var ErrInvalid = apperr.New(apperr.CodeLocatorInvalid, apperr.KindUser)

const (
	TypePage      = "page"
	TypeRegion    = "region"
	TypeTextQuote = "text_quote"
)

type document struct {
	Version   int             `json:"version"`
	Selectors []selectorRaw   `json:"selectors"`
}

type selectorRaw struct {
	Type string `json:"type"`
	// page
	ArtifactPage *int    `json:"artifact_page"`
	PageLabel    *string `json:"page_label"`
	// region
	Points []point `json:"points"`
	Unit   string  `json:"unit"`
	// text_quote
	Exact  string `json:"exact"`
	Prefix string `json:"prefix"`
	Suffix string `json:"suffix"`
}

type point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Validate checks locator JSON against version-1 invariants for known selector types.
func Validate(locatorJSON string) error {
	locatorJSON = strings.TrimSpace(locatorJSON)
	if locatorJSON == "" {
		return ErrInvalid
	}
	var doc document
	if err := json.Unmarshal([]byte(locatorJSON), &doc); err != nil {
		return ErrInvalid
	}
	if doc.Version != 1 || len(doc.Selectors) == 0 {
		return ErrInvalid
	}
	for _, sel := range doc.Selectors {
		switch strings.TrimSpace(sel.Type) {
		case TypePage:
			if err := validatePage(sel); err != nil {
				return err
			}
		case TypeRegion:
			if err := validateRegion(sel); err != nil {
				return err
			}
		case TypeTextQuote:
			if err := validateTextQuote(sel); err != nil {
				return err
			}
		default:
			// Unknown types preserved; no further validation.
		}
	}
	return nil
}

func validatePage(sel selectorRaw) error {
	if sel.ArtifactPage == nil || *sel.ArtifactPage < 1 {
		return ErrInvalid
	}
	return nil
}

func validateTextQuote(sel selectorRaw) error {
	if strings.TrimSpace(sel.Exact) == "" {
		return ErrInvalid
	}
	return nil
}

func validateRegion(sel selectorRaw) error {
	if strings.TrimSpace(sel.Unit) != "normalized" {
		return ErrInvalid
	}
	pts := sel.Points
	if len(pts) < 3 {
		return ErrInvalid
	}
	// First point must not be repeated at the end.
	first, last := pts[0], pts[len(pts)-1]
	if nearlyEqual(first.X, last.X) && nearlyEqual(first.Y, last.Y) {
		return ErrInvalid
	}
	for _, p := range pts {
		if p.X < 0 || p.X > 1 || p.Y < 0 || p.Y > 1 {
			return ErrInvalid
		}
	}
	if polygonArea(pts) < 1e-12 {
		return ErrInvalid
	}
	if selfIntersects(pts) {
		return ErrInvalid
	}
	return nil
}

func nearlyEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-12
}

func polygonArea(pts []point) float64 {
	n := len(pts)
	var sum float64
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		sum += pts[i].X*pts[j].Y - pts[j].X*pts[i].Y
	}
	return math.Abs(sum) / 2
}

// selfIntersects reports whether any non-adjacent edges of the closed polygon intersect.
func selfIntersects(pts []point) bool {
	n := len(pts)
	for i := 0; i < n; i++ {
		a1, a2 := pts[i], pts[(i+1)%n]
		for j := i + 1; j < n; j++ {
			// Skip adjacent edges and the closing edge pair that shares a vertex.
			if j == i || (j+1)%n == i || (i+1)%n == j {
				continue
			}
			b1, b2 := pts[j], pts[(j+1)%n]
			if segmentsIntersect(a1, a2, b1, b2) {
				return true
			}
		}
	}
	return false
}

func segmentsIntersect(a1, a2, b1, b2 point) bool {
	d1 := cross(a2.X-a1.X, a2.Y-a1.Y, b1.X-a1.X, b1.Y-a1.Y)
	d2 := cross(a2.X-a1.X, a2.Y-a1.Y, b2.X-a1.X, b2.Y-a1.Y)
	d3 := cross(b2.X-b1.X, b2.Y-b1.Y, a1.X-b1.X, a1.Y-b1.Y)
	d4 := cross(b2.X-b1.X, b2.Y-b1.Y, a2.X-b1.X, a2.Y-b1.Y)
	if ((d1 > 0 && d2 < 0) || (d1 < 0 && d2 > 0)) &&
		((d3 > 0 && d4 < 0) || (d3 < 0 && d4 > 0)) {
		return true
	}
	return false
}

func cross(ax, ay, bx, by float64) float64 {
	return ax*by - ay*bx
}
