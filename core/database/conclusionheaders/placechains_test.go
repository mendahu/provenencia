package conclusionheaders_test

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/conclusiondetails"
	"github.com/mendahu/provenencia/core/database/conclusionheaders"
	"github.com/mendahu/provenencia/core/database/connect"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjects/subjectstest"

	// placeSubject cites a Place (and optional period) but does not promote — call
	// promoteSubject after placeRel so bridges file with the ends.
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/writes"
)

func year(y int) *int { return &y }

func (f *fixture) placeSubject(name string, start, end *int, x, y int64) subjects.Subject {
	f.t.Helper()
	s := subjectstest.On(f.t, f.c, userID, f.graph, "place", "")
	f.at(s, x, y)
	in := []observations.Input{
		{PropertyID: f.prop("toponym").ID, ValueText: name, HasText: true},
	}
	if start != nil {
		in = append(in, observations.Input{PropertyID: f.prop("start_date").ID, Date: pointYear(*start)})
	}
	if end != nil {
		in = append(in, observations.Input{PropertyID: f.prop("end_date").ID, Date: pointYear(*end)})
	}
	f.cite(s, in...)
	return s
}

func (f *fixture) placeRel(from, to subjects.Subject, kind string, linkStart, linkEnd *int) {
	f.t.Helper()
	term := f.term("place_relationship_type", kind)
	fromProp := f.prop("from").ID
	toProp := f.prop("to").ID
	typeProp := f.prop("place_relationship_type").ID
	res, err := writes.Call(f.c, writes.Op{Action: "create_cited_bridge", UserID: userID}, func(tx *database.Tx) (connect.Result, []rowchange.Change, error) {
		return connect.CreateCitedBridge(tx, userID, connect.CreateInput{
			FromSubjectID: from.ID, ToSubjectID: to.ID, BridgeTypeKey: "place_relationship",
			Citation: citations.CreateInput{ArtifactID: f.artifact.ID, LocatorJSON: locator},
			Observations: []observations.Input{
				{PropertyID: fromProp, ValueSubjectID: from.ID},
				{PropertyID: toProp, ValueSubjectID: to.ID},
				{PropertyID: typeProp, ValueTermID: term.ID},
			},
		})
	})
	must(f.t, err)
	if linkStart != nil || linkEnd != nil {
		var extra []observations.Input
		if linkStart != nil {
			extra = append(extra, observations.Input{
				SubjectID: res.Subject.ID, PropertyID: f.prop("start_date").ID, Date: pointYear(*linkStart),
			})
		}
		if linkEnd != nil {
			extra = append(extra, observations.Input{
				SubjectID: res.Subject.ID, PropertyID: f.prop("end_date").ID, Date: pointYear(*linkEnd),
			})
		}
		_, err := writes.Call(f.c, writes.Op{Action: "add_observations", UserID: userID}, func(tx *database.Tx) ([]observations.Observation, []rowchange.Change, error) {
			return observations.AddToCitation(tx, userID, res.Citation.ID, extra)
		})
		must(f.t, err)
	}
}

func TestPlaceParentsAtDateFromPeriods(t *testing.T) {
	f := newFixture(t)
	// Toronto itself is undated (always) so undated links follow the parents'
	// periods alone — 1820 / 1850 / 1950 pick UC / PC / Ontario.
	toronto := f.placeSubject("Toronto", nil, nil, 0, 0)
	uc := f.placeSubject("Upper Canada", year(1791), year(1841), 2, 0)
	pc := f.placeSubject("Province of Canada", year(1841), year(1867), 4, 0)
	ontario := f.placeSubject("Ontario", year(1867), nil, 6, 0)
	f.placeRel(toronto, uc, propertyterms.KeyPartOf, nil, nil)
	f.placeRel(toronto, pc, propertyterms.KeyPartOf, nil, nil)
	f.placeRel(toronto, ontario, propertyterms.KeyPartOf, nil, nil)
	torontoID := f.promoteSubject(toronto)
	f.promoteSubject(uc)
	f.promoteSubject(pc)
	f.promoteSubject(ontario)

	parents := func(y int) []string {
		t.Helper()
		got, err := conclusionheaders.ParentsAtDate(f.c, torontoID, *pointYear(y))
		must(t, err)
		return got
	}
	if got := parents(1820); !sameSet(got, []string{"Upper Canada"}) {
		t.Fatalf("1820 parents %v", got)
	}
	if got := parents(1850); !sameSet(got, []string{"Province of Canada"}) {
		t.Fatalf("1850 parents %v", got)
	}
	if got := parents(1950); !sameSet(got, []string{"Ontario"}) {
		t.Fatalf("1950 parents %v", got)
	}
	// About 1841 → both UC and PC candidates.
	abt := *pointYear(1841)
	abt.Qualifier = datevalue.QualifierABT
	got, err := conclusionheaders.ParentsAtDate(f.c, torontoID, abt)
	must(t, err)
	if !sameSet(got, []string{"Upper Canada", "Province of Canada"}) {
		t.Fatalf("about 1841 parents %v", got)
	}
}

