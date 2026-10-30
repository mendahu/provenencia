package sources

import (
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/dbtest"
)

// TestNotesAttributionQueryPlan: note create attribution subqueries seek
// audit_changes by entity_id instead of scanning the whole history table.
func TestNotesAttributionQueryPlan(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	noteID := make([]byte, 16)
	noteID[15] = 1
	sourceID := make([]byte, 16)
	sourceID[15] = 2
	for _, tc := range []struct {
		name string
		q    string
		bind []byte
	}{
		{"list", sqlListNotes, sourceID},
		{"get", sqlGetNoteAttributed, noteID},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, detail := range dbtest.QueryPlan(t, db, tc.q, tc.bind) {
				if detail == "SCAN c" || strings.HasPrefix(detail, "SCAN c ") {
					t.Errorf("query plan scans audit_changes alias c: %s", detail)
				}
			}
		})
	}
}
