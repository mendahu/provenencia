// Package artifactstest holds shared Artifact fixtures for tests: creating an
// Artifact, splitting an Evidence graph off a Source, and putting a Subject on
// a graph.
package artifactstest

import (
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/evidencegraphs"
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
)

// Create makes an Artifact through the write pipeline.
func Create(c *database.Catalog, userID []byte, in artifacts.CreateInput) (artifacts.Artifact, error) {
	a, _, err := writes.Run(c, writes.Op{Action: "create_artifact", UserID: userID},
		func(tx *database.Tx) (artifacts.Artifact, []rowchange.Change, error) {
			return artifacts.Create(tx, userID, in)
		})
	return a, err
}

// Split creates an Artifact labelled name on the Source, then moves it out of
// the Source's Default graph into a new graph of the same name. It returns both
// graphs and the Artifact. The Artifact is made first because a graph cannot
// lose its last link: the Default graph must still hold another Artifact.
func Split(t testing.TB, c *database.Catalog, userID, sourceID []byte, name string) (def, graph evidencegraphs.Graph, page artifacts.Artifact) {
	t.Helper()
	page, err := Create(c, userID, artifacts.CreateInput{SourceID: sourceID, Label: name})
	if err != nil {
		t.Fatal(err)
	}
	def, graph = evidencegraphstest.Split(t, c, userID, sourceID, name, page.ID)
	return def, graph, page
}

// SubjectOn creates a Subject of the provenencia type typeKey on the graph.
func SubjectOn(t testing.TB, c *database.Catalog, userID, sourceID, graphID []byte, typeKey, label string) subjects.Subject {
	t.Helper()
	typ, err := subjecttypes.Lookup(c, typeKey, vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	s, err := writes.Call(c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (subjects.Subject, []rowchange.Change, error) {
		return subjects.Create(tx, userID, subjects.CreateInput{SourceID: sourceID, GraphID: graphID, SubjectTypeID: typ.ID, Label: label}, nil)
	})
	if err != nil {
		t.Fatal(err)
	}
	return s
}
