package handlers

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/connect"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/ingest"
	"google.golang.org/protobuf/proto"
)

func connectFixture(t *testing.T) (
	dir, userID, sourceID, artifactID, personID, eventID, personPropID, eventPropID, rolePropID, roleTermID string,
) {
	t.Helper()
	dir, userID, sourceID, _ = subjectFixture(t)
	aout, err := CreateArtifact(marshalProto(t, &engine.CreateArtifactRequest{
		ProjectDir: dir,
		UserId:     userID,
		SourceId:   sourceID,
		Label:      "Scan",
	}))
	if err != nil {
		t.Fatal(err)
	}
	var art engine.CreateArtifactResponse
	if err := proto.Unmarshal(aout, &art); err != nil {
		t.Fatal(err)
	}
	personTypeID, eventTypeID := subjectTypeIDByKey(t, dir, "person"), subjectTypeIDByKey(t, dir, "event")
	personOut, err := CreateSubject(marshalProto(t, &engine.CreateSubjectRequest{
		ProjectDir: dir, UserId: userID, GraphId: soleGraphID(t, dir, sourceID), SubjectTypeId: personTypeID, Label: "Ada",
	}))
	if err != nil {
		t.Fatal(err)
	}
	var person engine.CreateSubjectResponse
	if err := proto.Unmarshal(personOut, &person); err != nil {
		t.Fatal(err)
	}
	if _, err := SetSubjectPosition(marshalProto(t, &engine.SetSubjectPositionRequest{
		ProjectDir: dir, SubjectId: person.Subject.GetId(), GridX: 0, GridY: 0,
	})); err != nil {
		t.Fatal(err)
	}
	eventOut, err := CreateSubject(marshalProto(t, &engine.CreateSubjectRequest{
		ProjectDir: dir, UserId: userID, GraphId: soleGraphID(t, dir, sourceID), SubjectTypeId: eventTypeID, Label: "Census",
	}))
	if err != nil {
		t.Fatal(err)
	}
	var event engine.CreateSubjectResponse
	if err := proto.Unmarshal(eventOut, &event); err != nil {
		t.Fatal(err)
	}
	if _, err := SetSubjectPosition(marshalProto(t, &engine.SetSubjectPositionRequest{
		ProjectDir: dir, SubjectId: event.Subject.GetId(), GridX: 2, GridY: 4,
	})); err != nil {
		t.Fatal(err)
	}
	var props engine.GetPropertiesWorkspaceResponse
	mustRPC(t, GetPropertiesWorkspace, &engine.GetPropertiesWorkspaceRequest{ProjectDir: dir}, &props)
	for _, p := range props.Properties {
		switch p.GetKey() {
		case "person":
			personPropID = p.GetId()
		case "event":
			eventPropID = p.GetId()
		case "role":
			rolePropID = p.GetId()
		}
	}
	termsOut, err := ListPropertyTerms(marshalProto(t, &engine.ListPropertyTermsRequest{
		ProjectDir: dir, PropertyId: rolePropID,
	}))
	if err != nil {
		t.Fatal(err)
	}
	var terms engine.ListPropertyTermsResponse
	if err := proto.Unmarshal(termsOut, &terms); err != nil {
		t.Fatal(err)
	}
	for _, term := range terms.Terms {
		if term.GetKey() == "witness" {
			roleTermID = term.GetId()
			break
		}
	}
	if roleTermID == "" && len(terms.Terms) > 0 {
		roleTermID = terms.Terms[0].GetId()
	}
	return dir, userID, sourceID, art.Artifact.GetId(), person.Subject.GetId(), event.Subject.GetId(),
		personPropID, eventPropID, rolePropID, roleTermID
}

// requireEndsNamed fails unless the effects name every end handle, in handles
// or in summary_handles, so a cached detail page of either end reloads.
func requireEndsNamed(t *testing.T, e *engine.WriteEffects, ends [][]byte) {
	t.Helper()
	named := append(slices.Clone(e.GetHandles()), e.GetSummaryHandles()...)
	for _, end := range ends {
		if !slices.Contains(named, database.UUIDString(end)) {
			t.Errorf("effects handles %v and summary_handles %v lack the end %x", e.GetHandles(), e.GetSummaryHandles(), end)
		}
	}
}

