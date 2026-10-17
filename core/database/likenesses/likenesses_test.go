package likenesses_test

import (
	"bytes"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/claimconfidencegrades"
	"github.com/mendahu/provenencia/core/database/likenesses"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/observations/observationstest"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sourcecredibilitygrades"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
)

var userID = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

type fixture struct {
	t        *testing.T
	c        *database.Catalog
	typeID   []byte
	person   subjecttypes.Type
	likeness properties.Property
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
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
	must(t, sourcecredibilitygrades.Install(c))
	must(t, claimconfidencegrades.Install(c))
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book"})
	must(t, err)
	f := &fixture{t: t, c: c, typeID: typeID}
	f.person, err = subjecttypes.Lookup(c, "person", vocabkey.OriginProvenencia)
	must(t, err)
	f.likeness, err = properties.Lookup(c, "likeness", vocabkey.OriginProvenencia)
	must(t, err)
	return f
}

func (f *fixture) source(title string) sources.Source {
	f.t.Helper()
	src, _, err := writes.Run(f.c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: f.typeID, Title: title})
	})
	must(f.t, err)
	return src
}

func (f *fixture) personOn(src sources.Source) subjects.Subject {
	f.t.Helper()
	s, err := writes.Call(f.c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (subjects.Subject, []rowchange.Change, error) {
		return subjects.Create(tx, userID, subjects.CreateInput{SourceID: src.ID, SubjectTypeID: f.person.ID}, nil)
	})
	must(f.t, err)
	return s
}

func (f *fixture) likenessOn(src sources.Source, s subjects.Subject, shade uint8) observations.Observation {
	f.t.Helper()
	art := observationstest.FileArtifact(f.t, f.c, userID, src.ID)
	res, _, err := observationstest.SaveCitation(f.c, userID, art, observations.Input{
		SubjectID: s.ID, PropertyID: f.likeness.ID,
		Depiction: &observations.DepictionInput{RegionJSON: observationstest.LikenessRegion, ImagePath: observationstest.Crop(f.t, 40, 30, shade)},
	})
	must(f.t, err)
	return res.Observations[0]
}

func (f *fixture) promote(s subjects.Subject, onto []byte) []byte {
	f.t.Helper()
	res, err := writes.Call(f.c, writes.Op{Action: "promote_subject", UserID: userID}, func(tx *database.Tx) (promote.Result, []rowchange.Change, error) {
		return promote.Save(tx, userID, promote.Input{SubjectID: s.ID, EntityID: onto})
	})
	must(f.t, err)
	return res.Entity.ID
}

func (f *fixture) imageOf(observationID []byte) []byte {
	f.t.Helper()
	db, err := f.c.DB()
	must(f.t, err)
	var id []byte
	must(f.t, db.QueryRow(`SELECT dv.image_file_id FROM observations o
		JOIN depiction_values dv ON dv.id = o.value_depiction_id WHERE o.id = ?`, observationID).Scan(&id))
	return id
}

func TestForEntities(t *testing.T) {
	t.Run("no ids", func(t *testing.T) {
		f := newFixture(t)
		got, err := likenesses.ForEntities(f.c.Graph(), nil)
		must(t, err)
		if got != nil {
			t.Fatalf("got %+v", got)
		}
	})
	t.Run("a handle with no likeness is absent", func(t *testing.T) {
		f := newFixture(t)
		a := f.personOn(f.source("Register"))
		h := f.promote(a, nil)
		got, err := likenesses.ForEntities(f.c.Graph(), [][]byte{h})
		must(t, err)
		if _, ok := got[string(h)]; ok || len(got) != 0 {
			t.Fatalf("got %+v", got)
		}
	})
	t.Run("earliest first, and a redraw keeps its place", func(t *testing.T) {
		f := newFixture(t)
		reg, cen := f.source("Register"), f.source("Census")
		a, b := f.personOn(reg), f.personOn(cen)
		oa := f.likenessOn(reg, a, 100)
		ob := f.likenessOn(cen, b, 200)
		h := f.promote(a, nil)
		f.promote(b, h)
		got, err := likenesses.ForEntities(f.c.Graph(), [][]byte{h})
		must(t, err)
		g := got[string(h)]
		if len(g) != 2 || !bytes.Equal(g[0].ObservationIDs[0], oa.ID) || !bytes.Equal(g[1].ObservationIDs[0], ob.ID) {
			t.Fatalf("gallery %+v", g)
		}
		_, err = observationstest.UpdateObservation(f.c, userID, observations.Input{
			ID: oa.ID, SubjectID: a.ID, PropertyID: f.likeness.ID,
			Depiction: &observations.DepictionInput{RegionJSON: observationstest.LikenessRegion, ImagePath: observationstest.Crop(t, 40, 30, 150)}})
		must(t, err)
		got, err = likenesses.ForEntities(f.c.Graph(), [][]byte{h})
		must(t, err)
		g = got[string(h)]
		if len(g) != 2 || !bytes.Equal(g[0].ObservationIDs[0], oa.ID) || !bytes.Equal(g[0].ImageFileID, f.imageOf(oa.ID)) ||
			!bytes.Equal(g[1].ObservationIDs[0], ob.ID) {
			t.Fatalf("after redraw %+v", g)
		}
	})
}
