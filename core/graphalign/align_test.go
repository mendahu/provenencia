package graphalign_test

import (
	"bytes"
	"math"
	"testing"

	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/match"
)

func id(s string) []byte { return []byte(s) }

func textProp(key, text string) (match.Property, match.Value) {
	p := match.Property{Key: key, Origin: "provenencia"}
	return p, match.Value{Text: text, HasText: true}
}

func termProp(key, term string) (match.Property, match.Value) {
	p := match.Property{Key: key, Origin: "provenencia"}
	return p, match.Value{Term: term}
}

func vals(pairs ...any) match.Values {
	out := match.Values{}
	for i := 0; i+1 < len(pairs); i += 2 {
		p := pairs[i].(match.Property)
		v := pairs[i+1].(match.Value)
		out[p] = append(out[p], v)
	}
	return out
}

func personMetas() []match.PropertyMeta {
	return []match.PropertyMeta{
		{Property: match.Property{Key: "name", Origin: "provenencia"}, ValueType: properties.ValueTypeText},
		{Property: match.Property{Key: "sex_at_birth", Origin: "provenencia"}, ValueType: properties.ValueTypeTerm},
	}
}

func placeMetas() []match.PropertyMeta {
	return []match.PropertyMeta{
		{Property: match.Property{Key: "toponym", Origin: "provenencia"}, ValueType: properties.ValueTypeText},
	}
}

func eventMetas() []match.PropertyMeta {
	return []match.PropertyMeta{
		{Property: match.Property{Key: "event_type", Origin: "provenencia"}, ValueType: properties.ValueTypeTerm},
	}
}

func partSig(role, neighborKind, typeTerm string) graphalign.EdgeSignature {
	return graphalign.EdgeSignature{
		BridgeType: "participation", RoleOrType: role,
		NeighborKind: neighborKind, NeighborTypeTerm: typeTerm,
	}
}

func relSig(typ string) graphalign.EdgeSignature {
	return graphalign.EdgeSignature{
		BridgeType: "relationship", RoleOrType: typ,
		NeighborKind: "person",
	}
}

func rowBySubject(p graphalign.Proposal, subjectID []byte) graphalign.Row {
	for _, r := range p.Rows {
		if bytes.Equal(r.SubjectID, subjectID) {
			return r
		}
	}
	return graphalign.Row{}
}

func TestSeedFromAlreadyPromotedNeighbor(t *testing.T) {
	// Event already promoted; unpromoted person participates → seed person.
	evSub := id("s-event")
	perSub := id("s-person")
	evH := id("h-event")
	perH := id("h-person")
	np, nv := textProp("name", "Gracie")
	et, etv := termProp("event_type", "birth")
	sig := partSig("subject", "event", "birth")

	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: evSub, Ref: "EVT-S", Kind: "event", Values: vals(et, etv)},
			{ID: perSub, Ref: "PER-S", Kind: "person", Values: vals(np, nv)},
		},
		Bridges: []graphalign.Bridge{{A: perSub, B: evSub, Signature: sig}},
		Metas:   append(personMetas(), eventMetas()...),
	}
	canon := graphalign.Canon{
		Handles: []graphalign.Handle{
			{ID: evH, Ref: "EVT-1", Kind: "event", Values: vals(et, etv)},
			{ID: perH, Ref: "PER-1", Kind: "person", Values: vals(np, nv)},
		},
		Edges: []graphalign.CanonEdge{{From: perH, To: evH, Signature: sig}},
	}
	fixed := []graphalign.Fixed{{SubjectID: evSub, HandleID: evH}}
	p := graphalign.Align(layer, canon, graphalign.Stats{}, fixed, nil)
	row := rowBySubject(p, perSub)
	if row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, perH) {
		t.Fatalf("person from promoted event: %+v", row)
	}
	if row.Via == nil || !bytes.Equal(row.Via.NeighborSubjectID, evSub) {
		t.Fatalf("person should be reached via the event: %+v", row.Via)
	}
}

func TestSeedFromNewlyFixedIntoUnpromoted(t *testing.T) {
	// Person fixed; unpromoted event → seed event (reverse of above).
	evSub := id("s-event")
	perSub := id("s-person")
	evH := id("h-event")
	perH := id("h-person")
	np, nv := textProp("name", "Gracie")
	et, etv := termProp("event_type", "birth")
	sig := partSig("subject", "event", "birth")

	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: evSub, Ref: "EVT-S", Kind: "event", Values: vals(et, etv)},
			{ID: perSub, Ref: "PER-S", Kind: "person", Values: vals(np, nv)},
		},
		Bridges: []graphalign.Bridge{{A: perSub, B: evSub, Signature: sig}},
		Metas:   append(personMetas(), eventMetas()...),
	}
	canon := graphalign.Canon{
		Handles: []graphalign.Handle{
			{ID: evH, Ref: "EVT-1", Kind: "event", Values: vals(et, etv)},
			{ID: perH, Ref: "PER-1", Kind: "person", Values: vals(np, nv)},
		},
		Edges: []graphalign.CanonEdge{{From: perH, To: evH, Signature: sig}},
	}
	fixed := []graphalign.Fixed{{SubjectID: perSub, HandleID: perH}}
	p := graphalign.Align(layer, canon, graphalign.Stats{}, fixed, nil)
	row := rowBySubject(p, evSub)
	if row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, evH) {
		t.Fatalf("event from fixed person: %+v", row)
	}
}

func TestBothUnpromotedOnceOneFixed(t *testing.T) {
	a, b := id("s-a"), id("s-b")
	ha, hb := id("h-a"), id("h-b")
	na, va := textProp("name", "Ann")
	nb, vb := textProp("name", "Bob")
	sig := relSig("spouse")
	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: a, Ref: "PER-A", Kind: "person", Values: vals(na, va)},
			{ID: b, Ref: "PER-B", Kind: "person", Values: vals(nb, vb)},
		},
		Bridges: []graphalign.Bridge{{A: a, B: b, Signature: sig}},
		Metas:   personMetas(),
	}
	canon := graphalign.Canon{
		Handles: []graphalign.Handle{
			{ID: ha, Ref: "PER-1", Kind: "person", Values: vals(na, va)},
			{ID: hb, Ref: "PER-2", Kind: "person", Values: vals(nb, vb)},
		},
		Edges: []graphalign.CanonEdge{{From: ha, To: hb, Signature: sig}},
	}
	fixed := []graphalign.Fixed{{SubjectID: a, HandleID: ha}}
	p := graphalign.Align(layer, canon, graphalign.Stats{}, fixed, nil)
	if row := rowBySubject(p, b); row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, hb) {
		t.Fatalf("spouse: %+v", row)
	}
}

func TestOneAgreeingPropertyStaysAWeakMatch(t *testing.T) {
	// York is one of two toponyms, so its frequency is 0.5 and the raw
	// log-odds sits under the accept bar. The agreement is still a weak match.
	orphan := id("s-orphan")
	known := id("h-known")
	n, v := textProp("toponym", "York")
	stats := graphalign.Stats{ValueFreq: map[string]map[string]float64{
		"toponym": {"York": 0.5},
	}}
	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: orphan, Ref: "PLC-O", Kind: "place", Values: vals(n, v)},
		},
		Metas: placeMetas(),
	}
	canon := graphalign.Canon{
		Handles: []graphalign.Handle{
			{ID: known, Ref: "PLC-K", Kind: "place", Values: vals(n, v)},
		},
	}
	row := rowBySubject(graphalign.Align(layer, canon, stats, nil, nil), orphan)
	if row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, known) ||
		row.Assessment != graphalign.AssessMedium || row.Reason != graphalign.ReasonAgrees ||
		row.ReasonProperty.Key != "toponym" {
		t.Fatalf("same toponym: %+v", row)
	}
	if row.Score >= graphalign.DefaultConfig().StrongScore {
		t.Fatalf("toponym alone is not strong: score %.2f", row.Score)
	}

	// Without the registry weight, the same common toponym stays weak.
	cfg := graphalign.DefaultConfig()
	cfg.PropertyAgreement = nil
	row = rowBySubject(graphalign.Align(layer, canon, stats, nil, &cfg), orphan)
	if row.Assessment != graphalign.AssessWeak {
		t.Fatalf("toponym without the registry weight: %+v", row)
	}

	other, ov := textProp("toponym", "Leeds")
	canon.Handles[0].Values = vals(other, ov)
	row = rowBySubject(graphalign.Align(layer, canon, stats, nil, nil), orphan)
	if row.Target == graphalign.TargetHandle || row.Assessment != graphalign.AssessNone {
		t.Fatalf("different toponym should not match: %+v", row)
	}
}

