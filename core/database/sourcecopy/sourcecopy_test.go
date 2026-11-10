package sourcecopy

import (
	"bytes"
	"context"
	"encoding/json"
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
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
	"github.com/mendahu/provenencia/core/database/metadatafields"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sourcecredibility"
	"github.com/mendahu/provenencia/core/database/sourcecredibilitygrades"
	"github.com/mendahu/provenencia/core/database/sourcemetadata"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/sourcevocab"
	"github.com/mendahu/provenencia/core/database/subjects/subjectstest"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/datevalue/datevaluetest"
	"github.com/mendahu/provenencia/core/derivatives"
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
	if _, err := derivatives.EnsureThumbnail(c, arts[0].FileID); err != nil {
		t.Fatal(err)
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
	graph := evidencegraphstest.SoleGraph(t, c, orig.ID).ID
	subjectstest.On(t, c, userID, graph, "person", "Ann")
	orig, err = sources.Get(c, orig.ID)
	if err != nil {
		t.Fatal(err)
	}
	return c, userID, orig, arts
}

// setDate dates a Source through the one owner write.
func setDate(t *testing.T, c *database.Catalog, userID, sourceID []byte, d *datevalue.Value) sources.Source {
	t.Helper()
	src, err := writeOne(c, userID, "set_source_date", func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.SetDate(tx, userID, sourceID, d)
	})
	if err != nil {
		t.Fatal(err)
	}
	return src
}

