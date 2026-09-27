package deleteimpact

import (
	"database/sql"
	"fmt"
)

// OwnedSnapshot is owned-outbound child IDs captured while the parent still
// exists. ReleaseSnapshot walks those IDs after DELETE (looking up
// artifacts.file_id after erase is a no-op).
type OwnedSnapshot struct {
	kind Kind
	cols map[string][]byte
}

// SnapshotOwned reads owned-outbound columns on the live parent row.
func SnapshotOwned(tx *sql.Tx, kind Kind, id []byte) (OwnedSnapshot, error) {
	if tx == nil || len(id) != 16 {
		return OwnedSnapshot{}, ErrInvalid
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

// ReleaseSnapshot deletes owned-outbound children from a pre-DELETE snapshot.
func ReleaseSnapshot(tx *sql.Tx, snap OwnedSnapshot) error {
	if tx == nil {
		return ErrInvalid
	}
	for _, rel := range ownedFor(snap.kind) {
		if err := releaseFromSnapshot(tx, rel, snap); err != nil {
			return err
		}
	}
	return nil
}

// ReleaseOwned snapshots then releases. Call it only while the parent row
// still exists; after DELETE use a SnapshotOwned taken beforehand.
func ReleaseOwned(tx *sql.Tx, kind Kind, id []byte) error {
	snap, err := SnapshotOwned(tx, kind, id)
	if err != nil {
		return err
	}
	return ReleaseSnapshot(tx, snap)
}

func releaseFromSnapshot(tx *sql.Tx, rel ownedRelease, snap OwnedSnapshot) error {
	switch rel.Child {
	case "date_values", "name_values":
		childID := snap.cols[rel.Column]
		if len(childID) != 16 {
			return nil
		}
		return deleteValueRow(tx, rel.Child, childID)
	case "files":
		fileID := snap.cols[rel.Column]
		if len(fileID) != 16 {
			return nil
		}
		return releaseFileIfUnused(tx, fileID)
	case "file_derivatives":
		return nil
	default:
		return nil
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

// CountFilePointers is the ifUnused walker: remaining artifacts.file_id rows
// that still name this file. Derivatives are released after the last artifact.
func CountFilePointers(tx *sql.Tx, fileID []byte) (int, error) {
	if tx == nil || len(fileID) != 16 {
		return 0, ErrInvalid
	}
	var n int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM artifacts WHERE file_id = ?`, fileID).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func releaseFileIfUnused(tx *sql.Tx, fileID []byte) error {
	n, err := CountFilePointers(tx, fileID)
	if err != nil || n > 0 {
		return err
	}
	derived, err := listDerivedFileIDs(tx, fileID)
	if err != nil {
		return err
	}
	if err := releaseFileDerivatives(tx, fileID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM files WHERE id = ?`, fileID); err != nil {
		return err
	}
	for _, derivedID := range derived {
		if err := releaseFileIfUnused(tx, derivedID); err != nil {
			return err
		}
	}
	return nil
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

func releaseFileDerivatives(tx *sql.Tx, fileID []byte) error {
	_, err := tx.Exec(
		`DELETE FROM file_derivatives WHERE source_file_id = ? OR derived_file_id = ?`,
		fileID, fileID,
	)
	return err
}
