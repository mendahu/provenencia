package gallery_test

import (
	"bytes"
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/dbtest"
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
	"github.com/mendahu/provenencia/core/database/gallery"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/observations/observationstest"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/promote/promotetest"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sourcecredibility"
	"github.com/mendahu/provenencia/core/database/sourcecredibilitygrades"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjects/subjectstest"
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
	return f.subjectOn(src, "person")
}

func (f *fixture) subjectOn(src sources.Source, kind string) subjects.Subject {
	f.t.Helper()
	ty, err := subjecttypes.Lookup(f.c, kind, vocabkey.OriginProvenencia)
	must(f.t, err)
	graph := evidencegraphstest.SoleGraph(f.t, f.c, src.ID).ID
	return subjectstest.OfType(f.t, f.c, userID, graph, ty.ID, "")
}

func (f *fixture) promote(s subjects.Subject, onto []byte) []byte {
	f.t.Helper()
	res := promotetest.Save(f.t, f.c, userID, promote.Input{SubjectID: s.ID, EntityID: onto})
	return res.Entity.ID
}

// credibility grades a Source with a seeded credibility grade.
func (f *fixture) credibility(src sources.Source, key string) {
	f.t.Helper()
	g, err := sourcecredibilitygrades.Lookup(f.c, key, vocabkey.OriginProvenencia)
	must(f.t, err)
	_, err = writes.Call(f.c, writes.Op{Action: "create_source_credibility_assessment", UserID: userID}, func(tx *database.Tx) (sourcecredibility.Assessment, []rowchange.Change, error) {
		return sourcecredibility.Upsert(tx, userID, sourcecredibility.UpsertInput{SourceID: src.ID, CredibilityGradeID: g.ID})
	})
	must(f.t, err)
}

// provisional links the Subject to the handle by a provisional Identity Claim;
// the Subject is never promoted.
func (f *fixture) provisional(s subjects.Subject, onto []byte) {
	f.t.Helper()
	_, err := writes.Call(f.c, writes.Op{Action: "create_identity_claim", UserID: userID}, func(tx *database.Tx) (identityclaims.Claim, []rowchange.Change, error) {
		return identityclaims.Create(tx, userID, identityclaims.CreateInput{SubjectID: s.ID, EntityID: onto, Status: identityclaims.StatusProvisional})
	})
	must(f.t, err)
}

// negativeTwin files a negative N (Album, shade 100), then B (Census, shade
// 200), then A (Register, shade 100: N's pixels), and promotes all three
// Subjects onto one Person.
func (f *fixture) negativeTwin() (h []byte, n, a, b observations.Observation) {
	f.t.Helper()
	alb, cen, reg := f.source("Album"), f.source("Census"), f.source("Register")
	sn, sb, sa := f.personOn(alb), f.personOn(cen), f.personOn(reg)
	n = observationstest.SaveNegativeLikeness(f.t, f.c, userID, alb.ID, sn.ID, 100)
	b = observationstest.SaveLikeness(f.t, f.c, userID, cen.ID, sb.ID, 200)
	a = observationstest.SaveLikeness(f.t, f.c, userID, reg.ID, sa.ID, 100)
	h = f.promote(sa, nil)
	f.promote(sb, h)
	f.promote(sn, h)
	return h, n, a, b
}

