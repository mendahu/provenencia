package audit

import (
	"bytes"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

// ScopeSource is the scope_type for work under a Source: its own row and
// everything hanging off it (notes, metadata, credibility, artifacts, files,
// subjects, citations, observations and their notes and values).
const ScopeSource = "source"

const sqlInsertScope = `INSERT OR IGNORE INTO audit_transaction_scopes (audit_transaction_id, scope_type, scope_id)
	VALUES (?, ?, ?)`

// Scope is one aggregate a revision touched.
type Scope struct {
	Type string
	ID   []byte
}

// ghostMap holds the old fields of rows deleted in this revision, keyed by
// entity type + id. Deletes run before Record, so a released note's parent
// citation may already be gone; its own delete change still names the parent.
type ghostMap map[string]map[string]any

func ghostKey(entityType string, id []byte) string { return entityType + ":" + string(id) }

func (g ghostMap) id(entityType string, id []byte, field string) []byte {
	return uuidValue(g[ghostKey(entityType, id)][field])
}

// resolver maps one change to the scopes it belongs to. A parent that cannot
// be found is no scope, never an error: another change in the same revision
// normally covers it.
type resolver func(r scopeResolver, ch Change) ([]Scope, error)

// resolvers covers every audited entity type. Record rejects a type missing
// here, so a new audited table has to decide its scope.
var resolvers = map[string]resolver{
	"source": func(_ scopeResolver, ch Change) ([]Scope, error) {
		return sourceScopes(ch.EntityID), nil
	},
	"source_note":                   directSource("source_notes"),
	"source_metadata":               directSource("source_metadata"),
	"source_credibility_assessment": directSource("source_credibility_assessments"),
	"subject":                       directSource("subjects"),
	"artifact":                      directSource("artifacts"),
	// Layout rows are audited under field_id; source_id is always in Fields.
	"source_metadata_layout": func(_ scopeResolver, ch Change) ([]Scope, error) {
		return sourceScopes(fieldIDs(ch, "source_id")...), nil
	},
	"citation": func(r scopeResolver, ch Change) ([]Scope, error) {
		return r.via(ch, "artifact_id", r.sourceOfArtifact, r.sourceOfCitation)
	},
	"citation_note": func(r scopeResolver, ch Change) ([]Scope, error) {
		return r.via(ch, "citation_id", r.sourceOfCitation, r.parentLookup(
			"citation_note", `SELECT citation_id FROM citation_notes WHERE id = ?`, "citation_id", r.sourceOfCitation))
	},
	"observation": func(r scopeResolver, ch Change) ([]Scope, error) {
		return r.via(ch, "citation_id", r.sourceOfCitation, r.sourceOfObservation)
	},
	"observation_note": func(r scopeResolver, ch Change) ([]Scope, error) {
		return r.via(ch, "observation_id", r.sourceOfObservation, r.parentLookup(
			"observation_note", `SELECT observation_id FROM observation_notes WHERE id = ?`, "observation_id", r.sourceOfObservation))
	},
	"date_value": func(r scopeResolver, ch Change) ([]Scope, error) {
		return r.sources(
			`SELECT a.source_id FROM observations o
				JOIN citations c ON c.id = o.citation_id
				JOIN artifacts a ON a.id = c.artifact_id
				WHERE o.value_date_id = ?`,
			ch.EntityID)
	},
	"name_value": func(r scopeResolver, ch Change) ([]Scope, error) {
		return r.sources(
			`SELECT a.source_id FROM observations o
				JOIN citations c ON c.id = o.citation_id
				JOIN artifacts a ON a.id = c.artifact_id
				WHERE o.value_name_id = ?`,
			ch.EntityID)
	},
	// A file belongs to every Source with an artifact on it (none at first
	// ingest; the artifact create that attaches it carries the Source).
	"file": func(r scopeResolver, ch Change) ([]Scope, error) {
		return r.sources(`SELECT DISTINCT source_id FROM artifacts WHERE file_id = ?`, ch.EntityID)
	},

	// Vocabulary is catalog-wide, and conclusion-layer work spans Sources.
	"source_type":             noScope,
	"metadata_field":          noScope,
	"property":                noScope,
	"property_term":           noScope,
	"identity_claim":          noScope,
	"identity_claim_evidence": noScope,
	"canonical_entity":        noScope,
}

func noScope(scopeResolver, Change) ([]Scope, error) { return nil, nil }

// directSource resolves rows that carry source_id themselves.
func directSource(table string) resolver {
	return func(r scopeResolver, ch Change) ([]Scope, error) {
		if ids := fieldIDs(ch, "source_id"); len(ids) > 0 {
			return sourceScopes(ids...), nil
		}
		id, err := r.lookup(`SELECT source_id FROM `+table+` WHERE id = ?`, ch.EntityID)
		if err != nil {
			return nil, err
		}
		if id == nil {
			id = r.ghosts.id(ch.EntityType, ch.EntityID, "source_id")
		}
		return sourceScopes(id), nil
	}
}

type scopeResolver struct {
	tx     *sql.Tx
	ghosts ghostMap
}

// via resolves a change through the parent named in its fields, or failing
// that through the change's own id.
func (r scopeResolver) via(ch Change, field string, parent, self func([]byte) ([]byte, error)) ([]Scope, error) {
	ids := fieldIDs(ch, field)
	if len(ids) == 0 {
		id, err := self(ch.EntityID)
		if err != nil {
			return nil, err
		}
		return sourceScopes(id), nil
	}
	var out []Scope
	for _, pid := range ids {
		id, err := parent(pid)
		if err != nil {
			return nil, err
		}
		out = append(out, sourceScopes(id)...)
	}
	return out, nil
}

// parentLookup reads a row's parent id (live, then ghost) and resolves it.
func (r scopeResolver) parentLookup(entityType, query, field string, next func([]byte) ([]byte, error)) func([]byte) ([]byte, error) {
	return func(id []byte) ([]byte, error) {
		pid, err := r.lookup(query, id)
		if err != nil {
			return nil, err
		}
		if pid == nil {
			pid = r.ghosts.id(entityType, id, field)
		}
		if pid == nil {
			return nil, nil
		}
		return next(pid)
	}
}

func (r scopeResolver) sourceOfArtifact(id []byte) ([]byte, error) {
	src, err := r.lookup(`SELECT source_id FROM artifacts WHERE id = ?`, id)
	if err != nil || src != nil {
		return src, err
	}
	return r.ghosts.id("artifact", id, "source_id"), nil
}

func (r scopeResolver) sourceOfCitation(id []byte) ([]byte, error) {
	return r.parentLookup("citation", `SELECT artifact_id FROM citations WHERE id = ?`, "artifact_id", r.sourceOfArtifact)(id)
}

func (r scopeResolver) sourceOfObservation(id []byte) ([]byte, error) {
	return r.parentLookup("observation", `SELECT citation_id FROM observations WHERE id = ?`, "citation_id", r.sourceOfCitation)(id)
}

// lookup returns the single id column of query, or nil when there is no row.
func (r scopeResolver) lookup(query string, args ...any) ([]byte, error) {
	var id []byte
	err := r.tx.QueryRow(query, args...).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return id, nil
}

// sources returns a Source scope for each source_id row of query.
func (r scopeResolver) sources(query string, args ...any) ([]Scope, error) {
	rows, err := r.tx.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Scope
	for rows.Next() {
		var id []byte
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, sourceScopes(id)...)
	}
	return out, rows.Err()
}

