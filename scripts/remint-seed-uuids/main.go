// Command remint-seed-uuids rewrites proveniencia seed rows that migrations
// inserted with randomblob(16) (not UUIDv7) so FFI parseID accepts them.
//
// Usage (quit the Mac app first — exclusive catalog lock):
//
//	go run ./scripts/remint-seed-uuids /path/to/family.provenencia
package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: remint-seed-uuids <project.provenencia>\n")
		os.Exit(2)
	}
	dir := os.Args[1]
	c, err := database.Open(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open: %v\n", err)
		os.Exit(1)
	}
	defer c.Close()
	db, err := c.DB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "db: %v\n", err)
		os.Exit(1)
	}
	n, err := remint(db)
	if err != nil {
		fmt.Fprintf(os.Stderr, "remint: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("reminted %d seed id(s) in %s\n", n, dir)
}

func remint(db *sql.DB) (int, error) {
	if _, err := db.Exec(`PRAGMA foreign_keys = OFF`); err != nil {
		return 0, err
	}
	defer func() { _, _ = db.Exec(`PRAGMA foreign_keys = ON`) }()

	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	count := 0
	for _, key := range []string{"place_relationship"} {
		n, err := remintSubjectType(tx, key)
		if err != nil {
			return 0, err
		}
		count += n
	}
	for _, key := range []string{"from", "to", "place_relationship_type", "event_name"} {
		n, err := remintProperty(tx, key)
		if err != nil {
			return 0, err
		}
		count += n
	}
	for _, key := range []string{"part_of", "succeeded_by"} {
		n, err := remintTerm(tx, "place_relationship_type", key)
		if err != nil {
			return 0, err
		}
		count += n
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return count, nil
}

func needsRemint(id []byte) bool {
	if len(id) != 16 {
		return true
	}
	u, err := uuid.FromBytes(id)
	return err != nil || u.Version() != 7
}

func remintSubjectType(tx *sql.Tx, key string) (int, error) {
	var id []byte
	err := tx.QueryRow(
		`SELECT id FROM subject_types WHERE key = ? AND origin = 'provenencia'`, key,
	).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	if !needsRemint(id) {
		return 0, nil
	}
	newID, err := uuid.NewV7()
	if err != nil {
		return 0, err
	}
	nid := newID[:]
	stmts := []string{
		`UPDATE subject_type_properties SET subject_type_id = ? WHERE subject_type_id = ?`,
		`UPDATE subjects SET subject_type_id = ? WHERE subject_type_id = ?`,
		`UPDATE canonical_entities SET subject_type_id = ? WHERE subject_type_id = ?`,
		`UPDATE identity_claims SET subject_type_id = ? WHERE subject_type_id = ?`,
		`UPDATE subject_types SET id = ? WHERE id = ?`,
	}
	for _, q := range stmts {
		if _, err := tx.Exec(q, nid, id); err != nil {
			return 0, fmt.Errorf("%s (%s): %w", key, q, err)
		}
	}
	fmt.Printf("  subject_types.%s %x → %s\n", key, id, newID)
	return 1, nil
}

func remintProperty(tx *sql.Tx, key string) (int, error) {
	var id []byte
	err := tx.QueryRow(
		`SELECT id FROM properties WHERE key = ? AND origin = 'provenencia'`, key,
	).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	if !needsRemint(id) {
		return 0, nil
	}
	newID, err := uuid.NewV7()
	if err != nil {
		return 0, err
	}
	nid := newID[:]
	stmts := []string{
		`UPDATE property_terms SET property_id = ? WHERE property_id = ?`,
		`UPDATE subject_type_properties SET property_id = ? WHERE property_id = ?`,
		`UPDATE observations SET property_id = ? WHERE property_id = ?`,
		`UPDATE auto_reconciler_values SET property_id = ? WHERE property_id = ?`,
		`UPDATE auto_reconciler_outcomes SET property_id = ? WHERE property_id = ?`,
		`UPDATE properties SET id = ? WHERE id = ?`,
	}
	for _, q := range stmts {
		if _, err := tx.Exec(q, nid, id); err != nil {
			return 0, fmt.Errorf("%s (%s): %w", key, q, err)
		}
	}
	fmt.Printf("  properties.%s %x → %s\n", key, id, newID)
	return 1, nil
}

func remintTerm(tx *sql.Tx, propertyKey, termKey string) (int, error) {
	var id []byte
	err := tx.QueryRow(`
		SELECT t.id FROM property_terms t
		JOIN properties p ON p.id = t.property_id
		WHERE p.key = ? AND p.origin = 'provenencia'
		  AND t.key = ? AND t.origin = 'provenencia'`, propertyKey, termKey,
	).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	if !needsRemint(id) {
		return 0, nil
	}
	newID, err := uuid.NewV7()
	if err != nil {
		return 0, err
	}
	nid := newID[:]
	stmts := []string{
		`UPDATE observations SET value_term_id = ? WHERE value_term_id = ?`,
		`UPDATE auto_reconciler_values SET value_term_id = ? WHERE value_term_id = ?`,
		`UPDATE property_terms SET id = ? WHERE id = ?`,
	}
	for _, q := range stmts {
		if _, err := tx.Exec(q, nid, id); err != nil {
			return 0, fmt.Errorf("%s/%s (%s): %w", propertyKey, termKey, q, err)
		}
	}
	fmt.Printf("  property_terms.%s %x → %s\n", termKey, id, newID)
	return 1, nil
}
