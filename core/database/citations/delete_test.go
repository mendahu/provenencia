package citations_test

import (
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/connect"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/writes"
)

const testLocator = `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`

func TestCitationDeleteConnectionOnly(t *testing.T) {
	userID := []byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}

	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	r, err := ref.Mint(ref.PrefixUser)
	if err != nil {
		t.Fatal(err)
	}
	if err := users.Upsert(c, userID, "Jake", r); err != nil {
		t.Fatal(err)
	}
	if err := subjectvocab.Install(c); err != nil {
		t.Fatal(err)
	}
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
		Key: "book", Origin: sourcetypes.OriginProvenencia, Label: "Book",
	})
	if err != nil {
		t.Fatal(err)
	}
	src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, userID, sources.CreateInput{
			SourceTypeID: typeID, Title: "Register",
		})
	})
	if err != nil {
		t.Fatal(err)
	}
	art, err := runArtifactCreate(c, userID, artifacts.CreateInput{
		SourceID: src.ID, Label: "Scan",
	})
	if err != nil {
		t.Fatal(err)
	}
	personType, err := subjecttypes.Lookup(c, "person", subjecttypes.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	eventType, err := subjecttypes.Lookup(c, "event", subjecttypes.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	alice, err := writes.Call(c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (subjects.Subject, []rowchange.Change, error) {
		return subjects.Create(tx, userID, subjects.CreateInput{
			SourceID: src.ID, SubjectTypeID: personType.ID, Label: "Alice",
		}, &subjects.Placement{GridX: 0, GridY: 0})
	})
	if err != nil {
		t.Fatal(err)
	}
	wedding, err := writes.Call(c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (subjects.Subject, []rowchange.Change, error) {
		return subjects.Create(tx, userID, subjects.CreateInput{
			SourceID: src.ID, SubjectTypeID: eventType.ID, Label: "Wedding",
		}, &subjects.Placement{GridX: 4, GridY: 4})
	})
	if err != nil {
		t.Fatal(err)
	}
	roleProp, err := properties.Lookup(c, "role", properties.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	personProp, err := properties.Lookup(c, "person", properties.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	eventProp, err := properties.Lookup(c, "event", properties.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	roleTerm, err := propertyterms.Lookup(c, roleProp.ID, "witness", propertyterms.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	bridge, err := writes.Call(c, writes.Op{Action: "create_cited_bridge", UserID: userID}, func(tx *database.Tx) (connect.Result, []rowchange.Change, error) {
		return connect.CreateCitedBridge(tx, userID, connect.CreateInput{
			SourceID: src.ID, FromSubjectID: alice.ID, ToSubjectID: wedding.ID,
			BridgeTypeKey: "participation",
			Citation:      citations.CreateInput{ArtifactID: art.ID, LocatorJSON: testLocator},
			Observations: []observations.Input{
				{PropertyID: personProp.ID, ValueSubjectID: alice.ID},
				{PropertyID: eventProp.ID, ValueSubjectID: wedding.ID},
				{PropertyID: roleProp.ID, ValueTermID: roleTerm.ID},
			},
		})
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writes.Call(c, writes.Op{Action: "delete_citation", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
		changes, err := citations.Delete(tx, userID, bridge.Citation.ID)
		return struct{}{}, changes, err
	}); !errors.Is(err, citations.ErrInUse) {
		t.Fatalf("got %v want ErrInUse", err)
	}
	if _, err := citations.Get(c, bridge.Citation.ID); err != nil {
		t.Fatalf("citation gone: %v", err)
	}
}

func runArtifactCreate(c *database.Catalog, userID []byte, in artifacts.CreateInput) (artifacts.Artifact, error) {
	a, _, err := writes.Run(c, writes.Op{Action: "create_artifact", UserID: userID},
		func(tx *database.Tx) (artifacts.Artifact, []rowchange.Change, error) {
			return artifacts.Create(tx, userID, in)
		})
	return a, err
}
