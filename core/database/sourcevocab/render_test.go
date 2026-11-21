package sourcevocab

import (
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/Jake-Robins-Studios/provenencia/core/citationformat"
	"github.com/Jake-Robins-Studios/provenencia/core/database"
	"github.com/Jake-Robins-Studios/provenencia/core/database/metadatafields"
	"github.com/Jake-Robins-Studios/provenencia/core/database/sourcetypes"
	"github.com/Jake-Robins-Studios/provenencia/core/datevalue/datevaluetest"
	"github.com/Jake-Robins-Studios/provenencia/core/vocabkey"
	"github.com/google/uuid"
)

// update rewrites testdata/seed-citations.txt from the current render:
// go test -tags fts5 -run TestSeedCitations ./core/database/sourcevocab -update
var update = flag.Bool("update", false, "rewrite testdata/seed-citations.txt")

const seedCitationsFile = "testdata/seed-citations.txt"

// fieldValues is what the full fixture answers for each seeded field, at every
// level. Every value is invented. creator is per type (fixtureCreators).
var fieldValues = map[string]string{
	"creator":              "Calder County Clerk",
	"author":               "Mabel Thorne",
	"author_sort":          "Thorne, Mabel",
	"editor":               "Owen Pryce",
	"compiler":             "Lena Marsh",
	"translator":           "Ivo Karsk",
	"photographer":         "Silas Pell",
	"studio":               "Pell Studio",
	"interviewer":          "Ruth Abernant",
	"interviewee":          "Clara Venn",
	"recipient":            "Harold Venn",
	"testing_company":      "GeneLine Labs",
	"software":             "FamilyLedger 4.2",
	"jurisdiction":         "Calder County, Westmark",
	"church_location":      "Port Ellis, Westmark",
	"cemetery_location":    "Port Ellis, Westmark",
	"interview_place":      "Port Ellis, Westmark",
	"written_from":         "Port Ellis, Westmark",
	"published_in":         "Calder Valley Gazette",
	"publisher":            "Ashby Press",
	"publication_place":    "Port Ellis",
	"edition":              "2nd",
	"series":               "Calder County Records Series",
	"volume":               "3",
	"issue":                "12",
	"page_range":           "41–58",
	"site_name":            "Calder Archives Online",
	"short_title":          "Calder Register",
	"coverage":             "1850–1899",
	"repository":           "Calder County Archives, Port Ellis",
	"collection":           "Venn Family Papers",
	"record_group":         "45",
	"call_number":          "MS 1123",
	"box":                  "4",
	"folder":               "17",
	"document_number":      "1880-0412",
	"book":                 "Deed Book F",
	"case_number":          "no. 2231",
	"enumeration_district": "236",
	"schedule":             "population schedule",
	"piece":                "1047",
	"chapter":              "88",
	"section":              "4",
	"lot":                  "112",
	"kit_number":           "GL-55120",
	"match_identifier":     "R. Venn",
	"holder":               "Clara Venn, Port Ellis, 2025",
	"medium":               "cabinet card",
	"format":               "digital image",
	"provider":             "Westmark Heritage",
	"database_title":       "Calder County Censuses, 1850–1900",
	"url":                  "https://example.org/calder/1880",
	"accessed_date":        "12 March 2025",
	"citing":               "county film 7, item 3",
	"film_publication":     "Westmark microfilm publication W12",
	"roll":                 "393",
	"film_number":          "1022345",
	"item_number":          "3",
	"item":                 "Ada Robins",
	"entry_number":         "57",
	"line_number":          "14",
	"dwelling_number":      "112",
	"family_number":        "118",
	"schedule_number":      "9",
	"folio":                "22",
	"column":               "3",
	"record_id":            "R-77810",
}

