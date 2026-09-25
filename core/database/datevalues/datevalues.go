// Package datevalues stores shared genealogical DateValue rows.
//
// Kinds validated by Insert (DDL stays flexible for later kinds):
//
//   - kind "point": one date; precision = whichever start_* components the
//     evidence asserts (gaps allowed); optional start_tz; qualifier
//     "" / "ABT" / "BEF" / "AFT"; no end_* (including end_tz). Phrase-only
//     points are allowed when no civil components are set.
//   - kind "range": one date in a bounded uncertainty window (BET); each side
//     needs at least one civil component; optional start_tz / end_tz;
//     start <= end when shared fields can be compared; qualifier empty only
//
// Missing fields mean unknown/not asserted, not midnight. Year is not
// required when a month, day, or clock field is set.
//
// Qualifiers on point: ABT ≈ about; BEF = before / no later than;
// AFT = after / no earlier than. The bound is the start_* civil components.
// Between is kind "range", not a qualifier.
//
// start_tz / end_tz are free-text zone labels as stated (IANA id, offset,
// historical name, or "local time"). Empty means unspecified. They are not
// parsed into UTC offsets in this package.
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
	KindPoint = "point"
	KindRange = "range"

	QualifierABT = "ABT" // about / approximately
	QualifierBEF = "BEF" // before / no later than start_*
	QualifierAFT = "AFT" // after / no earlier than start_*

	sqlInsert = `INSERT INTO date_values (
		id, kind, qualifier, calendar,
		start_year, start_month, start_day,
		start_hour, start_minute, start_second, start_millisecond, start_tz,
		end_year, end_month, end_day,
		end_hour, end_minute, end_second, end_millisecond, end_tz,
		phrase
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	sqlUpdate = `UPDATE date_values SET
		kind = ?, qualifier = ?, calendar = ?,
		start_year = ?, start_month = ?, start_day = ?,
		start_hour = ?, start_minute = ?, start_second = ?, start_millisecond = ?, start_tz = ?,
		end_year = ?, end_month = ?, end_day = ?,
		end_hour = ?, end_minute = ?, end_second = ?, end_millisecond = ?, end_tz = ?,
		phrase = ?
		WHERE id = ?`

	sqlExists = `SELECT 1 FROM date_values WHERE id = ?`

	sqlLookup = `SELECT kind, qualifier, calendar,
		start_year, start_month, start_day,
		start_hour, start_minute, start_second, start_millisecond, start_tz,
		end_year, end_month, end_day,
		end_hour, end_minute, end_second, end_millisecond, end_tz,
		phrase
		FROM date_values WHERE id = ?`
)

// Value is one date_values row.
type Value struct {
	ID               []byte
	Kind             string
	Qualifier        string // "" | ABT | BEF | AFT on point kinds
	Calendar         string
	StartYear        *int
	StartMonth       *int
	StartDay         *int
	StartHour        *int
	StartMinute      *int
	StartSecond      *int
	StartMillisecond *int
	StartTZ          string // free-text zone; empty = unspecified
	EndYear          *int
	EndMonth         *int
	EndDay           *int
	EndHour          *int
	EndMinute        *int
	EndSecond        *int
	EndMillisecond   *int
	EndTZ            string // free-text zone; empty = unspecified
	Phrase           string
}

type side struct {
	year, month, day     *int
	hour, minute, second *int
	millisecond          *int
}

// Insert mints a UUIDv7 id, validates, inserts, and returns the id.
func Insert(c *database.Catalog, v Value) ([]byte, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	return InsertTx(db, v)
}

// InsertTx validates and inserts on an existing connection or transaction.
func InsertTx(q interface {
	Exec(query string, args ...any) (sql.Result, error)
}, v Value) ([]byte, error) {
	v.Kind = strings.TrimSpace(v.Kind)
	v.Qualifier = strings.TrimSpace(v.Qualifier)
	v.Calendar = strings.TrimSpace(v.Calendar)
	v.Phrase = strings.TrimSpace(v.Phrase)
	v.StartTZ = strings.TrimSpace(v.StartTZ)
	v.EndTZ = strings.TrimSpace(v.EndTZ)
	if err := validate(v); err != nil {
		return nil, err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	_, err = q.Exec(
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
		nullIfEmpty(v.StartTZ),
		nullInt(v.EndYear),
		nullInt(v.EndMonth),
		nullInt(v.EndDay),
		nullInt(v.EndHour),
		nullInt(v.EndMinute),
		nullInt(v.EndSecond),
		nullInt(v.EndMillisecond),
		nullIfEmpty(v.EndTZ),
		nullIfEmpty(v.Phrase),
	)
	if err != nil {
		return nil, err
	}
	return id[:], nil
}

// UpdateTx rewrites the date_values row id. The id stays the same.
func UpdateTx(tx *sql.Tx, id []byte, v Value) error {
	if tx == nil || len(id) != 16 {
		return ErrInvalid
	}
	v.Kind = strings.TrimSpace(v.Kind)
	v.Qualifier = strings.TrimSpace(v.Qualifier)
	v.Calendar = strings.TrimSpace(v.Calendar)
	v.Phrase = strings.TrimSpace(v.Phrase)
	v.StartTZ = strings.TrimSpace(v.StartTZ)
	v.EndTZ = strings.TrimSpace(v.EndTZ)
	if err := validate(v); err != nil {
		return err
	}
	var one int
	if err := tx.QueryRow(sqlExists, id).Scan(&one); err != nil {
		if err == sql.ErrNoRows {
			return ErrInvalid
		}
		return err
	}
	_, err := tx.Exec(
		sqlUpdate,
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
		nullIfEmpty(v.StartTZ),
		nullInt(v.EndYear),
		nullInt(v.EndMonth),
		nullInt(v.EndDay),
		nullInt(v.EndHour),
		nullInt(v.EndMinute),
		nullInt(v.EndSecond),
		nullInt(v.EndMillisecond),
		nullIfEmpty(v.EndTZ),
		nullIfEmpty(v.Phrase),
		id,
	)
	return err
}

// Lookup returns the date_values row for id, or sql.ErrNoRows.
func Lookup(c *database.Catalog, id []byte) (Value, error) {
	db, err := c.DB()
	if err != nil {
		return Value{}, err
	}
	return LookupTx(db, id)
}

// LookupTx returns the date_values row for id on an existing connection or transaction.
func LookupTx(q interface {
	QueryRow(query string, args ...any) *sql.Row
}, id []byte) (Value, error) {
	if q == nil || len(id) != 16 {
		return Value{}, ErrInvalid
	}
	var (
		v                                                         Value
		qual, cal, phrase, startTZ, endTZ                         sql.NullString
		startY, startM, startD, startH, startMin, startS, startMs sql.NullInt64
		endY, endM, endD, endH, endMin, endS, endMs               sql.NullInt64
	)
	v.ID = append([]byte(nil), id...)
	err := q.QueryRow(sqlLookup, id).Scan(
		&v.Kind,
		&qual,
		&cal,
		&startY, &startM, &startD,
		&startH, &startMin, &startS, &startMs, &startTZ,
		&endY, &endM, &endD,
		&endH, &endMin, &endS, &endMs, &endTZ,
		&phrase,
	)
	if err != nil {
		return Value{}, err
	}
	v.Qualifier = qual.String
	v.Calendar = cal.String
	v.Phrase = phrase.String
	v.StartTZ = startTZ.String
	v.EndTZ = endTZ.String
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
	case KindPoint:
		if !pointQualifierOK(qual) {
			return ErrInvalid
		}
		if !end.empty() || v.EndTZ != "" {
			return ErrInvalid
		}
		if !hasCivil(start) {
			// Phrase-forward point: no year/month/day/time. Timezone or
			// millisecond-only is not enough.
			if v.Phrase == "" || v.StartTZ != "" || start.millisecond != nil {
				return ErrInvalid
			}
			return nil
		}
		return validateComponentRanges(start)
	case KindRange:
		if qual != "" {
			return ErrInvalid
		}
		if err := validateComponentRanges(start); err != nil {
			return err
		}
		if err := validateComponentRanges(end); err != nil {
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

func pointQualifierOK(qual string) bool {
	switch qual {
	case "", QualifierABT, QualifierBEF, QualifierAFT:
		return true
	default:
		return false
	}
}

func hasCivil(s side) bool {
	return s.year != nil || s.month != nil || s.day != nil ||
		s.hour != nil || s.minute != nil || s.second != nil
}

// validateComponentRanges requires at least one civil field and in-range
// values. Gaps (day without month, hour without year) are allowed.
func validateComponentRanges(s side) error {
	if !hasCivil(s) {
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
