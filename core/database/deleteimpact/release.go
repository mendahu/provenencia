package deleteimpact

import (
	"database/sql"
	"fmt"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/catalogmodel"
	"github.com/mendahu/provenencia/core/database/files"
)

// OwnedSnapshot is the owned-outbound child ids captured while the parent still
// exists, read from the declared catalog model. ReleaseSnapshot walks them after
// the parent DELETE (looking up artifacts.file_id after the erase finds nothing).
type OwnedSnapshot struct {
	root *rowSnap
}

// rowSnap is one row's owned children: each declared owned FK that held a
// value, with the child row's own snapshot when the child is a row, not a File.
type rowSnap struct {
	items []snapItem
}

type snapItem struct {
	edge    catalogmodel.FK
	childID []byte
	sub     *rowSnap // nil for a File child
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
	root, err := snapshotOwnedRow(tx, productionPlan(), spec.Name, id)
	if err != nil {
		return OwnedSnapshot{}, err
	}
	return OwnedSnapshot{root: root}, nil
}

// ReleaseSnapshot deletes owned-outbound children from a pre-DELETE snapshot
// and schedules the purge of the Files it released, after the write commits.
func ReleaseSnapshot(tx *database.Tx, snap OwnedSnapshot) error {
	if tx == nil {
		return ErrInvalid
	}
	objs, err := releaseSnapshot(tx.Tx, snap)
	if err != nil {
		return err
	}
	files.PurgeAfterCommit(tx, objs)
	return nil
}

// releaseSnapshot deletes the snapshot's children and returns the objects of
// the Files it released, in release order.
func releaseSnapshot(tx *sql.Tx, snap OwnedSnapshot) ([]files.Object, error) {
	if tx == nil {
		return nil, ErrInvalid
	}
	return releaseOwnedSnap(tx, productionPlan(), snap.root)
}

// ReleaseOwned snapshots then releases, and schedules the purge. Call it only
// while the parent row still exists; after DELETE use a SnapshotOwned taken
// beforehand.
func ReleaseOwned(tx *database.Tx, kind catalogmodel.Kind, id []byte) error {
	if tx == nil {
		return ErrInvalid
	}
	objs, err := releaseOwned(tx.Tx, kind, id)
	if err != nil {
		return err
	}
	files.PurgeAfterCommit(tx, objs)
	return nil
}

func releaseOwned(tx *sql.Tx, kind catalogmodel.Kind, id []byte) ([]files.Object, error) {
	snap, err := SnapshotOwned(tx, kind, id)
	if err != nil {
		return nil, err
	}
	return releaseSnapshot(tx, snap)
}

// ReleaseReplaced releases the old child of an owned FK after an update has
// pointed the parent at a new one. column is "table.column" of a declared
// owned-outbound FK. The child is a row (deleted by its PK, its own children
// released after it) or a File (deleted when nothing else points at it). Call
// it after the parent row already holds the new value. It schedules the purge
// of any File it released. An empty oldID is a no-op; a column that is not an
// owned FK is ErrInvalid.
func ReleaseReplaced(tx *database.Tx, column string, oldID []byte) error {
	if tx == nil {
		return ErrInvalid
	}
	objs, err := releaseReplaced(tx.Tx, productionPlan(), column, oldID)
	if err != nil {
		return err
	}
	files.PurgeAfterCommit(tx, objs)
	return nil
}

func releaseReplaced(tx *sql.Tx, p *ownedPlan, column string, oldID []byte) ([]files.Object, error) {
	if tx == nil {
		return nil, ErrInvalid
	}
	table, col, ok := splitColumn(column)
	if !ok {
		return nil, ErrInvalid
	}
	var edge *catalogmodel.FK
	for i := range p.edges[table] {
		if p.edges[table][i].Column == col {
			edge = &p.edges[table][i]
			break
		}
	}
	if edge == nil {
		return nil, ErrInvalid
	}
	if len(oldID) == 0 {
		return nil, nil
	}
	if len(oldID) != 16 {
		return nil, ErrInvalid
	}
	item, err := snapshotEdgeChild(tx, p, *edge, oldID)
	if err != nil || item == nil {
		return nil, err
	}
	return releaseSnapItem(tx, p, *item)
}

// snapshotOwnedRow reads the owned columns of the row in table whose key is id.
// It deletes nothing; the caller deletes the row, then calls releaseOwnedSnap.
// A row that is gone, and a NULL column, are skipped.
func snapshotOwnedRow(tx *sql.Tx, p *ownedPlan, table string, id []byte) (*rowSnap, error) {
	snap := &rowSnap{}
	pk := p.pks[table]
	if len(p.edges[table]) > 0 && pk == "" {
		return nil, fmt.Errorf("deleteimpact: %s owns children but declares no PK", table)
	}
	for _, edge := range p.edges[table] {
		var childID []byte
		err := tx.QueryRow(
			fmt.Sprintf(`SELECT %s FROM %s WHERE %s = ?`, edge.Column, table, pk), id,
		).Scan(&childID)
		if err == sql.ErrNoRows {
			return snap, nil
		}
		if err != nil {
			return nil, err
		}
		if len(childID) != 16 {
			continue
		}
		item, err := snapshotEdgeChild(tx, p, edge, append([]byte(nil), childID...))
		if err != nil {
			return nil, err
		}
		if item != nil {
			snap.items = append(snap.items, *item)
		}
	}
	return snap, nil
}

// snapshotEdgeChild captures one child. A File child needs no further reading;
// a row child is snapshotted for its own owned edges while it still exists.
func snapshotEdgeChild(tx *sql.Tx, p *ownedPlan, edge catalogmodel.FK, childID []byte) (*snapItem, error) {
	item := &snapItem{edge: edge, childID: childID}
	if edge.To == "files" {
		return item, nil
	}
	if p.pks[edge.To] == "" {
		return nil, fmt.Errorf("deleteimpact: owned child %s declares no PK", edge.To)
	}
	sub, err := snapshotOwnedRow(tx, p, edge.To, childID)
	if err != nil {
		return nil, err
	}
	item.sub = sub
	return item, nil
}

// releaseOwnedSnap releases every child a snapshot captured, in edge order,
// after the parent row is gone.
func releaseOwnedSnap(tx *sql.Tx, p *ownedPlan, snap *rowSnap) ([]files.Object, error) {
	if snap == nil {
		return nil, nil
	}
	var out []files.Object
	for _, item := range snap.items {
		objs, err := releaseSnapItem(tx, p, item)
		if err != nil {
			return nil, err
		}
		out = append(out, objs...)
	}
	return out, nil
}

func releaseSnapItem(tx *sql.Tx, p *ownedPlan, item snapItem) ([]files.Object, error) {
	if item.edge.To == "files" {
		return releaseFile(tx, p, item.childID)
	}
	if _, err := tx.Exec(
		fmt.Sprintf(`DELETE FROM %s WHERE %s = ?`, item.edge.To, p.pks[item.edge.To]), item.childID,
	); err != nil {
		return nil, err
	}
	return releaseOwnedSnap(tx, p, item.sub)
}

func splitColumn(column string) (table, col string, ok bool) {
	for i := 0; i < len(column); i++ {
		if column[i] == '.' {
			return column[:i], column[i+1:], column[:i] != "" && column[i+1:] != ""
		}
	}
	return "", "", false
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
