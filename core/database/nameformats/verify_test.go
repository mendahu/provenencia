package nameformats_test

import (
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	_ "github.com/mendahu/provenencia/core/database/autoreconciler" // registers Verify's truth recompute
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/nameformats"
)

// With PROVENENCIA_GRAPH_VERIFY=1 a commit reloads the default name format and
// compares it with the stored one.
func TestVerifyChecksNameFormat(t *testing.T) {
	t.Setenv("PROVENENCIA_GRAPH_VERIFY", "1")
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	if err := nameformats.Install(c); err != nil {
		t.Fatal(err)
	}
	if err := nameformats.Warm(c); err != nil {
		t.Fatal(err)
	}
	g := c.Graph()
	if err := g.OnCommit(effects.Commit{}); err != nil {
		t.Fatalf("a fresh name format failed Verify: %v", err)
	}
	loaded, err := nameformats.LoadFor(g)
	if err != nil {
		t.Fatal(err)
	}
	loaded.Key += "-other"
	nameformats.Default.Set(g, loaded)
	err = g.OnCommit(effects.Commit{})
	if err == nil || !strings.Contains(err.Error(), "nameformats.default") {
		t.Fatalf("a Loaded with another key: error = %v, want one naming nameformats.default", err)
	}
}