// fixtureTitles is source.title per type.
var fixtureTitles = map[string]string{
	"census":               "1880 census of Calder County",
	"birth_certificate":    "Register of births",
	"marriage_certificate": "Register of marriages",
	"death_certificate":    "Register of deaths",
	"church_register":      "Baptisms and burials, 1850–1899",
	"record_book":          "Deeds, 1878–1882",
	"passenger_list":       "Passengers of the SS Marigold, June 1880",
	"case_file":            "Venn v. Thorne",
	"manuscript":           "Letter book of Ellis Trading Company",
	"government_file":      "Pension file of Harold Venn",
	"cemetery_marker":      "Grave of Ada Robins",
	"family_bible":         "Venn family Bible",
	"photograph":           "Portrait of Ada Robins",
	"private_papers":       "Notes on the Venn family",
	"dna_match_report":     "Match list for Clara Venn",
	"gedcom_file":          "Venn family tree",
	"interview":            "Clara Venn on her grandmother",
	"oral_testimony":       "Clara Venn on her grandmother",
	"correspondence":       "Letter about Ada Robins",
	"book":                 "The Families of Calder County",
	"journal_article":      "Ada Robins and the Calder flood",
	"broadcast":            "Ada Robins and the Calder flood",
	"newspaper_item":       "Death of Ada Robins",
	"website":              "Ada Robins memorial",
	"statute":              "An act to divide Calder County",
	"court_report":         "Venn v. Thorne",
}

// fixtureCreators is the creator per type; every other type uses
// defaultCreator.
var fixtureCreators = map[string]string{
	"census":               "Westmark",
	"birth_certificate":    "Calder County Clerk",
	"marriage_certificate": "Calder County Clerk",
	"death_certificate":    "Calder County Clerk",
	"church_register":      "St. Brendan's Church, Port Ellis",
	"record_book":          "Calder County Recorder",
	"passenger_list":       "Port Ellis Customs House",
	"case_file":            "Calder County Court",
	"manuscript":           "Ellis Trading Company",
	"cemetery_marker":      "Hillcrest Cemetery",
	"private_papers":       "Clara Venn",
	"correspondence":       "Clara Venn",
	"gedcom_file":          "Mabel Thorne",
	"website":              "Calder Historical Society",
}

const defaultCreator = "Calder County Clerk"

// fixture answers the targets of one render. The full fixture answers every
// field at every level (so a chain takes its first option), the title, the date
// and the page; the bare one answers only the title and the date.
type fixture struct {
	typeKey string
	keys    map[uuid.UUID]string // field id -> key
	bare    bool
	noPage  bool
}

func (f fixture) Resolve(t citationformat.Target) citationformat.Value {
	if t.Builtin != "" {
		switch {
		case t.Level == vocabkey.LevelSource && t.Builtin == "title":
			return citationformat.Value{Text: fixtureTitles[f.typeKey]}
		case t.Level == vocabkey.LevelSource && t.Builtin == "date":
			return citationformat.Value{Date: datevaluetest.YMD(1880, 6, 14)}
		case t.Level == vocabkey.LevelCitation && t.Builtin == "page" && !f.bare && !f.noPage:
			return citationformat.Value{Text: "14"}
		}
		return citationformat.Value{}
	}
	if f.bare {
		return citationformat.Value{}
	}
	key := f.keys[t.Field.ID]
	if key == "creator" {
		if c, ok := fixtureCreators[f.typeKey]; ok {
			return citationformat.Value{Text: c}
		}
		return citationformat.Value{Text: defaultCreator}
	}
	return citationformat.Value{Text: fieldValues[key]}
}

// plainWithItalics is the text of a render with each italic run in asterisks.
func plainWithItalics(r citationformat.Rendered) string {
	var b strings.Builder
	for _, run := range r.Runs {
		if run.Italic {
			b.WriteString("*" + run.Text + "*")
		} else {
			b.WriteString(run.Text)
		}
	}
	return b.String()
}

