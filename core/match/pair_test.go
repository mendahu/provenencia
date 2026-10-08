package match_test

import (
	"testing"

	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/match"
)

func TestEvaluateAgreeConflictUnknown(t *testing.T) {
	name := match.Property{Key: "toponym", Origin: "provenencia"}
	meta := []match.PropertyMeta{{
		Property: name, ValueType: properties.ValueTypeText,
		Cardinality: properties.CardinalitySingle,
	}}
	text := func(s string) match.Values {
		return match.Values{name: {{Text: s, HasText: true}}}
	}
	ev := match.Evaluate(text("Toronto"), text("Toronto"), meta)
	if len(ev.Comparisons) != 1 || ev.Comparisons[0].Outcome != match.OutcomeAgree {
		t.Fatalf("agree: %+v", ev)
	}
	ev = match.Evaluate(text("Toronto"), text("York"), meta)
	if ev.Comparisons[0].Outcome != match.OutcomeConflict {
		t.Fatalf("conflict: %+v", ev)
	}
	ev = match.Evaluate(text("Toronto"), match.Values{}, meta)
	if ev.Comparisons[0].Outcome != match.OutcomeUnknown {
		t.Fatalf("unknown: %+v", ev)
	}
}

func TestEvaluateMultipleCardinalityNeverConflicts(t *testing.T) {
	res := match.Property{Key: "residence", Origin: "provenencia"}
	meta := []match.PropertyMeta{{
		Property: res, ValueType: properties.ValueTypeText,
		Cardinality: properties.CardinalityMultiple,
	}}
	a := match.Values{res: {
		{Text: "York", HasText: true},
		{Text: "Toronto", HasText: true},
	}}
	b := match.Values{res: {{Text: "Medicine Hat", HasText: true}}}
	ev := match.Evaluate(a, b, meta)
	if ev.Comparisons[0].Outcome != match.OutcomeUnknown {
		t.Fatalf("multiple differing: want unknown, got %+v", ev.Comparisons[0])
	}
	bShare := match.Values{res: {{Text: "Toronto", HasText: true}}}
	ev = match.Evaluate(a, bShare, meta)
	if ev.Comparisons[0].Outcome != match.OutcomeAgree {
		t.Fatalf("multiple share: want agree, got %+v", ev.Comparisons[0])
	}
}

func TestEvaluateTermByKey(t *testing.T) {
	sex := match.Property{Key: "sex_at_birth", Origin: "provenencia"}
	meta := []match.PropertyMeta{{
		Property: sex, ValueType: properties.ValueTypeTerm,
	}}
	male := match.Values{sex: {{Term: "male"}}}
	female := match.Values{sex: {{Term: "female"}}}
	ev := match.Evaluate(male, male, meta)
	if ev.Comparisons[0].Outcome != match.OutcomeAgree {
		t.Fatalf("term agree: %+v", ev)
	}
	ev = match.Evaluate(male, female, meta)
	if ev.Comparisons[0].Outcome != match.OutcomeConflict {
		t.Fatalf("term conflict: %+v", ev)
	}
}
