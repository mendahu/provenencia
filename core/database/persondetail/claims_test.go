package persondetail_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/conclusiondetails"
	"github.com/mendahu/provenencia/core/database/evidencegraphs"
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
	"github.com/mendahu/provenencia/core/database/evidencelayer/evidencelayertest"
	"github.com/mendahu/provenencia/core/database/graphfill"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/observations/observationstest"
	"github.com/mendahu/provenencia/core/database/persondetail"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/promote/promotetest"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjects/subjectstest"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/namevalue"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
)

var userID = evidencelayertest.UserID

const locator = evidencelayertest.Locator

type fixture struct {
	*evidencelayertest.Records
	name properties.Property
	arts map[string]artifacts.Artifact
}

func must(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	r := evidencelayertest.New(t)
	return &fixture{Records: r, name: r.Prop("name"), arts: map[string]artifacts.Artifact{}}
}

// source is another Source, with its Artifact kept for cite.
func (f *fixture) source(title string) sources.Source {
	f.T.Helper()
	src, art := f.NewSource(title)
	f.arts[string(src.ID)] = art
	return src
}

func (f *fixture) person(graphID []byte, label string) subjects.Subject {
	f.T.Helper()
	return subjectstest.On(f.T, f.C, userID, graphID, "person", label)
}

// cite files the Observations on the Subject's graph through the Source's Artifact.
func (f *fixture) cite(s subjects.Subject, in ...observations.Input) []observations.Observation {
	f.T.Helper()
	res, _, err := observationstest.Cite(f.C, userID, s.GraphID, f.arts[string(s.SourceID)].ID, locator, in...)
	must(f.T, err)
	return res.Observations
}

func (f *fixture) nameOf(s subjects.Subject, form string) observations.Input {
	return observations.Input{SubjectID: s.ID, PropertyID: f.name.ID, Name: &namevalue.Value{Form: form, Parts: []namevalue.Part{{Idx: 0, Type: "given", Value: form}}}}
}

// promote files an accepted claim for s onto onto, or onto a new handle.
func (f *fixture) promote(s subjects.Subject, onto []byte) canonicalentities.Entity {
	f.T.Helper()
	return promotetest.Save(f.T, f.C, userID, promote.Input{SubjectID: s.ID, EntityID: onto}).Entity
}

// provisional files a provisional claim for s on e.
func (f *fixture) provisional(s subjects.Subject, e canonicalentities.Entity) {
	f.T.Helper()
	_, err := writes.Call(f.C, writes.Op{Action: "create_identity_claim", UserID: userID}, func(tx *database.Tx) (identityclaims.Claim, []rowchange.Change, error) {
		return identityclaims.Create(tx, userID, identityclaims.CreateInput{SubjectID: s.ID, EntityID: e.ID, Status: identityclaims.StatusProvisional})
	})
	must(f.T, err)
}

func (f *fixture) claims(id []byte) []persondetail.ClaimGroup {
	f.T.Helper()
	groups, err := persondetail.Claims(f.C, id)
	must(f.T, err)
	return groups
}

// burialsAndBaptisms is the shape of AC-1.1: one Source with graphs "Burials"
// (renamed from its sole graph) and "Baptisms", a Subject on each accepted
// into one Person, and a third graph whose Subject is claimed provisionally.
type shape struct {
	src                            sources.Source
	burials, baptisms              evidencegraphs.Graph
	burialsSubject, baptismSubject subjects.Subject
	person                         canonicalentities.Entity
}

