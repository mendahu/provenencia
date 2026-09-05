// Package database owns the *.provenencia folder catalog (provenencia.sqlite).
// Identity files, FFI RPCs, and Source ingest do not live here.
package database

import (
	"database/sql"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mattn/go-sqlite3"
)

// ApplicationID is the SQLite application_id FourCC 'PROV' (0x50524F56).
// Folder suffix .provenencia is a hint; ownership is this id plus user_version.
const ApplicationID = 0x50524F56

const (
	catalogFile = "provenencia.sqlite"
	// Suffix is the project folder extension. It is a Finder hint, not catalog ownership.
	Suffix         = ".provenencia"
	objectsDir     = "objects"
	derivativesDir = "derivatives"
)

// Catalog is an exclusive connection to one provenencia.sqlite file.
type Catalog struct {
	dir string
	db  *sql.DB
}

func (c *Catalog) Dir() string {
	return c.dir
}

func dsn(sqlitePath string) string {
	return dsnOpts(sqlitePath, false)
}

func dsnReadOnly(sqlitePath string) string {
	return dsnOpts(sqlitePath, true)
}

func dsnOpts(sqlitePath string, readOnly bool) string {
	u := url.URL{
		Scheme: "file",
		Path:   filepath.ToSlash(sqlitePath),
	}
	q := url.Values{}
	q.Set("_busy_timeout", "100")
	q.Set("_foreign_keys", "1")
	if readOnly {
		q.Set("mode", "ro")
	} else {
		q.Set("_journal_mode", "WAL")
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func catalogPath(dir string) string {
	return filepath.Join(dir, catalogFile)
}

func openDB(sqlitePath string) (*sql.DB, error) {
	return openDBURI(dsn(sqlitePath))
}

func openDBURI(uri string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", uri)
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

// Create writes a new *.provenencia folder with an empty catalog and takes an exclusive lock.
func Create(parent, folderName string) (*Catalog, error) {
	if !strings.HasSuffix(folderName, Suffix) {
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
	return &Catalog{dir: dir, db: db}, nil
}

// Open opens an existing project directory by catalog content, not folder suffix.
func Open(dir string) (*Catalog, error) {
	sqlitePath, err := locateCatalog(dir)
	if err != nil {
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
	return &Catalog{dir: dir, db: db}, nil
}

// OpenReadOnly opens a catalog for queries without migrating or taking the exclusive lock.
func OpenReadOnly(dir string) (*Catalog, error) {
	sqlitePath, err := locateCatalog(dir)
	if err != nil {
		return nil, err
	}
	db, err := openDBURI(dsnReadOnly(sqlitePath))
	if err != nil {
		return nil, mapLockErr(err)
	}
	if err := checkApplicationID(db); err != nil {
		db.Close()
		return nil, err
	}
	if err := refuseNewerFormat(db); err != nil {
		db.Close()
		return nil, err
	}
	return &Catalog{dir: dir, db: db}, nil
}

func locateCatalog(dir string) (string, error) {
	st, err := os.Stat(dir)
	if err != nil {
		return "", err
	}
	if !st.IsDir() {
		return "", ErrNotAProject
	}
	sqlitePath := catalogPath(dir)
	if _, err := os.Stat(sqlitePath); err != nil {
		if os.IsNotExist(err) {
			return "", ErrNotAProject
		}
		return "", err
	}
	return sqlitePath, nil
}

func refuseNewerFormat(db *sql.DB) error {
	var ver int
	if err := db.QueryRow(`PRAGMA user_version`).Scan(&ver); err != nil {
		return err
	}
	if ver > formatVersion {
		return ErrUnsupportedVersion.WithParams(strconv.Itoa(ver))
	}
	return nil
}

func (c *Catalog) Close() error {
	if c == nil || c.db == nil {
		return nil
	}
	err := c.db.Close()
	c.db = nil
	return err
}

// DB is the exclusive SQLite handle. Nested table packages use this; callers should not.
func (c *Catalog) DB() (*sql.DB, error) {
	if c == nil || c.db == nil {
		return nil, ErrClosed
	}
	return c.db, nil
}
