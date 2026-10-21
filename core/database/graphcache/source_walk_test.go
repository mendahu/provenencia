package graphcache_test

import (
	"bytes"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/connect"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjectpositions"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/hops"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"

	_ "github.com/mendahu/provenencia/core/database/sourcefill"
)

var walkUser = []byte{9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9}

const walkLocator = `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`

type walkFixture struct {
	t    *testing.T
	c    *database.Catalog
	src  sources.Source
	art  artifacts.Artifact
	y    int64
	name map[string]subjects.Subject
}

func newWalkFixture(t *testing.T) *walkFixture {
	t.Helper()
	c := openCatalog(t)
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book"})
	if err != nil {
		t.Fatal(err)
	}
	f := &walkFixture{t: t, c: c, name: map[string]subjects.Subject{}}
	f.src, f.art = f.newSource(typeID, "Register")
	return f
}

func (f *walkFixture) newSource(typeID []byte, title string) (sources.Source, artifacts.Artifact) {
	f.t.Helper()
	src, _, err := writes.Run(f.c, writes.Op{Action: "create_source", UserID: walkUser}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, walkUser, sources.CreateInput{SourceTypeID: typeID, Title: title})
	})
	if err != nil {
		f.t.Fatal(err)
	}
	art, _, err := writes.Run(f.c, writes.Op{Action: "create_artifact", UserID: walkUser}, func(tx *database.Tx) (artifacts.Artifact, []rowchange.Change, error) {
		return artifacts.Create(tx, walkUser, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
	})
	if err != nil {
		f.t.Fatal(err)
	}
	return src, art
}

func (f *walkFixture) subject(src sources.Source, name, kind string) subjects.Subject {
	f.t.Helper()
	st, err := subjecttypes.Lookup(f.c, kind, vocabkey.OriginProvenencia)
	if err != nil {
		f.t.Fatal(err)
	}
	s, err := writes.Call(f.c, writes.Op{Action: "create_subject", UserID: walkUser}, func(tx *database.Tx) (subjects.Subject, []rowchange.Change, error) {
		return subjects.Create(tx, walkUser, subjects.CreateInput{SourceID: src.ID, SubjectTypeID: st.ID}, nil)
	})
	if err != nil {
		f.t.Fatal(err)
	}
	if _, err := writes.Call(f.c, writes.Op{}, func(tx *database.Tx) (subjectpositions.Position, []rowchange.Change, error) {
		return subjectpositions.Set(tx, s.ID, 0, f.y)
	}); err != nil {
		f.t.Fatal(err)
	}
	f.y += 2
	f.name[name] = s
	return s
}

func (f *walkFixture) prop(key string) []byte {
	f.t.Helper()
	p, err := properties.Lookup(f.c, key, vocabkey.OriginProvenencia)
	if err != nil {
		f.t.Fatal(err)
	}
	return p.ID
}

func (f *walkFixture) bridge(kind string, from, to subjects.Subject, obs ...observations.Input) subjects.Subject {
	f.t.Helper()
	res, err := writes.Call(f.c, writes.Op{Action: "create_cited_bridge", UserID: walkUser}, func(tx *database.Tx) (connect.Result, []rowchange.Change, error) {
		return connect.CreateCitedBridge(tx, walkUser, connect.CreateInput{
			SourceID: f.src.ID, FromSubjectID: from.ID, ToSubjectID: to.ID, BridgeTypeKey: kind,
			Citation:     citations.CreateInput{ArtifactID: f.art.ID, LocatorJSON: walkLocator},
			Observations: obs,
		})
	})
	if err != nil {
		f.t.Fatal(err)
	}
	return res.Subject
}

func (f *walkFixture) participation(person, event subjects.Subject, role string) subjects.Subject {
	f.t.Helper()
	term, err := propertyterms.Lookup(f.c, f.prop("role"), role, vocabkey.OriginProvenencia)
	if err != nil {
		f.t.Fatal(err)
	}
	return f.bridge("participation", person, event,
		observations.Input{PropertyID: f.prop("person"), ValueSubjectID: person.ID},
		observations.Input{PropertyID: f.prop("event"), ValueSubjectID: event.ID},
		observations.Input{PropertyID: f.prop("role"), ValueTermID: term.ID},
	)
}

