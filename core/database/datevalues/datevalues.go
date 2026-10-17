// Package datevalues stores shared genealogical DateValue rows. The value
// type, its kinds and qualifiers, validation and display live in the pure
// package core/datevalue; this package holds the queries.
package datevalues

import (
	"database/sql"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/sqlbatch"
	"github.com/mendahu/provenencia/core/datevalue"
)

const (
	sqlInsert = `INSERT INTO date_values (
		id, kind, qualifier, calendar,
		start_year, start_month, start_day,
		start_hour, start_minute, start_second, start_millisecond, start_tz,
		end_year, end_month, end_day,
		end_hour, end_minute, end_second, end_millisecond, end_tz,
		phrase
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	sqlUpdate = `UPDATE date_values SET
		kind = ?, qualifier = ?, calendar = ?,
		start_year = ?, start_month = ?, start_day = ?,
		start_hour = ?, start_minute = ?, start_second = ?, start_millisecond = ?, start_tz = ?,
		end_year = ?, end_month = ?, end_day = ?,
		end_hour = ?, end_minute = ?, end_second = ?, end_millisecond = ?, end_tz = ?,
		phrase = ?
		WHERE id = ?`

	sqlExists = `SELECT 1 FROM date_values WHERE id = ?`

	sqlLookup = `SELECT kind, qualifier, calendar,
		start_year, start_month, start_day,
		start_hour, start_minute, start_second, start_millisecond, start_tz,
		end_year, end_month, end_day,
		end_hour, end_minute, end_second, end_millisecond, end_tz,
		phrase
		FROM date_values WHERE id = ?`

	sqlLookupMany = `SELECT id, kind, qualifier, calendar,
		start_year, start_month, start_day,
		start_hour, start_minute, start_second, start_millisecond, start_tz,
		end_year, end_month, end_day,
		end_hour, end_minute, end_second, end_millisecond, end_tz,
		phrase
		FROM date_values WHERE id IN (`
)

// Insert mints a UUIDv7 id, validates, inserts, and returns the id.
func Insert(c *database.Catalog, v datevalue.Value) ([]byte, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	return InsertTx(db, v)
}

// InsertTx validates and inserts on an existing connection or transaction.
func InsertTx(q interface {
	Exec(query string, args ...any) (sql.Result, error)
}, v datevalue.Value) ([]byte, error) {
	v.Kind = strings.TrimSpace(v.Kind)
	v.Qualifier = strings.TrimSpace(v.Qualifier)
	v.Calendar = strings.TrimSpace(v.Calendar)
	v.Phrase = strings.TrimSpace(v.Phrase)
	v.StartTZ = strings.TrimSpace(v.StartTZ)
	v.EndTZ = strings.TrimSpace(v.EndTZ)
	if err := datevalue.Validate(v); err != nil {
		return nil, err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	_, err = q.Exec(
		sqlInsert,
		id[:],
		v.Kind,
		nullIfEmpty(v.Qualifier),
		nullIfEmpty(v.Calendar),
		nullInt(v.StartYear),
		nullInt(v.StartMonth),
		nullInt(v.StartDay),
		nullInt(v.StartHour),
		nullInt(v.StartMinute),
		nullInt(v.StartSecond),
		nullInt(v.StartMillisecond),
		nullIfEmpty(v.StartTZ),
		nullInt(v.EndYear),
		nullInt(v.EndMonth),
		nullInt(v.EndDay),
		nullInt(v.EndHour),
		nullInt(v.EndMinute),
		nullInt(v.EndSecond),
		nullInt(v.EndMillisecond),
		nullIfEmpty(v.EndTZ),
		nullIfEmpty(v.Phrase),
	)
	if err != nil {
		return nil, err
	}
	return id[:], nil
}

// UpdateTx rewrites the date_values row id. The id stays the same.
func UpdateTx(tx *sql.Tx, id []byte, v datevalue.Value) error {
	if tx == nil || len(id) != 16 {
		return datevalue.ErrInvalid
	}
	v.Kind = strings.TrimSpace(v.Kind)
	v.Qualifier = strings.TrimSpace(v.Qualifier)
	v.Calendar = strings.TrimSpace(v.Calendar)
	v.Phrase = strings.TrimSpace(v.Phrase)
	v.StartTZ = strings.TrimSpace(v.StartTZ)
	v.EndTZ = strings.TrimSpace(v.EndTZ)
	if err := datevalue.Validate(v); err != nil {
		return err
	}
	var one int
	if err := tx.QueryRow(sqlExists, id).Scan(&one); err != nil {
		if err == sql.ErrNoRows {
			return datevalue.ErrInvalid
		}
		return err
	}
	_, err := tx.Exec(
		sqlUpdate,
		v.Kind,
		nullIfEmpty(v.Qualifier),
		nullIfEmpty(v.Calendar),
		nullInt(v.StartYear),
		nullInt(v.StartMonth),
		nullInt(v.StartDay),
		nullInt(v.StartHour),
		nullInt(v.StartMinute),
		nullInt(v.StartSecond),
		nullInt(v.StartMillisecond),
		nullIfEmpty(v.StartTZ),
		nullInt(v.EndYear),
		nullInt(v.EndMonth),
		nullInt(v.EndDay),
		nullInt(v.EndHour),
		nullInt(v.EndMinute),
		nullInt(v.EndSecond),
		nullInt(v.EndMillisecond),
		nullIfEmpty(v.EndTZ),
		nullIfEmpty(v.Phrase),
		id,
	)
	return err
}

// Lookup returns the date_values row for id, or sql.ErrNoRows.
func Lookup(c *database.Catalog, id []byte) (datevalue.Value, error) {
	db, err := c.DB()
	if err != nil {
		return datevalue.Value{}, err
	}
	return LookupTx(db, id)
}

// LookupManyTx returns date_values rows for the given ids, one query per
// sqlbatch.Size ids, keyed by string(id). Missing ids are omitted.
func LookupManyTx(q interface {
	Query(query string, args ...any) (*sql.Rows, error)
}, ids [][]byte) (map[string]datevalue.Value, error) {
	ids = database.UniqueBlobIDs(ids)
	out := make(map[string]datevalue.Value, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	err := sqlbatch.Query(q, sqlLookupMany, `)`, nil, ids, func(rows *sql.Rows) error {
		var id []byte
		v, err := scanValue(rows, &id)
		if err != nil {
			return err
		}
		v.ID = append([]byte(nil), id...)
		out[string(v.ID)] = v
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// LookupTx returns the date_values row for id on an existing connection or transaction.
func LookupTx(q interface {
	QueryRow(query string, args ...any) *sql.Row
}, id []byte) (datevalue.Value, error) {
	if q == nil || len(id) != 16 {
		return datevalue.Value{}, datevalue.ErrInvalid
	}
	v, err := scanValue(q.QueryRow(sqlLookup, id))
	if err != nil {
		return datevalue.Value{}, err
	}
	v.ID = append([]byte(nil), id...)
	return v, nil
}

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

// scanValue reads the sqlLookup column list, after any leading columns.
func scanValue(row interface{ Scan(dest ...any) error }, lead ...any) (datevalue.Value, error) {
	var (
		v                                                         datevalue.Value
		qual, cal, phrase, startTZ, endTZ                         sql.NullString
		startY, startM, startD, startH, startMin, startS, startMs sql.NullInt64
		endY, endM, endD, endH, endMin, endS, endMs               sql.NullInt64
	)
	dest := append(lead,
		&v.Kind,
		&qual,
		&cal,
		&startY, &startM, &startD,
		&startH, &startMin, &startS, &startMs, &startTZ,
		&endY, &endM, &endD,
		&endH, &endMin, &endS, &endMs, &endTZ,
		&phrase,
	)
	if err := row.Scan(dest...); err != nil {
		return datevalue.Value{}, err
	}
	v.Qualifier = qual.String
	v.Calendar = cal.String
	v.Phrase = phrase.String
	v.StartTZ = startTZ.String
	v.EndTZ = endTZ.String
	v.StartYear = intPtr(startY)
	v.StartMonth = intPtr(startM)
	v.StartDay = intPtr(startD)
	v.StartHour = intPtr(startH)
	v.StartMinute = intPtr(startMin)
	v.StartSecond = intPtr(startS)
	v.StartMillisecond = intPtr(startMs)
	v.EndYear = intPtr(endY)
	v.EndMonth = intPtr(endM)
	v.EndDay = intPtr(endD)
	v.EndHour = intPtr(endH)
	v.EndMinute = intPtr(endMin)
	v.EndSecond = intPtr(endS)
	v.EndMillisecond = intPtr(endMs)
	return v, nil
}