// dateRows counts the date_values rows.
func dateRows(t *testing.T, c *database.Catalog) int {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM date_values`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// revisionChanges reads every change of the latest revision, in write order.
func revisionChanges(t *testing.T, c *database.Catalog) []rowchange.Change {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	rows, err := db.Query(`SELECT entity_type, entity_id, action, changes_json FROM audit_changes
		WHERE audit_transaction_id = (SELECT id FROM audit_transactions ORDER BY revision DESC LIMIT 1)
		ORDER BY rowid`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out []rowchange.Change
	for rows.Next() {
		var ch rowchange.Change
		var raw string
		if err := rows.Scan(&ch.EntityType, &ch.EntityID, &ch.Action, &raw); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal([]byte(raw), &ch.Fields); err != nil {
			t.Fatal(err)
		}
		out = append(out, ch)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
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

// graphLinks returns the one graph of a Source and the Artifact ids linked to it.
func graphLinks(t *testing.T, c *database.Catalog, sourceID []byte) ([]byte, []string) {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	var graph []byte
	if err := db.QueryRow(`SELECT id FROM evidence_graphs WHERE source_id = ?`, sourceID).Scan(&graph); err != nil {
		t.Fatal(err)
	}
	rows, err := db.Query(`SELECT artifact_id FROM evidence_graph_artifacts WHERE graph_id = ?`, graph)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var linked []string
	for rows.Next() {
		var id []byte
		if err := rows.Scan(&id); err != nil {
			t.Fatal(err)
		}
		linked = append(linked, string(id))
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	sort.Strings(linked)
	return graph, linked
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
				origGraph, origLinks := graphLinks(t, c, orig.ID)
				cp, err := runCopy(c, userID, allParts(orig, arts))
				if err != nil {
					t.Fatal(err)
				}
				copyGraph, copyLinks := graphLinks(t, c, cp.ID)
				if bytes.Equal(copyGraph, origGraph) {
					t.Fatal("copy shares the original's graph")
				}
				wantLinks := []string{}
				copied, err := artifacts.ListBySource(c, cp.ID)
				if err != nil {
					t.Fatal(err)
				}
				for _, a := range copied {
					wantLinks = append(wantLinks, string(a.ID))
				}
				sort.Strings(wantLinks)
				if len(wantLinks) != 2 || !reflect.DeepEqual(copyLinks, wantLinks) {
					t.Fatalf("copy links %x want %x", copyLinks, wantLinks)
				}
				if g, l := graphLinks(t, c, orig.ID); !bytes.Equal(g, origGraph) || !reflect.DeepEqual(l, origLinks) {
					t.Fatalf("original graph links changed: %x %x", g, l)
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
			name: "carries the date as its own row",
			run: func(t *testing.T) {
				for _, tc := range []struct {
					name string
					date *datevalue.Value
				}{
					{"range", datevaluetest.Between(1834, 0, 0, 1840, 0, 0)},
					{"phrase", datevaluetest.Phrase("in the reign of Victoria")},
				} {
					t.Run(tc.name, func(t *testing.T) {
						c, userID, orig, arts := seedObituary(t)
						orig = setDate(t, c, userID, orig.ID, tc.date)
						in := allParts(orig, arts)
						in.Date = true
						cp, err := runCopy(c, userID, in)
						if err != nil {
							t.Fatal(err)
						}
						if cp.Date == nil || bytes.Equal(cp.DateValueID, orig.DateValueID) || len(cp.DateValueID) != 16 {
							t.Fatalf("copy date %+v id %x original %x", cp.Date, cp.DateValueID, orig.DateValueID)
						}
						want, have := *orig.Date, *cp.Date
						want.ID, have.ID = nil, nil
						if !reflect.DeepEqual(want, have) {
							t.Fatalf("date\nwant %+v\nhave %+v", want, have)
						}
						var creates, updates int
						var seen []string
						for _, ch := range revisionChanges(t, c) {
							seen = append(seen, ch.EntityType+" "+string(ch.Action))
							switch {
							case ch.EntityType == "date_value" && ch.Action == rowchange.ActionCreate:
								creates++
							case ch.EntityType == "source" && ch.Action == rowchange.ActionUpdate && bytes.Equal(ch.EntityID, cp.ID):
								if _, ok := ch.Fields["date_value_id"]; ok {
									updates++
								}
							}
						}
						if creates != 1 || updates != 1 {
							t.Fatalf("revision changes %v", seen)
						}

						// The two rows are independent: dating the copy leaves the original.
						setDate(t, c, userID, cp.ID, datevaluetest.YMD(1841, 0, 0))
						got, err := sources.Get(c, orig.ID)
						if err != nil {
							t.Fatal(err)
						}
						if got.Date == nil || !bytes.Equal(got.DateValueID, orig.DateValueID) || !reflect.DeepEqual(*got.Date, *orig.Date) {
							t.Fatalf("original date changed: %+v", got.Date)
						}
					})
				}
			},
		},
		{
			name: "date chosen on an undated original",
			run: func(t *testing.T) {
				t.Run("undated", func(t *testing.T) {
					c, userID, orig, arts := seedObituary(t)
					in := allParts(orig, arts)
					in.Date = true
					cp, err := runCopy(c, userID, in)
					if err != nil {
						t.Fatal(err)
					}
					if cp.Date != nil || cp.DateValueID != nil || dateRows(t, c) != 0 {
						t.Fatalf("copy %+v, date rows %d", cp.Date, dateRows(t, c))
					}
					for _, ch := range revisionChanges(t, c) {
						if ch.EntityType == "date_value" {
							t.Fatalf("date_value change %+v", ch)
						}
					}
				})
				t.Run("date row missing", func(t *testing.T) {
					c, userID, orig, arts := seedObituary(t)
					orig = setDate(t, c, userID, orig.ID, datevaluetest.YMD(1880, 0, 0))
					db, err := c.DB()
					if err != nil {
						t.Fatal(err)
					}
					ctx := context.Background()
					conn, err := db.Conn(ctx)
					if err != nil {
						t.Fatal(err)
					}
					if _, err := conn.ExecContext(ctx, `PRAGMA foreign_keys=OFF`); err != nil {
						t.Fatal(err)
					}
					if _, err := conn.ExecContext(ctx, `DELETE FROM date_values WHERE id = ?`, orig.DateValueID); err != nil {
						t.Fatal(err)
					}
					if _, err := conn.ExecContext(ctx, `PRAGMA foreign_keys=ON`); err != nil {
						t.Fatal(err)
					}
					// Release the connection: the catalog holds one, and the copy needs it.
					if err := conn.Close(); err != nil {
						t.Fatal(err)
					}
					in := allParts(orig, arts)
					in.Date = true
					cp, err := runCopy(c, userID, in)
					if err != nil {
						t.Fatal(err)
					}
					if cp.Date != nil || cp.DateValueID != nil || dateRows(t, c) != 0 {
						t.Fatalf("copy %+v, date rows %d", cp.Date, dateRows(t, c))
					}
					for _, ch := range revisionChanges(t, c) {
						if ch.EntityType == "date_value" {
							t.Fatalf("date_value change %+v", ch)
						}
					}
				})
			},
		},
		{
			name: "date not chosen",
			run: func(t *testing.T) {
				c, userID, orig, arts := seedObituary(t)
				orig = setDate(t, c, userID, orig.ID, datevaluetest.YMD(1880, 7, 4))
				before := dateRows(t, c)
				cp, err := runCopy(c, userID, allParts(orig, arts))
				if err != nil {
					t.Fatal(err)
				}
				if cp.Date != nil || cp.DateValueID != nil || dateRows(t, c) != before {
					t.Fatalf("copy %+v, date rows %d -> %d", cp.Date, before, dateRows(t, c))
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
					want   error
				}{
					{"blank title", func(in *CopyInput) { in.Title = "   " }, sources.ErrInvalid},
					{"unknown source", func(in *CopyInput) { in.SourceID = missing }, sources.ErrNotFound},
					{"foreign Artifact", func(in *CopyInput) { in.ArtifactIDs = [][]byte{arts[0].ID, foreign.ID} }, sources.ErrInvalid},
					{"repeated Artifact", func(in *CopyInput) { in.ArtifactIDs = [][]byte{arts[0].ID, arts[0].ID} }, sources.ErrInvalid},
					{"unknown type", func(in *CopyInput) { in.SourceTypeID = missing }, sources.ErrInvalid},
				}
				for _, tc := range cases {
					t.Run(tc.name, func(t *testing.T) {
						before, err := sources.Count(c)
						if err != nil {
							t.Fatal(err)
						}
						in := allParts(orig, arts)
						tc.mutate(&in)
						if _, err := runCopy(c, userID, in); !errors.Is(err, tc.want) {
							t.Fatalf("got %v, want %v", err, tc.want)
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

func TestReadDraftMissingSourceIsNotFound(t *testing.T) {
	c, _, _, _ := seedObituary(t)
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	missing := make([]byte, 16)
	missing[15] = 9
	if _, err := ReadDraft(tx, missing); !errors.Is(err, sources.ErrNotFound) {
		t.Fatalf("got %v, want sources.ErrNotFound", err)
	}
}
