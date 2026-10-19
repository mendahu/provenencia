package catalogsession_test

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mendahu/provenencia/core/catalogsession"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/sourcevocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/onboarding"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
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
				sourceID := seedGraphRefless(t, c)
				t.Cleanup(func() { _ = catalogsession.Close(dir) })
				err = catalogsession.Do(dir, func(c *database.Catalog) error {
					if got := graphRef(t, c, sourceID); !ref.Valid(got) {
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

// seedGraphRefless builds a catalog with one Source, clears its Default
// graph's ref, and returns the Source id with the catalog closed.
func seedGraphRefless(t *testing.T, c *database.Catalog) []byte {
	t.Helper()
	userID := []byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	r, err := ref.Mint(ref.PrefixUser)
	if err != nil {
		t.Fatal(err)
	}
	if err := users.Upsert(c, userID, "Jake", r); err != nil {
		t.Fatal(err)
	}
	if err := sourcevocab.Install(c); err != nil {
		t.Fatal(err)
	}
	typ, err := sourcetypes.Lookup(c, "birth_certificate", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typ.ID, Title: "Register"})
	})
	if err != nil {
		t.Fatal(err)
	}
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE evidence_graphs SET ref = NULL`); err != nil {
		t.Fatal(err)
	}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	return src.ID
}

func graphRef(t *testing.T, c *database.Catalog, sourceID []byte) string {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	var r string
	if err := db.QueryRow(`SELECT COALESCE(ref, '') FROM evidence_graphs WHERE source_id = ?`, sourceID).Scan(&r); err != nil {
		t.Fatal(err)
	}
	return r
}
