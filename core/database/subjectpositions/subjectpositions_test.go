package subjectpositions_test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/artifacts/artifactstest"
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjectpositions"
	"github.com/mendahu/provenencia/core/database/subjects/subjectstest"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
)

func TestSubjectPositions(t *testing.T) {
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
	mustSource := func(t *testing.T, c *database.Catalog, title string) sources.Source {
		t.Helper()
		typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
			Key: "book_" + title, Origin: vocabkey.OriginProvenencia, Label: title,
		})
		if err != nil {
			t.Fatal(err)
		}
		src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
			return sources.Create(tx, userID, sources.CreateInput{
				SourceTypeID: typeID, Title: title,
			})
		})
		if err != nil {
			t.Fatal(err)
		}
		return src
	}
	mustPerson := func(t *testing.T, c *database.Catalog, sourceID []byte, label string) []byte {
		t.Helper()
		if err := subjectvocab.Install(c); err != nil {
			t.Fatal(err)
		}
		graph := evidencegraphstest.SoleGraph(t, c, sourceID).ID
		sub := subjectstest.On(t, c, userID, graph, "person", label)
		return sub.ID
	}
	auditCount := func(t *testing.T, c *database.Catalog) int {
		t.Helper()
		db, err := c.DB()
		if err != nil {
			t.Fatal(err)
		}
		var n int
		if err := db.QueryRow(`SELECT COUNT(*) FROM audit_transactions`).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}

	tests := []struct {
		name string
		run  func(t *testing.T, c *database.Catalog)
	}{
		{
			name: "set get round-trip including negatives",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				src := mustSource(t, c, "Deed")
				sub := mustPerson(t, c, src.ID, "Alice")
				before := auditCount(t, c)
				p, err := writes.Call(c, writes.Op{}, func(tx *database.Tx) (subjectpositions.Position, []rowchange.Change, error) {
					return subjectpositions.Set(tx, sub, -3, 7)
				})
				if err != nil {
					t.Fatal(err)
				}
				if p.GridX != -3 || p.GridY != 7 {
					t.Fatalf("%+v", p)
				}
				got, err := subjectpositions.Get(c, sub)
				if err != nil || got.GridX != -3 || got.GridY != 7 {
					t.Fatalf("%v %+v", err, got)
				}
				if auditCount(t, c) != before {
					t.Fatal("set wrote audit")
				}
			},
		},
		{
			name: "set overwrites same subject",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				src := mustSource(t, c, "Deed")
				sub := mustPerson(t, c, src.ID, "Alice")
				if _, err := writes.Call(c, writes.Op{}, func(tx *database.Tx) (subjectpositions.Position, []rowchange.Change, error) {
					return subjectpositions.Set(tx, sub, 1, 1)
				}); err != nil {
					t.Fatal(err)
				}
				if _, err := writes.Call(c, writes.Op{}, func(tx *database.Tx) (subjectpositions.Position, []rowchange.Change, error) {
					return subjectpositions.Set(tx, sub, 2, 3)
				}); err != nil {
					t.Fatal(err)
				}
				got, err := subjectpositions.Get(c, sub)
				if err != nil || got.GridX != 2 || got.GridY != 3 {
					t.Fatalf("%v %+v", err, got)
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				var n int
				if err := db.QueryRow(`SELECT COUNT(*) FROM subject_positions`).Scan(&n); err != nil {
					t.Fatal(err)
				}
				if n != 1 {
					t.Fatalf("rows %d", n)
				}
			},
		},
		{
			name: "get missing is tray; clear returns to tray",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				src := mustSource(t, c, "Deed")
				sub := mustPerson(t, c, src.ID, "Alice")
				_, err := subjectpositions.Get(c, sub)
				if !errors.Is(err, sql.ErrNoRows) {
					t.Fatalf("got %v", err)
				}
				if _, err := writes.Call(c, writes.Op{}, func(tx *database.Tx) (subjectpositions.Position, []rowchange.Change, error) {
					return subjectpositions.Set(tx, sub, 0, 0)
				}); err != nil {
					t.Fatal(err)
				}
				before := auditCount(t, c)
				if _, err := writes.Call(c, writes.Op{}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
					changes, err := subjectpositions.Clear(tx, sub)
					return struct{}{}, changes, err
				}); err != nil {
					t.Fatal(err)
				}
				if _, err := writes.Call(c, writes.Op{}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
					changes, err := subjectpositions.Clear(tx, sub)
					return struct{}{}, changes, err
				}); err != nil {
					t.Fatal(err)
				}
				_, err = subjectpositions.Get(c, sub)
				if !errors.Is(err, sql.ErrNoRows) {
					t.Fatalf("got %v", err)
				}
				if auditCount(t, c) != before {
					t.Fatal("clear wrote audit")
				}
			},
		},
		{
			name: "list by graph scoped; unplaced omitted",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				srcA := mustSource(t, c, "A")
				srcB := mustSource(t, c, "B")
				a1 := mustPerson(t, c, srcA.ID, "A1")
				a2 := mustPerson(t, c, srcA.ID, "A2")
				b1 := mustPerson(t, c, srcB.ID, "B1")
				if _, err := writes.Call(c, writes.Op{}, func(tx *database.Tx) (subjectpositions.Position, []rowchange.Change, error) {
					return subjectpositions.Set(tx, a1, 1, 0)
				}); err != nil {
					t.Fatal(err)
				}
				if _, err := writes.Call(c, writes.Op{}, func(tx *database.Tx) (subjectpositions.Position, []rowchange.

					// a2 stays unplaced
					Change, error) {
					return subjectpositions.Set(tx, b1, 9, 9)
				}); err != nil {
					t.Fatal(err)
				}

				_ = a2
				list, err := subjectpositions.ListByEvidenceGraph(c, evidencegraphstest.SoleGraph(t, c, srcA.ID).ID)
				if err != nil || len(list) != 1 || string(list[0].SubjectID) != string(a1) {
					t.Fatalf("%v %+v", err, list)
				}
			},
		},
		{
			name: "list by graph separates the two graphs of one Source",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				src := mustSource(t, c, "Census")
				ada := mustPerson(t, c, src.ID, "Ada")
				if _, err := artifactstest.Create(c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Page 1"}); err != nil {
					t.Fatal(err)
				}
				def, jones, _ := artifactstest.Split(t, c, userID, src.ID, "Jones")
				bea := subjectstest.On(t, c, userID, jones.ID, "person", "Bea")
				for _, p := range []struct {
					id   []byte
					x, y int64
				}{{ada, 1, 2}, {bea.ID, 3, 4}} {
					if _, err := writes.Call(c, writes.Op{}, func(tx *database.Tx) (subjectpositions.Position, []rowchange.Change, error) {
						return subjectpositions.Set(tx, p.id, p.x, p.y)
					}); err != nil {
						t.Fatal(err)
					}
				}
				for _, tc := range []struct {
					graph []byte
					want  []byte
					x, y  int64
				}{{jones.ID, bea.ID, 3, 4}, {def.ID, ada, 1, 2}} {
					list, err := subjectpositions.ListByEvidenceGraph(c, tc.graph)
					if err != nil || len(list) != 1 || string(list[0].SubjectID) != string(tc.want) || list[0].GridX != tc.x || list[0].GridY != tc.y {
						t.Fatalf("%v %+v", err, list)
					}
				}
			},
		},
		{
			name: "set unknown subject is invalid",
			run: func(t *testing.T, c *database.Catalog) {
				missing := make([]byte, 16)
				missing[0] = 1
				if _, err := writes.Call(c, writes.Op{}, func(tx *database.Tx) (subjectpositions.Position, []rowchange.Change, error) {
					return subjectpositions.Set(tx, missing, 0, 0)
				}); !errors.Is(err, subjectpositions.ErrInvalid) {
					t.Fatalf("got %v", err)
				}
				if _, err := writes.Call(c, writes.Op{}, func(tx *database.Tx) (subjectpositions.Position, []rowchange.Change, error) {
					return subjectpositions.Set(tx, []byte{1}, 0, 0)
				}); !errors.Is(err, subjectpositions.ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "reopen persists position",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				src := mustSource(t, c, "Deed")
				sub := mustPerson(t, c, src.ID, "Alice")
				if _, err := writes.Call(c, writes.Op{}, func(tx *database.Tx) (subjectpositions.Position, []rowchange.Change, error) {
					return subjectpositions.Set(tx, sub, 4, -1)
				}); err != nil {
					t.Fatal(err)
				}
				dir := c.Dir()
				if err := c.Close(); err != nil {
					t.Fatal(err)
				}
				reopened, err := database.Open(dir)
				if err != nil {
					t.Fatal(err)
				}
				defer reopened.Close()
				got, err := subjectpositions.Get(reopened, sub)
				if err != nil || got.GridX != 4 || got.GridY != -1 {
					t.Fatalf("%v %+v", err, got)
				}
				list, err := subjectpositions.ListByEvidenceGraph(reopened, evidencegraphstest.SoleGraph(t, reopened, src.ID).ID)
				if err != nil || len(list) != 1 {
					t.Fatalf("%v len=%d", err, len(list))
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
			closedEarly := false
			defer func() {
				if !closedEarly {
					_ = c.Close()
				}
			}()
			if tt.name == "reopen persists position" {
				closedEarly = true
			}
			tt.run(t, c)
		})
	}
}
