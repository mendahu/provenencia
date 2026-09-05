package users

import (
	"errors"
	"strings"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/ref"
)

var ErrInvalid = errors.New("Invalid user ID, display name, or ref.")

const (
	sqlUpsert = `INSERT INTO users (id, display_name, ref) VALUES (?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			display_name = excluded.display_name,
			ref = excluded.ref`
	sqlLookup = `SELECT display_name, COALESCE(ref, '') FROM users WHERE id = ?`
	sqlList   = `SELECT id, display_name, COALESCE(ref, '') FROM users ORDER BY display_name COLLATE NOCASE, id`
	sqlListMissingRef = `SELECT id, display_name FROM users WHERE ref IS NULL OR ref = ''`
	sqlSetRef = `UPDATE users SET ref = ? WHERE id = ?`
)

// User is one catalog contributor row.
type User struct {
	ID          []byte
	DisplayName string
	Ref         string
}

// Upsert inserts or updates a users row. id must be a 16-byte UUIDv7 blob; ref must be valid.
func Upsert(c *database.Catalog, id []byte, displayName, userRef string) error {
	db, err := c.DB()
	if err != nil {
		return err
	}
	name := strings.TrimSpace(displayName)
	userRef = strings.TrimSpace(userRef)
	if len(id) != 16 || name == "" || ref.Validate(userRef) != nil {
		return ErrInvalid
	}
	_, err = db.Exec(sqlUpsert, id, name, userRef)
	return err
}

// Lookup returns the user for id, or sql.ErrNoRows.
func Lookup(c *database.Catalog, id []byte) (User, error) {
	db, err := c.DB()
	if err != nil {
		return User{}, err
	}
	var u User
	u.ID = append([]byte(nil), id...)
	err = db.QueryRow(sqlLookup, id).Scan(&u.DisplayName, &u.Ref)
	if err != nil {
		return User{}, err
	}
	if u.Ref == "" {
		u.Ref = ""
	}
	return u, nil
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
		if err := rows.Scan(&u.ID, &u.DisplayName, &u.Ref); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// EnsureRefs mints USR-… refs for any users row missing one.
func EnsureRefs(c *database.Catalog) error {
	db, err := c.DB()
	if err != nil {
		return err
	}
	rows, err := db.Query(sqlListMissingRef)
	if err != nil {
		return err
	}
	defer rows.Close()
	type missing struct {
		id   []byte
		name string
	}
	var need []missing
	for rows.Next() {
		var m missing
		if err := rows.Scan(&m.id, &m.name); err != nil {
			return err
		}
		need = append(need, m)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, m := range need {
		r, err := ref.Mint(ref.PrefixUser)
		if err != nil {
			return err
		}
		if _, err := db.Exec(sqlSetRef, r, m.id); err != nil {
			return err
		}
	}
	return nil
}