func TestSpellingVariantIsAWeakerTextMatch(t *testing.T) {
	// One added letter on a toponym is a match, weaker than the exact spelling.
	// The frequency is the incoming spelling, so the small-catalog floor applies
	// and the resemblance scales it: about 0.91 × 3, which is weak, not medium.
	orphan := id("s-orphan")
	known := id("h-known")
	probe, pv := textProp("toponym", "Provenance")
	canonProp, cv := textProp("toponym", "Provenanced")
	stats := graphalign.Stats{ValueFreq: map[string]map[string]float64{
		"toponym": {"Provenance": 0.5},
	}}
	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: orphan, Ref: "PLC-O", Kind: "place", Values: vals(probe, pv)},
		},
		Metas: placeMetas(),
	}
	canon := graphalign.Canon{
		Handles: []graphalign.Handle{
			{ID: known, Ref: "PLC-K", Kind: "place", Values: vals(canonProp, cv)},
		},
	}
	cfg := graphalign.DefaultConfig()
	row := rowBySubject(graphalign.Align(layer, canon, stats, nil, nil), orphan)
	if row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, known) ||
		row.Assessment != graphalign.AssessWeak || row.Reason != graphalign.ReasonAgrees ||
		row.ReasonProperty.Key != "toponym" ||
		row.Score < cfg.WeakScore || row.Score >= cfg.MediumScore {
		t.Fatalf("spelling variant: %+v", row)
	}
	var partial bool
	for _, c := range row.Comparisons {
		if c.Property.Key != "toponym" {
			continue
		}
		if c.Outcome != match.OutcomePartial || c.Pinned {
			t.Fatalf("comparison %+v", c)
		}
		partial = true
	}
	if !partial {
		t.Fatal("missing toponym comparison")
	}

	// Without the toponym bonus the scaled floor (accept bar × resemblance)
	// stays under weak. The resemblance itself is any text property; the
	// bonus is what makes a toponym clear the bar.
	bare := graphalign.DefaultConfig()
	bare.PropertyAgreement = nil
	row = rowBySubject(graphalign.Align(layer, canon, stats, nil, &bare), orphan)
	if row.Target == graphalign.TargetHandle || row.Assessment != graphalign.AssessNone {
		t.Fatalf("spelling variant without the registry weight: %+v", row)
	}

	// Sharing one word of a longer toponym is a resemblance, not a match.
	york, yv := textProp("toponym", "York")
	longer, lv := textProp("toponym", "York, Upper Canada")
	layer.Metas = placeMetas()
	layer.Subjects[0].Values = vals(york, yv)
	canon.Handles[0].Values = vals(longer, lv)
	stats.ValueFreq = map[string]map[string]float64{"toponym": {"York": 0.5}}
	row = rowBySubject(graphalign.Align(layer, canon, stats, nil, nil), orphan)
	if row.Target == graphalign.TargetHandle || row.Assessment != graphalign.AssessNone {
		t.Fatalf("shared word should stay under the bar: %+v", row)
	}
}

func TestOrphanFallsBackToPropertyOnlyOrSkip(t *testing.T) {
	// Place orphans use toponym + DefaultText Rank (names need namevalues).
	orphan := id("s-orphan")
	known := id("h-known")
	n, v := textProp("toponym", "UniqueZebraTown")
	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: orphan, Ref: "PLC-O", Kind: "place", Values: vals(n, v)},
		},
		Metas: placeMetas(),
	}
	canon := graphalign.Canon{
		Handles: []graphalign.Handle{
			{ID: known, Ref: "PLC-K", Kind: "place", Values: vals(n, v)},
		},
	}
	p := graphalign.Align(layer, canon, graphalign.Stats{}, nil, nil)
	row := rowBySubject(p, orphan)
	if row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, known) {
		t.Fatalf("orphan with match: %+v", row)
	}

	weak := id("s-weak")
	wn, wv := textProp("toponym", "X")
	other, ov := textProp("toponym", "CompletelyDifferentPlace")
	layer2 := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: weak, Ref: "PLC-W", Kind: "place", Values: vals(wn, wv)},
		},
		Metas: placeMetas(),
	}
	canon2 := graphalign.Canon{
		Handles: []graphalign.Handle{
			{ID: known, Ref: "PLC-K", Kind: "place", Values: vals(other, ov)},
		},
	}
	p2 := graphalign.Align(layer2, canon2, graphalign.Stats{}, nil, nil)
	row2 := rowBySubject(p2, weak)
	if row2.Target == graphalign.TargetHandle {
		t.Fatalf("weak orphan should not merge: %+v", row2)
	}
}

func TestConfigOverrideWithoutEditingRegistry(t *testing.T) {
	cfg := graphalign.DefaultConfig()
	cfg.AcceptScore = 100 // impossible
	evSub := id("s-event")
	perSub := id("s-person")
	evH := id("h-event")
	perH := id("h-person")
	np, nv := textProp("name", "Gracie")
	et, etv := termProp("event_type", "birth")
	sig := partSig("subject", "event", "birth")
	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: evSub, Ref: "EVT-S", Kind: "event", Values: vals(et, etv)},
			{ID: perSub, Ref: "PER-S", Kind: "person", Values: vals(np, nv)},
		},
		Bridges: []graphalign.Bridge{{A: perSub, B: evSub, Signature: sig}},
		Metas:   append(personMetas(), eventMetas()...),
	}
	canon := graphalign.Canon{
		Handles: []graphalign.Handle{
			{ID: evH, Ref: "EVT-1", Kind: "event", Values: vals(et, etv)},
			{ID: perH, Ref: "PER-1", Kind: "person", Values: vals(np, nv)},
		},
		Edges: []graphalign.CanonEdge{{From: perH, To: evH, Signature: sig}},
	}
	fixed := []graphalign.Fixed{{SubjectID: evSub, HandleID: evH}}
	p := graphalign.Align(layer, canon, graphalign.Stats{}, fixed, &cfg)
	row := rowBySubject(p, perSub)
	// Structure path refused; Rank may still assign the same person.
	if row.Target == graphalign.TargetHandle && bytes.Equal(row.HandleID, perH) {
		// Rank fallback is OK
		return
	}
	if row.Target == graphalign.TargetHandle {
		t.Fatalf("unexpected handle: %+v", row)
	}
}

