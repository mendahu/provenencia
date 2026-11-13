package conclusionsummaries_test

import (
	"bytes"
	"slices"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/conclusionsummaries"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/promote/promotetest"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/subjectpositions"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjects/subjectstest"
	"github.com/mendahu/provenencia/core/datevalue/datevaluetest"
	"github.com/mendahu/provenencia/core/eventtitle"
	"github.com/mendahu/provenencia/core/namevalue/namevaluetest"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/writes"
)

func (f *fixture) at(s subjects.Subject, x, y int64) {
	f.T.Helper()
	_, err := writes.Call(f.C, writes.Op{}, func(tx *database.Tx) (subjectpositions.Position, []rowchange.Change, error) {
		return subjectpositions.Set(tx, s.ID, x, y)
	})
	must(f.T, err)
}

func (f *fixture) promoteSubject(s subjects.Subject) []byte {
	f.T.Helper()
	res := promotetest.Save(f.T, f.C, userID, promote.Input{SubjectID: s.ID})
	return res.Entity.ID
}

func (f *fixture) named(s subjects.Subject, form string) {
	f.T.Helper()
	f.Cite(f.Artifact, s, observations.Input{PropertyID: f.name.ID, Name: namevaluetest.Western(form)})
}

func TestCanonicalWalks(t *testing.T) {
	f := newFixture(t)
	james := subjectstest.On(f.T, f.C, userID, f.graph, "person", "")
	f.named(james, "James Robins")
	f.at(james, 0, 0)
	mary := subjectstest.On(f.T, f.C, userID, f.graph, "person", "")
	f.named(mary, "Mary Smith")
	f.at(mary, 0, 2)
	ann := subjectstest.On(f.T, f.C, userID, f.graph, "person", "")
	f.named(ann, "Ann Lee")
	f.at(ann, 0, 4)
	unnamed := subjectstest.On(f.T, f.C, userID, f.graph, "person", "")
	f.at(unnamed, 0, 6)
	witness := subjectstest.On(f.T, f.C, userID, f.graph, "person", "")
	f.named(witness, "Witness Webb")
	f.at(witness, 0, 8)

	birth := subjectstest.On(f.T, f.C, userID, f.graph, "event", "")
	f.Cite(f.Artifact, birth,
		observations.Input{PropertyID: f.Prop("event_type").ID, ValueTermID: f.Term("event_type", "birth").ID},
		observations.Input{PropertyID: f.Prop("date").ID, Date: datevaluetest.YMD(1817, 0, 0)},
	)
	f.at(birth, 4, 0)
	death := subjectstest.On(f.T, f.C, userID, f.graph, "event", "")
	f.Cite(f.Artifact, death,
		observations.Input{PropertyID: f.Prop("event_type").ID, ValueTermID: f.Term("event_type", "death").ID},
		observations.Input{PropertyID: f.Prop("start_date").ID, Date: datevaluetest.YMD(1880, 0, 0)},
	)
	f.at(death, 4, 2)
	marriage := subjectstest.On(f.T, f.C, userID, f.graph, "event", "")
	f.Cite(f.Artifact, marriage, observations.Input{PropertyID: f.Prop("event_type").ID, ValueTermID: f.Term("event_type", "marriage").ID})
	f.at(marriage, 4, 4)
	census := subjectstest.On(f.T, f.C, userID, f.graph, "event", "")
	f.Cite(f.Artifact, census, observations.Input{PropertyID: f.Prop("event_type").ID, ValueTermID: f.Term("event_type", "census").ID})
	f.at(census, 4, 6)
	baptism := subjectstest.On(f.T, f.C, userID, f.graph, "event", "")
	f.Cite(f.Artifact, baptism, observations.Input{PropertyID: f.Prop("event_type").ID, ValueTermID: f.Term("event_type", "baptism").ID})
	f.at(baptism, 4, 8)
	eventTypeID := f.Prop("event_type").ID
	fireTerm, _, err := writes.Run(f.C, writes.Op{Action: "create_property_term", UserID: userID},
		func(tx *database.Tx) (propertyterms.Term, []rowchange.Change, error) {
			return propertyterms.Create(tx, userID, eventTypeID, "Fire", "")
		})
	must(t, err)
	fire := subjectstest.On(f.T, f.C, userID, f.graph, "event", "")
	f.Cite(f.Artifact, fire, observations.Input{PropertyID: f.Prop("event_type").ID, ValueTermID: fireTerm.ID})
	f.at(fire, 4, 10)

	york := subjectstest.On(f.T, f.C, userID, f.graph, "place", "")
	f.Cite(f.Artifact, york,
		observations.Input{PropertyID: f.Prop("toponym").ID, Scalar: propertymeta.TextScalar("York")},
		observations.Input{PropertyID: f.Prop("toponym").ID, Scalar: propertymeta.TextScalar("Tkaronto")},
	)
	f.at(york, 8, 0)
	toronto := subjectstest.On(f.T, f.C, userID, f.graph, "place", "")
	f.Cite(f.Artifact, toronto, observations.Input{PropertyID: f.Prop("toponym").ID, Scalar: propertymeta.TextScalar("Toronto")})
	f.at(toronto, 8, 2)

	f.Participation(f.Source, f.Artifact, james, birth, "subject")
	f.Participation(f.Source, f.Artifact, witness, birth, "witness")
	f.Participation(f.Source, f.Artifact, james, death, "subject")
	f.Participation(f.Source, f.Artifact, james, marriage, "subject")
	f.Participation(f.Source, f.Artifact, mary, marriage, "subject")
	f.Participation(f.Source, f.Artifact, james, census, "subject")
	f.Participation(f.Source, f.Artifact, mary, census, "subject")
	f.Participation(f.Source, f.Artifact, ann, census, "subject")
	f.Participation(f.Source, f.Artifact, unnamed, baptism, "subject")
	f.Locate(f.Source, f.Artifact, birth, york)
	f.Locate(f.Source, f.Artifact, birth, toronto)
	f.Locate(f.Source, f.Artifact, death, toronto)
	f.Locate(f.Source, f.Artifact, fire, york)

	jamesID := f.promoteSubject(james)
	f.promoteSubject(mary)
	f.promoteSubject(ann)
	unnamedID := f.promoteSubject(unnamed)
	f.promoteSubject(witness)
	birthID := f.promoteSubject(birth)
	deathID := f.promoteSubject(death)
	marriageID := f.promoteSubject(marriage)
	censusID := f.promoteSubject(census)
	baptismID := f.promoteSubject(baptism)
	fireID := f.promoteSubject(fire)
	yorkID := f.promoteSubject(york)
	torontoID := f.promoteSubject(toronto)

	var jamesSummary conclusionsummaries.PersonSummary
	for _, h := range f.list() {
		if bytes.Equal(h.Entity.ID, jamesID) {
			jamesSummary = h
		}
	}
	if jamesSummary.Birth.Date == nil || jamesSummary.Birth.Date.StartYear == nil || *jamesSummary.Birth.Date.StartYear != 1817 {
		t.Fatalf("birth date %+v", jamesSummary.Birth.Date)
	}
	if jamesSummary.Birth.DateCount != 1 {
		t.Fatalf("birth date count %d", jamesSummary.Birth.DateCount)
	}
	if jamesSummary.Death.Date == nil || jamesSummary.Death.Date.StartYear == nil || *jamesSummary.Death.Date.StartYear != 1880 {
		t.Fatalf("death date %+v", jamesSummary.Death.Date)
	}
	if got := placeNames(jamesSummary.Birth.Places); !sameSet(got, []string{"York", "Tkaronto", "Toronto"}) {
		t.Fatalf("birth places %v", got)
	}
	for _, name := range placeNames(jamesSummary.Birth.Places) {
		if bytes.Contains([]byte(name), []byte(",")) {
			t.Fatalf("chain %q", name)
		}
	}
	yorkPlace := chainedPlace(jamesSummary.Birth.Places, "York")
	if !sameSet(yorkPlace.Names, []string{"York", "Tkaronto"}) {
		t.Fatalf("York names %+v", yorkPlace)
	}
	if got := placeNames(jamesSummary.Death.Places); !sameSet(got, []string{"Toronto"}) {
		t.Fatalf("death places %v", got)
	}
	t.Run("life facts name the event and places they read", func(t *testing.T) {
		if e := jamesSummary.Birth.Event; e == nil || !bytes.Equal(e.ID, birthID) || e.Ref == "" || len(e.SubjectTypeID) != 16 {
			t.Fatalf("birth event %+v", e)
		}
		if e := jamesSummary.Death.Event; e == nil || !bytes.Equal(e.ID, deathID) {
			t.Fatalf("death event %+v", e)
		}
		if p := chainedPlace(jamesSummary.Birth.Places, "York"); !bytes.Equal(p.Entity.ID, yorkID) || p.Entity.Ref == "" {
			t.Fatalf("York entity %+v", p.Entity)
		}
		if p := chainedPlace(jamesSummary.Death.Places, "Toronto"); !bytes.Equal(p.Entity.ID, torontoID) {
			t.Fatalf("Toronto entity %+v", p.Entity)
		}
	})

	events := map[string]conclusionsummaries.EventSummary{}
	for _, h := range f.events() {
		events[string(h.Entity.ID)] = h
	}
	birthSummary := events[string(birthID)]
	if forms := subjectForms(birthSummary); !sameSet(forms, []string{"James Robins"}) {
		t.Fatalf("birth subjects %v", forms)
	}
	if birthSummary.Type.Term == nil || birthSummary.Type.Term.Key != "birth" {
		t.Fatalf("birth type %+v", birthSummary.Type.Term)
	}
	if len(birthSummary.Places) != 2 {
		t.Fatalf("birth locations %d", len(birthSummary.Places))
	}
	marriageSummary := events[string(marriageID)]
	if forms := subjectForms(marriageSummary); !sameSet(forms, []string{"James Robins", "Mary Smith"}) {
		t.Fatalf("marriage subjects %v", forms)
	}
	if marriageSummary.Type.Term.Key != "marriage" {
		t.Fatalf("marriage type %s", marriageSummary.Type.Term.Key)
	}
	if forms := subjectForms(events[string(censusID)]); len(forms) != 3 {
		t.Fatalf("census subjects %v", forms)
	}
	if events[string(baptismID)].Subjects[0].Name != nil || !bytes.Equal(events[string(baptismID)].Subjects[0].Entity.ID, unnamedID) {
		t.Fatalf("unnamed %+v", events[string(baptismID)].Subjects)
	}
	fireSummary := events[string(fireID)]
	if len(fireSummary.Subjects) != 0 || fireSummary.Type.Term.Label != "Fire" {
		t.Fatalf("fire %+v subjects %d", fireSummary.Type.Term, len(fireSummary.Subjects))
	}
	if got := placeNames(fireSummary.Places); !sameSet(got, []string{"York", "Tkaronto"}) {
		t.Fatalf("fire places %v", got)
	}
	if len(fireSummary.Places) != 1 || !bytes.Equal(fireSummary.Places[0].Entity.ID, yorkID) {
		t.Fatalf("fire place entity %+v", fireSummary.Places)
	}

	t.Run("each event's title rule comes from its own parts", func(t *testing.T) {
		tests := []struct {
			name     string
			id       []byte
			want     eventtitle.Rule
			subjects int
		}{
			{name: "birth of one subject", id: birthID, want: eventtitle.RuleSubject, subjects: 1},
			{name: "marriage of two", id: marriageID, want: eventtitle.RuleCouple, subjects: 2},
			{name: "census of three", id: censusID, want: eventtitle.RuleSubjects, subjects: 3},
			{name: "baptism of an unnamed person", id: baptismID, want: eventtitle.RuleSubject, subjects: 1},
			{name: "fire at York", id: fireID, want: eventtitle.RuleTypeAtPlace},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				title := events[string(tt.id)].Title
				if title.Rule != tt.want || len(title.Parts.Subjects) != tt.subjects {
					t.Fatalf("rule %s with %d subjects", title.Rule, len(title.Parts.Subjects))
				}
			})
		}
		if p := events[string(birthID)].Title.Parts; p.TypeKey != "birth" || p.TypeLabel != "Birth" {
			t.Fatalf("birth parts %+v", p)
		}
		if p := events[string(fireID)].Title.Parts; p.Place != "York" || p.TypeLabel != "Fire" {
			t.Fatalf("fire parts %+v", p)
		}
		if s := events[string(baptismID)].Title.Parts.Subjects; s[0].Name != nil {
			t.Fatalf("unnamed subject carried a name %+v", s[0].Name)
		}
	})

	again := map[string][]string{}
	for _, h := range f.events() {
		again[string(h.Entity.ID)] = subjectForms(h)
	}
	if !sameSet(again[string(censusID)], subjectForms(events[string(censusID)])) ||
		!bytes.Equal([]byte(joinForms(again[string(censusID)])), []byte(joinForms(subjectForms(events[string(censusID)])))) {
		t.Fatal("subject order moved")
	}

	// A write reports the rows composed from what it moved. A toponym cited on
	// York reaches the events held there and the people they are about.
	f.Cite(f.Artifact, york, observations.Input{PropertyID: f.Prop("toponym").ID, Scalar: propertymeta.TextScalar("Yorke")})
	deps := f.Reported
	if !database.ContainsBlobID(deps, birthID) || !database.ContainsBlobID(deps, fireID) || !database.ContainsBlobID(deps, jamesID) {
		t.Fatalf("a toponym cited on York did not report the birth, the fire and James (%d ids)", len(deps))
	}
	// A name cited on James reaches the events he is a subject of.
	f.Cite(f.Artifact, james, observations.Input{PropertyID: f.Prop("name").ID, Name: namevaluetest.Western("Jim Robins")})
	fromJames := f.Reported
	if !database.ContainsBlobID(fromJames, birthID) || !database.ContainsBlobID(fromJames, deathID) || !database.ContainsBlobID(fromJames, marriageID) || !database.ContainsBlobID(fromJames, censusID) {
		t.Fatal("a name cited on James missed an event he is a subject of")
	}
	// James's summary embeds his birth's date: a date cited on the birth
	// reaches him.
	f.Cite(f.Artifact, birth, observations.Input{PropertyID: f.Prop("date").ID, Date: datevaluetest.YMD(1818, 0, 0)})
	fromBirth := f.Reported
	if !database.ContainsBlobID(fromBirth, jamesID) {
		t.Fatalf("a date cited on the birth did not report James (%d ids)", len(fromBirth))
	}
}

