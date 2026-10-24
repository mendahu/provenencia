// Package subjectstest holds shared Subject fixtures for tests: creating one
// Subject on an Evidence graph in its own write.
package subjectstest

import (
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
)

// On creates a Subject of the provenencia type typeKey on the graph.
func On(t testing.TB, c *database.Catalog, userID, graphID []byte, typeKey, label string) subjects.Subject {
	t.Helper()
	typ, err := subjecttypes.Lookup(c, typeKey, vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	return OfType(t, c, userID, graphID, typ.ID, label)
}

// OfType creates a Subject of the Subject type subjectTypeID on the graph, for
// a caller that already holds the type id.
func OfType(t testing.TB, c *database.Catalog, userID, graphID, subjectTypeID []byte, label string) subjects.Subject {
	t.Helper()
	return create(t, c, userID, graphID, subjectTypeID, label, nil)
}

// At is OfType for a Subject created with a Placement.
func At(t testing.TB, c *database.Catalog, userID, graphID, subjectTypeID []byte, label string, at subjects.Placement) subjects.Subject {
	t.Helper()
	return create(t, c, userID, graphID, subjectTypeID, label, &at)
}

func create(t testing.TB, c *database.Catalog, userID, graphID, subjectTypeID []byte, label string, at *subjects.Placement) subjects.Subject {
	t.Helper()
	s, err := writes.Call(c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (subjects.Subject, []rowchange.Change, error) {
		return subjects.Create(tx, userID, subjects.CreateInput{GraphID: graphID, SubjectTypeID: subjectTypeID, Label: label}, at)
	})
	if err != nil {
		t.Fatal(err)
	}
	return s
}
