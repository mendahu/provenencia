package conclusionsummaries_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/mendahu/provenencia/core/catalogsession/catalogsessiontest"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/conclusionsummaries"
	"github.com/mendahu/provenencia/core/database/conclusionsummaries/conclusionsummariestest"
	"github.com/mendahu/provenencia/core/database/databasetest"
	"github.com/mendahu/provenencia/core/database/graphcache/graphcachetest"
	"github.com/mendahu/provenencia/core/database/nameformats"
	"github.com/mendahu/provenencia/core/hops"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// placesWorld is a Person Pat born at an event, and the places the birth is
// held at, each with the Location that files it.
type placesWorld struct {
	c      *database.Catalog
	k      *graphcachetest.Canon
	pat    []byte
	birth  []byte
	places [][]byte
}

// newPlacesWorld files Pat's birth held at n places, one Location each, and
// commits it all.
func newPlacesWorld(t *testing.T, n int) *placesWorld {
	t.Helper()
	c := graphcachetest.OpenCatalog(t)
	if err := nameformats.Warm(c); err != nil {
		t.Fatal(err)
	}
	w := &placesWorld{c: c, k: graphcachetest.NewCanon(t, c)}
	w.pat = w.k.Person()
	w.birth, _ = w.k.Birth(w.pat, nil, nil)
	ids := [][]byte{w.pat, w.birth}
	w.k.Bulk(func() {
		for i := 0; i < n; i++ {
			place := w.k.Entity("place")
			loc := w.k.Location(w.birth, place)
			w.places = append(w.places, place)
			ids = append(ids, place, loc)
		}
	})
	w.k.Commit(ids...)
	return w
}

// location files the birth at place once more.
func (w *placesWorld) location(place []byte) []byte {
	loc := w.k.Location(w.birth, place)
	w.k.Commit(loc, w.birth)
	return loc
}

func (w *placesWorld) event(t *testing.T) conclusionsummaries.EventSummary {
	t.Helper()
	got, err := conclusionsummaries.EventsByIDs(w.c, [][]byte{w.birth})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("EventsByIDs returned %d summaries, want 1", len(got))
	}
	return got[0]
}

func (w *placesWorld) owners(t *testing.T) conclusionsummaries.LifeOwners {
	t.Helper()
	rows, err := conclusionsummaries.PersonsByIDs(w.c, [][]byte{w.pat})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("PersonsByIDs returned %d rows, want 1", len(rows))
	}
	return rows[0].Birth.Owners
}

// firstPlaces is the first 20 distinct places of the birth's location links, in link order.
func (w *placesWorld) firstPlaces(t *testing.T) [][]byte {
	t.Helper()
	links, err := w.c.Graph().Follow(w.birth, hops.PlacesOfEvent)
	if err != nil {
		t.Fatal(err)
	}
	var out [][]byte
	for _, l := range links {
		if len(out) < 20 && !contains(out, l.Neighbor) {
			out = append(out, l.Neighbor)
		}
	}
	return out
}

func placeIDs(ps []conclusionsummaries.ChainedPlace) [][]byte {
	var out [][]byte
	for _, p := range ps {
		out = append(out, p.Entity.ID)
	}
	return out
}