func placeNames(places []conclusionsummaries.ChainedPlace) []string {
	var out []string
	for _, p := range places {
		out = append(out, p.Names...)
	}
	return out
}

func chainedPlace(places []conclusionsummaries.ChainedPlace, name string) conclusionsummaries.ChainedPlace {
	for _, p := range places {
		for _, n := range p.Names {
			if n == name {
				return p
			}
		}
	}
	return conclusionsummaries.ChainedPlace{}
}

func subjectForms(h conclusionsummaries.EventSummary) []string {
	var out []string
	for _, s := range h.Subjects {
		if s.Name == nil {
			out = append(out, "")
			continue
		}
		out = append(out, s.Name.Form)
	}
	return out
}

func joinForms(forms []string) string {
	out := ""
	for _, f := range forms {
		out += f + "\x00"
	}
	return out
}

func sameSet(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	seen := map[string]int{}
	for _, g := range got {
		seen[g]++
	}
	for _, w := range want {
		seen[w]--
		if seen[w] < 0 {
			return false
		}
	}
	return true
}

// Two surviving births are a disagreement: the summary counts them and leads
// with the earliest dated one, whatever the refs say.
func TestCompetingLifeEvents(t *testing.T) {
	tests := []struct {
		name      string
		years     []int // 0 = undated
		wantCount int
		wantYear  int // 0 = no date
	}{
		{name: "one birth", years: []int{1817}, wantCount: 1, wantYear: 1817},
		{name: "the earliest dated birth leads", years: []int{1819, 1815, 1817}, wantCount: 3, wantYear: 1815},
		{name: "a dated birth leads an undated one", years: []int{0, 1821}, wantCount: 2, wantYear: 1821},
		{name: "undated births still count", years: []int{0, 0}, wantCount: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			james := subjectstest.On(f.T, f.C, userID, f.graph, "person", "")
			f.named(james, "James Robins")
			f.at(james, 0, 0)
			var births []subjects.Subject
			for i, year := range tt.years {
				birth := subjectstest.On(f.T, f.C, userID, f.graph, "event", "")
				in := []observations.Input{{PropertyID: f.Prop("event_type").ID, ValueTermID: f.Term("event_type", "birth").ID}}
				if year != 0 {
					in = append(in, observations.Input{PropertyID: f.Prop("date").ID, Date: datevaluetest.YMD(year, 0, 0)})
				}
				f.Cite(f.Artifact, birth, in...)
				f.at(birth, 4, int64(2*i))
				f.Participation(f.Source, f.Artifact, james, birth, "subject")
				births = append(births, birth)
			}
			jamesID := f.promoteSubject(james)
			for _, b := range births {
				f.promoteSubject(b)
			}
			var birth conclusionsummaries.LifeFacts
			for _, h := range f.list() {
				if bytes.Equal(h.Entity.ID, jamesID) {
					birth = h.Birth
				}
			}
			if birth.EventCount != tt.wantCount || birth.Event == nil {
				t.Fatalf("event count %d event %+v", birth.EventCount, birth.Event)
			}
			var year int
			if birth.Date != nil && birth.Date.StartYear != nil {
				year = *birth.Date.StartYear
			}
			if year != tt.wantYear {
				t.Fatalf("lead year %d, want %d", year, tt.wantYear)
			}
		})
	}
}