func (f *walkFixture) location(event, place subjects.Subject) subjects.Subject {
	f.t.Helper()
	return f.bridge("location", event, place,
		observations.Input{PropertyID: f.prop("event"), ValueSubjectID: event.ID},
		observations.Input{PropertyID: f.prop("place"), ValueSubjectID: place.ID},
	)
}

func (f *walkFixture) exec(query string, args ...any) {
	f.t.Helper()
	db, err := f.c.DB()
	if err != nil {
		f.t.Fatal(err)
	}
	if _, err := db.Exec(query, args...); err != nil {
		f.t.Fatal(err)
	}
}

// walkCase is one hop and its expected edges, as (from, bridge, to) names.
type walkCase struct {
	name string
	hop  hops.Hop
	from []string
	want [][3]string
}

// TestWalkSourceMatchesSQL walks the Evidence graph of one Source in memory.
// Its expected edges are literals: the edges canonicalgraph.WalkSource's SQL
// returned for this fixture before it was deleted (checked equal, edge for edge
// and in order, when the in-memory walk landed).
func TestWalkSourceMatchesSQL(t *testing.T) {
	f := newWalkFixture(t)
	james, mary := f.subject(f.src, "james", "person"), f.subject(f.src, "mary", "person")
	infant, witness := f.subject(f.src, "infant", "person"), f.subject(f.src, "witness", "person")
	birth, wedding := f.subject(f.src, "birth", "event"), f.subject(f.src, "wedding", "event")
	york, hull := f.subject(f.src, "york", "place"), f.subject(f.src, "hull", "place")

	f.name["j-birth"] = f.participation(james, birth, "subject")
	f.name["j-wedding"] = f.participation(james, wedding, "witness")
	f.name["m-wedding"] = f.participation(mary, wedding, "subject")
	f.name["w-wedding"] = f.participation(witness, wedding, "witness")
	f.name["i-birth"] = f.participation(infant, birth, "subject")
	f.name["birth-york"] = f.location(birth, york)
	f.name["wedding-hull"] = f.location(wedding, hull)

	// A negative from-endpoint: the infant's participation no longer names them.
	f.exec(`UPDATE observations SET polarity = 'negative'
		WHERE subject_id = ? AND property_id = ?`, f.name["i-birth"].ID, f.prop("person"))
	// A filter term recorded negatively: the witness' role no longer counts.
	f.exec(`UPDATE observations SET polarity = 'negative'
		WHERE subject_id = ? AND property_id = ?`, f.name["w-wedding"].ID, f.prop("role"))
	// A to-end on another Source: Hull's location now names a Subject elsewhere.
	typeID, err := sourcetypes.Upsert(f.c, sourcetypes.Type{Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book"})
	if err != nil {
		t.Fatal(err)
	}
	other, _ := f.newSource(typeID, "Elsewhere")
	away := f.subject(other, "away", "place")
	f.exec(`UPDATE observations SET value_subject_id = ? WHERE subject_id = ? AND property_id = ?`,
		away.ID, f.name["wedding-hull"].ID, f.prop("place"))

	participates := hops.MustHop("participation", "person", "event", nil)
	asSubject := hops.MustHop("participation", "person", "event", &hops.TermFilter{PropertyKey: "role", TermKey: "subject"})
	asWitness := hops.MustHop("participation", "person", "event", &hops.TermFilter{PropertyKey: "role", TermKey: "witness"})

	tests := []walkCase{
		{name: "every participation", hop: participates, from: []string{"james", "mary", "witness"}, want: [][3]string{
			{"james", "j-birth", "birth"}, {"james", "j-wedding", "wedding"},
			{"mary", "m-wedding", "wedding"}, {"witness", "w-wedding", "wedding"},
		}},
		{name: "a negative from-endpoint is not walked", hop: participates, from: []string{"infant"}},
		{name: "subject role only", hop: asSubject, from: []string{"james", "mary", "witness"}, want: [][3]string{
			{"james", "j-birth", "birth"}, {"mary", "m-wedding", "wedding"},
		}},
		{name: "a negative filter term is not read", hop: asWitness, from: []string{"james", "witness"}, want: [][3]string{
			{"james", "j-wedding", "wedding"},
		}},
		{name: "an event's people", hop: hops.SubjectsOfEvent, from: []string{"birth", "wedding"}, want: [][3]string{
			{"birth", "j-birth", "james"}, {"wedding", "m-wedding", "mary"},
		}},
		{name: "a location", hop: hops.PlacesOfEvent, from: []string{"birth", "wedding"}, want: [][3]string{
			{"birth", "birth-york", "york"},
		}},
		{name: "a place's events", hop: hops.EventsAtPlace, from: []string{"york", "hull"}, want: [][3]string{
			{"york", "birth-york", "birth"},
		}},
		{name: "no from Subjects", hop: participates},
	}
	g := f.c.Graph()
	sg, err := g.Source(f.src.ID)
	if err != nil {
		t.Fatal(err)
	}
	ids := func(names []string) [][]byte {
		var out [][]byte
		for _, n := range names {
			out = append(out, f.name[n].ID)
		}
		return out
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := g.WalkSource(sg, tt.hop, ids(tt.from))
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %d edges, want %d: %+v", len(got), len(tt.want), got)
			}
			for _, w := range tt.want {
				found := false
				for _, e := range got {
					if bytes.Equal(e.From, f.name[w[0]].ID) && bytes.Equal(e.Association, f.name[w[1]].ID) &&
						bytes.Equal(e.To, f.name[w[2]].ID) && e.AssociationRef == f.name[w[1]].Ref && e.ToRef == f.name[w[2]].Ref {
						found = true
					}
				}
				if !found {
					t.Fatalf("no edge %v in %+v", w, got)
				}
			}
		})
	}
}

