package canonicalgraph_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database/canonicalgraph"
	"github.com/mendahu/provenencia/core/database/dbtest"
	"github.com/mendahu/provenencia/core/database/evidencelayer/evidencelayertest"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/promote/promotetest"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/hops"
)

var userID = evidencelayertest.UserID

var subjectRole = &hops.TermFilter{PropertyKey: "role", TermKey: "subject"}

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
		filter   *hops.TermFilter
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
		{name: "filter with no term", bridge: "participation", from: "person", to: "event", filter: &hops.TermFilter{PropertyKey: "role"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := hops.NewHop(tt.bridge, tt.from, tt.to, tt.filter)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type fixture struct {
	*evidencelayertest.Records
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	return &fixture{evidencelayertest.New(t)}
}

func (f *fixture) promote(s subjects.Subject) []byte {
	f.T.Helper()
	res := promotetest.Save(f.T, f.C, userID, promote.Input{SubjectID: s.ID})
	return res.Entity.ID
}

func TestWalk(t *testing.T) {
	f := newFixture(t)
	james, birth, wedding, york := f.Bare(f.Source, "person"), f.Bare(f.Source, "event"), f.Bare(f.Source, "event"), f.Bare(f.Source, "place")
	f.Participation(f.Source, f.Artifact, james, birth, "subject")
	f.Participation(f.Source, f.Artifact, james, wedding, "witness")
	f.Locate(f.Source, f.Artifact, birth, york)
	jamesID, birthID, weddingID, yorkID := f.promote(james), f.promote(birth), f.promote(wedding), f.promote(york)

	participates := hops.MustHop("participation", "person", "event", nil)
	asSubject := hops.MustHop("participation", "person", "event", subjectRole)
	located := hops.MustHop("location", "event", "place", nil)

	tests := []struct {
		name string
		hop  hops.Hop
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
	db, err := f.C.DB()
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
		for _, hop := range []hops.Hop{participates, asSubject, located.Reverse()} {
			query, args := canonicalgraph.WalkSQL(hop, [][]byte{jamesID, birthID})
			for _, detail := range dbtest.QueryPlan(t, db, query, args...) {
				if strings.HasPrefix(detail, "SCAN ") {
					t.Fatalf("full scan: %s", detail)
				}
			}
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