func TestPlaceSeveralParentsAreAllCandidates(t *testing.T) {
	f := newFixture(t)
	farm := f.placeSubject("The Farm", nil, nil, 0, 0)
	township := f.placeSubject("York Township", nil, nil, 2, 0)
	county := f.placeSubject("York County", nil, nil, 4, 0)
	f.placeRel(farm, township, propertyterms.KeyPartOf, nil, nil)
	f.placeRel(farm, county, propertyterms.KeyPartOf, nil, nil)
	farmID := f.promoteSubject(farm)
	f.promoteSubject(township)
	f.promoteSubject(county)
	got, err := conclusionheaders.ParentsAtDate(f.c, farmID, conclusionheaders.TodayDate())
	must(t, err)
	if !sameSet(got, []string{"York Township", "York County"}) {
		t.Fatalf("parents %v", got)
	}
}

func TestIrelandUKDatedMembershipDropsAfter1922(t *testing.T) {
	f := newFixture(t)
	ireland := f.placeSubject("Ireland", nil, nil, 0, 0)
	uk := f.placeSubject("United Kingdom", nil, nil, 2, 0)
	f.placeRel(ireland, uk, propertyterms.KeyPartOf, nil, year(1922))
	irelandID := f.promoteSubject(ireland)
	ukID := f.promoteSubject(uk)
	before, err := conclusionheaders.ParentsAtDate(f.c, irelandID, *pointYear(1900))
	must(t, err)
	if !sameSet(before, []string{"United Kingdom"}) {
		t.Fatalf("1900 %v", before)
	}
	after, err := conclusionheaders.ParentsAtDate(f.c, irelandID, *pointYear(1923))
	must(t, err)
	if len(after) != 0 {
		t.Fatalf("1923 still has parents %v", after)
	}
	// Both places continue: UK still lists Ireland as a former part only via
	// the dated link — PartsAtDate at 1923 is empty.
	parts, err := conclusionheaders.PartsAtDate(f.c, ukID, *pointYear(1923))
	must(t, err)
	if len(parts) != 0 {
		t.Fatalf("UK parts in 1923 %v", parts)
	}
	parts1900, err := conclusionheaders.PartsAtDate(f.c, ukID, *pointYear(1900))
	must(t, err)
	if !sameSet(parts1900, []string{"Ireland"}) {
		t.Fatalf("UK parts in 1900 %v", parts1900)
	}
}

func TestSuccessionNeverBuildsChain(t *testing.T) {
	f := newFixture(t)
	york := f.placeSubject("York", year(1793), year(1834), 0, 0)
	toronto := f.placeSubject("Toronto", year(1834), nil, 2, 0)
	f.placeRel(york, toronto, propertyterms.KeySucceededBy, nil, nil)
	yorkID := f.promoteSubject(york)
	torontoID := f.promoteSubject(toronto)
	parents, err := conclusionheaders.ParentsAtDate(f.c, yorkID, *pointYear(1834))
	must(t, err)
	if len(parents) != 0 {
		t.Fatalf("succession leaked into chain %v", parents)
	}
	pred, succ, err := conclusionheaders.SuccessionNames(f.c, yorkID)
	must(t, err)
	if len(pred) != 0 || !sameSet(succ, []string{"Toronto"}) {
		t.Fatalf("york succession pred=%v succ=%v", pred, succ)
	}
	pred, succ, err = conclusionheaders.SuccessionNames(f.c, torontoID)
	must(t, err)
	if !sameSet(pred, []string{"York"}) || len(succ) != 0 {
		t.Fatalf("toronto succession pred=%v succ=%v", pred, succ)
	}
}

