package audit

// ScopeSource is the scope_type for work under a Source: its own row and
// everything hanging off it (notes, metadata, credibility, artifacts, files,
// subjects, citations, observations and their notes and values).
const ScopeSource = "source"

const sqlInsertScope = `INSERT OR IGNORE INTO audit_transaction_scopes (audit_transaction_id, scope_type, scope_id)
	VALUES (?, ?, ?)`

// Scope is one aggregate a revision touched. Record fills these from
// effects.Sources.
type Scope struct {
	Type string
	ID   []byte
}
