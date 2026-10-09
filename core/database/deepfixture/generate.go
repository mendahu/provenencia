// Package deepfixture builds a temporary catalog large enough to time the
// conclusion reads and Promote writes the Spike 9 ledger records.
//
// One Person is a member of about ten Sources. Birth, death, and marriage
// each have more than one member and a Location. A few relationships and a
// place chain several levels deep sit beside them. Filler handles bring the
// catalog to a few hundred.
package deepfixture

import (
	"bytes"
	"fmt"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/connect"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/namevalues/namevaluestest"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjectpositions"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
)

// FillerPeople is how many extra Persons (each with a birth) pad the catalog.
const FillerPeople = 150

// UserID is the researcher every write in the fixture is attributed to.
var UserID = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

const locator = `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`

// Catalog is an open deep fixture and the handles the timings read.
type Catalog struct {
	Cat *database.Catalog
	// Person is the handle named on about ten Sources.
	Person []byte
	// NameObservation is one of that Person's name records, for a one-write upkeep.
	NameObservation observations.Observation
	// Place is the deepest Place in the chain.
	Place []byte
	// DeepSource is a Source whose Subjects are already filed.
	DeepSource []byte
	// ObitSource is a small unpromoted layer (obituary scale).
	ObitSource []byte
	// ObitSubjects are the primary Subjects on ObitSource, still unfiled.
	ObitSubjects [][]byte
}

// Unpromoted adds a Source of people and events that are not yet filed.
// Done benchmarks call it outside the timed section.
func (out *Catalog) Unpromoted(title string, people, events int) (sourceID []byte, subjectIDs [][]byte, err error) {
	b := &bld{c: out, prop: map[string]properties.Property{}, term: map[string]propertyterms.Term{}, typ: map[string][]byte{}}
	typeID, err := sourcetypes.Upsert(out.Cat, sourcetypes.Type{Key: "book", Origin: sourcetypes.OriginProvenencia, Label: "Book"})
	if err != nil {
		return nil, nil, err
	}
	s, err := b.openSource(typeID, title)
	if err != nil {
		return nil, nil, err
	}
	for i := 0; i < people; i++ {
		sub, err := b.bare(s, "person")
		if err != nil {
			return nil, nil, err
		}
		if err := b.cite(s, sub, observations.Input{
			PropertyID: b.mustProp("name").ID,
			Name:       namevaluestest.Western(fmt.Sprintf("%s %d", title, i+1)),
		}); err != nil {
			return nil, nil, err
		}
		subjectIDs = append(subjectIDs, append([]byte(nil), sub.ID...))
	}
	for i := 0; i < events; i++ {
		ev, err := b.bare(s, "event")
		if err != nil {
			return nil, nil, err
		}
		if err := b.cite(s, ev, observations.Input{
			PropertyID:  b.mustProp("event_type").ID,
			ValueTermID: b.mustTerm("event_type", "residence").ID,
		}); err != nil {
			return nil, nil, err
		}
		subjectIDs = append(subjectIDs, append([]byte(nil), ev.ID...))
	}
	return append([]byte(nil), s.source.ID...), subjectIDs, nil
}

// Generate creates the catalog in dir and returns it open. The caller closes it.
func Generate(dir string) (*Catalog, error) {
	c, err := database.Create(dir, "deep.provenencia")
	if err != nil {
		return nil, err
	}
	out := &Catalog{Cat: c}
	if err := out.build(); err != nil {
		_ = c.Close()
		return nil, err
	}
	return out, nil
}

type layer struct {
	source   sources.Source
	artifact artifacts.Artifact
}

type bld struct {
	c    *Catalog
	prop map[string]properties.Property
	term map[string]propertyterms.Term
	typ  map[string][]byte
	slot int64
}

