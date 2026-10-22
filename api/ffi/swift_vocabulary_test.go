package ffi

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/locator"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/objectpath"
	"github.com/mendahu/provenencia/core/propertycompare"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/raster"
	"github.com/mendahu/provenencia/core/segmentcompare"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// The app reads Go's wire vocabulary, and mirrors some of Go's rules and
// limits, through Swift files this test reads. A rename or a changed rule or
// limit on either side fails here, not as a silently missing badge, row or
// match, or a wrongly refused input.
const (
	swiftVocabulary = "../../macos/App/Platform/ConclusionVocabulary.swift"
	swiftValueKinds = "../../macos/App/Platform/InterpretationValueKinds.swift"
	swiftMarks      = "../../macos/App/DesignSystem/Recipes/Marks/PVMark.swift"
	swiftGraphKinds = "../../macos/App/Features/Workspace/Session/SourceGraphSnapshot.swift"
	swiftCatalogMut = "../../macos/App/Features/Workspace/Session/CatalogMutation.swift"

	swiftRegionGeometry = "../../macos/App/Features/ArtifactViewer/ArtifactRegionGeometry.swift"
	swiftProjectFiles   = "../../macos/App/Platform/ProjectFiles.swift"
	swiftTextMatch      = "../../macos/App/Platform/TextMatch.swift"
	swiftThumbnailCache = "../../macos/App/Platform/ThumbnailCache.swift"
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
	swiftFixturePair = regexp.MustCompile(`\("((?:[^"\\]|\\.)*)", "((?:[^"\\]|\\.)*)"\)`)
	swiftFoldEntry   = regexp.MustCompile(`"(.)": "([^"\\]+)"`)
)

// swiftArrayLiteral returns the array literal assigned to `static let <name>`
// inside an enum body: from its `= [` to the `]` alone on an indented line.
func swiftArrayLiteral(t *testing.T, body, name string) string {
	t.Helper()
	start := strings.Index(body, "static let "+name)
	if start < 0 {
		t.Fatalf("no static let %s", name)
	}
	open := strings.Index(body[start:], "= [")
	if open < 0 {
		t.Fatalf("static let %s is not an array literal", name)
	}
	rest := body[start+open:]
	end := strings.Index(rest, "\n    ]")
	if end < 0 {
		t.Fatalf("static let %s does not close", name)
	}
	return rest[:end]
}