func TestSuccessionSplitAndAmalgamation(t *testing.T) {
	f := newFixture(t)
	oldA := f.placeSubject("Old A", nil, nil, 0, 0)
	oldB := f.placeSubject("Old B", nil, nil, 2, 0)
	newOne := f.placeSubject("New", nil, nil, 4, 0)
	left := f.placeSubject("Left", nil, nil, 0, 2)
	right := f.placeSubject("Right", nil, nil, 2, 2)
	f.placeRel(oldA, newOne, propertyterms.KeySucceededBy, nil, nil)
	f.placeRel(oldB, newOne, propertyterms.KeySucceededBy, nil, nil)
	f.placeRel(newOne, left, propertyterms.KeySucceededBy, nil, nil)
	f.placeRel(newOne, right, propertyterms.KeySucceededBy, nil, nil)
	f.promoteSubject(oldA)
	f.promoteSubject(oldB)
	newID := f.promoteSubject(newOne)
	f.promoteSubject(left)
	f.promoteSubject(right)
	pred, succ, err := conclusionheaders.SuccessionNames(f.c, newID)
	must(t, err)
	if !sameSet(pred, []string{"Old A", "Old B"}) || !sameSet(succ, []string{"Left", "Right"}) {
		t.Fatalf("amalgamation/split pred=%v succ=%v", pred, succ)
	}
}

func TestBirthLocationsFoldWhenPartOf(t *testing.T) {
	f := newFixture(t)
	toronto := f.placeSubject("Toronto", nil, nil, 0, 0)
	ontario := f.placeSubject("Ontario", nil, nil, 2, 0)
	scotland := f.placeSubject("Scotland", nil, nil, 4, 0)
	f.placeRel(toronto, ontario, propertyterms.KeyPartOf, nil, nil)

	person := subjectstest.On(f.t, f.c, userID, f.graph, "person", "")
	f.named(person, "James")
	f.at(person, 0, 4)
	birth := subjectstest.On(f.t, f.c, userID, f.graph, "event", "")
	f.at(birth, 4, 4)
	f.cite(birth,
		observations.Input{PropertyID: f.prop("event_type").ID, ValueTermID: f.term("event_type", "birth").ID},
		observations.Input{PropertyID: f.prop("date").ID, Date: pointYear(1850)},
	)
	f.participation(person, birth, "subject")
	f.location(birth, toronto)
	f.location(birth, ontario)
	f.promoteSubject(person)
	f.promoteSubject(birth)
	f.promoteSubject(toronto)
	f.promoteSubject(ontario)

	persons, err := conclusionheaders.ListPersons(f.c)
	must(t, err)
	if len(persons) != 1 {
		t.Fatalf("persons %d", len(persons))
	}
	places := persons[0].Birth.Places
	if len(places) != 1 {
		t.Fatalf("folded places %d (%+v)", len(places), places)
	}
	if !sameSet(places[0].Names, []string{"Toronto"}) {
		t.Fatalf("leaf names %v", places[0].Names)
	}
	if !sameSet(places[0].Parents, []string{"Ontario"}) {
		t.Fatalf("chain parents %v", places[0].Parents)
	}

	// Unrelated Scotland stays a competing value.
	birth2 := subjectstest.On(f.t, f.c, userID, f.graph, "event", "")
	f.at(birth2, 4, 6)
	f.cite(birth2,
		observations.Input{PropertyID: f.prop("event_type").ID, ValueTermID: f.term("event_type", "birth").ID},
		observations.Input{PropertyID: f.prop("date").ID, Date: pointYear(1850)},
	)
	person2 := subjectstest.On(f.t, f.c, userID, f.graph, "person", "")
	f.named(person2, "Mary")
	f.at(person2, 0, 6)
	f.participation(person2, birth2, "subject")
	f.location(birth2, toronto)
	f.location(birth2, scotland)
	f.promoteSubject(person2)
	f.promoteSubject(birth2)
	f.promoteSubject(scotland)
	persons, err = conclusionheaders.ListPersons(f.c)
	must(t, err)
	var mary *conclusionheaders.PersonHeader
	for i := range persons {
		if persons[i].Name != nil && persons[i].Name.Form == "Mary" {
			mary = &persons[i]
			break
		}
	}
	if mary == nil || len(mary.Birth.Places) != 2 {
		t.Fatalf("competing places %+v", mary)
	}
}