func (out *Catalog) build() error {
	r, err := ref.Mint(ref.PrefixUser)
	if err != nil {
		return err
	}
	if err := users.Upsert(out.Cat, UserID, "Tester", r); err != nil {
		return err
	}
	if err := subjectvocab.Install(out.Cat); err != nil {
		return err
	}
	b := &bld{
		c:    out,
		prop: map[string]properties.Property{},
		term: map[string]propertyterms.Term{},
		typ:  map[string][]byte{},
	}
	typeID, err := sourcetypes.Upsert(out.Cat, sourcetypes.Type{Key: "book", Origin: sourcetypes.OriginProvenencia, Label: "Book"})
	if err != nil {
		return err
	}
	regs := make([]layer, 0, 10)
	for i := 0; i < 10; i++ {
		s, err := b.openSource(typeID, fmt.Sprintf("Register %d", i+1))
		if err != nil {
			return err
		}
		regs = append(regs, s)
	}
	out.DeepSource = append([]byte(nil), regs[0].source.ID...)

	places, err := b.placeChain(regs[0])
	if err != nil {
		return fmt.Errorf("places: %w", err)
	}
	out.Place = places[len(places)-1].id

	person, nameObs, err := b.focalPerson(regs)
	if err != nil {
		return fmt.Errorf("focal: %w", err)
	}
	out.Person = person
	out.NameObservation = nameObs

	spouseSub, spouseID, err := b.namedPerson(regs[0], "Mary Smith")
	if err != nil {
		return err
	}
	town := places[len(places)-2]
	if err := b.lifeEvents(regs, person, spouseID, town); err != nil {
		return fmt.Errorf("life: %w", err)
	}
	if err := b.relationship(regs[0], person, spouseSub, "spouse"); err != nil {
		return err
	}
	childSub, _, err := b.namedPerson(regs[1], "Ann Robins")
	if err != nil {
		return err
	}
	if err := b.relationship(regs[1], person, childSub, "parent"); err != nil {
		return err
	}
	for i := 0; i < FillerPeople; i++ {
		s := regs[i%len(regs)]
		_, who, err := b.namedPerson(s, fmt.Sprintf("Filler %d Person", i+1))
		if err != nil {
			return err
		}
		if err := b.simpleEvent(s, "birth", who); err != nil {
			return err
		}
	}
	for _, s := range regs {
		if err := out.fileBridges(s.source.ID); err != nil {
			return fmt.Errorf("file bridges: %w", err)
		}
	}
	return b.obituary(typeID)
}

// fileBridges files Evidence bridges whose ends are already handles.
func (out *Catalog) fileBridges(sourceID []byte) error {
	db, err := out.Cat.DB()
	if err != nil {
		return err
	}
	var rev int64
	if err := db.QueryRow(`SELECT COALESCE(MAX(revision), 0) FROM audit_transactions`).Scan(&rev); err != nil {
		return err
	}
	_, err = promote.SaveBatch(out.Cat, UserID, promote.Batch{SourceID: sourceID, SeenRevision: rev})
	return err
}

func (b *bld) openSource(typeID []byte, title string) (layer, error) {
	s, err := sources.Create(b.c.Cat, UserID, sources.CreateInput{SourceTypeID: typeID, Title: title})
	if err != nil {
		return layer{}, err
	}
	a, err := artifacts.Create(b.c.Cat, UserID, artifacts.CreateInput{SourceID: s.ID, Label: title})
	if err != nil {
		return layer{}, err
	}
	return layer{source: s, artifact: a}, nil
}

type filed struct {
	sub subjects.Subject
	id  []byte
}

func (b *bld) placeChain(s layer) ([]filed, error) {
	names := []string{"England", "Yorkshire", "York", "Parish", "Street", "House"}
	out := make([]filed, len(names))
	for i, name := range names {
		sub, err := b.bare(s, "place")
		if err != nil {
			return nil, err
		}
		if err := b.cite(s, sub, observations.Input{PropertyID: b.mustProp("toponym").ID, ValueText: name, HasText: true}); err != nil {
			return nil, err
		}
		res, err := promote.Save(b.c.Cat, UserID, promote.Input{SubjectID: sub.ID})
		if err != nil {
			return nil, err
		}
		out[i] = filed{sub: sub, id: res.Entity.ID}
		if i == 0 {
			continue
		}
		if err := b.placeRel(s, sub, out[i-1].sub, "part_of"); err != nil {
			return nil, fmt.Errorf("part_of %s: %w", name, err)
		}
	}
	return out, nil
}

