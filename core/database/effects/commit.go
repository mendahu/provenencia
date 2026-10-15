package effects

// Activity is a scope's newest audit revision and when it was recorded
// (RFC3339 UTC, the audit_transactions.created_at of that revision). The zero
// value means no revision.
type Activity struct {
	Revision int64
	At       string
}

// Commit is a write that has landed: its Activity (revision and time) and what
// it touched. It lives here because the display cache must accept it and
// core/database, which imports the cache, hands it to listeners.
type Commit struct {
	Activity
	Effects Set
}