func TestForEntities(t *testing.T) {
	t.Run("no ids", func(t *testing.T) {
		f := newFixture(t)
		got, err := gallery.ForEntities(f.c.Graph(), nil)
		must(t, err)
		if got != nil {
			t.Fatalf("got %+v", got)
		}
	})
	t.Run("a handle with no likeness is absent", func(t *testing.T) {
		f := newFixture(t)
		a := f.personOn(f.source("Register"))
		h := f.promote(a, nil)
		got, err := gallery.ForEntities(f.c.Graph(), [][]byte{h})
		must(t, err)
		if _, ok := got[string(h)]; ok || len(got) != 0 {
			t.Fatalf("got %+v", got)
		}
	})
	t.Run("earliest first, and a redraw keeps its place", func(t *testing.T) {
		f := newFixture(t)
		reg, cen := f.source("Register"), f.source("Census")
		a, b := f.personOn(reg), f.personOn(cen)
		oa := observationstest.SaveLikeness(t, f.c, userID, reg.ID, a.ID, 100)
		ob := observationstest.SaveLikeness(t, f.c, userID, cen.ID, b.ID, 200)
		h := f.promote(a, nil)
		f.promote(b, h)
		got, err := gallery.ForEntities(f.c.Graph(), [][]byte{h})
		must(t, err)
		g := got[string(h)]
		if len(g) != 2 || !bytes.Equal(g[0].ObservationIDs[0], oa.ID) || !bytes.Equal(g[1].ObservationIDs[0], ob.ID) {
			t.Fatalf("gallery %+v", g)
		}
		_, err = observationstest.UpdateObservation(f.c, userID, observations.Input{
			ID: oa.ID, SubjectID: a.ID, PropertyID: f.likeness.ID,
			Depiction: &observations.DepictionInput{RegionJSON: observationstest.LikenessRegion, ImagePath: observationstest.Crop(t, 40, 30, 150)}})
		must(t, err)
		got, err = gallery.ForEntities(f.c.Graph(), [][]byte{h})
		must(t, err)
		g = got[string(h)]
		if len(g) != 2 || !bytes.Equal(g[0].ObservationIDs[0], oa.ID) || !bytes.Equal(g[0].ImageFileID, observationstest.ImageOf(t, f.c, oa.ID)) ||
			!bytes.Equal(g[1].ObservationIDs[0], ob.ID) {
			t.Fatalf("after redraw %+v", g)
		}
	})
	t.Run("a negative of the same picture is not one of its records", func(t *testing.T) {
		f := newFixture(t)
		h, n, a, b := f.negativeTwin()
		g := f.gallery(h)
		if len(g) != 2 || len(g[0].ObservationIDs) != 1 || !bytes.Equal(g[0].ObservationIDs[0], b.ID) ||
			len(g[1].ObservationIDs) != 1 || !bytes.Equal(g[1].ObservationIDs[0], a.ID) || !g[0].Leads {
			t.Fatalf("gallery %+v", g)
		}
		if _, ok := gallery.ByObservation(g, n.ID); ok {
			t.Fatalf("negative is in the gallery %+v", g)
		}
	})
	t.Run("a provisional record of the same picture is not one of its records", func(t *testing.T) {
		f := newFixture(t)
		reg, diary := f.source("Register"), f.source("Diary")
		sa, sp := f.personOn(reg), f.personOn(diary)
		p := observationstest.SaveLikeness(t, f.c, userID, diary.ID, sp.ID, 100)
		a := observationstest.SaveLikeness(t, f.c, userID, reg.ID, sa.ID, 100)
		h := f.promote(sa, nil)
		f.provisional(sp, h)
		g := f.gallery(h)
		if len(g) != 1 || len(g[0].ObservationIDs) != 1 || !bytes.Equal(g[0].ObservationIDs[0], a.ID) {
			t.Fatalf("gallery %+v", g)
		}
		if _, ok := gallery.ByObservation(g, p.ID); ok {
			t.Fatalf("provisional record is in the gallery %+v", g)
		}
	})
	t.Run("a provisional record does not hide a weak likeness", func(t *testing.T) {
		f := newFixture(t)
		reg, gossip, diary := f.source("Register"), f.source("Gossip"), f.source("Diary")
		sa, sw, sp := f.personOn(reg), f.personOn(gossip), f.personOn(diary)
		observationstest.SaveLikeness(t, f.c, userID, reg.ID, sa.ID, 100)
		w := observationstest.SaveLikeness(t, f.c, userID, gossip.ID, sw.ID, 200)
		p := observationstest.SaveLikeness(t, f.c, userID, diary.ID, sp.ID, 200)
		f.credibility(gossip, "low_trust")
		h := f.promote(sa, nil)
		f.promote(sw, h)
		f.provisional(sp, h)
		g := f.gallery(h)
		if len(g) != 2 || len(g[1].ObservationIDs) != 1 || !bytes.Equal(g[1].ObservationIDs[0], w.ID) {
			t.Fatalf("gallery %+v", g)
		}
		if _, ok := gallery.ByObservation(g, p.ID); ok {
			t.Fatalf("provisional record is in the gallery %+v", g)
		}
	})
	t.Run("place picture on a Place handle", func(t *testing.T) {
		f := newFixture(t)
		src := f.source("Map")
		s := f.subjectOn(src, "place")
		o := observationstest.SaveDepiction(t, f.c, userID, src.ID, s.ID, vocabkey.PropertyPlacePicture, 100)
		h := f.promote(s, nil)
		g := f.gallery(h)
		if len(g) != 1 || len(g[0].ObservationIDs) != 1 || !bytes.Equal(g[0].ObservationIDs[0], o.ID) {
			t.Fatalf("gallery %+v", g)
		}
	})
	t.Run("event picture on an Event handle", func(t *testing.T) {
		f := newFixture(t)
		src := f.source("Photo")
		s := f.subjectOn(src, "event")
		o := observationstest.SaveDepiction(t, f.c, userID, src.ID, s.ID, vocabkey.PropertyEventPicture, 100)
		h := f.promote(s, nil)
		g := f.gallery(h)
		if len(g) != 1 || len(g[0].ObservationIDs) != 1 || !bytes.Equal(g[0].ObservationIDs[0], o.ID) {
			t.Fatalf("gallery %+v", g)
		}
	})
	t.Run("tile image is the depiction row, not the reconciler cache", func(t *testing.T) {
		f := newFixture(t)
		src := f.source("Register")
		s := f.personOn(src)
		o := observationstest.SaveLikeness(t, f.c, userID, src.ID, s.ID, 100)
		h := f.promote(s, nil)
		want := observationstest.ImageOf(t, f.c, o.ID)
		other := observationstest.FileArtifact(t, f.c, userID, src.ID).FileID
		db, err := f.c.DB()
		must(t, err)
		_, err = db.Exec(`UPDATE auto_reconciler_values SET value_image_file_id = ? WHERE entity_id = ?`, other, h)
		must(t, err)
		g := f.gallery(h)
		if len(g) != 1 || !bytes.Equal(g[0].ImageFileID, want) || bytes.Equal(g[0].ImageFileID, other) {
			t.Fatalf("gallery image %x, want depiction %x (not cache %x)", g[0].ImageFileID, want, other)
		}
	})
}

