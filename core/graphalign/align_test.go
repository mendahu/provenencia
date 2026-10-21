package graphalign_test

import (
	"bytes"
	"maps"
	"math"
	"slices"
	"testing"

	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/datevalue/datevaluetest"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/namevalue"
	"github.com/mendahu/provenencia/core/propertycompare"
	"github.com/mendahu/provenencia/core/propertycompare/propertycomparetest"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/segmentcompare"
	"github.com/mendahu/provenencia/core/segmentcompare/segmentcomparetest"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/subjectcompare/subjectcomparetest"
	"github.com/mendahu/provenencia/core/subjectprofile"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

func id(s string) []byte { return []byte(s) }

func textProp(key, text string) (vocabkey.PropertyKey, propertycompare.Value) {
	return vocabkey.PropertyKey(key), propertycomparetest.Text(text)
}

func dateProp(year int) (vocabkey.PropertyKey, propertycompare.Value) {
	return vocabkey.PropertyDate, propertycomparetest.Date(datevaluetest.YMD(year, 0, 0))
}

func termProp(key, term string) (vocabkey.PropertyKey, propertycompare.Value) {
	return vocabkey.PropertyKey(key), propertycomparetest.Term(term)
}

func vals(pairs ...any) subjectcompare.Values {
	out := subjectcompare.Values{}
	for i := 0; i+1 < len(pairs); i += 2 {
		p := pairs[i].(vocabkey.PropertyKey)
		v := pairs[i+1].(propertycompare.Value)
		out[p] = append(out[p], v)
	}
	return out
}

func metasOf(valueTypes map[vocabkey.PropertyKey]string) subjectcompare.Metas {
	out := subjectcompare.Metas{}
	for k, t := range valueTypes {
		out[k] = propertymeta.Meta{ValueType: t, Cardinality: propertymeta.CardinalitySingle}
	}
	return out
}

func personMetas() subjectcompare.Metas {
	return metasOf(map[vocabkey.PropertyKey]string{
		vocabkey.PropertyName:       propertymeta.ValueTypeText,
		vocabkey.PropertySexAtBirth: propertymeta.ValueTypeTerm,
	})
}

func placeMetas() subjectcompare.Metas {
	return metasOf(map[vocabkey.PropertyKey]string{
		vocabkey.PropertyToponym: propertymeta.ValueTypeText,
	})
}

func eventMetas() subjectcompare.Metas {
	return metasOf(map[vocabkey.PropertyKey]string{
		vocabkey.PropertyEventType: propertymeta.ValueTypeTerm,
		vocabkey.PropertyDate:      propertymeta.ValueTypeDate,
	})
}

// mergeMetas is the union of metas.
func mergeMetas(ms ...subjectcompare.Metas) subjectcompare.Metas {
	out := subjectcompare.Metas{}
	for _, m := range ms {
		maps.Copy(out, m)
	}
	return out
}

// link is one bridge read from both ends: a is the key at the first end
// (Bridge.A, CanonEdge.From), b at the second.
type link struct{ a, b vocabkey.BridgeKey }

func bridge(a, b []byte, l link) graphalign.Bridge {
	return graphalign.Bridge{A: a, B: b, KeyA: l.a, KeyB: l.b}
}

func edge(from, to []byte, l link) graphalign.CanonEdge {
	return graphalign.CanonEdge{From: from, To: to, KeyFrom: l.a, KeyTo: l.b}
}

// partKeys are a participation read from the person's end and the event's,
// built through segmentcompare.Key over the seeded terms.
func partKeys(role vocabkey.BridgeTerm, eventType vocabkey.EventType) (person, event vocabkey.BridgeKey) {
	terms := segmentcomparetest.Terms()
	person = segmentcompare.Key(segmentcompare.Link{Type: vocabkey.BridgeParticipation, Term: role,
		End: vocabkey.PropertyPerson, Other: vocabkey.PropertyEvent, EventType: eventType}, terms)
	event = segmentcompare.Key(segmentcompare.Link{Type: vocabkey.BridgeParticipation, Term: role,
		End: vocabkey.PropertyEvent, Other: vocabkey.PropertyPerson, EventType: eventType}, terms)
	return person, event
}

// relKeys are a relationship read from the person's end and the related_to
// end.
func relKeys(term vocabkey.BridgeTerm) (a, b vocabkey.BridgeKey) {
	terms := segmentcomparetest.Terms()
	a = segmentcompare.Key(segmentcompare.Link{Type: vocabkey.BridgeRelationship, Term: term,
		End: vocabkey.PropertyPerson, Other: vocabkey.PropertyRelatedTo, SameType: true}, terms)
	b = segmentcompare.Key(segmentcompare.Link{Type: vocabkey.BridgeRelationship, Term: term,
		End: vocabkey.PropertyRelatedTo, Other: vocabkey.PropertyPerson, SameType: true}, terms)
	return a, b
}

// partOfKeys are part_of read from the part's end (from) and the whole's (to).
func partOfKeys() (part, whole vocabkey.BridgeKey) {
	terms := segmentcomparetest.Terms()
	part = segmentcompare.Key(segmentcompare.Link{Type: vocabkey.BridgePlaceRelationship, Term: vocabkey.TermPartOf,
		End: vocabkey.PropertyFrom, Other: vocabkey.PropertyTo, SameType: true}, terms)
	whole = segmentcompare.Key(segmentcompare.Link{Type: vocabkey.BridgePlaceRelationship, Term: vocabkey.TermPartOf,
		End: vocabkey.PropertyTo, Other: vocabkey.PropertyFrom, SameType: true}, terms)
	return part, whole
}

func partLink(role vocabkey.BridgeTerm, eventType vocabkey.EventType) link {
	a, b := partKeys(role, eventType)
	return link{a, b}
}

func relLink(term vocabkey.BridgeTerm) link {
	a, b := relKeys(term)
	return link{a, b}
}

// locLink is a location read from the event's end and the place's.
func locLink() link {
	terms := segmentcomparetest.Terms()
	return link{
		segmentcompare.Key(segmentcompare.Link{Type: vocabkey.BridgeLocation, End: vocabkey.PropertyEvent, Other: vocabkey.PropertyPlace}, terms),
		segmentcompare.Key(segmentcompare.Link{Type: vocabkey.BridgeLocation, End: vocabkey.PropertyPlace, Other: vocabkey.PropertyEvent}, terms),
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
	sig := partLink("subject", "birth")

	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: evSub, Ref: "EVT-S", Kind: "event", Values: vals(et, etv)},
			{ID: perSub, Ref: "PER-S", Kind: "person", Values: vals(np, nv)},
		},
		Bridges: []graphalign.Bridge{bridge(perSub, evSub, sig)},
		Metas:   mergeMetas(personMetas(), eventMetas()),
	}
	canon := graphalign.Canon{
		Handles: []graphalign.Handle{
			{ID: evH, Ref: "EVT-1", Kind: "event", Values: vals(et, etv)},
			{ID: perH, Ref: "PER-1", Kind: "person", Values: vals(np, nv)},
		},
		Edges: []graphalign.CanonEdge{edge(perH, evH, sig)},
	}
	fixed := []graphalign.Fixed{{SubjectID: evSub, HandleID: evH}}
	p := alignWith(layer, canon, fixed, matchregistry.Default().AlignConfig())
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
	sig := partLink("subject", "birth")

	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: evSub, Ref: "EVT-S", Kind: "event", Values: vals(et, etv)},
			{ID: perSub, Ref: "PER-S", Kind: "person", Values: vals(np, nv)},
		},
		Bridges: []graphalign.Bridge{bridge(perSub, evSub, sig)},
		Metas:   mergeMetas(personMetas(), eventMetas()),
	}
	canon := graphalign.Canon{
		Handles: []graphalign.Handle{
			{ID: evH, Ref: "EVT-1", Kind: "event", Values: vals(et, etv)},
			{ID: perH, Ref: "PER-1", Kind: "person", Values: vals(np, nv)},
		},
		Edges: []graphalign.CanonEdge{edge(perH, evH, sig)},
	}
	fixed := []graphalign.Fixed{{SubjectID: perSub, HandleID: perH}}
	p := alignWith(layer, canon, fixed, matchregistry.Default().AlignConfig())
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
	sig := relLink("spouse")
	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: a, Ref: "PER-A", Kind: "person", Values: vals(na, va)},
			{ID: b, Ref: "PER-B", Kind: "person", Values: vals(nb, vb)},
		},
		Bridges: []graphalign.Bridge{bridge(a, b, sig)},
		Metas:   personMetas(),
	}
	canon := graphalign.Canon{
		Handles: []graphalign.Handle{
			{ID: ha, Ref: "PER-1", Kind: "person", Values: vals(na, va)},
			{ID: hb, Ref: "PER-2", Kind: "person", Values: vals(nb, vb)},
		},
		Edges: []graphalign.CanonEdge{edge(ha, hb, sig)},
	}
	fixed := []graphalign.Fixed{{SubjectID: a, HandleID: ha}}
	p := alignWith(layer, canon, fixed, matchregistry.Default().AlignConfig())
	if row := rowBySubject(p, b); row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, hb) {
		t.Fatalf("spouse: %+v", row)
	}
}

func TestOneAgreeingPropertyStaysAWeakMatch(t *testing.T) {
	// An exact toponym alone is its scale's 3 points: medium, short of strong.
	orphan := id("s-orphan")
	known := id("h-known")
	n, v := textProp("toponym", "York")
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
	row := rowBySubject(alignWith(layer, canon, nil, matchregistry.Default().AlignConfig()), orphan)
	if row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, known) ||
		row.Assessment != graphalign.AssessMedium || row.Reason != graphalign.ReasonAgrees ||
		row.ReasonProperty != vocabkey.PropertyToponym || math.Abs(row.Score-3) > 1e-9 {
		t.Fatalf("same toponym: %+v", row)
	}
	if row.Score >= matchregistry.Default().AlignConfig().Alignment.Bands.Strong {
		t.Fatalf("toponym alone is not strong: score %.2f", row.Score)
	}

	// Without the toponym scale, the text weight (2) is a weak match.
	cfg := matchregistry.Default().AlignConfig()
	cfg.Subject.Weights[vocabkey.Place][vocabkey.PropertyToponym] = subjectcompare.Weights{Agree: 2, Conflict: 3}
	row = rowBySubject(alignWith(layer, canon, nil, cfg), orphan)
	if row.Assessment != graphalign.AssessWeak {
		t.Fatalf("toponym without the registry weight: %+v", row)
	}

	other, ov := textProp("toponym", "Leeds")
	canon.Handles[0].Values = vals(other, ov)
	row = rowBySubject(alignWith(layer, canon, nil, matchregistry.Default().AlignConfig()), orphan)
	if row.Target == graphalign.TargetHandle || row.Assessment != graphalign.AssessNone {
		t.Fatalf("different toponym should not match: %+v", row)
	}
}

