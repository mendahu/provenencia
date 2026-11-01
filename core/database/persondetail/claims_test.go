package persondetail_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/artifacts/artifactstest"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/conclusiondetails"
	"github.com/mendahu/provenencia/core/database/evidencegraphs"
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/persondetail"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/promote/promotetest"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjects/subjectstest"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/namevalue"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"

	_ "github.com/mendahu/provenencia/core/database/graphfill"
)

var userID = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

const locator = `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`

type fixture struct {
	t      *testing.T
	c      *database.Catalog
	typeID []byte
	name   properties.Property
	arts   map[string]artifacts.Artifact
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
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book"})
	must(t, err)
	name, err := properties.Lookup(c, "name", vocabkey.OriginProvenencia)
	must(t, err)
	return &fixture{t: t, c: c, typeID: typeID, name: name, arts: map[string]artifacts.Artifact{}}
}

func (f *fixture) source(title string) sources.Source {
	f.t.Helper()
	src, _, err := writes.Run(f.c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: f.typeID, Title: title})
	})
	must(f.t, err)
	art, err := artifactstest.Create(f.c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
	must(f.t, err)
	f.arts[string(src.ID)] = art
	return src
}

func (f *fixture) person(graphID []byte, label string) subjects.Subject {
	f.t.Helper()
	return subjectstest.On(f.t, f.c, userID, graphID, "person", label)
}

// cite files the Observations on the Subject's graph through the Source's Artifact.
func (f *fixture) cite(s subjects.Subject, in ...observations.Input) []observations.Observation {
	f.t.Helper()
	res, err := writes.Call(f.c, writes.Op{Action: "create_citation_with_observations", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
		return citations.CreateWithObservations(tx, userID, citations.CreateInput{GraphID: s.GraphID, ArtifactID: f.arts[string(s.SourceID)].ID, LocatorJSON: locator}, in)
	})
	must(f.t, err)
	return res.Observations
}

func (f *fixture) nameOf(s subjects.Subject, form string) observations.Input {
	return observations.Input{SubjectID: s.ID, PropertyID: f.name.ID, Name: &namevalue.Value{Form: form, Parts: []namevalue.Part{{Idx: 0, Type: "given", Value: form}}}}
}

// promote files an accepted claim for s onto onto, or onto a new handle.
func (f *fixture) promote(s subjects.Subject, onto []byte) canonicalentities.Entity {
	f.t.Helper()
	return promotetest.Save(f.t, f.c, userID, promote.Input{SubjectID: s.ID, EntityID: onto}).Entity
}

// provisional files a provisional claim for s on e.
func (f *fixture) provisional(s subjects.Subject, e canonicalentities.Entity) {
	f.t.Helper()
	_, err := writes.Call(f.c, writes.Op{Action: "create_identity_claim", UserID: userID}, func(tx *database.Tx) (identityclaims.Claim, []rowchange.Change, error) {
		return identityclaims.Create(tx, userID, identityclaims.CreateInput{SubjectID: s.ID, EntityID: e.ID, Status: identityclaims.StatusProvisional})
	})
	must(f.t, err)
}

func (f *fixture) claims(id []byte) []persondetail.ClaimGroup {
	f.t.Helper()
	groups, err := persondetail.Claims(f.c, id)
	must(f.t, err)
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
	f.t.Helper()
	src := f.source("Parish register")
	burials := evidencegraphstest.SoleGraph(f.t, f.c, src.ID)
	_, _, err := writes.Run(f.c, writes.Op{Action: "update_evidence_graph", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
		ch, err := evidencegraphs.Update(tx, userID, burials.ID, evidencegraphs.UpdateInput{Name: "Burials"})
		return struct{}{}, ch, err
	})
	must(f.t, err)
	burials.Name = "Burials"
	baptisms := evidencegraphstest.Create(f.t, f.c, userID, src.ID, "Baptisms", f.arts[string(src.ID)].ID)
	third := evidencegraphstest.Create(f.t, f.c, userID, src.ID, "Marriages", f.arts[string(src.ID)].ID)

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
	if !bytes.Equal(g.SourceID, s.src.ID) || g.SourceRef != s.src.Ref || g.Title != "Parish register" || !bytes.Equal(g.SourceTypeID, f.typeID) {
		t.Fatalf("group %+v", g)
	}
	if len(g.Claims) != 2 {
		t.Fatalf("%d claims, want 2 (the provisional one is not returned)", len(g.Claims))
	}
	baptisms, burials := g.Claims[0], g.Claims[1]
	if baptisms.Graph.GraphName != "Baptisms" || burials.Graph.GraphName != "Burials" {
		t.Fatalf("claims ordered %q, %q", baptisms.Graph.GraphName, burials.Graph.GraphName)
	}
	if !bytes.Equal(burials.Graph.GraphID, s.burials.ID) || burials.Graph.GraphRef != s.burials.Ref || !bytes.Equal(burials.Subject.ID, s.burialsSubject.ID) || len(burials.ClaimID) != 16 {
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

func TestClaimsOrderSources(t *testing.T) {
	f := newFixture(t)
	var p canonicalentities.Entity
	titles := []string{"Zion census", "Abbey roll", "Écluse ledger", "Abbey roll"}
	refs := map[string]string{}
	for i, title := range titles {
		src := f.source(title)
		s := f.person(evidencegraphstest.SoleGraph(t, f.c, src.ID).ID, "")
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
		got = append(got, g.Title)
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
	if groups[0].SourceRef >= groups[1].SourceRef {
		t.Fatalf("the two Abbey rolls are ordered %q, %q, want by ref", groups[0].SourceRef, groups[1].SourceRef)
	}
}

func TestClaimsMissingAndEmpty(t *testing.T) {
	f := newFixture(t)
	src := f.source("Register")
	graph := evidencegraphstest.SoleGraph(t, f.c, src.ID).ID
	merged := f.promote(f.person(graph, "A"), nil)
	into := f.promote(f.person(graph, "B"), nil)
	provisionalOnly := f.promote(f.person(graph, "C"), nil)
	db, err := f.c.DB()
	must(t, err)
	_, err = db.Exec(`UPDATE canonical_entities SET merged_into_id = ? WHERE id = ?`, into.ID, merged.ID)
	must(t, err)
	f.c.Graph().Drop()
	// A handle whose only claim is provisional: a second handle with a claim
	// moved to provisional status.
	_, err = db.Exec(`UPDATE identity_claims SET status = ? WHERE entity_id = ?`, identityclaims.StatusProvisional, provisionalOnly.ID)
	must(t, err)
	f.c.Graph().Drop()

	tests := []struct {
		name    string
		id      []byte
		wantErr error
		empty   bool
	}{
		{"unknown id", make([]byte, 16), conclusiondetails.ErrNotFound, false},
		{"id that does not parse", []byte{1, 2}, conclusiondetails.ErrNotFound, false},
		{"merged handle", merged.ID, conclusiondetails.ErrNotFound, false},
		{"only a provisional member", provisionalOnly.ID, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groups, err := persondetail.Claims(f.c, tt.id)
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
	f.c.Graph().SetQueryCounterForTest(func() { n++ })
	second := f.claims(s.person.ID)
	if n != 0 {
		t.Fatalf("the warm read issued %d queries", n)
	}
	if len(second) != len(first) {
		t.Fatalf("second read %d groups, first %d", len(second), len(first))
	}
}
