package promotecompare_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/promotecompare"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
)

var userID = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

const locator = `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`

type fixture struct {
	t        *testing.T
	c        *database.Catalog
	artifact artifacts.Artifact
	types    map[string]subjecttypes.Type
	props    map[string]properties.Property
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	f := &fixture{t: t, c: c, types: map[string]subjecttypes.Type{}, props: map[string]properties.Property{}}
	r, err := ref.Mint(ref.PrefixUser)
	f.must(err)
	f.must(users.Upsert(c, userID, "Tester", r))
	f.must(subjectvocab.Install(c))
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{Key: "book", Origin: sourcetypes.OriginProvenencia, Label: "Book"})
	f.must(err)
	src, err := sources.Create(c, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Register"})
	f.must(err)
	f.artifact, err = artifacts.Create(c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
	f.must(err)
	for _, k := range []string{"person", "place"} {
		st, err := subjecttypes.Lookup(c, k, subjecttypes.OriginProvenencia)
		f.must(err)
		f.types[k] = st
	}
	for _, k := range []string{"name", "toponym"} {
		p, err := properties.Lookup(c, k, properties.OriginProvenencia)
		f.must(err)
		f.props[k] = p
	}
	return f
}

func (f *fixture) must(err error) {
	f.t.Helper()
	if err != nil {
		f.t.Fatal(err)
	}
}

// person is a person Subject with one name Observation per spec
// ("given=James|surname=Robins"), on one Citation. A spec starting with "-"
// is negative.
func (f *fixture) person(specs ...string) (subjects.Subject, []observations.Observation) {
	f.t.Helper()
	s, err := subjects.Create(f.c, userID, subjects.CreateInput{SourceID: f.artifact.SourceID, SubjectTypeID: f.types["person"].ID}, nil)
	f.must(err)
	var in []observations.Input
	for _, spec := range specs {
		o := observations.Input{SubjectID: s.ID, PropertyID: f.props["name"].ID, Name: nameValue(strings.TrimPrefix(spec, "-"))}
		if strings.HasPrefix(spec, "-") {
			o.Polarity = observations.PolarityNegative
		}
		in = append(in, o)
	}
	if len(in) == 0 {
		return s, nil
	}
	res, err := citations.CreateWithObservations(f.c, userID, citations.CreateInput{ArtifactID: f.artifact.ID, LocatorJSON: locator}, in)
	f.must(err)
	return s, res.Observations
}

func nameValue(spec string) *namevalues.Value {
	v := &namevalues.Value{}
	var words []string
	for i, item := range strings.Split(spec, "|") {
		typ, val, _ := strings.Cut(item, "=")
		v.Parts = append(v.Parts, namevalues.Part{Idx: i, Type: typ, Value: val})
		words = append(words, val)
	}
	v.Form = strings.Join(words, " ")
	return v
}

func (f *fixture) compare(s subjects.Subject, entityID []byte) promotecompare.Comparison {
	f.t.Helper()
	db, err := f.c.DB()
	f.must(err)
	got, err := promotecompare.Compare(db, s.ID, entityID)
	f.must(err)
	return got
}

// shape is "property: incoming form → member form ✓|✗, …; …".
func shape(c promotecompare.Comparison) string {
	var props []string
	for _, p := range c.Properties {
		var ins []string
		for _, in := range p.Incoming {
			var pairs []string
			for _, pr := range in.Pairs {
				mark := "✗"
				if pr.Compatible {
					mark = "✓"
				}
				pairs = append(pairs, pr.Member.Value.Name.Form+" "+mark)
			}
			ins = append(ins, in.Record.Value.Name.Form+" → "+strings.Join(pairs, ", "))
		}
		props = append(props, p.Key+": "+strings.Join(ins, "; "))
	}
	return strings.Join(props, " | ")
}

func TestCompare(t *testing.T) {
	t.Run("two members, mostly compatible", func(t *testing.T) {
		f := newFixture(t)
		birth, _ := f.person("given=James|surname=Robins")
		first, err := promote.Save(f.c, userID, promote.Input{SubjectID: birth.ID})
		f.must(err)
		marriage, _ := f.person("given=J.|surname=Robins")
		_, err = promote.Save(f.c, userID, promote.Input{SubjectID: marriage.ID, EntityID: first.Entity.ID})
		f.must(err)
		census, _ := f.person("given=James|surname=Robins")
		got := f.compare(census, first.Entity.ID)
		if got.Members != 2 {
			t.Fatalf("Members = %d", got.Members)
		}
		if s, want := shape(got), "name: James Robins → James Robins ✓, J. Robins ✓"; s != want {
			t.Fatalf("got  %s\nwant %s", s, want)
		}
		pair := got.Properties[0].Incoming[0].Pairs[0]
		if string(pair.Member.ClaimID) != string(first.Claim.ID) || pair.Member.SourceTitle != "Register" ||
			pair.Member.SubjectRef != birth.Ref || got.Properties[0].Incoming[0].Record.ClaimID != nil {
			t.Fatalf("records %+v", pair.Member)
		}
	})

	t.Run("a differing pair is not compatible", func(t *testing.T) {
		f := newFixture(t)
		a, _ := f.person("given=James|surname=Robbins")
		first, err := promote.Save(f.c, userID, promote.Input{SubjectID: a.ID})
		f.must(err)
		b, _ := f.person("given=James|surname=Robins")
		if s, want := shape(f.compare(b, first.Entity.ID)), "name: James Robins → James Robbins ✗"; s != want {
			t.Fatalf("got %s", s)
		}
	})

	t.Run("several Observations on each side are each their own row", func(t *testing.T) {
		f := newFixture(t)
		a, _ := f.person("given=James|surname=Robins", "given=Jim|surname=Robins")
		first, err := promote.Save(f.c, userID, promote.Input{SubjectID: a.ID})
		f.must(err)
		b, _ := f.person("given=James", "-given=Jim")
		want := "name: James → James Robins ✓, Jim Robins ✗; Jim → James Robins ✗, Jim Robins ✗"
		if s := shape(f.compare(b, first.Entity.ID)); s != want {
			t.Fatalf("got  %s\nwant %s", s, want)
		}
	})

	t.Run("provisional members and other handles are left out", func(t *testing.T) {
		f := newFixture(t)
		a, _ := f.person("given=James")
		first, err := promote.Save(f.c, userID, promote.Input{SubjectID: a.ID})
		f.must(err)
		p, _ := f.person("given=Jim")
		_, err = identityclaims.Create(f.c, userID, identityclaims.CreateInput{
			SubjectID: p.ID, EntityID: first.Entity.ID, Status: identityclaims.StatusProvisional})
		f.must(err)
		other, _ := f.person("given=Jack")
		_, err = promote.Save(f.c, userID, promote.Input{SubjectID: other.ID})
		f.must(err)
		b, _ := f.person("given=James")
		got := f.compare(b, first.Entity.ID)
		if got.Members != 1 || shape(got) != "name: James → James ✓" {
			t.Fatalf("%d %s", got.Members, shape(got))
		}
	})

	t.Run("nothing to compare", func(t *testing.T) {
		f := newFixture(t)
		a, _ := f.person()
		first, err := promote.Save(f.c, userID, promote.Input{SubjectID: a.ID})
		f.must(err)
		b, _ := f.person("given=James")
		if s := shape(f.compare(b, first.Entity.ID)); s != "name: James → " {
			t.Fatalf("member with no Observations: %s", s)
		}
		c, _ := f.person()
		if got := f.compare(c, first.Entity.ID); len(got.Properties) != 0 || got.Members != 1 {
			t.Fatalf("incoming with no Observations: %+v", got)
		}
	})

	t.Run("the target is refused when missing, merged or of another type", func(t *testing.T) {
		f := newFixture(t)
		a, _ := f.person("given=James")
		first, err := promote.Save(f.c, userID, promote.Input{SubjectID: a.ID})
		f.must(err)
		place, err := subjects.Create(f.c, userID, subjects.CreateInput{SourceID: f.artifact.SourceID, SubjectTypeID: f.types["place"].ID}, nil)
		f.must(err)
		db, err := f.c.DB()
		f.must(err)
		if _, err := promotecompare.Compare(db, place.ID, first.Entity.ID); !errors.Is(err, identityclaims.ErrTypeMismatch) {
			t.Fatalf("other type: %v", err)
		}
		b, _ := f.person("given=James")
		missing := make([]byte, 16)
		if _, err := promotecompare.Compare(db, b.ID, missing); !errors.Is(err, promote.ErrInvalid) {
			t.Fatalf("missing: %v", err)
		}
		other, _ := f.person("given=James")
		into, err := promote.Save(f.c, userID, promote.Input{SubjectID: other.ID})
		f.must(err)
		_, err = db.Exec(`UPDATE canonical_entities SET merged_into_id = ? WHERE id = ?`, into.Entity.ID, first.Entity.ID)
		f.must(err)
		if _, err := promotecompare.Compare(db, b.ID, first.Entity.ID); !errors.Is(err, promote.ErrInvalid) {
			t.Fatalf("merged: %v", err)
		}
	})

	t.Run("confirmed pairs from the comparison save as pins", func(t *testing.T) {
		f := newFixture(t)
		a, _ := f.person("given=James|surname=Robins")
		first, err := promote.Save(f.c, userID, promote.Input{SubjectID: a.ID})
		f.must(err)
		b, _ := f.person("given=J.|surname=Robins")
		var pairs []promote.Pair
		for _, p := range f.compare(b, first.Entity.ID).Properties {
			for _, in := range p.Incoming {
				for _, pr := range in.Pairs {
					if pr.Compatible {
						pairs = append(pairs, promote.Pair{IncomingObservationID: in.Record.ObservationID, MemberObservationID: pr.Member.ObservationID})
					}
				}
			}
		}
		res, err := promote.Save(f.c, userID, promote.Input{SubjectID: b.ID, EntityID: first.Entity.ID, Pairs: pairs})
		f.must(err)
		if fmt.Sprint(len(pairs), res.Pins) != "1 2" {
			t.Fatalf("pairs %d pins %d", len(pairs), res.Pins)
		}
	})
}
