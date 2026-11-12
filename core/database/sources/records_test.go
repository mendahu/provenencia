package sources_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/artifacts/artifactstest"
	"github.com/mendahu/provenencia/core/database/evidencegraphs"
	"github.com/mendahu/provenencia/core/database/evidencelayer/evidencelayertest"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sourcecredibility"
	"github.com/mendahu/provenencia/core/database/sourcecredibilitygrades"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"

	_ "github.com/mendahu/provenencia/core/database/autoreconciler" // registers Verify's truth recompute
)

// derivedFromRecords is an entry another PR would register: it is emptied by
// any commit that reloads a Source record.
var derivedFromRecords = graphcache.RegisterWhole(graphcache.WholeEntry[string]{
	Name:        "sources_test.derived",
	DerivedFrom: []graphcache.Source{sources.Records},
})

// recordsFixture is evidencelayertest's catalog with the credibility grades.
func recordsFixture(t *testing.T) *evidencelayertest.Records {
	t.Helper()
	r := evidencelayertest.New(t)
	evidencelayertest.Must(t, sourcecredibilitygrades.Install(r.C))
	return r
}

func gradeID(t *testing.T, r *evidencelayertest.Records, key string) []byte {
	t.Helper()
	g, err := sourcecredibilitygrades.Lookup(r.C, key, vocabkey.OriginProvenencia)
	evidencelayertest.Must(t, err)
	return g.ID
}

func setGrade(t *testing.T, r *evidencelayertest.Records, src sources.Source, key string) {
	t.Helper()
	id := gradeID(t, r, key)
	r.Run("set_source_credibility", func(tx *database.Tx) ([]rowchange.Change, error) {
		_, ch, err := sourcecredibility.Upsert(tx, evidencelayertest.UserID, sourcecredibility.UpsertInput{SourceID: src.ID, CredibilityGradeID: id})
		return ch, err
	})
}

func records(t *testing.T, r *evidencelayertest.Records) *sources.RecordIndex {
	t.Helper()
	idx, err := sources.RecordsOf(r.C.Graph())
	evidencelayertest.Must(t, err)
	return idx
}

// TestRecordsBuild: each Source's record carries its words, its grade read
// relative to the default, and the labels of its Artifacts.
func TestRecordsBuild(t *testing.T) {
	r := recordsFixture(t)
	register := r.Source
	page2, err := artifactstest.Create(r.C, evidencelayertest.UserID, artifacts.CreateInput{SourceID: register.ID, Label: "Page 2"})
	evidencelayertest.Must(t, err)
	census, _ := r.NewSource("Census")
	gazette, _ := r.NewSource("Gazette")
	setGrade(t, r, register, "low_trust")
	setGrade(t, r, gazette, "high_trust")
	// Census keeps no Artifact in this case: its only Artifact is deleted.
	cArts, err := artifacts.ListBySource(r.C, census.ID)
	evidencelayertest.Must(t, err)
	r.Run("delete_artifact", func(tx *database.Tx) ([]rowchange.Change, error) {
		return artifacts.Delete(tx, evidencelayertest.UserID, cArts[0].ID)
	})

	idx := records(t, r)
	cases := []struct {
		name   string
		src    sources.Source
		key    string
		cred   int
		labels int
	}{
		{"low trust", register, "low_trust", -1, 2},
		{"no assessment", census, "", 0, 0},
		{"high trust", gazette, "high_trust", 1, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := idx.Source(tc.src.ID)
			if !ok {
				t.Fatal("no record")
			}
			if got.Ref != tc.src.Ref || got.Title != tc.src.Title || !bytes.Equal(got.SourceTypeID, tc.src.SourceTypeID) {
				t.Fatalf("words %q %q %x", got.Ref, got.Title, got.SourceTypeID)
			}
			if got.CredibilityKey != tc.key || got.Credibility != tc.cred {
				t.Fatalf("credibility %q %d, want %q %d", got.CredibilityKey, got.Credibility, tc.key, tc.cred)
			}
			if got.ArtifactLabels == nil || len(got.ArtifactLabels) != tc.labels {
				t.Fatalf("artifact labels %v, want %d", got.ArtifactLabels, tc.labels)
			}
		})
	}
	rec, label, ok := idx.Artifact(page2.ID)
	if !ok || rec.Title != "Register" || label != "Page 2" {
		t.Fatalf("Artifact(page 2) = %q %q %v", rec.Title, label, ok)
	}
	if _, _, ok := idx.Artifact(make([]byte, 16)); ok {
		t.Fatal("an unknown Artifact resolved")
	}
	if _, ok := idx.Source(make([]byte, 16)); ok {
		t.Fatal("an unknown Source resolved")
	}
}

