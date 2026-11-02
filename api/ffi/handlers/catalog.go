package handlers

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/catalogsession"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/evidencegraphs"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/ingest"
)

var (
	// errInvalidID is any malformed record id: a Subject's, a handle's, an
	// Observation's, not only a Source's.
	errInvalidID     = apperr.New(apperr.CodeRequestInvalidID, apperr.KindUser)
	errInvalidUserID = apperr.New(apperr.CodeUsersInvalid, apperr.KindUser)
)

// withProjectCatalog runs fn on the held exclusive catalog session for projectDir.
func withProjectCatalog(projectDir string, fn func(*database.Catalog) error) error {
	return catalogsession.Do(projectDir, fn)
}

// withStaged is the one stage-then-write entry point. It does the file work
// first, outside the catalog session: stage makes the request's Stages in the
// project folder Do will use (a stage that fails discards what it staged
// itself). Then fn runs on the held session with the stages' Settle deferred
// first inside it, which purges an object the write installed and did not use;
// the stages' Discard is deferred outside the session and removes any temp the
// session never reached. A client file (an attachment, a scan) is a stage func
// for withStaged; a handler never copies, hashes or decodes inside
// withProjectCatalog.
func withStaged(projectDir string, stage func(dir string) (ingest.Stages, error), fn func(*database.Catalog) error) error {
	dir, err := catalogsession.Key(projectDir)
	if err != nil {
		return err
	}
	stages, err := stage(dir)
	if err != nil {
		return err
	}
	defer stages.Discard()
	return withProjectCatalog(dir, func(c *database.Catalog) error {
		defer stages.Settle(c)
		return fn(c)
	})
}

// withStagedDepictions is withStaged with the images of the inputs' depictions
// as the stages (ingest.StageDepictions).
func withStagedDepictions(projectDir string, inputs []observations.Input, fn func(*database.Catalog) error) error {
	return withStaged(projectDir, func(dir string) (ingest.Stages, error) {
		return ingest.StageDepictions(dir, inputs)
	}, fn)
}

func parseUserID(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, errInvalidUserID
	}
	return parseID(s)
}

func parseID(s string) ([]byte, error) {
	u, err := uuid.Parse(strings.TrimSpace(s))
	if err != nil || u.Version() != 7 {
		return nil, errInvalidID
	}
	return u[:], nil
}

func optionalID(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	return parseID(s)
}

// parseIDs parses a repeated id field; the first malformed id fails it. An
// empty field returns nil.
func parseIDs(raws []string) ([][]byte, error) {
	var out [][]byte
	for _, raw := range raws {
		id, err := parseID(raw)
		if err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, nil
}

// parseGraphLinks reads an Artifact's graph links with their presence: an
// absent message is the zero value (the default); a present one is Given, and
// an empty one means none.
func parseGraphLinks(m *engine.GraphLinks) (evidencegraphs.GraphLinks, error) {
	if m == nil {
		return evidencegraphs.GraphLinks{}, nil
	}
	ids, err := parseIDs(m.GetGraphIds())
	if err != nil {
		return evidencegraphs.GraphLinks{}, err
	}
	return evidencegraphs.GraphLinks{Given: true, IDs: ids}, nil
}

func unmarshalErr(op string, err error) error {
	return fmt.Errorf("%s: unmarshal: %w", op, err)
}