func TestEventRowReadsPlaceEntries(t *testing.T) {
	f := newFixture(t)
	city := f.placeSubject("City", nil, nil, 0, 0)
	p := f.lifePerson()
	var events [][]byte
	for i := range 3 {
		events = append(events, f.promoteSubject(f.lifeEvent(p, "birth", "date", 1800+i, int64(2+2*i), city)))
	}
	cityID := f.promoteSubject(city)
	f.promoteSubject(p)
	g := f.C.Graph()
	g.Drop()
	rows, err := conclusionsummaries.EventsByIDs(f.C, events[:1])
	must(t, err)
	if len(rows) != 1 || len(rows[0].Places) != 1 || !slices.Equal(rows[0].Places[0].Names, []string{"City"}) {
		t.Fatalf("places %+v", rows)
	}
	if _, ok := g.Cached(cityID); ok {
		t.Error("the city node was loaded to name an event's place")
	}
}

func TestPersonRowReadsPlaceEntries(t *testing.T) {
	f := newFixture(t)
	town := f.placeSubject("Town", nil, nil, 0, 0)
	city := f.placeSubject("City", nil, nil, 2, 0)
	f.placeRel(town, city, propertyterms.KeyPartOf, nil, nil)
	p := f.lifePerson()
	birth := f.lifeEvent(p, "birth", "date", 1800, 2, town)
	townID, cityID := f.promoteSubject(town), f.promoteSubject(city)
	f.promoteSubject(birth)
	pid := f.promoteSubject(p)
	g := f.C.Graph()
	g.Drop()
	rows, err := conclusionsummaries.PersonsByIDs(f.C, [][]byte{pid})
	must(t, err)
	if len(rows) != 1 || len(rows[0].Birth.Places) != 1 {
		t.Fatalf("rows %+v", rows)
	}
	got := rows[0].Birth.Places[0]
	if !slices.Equal(got.Names, []string{"Town"}) || !slices.Equal(got.Parents, []string{"City"}) {
		t.Fatalf("birth place %+v, want Town under City", got)
	}
	for name, id := range map[string][]byte{"town": townID, "city": cityID} {
		if _, ok := g.Cached(id); ok {
			t.Errorf("the %s node was loaded to name a person's birth place", name)
		}
	}
}
