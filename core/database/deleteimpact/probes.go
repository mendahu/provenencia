package deleteimpact

import (
	"database/sql"
	"fmt"
)

func fkEdge(parent Kind, via string, child Kind, table, fkCol, refCol string) inboundEdge {
	countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE %s = ?`, table, fkCol)
	listSQL := fmt.Sprintf(
		`SELECT id, COALESCE(%s, '') FROM %s WHERE %s = ? ORDER BY %s COLLATE NOCASE LIMIT ?`,
		refCol, table, fkCol, refCol,
	)
	return inboundEdge{
		Parent: parent,
		Via:    via,
		Child:  child,
		Bucket: BucketResource,
		Count:  countSQLFn(countSQL),
		List:   listSQLFn(listSQL),
	}
}

func joinEdge(parent Kind, via string, child Kind, countSQL, listSQL string) inboundEdge {
	return inboundEdge{
		Parent: parent,
		Via:    via,
		Child:  child,
		Bucket: BucketResource,
		Count:  countSQLFn(countSQL),
		List:   listSQLFn(listSQL),
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

func reservedStub(parent Kind, via string, child Kind) inboundEdge {
	return inboundEdge{
		Parent: parent,
		Via:    via,
		Child:  child,
		Bucket: BucketResource,
		Count:  func(*sql.Tx, []byte) (int, error) { return 0, nil },
		List:   func(*sql.Tx, []byte, int) ([]probeRow, error) { return nil, nil },
	}
}
