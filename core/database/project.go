package database

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mattn/go-sqlite3"
)

// ApplicationID is the SQLite application_id FourCC 'PROV' (0x50524F56).
// Folder suffix .provenance is a hint; ownership is this id plus user_version.
const ApplicationID = 0x50524F56

const (
	catalogFile    = "provenance.sqlite"
	formatVersion  = 1
	projectSuffix  = ".provenance"
	objectsDir     = "objects"
	derivativesDir = "derivatives"
)

var (
	ErrAlreadyExists      = errors.New("project already exists")
	ErrAlreadyOpen        = errors.New("project already open")
	ErrNotAProject        = errors.New("not a provenance catalog")
	ErrUnsupportedVersion = errors.New("unsupported catalog version")
	ErrInvalidFolderName  = errors.New("folder name must end in .provenance")
)

// Project is an exclusive connection to one provenance.sqlite catalog.
type Project struct {
	dir string
	db  *sql.DB
}

func (p *Project) Dir() string {
	return p.dir
}

func dsn(sqlitePath string) string {
	return fmt.Sprintf("file:%s?_busy_timeout=100&_foreign_keys=1&_journal_mode=WAL", sqlitePath)
}

func catalogPath(dir string) string {
	return filepath.Join(dir, catalogFile)
}

func openDB(sqlitePath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dsn(sqlitePath))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func takeExclusiveLock(db *sql.DB) error {
	if _, err := db.Exec(`PRAGMA locking_mode=EXCLUSIVE`); err != nil {
		return err
	}
	if _, err := db.Exec(`BEGIN IMMEDIATE`); err != nil {
		return mapLockErr(err)
	}
	if _, err := db.Exec(`COMMIT`); err != nil {
		return mapLockErr(err)
	}
	return nil
}

func mapLockErr(err error) error {
	var se sqlite3.Error
	if errors.As(err, &se) && (se.Code == sqlite3.ErrBusy || se.Code == sqlite3.ErrLocked) {
		return ErrAlreadyOpen
	}
	return err
}

func migrate(db *sql.DB) error {
	var ver int
	if err := db.QueryRow(`PRAGMA user_version`).Scan(&ver); err != nil {
		return err
	}
	if ver > formatVersion {
		return fmt.Errorf("%w: %d", ErrUnsupportedVersion, ver)
	}
	if ver < 1 {
		if _, err := db.Exec(`CREATE TABLE users (
			id BLOB PRIMARY KEY,
			display_name TEXT NOT NULL
		) STRICT`); err != nil {
			return err
		}
		if _, err := db.Exec(fmt.Sprintf(`PRAGMA user_version = %d`, formatVersion)); err != nil {
			return err
		}
	}
	return nil
}

func setApplicationID(db *sql.DB) error {
	_, err := db.Exec(fmt.Sprintf(`PRAGMA application_id = %d`, ApplicationID))
	return err
}

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

// Create writes a new *.provenance folder with an empty catalog and takes an exclusive lock.
func Create(parent, folderName string) (*Project, error) {
	if !strings.HasSuffix(folderName, projectSuffix) {
		return nil, ErrInvalidFolderName
	}
	dir := filepath.Join(parent, folderName)
	if _, err := os.Stat(dir); err == nil {
		return nil, ErrAlreadyExists
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(dir, objectsDir), 0o755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(dir, derivativesDir), 0o755); err != nil {
		_ = os.RemoveAll(dir)
		return nil, err
	}
	db, err := openDB(catalogPath(dir))
	if err != nil {
		_ = os.RemoveAll(dir)
		return nil, err
	}
	if err := setApplicationID(db); err != nil {
		db.Close()
		_ = os.RemoveAll(dir)
		return nil, err
	}
	if err := migrate(db); err != nil {
		db.Close()
		_ = os.RemoveAll(dir)
		return nil, err
	}
	if err := takeExclusiveLock(db); err != nil {
		db.Close()
		_ = os.RemoveAll(dir)
		return nil, err
	}
	return &Project{dir: dir, db: db}, nil
}

// Open opens an existing project directory by catalog content, not folder suffix.
func Open(dir string) (*Project, error) {
	st, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !st.IsDir() {
		return nil, ErrNotAProject
	}
	sqlitePath := catalogPath(dir)
	if _, err := os.Stat(sqlitePath); err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotAProject
		}
		return nil, err
	}
	db, err := openDB(sqlitePath)
	if err != nil {
		return nil, mapLockErr(err)
	}
	if err := checkApplicationID(db); err != nil {
		db.Close()
		return nil, err
	}
	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}
	if err := takeExclusiveLock(db); err != nil {
		db.Close()
		return nil, err
	}
	return &Project{dir: dir, db: db}, nil
}

func (p *Project) Close() error {
	if p == nil || p.db == nil {
		return nil
	}
	err := p.db.Close()
	p.db = nil
	return err
}
