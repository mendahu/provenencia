// Package canonicalentities accesses the canonical_entities catalog table:
// Conclusion handles (Persons, Events, Places, …) with audited mutations.
package canonicalentities

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/audit"
	"github.com/mendahu/provenencia/core/database/project"
	"github.com/mendahu/provenencia/core/ref"
)

var ErrInvalid = apperr.New(apperr.CodeCanonicalEntitiesInvalid, apperr.KindUser)

const (
	sqlInsert = `INSERT INTO canonical_entities (id, subject_type_id, ref, argument, label)
		VALUES (?, ?, ?, ?, ?)`
	sqlColumns    = `id, subject_type_id, ref, COALESCE(argument, ''), COALESCE(label, ''), merged_into_id`
	sqlGet        = `SELECT ` + sqlColumns + ` FROM canonical_entities WHERE id = ?`
	sqlGetByRef   = `SELECT ` + sqlColumns + ` FROM canonical_entities WHERE ref = ?`
	sqlListByType = `SELECT ` + sqlColumns + ` FROM canonical_entities
		WHERE subject_type_id = ?
		ORDER BY ref COLLATE NOCASE`
	sqlTypePrefix = `SELECT ref_prefix FROM subject_types WHERE id = ?`
	// Merged handles are folded into their target and never counted.
	sqlCountByTypeKey = `SELECT COUNT(*) FROM canonical_entities e
		JOIN subject_types st ON st.id = e.subject_type_id
		WHERE st.key = ? AND st.origin = 'provenencia' AND e.merged_into_id IS NULL`
	maxRefRetries = 8
)

// Entity is one canonical_entities row.
type Entity struct {
	ID            []byte
	SubjectTypeID []byte
	Ref           string
	Argument      string
	Label         string
	MergedIntoID  []byte // nil unless merged
}

// CreateInput is the fields for a new handle. Subject type is immutable after insert.
type CreateInput struct {
	SubjectTypeID []byte
	Label         string
	Argument      string
}

// Create inserts a handle, mints a ref from the type's ref_prefix, and records create_canonical_entity.
func Create(c *database.Catalog, userID []byte, in CreateInput) (Entity, error) {
	db, err := c.DB()
	if err != nil {
		return Entity{}, err
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return Entity{}, err
	}

	tx, err := db.Begin()
	if err != nil {
		return Entity{}, err
	}
	defer func() { _ = tx.Rollback() }()

	e, change, err := InsertTx(tx, in)
	if err != nil {
		return Entity{}, err
	}
	if _, err := audit.Record(tx, audit.Revision{
		UserID:     userID,
		ActionType: "create_canonical_entity",
		CreatedAt:  project.NowUTC(),
		Changes:    []audit.Change{change},
	}); err != nil {
		return Entity{}, err
	}
	if err := tx.Commit(); err != nil {
		return Entity{}, err
	}
	return e, nil
}

// InsertTx inserts a handle on an open transaction (no commit, no revision).
// The ref is minted from subject_types.ref_prefix (PER-…), never the candidate prefix.
func InsertTx(tx *sql.Tx, in CreateInput) (Entity, audit.Change, error) {
	in.Label = strings.TrimSpace(in.Label)
	in.Argument = strings.TrimSpace(in.Argument)
	if len(in.SubjectTypeID) != 16 {
		return Entity{}, audit.Change{}, ErrInvalid
	}
	prefix, err := requireTypePrefix(tx, in.SubjectTypeID)
	if err != nil {
		return Entity{}, audit.Change{}, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return Entity{}, audit.Change{}, err
	}
	idBytes := id[:]

	var entityRef string
	for attempt := 0; attempt < maxRefRetries; attempt++ {
		entityRef, err = ref.Mint(prefix)
		if err != nil {
			return Entity{}, audit.Change{}, err
		}
		_, err = tx.Exec(sqlInsert, idBytes, in.SubjectTypeID, entityRef, nullStr(in.Argument), nullStr(in.Label))
		if err == nil {
			break
		}
		if !database.IsUniqueConflict(err) {
			return Entity{}, audit.Change{}, mapConstraint(err)
		}
	}
	if err != nil {
		return Entity{}, audit.Change{}, ErrInvalid
	}

	change := audit.Change{
		EntityType: "canonical_entity",
		EntityID:   idBytes,
		Action:     audit.ActionCreate,
		Fields: audit.FullRow(map[string]any{
			"id":              id.String(),
			"subject_type_id": uuidJSON(in.SubjectTypeID),
			"ref":             entityRef,
			"argument":        nullStr(in.Argument),
			"label":           nullStr(in.Label),
			"merged_into_id":  nil,
		}),
	}
	return Entity{
		ID:            append([]byte(nil), idBytes...),
		SubjectTypeID: append([]byte(nil), in.SubjectTypeID...),
		Ref:           entityRef,
		Argument:      in.Argument,
		Label:         in.Label,
	}, change, nil
}