// twoLikenesses files likeness A (Register) then B (Census) on two Subjects
// promoted onto one Person.
func (f *fixture) twoLikenesses() (h []byte, a, b observations.Observation, subjectB subjects.Subject) {
	f.t.Helper()
	reg, cen := f.source("Register"), f.source("Census")
	sa, sb := f.personOn(reg), f.personOn(cen)
	a = observationstest.SaveLikeness(f.t, f.c, userID, reg.ID, sa.ID, 100)
	b = observationstest.SaveLikeness(f.t, f.c, userID, cen.ID, sb.ID, 200)
	h = f.promote(sa, nil)
	f.promote(sb, h)
	return h, a, b, sb
}

func (f *fixture) setPreferred(h, observationID []byte) {
	f.t.Helper()
	_, err := writes.Call(f.c, writes.Op{Action: "set_preferred_likeness", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
		ch, err := canonicalentities.SetPreferredLikeness(tx, userID, h, observationID)
		return struct{}{}, ch, err
	})
	must(f.t, err)
}

func (f *fixture) gallery(h []byte) []gallery.Picture {
	f.t.Helper()
	got, err := gallery.ForEntities(f.c.Graph(), [][]byte{h})
	must(f.t, err)
	return got[string(h)]
}

func (f *fixture) pointer(h []byte) []byte {
	f.t.Helper()
	db, err := f.c.DB()
	must(f.t, err)
	var p []byte
	must(f.t, db.QueryRow(`SELECT preferred_likeness_observation_id FROM canonical_entities WHERE id = ?`, h).Scan(&p))
	return p
}

