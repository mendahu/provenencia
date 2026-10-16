package deleteimpact

import (
	"database/sql"
	"fmt"

	"github.com/mendahu/provenencia/core/database/catalogmodel"
	"github.com/mendahu/provenencia/core/database/files"
)

// OwnedSnapshot is owned-outbound child IDs captured while the parent still
// exists. ReleaseSnapshot walks those IDs after DELETE (looking up
// artifacts.file_id after erase is a no-op).
type OwnedSnapshot struct {
	kind catalogmodel.Kind
	cols map[string][]byte
}

// SnapshotOwned reads owned-outbound columns on the live parent row.
// Missing parent returns ErrInvalid — after DELETE use a snapshot taken first.
func SnapshotOwned(tx *sql.Tx, kind catalogmodel.Kind, id []byte) (OwnedSnapshot, error) {
	if tx == nil || len(id) != 16 {
		return OwnedSnapshot{}, ErrInvalid
	}
	spec, ok := tableByKind(kind)
	if !ok {
		return OwnedSnapshot{}, ErrInvalid
	}
	var one int
	err := tx.QueryRow(spec.Exists, id).Scan(&one)
	if err == sql.ErrNoRows {
		return OwnedSnapshot{}, ErrInvalid
	}
	if err != nil {
		return OwnedSnapshot{}, err
	}
	snap := OwnedSnapshot{kind: kind, cols: map[string][]byte{}}
	for _, rel := range ownedFor(kind) {
		switch rel.Child {
		case "files", "date_values", "name_values":
			childID, err := lookupParentColumn(tx, rel.Column, id)
			if err != nil {
				return OwnedSnapshot{}, err
			}
			if len(childID) == 16 {
				snap.cols[rel.Column] = childID
			}
		}
	}
	return snap, nil
}

// ReleaseSnapshot deletes owned-outbound children from a pre-DELETE snapshot
// and returns the objects of the Files it released, in release order.
func ReleaseSnapshot(tx *sql.Tx, snap OwnedSnapshot) ([]files.Object, error) {
	if tx == nil {
		return nil, ErrInvalid
	}
	var out []files.Object
	for _, rel := range ownedFor(snap.kind) {
		objs, err := releaseFromSnapshot(tx, rel, snap)
		if err != nil {
			return nil, err
		}
		out = append(out, objs...)
	}
	return out, nil
}

// ReleaseOwned snapshots then releases. Call it only while the parent row
// still exists; after DELETE use a SnapshotOwned taken beforehand.
func ReleaseOwned(tx *sql.Tx, kind catalogmodel.Kind, id []byte) ([]files.Object, error) {
	snap, err := SnapshotOwned(tx, kind, id)
	if err != nil {
		return nil, err
	}
	return ReleaseSnapshot(tx, snap)
}

func releaseFromSnapshot(tx *sql.Tx, rel ownedRelease, snap OwnedSnapshot) ([]files.Object, error) {
	switch rel.Child {
	case "date_values", "name_values":
		childID := snap.cols[rel.Column]
		if len(childID) != 16 {
			return nil, nil
		}
		return nil, deleteValueRow(tx, rel.Child, childID)
	case "files":
		fileID := snap.cols[rel.Column]
		if len(fileID) != 16 {
			return nil, nil
		}
		return releaseFile(tx, productionPlan(), fileID)
	default:
		return nil, nil
	}
}

func lookupParentColumn(tx *sql.Tx, column string, parentID []byte) ([]byte, error) {
	table, col, ok := splitColumn(column)
	if !ok {
		return nil, ErrInvalid
	}
	var id []byte
	err := tx.QueryRow(fmt.Sprintf(`SELECT %s FROM %s WHERE id = ?`, col, table), parentID).Scan(&id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return append([]byte(nil), id...), nil
}

func splitColumn(column string) (table, col string, ok bool) {
	for i := 0; i < len(column); i++ {
		if column[i] == '.' {
			return column[:i], column[i+1:], column[:i] != "" && column[i+1:] != ""
		}
	}
	return "", "", false
}

func deleteValueRow(tx *sql.Tx, table string, id []byte) error {
	if len(id) != 16 {
		return nil
	}
	_, err := tx.Exec(fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, table), id)
	return err
}

// CountFilePointers counts the rows that still point at a File through any
// declared foreign key into files, except the File's own thumbnail links
// (file_derivatives.source_file_id). A File with a nonzero count is in use.
func CountFilePointers(tx *sql.Tx, fileID []byte) (int, error) {
	return countFilePointers(tx, productionPlan(), fileID)
}

func countFilePointers(tx *sql.Tx, p *ownedPlan, fileID []byte) (int, error) {
	if tx == nil || len(fileID) != 16 {
		return 0, ErrInvalid
	}
	total := 0
	for _, fk := range p.pointers {
		var n int
		q := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE %s = ?`, fk.From, fk.Column)
		if err := tx.QueryRow(q, fileID).Scan(&n); err != nil {
			return 0, err
		}
		total += n
	}
	return total, nil
}

// releaseFile deletes a File nothing points at, along with its own derivative
// links, then releases each File it derived the same way. It returns the
// objects of every File it deleted, the File's own first. A File another
// source's link still names keeps a nonzero pointer count and stays.
func releaseFile(tx *sql.Tx, p *ownedPlan, fileID []byte) ([]files.Object, error) {
	n, err := countFilePointers(tx, p, fileID)
	if err != nil || n > 0 {
		return nil, err
	}
	derived, err := listDerivedFileIDs(tx, fileID)
	if err != nil {
		return nil, err
	}
	obj, found, err := lookupFileObject(tx, fileID)
	if err != nil {
		return nil, err
	}
	if _, err := tx.Exec(`DELETE FROM file_derivatives WHERE source_file_id = ?`, fileID); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(`DELETE FROM files WHERE id = ?`, fileID); err != nil {
		return nil, err
	}
	var out []files.Object
	if found {
		out = append(out, obj)
	}
	for _, derivedID := range derived {
		objs, err := releaseFile(tx, p, derivedID)
		if err != nil {
			return nil, err
		}
		out = append(out, objs...)
	}
	return out, nil
}

func lookupFileObject(tx *sql.Tx, fileID []byte) (files.Object, bool, error) {
	var checksum, media string
	err := tx.QueryRow(
		`SELECT checksum_sha256, COALESCE(media_type, '') FROM files WHERE id = ?`,
		fileID,
	).Scan(&checksum, &media)
	if err == sql.ErrNoRows {
		return files.Object{}, false, nil
	}
	if err != nil {
		return files.Object{}, false, err
	}
	return files.Object{ChecksumSHA256: checksum, MediaType: media}, true, nil
}

func listDerivedFileIDs(tx *sql.Tx, fileID []byte) ([][]byte, error) {
	rows, err := tx.Query(
		`SELECT derived_file_id FROM file_derivatives WHERE source_file_id = ?`,
		fileID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out [][]byte
	for rows.Next() {
		var id []byte
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		if len(id) == 16 {
			out = append(out, append([]byte(nil), id...))
		}
	}
	return out, rows.Err()
}
