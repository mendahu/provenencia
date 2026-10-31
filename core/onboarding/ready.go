package onboarding

import (
	"github.com/mendahu/provenencia/core/catalogsession"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/namealiases"
	"github.com/mendahu/provenencia/core/database/nameformats"
	"github.com/mendahu/provenencia/core/database/sourcecredibilitygrades"
	"github.com/mendahu/provenencia/core/database/sourcevocab"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
)

// createCatalog installs a new catalog (create path). OpenCatalog is a one-shot
// open for tests and low-level use. Researcher FFI and onboarding open/list
// paths use core/catalogsession (held session + serial queue). database.Create/Open
// stay migrate-only. User refs and project.uuid are reconciled on create and open;
// Source vocabulary, credibility grades, Interpretation subject vocabulary
// (types + Properties + bindings), the Western name format, and abbreviation
// and nickname tables are installed once at create only (never healed on open).
func createCatalog(parent, folder string) (*database.Catalog, error) {
	c, err := database.Create(parent, folder)
	if err != nil {
		return nil, err
	}
	if err := sourcevocab.Install(c); err != nil {
		_ = c.Close()
		return nil, err
	}
	if err := sourcecredibilitygrades.Install(c); err != nil {
		_ = c.Close()
		return nil, err
	}
	if err := subjectvocab.Install(c); err != nil {
		_ = c.Close()
		return nil, err
	}
	if err := nameformats.Install(c); err != nil {
		_ = c.Close()
		return nil, err
	}
	if err := namealiases.Install(c); err != nil {
		_ = c.Close()
		return nil, err
	}
	if err := catalogsession.ReconcileOnOpen(c); err != nil {
		_ = c.Close()
		return nil, err
	}
	return c, nil
}

// OpenCatalog opens a project once (migrate + the open-time repairs of catalogsession.ReconcileOnOpen).
// Prefer catalogsession.Do for researcher paths so opens are amortized and serialized.
// Does not re-install or heal Source vocabulary, credibility grades, Interpretation subject vocabulary, name formats, or alias tables.
func OpenCatalog(projectDir string) (*database.Catalog, error) {
	c, err := database.Open(projectDir)
	if err != nil {
		return nil, err
	}
	if err := catalogsession.ReconcileOnOpen(c); err != nil {
		_ = c.Close()
		return nil, err
	}
	return c, nil
}
