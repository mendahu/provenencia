package ffi

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
)

// The app reads Go's wire vocabulary through one Swift file. This test reads
// that file so a rename on either side fails here, not as a silently missing
// badge or row.
const (
	swiftVocabulary = "../../macos/App/Platform/ConclusionVocabulary.swift"
	swiftValueKinds = "../../macos/App/Platform/InterpretationValueKinds.swift"
	swiftMarks      = "../../macos/App/DesignSystem/Recipes/Marks/PVMark.swift"
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
			want: sorted(datevalues.QualifierABT, datevalues.QualifierBEF, datevalues.QualifierAFT),
		},
		{
			name: "value types", path: swiftValueKinds, enum: "PropertyValueType", re: rawCase,
			want: sorted(properties.ValueTypeText, properties.ValueTypeInteger, properties.ValueTypeDate,
				properties.ValueTypeName, properties.ValueTypeSubject, properties.ValueTypeTerm, properties.ValueTypeDepiction),
		},
		{
			name: "subject marks", path: swiftMarks, enum: "PVMarkKey", re: subjectMark,
			want: subjectMarkSymbols(),
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
			if _, err := properties.Lookup(c, key, properties.OriginProvenencia); err != nil {
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
			prop, err := properties.Lookup(c, tc.property, properties.OriginProvenencia)
			if err != nil {
				t.Fatal(err)
			}
			for _, key := range keys {
				if _, err := propertyterms.Lookup(c, prop.ID, key, properties.OriginProvenencia); err != nil {
					t.Errorf("Swift names %s term %q, which is not seeded: %v", tc.property, key, err)
				}
			}
		}
	})
}