func (b *bld) focalPerson(regs []layer) ([]byte, observations.Observation, error) {
	var entity []byte
	var nameObs observations.Observation
	for i, s := range regs {
		sub, err := b.bare(s, "person")
		if err != nil {
			return nil, observations.Observation{}, err
		}
		obs, err := b.citeObs(s, sub, observations.Input{
			PropertyID: b.mustProp("name").ID,
			Name:       namevaluestest.Western("James Kenneth Robins"),
		})
		if err != nil {
			return nil, observations.Observation{}, err
		}
		res, err := promote.Save(b.c.Cat, UserID, promote.Input{SubjectID: sub.ID, EntityID: entity})
		if err != nil {
			return nil, observations.Observation{}, err
		}
		entity = res.Entity.ID
		if i == 0 {
			nameObs = obs
		}
	}
	return entity, nameObs, nil
}

func (b *bld) namedPerson(s layer, form string) (subjects.Subject, []byte, error) {
	sub, err := b.bare(s, "person")
	if err != nil {
		return subjects.Subject{}, nil, err
	}
	if err := b.cite(s, sub, observations.Input{PropertyID: b.mustProp("name").ID, Name: namevaluestest.Western(form)}); err != nil {
		return subjects.Subject{}, nil, err
	}
	res, err := promote.Save(b.c.Cat, UserID, promote.Input{SubjectID: sub.ID})
	if err != nil {
		return subjects.Subject{}, nil, err
	}
	return sub, res.Entity.ID, nil
}

func (b *bld) lifeEvents(regs []layer, person, spouse []byte, town filed) error {
	for _, kind := range []string{"birth", "death", "marriage"} {
		first, err := b.eventSubject(regs[0], kind, town.sub)
		if err != nil {
			return err
		}
		if err := b.participation(regs[0], first, person, "subject"); err != nil {
			return fmt.Errorf("participation %s: %w", kind, err)
		}
		res, err := promote.Save(b.c.Cat, UserID, promote.Input{SubjectID: first.ID})
		if err != nil {
			return err
		}
		second, err := b.eventSubject(regs[1], kind, town.sub)
		if err != nil {
			return err
		}
		other := person
		if kind == "marriage" {
			other = spouse
		}
		if err := b.participation(regs[1], second, other, "subject"); err != nil {
			return err
		}
		if _, err := promote.Save(b.c.Cat, UserID, promote.Input{SubjectID: second.ID, EntityID: res.Entity.ID}); err != nil {
			return err
		}
	}
	return nil
}

func (b *bld) simpleEvent(s layer, kind string, person []byte) error {
	ev, err := b.eventSubject(s, kind, subjects.Subject{})
	if err != nil {
		return err
	}
	if err := b.participation(s, ev, person, "subject"); err != nil {
		return err
	}
	_, err = promote.Save(b.c.Cat, UserID, promote.Input{SubjectID: ev.ID})
	return err
}

func (b *bld) eventSubject(s layer, kind string, place subjects.Subject) (subjects.Subject, error) {
	ev, err := b.bare(s, "event")
	if err != nil {
		return subjects.Subject{}, err
	}
	y := 1817
	if err := b.cite(s, ev,
		observations.Input{PropertyID: b.mustProp("event_type").ID, ValueTermID: b.mustTerm("event_type", kind).ID},
		observations.Input{PropertyID: b.mustProp("date").ID, Date: &datevalues.Value{Kind: datevalues.KindPoint, Calendar: "gregorian", StartYear: &y}},
	); err != nil {
		return subjects.Subject{}, err
	}
	if len(place.ID) != 0 && bytes.Equal(place.SourceID, s.source.ID) {
		if err := b.location(s, ev, place); err != nil {
			return subjects.Subject{}, fmt.Errorf("location: %w", err)
		}
	}
	return ev, nil
}