func TestSpellingVariantIsAWeakerTextMatch(t *testing.T) {
	// One added letter on a toponym is a match, weaker than the exact spelling: about 0.91 × 3, which is weak, not medium.
	orphan := id("s-orphan")
	known := id("h-known")
	probe, pv := textProp("toponym", "Provenance")
	canonProp, cv := textProp("toponym", "Provenanced")
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
	cfg := matchregistry.Default().AlignConfig()
	row := rowBySubject(alignWith(layer, canon, nil, matchregistry.Default().AlignConfig()), orphan)
	if row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, known) ||
		row.Assessment != graphalign.AssessWeak || row.Reason != graphalign.ReasonAgrees ||
		row.ReasonProperty != vocabkey.PropertyToponym ||
		row.Score < cfg.Alignment.Bands.Weak || row.Score >= cfg.Alignment.Bands.Medium {
		t.Fatalf("spelling variant: %+v", row)
	}
	var partial bool
	for _, c := range row.Comparisons {
		if c.Key != vocabkey.PropertyToponym {
			continue
		}
		if c.Result.Outcome != propertycompare.OutcomePartial {
			t.Fatalf("comparison %+v", c)
		}
		partial = true
	}
	if !partial {
		t.Fatal("missing toponym comparison")
	}

	// Without the toponym scale, the text weight × the resemblance stays under weak. The toponym scale is what makes a spelling variant clear the bar.
	bare := matchregistry.Default().AlignConfig()
	bare.Subject.Weights[vocabkey.Place][vocabkey.PropertyToponym] = subjectcompare.Weights{Agree: 2, Conflict: 3}
	row = rowBySubject(alignWith(layer, canon, nil, bare), orphan)
	if row.Target == graphalign.TargetHandle || row.Assessment != graphalign.AssessNone {
		t.Fatalf("spelling variant without the registry weight: %+v", row)
	}

	// Sharing one word of a longer toponym is a resemblance, not a match.
	york, yv := textProp("toponym", "York")
	longer, lv := textProp("toponym", "York, Upper Canada")
	layer.Metas = placeMetas()
	layer.Subjects[0].Values = vals(york, yv)
	canon.Handles[0].Values = vals(longer, lv)
	row = rowBySubject(alignWith(layer, canon, nil, matchregistry.Default().AlignConfig()), orphan)
	if row.Target == graphalign.TargetHandle || row.Assessment != graphalign.AssessNone {
		t.Fatalf("shared word should stay under the bar: %+v", row)
	}
}

func TestOrphanMatchesOnItsPropertiesOrSkips(t *testing.T) {
	// Place orphans use toponym + text value compare in Candidates (names need namevalues).
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
	p := alignWith(layer, canon, nil, matchregistry.Default().AlignConfig())
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
	p2 := alignWith(layer2, canon2, nil, matchregistry.Default().AlignConfig())
	row2 := rowBySubject(p2, weak)
	if row2.Target == graphalign.TargetHandle {
		t.Fatalf("weak orphan should not merge: %+v", row2)
	}
}

func TestConfigOverrideWithoutEditingRegistry(t *testing.T) {
	cfg := matchregistry.Default().AlignConfig()
	cfg.Alignment.AcceptBar = 100 // impossible
	evSub := id("s-event")
	perSub := id("s-person")
	evH := id("h-event")
	perH := id("h-person")
	np, nv := textProp("name", "Gracie")
	et, etv := termProp("event_type", "birth")
	sig := partLink("subject", "birth")
	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: evSub, Ref: "EVT-S", Kind: "event", Values: vals(et, etv)},
			{ID: perSub, Ref: "PER-S", Kind: "person", Values: vals(np, nv)},
		},
		Bridges: []graphalign.Bridge{bridge(perSub, evSub, sig)},
		Metas:   mergeMetas(personMetas(), eventMetas()),
	}
	canon := graphalign.Canon{
		Handles: []graphalign.Handle{
			{ID: evH, Ref: "EVT-1", Kind: "event", Values: vals(et, etv)},
			{ID: perH, Ref: "PER-1", Kind: "person", Values: vals(np, nv)},
		},
		Edges: []graphalign.CanonEdge{edge(perH, evH, sig)},
	}
	fixed := []graphalign.Fixed{{SubjectID: evSub, HandleID: evH}}
	p := alignWith(layer, canon, fixed, cfg)
	row := rowBySubject(p, perSub)
	// The link path is refused; the person may still match on its own properties.
	if row.Target == graphalign.TargetHandle && bytes.Equal(row.HandleID, perH) {
		// matching on its own properties is OK
		return
	}
	if row.Target == graphalign.TargetHandle {
		t.Fatalf("unexpected handle: %+v", row)
	}
}

// gracieObituary is the obituary golden's layer and canon.
func gracieObituary() (graphalign.Layer, graphalign.Canon) {
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

	birthSig := partLink("subject", "birth")
	deathSig := partLink("subject", "death")
	burialSig := partLink("subject", "burial")
	childSig := relLink("parent")

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
			bridge(gracie, birth, birthSig),
			bridge(gracie, death, deathSig),
			bridge(gracie, burial, burialSig),
			bridge(gracie, child1, childSig),
			bridge(gracie, child2, childSig),
		},
		Metas: mergeMetas(personMetas(), eventMetas(), placeMetas()),
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
			edge(hGracie, hBirth, birthSig),
			edge(hGracie, hDeath, deathSig),
			edge(hGracie, hBurial, burialSig),
			edge(hGracie, hChild1, childSig),
			edge(hGracie, hChild2, childSig),
		},
	}
	return layer, canon
}

func TestGracieObituaryGolden(t *testing.T) {
	gracie, birth, death, burial := id("s-gracie"), id("s-birth"), id("s-death"), id("s-burial")
	child1, child2, res1, medHat := id("s-child1"), id("s-child2"), id("s-res1"), id("s-medhat")
	hGracie, hBirth, hDeath, hChild1, hChild2 := id("h-gracie"), id("h-birth"), id("h-death"), id("h-child1"), id("h-child2")
	layer, canon := gracieObituary()
	fixed := []graphalign.Fixed{{SubjectID: gracie, HandleID: hGracie}}

	p1 := alignWith(layer, canon, fixed, matchregistry.Default().AlignConfig())
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
	// dogfood: the burial (event type only) is an Other link and no longer
	// files; revisit when the whole algorithm is dogfooded (Spike 11 matching
	// dogfood).
	if r := rowBySubject(p1, burial); r.Target == graphalign.TargetHandle {
		t.Fatalf("burial: want not a handle, got %+v", r)
	}
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
	p2 := alignWith(layer, canon, fixed, matchregistry.Default().AlignConfig())
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

// nameOf is a name property value.
func nameOf(n *namevalue.Value) propertycompare.Value {
	return propertycompare.Value{Value: valuecompare.Value{Name: n}}
}

// propResult is the result for one lead property of a subject compare.
func propResult(t *testing.T, r subjectcompare.Result, key vocabkey.PropertyKey) subjectcompare.PropertyResult {
	t.Helper()
	for _, p := range r.Properties {
		if p.Key == key {
			return p
		}
	}
	t.Fatalf("no result for %s", key)
	return subjectcompare.PropertyResult{}
}

func nameVals(given, surname string) subjectcompare.Values {
	p := vocabkey.PropertyKey(vocabkey.PropertyName)
	return subjectcompare.Values{p: {nameOf(&namevalue.Value{
		Form: given + " " + surname,
		Parts: []namevalue.Part{
			{Idx: 0, Type: namevalue.PartTypeGiven, Value: given},
			{Idx: 1, Type: namevalue.PartTypeSurname, Value: surname},
		},
	})}}
}

// surnameVals is a name with a surname part and nothing else.
func surnameVals(surname string) subjectcompare.Values {
	p := vocabkey.PropertyKey(vocabkey.PropertyName)
	return subjectcompare.Values{p: {nameOf(&namevalue.Value{
		Form:  surname,
		Parts: []namevalue.Part{{Idx: 0, Type: namevalue.PartTypeSurname, Value: surname}},
	})}}
}

func nameMetas() subjectcompare.Metas {
	return metasOf(map[vocabkey.PropertyKey]string{vocabkey.PropertyName: propertymeta.ValueTypeName})
}

func TestIsolatedSubjectScoresOnTheWalkScale(t *testing.T) {
	tests := []struct {
		name       string
		probe      subjectcompare.Values
		handle     subjectcompare.Values
		wantTarget graphalign.Target
		wantAssess graphalign.Assessment
		wantReason graphalign.Reason
	}{
		{
			name:  "a different first name is no candidate",
			probe: nameVals("Mary", "Robins"), handle: nameVals("James", "Robins"),
			wantTarget: graphalign.TargetNew, wantAssess: graphalign.AssessNone, wantReason: graphalign.ReasonNoMatch,
		},
		{
			name:  "the same name alone is a weak match, never strong",
			probe: nameVals("Mary", "Robins"), handle: nameVals("Mary", "Robins"),
			wantTarget: graphalign.TargetHandle, wantAssess: graphalign.AssessWeak, wantReason: graphalign.ReasonAgrees,
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
			row := rowBySubject(alignWith(layer, canon, nil, matchregistry.Default().AlignConfig()), s)
			if row.Target != tt.wantTarget || row.Assessment != tt.wantAssess || row.Reason != tt.wantReason {
				t.Fatalf("got %s / %s / %s (score %.2f), want %s / %s / %s", row.Target, row.Assessment, row.Reason, row.Score, tt.wantTarget, tt.wantAssess, tt.wantReason)
			}
		})
	}
}

