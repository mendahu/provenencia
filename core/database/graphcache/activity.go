package graphcache

import (
	"fmt"

	"github.com/mendahu/provenencia/core/database/effects"
)

// Activity returns the last activity of one scope id (a Source id for
// effects.ScopeSource, a handle id for effects.ScopeCanonicalEntity). The first
// read of a scope type loads every id's activity with one grouped query; a
// commit keeps the loaded map current (OnCommit), so a list read never scans
// the audit history. A nil Graph and an id with no revision give the zero
// value. A type not in effects.Scopes is an error.
func (g *Graph) Activity(scopeType string, id []byte) (effects.Activity, error) {
	if g == nil {
		return effects.Activity{}, nil
	}
	if !knownScope(scopeType) {
		return effects.Activity{}, fmt.Errorf("graphcache: unknown scope type %q", scopeType)
	}
	m, ok := g.activity[scopeType]
	if !ok {
		var err error
		if m, err = g.loadActivity(scopeType); err != nil {
			return effects.Activity{}, err
		}
		g.activity[scopeType] = m
	}
	return m[string(id)], nil
}

func knownScope(scopeType string) bool {
	for _, sc := range effects.Scopes {
		if sc.Type == scopeType {
			return true
		}
	}
	return false
}

func (g *Graph) loadActivity(scopeType string) (map[string]effects.Activity, error) {
	rows, err := g.query(sqlActivityByScope, scopeType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]effects.Activity{}
	for rows.Next() {
		var id []byte
		var a effects.Activity
		if err := rows.Scan(&id, &a.Revision, &a.At); err != nil {
			return nil, err
		}
		out[string(id)] = a
	}
	return out, rows.Err()
}

// applyActivity moves every loaded scope map forward to the commit. A commit
// older than a held revision changes nothing.
func (g *Graph) applyActivity(c effects.Commit) {
	if c.Revision <= 0 {
		return
	}
	for _, sc := range effects.Scopes {
		m, ok := g.activity[sc.Type]
		if !ok {
			continue
		}
		for _, id := range sc.IDs(c.Effects) {
			if c.Revision > m[string(id)].Revision {
				m[string(id)] = c.Activity
			}
		}
	}
}
