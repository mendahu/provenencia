package match

// Feature is one Property a Profile weighs.
type Feature struct {
	Property Property
	Comparer Comparer
	// Weight is the points a perfect resemblance adds; a partial one adds
	// Weight × similarity.
	Weight float64
	// Contradiction is the points a clear disagreement (similarity 0) takes
	// away. 0 for Properties where differing values say little (names vary
	// across records); high where they rule a match out (sex at birth).
	Contradiction float64
}

// Profile is the matching algorithm for one Subject type.
type Profile struct {
	// Kind is the Subject type key (person, event, place).
	Kind     string
	Features []Feature
	// MinScore is the least score a candidate needs to be returned.
	MinScore float64
}

// With returns a copy of p with f replacing the Feature for the same
// Property, or added after the others. Use it to tune a default profile.
func (p Profile) With(f Feature) Profile {
	out := p
	out.Features = append([]Feature(nil), p.Features...)
	for i := range out.Features {
		if out.Features[i].Property == f.Property {
			out.Features[i] = f
			return out
		}
	}
	out.Features = append(out.Features, f)
	return out
}

// Properties are the Properties the profile reads, in profile order.
func (p Profile) Properties() []Property {
	out := make([]Property, 0, len(p.Features))
	for _, f := range p.Features {
		out = append(out, f.Property)
	}
	return out
}

func product(key string) Property { return Property{Key: key, Origin: "provenencia"} }

// DefaultProfile is the shipped algorithm for a Subject type key, or false
// when the type has none. Tune here; every consumer (Promote suggestions,
// merge hints) reads the same defaults unless it passes its own Profile.
//
// The weights are points. As a guide: a candidate worth showing scores at
// least MinScore; the same name alone clears it, a shared surname alone
// clears it barely, and one hard contradiction sinks any name match.
func DefaultProfile(kind string) (Profile, bool) {
	switch kind {
	case "person":
		return Profile{
			Kind: "person",
			Features: []Feature{
				{Property: product("name"), Comparer: NameComparer{}, Weight: 10},
				{Property: product("sex_at_birth"), Comparer: TermComparer{
					Neutral: map[string]bool{"unknown": true, "indeterminate": true},
				}, Weight: 1, Contradiction: 8},
			},
			// "Mary Robins" ~ "James Robins" is 10 × 0.55 = 5.5 typed (5 as
			// forms): shown, low. "James Smith" is 10 × 0.2 = 2: hidden.
			MinScore: 3,
		}, true
	case "event":
		return Profile{
			Kind: "event",
			Features: []Feature{
				{Property: product("event_type"), Comparer: TermComparer{}, Weight: 4, Contradiction: 6},
				{Property: product("date"), Comparer: DateComparer{}, Weight: 6, Contradiction: 4},
				{Property: product("start_date"), Comparer: DateComparer{}, Weight: 3, Contradiction: 2},
				{Property: product("end_date"), Comparer: DateComparer{}, Weight: 3, Contradiction: 2},
			},
			// The same type alone (every birth) is not enough; a date is.
			MinScore: 5,
		}, true
	case "place":
		return Profile{
			Kind: "place",
			Features: []Feature{
				{Property: product("toponym"), Comparer: TextComparer{}, Weight: 10},
			},
			MinScore: 3,
		}, true
	}
	return Profile{}, false
}
