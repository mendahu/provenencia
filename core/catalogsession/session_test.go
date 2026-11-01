package catalogsession_test

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mendahu/provenencia/core/catalogsession"
	"github.com/mendahu/provenencia/core/catalogsession/catalogsessiontest"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
	"github.com/mendahu/provenencia/core/onboarding"
	"github.com/mendahu/provenencia/core/ref"
)

func TestDo(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, projectDir string)
	}{
		{
			name: "overlapping Do same dir both succeed",
			run: func(t *testing.T, projectDir string) {
				var wg sync.WaitGroup
				var fails atomic.Int32
				started := make(chan struct{})
				var startOnce sync.Once
				for i := 0; i < 8; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						err := catalogsession.Do(projectDir, func(c *database.Catalog) error {
							startOnce.Do(func() { close(started) })
							<-started
							time.Sleep(5 * time.Millisecond)
							_, err := c.DB()
							return err
						})
						if err != nil {
							fails.Add(1)
							t.Errorf("Do: %v", err)
						}
					}()
				}
				wg.Wait()
				if fails.Load() != 0 {
					t.Fatalf("%d Do calls failed", fails.Load())
				}
			},
		},
		{
			name: "serializes ops on one session",
			run: func(t *testing.T, projectDir string) {
				var concurrent atomic.Int32
				var max atomic.Int32
				var wg sync.WaitGroup
				for i := 0; i < 6; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						_ = catalogsession.Do(projectDir, func(*database.Catalog) error {
							n := concurrent.Add(1)
							for {
								cur := max.Load()
								if n <= cur || max.CompareAndSwap(cur, n) {
									break
								}
							}
							time.Sleep(2 * time.Millisecond)
							concurrent.Add(-1)
							return nil
						})
					}()
				}
				wg.Wait()
				if max.Load() != 1 {
					t.Fatalf("max concurrent ops %d want 1", max.Load())
				}
			},
		},
		{
			name: "Close then Do reopens",
			run: func(t *testing.T, projectDir string) {
				if err := catalogsession.Do(projectDir, func(*database.Catalog) error { return nil }); err != nil {
					t.Fatal(err)
				}
				if err := catalogsession.Close(projectDir); err != nil {
					t.Fatal(err)
				}
				c, err := database.Open(projectDir)
				if err != nil {
					t.Fatal(err)
				}
				_ = c.Close()
				if err := catalogsession.Do(projectDir, func(*database.Catalog) error { return nil }); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "panic drops the session",
			run: func(t *testing.T, projectDir string) {
				var first *database.Catalog
				func() {
					defer func() {
						if recover() == nil {
							t.Error("the panic did not reach the caller")
						}
					}()
					_ = catalogsession.Do(projectDir, func(c *database.Catalog) error {
						first = c
						panic("x")
					})
				}()
				done := make(chan error, 1)
				var second *database.Catalog
				go func() {
					done <- catalogsession.Do(projectDir, func(c *database.Catalog) error {
						second = c
						_, err := c.DB()
						return err
					})
				}()
				select {
				case err := <-done:
					if err != nil {
						t.Fatal(err)
					}
				case <-time.After(5 * time.Second):
					t.Fatal("the next Do waited on a lock the panic left held")
				}
				if second == nil || second == first {
					t.Fatalf("the next Do ran on the dropped catalog: %p %p", first, second)
				}
			},
		},
		{
			name: "switch dir closes previous",
			run: func(t *testing.T, projectDir string) {
				ident := t.TempDir()
				parent := t.TempDir()
				other, err := onboarding.Complete(ident, parent, "Other", "Other Family")
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = catalogsession.CloseAll() })

				if err := catalogsession.Do(projectDir, func(*database.Catalog) error { return nil }); err != nil {
					t.Fatal(err)
				}
				if err := catalogsession.Do(other.ProjectDir, func(*database.Catalog) error { return nil }); err != nil {
					t.Fatal(err)
				}
				c, err := database.Open(projectDir)
				if err != nil {
					t.Fatalf("previous session should be closed: %v", err)
				}
				_ = c.Close()
			},
		},
		{
			name: "direct Open while session held is already_open",
			run: func(t *testing.T, projectDir string) {
				if err := catalogsession.Do(projectDir, func(*database.Catalog) error { return nil }); err != nil {
					t.Fatal(err)
				}
				_, err := database.Open(projectDir)
				if !errors.Is(err, database.ErrAlreadyOpen) {
					t.Fatalf("got %v want %v", err, database.ErrAlreadyOpen)
				}
			},
		},
		{
			name: "empty projectDir",
			run: func(t *testing.T, _ string) {
				err := catalogsession.Do("  ", func(*database.Catalog) error { return nil })
				if !errors.Is(err, database.ErrNotAProject) {
					t.Fatalf("got %v want %v", err, database.ErrNotAProject)
				}
			},
		},
		{
			name: "open mints missing graph refs",
			run: func(t *testing.T, _ string) {
				c, err := database.Create(t.TempDir(), "graphs.provenencia")
				if err != nil {
					t.Fatal(err)
				}
				dir := c.Dir()
				sourceID := evidencegraphstest.SeedRefless(t, c)
				t.Cleanup(func() { _ = catalogsession.Close(dir) })
				err = catalogsession.Do(dir, func(c *database.Catalog) error {
					if got := evidencegraphstest.SoleGraph(t, c, sourceID).Ref; !ref.Valid(got) {
						t.Errorf("graph ref %q", got)
					}
					return nil
				})
				if err != nil {
					t.Fatal(err)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() { _ = catalogsession.CloseAll() })
			ident := t.TempDir()
			parent := t.TempDir()
			created, err := onboarding.Complete(ident, parent, "Jake", "Robins")
			if err != nil {
				t.Fatal(err)
			}
			tt.run(t, created.ProjectDir)
		})
	}
}

func TestCloseAll(t *testing.T) {
	ident := t.TempDir()
	parent := t.TempDir()
	created, err := onboarding.Complete(ident, parent, "Jake", "Robins")
	if err != nil {
		t.Fatal(err)
	}
	if err := catalogsession.Do(created.ProjectDir, func(*database.Catalog) error { return nil }); err != nil {
		t.Fatal(err)
	}
	if err := catalogsession.CloseAll(); err != nil {
		t.Fatal(err)
	}
	c, err := database.Open(created.ProjectDir)
	if err != nil {
		t.Fatal(err)
	}
	_ = c.Close()
}

func TestHold(t *testing.T) {
	projectDir := catalogsessiontest.Project(t)
	release := catalogsessiontest.Hold(t, projectDir)

	entered := make(chan struct{})
	go func() {
		_ = catalogsession.Do(projectDir, func(*database.Catalog) error {
			close(entered)
			return nil
		})
	}()
	select {
	case <-entered:
		t.Fatal("a second Do ran while the session was held")
	case <-time.After(100 * time.Millisecond):
	}
	release()
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("the second Do did not run after release")
	}
	release() // idempotent
}
