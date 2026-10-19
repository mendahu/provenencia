package matchregistry

import (
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

const modulePath = "github.com/mendahu/provenencia/"

// layerPackages are the packages the matching stack builds on. Each is pure:
// it builds and tests with no build tags and no path to core/database.
// numbers marks a layer package that may hold no numeric literal other than 0
// and 1: the registry owns every tunable. A fact that is not a tunable (a
// calendar fact) belongs in its pure value package, where numbers is false.
// A new layer package or pure value package adds a row.
var layerPackages = []struct {
	dir     string
	numbers bool
}{
	{"core/valuecompare", true},
	{"core/propertycompare", true},
	{"core/propertycompare/propertycomparetest", false},
	{"core/subjectcompare", true},
	{"core/subjectcompare/subjectcomparetest", false},
	{"core/subjectprofile", true},
	{"core/segmentcompare", true},
	{"core/segmentcompare/segmentcomparetest", false},
	{"core/shortlist", true},
	{"core/graphalign", true},
	{"core/matchregistry", false},
	{"core/datevalue", false},
	{"core/datevalue/datevaluetest", false},
	{"core/namevalue", false},
	{"core/namevalue/namevaluetest", false},
	{"core/depictionvalue", false},
	{"core/depictionvalue/depictionvaluetest", false},
	{"core/propertymeta", false},
	{"core/vocabkey", false},
}

func TestLayerPackages(t *testing.T) {
	root := filepath.Join("..", "..")
	for _, p := range layerPackages {
		t.Run(p.dir+"/pure", func(t *testing.T) { checkPure(t, root, p.dir, p.numbers) })
		if p.numbers {
			t.Run(p.dir+"/numbers", func(t *testing.T) { checkNumbers(t, root, p.dir) })
		}
	}
}

// checkPure fails when the package has a file that builds only with a tag, or
// imports, directly or through another package of the module, anything under
// core/database. Following only non-test imports, it also fails when a numbers
// layer reaches core/matchregistry, and when a package other than
// core/subjectprofile reaches core/autoreconcile. It is one walker with three
// rules (D-11).
func checkPure(t *testing.T, root, dir string, numbers bool) {
	ctx := build.Default
	ctx.BuildTags = nil
	pkg, err := ctx.ImportDir(filepath.Join(root, dir), 0)
	if err != nil {
		t.Fatalf("%s: %v", dir, err)
	}
	if len(pkg.GoFiles) == 0 {
		t.Errorf("%s has no Go files without build tags", dir)
	}
	if len(pkg.IgnoredGoFiles) > 0 {
		t.Errorf("%s has files that build only with a tag or a platform: %s", dir, strings.Join(pkg.IgnoredGoFiles, ", "))
	}
	// One walker, three rules (D-11). The catalog rule holds on every hop; the
	// registry and autoreconcile rules hold only along production imports.
	// visited is keyed by import path and prodOnly, so a package first reached
	// through a test import is walked again through a production one.
	type visit struct {
		imp      string
		prodOnly bool
	}
	visited := map[visit]bool{}
	var walk func(pkg *build.Package, chain []string, prodOnly bool)
	walk = func(pkg *build.Package, chain []string, prodOnly bool) {
		step := func(imp string, prodOnly bool) {
			key := visit{imp, prodOnly}
			if !strings.HasPrefix(imp, modulePath) || visited[key] {
				return
			}
			visited[key] = true
			next := append(slices.Clone(chain), imp)
			rest := strings.TrimPrefix(imp, modulePath)
			if rest == "core/database" || strings.HasPrefix(rest, "core/database/") {
				t.Errorf("%s reaches the catalog: %s", dir, strings.Join(next, " -> "))
				return
			}
			if prodOnly && numbers && rest == "core/matchregistry" {
				t.Errorf("%s is a numbers layer and reaches the registry: %s", dir, strings.Join(next, " -> "))
				return
			}
			if prodOnly && rest == "core/autoreconcile" && dir != "core/subjectprofile" {
				t.Errorf("%s reaches core/autoreconcile: %s", dir, strings.Join(next, " -> "))
				return
			}
			dep, err := ctx.ImportDir(filepath.Join(root, rest), 0)
			if err != nil {
				t.Errorf("%s: reading %s: %v", dir, imp, err)
				return
			}
			walk(dep, next, prodOnly)
		}
		for _, imp := range pkg.Imports {
			step(imp, prodOnly)
		}
		for _, imp := range slices.Concat(pkg.TestImports, pkg.XTestImports) {
			step(imp, false)
		}
	}
	walk(pkg, []string{modulePath + dir}, true)
}

// checkNumbers fails on every numeric literal in a non-test file other than 0
// and 1 (0.0 and 1.0 pass), and on any imaginary literal.
func checkNumbers(t *testing.T, root, dir string) {
	entries, err := os.ReadDir(filepath.Join(root, dir))
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		path := filepath.Join(root, dir, name)
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			lit, ok := n.(*ast.BasicLit)
			if !ok {
				return true
			}
			if !literalAllowed(lit) {
				t.Errorf("%s:%d %s: a number belongs in core/matchregistry", filepath.ToSlash(filepath.Join(dir, name)), fset.Position(lit.Pos()).Line, lit.Value)
			}
			return true
		})
	}
}

func literalAllowed(lit *ast.BasicLit) bool {
	switch lit.Kind {
	case token.IMAG:
		return false
	case token.INT:
		v, err := strconv.ParseInt(lit.Value, 0, 64)
		return err == nil && (v == 0 || v == 1)
	case token.FLOAT:
		v, err := strconv.ParseFloat(strings.ReplaceAll(lit.Value, "_", ""), 64)
		return err == nil && (v == 0 || v == 1)
	}
	return true
}