func preferredFlags(g []gallery.Picture) []bool {
	out := make([]bool, len(g))
	for i := range g {
		out[i] = g[i].Preferred
	}
	return out
}

func leadsFlags(g []gallery.Picture) []bool {
	out := make([]bool, len(g))
	for i := range g {
		out[i] = g[i].Leads
	}
	return out
}

func TestForEntitiesPreferred(t *testing.T) {
	t.Run("no pointer leads with the earliest", func(t *testing.T) {
		f := newFixture(t)
		h, a, _, _ := f.twoLikenesses()
		g := f.gallery(h)
		if !slices.Equal(preferredFlags(g), []bool{false, false}) {
			t.Fatalf("flags %v", preferredFlags(g))
		}
		if !slices.Equal(leadsFlags(g), []bool{true, false}) {
			t.Fatalf("leads %v", leadsFlags(g))
		}
		l, ok := gallery.Lead(g)
		if !ok || !bytes.Equal(l.ObservationIDs[0], a.ID) {
			t.Fatalf("lead %+v", l)
		}
	})
	t.Run("pointer leads and keeps the order", func(t *testing.T) {
		f := newFixture(t)
		h, a, b, _ := f.twoLikenesses()
		f.setPreferred(h, b.ID)
		g := f.gallery(h)
		if len(g) != 2 || !bytes.Equal(g[0].ObservationIDs[0], a.ID) || !slices.Equal(preferredFlags(g), []bool{false, true}) {
			t.Fatalf("gallery %+v", g)
		}
		if !slices.Equal(leadsFlags(g), []bool{false, true}) {
			t.Fatalf("leads %v", leadsFlags(g))
		}
		l, ok := gallery.Lead(g)
		if !ok || !bytes.Equal(l.ObservationIDs[0], b.ID) {
			t.Fatalf("lead %+v", l)
		}
	})
	t.Run("unpromoted likeness falls back and keeps the pointer row", func(t *testing.T) {
		f := newFixture(t)
		h, a, b, sb := f.twoLikenesses()
		f.setPreferred(h, b.ID)
		_, err := writes.Call(f.c, writes.Op{Action: "unpromote_subject", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			ch, err := promote.Unpromote(tx, userID, sb.ID)
			return struct{}{}, ch, err
		})
		must(t, err)
		g := f.gallery(h)
		l, ok := gallery.Lead(g)
		if len(g) != 1 || !ok || !bytes.Equal(l.ObservationIDs[0], a.ID) || l.Preferred {
			t.Fatalf("gallery %+v lead %+v", g, l)
		}
		if !slices.Equal(leadsFlags(g), []bool{true}) {
			t.Fatalf("leads %v", leadsFlags(g))
		}
		if !bytes.Equal(f.pointer(h), b.ID) {
			t.Fatalf("pointer row %x", f.pointer(h))
		}
	})
	t.Run("stronger negative denying the pointed image falls back", func(t *testing.T) {
		f := newFixture(t)
		h, a, b, _ := f.twoLikenesses()
		f.setPreferred(h, b.ID)
		album := f.source("Album")
		c := f.personOn(album)
		observationstest.SaveNegativeLikeness(t, f.c, userID, album.ID, c.ID, 200)
		f.credibility(album, "high_trust")
		f.promote(c, h)
		g := f.gallery(h)
		l, ok := gallery.Lead(g)
		if len(g) != 1 || !ok || !bytes.Equal(l.ObservationIDs[0], a.ID) || !slices.Equal(preferredFlags(g), []bool{false}) {
			t.Fatalf("gallery %+v lead %+v", g, l)
		}
		if !slices.Equal(leadsFlags(g), []bool{true}) {
			t.Fatalf("leads %v", leadsFlags(g))
		}
		if !bytes.Equal(f.pointer(h), b.ID) {
			t.Fatalf("pointer row %x", f.pointer(h))
		}
	})
	t.Run("a stored pointer at a negative falls back to the earliest", func(t *testing.T) {
		f := newFixture(t)
		h, n, _, b := f.negativeTwin()
		db, err := f.c.DB()
		must(t, err)
		_, err = db.Exec(`UPDATE canonical_entities SET preferred_likeness_observation_id = ? WHERE id = ?`, n.ID, h)
		must(t, err)
		g := f.gallery(h)
		l, ok := gallery.Lead(g)
		if !ok || !bytes.Equal(l.ObservationIDs[0], b.ID) {
			t.Fatalf("gallery %+v lead %+v", g, l)
		}
		if slices.Contains(preferredFlags(g), true) {
			t.Fatalf("preferred %v", preferredFlags(g))
		}
		if !bytes.Equal(f.pointer(h), n.ID) {
			t.Fatalf("pointer row %x", f.pointer(h))
		}
	})
	t.Run("deleting the Observation clears the pointer", func(t *testing.T) {
		t.Run("clears preferred pointer row", func(t *testing.T) {
			f := newFixture(t)
			h, a, b, _ := f.twoLikenesses()
			f.setPreferred(h, b.ID)
			_, err := writes.Call(f.c, writes.Op{Action: "delete_observation", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
				ch, err := observations.Delete(tx, userID, b.ID)
				return struct{}{}, ch, err
			})
			must(t, err)
			if p := f.pointer(h); p != nil {
				t.Fatalf("pointer %x", p)
			}
			g := f.gallery(h)
			if !slices.Equal(leadsFlags(g), []bool{true}) {
				t.Fatalf("leads %v", leadsFlags(g))
			}
			l, ok := gallery.Lead(g)
			if !ok || !bytes.Equal(l.ObservationIDs[0], a.ID) {
				t.Fatalf("lead %+v", l)
			}
		})
		t.Run("records pointer clear in revision", func(t *testing.T) {
			f := newFixture(t)
			h, _, b, _ := f.twoLikenesses()
			diary := f.source("Diary")
			sd := f.personOn(diary)
			h2 := f.promote(sd, nil)
			ob2 := observationstest.SaveLikeness(f.t, f.c, userID, diary.ID, sd.ID, 150)
			f.setPreferred(h2, ob2.ID)
			var wantFields map[string]rowchange.FieldDiff
			_, err := writes.Call(f.c, writes.Op{Action: "set_preferred_likeness", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
				ch, err := canonicalentities.SetPreferredLikeness(tx, userID, h2, nil)
				if len(ch) == 1 {
					wantFields = ch[0].Fields
				}
				return struct{}{}, ch, err
			})
			must(t, err)
			f.setPreferred(h2, ob2.ID)
			f.setPreferred(h, b.ID)
			var got []rowchange.Change
			_, err = writes.Call(f.c, writes.Op{Action: "delete_observation", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
				ch, err := observations.Delete(tx, userID, b.ID)
				got = ch
				return struct{}{}, ch, err
			})
			must(t, err)
			dbtest.RequireUpdateBeforeDelete(t, got, "observation", "canonical_entity", h, wantFields)
		})
		t.Run("non-preferred delete emits no canonical_entity change", func(t *testing.T) {
			f := newFixture(t)
			_, _, b, _ := f.twoLikenesses()
			var got []rowchange.Change
			_, err := writes.Call(f.c, writes.Op{Action: "delete_observation", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
				ch, err := observations.Delete(tx, userID, b.ID)
				got = ch
				return struct{}{}, ch, err
			})
			must(t, err)
			for _, ch := range got {
				if ch.EntityType == "canonical_entity" {
					t.Fatalf("unexpected canonical_entity change %+v", ch)
				}
			}
		})
	})
	t.Run("empty gallery has no lead", func(t *testing.T) {
		if len(leadsFlags(nil)) != 0 {
			t.Fatal("flags")
		}
		if _, ok := gallery.Lead(nil); ok {
			t.Fatal("ok")
		}
	})
}

