package locator

import (
	"errors"
	"testing"
)

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
