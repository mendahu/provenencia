package ingest

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/writes"
)

// symlinkObjects replaces the project's objects/ folder with a symlink to a
// folder outside the project.
func symlinkObjects(t *testing.T, c *database.Catalog) {
	t.Helper()
	objects := filepath.Join(c.Dir(), "objects")
	if err := os.RemoveAll(objects); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(t.TempDir(), objects); err != nil {
		t.Fatal(err)
	}
}

func wantUnsafeObjects(t *testing.T, err error) {
	t.Helper()
	if !errors.Is(err, ErrUnsafeObjects) {
		t.Fatalf("err %v, want ErrUnsafeObjects", err)
	}
	if code := apperr.From(err).Code(); code != "ingest.unsafe_objects" {
		t.Fatalf("code %q, want ingest.unsafe_objects", code)
	}
}

func TestFileRefusesSymlinkedObjects(t *testing.T) {
	c := stageCatalog(t)
	userID := stageUser(t, c)
	path := stageWrite(t, t.TempDir(), "in.png", stagePNG(t, 8, 8))
	symlinkObjects(t, c)
	_, err := File(c, path, userID)
	wantUnsafeObjects(t, err)
}

func TestStagedImageInsertTxRefusesSymlinkedObjects(t *testing.T) {
	c := stageCatalog(t)
	userID := stageUser(t, c)
	path := stageWrite(t, t.TempDir(), "in.png", stagePNG(t, 8, 8))
	s, err := StageImage(c, path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Discard()
	symlinkObjects(t, c)
	_, _, err = writes.Run(c, writes.Op{Action: "create_file", UserID: userID},
		func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			_, ch, err := s.InsertTx(tx.Tx)
			return struct{}{}, ch, err
		})
	wantUnsafeObjects(t, err)
}