func (f *fixture) burialsAndBaptisms() shape {
	f.T.Helper()
	src := f.source("Parish register")
	burials := evidencegraphstest.SoleGraph(f.T, f.C, src.ID)
	_, _, err := writes.Run(f.C, writes.Op{Action: "update_evidence_graph", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
		ch, err := evidencegraphs.Update(tx, userID, burials.ID, evidencegraphs.UpdateInput{Name: "Burials"})
		return struct{}{}, ch, err
	})
	must(f.T, err)
	burials.Name = "Burials"
	baptisms := evidencegraphstest.Create(f.T, f.C, userID, src.ID, "Baptisms", f.arts[string(src.ID)].ID)
	third := evidencegraphstest.Create(f.T, f.C, userID, src.ID, "Marriages", f.arts[string(src.ID)].ID)

	bs := f.person(burials.ID, "Ann")
	f.cite(bs, f.nameOf(bs, "Ann Lee"), f.nameOf(bs, "Annie Lee"))
	ps := f.person(baptisms.ID, "Baby")
	p := f.promote(bs, nil)
	f.promote(ps, p.ID)
	ms := f.person(third.ID, "Wife")
	f.provisional(ms, p)
	return shape{src: src, burials: burials, baptisms: baptisms, burialsSubject: bs, baptismSubject: ps, person: p}
}

func TestClaimsGroupsBySource(t *testing.T) {
	f := newFixture(t)
	s := f.burialsAndBaptisms()
	groups := f.claims(s.person.ID)
	if len(groups) != 1 {
		t.Fatalf("%d groups, want 1", len(groups))
	}
	g := groups[0]
	if !bytes.Equal(g.Source.ID, s.src.ID) || g.Source.Ref != s.src.Ref || g.Source.Title != "Parish register" || !bytes.Equal(g.Source.SourceTypeID, f.Source.SourceTypeID) {
		t.Fatalf("group %+v", g)
	}
	if len(g.Claims) != 2 {
		t.Fatalf("%d claims, want 2 (the provisional one is not returned)", len(g.Claims))
	}
	baptisms, burials := g.Claims[0], g.Claims[1]
	if baptisms.Graph.Name != "Baptisms" || burials.Graph.Name != "Burials" {
		t.Fatalf("claims ordered %q, %q", baptisms.Graph.Name, burials.Graph.Name)
	}
	if !bytes.Equal(burials.Graph.ID, s.burials.ID) || burials.Graph.Ref != s.burials.Ref || !bytes.Equal(burials.Subject.ID, s.burialsSubject.ID) || len(burials.Accepted.ClaimID) != 16 {
		t.Fatalf("burials claim %+v", burials)
	}
	if len(burials.Observations) != 2 {
		t.Fatalf("burials Observations %d, want 2", len(burials.Observations))
	}
	for _, o := range burials.Observations {
		if !bytes.Equal(o.SubjectID, s.burialsSubject.ID) {
			t.Fatalf("Observation of another Subject: %+v", o)
		}
	}
	if !bytes.Equal(baptisms.Subject.ID, s.baptismSubject.ID) || len(baptisms.Observations) != 0 {
		t.Fatalf("baptisms claim %+v", baptisms)
	}
}

// TestClaimsAfterSourceEdit: a Source rename and a type change reach the claims
// group, and a loaded graph of the Source stays loaded.
func TestClaimsAfterSourceEdit(t *testing.T) {
	f := newFixture(t)
	s := f.burialsAndBaptisms()
	if got := f.claims(s.person.ID); got[0].Source.Title != "Parish register" {
		t.Fatalf("title before %q", got[0].Source.Title)
	}
	letter, err := sourcetypes.Upsert(f.C, sourcetypes.Type{Key: "letter", Origin: vocabkey.OriginProvenencia, Label: "Letter"})
	must(t, err)
	src := s.src
	src.Title = "Registry of parish"
	src.SourceTypeID = letter
	f.Run("update_source", func(tx *database.Tx) ([]rowchange.Change, error) {
		return sources.Update(tx, userID, src)
	})
	g := f.claims(s.person.ID)[0]
	if g.Source.Title != "Registry of parish" || !bytes.Equal(g.Source.SourceTypeID, letter) {
		t.Fatalf("group after the edit %q %x", g.Source.Title, g.Source.SourceTypeID)
	}
	if rl := f.recordLinks(s.person.ID); rl.Graphs[0].Source.Title != "Registry of parish" {
		t.Fatalf("record links title after the edit %q", rl.Graphs[0].Source.Title)
	}
	if _, ok := graphfill.Graphs.Peek(f.C.Graph(), s.burials.ID); !ok {
		t.Fatal("a Source edit dropped a loaded Evidence graph")
	}
}

