package evidencelayer_test

import (
	"crypto/rand"
	"errors"
	"slices"
	"testing"

	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database/evidencelayer"
	"github.com/mendahu/provenencia/core/database/evidencelayer/evidencelayertest"
	"github.com/mendahu/provenencia/core/database/valuerules"
	"github.com/mendahu/provenencia/core/kinship"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// links is the Evidence layer's links of src's graph, built under the
// catalog's name pattern.
func links(t *testing.T, r *evidencelayertest.Records, graph []byte) []kinship.Link {
	t.Helper()
	db, err := r.C.DB()
	evidencelayertest.Must(t, err)
	_, namePattern, err := valuerules.Load(r.C.Graph(), db)
	evidencelayertest.Must(t, err)
	layer, err := evidencelayer.Load(r.C.Graph(), graph, namePattern)
	evidencelayertest.Must(t, err)
	return layer.Links
}

// The loader reads a Source's links from the connect rules: one link for
// each registered bridge type, whatever its term.
func TestLoadCoversEveryConnectRule(t *testing.T) {
	r := evidencelayertest.New(t)
	ann := r.Person(r.Source, r.Artifact, "Ann Ames")
	bob := r.Person(r.Source, r.Artifact, "Bob Ames")
	birth := r.Birth(r.Source, r.Artifact)
	york := r.Place(r.Source, r.Artifact, "York")
	ontario := r.Place(r.Source, r.Artifact, "Ontario")
	r.Participation(r.Source, r.Artifact, ann, birth, "subject")
	r.Locate(r.Source, r.Artifact, birth, york)
	r.Relationship(r.Source, r.Artifact, ann, bob, "parent")
	r.PlaceRel(r.Source, r.Artifact, york, ontario, "part_of")

	var got []string
	for _, l := range links(t, r, r.Graph(r.Source)) {
		got = append(got, string(l.KeyA.Type))
	}
	want := connectrules.BridgeTypeKeys()
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Fatalf("bridge types %v, want %v", got, want)
	}
}

// A Source that records a spouse relationship and a marriage has one spouse
// link, with every association behind it; a marriage alone has an implied one.
func TestLoadMergesKinshipShapes(t *testing.T) {
	spouses := func(ls []kinship.Link) []kinship.Link {
		var out []kinship.Link
		for _, l := range ls {
			if l.KeyA.Type == vocabkey.BridgeRelationship && l.KeyA.Term == "spouse" {
				out = append(out, l)
			}
		}
		return out
	}
	build := func(withRelationship bool) []kinship.Link {
		r := evidencelayertest.New(t)
		ann := r.Person(r.Source, r.Artifact, "Ann Ames")
		bob := r.Person(r.Source, r.Artifact, "Bob Ames")
		r.Marriage(r.Source, r.Artifact, ann, bob, "subject")
		if withRelationship {
			r.Relationship(r.Source, r.Artifact, ann, bob, "spouse")
		}
		return links(t, r, r.Graph(r.Source))
	}
	both := spouses(build(true))
	if len(both) != 1 || len(both[0].Associations) != 3 || len(both[0].Shapes) != 1 || both[0].OnlyImplied {
		t.Fatalf("recorded and implied: %+v, want one recorded link with 3 associations and one shape", both)
	}
	only := spouses(build(false))
	if len(only) != 1 || len(only[0].Associations) != 2 || !only[0].OnlyImplied {
		t.Fatalf("marriage only: %+v, want one implied link with 2 associations", only)
	}
}

// A graph the cache does not find is ErrNoGraph, and nothing is memoized for it.
func TestLoadUnknownGraph(t *testing.T) {
	r := evidencelayertest.New(t)
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if _, err := evidencelayer.Load(r.C.Graph(), id, ""); !errors.Is(err, evidencelayer.ErrNoGraph) {
			t.Fatalf("call %d: err %v, want ErrNoGraph", i+1, err)
		}
	}
	if _, ok := r.C.Graph().EvidenceGraphMemo(id, "evidencelayer\x00"); ok {
		t.Fatal("a memo was stored for an unknown graph")
	}
}