// readSwiftFixturePairs returns the ("a", "b") literal pairs of the array
// `static let <decl>` in `enum <enum>`. A pair with a backslash is an error:
// the reader does not decode escapes, and a skipped row would be an
// unchecked mirror.
func readSwiftFixturePairs(t *testing.T, path, enum, decl string) ([][2]string, error) {
	t.Helper()
	literal := swiftArrayLiteral(t, swiftEnumBody(t, path, enum), decl)
	var out [][2]string
	for _, m := range swiftFixturePair.FindAllStringSubmatch(literal, -1) {
		if strings.Contains(m[1], `\`) || strings.Contains(m[2], `\`) {
			return nil, fmt.Errorf("%s: fixture has a backslash escape: %s", path, m[0])
		}
		out = append(out, [2]string{m[1], m[2]})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%s: %s.%s has no fixture pairs", path, enum, decl)
	}
	return out, nil
}

func swiftFixturePairs(t *testing.T, path, enum, decl string) [][2]string {
	t.Helper()
	out, err := readSwiftFixturePairs(t, path, enum, decl)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestSwiftFixturePairsRefusesEscapes(t *testing.T) {
	tests := []struct {
		name    string
		array   string
		want    [][2]string
		wantErr string
	}{
		{
			name:    "escape in a text",
			array:   `("God\u{2010}father", "godfather"), ("a", "b")`,
			wantErr: "backslash",
		},
		{
			name:    "escape in a key",
			array:   `("a", "b\\c")`,
			wantErr: "backslash",
		},
		{
			name:  "plain pairs",
			array: `("a", "b"), ("C d", "cd")`,
			want:  [][2]string{{"a", "b"}, {"C d", "cd"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "E.swift")
			src := "enum E {\n    static let fixtures: [(a: String, b: String)] = [\n        " +
				tt.array + ",\n    ]\n}\n"
			if err := os.WriteFile(path, []byte(src), 0o600); err != nil {
				t.Fatal(err)
			}
			got, err := readSwiftFixturePairs(t, path, "E", "fixtures")
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("got %v, %v; want an error containing %q", got, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
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
	listCase    = regexp.MustCompile(`case \w+ = "([a-z_]+)"`)
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

// listStrings is every effects.List wire key; Swift's CatalogList decodes
// WriteEffects.lists with them.
func listStrings() []string {
	var out []string
	for _, l := range effects.Lists {
		out = append(out, string(l))
	}
	return out
}

// sortedRelations is every valuecompare relation, sorted.
func sortedRelations() []string {
	var out []string
	for _, r := range valuecompare.Relations() {
		out = append(out, string(r))
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
			name: "promote reasons", path: swiftVocabulary, enum: "PromoteReasonCode", re: wireCase,
			want: sorted(string(graphalign.ReasonVia), string(graphalign.ReasonDecided), string(graphalign.ReasonAgrees),
				string(graphalign.ReasonWeak), string(graphalign.ReasonTaken), string(graphalign.ReasonNoMatch),
				string(graphalign.ReasonEmpty), string(graphalign.ReasonNeighborPlaced), string(graphalign.ReasonNeighborMoved),
				string(graphalign.ReasonNeighborReleased)),
		},
		{
			name: "evidence outcomes", path: swiftVocabulary, enum: "EvidenceOutcome", re: wireCase,
			want: sorted(string(propertycompare.OutcomeAgree), string(propertycompare.OutcomePartial),
				string(propertycompare.OutcomeConflict), string(propertycompare.OutcomeUnknown)),
		},
		{
			name: "evidence property causes", path: swiftVocabulary, enum: "EvidencePropertyCause", re: wireCase,
			want: sorted("", string(propertycompare.CauseRelation), string(propertycompare.CauseNegative),
				string(propertycompare.CauseNoValues), string(propertycompare.CauseCardinality)),
		},
		{
			name: "value relations", path: swiftVocabulary, enum: "ValueRelation", re: wireCase,
			want: sorted(append(sortedRelations(), "")...),
		},
		{
			name: "value causes", path: swiftVocabulary, enum: "ValueCause", re: wireCase,
			want: sorted("", string(valuecompare.CauseExact), string(valuecompare.CauseCase), string(valuecompare.CauseFold),
				string(valuecompare.CauseAlias), string(valuecompare.CauseAbbreviation), string(valuecompare.CauseInitial),
				string(valuecompare.CauseSpelling), string(valuecompare.CausePrecision), string(valuecompare.CauseMissingPart),
				string(valuecompare.CauseOverlap), string(valuecompare.CauseGap), string(valuecompare.CauseBeyondTolerance),
				string(valuecompare.CauseGenerationConflict), string(valuecompare.CauseEmpty), string(valuecompare.CausePhraseOnly),
				string(valuecompare.CauseBelowMinimum), string(valuecompare.CauseNoMatch), string(valuecompare.CauseTooLong),
				string(valuecompare.CauseOrder), string(valuecompare.CauseRole), string(valuecompare.CauseUnsupported), string(valuecompare.CauseUnjudged)),
		},
		{
			name: "segment causes", path: swiftVocabulary, enum: "SegmentCause", re: wireCase,
			want: sorted("", string(segmentcompare.CauseNotCorresponding), string(segmentcompare.CauseGated),
				string(segmentcompare.CausePlacedApart), string(segmentcompare.CauseUnplacedDiffers),
				string(segmentcompare.CauseUnplacedUnsure)),
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
			name: "catalog-wide lists", path: swiftCatalogMut, enum: "CatalogList", re: listCase,
			want: sorted(listStrings()...),
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
		{
			name: "objects folder", path: swiftProjectFiles, enum: "ProjectFiles",
			re:   regexp.MustCompile(`static let objectsDir = "(\w+)"`),
			want: []string{objectpath.Dir},
		},
		{
			name: "thumbnail decode edge", path: swiftThumbnailCache, enum: "ThumbnailDecodePolicy",
			re:   regexp.MustCompile(`static let maxDeclaredEdge = (\d+)`),
			want: []string{strconv.Itoa(raster.MaxEdgeLimit)},
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
		for _, enum := range []string{"ReconciledState", "ReconcilerReason", "PromoteReasonCode",
			"EvidenceOutcome", "EvidencePropertyCause", "ValueRelation", "ValueCause", "SegmentCause"} {
			if !unrecogniz.MatchString(swiftEnumBody(t, swiftVocabulary, enum)) {
				t.Fatalf("%s has no unrecognized case", enum)
			}
		}
	})
	t.Run("the product's origin is Go's", func(t *testing.T) {
		got := matches(staticLet, swiftEnumBody(t, swiftVocabulary, "CatalogOrigin"))
		want := []string{vocabkey.OriginProvenencia, vocabkey.OriginUser}
		if strings.Join(got, ",") != strings.Join(want, ",") {
			t.Fatalf("Swift CatalogOrigin %q, Go %q", got, want)
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
			{"SeededRoleTerm", "role"},
			{"SeededEventTypeTerm", "event_type"},
			{"SeededRelationshipTypeTerm", "relationship_type"},
			{"SeededPlaceRelationshipTypeTerm", "place_relationship_type"},
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
	for _, p := range swiftFixturePairs(t, swiftTextMatch, "TextMatch", "fixtures") {
		want, ok := valuecompare.TextKey(p[0], valuecompare.Equivalent, words)
		if !ok {
			want = ""
		}
		if want != p[1] {
			t.Errorf("TextMatch.fixtures %q: Swift says %q, Go TextKey says %q", p[0], p[1], want)
		}
	}
	folds := map[rune]string{}
	for _, m := range swiftFoldEntry.FindAllStringSubmatch(swiftArrayLiteral(t, swiftEnumBody(t, swiftTextMatch, "TextMatch"), "letterFolds"), -1) {
		folds[[]rune(m[1])[0]] = m[2]
	}
	if !reflect.DeepEqual(folds, words.Folds) {
		t.Errorf("TextMatch.letterFolds %v, Go Words.Folds %v", folds, words.Folds)
	}
}