// TestSeedCitations renders every seeded format of an installed catalog, in
// all three forms, from the full fixture and the bare one (and court_report's
// full form with no page), and compares each with its line in
// testdata/seed-citations.txt. A person reads that file against the edition on
// hand; -update rewrites it.
func TestSeedCitations(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if err := install(c); err != nil {
		t.Fatal(err)
	}
	fields, err := metadatafields.List(c)
	if err != nil {
		t.Fatal(err)
	}
	keys := map[uuid.UUID]string{}
	for _, f := range fields {
		id, err := uuid.FromBytes(f.ID)
		if err != nil {
			t.Fatal(err)
		}
		keys[id] = f.Key
	}
	typeIDs := map[string][]byte{}
	for _, st := range seedTypes {
		ty, err := sourcetypes.Lookup(c, st.Key, vocabkey.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		typeIDs[st.Key] = ty.ID
	}

	var lines []string
	got := map[string]string{} // "type<TAB>fixture<TAB>form" -> text
	add := func(typeKey, name string, form citationformat.Form, text string) {
		id := typeKey + "\t" + name + "\t" + string(form)
		got[id] = text
		lines = append(lines, id+"\t"+text)
	}
	for _, st := range seedTypes {
		stored := storedFormats(t, c, typeIDs[st.Key])
		render := func(name string, fx fixture, form citationformat.Form) {
			tpl, diags := citationformat.Parse(stored.Text(form), form)
			if len(diags) > 0 {
				t.Fatalf("%s %s: %v", st.Key, form, diags)
			}
			add(st.Key, name, form, plainWithItalics(citationformat.Render(tpl, fx)))
		}
		for _, name := range []string{"full", "bare"} {
			fx := fixture{typeKey: st.Key, keys: keys, bare: name == "bare"}
			for _, form := range citationformat.Forms() {
				render(name, fx, form)
			}
		}
		if st.Key == "court_report" {
			render("nopage", fixture{typeKey: st.Key, keys: keys, noPage: true}, citationformat.FormFull)
		}
	}

	// The census lines, asserted literally as well as by the golden.
	for id, want := range map[string]string{
		"census\tfull\tfull":  "1880 Westmark census, Calder County, Westmark, population schedule, ED 236, piece 1047, fol. 22, p. 14, schedule 9, dwelling 112, family 118, line 14, Ada Robins; digital image, *Westmark Heritage* (https://example.org/calder/1880 : accessed 12 March 2025); citing county film 7, item 3.",
		"census\tfull\tshort": "1880 Westmark census, Calder County, Westmark, ED 236, fol. 22, p. 14, dwell. 112, fam. 118, Ada Robins.",
		"census\tfull\tlist":  "Calder County, Westmark. 1880 Westmark census, population schedule.",
	} {
		if got[id] != want {
			t.Errorf("%s =\n%q\nwant\n%q", id, got[id], want)
		}
	}

	// A value that carries its own "ed." or "held in" would collide with the
	// format around it: no full-fixture line has a doubled period or starts a
	// sentence in lowercase "held".
	for id, text := range got {
		if !strings.Contains(id, "\tfull\t") {
			continue
		}
		if strings.Contains(text, "..") || strings.Contains(text, ". held") {
			t.Errorf("%s = %q has a doubled period or a lowercase sentence start", id, text)
		}
	}

	body := strings.Join(lines, "\n") + "\n"
	if *update {
		if err := os.WriteFile(seedCitationsFile, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	raw, err := os.ReadFile(seedCitationsFile)
	if err != nil {
		t.Fatal(err)
	}
	want := strings.Split(strings.TrimRight(string(raw), "\n"), "\n")
	if len(want) != len(lines) {
		t.Errorf("%s has %d lines, the render has %d", seedCitationsFile, len(want), len(lines))
	}
	for i := 0; i < len(lines) && i < len(want); i++ {
		if lines[i] != want[i] {
			t.Errorf("line %d:\n got %q\nwant %q", i+1, lines[i], want[i])
		}
	}
}