// TestRecordsApply: every write that changes a Source's words leaves the held
// records equal to a fresh build, and empties what is derived from them.
func TestRecordsApply(t *testing.T) {
	r := recordsFixture(t)
	g := r.C.Graph()
	src := r.Source
	records(t, r)
	letter, err := sourcetypes.Upsert(r.C, sourcetypes.Type{Key: "letter", Origin: vocabkey.OriginProvenencia, Label: "Letter"})
	evidencelayertest.Must(t, err)
	var extra artifacts.Artifact

	check := func(t *testing.T, write func()) {
		t.Helper()
		derivedFromRecords.Set(g, "held")
		write()
		if err := g.Verify(); err != nil {
			t.Fatalf("Verify after the write: %v", err)
		}
		if _, ok, _ := derivedFromRecords.Get(g); ok {
			t.Fatal("the derived entry survived a commit that names a Source")
		}
	}
	current := func(t *testing.T) sources.Record {
		t.Helper()
		rec, ok := records(t, r).Source(src.ID)
		if !ok {
			t.Fatal("the Source has no record")
		}
		return rec
	}
	update := func(s sources.Source) {
		r.Run("update_source", func(tx *database.Tx) ([]rowchange.Change, error) {
			return sources.Update(tx, evidencelayertest.UserID, s)
		})
	}

	t.Run("rename", func(t *testing.T) {
		check(t, func() { src.Title = "Parish register"; update(src) })
		if current(t).Title != "Parish register" {
			t.Fatal("title not updated")
		}
	})
	t.Run("type change", func(t *testing.T) {
		check(t, func() { src.SourceTypeID = letter; update(src) })
		if !bytes.Equal(current(t).SourceTypeID, letter) {
			t.Fatal("type not updated")
		}
	})
	t.Run("artifact create", func(t *testing.T) {
		check(t, func() {
			var err error
			extra, err = artifactstest.Create(r.C, evidencelayertest.UserID, artifacts.CreateInput{SourceID: src.ID, Label: "Page 1"})
			evidencelayertest.Must(t, err)
		})
		if _, label, ok := records(t, r).Artifact(extra.ID); !ok || label != "Page 1" {
			t.Fatalf("new Artifact resolves to %q %v", label, ok)
		}
	})
	t.Run("artifact relabel", func(t *testing.T) {
		check(t, func() {
			extra.Label = "Folio 1"
			r.Run("update_artifact", func(tx *database.Tx) ([]rowchange.Change, error) {
				return artifacts.Update(tx, evidencelayertest.UserID, extra, evidencegraphs.GraphLinks{})
			})
		})
		if _, label, _ := records(t, r).Artifact(extra.ID); label != "Folio 1" {
			t.Fatalf("label %q", label)
		}
	})
	t.Run("credibility set", func(t *testing.T) {
		check(t, func() { setGrade(t, r, src, "high_trust") })
		if rec := current(t); rec.CredibilityKey != "high_trust" || rec.Credibility != 1 {
			t.Fatalf("credibility %q %d", rec.CredibilityKey, rec.Credibility)
		}
	})
	t.Run("observation edit", func(t *testing.T) {
		check(t, func() { r.Person(src, r.Artifact, "Ada") })
	})
	t.Run("artifact delete", func(t *testing.T) {
		check(t, func() {
			r.Run("delete_artifact", func(tx *database.Tx) ([]rowchange.Change, error) {
				return artifacts.Delete(tx, evidencelayertest.UserID, extra.ID)
			})
		})
		if _, _, ok := records(t, r).Artifact(extra.ID); ok {
			t.Fatal("the deleted Artifact still resolves")
		}
	})
	t.Run("source delete", func(t *testing.T) {
		var second sources.Source
		evidencelayertest.Must(t, writesCreate(r, &second))
		records(t, r)
		check(t, func() {
			r.Run("delete_source", func(tx *database.Tx) ([]rowchange.Change, error) {
				return sources.Delete(tx, evidencelayertest.UserID, second.ID)
			})
		})
		if _, ok := records(t, r).Source(second.ID); ok {
			t.Fatal("the deleted Source still resolves")
		}
	})
}

func writesCreate(r *evidencelayertest.Records, out *sources.Source) error {
	typeID := r.Source.SourceTypeID
	s, _, err := writes.Run(r.C, writes.Op{Action: "create_source", UserID: evidencelayertest.UserID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, evidencelayertest.UserID, sources.CreateInput{SourceTypeID: typeID, Title: "Second"})
	})
	*out = s
	return err
}

// TestRecordsOf: a write's graph reads no records and builds none; a nil graph
// is a closed catalog.
func TestRecordsOf(t *testing.T) {
	r := recordsFixture(t)
	r.Run("probe", func(tx *database.Tx) ([]rowchange.Change, error) {
		queries := 0
		over := graphcache.Over(tx.Tx)
		over.SetQueryCounterForTest(func() { queries++ })
		if _, err := sources.RecordsOf(over); !errors.Is(err, sources.ErrRecordsOnWrite) {
			t.Fatalf("RecordsOf on a write's graph: %v", err)
		}
		if queries != 0 {
			t.Fatalf("a write's graph ran %d queries", queries)
		}
		return nil, nil
	})
	if _, err := sources.RecordsOf(nil); !errors.Is(err, database.ErrClosed) {
		t.Fatalf("RecordsOf(nil): %v", err)
	}
}