// Two places under one parent are two readings, even when a third record
// names only that parent: Toronto and Hamilton stay competing beside Ontario.
func TestSiblingBirthPlacesStayCompetingUnderASharedParent(t *testing.T) {
	f := newFixture(t)
	toronto := f.placeSubject("Toronto", nil, nil, 0, 0)
	hamilton := f.placeSubject("Hamilton", nil, nil, 2, 0)
	ontario := f.placeSubject("Ontario", nil, nil, 4, 0)
	f.placeRel(toronto, ontario, propertyterms.KeyPartOf, nil, nil)
	f.placeRel(hamilton, ontario, propertyterms.KeyPartOf, nil, nil)

	person := subjectstest.On(f.t, f.c, userID, f.graph, "person", "")
	f.named(person, "James")
	f.at(person, 0, 4)
	birth := subjectstest.On(f.t, f.c, userID, f.graph, "event", "")
	f.at(birth, 4, 4)
	f.cite(birth,
		observations.Input{PropertyID: f.prop("event_type").ID, ValueTermID: f.term("event_type", "birth").ID},
		observations.Input{PropertyID: f.prop("date").ID, Date: pointYear(1850)},
	)
	f.participation(person, birth, "subject")
	f.location(birth, toronto)
	f.location(birth, hamilton)
	f.location(birth, ontario)
	f.promoteSubject(person)
	f.promoteSubject(birth)
	f.promoteSubject(toronto)
	f.promoteSubject(hamilton)
	f.promoteSubject(ontario)

	persons, err := conclusionheaders.ListPersons(f.c)
	must(t, err)
	if len(persons) != 1 {
		t.Fatalf("persons %d", len(persons))
	}
	var leaves []string
	for _, p := range persons[0].Birth.Places {
		leaves = append(leaves, p.Names...)
	}
	if !sameSet(leaves, []string{"Toronto", "Hamilton"}) {
		t.Fatalf("birth places %v, want Toronto and Hamilton competing (Ontario folded into each)", leaves)
	}
}

func TestPlaceHeaderParentsAndPeriod(t *testing.T) {
	f := newFixture(t)
	toronto := f.placeSubject("Toronto", year(1834), nil, 0, 0)
	ontario := f.placeSubject("Ontario", year(1867), nil, 2, 0)
	f.placeRel(toronto, ontario, propertyterms.KeyPartOf, nil, nil)
	f.promoteSubject(toronto)
	f.promoteSubject(ontario)
	got := f.places()
	var h *conclusionheaders.PlaceHeader
	for i := range got {
		if sameSet(got[i].Names, []string{"Toronto"}) {
			h = &got[i]
			break
		}
	}
	if h == nil || h.StartDate == nil || *h.StartDate.StartYear != 1834 {
		t.Fatalf("toronto period %+v", h)
	}
	if !sameSet(h.Parents, []string{"Ontario"}) {
		t.Fatalf("toronto parents %v", h.Parents)
	}
}

