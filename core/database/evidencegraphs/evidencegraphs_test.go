package evidencegraphs_test

import (
	"bytes"
	"errors"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/catalogmodel"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/deleteimpact"
	"github.com/mendahu/provenencia/core/database/evidencegraphs"
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/namevalue"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
)

var userID = []byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}

type fixture struct {
	c      *database.Catalog
	typeID []byte
}

func newFixture(t *testing.T) fixture {
	t.Helper()
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	r, err := ref.Mint(ref.PrefixUser)
	if err != nil {
		t.Fatal(err)
	}
	if err := users.Upsert(c, userID, "Jake", r); err != nil {
		t.Fatal(err)
	}
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
		Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book",
	})
	if err != nil {
		t.Fatal(err)
	}
	return fixture{c: c, typeID: typeID}
}

// createSource creates a Source and returns it with the changes of its create.
func (f fixture) createSource(t *testing.T, title string) (sources.Source, []rowchange.Change) {
	t.Helper()
	var changes []rowchange.Change
	src, err := writes.Call(f.c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		s, ch, err := sources.Create(tx, userID, sources.CreateInput{SourceTypeID: f.typeID, Title: title})
		changes = ch
		return s, ch, err
	})
	if err != nil {
		t.Fatal(err)
	}
	return src, changes
}

type graphRow struct {
	id          []byte
	ref         string
	name        string
	description string
}

func graphRows(t *testing.T, c *database.Catalog, sourceID []byte) []graphRow {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	rows, err := db.Query(`SELECT id, COALESCE(ref, ''), name, COALESCE(description, '') FROM evidence_graphs WHERE source_id = ? ORDER BY id`, sourceID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out []graphRow
	for rows.Next() {
		var g graphRow
		if err := rows.Scan(&g.id, &g.ref, &g.name, &g.description); err != nil {
			t.Fatal(err)
		}
		out = append(out, g)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestCreateDefault(t *testing.T) {
	f := newFixture(t)
	src, changes := f.createSource(t, "Census 1881")

	graphs := graphRows(t, f.c, src.ID)
	if len(graphs) != 1 {
		t.Fatalf("graphs %d", len(graphs))
	}
	g := graphs[0]
	if g.name != evidencegraphs.DefaultName {
		t.Fatalf("name %q", g.name)
	}
	if !ref.Valid(g.ref) || !strings.HasPrefix(g.ref, "GPH-") {
		t.Fatalf("ref %q", g.ref)
	}
	if bytes.Equal(g.id, src.ID) {
		t.Fatal("new Default graph reuses the Source id")
	}
	if g.description != "" {
		t.Fatalf("description %q", g.description)
	}
	if len(changes) != 2 || changes[0].EntityType != "source" || changes[1].EntityType != "evidence_graph" {
		t.Fatalf("changes %+v", changes)
	}
	if got := changes[1].Fields["source_id"].New; got != uuid.UUID(src.ID).String() {
		t.Fatalf("graph change source_id %v", got)
	}
	if !bytes.Equal(changes[1].EntityID, g.id) {
		t.Fatal("graph change names another graph")
	}
}

func TestSoleGraphTx(t *testing.T) {
	f := newFixture(t)
	src, _ := f.createSource(t, "Census 1881")
	want := graphRows(t, f.c, src.ID)[0].id

	unknown := make([]byte, 16)
	unknown[0] = 9
	tests := []struct {
		name     string
		sourceID []byte
		want     []byte
		wantErr  error
	}{
		{name: "the Default graph", sourceID: src.ID, want: want},
		{name: "unknown Source", sourceID: unknown, wantErr: evidencegraphs.ErrInvalid},
		{name: "short id", sourceID: unknown[:15], wantErr: evidencegraphs.ErrInvalid},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := f.c.DB()
			if err != nil {
				t.Fatal(err)
			}
			tx, err := db.Begin()
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback()
			got, err := evidencegraphs.SoleGraphTx(tx, tt.sourceID)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("got %v want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Fatalf("got %x want %x", got, tt.want)
			}
		})
	}

	t.Run("two graphs is not ErrInvalid", func(t *testing.T) {
		db, err := f.c.DB()
		if err != nil {
			t.Fatal(err)
		}
		second, err := uuid.NewV7()
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`INSERT INTO evidence_graphs (id, source_id, name) VALUES (?, ?, 'Second')`, second[:], src.ID); err != nil {
			t.Fatal(err)
		}
		tx, err := db.Begin()
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback()
		_, err = evidencegraphs.SoleGraphTx(tx, src.ID)
		if err == nil || errors.Is(err, evidencegraphs.ErrInvalid) {
			t.Fatalf("got %v", err)
		}
	})
}

const pageLocator = `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`

// run runs fn as one write and returns the changes it produced.
func run(t *testing.T, c *database.Catalog, fn func(tx *database.Tx) ([]rowchange.Change, error)) ([]rowchange.Change, error) {
	t.Helper()
	var out []rowchange.Change
	_, err := writes.Call(c, writes.Op{Action: "test_write", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
		ch, err := fn(tx)
		out = ch
		return struct{}{}, ch, err
	})
	return out, err
}

// newArtifact creates an Artifact and clears whatever links its create made,
// so each test starts from an Artifact on no graph and links it itself.
func (f fixture) newArtifact(t *testing.T, src sources.Source, label string) []byte {
	t.Helper()
	a, err := writes.Call(f.c, writes.Op{Action: "create_artifact", UserID: userID}, func(tx *database.Tx) (artifacts.Artifact, []rowchange.Change, error) {
		return artifacts.Create(tx, userID, artifacts.CreateInput{SourceID: src.ID, Label: label})
	})
	if err != nil {
		t.Fatal(err)
	}
	db, err := f.c.DB()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`DELETE FROM evidence_graph_artifacts WHERE artifact_id = ?`, a.ID); err != nil {
		t.Fatal(err)
	}
	return a.ID
}

