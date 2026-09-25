package handlers

import (
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database/connect"
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
	typesOut, err := ListSubjectTypes(marshalProto(t, &engine.ListSubjectTypesRequest{ProjectDir: dir}))
	if err != nil {
		t.Fatal(err)
	}
	var types engine.ListSubjectTypesResponse
	if err := proto.Unmarshal(typesOut, &types); err != nil {
		t.Fatal(err)
	}
	personTypeID, eventTypeID := "", ""
	for _, typ := range types.Types {
		switch typ.GetKey() {
		case "person":
			personTypeID = typ.GetId()
		case "event":
			eventTypeID = typ.GetId()
		}
	}
	personOut, err := CreateSubject(marshalProto(t, &engine.CreateSubjectRequest{
		ProjectDir: dir, UserId: userID, SourceId: sourceID, SubjectTypeId: personTypeID, Label: "Ada",
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
		ProjectDir: dir, UserId: userID, SourceId: sourceID, SubjectTypeId: eventTypeID, Label: "Census",
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
	propsOut, err := ListProperties(marshalProto(t, &engine.ListPropertiesRequest{ProjectDir: dir}))
	if err != nil {
		t.Fatal(err)
	}
	var props engine.ListPropertiesResponse
	if err := proto.Unmarshal(propsOut, &props); err != nil {
		t.Fatal(err)
	}
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

func TestCreateCitedBridge(t *testing.T) {
	runRPC(t, CreateCitedBridge, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "creates participation",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, sourceID, artifactID, personID, eventID, personProp, eventProp, roleProp, roleTerm := connectFixture(t)
				return &engine.CreateCitedBridgeRequest{
					ProjectDir:    dir,
					UserId:        userID,
					SourceId:      sourceID,
					FromSubjectId: personID,
					ToSubjectId:   eventID,
					BridgeTypeKey: "participation",
					GridX:         4,
					GridY:         5,
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
			name: "refuses person place",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, sourceID, artifactID, personID, _, personProp, _, _, _ := connectFixture(t)
				typesOut, err := ListSubjectTypes(marshalProto(t, &engine.ListSubjectTypesRequest{ProjectDir: dir}))
				if err != nil {
					t.Fatal(err)
				}
				var types engine.ListSubjectTypesResponse
				if err := proto.Unmarshal(typesOut, &types); err != nil {
					t.Fatal(err)
				}
				placeTypeID := ""
				for _, typ := range types.Types {
					if typ.GetKey() == "place" {
						placeTypeID = typ.GetId()
					}
				}
				placeOut, err := CreateSubject(marshalProto(t, &engine.CreateSubjectRequest{
					ProjectDir: dir, UserId: userID, SourceId: sourceID, SubjectTypeId: placeTypeID, Label: "Leeds",
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
				propsOut, err := ListProperties(marshalProto(t, &engine.ListPropertiesRequest{ProjectDir: dir}))
				if err != nil {
					t.Fatal(err)
				}
				var props engine.ListPropertiesResponse
				if err := proto.Unmarshal(propsOut, &props); err != nil {
					t.Fatal(err)
				}
				placeProp := ""
				for _, p := range props.Properties {
					if p.GetKey() == "place" {
						placeProp = p.GetId()
					}
				}
				return &engine.CreateCitedBridgeRequest{
					ProjectDir:    dir,
					UserId:        userID,
					SourceId:      sourceID,
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