func TestPartialNamesAreWeak(t *testing.T) {
	given := func(i int, v string) namevalue.Part {
		return namevalue.Part{Idx: i, Type: namevalue.PartTypeGiven, Value: v}
	}
	sur := func(i int, v string) namevalue.Part {
		return namevalue.Part{Idx: i, Type: namevalue.PartTypeSurname, Value: v}
	}
	full := func(parts ...namevalue.Part) subjectcompare.Values {
		p := vocabkey.PropertyKey(vocabkey.PropertyName)
		return subjectcompare.Values{p: {nameOf(&namevalue.Value{Parts: parts})}}
	}
	alignOne := func(probe, handle subjectcompare.Values) graphalign.Row {
		s, h := id("s"), id("h")
		return rowBySubject(alignWith(
			graphalign.Layer{Subjects: []graphalign.Subject{{ID: s, Ref: "SUB", Kind: "person", Values: probe}}, Metas: nameMetas()},
			graphalign.Canon{Handles: []graphalign.Handle{{ID: h, Ref: "PER-1", Kind: "person", Values: handle}}},
			nil, matchregistry.Default().AlignConfig(),
		), s)
	}
	james := alignOne(
		full(given(0, "James"), sur(1, "Robins")),
		full(given(0, "James"), given(1, "Kenneth"), sur(2, "Robins")),
	)
	if james.Target != graphalign.TargetHandle || james.Assessment != graphalign.AssessWeak || james.Score >= matchregistry.Default().AlignConfig().Alignment.Bands.Medium {
		t.Fatalf("James / James Kenneth: %+v", james)
	}
	// 2·G1 + 2·F of those plus the other given name G: 5 / 5.5 of the name weight.
	if want := 2.5 * 5 / 5.5; math.Abs(james.Score-want) > 1e-9 {
		t.Fatalf("James / James Kenneth scored %.4f, want %.4f", james.Score, want)
	}
	lee := alignOne(
		full(given(0, "Lee"), sur(1, "Breakell")),
		full(given(0, "Lee-Ellen"), given(1, "Matilda"), sur(2, "Breakell")),
	)
	if lee.Target != graphalign.TargetHandle || lee.Assessment != graphalign.AssessWeak || lee.Score >= matchregistry.Default().AlignConfig().Alignment.Bands.Medium {
		t.Fatalf("Lee / Lee-Ellen: %+v", lee)
	}
	// The stored forms disagree. The parts are what match: surname first on
	// the certificate, given name first on the new subject.
	withForm := func(base subjectcompare.Values, form string) subjectcompare.Values {
		out := subjectcompare.Values{}
		for k, vs := range base {
			cp := append([]propertycompare.Value(nil), vs...)
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

	sex := vocabkey.PropertyKey(vocabkey.PropertySexAtBirth)
	metas := mergeMetas(nameMetas(), metasOf(map[vocabkey.PropertyKey]string{sex: propertymeta.ValueTypeTerm}))
	withSex := func(base subjectcompare.Values, term string) subjectcompare.Values {
		out := subjectcompare.Values{}
		for k, v := range base {
			out[k] = v
		}
		out[sex] = []propertycompare.Value{propertycomparetest.Term(term)}
		return out
	}
	probe := full(given(0, "James"), sur(1, "Robins"))
	handle := full(given(0, "James"), given(1, "Kenneth"), sur(2, "Robins"))
	s, h := id("s"), id("h")
	mismatch := rowBySubject(alignWith(
		graphalign.Layer{Subjects: []graphalign.Subject{{ID: s, Ref: "SUB", Kind: "person", Values: withSex(probe, "female")}}, Metas: metas},
		graphalign.Canon{Handles: []graphalign.Handle{{ID: h, Ref: "PER-1", Kind: "person", Values: withSex(handle, "male")}}},
		[]graphalign.Fixed{{SubjectID: s, HandleID: h}}, matchregistry.Default().AlignConfig(),
	), s)
	if diff := james.Score - mismatch.Score; diff < 0.99 || diff > 1.01 {
		t.Fatalf("sex mismatch subtracted %.3f, want 1 (name %.3f, with sex %.3f)", diff, james.Score, mismatch.Score)
	}
}

func TestMissingGivenOrPrefixIsNotANameConflict(t *testing.T) {
	name := vocabkey.PropertyKey(vocabkey.PropertyName)
	parts := func(ps ...namevalue.Part) subjectcompare.Values {
		return subjectcompare.Values{name: {nameOf(&namevalue.Value{Parts: ps})}}
	}
	mrs := parts(
		namevalue.Part{Idx: 0, Type: namevalue.PartTypePrefix, Value: "Mrs."},
		namevalue.Part{Idx: 1, Type: namevalue.PartTypeSurname, Value: "Robins"},
	)
	marion := parts(
		namevalue.Part{Idx: 0, Type: namevalue.PartTypeGiven, Value: "Marion"},
		namevalue.Part{Idx: 1, Type: namevalue.PartTypeSurname, Value: "Robins"},
	)
	cfg := matchregistry.Default().AlignConfig()
	r := subjectcompare.Compare(vocabkey.Person, mrs, marion, nameMetas(), cfg.Subject, cfg.Value)
	sc := graphalign.ScoreCandidate(r, 0, 1)
	nameRes := propResult(t, sc.Result, vocabkey.PropertyName).Result
	c := nameRes.Speaking.Result
	// The title is ignored. Marion exists only on one side, so the name is a
	// resemblance: the surnames earn 2·F of 2·F + G1.
	want := 2 * 1.5 / (2*1.5 + 1)
	if o := nameRes.Outcome; o != propertycompare.OutcomePartial || math.Abs(c.Similarity-want) > 1e-9 {
		t.Fatalf("Mrs. Robins / Marion Robins: outcome %s similarity %.4f, want partial %.4f", o, c.Similarity, want)
	}
	if math.Abs(sc.Score-cfg.Subject.Weights[vocabkey.Person][vocabkey.PropertyName].Agree*want) > 1e-9 || sc.Score >= cfg.Alignment.Bands.Medium {
		t.Fatalf("surname with a one-sided given name scored %.4f, want the name weight times %.2f under medium", sc.Score, want)
	}

	differ := graphalign.ScoreCandidate(
		subjectcompare.Compare(vocabkey.Person, nameVals("Mary", "Robins"), nameVals("James", "Robins"), nameMetas(), cfg.Subject, cfg.Value),
		0, 1,
	)
	if o := propResult(t, differ.Result, vocabkey.PropertyName).Result.Outcome; o != propertycompare.OutcomeConflict || differ.Score >= 0 {
		t.Fatalf("different given names: outcome %s score %.3f", o, differ.Score)
	}
}

func TestSexAgreementAloneStaysUnderWeak(t *testing.T) {
	sex := vocabkey.PropertyKey(vocabkey.PropertySexAtBirth)
	meta := metasOf(map[vocabkey.PropertyKey]string{sex: propertymeta.ValueTypeTerm})
	male := subjectcompare.Values{sex: {propertycomparetest.Term("male")}}
	cfg := matchregistry.Default().AlignConfig()
	sc := graphalign.ScoreCandidate(subjectcompare.Compare(vocabkey.Person, male, male, meta, cfg.Subject, cfg.Value), 0, 1)
	if sc.Score < 0.59 || sc.Score > 0.61 || sc.Score >= matchregistry.Default().AlignConfig().Alignment.Bands.Weak {
		t.Fatalf("sex alone scored %.3f", sc.Score)
	}
	s, h := id("s"), id("h")
	row := rowBySubject(alignWith(
		graphalign.Layer{Subjects: []graphalign.Subject{{ID: s, Ref: "SUB", Kind: "person", Values: male}}, Metas: meta},
		graphalign.Canon{Handles: []graphalign.Handle{{ID: h, Ref: "PER-1", Kind: "person", Values: male}}},
		nil, matchregistry.Default().AlignConfig(),
	), s)
	if row.Target == graphalign.TargetHandle || row.Assessment != graphalign.AssessNone {
		t.Fatalf("sex alone published %+v", row)
	}
}

func TestSharedHandleIsDecidedTheSameEveryRun(t *testing.T) {
	tests := []struct {
		name   string
		a, b   subjectcompare.Values
		winner string
	}{
		{name: "equal claims go to the first ref", a: nameVals("John", "Smith"), b: nameVals("John", "Smith"), winner: "s-a"},
		{name: "the stronger claim wins whatever its ref", a: nameVals("Jon", "Smith"), b: nameVals("John", "Smith"), winner: "s-b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sex := vocabkey.PropertyKey(vocabkey.PropertySexAtBirth)
			male := propertycomparetest.Term("male")
			b := tt.b
			if tt.winner == "s-b" {
				b = subjectcompare.Values{}
				for k, v := range tt.b {
					b[k] = v
				}
				b[sex] = []propertycompare.Value{male}
			}
			handle := nameVals("John", "Smith")
			handle[sex] = []propertycompare.Value{male}
			metas := mergeMetas(nameMetas(), metasOf(map[vocabkey.PropertyKey]string{sex: propertymeta.ValueTypeTerm}))
			layer := graphalign.Layer{
				Subjects: []graphalign.Subject{
					{ID: id("s-a"), Ref: "PER-A", Kind: "person", Values: tt.a},
					{ID: id("s-b"), Ref: "PER-B", Kind: "person", Values: b},
				},
				Metas: metas,
			}
			canon := graphalign.Canon{Handles: []graphalign.Handle{{ID: id("h-1"), Ref: "PER-1", Kind: "person", Values: handle}}}
			for run := 0; run < 50; run++ {
				p := alignWith(layer, canon, nil, matchregistry.Default().AlignConfig())
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
	sig := partLink("subject", "birth")
	evSub, perSub, evH, perH := id("s-event"), id("s-person"), id("h-event"), id("h-person")
	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: evSub, Ref: "EVT-S", Kind: "event"},
			{ID: perSub, Ref: "PER-S", Kind: "person"},
		},
		Bridges: []graphalign.Bridge{bridge(perSub, evSub, sig)},
		Metas:   personMetas(),
	}
	canon := graphalign.Canon{
		Handles: []graphalign.Handle{{ID: evH, Ref: "EVT-1", Kind: "event"}, {ID: perH, Ref: "PER-1", Kind: "person"}},
		Edges:   []graphalign.CanonEdge{edge(perH, evH, sig)},
	}
	fixed := []graphalign.Fixed{{SubjectID: evSub, HandleID: evH}}
	row := rowBySubject(alignWith(layer, canon, fixed, matchregistry.Default().AlignConfig()), perSub)
	cfg := matchregistry.Default().AlignConfig()
	if row.Score != cfg.Segment.OneSupport {
		t.Fatalf("score %.2f, want one edge's support %.2f", row.Score, cfg.Segment.OneSupport)
	}
	if row.Target == graphalign.TargetHandle {
		t.Fatalf("one edge and no shared values should not merge: %+v", row)
	}
}

func TestExactNameCreditsTheNeighborOnce(t *testing.T) {
	sig := partLink("subject", "birth")
	evSub, perSub, evH, perH := id("s-event"), id("s-person"), id("h-event"), id("h-person")
	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: evSub, Ref: "EVT-S", Kind: "event"},
			{ID: perSub, Ref: "PER-S", Kind: "person", Values: nameVals("Ann", "Ames")},
		},
		Bridges: []graphalign.Bridge{bridge(perSub, evSub, sig)},
		Metas:   mergeMetas(nameMetas(), eventMetas()),
	}
	canon := graphalign.Canon{
		Handles: []graphalign.Handle{
			{ID: evH, Ref: "EVT-1", Kind: "event"},
			{ID: perH, Ref: "PER-1", Kind: "person", Values: nameVals("Ann", "Ames")},
		},
		Edges: []graphalign.CanonEdge{edge(perH, evH, sig)},
	}
	fixed := []graphalign.Fixed{{SubjectID: perSub, HandleID: perH}}
	row := rowBySubject(alignWith(layer, canon, fixed, matchregistry.Default().AlignConfig()), evSub)
	cfg := matchregistry.Default().AlignConfig()
	name := cfg.Subject.Weights[vocabkey.Person][vocabkey.PropertyName].Agree
	want := cfg.Segment.OneSupport + cfg.Segment.OneCredit*name
	if row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, evH) ||
		math.Abs(row.Score-want) > 1e-9 || row.Assessment != graphalign.AssessWeak {
		t.Fatalf("event %+v, want %s at %.2f", row, evH, want)
	}
}

func TestOtherLinkCreditStaysUnderTheBar(t *testing.T) {
	sig := partLink("subject", "residence")
	evSub, perSub, evH, perH := id("s-event"), id("s-person"), id("h-event"), id("h-person")
	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: evSub, Ref: "EVT-S", Kind: "event"},
			{ID: perSub, Ref: "PER-S", Kind: "person", Values: nameVals("Ann", "Ames")},
		},
		Bridges: []graphalign.Bridge{bridge(perSub, evSub, sig)},
		Metas:   mergeMetas(nameMetas(), eventMetas()),
	}
	canon := graphalign.Canon{
		Handles: []graphalign.Handle{
			{ID: evH, Ref: "EVT-1", Kind: "event"},
			{ID: perH, Ref: "PER-1", Kind: "person", Values: nameVals("Ann", "Ames")},
		},
		Edges: []graphalign.CanonEdge{edge(perH, evH, sig)},
	}
	fixed := []graphalign.Fixed{{SubjectID: perSub, HandleID: perH}}
	row := rowBySubject(alignWith(layer, canon, fixed, matchregistry.Default().AlignConfig()), evSub)
	cfg := matchregistry.Default().AlignConfig()
	name := cfg.Subject.Weights[vocabkey.Person][vocabkey.PropertyName].Agree
	want := cfg.Segment.OtherSupport + cfg.Segment.OtherCredit*name
	if row.Target == graphalign.TargetHandle || math.Abs(row.Score-want) > 1e-9 || row.Score >= cfg.Alignment.AcceptBar {
		t.Fatalf("event %+v, want a suggestion at %.2f", row, want)
	}
}

