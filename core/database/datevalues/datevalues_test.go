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
			name: "exact day round trip",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Insert(c, Value{
					Kind:       KindExact,
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
				if got.Kind != KindExact || got.Qualifier != "" {
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
			name: "exact with time through millisecond",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Insert(c, Value{
					Kind:             KindExact,
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
			name: "exact with hour only",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Insert(c, Value{
					Kind:       KindExact,
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
			name: "year round trip",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Insert(c, Value{
					Kind:      KindYear,
					StartYear: intVal(1985),
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := Lookup(c, id)
				if err != nil {
					t.Fatal(err)
				}
				if got.Kind != KindYear || got.StartYear == nil || *got.StartYear != 1985 {
					t.Fatalf("got %+v", got)
				}
				if got.StartMonth != nil || got.StartDay != nil || got.StartHour != nil {
					t.Fatalf("finer want nil")
				}
			},
		},
		{
			name: "year with ABT",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Insert(c, Value{
					Kind:      KindYear,
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
				if got.Qualifier != QualifierABT || got.Kind != KindYear {
					t.Fatalf("got kind=%q qual=%q", got.Kind, got.Qualifier)
				}
			},
		},
		{
			name: "exact with ABT",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Insert(c, Value{
					Kind:       KindExact,
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
			name: "rejects exact missing day",
			run: func(t *testing.T, c *database.Catalog) {
				_, err := Insert(c, Value{
					Kind:       KindExact,
					StartYear:  intVal(1985),
					StartMonth: intVal(5),
				})
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "rejects minute without hour",
			run: func(t *testing.T, c *database.Catalog) {
				_, err := Insert(c, Value{
					Kind:        KindExact,
					StartYear:   intVal(2020),
					StartMonth:  intVal(1),
					StartDay:    intVal(1),
					StartMinute: intVal(30),
				})
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "rejects year with month set",
			run: func(t *testing.T, c *database.Catalog) {
				_, err := Insert(c, Value{
					Kind:       KindYear,
					StartYear:  intVal(1985),
					StartMonth: intVal(5),
				})
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "rejects year with hour set",
			run: func(t *testing.T, c *database.Catalog) {
				_, err := Insert(c, Value{
					Kind:      KindYear,
					StartYear: intVal(1985),
					StartHour: intVal(12),
				})
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
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
				_, err := Insert(c, Value{Kind: KindYear, StartYear: intVal(1900)})
				if !errors.Is(err, database.ErrClosed) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "create migrates date_values to user_version 4",
			run: func(t *testing.T, c *database.Catalog) {
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				var ver int
				if err := db.QueryRow(`PRAGMA user_version`).Scan(&ver); err != nil {
					t.Fatal(err)
				}
				if ver != 4 {
					t.Fatalf("user_version %d want 4", ver)
				}
				var name string
				err = db.QueryRow(
					`SELECT name FROM sqlite_master WHERE type='table' AND name='date_values'`,
				).Scan(&name)
				if err != nil {
					t.Fatal(err)
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
