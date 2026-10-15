package audit

import "database/sql"

const sqlInsertScope = `INSERT OR IGNORE INTO audit_transaction_scopes (audit_transaction_id, scope_type, scope_id)
	VALUES (?, ?, ?)`

const sqlLatestByScope = `SELECT sc.scope_id, MAX(t.revision), t.created_at
	FROM audit_transaction_scopes sc
	JOIN audit_transactions t ON t.id = sc.audit_transaction_id
	WHERE sc.scope_type = ?
	GROUP BY sc.scope_id`

// Querier is the read side of a database handle.
type Querier interface {
	Query(query string, args ...any) (*sql.Rows, error)
}

// Latest is a scope's newest revision and when it was recorded (RFC3339 UTC).
// The zero value means no revision.
type Latest struct {
	Revision int64
	At       string
}

// LatestByScope returns the newest revision of every scope id of one type,
// keyed by string(scope_id).
func LatestByScope(q Querier, scopeType string) (map[string]Latest, error) {
	if q == nil || scopeType == "" {
		return nil, ErrInvalid
	}
	rows, err := q.Query(sqlLatestByScope, scopeType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]Latest{}
	for rows.Next() {
		var id []byte
		var l Latest
		if err := rows.Scan(&id, &l.Revision, &l.At); err != nil {
			return nil, err
		}
		out[string(id)] = l
	}
	return out, rows.Err()
}