func TestAttachPlaceRelationships(t *testing.T) {
	f := newFixture(t)
	toronto := f.placeSubject("Toronto", year(1834), nil, 0, 0)
	uc := f.placeSubject("Upper Canada", year(1791), year(1841), 2, 0)
	ontario := f.placeSubject("Ontario", year(1867), nil, 4, 0)
	york := f.placeSubject("York", year(1793), year(1834), 6, 0)
	f.placeRel(toronto, uc, propertyterms.KeyPartOf, nil, nil)
	f.placeRel(toronto, ontario, propertyterms.KeyPartOf, nil, nil)
	f.placeRel(york, toronto, propertyterms.KeySucceededBy, nil, nil)
	torontoID := f.promoteSubject(toronto)
	f.promoteSubject(uc)
	f.promoteSubject(ontario)
	yorkID := f.promoteSubject(york)
	headers, err := conclusionheaders.PlacesByIDs(f.c, [][]byte{torontoID})
	must(t, err)
	if len(headers) != 1 {
		t.Fatalf("%d", len(headers))
	}
	h := headers[0]
	must(t, conclusionheaders.AttachPlaceRelationships(f.c, &h))
	if len(h.PartOf) != 2 {
		t.Fatalf("part_of %+v", h.PartOf)
	}
	titles := map[string]bool{}
	for _, r := range h.PartOf {
		titles[r.Title] = true
		if r.Kind != conclusionheaders.RelPartOf {
			t.Fatalf("kind %s", r.Kind)
		}
	}
	if !titles["Upper Canada"] || !titles["Ontario"] {
		t.Fatalf("part_of titles %v", titles)
	}
	if len(h.Predecessors) != 1 || h.Predecessors[0].Title != "York" {
		t.Fatalf("predecessors %+v", h.Predecessors)
	}
	if len(h.Successors) != 0 {
		t.Fatalf("successors %+v", h.Successors)
	}
	// York → Toronto succession both ways; successor shows Toronto's period.
	yh, err := conclusionheaders.PlacesByIDs(f.c, [][]byte{yorkID})
	must(t, err)
	must(t, conclusionheaders.AttachPlaceRelationships(f.c, &yh[0]))
	if len(yh[0].Successors) != 1 || yh[0].Successors[0].Title != "Toronto" {
		t.Fatalf("york successors %+v", yh[0].Successors)
	}
	if yh[0].Successors[0].StartDate == nil || yh[0].Successors[0].StartDate.StartYear == nil ||
		*yh[0].Successors[0].StartDate.StartYear != 1834 {
		t.Fatalf("york successor period %+v", yh[0].Successors[0])
	}
}

func TestAttachPlaceRelationshipsIrelandMembershipSpan(t *testing.T) {
	f := newFixture(t)
	ireland := f.placeSubject("Ireland", nil, nil, 0, 0)
	uk := f.placeSubject("United Kingdom", nil, nil, 2, 0)
	f.placeRel(ireland, uk, propertyterms.KeyPartOf, nil, year(1922))
	irelandID := f.promoteSubject(ireland)
	ukID := f.promoteSubject(uk)
	ih, err := conclusionheaders.PlacesByIDs(f.c, [][]byte{irelandID})
	must(t, err)
	must(t, conclusionheaders.AttachPlaceRelationships(f.c, &ih[0]))
	if len(ih[0].PartOf) != 1 || ih[0].PartOf[0].Title != "United Kingdom" {
		t.Fatalf("ireland part_of %+v", ih[0].PartOf)
	}
	if ih[0].PartOf[0].EndDate == nil || ih[0].PartOf[0].EndDate.StartYear == nil ||
		*ih[0].PartOf[0].EndDate.StartYear != 1922 {
		t.Fatalf("ireland membership end %+v", ih[0].PartOf[0])
	}
	uh, err := conclusionheaders.PlacesByIDs(f.c, [][]byte{ukID})
	must(t, err)
	must(t, conclusionheaders.AttachPlaceRelationships(f.c, &uh[0]))
	// Contains is "today"; a membership that ended in 1922 is not listed.
	for _, c := range uh[0].Contains {
		if c.Title == "Ireland" {
			t.Fatalf("UK contains former Ireland %+v", uh[0].Contains)
		}
	}
}

