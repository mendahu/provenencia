package handlers

import (
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
	"google.golang.org/protobuf/proto"
)

func TestListPropertyTerms(t *testing.T) {
	runRPC(t, ListPropertyTerms, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "lists seeded event_type terms",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, _ := sourceFixture(t)
				propID := propertyIDByKey(t, dir, "event_type")
				return &engine.ListPropertyTermsRequest{ProjectDir: dir, PropertyId: propID}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var resp engine.ListPropertyTermsResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				if len(resp.Terms) != 11 {
					t.Fatalf("len=%d want 11", len(resp.Terms))
				}
				found := false
				for _, term := range resp.Terms {
					if term.GetKey() == "birth" && term.GetOrigin() == "provenencia" {
						found = true
						break
					}
				}
				if !found {
					t.Fatal("missing birth term")
				}
			},
		},
	})
}

func TestCreateUpdateDeletePropertyTerm(t *testing.T) {
	runRPC(t, CreatePropertyTerm, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "create update delete round trip",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _ := sourceFixture(t)
				propID := propertyIDByKey(t, dir, "event_type")
				return &engine.CreatePropertyTermRequest{
					ProjectDir: dir, UserId: userID, PropertyId: propID,
					Label: "Land Grant", Description: "custom",
				}
			},
			after: func(t *testing.T, out []byte, req proto.Message) {
				var created engine.CreatePropertyTermResponse
				if err := proto.Unmarshal(out, &created); err != nil {
					t.Fatal(err)
				}
				if created.GetTerm().GetKey() != "land-grant" || created.GetTerm().GetOrigin() != "user" {
					t.Fatalf("%+v", created.GetTerm())
				}
				cr := req.(*engine.CreatePropertyTermRequest)
				uout, err := UpdatePropertyTerm(marshalProto(t, &engine.UpdatePropertyTermRequest{
					ProjectDir: cr.ProjectDir, UserId: cr.UserId,
					TermId: created.GetTerm().GetId(), Label: "Land Grant 2", Description: "",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var updated engine.UpdatePropertyTermResponse
				if err := proto.Unmarshal(uout, &updated); err != nil {
					t.Fatal(err)
				}
				if updated.GetTerm().GetLabel() != "Land Grant 2" {
					t.Fatalf("%+v", updated.GetTerm())
				}
				if _, err := DeletePropertyTerm(marshalProto(t, &engine.DeletePropertyTermRequest{
					ProjectDir: cr.ProjectDir, UserId: cr.UserId, TermId: created.GetTerm().GetId(),
				})); err != nil {
					t.Fatal(err)
				}
			},
		},
	})
	runRPC(t, UpdatePropertyTerm, []rpcTest{
		{
			name: "refuse update product term",
			reqFn: func(t *testing.T) proto.Message {
				d, uid, _ := sourceFixture(t)
				pid := propertyIDByKey(t, d, "event_type")
				out, err := ListPropertyTerms(marshalProto(t, &engine.ListPropertyTermsRequest{
					ProjectDir: d, PropertyId: pid,
				}))
				if err != nil {
					t.Fatal(err)
				}
				var list engine.ListPropertyTermsResponse
				if err := proto.Unmarshal(out, &list); err != nil {
					t.Fatal(err)
				}
				var birthID string
				for _, term := range list.Terms {
					if term.GetKey() == "birth" {
						birthID = term.GetId()
						break
					}
				}
				if birthID == "" {
					t.Fatal("missing birth")
				}
				return &engine.UpdatePropertyTermRequest{
					ProjectDir: d, UserId: uid, TermId: birthID, Label: "Nope",
				}
			},
			wantErr: true,
		},
	})
}

func propertyIDByKey(t *testing.T, dir, key string) string {
	t.Helper()
	out, err := ListProperties(marshalProto(t, &engine.ListPropertiesRequest{ProjectDir: dir}))
	if err != nil {
		t.Fatal(err)
	}
	var resp engine.ListPropertiesResponse
	if err := proto.Unmarshal(out, &resp); err != nil {
		t.Fatal(err)
	}
	for _, p := range resp.Properties {
		if p.GetKey() == key {
			return p.GetId()
		}
	}
	t.Fatalf("missing property %q", key)
	return ""
}