func TestGracieObituaryGolden(t *testing.T) {
	gracie := id("s-gracie")
	birth := id("s-birth")
	death := id("s-death")
	burial := id("s-burial")
	child1 := id("s-child1")
	child2 := id("s-child2")
	res1 := id("s-res1")
	medHat := id("s-medhat")

	hGracie := id("h-gracie")
	hBirth := id("h-birth")
	hDeath := id("h-death")
	hBurial := id("h-burial")
	hChild1 := id("h-child1")
	hChild2 := id("h-child2")

	gn, gv := textProp("name", "Gracie Gray Gates")
	c1n, c1v := textProp("name", "Alice Gates")
	c2n, c2v := textProp("name", "Bob Gates")
	birthT, birthV := termProp("event_type", "birth")
	deathT, deathV := termProp("event_type", "death")
	burialT, burialV := termProp("event_type", "burial")
	top, topV := textProp("toponym", "Medicine Hat")
	resTop, resV := textProp("toponym", "York Township")

	birthSig := partSig("subject", "event", "birth")
	deathSig := partSig("subject", "event", "death")
	burialSig := partSig("subject", "event", "burial")
	childSig := relSig("parent")

	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: gracie, Ref: "PER-G", Kind: "person", Values: vals(gn, gv)},
			{ID: birth, Ref: "EVT-B", Kind: "event", Values: vals(birthT, birthV)},
			{ID: death, Ref: "EVT-D", Kind: "event", Values: vals(deathT, deathV)},
			{ID: burial, Ref: "EVT-U", Kind: "event", Values: vals(burialT, burialV)},
			{ID: child1, Ref: "PER-C1", Kind: "person", Values: vals(c1n, c1v)},
			{ID: child2, Ref: "PER-C2", Kind: "person", Values: vals(c2n, c2v)},
			{ID: res1, Ref: "PLC-R", Kind: "place", Values: vals(resTop, resV)},
			{ID: medHat, Ref: "PLC-M", Kind: "place", Values: vals(top, topV)},
		},
		Bridges: []graphalign.Bridge{
			{A: gracie, B: birth, Signature: birthSig},
			{A: gracie, B: death, Signature: deathSig},
			{A: gracie, B: burial, Signature: burialSig},
			{A: gracie, B: child1, Signature: childSig},
			{A: gracie, B: child2, Signature: childSig},
		},
		Metas: append(append(personMetas(), eventMetas()...), placeMetas()...),
	}
	canon := graphalign.Canon{
		Handles: []graphalign.Handle{
			{ID: hGracie, Ref: "PER-Z", Kind: "person", Values: vals(gn, gv)},
			{ID: hBirth, Ref: "EVT-B1", Kind: "event", Values: vals(birthT, birthV)},
			{ID: hDeath, Ref: "EVT-D1", Kind: "event", Values: vals(deathT, deathV)},
			{ID: hBurial, Ref: "EVT-U1", Kind: "event", Values: vals(burialT, burialV)},
			{ID: hChild1, Ref: "PER-A", Kind: "person", Values: vals(c1n, c1v)},
			{ID: hChild2, Ref: "PER-B", Kind: "person", Values: vals(c2n, c2v)},
		},
		Edges: []graphalign.CanonEdge{
			{From: hGracie, To: hBirth, Signature: birthSig},
			{From: hGracie, To: hDeath, Signature: deathSig},
			{From: hGracie, To: hBurial, Signature: burialSig},
			{From: hGracie, To: hChild1, Signature: childSig},
			{From: hGracie, To: hChild2, Signature: childSig},
		},
	}
	stats := graphalign.Stats{
		FanOut: map[string]float64{
			birthSig.Key():  1,
			deathSig.Key():  1,
			burialSig.Key(): 1,
			childSig.Key():  3,
		},
	}
	fixed := []graphalign.Fixed{{SubjectID: gracie, HandleID: hGracie}}

	p1 := graphalign.Align(layer, canon, stats, fixed, nil)
	mustHandle := func(sid, hid []byte) {
		t.Helper()
		r := rowBySubject(p1, sid)
		if r.Target != graphalign.TargetHandle || !bytes.Equal(r.HandleID, hid) {
			t.Fatalf("%s: want handle %s, got %+v", sid, hid, r)
		}
	}
	mustHandle(gracie, hGracie)
	mustHandle(birth, hBirth)
	mustHandle(death, hDeath)
	mustHandle(burial, hBurial)
	mustHandle(child1, hChild1)
	mustHandle(child2, hChild2)

	for _, sid := range [][]byte{res1, medHat} {
		r := rowBySubject(p1, sid)
		if r.Target == graphalign.TargetHandle {
			t.Fatalf("place %s should be New or Skip, got %+v", sid, r)
		}
		if r.Target != graphalign.TargetNew && r.Target != graphalign.TargetSkip {
			t.Fatalf("place %s: %+v", sid, r)
		}
	}

	// Deterministic re-run.
	p2 := graphalign.Align(layer, canon, stats, fixed, nil)
	if len(p1.Rows) != len(p2.Rows) {
		t.Fatalf("row count %d vs %d", len(p1.Rows), len(p2.Rows))
	}
	for i := range p1.Rows {
		a, b := p1.Rows[i], p2.Rows[i]
		if a.Target != b.Target || !bytes.Equal(a.HandleID, b.HandleID) || a.HandleRef != b.HandleRef {
			t.Fatalf("nondeterministic row %d: %+v vs %+v", i, a, b)
		}
	}

	// Fixed row stays Gracie even if we only pass that fixed pair again.
	rG := rowBySubject(p1, gracie)
	if rG.Target != graphalign.TargetHandle || !bytes.Equal(rG.HandleID, hGracie) {
		t.Fatalf("fixed gracie: %+v", rG)
	}
}

func nameVals(given, surname string) match.Values {
	p := match.Property{Key: "name", Origin: "provenencia"}
	return match.Values{p: {{Name: &namevalues.Value{
		Form: given + " " + surname,
		Parts: []namevalues.Part{
			{Idx: 0, Type: namevalues.PartTypeGiven, Value: given},
			{Idx: 1, Type: namevalues.PartTypeSurname, Value: surname},
		},
	}}}}
}

func nameMetas() []match.PropertyMeta {
	return []match.PropertyMeta{
		{Property: match.Property{Key: "name", Origin: "provenencia"}, ValueType: properties.ValueTypeName},
	}
}

func TestUnreachableFallbackScoresOnTheWalkScale(t *testing.T) {
	tests := []struct {
		name       string
		probe      match.Values
		handle     match.Values
		wantTarget graphalign.Target
		wantAssess graphalign.Assessment
	}{
		{
			name:  "a shared surname alone is not a match",
			probe: nameVals("Mary", "Robins"), handle: nameVals("James", "Robins"),
			wantTarget: graphalign.TargetSkip, wantAssess: graphalign.AssessNone,
		},
		{
			name:  "the same name alone is a weak match, never strong",
			probe: nameVals("Mary", "Robins"), handle: nameVals("Mary", "Robins"),
			wantTarget: graphalign.TargetHandle, wantAssess: graphalign.AssessWeak,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, h := id("s-orphan"), id("h-known")
			layer := graphalign.Layer{
				Subjects: []graphalign.Subject{{ID: s, Ref: "PER-S", Kind: "person", Values: tt.probe}},
				Metas:    nameMetas(),
			}
			canon := graphalign.Canon{Handles: []graphalign.Handle{{ID: h, Ref: "PER-1", Kind: "person", Values: tt.handle}}}
			row := rowBySubject(graphalign.Align(layer, canon, graphalign.Stats{}, nil, nil), s)
			if row.Target != tt.wantTarget || row.Assessment != tt.wantAssess {
				t.Fatalf("got %s / %s (score %.2f), want %s / %s", row.Target, row.Assessment, row.Score, tt.wantTarget, tt.wantAssess)
			}
		})
	}
}

func TestPartialNamesAreWeak(t *testing.T) {
	given := func(i int, v string) namevalues.Part {
		return namevalues.Part{Idx: i, Type: namevalues.PartTypeGiven, Value: v}
	}
	sur := func(i int, v string) namevalues.Part {
		return namevalues.Part{Idx: i, Type: namevalues.PartTypeSurname, Value: v}
	}
	full := func(parts ...namevalues.Part) match.Values {
		p := match.Property{Key: "name", Origin: "provenencia"}
		return match.Values{p: {{Name: &namevalues.Value{Parts: parts}}}}
	}
	alignOne := func(probe, handle match.Values) graphalign.Row {
		s, h := id("s"), id("h")
		return rowBySubject(graphalign.Align(
			graphalign.Layer{Subjects: []graphalign.Subject{{ID: s, Ref: "SUB", Kind: "person", Values: probe}}, Metas: nameMetas()},
			graphalign.Canon{Handles: []graphalign.Handle{{ID: h, Ref: "PER-1", Kind: "person", Values: handle}}},
			graphalign.Stats{}, nil, nil,
		), s)
	}
	james := alignOne(
		full(given(0, "James"), sur(1, "Robins")),
		full(given(0, "James"), given(1, "Kenneth"), sur(2, "Robins")),
	)
	if james.Target != graphalign.TargetHandle || james.Assessment != graphalign.AssessWeak || james.Score >= graphalign.DefaultConfig().MediumScore {
		t.Fatalf("James / James Kenneth: %+v", james)
	}
	lee := alignOne(
		full(given(0, "Lee"), sur(1, "Breakell")),
		full(given(0, "Lee-Ellen"), given(1, "Matilda"), sur(2, "Breakell")),
	)
	if lee.Target != graphalign.TargetHandle || lee.Assessment != graphalign.AssessWeak || lee.Score >= graphalign.DefaultConfig().MediumScore {
		t.Fatalf("Lee / Lee-Ellen: %+v", lee)
	}
	// The stored forms disagree. The parts are what match: surname first on
	// the certificate, given name first on the new subject.
	withForm := func(base match.Values, form string) match.Values {
		out := match.Values{}
		for k, vs := range base {
			cp := append([]match.Value(nil), vs...)
			if cp[0].Name != nil {
				n := *cp[0].Name
				n.Form = form
				cp[0].Name = &n
			}
			out[k] = cp
		}
		return out
	}
	certificate := alignOne(
		withForm(full(given(0, "James"), sur(1, "Robins")), "unrelated transcription"),
		withForm(full(sur(0, "Robins"), given(1, "James"), given(2, "Kenneth")), "Robins, James Kenneth"),
	)
	if certificate.Target != graphalign.TargetHandle || certificate.Assessment != graphalign.AssessWeak ||
		math.Abs(certificate.Score-james.Score) > 1e-9 {
		t.Fatalf("parts James / certificate parts: %+v, james score %.3f", certificate, james.Score)
	}

	sex := match.Property{Key: "sex_at_birth", Origin: "provenencia"}
	metas := append(nameMetas(), match.PropertyMeta{Property: sex, ValueType: properties.ValueTypeTerm})
	withSex := func(base match.Values, term string) match.Values {
		out := match.Values{}
		for k, v := range base {
			out[k] = v
		}
		out[sex] = []match.Value{{Term: term}}
		return out
	}
	probe := full(given(0, "James"), sur(1, "Robins"))
	handle := full(given(0, "James"), given(1, "Kenneth"), sur(2, "Robins"))
	s, h := id("s"), id("h")
	mismatch := rowBySubject(graphalign.Align(
		graphalign.Layer{Subjects: []graphalign.Subject{{ID: s, Ref: "SUB", Kind: "person", Values: withSex(probe, "female")}}, Metas: metas},
		graphalign.Canon{Handles: []graphalign.Handle{{ID: h, Ref: "PER-1", Kind: "person", Values: withSex(handle, "male")}}},
		graphalign.Stats{}, nil, nil,
	), s)
	if diff := james.Score - mismatch.Score; diff < 0.99 || diff > 1.01 {
		t.Fatalf("sex mismatch subtracted %.3f, want 1 (name %.3f, with sex %.3f)", diff, james.Score, mismatch.Score)
	}
}

