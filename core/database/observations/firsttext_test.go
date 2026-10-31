package observations_test

import (
	"testing"

	"github.com/mendahu/provenencia/core/database/observations"
)

func TestFirstText(t *testing.T) {
	pos, neg := observations.PolarityPositive, observations.PolarityNegative
	tests := []struct {
		name string
		in   []observations.TextRecord
		want string
	}{
		{"none", nil, ""},
		{"one", []observations.TextRecord{{"OBS-B", pos, "York"}}, "York"},
		{"negative skipped", []observations.TextRecord{{"OBS-A", neg, "Leeds"}, {"OBS-B", pos, "York"}}, "York"},
		{"blank skipped", []observations.TextRecord{{"OBS-A", pos, "  \t"}, {"OBS-B", pos, " York "}}, "York"},
		{"lowest ref wins", []observations.TextRecord{{"OBS-C", pos, "Hull"}, {"OBS-B", pos, "York"}}, "York"},
		{"refs compare without case", []observations.TextRecord{{"obs-c", pos, "Hull"}, {"OBS-D", pos, "York"}}, "Hull"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := observations.FirstText(tc.in); got != tc.want {
				t.Fatalf("FirstText = %q, want %q", got, tc.want)
			}
		})
	}
}