func TestHeaderDependentsIncludeChildPlaces(t *testing.T) {
	f := newFixture(t)
	toronto := f.placeSubject("Toronto", nil, nil, 0, 0)
	ontario := f.placeSubject("Ontario", nil, nil, 2, 0)
	f.placeRel(toronto, ontario, propertyterms.KeyPartOf, nil, nil)
	torontoID := f.promoteSubject(toronto)
	ontarioID := f.promoteSubject(ontario)
	db, err := f.c.DB()
	must(t, err)
	deps, err := conclusionheaders.HeaderDependents(db, [][]byte{ontarioID})
	must(t, err)
	if !database.ContainsBlobID(deps, torontoID) {
		t.Fatalf("parent dependents missing child: %d", len(deps))
	}
}

func TestUndatedPlacesAlwaysHold(t *testing.T) {
	f := newFixture(t)
	child := f.placeSubject("Child", nil, nil, 0, 0)
	parent := f.placeSubject("Parent", nil, nil, 2, 0)
	f.placeRel(child, parent, propertyterms.KeyPartOf, nil, nil)
	childID := f.promoteSubject(child)
	f.promoteSubject(parent)
	for _, y := range []int{1, 1000, 2020} {
		got, err := conclusionheaders.ParentsAtDate(f.c, childID, *pointYear(y))
		must(t, err)
		if !sameSet(got, []string{"Parent"}) {
			t.Fatalf("%d parents %v", y, got)
		}
	}
}

// lifeEvent cites a birth or death event for person, dated by the Property
// dateKey ("date" or "start_date"; "" for no date) at year, located at places.
// It does not promote.
func (f *fixture) lifeEvent(person subjects.Subject, kind, dateKey string, y int, y2 int64, places ...subjects.Subject) subjects.Subject {
	f.t.Helper()
	ev := subjectstest.On(f.t, f.c, userID, f.graph, "event", "")
	f.at(ev, 6, y2)
	in := []observations.Input{{PropertyID: f.prop("event_type").ID, ValueTermID: f.term("event_type", kind).ID}}
	if dateKey != "" {
		in = append(in, observations.Input{PropertyID: f.prop(dateKey).ID, Date: pointYear(y)})
	}
	f.cite(ev, in...)
	f.participation(person, ev, "subject")
	for _, p := range places {
		f.location(ev, p)
	}
	return ev
}

func (f *fixture) lifePerson() subjects.Subject {
	f.t.Helper()
	p := subjectstest.On(f.t, f.c, userID, f.graph, "person", "")
	f.named(p, "Ann")
	f.at(p, 0, 20)
	return p
}