func TestSexAgreementAloneStaysUnderWeak(t *testing.T) {
	sex := match.Property{Key: "sex_at_birth", Origin: "provenencia"}
	meta := []match.PropertyMeta{{Property: sex, ValueType: properties.ValueTypeTerm}}
	male := match.Values{sex: {{Term: "male"}}}
	ev := match.Evaluate(male, male, meta)
	sc := graphalign.ScoreCandidate(ev, male, 0, 1, graphalign.DefaultConfig(), graphalign.Stats{})
	if sc.Score < 0.59 || sc.Score > 0.61 || sc.Score >= graphalign.DefaultConfig().WeakScore {
		t.Fatalf("sex alone scored %.3f", sc.Score)
	}
	s, h := id("s"), id("h")
	row := rowBySubject(graphalign.Align(
		graphalign.Layer{Subjects: []graphalign.Subject{{ID: s, Ref: "SUB", Kind: "person", Values: male}}, Metas: meta},
		graphalign.Canon{Handles: []graphalign.Handle{{ID: h, Ref: "PER-1", Kind: "person", Values: male}}},
		graphalign.Stats{}, nil, nil,
	), s)
	if row.Target == graphalign.TargetHandle || row.Assessment != graphalign.AssessNone {
		t.Fatalf("sex alone published %+v", row)
	}
}

func TestContestedHandleIsDecidedTheSameEveryRun(t *testing.T) {
	tests := []struct {
		name   string
		a, b   match.Values
		winner string
	}{
		{name: "equal claims go to the first ref", a: nameVals("John", "Smith"), b: nameVals("John", "Smith"), winner: "s-a"},
		{name: "the stronger claim wins whatever its ref", a: nameVals("Jon", "Smith"), b: nameVals("John", "Smith"), winner: "s-b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sex := match.Property{Key: "sex_at_birth", Origin: "provenencia"}
			male := match.Value{Term: "male"}
			b := tt.b
			if tt.winner == "s-b" {
				b = match.Values{}
				for k, v := range tt.b {
					b[k] = v
				}
				b[sex] = []match.Value{male}
			}
			handle := nameVals("John", "Smith")
			handle[sex] = []match.Value{male}
			metas := append(nameMetas(), match.PropertyMeta{Property: sex, ValueType: properties.ValueTypeTerm})
			layer := graphalign.Layer{
				Subjects: []graphalign.Subject{
					{ID: id("s-a"), Ref: "PER-A", Kind: "person", Values: tt.a},
					{ID: id("s-b"), Ref: "PER-B", Kind: "person", Values: b},
				},
				Metas: metas,
			}
			canon := graphalign.Canon{Handles: []graphalign.Handle{{ID: id("h-1"), Ref: "PER-1", Kind: "person", Values: handle}}}
			for run := 0; run < 50; run++ {
				p := graphalign.Align(layer, canon, graphalign.Stats{}, nil, nil)
				var got []string
				for _, r := range p.Rows {
					if r.Target == graphalign.TargetHandle {
						got = append(got, string(r.SubjectID))
					}
				}
				if len(got) != 1 || got[0] != tt.winner {
					t.Fatalf("run %d: handle went to %v, want %s", run, got, tt.winner)
				}
			}
		})
	}
}

func TestEdgeSupportCountsTheSeedingNeighborOnce(t *testing.T) {
	sig := partSig("subject", "event", "birth")
	evSub, perSub, evH, perH := id("s-event"), id("s-person"), id("h-event"), id("h-person")
	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: evSub, Ref: "EVT-S", Kind: "event"},
			{ID: perSub, Ref: "PER-S", Kind: "person"},
		},
		Bridges: []graphalign.Bridge{{A: perSub, B: evSub, Signature: sig}},
		Metas:   personMetas(),
	}
	canon := graphalign.Canon{
		Handles: []graphalign.Handle{{ID: evH, Ref: "EVT-1", Kind: "event"}, {ID: perH, Ref: "PER-1", Kind: "person"}},
		Edges:   []graphalign.CanonEdge{{From: perH, To: evH, Signature: sig}},
	}
	stats := graphalign.Stats{FanOut: map[string]float64{sig.Key(): 1}}
	fixed := []graphalign.Fixed{{SubjectID: evSub, HandleID: evH}}
	row := rowBySubject(graphalign.Align(layer, canon, stats, fixed, nil), perSub)
	cfg := graphalign.DefaultConfig()
	if row.Score != cfg.EdgeSupportLow {
		t.Fatalf("score %.2f, want one edge's support %.2f", row.Score, cfg.EdgeSupportLow)
	}
	if row.Target == graphalign.TargetHandle {
		t.Fatalf("one edge and no shared values should not merge: %+v", row)
	}
}

func TestExactNameCreditsTheNeighborOnce(t *testing.T) {
	sig := partSig("subject", "event", "birth")
	evSub, perSub, evH, perH := id("s-event"), id("s-person"), id("h-event"), id("h-person")
	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: evSub, Ref: "EVT-S", Kind: "event"},
			{ID: perSub, Ref: "PER-S", Kind: "person", Values: nameVals("Ann", "Ames")},
		},
		Bridges: []graphalign.Bridge{{A: perSub, B: evSub, Signature: sig}},
		Metas:   append(nameMetas(), eventMetas()...),
	}
	canon := graphalign.Canon{
		Handles: []graphalign.Handle{
			{ID: evH, Ref: "EVT-1", Kind: "event"},
			{ID: perH, Ref: "PER-1", Kind: "person", Values: nameVals("Ann", "Ames")},
		},
		Edges: []graphalign.CanonEdge{{From: perH, To: evH, Signature: sig}},
	}
	stats := graphalign.Stats{FanOut: map[string]float64{sig.Key(): 1}}
	fixed := []graphalign.Fixed{{SubjectID: perSub, HandleID: perH}}
	row := rowBySubject(graphalign.Align(layer, canon, stats, fixed, nil), evSub)
	cfg := graphalign.DefaultConfig()
	name := cfg.PropertyScale[match.Property{Key: "name", Origin: "provenencia"}].Weight
	want := cfg.EdgeSupportLow + cfg.NeighborCreditLow*name
	if row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, evH) ||
		math.Abs(row.Score-want) > 1e-9 || row.Assessment != graphalign.AssessWeak {
		t.Fatalf("event %+v, want %s at %.2f", row, evH, want)
	}
}

func TestHighFanOutNameCreditStaysUnderTheBar(t *testing.T) {
	sig := partSig("subject", "event", "residence")
	evSub, perSub, evH, perH := id("s-event"), id("s-person"), id("h-event"), id("h-person")
	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: evSub, Ref: "EVT-S", Kind: "event"},
			{ID: perSub, Ref: "PER-S", Kind: "person", Values: nameVals("Ann", "Ames")},
		},
		Bridges: []graphalign.Bridge{{A: perSub, B: evSub, Signature: sig}},
		Metas:   append(nameMetas(), eventMetas()...),
	}
	canon := graphalign.Canon{
		Handles: []graphalign.Handle{
			{ID: evH, Ref: "EVT-1", Kind: "event"},
			{ID: perH, Ref: "PER-1", Kind: "person", Values: nameVals("Ann", "Ames")},
		},
		Edges: []graphalign.CanonEdge{{From: perH, To: evH, Signature: sig}},
	}
	stats := graphalign.Stats{FanOut: map[string]float64{sig.Key(): 3}}
	fixed := []graphalign.Fixed{{SubjectID: perSub, HandleID: perH}}
	row := rowBySubject(graphalign.Align(layer, canon, stats, fixed, nil), evSub)
	cfg := graphalign.DefaultConfig()
	name := cfg.PropertyScale[match.Property{Key: "name", Origin: "provenencia"}].Weight
	want := cfg.EdgeSupportHigh + cfg.NeighborCreditHigh*name
	if row.Target == graphalign.TargetHandle || math.Abs(row.Score-want) > 1e-9 || row.Score >= cfg.AcceptScore {
		t.Fatalf("event %+v, want a suggestion at %.2f", row, want)
	}
}

