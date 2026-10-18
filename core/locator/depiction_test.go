package locator

import (
	"errors"
	"testing"
)

func TestValidateDepictionRegion(t *testing.T) {
	const tri = `{"type":"region","unit":"normalized","points":[{"x":0.1,"y":0.1},{"x":0.4,"y":0.1},{"x":0.4,"y":0.5}]}`
	const whole = `{"type":"region","unit":"normalized","points":[{"x":0,"y":0},{"x":1,"y":0},{"x":1,"y":1},{"x":0,"y":1}]}`
	doc := func(sels ...string) string {
		out := `{"version":1,"selectors":[`
		for i, s := range sels {
			if i > 0 {
				out += ","
			}
			out += s
		}
		return out + `]}`
	}
	page := func(n string) string { return `{"type":"page","artifact_page":` + n + `}` }
	art := `{"type":"artifact"}`
	tests := []struct {
		name     string
		region   string
		citation string
		wantErr  bool
	}{
		{"page 1 both", doc(page("1"), tri), doc(page("1")), false},
		{"whole file, page 1 both", doc(page("1"), whole), doc(page("1")), false},
		{"whole file, no page", doc(whole), doc(art), false},
		{"unknown selector type on region", doc(page("1"), tri, `{"type":"mystery"}`), doc(page("1")), true},
		{"page_label not compared", doc(`{"type":"page","artifact_page":1,"page_label":"iv"}`, tri), doc(page("1")), false},
		{"artifact first on both", doc(art, page("1"), tri), doc(art, page("1")), false},
		{"region page 2", doc(page("2"), tri), doc(page("1")), true},
		{"region page, citation none", doc(page("1"), tri), doc(art), true},
		{"citation page, region none", doc(tri), doc(page("1")), true},
		{"no page on either", doc(tri), doc(art), false},
		{"two pages on region", doc(page("1"), page("1"), tri), doc(page("1")), true},
		{"two pages on citation", doc(page("1"), tri), doc(page("1"), page("1")), true},
		{"artifact then region, no pages", doc(art, tri), doc(art), false},
		{"no region selector", doc(page("1")), doc(page("1")), true},
		{"two region selectors", doc(page("1"), tri, tri), doc(page("1")), true},
		{"text_quote on region", doc(page("1"), tri, `{"type":"text_quote","exact":"x"}`), doc(page("1")), true},
		{"region over the point cap", doc(page("1"), ringRegion(MaxRegionPoints+1)), doc(page("1")), true},
		{"citation locator over the point cap", doc(page("1"), tri), doc(page("1"), ringRegion(MaxRegionPoints+1)), true},
		{"citation locator points over the cap across selectors", doc(page("1"), tri), doc(page("1"), ringRegion(129), ringRegion(128)), true},
		{"invalid region json", `{`, doc(page("1")), true},
		{"invalid citation json", doc(page("1"), tri), `{`, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateDepictionRegion(tc.region, tc.citation)
			if tc.wantErr {
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v, want ErrInvalid", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
