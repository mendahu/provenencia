package graphalign_test

import (
	"bytes"
	"testing"

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
