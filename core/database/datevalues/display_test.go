package datevalues_test

import (
	"testing"

	"github.com/mendahu/provenencia/core/database/datevalues"
)

func TestCompactDisplay(t *testing.T) {
	y, m, d := 1882, 4, 3
	tests := []struct {
		name string
		v    datevalues.Value
		want string
	}{
		{
			name: "year only",
			v:    datevalues.Value{Kind: datevalues.KindPoint, StartYear: &y},
			want: "1882",
		},
		{
			name: "about month",
			v: datevalues.Value{
				Kind: datevalues.KindPoint, Qualifier: datevalues.QualifierABT,
				StartYear: &y, StartMonth: &m,
			},
			want: "About 1882-04",
		},
		{
			name: "full day",
			v: datevalues.Value{
				Kind: datevalues.KindPoint,
				StartYear: &y, StartMonth: &m, StartDay: &d,
			},
			want: "1882-04-03",
		},
		{
			name: "month only",
			v:    datevalues.Value{Kind: datevalues.KindPoint, StartMonth: &m},
			want: "--04",
		},
		{
			name: "day without month",
			v:    datevalues.Value{Kind: datevalues.KindPoint, StartYear: &y, StartDay: &d},
			want: "1882--03",
		},
		{
			name: "phrase only",
			v:    datevalues.Value{Kind: datevalues.KindPoint, Phrase: "Christmas"},
			want: "Christmas",
		},
		{
			name: "range",
			v: datevalues.Value{
				Kind: datevalues.KindRange,
				StartYear: &y, EndYear: &y, EndMonth: &m, EndDay: &d,
			},
			want: "Between 1882 and 1882-04-03",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := datevalues.CompactDisplay(tt.v); got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}
