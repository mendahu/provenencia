package audit

import (
	"bytes"
	"errors"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/effects"
)

// TestResolversCoverEveryEntityType parses non-test Go under core/ for every
// audited entity type (EntityType: "…" literals and deleteimpact rowFacet
// arguments) and wants a resolver for each, so a new audited table decides
// its scope instead of failing at runtime. An entity whose effect is None is
// committed without a revision and is left out of an audited change list, so
// it has no resolver.
func TestResolversCoverEveryEntityType(t *testing.T) {
	root := filepath.Join("..", "..")
	found := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.KeyValueExpr:
				if k, ok := x.Key.(*ast.Ident); ok && k.Name == "EntityType" {
					if s, ok := stringLit(x.Value); ok {
						found[s] = path
					}
				}
			case *ast.CallExpr:
				// rowFacet(parent, table, fkCol, entityType, entityIDCol, cols...)
				if fn, ok := x.Fun.(*ast.Ident); ok && fn.Name == "rowFacet" && len(x.Args) >= 4 {
					if s, ok := stringLit(x.Args[3]); ok {
						found[s] = path
					}
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(found) < 10 {
		t.Fatalf("found only %d entity types; parser walk broken? %v", len(found), found)
	}
	for et, path := range found {
		none, err := effects.AllNone([]rowchange.Change{{EntityType: et}})
		if err != nil {
			t.Fatal(err)
		}
		if none {
			continue
		}
		if _, ok := resolvers[et]; !ok {
			t.Errorf("entity type %q (%s) has no scope resolver in audit/scopes.go", et, path)
		}
	}
}

func stringLit(e ast.Expr) (string, bool) {
	lit, ok := e.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}
	s, err := strconv.Unquote(lit.Value)
	return s, err == nil
}

func TestResolveScopes(t *testing.T) {
	id := func(b byte) []byte { return bytes.Repeat([]byte{b}, 16) }
	uid := func(b []byte) string { u, _ := uuid.FromBytes(b); return u.String() }
	typeID, srcA, srcB, art := id(1), id(2), id(3), id(4)
	goneCitation, note := id(5), id(6)

	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	for _, stmt := range []struct {
		q    string
		args []any
	}{
		{`INSERT INTO source_types (id, key, origin, label) VALUES (?, 'book', 'test', 'Book')`, []any{typeID}},
		{`INSERT INTO sources (id, ref, source_type_id, title) VALUES (?, 'SRC-A', ?, 'A')`, []any{srcA, typeID}},
		{`INSERT INTO sources (id, ref, source_type_id, title) VALUES (?, 'SRC-B', ?, 'B')`, []any{srcB, typeID}},
		{`INSERT INTO artifacts (id, ref, source_id) VALUES (?, 'ART-A', ?)`, []any{art, srcA}},
	} {
		if _, err := db.Exec(stmt.q, stmt.args...); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name    string
		changes []rowchange.Change
		want    [][]byte
	}{
		{
			name: "released note resolves through its deleted citation",
			changes: []rowchange.Change{
				{EntityType: "citation_note", EntityID: note, Action: rowchange.ActionDelete,
					Fields: rowchange.DeletedRow(map[string]any{"citation_id": uid(goneCitation)})},
				{EntityType: "citation", EntityID: goneCitation, Action: rowchange.ActionDelete,
					Fields: rowchange.DeletedRow(map[string]any{"artifact_id": uid(art)})},
			},
			want: [][]byte{srcA},
		},
		{
			name: "row moved between sources scopes to both",
			changes: []rowchange.Change{{EntityType: "artifact", EntityID: art, Action: rowchange.ActionUpdate,
				Fields: map[string]rowchange.FieldDiff{"source_id": {Old: uid(srcB), New: uid(srcA)}}}},
			want: [][]byte{srcA, srcB},
		},
		{
			name: "duplicates collapse",
			changes: []rowchange.Change{
				{EntityType: "source", EntityID: srcA, Action: rowchange.ActionUpdate, Fields: map[string]rowchange.FieldDiff{}},
				{EntityType: "artifact", EntityID: art, Action: rowchange.ActionUpdate, Fields: map[string]rowchange.FieldDiff{}},
			},
			want: [][]byte{srcA},
		},
		{
			name:    "unresolvable parent is no scope",
			changes: []rowchange.Change{{EntityType: "citation", EntityID: goneCitation, Action: rowchange.ActionUpdate, Fields: map[string]rowchange.FieldDiff{}}},
		},
		{
			name:    "vocabulary has no scope",
			changes: []rowchange.Change{{EntityType: "source_type", EntityID: typeID, Action: rowchange.ActionUpdate, Fields: map[string]rowchange.FieldDiff{}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx, err := db.Begin()
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback() }()
			got, err := resolveScopes(tx, tt.changes)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %d scopes %v want %d", len(got), got, len(tt.want))
			}
			for _, w := range tt.want {
				if !containsScope(got, Scope{Type: ScopeSource, ID: w}) {
					t.Fatalf("missing %x in %v", w, got)
				}
			}
		})
	}

	t.Run("note resolver alone reaches the ghost", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback() }()
		r := scopeResolver{tx: tx, ghosts: ghostMap{
			ghostKey("citation", goneCitation): {"artifact_id": uid(art)},
		}}
		got, err := resolvers["citation_note"](r, rowchange.Change{EntityType: "citation_note", EntityID: note,
			Action: rowchange.ActionDelete, Fields: rowchange.DeletedRow(map[string]any{"citation_id": uid(goneCitation)})})
		if err != nil || len(got) != 1 || !bytes.Equal(got[0].ID, srcA) {
			t.Fatalf("got %v %v", got, err)
		}
	})

	t.Run("unknown entity type rejected by Record", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback() }()
		_, err = Record(tx, Revision{ActionType: "x", CreatedAt: "2026-10-04T00:00:00Z", Changes: []rowchange.Change{{
			EntityType: "mystery", EntityID: id(9), Action: rowchange.ActionCreate, Fields: map[string]rowchange.FieldDiff{},
		}}})
		if !errors.Is(err, ErrInvalid) {
			t.Fatalf("want ErrInvalid got %v", err)
		}
	})
}
