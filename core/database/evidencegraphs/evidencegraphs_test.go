package evidencegraphs_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/evidencegraphs"
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

func defaultGraphID(t *testing.T, c *database.Catalog, sourceID []byte) []byte {
	t.Helper()
	g := graphRows(t, c, sourceID)
	if len(g) != 1 {
		t.Fatalf("graphs %d", len(g))
	}
	return g[0].id
}

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
	graph := defaultGraphID(t, f.c, src.ID)
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
	graph := defaultGraphID(t, f.c, src.ID)
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
	graph := defaultGraphID(t, f.c, src.ID)
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
