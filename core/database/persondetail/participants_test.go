package persondetail_test

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database/persondetail"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// roleTerm adds a researcher's own role term.
func (w *kinWorld) roleTerm(key, label string) {
	w.t.Helper()
	db, err := w.c.DB()
	if err != nil {
		w.t.Fatal(err)
	}
	var propID []byte
	if err := db.QueryRow(`SELECT id FROM properties WHERE key = 'role' AND origin = 'provenencia'`).Scan(&propID); err != nil {
		w.t.Fatal(err)
	}
	termID := []byte(fmt.Sprintf("%016s", key))[:16]
	if _, err := db.Exec(`INSERT INTO property_terms (id, property_id, key, origin, label) VALUES (?, ?, ?, 'user', ?)`, termID, propID, key, label); err != nil {
		w.t.Fatal(err)
	}
	w.k.CommitVocabulary()
}

// participantNames writes the participants of the one event of pat in order.
func (w *kinWorld) participantNames(pat []byte) string {
	w.t.Helper()
	rows, _ := w.events(pat)
	if len(rows) != 1 {
		w.t.Fatalf("%d rows", len(rows))
	}
	var out []string
	for _, p := range rows[0].Participants {
		out = append(out, w.name(p.ID))
	}
	return strings.Join(out, ",")
}

func TestParticipantOrder(t *testing.T) {
	t.Run("AC-4.1", func(t *testing.T) {
		w := newKinWorld(t)
		pat, kohl, anna, heinrich := w.person("Pat"), w.person("Kohl"), w.person("Anna"), w.person("Heinrich")
		e := w.event("Baptism", vocabkey.EventType("baptism"), 1820)
		w.k.Join(pat, e, "subject")
		w.k.Join(kohl, e, "witness")
		w.k.Join(anna, e, "mother")
		w.k.Join(heinrich, e, "father")
		if got := w.participantNames(pat); got != "Heinrich,Anna,Kohl" {
			t.Fatalf("got %s", got)
		}
	})
	t.Run("AC-4.2", func(t *testing.T) {
		w := newKinWorld(t)
		w.roleTerm("officiant", "Officiant")
		w.roleTerm("mourner", "Mourner")
		pat := w.person("Pat")
		zimmer, abel, off, mourn, bare := w.person("Zimmer"), w.person("Abel"), w.person("Off"), w.person("Mourn"), w.person("Bare")
		e := w.event("Burial", vocabkey.EventBurial, 1890)
		w.k.Join(pat, e, "subject")
		w.k.Link("participation", map[string][]byte{"person": bare, "event": e}, "", "")
		w.k.Join(off, e, "officiant")
		w.k.Join(zimmer, e, "pallbearer")
		w.k.Join(mourn, e, "mourner")
		w.k.Join(abel, e, "pallbearer")
		// People here are unnamed, so a rank's ties come in ref order.
		want := append(w.order(w.byRef(abel, zimmer)...), "Mourn", "Off", "Bare")
		if got := w.participantNames(pat); got != strings.Join(want, ",") {
			t.Fatalf("got %s, want %v", got, want)
		}
	})
	t.Run("AC-4.3", func(t *testing.T) {
		w := newKinWorld(t)
		pat, a, b := w.person("Pat"), w.person("Aaa"), w.person("Bbb")
		e := w.event("Census", vocabkey.EventType("census"), 1851)
		w.k.Join(pat, e, "subject")
		w.k.Join(a, e, "informant")
		w.k.Join(b, e, "informant")
		w.k.Join(b, e, "witness")
		if got := w.participantNames(pat); got != "Bbb,Aaa" {
			t.Fatalf("got %s", got)
		}
	})
	t.Run("unlisted roles by label, ignoring case", func(t *testing.T) {
		w := newKinWorld(t)
		w.roleTerm("officiant", "Officiant")
		w.roleTerm("mourner", "mourner")
		pat, off, mourn := w.person("Pat"), w.person("Off"), w.person("Mourn")
		e := w.event("Burial", vocabkey.EventBurial, 1890)
		w.k.Join(pat, e, "subject")
		w.k.Join(off, e, "officiant")
		w.k.Join(mourn, e, "mourner")
		if got := w.participantNames(pat); got != "Mourn,Off" {
			t.Fatalf("got %s", got)
		}
	})
	t.Run("equal rank keeps name order", func(t *testing.T) {
		w := newKinWorld(t)
		pat, z, a := w.person("Pat"), w.person("Zed"), w.person("Abe")
		e := w.event("Census", vocabkey.EventType("census"), 1851)
		w.k.Join(pat, e, "subject")
		w.k.Join(z, e, "witness")
		w.k.Join(a, e, "witness")
		if got, want := w.participantNames(pat), strings.Join(w.order(w.byRef(z, a)...), ","); got != want {
			t.Fatalf("got %s, want %s", got, want)
		}
	})
}

// A role the seed adds must be placed in participantRoleOrder.
func TestParticipantRolesCoverSeed(t *testing.T) {
	w := newKinWorld(t)
	db, err := w.c.DB()
	if err != nil {
		t.Fatal(err)
	}
	rows, err := db.Query(`SELECT t.key FROM property_terms t JOIN properties p ON p.id = t.property_id WHERE p.key = 'role' AND t.origin = 'provenencia'`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	order := persondetail.ParticipantRoleOrder()
	n := 0
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			t.Fatal(err)
		}
		n++
		if !slices.Contains(order, vocabkey.Role(key)) {
			t.Errorf("seeded role %q is not in participantRoleOrder", key)
		}
	}
	if n == 0 {
		t.Fatal("no seeded role terms")
	}
}
