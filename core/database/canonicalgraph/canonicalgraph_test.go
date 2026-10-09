package canonicalgraph_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/canonicalgraph"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/connect"
	"github.com/mendahu/provenencia/core/database/evrun"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/writes"
)

var userID = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

const locator = `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`

var subjectRole = &canonicalgraph.TermFilter{PropertyKey: "role", TermKey: "subject"}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewHop(t *testing.T) {
	tests := []struct {
		name     string
		bridge   string
		from, to string
		filter   *canonicalgraph.TermFilter
		wantErr  bool
	}{
		{name: "participation person to event", bridge: "participation", from: "person", to: "event"},
		{name: "participation filtered by role", bridge: "participation", from: "event", to: "person", filter: subjectRole},
		{name: "location event to place", bridge: "location", from: "event", to: "place"},
		{name: "relationship between two people", bridge: "relationship", from: "person", to: "related_to"},
		{name: "unknown bridge", bridge: "kinship", from: "person", to: "event", wantErr: true},
		{name: "not an endpoint of the bridge", bridge: "location", from: "person", to: "place", wantErr: true},
		{name: "an endpoint to itself", bridge: "location", from: "place", to: "place", wantErr: true},
		{name: "filter on a property that does not disambiguate", bridge: "location", from: "event", to: "place", filter: subjectRole, wantErr: true},
		{name: "filter with no term", bridge: "participation", from: "person", to: "event", filter: &canonicalgraph.TermFilter{PropertyKey: "role"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := canonicalgraph.NewHop(tt.bridge, tt.from, tt.to, tt.filter)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type fixture struct {
	t   *testing.T
	c   *database.Catalog
	src sources.Source
	art artifacts.Artifact
	y   int64
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
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{Key: "book", Origin: sourcetypes.OriginProvenencia, Label: "Book"})
	must(t, err)
	src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Register"})
	})
	must(t, err)
	art, err := runArtifactCreate(c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
	must(t, err)
	return &fixture{t: t, c: c, src: src, art: art}
}

func (f *fixture) subject(kind string) subjects.Subject {
	f.t.Helper()
	st, err := subjecttypes.Lookup(f.c, kind, subjecttypes.OriginProvenencia)
	must(f.t, err)
	s, err := evrun.CreateSubject(f.c, userID, subjects.CreateInput{SourceID: f.src.ID, SubjectTypeID: st.ID}, nil)
	must(f.t, err)
	_, err = evrun.SetPosition(f.c, s.ID, 0, f.y)
	must(f.t, err)
	f.y += 2
	return s
}

func (f *fixture) prop(key string) []byte {
	f.t.Helper()
	p, err := properties.Lookup(f.c, key, properties.OriginProvenencia)
	must(f.t, err)
	return p.ID
}

func (f *fixture) bridge(kind string, from, to subjects.Subject, obs ...observations.Input) {
	f.t.Helper()
	_, err := evrun.CreateBridge(f.c, userID, connect.CreateInput{
		SourceID: f.src.ID, FromSubjectID: from.ID, ToSubjectID: to.ID, BridgeTypeKey: kind,
		Citation:     citations.CreateInput{ArtifactID: f.art.ID, LocatorJSON: locator},
		Observations: obs,
	})
	must(f.t, err)
}

func (f *fixture) participation(person, event subjects.Subject, role string) {
	f.t.Helper()
	term, err := propertyterms.Lookup(f.c, f.prop("role"), role, propertyterms.OriginProvenencia)
	must(f.t, err)
	f.bridge("participation", person, event,
		observations.Input{PropertyID: f.prop("person"), ValueSubjectID: person.ID},
		observations.Input{PropertyID: f.prop("event"), ValueSubjectID: event.ID},
		observations.Input{PropertyID: f.prop("role"), ValueTermID: term.ID},
	)
}

func (f *fixture) promote(s subjects.Subject) []byte {
	f.t.Helper()
	res, err := promote.Save(f.c, userID, promote.Input{SubjectID: s.ID})
	must(f.t, err)
	return res.Entity.ID
}

func TestWalk(t *testing.T) {
	f := newFixture(t)
	james, birth, wedding, york := f.subject("person"), f.subject("event"), f.subject("event"), f.subject("place")
	f.participation(james, birth, "subject")
	f.participation(james, wedding, "witness")
	f.bridge("location", birth, york,
		observations.Input{PropertyID: f.prop("event"), ValueSubjectID: birth.ID},
		observations.Input{PropertyID: f.prop("place"), ValueSubjectID: york.ID},
	)
	jamesID, birthID, weddingID, yorkID := f.promote(james), f.promote(birth), f.promote(wedding), f.promote(york)

	participates := canonicalgraph.MustHop("participation", "person", "event", nil)
	asSubject := canonicalgraph.MustHop("participation", "person", "event", subjectRole)
	located := canonicalgraph.MustHop("location", "event", "place", nil)

	tests := []struct {
		name string
		hop  canonicalgraph.Hop
		from [][]byte
		want [][]byte
	}{
		{name: "every participation", hop: participates, from: [][]byte{jamesID}, want: [][]byte{birthID, weddingID}},
		{name: "subject role only", hop: asSubject, from: [][]byte{jamesID}, want: [][]byte{birthID}},
		{name: "reversed, from the event", hop: asSubject.Reverse(), from: [][]byte{birthID}, want: [][]byte{jamesID}},
		{name: "a witness is not a subject", hop: asSubject.Reverse(), from: [][]byte{weddingID}},
		{name: "location", hop: located, from: [][]byte{birthID}, want: [][]byte{yorkID}},
		{name: "location reversed", hop: located.Reverse(), from: [][]byte{yorkID}, want: [][]byte{birthID}},
		{name: "a hop from the wrong kind finds nothing", hop: located, from: [][]byte{jamesID}},
		{name: "no handles, no edges", hop: participates},
	}
	db, err := f.c.DB()
	must(t, err)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edges, err := canonicalgraph.Walk(db, tt.hop, tt.from)
			must(t, err)
			if got := canonicalgraph.Targets(edges); !sameIDs(got, tt.want) {
				t.Fatalf("got %d targets, want %d", len(got), len(tt.want))
			}
			for _, e := range edges {
				if len(e.Association) != 16 || e.AssociationRef == "" || e.ToRef == "" {
					t.Fatalf("edge %+v", e)
				}
			}
		})
	}

	t.Run("a merged handle is not reached", func(t *testing.T) {
		if _, err := db.Exec(`UPDATE canonical_entities SET merged_into_id = ? WHERE id = ?`, birthID, weddingID); err != nil {
			t.Fatal(err)
		}
		edges, err := canonicalgraph.Walk(db, participates, [][]byte{jamesID})
		must(t, err)
		if got := canonicalgraph.Targets(edges); !sameIDs(got, [][]byte{birthID}) {
			t.Fatalf("reached %d events", len(got))
		}
	})

	t.Run("the walk reads indexes, not whole tables", func(t *testing.T) {
		for _, hop := range []canonicalgraph.Hop{participates, asSubject, located.Reverse()} {
			query, args := canonicalgraph.WalkSQL(hop, [][]byte{jamesID, birthID})
			rows, err := db.Query(`EXPLAIN QUERY PLAN `+query, args...)
			must(t, err)
			for rows.Next() {
				var id, parent, notUsed int
				var detail string
				must(t, rows.Scan(&id, &parent, &notUsed, &detail))
				if strings.HasPrefix(detail, "SCAN ") {
					_ = rows.Close()
					t.Fatalf("full scan: %s", detail)
				}
			}
			must(t, rows.Err())
			_ = rows.Close()
		}
	})
}

