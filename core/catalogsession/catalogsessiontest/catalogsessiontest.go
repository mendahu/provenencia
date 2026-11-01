// Package catalogsessiontest holds test helpers for code that must run while
// another call has the project's catalog session.
package catalogsessiontest

import (
	"sync"
	"testing"
	"time"

	"github.com/mendahu/provenencia/core/catalogsession"
	"github.com/mendahu/provenencia/core/database"
)

// Hold starts a catalogsession.Do on projectDir whose function blocks until
// release is called, and returns once that function is running, so the
// session is held. release is idempotent and is also registered with
// t.Cleanup.
//
// A test that holds a project builds its fixtures through catalogsession.Do
// (or before the first Do), never on a catalog it opened with database.Open,
// which would hold the project's exclusive lock; it calls Hold only around the
// call under test and releases before its next Do.
func Hold(t testing.TB, projectDir string) (release func()) {
	t.Helper()
	running := make(chan struct{})
	block := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- catalogsession.Do(projectDir, func(*database.Catalog) error {
			close(running)
			<-block
			return nil
		})
	}()
	select {
	case <-running:
	case err := <-done:
		t.Fatalf("catalogsessiontest.Hold: Do returned before running: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("catalogsessiontest.Hold: the session was not taken in 5s")
	}
	var once sync.Once
	release = func() {
		once.Do(func() {
			close(block)
			<-done
		})
	}
	t.Cleanup(release)
	return release
}
