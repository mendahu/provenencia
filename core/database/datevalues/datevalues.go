// Package datevalues stores shared genealogical DateValue rows.
//
// Dogfood subset validated by Insert (full DDL remains flexible for later kinds):
//
//   - kind "exact": start_year/month/day required; qualifier "" or "ABT"
//   - kind "year":  start_year only; qualifier "" or "ABT"
//   - kind "range": start_year + end_year (start <= end); qualifier empty only
package datevalues

import (
	"database/sql"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
)

var ErrInvalid = apperr.New(apperr.CodeDateValuesInvalid, apperr.KindUser)

const (
	KindExact = "exact"
	KindYear  = "year"
	KindRange = "range"

	QualifierABT = "ABT"

	sqlInsert = `INSERT INTO date_values (
		id, kind, qualifier, calendar,
		start_year, start_month, start_day,
		end_year, end_month, end_day,
		phrase
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	sqlLookup = `SELECT kind, qualifier, calendar,
		start_year, start_month, start_day,
		end_year, end_month, end_day,
		phrase
		FROM date_values WHERE id = ?`
)

// Value is one date_values row.
type Value struct {
	ID         []byte
	Kind       string
	Qualifier  string // "" or QualifierABT for dogfood kinds
	Calendar   string
	StartYear  *int
	StartMonth *int
	StartDay   *int
	EndYear    *int
	EndMonth   *int
	EndDay     *int
	Phrase     string
}

// Insert mints a UUIDv7 id, validates the dogfood subset, inserts, and returns the id.
func Insert(c *database.Catalog, v Value) ([]byte, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	v.Kind = strings.TrimSpace(v.Kind)
	v.Qualifier = strings.TrimSpace(v.Qualifier)
	v.Calendar = strings.TrimSpace(v.Calendar)
	v.Phrase = strings.TrimSpace(v.Phrase)
	if err := validate(v); err != nil {
		return nil, err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	qual := nullIfEmpty(v.Qualifier)
	cal := nullIfEmpty(v.Calendar)
	phrase := nullIfEmpty(v.Phrase)
	_, err = db.Exec(
		sqlInsert,
		id[:],
		v.Kind,
		qual,
		cal,
		nullInt(v.StartYear),
		nullInt(v.StartMonth),
		nullInt(v.StartDay),
		nullInt(v.EndYear),
		nullInt(v.EndMonth),
		nullInt(v.EndDay),
		phrase,
	)
	if err != nil {
		return nil, err
	}
	return id[:], nil
}

// Lookup returns the date_values row for id, or sql.ErrNoRows.
func Lookup(c *database.Catalog, id []byte) (Value, error) {
	db, err := c.DB()
	if err != nil {
		return Value{}, err
	}
	if len(id) != 16 {
		return Value{}, ErrInvalid
	}
	var (
		v                                        Value
		qual, cal, phrase                        sql.NullString
		startY, startM, startD, endY, endM, endD sql.NullInt64
	)
	v.ID = append([]byte(nil), id...)
	err = db.QueryRow(sqlLookup, id).Scan(
		&v.Kind,
		&qual,
		&cal,
		&startY, &startM, &startD,
		&endY, &endM, &endD,
		&phrase,
	)
	if err != nil {
		return Value{}, err
	}
	v.Qualifier = qual.String
	v.Calendar = cal.String
	v.Phrase = phrase.String
	v.StartYear = intPtr(startY)
	v.StartMonth = intPtr(startM)
	v.StartDay = intPtr(startD)
	v.EndYear = intPtr(endY)
	v.EndMonth = intPtr(endM)
	v.EndDay = intPtr(endD)
	return v, nil
}

func validate(v Value) error {
	qual := strings.TrimSpace(v.Qualifier)
	switch v.Kind {
	case KindExact:
		if qual != "" && qual != QualifierABT {
			return ErrInvalid
		}
		if v.StartYear == nil || v.StartMonth == nil || v.StartDay == nil {
			return ErrInvalid
		}
		if !monthOK(*v.StartMonth) || !dayOK(*v.StartDay) {
			return ErrInvalid
		}
		if v.EndYear != nil || v.EndMonth != nil || v.EndDay != nil {
			return ErrInvalid
		}
	case KindYear:
		if qual != "" && qual != QualifierABT {
			return ErrInvalid
		}
		if v.StartYear == nil || v.StartMonth != nil || v.StartDay != nil {
			return ErrInvalid
		}
		if v.EndYear != nil || v.EndMonth != nil || v.EndDay != nil {
			return ErrInvalid
		}
	case KindRange:
		if qual != "" {
			return ErrInvalid
		}
		if v.StartYear == nil || v.EndYear == nil {
			return ErrInvalid
		}
		if *v.EndYear < *v.StartYear {
			return ErrInvalid
		}
		if v.StartMonth != nil && !monthOK(*v.StartMonth) {
			return ErrInvalid
		}
		if v.EndMonth != nil && !monthOK(*v.EndMonth) {
			return ErrInvalid
		}
		if v.StartDay != nil && !dayOK(*v.StartDay) {
			return ErrInvalid
		}
		if v.EndDay != nil && !dayOK(*v.EndDay) {
			return ErrInvalid
		}
		// Dogfood simple range: year bounds only (no month/day on either side).
		if v.StartMonth != nil || v.StartDay != nil || v.EndMonth != nil || v.EndDay != nil {
			return ErrInvalid
		}
	default:
		return ErrInvalid
	}
	return nil
}

func monthOK(m int) bool { return m >= 1 && m <= 12 }
func dayOK(d int) bool   { return d >= 1 && d <= 31 }

func nullIfEmpty(s string) any {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return s
}

func nullInt(p *int) any {
	if p == nil {
		return nil
	}
	return *p
}

func intPtr(n sql.NullInt64) *int {
	if !n.Valid {
		return nil
	}
	v := int(n.Int64)
	return &v
}