func TestPersonLifeOwners(t *testing.T) {
	type owner struct {
		entity []byte
		prop   string
	}
	type world struct {
		person       []byte
		birth, death struct{ dates, places []owner }
	}
	cases := []struct {
		name  string
		build func(f *fixture) world
	}{
		{"one birth event with a date at York", func(f *fixture) world {
			york := f.placeSubject("York", nil, nil, 0, 0)
			p := f.lifePerson()
			ev := f.lifeEvent(p, "birth", "date", 1817, 2, york)
			var w world
			w.person = f.promoteSubject(p)
			e, y := f.promoteSubject(ev), f.promoteSubject(york)
			w.birth.dates = []owner{{e, "date"}}
			w.birth.places = []owner{{y, "toponym"}}
			return w
		}},
		{"start_date only", func(f *fixture) world {
			p := f.lifePerson()
			ev := f.lifeEvent(p, "birth", "start_date", 1817, 2)
			var w world
			w.person = f.promoteSubject(p)
			w.birth.dates = []owner{{f.promoteSubject(ev), "start_date"}}
			return w
		}},
		{"earlier birth event first", func(f *fixture) world {
			p := f.lifePerson()
			late := f.lifeEvent(p, "birth", "date", 1818, 2)
			early := f.lifeEvent(p, "birth", "date", 1817, 4)
			var w world
			w.person = f.promoteSubject(p)
			l, e := f.promoteSubject(late), f.promoteSubject(early)
			w.birth.dates = []owner{{e, "date"}, {l, "date"}}
			return w
		}},
		{"death at two Places that do not fold", func(f *fixture) world {
			mtl := f.placeSubject("Montréal", nil, nil, 0, 0)
			paris := f.placeSubject("Paris", nil, nil, 0, 2)
			p := f.lifePerson()
			ev := f.lifeEvent(p, "death", "date", 1900, 4, mtl, paris)
			var w world
			w.person = f.promoteSubject(p)
			w.death.dates = []owner{{f.promoteSubject(ev), "date"}}
			// The order the header gives the Places: compared below against
			// the header's own walk.
			f.promoteSubject(mtl)
			f.promoteSubject(paris)
			return w
		}},
		{"York on two birth events is listed once", func(f *fixture) world {
			york := f.placeSubject("York", nil, nil, 0, 0)
			p := f.lifePerson()
			a := f.lifeEvent(p, "birth", "date", 1817, 2, york)
			b := f.lifeEvent(p, "birth", "date", 1818, 4, york)
			var w world
			w.person = f.promoteSubject(p)
			ae, be := f.promoteSubject(a), f.promoteSubject(b)
			y := f.promoteSubject(york)
			w.birth.dates = []owner{{ae, "date"}, {be, "date"}}
			w.birth.places = []owner{{y, "toponym"}}
			return w
		}},
		{"a parent Place is not an owner", func(f *fixture) world {
			york := f.placeSubject("York", nil, nil, 0, 0)
			shire := f.placeSubject("Yorkshire", nil, nil, 2, 0)
			f.placeRel(york, shire, propertyterms.KeyPartOf, nil, nil)
			p := f.lifePerson()
			ev := f.lifeEvent(p, "birth", "date", 1817, 2, york, shire)
			var w world
			w.person = f.promoteSubject(p)
			e, y := f.promoteSubject(ev), f.promoteSubject(york)
			f.promoteSubject(shire)
			w.birth.dates = []owner{{e, "date"}}
			w.birth.places = []owner{{y, "toponym"}}
			return w
		}},
		{"no events", func(f *fixture) world {
			p := f.lifePerson()
			return world{person: f.promoteSubject(p)}
		}},
		{"unknown id", func(f *fixture) world {
			return world{person: bytes16(0x7a)}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			w := tc.build(f)
			birth, death, err := conclusionheaders.PersonLifeOwners(f.c, w.person)
			must(t, err)
			check := func(side string, got conclusionheaders.LifeOwners, want struct{ dates, places []owner }) {
				t.Helper()
				for _, part := range []struct {
					name string
					got  []conclusionheaders.FieldOwner
					want []owner
				}{{"dates", got.Dates, want.dates}, {"places", got.Places, want.places}} {
					if len(part.got) != len(part.want) {
						// A case that leaves the order to the header's walk.
						if part.want == nil && side == "death" && part.name == "places" && len(part.got) == 2 {
							continue
						}
						t.Fatalf("%s %s: %d owners, want %d", side, part.name, len(part.got), len(part.want))
					}
					for i := range part.got {
						if !bytes.Equal(part.got[i].Entity.ID, part.want[i].entity) {
							t.Fatalf("%s %s[%d]: wrong entity", side, part.name, i)
						}
						if wantID := f.prop(part.want[i].prop).ID; !bytes.Equal(part.got[i].PropertyID, wantID) {
							t.Fatalf("%s %s[%d]: property %x, want %s", side, part.name, i, part.got[i].PropertyID, part.want[i].prop)
						}
					}
				}
			}
			check("birth", birth, w.birth)
			check("death", death, w.death)
			if tc.name == "death at two Places that do not fold" {
				persons, err := conclusionheaders.PersonsByIDs(f.c, [][]byte{w.person})
				must(t, err)
				header := persons[0].Death.Places
				if len(death.Places) != 2 || len(header) != 2 {
					t.Fatalf("death places %d, header %d", len(death.Places), len(header))
				}
				for i := range header {
					if !bytes.Equal(death.Places[i].Entity.ID, header[i].Entity.ID) {
						t.Fatalf("death place %d is not the header's", i)
					}
				}
			}
		})
	}
}