func TestConflictingNeighborLowersTheScore(t *testing.T) {
	sig := partSig("subject", "event", "birth")
	evSub, perSub, evH, perH := id("s-event"), id("s-person"), id("h-event"), id("h-person")
	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: evSub, Ref: "EVT-S", Kind: "event"},
			{ID: perSub, Ref: "PER-S", Kind: "person", Values: nameVals("Mary", "Robins")},
		},
		Bridges: []graphalign.Bridge{{A: perSub, B: evSub, Signature: sig}},
		Metas:   append(nameMetas(), eventMetas()...),
	}
	canon := graphalign.Canon{
		Handles: []graphalign.Handle{
			{ID: evH, Ref: "EVT-1", Kind: "event"},
			{ID: perH, Ref: "PER-1", Kind: "person", Values: nameVals("James", "Robins")},
		},
		Edges: []graphalign.CanonEdge{{From: perH, To: evH, Signature: sig}},
	}
	stats := graphalign.Stats{FanOut: map[string]float64{sig.Key(): 1}}
	fixed := []graphalign.Fixed{{SubjectID: perSub, HandleID: perH}}
	row := rowBySubject(graphalign.Align(layer, canon, stats, fixed, nil), evSub)
	cfg := graphalign.DefaultConfig()
	penalty := cfg.PropertyScale[match.Property{Key: "name", Origin: "provenencia"}].Contradiction
	want := cfg.EdgeSupportLow + cfg.NeighborCreditLow*(-penalty)
	if row.Target == graphalign.TargetHandle || math.Abs(row.Score-want) > 1e-9 || row.Score >= cfg.EdgeSupportLow {
		t.Fatalf("event %+v, want %.2f, under one edge's support", row, want)
	}
}

func hasAlt(row graphalign.Row, handle []byte) bool {
	for _, alt := range row.Alternatives {
		if bytes.Equal(alt.HandleID, handle) {
			return true
		}
	}
	return false
}

func TestAcceptedPropertyMatchNominatesItsNeighbor(t *testing.T) {
	// Two names clear the bar. The third subject does not: its only property
	// is a shared term, under that kind's Rank minimum. Both accepted handles
	// bridge to it, and both canon edges share that bridge's signature, so
	// the walk nominates the neighbor. The score counts both edges and a
	// fraction of each neighbor's name. A canon neighbor on a different
	// signature is not a suggestion.
	sig := partSig("subject", "event", "gathering")
	other := partSig("subject", "event", "other")
	a, b, mid := id("s-a"), id("s-b"), id("s-mid")
	ha, hb, hMid, hWrong := id("h-a"), id("h-b"), id("h-mid"), id("h-wrong")
	et, etv := termProp("event_type", "gathering")
	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: a, Ref: "PER-A", Kind: "person", Values: nameVals("Ann", "Ames")},
			{ID: b, Ref: "PER-B", Kind: "person", Values: nameVals("Bob", "Ames")},
			{ID: mid, Ref: "EVT-M", Kind: "event", Values: vals(et, etv)},
		},
		Bridges: []graphalign.Bridge{
			{A: a, B: mid, Signature: sig},
			{A: b, B: mid, Signature: sig},
		},
		Metas: append(nameMetas(), eventMetas()...),
	}
	canon := graphalign.Canon{
		Handles: []graphalign.Handle{
			{ID: ha, Ref: "PER-1", Kind: "person", Values: nameVals("Ann", "Ames")},
			{ID: hb, Ref: "PER-2", Kind: "person", Values: nameVals("Bob", "Ames")},
			{ID: hMid, Ref: "EVT-1", Kind: "event", Values: vals(et, etv)},
			{ID: hWrong, Ref: "EVT-9", Kind: "event", Values: vals(et, etv)},
		},
		Edges: []graphalign.CanonEdge{
			{From: ha, To: hMid, Signature: sig},
			{From: hb, To: hMid, Signature: sig},
			{From: ha, To: hWrong, Signature: other},
			{From: hb, To: hWrong, Signature: other},
		},
	}
	stats := graphalign.Stats{FanOut: map[string]float64{sig.Key(): 1, other.Key(): 1}}
	row := rowBySubject(graphalign.Align(layer, canon, stats, nil, nil), mid)
	cfg := graphalign.DefaultConfig()
	term := math.Log(cfg.MPrior["term"] / cfg.UPrior)
	name := cfg.PropertyScale[match.Property{Key: "name", Origin: "provenencia"}].Weight
	want := term + 2*(cfg.EdgeSupportLow+cfg.NeighborCreditLow*name)
	if row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, hMid) ||
		math.Abs(row.Score-want) > 1e-9 || hasAlt(row, hWrong) {
		t.Fatalf("nominated neighbor %+v, want %s at %.4f", row, hMid, want)
	}
}

func TestRankLoserDoesNotNominate(t *testing.T) {
	// The exact name wins the row. Anne still clears Rank, and only that
	// handle bridges to the third subject. Walking the loser would nominate
	// it; walking the winner must not.
	sig := partSig("subject", "event", "gathering")
	person, mid := id("s-person"), id("s-mid")
	win, lose, hMid := id("h-win"), id("h-lose"), id("h-mid")
	et, etv := termProp("event_type", "gathering")
	lesser := nameVals("Ann", "Ames")
	lesser[match.Property{Key: "name", Origin: "provenencia"}][0].Name.Parts[0].Value = "Anne"
	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: person, Ref: "PER-A", Kind: "person", Values: nameVals("Ann", "Ames")},
			{ID: mid, Ref: "EVT-M", Kind: "event", Values: vals(et, etv)},
		},
		Bridges: []graphalign.Bridge{{A: person, B: mid, Signature: sig}},
		Metas:   append(nameMetas(), eventMetas()...),
	}
	canon := graphalign.Canon{
		Handles: []graphalign.Handle{
			{ID: win, Ref: "PER-1", Kind: "person", Values: nameVals("Ann", "Ames")},
			{ID: lose, Ref: "PER-9", Kind: "person", Values: lesser},
			{ID: hMid, Ref: "EVT-1", Kind: "event", Values: vals(et, etv)},
		},
		Edges: []graphalign.CanonEdge{{From: lose, To: hMid, Signature: sig}},
	}
	p := graphalign.Align(layer, canon, graphalign.Stats{}, nil, nil)
	personRow := rowBySubject(p, person)
	if personRow.Target != graphalign.TargetHandle || !bytes.Equal(personRow.HandleID, win) || !hasAlt(personRow, lose) {
		t.Fatalf("winner %+v, want %s with %s still suggested", personRow, win, lose)
	}
	row := rowBySubject(p, mid)
	if row.Target == graphalign.TargetHandle || hasAlt(row, hMid) {
		t.Fatalf("loser nominated a neighbor: %+v", row)
	}
}

func TestNeighborCreditStopsAtOneHop(t *testing.T) {
	// The accepted name nominates a neighbor with no comparable values.
	// That neighbor's score includes the name, so it clears the bar and
	// nominates the subject beyond it. The far subject gets only the
	// structural bonus: the middle has no properties, and the name does
	// not travel a second hop.
	sig := relSig("kin")
	person, mid, far := id("s-person"), id("s-mid"), id("s-far")
	hp, hm, hf := id("h-person"), id("h-mid"), id("h-far")
	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: person, Ref: "PER-A", Kind: "person", Values: nameVals("Ann", "Ames")},
			{ID: mid, Ref: "PER-M", Kind: "person"},
			{ID: far, Ref: "PER-F", Kind: "person"},
		},
		Bridges: []graphalign.Bridge{
			{A: person, B: mid, Signature: sig},
			{A: mid, B: far, Signature: sig},
		},
		Metas: nameMetas(),
	}
	canon := graphalign.Canon{
		Handles: []graphalign.Handle{
			{ID: hp, Ref: "PER-1", Kind: "person", Values: nameVals("Ann", "Ames")},
			{ID: hm, Ref: "PER-2", Kind: "person"},
			{ID: hf, Ref: "PER-3", Kind: "person"},
		},
		Edges: []graphalign.CanonEdge{
			{From: hp, To: hm, Signature: sig},
			{From: hm, To: hf, Signature: sig},
		},
	}
	stats := graphalign.Stats{FanOut: map[string]float64{sig.Key(): 1}}
	p := graphalign.Align(layer, canon, stats, nil, nil)
	cfg := graphalign.DefaultConfig()
	name := cfg.PropertyScale[match.Property{Key: "name", Origin: "provenencia"}].Weight
	midWant := cfg.EdgeSupportLow + cfg.NeighborCreditLow*name
	midRow := rowBySubject(p, mid)
	if midRow.Target != graphalign.TargetHandle || !bytes.Equal(midRow.HandleID, hm) ||
		math.Abs(midRow.Score-midWant) > 1e-9 {
		t.Fatalf("middle %+v, want %s at %.2f", midRow, hm, midWant)
	}
	farRow := rowBySubject(p, far)
	if farRow.Target == graphalign.TargetHandle || !hasAlt(farRow, hf) ||
		math.Abs(farRow.Score-cfg.EdgeSupportLow) > 1e-9 || farRow.Score >= cfg.AcceptScore {
		t.Fatalf("far hop %+v, want %s suggested at %.2f", farRow, hf, cfg.EdgeSupportLow)
	}
}

