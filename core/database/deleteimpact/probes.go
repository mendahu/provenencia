package deleteimpact

import (
	"database/sql"
	"fmt"

	"github.com/mendahu/provenencia/core/database/catalogmodel"
	"github.com/mendahu/provenencia/core/database/sqlbatch"
)

func fkEdge(parent catalogmodel.Kind, via string, child catalogmodel.Kind, table, fkCol, refCol string) inboundEdge {
	countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE %s = ?`, table, fkCol)
	listSQL := fmt.Sprintf(
		`SELECT id, COALESCE(%s, '') FROM %s WHERE %s = ? ORDER BY %s COLLATE NOCASE LIMIT ?`,
		refCol, table, fkCol, refCol,
	)
	return inboundEdge{
		Parent: parent,
		Via:    via,
		Child:  child,
		Bucket: catalogmodel.BucketResource,
		Count:  countSQLFn(countSQL),
		Totals: groupedTotalsFn(table, fkCol),
		List:   listSQLFn(listSQL),
	}
}

func joinEdge(parent catalogmodel.Kind, via string, child catalogmodel.Kind, countSQL, listSQL string) inboundEdge {
	return inboundEdge{
		Parent: parent,
		Via:    via,
		Child:  child,
		Bucket: catalogmodel.BucketResource,
		Count:  countSQLFn(countSQL),
		List:   listSQLFn(listSQL),
	}
}

// withTotals gives a join edge whose count is COUNT(*) of table rows with col
// equal to the parent its grouped form.
func withTotals(e inboundEdge, table, col string) inboundEdge {
	e.Totals = groupedTotalsFn(table, col)
	return e
}

// groupedTotalsFn counts the rows of table per value of col among a batch of
// parent ids: the grouped form of `SELECT COUNT(*) FROM table WHERE col = ?`.
func groupedTotalsFn(table, col string) func(*sql.Tx, [][]byte) (map[string]int, error) {
	head := fmt.Sprintf(`SELECT %s, COUNT(*) FROM %s WHERE %s IN (`, col, table, col)
	return func(tx *sql.Tx, ids [][]byte) (map[string]int, error) {
		out := map[string]int{}
		err := sqlbatch.Query(tx, head, fmt.Sprintf(`) GROUP BY %s`, col), nil, ids, func(rows *sql.Rows) error {
			var id []byte
			var n int
			if err := rows.Scan(&id, &n); err != nil {
				return err
			}
			out[string(id)] += n
			return nil
		})
		return out, err
	}
}

func countSQLFn(q string) func(*sql.Tx, []byte) (int, error) {
	return func(tx *sql.Tx, parentID []byte) (int, error) {
		var n int
		if err := tx.QueryRow(q, parentID).Scan(&n); err != nil {
			return 0, err
		}
		return n, nil
	}
}

func listSQLFn(q string) func(*sql.Tx, []byte, int) ([]probeRow, error) {
	return func(tx *sql.Tx, parentID []byte, limit int) ([]probeRow, error) {
		rows, err := tx.Query(q, parentID, limit)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var out []probeRow
		seen := map[string]bool{}
		for rows.Next() {
			var r probeRow
			if err := rows.Scan(&r.ID, &r.Ref); err != nil {
				return nil, err
			}
			key := string(r.ID)
			if seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, r)
		}
		return out, rows.Err()
	}
}