// A Source's graph is walked alone: another Source's Subjects reach nothing.
func TestWalkSourceStaysOnItsSource(t *testing.T) {
	f := newWalkFixture(t)
	james, birth := f.subject(f.src, "james", "person"), f.subject(f.src, "birth", "event")
	f.participation(james, birth, "subject")
	typeID, err := sourcetypes.Upsert(f.c, sourcetypes.Type{Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book"})
	if err != nil {
		t.Fatal(err)
	}
	other, _ := f.newSource(typeID, "Elsewhere")
	g := f.c.Graph()
	sg, err := g.Source(other.ID)
	if err != nil {
		t.Fatal(err)
	}
	edges, err := g.WalkSource(sg, hops.EventsOfSubject, [][]byte{james.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(edges) != 0 {
		t.Fatalf("walked %d edges", len(edges))
	}
	if _, err := g.WalkSource(sg, hops.Hop{}, nil); err == nil {
		t.Fatal("a zero Hop was walked")
	}
}

// T-22: a loaded handle's members carry the Source of each claimed Subject,
// and Verify compares it.
func TestMemberSourceID(t *testing.T) {
	f := newWalkFixture(t)
	james := f.subject(f.src, "james", "person")
	res, err := writes.Call(f.c, writes.Op{Action: "promote_subject", UserID: walkUser}, func(tx *database.Tx) (promote.Result, []rowchange.Change, error) {
		return promote.Save(tx, walkUser, promote.Input{SubjectID: james.ID})
	})
	if err != nil {
		t.Fatal(err)
	}
	g := f.c.Graph()
	g.Drop()
	n, err := g.Node(res.Entity.ID)
	if err != nil || n == nil || len(n.Members) != 1 {
		t.Fatalf("node %+v, err %v", n, err)
	}
	if !bytes.Equal(n.Members[0].SubjectID, james.ID) || !bytes.Equal(n.Members[0].SourceID, f.src.ID) {
		t.Fatalf("member %+v, want Subject %x on Source %x", n.Members[0], james.ID, f.src.ID)
	}
	if err := g.Verify(); err != nil {
		t.Fatal(err)
	}
	g.CorruptMemberSourceForTest(res.Entity.ID)
	if err := g.Verify(); err == nil {
		t.Fatal("Verify accepted a member whose Source does not match the catalog")
	}
}