func TestConflictingNeighborLowersTheScore(t *testing.T) {
	sig := partLink("subject", "birth")
	evSub, perSub, evH, perH := id("s-event"), id("s-person"), id("h-event"), id("h-person")
	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: evSub, Ref: "EVT-S", Kind: "event"},
			{ID: perSub, Ref: "PER-S", Kind: "person", Values: nameVals("Mary", "Robins")},
		},
		Bridges: []graphalign.Bridge{bridge(perSub, evSub, sig)},
		Metas:   mergeMetas(nameMetas(), eventMetas()),
	}
	canon := graphalign.Canon{
		Handles: []graphalign.Handle{
			{ID: evH, Ref: "EVT-1", Kind: "event"},
			{ID: perH, Ref: "PER-1", Kind: "person", Values: nameVals("James", "Robins")},
		},
		Edges: []graphalign.CanonEdge{edge(perH, evH, sig)},
	}
	fixed := []graphalign.Fixed{{SubjectID: perSub, HandleID: perH}}
	row := rowBySubject(alignWith(layer, canon, fixed, matchregistry.Default().AlignConfig()), evSub)
	cfg := matchregistry.Default().AlignConfig()
	penalty := cfg.Subject.Weights[vocabkey.Person][vocabkey.PropertyName].Conflict
	want := cfg.Segment.OneSupport + cfg.Segment.OneCredit*(-penalty)
	if row.Target == graphalign.TargetHandle || math.Abs(row.Score-want) > 1e-9 || row.Score >= cfg.Segment.OneSupport {
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
	// is a shared term, under the weak band. Both accepted handles
	// bridge to it, and both canon edges share that bridge's signature, so
	// the walk nominates the neighbor. The score counts both edges and a
	// fraction of each neighbor's name. A canon neighbor on a different
	// signature is not nominated: it is offered only on its own properties,
	// with no support.
	sig := partLink("subject", "gathering")
	other := partLink("subject", "other")
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
			bridge(a, mid, sig),
			bridge(b, mid, sig),
		},
		Metas: mergeMetas(nameMetas(), eventMetas()),
	}
	canon := graphalign.Canon{
		Handles: []graphalign.Handle{
			{ID: ha, Ref: "PER-1", Kind: "person", Values: nameVals("Ann", "Ames")},
			{ID: hb, Ref: "PER-2", Kind: "person", Values: nameVals("Bob", "Ames")},
			{ID: hMid, Ref: "EVT-1", Kind: "event", Values: vals(et, etv)},
			{ID: hWrong, Ref: "EVT-9", Kind: "event", Values: vals(et, etv)},
		},
		Edges: []graphalign.CanonEdge{
			edge(ha, hMid, sig),
			edge(hb, hMid, sig),
			edge(ha, hWrong, other),
			edge(hb, hWrong, other),
		},
	}
	row := rowBySubject(alignWith(layer, canon, nil, matchregistry.Default().AlignConfig()), mid)
	cfg := matchregistry.Default().AlignConfig()
	term := cfg.Subject.Weights[vocabkey.Event][vocabkey.PropertyEventType].Agree
	name := cfg.Subject.Weights[vocabkey.Person][vocabkey.PropertyName].Agree
	want := term + 2*(cfg.Segment.OtherSupport+cfg.Segment.OtherCredit*name)
	if row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, hMid) ||
		math.Abs(row.Score-want) > 1e-9 {
		t.Fatalf("nominated neighbor %+v, want %s at %.4f", row, hMid, want)
	}
	if len(row.Alternatives) != 1 || !bytes.Equal(row.Alternatives[0].HandleID, hWrong) ||
		len(row.Alternatives[0].ViaNeighbor) != 0 || row.Alternatives[0].Assessment != graphalign.AssessNone ||
		math.Abs(row.Alternatives[0].Score-term) > 1e-9 {
		t.Fatalf("alternatives %+v, want only %s on its own properties at %.4f", row.Alternatives, hWrong, term)
	}
}

func TestCandidateLoserDoesNotNominate(t *testing.T) {
	// The exact name wins the row. Anne is still a candidate, and only that
	// handle bridges to the third subject. Walking the loser would nominate
	// it; walking the winner must not.
	sig := partLink("subject", "gathering")
	person, mid := id("s-person"), id("s-mid")
	win, lose, hMid := id("h-win"), id("h-lose"), id("h-mid")
	et, etv := termProp("event_type", "gathering")
	lesser := nameVals("Ann", "Ames")
	lesser[vocabkey.PropertyName][0].Name.Parts[0].Value = "Anne"
	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: person, Ref: "PER-A", Kind: "person", Values: nameVals("Ann", "Ames")},
			{ID: mid, Ref: "EVT-M", Kind: "event", Values: vals(et, etv)},
		},
		Bridges: []graphalign.Bridge{bridge(person, mid, sig)},
		Metas:   mergeMetas(nameMetas(), eventMetas()),
	}
	canon := graphalign.Canon{
		Handles: []graphalign.Handle{
			{ID: win, Ref: "PER-1", Kind: "person", Values: nameVals("Ann", "Ames")},
			{ID: lose, Ref: "PER-9", Kind: "person", Values: lesser},
			{ID: hMid, Ref: "EVT-1", Kind: "event", Values: vals(et, etv)},
		},
		Edges: []graphalign.CanonEdge{edge(lose, hMid, sig)},
	}
	p := alignWith(layer, canon, nil, matchregistry.Default().AlignConfig())
	personRow := rowBySubject(p, person)
	if personRow.Target != graphalign.TargetHandle || !bytes.Equal(personRow.HandleID, win) || !hasAlt(personRow, lose) {
		t.Fatalf("winner %+v, want %s with %s still suggested", personRow, win, lose)
	}
	// The person is placed on PER-1, so the gathering link to PER-9 is apart
	// and adds nothing: EVT-1 is offered only on its own properties.
	row := rowBySubject(p, mid)
	term := matchregistry.Default().AlignConfig().Subject.Weights[vocabkey.Event][vocabkey.PropertyEventType].Agree
	if row.Target != graphalign.TargetSkip || row.Reason != graphalign.ReasonWeak || row.Via != nil {
		t.Fatalf("loser nominated a neighbor: %+v", row)
	}
	if len(row.Alternatives) != 1 || !bytes.Equal(row.Alternatives[0].HandleID, hMid) ||
		len(row.Alternatives[0].ViaNeighbor) != 0 || row.Alternatives[0].Assessment != graphalign.AssessNone ||
		math.Abs(row.Alternatives[0].Score-term) > 1e-9 {
		t.Fatalf("alternatives %+v, want only %s on its own properties at %.4f", row.Alternatives, hMid, term)
	}
}

func TestNeighborCreditStopsAtOneHop(t *testing.T) {
	// The accepted name nominates a neighbor with no comparable values.
	// That neighbor's score includes the name, so it clears the bar and
	// nominates the subject beyond it. The far subject gets only the
	// structural bonus: the middle has no properties, and the name does
	// not travel a second hop.
	birth, loc := partLink("subject", "birth"), locLink()
	person, mid, far := id("s-person"), id("s-mid"), id("s-far")
	hp, hm, hf := id("h-person"), id("h-mid"), id("h-far")
	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: person, Ref: "PER-A", Kind: "person", Values: nameVals("Ann", "Ames")},
			{ID: mid, Ref: "EVT-M", Kind: "event"},
			{ID: far, Ref: "PLC-F", Kind: "place"},
		},
		Bridges: []graphalign.Bridge{
			bridge(person, mid, birth),
			bridge(mid, far, loc),
		},
		Metas: mergeMetas(nameMetas(), eventMetas(), placeMetas()),
	}
	canon := graphalign.Canon{
		Handles: []graphalign.Handle{
			{ID: hp, Ref: "PER-1", Kind: "person", Values: nameVals("Ann", "Ames")},
			{ID: hm, Ref: "EVT-2", Kind: "event"},
			{ID: hf, Ref: "PLC-3", Kind: "place"},
		},
		Edges: []graphalign.CanonEdge{
			edge(hp, hm, birth),
			edge(hm, hf, loc),
		},
	}
	p := alignWith(layer, canon, nil, matchregistry.Default().AlignConfig())
	cfg := matchregistry.Default().AlignConfig()
	name := cfg.Subject.Weights[vocabkey.Person][vocabkey.PropertyName].Agree
	midWant := cfg.Segment.OneSupport + cfg.Segment.OneCredit*name
	midRow := rowBySubject(p, mid)
	if midRow.Target != graphalign.TargetHandle || !bytes.Equal(midRow.HandleID, hm) ||
		math.Abs(midRow.Score-midWant) > 1e-9 {
		t.Fatalf("middle %+v, want %s at %.2f", midRow, hm, midWant)
	}
	farRow := rowBySubject(p, far)
	if farRow.Target == graphalign.TargetHandle || !hasAlt(farRow, hf) ||
		math.Abs(farRow.Score-cfg.Segment.OtherSupport) > 1e-9 || farRow.Score >= cfg.Alignment.AcceptBar {
		t.Fatalf("far hop %+v, want %s suggested at %.2f", farRow, hf, cfg.Segment.OtherSupport)
	}
}

func TestHeldNewAndSkipDecisions(t *testing.T) {
	sig := relLink("spouse")
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
					bridge(id("s-a"), id("s-x"), sig),
					bridge(id("s-x"), id("s-z"), sig),
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
					edge(id("h-a"), id("h-x"), sig),
					edge(id("h-x"), id("h-z"), sig),
				},
			}
			fixed := []graphalign.Fixed{
				{SubjectID: id("s-a"), HandleID: id("h-a")},
				{SubjectID: id("s-x"), Target: tt.hold},
			}
			tt.check(t, alignWith(layer, canon, fixed, matchregistry.Default().AlignConfig()))
		})
	}
}