func (b *bld) obituary(typeID []byte) error {
	s, err := b.openSource(typeID, "Obituary")
	if err != nil {
		return err
	}
	b.c.ObitSource = append([]byte(nil), s.source.ID...)
	for i := 0; i < 6; i++ {
		sub, err := b.bare(s, "person")
		if err != nil {
			return err
		}
		if err := b.cite(s, sub, observations.Input{
			PropertyID: b.mustProp("name").ID,
			Name:       namevaluestest.Western(fmt.Sprintf("Obit %d Child", i+1)),
		}); err != nil {
			return err
		}
		b.c.ObitSubjects = append(b.c.ObitSubjects, append([]byte(nil), sub.ID...))
	}
	for i := 0; i < 4; i++ {
		ev, err := b.bare(s, "event")
		if err != nil {
			return err
		}
		if err := b.cite(s, ev, observations.Input{
			PropertyID:  b.mustProp("event_type").ID,
			ValueTermID: b.mustTerm("event_type", "residence").ID,
		}); err != nil {
			return err
		}
		b.c.ObitSubjects = append(b.c.ObitSubjects, append([]byte(nil), ev.ID...))
	}
	return nil
}

func (b *bld) bare(s layer, kind string) (subjects.Subject, error) {
	sub, err := subjects.Create(b.c.Cat, UserID, subjects.CreateInput{
		SourceID: s.source.ID, SubjectTypeID: b.typeID(kind),
	}, nil)
	if err != nil {
		return subjects.Subject{}, err
	}
	b.slot++
	if _, err := subjectpositions.Set(b.c.Cat, sub.ID, b.slot, 0); err != nil {
		return subjects.Subject{}, err
	}
	return sub, nil
}

func (b *bld) cite(s layer, sub subjects.Subject, in ...observations.Input) error {
	_, err := b.citeObs(s, sub, in...)
	return err
}

func (b *bld) citeObs(s layer, sub subjects.Subject, in ...observations.Input) (observations.Observation, error) {
	for i := range in {
		in[i].SubjectID = sub.ID
	}
	res, err := citations.CreateWithObservations(b.c.Cat, UserID, citations.CreateInput{
		ArtifactID: s.artifact.ID, LocatorJSON: locator,
	}, in)
	if err != nil {
		return observations.Observation{}, err
	}
	if len(res.Observations) == 0 {
		return observations.Observation{}, fmt.Errorf("citation wrote no observations")
	}
	return res.Observations[0], nil
}

func (b *bld) participation(s layer, event subjects.Subject, person []byte, role string) error {
	// person is a handle. Participation links Subjects. Resolve a member subject
	// of the person on this source when one exists; otherwise skip the link
	// (filler births are typed events without a same-source member).
	member, err := b.memberOn(person, s.source.ID)
	if err != nil || len(member) == 0 {
		return err
	}
	_, err = connect.CreateCitedBridge(b.c.Cat, UserID, connect.CreateInput{
		SourceID: s.source.ID, FromSubjectID: member, ToSubjectID: event.ID, BridgeTypeKey: "participation",
		Citation: citations.CreateInput{ArtifactID: s.artifact.ID, LocatorJSON: locator},
		Observations: []observations.Input{
			{PropertyID: b.mustProp("person").ID, ValueSubjectID: member},
			{PropertyID: b.mustProp("event").ID, ValueSubjectID: event.ID},
			{PropertyID: b.mustProp("role").ID, ValueTermID: b.mustTerm("role", role).ID},
		},
	})
	return err
}

