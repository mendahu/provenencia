package datevalues

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/database"
)

func intVal(n int) *int { return &n }

func TestInsertLookup(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, c *database.Catalog)
	}{
		{
			name: "point day round trip",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Insert(c, Value{
					Kind:       KindPoint,
					StartYear:  intVal(1985),
					StartMonth: intVal(5),
					StartDay:   intVal(14),
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := Lookup(c, id)
				if err != nil {
					t.Fatal(err)
				}
				if got.Kind != KindPoint || got.Qualifier != "" {
					t.Fatalf("got kind=%q qual=%q", got.Kind, got.Qualifier)
				}
				if got.StartYear == nil || *got.StartYear != 1985 ||
					got.StartMonth == nil || *got.StartMonth != 5 ||
					got.StartDay == nil || *got.StartDay != 14 {
					t.Fatalf("start %+v %+v %+v", got.StartYear, got.StartMonth, got.StartDay)
				}
				if got.StartHour != nil || got.EndYear != nil {
					t.Fatalf("unexpected time/end: %+v", got)
				}
			},
		},
		{
			name: "point year only",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Insert(c, Value{
					Kind:      KindPoint,
					StartYear: intVal(1985),
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := Lookup(c, id)
				if err != nil {
					t.Fatal(err)
				}
				if got.Kind != KindPoint || got.StartYear == nil || *got.StartYear != 1985 {
					t.Fatalf("got %+v", got)
				}
				if got.StartMonth != nil || got.StartDay != nil || got.StartHour != nil {
					t.Fatalf("finer want nil")
				}
			},
		},
		{
			name: "point month precision",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Insert(c, Value{
					Kind:       KindPoint,
					StartYear:  intVal(1985),
					StartMonth: intVal(5),
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := Lookup(c, id)
				if err != nil {
					t.Fatal(err)
				}
				if got.StartMonth == nil || *got.StartMonth != 5 || got.StartDay != nil {
					t.Fatalf("got %+v", got)
				}
			},
		},
		{
			name: "point with time through millisecond",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Insert(c, Value{
					Kind:             KindPoint,
					StartYear:        intVal(2020),
					StartMonth:       intVal(1),
					StartDay:         intVal(2),
					StartHour:        intVal(15),
					StartMinute:      intVal(30),
					StartSecond:      intVal(45),
					StartMillisecond: intVal(123),
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := Lookup(c, id)
				if err != nil {
					t.Fatal(err)
				}
				if got.StartHour == nil || *got.StartHour != 15 ||
					got.StartMinute == nil || *got.StartMinute != 30 ||
					got.StartSecond == nil || *got.StartSecond != 45 ||
					got.StartMillisecond == nil || *got.StartMillisecond != 123 {
					t.Fatalf("time %+v", got)
				}
			},
		},
		{
			name: "point with hour only",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Insert(c, Value{
					Kind:       KindPoint,
					StartYear:  intVal(2020),
					StartMonth: intVal(6),
					StartDay:   intVal(1),
					StartHour:  intVal(9),
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := Lookup(c, id)
				if err != nil {
					t.Fatal(err)
				}
				if got.StartHour == nil || *got.StartHour != 9 || got.StartMinute != nil {
					t.Fatalf("got %+v", got)
				}
			},
		},
		{
			name: "point with free-text timezone",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Insert(c, Value{
					Kind:        KindPoint,
					StartYear:   intVal(1985),
					StartMonth:  intVal(5),
					StartDay:    intVal(14),
					StartHour:   intVal(15),
					StartMinute: intVal(30),
					StartTZ:     "Eastern Standard Time",
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := Lookup(c, id)
				if err != nil {
					t.Fatal(err)
				}
				if got.StartTZ != "Eastern Standard Time" || got.EndTZ != "" {
					t.Fatalf("tz start=%q end=%q", got.StartTZ, got.EndTZ)
				}
			},
		},
		{
			name: "point phrase only",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Insert(c, Value{
					Kind:   KindPoint,
					Phrase: "Christmas",
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := Lookup(c, id)
				if err != nil {
					t.Fatal(err)
				}
				if got.Phrase != "Christmas" || got.StartYear != nil {
					t.Fatalf("got %+v", got)
				}
			},
		},
		{
			name: "point phrase with year",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Insert(c, Value{
					Kind:      KindPoint,
					StartYear: intVal(1887),
					Phrase:    "Christmas",
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := Lookup(c, id)
				if err != nil {
					t.Fatal(err)
				}
				if got.Phrase != "Christmas" || got.StartYear == nil || *got.StartYear != 1887 {
					t.Fatalf("got %+v", got)
				}
			},
		},
		{
			name: "range with distinct free-text timezones",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Insert(c, Value{
					Kind:       KindRange,
					StartYear:  intVal(2020),
					StartMonth: intVal(1),
					StartDay:   intVal(1),
					StartHour:  intVal(10),
					StartTZ:    "America/New_York",
					EndYear:    intVal(2020),
					EndMonth:   intVal(1),
					EndDay:     intVal(1),
					EndHour:    intVal(11),
					EndTZ:      "local mean time at York",
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := Lookup(c, id)
				if err != nil {
					t.Fatal(err)
				}
				if got.StartTZ != "America/New_York" || got.EndTZ != "local mean time at York" {
					t.Fatalf("tz start=%q end=%q", got.StartTZ, got.EndTZ)
				}
			},
		},
		{
			name: "rejects end_tz on point",
			run: func(t *testing.T, c *database.Catalog) {
				_, err := Insert(c, Value{
					Kind:       KindPoint,
					StartYear:  intVal(1985),
					StartMonth: intVal(5),
					StartDay:   intVal(14),
					EndTZ:      "UTC",
				})
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "rejects empty point without phrase",
			run: func(t *testing.T, c *database.Catalog) {
				_, err := Insert(c, Value{Kind: KindPoint})
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "rejects phrase-only point with start_tz",
			run: func(t *testing.T, c *database.Catalog) {
				_, err := Insert(c, Value{
					Kind:    KindPoint,
					Phrase:  "Christmas",
					StartTZ: "UTC",
				})
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "rejects legacy exact kind",
			run: func(t *testing.T, c *database.Catalog) {
				_, err := Insert(c, Value{
					Kind:       "exact",
					StartYear:  intVal(1985),
					StartMonth: intVal(5),
					StartDay:   intVal(14),
				})
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "rejects legacy year kind",
			run: func(t *testing.T, c *database.Catalog) {
				_, err := Insert(c, Value{
					Kind:      "year",
					StartYear: intVal(1985),
				})
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "point with ABT",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Insert(c, Value{
					Kind:      KindPoint,
					Qualifier: QualifierABT,
					StartYear: intVal(1890),
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := Lookup(c, id)
				if err != nil {
					t.Fatal(err)
				}
				if got.Qualifier != QualifierABT || got.Kind != KindPoint {
					t.Fatalf("got kind=%q qual=%q", got.Kind, got.Qualifier)
				}
			},
		},
		{
			name: "point day with ABT",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Insert(c, Value{
					Kind:       KindPoint,
					Qualifier:  QualifierABT,
					StartYear:  intVal(1985),
					StartMonth: intVal(5),
					StartDay:   intVal(14),
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := Lookup(c, id)
				if err != nil {
					t.Fatal(err)
				}
				if got.Qualifier != QualifierABT {
					t.Fatalf("qual %q", got.Qualifier)
				}
			},
		},
		{
			name: "point with BEF",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Insert(c, Value{
					Kind:      KindPoint,
					Qualifier: QualifierBEF,
					StartYear: intVal(1900),
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := Lookup(c, id)
				if err != nil {
					t.Fatal(err)
				}
				if got.Qualifier != QualifierBEF {
					t.Fatalf("qual %q", got.Qualifier)
				}
			},
		},
		{
			name: "point with AFT",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Insert(c, Value{
					Kind:       KindPoint,
					Qualifier:  QualifierAFT,
					StartYear:  intVal(1872),
					StartMonth: intVal(3),
					StartDay:   intVal(1),
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := Lookup(c, id)
				if err != nil {
					t.Fatal(err)
				}
				if got.Qualifier != QualifierAFT {
					t.Fatalf("qual %q", got.Qualifier)
				}
			},
		},
		{
			name: "year range round trip",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Insert(c, Value{
					Kind:      KindRange,
					StartYear: intVal(1880),
					EndYear:   intVal(1885),
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := Lookup(c, id)
				if err != nil {
					t.Fatal(err)
				}
				if got.Kind != KindRange ||
					got.StartYear == nil || *got.StartYear != 1880 ||
					got.EndYear == nil || *got.EndYear != 1885 {
					t.Fatalf("got %+v", got)
				}
			},
		},
		{
			name: "datetime range round trip",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Insert(c, Value{
					Kind:        KindRange,
					StartYear:   intVal(2020),
					StartMonth:  intVal(1),
					StartDay:    intVal(1),
					StartHour:   intVal(10),
					StartMinute: intVal(0),
					EndYear:     intVal(2020),
					EndMonth:    intVal(1),
					EndDay:      intVal(1),
					EndHour:     intVal(11),
					EndMinute:   intVal(30),
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := Lookup(c, id)
				if err != nil {
					t.Fatal(err)
				}
				if got.StartHour == nil || *got.StartHour != 10 ||
					got.EndMinute == nil || *got.EndMinute != 30 {
					t.Fatalf("got %+v", got)
				}
			},
		},
		{
			name: "month range round trip",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Insert(c, Value{
					Kind:       KindRange,
					StartYear:  intVal(1880),
					StartMonth: intVal(1),
					EndYear:    intVal(1880),
					EndMonth:   intVal(6),
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := Lookup(c, id)
				if err != nil {
					t.Fatal(err)
				}
				if got.StartMonth == nil || *got.StartMonth != 1 ||
					got.EndMonth == nil || *got.EndMonth != 6 {
					t.Fatalf("got %+v", got)
				}
			},
		},
		{
			name: "rejects unknown kind",
			run: func(t *testing.T, c *database.Catalog) {
				_, err := Insert(c, Value{Kind: "phrase", StartYear: intVal(1900)})
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "rejects ABT on range",
			run: func(t *testing.T, c *database.Catalog) {
				_, err := Insert(c, Value{
					Kind:      KindRange,
					Qualifier: QualifierABT,
					StartYear: intVal(1880),
					EndYear:   intVal(1885),
				})
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "rejects BEF on range",
			run: func(t *testing.T, c *database.Catalog) {
				_, err := Insert(c, Value{
					Kind:      KindRange,
					Qualifier: QualifierBEF,
					StartYear: intVal(1880),
					EndYear:   intVal(1885),
				})
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "rejects unknown qualifier",
			run: func(t *testing.T, c *database.Catalog) {
				_, err := Insert(c, Value{
					Kind:      KindPoint,
					Qualifier: "CIR",
					StartYear: intVal(1900),
				})
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "accepts minute without hour",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Insert(c, Value{
					Kind:        KindPoint,
					StartYear:   intVal(2020),
					StartMonth:  intVal(1),
					StartDay:    intVal(1),
					StartMinute: intVal(30),
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := Lookup(c, id)
				if err != nil {
					t.Fatal(err)
				}
				if got.StartMinute == nil || *got.StartMinute != 30 {
					t.Fatalf("got %+v", got)
				}
			},
		},
		{
			name: "accepts day without month",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Insert(c, Value{
					Kind:      KindPoint,
					StartYear: intVal(1985),
					StartDay:  intVal(14),
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := Lookup(c, id)
				if err != nil {
					t.Fatal(err)
				}
				if got.StartDay == nil || *got.StartDay != 14 || got.StartMonth != nil {
					t.Fatalf("got %+v", got)
				}
			},
		},
		{
			name: "accepts hour without year",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Insert(c, Value{
					Kind:      KindPoint,
					StartHour: intVal(12),
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := Lookup(c, id)
				if err != nil {
					t.Fatal(err)
				}
				if got.StartHour == nil || *got.StartHour != 12 || got.StartYear != nil {
					t.Fatalf("got %+v", got)
				}
			},
		},
		{
			name: "accepts month only",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Insert(c, Value{
					Kind:       KindPoint,
					StartMonth: intVal(5),
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := Lookup(c, id)
				if err != nil {
					t.Fatal(err)
				}
				if got.StartMonth == nil || *got.StartMonth != 5 {
					t.Fatalf("got %+v", got)
				}
			},
		},
		{
			name: "rejects range end before start",
			run: func(t *testing.T, c *database.Catalog) {
				_, err := Insert(c, Value{
					Kind:      KindRange,
					StartYear: intVal(1885),
					EndYear:   intVal(1880),
				})
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "rejects range end time before start time",
			run: func(t *testing.T, c *database.Catalog) {
				_, err := Insert(c, Value{
					Kind:       KindRange,
					StartYear:  intVal(2020),
					StartMonth: intVal(1),
					StartDay:   intVal(1),
					StartHour:  intVal(12),
					EndYear:    intVal(2020),
					EndMonth:   intVal(1),
					EndDay:     intVal(1),
					EndHour:    intVal(11),
				})
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "lookup rejects short id",
			run: func(t *testing.T, c *database.Catalog) {
				_, err := Lookup(c, []byte{1})
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "lookup missing row",
			run: func(t *testing.T, c *database.Catalog) {
				id := make([]byte, 16)
				_, err := Lookup(c, id)
				if !errors.Is(err, sql.ErrNoRows) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "rejects closed catalog",
			run: func(t *testing.T, c *database.Catalog) {
				c.Close()
				_, err := Insert(c, Value{Kind: KindPoint, StartYear: intVal(1900)})
				if !errors.Is(err, database.ErrClosed) {
					t.Fatalf("got %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := database.Create(t.TempDir(), "t.provenencia")
			if err != nil {
				t.Fatal(err)
			}
			defer c.Close()
			tt.run(t, c)
		})
	}
}
