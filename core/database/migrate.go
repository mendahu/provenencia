package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"regexp"
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

// foreignKeysOffDirective is the one line a step uses to ask migrate to run it
// with foreign keys off (a table rebuild). migrate checks every foreign key
// before it commits such a step.
const foreignKeysOffDirective = "-- migrate: foreign_keys off"

type migration struct {
	to             int
	sql            string
	foreignKeysOff bool
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
	if err := initExpectedSchemaHash(); err != nil {
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
		upper := strings.ToUpper(sqlText)
		if strings.Contains(upper, "PRAGMA FOREIGN_KEYS") || strings.Contains(upper, "PRAGMA FOREIGN_KEY_CHECK") {
			return nil, 0, apperr.New(apperr.CodeInternalMigrations, apperr.KindInternal, name+" toggles foreign_keys; use the "+foreignKeysOffDirective+" line")
		}
		off, err := scanDirectives(name, sqlText)
		if err != nil {
			return nil, 0, err
		}
		steps = append(steps, migration{to: n, sql: sqlText, foreignKeysOff: off})
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

// transactionControl matches a line that is a transaction statement on its
// own. migrate owns each step's transaction. A trigger body's BEGIN has no
// semicolon and its END; is not matched, so CREATE TRIGGER still loads; a lone
// END; is therefore not refused.
var transactionControl = regexp.MustCompile(`^(begin( (deferred|immediate|exclusive))?( transaction)?|commit( transaction)?|rollback( transaction)?)\s*;$`)

// scanDirectives reads a step's lines. Any line starting "-- migrate:" (after
// trimming, any case) must be exactly the foreign_keys off directive, and a
// line that is only a transaction statement is refused. It reports whether the
// step carries the directive.
func scanDirectives(name, sqlText string) (foreignKeysOff bool, err error) {
	for i, raw := range strings.Split(sqlText, "\n") {
		line := strings.TrimRight(raw, "\r")
		lower := strings.ToLower(strings.TrimSpace(line))
		if strings.HasPrefix(lower, "-- migrate:") {
			if line != foreignKeysOffDirective {
				return false, apperr.New(apperr.CodeInternalMigrations, apperr.KindInternal,
					fmt.Sprintf("%s line %d: only %q is allowed", name, i+1, foreignKeysOffDirective))
			}
			foreignKeysOff = true
		}
		if transactionControl.MatchString(lower) {
			return false, apperr.New(apperr.CodeInternalMigrations, apperr.KindInternal,
				fmt.Sprintf("%s line %d: migrate owns the transaction; remove %q", name, i+1, strings.TrimSpace(line)))
		}
	}
	return foreignKeysOff, nil
}

// migrate applies embedded SQL until user_version matches formatVersion.
// Open and Create call this; there is no migrate CLI. A catalog with a higher
// user_version is refused (ErrUnsupportedVersion).
//
// Each step runs in one transaction with its user_version write, so a failed
// step leaves the catalog at the previous version. A step marked
// foreignKeysOff runs with foreign keys off and is refused, rolled back, if
// PRAGMA foreign_key_check reports a row before it would commit. Foreign keys
// are back on when migrate returns.
func migrate(db *sql.DB) error {
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	var ver int
	if err := conn.QueryRowContext(ctx, `PRAGMA user_version`).Scan(&ver); err != nil {
		return err
	}
	if ver > formatVersion {
		return ErrUnsupportedVersion.WithParams(strconv.Itoa(ver))
	}
	for _, step := range migrations {
		if ver >= step.to {
			continue
		}
		if err := applyStep(ctx, conn, step); err != nil {
			return err
		}
		ver = step.to
	}
	return nil
}

func applyStep(ctx context.Context, conn *sql.Conn, step migration) (err error) {
	if step.foreignKeysOff {
		if _, err := conn.ExecContext(ctx, `PRAGMA foreign_keys=OFF`); err != nil {
			return err
		}
		defer func() {
			if _, onErr := conn.ExecContext(ctx, `PRAGMA foreign_keys=ON`); onErr != nil && err == nil {
				err = onErr
			}
		}()
	}
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if _, err = tx.ExecContext(ctx, step.sql); err != nil {
		return err
	}
	if step.foreignKeysOff {
		if err = checkForeignKeys(ctx, tx, step.to); err != nil {
			return err
		}
	}
	if _, err = tx.ExecContext(ctx, fmt.Sprintf(`PRAGMA user_version = %d`, step.to)); err != nil {
		return err
	}
	return tx.Commit()
}

// checkForeignKeys fails when any foreign key is broken after a step that ran
// with foreign keys off.
func checkForeignKeys(ctx context.Context, tx *sql.Tx, to int) error {
	rows, err := tx.QueryContext(ctx, `PRAGMA foreign_key_check`)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		var table, parent string
		var rowid sql.NullInt64
		var fkid int
		if err := rows.Scan(&table, &rowid, &parent, &fkid); err != nil {
			return err
		}
		return apperr.New(apperr.CodeInternalMigrations, apperr.KindInternal,
			fmt.Sprintf("step %d leaves a broken foreign key in %s (parent %s)", to, table, parent))
	}
	return rows.Err()
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
