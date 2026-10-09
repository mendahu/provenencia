package conclusionheaders_test

import (
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/conclusionheaders"
	"github.com/mendahu/provenencia/core/database/connect"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/subjects"
)

func year(y int) *int { return &y }

// placeSubject cites a Place (and optional period) but does not promote — call
// promoteSubject after placeRel so bridges file with the ends.
func (f *fixture) placeSubject(name string, start, end *int, x, y int64) subjects.Subject {
	f.t.Helper()
	s := f.bare("place")
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
	res, err := connect.CreateCitedBridge(f.c, userID, connect.CreateInput{
		SourceID: f.source.ID, FromSubjectID: from.ID, ToSubjectID: to.ID, BridgeTypeKey: "place_relationship",
		Citation: citations.CreateInput{ArtifactID: f.artifact.ID, LocatorJSON: locator},
		Observations: []observations.Input{
			{PropertyID: f.prop("from").ID, ValueSubjectID: from.ID},
			{PropertyID: f.prop("to").ID, ValueSubjectID: to.ID},
			{PropertyID: f.prop("place_relationship_type").ID, ValueTermID: term.ID},
		},
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
		_, err := observations.AddToCitation(f.c, userID, res.Citation.ID, extra)
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
	db, err := f.c.DB()
	must(t, err)

	parents := func(y int) []string {
		t.Helper()
		got, err := conclusionheaders.ParentsAtDate(db, torontoID, *pointYear(y))
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
	abt.Qualifier = datevalues.QualifierABT
	got, err := conclusionheaders.ParentsAtDate(db, torontoID, abt)
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
	db, err := f.c.DB()
	must(t, err)
	got, err := conclusionheaders.ParentsAtDate(db, farmID, conclusionheaders.TodayDate())
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
	db, err := f.c.DB()
	must(t, err)
	before, err := conclusionheaders.ParentsAtDate(db, irelandID, *pointYear(1900))
	must(t, err)
	if !sameSet(before, []string{"United Kingdom"}) {
		t.Fatalf("1900 %v", before)
	}
	after, err := conclusionheaders.ParentsAtDate(db, irelandID, *pointYear(1923))
	must(t, err)
	if len(after) != 0 {
		t.Fatalf("1923 still has parents %v", after)
	}
	// Both places continue: UK still lists Ireland as a former part only via
	// the dated link — PartsAtDate at 1923 is empty.
	parts, err := conclusionheaders.PartsAtDate(db, ukID, *pointYear(1923))
	must(t, err)
	if len(parts) != 0 {
		t.Fatalf("UK parts in 1923 %v", parts)
	}
	parts1900, err := conclusionheaders.PartsAtDate(db, ukID, *pointYear(1900))
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
	db, err := f.c.DB()
	must(t, err)
	parents, err := conclusionheaders.ParentsAtDate(db, yorkID, *pointYear(1834))
	must(t, err)
	if len(parents) != 0 {
		t.Fatalf("succession leaked into chain %v", parents)
	}
	pred, succ, err := conclusionheaders.SuccessionNames(db, yorkID)
	must(t, err)
	if len(pred) != 0 || !sameSet(succ, []string{"Toronto"}) {
		t.Fatalf("york succession pred=%v succ=%v", pred, succ)
	}
	pred, succ, err = conclusionheaders.SuccessionNames(db, torontoID)
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
	db, err := f.c.DB()
	must(t, err)
	pred, succ, err := conclusionheaders.SuccessionNames(db, newID)
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

	person := f.bare("person")
	f.named(person, "James")
	f.at(person, 0, 4)
	birth := f.bare("event")
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

	db, err := f.c.DB()
	must(t, err)
	persons, err := conclusionheaders.ListPersons(db)
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
	birth2 := f.bare("event")
	f.at(birth2, 4, 6)
	f.cite(birth2,
		observations.Input{PropertyID: f.prop("event_type").ID, ValueTermID: f.term("event_type", "birth").ID},
		observations.Input{PropertyID: f.prop("date").ID, Date: pointYear(1850)},
	)
	person2 := f.bare("person")
	f.named(person2, "Mary")
	f.at(person2, 0, 6)
	f.participation(person2, birth2, "subject")
	f.location(birth2, toronto)
	f.location(birth2, scotland)
	f.promoteSubject(person2)
	f.promoteSubject(birth2)
	f.promoteSubject(scotland)
	persons, err = conclusionheaders.ListPersons(db)
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
	db, err := f.c.DB()
	must(t, err)
	headers, err := conclusionheaders.PlacesByIDs(db, [][]byte{torontoID})
	must(t, err)
	if len(headers) != 1 {
		t.Fatalf("%d", len(headers))
	}
	h := headers[0]
	must(t, conclusionheaders.AttachPlaceRelationships(db, &h))
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
	yh, err := conclusionheaders.PlacesByIDs(db, [][]byte{yorkID})
	must(t, err)
	must(t, conclusionheaders.AttachPlaceRelationships(db, &yh[0]))
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
	db, err := f.c.DB()
	must(t, err)
	ih, err := conclusionheaders.PlacesByIDs(db, [][]byte{irelandID})
	must(t, err)
	must(t, conclusionheaders.AttachPlaceRelationships(db, &ih[0]))
	if len(ih[0].PartOf) != 1 || ih[0].PartOf[0].Title != "United Kingdom" {
		t.Fatalf("ireland part_of %+v", ih[0].PartOf)
	}
	if ih[0].PartOf[0].EndDate == nil || ih[0].PartOf[0].EndDate.StartYear == nil ||
		*ih[0].PartOf[0].EndDate.StartYear != 1922 {
		t.Fatalf("ireland membership end %+v", ih[0].PartOf[0])
	}
	uh, err := conclusionheaders.PlacesByIDs(db, [][]byte{ukID})
	must(t, err)
	must(t, conclusionheaders.AttachPlaceRelationships(db, &uh[0]))
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
	if !hasID(deps, torontoID) {
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
	db, err := f.c.DB()
	must(t, err)
	for _, y := range []int{1, 1000, 2020} {
		got, err := conclusionheaders.ParentsAtDate(db, childID, *pointYear(y))
		must(t, err)
		if !sameSet(got, []string{"Parent"}) {
			t.Fatalf("%d parents %v", y, got)
		}
	}
}

// A chain or a fold only looks up, so its graph holds the seed's ancestors
// and never their other parts. A Place page also reads one hop of parts.
func TestPlaceGraphReadsOnlyWhatTheCallerNeeds(t *testing.T) {
	f := newFixture(t)
	toronto := f.placeSubject("Toronto", nil, nil, 0, 0)
	hamilton := f.placeSubject("Hamilton", nil, nil, 2, 0)
	ontario := f.placeSubject("Ontario", nil, nil, 4, 0)
	canada := f.placeSubject("Canada", nil, nil, 6, 0)
	f.placeRel(toronto, ontario, propertyterms.KeyPartOf, nil, nil)
	f.placeRel(hamilton, ontario, propertyterms.KeyPartOf, nil, nil)
	f.placeRel(ontario, canada, propertyterms.KeyPartOf, nil, nil)
	torontoID := f.promoteSubject(toronto)
	f.promoteSubject(hamilton)
	ontarioID := f.promoteSubject(ontario)
	f.promoteSubject(canada)
	db, err := f.c.DB()
	must(t, err)

	up, err := conclusionheaders.PlaceGraphRefsForTest(db, torontoID, false)
	must(t, err)
	if got := strings.Join(up, ","); got != "Canada,Ontario,Toronto" {
		t.Fatalf("ancestor reach read %s, want Canada,Ontario,Toronto", got)
	}
	detail, err := conclusionheaders.PlaceGraphRefsForTest(db, ontarioID, true)
	must(t, err)
	if got := strings.Join(detail, ","); got != "Canada,Hamilton,Ontario,Toronto" {
		t.Fatalf("detail reach read %s, want Canada,Hamilton,Ontario,Toronto", got)
	}
}
