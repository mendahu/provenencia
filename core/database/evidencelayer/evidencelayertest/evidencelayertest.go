// Package evidencelayertest builds catalogs whose Sources hold people, events,
// places and the bridges between them, for tests of the Evidence layer and the
// packages that read it.
package evidencelayertest

import (
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/artifacts/artifactstest"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/connect"
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
	"github.com/mendahu/provenencia/core/database/evidencelayer"
	_ "github.com/mendahu/provenencia/core/database/graphfill"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjectpositions"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjects/subjectstest"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/database/valuerules"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/datevalue/datevaluetest"
	"github.com/mendahu/provenencia/core/namevalue/namevaluetest"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
)

// UserID is the user every Records catalog is created for.
var UserID = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

// Locator is the citation locator every Records citation uses.
const Locator = `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`

// Records is a catalog with one Source and its Artifact, and the builders for
// the Subjects, Observations and bridges cited on any Source in it.
type Records struct {
	T        testing.TB
	C        *database.Catalog
	Source   sources.Source
	Artifact artifacts.Artifact
	// Reported is the SummaryDeps of the last write a Records helper made.
	Reported [][]byte
	y        int64
}

// run is writes.Run with the action, recording the write's SummaryDeps in
// r.Reported.
func run[T any](r *Records, action string, fn func(tx *database.Tx) (T, []rowchange.Change, error)) T {
	r.T.Helper()
	v, res, err := writes.Run(r.C, writes.Op{Action: action, UserID: UserID}, fn)
	Must(r.T, err)
	r.Reported = res.Effects.SummaryDeps
	return v
}

// Run makes one write of the action through writes.Run, as the app does, and
// returns the handles it reported as moved (Result.Effects.SummaryDeps).
func (r *Records) Run(action string, fn func(tx *database.Tx) ([]rowchange.Change, error)) [][]byte {
	r.T.Helper()
	run(r, action, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
		ch, err := fn(tx)
		return struct{}{}, ch, err
	})
	return r.Reported
}

// Must fails the test on err.
func Must(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

// New is a fresh catalog with the user, the seeded vocabulary, and one Source
// ("Register") with one Artifact ("Scan").
func New(t testing.TB) *Records {
	t.Helper()
	c, err := database.Create(t.TempDir(), "t.provenencia")
	Must(t, err)
	t.Cleanup(func() { _ = c.Close() })
	Must(t, subjectvocab.Install(c))
	return Open(t, c)
}

// Open is Records on a catalog that already exists, such as a project's; the
// catalog must hold the seeded vocabulary. It upserts UserID beside any user
// the catalog has, creates the Source ("Register") and its Artifact ("Scan"),
// and registers no close: the caller owns c.
func Open(t testing.TB, c *database.Catalog) *Records {
	t.Helper()
	r, err := ref.Mint(ref.PrefixUser)
	Must(t, err)
	Must(t, users.Upsert(c, UserID, "Tester", r))
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book"})
	Must(t, err)
	src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: UserID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, UserID, sources.CreateInput{SourceTypeID: typeID, Title: "Register"})
	})
	Must(t, err)
	art, err := artifactstest.Create(c, UserID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
	Must(t, err)
	return &Records{T: t, C: c, Source: src, Artifact: art}
}

