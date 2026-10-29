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
	"github.com/mendahu/provenencia/core/database/autoreconciler"
	"github.com/mendahu/provenencia/core/database/evidencegraphs"
	"github.com/mendahu/provenencia/core/database/namealiases"
	"github.com/mendahu/provenencia/core/database/nameformats"
	"github.com/mendahu/provenencia/core/database/project"
	"github.com/mendahu/provenencia/core/database/promotealign"
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

// ReconcileOnOpen runs the open-time repairs of a catalog, in order: refs for
// rows a format step added without one, the project uuid, then the derived
// caches. onboarding.createCatalog, onboarding's OpenCatalog and Do all call
// it, so the next repair is one line here. It does not close c.
func ReconcileOnOpen(c *database.Catalog) error {
	if err := users.EnsureRefs(c); err != nil {
		return err
	}
	if err := evidencegraphs.EnsureRefs(c); err != nil {
		return err
	}
	if err := project.EnsureUUID(c); err != nil {
		return err
	}
	// The cache first: the search index's handle documents read it.
	if err := autoreconciler.EnsureCatalog(c); err != nil {
		return err
	}
	return searchindex.EnsureCatalog(c)
}

// warmGraph loads what Promote and the lists read on every call onto the
// catalog graph: the built alias index, then the default name format, then it
// marks the graph warm. Its error is the only one WarmOnOpen returns.
func warmGraph(c *database.Catalog) error {
	if err := namealiases.Warm(c); err != nil {
		return err
	}
	if err := nameformats.Warm(c); err != nil {
		return err
	}
	c.Graph().MarkWarm()
	return nil
}

// WarmOnOpen loads what Promote and the lists read on every call onto the
// catalog graph when a project opens: the built alias index and the default
// name format (warmGraph), then matching's shortlist kinds and index. Matching
// goes last because setting the first two drops the match index. Matching's
// error is ignored: a value that does not decode leaves the shortlist index
// unbuilt, and the next proposal builds it and reports the error. WarmOnOpen
// returns warmGraph's error only. Do runs warmGraph again after a call that
// dropped the graph or its vocabulary, not matching's warm, so the shortlist
// index waits for the next proposal. One list, so a catalog opened for a test
// warms what the app does.
func WarmOnOpen(c *database.Catalog) error {
	if err := warmGraph(c); err != nil {
		return err
	}
	_ = promotealign.Warm(c)
	return nil
}

func openResearcher(projectDir string) (*database.Catalog, error) {
	c, err := database.Open(projectDir)
	if err != nil {
		return nil, err
	}
	if err := ReconcileOnOpen(c); err != nil {
		_ = c.Close()
		return nil, err
	}
	// What scoring reads stays warm for the life of this project, the shortlist
	// index too when it builds. Close drops it.
	if err := WarmOnOpen(c); err != nil {
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
	return closeLocked(s)
}

// closeLocked is the one body of closing a session. The caller holds s.mu.
func closeLocked(s *session) error {
	s.closed = true
	if s.cat == nil {
		return nil
	}
	err := s.cat.Close()
	s.cat = nil
	return err
}

// forget drops s from the registry under dir, only when dir still holds s.
// The caller holds no s.mu (lock order is reg.mu then s.mu).
func forget(dir string, s *session) {
	reg.mu.Lock()
	defer reg.mu.Unlock()
	if reg.sessions[dir] == s {
		delete(reg.sessions, dir)
	}
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
// Opening a different project closes any other held session first. When fn
// returns nil and left the graph needing a warm, Do runs warmGraph (the alias
// index and name format, not the shortlist index) before it returns; when fn
// fails it does not. A warm error never replaces fn's result, and leaves the
// flag set so the next call tries again.
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
		return runLocked(dir, s, fn)
	}
}

// runLocked runs fn on s, which Do has locked, and unlocks it. A panic in fn or
// in the warm closes the session (its caches may disagree with the catalog) and
// drops it from the registry, so the next Do reopens; the panic goes on to the
// caller.
func runLocked(dir string, s *session, fn func(*database.Catalog) error) error {
	done := false
	defer func() {
		if !done {
			_ = closeLocked(s)
		}
		s.mu.Unlock()
		if !done {
			forget(dir, s)
		}
	}()
	err := fn(s.cat)
	if err == nil && s.cat.Graph().WarmNeeded() {
		_ = warmGraph(s.cat)
	}
	done = true
	return err
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
