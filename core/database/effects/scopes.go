package effects

// ScopeSource is the scope_type for work under a Source: its own row and
// everything hanging off it (notes, metadata, credibility, artifacts, files,
// subjects, citations, observations and their notes and values).
const ScopeSource = "source"

// ScopeCanonicalEntity is the scope_type for work on a Person, Event, or Place
// handle: its row, its claims and pins, and evidence on a member Subject.
const ScopeCanonicalEntity = "canonical_entity"

// Scope names an audit scope type and says which ids of a resolved Set belong
// to it. audit.Record stores one audit_transaction_scopes row per id, and the
// display cache keeps each type's last activity current from the same ids, so
// a scope type is declared here once.
type Scope struct {
	Type string
	IDs  func(Set) [][]byte
}

// Scopes lists every audit scope type, in storage order.
var Scopes = []Scope{
	{ScopeSource, func(s Set) [][]byte { return s.SourceScopes }},
	{ScopeCanonicalEntity, func(s Set) [][]byte { return s.HandleScopes }},
}
