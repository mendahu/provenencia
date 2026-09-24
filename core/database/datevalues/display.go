package datevalues

import (
	"fmt"
	"strings"
)

// CompactDisplay returns a portable, locale-independent summary for list/card
// rows when the client has not yet formatted with a locale calendar.
// Prefer client-side locale formatting (e.g. macOS DateValueDisplay) when the
// full Value is available on the wire.
func CompactDisplay(v Value) string {
	phrase := strings.TrimSpace(v.Phrase)
	switch strings.TrimSpace(v.Kind) {
	case KindPoint:
		point := formatSide(v.StartYear, v.StartMonth, v.StartDay, v.StartHour, v.StartMinute, v.StartSecond)
		if point == "" {
			return phrase
		}
		qualified := applyQualifier(v.Qualifier, point)
		return appendPhrase(phrase, qualified)
	case KindRange:
		start := formatSide(v.StartYear, v.StartMonth, v.StartDay, v.StartHour, v.StartMinute, v.StartSecond)
		end := formatSide(v.EndYear, v.EndMonth, v.EndDay, v.EndHour, v.EndMinute, v.EndSecond)
		if start == "" || end == "" {
			return phrase
		}
		between := fmt.Sprintf("Between %s and %s", start, end)
		return appendPhrase(phrase, between)
	default:
		return phrase
	}
}

func appendPhrase(phrase, structured string) string {
	if phrase == "" {
		return structured
	}
	if structured == "" {
		return phrase
	}
	return structured + " · " + phrase
}

func applyQualifier(qual, point string) string {
	switch strings.TrimSpace(qual) {
	case QualifierABT:
		return "About " + point
	case QualifierBEF:
		return "Before " + point
	case QualifierAFT:
		return "After " + point
	default:
		return point
	}
}

func formatSide(year, month, day, hour, minute, second *int) string {
	var date string
	switch {
	case year != nil && month != nil && day != nil:
		date = fmt.Sprintf("%d-%02d-%02d", *year, *month, *day)
	case year != nil && month != nil:
		date = fmt.Sprintf("%d-%02d", *year, *month)
	case year != nil && day != nil:
		date = fmt.Sprintf("%d--%02d", *year, *day)
	case year != nil:
		date = fmt.Sprintf("%d", *year)
	case month != nil && day != nil:
		date = fmt.Sprintf("--%02d-%02d", *month, *day)
	case month != nil:
		date = fmt.Sprintf("--%02d", *month)
	case day != nil:
		date = fmt.Sprintf("---%02d", *day)
	}

	var clock string
	switch {
	case hour != nil && minute != nil && second != nil:
		clock = fmt.Sprintf("%02d:%02d:%02d", *hour, *minute, *second)
	case hour != nil && minute != nil:
		clock = fmt.Sprintf("%02d:%02d", *hour, *minute)
	case hour != nil && second != nil:
		clock = fmt.Sprintf("%02d::%02d", *hour, *second)
	case hour != nil:
		clock = fmt.Sprintf("%02d", *hour)
	case minute != nil && second != nil:
		clock = fmt.Sprintf(":%02d:%02d", *minute, *second)
	case minute != nil:
		clock = fmt.Sprintf(":%02d", *minute)
	case second != nil:
		clock = fmt.Sprintf("::%02d", *second)
	}

	switch {
	case date != "" && clock != "":
		return date + " " + clock
	case date != "":
		return date
	default:
		return clock
	}
}