// resolveScopes derives the de-duplicated scopes of a revision's changes.
func resolveScopes(tx *sql.Tx, changes []Change) ([]Scope, error) {
	ghosts := ghostMap{}
	for _, ch := range changes {
		if ch.Action != ActionDelete {
			continue
		}
		old := make(map[string]any, len(ch.Fields))
		for k, v := range ch.Fields {
			old[k] = v.Old
		}
		ghosts[ghostKey(ch.EntityType, ch.EntityID)] = old
	}
	r := scopeResolver{tx: tx, ghosts: ghosts}
	var out []Scope
	for _, ch := range changes {
		resolve, ok := resolvers[ch.EntityType]
		if !ok {
			return nil, ErrInvalid
		}
		scopes, err := resolve(r, ch)
		if err != nil {
			return nil, err
		}
		for _, s := range scopes {
			if !containsScope(out, s) {
				out = append(out, s)
			}
		}
	}
	return out, nil
}

func containsScope(list []Scope, s Scope) bool {
	for _, x := range list {
		if x.Type == s.Type && bytes.Equal(x.ID, s.ID) {
			return true
		}
	}
	return false
}

func sourceScopes(ids ...[]byte) []Scope {
	var out []Scope
	for _, id := range ids {
		if len(id) == 16 {
			out = append(out, Scope{Type: ScopeSource, ID: id})
		}
	}
	return out
}

// fieldIDs returns the distinct UUIDs in a field's new and old values, so a
// row moved between parents scopes to both.
func fieldIDs(ch Change, field string) [][]byte {
	diff, ok := ch.Fields[field]
	if !ok {
		return nil
	}
	var out [][]byte
	for _, v := range []any{diff.New, diff.Old} {
		id := uuidValue(v)
		if id == nil {
			continue
		}
		if len(out) == 1 && bytes.Equal(out[0], id) {
			continue
		}
		out = append(out, id)
	}
	return out
}

// uuidValue reads a UUID stored in changes as a dashed string or raw bytes.
func uuidValue(v any) []byte {
	switch x := v.(type) {
	case string:
		u, err := uuid.Parse(x)
		if err != nil {
			return nil
		}
		return u[:]
	case []byte:
		if len(x) == 16 {
			return x
		}
	}
	return nil
}
