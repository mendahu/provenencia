package files

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"

	"github.com/mendahu/provenencia/core/database"
)

// Object is the checksum and media type of a files row: enough for
// StorageRelPath, and what a release reports so the bytes can be removed after
// COMMIT.
type Object struct {
	ChecksumSHA256 string
	MediaType      string
}

// PurgeUnreferenced removes objects/ bytes for released Files whose catalog row
// is gone. Best-effort after COMMIT: a crash leaves orphan bytes, not a live
// pointer at a missing object. A checksum a files row still carries stays on
// disk, and a checksum listed twice is removed once.
//
// Object bytes are removed only here (S10-53 moves ingest.File's rollback and
// StagedImage.Discard onto this function).
func PurgeUnreferenced(c *database.Catalog, objs []Object) {
	for _, obj := range objs {
		if _, err := LookupByChecksum(c, obj.ChecksumSHA256); err == nil {
			continue
		} else if !errors.Is(err, sql.ErrNoRows) {
			continue
		}
		rel, err := StorageRelPath(obj.ChecksumSHA256, obj.MediaType)
		if err != nil {
			continue
		}
		_ = os.Remove(filepath.Join(c.Dir(), filepath.FromSlash(rel)))
	}
}

// PurgeAfterCommit schedules PurgeUnreferenced for objs once tx commits. A
// rollback drops it. An empty list schedules nothing; a non-empty list on a Tx
// that did not come from BeginWrite panics.
func PurgeAfterCommit(tx *database.Tx, objs []Object) {
	if len(objs) == 0 {
		return
	}
	c := tx.Catalog()
	if c == nil {
		panic("files: purge needs the write's catalog")
	}
	tx.AfterCommit(func() { PurgeUnreferenced(c, objs) })
}