func (b *bld) relationship(s layer, person []byte, other subjects.Subject, kind string) error {
	member, err := b.memberOn(person, s.source.ID)
	if err != nil || len(member) == 0 {
		return err
	}
	_, err = connect.CreateCitedBridge(b.c.Cat, UserID, connect.CreateInput{
		SourceID: s.source.ID, FromSubjectID: member, ToSubjectID: other.ID, BridgeTypeKey: "relationship",
		Citation: citations.CreateInput{ArtifactID: s.artifact.ID, LocatorJSON: locator},
		Observations: []observations.Input{
			{PropertyID: b.mustProp("person").ID, ValueSubjectID: member},
			{PropertyID: b.mustProp("related_to").ID, ValueSubjectID: other.ID},
			{PropertyID: b.mustProp("relationship_type").ID, ValueTermID: b.mustTerm("relationship_type", kind).ID},
		},
	})
	return err
}

func (b *bld) placeRel(s layer, from, to subjects.Subject, kind string) error {
	_, err := connect.CreateCitedBridge(b.c.Cat, UserID, connect.CreateInput{
		SourceID: s.source.ID, FromSubjectID: from.ID, ToSubjectID: to.ID, BridgeTypeKey: "place_relationship",
		Citation: citations.CreateInput{ArtifactID: s.artifact.ID, LocatorJSON: locator},
		Observations: []observations.Input{
			{PropertyID: b.mustProp("from").ID, ValueSubjectID: from.ID},
			{PropertyID: b.mustProp("to").ID, ValueSubjectID: to.ID},
			{PropertyID: b.mustProp("place_relationship_type").ID, ValueTermID: b.mustTerm("place_relationship_type", kind).ID},
		},
	})
	return err
}

func (b *bld) location(s layer, event, place subjects.Subject) error {
	_, err := connect.CreateCitedBridge(b.c.Cat, UserID, connect.CreateInput{
		SourceID: s.source.ID, FromSubjectID: event.ID, ToSubjectID: place.ID, BridgeTypeKey: "location",
		Citation: citations.CreateInput{ArtifactID: s.artifact.ID, LocatorJSON: locator},
		Observations: []observations.Input{
			{PropertyID: b.mustProp("event").ID, ValueSubjectID: event.ID},
			{PropertyID: b.mustProp("place").ID, ValueSubjectID: place.ID},
		},
	})
	return err
}

func (b *bld) memberOn(entity, sourceID []byte) ([]byte, error) {
	db, err := b.c.Cat.DB()
	if err != nil {
		return nil, err
	}
	var id []byte
	err = db.QueryRow(`
		SELECT ic.subject_id FROM identity_claims ic
		JOIN subjects s ON s.id = ic.subject_id
		WHERE ic.entity_id = ? AND s.source_id = ? AND ic.status = 'accepted'
		LIMIT 1`, entity, sourceID).Scan(&id)
	if err != nil {
		return nil, nil
	}
	return id, nil
}

func (b *bld) typeID(key string) []byte {
	if id, ok := b.typ[key]; ok {
		return id
	}
	st, err := subjecttypes.Lookup(b.c.Cat, key, subjecttypes.OriginProvenencia)
	if err != nil {
		panic(err)
	}
	b.typ[key] = st.ID
	return st.ID
}

func (b *bld) mustProp(key string) properties.Property {
	if p, ok := b.prop[key]; ok {
		return p
	}
	p, err := properties.Lookup(b.c.Cat, key, properties.OriginProvenencia)
	if err != nil {
		panic(err)
	}
	b.prop[key] = p
	return p
}

func (b *bld) mustTerm(prop, key string) propertyterms.Term {
	k := prop + "/" + key
	if t, ok := b.term[k]; ok {
		return t
	}
	t, err := propertyterms.Lookup(b.c.Cat, b.mustProp(prop).ID, key, propertyterms.OriginProvenencia)
	if err != nil {
		panic(err)
	}
	b.term[k] = t
	return t
}
