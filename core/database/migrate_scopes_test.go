package database

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

// TestMigrateBackfillsAuditScopes seeds a v35 catalog with audit history and
// checks 000036 scopes each revision to the Source it worked on.
func TestMigrateBackfillsAuditScopes(t *testing.T) {
	id := func(b byte) []byte { return bytes.Repeat([]byte{b}, 16) }
	uid := func(b []byte) string { u, _ := uuid.FromBytes(b); return u.String() }
	typeID, srcA, srcB, art, cit, file := id(1), id(2), id(3), id(4), id(5), id(6)
	goneSubject, goneObs, goneNote, sourceNote, claim := id(7), id(8), id(9), id(10), id(11)

	db, err := openDBURI(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, step := range migrations {
		if step.to > 35 {
			break
		}
		if _, err := db.Exec(step.sql); err != nil {
			t.Fatalf("step %d: %v", step.to, err)
		}
		if _, err := db.Exec(fmt.Sprintf(`PRAGMA user_version = %d`, step.to)); err != nil {
			t.Fatal(err)
		}
	}

	exec := func(q string, args ...any) {
		t.Helper()
		if _, err := db.Exec(q, args...); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
	exec(`INSERT INTO source_types (id, key, origin, label) VALUES (?, 'book', 'test', 'Book')`, typeID)
	exec(`INSERT INTO sources (id, ref, source_type_id, title) VALUES (?, 'SRC-A', ?, 'A')`, srcA, typeID)
	exec(`INSERT INTO sources (id, ref, source_type_id, title) VALUES (?, 'SRC-B', ?, 'B')`, srcB, typeID)
	exec(`INSERT INTO files (id, checksum_sha256, byte_size) VALUES (?, 'abc', 1)`, file)
	exec(`INSERT INTO artifacts (id, ref, source_id, file_id) VALUES (?, 'ART-A', ?, ?)`, art, srcA, file)
	exec(`INSERT INTO citations (id, ref, artifact_id, locator_json) VALUES (?, 'CIT-A', ?, '{}')`, cit, art)

	type change struct {
		entityType string
		entityID   []byte
		action     string
		fields     map[string]any // field → [old, new]
	}
	rev := int64(0)
	record := func(cs ...change) int64 {
		t.Helper()
		rev++
		txID := id(byte(100 + rev))
		exec(`INSERT INTO audit_transactions (id, revision, created_at) VALUES (?, ?, '2026-10-04T00:00:00Z')`, txID, rev)
		for i, c := range cs {
			fields := map[string]map[string]any{}
			for k, v := range c.fields {
				pair := v.([2]any)
				fields[k] = map[string]any{"old": pair[0], "new": pair[1]}
			}
			raw, err := json.Marshal(fields)
			if err != nil {
				t.Fatal(err)
			}
			exec(`INSERT INTO audit_changes (id, audit_transaction_id, entity_type, entity_id, action, changes_json)
				VALUES (?, ?, ?, ?, ?, ?)`, id(byte(200+int(rev)*4+i)), txID, c.entityType, c.entityID, c.action, string(raw))
		}
		return rev
	}

	cases := []struct {
		name string
		rev  int64
		want [][]byte
	}{
		{"source create", record(change{"source", srcA, "create", map[string]any{"title": [2]any{nil, "A"}}}), [][]byte{srcA}},
		{"note create via field", record(change{"source_note", sourceNote, "create",
			map[string]any{"source_id": [2]any{nil, uid(srcA)}}}), [][]byte{srcA}},
		{"deleted subject via old field", record(change{"subject", goneSubject, "delete",
			map[string]any{"source_id": [2]any{uid(srcB), nil}}}), [][]byte{srcB}},
		{"citation update via live row", record(change{"citation", cit, "update",
			map[string]any{"transcription": [2]any{nil, "x"}}}), [][]byte{srcA}},
		{"deleted observation via citation field", record(change{"observation", goneObs, "delete",
			map[string]any{"citation_id": [2]any{uid(cit), nil}}}), [][]byte{srcA}},
		{"deleted citation note via citation field", record(change{"citation_note", goneNote, "delete",
			map[string]any{"citation_id": [2]any{uid(cit), nil}}}), [][]byte{srcA}},
		{"file via artifact", record(change{"file", file, "update",
			map[string]any{"original_filename": [2]any{"a", "b"}}}), [][]byte{srcA}},
		{"conclusion layer unscoped", record(change{"identity_claim", claim, "create",
			map[string]any{"status": [2]any{nil, "accepted"}}}), nil},
	}

	if err := migrate(db); err != nil {
		t.Fatal(err)
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rows, err := db.Query(`SELECT s.scope_id FROM audit_transaction_scopes s
				JOIN audit_transactions t ON t.id = s.audit_transaction_id
				WHERE t.revision = ? AND s.scope_type = 'source'`, tc.rev)
			if err != nil {
				t.Fatal(err)
			}
			defer rows.Close()
			var got [][]byte
			for rows.Next() {
				var sid []byte
				if err := rows.Scan(&sid); err != nil {
					t.Fatal(err)
				}
				got = append(got, sid)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("got %x want %x", got, tc.want)
			}
			for i := range got {
				if !bytes.Equal(got[i], tc.want[i]) {
					t.Fatalf("got %x want %x", got, tc.want)
				}
			}
		})
	}
}
