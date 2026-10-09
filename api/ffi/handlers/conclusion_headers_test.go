package handlers

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/catalogsession"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/conclusionheaders"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/evrun"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/database/namevalues/namevaluestest"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/writes"
	"google.golang.org/protobuf/proto"
)

// namedPerson promotes a person Subject with two disagreeing name
// Observations and returns the project dir and the handle ref.
func namedPerson(t *testing.T) (dir, handleRef string) {
	t.Helper()
	req := promotableSubject(t)
	subjectID, err := uuid.Parse(req.SubjectId)
	if err != nil {
		t.Fatal(err)
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		t.Fatal(err)
	}
	if err := withProjectCatalog(req.ProjectDir, func(c *database.Catalog) error {
		name, err := properties.Lookup(c, "name", properties.OriginProvenencia)
		if err != nil {
			return err
		}
		var sourceID []byte
		db, err := c.DB()
		if err != nil {
			return err
		}
		if err := db.QueryRow(`SELECT source_id FROM subjects WHERE id = ?`, subjectID[:]).Scan(&sourceID); err != nil {
			return err
		}
		art, err := runArtifactCreate(c, userID[:], artifacts.CreateInput{SourceID: sourceID, Label: "Scan"})
		if err != nil {
			return err
		}
		_, err = evrun.CreateCitation(c, userID[:], citations.CreateInput{
			ArtifactID: art.ID, LocatorJSON: `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`,
		}, []observations.Input{
			{SubjectID: subjectID[:], PropertyID: name.ID, Name: &namevalues.Value{Form: "James Robins", Parts: []namevalues.Part{
				{Idx: 0, Value: "James", Type: namevalues.PartTypeGiven}, {Idx: 1, Value: "Robins", Type: namevalues.PartTypeSurname},
			}}},
			{SubjectID: subjectID[:], PropertyID: name.ID, Name: namevaluestest.Western("Jim Robins")},
		})
		return err
	}); err != nil {
		t.Fatal(err)
	}
	out, err := PromoteSubject(marshalProto(t, req))
	if err != nil {
		t.Fatal(err)
	}
	var promoted engine.PromoteSubjectResponse
	if err := proto.Unmarshal(out, &promoted); err != nil {
		t.Fatal(err)
	}
	return req.ProjectDir, promoted.Entity.GetRef()
}

func TestListPersonHeaders(t *testing.T) {
	runRPC(t, ListPersonHeaders, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "empty project lists nothing",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, _, _ := subjectFixture(t)
				return &engine.ListPersonHeadersRequest{ProjectDir: dir}
			},
			want:  &engine.ListPersonHeadersResponse{},
			exact: true,
		},
		{
			name: "promoted Person carries its auto-reconciled name and value count",
			reqFn: func(t *testing.T) proto.Message {
				dir, _ := namedPerson(t)
				return &engine.ListPersonHeadersRequest{ProjectDir: dir}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var resp engine.ListPersonHeadersResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				if len(resp.Headers) != 1 {
					t.Fatalf("%+v", resp.Headers)
				}
				h := resp.Headers[0]
				// James Robins + Jim Robins reconcile into one name (S9-13b).
				if !strings.HasPrefix(h.Entity.GetRef(), "PER-") || h.GetNameValueCount() != 1 {
					t.Fatalf("%+v", h)
				}
				n := h.GetName()
				if n.GetForm() != "James Jim Robins" || len(n.GetParts()) != 3 || n.GetParts()[2].GetType() != namevalues.PartTypeSurname {
					t.Fatalf("name %+v", n)
				}
			},
		},
	})
}

func namedPersonID(t *testing.T) (dir, id string) {
	t.Helper()
	dir, ref := namedPerson(t)
	raw, err := ListPersonHeaders(marshalProto(t, &engine.ListPersonHeadersRequest{ProjectDir: dir}))
	if err != nil {
		t.Fatal(err)
	}
	var resp engine.ListPersonHeadersResponse
	if err := proto.Unmarshal(raw, &resp); err != nil {
		t.Fatal(err)
	}
	for _, h := range resp.GetHeaders() {
		if h.GetEntity().GetRef() == ref {
			return dir, h.GetEntity().GetId()
		}
	}
	t.Fatalf("no header for %s", ref)
	return "", ""
}

