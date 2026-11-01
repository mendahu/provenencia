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