func TestClaimsOrderSources(t *testing.T) {
	f := newFixture(t)
	var p canonicalentities.Entity
	titles := []string{"Zion census", "Abbey roll", "Écluse ledger", "Abbey roll"}
	refs := map[string]string{}
	for i, title := range titles {
		src := f.source(title)
		s := f.person(evidencegraphstest.SoleGraph(t, f.C, src.ID).ID, "")
		if i == 0 {
			p = f.promote(s, nil)
		} else {
			f.promote(s, p.ID)
		}
		refs[string(src.ID)] = src.Ref
	}
	groups := f.claims(p.ID)
	var got []string
	for _, g := range groups {
		got = append(got, g.Source.Title)
	}
	want := []string{"Abbey roll", "Abbey roll", "Écluse ledger", "Zion census"}
	if len(got) != len(want) {
		t.Fatalf("order %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order %v, want %v", got, want)
		}
	}
	if groups[0].Source.Ref >= groups[1].Source.Ref {
		t.Fatalf("the two Abbey rolls are ordered %q, %q, want by ref", groups[0].Source.Ref, groups[1].Source.Ref)
	}
}

func TestClaimsMissingAndEmpty(t *testing.T) {
	f := newFixture(t)
	src := f.source("Register")
	graph := evidencegraphstest.SoleGraph(t, f.C, src.ID).ID
	merged := f.promote(f.person(graph, "A"), nil)
	into := f.promote(f.person(graph, "B"), nil)
	provisionalOnly := f.promote(f.person(graph, "C"), nil)
	event := f.promote(f.Birth(src, f.arts[string(src.ID)]), nil)
	db, err := f.C.DB()
	must(t, err)
	_, err = db.Exec(`UPDATE canonical_entities SET merged_into_id = ? WHERE id = ?`, into.ID, merged.ID)
	must(t, err)
	f.C.Graph().Drop()
	// A handle whose only claim is provisional: a second handle with a claim
	// moved to provisional status.
	_, err = db.Exec(`UPDATE identity_claims SET status = ? WHERE entity_id = ?`, identityclaims.StatusProvisional, provisionalOnly.ID)
	must(t, err)
	f.C.Graph().Drop()

	tests := []struct {
		name    string
		id      []byte
		wantErr error
		empty   bool
	}{
		{"unknown id", make([]byte, 16), conclusiondetails.ErrNotFound, false},
		{"id that does not parse", []byte{1, 2}, conclusiondetails.ErrNotFound, false},
		{"merged handle", merged.ID, conclusiondetails.ErrNotFound, false},
		{"an Event handle", event.ID, conclusiondetails.ErrNotFound, false},
		{"only a provisional member", provisionalOnly.ID, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groups, err := persondetail.Claims(f.C, tt.id)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error %v, want %v", err, tt.wantErr)
			}
			if tt.empty && len(groups) != 0 {
				t.Fatalf("groups %+v, want none", groups)
			}
		})
	}
}

func TestClaimsWarmReadIssuesNoQuery(t *testing.T) {
	f := newFixture(t)
	s := f.burialsAndBaptisms()
	first := f.claims(s.person.ID)
	var n int
	f.C.Graph().SetQueryCounterForTest(func() { n++ })
	second := f.claims(s.person.ID)
	if n != 0 {
		t.Fatalf("the warm read issued %d queries", n)
	}
	if len(second) != len(first) {
		t.Fatalf("second read %d groups, first %d", len(second), len(first))
	}

	// A person in 80 graphs, past the 64 the loaded-graph store keeps: the
	// read loads no Evidence graph, and the second read issues no query.
	t.Run("80 graphs", func(t *testing.T) {
		f := newFixture(t)
		son := households(f, 80)
		g := f.C.Graph()
		g.Drop()
		first := f.claims(son)
		if len(first) != 80 {
			t.Fatalf("%d groups, want 80", len(first))
		}
		for _, grp := range first {
			if len(grp.Claims) != 1 {
				t.Fatalf("group %q has %d claims, want 1", grp.Source.Title, len(grp.Claims))
			}
		}
		if n := g.KeyCounts()["graphfill.graphs"]; n != 0 {
			t.Fatalf("the read loaded %d Evidence graphs", n)
		}
		n := 0
		g.SetQueryCounterForTest(func() { n++ })
		defer g.SetQueryCounterForTest(nil)
		second := f.claims(son)
		if n != 0 {
			t.Fatalf("the second read issued %d queries", n)
		}
		if len(second) != 80 {
			t.Fatalf("second read %d groups, want 80", len(second))
		}
	})
}

