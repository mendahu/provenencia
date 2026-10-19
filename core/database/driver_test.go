package database

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestUUID7Function(t *testing.T) {
	db, err := openDBURI(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var a, b []byte
	if err := db.QueryRow(`SELECT uuid7(), uuid7()`).Scan(&a, &b); err != nil {
		t.Fatal(err)
	}
	if string(a) == string(b) {
		t.Fatalf("two calls returned the same id %x", a)
	}
	for _, raw := range [][]byte{a, b} {
		id, err := uuid.FromBytes(raw)
		if err != nil {
			t.Fatalf("%x: %v", raw, err)
		}
		if id.Version() != 7 || id.Variant() != uuid.RFC4122 {
			t.Fatalf("%s: version %d variant %v", id, id.Version(), id.Variant())
		}
	}
}

func TestSchemaCallsNoAppFunction(t *testing.T) {
	c, err := Create(t.TempDir(), "Schema.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	rows, err := db.Query(`SELECT name, sql FROM sqlite_schema WHERE sql IS NOT NULL`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var name, sql string
		if err := rows.Scan(&name, &sql); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(strings.ToLower(sql), "uuid7(") {
			t.Errorf("%s calls uuid7() in the stored schema", name)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
}
