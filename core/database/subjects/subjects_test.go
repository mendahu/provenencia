package subjects

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
	_ "github.com/mendahu/provenencia/core/database/graphfill"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjectpositions"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
)

func TestSubjects(t *testing.T) {
	userID := []byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}

	mustUser := func(t *testing.T, c *database.Catalog) {
		t.Helper()
		r, err := ref.Mint(ref.PrefixUser)
		if err != nil {
			t.Fatal(err)
		}
		if err := users.Upsert(c, userID, "Jake", r); err != nil {
			t.Fatal(err)
		}
	}
	mustSource := func(t *testing.T, c *database.Catalog) sources.Source {
		t.Helper()
		typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
			Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book",
		})
		if err != nil {
			t.Fatal(err)
		}
		src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
			return sources.Create(tx, userID, sources.CreateInput{
				SourceTypeID: typeID,
				Title:        "Deed",
			})
		})
		if err != nil {
			t.Fatal(err)
		}
		return src
	}
	mustArtifact := func(t *testing.T, c *database.Catalog, sourceID []byte) artifacts.Artifact {
		t.Helper()
		a, _, err := writes.Run(c, writes.Op{Action: "create_artifact", UserID: userID}, func(tx *database.Tx) (artifacts.Artifact, []rowchange.Change, error) {
			return artifacts.Create(tx, userID, artifacts.CreateInput{SourceID: sourceID, Label: "Scan"})
		})
		if err != nil {
			t.Fatal(err)
		}
		return a
	}
	mustSubjectTypes := func(t *testing.T, c *database.Catalog) {
		t.Helper()
		if err := subjectvocab.Install(c); err != nil {
			t.Fatal(err)
		}
	}
	latestAction := func(t *testing.T, c *database.Catalog) string {
		t.Helper()
		db, err := c.DB()
		if err != nil {
			t.Fatal(err)
		}
		var actionType string
		if err := db.QueryRow(`SELECT action_type FROM audit_transactions ORDER BY revision DESC LIMIT 1`).Scan(&actionType); err != nil {
			t.Fatal(err)
		}
		return actionType
	}
	latestEntityType := func(t *testing.T, c *database.Catalog) string {
		t.Helper()
		db, err := c.DB()
		if err != nil {
			t.Fatal(err)
		}
		var entityType string
		if err := db.QueryRow(`
			SELECT entity_type FROM audit_changes
			ORDER BY rowid DESC LIMIT 1`).Scan(&entityType); err != nil {
			t.Fatal(err)
		}
		return entityType
	}

	tests := []struct {
		name string
		run  func(t *testing.T, c *database.Catalog)
	}{
		{
			name: "create joins the Source's graph",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				mustSubjectTypes(t, c)
				src := mustSource(t, c)
				person, err := subjecttypes.Lookup(c, "person", vocabkey.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				want := evidencegraphstest.SoleGraph(t, c, src.ID).ID
				var created []rowchange.Change
				s, err := writes.Call(c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (Subject, []rowchange.Change, error) {
					s, changes, err := Create(tx, userID, CreateInput{
						SourceID: src.ID, SubjectTypeID: person.ID, Label: "Alice",
					}, nil)
					created = changes
					return s, changes, err
				})
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(s.GraphID, want) {
					t.Fatalf("created graph %x want %x", s.GraphID, want)
				}
				got, err := Get(c, s.ID)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(got.GraphID, want) {
					t.Fatalf("read graph %x want %x", got.GraphID, want)
				}
				if len(created) == 0 || created[0].Fields["graph_id"].New == nil {
					t.Fatalf("create change lacks graph_id: %+v", created)
				}
				var deleted []rowchange.Change
				if _, err := writes.Call(c, writes.Op{Action: "delete_subject", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
					changes, err := Delete(tx, userID, s.ID)
					deleted = changes
					return struct{}{}, changes, err
				}); err != nil {
					t.Fatal(err)
				}
				last := deleted[len(deleted)-1]
				if last.EntityType != "subject" || last.Fields["graph_id"].Old == nil {
					t.Fatalf("delete change lacks graph_id: %+v", last)
				}
			},
		},
		{
			name: "graph id given",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				mustSubjectTypes(t, c)
				src := mustSource(t, c)
				person, err := subjecttypes.Lookup(c, "person", vocabkey.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				art := mustArtifact(t, c, src.ID)
				jones := evidencegraphstest.Create(t, c, userID, src.ID, "Jones", art.ID)
				var created []rowchange.Change
				s, err := writes.Call(c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (Subject, []rowchange.Change, error) {
					s, changes, err := Create(tx, userID, CreateInput{
						SourceID: src.ID, GraphID: jones.ID, SubjectTypeID: person.ID, Label: "Ben",
					}, nil)
					created = changes
					return s, changes, err
				})
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(s.GraphID, jones.ID) {
					t.Fatalf("created graph %x want %x", s.GraphID, jones.ID)
				}
				got, err := Get(c, s.ID)
				if err != nil || !bytes.Equal(got.GraphID, jones.ID) {
					t.Fatalf("read graph %x want %x (%v)", got.GraphID, jones.ID, err)
				}
				if len(created) == 0 || created[0].Fields["graph_id"].New != database.UUIDJSON(jones.ID) {
					t.Fatalf("create change graph_id: %+v", created)
				}
			},
		},
		{
			name: "graph of another source",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				mustSubjectTypes(t, c)
				srcA := mustSource(t, c)
				typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
					Key: "census", Origin: vocabkey.OriginProvenencia, Label: "Census",
				})
				if err != nil {
					t.Fatal(err)
				}
				srcB, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
					return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Other"})
				})
				if err != nil {
					t.Fatal(err)
				}
				person, err := subjecttypes.Lookup(c, "person", vocabkey.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				other := evidencegraphstest.SoleGraph(t, c, srcB.ID)
				_, err = writes.Call(c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (Subject, []rowchange.Change, error) {
					return Create(tx, userID, CreateInput{
						SourceID: srcA.ID, GraphID: other.ID, SubjectTypeID: person.ID, Label: "X",
					}, nil)
				})
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v want ErrInvalid", err)
				}
			},
		},
		{
			name: "graph id wrong length",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				mustSubjectTypes(t, c)
				src := mustSource(t, c)
				person, err := subjecttypes.Lookup(c, "person", vocabkey.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				_, err = writes.Call(c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (Subject, []rowchange.Change, error) {
					return Create(tx, userID, CreateInput{
						SourceID: src.ID, GraphID: []byte{1, 2, 3}, SubjectTypeID: person.ID, Label: "X",
					}, nil)
				})
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v want ErrInvalid", err)
				}
			},
		},
		{
			name: "create rejects empty user id",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				mustSubjectTypes(t, c)
				src := mustSource(t, c)
				person, err := subjecttypes.Lookup(c, "person", vocabkey.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				_, err = writes.Call(c, writes.Op{Action: "create_subject", UserID: nil}, func(tx *database.Tx) (Subject, []rowchange.Change, error) {
					return Create(tx, nil, CreateInput{
						SourceID:      src.ID,
						SubjectTypeID: person.ID,
						Label:         "Alice",
					}, nil)
				})
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v want ErrInvalid", err)
				}
			},
		},
		{
			name: "create person mints CPR ref and audits",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				mustSubjectTypes(t, c)
				src := mustSource(t, c)
				person, err := subjecttypes.Lookup(c, "person", vocabkey.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				s, err := writes.Call(c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (Subject, []rowchange.Change, error) {
					return Create(tx, userID, CreateInput{
						SourceID:      src.ID,
						SubjectTypeID: person.ID,
						Label:         "Alice",
					}, nil)
				})
				if err != nil {
					t.Fatal(err)
				}
				if !strings.HasPrefix(s.Ref, "CPR-") {
					t.Fatalf("ref %q", s.Ref)
				}
				if latestAction(t, c) != "create_subject" {
					t.Fatalf("action %s", latestAction(t, c))
				}
				if latestEntityType(t, c) != "subject" {
					t.Fatalf("entity %s", latestEntityType(t, c))
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				var raw string
				if err := db.QueryRow(`
					SELECT changes_json FROM audit_changes
					ORDER BY rowid DESC LIMIT 1`).Scan(&raw); err != nil {
					t.Fatal(err)
				}
				var fields map[string]any
				if err := json.Unmarshal([]byte(raw), &fields); err != nil {
					t.Fatal(err)
				}
				for _, key := range []string{"id", "ref", "source_id", "subject_type_id", "label", "description"} {
					if _, ok := fields[key]; !ok {
						t.Fatalf("subject create missing %q in %+v", key, fields)
					}
				}
				desc, _ := fields["description"].(map[string]any)
				if desc["new"] != nil {
					t.Fatalf("empty description should be null: %+v", fields)
				}
				got, err := Get(c, s.ID)
				if err != nil || got.Label != "Alice" {
					t.Fatalf("%v %+v", err, got)
				}
				byRef, err := getByRef(c, s.Ref)
				if err != nil || string(byRef.ID) != string(s.ID) {
					t.Fatalf("%v %+v", err, byRef)
				}
				if _, err := subjectpositions.Get(c, s.ID); !errors.Is(err, sql.ErrNoRows) {
					t.Fatalf("no placement wrote position: %v", err)
				}
			},
		},
		{
			name: "create with placement writes position",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				mustSubjectTypes(t, c)
				src := mustSource(t, c)
				person, err := subjecttypes.Lookup(c, "person", vocabkey.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				s, err := writes.Call(c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (Subject, []rowchange.Change, error) {
					return Create(tx, userID, CreateInput{
						SourceID:      src.ID,
						SubjectTypeID: person.ID,
						Label:         "Placed",
					}, &Placement{GridX: -2, GridY: 5})
				})
				if err != nil {
					t.Fatal(err)
				}
				pos, err := subjectpositions.Get(c, s.ID)
				if err != nil {
					t.Fatal(err)
				}
				if pos.GridX != -2 || pos.GridY != 5 {
					t.Fatalf("position %+v", pos)
				}
			},
		},
		{
			name: "create event mints CEV ref",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				mustSubjectTypes(t, c)
				src := mustSource(t, c)
				event, err := subjecttypes.Lookup(c, "event", vocabkey.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				s, err := writes.Call(c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (Subject, []rowchange.Change, error) {
					return Create(tx, userID, CreateInput{
						SourceID:      src.ID,
						SubjectTypeID: event.ID,
					}, nil)
				})
				if err != nil {
					t.Fatal(err)
				}
				if !strings.HasPrefix(s.Ref, "CEV-") {
					t.Fatalf("ref %q", s.Ref)
				}
			},
		},
		{
			name: "update label records update_subject",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				mustSubjectTypes(t, c)
				src := mustSource(t, c)
				person, err := subjecttypes.Lookup(c, "person", vocabkey.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				s, err := writes.Call(c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (Subject, []rowchange.Change, error) {
					return Create(tx, userID, CreateInput{
						SourceID: src.ID, SubjectTypeID: person.ID, Label: "Alice",
					}, nil)
				})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := writes.Call(c, writes.Op{Action: "update_subject", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
					changes, err := Update(tx, userID, s.ID, "Alicia", "")
					return struct{}{}, changes, err
				}); err != nil {
					t.Fatal(err)
				}
				if latestAction(t, c) != "update_subject" {
					t.Fatalf("action %s", latestAction(t, c))
				}
				got, err := Get(c, s.ID)
				if err != nil {
					t.Fatal(err)
				}
				if got.Label != "Alicia" || string(got.SubjectTypeID) != string(person.ID) {
					t.Fatalf("%+v", got)
				}
				if _, err := writes.Call(c, writes.Op{Action: "update_subject", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
					changes, err := Update(tx, userID, s.ID, "Alicia", "")
					return struct{}{}, changes, err
				}); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "delete records delete_subject and cascades positions",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				mustSubjectTypes(t, c)
				src := mustSource(t, c)
				person, err := subjecttypes.Lookup(c, "person", vocabkey.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				s, err := writes.Call(c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (Subject, []rowchange.Change, error) {
					return Create(tx, userID, CreateInput{
						SourceID: src.ID, SubjectTypeID: person.ID, Label: "Bob",
					}, nil)
				})
				if err != nil {
					t.Fatal(err)
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				if _, err := db.Exec(
					`INSERT INTO subject_positions (subject_id, grid_x, grid_y) VALUES (?, 1, 2)`,
					s.ID,
				); err != nil {
					t.Fatal(err)
				}
				if _, err := writes.Call(c, writes.Op{Action: "delete_subject", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
					changes, err := Delete(tx, userID, s.ID)
					return struct{}{}, changes, err
				}); err != nil {
					t.Fatal(err)
				}
				if latestAction(t, c) != "delete_subject" {
					t.Fatalf("action %s", latestAction(t, c))
				}
				_, err = Get(c, s.ID)
				if !errors.Is(err, sql.ErrNoRows) {
					t.Fatalf("got %v", err)
				}
				var n int
				if err := db.QueryRow(`SELECT COUNT(*) FROM subject_positions WHERE subject_id = ?`, s.ID).Scan(&n); err != nil {
					t.Fatal(err)
				}
				if n != 0 {
					t.Fatalf("positions left %d", n)
				}
			},
		},
		{
			name: "list by source is scoped",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				mustSubjectTypes(t, c)
				srcA := mustSource(t, c)
				typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
					Key: "census", Origin: vocabkey.OriginProvenencia, Label: "Census",
				})
				if err != nil {
					t.Fatal(err)
				}
				srcB, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
					return sources.Create(tx, userID, sources.CreateInput{
						SourceTypeID: typeID, Title: "Other",
					})
				})
				if err != nil {
					t.Fatal(err)
				}
				person, err := subjecttypes.Lookup(c, "person", vocabkey.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				a, err := writes.Call(c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (Subject, []rowchange.Change, error) {
					return Create(tx, userID, CreateInput{SourceID: srcA.ID, SubjectTypeID: person.ID, Label: "A"}, nil)
				})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := writes.Call(c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (Subject, []rowchange.Change, error) {
					return Create(tx, userID, CreateInput{SourceID: srcB.ID, SubjectTypeID: person.ID, Label: "B"}, nil)
				}); err != nil {
					t.Fatal(err)
				}
				eg, err := c.Graph().EvidenceGraph(a.GraphID)
				if err != nil || len(eg.Subjects) != 1 || eg.Subjects[0].Label != "A" {
					t.Fatalf("%v %+v", err, eg.Subjects)
				}
			},
		},
		{
			name: "reject invalid ids",
			run: func(t *testing.T, c *database.Catalog) {
				if _, err := writes.Call(c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (Subject, []rowchange.Change, error) {
					return Create(tx, userID, CreateInput{}, nil)
				}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
				if _, err := Get(c, []byte{1}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := database.Create(t.TempDir(), "t.provenencia")
			if err != nil {
				t.Fatal(err)
			}
			defer c.Close()
			tt.run(t, c)
		})
	}
}

