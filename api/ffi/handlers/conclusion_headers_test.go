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
	"github.com/mendahu/provenencia/core/database/conclusiondetails"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/database/namevalues/namevaluestest"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
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
		art, err := artifacts.Create(c, userID[:], artifacts.CreateInput{SourceID: sourceID, Label: "Scan"})
		if err != nil {
			return err
		}
		_, err = citations.CreateWithObservations(c, userID[:], citations.CreateInput{
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

func TestGetEventHeader(t *testing.T) {
	runRPC(t, GetEventHeader, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "unknown id",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, _, _ := subjectFixture(t)
				return &engine.GetEventHeaderRequest{ProjectDir: dir, EntityId: uuid.Must(uuid.NewV7()).String()}
			},
			wantErr:   true,
			wantErrIs: conclusiondetails.ErrNotFound,
		},
		{
			name: "returns the one event",
			reqFn: func(t *testing.T) proto.Message {
				dir, id := citedEvent(t)
				return &engine.GetEventHeaderRequest{ProjectDir: dir, EntityId: id}
			},
			after: func(t *testing.T, out []byte, req proto.Message) {
				var resp engine.GetEventHeaderResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				in := req.(*engine.GetEventHeaderRequest)
				if resp.GetHeader().GetEntity().GetId() != in.GetEntityId() || resp.GetHeader().GetEventName() != "The Great Fire" {
					t.Fatalf("%+v", resp.GetHeader())
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
		s, err := subjects.Create(c, userID[:], subjects.CreateInput{SourceID: source[:], SubjectTypeID: st.ID}, nil)
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
		art, err := artifacts.Create(c, userID[:], artifacts.CreateInput{SourceID: source[:], Label: "Scan"})
		if err != nil {
			return err
		}
		year := 1849
		_, err = citations.CreateWithObservations(c, userID[:], citations.CreateInput{
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

func TestGetPlaceHeader(t *testing.T) {
	runRPC(t, GetPlaceHeader, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "unknown id",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, _, _ := subjectFixture(t)
				return &engine.GetPlaceHeaderRequest{ProjectDir: dir, EntityId: uuid.Must(uuid.NewV7()).String()}
			},
			wantErr:   true,
			wantErrIs: conclusiondetails.ErrNotFound,
		},
		{
			name: "a Person is not a Place",
			reqFn: func(t *testing.T) proto.Message {
				dir, ref := namedPerson(t)
				var entityID string
				if err := withProjectCatalog(dir, func(c *database.Catalog) error {
					db, err := c.DB()
					if err != nil {
						return err
					}
					var id []byte
					if err := db.QueryRow(`SELECT id FROM canonical_entities WHERE ref = ?`, ref).Scan(&id); err != nil {
						return err
					}
					entityID = uuidString(id)
					return nil
				}); err != nil {
					t.Fatal(err)
				}
				return &engine.GetPlaceHeaderRequest{ProjectDir: dir, EntityId: entityID}
			},
			wantErr:   true,
			wantErrIs: conclusiondetails.ErrNotFound,
		},
		{
			name: "returns the one place",
			reqFn: func(t *testing.T) proto.Message {
				dir, id := citedPlace(t)
				return &engine.GetPlaceHeaderRequest{ProjectDir: dir, EntityId: id}
			},
			after: func(t *testing.T, out []byte, req proto.Message) {
				var resp engine.GetPlaceHeaderResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				in := req.(*engine.GetPlaceHeaderRequest)
				names := resp.GetHeader().GetNames()
				if resp.GetHeader().GetEntity().GetId() != in.GetEntityId() || len(names) != 2 || names[0] != "York" {
					t.Fatalf("%+v", resp.GetHeader())
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
		s, err := subjects.Create(c, userID[:], subjects.CreateInput{SourceID: source[:], SubjectTypeID: st.ID}, nil)
		if err != nil {
			return err
		}
		toponym, err := properties.Lookup(c, "toponym", properties.OriginProvenencia)
		if err != nil {
			return err
		}
		art, err := artifacts.Create(c, userID[:], artifacts.CreateInput{SourceID: source[:], Label: "Scan"})
		if err != nil {
			return err
		}
		_, err = citations.CreateWithObservations(c, userID[:], citations.CreateInput{
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