func TestHeldNewAndSkipDecisions(t *testing.T) {
	sig := relSig("spouse")
	tests := []struct {
		name  string
		hold  graphalign.Target
		check func(t *testing.T, p graphalign.Proposal)
	}{
		{
			name: "a row held New frees the handle for its twin",
			hold: graphalign.TargetNew,
			check: func(t *testing.T, p graphalign.Proposal) {
				x, y := rowBySubject(p, id("s-x")), rowBySubject(p, id("s-y"))
				if x.Target != graphalign.TargetNew {
					t.Fatalf("held row %+v, want New", x)
				}
				if y.Target != graphalign.TargetHandle || !bytes.Equal(y.HandleID, id("h-x")) {
					t.Fatalf("twin %+v, want the freed handle", y)
				}
			},
		},
		{
			name: "a row held Skip is not walked through",
			hold: graphalign.TargetSkip,
			check: func(t *testing.T, p graphalign.Proposal) {
				if x := rowBySubject(p, id("s-x")); x.Target != graphalign.TargetSkip {
					t.Fatalf("held row %+v, want Skip", x)
				}
				if z := rowBySubject(p, id("s-z")); z.Via != nil && bytes.Equal(z.Via.NeighborSubjectID, id("s-x")) {
					t.Fatalf("walk went through the skipped row: %+v", z.Via)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// anchor —spouse— x (held) —spouse— z; y is x's twin with no edges.
			layer := graphalign.Layer{
				Subjects: []graphalign.Subject{
					{ID: id("s-a"), Ref: "PER-A", Kind: "person", Values: nameVals("Ann", "Ames")},
					{ID: id("s-x"), Ref: "PER-X", Kind: "person", Values: nameVals("Xavier", "Xu")},
					{ID: id("s-y"), Ref: "PER-Y", Kind: "person", Values: nameVals("Xavier", "Xu")},
					{ID: id("s-z"), Ref: "PER-Z", Kind: "person"},
				},
				Bridges: []graphalign.Bridge{
					{A: id("s-a"), B: id("s-x"), Signature: sig},
					{A: id("s-x"), B: id("s-z"), Signature: sig},
				},
				Metas: nameMetas(),
			}
			canon := graphalign.Canon{
				Handles: []graphalign.Handle{
					{ID: id("h-a"), Ref: "PER-1", Kind: "person", Values: nameVals("Ann", "Ames")},
					{ID: id("h-x"), Ref: "PER-2", Kind: "person", Values: nameVals("Xavier", "Xu")},
					{ID: id("h-z"), Ref: "PER-3", Kind: "person"},
				},
				Edges: []graphalign.CanonEdge{
					{From: id("h-a"), To: id("h-x"), Signature: sig},
					{From: id("h-x"), To: id("h-z"), Signature: sig},
				},
			}
			fixed := []graphalign.Fixed{
				{SubjectID: id("s-a"), HandleID: id("h-a")},
				{SubjectID: id("s-x"), Target: tt.hold},
			}
			tt.check(t, graphalign.Align(layer, canon, graphalign.Stats{}, fixed, nil))
		})
	}
}

func TestDecidedRowTheNeighborsContradict(t *testing.T) {
	sig := partSig("subject", "event", "birth")
	et, etv := termProp("event_type", "birth")
	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: id("s-person"), Ref: "PER-S", Kind: "person", Values: nameVals("Gracie", "Gates")},
			{ID: id("s-birth"), Ref: "EVT-S", Kind: "event", Values: vals(et, etv)},
		},
		Bridges: []graphalign.Bridge{{A: id("s-person"), B: id("s-birth"), Signature: sig}},
		Metas:   append(nameMetas(), eventMetas()...),
	}
	canon := graphalign.Canon{
		Handles: []graphalign.Handle{
			{ID: id("h-chosen"), Ref: "PER-1", Kind: "person", Values: nameVals("Mabel", "Moss")},
			{ID: id("h-gracie"), Ref: "PER-2", Kind: "person", Values: nameVals("Gracie", "Gates")},
			{ID: id("h-birth"), Ref: "EVT-1", Kind: "event", Values: vals(et, etv)},
		},
		Edges: []graphalign.CanonEdge{{From: id("h-gracie"), To: id("h-birth"), Signature: sig}},
	}
	fixed := []graphalign.Fixed{
		{SubjectID: id("s-person"), HandleID: id("h-chosen")},
		{SubjectID: id("s-birth"), HandleID: id("h-birth")},
	}
	row := rowBySubject(graphalign.Align(layer, canon, graphalign.Stats{}, fixed, nil), id("s-person"))
	if !bytes.Equal(row.HandleID, id("h-chosen")) {
		t.Fatalf("decided row moved: %+v", row)
	}
	if !row.Flags.ConflictWithFixed || len(row.Alternatives) == 0 || !bytes.Equal(row.Alternatives[0].HandleID, id("h-gracie")) {
		t.Fatalf("want a conflict naming PER-2 first, got flags %+v alts %+v", row.Flags, row.Alternatives)
	}
	alt := row.Alternatives[0]
	if alt.Assessment == "" || (alt.Reason != graphalign.ReasonAgrees && alt.Reason != graphalign.ReasonVia) {
		t.Fatalf("the other record should carry its own band, got %+v", alt)
	}
	if row.Assessment == graphalign.AssessStrong {
		t.Fatalf("a conflicting decision should not read strong: %+v", row)
	}
}

func TestPossibleDuplicates(t *testing.T) {
	tests := []struct {
		name    string
		handles []graphalign.Handle
		want    map[string]string // subject → duplicate of
	}{
		{
			name:    "the row that lost a handle names the row that took it",
			handles: []graphalign.Handle{{ID: id("h-1"), Ref: "PER-1", Kind: "person", Values: nameVals("John", "Smith")}},
			want:    map[string]string{"s-b": "s-a"},
		},
		{
			name: "two New rows that match each other name each other",
			want: map[string]string{"s-a": "s-b", "s-b": "s-a"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layer := graphalign.Layer{
				Subjects: []graphalign.Subject{
					{ID: id("s-a"), Ref: "PER-A", Kind: "person", Values: nameVals("John", "Smith")},
					{ID: id("s-b"), Ref: "PER-B", Kind: "person", Values: nameVals("John", "Smith")},
				},
				Metas: nameMetas(),
			}
			p := graphalign.Align(layer, graphalign.Canon{Handles: tt.handles}, graphalign.Stats{}, nil, nil)
			for _, r := range p.Rows {
				want, flagged := tt.want[string(r.SubjectID)]
				if r.Flags.PossibleDuplicate != flagged || string(r.Flags.DuplicateOf) != want {
					t.Fatalf("%s: flags %+v, want duplicate of %q", r.SubjectID, r.Flags, want)
				}
				if r.Target == graphalign.TargetSkip && flagged && r.Reason != graphalign.ReasonTaken {
					t.Fatalf("%s: reason %q, want taken", r.SubjectID, r.Reason)
				}
			}
		})
	}
}

func TestNewEventsDisagreeWhenTheirParticipantsDo(t *testing.T) {
	part := partSig("subject", "person", "")
	place := graphalign.EdgeSignature{BridgeType: "location", RoleOrType: "took_place_in", NeighborKind: "place"}
	et, residence := termProp("event_type", "residence")
	hatP, hatV := textProp("toponym", "Medicine Hat")
	calP, calV := textProp("toponym", "Calgary")
	base := []graphalign.Subject{
		{ID: id("e-a"), Ref: "EVT-A", Kind: "event", Values: vals(et, residence)},
		{ID: id("e-b"), Ref: "EVT-B", Kind: "event", Values: vals(et, residence)},
		{ID: id("p-paty"), Ref: "PER-A", Kind: "person", Values: nameVals("Paty", "Robins")},
		{ID: id("p-sandra"), Ref: "PER-B", Kind: "person", Values: nameVals("Sandra", "Robins")},
	}
	places := []graphalign.Subject{
		{ID: id("l-hat"), Ref: "PLC-H", Kind: "place", Values: vals(hatP, hatV)},
		{ID: id("l-calgary"), Ref: "PLC-C", Kind: "place", Values: vals(calP, calV)},
	}
	tests := []struct {
		name     string
		bridges  []graphalign.Bridge
		subjects []graphalign.Subject
		dup      bool
	}{
		{
			name: "different participants",
			bridges: []graphalign.Bridge{
				{A: id("p-paty"), B: id("e-a"), Signature: part},
				{A: id("p-sandra"), B: id("e-b"), Signature: part},
			},
			subjects: base,
		},
		{
			name: "the same participant",
			bridges: []graphalign.Bridge{
				{A: id("p-paty"), B: id("e-a"), Signature: part},
				{A: id("p-paty"), B: id("e-b"), Signature: part},
			},
			subjects: base,
			dup:      true,
		},
		{name: "no participants", subjects: base, dup: true},
		{
			name: "the same participant and different places",
			bridges: []graphalign.Bridge{
				{A: id("p-paty"), B: id("e-a"), Signature: part},
				{A: id("p-paty"), B: id("e-b"), Signature: part},
				{A: id("e-a"), B: id("l-hat"), Signature: place},
				{A: id("e-b"), B: id("l-calgary"), Signature: place},
			},
			subjects: append(append([]graphalign.Subject{}, base...), places...),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layer := graphalign.Layer{
				Subjects: tt.subjects,
				Bridges:  tt.bridges,
				Metas:    append(nameMetas(), append(eventMetas(), placeMetas()...)...),
			}
			p := graphalign.Align(layer, graphalign.Canon{}, graphalign.Stats{}, nil, nil)
			for _, sub := range []string{"e-a", "e-b"} {
				row := rowBySubject(p, id(sub))
				if row.Flags.PossibleDuplicate != tt.dup {
					t.Fatalf("%s duplicate %v, want %v (%+v)", sub, row.Flags.PossibleDuplicate, tt.dup, row.Flags)
				}
			}
		})
	}
}