func TestCreateCitedBridge(t *testing.T) {
	var dir string
	var ends [][]byte
	runRPC(t, CreateCitedBridge, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "creates participation",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _, artifactID, personID, eventID, personProp, eventProp, roleProp, roleTerm := connectFixture(t)
				return &engine.CreateCitedBridgeRequest{
					ProjectDir:    dir,
					UserId:        userID,
					FromSubjectId: personID,
					ToSubjectId:   eventID,
					BridgeTypeKey: "participation",
					ArtifactId:    artifactID,
					LocatorJson:   validLocatorJSON,
					Observations: []*engine.ObservationDraft{
						{PropertyId: personProp, ValueSubjectId: personID},
						{PropertyId: eventProp, ValueSubjectId: eventID},
						{PropertyId: roleProp, ValueTermId: roleTerm},
					},
				}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var created engine.CreateCitedBridgeResponse
				if err := proto.Unmarshal(out, &created); err != nil {
					t.Fatal(err)
				}
				if created.GetSubject().GetId() == "" || created.GetCitation().GetId() == "" {
					t.Fatal("missing subject or citation")
				}
				if len(created.GetObservations()) != 3 {
					t.Fatalf("observations %d", len(created.GetObservations()))
				}
			},
		},
		{
			name: "files a participation between two promoted subjects",
			reqFn: func(t *testing.T) proto.Message {
				d, userID, _, artifactID, personID, eventID, personProp, eventProp, roleProp, roleTerm := connectFixture(t)
				dir = d
				ends = [][]byte{
					promoteInDir(t, dir, userID, personID, "").Entity.ID,
					promoteInDir(t, dir, userID, eventID, "").Entity.ID,
				}
				return &engine.CreateCitedBridgeRequest{
					ProjectDir:    dir,
					UserId:        userID,
					FromSubjectId: personID,
					ToSubjectId:   eventID,
					BridgeTypeKey: "participation",
					ArtifactId:    artifactID,
					LocatorJson:   validLocatorJSON,
					Observations: []*engine.ObservationDraft{
						{PropertyId: personProp, ValueSubjectId: personID},
						{PropertyId: eventProp, ValueSubjectId: eventID},
						{PropertyId: roleProp, ValueTermId: roleTerm},
					},
				}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var created engine.CreateCitedBridgeResponse
				if err := proto.Unmarshal(out, &created); err != nil {
					t.Fatal(err)
				}
				bridge := uuid.MustParse(created.GetSubject().GetId())
				if err := withProjectCatalog(dir, func(c *database.Catalog) error {
					claim, err := identityclaims.AcceptedEntityForSubject(c, bridge[:])
					if err != nil {
						return err
					}
					if !slices.Contains(created.GetEffects().GetHandles(), database.UUIDString(claim.EntityID)) {
						t.Fatalf("effects.handles %v lacks the association %x", created.GetEffects().GetHandles(), claim.EntityID)
					}
					return nil
				}); err != nil {
					t.Fatal(err)
				}
				requireEndsNamed(t, created.GetEffects(), ends)
			},
		},
		{
			name: "files a relationship between two promoted people",
			reqFn: func(t *testing.T) proto.Message {
				d, userID, sourceID, artifactID, personID, _, personProp, _, _, _ := connectFixture(t)
				dir = d
				personOut, err := CreateSubject(marshalProto(t, &engine.CreateSubjectRequest{
					ProjectDir: dir, UserId: userID, GraphId: soleGraphID(t, dir, sourceID),
					SubjectTypeId: subjectTypeIDByKey(t, dir, "person"), Label: "Bea",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var other engine.CreateSubjectResponse
				if err := proto.Unmarshal(personOut, &other); err != nil {
					t.Fatal(err)
				}
				otherID := other.Subject.GetId()
				if _, err := SetSubjectPosition(marshalProto(t, &engine.SetSubjectPositionRequest{
					ProjectDir: dir, SubjectId: otherID, GridX: 4, GridY: 0,
				})); err != nil {
					t.Fatal(err)
				}
				ends = [][]byte{
					promoteInDir(t, dir, userID, personID, "").Entity.ID,
					promoteInDir(t, dir, userID, otherID, "").Entity.ID,
				}
				typeProp := propertyIDByKey(t, dir, "relationship_type")
				var terms engine.ListPropertyTermsResponse
				mustRPC(t, ListPropertyTerms, &engine.ListPropertyTermsRequest{ProjectDir: dir, PropertyId: typeProp}, &terms)
				return &engine.CreateCitedBridgeRequest{
					ProjectDir:    dir,
					UserId:        userID,
					FromSubjectId: personID,
					ToSubjectId:   otherID,
					BridgeTypeKey: "relationship",
					ArtifactId:    artifactID,
					LocatorJson:   validLocatorJSON,
					Observations: []*engine.ObservationDraft{
						{PropertyId: personProp, ValueSubjectId: personID},
						{PropertyId: propertyIDByKey(t, dir, "related_to"), ValueSubjectId: otherID},
						{PropertyId: typeProp, ValueTermId: terms.Terms[0].GetId()},
					},
				}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var created engine.CreateCitedBridgeResponse
				if err := proto.Unmarshal(out, &created); err != nil {
					t.Fatal(err)
				}
				requireEndsNamed(t, created.GetEffects(), ends)
			},
		},
		{
			name: "stages a depiction draft",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _, artifactID, personID, eventID, personProp, eventProp, roleProp, roleTerm := connectFixture(t)
				link := filepath.Join(t.TempDir(), "link.jpg")
				if err := os.Symlink(writeCrop(t, 300, 200, 128), link); err != nil {
					t.Fatal(err)
				}
				return &engine.CreateCitedBridgeRequest{
					ProjectDir:    dir,
					UserId:        userID,
					FromSubjectId: personID,
					ToSubjectId:   eventID,
					BridgeTypeKey: "participation",
					ArtifactId:    artifactID,
					LocatorJson:   validLocatorJSON,
					Observations: []*engine.ObservationDraft{
						{PropertyId: personProp, ValueSubjectId: personID},
						{PropertyId: eventProp, ValueSubjectId: eventID},
						{PropertyId: roleProp, ValueTermId: roleTerm},
						{PropertyId: personProp, Depiction: &engine.DepictionValueInput{RegionJson: likenessRegionJSON, ImagePath: link}},
					},
				}
			},
			wantErr:   true,
			wantErrIs: ingest.ErrSymlink,
		},
		{
			name: "refuses person place",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, sourceID, artifactID, personID, _, personProp, _, _, _ := connectFixture(t)
				placeTypeID := subjectTypeIDByKey(t, dir, "place")
				placeOut, err := CreateSubject(marshalProto(t, &engine.CreateSubjectRequest{
					ProjectDir: dir, UserId: userID, GraphId: soleGraphID(t, dir, sourceID), SubjectTypeId: placeTypeID, Label: "Leeds",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var place engine.CreateSubjectResponse
				if err := proto.Unmarshal(placeOut, &place); err != nil {
					t.Fatal(err)
				}
				if _, err := SetSubjectPosition(marshalProto(t, &engine.SetSubjectPositionRequest{
					ProjectDir: dir, SubjectId: place.Subject.GetId(), GridX: 6, GridY: 4,
				})); err != nil {
					t.Fatal(err)
				}
				placeProp := propertyIDByKey(t, dir, "place")
				return &engine.CreateCitedBridgeRequest{
					ProjectDir:    dir,
					UserId:        userID,
					FromSubjectId: personID,
					ToSubjectId:   place.Subject.GetId(),
					BridgeTypeKey: "location",
					ArtifactId:    artifactID,
					LocatorJson:   validLocatorJSON,
					Observations: []*engine.ObservationDraft{
						{PropertyId: personProp, ValueSubjectId: personID},
						{PropertyId: placeProp, ValueSubjectId: place.Subject.GetId()},
					},
				}
			},
			wantErr:   true,
			wantErrIs: connect.ErrRefused,
		},
	})
}

// Deleting a bridge Subject releases its connection Observations, so the
// effects report property_usage and the app reloads each Property's usedBy from
// them, with no mutation of its own.
func TestDeleteBridgeReportsPropertyUsage(t *testing.T) {
	dir, userID, _, artifactID, personID, eventID, personProp, eventProp, roleProp, roleTerm := connectFixture(t)
	var bridge engine.CreateCitedBridgeResponse
	mustRPC(t, CreateCitedBridge, &engine.CreateCitedBridgeRequest{
		ProjectDir: dir, UserId: userID, FromSubjectId: personID, ToSubjectId: eventID,
		BridgeTypeKey: "participation", ArtifactId: artifactID, LocatorJson: validLocatorJSON,
		Observations: []*engine.ObservationDraft{
			{PropertyId: personProp, ValueSubjectId: personID},
			{PropertyId: eventProp, ValueSubjectId: eventID},
			{PropertyId: roleProp, ValueTermId: roleTerm},
		},
	}, &bridge)
	var resp engine.DeleteSubjectResponse
	mustRPC(t, DeleteSubject, &engine.DeleteSubjectRequest{
		ProjectDir: dir, UserId: userID, SubjectId: bridge.GetSubject().GetId(),
	}, &resp)
	if !contains(resp.GetEffects().GetLists(), "property_usage") {
		t.Fatalf("lists = %v, want property_usage", resp.GetEffects().GetLists())
	}
}
