// Package datevalues stores shared genealogical DateValue rows.
//
// Dogfood kinds validated by Insert (DDL stays flexible for later kinds):
//
//   - kind "exact": calendar day required (Y/M/D); optional cascading time
//     (hour→minute→second→millisecond); qualifier "" or "ABT"; no end_*
//   - kind "year":  start_year only; qualifier "" or "ABT"
//   - kind "range": start and end each cascading from year through optional
//     time; start <= end; qualifier empty only
//
// Finer components require all coarser ones (no gaps). Missing time means
// unknown/not asserted, not midnight. Timezone is not modeled yet.
package datevalues

import (
	"database/sql"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
)

var ErrInvalid = apperr.New(apperr.CodeDateValuesInvalid, apperr.KindUser)

const (
	KindExact = "exact"
	KindYear  = "year"
	KindRange = "range"

	QualifierABT = "ABT"

	sqlInsert = `INSERT INTO date_values (
		id, kind, qualifier, calendar,
		start_year, start_month, start_day,
		start_hour, start_minute, start_second, start_millisecond,
		end_year, end_month, end_day,
		end_hour, end_minute, end_second, end_millisecond,
		phrase
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	sqlLookup = `SELECT kind, qualifier, calendar,
		start_year, start_month, start_day,
		start_hour, start_minute, start_second, start_millisecond,
		end_year, end_month, end_day,
		end_hour, end_minute, end_second, end_millisecond,
		phrase
		FROM date_values WHERE id = ?`
)

// Value is one date_values row.
type Value struct {
	ID               []byte
	Kind             string
	Qualifier        string // "" or QualifierABT for dogfood kinds
	Calendar         string
	StartYear        *int
	StartMonth       *int
	StartDay         *int
	StartHour        *int
	StartMinute      *int
	StartSecond      *int
	StartMillisecond *int
	EndYear          *int
	EndMonth         *int
	EndDay           *int
	EndHour          *int
	EndMinute        *int
	EndSecond        *int
	EndMillisecond   *int
	Phrase           string
}

type side struct {
	year, month, day     *int
	hour, minute, second *int
	millisecond          *int
}

// Insert mints a UUIDv7 id, validates the dogfood subset, inserts, and returns the id.
func Insert(c *database.Catalog, v Value) ([]byte, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	v.Kind = strings.TrimSpace(v.Kind)
	v.Qualifier = strings.TrimSpace(v.Qualifier)
	v.Calendar = strings.TrimSpace(v.Calendar)
	v.Phrase = strings.TrimSpace(v.Phrase)
	if err := validate(v); err != nil {
		return nil, err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(
		sqlInsert,
		id[:],
		v.Kind,
		nullIfEmpty(v.Qualifier),
		nullIfEmpty(v.Calendar),
		nullInt(v.StartYear),
		nullInt(v.StartMonth),
		nullInt(v.StartDay),
		nullInt(v.StartHour),
		nullInt(v.StartMinute),
		nullInt(v.StartSecond),
		nullInt(v.StartMillisecond),
		nullInt(v.EndYear),
		nullInt(v.EndMonth),
		nullInt(v.EndDay),
		nullInt(v.EndHour),
		nullInt(v.EndMinute),
		nullInt(v.EndSecond),
		nullInt(v.EndMillisecond),
		nullIfEmpty(v.Phrase),
	)
	if err != nil {
		return nil, err
	}
	return id[:], nil
}

// Lookup returns the date_values row for id, or sql.ErrNoRows.
func Lookup(c *database.Catalog, id []byte) (Value, error) {
	db, err := c.DB()
	if err != nil {
		return Value{}, err
	}
	if len(id) != 16 {
		return Value{}, ErrInvalid
	}
	var (
		v                                                         Value
		qual, cal, phrase                                         sql.NullString
		startY, startM, startD, startH, startMin, startS, startMs sql.NullInt64
		endY, endM, endD, endH, endMin, endS, endMs               sql.NullInt64
	)
	v.ID = append([]byte(nil), id...)
	err = db.QueryRow(sqlLookup, id).Scan(
		&v.Kind,
		&qual,
		&cal,
		&startY, &startM, &startD,
		&startH, &startMin, &startS, &startMs,
		&endY, &endM, &endD,
		&endH, &endMin, &endS, &endMs,
		&phrase,
	)
	if err != nil {
		return Value{}, err
	}
	v.Qualifier = qual.String
	v.Calendar = cal.String
	v.Phrase = phrase.String
	v.StartYear = intPtr(startY)
	v.StartMonth = intPtr(startM)
	v.StartDay = intPtr(startD)
	v.StartHour = intPtr(startH)
	v.StartMinute = intPtr(startMin)
	v.StartSecond = intPtr(startS)
	v.StartMillisecond = intPtr(startMs)
	v.EndYear = intPtr(endY)
	v.EndMonth = intPtr(endM)
	v.EndDay = intPtr(endD)
	v.EndHour = intPtr(endH)
	v.EndMinute = intPtr(endMin)
	v.EndSecond = intPtr(endS)
	v.EndMillisecond = intPtr(endMs)
	return v, nil
}

func validate(v Value) error {
	start := side{
		year: v.StartYear, month: v.StartMonth, day: v.StartDay,
		hour: v.StartHour, minute: v.StartMinute, second: v.StartSecond,
		millisecond: v.StartMillisecond,
	}
	end := side{
		year: v.EndYear, month: v.EndMonth, day: v.EndDay,
		hour: v.EndHour, minute: v.EndMinute, second: v.EndSecond,
		millisecond: v.EndMillisecond,
	}
	qual := v.Qualifier

	switch v.Kind {
	case KindExact:
		if qual != "" && qual != QualifierABT {
			return ErrInvalid
		}
		if !end.empty() {
			return ErrInvalid
		}
		// Calendar day required; optional cascading time.
		if err := validateCascade(start, true); err != nil {
			return err
		}
	case KindYear:
		if qual != "" && qual != QualifierABT {
			return ErrInvalid
		}
		if !end.empty() {
			return ErrInvalid
		}
		if start.year == nil || start.month != nil || start.day != nil ||
			start.hour != nil || start.minute != nil || start.second != nil || start.millisecond != nil {
			return ErrInvalid
		}
	case KindRange:
		if qual != "" {
			return ErrInvalid
		}
		if err := validateCascade(start, false); err != nil {
			return err
		}
		if err := validateCascade(end, false); err != nil {
			return err
		}
		if !sideLessOrEqual(start, end) {
			return ErrInvalid
		}
	default:
		return ErrInvalid
	}
	return nil
}

// validateCascade requires year, optional finer fields with no gaps, and
// valid ranges. If requireDay, month and day must be present (exact points).
func validateCascade(s side, requireDay bool) error {
	levels := []*int{s.year, s.month, s.day, s.hour, s.minute, s.second, s.millisecond}
	if levels[0] == nil {
		return ErrInvalid
	}
	seenNil := false
	for _, p := range levels {
		if p == nil {
			seenNil = true
			continue
		}
		if seenNil {
			return ErrInvalid
		}
	}
	if requireDay && (s.month == nil || s.day == nil) {
		return ErrInvalid
	}
	if s.month != nil && !monthOK(*s.month) {
		return ErrInvalid
	}
	if s.day != nil && !dayOK(*s.day) {
		return ErrInvalid
	}
	if s.hour != nil && (*s.hour < 0 || *s.hour > 23) {
		return ErrInvalid
	}
	if s.minute != nil && (*s.minute < 0 || *s.minute > 59) {
		return ErrInvalid
	}
	if s.second != nil && (*s.second < 0 || *s.second > 59) {
		return ErrInvalid
	}
	if s.millisecond != nil && (*s.millisecond < 0 || *s.millisecond > 999) {
		return ErrInvalid
	}
	return nil
}

func (s side) empty() bool {
	return s.year == nil && s.month == nil && s.day == nil &&
		s.hour == nil && s.minute == nil && s.second == nil && s.millisecond == nil
}

// sideLessOrEqual compares cascading components; a missing field on either
// side stops the comparison (treat as equal at that precision).
func sideLessOrEqual(a, b side) bool {
	levelsA := []*int{a.year, a.month, a.day, a.hour, a.minute, a.second, a.millisecond}
	levelsB := []*int{b.year, b.month, b.day, b.hour, b.minute, b.second, b.millisecond}
	for i := range levelsA {
		if levelsA[i] == nil || levelsB[i] == nil {
			return true
		}
		if *levelsA[i] < *levelsB[i] {
			return true
		}
		if *levelsA[i] > *levelsB[i] {
			return false
		}
	}
	return true
}

func monthOK(m int) bool { return m >= 1 && m <= 12 }
func dayOK(d int) bool   { return d >= 1 && d <= 31 }

func nullIfEmpty(s string) any {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return s
}

func nullInt(p *int) any {
	if p == nil {
		return nil
	}
	return *p
}

func intPtr(n sql.NullInt64) *int {
	if !n.Valid {
		return nil
	}
	v := int(n.Int64)
	return &v
}
