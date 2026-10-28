package audit_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/artifacts/artifactstest"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/evidencegraphs"
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/promote/promotetest"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/ingest"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
)

// scopeUserID is the researcher every scope test writes as.
var scopeUserID = []byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}

const scopeLocator = `{"version":1,"selectors":[{"type":"page","artifact_page":3}]}`

// scopeSeed is the shared world of the scope tests: two Sources, one
// artifact, and a Subject of each kind on Source A. It promotes nothing.
type scopeSeed struct {
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

func newScopeSeed(t *testing.T) scopeSeed {
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
	if err := users.Upsert(c, scopeUserID, "Tester", ur); err != nil {
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
	a, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: scopeUserID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, scopeUserID, sources.CreateInput{SourceTypeID: typeID, Title: "A"})
	})
	if err != nil {
		t.Fatal(err)
	}
	art, err := artifactstest.Create(c, scopeUserID, artifacts.CreateInput{SourceID: a.ID, Label: "Scan"})
	if err != nil {
		t.Fatal(err)
	}
	personType, err := subjecttypes.Lookup(c, "person", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	eventType, err := subjecttypes.Lookup(c, "event", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	mkSubject := func(typeID []byte, label string, x int64) subjects.Subject {
		graph := evidencegraphstest.SoleGraph(t, c, a.ID).ID
		s, err := writes.Call(c, writes.Op{Action: "create_subject", UserID: scopeUserID}, func(tx *database.Tx) (subjects.Subject, []rowchange.Change, error) {
			return subjects.Create(tx, scopeUserID, subjects.CreateInput{
				GraphID: graph, SubjectTypeID: typeID, Label: label,
			}, &subjects.Placement{GridX: x, GridY: 0})
		})
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
	person := mkSubject(personType.ID, "Bob", 0)
	event := mkSubject(eventType.ID, "Birth", 4)
	placeType, err := subjecttypes.Lookup(c, "place", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	place := mkSubject(placeType.ID, "Boston", 8)
	place2 := mkSubject(placeType.ID, "Salem", 12)
	toponymProp, err := properties.Lookup(c, "toponym", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	dateProp, err := properties.Lookup(c, "date", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	// b is created last so it starts ahead of a on "Updated".
	b, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: scopeUserID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, scopeUserID, sources.CreateInput{SourceTypeID: typeID, Title: "B"})
	})
	if err != nil {
		t.Fatal(err)
	}
	s := scopeSeed{
		c: c, a: a, b: b, artifact: art, person: person, event: event, place: place, place2: place2,
		toponymProp: toponymProp, dateProp: dateProp, personTy: personType.ID,
	}
	checkNewEffects(t, c)
	return s
}

func (s scopeSeed) sourceRevisions(t *testing.T) (a, b int64) {
	t.Helper()
	all, err := sources.List(s.c)
	if err != nil {
		t.Fatal(err)
	}
	for _, src := range all {
		switch {
		case bytes.Equal(src.ID, s.a.ID):
			a = src.Updated.Revision
		case bytes.Equal(src.ID, s.b.ID):
			b = src.Updated.Revision
		}
	}
	return a, b
}

func (s scopeSeed) latest(t *testing.T) int64 {
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
func (s scopeSeed) bumpsA(t *testing.T, act func()) {
	t.Helper()
	_, b0 := s.sourceRevisions(t)
	act()
	checkNewEffects(t, s.c)
	a1, b1 := s.sourceRevisions(t)
	if want := s.latest(t); a1 != want {
		t.Fatalf("source A revision %d, want latest %d", a1, want)
	}
	if b1 != b0 {
		t.Fatalf("source B moved %d → %d", b0, b1)
	}
}

func (s scopeSeed) citation(t *testing.T, in citations.CreateInput, obs ...observations.Input) citations.CreateResult {
	t.Helper()
	in.ArtifactID = s.artifact.ID
	in.LocatorJSON = scopeLocator
	res, err := writes.Call(s.c, writes.Op{Action: "create_citation_with_observations", UserID: scopeUserID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
		return citations.CreateWithObservations(tx, scopeUserID, in, obs)
	})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func (s scopeSeed) textObs(subject subjects.Subject, text string) observations.Input {
	return observations.Input{SubjectID: subject.ID, PropertyID: s.toponymProp.ID, ValueText: text, HasText: true}
}

// promote promotes a Subject to a new handle and returns the handle id.
func (s scopeSeed) promote(t *testing.T, subject subjects.Subject) []byte {
	t.Helper()
	res := promotetest.Save(t, s.c, scopeUserID, promote.Input{SubjectID: subject.ID})
	return res.Entity.ID
}

func (s scopeSeed) updateObservation(t *testing.T, in observations.Input) {
	t.Helper()
	if _, err := writes.Call(s.c, writes.Op{Action: "update_observation", UserID: scopeUserID}, func(tx *database.Tx) (observations.Listed, []rowchange.Change, error) {
		return observations.Update(tx, scopeUserID, in)
	}); err != nil {
		t.Fatal(err)
	}
}

func (s scopeSeed) deleteObservation(t *testing.T, id []byte) {
	t.Helper()
	if _, err := writes.Call(s.c, writes.Op{Action: "delete_observation", UserID: scopeUserID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
		changes, err := observations.Delete(tx, scopeUserID, id)
		return struct{}{}, changes, err
	}); err != nil {
		t.Fatal(err)
	}
}

// handleRevision is the handle's latest canonical_entity revision, 0 for none.
func (s scopeSeed) handleRevision(t *testing.T, id []byte) int64 {
	t.Helper()
	db, err := s.c.DB()
	if err != nil {
		t.Fatal(err)
	}
	var rev int64
	if err := db.QueryRow(`SELECT COALESCE(MAX(t.revision), 0)
		FROM audit_transaction_scopes sc
		JOIN audit_transactions t ON t.id = sc.audit_transaction_id
		WHERE sc.scope_type = ? AND sc.scope_id = ?`, effects.ScopeCanonicalEntity, id).Scan(&rev); err != nil {
		t.Fatal(err)
	}
	return rev
}

// TestSourceScopes drives real write paths and checks each one moves the
// Sources list "Updated" revision of exactly the Sources it touched.
func TestSourceScopes(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, s scopeSeed)
	}{
		{
			name: "seed: newer source leads",
			run: func(t *testing.T, s scopeSeed) {
				a, b := s.sourceRevisions(t)
				if a >= b {
					t.Fatalf("want b ahead after seed: a=%d b=%d", a, b)
				}
			},
		},
		{
			name: "source note add, edit, delete",
			run: func(t *testing.T, s scopeSeed) {
				var note sources.Note
				s.bumpsA(t, func() {
					n, _, err := writes.Run(s.c, writes.Op{Action: "create_source_note", UserID: scopeUserID}, func(tx *database.Tx) (sources.Note, []rowchange.Change, error) {
						return sources.AddNote(tx, scopeUserID, s.a.ID, "first")
					})
					if err != nil {
						t.Fatal(err)
					}
					note = n
				})
				s.bumpsA(t, func() {
					if _, _, err := writes.Run(s.c, writes.Op{Action: "update_source_note", UserID: scopeUserID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
						changes, err := sources.UpdateNote(tx, scopeUserID, note.ID, "second")
						return struct{}{}, changes, err
					}); err != nil {
						t.Fatal(err)
					}
				})
				s.bumpsA(t, func() {
					if _, _, err := writes.Run(s.c, writes.Op{Action: "delete_source_note", UserID: scopeUserID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
						changes, err := sources.DeleteNote(tx, scopeUserID, note.ID)
						return struct{}{}, changes, err
					}); err != nil {
						t.Fatal(err)
					}
				})
			},
		},
		{
			name: "artifact create and update",
			run: func(t *testing.T, s scopeSeed) {
				var art artifacts.Artifact
				s.bumpsA(t, func() {
					a, err := artifactstest.Create(s.c, scopeUserID, artifacts.CreateInput{SourceID: s.a.ID, Label: "Back"})
					if err != nil {
						t.Fatal(err)
					}
					art = a
				})
				s.bumpsA(t, func() {
					art.Label = "Reverse"
					if err := runArtifactUpdate(s.c, scopeUserID, art); err != nil {
						t.Fatal(err)
					}
				})
			},
		},
		{
			name: "graph link and unlink",
			run: func(t *testing.T, s scopeSeed) {
				second, err := artifactstest.Create(s.c, scopeUserID, artifacts.CreateInput{SourceID: s.a.ID, Label: "Second"})
				if err != nil {
					t.Fatal(err)
				}
				linkWrite := func(action string, fn func(tx *database.Tx, graphID []byte) ([]rowchange.Change, error)) {
					t.Helper()
					graphID := evidencegraphstest.SoleGraph(t, s.c, s.a.ID).ID
					if _, err := writes.Call(s.c, writes.Op{Action: action, UserID: scopeUserID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
						changes, err := fn(tx, graphID)
						return struct{}{}, changes, err
					}); err != nil {
						t.Fatal(err)
					}
				}
				s.bumpsA(t, func() {
					linkWrite("unlink_artifact", func(tx *database.Tx, graphID []byte) ([]rowchange.Change, error) {
						return evidencegraphs.Unlink(tx, scopeUserID, graphID, second.ID)
					})
				})
				s.bumpsA(t, func() {
					linkWrite("link_artifact", func(tx *database.Tx, graphID []byte) ([]rowchange.Change, error) {
						return evidencegraphs.Link(tx, scopeUserID, graphID, second.ID)
					})
				})
			},
		},
		{
			name: "graph create, rename and delete",
			run: func(t *testing.T, s scopeSeed) {
				var jones evidencegraphs.Graph
				s.bumpsA(t, func() {
					got, err := writes.Call(s.c, writes.Op{Action: "create_evidence_graph", UserID: scopeUserID}, func(tx *database.Tx) (evidencegraphs.Graph, []rowchange.Change, error) {
						return evidencegraphs.Create(tx, scopeUserID, evidencegraphs.CreateInput{
							SourceID: s.a.ID, Name: "Jones", ArtifactIDs: [][]byte{s.artifact.ID},
						})
					})
					if err != nil {
						t.Fatal(err)
					}
					jones = got
				})
				s.bumpsA(t, func() {
					if _, err := writes.Call(s.c, writes.Op{Action: "update_evidence_graph", UserID: scopeUserID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
						changes, err := evidencegraphs.Update(tx, scopeUserID, jones.ID, evidencegraphs.UpdateInput{Name: "Jones household"})
						return struct{}{}, changes, err
					}); err != nil {
						t.Fatal(err)
					}
				})
				s.bumpsA(t, func() {
					if _, err := writes.Call(s.c, writes.Op{Action: "delete_evidence_graph", UserID: scopeUserID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
						changes, err := evidencegraphs.Delete(tx, scopeUserID, jones.ID)
						return struct{}{}, changes, err
					}); err != nil {
						t.Fatal(err)
					}
				})
			},
		},
		{
			name: "subject create, edit, delete",
			run: func(t *testing.T, s scopeSeed) {
				var subj subjects.Subject
				s.bumpsA(t, func() {
					graph := evidencegraphstest.SoleGraph(t, s.c, s.a.ID).ID
					got, err := writes.Call(s.c, writes.Op{Action: "create_subject", UserID: scopeUserID}, func(tx *database.Tx) (subjects.Subject, []rowchange.Change, error) {
						return subjects.Create(tx, scopeUserID, subjects.CreateInput{
							GraphID: graph, SubjectTypeID: s.personTy, Label: "Cy",
						}, &subjects.Placement{GridX: 16, GridY: 0})
					})
					if err != nil {
						t.Fatal(err)
					}
					subj = got
				})
				s.bumpsA(t, func() {
					if _, err := writes.Call(s.c, writes.Op{Action: "update_subject", UserID: scopeUserID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
						changes, err := subjects.Update(tx, scopeUserID, subj.ID, "Cyril", "")
						return struct{}{}, changes, err
					}); err != nil {
						t.Fatal(err)
					}
				})
				s.bumpsA(t, func() {
					if _, err := writes.Call(s.c, writes.Op{Action: "delete_subject", UserID: scopeUserID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
						changes, err := subjects.Delete(tx, scopeUserID, subj.ID)
						return struct{}{}, changes, err
					}); err != nil {
						t.Fatal(err)
					}
				})
			},
		},
		{
			name: "citation create, edit; observation add, edit, delete; citation delete with notes",
			run: func(t *testing.T, s scopeSeed) {
				var res citations.CreateResult
				s.bumpsA(t, func() {
					in := s.textObs(s.place, "Boston")
					in.Notes = []string{"faded ink"}
					res = s.citation(t, citations.CreateInput{Notes: []string{"p. 3"}}, in)
				})
				s.bumpsA(t, func() {
					if _, err := writes.Call(s.c, writes.Op{Action: "update_citation", UserID: scopeUserID}, func(tx *database.Tx) (citations.Citation, []rowchange.Change, error) {
						return citations.Update(tx, scopeUserID, res.Citation.ID, citations.CitationFieldsInput{
							LocatorJSON: scopeLocator, Transcription: "Bob Smith",
						})
					}); err != nil {
						t.Fatal(err)
					}
				})
				var added []observations.Observation
				s.bumpsA(t, func() {
					got, err := writes.Call(s.c, writes.Op{Action: "add_observations", UserID: scopeUserID}, func(tx *database.Tx) ([]observations.Observation, []rowchange.Change, error) {
						return observations.AddToCitation(tx, scopeUserID, res.Citation.ID,
							[]observations.Input{s.textObs(s.place2, "Salem")})
					})
					if err != nil {
						t.Fatal(err)
					}
					added = got
				})
				s.bumpsA(t, func() {
					s.updateObservation(t, observations.Input{
						ID: added[0].ID, SubjectID: s.place2.ID, PropertyID: s.toponymProp.ID,
						ValueText: "Salem Town", HasText: true,
					})
				})
				// Deleting an observation releases its notes first; the
				// note's change resolves through the deleted observation.
				s.bumpsA(t, func() {
					s.deleteObservation(t, res.Observations[0].ID)
				})
				s.bumpsA(t, func() {
					s.deleteObservation(t, added[0].ID)
				})
				s.bumpsA(t, func() {
					if _, err := writes.Call(s.c, writes.Op{Action: "delete_citation", UserID: scopeUserID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
						changes, err := citations.Delete(tx, scopeUserID, res.Citation.ID)
						return struct{}{}, changes, err
					}); err != nil {
						t.Fatal(err)
					}
				})
			},
		},
		{
			name: "polarity edit keeps the source",
			run: func(t *testing.T, s scopeSeed) {
				res := s.citation(t, citations.CreateInput{}, s.textObs(s.place, "Boston"))
				s.bumpsA(t, func() {
					s.updateObservation(t, observations.Input{
						ID: res.Observations[0].ID, SubjectID: s.place.ID, PropertyID: s.toponymProp.ID,
						ValueText: "Boston", HasText: true, Polarity: observations.PolarityNegative,
					})
				})
			},
		},
		{
			name: "date value edited in place",
			run: func(t *testing.T, s scopeSeed) {
				year := 1842
				res := s.citation(t, citations.CreateInput{}, observations.Input{
					SubjectID: s.event.ID, PropertyID: s.dateProp.ID,
					Date: &datevalue.Value{Kind: datevalue.KindPoint, StartYear: &year},
				})
				year2 := 1843
				s.bumpsA(t, func() {
					s.updateObservation(t, observations.Input{
						ID: res.Observations[0].ID, SubjectID: s.event.ID, PropertyID: s.dateProp.ID,
						Date: &datevalue.Value{Kind: datevalue.KindPoint, StartYear: &year2},
					})
				})
			},
		},
		{
			name: "file rename moves every source using it",
			run: func(t *testing.T, s scopeSeed) {
				path := filepath.Join(t.TempDir(), "scan.txt")
				writeFile(t, path, "page one")
				fres, err := ingest.File(s.c, path, scopeUserID)
				if err != nil {
					t.Fatal(err)
				}
				// The file create has no artifact yet. Checking it after the
				// artifacts exist would see sources the revision did not.
				checkNewEffects(t, s.c)
				for _, src := range []sources.Source{s.a, s.b} {
					if _, err := artifactstest.Create(s.c, scopeUserID, artifacts.CreateInput{
						SourceID: src.ID, FileID: fres.File.ID, Label: "Scan",
					}); err != nil {
						t.Fatal(err)
					}
				}
				if err := runSetFilename(s.c, fres.File.ID, "renamed.txt", scopeUserID); err != nil {
					t.Fatal(err)
				}
				a, b := s.sourceRevisions(t)
				want := s.latest(t)
				if a != want || b != want {
					t.Fatalf("want both at %d got a=%d b=%d", want, a, b)
				}
			},
		},
		{
			name: "promote does not move the source",
			run: func(t *testing.T, s scopeSeed) {
				a0, b0 := s.sourceRevisions(t)
				s.promote(t, s.person)
				if s.latest(t) <= b0 {
					t.Fatal("promote recorded no revision")
				}
				a1, b1 := s.sourceRevisions(t)
				if a1 != a0 || b1 != b0 {
					t.Fatalf("promote moved sources: a %d→%d b %d→%d", a0, a1, b0, b1)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t, newScopeSeed(t))
		})
	}
}

// movesBoth runs act and wants Source A and the handle to sit at the latest
// revision.
func (s scopeSeed) movesBoth(t *testing.T, handle []byte, act func()) {
	t.Helper()
	act()
	checkNewEffects(t, s.c)
	want := s.latest(t)
	if a, _ := s.sourceRevisions(t); a != want {
		t.Fatalf("source A revision %d, want latest %d", a, want)
	}
	if got := s.handleRevision(t, handle); got != want {
		t.Fatalf("handle revision %d, want latest %d", got, want)
	}
}

// TestHandleScopes drives real write paths and checks each one moves the
// canonical_entity scope of exactly the handles it worked on, and a Source
// only where the Source scope says so.
func TestHandleScopes(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, s scopeSeed)
	}{
		{
			name: "promote moves the new handle, not the source",
			run: func(t *testing.T, s scopeSeed) {
				a0, _ := s.sourceRevisions(t)
				h := s.promote(t, s.person)
				if got, want := s.handleRevision(t, h), s.latest(t); got != want {
					t.Fatalf("handle revision %d, want latest %d", got, want)
				}
				if a1, _ := s.sourceRevisions(t); a1 != a0 {
					t.Fatalf("promote moved source A %d → %d", a0, a1)
				}
			},
		},
		{
			name: "unpromote moves the handle, not the source",
			run: func(t *testing.T, s scopeSeed) {
				h := s.promote(t, s.person)
				a0, b0 := s.sourceRevisions(t)
				if _, err := writes.Call(s.c, writes.Op{Action: "unpromote_subject", UserID: scopeUserID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
					changes, err := promote.Unpromote(tx, scopeUserID, s.person.ID)
					return struct{}{}, changes, err
				}); err != nil {
					t.Fatal(err)
				}
				if got, want := s.handleRevision(t, h), s.latest(t); got != want {
					t.Fatalf("handle revision %d, want latest %d", got, want)
				}
				if a1, b1 := s.sourceRevisions(t); a1 != a0 || b1 != b0 {
					t.Fatalf("unpromote moved the sources (%d,%d) → (%d,%d)", a0, b0, a1, b1)
				}
			},
		},
		{
			name: "member observation add, edit, note, delete",
			run: func(t *testing.T, s scopeSeed) {
				h := s.promote(t, s.place)
				var res citations.CreateResult
				s.movesBoth(t, h, func() {
					res = s.citation(t, citations.CreateInput{}, s.textObs(s.place, "Boston"))
				})
				edit := func(text string, notes ...string) observations.Input {
					return observations.Input{
						ID: res.Observations[0].ID, SubjectID: s.place.ID, PropertyID: s.toponymProp.ID,
						ValueText: text, HasText: true, Notes: notes,
					}
				}
				s.movesBoth(t, h, func() { s.updateObservation(t, edit("Boston Town")) })
				s.movesBoth(t, h, func() { s.updateObservation(t, edit("Boston Town", "faded")) })
				s.movesBoth(t, h, func() { s.updateObservation(t, edit("Boston Town", "faded ink")) })
				s.movesBoth(t, h, func() { s.deleteObservation(t, res.Observations[0].ID) })
			},
		},
		{
			name: "member date edited in place",
			run: func(t *testing.T, s scopeSeed) {
				h := s.promote(t, s.event)
				year := 1842
				var res citations.CreateResult
				s.movesBoth(t, h, func() {
					res = s.citation(t, citations.CreateInput{}, observations.Input{
						SubjectID: s.event.ID, PropertyID: s.dateProp.ID,
						Date: &datevalue.Value{Kind: datevalue.KindPoint, StartYear: &year},
					})
				})
				year2 := 1843
				s.movesBoth(t, h, func() {
					s.updateObservation(t, observations.Input{
						ID: res.Observations[0].ID, SubjectID: s.event.ID, PropertyID: s.dateProp.ID,
						Date: &datevalue.Value{Kind: datevalue.KindPoint, StartYear: &year2},
					})
				})
			},
		},
		{
			name: "observation on a Subject with no claim moves no handle",
			run: func(t *testing.T, s scopeSeed) {
				h := s.promote(t, s.place)
				before := s.handleRevision(t, h)
				s.bumpsA(t, func() { s.citation(t, citations.CreateInput{}, s.textObs(s.place2, "Salem")) })
				if got := s.handleRevision(t, h); got != before {
					t.Fatalf("handle moved %d → %d", before, got)
				}
			},
		},
		{
			name: "deleting a member Subject moves its handle and the source",
			run: func(t *testing.T, s scopeSeed) {
				h := s.promote(t, s.person)
				s.movesBoth(t, h, func() {
					if _, err := writes.Call(s.c, writes.Op{Action: "delete_subject", UserID: scopeUserID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
						changes, err := subjects.Delete(tx, scopeUserID, s.person.ID)
						return struct{}{}, changes, err
					}); err != nil {
						t.Fatal(err)
					}
				})
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t, newScopeSeed(t))
		})
	}
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func runArtifactUpdate(c *database.Catalog, userID []byte, a artifacts.Artifact) error {
	_, _, err := writes.Run(c, writes.Op{Action: "update_artifact", UserID: userID},
		func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := artifacts.Update(tx, userID, a, evidencegraphs.GraphLinks{})
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