func bytes16(b byte) []byte { return bytes.Repeat([]byte{b}, 16) }

func TestPersonDerivedRows(t *testing.T) {
	t.Run("graph queries do not grow with events and Places", func(t *testing.T) {
		f := newFixture(t)
		york := f.placeSubject("York", nil, nil, 0, 0)
		small := f.lifePerson()
		smallBirth := f.lifeEvent(small, "birth", "date", 1817, 2, york)
		big := f.lifePerson()
		var bigPlaces []subjects.Subject
		for i := range 2 {
			bigPlaces = append(bigPlaces, f.placeSubject(fmt.Sprintf("Town %d", i), nil, nil, int64(4+2*i), 0))
		}
		bigBirths := []subjects.Subject{
			f.lifeEvent(big, "birth", "date", 1818, 10, bigPlaces...),
			f.lifeEvent(big, "birth", "date", 1819, 12, bigPlaces...),
		}
		smallID, bigID := f.promoteSubject(small), f.promoteSubject(big)
		f.promoteSubject(smallBirth)
		for _, b := range bigBirths {
			f.promoteSubject(b)
		}
		f.promoteSubject(york)
		for _, p := range bigPlaces {
			f.promoteSubject(p)
		}
		count := func(id []byte) (int, []conclusiondetails.DerivedRow) {
			t.Helper()
			g := f.c.Graph()
			g.Drop()
			_, err := g.Vocabulary()
			must(t, err)
			n := 0
			g.SetQueryCounterForTest(func() { n++ })
			defer g.SetQueryCounterForTest(nil)
			rows, err := conclusiondetails.PersonDerivedRows(f.c, id)
			must(t, err)
			return n, rows
		}
		smallN, smallRows := count(smallID)
		bigN, bigRows := count(bigID)
		if len(smallRows) != 2 || len(bigRows) != 2 {
			t.Fatalf("rows: small %d, big %d, want a birth date and a birth place each", len(smallRows), len(bigRows))
		}
		if len(bigRows[0].Owners) != 2 || len(bigRows[1].Owners) != 2 {
			t.Fatalf("big owners: %d dates, %d places", len(bigRows[0].Owners), len(bigRows[1].Owners))
		}
		if smallN == 0 || smallN != bigN {
			t.Fatalf("graph queries: %d for one event and Place, %d for two events and two Places", smallN, bigN)
		}
	})

	t.Run("born and died at York", func(t *testing.T) {
		f := newFixture(t)
		york := f.placeSubject("York", nil, nil, 0, 0)
		p := f.lifePerson()
		birth := f.lifeEvent(p, "birth", "date", 1817, 2, york)
		death := f.lifeEvent(p, "death", "date", 1890, 4, york)
		id := f.promoteSubject(p)
		f.promoteSubject(birth)
		f.promoteSubject(death)
		yorkID := f.promoteSubject(york)
		rows, err := conclusiondetails.PersonDerivedRows(f.c, id)
		must(t, err)
		kinds := map[conclusiondetails.DerivedRowKind]conclusiondetails.DerivedRow{}
		for _, r := range rows {
			kinds[r.Kind] = r
		}
		bp, dp := kinds[conclusiondetails.DerivedRowBirthPlace], kinds[conclusiondetails.DerivedRowDeathPlace]
		if len(bp.Owners) != 1 || len(dp.Owners) != 1 ||
			!bytes.Equal(bp.Owners[0].Entity.ID, yorkID) || !bytes.Equal(dp.Owners[0].Entity.ID, yorkID) {
			t.Fatalf("place rows %+v / %+v", bp, dp)
		}
		if !reflect.DeepEqual(bp.Owners[0].Field, dp.Owners[0].Field) || len(bp.Owners[0].Field.Outcomes) != 1 {
			t.Fatalf("York fields differ: %+v / %+v", bp.Owners[0].Field, dp.Owners[0].Field)
		}
		if bp.Owners[0].Event != nil || kinds[conclusiondetails.DerivedRowBirthDate].Owners[0].Event == nil {
			t.Fatal("only date owners carry an Event header")
		}
	})
}
