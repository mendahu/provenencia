package sourcecopy

import (
	"bytes"
	"errors"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/metadatafields"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sourcecredibility"
	"github.com/mendahu/provenencia/core/database/sourcecredibilitygrades"
	"github.com/mendahu/provenencia/core/database/sourcemetadata"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/sourcevocab"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ingest"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
)

const originalDescription = "Medicine Hat News"

func writeTinyPNG(t *testing.T, path string) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}

func runCopy(c *database.Catalog, userID []byte, in CopyInput) (sources.Source, error) {
	src, _, err := writes.Run(c, writes.Op{Action: "copy_source", UserID: userID},
		func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
			return Copy(tx, userID, in)
		})
	return src, err
}

// seedObituary builds the AC-1.1 original: type obituary, a description, two
// metadata values in a hand-sorted layout with one dismissed suggestion, a
// credibility grade and argument, two Artifacts with Files (the first a real
// PNG and the Source's cover), a note and a Subject.
func seedObituary(t *testing.T) (c *database.Catalog, userID []byte, orig sources.Source, arts []artifacts.Artifact) {
	t.Helper()
	userID = []byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	ur, err := ref.Mint(ref.PrefixUser)
	if err != nil {
		t.Fatal(err)
	}
	if err := users.Upsert(c, userID, "Jake", ur); err != nil {
		t.Fatal(err)
	}

	obit, err := sourcetypes.Upsert(c, sourcetypes.Type{Key: "obituary", Origin: vocabkey.OriginUser, Label: "Obituary"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := sourcetypes.Upsert(c, sourcetypes.Type{Key: "newspaper", Origin: vocabkey.OriginUser, Label: "Newspaper"}); err != nil {
		t.Fatal(err)
	}
	field := func(key string, dataType string) []byte {
		id, err := metadatafields.Upsert(c, metadatafields.Field{Key: key, Origin: vocabkey.OriginUser, Label: key, DataType: dataType})
		if err != nil {
			t.Fatal(err)
		}
		return id
	}
	publication := field("publication", metadatafields.DataTypeText)
	link := field("url", metadatafields.DataTypeURL)
	page := field("page", metadatafields.DataTypeText)
	for i, f := range [][]byte{publication, link, page} {
		if err := sourcevocab.EnsureSuggestion(c, obit, f, i); err != nil {
			t.Fatal(err)
		}
	}
	if err := sourcecredibilitygrades.Install(c); err != nil {
		t.Fatal(err)
	}
	grades, err := sourcecredibilitygrades.List(c)
	if err != nil || len(grades) == 0 {
		t.Fatalf("grades %v err %v", grades, err)
	}
	grade := grades[0]

	orig, err = writeOne(c, userID, "create_source", func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, userID, sources.CreateInput{
			SourceTypeID: obit, Title: "Obituary of Ann Smith", Description: originalDescription,
		})
	})
	if err != nil {
		t.Fatal(err)
	}
	steps := []func(tx *database.Tx) ([]rowchange.Change, error){
		func(tx *database.Tx) ([]rowchange.Change, error) {
			_, ch, err := sourcemetadata.Set(tx, userID, sourcemetadata.Input{SourceID: orig.ID, FieldID: publication, ValueText: "Medicine Hat News"})
			return ch, err
		},
		func(tx *database.Tx) ([]rowchange.Change, error) {
			_, ch, err := sourcemetadata.Set(tx, userID, sourcemetadata.Input{SourceID: orig.ID, FieldID: link, ValueText: "https://example.org/a"})
			return ch, err
		},
		func(tx *database.Tx) ([]rowchange.Change, error) {
			return sourcemetadata.DismissSuggestion(tx, userID, orig.ID, page)
		},
		func(tx *database.Tx) ([]rowchange.Change, error) {
			return sourcemetadata.Reorder(tx, userID, orig.ID, [][]byte{link, publication, page})
		},
		func(tx *database.Tx) ([]rowchange.Change, error) {
			_, ch, err := sourcecredibility.Upsert(tx, userID, sourcecredibility.UpsertInput{
				SourceID: orig.ID, CredibilityGradeID: grade.ID, Argument: "Family-placed notice",
			})
			return ch, err
		},
		func(tx *database.Tx) ([]rowchange.Change, error) {
			_, ch, err := sources.AddNote(tx, userID, orig.ID, "Check the date")
			return ch, err
		},
	}
	for _, step := range steps {
		if _, _, err := writes.Run(c, writes.Op{Action: "seed", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			ch, err := step(tx)
			return struct{}{}, ch, err
		}); err != nil {
			t.Fatal(err)
		}
	}

	pngPath := filepath.Join(t.TempDir(), "page1.png")
	writeTinyPNG(t, pngPath)
	otherPath := filepath.Join(t.TempDir(), "page2.bin")
	if err := os.WriteFile(otherPath, []byte("page two bytes"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, spec := range []struct{ path, label, desc string }{
		{pngPath, "Page 1", "Front page"},
		{otherPath, "Page 2", "Obituary column"},
	} {
		res, err := ingest.File(c, spec.path, userID)
		if err != nil {
			t.Fatal(err)
		}
		art, err := writeOne(c, userID, "create_artifact", func(tx *database.Tx) (artifacts.Artifact, []rowchange.Change, error) {
			return artifacts.Create(tx, userID, artifacts.CreateInput{
				SourceID: orig.ID, FileID: res.File.ID, Label: spec.label, Description: spec.desc,
			})
		})
		if err != nil {
			t.Fatal(err)
		}
		arts = append(arts, art)
	}
	mode, artID, err := sources.NormalizeCover(c, orig.ID, sources.CoverModeArtifact, arts[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := writes.Run(c, writes.Op{Action: "set_source_cover", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
		_, ch, err := sources.SetCover(tx, userID, orig.ID, mode, artID)
		return struct{}{}, ch, err
	}); err != nil {
		t.Fatal(err)
	}

	if err := subjectvocab.Install(c); err != nil {
		t.Fatal(err)
	}
	person, err := subjecttypes.Lookup(c, "person", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writes.Call(c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (subjects.Subject, []rowchange.Change, error) {
		return subjects.Create(tx, userID, subjects.CreateInput{SourceID: orig.ID, SubjectTypeID: person.ID, Label: "Ann"}, nil)
	}); err != nil {
		t.Fatal(err)
	}
	orig, err = sources.Get(c, orig.ID)
	if err != nil {
		t.Fatal(err)
	}
	return c, userID, orig, arts
}

func writeOne[T any](c *database.Catalog, userID []byte, action string, fn func(tx *database.Tx) (T, []rowchange.Change, error)) (T, error) {
	v, _, err := writes.Run(c, writes.Op{Action: action, UserID: userID}, fn)
	return v, err
}

type wsKey struct {
	FieldID   string
	ValueText string
	SortOrder int
	Suggested bool
	HasValue  bool
}

func workspaceKeys(t *testing.T, c *database.Catalog, sourceID []byte) []wsKey {
	t.Helper()
	ws, err := sourcemetadata.ListWorkspace(c, sourceID)
	if err != nil {
		t.Fatal(err)
	}
	out := make([]wsKey, 0, len(ws))
	for _, e := range ws {
		k := wsKey{FieldID: string(e.Field.ID), SortOrder: e.SortOrder, Suggested: e.Suggested}
		if e.Value != nil {
			k.HasValue = true
			k.ValueText = e.Value.ValueText
		}
		out = append(out, k)
	}
	return out
}

type artKey struct{ Label, Description, FileID string }

func artifactKeys(t *testing.T, c *database.Catalog, sourceID []byte) []artKey {
	t.Helper()
	list, err := artifacts.ListBySource(c, sourceID)
	if err != nil {
		t.Fatal(err)
	}
	out := make([]artKey, 0, len(list))
	for _, a := range list {
		out = append(out, artKey{a.Label, a.Description, string(a.FileID)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Label < out[j].Label })
	return out
}

func countFor(t *testing.T, c *database.Catalog, query string, sourceID []byte) int {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.QueryRow(query, sourceID).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func layoutOf(t *testing.T, c *database.Catalog, sourceID []byte) []sourcemetadata.LayoutRow {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	rows, err := sourcemetadata.ListLayoutTx(tx, sourceID)
	if err != nil {
		t.Fatal(err)
	}
	return rows
}

type snapshot struct {
	Source      sources.Source
	Notes       []sources.Note
	Workspace   []wsKey
	Layout      []sourcemetadata.LayoutRow
	Credibility sourcecredibility.Assessment
	Artifacts   []artifacts.Artifact
}

func snap(t *testing.T, c *database.Catalog, id []byte) snapshot {
	t.Helper()
	src, err := sources.Get(c, id)
	if err != nil {
		t.Fatal(err)
	}
	notes, err := sources.ListNotes(c, id)
	if err != nil {
		t.Fatal(err)
	}
	cred, err := sourcecredibility.GetBySource(c, id)
	if err != nil {
		t.Fatal(err)
	}
	arts, err := artifacts.ListBySource(c, id)
	if err != nil {
		t.Fatal(err)
	}
	return snapshot{
		Source:      src,
		Notes:       notes,
		Workspace:   workspaceKeys(t, c, id),
		Layout:      layoutOf(t, c, id),
		Credibility: cred,
		Artifacts:   arts,
	}
}

func TestCopy(t *testing.T) {
	allParts := func(orig sources.Source, arts []artifacts.Artifact) CopyInput {
		return CopyInput{
			SourceID: orig.ID, Title: "Copy of Obituary of Ann Smith", SourceTypeID: orig.SourceTypeID,
			Description: true, Metadata: true, Credibility: true,
			ArtifactIDs: [][]byte{arts[0].ID, arts[1].ID},
		}
	}
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "all parts",
			run: func(t *testing.T) {
				c, userID, orig, arts := seedObituary(t)
				cp, err := runCopy(c, userID, allParts(orig, arts))
				if err != nil {
					t.Fatal(err)
				}
				if cp.Ref == orig.Ref || cp.Ref == "" || bytes.Equal(cp.ID, orig.ID) {
					t.Fatalf("ref/id %q %q", cp.Ref, orig.Ref)
				}
				got, err := sources.Get(c, cp.ID)
				if err != nil {
					t.Fatal(err)
				}
				if got.Title != "Copy of Obituary of Ann Smith" || !bytes.Equal(got.SourceTypeID, orig.SourceTypeID) ||
					got.Description != originalDescription {
					t.Fatalf("copy %+v", got)
				}
				if got.CoverMode != sources.CoverModeTypeIcon || got.PrimaryArtifactID != nil {
					t.Fatalf("cover %q %x", got.CoverMode, got.PrimaryArtifactID)
				}
				want, have := workspaceKeys(t, c, orig.ID), workspaceKeys(t, c, cp.ID)
				// The dismissed suggestion is hidden, so the workspace holds two entries.
				if len(want) != 2 || !reflect.DeepEqual(want, have) {
					t.Fatalf("workspace\nwant %+v\nhave %+v", want, have)
				}
				if want, have := layoutOf(t, c, orig.ID), layoutOf(t, c, cp.ID); len(want) != 3 || !reflect.DeepEqual(want, have) {
					t.Fatalf("layout\nwant %+v\nhave %+v", want, have)
				}
				origCred, err := sourcecredibility.GetBySource(c, orig.ID)
				if err != nil {
					t.Fatal(err)
				}
				cred, err := sourcecredibility.GetBySource(c, cp.ID)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(cred.CredibilityGradeID, origCred.CredibilityGradeID) || cred.Argument != "Family-placed notice" ||
					bytes.Equal(cred.ID, origCred.ID) {
					t.Fatalf("credibility %+v", cred)
				}
				if want, have := artifactKeys(t, c, orig.ID), artifactKeys(t, c, cp.ID); len(want) != 2 || !reflect.DeepEqual(want, have) {
					t.Fatalf("artifacts\nwant %+v\nhave %+v", want, have)
				}
				origRefs := map[string]bool{}
				origList, err := artifacts.ListBySource(c, orig.ID)
				if err != nil {
					t.Fatal(err)
				}
				for _, a := range origList {
					origRefs[a.Ref] = true
				}
				copyList, err := artifacts.ListBySource(c, cp.ID)
				if err != nil {
					t.Fatal(err)
				}
				for _, a := range copyList {
					if origRefs[a.Ref] || a.Ref == "" {
						t.Fatalf("artifact ref %q reused", a.Ref)
					}
				}
			},
		},
		{
			name: "original untouched",
			run: func(t *testing.T) {
				c, userID, orig, arts := seedObituary(t)
				before := snap(t, c, orig.ID)
				if _, err := runCopy(c, userID, allParts(orig, arts)); err != nil {
					t.Fatal(err)
				}
				after := snap(t, c, orig.ID)
				if !reflect.DeepEqual(before, after) {
					t.Fatalf("original changed\nbefore %+v\nafter  %+v", before, after)
				}
				if after.Source.CoverMode != sources.CoverModeArtifact {
					t.Fatalf("cover %q", after.Source.CoverMode)
				}
			},
		},
		{
			name: "nothing chosen",
			run: func(t *testing.T) {
				c, userID, orig, _ := seedObituary(t)
				cp, err := runCopy(c, userID, CopyInput{SourceID: orig.ID, Title: "Bare", SourceTypeID: orig.SourceTypeID})
				if err != nil {
					t.Fatal(err)
				}
				got, err := sources.Get(c, cp.ID)
				if err != nil {
					t.Fatal(err)
				}
				if got.Description != "" {
					t.Fatalf("description %q", got.Description)
				}
				for _, table := range []string{"source_metadata", "source_metadata_layout", "source_credibility_assessments", "artifacts"} {
					if n := countFor(t, c, `SELECT COUNT(*) FROM `+table+` WHERE source_id = ?`, cp.ID); n != 0 {
						t.Fatalf("%s rows %d", table, n)
					}
				}
			},
		},
		{
			name: "second Artifact only",
			run: func(t *testing.T) {
				c, userID, orig, arts := seedObituary(t)
				in := allParts(orig, arts)
				in.ArtifactIDs = [][]byte{arts[1].ID}
				cp, err := runCopy(c, userID, in)
				if err != nil {
					t.Fatal(err)
				}
				got := artifactKeys(t, c, cp.ID)
				want := artKey{arts[1].Label, arts[1].Description, string(arts[1].FileID)}
				if len(got) != 1 || got[0] != want {
					t.Fatalf("artifacts %+v want %+v", got, want)
				}
			},
		},
		{
			name: "note and Subject not copied",
			run: func(t *testing.T) {
				c, userID, orig, arts := seedObituary(t)
				cp, err := runCopy(c, userID, allParts(orig, arts))
				if err != nil {
					t.Fatal(err)
				}
				if n := countFor(t, c, `SELECT COUNT(*) FROM source_notes WHERE source_id = ?`, cp.ID); n != 0 {
					t.Fatalf("notes %d", n)
				}
				if n := countFor(t, c, `SELECT COUNT(*) FROM subjects WHERE source_id = ?`, cp.ID); n != 0 {
					t.Fatalf("subjects %d", n)
				}
				if n := countFor(t, c, `SELECT COUNT(*) FROM citations WHERE artifact_id IN (SELECT id FROM artifacts WHERE source_id = ?)`, cp.ID); n != 0 {
					t.Fatalf("citations %d", n)
				}
			},
		},
		{
			name: "other type",
			run: func(t *testing.T) {
				c, userID, orig, arts := seedObituary(t)
				newspaper, err := sourcetypes.Lookup(c, "newspaper", vocabkey.OriginUser)
				if err != nil {
					t.Fatal(err)
				}
				in := allParts(orig, arts)
				in.SourceTypeID = newspaper.ID
				cp, err := runCopy(c, userID, in)
				if err != nil {
					t.Fatal(err)
				}
				got, err := sources.Get(c, cp.ID)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(got.SourceTypeID, newspaper.ID) {
					t.Fatalf("type %x", got.SourceTypeID)
				}
				if want, have := layoutOf(t, c, orig.ID), layoutOf(t, c, cp.ID); len(want) != 3 || !reflect.DeepEqual(want, have) {
					t.Fatalf("layout\nwant %+v\nhave %+v", want, have)
				}
				type valueKey struct{ FieldID, ValueText string }
				valueKeys := func(id []byte) []valueKey {
					rows, err := sourcemetadata.ListBySource(c, id)
					if err != nil {
						t.Fatal(err)
					}
					out := make([]valueKey, 0, len(rows))
					for _, r := range rows {
						out = append(out, valueKey{string(r.FieldID), r.ValueText})
					}
					sort.Slice(out, func(i, j int) bool { return out[i].FieldID < out[j].FieldID })
					return out
				}
				if want, have := valueKeys(orig.ID), valueKeys(cp.ID); len(want) != 2 || !reflect.DeepEqual(want, have) {
					t.Fatalf("values\nwant %+v\nhave %+v", want, have)
				}
			},
		},
		{
			name: "invalid input writes nothing",
			run: func(t *testing.T) {
				c, userID, orig, arts := seedObituary(t)
				otherSrc, err := writeOne(c, userID, "create_source", func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
					return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: orig.SourceTypeID, Title: "Other"})
				})
				if err != nil {
					t.Fatal(err)
				}
				foreign, err := writeOne(c, userID, "create_artifact", func(tx *database.Tx) (artifacts.Artifact, []rowchange.Change, error) {
					return artifacts.Create(tx, userID, artifacts.CreateInput{SourceID: otherSrc.ID, Label: "Scan"})
				})
				if err != nil {
					t.Fatal(err)
				}
				missing := make([]byte, 16)
				missing[15] = 9

				cases := []struct {
					name   string
					mutate func(in *CopyInput)
				}{
					{"blank title", func(in *CopyInput) { in.Title = "   " }},
					{"unknown source", func(in *CopyInput) { in.SourceID = missing }},
					{"foreign Artifact", func(in *CopyInput) { in.ArtifactIDs = [][]byte{arts[0].ID, foreign.ID} }},
					{"repeated Artifact", func(in *CopyInput) { in.ArtifactIDs = [][]byte{arts[0].ID, arts[0].ID} }},
					{"unknown type", func(in *CopyInput) { in.SourceTypeID = missing }},
				}
				for _, tc := range cases {
					t.Run(tc.name, func(t *testing.T) {
						before, err := sources.Count(c)
						if err != nil {
							t.Fatal(err)
						}
						in := allParts(orig, arts)
						tc.mutate(&in)
						if _, err := runCopy(c, userID, in); !errors.Is(err, sources.ErrInvalid) {
							t.Fatalf("got %v", err)
						}
						after, err := sources.Count(c)
						if err != nil {
							t.Fatal(err)
						}
						if after != before {
							t.Fatalf("sources %d -> %d", before, after)
						}
					})
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}
