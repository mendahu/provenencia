// Package depictionvalues stores the depiction_values rows an Observation owns:
// a region of the Citation's page and the image File Go stored from it. The
// value type, its shape check and the read rules live in the pure package
// core/depictionvalue; this package holds the queries and the join helper.
//
// A row is replaced, never rewritten: a new region always comes with a new
// image. The caller checks the region against the Citation
// (locator.ValidateDepictionRegion).
package depictionvalues

import (
	"database/sql"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/depictionvalue"
)

const (
	sqlInsert = `INSERT INTO depiction_values (id, region_json, image_file_id) VALUES (?, ?, ?)`
	sqlLookup = `SELECT region_json, image_file_id FROM depiction_values WHERE id = ?`
)

// InsertTx mints a UUIDv7 id, inserts, and returns the id.
func InsertTx(tx *sql.Tx, v depictionvalue.Value) ([]byte, error) {
	if tx == nil {
		return nil, depictionvalue.ErrInvalid
	}
	v.RegionJSON = strings.TrimSpace(v.RegionJSON)
	if err := depictionvalue.Validate(v); err != nil {
		return nil, err
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

// RowMap is the audit row of a depiction_values row.
func RowMap(id []byte, v depictionvalue.Value) map[string]any {
	return map[string]any{
		"id":            database.UUIDJSON(id),
		"region_json":   v.RegionJSON,
		"image_file_id": database.UUIDJSON(v.ImageFileID),
	}
}

// CreateChange is the audit change of a new depiction_values row.
func CreateChange(id []byte, v depictionvalue.Value) rowchange.Change {
	return rowchange.Change{
		EntityType: "depiction_value",
		EntityID:   append([]byte(nil), id...),
		Action:     rowchange.ActionCreate,
		Fields:     rowchange.FullRow(RowMap(id, v)),
	}
}

// LookupTx returns the row with id, or sql.ErrNoRows.
func LookupTx(tx *sql.Tx, id []byte) (depictionvalue.Value, error) {
	if tx == nil || len(id) != 16 {
		return depictionvalue.Value{}, depictionvalue.ErrInvalid
	}
	v := depictionvalue.Value{ID: append([]byte(nil), id...)}
	if err := tx.QueryRow(sqlLookup, id).Scan(&v.RegionJSON, &v.ImageFileID); err != nil {
		return depictionvalue.Value{}, err
	}
	return v, nil
}

// FromJoin turns the columns of a join into the value to show, or nil when the
// Observation has no depiction (imageFileID empty: depiction_values.image_file_id
// is NOT NULL). The caller joins depiction_values on observations.value_depiction_id,
// citations on observations.citation_id (for the locator), and files on
// depiction_values.image_file_id (for the media type). The result has passed
// depictionvalue.Readable and shares no slice with its inputs.
func FromJoin(id []byte, region sql.NullString, imageFileID []byte, citationLocator, imageMediaType sql.NullString) *depictionvalue.Value {
	if len(imageFileID) == 0 {
		return nil
	}
	v := depictionvalue.Readable(depictionvalue.Value{
		ID:          append([]byte(nil), id...),
		RegionJSON:  region.String,
		ImageFileID: append([]byte(nil), imageFileID...),
	}, citationLocator.String, imageMediaType.String)
	return &v
}

// RowMapTx is the release row-map of the depiction_values row id for
// deleteimpact.ReleaseReplaced: it looks the row up when called, then returns
// its audit row.
func RowMapTx(tx *sql.Tx, id []byte) func() (map[string]any, error) {
	return func() (map[string]any, error) {
		v, err := LookupTx(tx, id)
		return RowMap(id, v), err
	}
}
