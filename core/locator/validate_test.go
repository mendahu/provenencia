package locator

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"
)

// ringPoints returns n points on a circle of radius 0.4 around (0.5, 0.5),
// counter-clockwise from angle 0.
func ringPoints(n int) []point {
	pts := make([]point, n)
	for i := range pts {
		a := 2 * math.Pi * float64(i) / float64(n)
		pts[i] = point{X: 0.5 + 0.4*math.Cos(a), Y: 0.5 + 0.4*math.Sin(a)}
	}
	return pts
}

// ringRegion returns a region selector object holding an n-point ring.
func ringRegion(n int) string {
	var b strings.Builder
	b.WriteString(`{"type":"region","unit":"normalized","points":[`)
	for i, p := range ringPoints(n) {
		if i > 0 {
			b.WriteString(",")
		}
		fmt.Fprintf(&b, `{"x":%s,"y":%s}`, strconv.FormatFloat(p.X, 'f', 6, 64), strconv.FormatFloat(p.Y, 'f', 6, 64))
	}
	b.WriteString(`]}`)
	return b.String()
}

// paddedArtifactDoc returns a valid artifact-only document of n bytes, padded
// with spaces inside the JSON.
func paddedArtifactDoc(n int) string {
	const base = `{"version":1,"selectors":[{"type":"artifact"}]}`
	return `{"version":1,"selectors":[{"type":"artifact"}` + strings.Repeat(" ", n-len(base)) + `]}`
}

func TestValidateDocCapBeforeCrossing(t *testing.T) {
	doc := document{Version: 1, Selectors: []selectorRaw{{Type: TypeRegion, Unit: "normalized", Points: ringPoints(200_000)}}}
	start := time.Now()
	err := validateDoc(doc)
	elapsed := time.Since(start)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("got %v, want ErrInvalid", err)
	}
	if elapsed > 100*time.Millisecond {
		t.Fatalf("took %v, want under 100ms", elapsed)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name: "artifact only",
			json: `{"version":1,"selectors":[{"type":"artifact"}]}`,
		},
		{
			name: "artifact then page then region",
			json: `{"version":1,"selectors":[
				{"type":"artifact"},
				{"type":"page","artifact_page":2},
				{"type":"region","unit":"normalized","points":[
					{"x":0.1,"y":0.1},{"x":0.4,"y":0.1},{"x":0.4,"y":0.3},{"x":0.1,"y":0.3}
				]}
			]}`,
		},
		{
			name: "page only",
			json: `{"version":1,"selectors":[{"type":"page","artifact_page":37,"page_label":"23"}]}`,
		},
		{
			name: "region rectangle",
			json: `{"version":1,"selectors":[{"type":"region","unit":"normalized","points":[
				{"x":0.1,"y":0.1},{"x":0.5,"y":0.1},{"x":0.5,"y":0.4},{"x":0.1,"y":0.4}
			]}]}`,
		},
		{
			name: "page then region",
			json: `{"version":1,"selectors":[
				{"type":"page","artifact_page":1},
				{"type":"region","unit":"normalized","points":[
					{"x":0.2,"y":0.2},{"x":0.8,"y":0.2},{"x":0.8,"y":0.8},{"x":0.2,"y":0.8}
				]}
			]}`,
		},
		{
			name: "text_quote",
			json: `{"version":1,"selectors":[{"type":"text_quote","exact":"William Robins","prefix":"of ","suffix":" aged"}]}`,
		},
		{
			name: "unknown type preserved",
			json: `{"version":1,"selectors":[{"type":"time_range","start_ms":0,"end_ms":10}]}`,
		},
		{
			name:    "empty",
			json:    "",
			wantErr: true,
		},
		{
			name:    "bad json",
			json:    `{`,
			wantErr: true,
		},
		{
			name:    "wrong version",
			json:    `{"version":2,"selectors":[{"type":"page","artifact_page":1}]}`,
			wantErr: true,
		},
		{
			name:    "no selectors",
			json:    `{"version":1,"selectors":[]}`,
			wantErr: true,
		},
		{
			name:    "artifact not first",
			json:    `{"version":1,"selectors":[{"type":"page","artifact_page":1},{"type":"artifact"}]}`,
			wantErr: true,
		},
		{
			name:    "page zero",
			json:    `{"version":1,"selectors":[{"type":"page","artifact_page":0}]}`,
			wantErr: true,
		},
		{
			name:    "text_quote blank exact",
			json:    `{"version":1,"selectors":[{"type":"text_quote","exact":"  "}]}`,
			wantErr: true,
		},
		{
			name: "region too few points",
			json: `{"version":1,"selectors":[{"type":"region","unit":"normalized","points":[
				{"x":0.1,"y":0.1},{"x":0.5,"y":0.1}
			]}]}`,
			wantErr: true,
		},
		{
			name: "region wrong unit",
			json: `{"version":1,"selectors":[{"type":"region","unit":"pixels","points":[
				{"x":0.1,"y":0.1},{"x":0.5,"y":0.1},{"x":0.5,"y":0.4}
			]}]}`,
			wantErr: true,
		},
		{
			name: "region closed ring repeated",
			json: `{"version":1,"selectors":[{"type":"region","unit":"normalized","points":[
				{"x":0.1,"y":0.1},{"x":0.5,"y":0.1},{"x":0.5,"y":0.4},{"x":0.1,"y":0.1}
			]}]}`,
			wantErr: true,
		},
		{
			name: "region out of bounds",
			json: `{"version":1,"selectors":[{"type":"region","unit":"normalized","points":[
				{"x":0.1,"y":0.1},{"x":1.5,"y":0.1},{"x":0.5,"y":0.4}
			]}]}`,
			wantErr: true,
		},
		{
			name: "region zero area",
			json: `{"version":1,"selectors":[{"type":"region","unit":"normalized","points":[
				{"x":0.1,"y":0.1},{"x":0.5,"y":0.1},{"x":0.9,"y":0.1}
			]}]}`,
			wantErr: true,
		},
		{
			name: "region at the point cap",
			json: fmt.Sprintf(`{"version":1,"selectors":[%s]}`, ringRegion(MaxRegionPoints)),
		},
		{
			name:    "region over the point cap",
			json:    fmt.Sprintf(`{"version":1,"selectors":[%s]}`, ringRegion(MaxRegionPoints+1)),
			wantErr: true,
		},
		{
			name:    "points over the cap across two region selectors",
			json:    fmt.Sprintf(`{"version":1,"selectors":[{"type":"page","artifact_page":1},%s,%s]}`, ringRegion(129), ringRegion(128)),
			wantErr: true,
		},
		{
			name: "document at the byte cap",
			json: paddedArtifactDoc(MaxDocumentBytes),
		},
		{
			name:    "document over the byte cap",
			json:    paddedArtifactDoc(MaxDocumentBytes + 1),
			wantErr: true,
		},
		{
			name: "region self intersecting bowtie",
			json: `{"version":1,"selectors":[{"type":"region","unit":"normalized","points":[
				{"x":0.1,"y":0.1},{"x":0.9,"y":0.9},{"x":0.1,"y":0.9},{"x":0.9,"y":0.1}
			]}]}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.json)
			if tt.wantErr {
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