func TestSetPreferredLikeness(t *testing.T) {
	set := func(f *fixture, h, obs []byte, user []byte) ([]rowchange.Change, writes.Result, error) {
		return writes.Run(f.c, writes.Op{Action: "set_preferred_likeness", UserID: userID}, func(tx *database.Tx) ([]rowchange.Change, []rowchange.Change, error) {
			ch, err := canonicalentities.SetPreferredLikeness(tx, user, h, obs)
			return ch, ch, err
		})
	}
	t.Run("choose, repeat, switch, clear", func(t *testing.T) {
		f := newFixture(t)
		h, a, b, _ := f.twoLikenesses()
		ch, res, err := set(f, h, b.ID, userID)
		must(t, err)
		if len(ch) != 1 {
			t.Fatalf("changes %+v", ch)
		}
		d := ch[0].Fields["preferred_likeness_observation_id"]
		if d.Old != nil || d.New != uuid.UUID(b.ID).String() {
			t.Fatalf("diff %+v", d)
		}
		if len(res.Effects.Handles) != 1 || !bytes.Equal(res.Effects.Handles[0], h) ||
			len(res.Effects.HandleScopes) != 1 || !bytes.Equal(res.Effects.HandleScopes[0], h) {
			t.Fatalf("effects %+v", res.Effects)
		}
		ch, res, err = set(f, h, b.ID, userID)
		must(t, err)
		if len(ch) != 0 || res.Revision != 0 {
			t.Fatalf("repeat wrote %+v rev %d", ch, res.Revision)
		}
		_, _, err = set(f, h, a.ID, userID)
		must(t, err)
		if !bytes.Equal(f.pointer(h), a.ID) {
			t.Fatalf("pointer %x", f.pointer(h))
		}
		ch, _, err = set(f, h, nil, userID)
		must(t, err)
		if f.pointer(h) != nil || len(ch) != 1 || ch[0].Fields["preferred_likeness_observation_id"].New != nil {
			t.Fatalf("clear %+v pointer %x", ch, f.pointer(h))
		}
	})
	t.Run("refused", func(t *testing.T) {
		f := newFixture(t)
		h, _, b, _ := f.twoLikenesses()
		other := f.source("Diary")
		so := f.personOn(other)
		o := observationstest.SaveLikeness(t, f.c, userID, other.ID, so.ID, 150)
		f.promote(so, nil)
		album := f.source("Album")
		sn := f.personOn(album)
		n := observationstest.SaveNegativeLikeness(t, f.c, userID, album.ID, sn.ID, 100)
		f.promote(sn, h)
		fresh, err := uuid.NewV7()
		must(t, err)
		unknown, err := uuid.NewV7()
		must(t, err)
		f.setPreferred(h, b.ID)
		cases := []struct {
			name      string
			h, obs, u []byte
		}{
			{"another Person's likeness", h, o.ID, userID},
			{"a negative of the same picture", h, n.ID, userID},
			{"no such Observation", h, fresh[:], userID},
			{"unknown entity", unknown[:], b.ID, userID},
			{"nil user", h, b.ID, nil},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				_, _, err := set(f, tc.h, tc.obs, tc.u)
				if !errors.Is(err, canonicalentities.ErrInvalid) {
					t.Fatalf("err %v", err)
				}
				if !bytes.Equal(f.pointer(h), b.ID) {
					t.Fatalf("pointer changed %x", f.pointer(h))
				}
			})
		}
	})
}
