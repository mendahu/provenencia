package database

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"

	"github.com/mendahu/provenencia/core/apperr"
)

// Catalog format is SQLite user_version, not product SemVer and not a
// timestamped schema_migrations table. Each migrations/NNNNNN.sql file is one
// integer step. Mac and Windows must agree on that number.
//
//go:embed migrations/*.sql
var migrationFS embed.FS

type migration struct {
	to  int
	sql string
}

var (
	migrations    []migration
	formatVersion int // highest NNNNNN.sql; also the current user_version this engine writes
)

func init() {
	var err error
	migrations, formatVersion, err = parseMigrations(migrationFS, "migrations")
	if err != nil {
		panic(err)
	}
}

// parseMigrations loads NNNNNN.sql files (six decimal digits, no gaps from 1).
// The returned version is the last file number. Empty SQL is rejected.
func parseMigrations(fsys fs.FS, dir string) ([]migration, int, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, 0, err
	}
	var steps []migration
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".sql") {
			continue
		}
		base := strings.TrimSuffix(name, ".sql")
		if len(base) != 6 {
			return nil, 0, apperr.New(apperr.CodeInternalMigrations, apperr.KindInternal, "want NNNNNN.sql, got "+name)
		}
		n, err := strconv.Atoi(base)
		if err != nil || n < 1 {
			return nil, 0, apperr.New(apperr.CodeInternalMigrations, apperr.KindInternal, "bad name "+name)
		}
		body, err := fs.ReadFile(fsys, dir+"/"+name)
		if err != nil {
			return nil, 0, err
		}
		sqlText := strings.TrimSpace(string(body))
		if sqlText == "" {
			return nil, 0, apperr.New(apperr.CodeInternalMigrations, apperr.KindInternal, "empty "+name)
		}
		steps = append(steps, migration{to: n, sql: sqlText})
	}
	sort.Slice(steps, func(i, j int) bool { return steps[i].to < steps[j].to })
	for i, s := range steps {
		if s.to != i+1 {
			return nil, 0, apperr.New(apperr.CodeInternalMigrations, apperr.KindInternal, fmt.Sprintf("missing step %d", i+1))
		}
	}
	if len(steps) == 0 {
		return nil, 0, apperr.New(apperr.CodeInternalMigrations, apperr.KindInternal, "none")
	}
	return steps, steps[len(steps)-1].to, nil
}

// migrate applies embedded SQL until user_version matches formatVersion.
// Open and Create call this; there is no migrate CLI. A catalog with a higher
// user_version is refused (ErrUnsupportedVersion).
func migrate(db *sql.DB) error {
	var ver int
	if err := db.QueryRow(`PRAGMA user_version`).Scan(&ver); err != nil {
		return err
	}
	if ver > formatVersion {
		return ErrUnsupportedVersion.WithParams(strconv.Itoa(ver))
	}
	for _, step := range migrations {
		if ver >= step.to {
			continue
		}
		if _, err := db.Exec(step.sql); err != nil {
			return err
		}
		if _, err := db.Exec(fmt.Sprintf(`PRAGMA user_version = %d`, step.to)); err != nil {
			return err
		}
		ver = step.to
	}
	return nil
}

// setApplicationID stamps FourCC 'PROV' on a new catalog.
// The .provenencia suffix is not proof of ownership.
func setApplicationID(db *sql.DB) error {
	_, err := db.Exec(fmt.Sprintf(`PRAGMA application_id = %d`, ApplicationID))
	return err
}

// checkApplicationID refuses a foreign SQLite file in a project folder.
func checkApplicationID(db *sql.DB) error {
	var id int
	if err := db.QueryRow(`PRAGMA application_id`).Scan(&id); err != nil {
		return err
	}
	if id != ApplicationID {
		return ErrNotAProject
	}
	return nil
}