func TestDecidedRowTheNeighborsContradict(t *testing.T) {
	sig := partLink("subject", "birth")
	et, etv := termProp("event_type", "birth")
	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: id("s-person"), Ref: "PER-S", Kind: "person", Values: nameVals("Gracie", "Gates")},
			{ID: id("s-birth"), Ref: "EVT-S", Kind: "event", Values: vals(et, etv)},
		},
		Bridges: []graphalign.Bridge{bridge(id("s-person"), id("s-birth"), sig)},
		Metas:   mergeMetas(nameMetas(), eventMetas()),
	}
	canon := graphalign.Canon{
		Handles: []graphalign.Handle{
			{ID: id("h-chosen"), Ref: "PER-1", Kind: "person", Values: nameVals("Mabel", "Moss")},
			{ID: id("h-gracie"), Ref: "PER-2", Kind: "person", Values: nameVals("Gracie", "Gates")},
			{ID: id("h-birth"), Ref: "EVT-1", Kind: "event", Values: vals(et, etv)},
		},
		Edges: []graphalign.CanonEdge{edge(id("h-gracie"), id("h-birth"), sig)},
	}
	fixed := []graphalign.Fixed{
		{SubjectID: id("s-person"), HandleID: id("h-chosen")},
		{SubjectID: id("s-birth"), HandleID: id("h-birth")},
	}
	row := rowBySubject(alignWith(layer, canon, fixed, matchregistry.Default().AlignConfig()), id("s-person"))
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
			p := alignWith(layer, graphalign.Canon{Handles: tt.handles}, nil, matchregistry.Default().AlignConfig())
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
	part := partLink("subject", "")
	place := locLink()
	et, residence := termProp("event_type", "residence")
	dp, dv := dateProp(1850)
	hatP, hatV := textProp("toponym", "Medicine Hat")
	calP, calV := textProp("toponym", "Calgary")
	base := []graphalign.Subject{
		{ID: id("e-a"), Ref: "EVT-A", Kind: "event", Values: vals(et, residence, dp, dv)},
		{ID: id("e-b"), Ref: "EVT-B", Kind: "event", Values: vals(et, residence, dp, dv)},
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
				bridge(id("p-paty"), id("e-a"), part),
				bridge(id("p-sandra"), id("e-b"), part),
			},
			subjects: base,
		},
		{
			name: "the same participant",
			bridges: []graphalign.Bridge{
				bridge(id("p-paty"), id("e-a"), part),
				bridge(id("p-paty"), id("e-b"), part),
			},
			subjects: base,
			dup:      true,
		},
		{name: "no participants", subjects: base, dup: true},
		{
			name: "the same participant and different places",
			bridges: []graphalign.Bridge{
				bridge(id("p-paty"), id("e-a"), part),
				bridge(id("p-paty"), id("e-b"), part),
				bridge(id("e-a"), id("l-hat"), place),
				bridge(id("e-b"), id("l-calgary"), place),
			},
			subjects: append(append([]graphalign.Subject{}, base...), places...),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layer := graphalign.Layer{
				Subjects: tt.subjects,
				Bridges:  tt.bridges,
				Metas:    mergeMetas(nameMetas(), eventMetas(), placeMetas()),
			}
			p := alignWith(layer, graphalign.Canon{}, nil, matchregistry.Default().AlignConfig())
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
	nameKey := vocabkey.PropertyKey(vocabkey.PropertyName)
	sig := relLink("spouse")
	tests := []struct {
		name     string
		layer    []graphalign.Subject
		handles  []graphalign.Handle
		bridges  []graphalign.Bridge
		edges    []graphalign.CanonEdge
		fixed    []graphalign.Fixed
		subject  string
		want     graphalign.Reason
		wantProp vocabkey.PropertyKey
	}{
		{
			name:    "a property-only match names the agreeing property",
			layer:   []graphalign.Subject{{ID: id("s-a"), Ref: "A", Kind: "person", Values: nameVals("Ann", "Ames")}},
			handles: []graphalign.Handle{{ID: id("h-a"), Ref: "PER-1", Kind: "person", Values: nameVals("Ann", "Ames")}},
			subject: "s-a", want: graphalign.ReasonAgrees, wantProp: nameKey,
		},
		{
			name: "a neighbor-reached candidate below the bar is weak",
			layer: []graphalign.Subject{
				{ID: id("s-a"), Ref: "A", Kind: "person", Values: nameVals("Ann", "Ames")},
				{ID: id("s-e"), Ref: "E", Kind: "event", Values: vals(termProp("event_type", "residence"))},
			},
			handles: []graphalign.Handle{
				{ID: id("h-a"), Ref: "PER-1", Kind: "person", Values: nameVals("Ann", "Ames")},
				{ID: id("h-e"), Ref: "EVT-1", Kind: "event", Values: vals(termProp("event_type", "residence"))},
			},
			bridges: []graphalign.Bridge{bridge(id("s-a"), id("s-e"), partLink("subject", "residence"))},
			edges:   []graphalign.CanonEdge{edge(id("h-a"), id("h-e"), partLink("subject", "residence"))},
			fixed:   []graphalign.Fixed{{SubjectID: id("s-a"), HandleID: id("h-a")}},
			subject: "s-e", want: graphalign.ReasonWeak,
		},
		{
			name:    "a surname-only look-alike is offered as weak",
			layer:   []graphalign.Subject{{ID: id("s-a"), Ref: "A", Kind: "person", Values: nameVals("Ann", "Ames")}},
			handles: []graphalign.Handle{{ID: id("h-a"), Ref: "PER-1", Kind: "person", Values: surnameVals("Ames")}},
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
			bridges: []graphalign.Bridge{bridge(id("s-a"), id("s-b"), sig)},
			edges:   []graphalign.CanonEdge{edge(id("h-a"), id("h-b"), sig)},
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
			layer := graphalign.Layer{Subjects: tt.layer, Bridges: tt.bridges, Metas: mergeMetas(nameMetas(), eventMetas())}
			canon := graphalign.Canon{Handles: tt.handles, Edges: tt.edges}
			row := rowBySubject(alignWith(layer, canon, tt.fixed, matchregistry.Default().AlignConfig()), id(tt.subject))
			if row.Reason != tt.want || row.ReasonProperty != tt.wantProp {
				t.Fatalf("reason %q / %+v, want %q / %+v", row.Reason, row.ReasonProperty, tt.want, tt.wantProp)
			}
		})
	}
}

func partOfLink() link {
	a, b := partOfKeys()
	return link{a, b}
}

// partOfChain is a city in a state in a country, each on both sides.
func partOfChain() (graphalign.Layer, graphalign.Canon) {
	sig := partOfLink()
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
			bridge(city, state, sig),
			bridge(state, country, sig),
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
			edge(hCity, hState, sig),
			edge(hState, hCountry, sig),
		},
	}
	return layer, canon
}

func TestPartOfChainIsStrongWithoutADecision(t *testing.T) {
	city, state, country := id("s-city"), id("s-state"), id("s-country")
	hCity, hState, hCountry := id("h-city"), id("h-state"), id("h-country")
	layer, canon := partOfChain()
	p := alignWith(layer, canon, nil, matchregistry.Default().AlignConfig())
	cfg := matchregistry.Default().AlignConfig()
	for _, pair := range []struct{ sub, handle []byte }{
		{city, hCity}, {state, hState}, {country, hCountry},
	} {
		row := rowBySubject(p, pair.sub)
		if row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, pair.handle) ||
			row.Assessment != graphalign.AssessStrong || row.Score < cfg.Alignment.Bands.Strong ||
			row.Reason != graphalign.ReasonAgrees || row.ReasonProperty != vocabkey.PropertyToponym {
			t.Fatalf("%s: %+v", pair.sub, row)
		}
	}
}

func TestPartOfDoesNotFileADifferentToponym(t *testing.T) {
	// Calgary and Edmonton are both part of Alberta. The shared parent must
	// not file Calgary onto Edmonton: a different toponym adds nothing, and
	// part_of only reinforces a property score that already cleared the bar.
	sig := partOfLink()
	child, parent := id("s-calgary"), id("s-alberta")
	hSibling, hParent := id("h-edmonton"), id("h-alberta")
	cn, cv := textProp("toponym", "Calgary")
	sn, sv := textProp("toponym", "Edmonton")
	pn, pv := textProp("toponym", "Alberta")
	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: child, Ref: "CPL-C", Kind: "place", Values: vals(cn, cv)},
			{ID: parent, Ref: "CPL-A", Kind: "place", Values: vals(pn, pv)},
		},
		Bridges: []graphalign.Bridge{bridge(child, parent, sig)},
		Metas: subjectcompare.Metas{cn: {
			ValueType: propertymeta.ValueTypeText, Cardinality: propertymeta.CardinalityMultiple,
		}},
	}
	canon := graphalign.Canon{
		Handles: []graphalign.Handle{
			{ID: hSibling, Ref: "PLC-E", Kind: "place", Values: vals(sn, sv)},
			{ID: hParent, Ref: "PLC-A", Kind: "place", Values: vals(pn, pv)},
		},
		Edges: []graphalign.CanonEdge{edge(hSibling, hParent, sig)},
	}
	fixed := []graphalign.Fixed{{SubjectID: parent, HandleID: hParent}}
	row := rowBySubject(alignWith(layer, canon, fixed, matchregistry.Default().AlignConfig()), child)
	cfg := matchregistry.Default().AlignConfig()
	if row.Target == graphalign.TargetHandle || bytes.Equal(row.HandleID, hSibling) || row.Score >= cfg.Alignment.AcceptBar {
		t.Fatalf("calgary %+v, want it off Edmonton", row)
	}
}

func TestPartOfBreaksASharedToponym(t *testing.T) {
	// Both Springfields score the same on the name. PLC-A sorts first, and
	// it is the one whose parent is not Illinois. The part-of link has to
	// move the row onto PLC-Z.
	sig := partOfLink()
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
		Bridges: []graphalign.Bridge{bridge(spring, ill, sig)},
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
			edge(hWrong, hOhio, sig),
			edge(hRight, hIll, sig),
		},
	}
	p := alignWith(layer, canon, nil, matchregistry.Default().AlignConfig())
	cfg := matchregistry.Default().AlignConfig()
	got := rowBySubject(p, spring)
	if got.Target != graphalign.TargetHandle || !bytes.Equal(got.HandleID, hRight) ||
		got.Assessment != graphalign.AssessStrong || got.Score < cfg.Alignment.Bands.Strong {
		t.Fatalf("springfield %+v, want %s", got, hRight)
	}
	parent := rowBySubject(p, ill)
	if parent.Target != graphalign.TargetHandle || !bytes.Equal(parent.HandleID, hIll) ||
		parent.Assessment != graphalign.AssessStrong || parent.Score < cfg.Alignment.Bands.Strong {
		t.Fatalf("illinois %+v", parent)
	}
	var wrong *graphalign.Alternative
	for i := range got.Alternatives {
		if bytes.Equal(got.Alternatives[i].HandleID, hWrong) {
			wrong = &got.Alternatives[i]
			break
		}
	}
	if wrong == nil || wrong.Score >= cfg.Alignment.AcceptBar {
		t.Fatalf("ohio springfield %+v, want it under the accept bar", got.Alternatives)
	}
}