func (f fixture) link(t *testing.T, graphID, artifactID []byte) {
	t.Helper()
	if _, err := run(t, f.c, func(tx *database.Tx) ([]rowchange.Change, error) {
		return evidencegraphs.Link(tx, userID, graphID, artifactID)
	}); err != nil {
		t.Fatal(err)
	}
}

func linkedArtifacts(t *testing.T, c *database.Catalog, graphID []byte) int {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM evidence_graph_artifacts WHERE graph_id = ?`, graphID).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// subjectAndCitation creates a person Subject on src's Default graph and a
// Citation on the Artifact with an Observation of that Subject.
func (f fixture) subjectAndCitation(t *testing.T, src sources.Source, artifactID []byte) (subjects.Subject, citations.Citation) {
	t.Helper()
	if err := subjectvocab.Install(f.c); err != nil {
		t.Fatal(err)
	}
	person, err := subjecttypes.Lookup(f.c, "person", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	nameProp, err := properties.Lookup(f.c, "name", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	subject, err := writes.Call(f.c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (subjects.Subject, []rowchange.Change, error) {
		return subjects.Create(tx, userID, subjects.CreateInput{
			SourceID: src.ID, SubjectTypeID: person.ID, Label: "Ada",
		}, nil)
	})
	if err != nil {
		t.Fatal(err)
	}
	res, err := writes.Call(f.c, writes.Op{Action: "create_citation", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
		return citations.CreateWithObservations(tx, userID, citations.CreateInput{
			ArtifactID: artifactID, LocatorJSON: pageLocator,
		}, []observations.Input{
			{SubjectID: subject.ID, PropertyID: nameProp.ID, Name: &namevalue.Value{Form: "Ada"}},
		})
	})
	if err != nil {
		t.Fatal(err)
	}
	return subject, res.Citation
}

func TestLink(t *testing.T) {
	f := newFixture(t)
	src, _ := f.createSource(t, "Census 1881")
	other, _ := f.createSource(t, "Parish register")
	graph := evidencegraphstest.SoleGraph(t, f.c, src.ID).ID
	a := f.newArtifact(t, src, "Page 1")
	foreign := f.newArtifact(t, other, "Page 9")
	unknown := make([]byte, 16)
	unknown[0] = 9

	t.Run("links once", func(t *testing.T) {
		ch, err := run(t, f.c, func(tx *database.Tx) ([]rowchange.Change, error) {
			return evidencegraphs.Link(tx, userID, graph, a)
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(ch) != 1 || ch[0].EntityType != "evidence_graph_artifact" || ch[0].Action != rowchange.ActionCreate {
			t.Fatalf("changes %+v", ch)
		}
		if !bytes.Equal(ch[0].EntityID, graph) {
			t.Fatal("link change is not audited under its graph")
		}
		if linkedArtifacts(t, f.c, graph) != 1 {
			t.Fatal("link row missing")
		}
	})
	t.Run("already linked is no change", func(t *testing.T) {
		ch, err := run(t, f.c, func(tx *database.Tx) ([]rowchange.Change, error) {
			return evidencegraphs.Link(tx, userID, graph, a)
		})
		if err != nil || len(ch) != 0 {
			t.Fatalf("got %v %+v", err, ch)
		}
	})
	refused := []struct {
		name     string
		userID   []byte
		graph    []byte
		artifact []byte
	}{
		{name: "Artifact of another Source", userID: userID, graph: graph, artifact: foreign},
		{name: "unknown graph", userID: userID, graph: unknown, artifact: a},
		{name: "unknown Artifact", userID: userID, graph: graph, artifact: unknown},
		{name: "empty user id", userID: nil, graph: graph, artifact: a},
		{name: "short graph id", userID: userID, graph: graph[:15], artifact: a},
	}
	for _, tt := range refused {
		t.Run(tt.name, func(t *testing.T) {
			_, err := run(t, f.c, func(tx *database.Tx) ([]rowchange.Change, error) {
				return evidencegraphs.Link(tx, tt.userID, tt.graph, tt.artifact)
			})
			if !errors.Is(err, evidencegraphs.ErrInvalid) {
				t.Fatalf("got %v", err)
			}
		})
	}
	t.Run("LinkArtifact refuses a repeated or short id", func(t *testing.T) {
		b := f.newArtifact(t, src, "Page 2")
		for name, ids := range map[string][][]byte{
			"repeated": {graph, graph},
			"short":    {graph[:15]},
		} {
			_, err := run(t, f.c, func(tx *database.Tx) ([]rowchange.Change, error) {
				return evidencegraphs.LinkArtifact(tx, userID, src.ID, b, ids)
			})
			if !errors.Is(err, evidencegraphs.ErrInvalid) {
				t.Fatalf("%s: got %v", name, err)
			}
		}
	})
}

func TestUnlink(t *testing.T) {
	f := newFixture(t)
	src, _ := f.createSource(t, "Census 1881")
	graph := evidencegraphstest.SoleGraph(t, f.c, src.ID).ID
	a1 := f.newArtifact(t, src, "Page 1")
	a2 := f.newArtifact(t, src, "Page 2")
	f.link(t, graph, a1)
	f.link(t, graph, a2)

	unlink := func(artifact []byte) ([]rowchange.Change, error) {
		return run(t, f.c, func(tx *database.Tx) ([]rowchange.Change, error) {
			return evidencegraphs.Unlink(tx, userID, graph, artifact)
		})
	}

	ch, err := unlink(a1)
	if err != nil {
		t.Fatal(err)
	}
	if len(ch) != 1 || ch[0].Action != rowchange.ActionDelete || ch[0].EntityType != "evidence_graph_artifact" {
		t.Fatalf("changes %+v", ch)
	}
	if linkedArtifacts(t, f.c, graph) != 1 {
		t.Fatal("link row not removed")
	}
	if _, err := unlink(a2); !errors.Is(err, evidencegraphs.ErrLastArtifact) {
		t.Fatalf("last link: %v", err)
	}
	if ch, err := unlink(a1); err != nil || len(ch) != 0 {
		t.Fatalf("not linked: %v %+v", err, ch)
	}

	f.link(t, graph, a1)
	f.subjectAndCitation(t, src, a2)
	_, err = unlink(a2)
	if !errors.Is(err, evidencegraphs.ErrArtifactInUse) {
		t.Fatalf("in use: %v", err)
	}
	var ae *apperr.Error
	if !errors.As(err, &ae) || len(ae.Params()) != 0 {
		t.Fatalf("refusal carries params: %v", err)
	}
	if linkedArtifacts(t, f.c, graph) != 2 {
		t.Fatal("refused unlink dropped the link")
	}
}

func TestMembership(t *testing.T) {
	f := newFixture(t)
	src, _ := f.createSource(t, "Census 1881")
	graph := evidencegraphstest.SoleGraph(t, f.c, src.ID).ID
	a := f.newArtifact(t, src, "Page 1")
	b := f.newArtifact(t, src, "Page 2")
	f.link(t, graph, a)
	f.link(t, graph, b)
	subject, cit := f.subjectAndCitation(t, src, a)
	// A second Citation on the same Artifact with no Observation belongs to no graph.
	if _, err := writes.Call(f.c, writes.Op{Action: "create_citation", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
		return citations.CreateWithObservations(tx, userID, citations.CreateInput{ArtifactID: a, LocatorJSON: pageLocator}, nil)
	}); err != nil {
		t.Fatal(err)
	}

	db, err := f.c.DB()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	if got, err := evidencegraphs.SubjectGraphTx(tx, subject.ID); err != nil || !bytes.Equal(got, graph) {
		t.Fatalf("SubjectGraphTx %x %v", got, err)
	}
	unknown := make([]byte, 16)
	unknown[0] = 9
	if _, err := evidencegraphs.SubjectGraphTx(tx, unknown); !errors.Is(err, evidencegraphs.ErrInvalid) {
		t.Fatalf("unknown Subject: %v", err)
	}
	if ok, err := evidencegraphs.IsLinkedTx(tx, graph, a); err != nil || !ok {
		t.Fatalf("IsLinkedTx %v %v", ok, err)
	}
	if ok, err := evidencegraphs.IsLinkedTx(tx, graph, unknown); err != nil || ok {
		t.Fatalf("IsLinkedTx unlinked %v %v", ok, err)
	}
	ids, err := evidencegraphs.CitationsUsingArtifactTx(tx, graph, a)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 || !bytes.Equal(ids[0], cit.ID) {
		t.Fatalf("citations %x want %x", ids, cit.ID)
	}
	if ids, err := evidencegraphs.CitationsUsingArtifactTx(tx, graph, b); err != nil || len(ids) != 0 {
		t.Fatalf("other Artifact: %v %x", err, ids)
	}
	if _, err := evidencegraphs.CitationsUsingArtifactTx(tx, graph[:15], a); !errors.Is(err, evidencegraphs.ErrInvalid) {
		t.Fatalf("short id: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}

	f.link(t, graph, b)
	if _, err := run(t, f.c, func(tx *database.Tx) ([]rowchange.Change, error) {
		return evidencegraphs.Unlink(tx, userID, graph, b)
	}); err != nil {
		t.Fatal(err)
	}
	tx2, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx2.Rollback()
	if ok, err := evidencegraphs.IsLinkedTx(tx2, graph, b); err != nil || ok {
		t.Fatalf("after Unlink %v %v", ok, err)
	}
}

func TestEnsureRefs(t *testing.T) {
	t.Run("mints missing refs", func(t *testing.T) {
		f := newFixture(t)
		a, _ := f.createSource(t, "Census 1881")
		b, _ := f.createSource(t, "Parish register")
		keep := graphRows(t, f.c, b.ID)[0].ref
		db, err := f.c.DB()
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`UPDATE evidence_graphs SET ref = NULL WHERE source_id = ?`, a.ID); err != nil {
			t.Fatal(err)
		}

		if err := evidencegraphs.EnsureRefs(f.c); err != nil {
			t.Fatal(err)
		}
		minted := graphRows(t, f.c, a.ID)[0].ref
		if !ref.Valid(minted) || !strings.HasPrefix(minted, "GPH-") {
			t.Fatalf("minted ref %q", minted)
		}
		if got := graphRows(t, f.c, b.ID)[0].ref; got != keep {
			t.Fatalf("existing ref changed %q -> %q", keep, got)
		}
		if err := evidencegraphs.EnsureRefs(f.c); err != nil {
			t.Fatal(err)
		}
		if got := graphRows(t, f.c, a.ID)[0].ref; got != minted {
			t.Fatalf("second call changed ref %q -> %q", minted, got)
		}
	})

	t.Run("retries a ref collision", func(t *testing.T) {
		f := newFixture(t)
		a, _ := f.createSource(t, "Census 1881")
		b, _ := f.createSource(t, "Parish register")
		keep := graphRows(t, f.c, a.ID)[0].ref
		db, err := f.c.DB()
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`UPDATE evidence_graphs SET ref = NULL WHERE source_id = ?`, b.ID); err != nil {
			t.Fatal(err)
		}
		calls := 0
		restore := evidencegraphs.SetMintRef(func() (string, error) {
			calls++
			if calls == 1 {
				return keep, nil
			}
			return ref.Mint(ref.PrefixEvidenceGraph)
		})
		t.Cleanup(restore)

		if err := evidencegraphs.EnsureRefs(f.c); err != nil {
			t.Fatal(err)
		}
		got := graphRows(t, f.c, b.ID)[0].ref
		if !ref.Valid(got) || !strings.HasPrefix(got, "GPH-") || got == keep {
			t.Fatalf("minted ref %q after a collision with %q", got, keep)
		}
		if calls != 2 {
			t.Fatalf("mint calls = %d, want 2", calls)
		}
		if after := graphRows(t, f.c, a.ID)[0].ref; after != keep {
			t.Fatalf("existing ref changed %q -> %q", keep, after)
		}
	})

	t.Run("exhausted retries roll back", func(t *testing.T) {
		f := newFixture(t)
		a, _ := f.createSource(t, "Census 1881")
		b, _ := f.createSource(t, "Parish register")
		c, _ := f.createSource(t, "Will")
		keep := graphRows(t, f.c, a.ID)[0].ref
		db, err := f.c.DB()
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`UPDATE evidence_graphs SET ref = NULL WHERE source_id IN (?, ?)`, b.ID, c.ID); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(evidencegraphs.SetMintRef(func() (string, error) { return keep, nil }))

		if err := evidencegraphs.EnsureRefs(f.c); !errors.Is(err, evidencegraphs.ErrInvalid) {
			t.Fatalf("err = %v, want ErrInvalid", err)
		}
		for _, id := range [][]byte{b.ID, c.ID} {
			if got := graphRows(t, f.c, id)[0].ref; got != "" {
				t.Fatalf("ref %q left after rollback", got)
			}
		}
	})
}

// artifactRef reads an Artifact's ref.
func artifactRef(t *testing.T, c *database.Catalog, id []byte) string {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	var r string
	if err := db.QueryRow(`SELECT ref FROM artifacts WHERE id = ?`, id).Scan(&r); err != nil {
		t.Fatal(err)
	}
	return r
}

// byArtifactRef sorts Artifact ids by their refs, case-insensitively.
func byArtifactRef(t *testing.T, c *database.Catalog, ids ...[]byte) [][]byte {
	t.Helper()
	out := append([][]byte(nil), ids...)
	sort.Slice(out, func(i, j int) bool {
		return strings.ToLower(artifactRef(t, c, out[i])) < strings.ToLower(artifactRef(t, c, out[j]))
	})
	return out
}

func sameIDs(a, b [][]byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !bytes.Equal(a[i], b[i]) {
			return false
		}
	}
	return true
}

func graphNames(gs []evidencegraphs.Graph) []string {
	var out []string
	for _, g := range gs {
		out = append(out, g.Name)
	}
	return out
}

// insertGraph adds a graph row by SQL and returns its id.
func insertGraph(t *testing.T, c *database.Catalog, sourceID []byte, name string) []byte {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO evidence_graphs (id, source_id, name) VALUES (?, ?, ?)`, id[:], sourceID, name); err != nil {
		t.Fatal(err)
	}
	return id[:]
}

