package audit_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ingest"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/writes"
)

// TestSourceScopes drives real write paths and checks each one moves the
// Sources list "Updated" revision of exactly the Sources it touched.
func TestSourceScopes(t *testing.T) {
	userID := []byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	validLocator := `{"version":1,"selectors":[{"type":"page","artifact_page":3}]}`

	type seed struct {
		c           *database.Catalog
		a, b        sources.Source
		artifact    artifacts.Artifact // on a
		person      subjects.Subject   // on a
		event       subjects.Subject   // on a
		place       subjects.Subject   // on a
		place2      subjects.Subject   // on a
		toponymProp properties.Property
		dateProp    properties.Property
		personTy    []byte
	}

	mustSeed := func(t *testing.T) seed {
		t.Helper()
		c, err := database.Create(t.TempDir(), "t.provenencia")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = c.Close() })
		// Registered second so it runs first, while the catalog is still open.
		// Writes that omit a foreign key are checked in bumpsA, before a later
		// delete removes the live row those lookups need.
		t.Cleanup(func() { checkNewEffects(t, c) })
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
			Key: "book", Origin: sourcetypes.OriginProvenencia, Label: "Book",
		})
		if err != nil {
			t.Fatal(err)
		}
		a, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
			return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "A"})
		})
		if err != nil {
			t.Fatal(err)
		}
		art, err := runArtifactCreate(c, userID, artifacts.CreateInput{SourceID: a.ID, Label: "Scan"})
		if err != nil {
			t.Fatal(err)
		}
		personType, err := subjecttypes.Lookup(c, "person", subjecttypes.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		eventType, err := subjecttypes.Lookup(c, "event", subjecttypes.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		mkSubject := func(typeID []byte, label string, x int64) subjects.Subject {
			s, err := subjects.Create(c, userID, subjects.CreateInput{
				SourceID: a.ID, SubjectTypeID: typeID, Label: label,
			}, &subjects.Placement{GridX: x, GridY: 0})
			if err != nil {
				t.Fatal(err)
			}
			return s
		}
		person := mkSubject(personType.ID, "Bob", 0)
		event := mkSubject(eventType.ID, "Birth", 4)
		placeType, err := subjecttypes.Lookup(c, "place", subjecttypes.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		place := mkSubject(placeType.ID, "Boston", 8)
		place2 := mkSubject(placeType.ID, "Salem", 12)
		toponymProp, err := properties.Lookup(c, "toponym", properties.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		dateProp, err := properties.Lookup(c, "date", properties.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		// b is created last so it starts ahead of a on "Updated".
		b, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
			return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "B"})
		})
		if err != nil {
			t.Fatal(err)
		}
		s := seed{
			c: c, a: a, b: b, artifact: art, person: person, event: event, place: place, place2: place2,
			toponymProp: toponymProp, dateProp: dateProp, personTy: personType.ID,
		}
		checkNewEffects(t, c)
		return s
	}

	revisions := func(t *testing.T, s seed) (a, b int64) {
		t.Helper()
		all, err := sources.List(s.c)
		if err != nil {
			t.Fatal(err)
		}
		for _, src := range all {
			switch {
			case bytes.Equal(src.ID, s.a.ID):
				a = src.UpdatedRevision
			case bytes.Equal(src.ID, s.b.ID):
				b = src.UpdatedRevision
			}
		}
		return a, b
	}
	latest := func(t *testing.T, s seed) int64 {
		t.Helper()
		db, err := s.c.DB()
		if err != nil {
			t.Fatal(err)
		}
		var rev int64
		if err := db.QueryRow(`SELECT MAX(revision) FROM audit_transactions`).Scan(&rev); err != nil {
			t.Fatal(err)
		}
		return rev
	}
	// bumpsA runs act and wants a (only) to move to the latest revision.
	bumpsA := func(t *testing.T, s seed, act func()) {
		t.Helper()
		_, b0 := revisions(t, s)
		act()
		checkNewEffects(t, s.c)
		a1, b1 := revisions(t, s)
		if want := latest(t, s); a1 != want {
			t.Fatalf("source A revision %d, want latest %d", a1, want)
		}
		if b1 != b0 {
			t.Fatalf("source B moved %d → %d", b0, b1)
		}
	}
	mustCitation := func(t *testing.T, s seed, in citations.CreateInput, obs ...observations.Input) citations.CreateResult {
		t.Helper()
		in.ArtifactID = s.artifact.ID
		in.LocatorJSON = validLocator
		res, err := citations.CreateWithObservations(s.c, userID, in, obs)
		if err != nil {
			t.Fatal(err)
		}
		return res
	}
	textObs := func(s seed, subject subjects.Subject, text string) observations.Input {
		return observations.Input{SubjectID: subject.ID, PropertyID: s.toponymProp.ID, ValueText: text, HasText: true}
	}

	tests := []struct {
		name string
		run  func(t *testing.T, s seed)
	}{
		{
			name: "seed: newer source leads",
			run: func(t *testing.T, s seed) {
				a, b := revisions(t, s)
				if a >= b {
					t.Fatalf("want b ahead after seed: a=%d b=%d", a, b)
				}
			},
		},
		{
			name: "source note add, edit, delete",
			run: func(t *testing.T, s seed) {
				var note sources.Note
				bumpsA(t, s, func() {
					n, _, err := writes.Run(s.c, writes.Op{Action: "create_source_note", UserID: userID}, func(tx *database.Tx) (sources.Note, []rowchange.Change, error) {
						return sources.AddNote(tx, userID, s.a.ID, "first")
					})
					if err != nil {
						t.Fatal(err)
					}
					note = n
				})
				bumpsA(t, s, func() {
					if _, _, err := writes.Run(s.c, writes.Op{Action: "update_source_note", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
						changes, err := sources.UpdateNote(tx, userID, note.ID, "second")
						return struct{}{}, changes, err
					}); err != nil {
						t.Fatal(err)
					}
				})
				bumpsA(t, s, func() {
					if _, _, err := writes.Run(s.c, writes.Op{Action: "delete_source_note", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
						changes, err := sources.DeleteNote(tx, userID, note.ID)
						return struct{}{}, changes, err
					}); err != nil {
						t.Fatal(err)
					}
				})
			},
		},
		{
			name: "artifact create and update",
			run: func(t *testing.T, s seed) {
				var art artifacts.Artifact
				bumpsA(t, s, func() {
					a, err := runArtifactCreate(s.c, userID, artifacts.CreateInput{SourceID: s.a.ID, Label: "Back"})
					if err != nil {
						t.Fatal(err)
					}
					art = a
				})
				bumpsA(t, s, func() {
					art.Label = "Reverse"
					if err := runArtifactUpdate(s.c, userID, art); err != nil {
						t.Fatal(err)
					}
				})
			},
		},
		{
			name: "subject create, edit, delete",
			run: func(t *testing.T, s seed) {
				var subj subjects.Subject
				bumpsA(t, s, func() {
					got, err := subjects.Create(s.c, userID, subjects.CreateInput{
						SourceID: s.a.ID, SubjectTypeID: s.personTy, Label: "Cy",
					}, &subjects.Placement{GridX: 16, GridY: 0})
					if err != nil {
						t.Fatal(err)
					}
					subj = got
				})
				bumpsA(t, s, func() {
					if err := subjects.Update(s.c, userID, subj.ID, "Cyril", ""); err != nil {
						t.Fatal(err)
					}
				})
				bumpsA(t, s, func() {
					if err := subjects.Delete(s.c, userID, subj.ID); err != nil {
						t.Fatal(err)
					}
				})
			},
		},
		{
			name: "citation create, edit; observation add, edit, delete; citation delete with notes",
			run: func(t *testing.T, s seed) {
				var res citations.CreateResult
				bumpsA(t, s, func() {
					in := textObs(s, s.place, "Boston")
					in.Notes = []string{"faded ink"}
					res = mustCitation(t, s, citations.CreateInput{Notes: []string{"p. 3"}}, in)
				})
				bumpsA(t, s, func() {
					if _, err := citations.Update(s.c, userID, res.Citation.ID, citations.CitationFieldsInput{
						LocatorJSON: validLocator, Transcription: "Bob Smith",
					}); err != nil {
						t.Fatal(err)
					}
				})
				var added []observations.Observation
				bumpsA(t, s, func() {
					got, err := observations.AddToCitation(s.c, userID, res.Citation.ID,
						[]observations.Input{textObs(s, s.place2, "Salem")})
					if err != nil {
						t.Fatal(err)
					}
					added = got
				})
				bumpsA(t, s, func() {
					if _, err := observations.Update(s.c, userID, observations.Input{
						ID: added[0].ID, SubjectID: s.place2.ID, PropertyID: s.toponymProp.ID,
						ValueText: "Salem Town", HasText: true,
					}); err != nil {
						t.Fatal(err)
					}
				})
				// Deleting an observation releases its notes first; the
				// note's change resolves through the deleted observation.
				bumpsA(t, s, func() {
					if err := observations.Delete(s.c, userID, res.Observations[0].ID); err != nil {
						t.Fatal(err)
					}
				})
				bumpsA(t, s, func() {
					if err := observations.Delete(s.c, userID, added[0].ID); err != nil {
						t.Fatal(err)
					}
				})
				bumpsA(t, s, func() {
					if err := citations.Delete(s.c, userID, res.Citation.ID); err != nil {
						t.Fatal(err)
					}
				})
			},
		},
		{
			name: "polarity edit keeps the source",
			run: func(t *testing.T, s seed) {
				res := mustCitation(t, s, citations.CreateInput{}, textObs(s, s.place, "Boston"))
				bumpsA(t, s, func() {
					if _, err := observations.Update(s.c, userID, observations.Input{
						ID: res.Observations[0].ID, SubjectID: s.place.ID, PropertyID: s.toponymProp.ID,
						ValueText: "Boston", HasText: true, Polarity: observations.PolarityNegative,
					}); err != nil {
						t.Fatal(err)
					}
				})
			},
		},
		{
			name: "date value edited in place",
			run: func(t *testing.T, s seed) {
				year := 1842
				res := mustCitation(t, s, citations.CreateInput{}, observations.Input{
					SubjectID: s.event.ID, PropertyID: s.dateProp.ID,
					Date: &datevalues.Value{Kind: datevalues.KindPoint, StartYear: &year},
				})
				year2 := 1843
				bumpsA(t, s, func() {
					if _, err := observations.Update(s.c, userID, observations.Input{
						ID: res.Observations[0].ID, SubjectID: s.event.ID, PropertyID: s.dateProp.ID,
						Date: &datevalues.Value{Kind: datevalues.KindPoint, StartYear: &year2},
					}); err != nil {
						t.Fatal(err)
					}
				})
			},
		},
		{
			name: "file rename moves every source using it",
			run: func(t *testing.T, s seed) {
				path := filepath.Join(t.TempDir(), "scan.txt")
				writeFile(t, path, "page one")
				fres, err := ingest.File(s.c, path, userID)
				if err != nil {
					t.Fatal(err)
				}
				// The file create has no artifact yet. Checking it after the
				// artifacts exist would see sources the revision did not.
				checkNewEffects(t, s.c)
				for _, src := range []sources.Source{s.a, s.b} {
					if _, err := runArtifactCreate(s.c, userID, artifacts.CreateInput{
						SourceID: src.ID, FileID: fres.File.ID, Label: "Scan",
					}); err != nil {
						t.Fatal(err)
					}
				}
				if err := runSetFilename(s.c, fres.File.ID, "renamed.txt", userID); err != nil {
					t.Fatal(err)
				}
				a, b := revisions(t, s)
				want := latest(t, s)
				if a != want || b != want {
					t.Fatalf("want both at %d got a=%d b=%d", want, a, b)
				}
			},
		},
		{
			name: "promote does not move the source",
			run: func(t *testing.T, s seed) {
				a0, b0 := revisions(t, s)
				if _, err := promote.Save(s.c, userID, promote.Input{SubjectID: s.person.ID}); err != nil {
					t.Fatal(err)
				}
				if latest(t, s) <= b0 {
					t.Fatal("promote recorded no revision")
				}
				a1, b1 := revisions(t, s)
				if a1 != a0 || b1 != b0 {
					t.Fatalf("promote moved sources: a %d→%d b %d→%d", a0, a1, b0, b1)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t, mustSeed(t))
		})
	}
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func runArtifactCreate(c *database.Catalog, userID []byte, in artifacts.CreateInput) (artifacts.Artifact, error) {
	a, _, err := writes.Run(c, writes.Op{Action: "create_artifact", UserID: userID},
		func(tx *database.Tx) (artifacts.Artifact, []rowchange.Change, error) {
			return artifacts.Create(tx, userID, in)
		})
	return a, err
}

func runArtifactUpdate(c *database.Catalog, userID []byte, a artifacts.Artifact) error {
	_, _, err := writes.Run(c, writes.Op{Action: "update_artifact", UserID: userID},
		func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := artifacts.Update(tx, userID, a)
			return struct{}{}, changes, err
		})
	return err
}

func runSetFilename(c *database.Catalog, fileID []byte, name string, userID []byte) error {
	_, _, err := writes.Run(c, writes.Op{Action: "update_file", UserID: userID},
		func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := ingest.SetFilename(tx, fileID, name, userID)
			return struct{}{}, changes, err
		})
	return err
}
