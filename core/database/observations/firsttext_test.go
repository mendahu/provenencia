package observations_test

import (
	"testing"

	"github.com/mendahu/provenencia/core/database/graphcache"
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

func TestFirstTextOf(t *testing.T) {
	toponym, name := []byte{1}, []byte{2}
	pos, neg := observations.PolarityPositive, observations.PolarityNegative
	obs := func(prop []byte, ref, polarity, text string) graphcache.EvidenceObservation {
		o := graphcache.EvidenceObservation{PropertyID: prop, Ref: ref, Polarity: polarity}
		o.Text, o.HasText = text, true
		return o
	}
	tests := []struct {
		name string
		in   []graphcache.EvidenceObservation
		want string
	}{
		{"lowest positive ref of the Property", []graphcache.EvidenceObservation{
			obs(toponym, "OBS-B", pos, "York"), obs(toponym, "OBS-A", pos, "Eboracum"), obs(toponym, "OBS-0", neg, "Leeds"),
			obs(toponym, "OBS-00", pos, "  "), obs(name, "OBS-000", pos, "Jorvik"),
		}, "Eboracum"},
		{"no toponym", []graphcache.EvidenceObservation{obs(name, "OBS-A", pos, "Jorvik")}, ""},
		{"no Observations", nil, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := observations.FirstTextOf(tc.in, toponym); got != tc.want {
				t.Fatalf("FirstTextOf = %q, want %q", got, tc.want)
			}
		})
	}
}
