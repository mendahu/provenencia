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
		if v.StartYear == nil {
			return phrase
		}
		point := formatSide(v.StartYear, v.StartMonth, v.StartDay, v.StartHour, v.StartMinute, v.StartSecond)
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
	if year == nil {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%d", *year)
	if month == nil {
		return b.String()
	}
	fmt.Fprintf(&b, "-%02d", *month)
	if day == nil {
		return b.String()
	}
	fmt.Fprintf(&b, "-%02d", *day)
	if hour == nil {
		return b.String()
	}
	fmt.Fprintf(&b, " %02d", *hour)
	if minute == nil {
		return b.String()
	}
	fmt.Fprintf(&b, ":%02d", *minute)
	if second == nil {
		return b.String()
	}
	fmt.Fprintf(&b, ":%02d", *second)
	return b.String()
}
