package conclusionheaders_test

import (
	"bytes"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/conclusionheaders"
	"github.com/mendahu/provenencia/core/database/connect"
	"github.com/mendahu/provenencia/core/database/namevalues/namevaluestest"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/subjectpositions"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/eventtitle"
	"github.com/mendahu/provenencia/core/writes"
)

func (f *fixture) bare(kind string) subjects.Subject {
	f.t.Helper()
	typeID := f.typeID(kind)
	s, err := writes.Call(f.c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (subjects.Subject, []rowchange.Change, error) {
		return subjects.Create(tx, userID, subjects.CreateInput{
			SourceID: f.source.ID, SubjectTypeID: typeID,
		}, nil)
	})
	must(f.t, err)
	return s
}

func (f *fixture) cite(s subjects.Subject, in ...observations.Input) {
	f.t.Helper()
	for i := range in {
		in[i].SubjectID = s.ID
	}
	_, err := writes.Call(f.c, writes.Op{Action: "create_citation_with_observations", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
		return citations.CreateWithObservations(tx, userID, citations.CreateInput{
			ArtifactID: f.artifact.ID, LocatorJSON: locator,
		}, in)
	})
	must(f.t, err)
}

func (f *fixture) at(s subjects.Subject, x, y int64) {
	f.t.Helper()
	_, err := writes.Call(f.c, writes.Op{}, func(tx *database.Tx) (subjectpositions.Position, []rowchange.Change, error) {
		return subjectpositions.Set(tx, s.ID, x, y)
	})
	must(f.t, err)
}

func (f *fixture) promoteSubject(s subjects.Subject) []byte {
	f.t.Helper()
	res, err := writes.Call(f.c, writes.Op{Action: "promote_subject", UserID: userID}, func(tx *database.Tx) (promote.Result, []rowchange.Change, error) {
		return promote.Save(tx, userID, promote.Input{SubjectID: s.ID})
	})
	must(f.t, err)
	return res.Entity.ID
}

func (f *fixture) participation(person, event subjects.Subject, role string) {
	f.t.Helper()
	f.bridge("participation", person, event, []observations.Input{
		{PropertyID: f.prop("person").ID, ValueSubjectID: person.ID},
		{PropertyID: f.prop("event").ID, ValueSubjectID: event.ID},
		{PropertyID: f.prop("role").ID, ValueTermID: f.term("role", role).ID},
	})
}

func (f *fixture) location(event, place subjects.Subject) {
	f.t.Helper()
	f.bridge("location", event, place, []observations.Input{
		{PropertyID: f.prop("event").ID, ValueSubjectID: event.ID},
		{PropertyID: f.prop("place").ID, ValueSubjectID: place.ID},
	})
}

func (f *fixture) bridge(kind string, from, to subjects.Subject, obs []observations.Input) {
	f.t.Helper()
	_, err := writes.Call(f.c, writes.Op{Action: "create_cited_bridge", UserID: userID}, func(tx *database.Tx) (connect.Result, []rowchange.Change, error) {
		return connect.CreateCitedBridge(tx, userID, connect.CreateInput{
			SourceID: f.source.ID, FromSubjectID: from.ID, ToSubjectID: to.ID, BridgeTypeKey: kind,
			Citation:     citations.CreateInput{ArtifactID: f.artifact.ID, LocatorJSON: locator},
			Observations: obs,
		})
	})
	must(f.t, err)
}

func (f *fixture) named(s subjects.Subject, form string) {
	f.t.Helper()
	f.cite(s, observations.Input{PropertyID: f.name.ID, Name: namevaluestest.Western(form)})
}

