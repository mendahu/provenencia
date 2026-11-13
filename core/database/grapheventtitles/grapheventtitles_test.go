package grapheventtitles_test

import (
	"testing"

	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
	"github.com/mendahu/provenencia/core/database/subjects/subjectstest"
	"github.com/mendahu/provenencia/core/propertymeta"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts/artifactstest"
	"github.com/mendahu/provenencia/core/database/evidencelayer/evidencelayertest"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/grapheventtitles"
	"github.com/mendahu/provenencia/core/database/graphfill"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/eventtitle"
	"github.com/mendahu/provenencia/core/namevalue/namevaluetest"
	"github.com/mendahu/provenencia/core/writes"
)

var userID = evidencelayertest.UserID

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

type fixture struct {
	*evidencelayertest.Records
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	return &fixture{evidencelayertest.New(t)}
}

// subject is a Subject of the kind on the fixture's Source, labelled, with the
// Observations cited on its Artifact.
func (f *fixture) subject(kind, label string, in ...observations.Input) subjects.Subject {
	f.T.Helper()
	s := f.Labelled(f.Source, kind, label)
	if len(in) > 0 {
		f.Cite(f.Artifact, s, in...)
	}
	return s
}

func (f *fixture) eventType(key string) observations.Input {
	return observations.Input{PropertyID: f.Prop("event_type").ID, ValueTermID: f.Term("event_type", key).ID}
}

func TestForEvidenceGraph(t *testing.T) {
	f := newFixture(t)
	name := func(form string) observations.Input {
		return observations.Input{PropertyID: f.Prop("name").ID, Name: namevaluetest.Western(form)}
	}
	text := func(key, v string) observations.Input {
		return observations.Input{PropertyID: f.Prop(key).ID, Scalar: propertymeta.TextScalar(v)}
	}
	james := f.subject("person", "James", name("James Robins"))
	mary := f.subject("person", "Mary", name("Mary Smith"))
	nobody := f.subject("person", "the infant")
	witness := f.subject("person", "Webb", name("Witness Webb"))
	york := f.subject("place", "the town", text("toponym", "York"))

	baptism := f.subject("event", "Baptism", f.eventType("baptism"))
	f.Participation(f.Source, f.Artifact, james, baptism, "subject")
	f.Participation(f.Source, f.Artifact, witness, baptism, "witness")
	f.Locate(f.Source, f.Artifact, baptism, york)
	marriage := f.subject("event", "", f.eventType("marriage"))
	f.Participation(f.Source, f.Artifact, james, marriage, "subject")
	f.Participation(f.Source, f.Artifact, mary, marriage, "subject")
	birth := f.subject("event", "", f.eventType("birth"))
	f.Participation(f.Source, f.Artifact, nobody, birth, "subject")
	eventTypeID := f.Prop("event_type").ID
	fireTerm, _, err := writes.Run(f.C, writes.Op{Action: "create_property_term", UserID: userID},
		func(tx *database.Tx) (propertyterms.Term, []rowchange.Change, error) {
			return propertyterms.Create(tx, userID, eventTypeID, "Fire", "")
		})
	must(t, err)
	fire := f.subject("event", "", observations.Input{PropertyID: f.Prop("event_type").ID, ValueTermID: fireTerm.ID})
	f.Locate(f.Source, f.Artifact, fire, york)
	named := f.subject("event", "Grandpa's fire", text("event_name", "The Great Fire"))
	labelled := f.subject("event", "Grandpa's house fire")
	typed := f.subject("event", "", f.eventType("death"))

	titles := titlesFromStored(t, f)
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
			if got := subjectForms(plan); !sameSet(got, tt.subjects) {
				t.Fatalf("subjects %q, want %q", got, tt.subjects)
			}
		})
	}
	if p := byID[string(named.ID)].Parts; p.RecordedName != "The Great Fire" {
		t.Fatalf("recorded name %q", p.RecordedName)
	}
	if p := byID[string(fire.ID)].Parts; p.TypeKey != fireTerm.Key || p.TypeLabel != "Fire" || p.TypeOrigin != fireTerm.Origin {
		t.Fatalf("type %+v", p)
	}

	// Subject order is by participation ref, so et al. always names the same
	// person, read after read.
	t.Run("subjects are in participation-ref order, every time", func(t *testing.T) {
		db, err := f.C.DB()
		must(t, err)
		var refs []string
		rows, err := db.Query(`SELECT s.ref FROM subjects s
			JOIN observations o ON o.subject_id = s.id AND o.value_subject_id = ?
			JOIN subject_types st ON st.id = s.subject_type_id AND st.key = 'participation'
			ORDER BY s.ref`, marriage.ID)
		must(t, err)
		for rows.Next() {
			var r string
			must(t, rows.Scan(&r))
			refs = append(refs, r)
		}
		must(t, rows.Err())
		_ = rows.Close()
		var want []string
		for _, r := range refs {
			for _, pair := range []struct {
				person subjects.Subject
				form   string
			}{{james, "James Robins"}, {mary, "Mary Smith"}} {
				var n int
				must(t, db.QueryRow(`SELECT COUNT(*) FROM observations o JOIN subjects b ON b.id = o.subject_id
					WHERE b.ref = ? AND o.value_subject_id = ?`, r, pair.person.ID).Scan(&n))
				if n > 0 {
					want = append(want, pair.form)
				}
			}
		}
		for i := 0; i < 3; i++ {
			again := titlesFromStored(t, f)
			for _, title := range again {
				if string(title.SubjectID) != string(marriage.ID) {
					continue
				}
				if got := subjectForms(title.Plan); len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
					t.Fatalf("read %d: %q, want %q", i, got, want)
				}
			}
		}
	})

	t.Run("one Event on each graph titles only its own graph's", func(t *testing.T) {
		f := newFixture(t)
		f.subject("event", "Default event")
		_, jones, _ := artifactstest.Split(t, f.C, userID, f.Source.ID, "Jones")
		jonesEvent := subjectstest.On(t, f.C, userID, jones.ID, "event", "Jones event")
		eg, err := graphfill.EvidenceGraph(f.C.Graph(), jones.ID)
		must(t, err)
		got := grapheventtitles.FromStored(eg, nil, nil)
		if len(got) != 1 || string(got[0].SubjectID) != string(jonesEvent.ID) {
			t.Fatalf("titles %+v, want one for %x", got, jonesEvent.ID)
		}
	})

	t.Run("an empty Source has no titles", func(t *testing.T) {
		got := grapheventtitles.FromStored(graphcache.EvidenceGraph{}, nil, nil)
		if len(got) != 0 {
			t.Fatalf("titled %d", len(got))
		}
	})
}

func titlesFromStored(t *testing.T, f *fixture) []grapheventtitles.Title {
	t.Helper()
	eg, err := graphfill.EvidenceGraph(f.C.Graph(), evidencegraphstest.SoleGraph(t, f.C, f.Source.ID).ID)
	must(t, err)
	v, err := f.C.Graph().Vocabulary()
	must(t, err)
	propKey := map[string]string{}
	for _, p := range v.Properties {
		propKey[string(p.ID)] = p.Key
	}
	terms := map[string]struct{ Key, Origin, Label string }{}
	for _, term := range v.Terms {
		terms[string(term.ID)] = struct{ Key, Origin, Label string }{Key: term.Key, Origin: term.Origin, Label: term.Label}
	}
	return grapheventtitles.FromStored(eg, propKey, terms)
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
		if seen[w] == 0 {
			return false
		}
		seen[w]--
	}
	return true
}
