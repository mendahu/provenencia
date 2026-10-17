package match

import (
	"github.com/mendahu/provenencia/core/propertymeta"
)

// The registry: the default profiles and the period map. The numbers value
// comparison uses (words, names, text, integers, dates) live in the match
// registry's value section and reach match as valuecompare.Rules, an argument
// to Rank, Score and Evaluate; match reads none.

// PeriodProperties are the Properties whose dates make a period: a thing's
// start and end.
type PeriodProperties struct{ Start, End Property }

// DatePeriods maps a date Property to the start and end Properties that stand
// for it when the other side has only a date (see poolPeriods). Another
// pair of start and end dates with no date of its own is one more entry.
var DatePeriods = map[Property]PeriodProperties{
	product("date"): {Start: product("start_date"), End: product("end_date")},
}

// ---------------------------------------------------------------------------
// Profiles and consumers

// DefaultSuggestionLimit caps Promote's target suggestions when the caller
// passes no limit.
const DefaultSuggestionLimit = 10

// DefaultProfile is the shipped algorithm for a Subject type key, or false
// when the type has none. Every consumer (Promote suggestions, merge hints)
// reads these unless it passes its own Profile.
//
// The weights are points. As a guide: a candidate worth showing scores at
// least MinScore; the same name alone clears it, a shared surname alone
// clears it barely. A sex or gender mismatch narrows the score; it does not
// hide a name that otherwise cleared MinScore.
func DefaultProfile(kind string) (Profile, bool) {
	switch kind {
	case "person":
		return Profile{
			Kind: "person",
			Features: []Feature{
				{Property: product("name"), ValueType: propertymeta.ValueTypeName, Weight: 10},
				{Property: product("sex_at_birth"), ValueType: propertymeta.ValueTypeTerm, Weight: 1, Contradiction: 2},
				{Property: product("gender"), ValueType: propertymeta.ValueTypeTerm, Weight: 1, Contradiction: 2},
			},
			// "Mary Robins" ~ "James Robins" is Different: 0, hidden.
			// "Robins" ~ "James Robins" is 10 × 0.75 = 7.5: shown.
			MinScore: 3,
		}, true
	case "event":
		return Profile{
			Kind: "event",
			Features: []Feature{
				{Property: product("event_type"), ValueType: propertymeta.ValueTypeTerm, Weight: 4, Contradiction: 6},
				{Property: product("date"), ValueType: propertymeta.ValueTypeDate, Weight: 6, Contradiction: 4},
				{Property: product("start_date"), ValueType: propertymeta.ValueTypeDate, Weight: 3, Contradiction: 2},
				{Property: product("end_date"), ValueType: propertymeta.ValueTypeDate, Weight: 3, Contradiction: 2},
			},
			// The same type alone (every birth) is not enough; a date is.
			MinScore: 5,
		}, true
	case "place":
		return Profile{
			Kind: "place",
			Features: []Feature{
				{Property: product("toponym"), ValueType: propertymeta.ValueTypeText, Weight: 10, Cardinality: propertymeta.CardinalityMultiple},
			},
			MinScore: 3,
		}, true
	}
	return Profile{}, false
}
