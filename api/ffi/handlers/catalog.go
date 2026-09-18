package handlers

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/catalogsession"
	"github.com/mendahu/provenencia/core/database"
)

var (
	errInvalidID     = apperr.New(apperr.CodeSourcesInvalid, apperr.KindUser)
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

func uuidString(id []byte) string {
	if len(id) != 16 {
		return ""
	}
	u, err := uuid.FromBytes(id)
	if err != nil {
		return ""
	}
	return u.String()
}

func unmarshalErr(op string, err error) error {
	return fmt.Errorf("%s: unmarshal: %w", op, err)
}
