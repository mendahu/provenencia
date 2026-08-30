package users

import (
	"errors"
	"strings"

	"github.com/mendahu/provenencia/core/database"
)

var ErrInvalid = errors.New("invalid user id or display name")

const (
	sqlUpsert = `INSERT INTO users (id, display_name) VALUES (?, ?)
		ON CONFLICT(id) DO UPDATE SET display_name = excluded.display_name`
	sqlLookup = `SELECT display_name FROM users WHERE id = ?`
	sqlList   = `SELECT id, display_name FROM users ORDER BY display_name COLLATE NOCASE, id`
)

// User is one catalog contributor row.
type User struct {
	ID          []byte
	DisplayName string
}

// Upsert inserts or updates a users row. id must be a 16-byte UUIDv7 blob.
func Upsert(c *database.Catalog, id []byte, displayName string) error {
	db, err := c.DB()
	if err != nil {
		return err
	}
	name := strings.TrimSpace(displayName)
	if len(id) != 16 || name == "" {
		return ErrInvalid
	}
	_, err = db.Exec(sqlUpsert, id, name)
	return err
}

// Lookup returns the display name for id, or sql.ErrNoRows.
func Lookup(c *database.Catalog, id []byte) (string, error) {
	db, err := c.DB()
	if err != nil {
		return "", err
	}
	var name string
	err = db.QueryRow(sqlLookup, id).Scan(&name)
	return name, err
}

// List returns all users rows. Order is display name, then id.
func List(c *database.Catalog) ([]User, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(sqlList)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.DisplayName); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}
