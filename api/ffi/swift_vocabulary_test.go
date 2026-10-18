package ffi

import (
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/locator"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// The app reads Go's wire vocabulary, and mirrors a few Go rules, in Swift.
// This test reads those Swift files so a rename or a rule change on either
// side fails here, not as a silently missing badge, row or match.
const (
	swiftVocabulary = "../../macos/App/Platform/ConclusionVocabulary.swift"
	swiftValueKinds = "../../macos/App/Platform/InterpretationValueKinds.swift"
	swiftMarks      = "../../macos/App/DesignSystem/Recipes/Marks/PVMark.swift"
	swiftGraphKinds = "../../macos/App/Features/Workspace/Session/SourceGraphSnapshot.swift"

	swiftRegionGeometry = "../../macos/App/Features/ArtifactViewer/ArtifactRegionGeometry.swift"
	swiftTextMatch      = "../../macos/App/Platform/TextMatch.swift"
)

// swiftEnumBody returns the source between `enum <name>` and its closing
// brace at column 0.
func swiftEnumBody(t *testing.T, path, name string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	src := string(b)
	start := strings.Index(src, "enum "+name)
	if start < 0 {
		t.Fatalf("%s: no enum %s", path, name)
	}
	end := strings.Index(src[start:], "\n}\n")
	if end < 0 {
		t.Fatalf("%s: enum %s does not close", path, name)
	}
	return src[start : start+end]
}

var (
	swiftFixturePair = regexp.MustCompile(`\("([^"\\]*)", "([^"\\]*)"\)`)
	swiftFoldEntry   = regexp.MustCompile(`"(.)": "([^"\\]+)"`)
)

// swiftFixturePairs returns the ("a", "b") literal pairs in the body of
// `enum <enum>`. A fixture line with a backslash is refused: the reader does
// not decode escapes.
func swiftFixturePairs(t *testing.T, path, enum string) [][2]string {
	t.Helper()
	body := swiftEnumBody(t, path, enum)
	for _, line := range strings.Split(body, "\n") {
		if strings.Contains(line, `\`) && swiftFixturePair.MatchString(line) {
			t.Fatalf("%s: fixture line has a backslash escape: %s", path, line)
		}
	}
	var out [][2]string
	for _, m := range swiftFixturePair.FindAllStringSubmatch(body, -1) {
		out = append(out, [2]string{m[1], m[2]})
	}
	if len(out) == 0 {
		t.Fatalf("%s: enum %s has no fixture pairs", path, enum)
	}
	return out
}

// installedCatalog is a fresh catalog with the seeded vocabulary installed,
// closed when the test ends.
func installedCatalog(t *testing.T) *database.Catalog {
	t.Helper()
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	if err := subjectvocab.Install(c); err != nil {
		t.Fatal(err)
	}
	return c
}

func matches(re *regexp.Regexp, body string) []string {
	var out []string
	for _, m := range re.FindAllStringSubmatch(body, -1) {
		out = append(out, m[1])
	}
	sort.Strings(out)
	return out
}

func sorted(in ...string) []string {
	out := append([]string(nil), in...)
	sort.Strings(out)
	return out
}

var (
	wireCase    = regexp.MustCompile(`case "([a-z_]*)": self = `)
	staticLet   = regexp.MustCompile(`static let \w+ = "([^"]*)"`)
	rawCase     = regexp.MustCompile(`(?m)^\s+case (\w+)\s*$`)
	unrecogniz  = regexp.MustCompile(`case unrecognized\(String\)`)
	subjectMark = regexp.MustCompile(`case \w+ = "(subject_[a-z_]+)"`)
)

// subjectMarkSymbols is the sorted IconSymbol of every registry subject type;
// Swift's PVMarkKey(subjectTypeKey:) builds "subject_<key>" and needs a case
// for each.
func subjectMarkSymbols() []string {
	var out []string
	for _, info := range subjectvocab.AllTypes() {
		out = append(out, info.Presentation.IconSymbol)
	}
	sort.Strings(out)
	return out
}

func TestSwiftVocabularyMatchesGo(t *testing.T) {
	tests := []struct {
		name string
		path string
		enum string
		re   *regexp.Regexp
		want []string
	}{
		{
			name: "field states", path: swiftVocabulary, enum: "ReconciledState", re: wireCase,
			want: sorted(string(autoreconcile.StateEmpty), string(autoreconcile.StateSingle), string(autoreconcile.StateMerged),
				string(autoreconcile.StateMixed), string(autoreconcile.StateMultiple), string(autoreconcile.StateConcluded)),
		},
		{
			name: "outcome reasons", path: swiftVocabulary, enum: "ReconcilerReason", re: wireCase,
			want: sorted(string(autoreconcile.ReasonKept), string(autoreconcile.ReasonFolded), string(autoreconcile.ReasonOutvoted),
				string(autoreconcile.ReasonWeak), string(autoreconcile.ReasonDenied), string(autoreconcile.ReasonProvisional),
				string(autoreconcile.ReasonNoEvidence), string(autoreconcile.ReasonAgainst)),
		},
		{
			name: "date qualifiers", path: swiftVocabulary, enum: "DateQualifier", re: staticLet,
			want: sorted(datevalue.QualifierABT, datevalue.QualifierBEF, datevalue.QualifierAFT),
		},
		{
			name: "value types", path: swiftValueKinds, enum: "PropertyValueType", re: rawCase,
			want: sorted(propertymeta.ValueTypeText, propertymeta.ValueTypeInteger, propertymeta.ValueTypeDate,
				propertymeta.ValueTypeName, propertymeta.ValueTypeSubject, propertymeta.ValueTypeTerm, propertymeta.ValueTypeDepiction),
		},
		{
			name: "primary kinds", path: swiftGraphKinds, enum: "EvidencePrimaryKind", re: rawCase,
			want: sorted(vocabkey.PrimaryKinds()...),
		},
		{
			name: "subject marks", path: swiftMarks, enum: "PVMarkKey", re: subjectMark,
			want: subjectMarkSymbols(),
		},
		{
			name: "region point cap", path: swiftRegionGeometry, enum: "ArtifactRegionGeometry",
			re:   regexp.MustCompile(`static let maxRegionPoints = (\d+)`),
			want: []string{strconv.Itoa(locator.MaxRegionPoints)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := swiftEnumBody(t, tt.path, tt.enum)
			if got := matches(tt.re, body); strings.Join(got, ",") != strings.Join(tt.want, ",") {
				t.Fatalf("Swift %s %q, Go %q", tt.enum, got, tt.want)
			}
		})
	}
	t.Run("states and reasons tolerate a value this build doesn't know", func(t *testing.T) {
		for _, enum := range []string{"ReconciledState", "ReconcilerReason"} {
			if !unrecogniz.MatchString(swiftEnumBody(t, swiftVocabulary, enum)) {
				t.Fatalf("%s has no unrecognized case", enum)
			}
		}
	})
	t.Run("seeded Property keys are seeded", func(t *testing.T) {
		keys := matches(staticLet, swiftEnumBody(t, swiftVocabulary, "SeededPropertyKey"))
		if len(keys) == 0 {
			t.Fatal("no keys read")
		}
		c := installedCatalog(t)
		for _, key := range keys {
			if _, err := properties.Lookup(c, key, vocabkey.OriginProvenencia); err != nil {
				t.Errorf("Swift names %q, which is not a seeded Property: %v", key, err)
			}
		}
	})
	t.Run("seeded term keys are seeded", func(t *testing.T) {
		c := installedCatalog(t)
		for _, tc := range []struct{ enum, property string }{
			{"SeededGenderTerm", "gender"},
			{"SeededSexAtBirthTerm", "sex_at_birth"},
		} {
			keys := matches(staticLet, swiftEnumBody(t, swiftVocabulary, tc.enum))
			if len(keys) == 0 {
				t.Fatalf("%s: no keys read", tc.enum)
			}
			prop, err := properties.Lookup(c, tc.property, vocabkey.OriginProvenencia)
			if err != nil {
				t.Fatal(err)
			}
			for _, key := range keys {
				if _, err := propertyterms.Lookup(c, prop.ID, key, vocabkey.OriginProvenencia); err != nil {
					t.Errorf("Swift names %s term %q, which is not seeded: %v", tc.property, key, err)
				}
			}
		}
	})
}

func TestSwiftTextMatchMirrorsTextKey(t *testing.T) {
	words := matchregistry.Default().Value.Words
	for _, p := range swiftFixturePairs(t, swiftTextMatch, "TextMatch") {
		want, ok := valuecompare.TextKey(p[0], valuecompare.Equivalent, words)
		if !ok {
			want = ""
		}
		if want != p[1] {
			t.Errorf("TextMatch.fixtures %q: Swift says %q, Go TextKey says %q", p[0], p[1], want)
		}
	}
	folds := map[rune]string{}
	for _, m := range swiftFoldEntry.FindAllStringSubmatch(swiftEnumBody(t, swiftTextMatch, "TextMatch"), -1) {
		folds[[]rune(m[1])[0]] = m[2]
	}
	if !reflect.DeepEqual(folds, words.Folds) {
		t.Errorf("TextMatch.letterFolds %v, Go Words.Folds %v", folds, words.Folds)
	}
}
