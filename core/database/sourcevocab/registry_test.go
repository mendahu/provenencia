package sourcevocab

import (
	"slices"
	"testing"

	"github.com/Jake-Robins-Studios/provenencia/core/citationformat"
	"github.com/Jake-Robins-Studios/provenencia/core/database/metadatafields"
	"github.com/Jake-Robins-Studios/provenencia/core/database/sourcetypes"
	"github.com/Jake-Robins-Studios/provenencia/core/vocabkey"
)

// copyFields is the "copy" group of the map's citation vocabulary, in order:
// the fields about the copy a researcher used, usually set on an Artifact.
var copyFields = []string{
	"format", "provider", "database_title", "url", "accessed_date", "citing",
	"film_publication", "roll", "film_number", "item_number",
}

// itemFields is the "item" group: what one Evidence graph is about, usually set
// on the graph.
var itemFields = []string{
	"item", "entry_number", "line_number", "dwelling_number", "family_number",
	"schedule_number", "folio", "column", "record_id",
}

// recordTypes are the types whose Artifacts suggest every copy field.
var recordTypes = []string{
	"census", "birth_certificate", "marriage_certificate", "death_certificate",
	"church_register", "record_book", "passenger_list", "case_file", "manuscript",
	"government_file",
}

// unreadFields are the two seeded fields no seeded format reads:
// database_title (suggested on record types' Artifacts by the copy exception)
// and record_id (there for researchers' own formats).
var unreadFields = []string{"database_title", "record_id"}

// usualLevel is where a field is usually set.
func usualLevel(key string) vocabkey.Level {
	switch {
	case slices.Contains(copyFields, key):
		return vocabkey.LevelArtifact
	case slices.Contains(itemFields, key):
		return vocabkey.LevelGraph
	}
	return vocabkey.LevelSource
}

func mustParseForm(t *testing.T, typeKey, text string, form citationformat.Form) *citationformat.Template {
	t.Helper()
	tpl, diags := citationformat.Parse(text, form)
	if len(diags) > 0 {
		t.Fatalf("%s %s: %v", typeKey, form, diags)
	}
	return tpl
}

// derivedSuggestions is the suggestion rule applied to a type's full format:
// each field once, at its usual level when the format reads it there and else
// at the level it first appears at, in the order the format first reads it; the
// record types add the copy fields their Artifacts do not have yet.
func derivedSuggestions(t *testing.T, st seedType) map[vocabkey.Level][]string {
	t.Helper()
	var order []string
	read := map[string][]vocabkey.Level{}
	// The full format's fields, then the ones the list and the short format
	// need outside any group that the full format does not use.
	full := true
	var fullKeys map[string]bool
	for _, form := range []citationformat.Form{citationformat.FormFull, citationformat.FormList, citationformat.FormShort} {
		tpl := mustParseForm(t, st.Key, st.Formats.Text(form), form)
		for _, req := range citationformat.Requirements(tpl) {
			if !full && req.Need != citationformat.NeedNeeded {
				continue
			}
			// A chain is met when one option is suggested, so a later form's
			// chain adds nothing when the full format already uses an option.
			met := false
			for _, target := range req.Targets {
				met = met || (!full && fullKeys[target.Field.Key])
			}
			if met {
				continue
			}
			for _, target := range req.Targets {
				key := target.Field.Key
				if key == "" {
					continue
				}
				if _, ok := read[key]; !ok {
					order = append(order, key)
				}
				read[key] = append(read[key], target.Level)
			}
		}
		if full {
			full = false
			fullKeys = map[string]bool{}
			for key := range read {
				fullKeys[key] = true
			}
		}
	}
	out := map[vocabkey.Level][]string{}
	for _, key := range order {
		level := read[key][0]
		if slices.Contains(read[key], usualLevel(key)) {
			level = usualLevel(key)
		}
		out[level] = append(out[level], key)
	}
	if slices.Contains(recordTypes, st.Key) {
		for _, key := range copyFields {
			if !slices.Contains(out[vocabkey.LevelArtifact], key) {
				out[vocabkey.LevelArtifact] = append(out[vocabkey.LevelArtifact], key)
			}
		}
	}
	return out
}

func TestSeedSuggestionsFollowFormats(t *testing.T) {
	types := map[string]seedType{}
	for _, st := range seedTypes {
		types[st.Key] = st
	}
	fields := map[string]bool{}
	for _, f := range seedFields {
		fields[f.Key] = true
	}
	seen := map[string]bool{}
	got := map[string]map[vocabkey.Level][]string{}
	for _, s := range seedSuggestions {
		id := s.TypeKey + "/" + string(s.Level)
		if seen[id] {
			t.Errorf("%s appears twice", id)
		}
		seen[id] = true
		if _, ok := types[s.TypeKey]; !ok {
			t.Errorf("%s names an unknown type", id)
		}
		if len(s.FieldKeys) == 0 {
			t.Errorf("%s has no field", id)
		}
		for _, key := range s.FieldKeys {
			if !fields[key] {
				t.Errorf("%s names unknown field %s", id, key)
			}
		}
		if got[s.TypeKey] == nil {
			got[s.TypeKey] = map[vocabkey.Level][]string{}
		}
		got[s.TypeKey][s.Level] = s.FieldKeys
	}
	for _, st := range seedTypes {
		want := derivedSuggestions(t, st)
		for _, level := range vocabkey.MetadataLevels() {
			if !slices.Equal(got[st.Key][level], want[level]) {
				t.Errorf("%s %s suggests %v, its formats give %v", st.Key, level, got[st.Key][level], want[level])
			}
		}
	}
}