// Get returns a handle by id, or sql.ErrNoRows.
func Get(c *database.Catalog, id []byte) (Entity, error) {
	db, err := c.DB()
	if err != nil {
		return Entity{}, err
	}
	if len(id) != 16 {
		return Entity{}, ErrInvalid
	}
	return scanEntity(db.QueryRow(sqlGet, id))
}

// GetTx returns a handle by id on an open transaction, or sql.ErrNoRows.
func GetTx(tx *sql.Tx, id []byte) (Entity, error) {
	if len(id) != 16 {
		return Entity{}, ErrInvalid
	}
	return scanEntity(tx.QueryRow(sqlGet, id))
}

// GetByRef returns a handle by ref (PER-…), or sql.ErrNoRows.
func GetByRef(c *database.Catalog, entityRef string) (Entity, error) {
	db, err := c.DB()
	if err != nil {
		return Entity{}, err
	}
	entityRef = strings.TrimSpace(entityRef)
	if ref.Validate(entityRef) != nil {
		return Entity{}, ErrInvalid
	}
	return scanEntity(db.QueryRow(sqlGetByRef, entityRef))
}

// ListByType returns every handle of one Subject type, ordered by ref.
func ListByType(c *database.Catalog, subjectTypeID []byte) ([]Entity, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	if len(subjectTypeID) != 16 {
		return nil, ErrInvalid
	}
	rows, err := db.Query(sqlListByType, subjectTypeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Entity
	for rows.Next() {
		e, err := scanEntity(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// CountByTypeKey counts unmerged handles of a product Subject type (person,
// event, place).
func CountByTypeKey(c *database.Catalog, typeKey string) (int, error) {
	db, err := c.DB()
	if err != nil {
		return 0, err
	}
	typeKey = strings.TrimSpace(typeKey)
	if typeKey == "" {
		return 0, ErrInvalid
	}
	var n int
	err = db.QueryRow(sqlCountByTypeKey, typeKey).Scan(&n)
	return n, err
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanEntity(row rowScanner) (Entity, error) {
	var e Entity
	if err := row.Scan(&e.ID, &e.SubjectTypeID, &e.Ref, &e.Argument, &e.Label, &e.MergedIntoID); err != nil {
		return Entity{}, err
	}
	return e, nil
}

func requireTypePrefix(tx *sql.Tx, typeID []byte) (string, error) {
	var prefix string
	err := tx.QueryRow(sqlTypePrefix, typeID).Scan(&prefix)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrInvalid
	}
	if err != nil {
		return "", err
	}
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return "", ErrInvalid
	}
	return prefix, nil
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func uuidJSON(id []byte) any {
	if len(id) != 16 {
		return nil
	}
	u, err := uuid.FromBytes(id)
	if err != nil {
		return nil
	}
	return u.String()
}

func mapConstraint(err error) error {
	if database.IsConstraintViolation(err) {
		return ErrInvalid
	}
	return err
}