// TestCraftedEventPlaces: an event with 4,000 places summarizes in under a
// second and names the first 20 (S16-01 AC-1.3).
func TestCraftedEventPlaces(t *testing.T) {
	const most = 20
	t.Run("table", func(t *testing.T) {
		for _, tc := range []struct {
			name        string
			places      int
			repeatFirst int // extra Locations to the first place
			wantPlaces  int
			wantOwners  int
		}{
			{"20 places", 20, 0, 20, 20},
			{"21 places", 21, 0, 20, 20},
			{"one place filed 21 times", 1, 20, 1, 21},
		} {
			t.Run(tc.name, func(t *testing.T) {
				w := newPlacesWorld(t, tc.places)
				for i := 0; i < tc.repeatFirst; i++ {
					w.location(w.places[0])
				}
				got := w.event(t)
				if len(got.Places) != tc.wantPlaces {
					t.Fatalf("the event holds %d places, want %d", len(got.Places), tc.wantPlaces)
				}
				if n := len(w.owners(t).Places); n != tc.wantOwners {
					t.Fatalf("the birth lists %d place owners, want %d", n, tc.wantOwners)
				}
				first := w.firstPlaces(t)
				for _, id := range placeIDs(got.Places) {
					if !contains(first, id) {
						t.Fatalf("place %x is not among the first %d in link order", id, most)
					}
				}
			})
		}
	})
	t.Run("4,000 places", func(t *testing.T) {
		w := newPlacesWorld(t, 4000)
		var ev conclusionsummaries.EventSummary
		var owners conclusionsummaries.LifeOwners
		var want conclusionsummaries.Want
		want.Add(vocabkey.Event, w.birth)
		want.Add(vocabkey.Person, w.pat)
		cold := func(what string, read func()) {
			t.Helper()
			d := catalogsessiontest.Cold(t, w.c, 2, read)
			databasetest.WithinBudget(t, what+" at 4,000 places", d, time.Second)
		}
		cold("EventsByIDs", func() { ev = w.event(t) })
		cold("PersonsByIDs owners", func() { owners = w.owners(t) })
		var rows conclusionsummaries.Summaries
		cold("RowsForSearch", func() {
			var err error
			if rows, err = conclusionsummaries.RowsForSearch(w.c.Graph(), want); err != nil {
				t.Fatal(err)
			}
		})
		if len(ev.Places) != most || len(owners.Places) != most {
			t.Fatalf("the event holds %d places and the birth lists %d owners, want %d each", len(ev.Places), len(owners.Places), most)
		}
		first := w.firstPlaces(t)
		for _, id := range placeIDs(ev.Places) {
			if !contains(first, id) {
				t.Fatalf("place %x is not among the first %d in link order", id, most)
			}
		}
		row, ok := rows.Event(w.birth)
		if !ok || len(row.Places) > most {
			t.Fatalf("the search row names %d places (found %v), want at most %d", len(row.Places), ok, most)
		}
	})
}

func contains(ids [][]byte, id []byte) bool {
	for _, x := range ids {
		if bytes.Equal(x, id) {
			return true
		}
	}
	return false
}

// TestEventPlacesCapInputs: the batch and the isolated composition read the
// same places past the cap, and a write to a kept place's link recomposes the
// row (S16-01 AC-1.5).
func TestEventPlacesCapInputs(t *testing.T) {
	w := newPlacesWorld(t, 25)
	db, err := w.c.DB()
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string][][]byte{vocabkey.Event: {w.birth}, vocabkey.Person: {w.pat}}
	conclusionsummariestest.CheckMatchIsolated(t, db, ids)

	before := w.event(t)
	if len(before.Places) != 20 {
		t.Fatalf("the event holds %d places, want 20", len(before.Places))
	}
	// Delete the first place's location link: the next place moves in.
	var firstLink []byte
	if err := db.QueryRow(`SELECT a.id FROM observations o JOIN subjects a ON a.id = o.subject_id
		WHERE o.value_subject_id = ? LIMIT 1`, w.places[0]).Scan(&firstLink); err != nil {
		t.Fatal(err)
	}
	w.k.Remove(firstLink)
	w.k.Commit(w.birth, firstLink)
	after := w.event(t)
	if len(after.Places) != 20 {
		t.Fatalf("after the removal the event holds %d places, want 20", len(after.Places))
	}
	if contains(placeIDs(after.Places), w.places[0]) {
		t.Fatal("the removed place is still named")
	}
	if !contains(placeIDs(after.Places), w.places[20]) {
		t.Fatal("the 21st place did not move in")
	}
}