func TestDirectedRelationshipsMatchFromTheSameEnd(t *testing.T) {
	tests := []struct {
		name string
		term vocabkey.BridgeTerm
		want string
	}{
		// PER-0 is the anchor's parent, PER-9 its child; Bob is the anchor's
		// child. A symmetric term, the tie goes to the first ref (the
		// parent).
		{name: "a symmetric term reaches either end", term: "spouse", want: "h-parent"},
		{name: "a directed term reaches only the same end", term: "parent", want: "h-child"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := relLink(tt.term)
			layer := graphalign.Layer{
				Subjects: []graphalign.Subject{
					{ID: id("s-a"), Ref: "A", Kind: "person", Values: nameVals("Ann", "Ames")},
					{ID: id("s-b"), Ref: "B", Kind: "person", Values: nameVals("Bob", "Ames")},
				},
				Bridges: []graphalign.Bridge{bridge(id("s-a"), id("s-b"), l)}, // Ann is Bob's term
				Metas:   nameMetas(),
			}
			canon := graphalign.Canon{
				Handles: []graphalign.Handle{
					{ID: id("h-a"), Ref: "PER-5", Kind: "person", Values: nameVals("Ann", "Ames")},
					{ID: id("h-parent"), Ref: "PER-0", Kind: "person", Values: nameVals("Bob", "Ames")},
					{ID: id("h-child"), Ref: "PER-9", Kind: "person", Values: nameVals("Bob", "Ames")},
				},
				Edges: []graphalign.CanonEdge{
					edge(id("h-parent"), id("h-a"), l), // PER-0 is Ann's term
					edge(id("h-a"), id("h-child"), l),  // Ann is PER-9's term
				},
			}
			fixed := []graphalign.Fixed{{SubjectID: id("s-a"), HandleID: id("h-a")}}
			row := rowBySubject(alignWith(layer, canon, fixed, matchregistry.Default().AlignConfig()), id("s-b"))
			if row.Target != graphalign.TargetHandle || string(row.HandleID) != tt.want || row.Reason != graphalign.ReasonVia {
				t.Fatalf("got %s %q (%s), want %q via the anchor", row.Target, row.HandleID, row.Reason, tt.want)
			}
		})
	}
}

// A pair the walk accepted on a neighbor's edge loses that support when
// settling moves the neighbor to another handle: the score is rescored, not
// kept at its old high, and the row no longer reads "via" that neighbor.
func TestSettlingDropsSupportFromAMovedNeighbor(t *testing.T) {
	merge := func(vs ...subjectcompare.Values) subjectcompare.Values {
		out := subjectcompare.Values{}
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
	metas := mergeMetas(nameMetas(), metasOf(map[vocabkey.PropertyKey]string{vocabkey.PropertySexAtBirth: propertymeta.ValueTypeTerm}))
	spouse, sibling := relLink("spouse"), relLink("sibling")

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
			bridge(ann, husband, spouse),
			bridge(ann, carl, sibling),
			bridge(ann, dora, sibling),
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
			edge(annA, bob, spouse),
			edge(annB, hCarl, sibling),
			edge(annB, hDora, sibling),
		},
	}
	p := alignWith(layer, canon, nil, matchregistry.Default().AlignConfig())

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
	parent := relLink("parent")
	a, b, son, father := id("s-a"), id("s-b"), id("s-son"), id("s-father")
	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: a, Ref: "S-A", Kind: "person", Values: nameVals("John", "Smith")},
			{ID: b, Ref: "S-B", Kind: "person", Values: nameVals("John", "Smith")},
			{ID: son, Ref: "S-SON", Kind: "person", Values: nameVals("Robert", "Smith")},
			{ID: father, Ref: "S-FATHER", Kind: "person", Values: nameVals("William", "Smith")},
		},
		Bridges: []graphalign.Bridge{
			bridge(a, son, parent),    // a is the son's parent
			bridge(father, b, parent), // b is the father's child
		},
		Metas: nameMetas(),
	}
	p := alignWith(layer, graphalign.Canon{}, nil, matchregistry.Default().AlignConfig())
	for _, sid := range [][]byte{a, b} {
		r := rowBySubject(p, sid)
		if r.Target != graphalign.TargetNew || !r.Flags.PossibleDuplicate {
			t.Fatalf("%s: target=%s duplicate=%v, want New and a possible duplicate", sid, r.Target, r.Flags.PossibleDuplicate)
		}
	}

	// The same end still rules a pair out: two Johns whose sons differ.
	layer.Subjects[3] = graphalign.Subject{ID: father, Ref: "S-OTHERSON", Kind: "person", Values: nameVals("William", "Smith")}
	layer.Bridges[1] = bridge(b, father, parent) // b is William's parent
	p = alignWith(layer, graphalign.Canon{}, nil, matchregistry.Default().AlignConfig())
	if r := rowBySubject(p, a); r.Flags.PossibleDuplicate {
		t.Fatalf("Johns with different sons still flagged as one")
	}
}

// A record that gives sex at birth as "unknown" says nothing about sex: an
// exact name still matches a handle of either sex.
func TestUnknownSexIsNoEvidence(t *testing.T) {
	merge := func(a, b subjectcompare.Values) subjectcompare.Values {
		for k, v := range b {
			a[k] = append(a[k], v...)
		}
		return a
	}
	sp, unknown := termProp("sex_at_birth", "unknown")
	unknown.Neutral = true // the seeded flag
	_, male := termProp("sex_at_birth", "male")
	s, h := id("s-john"), id("h-john")
	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{{ID: s, Ref: "S-J", Kind: "person", Values: merge(nameVals("John", "Smith"), vals(sp, unknown))}},
		Metas:    mergeMetas(nameMetas(), metasOf(map[vocabkey.PropertyKey]string{vocabkey.PropertySexAtBirth: propertymeta.ValueTypeTerm})),
	}
	canon := graphalign.Canon{Handles: []graphalign.Handle{
		{ID: h, Ref: "PER-1", Kind: "person", Values: merge(nameVals("John", "Smith"), vals(sp, male))},
	}}
	r := rowBySubject(alignWith(layer, canon, nil, matchregistry.Default().AlignConfig()), s)
	if r.Target != graphalign.TargetHandle || !bytes.Equal(r.HandleID, h) {
		t.Fatalf("target=%s score=%.2f, want John's handle", r.Target, r.Score)
	}
	for _, c := range r.Comparisons {
		if c.Key == vocabkey.PropertySexAtBirth && c.Result.Outcome != propertycompare.OutcomeUnknown {
			t.Fatalf("sex_at_birth outcome %s, want unknown", c.Result.Outcome)
		}
	}
}

func TestSparseNameAndBirthDate(t *testing.T) {
	// A person with only a name and a birth date files on the person whose
	// name and birth match, not on a look-alike born in another year.
	per, ev := id("s-person"), id("s-event")
	p1, e1, p2 := id("h-person-1"), id("h-event-1"), id("h-person-2")
	layer, canon := sparseAnnAmes()
	p := alignWith(layer, canon, nil, matchregistry.Default().AlignConfig())

	cfg := matchregistry.Default().AlignConfig()
	name := cfg.Subject.Weights[vocabkey.Person][vocabkey.PropertyName].Agree
	event := cfg.Subject.Weights[vocabkey.Event][vocabkey.PropertyEventType].Agree + cfg.Subject.Weights[vocabkey.Event][vocabkey.PropertyDate].Agree
	tests := []struct {
		name    string
		subject []byte
		handle  []byte
		score   float64
	}{
		{"person", per, p1, name + cfg.Segment.OneSupport + cfg.Segment.OneCredit*event},
		{"birth", ev, e1, event + cfg.Segment.OneSupport + cfg.Segment.OneCredit*name},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := rowBySubject(p, tt.subject)
			if row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, tt.handle) ||
				row.Assessment != graphalign.AssessStrong || math.Abs(row.Score-tt.score) > 1e-9 {
				t.Fatalf("row %+v, want %s strong at %.2f", row, tt.handle, tt.score)
			}
		})
	}

	// The unplaced EVT-S differs from EVT-2, so PER-2 is offered at the name's
	// 2.5 less the One link's 3.
	alts := rowBySubject(p, per).Alternatives
	if len(alts) != 1 || !bytes.Equal(alts[0].HandleID, p2) || math.Abs(alts[0].Score-(name-cfg.Segment.OneConflict)) > 1e-9 {
		t.Fatalf("PER-S alternatives %+v, want PER-2 at %.2f", alts, name-cfg.Segment.OneConflict)
	}
}

// eventVals is an event of one type in one year.
func eventVals(eventType string, year int) subjectcompare.Values {
	et, tv := termProp("event_type", eventType)
	dp, dv := dateProp(year)
	return vals(et, tv, dp, dv)
}

// withLife copies v and sets its person lifespan keys from life events.
func withLife(v subjectcompare.Values, events ...subjectprofile.LifeEvent) subjectcompare.Values {
	out := subjectcompare.Values{}
	maps.Copy(out, v)
	life := subjectprofile.ForSubject(subjectprofile.Subject{Life: events}, matchregistry.Default().Subject.Lifespans[vocabkey.Person])
	maps.Copy(out, life)
	return out
}

// lifeIn is a life event of the person as its subject, dated to one year.
func lifeIn(eventType vocabkey.EventType, year int) subjectprofile.LifeEvent {
	return subjectprofile.LifeEvent{
		EventType: eventType,
		Role:      vocabkey.RoleSubject,
		Dates:     []datevalue.Value{*datevaluetest.YMD(year, 0, 0)},
	}
}

func TestBirthIsNeverProposedOnADeath(t *testing.T) {
	metas := subjectcomparetest.Metas()
	t.Run("AC-1.2", func(t *testing.T) {
		s, h := id("s-event"), id("h-event")
		layer := graphalign.Layer{
			Subjects: []graphalign.Subject{{ID: s, Ref: "EVT-S", Kind: "event", Values: eventVals("birth", 1850)}},
			Metas:    metas,
		}
		canon := graphalign.Canon{Handles: []graphalign.Handle{{ID: h, Ref: "EVT-1", Kind: "event", Values: eventVals("death", 1850)}}}
		row := rowBySubject(alignWith(layer, canon, nil, matchregistry.Default().AlignConfig()), s)
		if row.Target != graphalign.TargetNew || row.Reason != graphalign.ReasonNoMatch || len(row.Alternatives) != 0 {
			t.Fatalf("birth against a death: %+v", row)
		}
	})
	t.Run("AC-1.3", func(t *testing.T) {
		sig := partLink("subject", "birth")
		p, e, hp, hd := id("s-person"), id("s-event"), id("h-person"), id("h-death")
		layer := graphalign.Layer{
			Subjects: []graphalign.Subject{
				{ID: p, Ref: "PER-S", Kind: "person", Values: vals(nameProp("Ann Ames"))},
				{ID: e, Ref: "EVT-S", Kind: "event", Values: eventVals("birth", 1850)},
			},
			Bridges: []graphalign.Bridge{bridge(p, e, sig)},
			Metas:   metas,
		}
		canon := graphalign.Canon{
			Handles: []graphalign.Handle{
				{ID: hp, Ref: "PER-1", Kind: "person", Values: vals(nameProp("Ann Ames"))},
				{ID: hd, Ref: "EVT-1", Kind: "event", Values: eventVals("death", 1850)},
			},
			Edges: []graphalign.CanonEdge{edge(hp, hd, sig)},
		}
		fixed := []graphalign.Fixed{{SubjectID: p, HandleID: hp}}
		row := rowBySubject(alignWith(layer, canon, fixed, matchregistry.Default().AlignConfig()), e)
		if row.Target != graphalign.TargetSkip || row.Assessment != graphalign.AssessNone || row.Reason != graphalign.ReasonWeak {
			t.Fatalf("birth under a death's signature: %+v", row)
		}
		if len(row.Alternatives) != 1 || !bytes.Equal(row.Alternatives[0].HandleID, hd) || row.Alternatives[0].Assessment != graphalign.AssessNone {
			t.Fatalf("alternatives %+v, want the death with no band", row.Alternatives)
		}
	})
	t.Run("AC-1.8", func(t *testing.T) {
		s, h := id("s-event"), id("h-event")
		et, bv := termProp("event_type", "birth")
		layer := graphalign.Layer{
			Subjects: []graphalign.Subject{{ID: s, Ref: "EVT-S", Kind: "event", Values: vals(et, bv)}},
			Metas:    metas,
		}
		canon := graphalign.Canon{Handles: []graphalign.Handle{{ID: h, Ref: "EVT-1", Kind: "event", Values: vals(et, bv)}}}
		row := rowBySubject(alignWith(layer, canon, nil, matchregistry.Default().AlignConfig()), s)
		if row.Target != graphalign.TargetSkip || row.Reason != graphalign.ReasonWeak || row.Assessment != graphalign.AssessNone ||
			len(row.Alternatives) == 0 || !bytes.Equal(row.Alternatives[0].HandleID, h) || row.Alternatives[0].Assessment != graphalign.AssessNone {
			t.Fatalf("undated birth against an undated birth: %+v", row)
		}
	})
}