func TestList(t *testing.T) {
	f := newFixture(t)
	s, _ := f.createSource(t, "Census 1881")
	other, _ := f.createSource(t, "Parish register")
	db, err := f.c.DB()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE evidence_graphs SET name = 'Archive' WHERE source_id = ?`, other.ID); err != nil {
		t.Fatal(err)
	}
	jones := insertGraph(t, f.c, s.ID, "Jones")
	insertGraph(t, f.c, s.ID, "smith")
	b := f.newArtifact(t, s, "B")
	a := f.newArtifact(t, s, "A")
	f.link(t, jones, b)
	f.link(t, jones, a)

	unknown := make([]byte, 16)
	unknown[0] = 9
	tests := []struct {
		name      string
		sourceID  []byte
		wantNames []string
		wantErr   error
	}{
		{name: "one Source, case-insensitive by name", sourceID: s.ID, wantNames: []string{"Default", "Jones", "smith"}},
		{name: "the whole project", sourceID: nil, wantNames: []string{"Archive", "Default", "Jones", "smith"}},
		{name: "empty id is the whole project", sourceID: []byte{}, wantNames: []string{"Archive", "Default", "Jones", "smith"}},
		{name: "another Source", sourceID: other.ID, wantNames: []string{"Archive"}},
		{name: "unknown Source", sourceID: unknown, wantNames: nil},
		{name: "short id", sourceID: unknown[:15], wantErr: evidencegraphs.ErrInvalid},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := evidencegraphs.List(f.c, tt.sourceID)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("got %v want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if names := graphNames(got); !reflect.DeepEqual(names, tt.wantNames) {
				t.Fatalf("names %v want %v", names, tt.wantNames)
			}
			for _, g := range got {
				if len(g.ID) != 16 || len(g.SourceID) != 16 {
					t.Fatalf("graph %q ids %x %x", g.Name, g.ID, g.SourceID)
				}
				if g.Name == "Jones" {
					if want := byArtifactRef(t, f.c, a, b); !sameIDs(g.ArtifactIDs, want) {
						t.Fatalf("Jones links %x want %x", g.ArtifactIDs, want)
					}
				} else if g.ArtifactIDs != nil {
					t.Fatalf("%s links %x", g.Name, g.ArtifactIDs)
				}
			}
		})
	}

	t.Run("fields", func(t *testing.T) {
		got, err := evidencegraphs.List(f.c, other.ID)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 || !ref.Valid(got[0].Ref) || got[0].Description != "" || !bytes.Equal(got[0].SourceID, other.ID) {
			t.Fatalf("got %+v", got)
		}
	})

	t.Run("closed catalog", func(t *testing.T) {
		g := newFixture(t)
		if err := g.c.Close(); err != nil {
			t.Fatal(err)
		}
		if _, err := evidencegraphs.List(g.c, nil); !errors.Is(err, database.ErrClosed) {
			t.Fatalf("got %v", err)
		}
	})
}

// createGraph runs Create in one write.
func (f fixture) createGraph(t *testing.T, in evidencegraphs.CreateInput) (evidencegraphs.Graph, []rowchange.Change, error) {
	t.Helper()
	var changes []rowchange.Change
	g, err := writes.Call(f.c, writes.Op{Action: "create_evidence_graph", UserID: userID}, func(tx *database.Tx) (evidencegraphs.Graph, []rowchange.Change, error) {
		g, ch, err := evidencegraphs.Create(tx, userID, in)
		changes = ch
		return g, ch, err
	})
	return g, changes, err
}

func changeKinds(changes []rowchange.Change) []string {
	var out []string
	for _, c := range changes {
		out = append(out, c.EntityType+" "+string(c.Action))
	}
	return out
}

func mustList(t *testing.T, c *database.Catalog, sourceID []byte) []evidencegraphs.Graph {
	t.Helper()
	got, err := evidencegraphs.List(c, sourceID)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func TestCreate(t *testing.T) {
	f := newFixture(t)
	s, _ := f.createSource(t, "Census 1881")
	other, _ := f.createSource(t, "Parish register")
	a1 := f.newArtifact(t, s, "Page 1")
	a2 := f.newArtifact(t, s, "Page 2")
	foreign := f.newArtifact(t, other, "Page 9")
	unknown := make([]byte, 16)
	unknown[0] = 9

	t.Run("creates and links", func(t *testing.T) {
		g, changes, err := f.createGraph(t, evidencegraphs.CreateInput{
			SourceID: s.ID, Name: "  Jones household  ", Description: "", ArtifactIDs: [][]byte{a2, a1},
		})
		if err != nil {
			t.Fatal(err)
		}
		if g.Name != "Jones household" || g.Description != "" || !bytes.Equal(g.SourceID, s.ID) {
			t.Fatalf("graph %+v", g)
		}
		if !ref.Valid(g.Ref) || !strings.HasPrefix(g.Ref, "GPH-") {
			t.Fatalf("ref %q", g.Ref)
		}
		if id, err := uuid.FromBytes(g.ID); err != nil || id.Version() != 7 {
			t.Fatalf("id %x: %v", g.ID, err)
		}
		if want := []string{"evidence_graph create", "evidence_graph_artifact create", "evidence_graph_artifact create"}; !reflect.DeepEqual(changeKinds(changes), want) {
			t.Fatalf("changes %v", changeKinds(changes))
		}
		if got := changes[1].Fields["artifact_id"].New; got != uuid.UUID(a2).String() {
			t.Fatalf("first link %v", got)
		}
		if got := changes[2].Fields["artifact_id"].New; got != uuid.UUID(a1).String() {
			t.Fatalf("second link %v", got)
		}
		if !sameIDs(g.ArtifactIDs, [][]byte{a2, a1}) {
			t.Fatalf("ArtifactIDs %x", g.ArtifactIDs)
		}
		var stored evidencegraphs.Graph
		for _, x := range mustList(t, f.c, s.ID) {
			if bytes.Equal(x.ID, g.ID) {
				stored = x
			}
		}
		if stored.Name != "Jones household" || stored.Description != "" || len(stored.ArtifactIDs) != 2 {
			t.Fatalf("stored %+v", stored)
		}
	})

	t.Run("a name that repeats is fine", func(t *testing.T) {
		if _, _, err := f.createGraph(t, evidencegraphs.CreateInput{
			SourceID: s.ID, Name: evidencegraphs.DefaultName, ArtifactIDs: [][]byte{a1},
		}); err != nil {
			t.Fatal(err)
		}
	})

	refused := []struct {
		name string
		in   evidencegraphs.CreateInput
	}{
		{"empty name", evidencegraphs.CreateInput{SourceID: s.ID, Name: "", ArtifactIDs: [][]byte{a1}}},
		{"blank name", evidencegraphs.CreateInput{SourceID: s.ID, Name: "   ", ArtifactIDs: [][]byte{a1}}},
		{"no Artifacts", evidencegraphs.CreateInput{SourceID: s.ID, Name: "X"}},
		{"repeated Artifact", evidencegraphs.CreateInput{SourceID: s.ID, Name: "X", ArtifactIDs: [][]byte{a1, a1}}},
		{"short Artifact id", evidencegraphs.CreateInput{SourceID: s.ID, Name: "X", ArtifactIDs: [][]byte{a1[:15]}}},
		{"Artifact of another Source", evidencegraphs.CreateInput{SourceID: s.ID, Name: "X", ArtifactIDs: [][]byte{a1, foreign}}},
		{"unknown Artifact", evidencegraphs.CreateInput{SourceID: s.ID, Name: "X", ArtifactIDs: [][]byte{unknown}}},
		{"unknown Source", evidencegraphs.CreateInput{SourceID: unknown, Name: "X", ArtifactIDs: [][]byte{a1}}},
		{"short Source id", evidencegraphs.CreateInput{SourceID: s.ID[:15], Name: "X", ArtifactIDs: [][]byte{a1}}},
	}
	for _, tt := range refused {
		t.Run(tt.name, func(t *testing.T) {
			before := mustList(t, f.c, nil)
			if _, _, err := f.createGraph(t, tt.in); !errors.Is(err, evidencegraphs.ErrInvalid) {
				t.Fatalf("got %v", err)
			}
			if after := mustList(t, f.c, nil); !reflect.DeepEqual(before, after) {
				t.Fatalf("graphs changed: %+v -> %+v", before, after)
			}
		})
	}

	t.Run("empty user id", func(t *testing.T) {
		_, err := run(t, f.c, func(tx *database.Tx) ([]rowchange.Change, error) {
			_, ch, err := evidencegraphs.Create(tx, nil, evidencegraphs.CreateInput{SourceID: s.ID, Name: "X", ArtifactIDs: [][]byte{a1}})
			return ch, err
		})
		if !errors.Is(err, evidencegraphs.ErrInvalid) {
			t.Fatalf("got %v", err)
		}
	})
}

// updateGraph runs Update in one write.
func (f fixture) updateGraph(t *testing.T, graphID []byte, in evidencegraphs.UpdateInput) ([]rowchange.Change, error) {
	t.Helper()
	return run(t, f.c, func(tx *database.Tx) ([]rowchange.Change, error) {
		return evidencegraphs.Update(tx, userID, graphID, in)
	})
}

func graphByID(t *testing.T, c *database.Catalog, sourceID, id []byte) evidencegraphs.Graph {
	t.Helper()
	for _, g := range mustList(t, c, sourceID) {
		if bytes.Equal(g.ID, id) {
			return g
		}
	}
	t.Fatalf("graph %x not listed", id)
	return evidencegraphs.Graph{}
}

func TestUpdate(t *testing.T) {
	f := newFixture(t)
	s, _ := f.createSource(t, "Census 1881")
	other, _ := f.createSource(t, "Parish register")
	def := evidencegraphstest.SoleGraph(t, f.c, s.ID).ID
	a1 := f.newArtifact(t, s, "Page 1")
	a2 := f.newArtifact(t, s, "Page 2")
	foreign := f.newArtifact(t, other, "Page 9")
	f.link(t, def, a1)
	unknown := make([]byte, 16)
	unknown[0] = 9

	t.Run("rename", func(t *testing.T) {
		ch, err := f.updateGraph(t, def, evidencegraphs.UpdateInput{Name: "Smith household"})
		if err != nil {
			t.Fatal(err)
		}
		if len(ch) != 1 || ch[0].EntityType != "evidence_graph" || ch[0].Action != rowchange.ActionUpdate || !bytes.Equal(ch[0].EntityID, def) {
			t.Fatalf("changes %+v", ch)
		}
		if len(ch[0].Fields) != 1 || ch[0].Fields["name"].Old != "Default" || ch[0].Fields["name"].New != "Smith household" {
			t.Fatalf("fields %+v", ch[0].Fields)
		}
		if g := graphByID(t, f.c, s.ID, def); g.Name != "Smith household" || len(g.ArtifactIDs) != 1 {
			t.Fatalf("graph %+v", g)
		}
	})

	t.Run("description", func(t *testing.T) {
		ch, err := f.updateGraph(t, def, evidencegraphs.UpdateInput{Name: "Smith household", Description: " Pages 4–6 "})
		if err != nil {
			t.Fatal(err)
		}
		if len(ch) != 1 || len(ch[0].Fields) != 1 || ch[0].Fields["description"].Old != nil || ch[0].Fields["description"].New != "Pages 4–6" {
			t.Fatalf("changes %+v", ch)
		}
		ch, err = f.updateGraph(t, def, evidencegraphs.UpdateInput{Name: "Smith household", Description: ""})
		if err != nil {
			t.Fatal(err)
		}
		if len(ch) != 1 || ch[0].Fields["description"].Old != "Pages 4–6" || ch[0].Fields["description"].New != nil {
			t.Fatalf("changes %+v", ch)
		}
	})

	t.Run("unchanged is no change", func(t *testing.T) {
		ch, err := f.updateGraph(t, def, evidencegraphs.UpdateInput{Name: "  Smith household ", Description: "  "})
		if err != nil || ch != nil {
			t.Fatalf("got %v %+v", err, ch)
		}
	})

	t.Run("rename with no Artifact ids keeps the links", func(t *testing.T) {
		ch, err := f.updateGraph(t, def, evidencegraphs.UpdateInput{Name: "Smith again"})
		if err != nil {
			t.Fatal(err)
		}
		if len(ch) != 1 || ch[0].EntityType != "evidence_graph" {
			t.Fatalf("changes %v", changeKinds(ch))
		}
		if g := graphByID(t, f.c, s.ID, def); len(g.ArtifactIDs) != 1 || !bytes.Equal(g.ArtifactIDs[0], a1) {
			t.Fatalf("links %x", g.ArtifactIDs)
		}
	})

	jones, _, err := f.createGraph(t, evidencegraphs.CreateInput{SourceID: s.ID, Name: "Jones", ArtifactIDs: [][]byte{a1}})
	if err != nil {
		t.Fatal(err)
	}

	refused := []struct {
		name string
		id   []byte
		in   evidencegraphs.UpdateInput
	}{
		{"empty name", jones.ID, evidencegraphs.UpdateInput{Name: " "}},
		{"unknown graph", unknown, evidencegraphs.UpdateInput{Name: "X"}},
		{"short graph id", jones.ID[:15], evidencegraphs.UpdateInput{Name: "X"}},
		{"Artifact of another Source", jones.ID, evidencegraphs.UpdateInput{Name: "Renamed", ArtifactIDs: [][]byte{a1, foreign}}},
		{"unknown Artifact", jones.ID, evidencegraphs.UpdateInput{Name: "Renamed", ArtifactIDs: [][]byte{a1, unknown}}},
		{"repeated Artifact", jones.ID, evidencegraphs.UpdateInput{Name: "Renamed", ArtifactIDs: [][]byte{a1, a1}}},
	}
	for _, tt := range refused {
		t.Run(tt.name, func(t *testing.T) {
			before := graphByID(t, f.c, s.ID, jones.ID)
			if _, err := f.updateGraph(t, tt.id, tt.in); !errors.Is(err, evidencegraphs.ErrInvalid) {
				t.Fatalf("got %v", err)
			}
			if after := graphByID(t, f.c, s.ID, jones.ID); !reflect.DeepEqual(before, after) {
				t.Fatalf("graph changed: %+v -> %+v", before, after)
			}
		})
	}
	t.Run("empty user id", func(t *testing.T) {
		_, err := run(t, f.c, func(tx *database.Tx) ([]rowchange.Change, error) {
			return evidencegraphs.Update(tx, nil, jones.ID, evidencegraphs.UpdateInput{Name: "X"})
		})
		if !errors.Is(err, evidencegraphs.ErrInvalid) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("swapping an Artifact links before it unlinks", func(t *testing.T) {
		ch, err := f.updateGraph(t, jones.ID, evidencegraphs.UpdateInput{Name: "Jones", ArtifactIDs: [][]byte{a2}})
		if err != nil {
			t.Fatal(err)
		}
		if want := []string{"evidence_graph_artifact create", "evidence_graph_artifact delete"}; !reflect.DeepEqual(changeKinds(ch), want) {
			t.Fatalf("changes %v", changeKinds(ch))
		}
		if got := ch[0].Fields["artifact_id"].New; got != uuid.UUID(a2).String() {
			t.Fatalf("link %v", got)
		}
		if got := ch[1].Fields["artifact_id"].Old; got != uuid.UUID(a1).String() {
			t.Fatalf("unlink %v", got)
		}
		if g := graphByID(t, f.c, s.ID, jones.ID); len(g.ArtifactIDs) != 1 || !bytes.Equal(g.ArtifactIDs[0], a2) {
			t.Fatalf("links %x", g.ArtifactIDs)
		}
	})

	t.Run("a refused unlink rolls back the rename", func(t *testing.T) {
		g := newFixture(t)
		src, _ := g.createSource(t, "Register")
		graph := evidencegraphstest.SoleGraph(t, g.c, src.ID).ID
		b1 := g.newArtifact(t, src, "Page 1")
		b2 := g.newArtifact(t, src, "Page 2")
		g.link(t, graph, b1)
		g.link(t, graph, b2)
		g.subjectAndCitation(t, src, b1)
		_, err := g.updateGraph(t, graph, evidencegraphs.UpdateInput{Name: "Smith household", ArtifactIDs: [][]byte{b2}})
		if !errors.Is(err, evidencegraphs.ErrArtifactInUse) {
			t.Fatalf("got %v", err)
		}
		got := graphByID(t, g.c, src.ID, graph)
		if got.Name != evidencegraphs.DefaultName || len(got.ArtifactIDs) != 2 {
			t.Fatalf("graph %+v", got)
		}
	})
}

func (f fixture) deleteGraph(t *testing.T, graphID []byte) ([]rowchange.Change, error) {
	t.Helper()
	return run(t, f.c, func(tx *database.Tx) ([]rowchange.Change, error) {
		return evidencegraphs.Delete(tx, userID, graphID)
	})
}

func graphImpact(t *testing.T, c *database.Catalog, graphID []byte) deleteimpact.Report {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	report, err := deleteimpact.Snapshot(db, catalogmodel.KindEvidenceGraph, graphID)
	if err != nil {
		t.Fatal(err)
	}
	return report
}

func TestDelete(t *testing.T) {
	unknown := make([]byte, 16)
	unknown[0] = 9

	t.Run("an empty graph that is not the last", func(t *testing.T) {
		f := newFixture(t)
		s, _ := f.createSource(t, "Census 1881")
		def := evidencegraphstest.SoleGraph(t, f.c, s.ID)
		a1 := f.newArtifact(t, s, "Page 1")
		f.link(t, def.ID, a1)
		jones, _, err := f.createGraph(t, evidencegraphs.CreateInput{SourceID: s.ID, Name: "Jones", Description: "Pages 4", ArtifactIDs: [][]byte{a1}})
		if err != nil {
			t.Fatal(err)
		}
		if report := graphImpact(t, f.c, jones.ID); !report.Allowed || report.Gate != deleteimpact.GateOK || len(report.Groups) != 0 {
			t.Fatalf("impact %+v", report)
		}
		ch, err := f.deleteGraph(t, jones.ID)
		if err != nil {
			t.Fatal(err)
		}
		if want := []string{"evidence_graph_artifact delete", "evidence_graph delete"}; !reflect.DeepEqual(changeKinds(ch), want) {
			t.Fatalf("changes %v", changeKinds(ch))
		}
		fields := ch[1].Fields
		for _, key := range []string{"id", "ref", "source_id", "name", "description"} {
			if _, ok := fields[key]; !ok {
				t.Fatalf("delete change lacks %q: %+v", key, fields)
			}
		}
		if len(fields) != 5 || fields["name"].Old != "Jones" || fields["description"].Old != "Pages 4" || fields["ref"].Old != jones.Ref {
			t.Fatalf("fields %+v", fields)
		}
		if got := mustList(t, f.c, s.ID); len(got) != 1 || !bytes.Equal(got[0].ID, def.ID) || len(got[0].ArtifactIDs) != 1 {
			t.Fatalf("graphs %+v", got)
		}
		if linkedArtifacts(t, f.c, jones.ID) != 0 {
			t.Fatal("links stayed")
		}
	})

	t.Run("a Source's last graph", func(t *testing.T) {
		f := newFixture(t)
		s, _ := f.createSource(t, "Census 1881")
		def := evidencegraphstest.SoleGraph(t, f.c, s.ID)
		a1 := f.newArtifact(t, s, "Page 1")
		f.link(t, def.ID, a1)
		check := func(t *testing.T) {
			t.Helper()
			if _, err := f.deleteGraph(t, def.ID); !errors.Is(err, evidencegraphs.ErrLastGraph) {
				t.Fatalf("got %v", err)
			}
			if got := mustList(t, f.c, s.ID); len(got) != 1 || len(got[0].ArtifactIDs) != 1 {
				t.Fatalf("graphs %+v", got)
			}
			if report := graphImpact(t, f.c, def.ID); report.Allowed || report.Gate != deleteimpact.GateLastGraph || len(report.Groups) != 0 {
				t.Fatalf("impact %+v", report)
			}
		}
		t.Run("without Subjects", check)
		f.subjectAndCitation(t, s, a1)
		t.Run("with a Subject", check)
	})

	t.Run("a graph with Subjects", func(t *testing.T) {
		f := newFixture(t)
		s, _ := f.createSource(t, "Census 1881")
		def := evidencegraphstest.SoleGraph(t, f.c, s.ID)
		a1 := f.newArtifact(t, s, "Page 1")
		f.link(t, def.ID, a1)
		subject, _ := f.subjectAndCitation(t, s, a1)
		if _, _, err := f.createGraph(t, evidencegraphs.CreateInput{SourceID: s.ID, Name: "Jones", ArtifactIDs: [][]byte{a1}}); err != nil {
			t.Fatal(err)
		}
		_, err := f.deleteGraph(t, def.ID)
		if !errors.Is(err, evidencegraphs.ErrInUse) {
			t.Fatalf("got %v", err)
		}
		var ae *apperr.Error
		if !errors.As(err, &ae) || ae.Kind() != apperr.KindConflict || len(ae.Params()) != 0 {
			t.Fatalf("refusal %v", err)
		}
		if got := mustList(t, f.c, s.ID); len(got) != 2 {
			t.Fatalf("graphs %+v", got)
		}
		report := graphImpact(t, f.c, def.ID)
		if report.Allowed || report.Gate != deleteimpact.GateInbound || len(report.Groups) != 1 ||
			report.Groups[0].Via != "subjects.graph_id" || report.Groups[0].Total != 1 ||
			len(report.Groups[0].Listed) != 1 || report.Groups[0].Listed[0].Ref != subject.Ref {
			t.Fatalf("impact %+v", report)
		}
	})

	t.Run("bad input", func(t *testing.T) {
		f := newFixture(t)
		s, _ := f.createSource(t, "Census 1881")
		def := evidencegraphstest.SoleGraph(t, f.c, s.ID)
		for name, id := range map[string][]byte{"unknown": unknown, "short": def.ID[:15]} {
			if _, err := f.deleteGraph(t, id); !errors.Is(err, evidencegraphs.ErrInvalid) {
				t.Fatalf("%s: got %v", name, err)
			}
		}
		_, err := run(t, f.c, func(tx *database.Tx) ([]rowchange.Change, error) {
			return evidencegraphs.Delete(tx, nil, def.ID)
		})
		if !errors.Is(err, evidencegraphs.ErrInvalid) {
			t.Fatalf("empty user: got %v", err)
		}
		if report := graphImpact(t, f.c, unknown); report.Gate != deleteimpact.GateNotFound {
			t.Fatalf("impact %+v", report)
		}
	})
}

// TestListQueryPlan: the per-Source read seeks evidence_graphs_source_id_idx
// instead of scanning every graph or link.
func TestListQueryPlan(t *testing.T) {
	f := newFixture(t)
	s, _ := f.createSource(t, "Census 1881")
	db, err := f.c.DB()
	if err != nil {
		t.Fatal(err)
	}
	graphs, links := evidencegraphs.ListSQL()
	for name, q := range map[string]string{"graphs": graphs, "links": links} {
		args := make([]any, strings.Count(q, "?"))
		for i := range args {
			args[i] = s.ID
		}
		rows, err := db.Query("EXPLAIN QUERY PLAN "+q, args...)
		if err != nil {
			t.Fatal(err)
		}
		for rows.Next() {
			var id, parent, unused int
			var detail string
			if err := rows.Scan(&id, &parent, &unused, &detail); err != nil {
				t.Fatal(err)
			}
			if strings.HasPrefix(detail, "SCAN") {
				t.Errorf("%s query plan scans: %s", name, detail)
			}
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		rows.Close()
	}
}
