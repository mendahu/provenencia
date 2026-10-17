// Package depictionvalues stores the depiction_values rows an Observation owns:
// the region of the Citation's page that shows a face (a locator document) and
// the image File Go stored from the app's rendering of that region.
//
// A row is replaced, never rewritten: a new region always comes with a new
// image. The caller checks the region against the Citation
// (locator.ValidateDepictionRegion); InsertTx checks only the row's own shape.
package depictionvalues

import (
	"database/sql"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/locator"
)

var ErrInvalid = apperr.New(apperr.CodeDepictionValuesInvalid, apperr.KindUser)

const (
	sqlInsert = `INSERT INTO depiction_values (id, region_json, image_file_id) VALUES (?, ?, ?)`
	sqlLookup = `SELECT region_json, image_file_id FROM depiction_values WHERE id = ?`
)

// Value is one depiction_values row.
type Value struct {
	ID          []byte
	RegionJSON  string
	ImageFileID []byte
}

// InsertTx mints a UUIDv7 id, inserts, and returns the id.
func InsertTx(tx *sql.Tx, v Value) ([]byte, error) {
	if tx == nil {
		return nil, ErrInvalid
	}
	v.RegionJSON = strings.TrimSpace(v.RegionJSON)
	if v.RegionJSON == "" || len(v.ImageFileID) != 16 {
		return nil, ErrInvalid
	}
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	if _, err := tx.Exec(sqlInsert, id[:], v.RegionJSON, v.ImageFileID); err != nil {
		return nil, err
	}
	return append([]byte(nil), id[:]...), nil
}

// LookupTx returns the row with id, or sql.ErrNoRows.
func LookupTx(tx *sql.Tx, id []byte) (Value, error) {
	if tx == nil || len(id) != 16 {
		return Value{}, ErrInvalid
	}
	v := Value{ID: append([]byte(nil), id...)}
	if err := tx.QueryRow(sqlLookup, id).Scan(&v.RegionJSON, &v.ImageFileID); err != nil {
		return Value{}, err
	}
	return v, nil
}

// Readable returns v as it may be shown. RegionJSON is emptied unless it
// passes locator.ValidateDepictionRegion against the locator of the Citation
// that holds the value; ImageFileID is emptied unless the stored File's media
// type is image/png, which is what ingest.StageImage always stores. Every
// reader that returns a stored depiction to the app goes through Readable.
// LookupTx does not: audit records what is stored.
func Readable(v Value, citationLocatorJSON, imageMediaType string) Value {
	if locator.ValidateDepictionRegion(v.RegionJSON, citationLocatorJSON) != nil {
		v.RegionJSON = ""
	}
	if imageMediaType != "image/png" {
		v.ImageFileID = nil
	}
	return v
}
