package datevalue

// Calendar facts for civil dates: proleptic Gregorian day numbers, month
// lengths and the span a bound covers. They are facts, not tunables, so they
// live here, outside the matching numbers guard.

// DaysFromCivil is the proleptic Gregorian day number, 1970-01-01 = 0.
func DaysFromCivil(year, month, day int) int {
	a := (14 - month) / 12
	y := year + 4800 - a
	m := month + 12*a - 3
	jdn := day + (153*m+2)/5 + 365*y + y/4 - y/100 + y/400 - 32045
	return jdn - 2440588
}

// civilFromDays is the inverse of DaysFromCivil (Howard Hinnant's
// civil_from_days).
func civilFromDays(days int) (year, month, day int) {
	z := days + 719468
	era := z / 146097
	if z < 0 {
		era = (z - 146096) / 146097
	}
	doe := z - era*146097
	yoe := (doe - doe/1460 + doe/36524 - doe/146096) / 365
	year = yoe + era*400
	doy := doe - (365*yoe + yoe/4 - yoe/100)
	mp := (5*doy + 2) / 153
	day = doy - (153*mp+2)/5 + 1
	if mp < 10 {
		month = mp + 3
	} else {
		month = mp - 9
	}
	if month <= 2 {
		year++
	}
	return year, month, day
}

func isLeap(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// DaysInMonth is the number of days in a month of a year.
func DaysInMonth(year, month int) int {
	switch month {
	case 1, 3, 5, 7, 8, 10, 12:
		return 31
	case 4, 6, 9, 11:
		return 30
	default:
		if isLeap(year) {
			return 29
		}
		return 28
	}
}

// BoundDays is the inclusive span of day numbers one bound covers at its
// precision: a month names that month's days, no month the whole year (a day
// with no month is ignored). ok is false with no year, a month that does not
// exist, or a day outside its month.
func BoundDays(year, month, day *int) (lo, hi int, ok bool) {
	if year == nil {
		return 0, 0, false
	}
	y := *year
	if month == nil {
		return DaysFromCivil(y, 1, 1), DaysFromCivil(y, 12, 31), true
	}
	m := *month
	if !monthOK(m) {
		return 0, 0, false
	}
	if day == nil {
		return DaysFromCivil(y, m, 1), DaysFromCivil(y, m, DaysInMonth(y, m)), true
	}
	d := *day
	if d < 1 || d > DaysInMonth(y, m) {
		return 0, 0, false
	}
	n := DaysFromCivil(y, m, d)
	return n, n, true
}

// YearLine places the day span lo..hi, inside one civil year, on a line of
// years, where day d of year y covers [y + (d-1)/n, y + d/n) with n the days
// in y, so a whole year is [y, y+1). mid is the centre of the span.
func YearLine(lo, hi int) (from, to, mid float64) {
	y, _, _ := civilFromDays(lo)
	jan1 := DaysFromCivil(y, 1, 1)
	n := float64(DaysFromCivil(y+1, 1, 1) - jan1)
	from = float64(y) + float64(lo-jan1)/n
	to = float64(y) + float64(hi-jan1+1)/n
	return from, to, (from + to) / 2
}

// AddMonths moves a month of a year by n months, which may be negative.
func AddMonths(year, month, n int) (int, int) {
	total := year*12 + (month - 1) + n
	y, m := total/12, total%12
	if m < 0 {
		y, m = y-1, m+12
	}
	return y, m + 1
}

// AddDays moves a day by n days, which may be negative.
func AddDays(year, month, day, n int) (int, int, int) {
	return civilFromDays(DaysFromCivil(year, month, day) + n)
}

// AddYears moves a date by n civil years. A month or day of 0 (not given)
// stays 0; 29 February in a common year becomes 28 February.
func AddYears(year, month, day, n int) (int, int, int) {
	y := year + n
	if month != 0 && day != 0 {
		day = min(day, DaysInMonth(y, month))
	}
	return y, month, day
}