func sameIDs(got, want [][]byte) bool {
	if len(got) != len(want) {
		return false
	}
	for _, w := range want {
		found := false
		for _, g := range got {
			if bytes.Equal(g, w) {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func TestWalkSource(t *testing.T) {
	f := newFixture(t)
	james, birth, wedding, york := f.subject("person"), f.subject("event"), f.subject("event"), f.subject("place")
	f.participation(james, birth, "subject")
	f.participation(james, wedding, "witness")
	f.bridge("location", birth, york,
		observations.Input{PropertyID: f.prop("event"), ValueSubjectID: birth.ID},
		observations.Input{PropertyID: f.prop("place"), ValueSubjectID: york.ID},
	)
	participates := canonicalgraph.MustHop("participation", "person", "event", nil)

	tests := []struct {
		name string
		hop  canonicalgraph.Hop
		from []subjects.Subject
		want []subjects.Subject
	}{
		{name: "every participation", hop: participates, from: []subjects.Subject{james}, want: []subjects.Subject{birth, wedding}},
		{name: "subject role only", hop: canonicalgraph.EventsOfSubject, from: []subjects.Subject{james}, want: []subjects.Subject{birth}},
		{name: "an event's subjects", hop: canonicalgraph.SubjectsOfEvent, from: []subjects.Subject{birth, wedding}, want: []subjects.Subject{james}},
		{name: "a location", hop: canonicalgraph.PlacesOfEvent, from: []subjects.Subject{birth}, want: []subjects.Subject{york}},
		{name: "an event with no location", hop: canonicalgraph.PlacesOfEvent, from: []subjects.Subject{wedding}},
		{name: "a place's events", hop: canonicalgraph.EventsAtPlace, from: []subjects.Subject{york}, want: []subjects.Subject{birth}},
	}
	db, err := f.c.DB()
	must(t, err)
	ids := func(ss []subjects.Subject) [][]byte {
		var out [][]byte
		for _, s := range ss {
			out = append(out, s.ID)
		}
		return out
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edges, err := canonicalgraph.WalkSource(db, f.src.ID, tt.hop, ids(tt.from))
			must(t, err)
			if got := canonicalgraph.Targets(edges); !sameIDs(got, ids(tt.want)) {
				t.Fatalf("got %d targets, want %d", len(got), len(tt.want))
			}
		})
	}

	t.Run("another Source's graph is not walked", func(t *testing.T) {
		edges, err := canonicalgraph.WalkSource(db, make([]byte, 16), canonicalgraph.EventsOfSubject, [][]byte{james.ID})
		must(t, err)
		if len(edges) != 0 {
			t.Fatalf("walked %d edges", len(edges))
		}
	})

	t.Run("the walk reads indexes, not whole tables", func(t *testing.T) {
		query, args := canonicalgraph.WalkSourceSQL(canonicalgraph.EventsOfSubject, f.src.ID, [][]byte{james.ID})
		rows, err := db.Query(`EXPLAIN QUERY PLAN `+query, args...)
		must(t, err)
		defer rows.Close()
		for rows.Next() {
			var id, parent, notUsed int
			var detail string
			must(t, rows.Scan(&id, &parent, &notUsed, &detail))
			if strings.HasPrefix(detail, "SCAN ") {
				t.Fatalf("full scan: %s", detail)
			}
		}
		must(t, rows.Err())
	})
}

func runArtifactCreate(c *database.Catalog, userID []byte, in artifacts.CreateInput) (artifacts.Artifact, error) {
	a, _, err := writes.Run(c, writes.Op{Action: "create_artifact", UserID: userID},
		func(tx *database.Tx) (artifacts.Artifact, []rowchange.Change, error) {
			return artifacts.Create(tx, userID, in)
		})
	return a, err
}