func TestSeedFieldsAreRead(t *testing.T) {
	keys := map[string]bool{}
	for _, f := range seedFields {
		if keys[f.Key] {
			t.Errorf("field %s is seeded twice", f.Key)
		}
		keys[f.Key] = true
		if f.Label == "" || f.Description == "" {
			t.Errorf("field %s lacks a label or description", f.Key)
		}
		wantType := metadatafields.DataTypeText
		if f.Key == "url" {
			wantType = metadatafields.DataTypeURL
		}
		if f.DataType != wantType {
			t.Errorf("field %s has data type %s, want %s", f.Key, f.DataType, wantType)
		}
		wantUnknown := ""
		if f.Key == "publication_place" || f.Key == "publisher" {
			wantUnknown = "n.p."
		}
		if f.UnknownText != wantUnknown {
			t.Errorf("field %s has unknown text %q, want %q", f.Key, f.UnknownText, wantUnknown)
		}
	}
	read := map[string]bool{}
	for _, st := range seedTypes {
		for _, form := range citationformat.Forms() {
			tpl := mustParseForm(t, st.Key, st.Formats.Text(form), form)
			for _, f := range tpl.Fields() {
				read[f.Key] = true
			}
		}
	}
	for key := range read {
		if !keys[key] {
			t.Errorf("a seeded format reads %s, which is not seeded", key)
		}
	}
	for _, f := range seedFields {
		if !read[f.Key] && !slices.Contains(unreadFields, f.Key) {
			t.Errorf("field %s is read by no seeded format", f.Key)
		}
		if read[f.Key] && slices.Contains(unreadFields, f.Key) {
			t.Errorf("field %s is in unreadFields but a seeded format reads it", f.Key)
		}
	}
}

func TestSeedTypes(t *testing.T) {
	keys := map[string]bool{}
	for _, st := range seedTypes {
		if keys[st.Key] {
			t.Errorf("type %s is seeded twice", st.Key)
		}
		keys[st.Key] = true
		if st.Label == "" || st.Description == "" {
			t.Errorf("type %s lacks a label or description", st.Key)
		}
		if _, ok := sourcetypes.ValidIconKeys[st.IconKey]; !ok {
			t.Errorf("type %s has unknown icon %q", st.Key, st.IconKey)
		}
		for _, form := range citationformat.Forms() {
			if st.Formats.Text(form) == "" {
				t.Errorf("type %s has no %s format", st.Key, form)
				continue
			}
			tpl := mustParseForm(t, st.Key, st.Formats.Text(form), form)
			if form != citationformat.FormList {
				continue
			}
			for _, req := range citationformat.Requirements(tpl) {
				for _, target := range req.Targets {
					if target.Level != vocabkey.LevelSource {
						t.Errorf("type %s list format reads the %s level", st.Key, target.Level)
					}
				}
			}
		}
	}
}

// TestSeededFormatsNeedOnlySuggestedFields: a field a type's formats read
// outside any group is one a Source of that type must have, so the type page
// has to offer it: at least one option of each such requirement is suggested
// at its level.
func TestSeededFormatsNeedOnlySuggestedFields(t *testing.T) {
	suggested := map[string]bool{}
	for _, s := range seedSuggestions {
		for _, key := range s.FieldKeys {
			suggested[s.TypeKey+"/"+string(s.Level)+"/"+key] = true
		}
	}
	for _, st := range seedTypes {
		var tpls []*citationformat.Template
		for _, form := range citationformat.Forms() {
			tpls = append(tpls, mustParseForm(t, st.Key, st.Formats.Text(form), form))
		}
		for _, req := range citationformat.Requirements(tpls...) {
			if req.Need != citationformat.NeedNeeded {
				continue
			}
			var options []string
			met, hasField := false, false
			for _, target := range req.Targets {
				if target.Field.Key == "" {
					continue
				}
				hasField = true
				id := st.Key + "/" + string(target.Level) + "/" + target.Field.Key
				options = append(options, id)
				met = met || suggested[id]
			}
			if hasField && !met {
				t.Errorf("%s: a format needs one of %v, none suggested", st.Key, options)
			}
		}
	}
}

// TestSeedPlaceFields: each seeded type that names a place suggests the field
// that says what the place is. Only the government records, whose place is
// the area whose office keeps them, use jurisdiction.
func TestSeedPlaceFields(t *testing.T) {
	want := map[string]string{
		"census":               "jurisdiction",
		"birth_certificate":    "jurisdiction",
		"marriage_certificate": "jurisdiction",
		"death_certificate":    "jurisdiction",
		"record_book":          "jurisdiction",
		"case_file":            "jurisdiction",
		"statute":              "jurisdiction",
		"newspaper_item":       "publication_place",
		"church_register":      "church_location",
		"cemetery_marker":      "cemetery_location",
		"interview":            "interview_place",
		"oral_testimony":       "interview_place",
		"correspondence":       "written_from",
	}
	for _, s := range seedSuggestions {
		if s.Level != vocabkey.LevelSource {
			continue
		}
		key, ok := want[s.TypeKey]
		if ok && !slices.Contains(s.FieldKeys, key) {
			t.Errorf("type %s suggests %v at the source, want %s among them", s.TypeKey, s.FieldKeys, key)
		}
		if key != "jurisdiction" && slices.Contains(s.FieldKeys, "jurisdiction") {
			t.Errorf("type %s suggests jurisdiction, which only government records use", s.TypeKey)
		}
	}
}