func TestRowReasons(t *testing.T) {
	nameKey := match.Property{Key: "name", Origin: "provenencia"}
	sig := relSig("spouse")
	tests := []struct {
		name     string
		layer    []graphalign.Subject
		handles  []graphalign.Handle
		bridges  []graphalign.Bridge
		edges    []graphalign.CanonEdge
		fixed    []graphalign.Fixed
		subject  string
		want     graphalign.Reason
		wantProp match.Property
	}{
		{
			name:    "a property-only match names the agreeing property",
			layer:   []graphalign.Subject{{ID: id("s-a"), Ref: "A", Kind: "person", Values: nameVals("Ann", "Ames")}},
			handles: []graphalign.Handle{{ID: id("h-a"), Ref: "PER-1", Kind: "person", Values: nameVals("Ann", "Ames")}},
			subject: "s-a", want: graphalign.ReasonAgrees, wantProp: nameKey,
		},
		{
			name:    "a candidate below the bar is weak",
			layer:   []graphalign.Subject{{ID: id("s-a"), Ref: "A", Kind: "person", Values: nameVals("Ann", "Ames")}},
			handles: []graphalign.Handle{{ID: id("h-a"), Ref: "PER-1", Kind: "person", Values: nameVals("Bob", "Ames")}},
			subject: "s-a", want: graphalign.ReasonWeak,
		},
		{
			name:    "no candidate is no match",
			layer:   []graphalign.Subject{{ID: id("s-a"), Ref: "A", Kind: "person", Values: nameVals("Ann", "Ames")}},
			subject: "s-a", want: graphalign.ReasonNoMatch,
		},
		{
			name:    "no values is empty",
			layer:   []graphalign.Subject{{ID: id("s-a"), Ref: "A", Kind: "person"}},
			subject: "s-a", want: graphalign.ReasonEmpty,
		},
		{
			name: "a neighbor-reached row is via",
			layer: []graphalign.Subject{
				{ID: id("s-a"), Ref: "A", Kind: "person", Values: nameVals("Ann", "Ames")},
				{ID: id("s-b"), Ref: "B", Kind: "person", Values: nameVals("Bob", "Ames")},
			},
			handles: []graphalign.Handle{
				{ID: id("h-a"), Ref: "PER-1", Kind: "person", Values: nameVals("Ann", "Ames")},
				{ID: id("h-b"), Ref: "PER-2", Kind: "person", Values: nameVals("Bob", "Ames")},
			},
			bridges: []graphalign.Bridge{{A: id("s-a"), B: id("s-b"), Signature: sig}},
			edges:   []graphalign.CanonEdge{{From: id("h-a"), To: id("h-b"), Signature: sig}},
			fixed:   []graphalign.Fixed{{SubjectID: id("s-a"), HandleID: id("h-a")}},
			subject: "s-b", want: graphalign.ReasonVia,
		},
		{
			name:    "a held row is decided",
			layer:   []graphalign.Subject{{ID: id("s-a"), Ref: "A", Kind: "person", Values: nameVals("Ann", "Ames")}},
			fixed:   []graphalign.Fixed{{SubjectID: id("s-a"), Target: graphalign.TargetSkip}},
			subject: "s-a", want: graphalign.ReasonDecided,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layer := graphalign.Layer{Subjects: tt.layer, Bridges: tt.bridges, Metas: nameMetas()}
			canon := graphalign.Canon{Handles: tt.handles, Edges: tt.edges}
			row := rowBySubject(graphalign.Align(layer, canon, graphalign.Stats{}, tt.fixed, nil), id(tt.subject))
			if row.Reason != tt.want || row.ReasonProperty != tt.wantProp {
				t.Fatalf("reason %q / %+v, want %q / %+v", row.Reason, row.ReasonProperty, tt.want, tt.wantProp)
			}
		})
	}
}

func partOfSig() graphalign.EdgeSignature {
	return graphalign.EdgeSignature{
		BridgeType: "place_relationship", RoleOrType: "part_of",
		NeighborKind: "place", Directed: true,
	}
}

func TestPartOfChainIsStrongWithoutADecision(t *testing.T) {
	sig := partOfSig()
	city, state, country := id("s-city"), id("s-state"), id("s-country")
	hCity, hState, hCountry := id("h-city"), id("h-state"), id("h-country")
	cn, cv := textProp("toponym", "Gumptiontown")
	sn, sv := textProp("toponym", "Provenance")
	un, uv := textProp("toponym", "Arcadia")
	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: city, Ref: "CPL-C", Kind: "place", Values: vals(cn, cv)},
			{ID: state, Ref: "CPL-S", Kind: "place", Values: vals(sn, sv)},
			{ID: country, Ref: "CPL-U", Kind: "place", Values: vals(un, uv)},
		},
		Bridges: []graphalign.Bridge{
			{A: city, B: state, Signature: sig},
			{A: state, B: country, Signature: sig},
		},
		Metas: placeMetas(),
	}
	canon := graphalign.Canon{
		Handles: []graphalign.Handle{
			{ID: hCity, Ref: "PLC-C", Kind: "place", Values: vals(cn, cv)},
			{ID: hState, Ref: "PLC-S", Kind: "place", Values: vals(sn, sv)},
			{ID: hCountry, Ref: "PLC-U", Kind: "place", Values: vals(un, uv)},
		},
		Edges: []graphalign.CanonEdge{
			{From: hCity, To: hState, Signature: sig},
			{From: hState, To: hCountry, Signature: sig},
		},
	}
	stats := graphalign.Stats{FanOut: map[string]float64{sig.Key(): 1}}
	p := graphalign.Align(layer, canon, stats, nil, nil)
	cfg := graphalign.DefaultConfig()
	for _, pair := range []struct{ sub, handle []byte }{
		{city, hCity}, {state, hState}, {country, hCountry},
	} {
		row := rowBySubject(p, pair.sub)
		if row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, pair.handle) ||
			row.Assessment != graphalign.AssessStrong || row.Score < cfg.StrongScore ||
			row.Reason != graphalign.ReasonAgrees || row.ReasonProperty.Key != "toponym" {
			t.Fatalf("%s: %+v", pair.sub, row)
		}
	}
}

func TestPartOfBreaksASharedToponym(t *testing.T) {
	// Both Springfields score the same on the name. PLC-A sorts first, and
	// it is the one whose parent is not Illinois. The part-of link has to
	// move the row onto PLC-Z.
	sig := partOfSig()
	spring, ill := id("s-spring"), id("s-ill")
	hWrong, hRight := id("h-wrong"), id("h-right")
	hIll, hOhio := id("h-ill"), id("h-ohio")
	sn, sv := textProp("toponym", "Springfield")
	in, iv := textProp("toponym", "Illinois")
	on, ov := textProp("toponym", "Ohio")
	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: spring, Ref: "CPL-S", Kind: "place", Values: vals(sn, sv)},
			{ID: ill, Ref: "CPL-I", Kind: "place", Values: vals(in, iv)},
		},
		Bridges: []graphalign.Bridge{{A: spring, B: ill, Signature: sig}},
		Metas:   placeMetas(),
	}
	canon := graphalign.Canon{
		Handles: []graphalign.Handle{
			{ID: hWrong, Ref: "PLC-A", Kind: "place", Values: vals(sn, sv)},
			{ID: hRight, Ref: "PLC-Z", Kind: "place", Values: vals(sn, sv)},
			{ID: hIll, Ref: "PLC-I", Kind: "place", Values: vals(in, iv)},
			{ID: hOhio, Ref: "PLC-O", Kind: "place", Values: vals(on, ov)},
		},
		Edges: []graphalign.CanonEdge{
			{From: hWrong, To: hOhio, Signature: sig},
			{From: hRight, To: hIll, Signature: sig},
		},
	}
	stats := graphalign.Stats{FanOut: map[string]float64{sig.Key(): 1}}
	p := graphalign.Align(layer, canon, stats, nil, nil)
	cfg := graphalign.DefaultConfig()
	got := rowBySubject(p, spring)
	if got.Target != graphalign.TargetHandle || !bytes.Equal(got.HandleID, hRight) ||
		got.Assessment != graphalign.AssessStrong || got.Score < cfg.StrongScore {
		t.Fatalf("springfield %+v, want %s", got, hRight)
	}
	parent := rowBySubject(p, ill)
	if parent.Target != graphalign.TargetHandle || !bytes.Equal(parent.HandleID, hIll) ||
		parent.Assessment != graphalign.AssessStrong || parent.Score < cfg.StrongScore {
		t.Fatalf("illinois %+v", parent)
	}
}

