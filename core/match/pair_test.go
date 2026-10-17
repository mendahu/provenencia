package match_test

import (
	"math"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/match"
	"github.com/mendahu/provenencia/core/namevalue"
	"github.com/mendahu/provenencia/core/propertymeta"
)

func TestEvaluateAgreeConflictUnknown(t *testing.T) {
	name := match.Property{Key: "toponym", Origin: "provenencia"}
	meta := []match.PropertyMeta{{
		Property: name, ValueType: propertymeta.ValueTypeText,
		Cardinality: propertymeta.CardinalitySingle,
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
		Property: res, ValueType: propertymeta.ValueTypeText,
		Cardinality: propertymeta.CardinalityMultiple,
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
		Property: sex, ValueType: propertymeta.ValueTypeTerm,
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
		Property: top, ValueType: propertymeta.ValueTypeText,
		Cardinality: propertymeta.CardinalitySingle,
	}}
	multi := []match.PropertyMeta{{
		Property: top, ValueType: propertymeta.ValueTypeText,
		Cardinality: propertymeta.CardinalityMultiple,
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
	if ev.Comparisons[0].Outcome != match.OutcomeAgree || ev.Comparisons[0].Similarity != 1 {
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

func TestEvaluateStrayName(t *testing.T) {
	prop := match.Property{Key: "name", Origin: "provenencia"}
	meta := []match.PropertyMeta{{Property: prop, ValueType: propertymeta.ValueTypeName}}
	named := func(form string, support int) match.Value {
		v := match.Value{Name: &namevalue.Value{Form: form, Parts: []namevalue.Part{
			{Idx: 0, Value: strings.Fields(form)[0], Type: namevalue.PartTypeGiven},
			{Idx: 1, Value: strings.Fields(form)[1], Type: namevalue.PartTypeSurname},
		}}}
		if support > 0 {
			v.SetEvidence(support, 0, 1, "kept")
		}
		return v
	}
	probe := match.Values{prop: {named("James Robins", 0)}}
	handle := match.Values{prop: {named("Mary Smith", 5), named("James Robins", 1)}}
	ev := match.EvaluateWithNames(probe, handle, meta, nil)
	c := ev.Comparisons[0]
	if c.Outcome != match.OutcomeConflict || !c.Comparable || c.Probe.Label != "James Robins" || c.Candidate.Label != "Mary Smith" {
		t.Fatalf("%+v", c)
	}
}

func TestEvaluateDatePeriod(t *testing.T) {
	prop := func(k string) match.Property { return match.Property{Key: k, Origin: "provenencia"} }
	year := func(y int) match.Value {
		return match.Value{Date: &datevalue.Value{Kind: datevalue.KindPoint, StartYear: &y}}
	}
	metas := []match.PropertyMeta{
		{Property: prop("date"), ValueType: propertymeta.ValueTypeDate},
		{Property: prop("start_date"), ValueType: propertymeta.ValueTypeDate},
		{Property: prop("end_date"), ValueType: propertymeta.ValueTypeDate},
	}
	probe := match.Values{prop("date"): {year(1817)}}
	cand := match.Values{prop("start_date"): {year(1815)}, prop("end_date"): {year(1820)}}
	ev := match.Evaluate(probe, cand, metas)
	if len(ev.Comparisons) != 3 {
		t.Fatalf("%+v", ev.Comparisons)
	}
	c := ev.Comparisons[0]
	if !c.Comparable || c.Outcome != match.OutcomePartial || math.Abs(c.Similarity-0.55) > 1e-9 {
		t.Fatalf("date: %+v", c)
	}
	if ev.Comparisons[1].Comparable || ev.Comparisons[2].Comparable {
		t.Fatalf("start and end: %+v", ev.Comparisons[1:])
	}
}
