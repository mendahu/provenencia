// Package depictionvalue is the pure depiction value type: a region of the
// Citation's page and the image File Go stored from it, the row's shape check,
// and the rules a reader applies before showing it. It imports no database
// package; queries live in core/database/depictionvalues.
package depictionvalue

import (
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/locator"
)

var ErrInvalid = apperr.New(apperr.CodeDepictionValuesInvalid, apperr.KindUser)

// Value is one depiction_values row.
type Value struct {
	ID          []byte
	RegionJSON  string
	ImageFileID []byte
}

// Validate reports ErrInvalid when v is not a storable depiction value.
// Callers trim string fields first.
func Validate(v Value) error {
	if v.RegionJSON == "" || len(v.ImageFileID) != 16 {
		return ErrInvalid
	}
	return nil
}

// Readable returns v as it may be shown. RegionJSON is emptied unless it
// passes locator.ValidateDepictionRegion against the locator of the Citation
// that holds the value; ImageFileID is emptied unless the stored File's media
// type is image/png, which is what ingest.StageImage always stores. Every
// reader that returns a stored depiction to the app goes through Readable.
// LookupTx in core/database/depictionvalues does not: audit records what is
// stored.
func Readable(v Value, citationLocatorJSON, imageMediaType string) Value {
	if locator.ValidateDepictionRegion(v.RegionJSON, citationLocatorJSON) != nil {
		v.RegionJSON = ""
	}
	v.ImageFileID = ReadableImage(v.ImageFileID, imageMediaType)
	return v
}

// ReadableImage is the image half of Readable: the File id a reader may hand to
// the app. It returns imageFileID when it is non-empty and the stored File's
// media type is image/png, else nil.
func ReadableImage(imageFileID []byte, mediaType string) []byte {
	if len(imageFileID) == 0 || mediaType != "image/png" {
		return nil
	}
	return imageFileID
}
