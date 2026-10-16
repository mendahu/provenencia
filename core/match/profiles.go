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
	// Cardinality is the Property's cardinality (properties.Cardinality*);
	// empty means single. A multiple Property's displayed values all speak
	// at full weight, so any one of them can match.
	Cardinality string
}

// Profile is the matching algorithm for one Subject type.
type Profile struct {
	// Kind is the Subject type key (person, event, place).
	Kind     string
	Features []Feature
	// MinScore is the least score a candidate needs to be returned.
	MinScore float64
}

// WithNameComparer returns a copy of p whose name features use c. Other
// features stay as they are.
func (p Profile) WithNameComparer(c NameComparer) Profile {
	out := p
	out.Features = append([]Feature(nil), p.Features...)
	for i := range out.Features {
		if _, ok := out.Features[i].Comparer.(NameComparer); ok {
			out.Features[i].Comparer = c
		}
	}
	return out
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
