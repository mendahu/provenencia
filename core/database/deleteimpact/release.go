package deleteimpact

import (
	"database/sql"
	"fmt"
)

// ReleaseOwned deletes owned-outbound children registered on kind after the
// parent row is gone (or using the parent's leftover column values).
func ReleaseOwned(tx *sql.Tx, kind Kind, id []byte) error {
	if tx == nil || len(id) != 16 {
		return ErrInvalid
	}
	for _, rel := range ownedFor(kind) {
		if err := runOwned(tx, rel, id); err != nil {
			return err
		}
	}
	return nil
}

func runOwned(tx *sql.Tx, rel ownedRelease, parentID []byte) error {
	switch rel.Child {
	case "date_values", "name_values":
		childID, err := lookupParentColumn(tx, rel.Column, parentID)
		if err != nil || len(childID) == 0 {
			return err
		}
		return deleteValueRow(tx, rel.Child, childID)
	case "files":
		fileID, err := lookupParentColumn(tx, rel.Column, parentID)
		if err != nil || len(fileID) == 0 {
			return err
		}
		return releaseFileIfUnused(tx, fileID)
	case "file_derivatives":
		return releaseFileDerivatives(tx, parentID)
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