// nameProp is a Western name property pair for a form.
func nameProp(form string) (vocabkey.PropertyKey, propertycompare.Value) {
	return vocabkey.PropertyName, propertycomparetest.Name(form)
}

func TestLifespanRulesOutAPerson(t *testing.T) {
	metas := subjectcomparetest.Metas()
	probe := func(deathYear int) subjectcompare.Values {
		return withLife(vals(nameProp("James Robins")), lifeIn(vocabkey.EventDeath, deathYear))
	}
	handle := withLife(vals(nameProp("James Robins")), lifeIn(vocabkey.EventBirth, 1850))
	align := func(values subjectcompare.Values) graphalign.Row {
		s, h := id("s-person"), id("h-person")
		layer := graphalign.Layer{Subjects: []graphalign.Subject{{ID: s, Ref: "PER-S", Kind: "person", Values: values}}, Metas: metas}
		canon := graphalign.Canon{Handles: []graphalign.Handle{{ID: h, Ref: "PER-1", Kind: "person", Values: handle}}}
		return rowBySubject(alignWith(layer, canon, nil, matchregistry.Default().AlignConfig()), s)
	}
	t.Run("AC-1.4", func(t *testing.T) {
		row := align(probe(1840))
		if row.Target != graphalign.TargetNew || row.Reason != graphalign.ReasonNoMatch {
			t.Fatalf("died before the handle was born: %+v", row)
		}
	})
	t.Run("AC-1.5", func(t *testing.T) {
		row := align(probe(1920))
		if row.Target != graphalign.TargetHandle || row.Assessment != graphalign.AssessWeak || math.Abs(row.Score-2.6) > 1e-9 {
			t.Fatalf("a consistent lifespan: %+v", row)
		}
		var found int
		for _, c := range row.Comparisons {
			if c.Key != vocabkey.PropertyLifespan {
				continue
			}
			found++
			if c.Result.Outcome != propertycompare.OutcomeAgree || math.Abs(c.Points-0.1) > 1e-9 || !c.Derived {
				t.Fatalf("lifespan comparison %+v", c)
			}
		}
		if found != 1 {
			t.Fatalf("%d lifespan comparisons, want 1", found)
		}
		if row.ReasonProperty != vocabkey.PropertyName {
			t.Fatalf("reason property %q, want name: the lifespan is never named", row.ReasonProperty)
		}
	})
}

func TestGenderPoolsWithSex(t *testing.T) {
	s, h := id("s-person"), id("h-person")
	name, nv := nameProp("Mary Robins")
	gp, gv := termProp("gender", "man")
	sp, sv := termProp("sex_at_birth", "male")
	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{{ID: s, Ref: "PER-S", Kind: "person", Values: vals(name, nv, gp, gv)}},
		Metas:    subjectcomparetest.Metas(),
	}
	canon := graphalign.Canon{Handles: []graphalign.Handle{{ID: h, Ref: "PER-1", Kind: "person", Values: vals(name, nv, sp, sv)}}}
	row := rowBySubject(alignWith(layer, canon, nil, matchregistry.Default().AlignConfig()), s)
	if row.Target != graphalign.TargetHandle || row.Assessment != graphalign.AssessMedium || math.Abs(row.Score-3.1) > 1e-9 {
		t.Fatalf("gender against sex: %+v", row)
	}
	var keys []vocabkey.PropertyKey
	for _, c := range row.Comparisons {
		keys = append(keys, c.Key)
		if c.Key == vocabkey.PropertySexAtBirth && c.Result.Outcome != propertycompare.OutcomeAgree {
			t.Fatalf("sex_at_birth outcome %s, want agree", c.Result.Outcome)
		}
	}
	if want := []vocabkey.PropertyKey{vocabkey.PropertyLifespan, vocabkey.PropertyName, vocabkey.PropertySexAtBirth}; !slices.Equal(keys, want) {
		t.Fatalf("comparison keys %v, want %v", keys, want)
	}
}

func TestEventPeriodStandsForTheDate(t *testing.T) {
	s, h := id("s-event"), id("h-event")
	et, rv := termProp("event_type", "residence")
	sd, sdv := vocabkey.PropertyKey(vocabkey.PropertyStartDate), propertycomparetest.Date(datevaluetest.YMD(1850, 0, 0))
	ed, edv := vocabkey.PropertyKey(vocabkey.PropertyEndDate), propertycomparetest.Date(datevaluetest.YMD(1860, 0, 0))
	dp, dv := dateProp(1855)
	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{{ID: s, Ref: "EVT-S", Kind: "event", Values: vals(et, rv, sd, sdv, ed, edv)}},
		Metas:    subjectcomparetest.Metas(),
	}
	canon := graphalign.Canon{Handles: []graphalign.Handle{{ID: h, Ref: "EVT-1", Kind: "event", Values: vals(et, rv, dp, dv)}}}
	fixed := []graphalign.Fixed{{SubjectID: s, HandleID: h}}
	row := rowBySubject(alignWith(layer, canon, fixed, matchregistry.Default().AlignConfig()), s)
	var keys []vocabkey.PropertyKey
	for _, c := range row.Comparisons {
		keys = append(keys, c.Key)
		if c.Key == vocabkey.PropertyDate && c.Result.Outcome != propertycompare.OutcomePartial {
			t.Fatalf("date outcome %s, want partial", c.Result.Outcome)
		}
	}
	if want := []vocabkey.PropertyKey{vocabkey.PropertyDate, vocabkey.PropertyEventType}; !slices.Equal(keys, want) {
		t.Fatalf("comparison keys %v, want %v", keys, want)
	}
}

// placedApart places Ann on one Ann and her event on the other Ann's event,
// and returns both rows' scores.
func placedApart(t *testing.T, eventType string) (person, event float64) {
	t.Helper()
	link := partLink("subject", vocabkey.EventType(eventType))
	et, etv := termProp("event_type", eventType)
	dp, dv := dateProp(1817)
	p, e := id("s-ann"), id("s-event")
	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: p, Ref: "PER-S", Kind: "person", Values: nameVals("Ann", "Ames")},
			{ID: e, Ref: "EVT-S", Kind: "event", Values: vals(et, etv, dp, dv)},
		},
		Bridges: []graphalign.Bridge{bridge(p, e, link)},
		Metas:   mergeMetas(nameMetas(), eventMetas()),
	}
	canon := graphalign.Canon{
		Handles: []graphalign.Handle{
			{ID: id("h-ann-1"), Ref: "PER-1", Kind: "person", Values: nameVals("Ann", "Ames")},
			{ID: id("h-ann-2"), Ref: "PER-2", Kind: "person", Values: nameVals("Ann", "Ames")},
			{ID: id("h-evt-1"), Ref: "EVT-1", Kind: "event", Values: vals(et, etv, dp, dv)},
			{ID: id("h-evt-2"), Ref: "EVT-2", Kind: "event", Values: vals(et, etv, dp, dv)},
		},
		Edges: []graphalign.CanonEdge{
			edge(id("h-ann-1"), id("h-evt-1"), link),
			edge(id("h-ann-2"), id("h-evt-2"), link),
		},
	}
	fixed := []graphalign.Fixed{
		{SubjectID: p, HandleID: id("h-ann-1")},
		{SubjectID: e, HandleID: id("h-evt-2")},
	}
	prop := alignWith(layer, canon, fixed, matchregistry.Default().AlignConfig())
	return rowBySubject(prop, p).Score, rowBySubject(prop, e).Score
}

// A person and her birth placed on different Anns' births: a birth is one per
// person and one per subject, so each end loses the one-cardinality conflict.
func TestBirthPlacedApartSubtracts(t *testing.T) {
	cfg := matchregistry.Default().AlignConfig()
	want := 2.5 - cfg.Segment.OneConflict
	person, event := placedApart(t, "birth")
	if math.Abs(person-want) > 1e-9 || math.Abs(event-want) > 1e-9 {
		t.Fatalf("person %.2f, event %.2f, want both %.2f", person, event, want)
	}
}

// A residence is not one per person: placed apart it says nothing.
func TestResidencePlacedApartIsUnknown(t *testing.T) {
	person, event := placedApart(t, "residence")
	if math.Abs(person-2.5) > 1e-9 || math.Abs(event-2.5) > 1e-9 {
		t.Fatalf("person %.2f, event %.2f, want both 2.5", person, event)
	}
}

// A spouse link never stands for a sibling link: the neighbor, however well
// its name matches, is reached by its own properties and not through Ann.
func TestSpouseNeverCorrespondsToSibling(t *testing.T) {
	ann, bob := id("s-ann"), id("s-bob")
	hAnn, hBob := id("h-ann"), id("h-bob")
	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: ann, Ref: "S-ANN", Kind: "person", Values: nameVals("Ann", "Ames")},
			{ID: bob, Ref: "S-BOB", Kind: "person", Values: nameVals("Bob", "Ames")},
		},
		Bridges: []graphalign.Bridge{bridge(ann, bob, relLink("spouse"))},
		Metas:   nameMetas(),
	}
	canon := graphalign.Canon{
		Handles: []graphalign.Handle{
			{ID: hAnn, Ref: "PER-1", Kind: "person", Values: nameVals("Ann", "Ames")},
			{ID: hBob, Ref: "PER-B", Kind: "person", Values: nameVals("Bob", "Ames")},
		},
		Edges: []graphalign.CanonEdge{edge(hAnn, hBob, relLink("sibling"))},
	}
	fixed := []graphalign.Fixed{{SubjectID: ann, HandleID: hAnn}}
	cfg := matchregistry.Default().AlignConfig()
	row := rowBySubject(alignWith(layer, canon, fixed, matchregistry.Default().AlignConfig()), bob)
	want := cfg.Subject.Weights[vocabkey.Person][vocabkey.PropertyName].Agree
	if row.Reason == graphalign.ReasonVia || math.Abs(row.Score-want) > 1e-9 {
		t.Fatalf("Bob reason %s score %.2f, want a property match at %.2f", row.Reason, row.Score, want)
	}
}