// TestClaimsSkipsWhatLeft: a claim whose graph's records no longer hold its
// Subject (the graph changed mid-read) is skipped, and the person's other claims
// stay.
func TestClaimsSkipsWhatLeft(t *testing.T) {
	cases := []struct {
		name string
		// records is what graph A's records become: empty records of another
		// graph, or graph B's.
		records func(t *testing.T, f *fixture, s shape) *graphfill.GraphRecords
	}{
		{"an empty graph's records", func(t *testing.T, f *fixture, s shape) *graphfill.GraphRecords {
			empty := evidencegraphstest.SoleGraph(t, f.C, f.source("Empty").ID).ID
			got, err := graphfill.RecordsOf(f.C.Graph(), [][]byte{empty})
			must(t, err)
			return got[string(empty)]
		}},
		{"records without the Subject", func(t *testing.T, f *fixture, s shape) *graphfill.GraphRecords {
			got, err := graphfill.RecordsOf(f.C.Graph(), [][]byte{s.baptisms.ID})
			must(t, err)
			return got[string(s.baptisms.ID)]
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			s := f.burialsAndBaptisms()
			if got := f.claims(s.person.ID); len(got) != 1 || len(got[0].Claims) != 2 {
				t.Fatalf("the warm read: %+v", got)
			}
			graphfill.Records.Set(f.C.Graph(), s.burials.ID, tc.records(t, f, s))
			groups := f.claims(s.person.ID)
			if len(groups) != 1 || len(groups[0].Claims) != 1 {
				t.Fatalf("groups %+v, want one group of the one claim that stayed", groups)
			}
			if !bytes.Equal(groups[0].Claims[0].Subject.ID, s.baptismSubject.ID) {
				t.Fatalf("the claim that stayed is %x, want the baptisms Subject", groups[0].Claims[0].Subject.ID)
			}
		})
	}
}

// TestClaimsRecordsPerGraph: a Person's records in one Evidence graph are kept
// to the registry's recordlinks MaxEntityMembers, the first ones in member order.
func TestClaimsRecordsPerGraph(t *testing.T) {
	most := matchregistry.Default().RecordLinks.MaxEntityMembers
	for _, tc := range []struct {
		name   string
		extra  int // records beside the Person's own
		want   int
		prefix bool
	}{
		{"at the cap", most - 1, most, false},
		{"one over", most, most, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := newKinWorld(t)
			pat := w.person("Pat")
			w.k.Members(pat, tc.extra)
			w.k.Commit(pat)
			groups, err := persondetail.Claims(w.c, pat)
			if err != nil {
				t.Fatal(err)
			}
			if len(groups) != 1 || len(groups[0].Claims) != tc.want {
				t.Fatalf("%d groups, %d claims in the first, want 1 and %d", len(groups), len(groups[0].Claims), tc.want)
			}
			if tc.prefix {
				n, err := w.c.Graph().Node(pat)
				if err != nil {
					t.Fatal(err)
				}
				for i, m := range n.AcceptedMembers()[:tc.want] {
					found := false
					for _, cl := range groups[0].Claims {
						found = found || bytes.Equal(cl.Subject.ID, m.SubjectID)
					}
					if !found {
						t.Fatalf("member %d is not among the claims", i)
					}
				}
			}
		})
	}
}