func TestInsertReadsTheGraphsSource(t *testing.T) {
	userID := []byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
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
	if err := subjectvocab.Install(c); err != nil {
		t.Fatal(err)
	}
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book"})
	if err != nil {
		t.Fatal(err)
	}
	newSource := func(title string) sources.Source {
		t.Helper()
		src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
			return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: title})
		})
		if err != nil {
			t.Fatal(err)
		}
		return src
	}
	src, other := newSource("Deed"), newSource("Other")
	art, _, err := writes.Run(c, writes.Op{Action: "create_artifact", UserID: userID}, func(tx *database.Tx) (artifacts.Artifact, []rowchange.Change, error) {
		return artifacts.Create(tx, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
	})
	if err != nil {
		t.Fatal(err)
	}
	jones := evidencegraphstest.Create(t, c, userID, src.ID, "Jones", art.ID)
	person, err := subjecttypes.Lookup(c, "person", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	unknown := make([]byte, 16)
	unknown[0] = 9
	create := func(in CreateInput) (Subject, error) {
		in.SubjectTypeID, in.Label = person.ID, "X"
		return writes.Call(c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (Subject, []rowchange.Change, error) {
			return Create(tx, userID, in, nil)
		})
	}

	s, err := create(CreateInput{GraphID: jones.ID})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(s.SourceID, src.ID) || !bytes.Equal(s.GraphID, jones.ID) {
		t.Fatalf("subject on %x of %x, want %x of %x", s.GraphID, s.SourceID, jones.ID, src.ID)
	}
	for name, in := range map[string]CreateInput{
		"another Source's id": {GraphID: jones.ID, SourceID: other.ID},
		"neither id":          {},
		"unknown graph":       {GraphID: unknown},
	} {
		if _, err := create(in); !errors.Is(err, ErrInvalid) {
			t.Fatalf("%s: got %v want ErrInvalid", name, err)
		}
	}
}