func TestCanonicalWalks(t *testing.T) {
	f := newFixture(t)
	james := f.bare("person")
	f.named(james, "James Robins")
	f.at(james, 0, 0)
	mary := f.bare("person")
	f.named(mary, "Mary Smith")
	f.at(mary, 0, 2)
	ann := f.bare("person")
	f.named(ann, "Ann Lee")
	f.at(ann, 0, 4)
	unnamed := f.bare("person")
	f.at(unnamed, 0, 6)
	witness := f.bare("person")
	f.named(witness, "Witness Webb")
	f.at(witness, 0, 8)

	birth := f.bare("event")
	f.cite(birth,
		observations.Input{PropertyID: f.prop("event_type").ID, ValueTermID: f.term("event_type", "birth").ID},
		observations.Input{PropertyID: f.prop("date").ID, Date: pointYear(1817)},
	)
	f.at(birth, 4, 0)
	death := f.bare("event")
	f.cite(death,
		observations.Input{PropertyID: f.prop("event_type").ID, ValueTermID: f.term("event_type", "death").ID},
		observations.Input{PropertyID: f.prop("start_date").ID, Date: pointYear(1880)},
	)
	f.at(death, 4, 2)
	marriage := f.bare("event")
	f.cite(marriage, observations.Input{PropertyID: f.prop("event_type").ID, ValueTermID: f.term("event_type", "marriage").ID})
	f.at(marriage, 4, 4)
	census := f.bare("event")
	f.cite(census, observations.Input{PropertyID: f.prop("event_type").ID, ValueTermID: f.term("event_type", "census").ID})
	f.at(census, 4, 6)
	baptism := f.bare("event")
	f.cite(baptism, observations.Input{PropertyID: f.prop("event_type").ID, ValueTermID: f.term("event_type", "baptism").ID})
	f.at(baptism, 4, 8)
	eventTypeID := f.prop("event_type").ID
	fireTerm, _, err := writes.Run(f.c, writes.Op{Action: "create_property_term", UserID: userID},
		func(tx *database.Tx) (propertyterms.Term, []rowchange.Change, error) {
			return propertyterms.Create(tx, userID, eventTypeID, "Fire", "")
		})
	must(t, err)
	fire := f.bare("event")
	f.cite(fire, observations.Input{PropertyID: f.prop("event_type").ID, ValueTermID: fireTerm.ID})
	f.at(fire, 4, 10)

	york := f.bare("place")
	f.cite(york,
		observations.Input{PropertyID: f.prop("toponym").ID, ValueText: "York", HasText: true},
		observations.Input{PropertyID: f.prop("toponym").ID, ValueText: "Tkaronto", HasText: true},
	)
	f.at(york, 8, 0)
	toronto := f.bare("place")
	f.cite(toronto, observations.Input{PropertyID: f.prop("toponym").ID, ValueText: "Toronto", HasText: true})
	f.at(toronto, 8, 2)

	f.participation(james, birth, "subject")
	f.participation(witness, birth, "witness")
	f.participation(james, death, "subject")
	f.participation(james, marriage, "subject")
	f.participation(mary, marriage, "subject")
	f.participation(james, census, "subject")
	f.participation(mary, census, "subject")
	f.participation(ann, census, "subject")
	f.participation(unnamed, baptism, "subject")
	f.location(birth, york)
	f.location(birth, toronto)
	f.location(death, toronto)
	f.location(fire, york)

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

	var jamesHeader conclusionheaders.PersonHeader
	for _, h := range f.list() {
		if bytes.Equal(h.Entity.ID, jamesID) {
			jamesHeader = h
		}
	}
	if jamesHeader.Birth.Date == nil || jamesHeader.Birth.Date.StartYear == nil || *jamesHeader.Birth.Date.StartYear != 1817 {
		t.Fatalf("birth date %+v", jamesHeader.Birth.Date)
	}
	if jamesHeader.Birth.DateCount != 1 {
		t.Fatalf("birth date count %d", jamesHeader.Birth.DateCount)
	}
	if jamesHeader.Death.Date == nil || jamesHeader.Death.Date.StartYear == nil || *jamesHeader.Death.Date.StartYear != 1880 {
		t.Fatalf("death date %+v", jamesHeader.Death.Date)
	}
	if got := placeNames(jamesHeader.Birth.Places); !sameSet(got, []string{"York", "Tkaronto", "Toronto"}) {
		t.Fatalf("birth places %v", got)
	}
	for _, name := range placeNames(jamesHeader.Birth.Places) {
		if bytes.Contains([]byte(name), []byte(",")) {
			t.Fatalf("chain %q", name)
		}
	}
	yorkPlace := headerPlace(jamesHeader.Birth.Places, "York")
	if !sameSet(yorkPlace.Names, []string{"York", "Tkaronto"}) {
		t.Fatalf("York names %+v", yorkPlace)
	}
	if got := placeNames(jamesHeader.Death.Places); !sameSet(got, []string{"Toronto"}) {
		t.Fatalf("death places %v", got)
	}
	t.Run("life facts name the event and places they read", func(t *testing.T) {
		if e := jamesHeader.Birth.Event; e == nil || !bytes.Equal(e.ID, birthID) || e.Ref == "" || len(e.SubjectTypeID) != 16 {
			t.Fatalf("birth event %+v", e)
		}
		if e := jamesHeader.Death.Event; e == nil || !bytes.Equal(e.ID, deathID) {
			t.Fatalf("death event %+v", e)
		}
		if p := headerPlace(jamesHeader.Birth.Places, "York"); !bytes.Equal(p.Entity.ID, yorkID) || p.Entity.Ref == "" {
			t.Fatalf("York entity %+v", p.Entity)
		}
		if p := headerPlace(jamesHeader.Death.Places, "Toronto"); !bytes.Equal(p.Entity.ID, torontoID) {
			t.Fatalf("Toronto entity %+v", p.Entity)
		}
	})

	events := map[string]conclusionheaders.EventHeader{}
	for _, h := range f.events() {
		events[string(h.Entity.ID)] = h
	}
	birthHeader := events[string(birthID)]
	if forms := subjectForms(birthHeader); !sameSet(forms, []string{"James Robins"}) {
		t.Fatalf("birth subjects %v", forms)
	}
	if birthHeader.EventType == nil || birthHeader.EventType.Key != "birth" {
		t.Fatalf("birth type %+v", birthHeader.EventType)
	}
	if len(birthHeader.Places) != 2 {
		t.Fatalf("birth locations %d", len(birthHeader.Places))
	}
	marriageHeader := events[string(marriageID)]
	if forms := subjectForms(marriageHeader); !sameSet(forms, []string{"James Robins", "Mary Smith"}) {
		t.Fatalf("marriage subjects %v", forms)
	}
	if marriageHeader.EventType.Key != "marriage" {
		t.Fatalf("marriage type %s", marriageHeader.EventType.Key)
	}
	if forms := subjectForms(events[string(censusID)]); len(forms) != 3 {
		t.Fatalf("census subjects %v", forms)
	}
	if events[string(baptismID)].Subjects[0].Name != nil || !bytes.Equal(events[string(baptismID)].Subjects[0].Entity.ID, unnamedID) {
		t.Fatalf("unnamed %+v", events[string(baptismID)].Subjects)
	}
	fireHeader := events[string(fireID)]
	if len(fireHeader.Subjects) != 0 || fireHeader.EventType.Label != "Fire" {
		t.Fatalf("fire %+v subjects %d", fireHeader.EventType, len(fireHeader.Subjects))
	}
	if got := placeNames(fireHeader.Places); !sameSet(got, []string{"York", "Tkaronto"}) {
		t.Fatalf("fire places %v", got)
	}
	if len(fireHeader.Places) != 1 || !bytes.Equal(fireHeader.Places[0].Entity.ID, yorkID) {
		t.Fatalf("fire place entity %+v", fireHeader.Places)
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

	db, err := f.c.DB()
	must(t, err)
	deps, err := conclusionheaders.HeaderDependents(db, [][]byte{yorkID})
	must(t, err)
	if !hasID(deps, birthID) || !hasID(deps, fireID) || !hasID(deps, jamesID) || hasID(deps, deathID) {
		t.Fatalf("place dependents missing birth/fire/james or included the death")
	}
	// The witness is not a subject of the birth, so the place walk does not reach them.
	fromJames, err := conclusionheaders.HeaderDependents(db, [][]byte{jamesID})
	must(t, err)
	if !hasID(fromJames, birthID) || !hasID(fromJames, deathID) || !hasID(fromJames, marriageID) || !hasID(fromJames, censusID) {
		t.Fatal("person dependents missed an event")
	}
	if hasID(fromJames, yorkID) || hasID(fromJames, baptismID) {
		t.Fatal("person walk reached a place or someone else's event")
	}
	// James's header embeds his birth's date: a change to the birth must
	// reach him. The witness is not a subject, so they are not reached.
	fromBirth, err := conclusionheaders.HeaderDependents(db, [][]byte{birthID})
	must(t, err)
	if !hasID(fromBirth, jamesID) || hasID(fromBirth, birthID) || hasID(fromBirth, yorkID) {
		t.Fatalf("event dependents %d: missing James, or included the event or its place", len(fromBirth))
	}
}

func placeNames(places []conclusionheaders.HeaderPlace) []string {
	var out []string
	for _, p := range places {
		out = append(out, p.Names...)
	}
	return out
}

func headerPlace(places []conclusionheaders.HeaderPlace, name string) conclusionheaders.HeaderPlace {
	for _, p := range places {
		for _, n := range p.Names {
			if n == name {
				return p
			}
		}
	}
	return conclusionheaders.HeaderPlace{}
}

func subjectForms(h conclusionheaders.EventHeader) []string {
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

func hasID(ids [][]byte, id []byte) bool {
	for _, got := range ids {
		if bytes.Equal(got, id) {
			return true
		}
	}
	return false
}

// Two surviving births are a disagreement: the header counts them and leads
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
			james := f.bare("person")
			f.named(james, "James Robins")
			f.at(james, 0, 0)
			var births []subjects.Subject
			for i, year := range tt.years {
				birth := f.bare("event")
				in := []observations.Input{{PropertyID: f.prop("event_type").ID, ValueTermID: f.term("event_type", "birth").ID}}
				if year != 0 {
					in = append(in, observations.Input{PropertyID: f.prop("date").ID, Date: pointYear(year)})
				}
				f.cite(birth, in...)
				f.at(birth, 4, int64(2*i))
				f.participation(james, birth, "subject")
				births = append(births, birth)
			}
			jamesID := f.promoteSubject(james)
			for _, b := range births {
				f.promoteSubject(b)
			}
			var birth conclusionheaders.LifeFacts
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
