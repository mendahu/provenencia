// Package scalarcols scans the value_text, value_integer and value_boolean
// columns every value-carrying table shares into a propertymeta.Scalar.
package scalarcols

import (
	"database/sql"

	"github.com/mendahu/provenencia/core/propertymeta"
)

// Columns is the three scalar value columns as a row scan fills them. Scan
// the integer and boolean columns as NullInt64: a stored boolean outside 0 and
// 1 is then read as absent instead of failing the whole read.
type Columns struct {
	Text    sql.NullString
	Integer sql.NullInt64
	Boolean sql.NullInt64
}

// Scalar is the scanned columns as a Scalar. HasBoolean is true only for 0 or 1.
func (c Columns) Scalar() propertymeta.Scalar {
	return propertymeta.Scalar{
		Text: c.Text.String, HasText: c.Text.Valid,
		Integer: c.Integer.Int64, HasInteger: c.Integer.Valid,
		Boolean: c.Boolean.Int64 == 1, HasBoolean: c.Boolean.Valid && (c.Boolean.Int64 == 0 || c.Boolean.Int64 == 1),
	}
}