// withCandidates returns canon with Candidates set the way the loader sets
// them: graphalign.Candidates for every Subject over all fixture handles of
// its kind.
func withCandidates(layer graphalign.Layer, canon graphalign.Canon, cfg graphalign.Config) graphalign.Canon {
	canon.Candidates = map[string][][]byte{}
	for _, s := range layer.Subjects {
		for _, h := range graphalign.Candidates(s.Kind, s.Values, canon.Handles, layer.Metas, cfg) {
			canon.Candidates[string(s.ID)] = append(canon.Candidates[string(s.ID)], h.ID)
		}
	}
	return canon
}

// alignWith is Align on a canon whose Candidates the loader would have filled.
func alignWith(layer graphalign.Layer, canon graphalign.Canon, fixed []graphalign.Fixed, cfg graphalign.Config) graphalign.Proposal {
	return graphalign.Align(layer, withCandidates(layer, canon, cfg), fixed, cfg)
}

// withSex adds a sex_at_birth term to values.
func withSex(v subjectcompare.Values, term string) subjectcompare.Values {
	k, x := termProp("sex_at_birth", term)
	out := subjectcompare.Values{}
	maps.Copy(out, v)
	out[k] = append(slices.Clone(out[k]), x)
	return out
}

func sexMetas() subjectcompare.Metas {
	return mergeMetas(nameMetas(), eventMetas(), metasOf(map[vocabkey.PropertyKey]string{vocabkey.PropertySexAtBirth: propertymeta.ValueTypeTerm}))
}

func TestOnlyCandidatesOrNominatedHandlesAreCompared(t *testing.T) {
	cfg := matchregistry.Default().AlignConfig()
	t.Run("a handle the shortlist did not offer is never scored", func(t *testing.T) {
		layer := graphalign.Layer{
			Subjects: []graphalign.Subject{{ID: id("s"), Ref: "PER-S", Kind: "person", Values: nameVals("Rob", "Robnis")}},
			Metas:    nameMetas(),
		}
		canon := graphalign.Canon{
			Handles:    []graphalign.Handle{{ID: id("h"), Ref: "PER-1", Kind: "person", Values: nameVals("Rob", "Robins")}},
			Candidates: map[string][][]byte{},
		}
		row := rowBySubject(graphalign.Align(layer, canon, nil, cfg), id("s"))
		if row.Target != graphalign.TargetNew || row.Reason != graphalign.ReasonNoMatch || len(row.Alternatives) != 0 {
			t.Fatalf("want New no_match with no alternatives, got %+v", row)
		}
	})
	t.Run("a handle reached through a placed link is nominated", func(t *testing.T) {
		et, birth := termProp("event_type", "birth")
		date := vocabkey.PropertyKey(vocabkey.PropertyDate)
		y1817 := propertycomparetest.Date(datevaluetest.YMD(1817, 0, 0))
		sig := partLink("subject", "birth")
		layer := graphalign.Layer{
			Subjects: []graphalign.Subject{
				{ID: id("s"), Ref: "PER-S", Kind: "person", Values: nameVals("Rob", "Robnis")},
				{ID: id("e"), Ref: "EVT-S", Kind: "event", Values: vals(et, birth, date, y1817)},
			},
			Bridges: []graphalign.Bridge{bridge(id("s"), id("e"), sig)},
			Metas:   mergeMetas(nameMetas(), eventMetas()),
		}
		canon := graphalign.Canon{
			Handles: []graphalign.Handle{
				{ID: id("h"), Ref: "PER-1", Kind: "person", Values: nameVals("Rob", "Robins")},
				{ID: id("he"), Ref: "EVT-1", Kind: "event", Values: vals(et, birth, date, y1817)},
			},
			Edges:      []graphalign.CanonEdge{edge(id("h"), id("he"), sig)},
			Candidates: map[string][][]byte{},
		}
		fixed := []graphalign.Fixed{{SubjectID: id("e"), HandleID: id("he")}}
		row := rowBySubject(graphalign.Align(layer, canon, fixed, cfg), id("s"))
		if row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, id("h")) ||
			row.Reason != graphalign.ReasonVia || row.Via == nil || !bytes.Equal(row.Via.NeighborSubjectID, id("e")) {
			t.Fatalf("want PER-1 via the birth row, got %+v", row)
		}
	})
	t.Run("a same-name look-alike whose sex conflicts is Skip weak", func(t *testing.T) {
		layer := graphalign.Layer{
			Subjects: []graphalign.Subject{{ID: id("s"), Ref: "PER-S", Kind: "person", Values: withSex(nameVals("Ann", "Ames"), "male")}},
			Metas:    sexMetas(),
		}
		canon := graphalign.Canon{Handles: []graphalign.Handle{{ID: id("h"), Ref: "PER-1", Kind: "person", Values: withSex(nameVals("Ann", "Ames"), "female")}}}
		row := rowBySubject(alignWith(layer, canon, nil, cfg), id("s"))
		if row.Target != graphalign.TargetSkip || row.Reason != graphalign.ReasonWeak || row.Assessment != graphalign.AssessNone ||
			len(row.Alternatives) == 0 || !bytes.Equal(row.Alternatives[0].HandleID, id("h")) {
			t.Fatalf("want Skip weak with PER-1 as the alternative, got %+v", row)
		}
	})
}

func TestSettlingMovesAPairAndSaysWhy(t *testing.T) {
	spouse := relLink("spouse")
	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: id("p"), Ref: "PER-P", Kind: "person", Values: withSex(nameVals("Ann", "Ames"), "male")},
			{ID: id("s"), Ref: "PER-S", Kind: "person", Values: nameVals("John", "Ames")},
		},
		Bridges: []graphalign.Bridge{bridge(id("p"), id("s"), spouse)},
		Metas:   sexMetas(),
	}
	canon := graphalign.Canon{
		Handles: []graphalign.Handle{
			{ID: id("h1"), Ref: "PER-1", Kind: "person", Values: withSex(nameVals("Ann", "Ames"), "male")},
			{ID: id("h2"), Ref: "PER-2", Kind: "person", Values: nameVals("Ann", "Ames")},
			{ID: id("h3"), Ref: "PER-3", Kind: "person", Values: nameVals("John", "Ames")},
		},
		Edges: []graphalign.CanonEdge{edge(id("h2"), id("h3"), spouse)},
	}
	base := matchregistry.Default().AlignConfig()
	settle := base
	settle.Segment.UnplacedAlike = 100
	settle.Alignment.SettleMargin = 0.1
	n := func(row graphalign.Row, handle string, score float64) {
		t.Helper()
		if row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, id(handle)) || math.Abs(row.Score-score) > 1e-9 {
			t.Fatalf("want %s at %.3f, got %+v", handle, score, row)
		}
	}
	name := base.Subject.Weights[vocabkey.Person][vocabkey.PropertyName].Agree
	linked := name + base.Segment.OtherSupport + base.Segment.OtherCredit*name

	t.Run("a neighbor placed after the pair moves it", func(t *testing.T) {
		p := alignWith(layer, canon, nil, settle)
		pr := rowBySubject(p, id("p"))
		n(pr, "h2", linked)
		if pr.Move == nil || pr.Move.Reason != graphalign.ReasonNeighborPlaced || !bytes.Equal(pr.Move.NeighborSubjectID, id("s")) {
			t.Fatalf("want Move neighbor_placed by S, got %+v", pr.Move)
		}
		sr := rowBySubject(p, id("s"))
		n(sr, "h3", sr.Score)
		if sr.Move != nil {
			t.Fatalf("S did not move: %+v", sr.Move)
		}
	})
	t.Run("no settle rounds leaves the first placement", func(t *testing.T) {
		c := settle
		c.Alignment.SettleRounds = 0
		pr := rowBySubject(alignWith(layer, canon, nil, c), id("p"))
		n(pr, "h1", name+base.Subject.Weights[vocabkey.Person][vocabkey.PropertySexAtBirth].Agree)
		if pr.Move != nil {
			t.Fatalf("no move expected: %+v", pr.Move)
		}
	})
	t.Run("an unplaced neighbor already lines up under the shipped numbers", func(t *testing.T) {
		pr := rowBySubject(alignWith(layer, canon, nil, base), id("p"))
		n(pr, "h2", linked)
		if pr.Move != nil {
			t.Fatalf("no move expected: %+v", pr.Move)
		}
	})
}

func TestContestedRow(t *testing.T) {
	cfg := matchregistry.Default().AlignConfig()
	et, birth := termProp("event_type", "birth")
	date := vocabkey.PropertyKey(vocabkey.PropertyDate)
	y1817 := propertycomparetest.Date(datevaluetest.YMD(1817, 0, 0))
	sig := partLink("subject", "birth")
	ev := vals(et, birth, date, y1817)
	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: id("p"), Ref: "PER-S", Kind: "person", Values: withSex(nameVals("Ann", "Ames"), "male")},
			{ID: id("e"), Ref: "EVT-S", Kind: "event", Values: ev},
			{ID: id("n"), Ref: "PER-N", Kind: "person", Values: nameVals("Zed", "Zimm")},
			{ID: id("k"), Ref: "PER-K", Kind: "person", Values: nameVals("Kay", "Kell")},
		},
		Bridges: []graphalign.Bridge{bridge(id("p"), id("e"), sig)},
		Metas:   sexMetas(),
	}
	canon := graphalign.Canon{
		Handles: []graphalign.Handle{
			{ID: id("h"), Ref: "PER-1", Kind: "person", Values: withSex(nameVals("Ann", "Ames"), "female")},
			{ID: id("he"), Ref: "EVT-1", Kind: "event", Values: ev},
		},
		Edges: []graphalign.CanonEdge{edge(id("h"), id("he"), sig)},
	}
	fixed := []graphalign.Fixed{{SubjectID: id("k"), Target: graphalign.TargetSkip}}
	p := alignWith(layer, canon, fixed, cfg)
	pr, er := rowBySubject(p, id("p")), rowBySubject(p, id("e"))
	want := func(row graphalign.Row, handle string, score float64, contested bool) {
		t.Helper()
		if !bytes.Equal(row.HandleID, id(handle)) || row.Assessment != graphalign.AssessStrong ||
			math.Abs(row.Score-score) > 1e-9 || row.Flags.Contested != contested {
			t.Fatalf("got %+v, want %s strong at %.2f contested %v", row, handle, score, contested)
		}
	}
	name := cfg.Subject.Weights[vocabkey.Person][vocabkey.PropertyName].Agree
	sexw := cfg.Subject.Weights[vocabkey.Person][vocabkey.PropertySexAtBirth]
	event := cfg.Subject.Weights[vocabkey.Event][vocabkey.PropertyEventType].Agree + cfg.Subject.Weights[vocabkey.Event][vocabkey.PropertyDate].Agree
	want(pr, "h", name-sexw.Conflict+cfg.Segment.OneSupport+cfg.Segment.OneCredit*event, true)
	want(er, "he", event+cfg.Segment.OneSupport+cfg.Segment.OneCredit*(name-sexw.Conflict), false)
	if r := rowBySubject(p, id("n")); r.Target != graphalign.TargetNew || r.Flags.Contested {
		t.Fatalf("a New row is not contested: %+v", r)
	}
	if r := rowBySubject(p, id("k")); r.Target != graphalign.TargetSkip || r.Flags.Contested {
		t.Fatalf("a decided Skip row is not contested: %+v", r)
	}
}
