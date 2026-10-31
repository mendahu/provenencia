package citations

import (
	"bytes"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/artifacts/artifactstest"
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjects/subjectstest"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/locator"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
)

func TestCitations(t *testing.T) {
	userID := []byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}

	validLocator := `{"version":1,"selectors":[{"type":"page","artifact_page":12,"page_label":"10"}]}`

	type seed struct {
		artifact    artifacts.Artifact
		graph       []byte
		person      subjects.Subject
		place       subjects.Subject
		toponymProp properties.Property
		sexProp     properties.Property
		femaleTerm  propertyterms.Term
	}

	mustSeed := func(t *testing.T) (*database.Catalog, seed) {
		t.Helper()
		c, err := database.Create(t.TempDir(), "t.provenencia")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = c.Close() })
		ur, err := ref.Mint(ref.PrefixUser)
		if err != nil {
			t.Fatal(err)
		}
		if err := users.Upsert(c, userID, "Tester", ur); err != nil {
			t.Fatal(err)
		}
		if err := subjectvocab.Install(c); err != nil {
			t.Fatal(err)
		}
		typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
			Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book",
		})
		if err != nil {
			t.Fatal(err)
		}
		src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
			return sources.Create(tx, userID, sources.CreateInput{
				SourceTypeID: typeID,
				Title:        "Family Bible",
			})
		})
		if err != nil {
			t.Fatal(err)
		}
		art, err := artifactstest.Create(c, userID, artifacts.CreateInput{
			SourceID: src.ID,
			Label:    "Scan",
		})
		if err != nil {
			t.Fatal(err)
		}
		personType, err := subjecttypes.Lookup(c, "person", vocabkey.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		placeType, err := subjecttypes.Lookup(c, "place", vocabkey.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		graph := evidencegraphstest.SoleGraph(t, c, src.ID).ID
		person := subjectstest.OfType(t, c, userID, graph, personType.ID, "Alice")
		place := subjectstest.OfType(t, c, userID, graph, placeType.ID, "Boston")
		toponymProp, err := properties.Lookup(c, "toponym", vocabkey.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		sexProp, err := properties.Lookup(c, "sex_at_birth", vocabkey.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		femaleTerm, err := propertyterms.Lookup(c, sexProp.ID, "female", vocabkey.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		return c, seed{
			artifact: art, graph: graph, person: person, place: place,
			toponymProp: toponymProp, sexProp: sexProp, femaleTerm: femaleTerm,
		}
	}
	latestAction := func(t *testing.T, c *database.Catalog) string {
		t.Helper()
		db, err := c.DB()
		if err != nil {
			t.Fatal(err)
		}
		var actionType string
		if err := db.QueryRow(
			`SELECT action_type FROM audit_transactions ORDER BY revision DESC LIMIT 1`,
		).Scan(&actionType); err != nil {
			t.Fatal(err)
		}
		return actionType
	}
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "migration creates citation tables",
			run: func(t *testing.T) {
				c, err := database.Create(t.TempDir(), "t.provenencia")
				if err != nil {
					t.Fatal(err)
				}
				defer c.Close()
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				for _, table := range []string{"citations", "citation_notes", "observations", "observation_notes"} {
					var name string
					if err := db.QueryRow(
						`SELECT name FROM sqlite_schema WHERE type='table' AND name=?`, table,
					).Scan(&name); err != nil || name != table {
						t.Fatalf("table %q: %v", table, err)
					}
				}
				var ver int
				if err := db.QueryRow(`PRAGMA user_version`).Scan(&ver); err != nil {
					t.Fatal(err)
				}
				if ver < 26 {
					t.Fatalf("user_version=%d want >= 26", ver)
				}
			},
		},
		{
			name: "create with observations text and term round-trip",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				res, err := createCitation(c, userID, CreateInput{
					GraphID:       s.graph,
					ArtifactID:    s.artifact.ID,
					LocatorJSON:   validLocator,
					Transcription: "Alice, female, of Boston",
					Label:         "entry on page 10",
				}, []observations.Input{
					{
						SubjectID: s.place.ID, PropertyID: s.toponymProp.ID,
						ValueText: "Boston", HasText: true,
					},
					{
						SubjectID: s.person.ID, PropertyID: s.sexProp.ID,
						ValueTermID: s.femaleTerm.ID,
					},
				})
				if err != nil {
					t.Fatal(err)
				}
				if !strings.HasPrefix(res.Citation.Ref, "CIT-") {
					t.Fatalf("citation ref %q", res.Citation.Ref)
				}
				if len(res.Observations) != 2 {
					t.Fatalf("obs count %d", len(res.Observations))
				}
				for _, o := range res.Observations {
					if !strings.HasPrefix(o.Ref, "OBS-") {
						t.Fatalf("obs ref %q", o.Ref)
					}
				}
				if latestAction(t, c) != "create_citation_with_observations" {
					t.Fatalf("action %q", latestAction(t, c))
				}
				got, err := Get(c, res.Citation.ID)
				if err != nil {
					t.Fatal(err)
				}
				if got.Transcription != "Alice, female, of Boston" || got.LocatorJSON != validLocator {
					t.Fatalf("got %+v", got)
				}
				bySubject, err := observations.ListBySubject(c, s.person.ID)
				if err != nil || len(bySubject) != 1 {
					t.Fatalf("person obs %v len=%d", err, len(bySubject))
				}
				if string(bySubject[0].ValueTermID) != string(s.femaleTerm.ID) {
					t.Fatalf("term id mismatch")
				}
				placeObs, err := observations.ListBySubject(c, s.place.ID)
				if err != nil || len(placeObs) != 1 {
					t.Fatalf("place obs %v len=%d", err, len(placeObs))
				}
				if !placeObs[0].HasText || placeObs[0].ValueText != "Boston" {
					t.Fatalf("toponym %+v", placeObs[0])
				}
			},
		},
		{
			name: "create audit is full-state including nulls and notes",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				res, err := createCitation(c, userID, CreateInput{
					GraphID:     s.graph,
					ArtifactID:  s.artifact.ID,
					LocatorJSON: validLocator,
					Notes:       []string{"see folio 12"},
				}, []observations.Input{{
					SubjectID: s.person.ID, PropertyID: s.sexProp.ID,
					ValueTermID: s.femaleTerm.ID,
				}})
				if err != nil {
					t.Fatal(err)
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				rows, err := db.Query(`
					SELECT entity_type, action, changes_json FROM audit_changes
					WHERE audit_transaction_id = (
						SELECT id FROM audit_transactions ORDER BY revision DESC LIMIT 1
					)
					ORDER BY rowid`)
				if err != nil {
					t.Fatal(err)
				}
				defer rows.Close()
				type row struct {
					entity, action string
					fields         map[string]any
				}
				var got []row
				for rows.Next() {
					var entity, action, raw string
					if err := rows.Scan(&entity, &action, &raw); err != nil {
						t.Fatal(err)
					}
					var fields map[string]any
					if err := json.Unmarshal([]byte(raw), &fields); err != nil {
						t.Fatal(err)
					}
					got = append(got, row{entity, action, fields})
				}
				if err := rows.Err(); err != nil {
					t.Fatal(err)
				}
				if len(got) < 3 {
					t.Fatalf("changes %d want >= 3", len(got))
				}
				if got[0].entity != "citation" || got[0].action != "create" {
					t.Fatalf("first %+v", got[0])
				}
				for _, key := range []string{
					"id", "ref", "artifact_id", "locator_json",
					"transcription", "label", "transcription_uncertain", "transcription_note",
				} {
					if _, ok := got[0].fields[key]; !ok {
						t.Fatalf("citation missing %q in %+v", key, got[0].fields)
					}
				}
				for _, key := range []string{"transcription", "label", "transcription_note"} {
					diff, _ := got[0].fields[key].(map[string]any)
					if diff["new"] != nil {
						t.Fatalf("empty citation %s should be null: %+v", key, got[0].fields)
					}
				}
				if got[1].entity != "citation_note" {
					t.Fatalf("note %+v", got[1])
				}
				if body, _ := got[1].fields["body"].(map[string]any); body["new"] != "see folio 12" {
					t.Fatalf("note body %+v", got[1].fields)
				}
				if got[2].entity != "observation" {
					t.Fatalf("obs %+v", got[2])
				}
				for _, key := range []string{
					"id", "ref", "citation_id", "subject_id", "property_id", "polarity",
					"value_text", "value_integer", "value_date_id", "value_name_id",
					"value_subject_id", "value_term_id",
				} {
					if _, ok := got[2].fields[key]; !ok {
						t.Fatalf("observation missing %q in %+v", key, got[2].fields)
					}
				}
				for _, key := range []string{"value_text", "value_date_id"} {
					diff, _ := got[2].fields[key].(map[string]any)
					if diff["new"] != nil {
						t.Fatalf("unused observation %s should be null: %+v", key, got[2].fields)
					}
				}
				if string(res.Citation.ID) == "" {
					t.Fatal("empty citation id")
				}
			},
		},
		{
			name: "Update changes citation fields only and skips no-op",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				res, err := createCitation(c, userID, CreateInput{
					GraphID:       s.graph,
					ArtifactID:    s.artifact.ID,
					LocatorJSON:   validLocator,
					Transcription: "was",
					Notes:         []string{"keep"},
				}, []observations.Input{{
					SubjectID: s.person.ID, PropertyID: s.sexProp.ID,
					ValueTermID: s.femaleTerm.ID,
				}})
				if err != nil {
					t.Fatal(err)
				}
				got, err := writes.Call(c, writes.Op{Action: "update_citation", UserID: userID}, func(tx *database.Tx) (Citation, []rowchange.Change, error) {
					return Update(tx, userID, res.Citation.ID, CitationFieldsInput{
						LocatorJSON:   validLocator,
						Transcription: "now",
					})
				})
				if err != nil {
					t.Fatal(err)
				}
				if got.Transcription != "now" {
					t.Fatalf("transcription %q", got.Transcription)
				}
				if latestAction(t, c) != "update_citation" {
					t.Fatalf("action %q", latestAction(t, c))
				}
				notes, err := ListNotes(c, res.Citation.ID)
				if err != nil || len(notes) != 1 || notes[0] != "keep" {
					t.Fatalf("notes %v %+v", err, notes)
				}
				obs, err := observations.ListByCitation(c, res.Citation.ID)
				if err != nil || len(obs) != 1 {
					t.Fatalf("obs %v len=%d", err, len(obs))
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				var n int
				if err := db.QueryRow(`SELECT COUNT(*) FROM audit_transactions`).Scan(&n); err != nil {
					t.Fatal(err)
				}
				if _, err := writes.Call(c, writes.Op{Action: "update_citation", UserID: userID}, func(tx *database.Tx) (Citation, []rowchange.Change, error) {
					return Update(tx, userID, res.Citation.ID, CitationFieldsInput{
						LocatorJSON:   validLocator,
						Transcription: "now",
					})
				}); err != nil {
					t.Fatal(err)
				}
				var after int
				if err := db.QueryRow(`SELECT COUNT(*) FROM audit_transactions`).Scan(&after); err != nil {
					t.Fatal(err)
				}
				if after != n {
					t.Fatalf("no-op recorded revision %d -> %d", n, after)
				}
			},
		},
		{
			name: "mapConstraint preserves non-constraint errors",
			run: func(t *testing.T) {
				raw := errors.New("disk full")
				if got := mapConstraint(raw); got != raw {
					t.Fatalf("got %v want raw", got)
				}
			},
		},
		{
			name: "invalid locator rejected",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				_, err := createCitation(c, userID, CreateInput{
					GraphID:     s.graph,
					ArtifactID:  s.artifact.ID,
					LocatorJSON: `{"version":1,"selectors":[{"type":"page","artifact_page":0}]}`,
				}, []observations.Input{{
					SubjectID: s.person.ID, PropertyID: s.sexProp.ID,
					ValueTermID: s.femaleTerm.ID,
				}})
				if !errors.Is(err, locator.ErrInvalid) {
					t.Fatalf("got %v want locator.ErrInvalid", err)
				}
			},
		},
		{
			name: "list by artifact",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				if _, err := createCitation(c, userID, CreateInput{
					GraphID:    s.graph,
					ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
				}, []observations.Input{{
					SubjectID: s.person.ID, PropertyID: s.sexProp.ID,
					ValueTermID: s.femaleTerm.ID,
				}}); err != nil {
					t.Fatal(err)
				}
				list, err := ListByArtifact(c, evidencegraphstest.SoleGraph(t, c, s.artifact.SourceID).ID, s.artifact.ID)
				if err != nil || len(list) != 1 {
					t.Fatalf("%v len=%d", err, len(list))
				}
				if list[0].Ref[:4] != "CIT-" {
					t.Fatalf("ref %q", list[0].Ref)
				}
				if list[0].ObservationCount != 1 {
					t.Fatalf("obs count %d", list[0].ObservationCount)
				}
				if len(list[0].SubjectNames) != 1 || list[0].SubjectNames[0] != "Alice" {
					t.Fatalf("names %v", list[0].SubjectNames)
				}
			},
		},
		{
			name: "list by artifact is per graph",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				def := evidencegraphstest.SoleGraph(t, c, s.artifact.SourceID)
				smith := evidencegraphstest.Create(t, c, userID, s.artifact.SourceID, "Smith", s.artifact.ID)
				clara := subjectstest.On(t, c, userID, smith.ID, "person", "Clara")
				mk := func(graph, subject []byte) []byte {
					var obs []observations.Input
					if subject != nil {
						obs = []observations.Input{{SubjectID: subject, PropertyID: s.sexProp.ID, ValueTermID: s.femaleTerm.ID}}
					}
					res, err := createCitation(c, userID, CreateInput{GraphID: graph, ArtifactID: s.artifact.ID, LocatorJSON: validLocator}, obs)
					if err != nil {
						t.Fatal(err)
					}
					return res.Citation.ID
				}
				a, b := mk(def.ID, s.person.ID), mk(smith.ID, clara.ID)
				cc, d := mk(def.ID, nil), mk(smith.ID, nil)
				refs := func(graph []byte) []string {
					list, err := ListByArtifact(c, graph, s.artifact.ID)
					if err != nil {
						t.Fatal(err)
					}
					var out []string
					for _, l := range list {
						out = append(out, string(l.ID))
					}
					return out
				}
				same := func(got []string, want ...[]byte) bool {
					if len(got) != len(want) {
						return false
					}
					for _, w := range want {
						if !slices.Contains(got, string(w)) {
							return false
						}
					}
					return true
				}
				if got := refs(def.ID); !same(got, a, cc) {
					t.Fatalf("default lists %x want A, C", got)
				}
				if got := refs(smith.ID); !same(got, b, d) {
					t.Fatalf("smith lists %x want B, D", got)
				}
				counts, err := CountByEvidenceGraph(c, def.ID)
				if err != nil {
					t.Fatal(err)
				}
				if got := counts[uuid.Must(uuid.FromBytes(s.artifact.ID)).String()]; got != 2 {
					t.Fatalf("default count %d want 2", got)
				}
			},
		},
		{
			name: "list counts and names the graph's Observations",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				def := evidencegraphstest.SoleGraph(t, c, s.artifact.SourceID)
				smith := evidencegraphstest.Create(t, c, userID, s.artifact.SourceID, "Smith", s.artifact.ID)
				ben := subjectstest.On(t, c, userID, def.ID, "person", "Ben")
				clara := subjectstest.On(t, c, userID, smith.ID, "person", "Clara")
				in := func(subject []byte) observations.Input {
					return observations.Input{SubjectID: subject, PropertyID: s.sexProp.ID, ValueTermID: s.femaleTerm.ID}
				}
				if _, err := createCitation(c, userID, CreateInput{GraphID: def.ID, ArtifactID: s.artifact.ID, LocatorJSON: validLocator}, []observations.Input{in(s.person.ID), in(ben.ID)}); err != nil {
					t.Fatal(err)
				}
				if _, err := createCitation(c, userID, CreateInput{GraphID: smith.ID, ArtifactID: s.artifact.ID, LocatorJSON: validLocator}, []observations.Input{in(clara.ID)}); err != nil {
					t.Fatal(err)
				}
				d, err := ListByArtifact(c, def.ID, s.artifact.ID)
				if err != nil || len(d) != 1 {
					t.Fatalf("%v %d", err, len(d))
				}
				if d[0].ObservationCount != 2 || strings.Join(d[0].SubjectNames, ",") != "Alice,Ben" {
					t.Fatalf("default %d %v", d[0].ObservationCount, d[0].SubjectNames)
				}
				sm, err := ListByArtifact(c, smith.ID, s.artifact.ID)
				if err != nil || len(sm) != 1 {
					t.Fatalf("%v %d", err, len(sm))
				}
				if sm[0].ObservationCount != 1 || strings.Join(sm[0].SubjectNames, ",") != "Clara" {
					t.Fatalf("smith %d %v", sm[0].ObservationCount, sm[0].SubjectNames)
				}
			},
		},
		{
			name: "count by evidence graph skips unlinked artifacts",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				art2, err := artifactstest.Create(c, userID, artifacts.CreateInput{SourceID: s.artifact.SourceID, Label: "Scan 2"})
				if err != nil {
					t.Fatal(err)
				}
				def := evidencegraphstest.SoleGraph(t, c, s.artifact.SourceID)
				smith := evidencegraphstest.Create(t, c, userID, s.artifact.SourceID, "Smith", s.artifact.ID)
				for _, in := range []CreateInput{
					{GraphID: def.ID, ArtifactID: s.artifact.ID},
					{GraphID: def.ID, ArtifactID: art2.ID},
					{GraphID: smith.ID, ArtifactID: s.artifact.ID},
				} {
					in.LocatorJSON = validLocator
					if _, err := createCitation(c, userID, in, nil); err != nil {
						t.Fatal(err)
					}
				}
				id1 := uuid.Must(uuid.FromBytes(s.artifact.ID)).String()
				id2 := uuid.Must(uuid.FromBytes(art2.ID)).String()
				dc, err := CountByEvidenceGraph(c, def.ID)
				if err != nil || dc[id1] != 1 || dc[id2] != 1 {
					t.Fatalf("default %v %v", dc, err)
				}
				sc, err := CountByEvidenceGraph(c, smith.ID)
				if err != nil || sc[id1] != 1 {
					t.Fatalf("smith %v %v", sc, err)
				}
				if _, ok := sc[id2]; ok {
					t.Fatalf("smith counts an unlinked artifact: %v", sc)
				}
			},
		},
		{
			name: "Citations belong to their graph",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				def := evidencegraphstest.SoleGraph(t, c, s.artifact.SourceID)
				jones := evidencegraphstest.Create(t, c, userID, s.artifact.SourceID, "Jones", s.artifact.ID)
				res, err := createCitation(c, userID, CreateInput{GraphID: jones.ID, ArtifactID: s.artifact.ID, LocatorJSON: validLocator}, nil)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(res.Citation.GraphID, jones.ID) {
					t.Fatalf("graph %x want jones", res.Citation.GraphID)
				}
				inDefault, err := ListByArtifact(c, def.ID, s.artifact.ID)
				if err != nil || len(inDefault) != 0 {
					t.Fatalf("default lists %d Citations, err %v", len(inDefault), err)
				}
				inJones, err := ListByArtifact(c, jones.ID, s.artifact.ID)
				if err != nil || len(inJones) != 1 {
					t.Fatalf("jones lists %d Citations, err %v", len(inJones), err)
				}
				art := uuid.Must(uuid.FromBytes(s.artifact.ID)).String()
				dc, err := CountByEvidenceGraph(c, def.ID)
				if err != nil {
					t.Fatal(err)
				}
				if _, ok := dc[art]; ok {
					t.Fatalf("default counts the Jones Citation: %v", dc)
				}
				jc, err := CountByEvidenceGraph(c, jones.ID)
				if err != nil || jc[art] != 1 {
					t.Fatalf("jones counts %v, err %v", jc, err)
				}
			},
		},
		{
			name: "create needs a linked graph",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				art2, err := artifactstest.Create(c, userID, artifacts.CreateInput{SourceID: s.artifact.SourceID, Label: "Scan 2"})
				if err != nil {
					t.Fatal(err)
				}
				jones := evidencegraphstest.Create(t, c, userID, s.artifact.SourceID, "Jones", s.artifact.ID)
				for name, in := range map[string]CreateInput{
					"no graph":       {ArtifactID: s.artifact.ID, LocatorJSON: validLocator},
					"unlinked graph": {GraphID: jones.ID, ArtifactID: art2.ID, LocatorJSON: validLocator},
					"unknown graph":  {GraphID: bytes.Repeat([]byte{9}, 16), ArtifactID: s.artifact.ID, LocatorJSON: validLocator},
				} {
					if _, err := createCitation(c, userID, in, nil); !errors.Is(err, ErrInvalid) {
						t.Fatalf("%s: err %v, want citations.invalid", name, err)
					}
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				var n int
				if err := db.QueryRow(`SELECT COUNT(*) FROM citations`).Scan(&n); err != nil || n != 0 {
					t.Fatalf("%d Citations written, err %v", n, err)
				}
			},
		},
		{
			name: "create with zero observations",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				res, err := createCitation(c, userID, CreateInput{
					GraphID:       s.graph,
					ArtifactID:    s.artifact.ID,
					LocatorJSON:   validLocator,
					Transcription: "transcribe first",
				}, nil)
				if err != nil {
					t.Fatal(err)
				}
				if !strings.HasPrefix(res.Citation.Ref, "CIT-") {
					t.Fatalf("citation ref %q", res.Citation.Ref)
				}
				if len(res.Observations) != 0 {
					t.Fatalf("obs count %d", len(res.Observations))
				}
				list, err := ListByArtifact(c, evidencegraphstest.SoleGraph(t, c, s.artifact.SourceID).ID, s.artifact.ID)
				if err != nil || len(list) != 1 {
					t.Fatalf("%v len=%d", err, len(list))
				}
				if list[0].ObservationCount != 0 {
					t.Fatalf("listed count %d", list[0].ObservationCount)
				}
			},
		},
		{
			name: "count by evidence graph",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				art2, err := artifactstest.Create(c, userID, artifacts.CreateInput{
					SourceID: s.artifact.SourceID,
					Label:    "Scan 2",
				})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := createCitation(c, userID, CreateInput{
					GraphID:    s.graph,
					ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
				}, []observations.Input{{
					SubjectID: s.person.ID, PropertyID: s.sexProp.ID,
					ValueTermID: s.femaleTerm.ID,
				}}); err != nil {
					t.Fatal(err)
				}
				if _, err := createCitation(c, userID, CreateInput{
					GraphID:    s.graph,
					ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
				}, []observations.Input{{
					SubjectID: s.person.ID, PropertyID: s.sexProp.ID,
					ValueTermID: s.femaleTerm.ID,
				}}); err != nil {
					t.Fatal(err)
				}
				if _, err := createCitation(c, userID, CreateInput{
					GraphID:    s.graph,
					ArtifactID: art2.ID, LocatorJSON: validLocator,
				}, []observations.Input{{
					SubjectID: s.person.ID, PropertyID: s.sexProp.ID,
					ValueTermID: s.femaleTerm.ID,
				}}); err != nil {
					t.Fatal(err)
				}
				counts, err := CountByEvidenceGraph(c, evidencegraphstest.SoleGraph(t, c, s.artifact.SourceID).ID)
				if err != nil {
					t.Fatal(err)
				}
				id1 := uuid.Must(uuid.FromBytes(s.artifact.ID)).String()
				id2 := uuid.Must(uuid.FromBytes(art2.ID)).String()
				if counts[id1] != 2 || counts[id2] != 1 {
					t.Fatalf("counts=%v want %s=2 %s=1", counts, id1, id2)
				}
			},
		},
		{
			name: "artifact delete refused while citation exists",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				if _, err := createCitation(c, userID, CreateInput{
					GraphID:    s.graph,
					ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
				}, nil); err != nil {
					t.Fatal(err)
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				if _, err := db.Exec(`DELETE FROM artifacts WHERE id = ?`, s.artifact.ID); err == nil {
					t.Fatal("expected artifact delete to fail while a citation remains")
				}
				var n int
				if err := db.QueryRow(`SELECT COUNT(*) FROM artifacts WHERE id = ?`, s.artifact.ID).Scan(&n); err != nil || n != 1 {
					t.Fatalf("artifact %d %v", n, err)
				}
			},
		},
		{
			name: "count by evidence graph rejects short id",
			run: func(t *testing.T) {
				c, _ := mustSeed(t)
				if _, err := CountByEvidenceGraph(c, []byte{1}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v want ErrInvalid", err)
				}
			},
		},
		{
			name: "empty citation erases and notes cascade",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				cit, err := createCitation(c, userID, CreateInput{
					GraphID:    s.graph,
					ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
					Notes: []string{"keep me", "and me"},
				}, nil)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := writes.Call(c, writes.Op{Action: "delete_citation", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
					changes, err := Delete(tx, userID, cit.Citation.ID)
					return struct{}{}, changes, err
				}); err != nil {
					t.Fatal(err)
				}
				if latestAction(t, c) != "delete_citation" {
					t.Fatalf("action %q", latestAction(t, c))
				}
				if _, err := Get(c, cit.Citation.ID); err == nil {
					t.Fatal("citation still present")
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				var n int
				if err := db.QueryRow(`SELECT COUNT(*) FROM citation_notes WHERE citation_id = ?`, cit.Citation.ID).Scan(&n); err != nil || n != 0 {
					t.Fatalf("notes leftover %d %v", n, err)
				}
			},
		},
		{
			name: "inbound observations refuse and row remains",
			run: func(t *testing.T) {
				c, s := mustSeed(t)
				cit, err := createCitation(c, userID, CreateInput{
					GraphID:    s.graph,
					ArtifactID: s.artifact.ID, LocatorJSON: validLocator,
				}, []observations.Input{{
					SubjectID: s.person.ID, PropertyID: s.sexProp.ID,
					ValueTermID: s.femaleTerm.ID,
				}})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := writes.Call(c, writes.Op{Action: "delete_citation", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
					changes, err := Delete(tx, userID, cit.Citation.ID)
					return struct{}{}, changes, err
				}); !errors.Is(err, ErrInUse) {
					t.Fatalf("got %v want ErrInUse", err)
				}
				if _, err := Get(c, cit.Citation.ID); err != nil {
					t.Fatalf("citation gone: %v", err)
				}
			},
		},
		{
			name: "missing citation is invalid",
			run: func(t *testing.T) {
				c, _ := mustSeed(t)
				missing := make([]byte, 16)
				missing[15] = 9
				if _, err := writes.Call(c, writes.Op{Action: "delete_citation", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
					changes, err := Delete(tx, userID, missing)
					return struct{}{}, changes, err
				}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v want ErrInvalid", err)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

// createCitation creates a Citation with its Observations in one audited write.
func createCitation(c *database.Catalog, userID []byte, in CreateInput, obs []observations.Input) (CreateResult, error) {
	return writes.Call(c, writes.Op{Action: "create_citation_with_observations", UserID: userID}, func(tx *database.Tx) (CreateResult, []rowchange.Change, error) {
		return CreateWithObservations(tx, userID, in, obs)
	})
}
