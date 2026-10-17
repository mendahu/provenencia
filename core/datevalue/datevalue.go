// Package datevalue is the pure DateValue type: the value, its validation and
// its compact display. It imports no database package; queries live in
// core/database/datevalues.
//
// Kinds validated by Validate (DDL stays flexible for later kinds):
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
package datevalue

import (
	"github.com/mendahu/provenencia/core/apperr"
)

var ErrInvalid = apperr.New(apperr.CodeDateValuesInvalid, apperr.KindUser)

const (
	KindPoint = "point"
	KindRange = "range"

	QualifierABT = "ABT" // about / approximately
	QualifierBEF = "BEF" // before / no later than start_*
	QualifierAFT = "AFT" // after / no earlier than start_*
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

// Validate reports ErrInvalid when v is not a storable DateValue. Callers
// trim string fields first.
func Validate(v Value) error {
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