func TestListEventHeaders(t *testing.T) {
	runRPC(t, ListEventHeaders, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "empty project lists nothing",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, _, _ := subjectFixture(t)
				return &engine.ListEventHeadersRequest{ProjectDir: dir}
			},
			want:  &engine.ListEventHeadersResponse{},
			exact: true,
		},
		{
			name: "promoted Event carries its name, type, and date",
			reqFn: func(t *testing.T) proto.Message {
				dir, _ := citedEvent(t)
				return &engine.ListEventHeadersRequest{ProjectDir: dir}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var resp engine.ListEventHeadersResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				if len(resp.Headers) != 1 {
					t.Fatalf("%+v", resp.Headers)
				}
				h := resp.Headers[0]
				if !strings.HasPrefix(h.Entity.GetRef(), "EVT-") || h.GetEventName() != "The Great Fire" || h.GetEventNameCount() != 1 {
					t.Fatalf("%+v", h)
				}
				if h.GetEventType().GetKey() != "birth" || h.GetEventType().GetLabel() != "Birth" || h.GetDate().GetStartYear() != 1849 {
					t.Fatalf("type %v date %v", h.GetEventType(), h.GetDate())
				}
			},
		},
	})
}

// citedEvent promotes an event with a recorded name, a birth type, and a date.
func citedEvent(t *testing.T) (dir, entityID string) {
	t.Helper()
	dir, user, sourceID, _ := subjectFixture(t)
	userID, err := uuid.Parse(user)
	if err != nil {
		t.Fatal(err)
	}
	source, err := uuid.Parse(sourceID)
	if err != nil {
		t.Fatal(err)
	}
	var subjectID []byte
	if err := withProjectCatalog(dir, func(c *database.Catalog) error {
		st, err := subjecttypes.Lookup(c, "event", subjecttypes.OriginProvenencia)
		if err != nil {
			return err
		}
		s, err := evrun.CreateSubject(c, userID[:], subjects.CreateInput{SourceID: source[:], SubjectTypeID: st.ID}, nil)
		if err != nil {
			return err
		}
		subjectID = s.ID
		name, err := properties.Lookup(c, "event_name", properties.OriginProvenencia)
		if err != nil {
			return err
		}
		eventType, err := properties.Lookup(c, "event_type", properties.OriginProvenencia)
		if err != nil {
			return err
		}
		term, err := propertyterms.Lookup(c, eventType.ID, "birth", propertyterms.OriginProvenencia)
		if err != nil {
			return err
		}
		dateProp, err := properties.Lookup(c, "date", properties.OriginProvenencia)
		if err != nil {
			return err
		}
		art, err := runArtifactCreate(c, userID[:], artifacts.CreateInput{SourceID: source[:], Label: "Scan"})
		if err != nil {
			return err
		}
		year := 1849
		_, err = evrun.CreateCitation(c, userID[:], citations.CreateInput{
			ArtifactID: art.ID, LocatorJSON: `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`,
		}, []observations.Input{
			{SubjectID: s.ID, PropertyID: name.ID, ValueText: "The Great Fire", HasText: true},
			{SubjectID: s.ID, PropertyID: eventType.ID, ValueTermID: term.ID},
			{SubjectID: s.ID, PropertyID: dateProp.ID, Date: &datevalues.Value{Kind: datevalues.KindPoint, Calendar: "gregorian", StartYear: &year}},
		})
		if err != nil {
			return err
		}
		p, err := promote.Save(c, userID[:], promote.Input{SubjectID: subjectID})
		if err != nil {
			return err
		}
		entityID = uuidString(p.Entity.ID)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return dir, entityID
}

func TestListPlaceHeaders(t *testing.T) {
	runRPC(t, ListPlaceHeaders, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "empty project lists nothing",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, _, _ := subjectFixture(t)
				return &engine.ListPlaceHeadersRequest{ProjectDir: dir}
			},
			want:  &engine.ListPlaceHeadersResponse{},
			exact: true,
		},
		{
			name: "promoted Place carries every kept toponym",
			reqFn: func(t *testing.T) proto.Message {
				dir, _ := citedPlace(t)
				return &engine.ListPlaceHeadersRequest{ProjectDir: dir}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var resp engine.ListPlaceHeadersResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				if len(resp.Headers) != 1 || len(resp.Headers[0].GetNames()) != 2 ||
					resp.Headers[0].GetNames()[0] != "York" || resp.Headers[0].GetNames()[1] != "Toronto" ||
					resp.Headers[0].GetKind() != "" || len(resp.Headers[0].GetParents()) != 0 ||
					resp.Headers[0].GetStartDate() != nil || resp.Headers[0].GetEndDate() != nil {
					t.Fatalf("%+v", resp.Headers)
				}
			},
		},
	})
}

