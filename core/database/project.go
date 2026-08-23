// Package database owns the *.provenance folder catalog (provenance.sqlite).
// Identity files, FFI RPCs, and Source ingest do not live here.
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
	ErrInvalidUser        = errors.New("invalid user id or display name")
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

// UpsertUser inserts or updates a users row. id must be a 16-byte UUIDv7 blob.
func (p *Project) UpsertUser(id []byte, displayName string) error {
	if p == nil || p.db == nil {
		return errors.New("project closed")
	}
	name := strings.TrimSpace(displayName)
	if len(id) != 16 || name == "" {
		return ErrInvalidUser
	}
	_, err := p.db.Exec(`INSERT INTO users (id, display_name) VALUES (?, ?)
		ON CONFLICT(id) DO UPDATE SET display_name = excluded.display_name`, id, name)
	return err
}

// LookupUser returns the display name for id, or sql.ErrNoRows.
func (p *Project) LookupUser(id []byte) (string, error) {
	if p == nil || p.db == nil {
		return "", errors.New("project closed")
	}
	var name string
	err := p.db.QueryRow(`SELECT display_name FROM users WHERE id = ?`, id).Scan(&name)
	return name, err
}
