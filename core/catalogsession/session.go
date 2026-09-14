// Package catalogsession holds an exclusive project catalog for the process
// lifetime of a workspace and serializes operations on it.
//
// Researcher FFI and onboarding paths must use Do (or Close) rather than
// open-per-call database.Open / onboarding.OpenCatalog. A leftover direct open
// while a session is held surfaces as catalog.already_open — a bypass bug, not
// normal overlap.
package catalogsession

import (
	"path/filepath"
	"strings"
	"sync"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/project"
	"github.com/mendahu/provenencia/core/database/searchindex"
	"github.com/mendahu/provenencia/core/database/users"
)

type session struct {
	mu     sync.Mutex
	cat    *database.Catalog
	closed bool
}

type registry struct {
	mu       sync.Mutex
	sessions map[string]*session
}

var reg = registry{sessions: make(map[string]*session)}

func key(projectDir string) (string, error) {
	projectDir = strings.TrimSpace(projectDir)
	if projectDir == "" {
		return "", database.ErrNotAProject
	}
	return filepath.Clean(projectDir), nil
}

func openResearcher(projectDir string) (*database.Catalog, error) {
	c, err := database.Open(projectDir)
	if err != nil {
		return nil, err
	}
	if err := users.EnsureRefs(c); err != nil {
		_ = c.Close()
		return nil, err
	}
	if err := project.EnsureUUID(c); err != nil {
		_ = c.Close()
		return nil, err
	}
	if err := searchindex.EnsureCatalog(c); err != nil {
		_ = c.Close()
		return nil, err
	}
	return c, nil
}

func closeSession(s *session) error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	if s.cat == nil {
		return nil
	}
	err := s.cat.Close()
	s.cat = nil
	return err
}

// getOrOpen returns the held session for projectDir, opening if needed.
// Opening a different project closes every other held session first.
// Caller must not hold any session.mu. Registry lock covers open so only one
// goroutine calls database.Open for a given miss.
func getOrOpen(projectDir string) (*session, error) {
	reg.mu.Lock()
	defer reg.mu.Unlock()

	if s, ok := reg.sessions[projectDir]; ok {
		return s, nil
	}

	for dir, s := range reg.sessions {
		delete(reg.sessions, dir)
		if err := closeSession(s); err != nil {
			return nil, err
		}
	}

	c, err := openResearcher(projectDir)
	if err != nil {
		return nil, err
	}
	s := &session{cat: c}
	reg.sessions[projectDir] = s
	return s, nil
}

// Do runs fn on the held catalog for projectDir, opening once if needed.
// Operations on the same project serialize; overlapping callers wait.
// Opening a different project closes any other held session first.
func Do(projectDir string, fn func(*database.Catalog) error) error {
	dir, err := key(projectDir)
	if err != nil {
		return err
	}
	if fn == nil {
		return nil
	}
	for {
		s, err := getOrOpen(dir)
		if err != nil {
			return err
		}
		s.mu.Lock()
		if s.closed || s.cat == nil {
			s.mu.Unlock()
			continue
		}
		err = fn(s.cat)
		s.mu.Unlock()
		return err
	}
}

// Close releases the held catalog for projectDir, if any.
func Close(projectDir string) error {
	dir, err := key(projectDir)
	if err != nil {
		return err
	}
	reg.mu.Lock()
	s, ok := reg.sessions[dir]
	if ok {
		delete(reg.sessions, dir)
	}
	reg.mu.Unlock()
	return closeSession(s)
}

// CloseAll releases every held catalog session.
func CloseAll() error {
	reg.mu.Lock()
	stale := make([]*session, 0, len(reg.sessions))
	for dir, s := range reg.sessions {
		delete(reg.sessions, dir)
		stale = append(stale, s)
	}
	reg.mu.Unlock()
	var first error
	for _, s := range stale {
		if err := closeSession(s); err != nil && first == nil {
			first = err
		}
	}
	return first
}
