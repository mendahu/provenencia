package onboarding

import (
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/project"
	"github.com/mendahu/provenencia/core/database/sourcecredibilitygrades"
	"github.com/mendahu/provenencia/core/database/sourcevocab"
	"github.com/mendahu/provenencia/core/database/users"
)

// createCatalog installs a new catalog (create path). OpenCatalog is a one-shot
// open for tests and low-level use. Researcher FFI and onboarding open/list
// paths use core/catalogsession (held session + serial queue). database.Create/Open
// stay migrate-only. User refs and project.uuid are reconciled on create and open;
// Source vocabulary and credibility grades are installed once at create only
// (never healed on open).
func createCatalog(parent, folder string) (*database.Catalog, error) {
	c, err := database.Create(parent, folder)
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
	if err := sourcevocab.Install(c); err != nil {
		_ = c.Close()
		return nil, err
	}
	if err := sourcecredibilitygrades.Install(c); err != nil {
		_ = c.Close()
		return nil, err
	}
	return c, nil
}

// OpenCatalog opens a project once (migrate + ensure user refs + project.uuid).
// Prefer catalogsession.Do for researcher paths so opens are amortized and serialized.
// Does not re-install or heal Source vocabulary or credibility grades.
func OpenCatalog(projectDir string) (*database.Catalog, error) {
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
	return c, nil
}
