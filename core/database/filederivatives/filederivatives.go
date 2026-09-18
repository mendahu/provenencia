// Package filederivatives links source Files to generated derivative Files.
package filederivatives

import (
	"bytes"
	"database/sql"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
)

var ErrInvalid = apperr.New(apperr.CodeFileDerivativesInvalid, apperr.KindUser)

// TypeThumbnail is the derivative_type for small JPEG previews.
const TypeThumbnail = "thumbnail"

const (
	sqlInsert = `INSERT INTO file_derivatives (id, source_file_id, derived_file_id, derivative_type)
		VALUES (?, ?, ?, ?)`
	sqlLookup = `SELECT id, source_file_id, derived_file_id, derivative_type
		FROM file_derivatives WHERE source_file_id = ? AND derivative_type = ?`
	sqlListBySource = `SELECT id, source_file_id, derived_file_id, derivative_type
		FROM file_derivatives WHERE source_file_id = ?
		ORDER BY derivative_type COLLATE NOCASE`
)

// Link is one file_derivatives row.
type Link struct {
	ID             []byte
	SourceFileID   []byte
	DerivedFileID  []byte
	DerivativeType string
}

// Lookup returns the derivative link for (sourceFileID, derivativeType), or sql.ErrNoRows.
func Lookup(c *database.Catalog, sourceFileID []byte, derivativeType string) (Link, error) {
	db, err := c.DB()
	if err != nil {
		return Link{}, err
	}
	derivativeType = strings.TrimSpace(derivativeType)
	if len(sourceFileID) != 16 || derivativeType == "" {
		return Link{}, ErrInvalid
	}
	return scanLink(db.QueryRow(sqlLookup, sourceFileID, derivativeType))
}

// Insert writes a new derivative link on tx.
func Insert(tx *sql.Tx, link Link) error {
	if tx == nil {
		return ErrInvalid
	}
	link.DerivativeType = strings.TrimSpace(link.DerivativeType)
	if len(link.ID) != 16 || len(link.SourceFileID) != 16 || len(link.DerivedFileID) != 16 ||
		link.DerivativeType == "" || bytes.Equal(link.SourceFileID, link.DerivedFileID) {
		return ErrInvalid
	}
	_, err := tx.Exec(sqlInsert, link.ID, link.SourceFileID, link.DerivedFileID, link.DerivativeType)
	if err != nil {
		return mapConstraint(err)
	}
	return nil
}

// ListBySourceFile returns all derivative links for a source File.
func ListBySourceFile(c *database.Catalog, sourceFileID []byte) ([]Link, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	if len(sourceFileID) != 16 {
		return nil, ErrInvalid
	}
	rows, err := db.Query(sqlListBySource, sourceFileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Link
	for rows.Next() {
		link, err := scanLink(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, link)
	}
	return out, rows.Err()
}

// NewID mints a UUIDv7 id for a new link row.
func NewID() ([]byte, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	return id[:], nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanLink(row rowScanner) (Link, error) {
	var link Link
	if err := row.Scan(&link.ID, &link.SourceFileID, &link.DerivedFileID, &link.DerivativeType); err != nil {
		return Link{}, err
	}
	return link, nil
}

func mapConstraint(err error) error {
	if database.IsUniqueConflict(err) {
		return err
	}
	if database.IsConstraint(err) {
		return ErrInvalid
	}
	return err
}

// IsUniqueConflict reports a SQLite UNIQUE constraint failure.
func IsUniqueConflict(err error) bool {
	return database.IsUniqueConflict(err)
}
