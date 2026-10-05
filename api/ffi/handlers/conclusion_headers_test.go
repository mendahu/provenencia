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
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/database/namevalues/namevaluestest"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/properties"
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
			name: "promoted Person carries its resolved name and cluster count",
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
				if !strings.HasPrefix(h.Entity.GetRef(), "PER-") || h.GetNameClusterCount() != 2 {
					t.Fatalf("%+v", h)
				}
				n := h.GetName()
				if n.GetForm() != "James Robins" || len(n.GetParts()) != 2 || n.GetParts()[1].GetType() != namevalues.PartTypeSurname {
					t.Fatalf("name %+v", n)
				}
			},
		},
	})
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

func TestListSubjectMembershipsCarriesResolvedName(t *testing.T) {
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
		resp.Memberships[0].GetName().GetForm() != "James Robins" {
		t.Fatalf("%+v", resp.Memberships)
	}
}
