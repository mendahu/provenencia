package graphcache_test

import (
	"bytes"
	"testing"

	"github.com/mendahu/provenencia/core/database/artifacts/artifactstest"
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
	"github.com/mendahu/provenencia/core/database/evidencelayer/evidencelayertest"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/graphfill"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/promote/promotetest"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjects/subjectstest"
	"github.com/mendahu/provenencia/core/hops"
	"github.com/mendahu/provenencia/core/vocabkey"
)

var walkUser = evidencelayertest.UserID

type walkFixture struct {
	*evidencelayertest.Records
	name map[string]subjects.Subject
}

func newWalkFixture(t *testing.T) *walkFixture {
	t.Helper()
	return &walkFixture{Records: evidencelayertest.New(t), name: map[string]subjects.Subject{}}
}

// subject is a Subject of the kind on src, found later under name.
func (f *walkFixture) subject(src sources.Source, name, kind string) subjects.Subject {
	f.T.Helper()
	s := f.Bare(src, kind)
	f.name[name] = s
	return s
}

func (f *walkFixture) exec(query string, args ...any) {
	f.T.Helper()
	db, err := f.C.DB()
	if err != nil {
		f.T.Fatal(err)
	}
	if _, err := db.Exec(query, args...); err != nil {
		f.T.Fatal(err)
	}
}

// walkCase is one hop and its expected edges, as (from, bridge, to) names.
type walkCase struct {
	name string
	hop  hops.Hop
	from []string
	want [][3]string
}

