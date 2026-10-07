package sourceeventtitles_test

import (
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/connect"
	"github.com/mendahu/provenencia/core/database/namevalues/namevaluestest"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/sourceeventtitles"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjectpositions"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/eventtitle"
	"github.com/mendahu/provenencia/core/ref"
)

var userID = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

const locator = `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

type fixture struct {
	t   *testing.T
	c   *database.Catalog
	src sources.Source
	art artifacts.Artifact
	y   int64
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	c, err := database.Create(t.TempDir(), "t.provenencia")
	must(t, err)
	t.Cleanup(func() { _ = c.Close() })
	r, err := ref.Mint(ref.PrefixUser)
	must(t, err)
	must(t, users.Upsert(c, userID, "Tester", r))
	must(t, subjectvocab.Install(c))
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{Key: "book", Origin: sourcetypes.OriginProvenencia, Label: "Book"})
	must(t, err)
	src, err := sources.Create(c, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Register"})
	must(t, err)
	art, err := artifacts.Create(c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
	must(t, err)
	return &fixture{t: t, c: c, src: src, art: art}
}

func (f *fixture) subject(kind, label string, in ...observations.Input) subjects.Subject {
	f.t.Helper()
	st, err := subjecttypes.Lookup(f.c, kind, subjecttypes.OriginProvenencia)
	must(f.t, err)
	s, err := subjects.Create(f.c, userID, subjects.CreateInput{SourceID: f.src.ID, SubjectTypeID: st.ID, Label: label}, nil)
	must(f.t, err)
	_, err = subjectpositions.Set(f.c, s.ID, 0, f.y)
	must(f.t, err)
	f.y += 2
	if len(in) > 0 {
		for i := range in {
			in[i].SubjectID = s.ID
		}
		_, err = citations.CreateWithObservations(f.c, userID, citations.CreateInput{ArtifactID: f.art.ID, LocatorJSON: locator}, in)
		must(f.t, err)
	}
	return s
}

func (f *fixture) prop(key string) []byte {
	f.t.Helper()
	p, err := properties.Lookup(f.c, key, properties.OriginProvenencia)
	must(f.t, err)
	return p.ID
}

func (f *fixture) term(propertyKey, termKey string) []byte {
	f.t.Helper()
	term, err := propertyterms.Lookup(f.c, f.prop(propertyKey), termKey, propertyterms.OriginProvenencia)
	must(f.t, err)
	return term.ID
}

func (f *fixture) eventType(key string) observations.Input {
	return observations.Input{PropertyID: f.prop("event_type"), ValueTermID: f.term("event_type", key)}
}

func (f *fixture) bridge(kind string, from, to subjects.Subject, obs ...observations.Input) {
	f.t.Helper()
	_, err := connect.CreateCitedBridge(f.c, userID, connect.CreateInput{
		SourceID: f.src.ID, FromSubjectID: from.ID, ToSubjectID: to.ID, BridgeTypeKey: kind,
		Citation:     citations.CreateInput{ArtifactID: f.art.ID, LocatorJSON: locator},
		Observations: obs,
	})
	must(f.t, err)
}

func (f *fixture) participation(person, event subjects.Subject, role string) {
	f.t.Helper()
	f.bridge("participation", person, event,
		observations.Input{PropertyID: f.prop("person"), ValueSubjectID: person.ID},
		observations.Input{PropertyID: f.prop("event"), ValueSubjectID: event.ID},
		observations.Input{PropertyID: f.prop("role"), ValueTermID: f.term("role", role)},
	)
}

func (f *fixture) location(event, place subjects.Subject) {
	f.t.Helper()
	f.bridge("location", event, place,
		observations.Input{PropertyID: f.prop("event"), ValueSubjectID: event.ID},
		observations.Input{PropertyID: f.prop("place"), ValueSubjectID: place.ID},
	)
}

func TestForSource(t *testing.T) {
	f := newFixture(t)
	name := func(form string) observations.Input {
		return observations.Input{PropertyID: f.prop("name"), Name: namevaluestest.Western(form)}
	}
	text := func(key, v string) observations.Input {
		return observations.Input{PropertyID: f.prop(key), ValueText: v, HasText: true}
	}
	james := f.subject("person", "James", name("James Robins"))
	mary := f.subject("person", "Mary", name("Mary Smith"))
	nobody := f.subject("person", "the infant")
	witness := f.subject("person", "Webb", name("Witness Webb"))
	york := f.subject("place", "the town", text("toponym", "York"))

	baptism := f.subject("event", "Baptism", f.eventType("baptism"))
	f.participation(james, baptism, "subject")
	f.participation(witness, baptism, "witness")
	f.location(baptism, york)
	marriage := f.subject("event", "", f.eventType("marriage"))
	f.participation(james, marriage, "subject")
	f.participation(mary, marriage, "subject")
	birth := f.subject("event", "", f.eventType("birth"))
	f.participation(nobody, birth, "subject")
	fireTerm, err := propertyterms.Create(f.c, userID, f.prop("event_type"), "Fire", "")
	must(t, err)
	fire := f.subject("event", "", observations.Input{PropertyID: f.prop("event_type"), ValueTermID: fireTerm.ID})
	f.location(fire, york)
	named := f.subject("event", "Grandpa's fire", text("event_name", "The Great Fire"))
	labelled := f.subject("event", "Grandpa's house fire")
	typed := f.subject("event", "", f.eventType("death"))

	db, err := f.c.DB()
	must(t, err)
	titles, err := sourceeventtitles.ForSource(db, f.src.ID)
	must(t, err)
	byID := map[string]eventtitle.Plan{}
	for _, title := range titles {
		byID[string(title.SubjectID)] = title.Plan
	}
	if len(titles) != 7 {
		t.Fatalf("titled %d events, want 7 (people and places are not titled)", len(titles))
	}

	subjectForms := func(p eventtitle.Plan) []string {
		var out []string
		for _, s := range p.Parts.Subjects {
			if s.Name == nil {
				out = append(out, "")
				continue
			}
			out = append(out, s.Name.Form)
		}
		return out
	}
	tests := []struct {
		name     string
		event    subjects.Subject
		rule     eventtitle.Rule
		subjects []string
		place    string
	}{
		{name: "the subject wins over the label; a witness is not a subject", event: baptism, rule: eventtitle.RuleSubject, subjects: []string{"James Robins"}, place: "York"},
		{name: "two subjects of a marriage", event: marriage, rule: eventtitle.RuleCouple, subjects: []string{"James Robins", "Mary Smith"}},
		{name: "an unnamed subject stays unnamed", event: birth, rule: eventtitle.RuleSubject, subjects: []string{""}},
		{name: "type at place", event: fire, rule: eventtitle.RuleTypeAtPlace, place: "York"},
		{name: "a recorded name wins", event: named, rule: eventtitle.RuleRecordedName},
		{name: "the label", event: labelled, rule: eventtitle.RuleLabel},
		{name: "type only", event: typed, rule: eventtitle.RuleType},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan, ok := byID[string(tt.event.ID)]
			if !ok {
				t.Fatal("not titled")
			}
			if plan.Rule != tt.rule || plan.Parts.Place != tt.place {
				t.Fatalf("rule %s place %q", plan.Rule, plan.Parts.Place)
			}
			if got := subjectForms(plan); len(got) != len(tt.subjects) || (len(got) > 0 && got[0] != tt.subjects[0]) {
				t.Fatalf("subjects %q, want %q", got, tt.subjects)
			}
		})
	}
	if p := byID[string(named.ID)].Parts; p.RecordedName != "The Great Fire" {
		t.Fatalf("recorded name %q", p.RecordedName)
	}
	if p := byID[string(fire.ID)].Parts; p.TypeKey != fireTerm.Key || p.TypeLabel != "Fire" {
		t.Fatalf("type %+v", p)
	}

	t.Run("an empty Source has no titles", func(t *testing.T) {
		got, err := sourceeventtitles.ForSource(db, make([]byte, 16))
		must(t, err)
		if len(got) != 0 {
			t.Fatalf("titled %d", len(got))
		}
	})
}
