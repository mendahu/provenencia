package database

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"
)

// driverName is the sqlite3 driver with the app's SQL functions registered on
// every connection. Migrations may call them to backfill ids; the stored
// schema must not (other tools open the file without them).
const driverName = "provenencia_sqlite3"

// registered runs before any init, so migrate.go's init can open a database.
var registered = registerDriver()

func registerDriver() bool {
	sql.Register(driverName, &sqlite3.SQLiteDriver{ConnectHook: registerFunctions})
	return true
}

func registerFunctions(conn *sqlite3.SQLiteConn) error {
	return conn.RegisterFunc("uuid7", newUUID7, false)
}

// newUUID7 returns a fresh UUIDv7 as the 16-byte blob catalog ids use.
func newUUID7() ([]byte, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	return id[:], nil
}