// TestWalkEvidenceGraphEdges walks one Evidence graph in memory.
// Its expected edges are literals: the edges the SQL walk S11-19 deleted
// returned for this fixture before it was deleted (checked equal, edge for edge
// and in order, when the in-memory walk landed).
func TestWalkEvidenceGraphEdges(t *testing.T) {
	f := newWalkFixture(t)
	james, mary := f.subject(f.Source, "james", "person"), f.subject(f.Source, "mary", "person")
	infant, witness := f.subject(f.Source, "infant", "person"), f.subject(f.Source, "witness", "person")
	birth, wedding := f.subject(f.Source, "birth", "event"), f.subject(f.Source, "wedding", "event")
	york, hull := f.subject(f.Source, "york", "place"), f.subject(f.Source, "hull", "place")

	f.name["j-birth"] = f.Participation(f.Source, f.Artifact, james, birth, "subject")
	f.name["j-wedding"] = f.Participation(f.Source, f.Artifact, james, wedding, "witness")
	f.name["m-wedding"] = f.Participation(f.Source, f.Artifact, mary, wedding, "subject")
	f.name["w-wedding"] = f.Participation(f.Source, f.Artifact, witness, wedding, "witness")
	f.name["i-birth"] = f.Participation(f.Source, f.Artifact, infant, birth, "subject")
	f.name["birth-york"] = f.Locate(f.Source, f.Artifact, birth, york)
	f.name["wedding-hull"] = f.Locate(f.Source, f.Artifact, wedding, hull)

	// A negative from-endpoint: the infant's participation no longer names them.
	f.exec(`UPDATE observations SET polarity = 'negative'
		WHERE subject_id = ? AND property_id = ?`, f.name["i-birth"].ID, f.Prop("person").ID)
	// A filter term recorded negatively: the witness' role no longer counts.
	f.exec(`UPDATE observations SET polarity = 'negative'
		WHERE subject_id = ? AND property_id = ?`, f.name["w-wedding"].ID, f.Prop("role").ID)
	// A to-end on another Source: Hull's location now names a Subject elsewhere.
	other, _ := f.NewSource("Elsewhere")
	away := f.subject(other, "away", "place")
	f.exec(`UPDATE observations SET value_subject_id = ? WHERE subject_id = ? AND property_id = ?`,
		away.ID, f.name["wedding-hull"].ID, f.Prop("place").ID)

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
	g := f.C.Graph()
	eg, err := graphfill.EvidenceGraph(g, evidencegraphstest.SoleGraph(t, f.C, f.Source.ID).ID)
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
			got, err := g.WalkEvidenceGraph(eg, tt.hop, ids(tt.from))
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

// TestWalkEvidenceGraphEdgeCap: one bridge citing many Subjects on each end
// builds edges as their product, so a walk over the cap reaches nothing.
func TestWalkEvidenceGraphEdgeCap(t *testing.T) {
	f := newWalkFixture(t)
	g := f.C.Graph()
	v, err := g.Vocabulary()
	if err != nil {
		t.Fatal(err)
	}
	var bridgeType []byte
	for _, ty := range v.Types {
		if ty.Key == "participation" && ty.Origin == vocabkey.OriginProvenencia {
			bridgeType = ty.ID
		}
	}
	id := func(n int) []byte {
		b := make([]byte, 16)
		b[0], b[1] = byte(n>>8), byte(n)
		return b
	}
	run := func(t *testing.T, froms, tos int) int {
		bridge := id(60000)
		eg := graphcache.EvidenceGraph{Found: true, Subjects: []graphcache.EvidenceSubject{{ID: bridge, SubjectTypeID: bridgeType, Ref: "SUB-B"}}}
		var from [][]byte
		for i := 0; i < tos; i++ {
			eg.Subjects = append(eg.Subjects, graphcache.EvidenceSubject{ID: id(i + 1), Ref: "SUB-P"})
		}
		for i := 0; i < froms; i++ {
			from = append(from, id(i+1))
			eg.Observations = append(eg.Observations, graphcache.EvidenceObservation{
				SubjectID: bridge, PropertyID: f.Prop("person").ID, ValueSubjectID: id(i + 1), Polarity: "positive"})
		}
		for i := 0; i < tos; i++ {
			eg.Observations = append(eg.Observations, graphcache.EvidenceObservation{
				SubjectID: bridge, PropertyID: f.Prop("event").ID, ValueSubjectID: id(i + 1), Polarity: "positive"})
		}
		edges, err := g.WalkEvidenceGraph(eg, hops.MustHop("participation", "person", "event", nil), from)
		if err != nil {
			t.Fatal(err)
		}
		return len(edges)
	}
	if got := run(t, 223, 224); got != 223*224 {
		t.Fatalf("under the cap: got %d edges, want %d", got, 223*224)
	}
	if got := run(t, 224, 224); got != 0 {
		t.Fatalf("over the cap: got %d edges, want none", got)
	}
}

// An Evidence graph is walked alone: another graph's Subjects reach nothing.
func TestWalkEvidenceGraphStaysOnItsGraph(t *testing.T) {
	f := newWalkFixture(t)
	james, birth := f.subject(f.Source, "james", "person"), f.subject(f.Source, "birth", "event")
	f.Participation(f.Source, f.Artifact, james, birth, "subject")
	other, _ := f.NewSource("Elsewhere")
	g := f.C.Graph()
	eg, err := graphfill.EvidenceGraph(g, evidencegraphstest.SoleGraph(t, f.C, other.ID).ID)
	if err != nil {
		t.Fatal(err)
	}
	edges, err := g.WalkEvidenceGraph(eg, hops.EventsOfSubject, [][]byte{james.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(edges) != 0 {
		t.Fatalf("walked %d edges", len(edges))
	}
	if _, err := g.WalkEvidenceGraph(eg, hops.Hop{}, nil); err == nil {
		t.Fatal("a zero Hop was walked")
	}
}

// T-22: a loaded handle's members carry the Source and the Evidence graph of
// each claimed Subject, and Verify compares them.
func TestMemberSourceAndGraph(t *testing.T) {
	check := func(t *testing.T, f *walkFixture, james subjects.Subject, wantGraph []byte) {
		t.Helper()
		res := promotetest.Save(t, f.C, walkUser, promote.Input{SubjectID: james.ID})
		g := f.C.Graph()
		g.Drop()
		n, err := g.Node(res.Entity.ID)
		if err != nil || n == nil || len(n.Members) != 1 {
			t.Fatalf("node %+v, err %v", n, err)
		}
		m := n.Members[0]
		if !bytes.Equal(m.SubjectID, james.ID) || !bytes.Equal(m.SourceID, f.Source.ID) || !bytes.Equal(m.GraphID, wantGraph) {
			t.Fatalf("member %+v, want Subject %x on Source %x, graph %x", m, james.ID, f.Source.ID, wantGraph)
		}
		if err := g.Verify(); err != nil {
			t.Fatal(err)
		}
		g.CorruptMemberSourceForTest(res.Entity.ID)
		if err := g.Verify(); err == nil {
			t.Fatal("Verify accepted a member whose Source does not match the catalog")
		}
	}
	t.Run("the Default graph", func(t *testing.T) {
		f := newWalkFixture(t)
		james := f.subject(f.Source, "james", "person")
		check(t, f, james, evidencegraphstest.SoleGraph(t, f.C, f.Source.ID).ID)
	})
	t.Run("a Split graph", func(t *testing.T) {
		f := newWalkFixture(t)
		_, jones, _ := artifactstest.Split(t, f.C, walkUser, f.Source.ID, "Jones")
		james := subjectstest.On(t, f.C, walkUser, jones.ID, "person", "")
		check(t, f, james, jones.ID)
	})
}
