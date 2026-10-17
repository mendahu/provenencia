package deleteimpact_test

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/catalogmodel"
	"github.com/mendahu/provenencia/core/database/deleteimpact"
	"github.com/mendahu/provenencia/core/database/files"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/objectstore"
	"github.com/mendahu/provenencia/core/writes"
)

// A release schedules the purge of the Files it freed; the caller passes no
// object list, and a rolled-back write removes nothing.
func TestReleaseSchedulesPurge(t *testing.T) {
	boom := errors.New("boom")
	// commitChange makes a write that changes no audited table commit.
	commitChange := []rowchange.Change{{
		EntityType: "subject_position",
		EntityID:   make([]byte, 16),
		Action:     rowchange.ActionCreate,
	}}

	tests := []struct {
		name     string
		fail     bool
		release  func(t *testing.T, tx *database.Tx, artID, fileID []byte)
		wantGone bool
	}{
		{
			name: "ReleaseReplaced after the commit",
			release: func(t *testing.T, tx *database.Tx, artID, fileID []byte) {
				if _, err := tx.Exec(`UPDATE artifacts SET file_id = NULL WHERE id = ?`, artID); err != nil {
					t.Fatal(err)
				}
				if err := deleteimpact.ReleaseReplaced(tx, "artifacts.file_id", fileID); err != nil {
					t.Fatal(err)
				}
			},
			wantGone: true,
		},
		{
			name: "ReleaseReplaced in a failed write",
			fail: true,
			release: func(t *testing.T, tx *database.Tx, artID, fileID []byte) {
				if _, err := tx.Exec(`UPDATE artifacts SET file_id = NULL WHERE id = ?`, artID); err != nil {
					t.Fatal(err)
				}
				if err := deleteimpact.ReleaseReplaced(tx, "artifacts.file_id", fileID); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "SnapshotOwned then ReleaseSnapshot",
			release: func(t *testing.T, tx *database.Tx, artID, fileID []byte) {
				snap, err := deleteimpact.SnapshotOwned(tx.Tx, catalogmodel.KindArtifact, artID)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := tx.Exec(`DELETE FROM artifacts WHERE id = ?`, artID); err != nil {
					t.Fatal(err)
				}
				if err := deleteimpact.ReleaseSnapshot(tx, snap); err != nil {
					t.Fatal(err)
				}
			},
			wantGone: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, userID := testCatalog(t)
			typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
				Key: "book", Origin: sourcetypes.OriginProvenencia, Label: "Book",
			})
			if err != nil {
				t.Fatal(err)
			}
			src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID},
				func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
					return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Deed"})
				})
			if err != nil {
				t.Fatal(err)
			}

			data := []byte("object bytes")
			sum := sha256.Sum256(data)
			checksum := hex.EncodeToString(sum[:])
			rel, err := files.StorageRelPath(checksum, "image/png")
			if err != nil {
				t.Fatal(err)
			}
			objPath := filepath.Join(c.Dir(), filepath.FromSlash(rel))
			if _, err := objectstore.WriteIfAbsent(objPath, checksum, data); err != nil {
				t.Fatal(err)
			}
			fileID, err := files.NewID()
			if err != nil {
				t.Fatal(err)
			}
			db, err := c.DB()
			if err != nil {
				t.Fatal(err)
			}
			sqlTx, err := db.Begin()
			if err != nil {
				t.Fatal(err)
			}
			if err := files.Insert(sqlTx, files.File{ID: fileID, ChecksumSHA256: checksum, MediaType: "image/png", ByteSize: int64(len(data))}); err != nil {
				t.Fatal(err)
			}
			if err := sqlTx.Commit(); err != nil {
				t.Fatal(err)
			}
			art, err := runArtifactCreate(c, userID, artifacts.CreateInput{SourceID: src.ID, FileID: fileID, Label: "Page"})
			if err != nil {
				t.Fatal(err)
			}

			_, err = writes.Call(c, writes.Op{Action: "t", UserID: userID},
				func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
					tt.release(t, tx, art.ID, fileID)
					if tt.fail {
						return struct{}{}, nil, boom
					}
					return struct{}{}, commitChange, nil
				})
			if tt.fail != errors.Is(err, boom) || (!tt.fail && err != nil) {
				t.Fatalf("write err = %v", err)
			}

			_, statErr := os.Stat(objPath)
			objectGone := errors.Is(statErr, os.ErrNotExist)
			var rows int
			if err := db.QueryRow(`SELECT COUNT(*) FROM files WHERE id = ?`, fileID).Scan(&rows); err != nil {
				t.Fatal(err)
			}
			if tt.wantGone && (rows != 0 || !objectGone) {
				t.Fatalf("rows=%d objectGone=%v, want both gone", rows, objectGone)
			}
			if !tt.wantGone && (rows != 1 || objectGone) {
				t.Fatalf("rows=%d objectGone=%v, want both kept", rows, objectGone)
			}
		})
	}
}
