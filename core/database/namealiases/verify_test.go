package namealiases_test

import (
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	_ "github.com/mendahu/provenencia/core/database/autoreconciler" // registers Verify's truth recompute
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/namealiases"
	"github.com/mendahu/provenencia/core/valuecompare"
)

// With PROVENENCIA_GRAPH_VERIFY=1 a commit rebuilds the alias index and
// compares it with the stored one.
func TestVerifyChecksAliasIndex(t *testing.T) {
	t.Setenv("PROVENENCIA_GRAPH_VERIFY", "1")
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	if err := namealiases.Install(c); err != nil {
		t.Fatal(err)
	}
	if err := namealiases.Warm(c); err != nil {
		t.Fatal(err)
	}
	g := c.Graph()
	if err := g.OnCommit(effects.Commit{}); err != nil {
		t.Fatalf("a fresh alias index failed Verify: %v", err)
	}
	namealiases.Index.Set(g, valuecompare.AliasIndex{})
	err = g.OnCommit(effects.Commit{})
	if err == nil || !strings.Contains(err.Error(), "namealiases.index") {
		t.Fatalf("an alias index with no tables over a catalog with some: error = %v, want one naming namealiases.index", err)
	}
}