func TestDirectedRelationshipsMatchFromTheSameEnd(t *testing.T) {
	tests := []struct {
		name     string
		directed bool
		want     string
	}{
		// PER-0 is the anchor's parent, PER-9 its child; Bob is the anchor's
		// child. Undirected, the tie goes to the first ref (the parent).
		{name: "a symmetric term reaches either end", directed: false, want: "h-parent"},
		{name: "a directed term reaches only the same end", directed: true, want: "h-child"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig := relSig("parent_of")
			sig.Directed = tt.directed
			layer := graphalign.Layer{
				Subjects: []graphalign.Subject{
					{ID: id("s-a"), Ref: "A", Kind: "person", Values: nameVals("Ann", "Ames")},
					{ID: id("s-b"), Ref: "B", Kind: "person", Values: nameVals("Bob", "Ames")},
				},
				Bridges: []graphalign.Bridge{{A: id("s-a"), B: id("s-b"), Signature: sig}}, // Ann parent_of Bob
				Metas:   nameMetas(),
			}
			canon := graphalign.Canon{
				Handles: []graphalign.Handle{
					{ID: id("h-a"), Ref: "PER-5", Kind: "person", Values: nameVals("Ann", "Ames")},
					{ID: id("h-parent"), Ref: "PER-0", Kind: "person", Values: nameVals("Bob", "Ames")},
					{ID: id("h-child"), Ref: "PER-9", Kind: "person", Values: nameVals("Bob", "Ames")},
				},
				Edges: []graphalign.CanonEdge{
					{From: id("h-parent"), To: id("h-a"), Signature: sig}, // PER-0 parent_of Ann
					{From: id("h-a"), To: id("h-child"), Signature: sig},  // Ann parent_of PER-9
				},
			}
			fixed := []graphalign.Fixed{{SubjectID: id("s-a"), HandleID: id("h-a")}}
			row := rowBySubject(graphalign.Align(layer, canon, graphalign.Stats{}, fixed, nil), id("s-b"))
			if row.Target != graphalign.TargetHandle || string(row.HandleID) != tt.want || row.Reason != graphalign.ReasonVia {
				t.Fatalf("got %s %q (%s), want %q via the anchor", row.Target, row.HandleID, row.Reason, tt.want)
			}
		})
	}
}

// A pair the walk accepted on a neighbor's edge loses that support when
// refinement moves the neighbor to another handle: the score is rescored, not
// kept at its old high, and the row no longer reads "via" that neighbor.
func TestRefinementDropsSupportFromAMovedNeighbor(t *testing.T) {
	merge := func(vs ...match.Values) match.Values {
		out := match.Values{}
		for _, v := range vs {
			for k, x := range v {
				out[k] = append(out[k], x...)
			}
		}
		return out
	}
	sp, female := termProp("sex_at_birth", "female")
	_, male := termProp("sex_at_birth", "male")
	fem, mal := vals(sp, female), vals(sp, male)
	metas := append(nameMetas(), personMetas()[1])
	spouse, sibling := relSig("spouse"), relSig("sibling")

	ann, husband, carl, dora := id("s-ann"), id("s-husband"), id("s-carl"), id("s-dora")
	annA, annB, bob, hCarl, hDora := id("h-ann-a"), id("h-ann-b"), id("h-bob"), id("h-carl"), id("h-dora")
	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: ann, Ref: "S-ANN", Kind: "person", Values: merge(nameVals("Ann", "Lee"), fem)},
			{ID: husband, Ref: "S-HUSBAND", Kind: "person", Values: mal}, // unnamed
			{ID: carl, Ref: "S-CARL", Kind: "person", Values: nameVals("Carl", "Lee")},
			{ID: dora, Ref: "S-DORA", Kind: "person", Values: nameVals("Dora", "Lee")},
		},
		Bridges: []graphalign.Bridge{
			{A: ann, B: husband, Signature: spouse},
			{A: ann, B: carl, Signature: sibling},
			{A: ann, B: dora, Signature: sibling},
		},
		Metas: metas,
	}
	// Two Ann Lees: A is married to Bob; B has Carl and Dora for siblings,
	// so the page's Ann is B once her siblings map.
	canon := graphalign.Canon{
		Handles: []graphalign.Handle{
			{ID: annA, Ref: "A", Kind: "person", Values: merge(nameVals("Ann", "Lee"), fem)},
			{ID: annB, Ref: "B", Kind: "person", Values: merge(nameVals("Ann", "Lee"), fem)},
			{ID: bob, Ref: "C", Kind: "person", Values: merge(nameVals("Bob", "Ray"), mal)},
			{ID: hCarl, Ref: "D", Kind: "person", Values: nameVals("Carl", "Lee")},
			{ID: hDora, Ref: "E", Kind: "person", Values: nameVals("Dora", "Lee")},
		},
		Edges: []graphalign.CanonEdge{
			{From: annA, To: bob, Signature: spouse},
			{From: annB, To: hCarl, Signature: sibling},
			{From: annB, To: hDora, Signature: sibling},
		},
	}
	stats := graphalign.Stats{FanOut: map[string]float64{spouse.Key(): 1, sibling.Key(): 1}}
	p := graphalign.Align(layer, canon, stats, nil, nil)

	if r := rowBySubject(p, ann); !bytes.Equal(r.HandleID, annB) {
		t.Fatalf("Ann on %s, want the Ann with Carl and Dora", r.HandleID)
	}
	r := rowBySubject(p, husband)
	if bytes.Equal(r.HandleID, bob) || r.Reason == graphalign.ReasonVia {
		t.Fatalf("husband: target=%s handle=%s score=%.2f reason=%s; Bob is the other Ann's spouse",
			r.Target, r.HandleID, r.Score, r.Reason)
	}
	for _, alt := range r.Alternatives {
		if bytes.Equal(alt.HandleID, bob) && (alt.Assessment != graphalign.AssessNone || alt.Reason == graphalign.ReasonVia) {
			t.Fatalf("Bob still offered as %s via Ann (score %.2f)", alt.Assessment, alt.Score)
		}
	}
}

// Two New rows with one name, where one record names his son and the other
// his father: those neighbors play different roles, so they neither confirm
// nor rule out the pair, and the duplicate warning stands on the names.
func TestNewDuplicateCheckComparesDirectedNeighborsByEnd(t *testing.T) {
	parent := relSig("parent")
	parent.Directed = true
	a, b, son, father := id("s-a"), id("s-b"), id("s-son"), id("s-father")
	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: a, Ref: "S-A", Kind: "person", Values: nameVals("John", "Smith")},
			{ID: b, Ref: "S-B", Kind: "person", Values: nameVals("John", "Smith")},
			{ID: son, Ref: "S-SON", Kind: "person", Values: nameVals("Robert", "Smith")},
			{ID: father, Ref: "S-FATHER", Kind: "person", Values: nameVals("William", "Smith")},
		},
		Bridges: []graphalign.Bridge{
			{A: a, B: son, Signature: parent},    // a is the son's parent
			{A: father, B: b, Signature: parent}, // b is the father's child
		},
		Metas: nameMetas(),
	}
	p := graphalign.Align(layer, graphalign.Canon{}, graphalign.Stats{}, nil, nil)
	for _, sid := range [][]byte{a, b} {
		r := rowBySubject(p, sid)
		if r.Target != graphalign.TargetNew || !r.Flags.PossibleDuplicate {
			t.Fatalf("%s: target=%s duplicate=%v, want New and a possible duplicate", sid, r.Target, r.Flags.PossibleDuplicate)
		}
	}

	// The same end still rules a pair out: two Johns whose sons differ.
	layer.Subjects[3] = graphalign.Subject{ID: father, Ref: "S-OTHERSON", Kind: "person", Values: nameVals("William", "Smith")}
	layer.Bridges[1] = graphalign.Bridge{A: b, B: father, Signature: parent} // b is William's parent
	p = graphalign.Align(layer, graphalign.Canon{}, graphalign.Stats{}, nil, nil)
	if r := rowBySubject(p, a); r.Flags.PossibleDuplicate {
		t.Fatalf("Johns with different sons still flagged as one")
	}
}