// NewSource is another Source in the catalog, with one Artifact.
func (r *Records) NewSource(title string) (sources.Source, artifacts.Artifact) {
	r.T.Helper()
	typeID, err := sourcetypes.Upsert(r.C, sourcetypes.Type{Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book"})
	Must(r.T, err)
	src, _, err := writes.Run(r.C, writes.Op{Action: "create_source", UserID: UserID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, UserID, sources.CreateInput{SourceTypeID: typeID, Title: title})
	})
	Must(r.T, err)
	art, err := artifactstest.Create(r.C, UserID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
	Must(r.T, err)
	return src, art
}

// Prop is the product Property with the key.
func (r *Records) Prop(key string) properties.Property {
	r.T.Helper()
	p, err := properties.Lookup(r.C, key, vocabkey.OriginProvenencia)
	Must(r.T, err)
	return p
}

// Term is the product term of the Property.
func (r *Records) Term(propKey, termKey string) propertyterms.Term {
	r.T.Helper()
	t, err := propertyterms.Lookup(r.C, r.Prop(propKey).ID, termKey, vocabkey.OriginProvenencia)
	Must(r.T, err)
	return t
}

// Bare is a Subject of the kind on src's Evidence graph, with no Observation,
// placed below the last one.
func (r *Records) Bare(src sources.Source, kind string) subjects.Subject {
	r.T.Helper()
	return r.Labelled(src, kind, "")
}

// Labelled is Bare with a label.
func (r *Records) Labelled(src sources.Source, kind, label string) subjects.Subject {
	r.T.Helper()
	s := subjectstest.On(r.T, r.C, UserID, r.Graph(src), kind, label)
	_, err := writes.Call(r.C, writes.Op{}, func(tx *database.Tx) (subjectpositions.Position, []rowchange.Change, error) {
		return subjectpositions.Set(tx, s.ID, 0, r.y)
	})
	Must(r.T, err)
	r.y += 2
	return s
}

// Cite records the Observations on the Subject, cited on art.
func (r *Records) Cite(art artifacts.Artifact, s subjects.Subject, in ...observations.Input) {
	r.T.Helper()
	for i := range in {
		in[i].SubjectID = s.ID
	}
	run(r, "create_citation_with_observations", func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
		return citations.CreateWithObservations(tx, UserID, citations.CreateInput{ArtifactID: art.ID, GraphID: s.GraphID, LocatorJSON: Locator}, in)
	})
}

// Person is a person named name.
func (r *Records) Person(src sources.Source, art artifacts.Artifact, name string) subjects.Subject {
	r.T.Helper()
	s := r.Bare(src, "person")
	r.Cite(art, s, observations.Input{
		PropertyID: r.Prop("name").ID, Name: namevaluetest.Western(name),
	})
	return s
}

// Birth is an event cited as a birth, with no date.
func (r *Records) Birth(src sources.Source, art artifacts.Artifact) subjects.Subject {
	r.T.Helper()
	s := r.Bare(src, "event")
	r.Cite(art, s, observations.Input{
		PropertyID: r.Prop("event_type").ID, ValueTermID: r.Term("event_type", "birth").ID,
	})
	return s
}

// Place is a place with the toponym name.
func (r *Records) Place(src sources.Source, art artifacts.Artifact, name string) subjects.Subject {
	r.T.Helper()
	s := r.Bare(src, "place")
	r.Cite(art, s, observations.Input{
		PropertyID: r.Prop("toponym").ID, Scalar: propertymeta.TextScalar(name),
	})
	return s
}

// Participation cites person as taking part in event, in role, and returns the
// bridge Subject.
func (r *Records) Participation(src sources.Source, art artifacts.Artifact, person, event subjects.Subject, role string) subjects.Subject {
	r.T.Helper()
	personProp := r.Prop("person").ID
	eventProp := r.Prop("event").ID
	roleProp := r.Prop("role").ID
	roleTerm := r.Term("role", role).ID
	res := run(r, "create_cited_bridge", func(tx *database.Tx) (connect.Result, []rowchange.Change, error) {
		return connect.CreateCitedBridge(tx, UserID, connect.CreateInput{
			FromSubjectID: person.ID, ToSubjectID: event.ID, BridgeTypeKey: "participation",
			Citation: citations.CreateInput{ArtifactID: art.ID, LocatorJSON: Locator},
			Observations: []observations.Input{
				{PropertyID: personProp, ValueSubjectID: person.ID},
				{PropertyID: eventProp, ValueSubjectID: event.ID},
				{PropertyID: roleProp, ValueTermID: roleTerm},
			},
		})
	})
	return res.Subject
}

// Locate cites event as taking place at place, and returns the bridge Subject.
func (r *Records) Locate(src sources.Source, art artifacts.Artifact, event, place subjects.Subject) subjects.Subject {
	r.T.Helper()
	eventProp := r.Prop("event").ID
	placeProp := r.Prop("place").ID
	res := run(r, "create_cited_bridge", func(tx *database.Tx) (connect.Result, []rowchange.Change, error) {
		return connect.CreateCitedBridge(tx, UserID, connect.CreateInput{
			FromSubjectID: event.ID, ToSubjectID: place.ID, BridgeTypeKey: "location",
			Citation: citations.CreateInput{ArtifactID: art.ID, LocatorJSON: Locator},
			Observations: []observations.Input{
				{PropertyID: eventProp, ValueSubjectID: event.ID},
				{PropertyID: placeProp, ValueSubjectID: place.ID},
			},
		})
	})
	return res.Subject
}

// PlaceRel cites from as a kind of place relationship to to.
func (r *Records) PlaceRel(src sources.Source, art artifacts.Artifact, from, to subjects.Subject, kind string) subjects.Subject {
	r.T.Helper()
	fromProp := r.Prop("from").ID
	toProp := r.Prop("to").ID
	typeProp := r.Prop("place_relationship_type").ID
	kindTerm := r.Term("place_relationship_type", kind).ID
	res := run(r, "create_cited_bridge", func(tx *database.Tx) (connect.Result, []rowchange.Change, error) {
		return connect.CreateCitedBridge(tx, UserID, connect.CreateInput{
			FromSubjectID: from.ID, ToSubjectID: to.ID, BridgeTypeKey: "place_relationship",
			Citation: citations.CreateInput{ArtifactID: art.ID, LocatorJSON: Locator},
			Observations: []observations.Input{
				{PropertyID: fromProp, ValueSubjectID: from.ID},
				{PropertyID: toProp, ValueSubjectID: to.ID},
				{PropertyID: typeProp, ValueTermID: kindTerm},
			},
		})
	})
	return res.Subject
}

// Relationship cites "person is typ of related" on src.
func (r *Records) Relationship(src sources.Source, art artifacts.Artifact, person, related subjects.Subject, typ string) subjects.Subject {
	r.T.Helper()
	personProp := r.Prop("person").ID
	relatedProp := r.Prop("related_to").ID
	typeProp := r.Prop("relationship_type").ID
	typTerm := r.Term("relationship_type", typ).ID
	res := run(r, "create_cited_bridge", func(tx *database.Tx) (connect.Result, []rowchange.Change, error) {
		return connect.CreateCitedBridge(tx, UserID, connect.CreateInput{
			FromSubjectID: person.ID, ToSubjectID: related.ID, BridgeTypeKey: "relationship",
			Citation: citations.CreateInput{ArtifactID: art.ID, LocatorJSON: Locator},
			Observations: []observations.Input{
				{PropertyID: personProp, ValueSubjectID: person.ID},
				{PropertyID: relatedProp, ValueSubjectID: related.ID},
				{PropertyID: typeProp, ValueTermID: typTerm},
			},
		})
	})
	return res.Subject
}

// SexOnly is a person whose record gives no name, only a sex at birth.
func (r *Records) SexOnly(src sources.Source, art artifacts.Artifact, sex string) subjects.Subject {
	r.T.Helper()
	s := r.Bare(src, "person")
	r.Cite(art, s, observations.Input{
		PropertyID: r.Prop("sex_at_birth").ID, ValueTermID: r.Term("sex_at_birth", sex).ID,
	})
	return s
}

// Graph is the one Evidence graph of src. Call it outside a write closure.
func (r *Records) Graph(src sources.Source) []byte {
	r.T.Helper()
	return evidencegraphstest.SoleGraph(r.T, r.C, src.ID).ID
}

// Layer is the Evidence layer of the graph, loaded under the catalog's name
// pattern: the memoized one, or one built.
func (r *Records) Layer(graph []byte) evidencelayer.Layer {
	r.T.Helper()
	_, namePattern, err := valuerules.Load(r.C.Graph())
	Must(r.T, err)
	layer, err := evidencelayer.Load(r.C.Graph(), graph, namePattern)
	Must(r.T, err)
	return layer
}

// DatedOn is an event cited as an event type on a date (m and d 0 for a year).
func (r *Records) DatedOn(src sources.Source, art artifacts.Artifact, eventType string, y, m, d int) subjects.Subject {
	r.T.Helper()
	return r.EventOn(src, art, eventType, datevaluetest.YMD(y, m, d))
}

// Marriage is a marriage event cited in src with a and b as its participants,
// each in role.
func (r *Records) Marriage(src sources.Source, art artifacts.Artifact, a, b subjects.Subject, role string) subjects.Subject {
	m := r.EventOn(src, art, "marriage", nil)
	r.Participation(src, art, a, m, role)
	r.Participation(src, art, b, m, role)
	return m
}

// EventOn is an event cited as an event type, on a date when d is not nil.
func (r *Records) EventOn(src sources.Source, art artifacts.Artifact, eventType string, d *datevalue.Value) subjects.Subject {
	r.T.Helper()
	s := r.Bare(src, "event")
	in := []observations.Input{{PropertyID: r.Prop("event_type").ID, ValueTermID: r.Term("event_type", eventType).ID}}
	if d != nil {
		in = append(in, observations.Input{PropertyID: r.Prop("date").ID, Date: d})
	}
	r.Cite(art, s, in...)
	return s
}

// DatedBirth is an event cited as a birth in a year.
func (r *Records) DatedBirth(src sources.Source, art artifacts.Artifact, y int) subjects.Subject {
	r.T.Helper()
	return r.DatedOn(src, art, "birth", y, 0, 0)
}
