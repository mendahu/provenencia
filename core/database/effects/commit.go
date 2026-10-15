package effects

// Commit is a write that has landed: its revision, when it was recorded
// (RFC3339 UTC, the audit_transactions.created_at of that revision), and what
// it touched. It lives here because the display cache must accept it and
// core/database, which imports the cache, hands it to listeners.
type Commit struct {
	Revision int64
	At       string
	Effects  Set
}
