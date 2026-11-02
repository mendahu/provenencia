package observations

import (
	"database/sql"
	"strings"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/depictionvalues"
	"github.com/mendahu/provenencia/core/database/files"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/depictionvalue"
	"github.com/mendahu/provenencia/core/locator"
)

// StagedFile is an image already checked and re-encoded by Go and not yet
// stored. *ingest.Staged satisfies it. The write calls InsertTx inside its
// own transaction; every decode and disk write happened before the write began.
type StagedFile interface {
	InsertTx(tx *sql.Tx) (files.File, []rowchange.Change, error)
}

// DepictionInput is the value of a depiction Property: a region of the
// Citation's page and the image of it. A caller sets
// RegionJSON and ImagePath; ingest.StageDepictions (withStagedDepictions, before the catalog session) reads ImagePath and sets
// Image. The write never reads ImagePath.
type DepictionInput struct {
	RegionJSON string
	ImagePath  string
	Image      StagedFile
}

const sqlCitationArtifactFile = `SELECT a.file_id, c.locator_json
	FROM citations c JOIN artifacts a ON a.id = c.artifact_id
	WHERE c.id = ?`

// resolveDepiction stores a depiction value for a Citation: it checks the
// region against the Citation's locator, stores the staged image, and inserts
// the value row.
func resolveDepiction(tx *sql.Tx, citationID []byte, d *DepictionInput) (resolvedValue, error) {
	if d == nil || d.Image == nil {
		return resolvedValue{}, ErrInvalid
	}
	var (
		fileID      []byte
		locatorJSON string
	)
	if err := tx.QueryRow(sqlCitationArtifactFile, citationID).Scan(&fileID, &locatorJSON); err != nil {
		if err == sql.ErrNoRows {
			return resolvedValue{}, ErrInvalid
		}
		return resolvedValue{}, err
	}
	if len(fileID) != 16 {
		return resolvedValue{}, ErrInvalid
	}
	d.RegionJSON = strings.TrimSpace(d.RegionJSON)
	if err := locator.ValidateDepictionRegion(d.RegionJSON, locatorJSON); err != nil {
		return resolvedValue{}, err
	}
	file, fileChanges, err := d.Image.InsertTx(tx)
	if err != nil {
		return resolvedValue{}, err
	}
	v := depictionvalue.Value{RegionJSON: d.RegionJSON, ImageFileID: file.ID}
	id, err := depictionvalues.InsertTx(tx, v)
	if err != nil {
		return resolvedValue{}, err
	}
	changes := append(fileChanges, depictionCreateChange(id, v))
	return resolvedValue{DepictionID: id, ValueChanges: changes}, nil
}

func depictionCreateChange(id []byte, v depictionvalue.Value) rowchange.Change {
	return rowchange.Change{
		EntityType: "depiction_value",
		EntityID:   append([]byte(nil), id...),
		Action:     rowchange.ActionCreate,
		Fields:     rowchange.FullRow(depictionValueMap(id, v)),
	}
}

func depictionValueMap(id []byte, v depictionvalue.Value) map[string]any {
	return map[string]any{
		"id":            database.UUIDJSON(id),
		"region_json":   v.RegionJSON,
		"image_file_id": database.UUIDJSON(v.ImageFileID),
	}
}