// citedPlace promotes a Place named York and Toronto.
func citedPlace(t *testing.T) (dir, entityID string) {
	t.Helper()
	dir, user, sourceID, _ := subjectFixture(t)
	userID, err := uuid.Parse(user)
	if err != nil {
		t.Fatal(err)
	}
	source, err := uuid.Parse(sourceID)
	if err != nil {
		t.Fatal(err)
	}
	if err := withProjectCatalog(dir, func(c *database.Catalog) error {
		st, err := subjecttypes.Lookup(c, "place", subjecttypes.OriginProvenencia)
		if err != nil {
			return err
		}
		s, err := evrun.CreateSubject(c, userID[:], subjects.CreateInput{SourceID: source[:], SubjectTypeID: st.ID}, nil)
		if err != nil {
			return err
		}
		toponym, err := properties.Lookup(c, "toponym", properties.OriginProvenencia)
		if err != nil {
			return err
		}
		art, err := runArtifactCreate(c, userID[:], artifacts.CreateInput{SourceID: source[:], Label: "Scan"})
		if err != nil {
			return err
		}
		_, err = evrun.CreateCitation(c, userID[:], citations.CreateInput{
			ArtifactID: art.ID, LocatorJSON: `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`,
		}, []observations.Input{
			{SubjectID: s.ID, PropertyID: toponym.ID, ValueText: "York", HasText: true},
			{SubjectID: s.ID, PropertyID: toponym.ID, ValueText: "Toronto", HasText: true},
		})
		if err != nil {
			return err
		}
		p, err := promote.Save(c, userID[:], promote.Input{SubjectID: s.ID})
		if err != nil {
			return err
		}
		entityID = uuidString(p.Entity.ID)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return dir, entityID
}

func TestWorkspaceNavCountsConclusionHandles(t *testing.T) {
	dir, _ := namedPerson(t)
	t.Cleanup(func() { _ = catalogsession.CloseAll() })
	// One event and two place handles, minted directly: the count is of
	// handles, whatever their members.
	if err := withProjectCatalog(dir, func(c *database.Catalog) error {
		db, err := c.DB()
		if err != nil {
			return err
		}
		var userID []byte
		if err := db.QueryRow(`SELECT id FROM users LIMIT 1`).Scan(&userID); err != nil {
			return err
		}
		for _, key := range []string{"event", "place", "place"} {
			st, err := subjecttypes.Lookup(c, key, subjecttypes.OriginProvenencia)
			if err != nil {
				return err
			}
			if _, err := canonicalentities.Create(c, userID, canonicalentities.CreateInput{SubjectTypeID: st.ID}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	out, err := GetWorkspaceNavCounts(marshalProto(t, &engine.GetWorkspaceNavCountsRequest{ProjectDir: dir}))
	if err != nil {
		t.Fatal(err)
	}
	var resp engine.GetWorkspaceNavCountsResponse
	if err := proto.Unmarshal(out, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.GetPersons() != 1 || resp.GetEvents() != 1 || resp.GetPlaces() != 2 {
		t.Fatalf("persons %d events %d places %d", resp.GetPersons(), resp.GetEvents(), resp.GetPlaces())
	}
}

func TestListSubjectMembershipsCarriesAutoReconciledName(t *testing.T) {
	dir, ref := namedPerson(t)
	t.Cleanup(func() { _ = catalogsession.CloseAll() })
	var sourceID string
	if err := withProjectCatalog(dir, func(c *database.Catalog) error {
		db, err := c.DB()
		if err != nil {
			return err
		}
		var id []byte
		if err := db.QueryRow(`SELECT source_id FROM subjects LIMIT 1`).Scan(&id); err != nil {
			return err
		}
		sourceID = uuidString(id)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	out, err := ListSubjectMemberships(marshalProto(t, &engine.ListSubjectMembershipsRequest{ProjectDir: dir, SourceId: sourceID}))
	if err != nil {
		t.Fatal(err)
	}
	var resp engine.ListSubjectMembershipsResponse
	if err := proto.Unmarshal(out, &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Memberships) != 1 || resp.Memberships[0].Entity.GetRef() != ref ||
		resp.Memberships[0].GetName().GetForm() != "James Jim Robins" {
		t.Fatalf("%+v", resp.Memberships)
	}
}

func TestLifeFactsProtoCarriesIdentities(t *testing.T) {
	event := canonicalentities.Entity{ID: mustUUIDBytes(t), Ref: "EVT-1"}
	place := canonicalentities.Entity{ID: mustUUIDBytes(t), Ref: "PLC-1", Label: "York"}
	tests := []struct {
		name      string
		in        conclusionheaders.LifeFacts
		wantEvent string
		wantPlace string
	}{
		{name: "no event linked", in: conclusionheaders.LifeFacts{}},
		{
			name: "event and place",
			in: conclusionheaders.LifeFacts{
				Event:  &event,
				Places: []conclusionheaders.HeaderPlace{{Entity: place, Names: []string{"York"}}},
			},
			wantEvent: "EVT-1",
			wantPlace: "PLC-1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lifeFactsProto(tt.in)
			if got.GetEvent().GetRef() != tt.wantEvent {
				t.Fatalf("event %+v", got.GetEvent())
			}
			if tt.wantEvent != "" && got.GetEvent().GetId() != uuidString(event.ID) {
				t.Fatalf("event id %q", got.GetEvent().GetId())
			}
			var placeRef string
			if len(got.GetPlaces()) > 0 {
				placeRef = got.GetPlaces()[0].GetEntity().GetRef()
			}
			if placeRef != tt.wantPlace {
				t.Fatalf("place %+v", got.GetPlaces())
			}
		})
	}
}

func mustUUIDBytes(t *testing.T) []byte {
	t.Helper()
	id := uuid.Must(uuid.NewV7())
	return id[:]
}

func eventSourceID(t *testing.T, dir string) string {
	t.Helper()
	var sourceID string
	if err := withProjectCatalog(dir, func(c *database.Catalog) error {
		db, err := c.DB()
		if err != nil {
			return err
		}
		var id []byte
		if err := db.QueryRow(`SELECT s.source_id FROM subjects s
			JOIN subject_types st ON st.id = s.subject_type_id AND st.key = 'event' LIMIT 1`).Scan(&id); err != nil {
			return err
		}
		sourceID = uuidString(id)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return sourceID
}

func TestEventTitlesUseOneRule(t *testing.T) {
	t.Run("an Event header carries its chosen title", func(t *testing.T) {
		dir, _ := citedEvent(t)
		t.Cleanup(func() { _ = catalogsession.CloseAll() })
		out, err := ListEventHeaders(marshalProto(t, &engine.ListEventHeadersRequest{ProjectDir: dir}))
		if err != nil {
			t.Fatal(err)
		}
		var resp engine.ListEventHeadersResponse
		if err := proto.Unmarshal(out, &resp); err != nil {
			t.Fatal(err)
		}
		title := resp.GetHeaders()[0].GetTitle()
		if title.GetRule() != engine.EventTitleRule_EVENT_TITLE_RULE_RECORDED_NAME || title.GetRecordedName() != "The Great Fire" ||
			title.GetTypeKey() != "birth" || !strings.HasPrefix(title.GetRef(), "EVT-") {
			t.Fatalf("%+v", title)
		}
	})
	runRPC(t, ListSourceEventTitles, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "bad source id",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, _, _ := subjectFixture(t)
				return &engine.ListSourceEventTitlesRequest{ProjectDir: dir, SourceId: "nope"}
			},
			wantErr: true,
		},
		{
			name: "the Source's Event card is titled by the same rule",
			reqFn: func(t *testing.T) proto.Message {
				dir, _ := citedEvent(t)
				return &engine.ListSourceEventTitlesRequest{ProjectDir: dir, SourceId: eventSourceID(t, dir)}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var resp engine.ListSourceEventTitlesResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				if len(resp.Titles) != 1 || resp.Titles[0].GetSubjectId() == "" {
					t.Fatalf("%+v", resp.Titles)
				}
				title := resp.Titles[0].GetTitle()
				if title.GetRule() != engine.EventTitleRule_EVENT_TITLE_RULE_RECORDED_NAME || title.GetRecordedName() != "The Great Fire" {
					t.Fatalf("%+v", title)
				}
			},
		},
	})
}

func runArtifactCreate(c *database.Catalog, userID []byte, in artifacts.CreateInput) (artifacts.Artifact, error) {
	a, _, err := writes.Run(c, writes.Op{Action: "create_artifact", UserID: userID},
		func(tx *database.Tx) (artifacts.Artifact, []rowchange.Change, error) {
			return artifacts.Create(tx, userID, in)
		})
	return a, err
}
