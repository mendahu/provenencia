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

func TestEvaluateTextResemblance(t *testing.T) {
	top := match.Property{Key: "toponym", Origin: "provenencia"}
	single := []match.PropertyMeta{{
		Property: top, ValueType: properties.ValueTypeText,
		Cardinality: properties.CardinalitySingle,
	}}
	multi := []match.PropertyMeta{{
		Property: top, ValueType: properties.ValueTypeText,
		Cardinality: properties.CardinalityMultiple,
	}}
	text := func(s string) match.Values {
		return match.Values{top: {{Text: s, HasText: true}}}
	}

	ev := match.Evaluate(text("Provenance"), text("Provenanced"), single)
	sim := match.TextResemblance("Provenance", "Provenanced")
	if ev.Comparisons[0].Outcome != match.OutcomePartial || ev.Comparisons[0].Similarity != sim || sim < 0.9 {
		t.Fatalf("spelling variant: %+v sim %v", ev.Comparisons[0], sim)
	}
	ev = match.Evaluate(text("Provenance"), text("Provenanced"), multi)
	if ev.Comparisons[0].Outcome != match.OutcomePartial {
		t.Fatalf("multiple spelling variant: %+v", ev.Comparisons[0])
	}

	ev = match.Evaluate(text("York."), text("York"), single)
	if ev.Comparisons[0].Outcome != match.OutcomeAgree || ev.Comparisons[0].Similarity != 0 {
		t.Fatalf("same normalized form: %+v", ev.Comparisons[0])
	}

	ev = match.Evaluate(text("York"), text("York, Upper Canada"), multi)
	shared := ev.Comparisons[0].Similarity
	if ev.Comparisons[0].Outcome != match.OutcomePartial || shared <= 0 || shared >= 0.5 {
		t.Fatalf("shared words: %+v", ev.Comparisons[0])
	}

	ev = match.Evaluate(text("York"), text("Toronto"), single)
	if ev.Comparisons[0].Outcome != match.OutcomeConflict {
		t.Fatalf("different text: %+v", ev.Comparisons[0])
	}
	ev = match.Evaluate(text("York"), text("Toronto"), multi)
	if ev.Comparisons[0].Outcome != match.OutcomeUnknown {
		t.Fatalf("multiple different text: %+v", ev.Comparisons[0])
	}
}
