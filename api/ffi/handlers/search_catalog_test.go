package handlers

import (
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/search"
	"google.golang.org/protobuf/proto"
)

func TestSearchCatalog(t *testing.T) {
	runRPC(t, SearchCatalog, []rpcTest{
		{name: "bad proto", raw: []byte{0xff, 0xff, 0xff, 0xff}, wantErr: true},
		{
			name: "empty query",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, _ := sourceFixture(t)
				return &engine.SearchCatalogRequest{
					ProjectDir: dir,
					Query:      "  ",
					Location:   &engine.WorkspaceLocation{Section: search.SectionSources},
				}
			},
			want: &engine.SearchCatalogResponse{},
		},
		{
			name: "finds created source by title",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, typeID := sourceFixture(t)
				if _, err := CreateSource(marshalProto(t, &engine.CreateSourceRequest{
					ProjectDir:   dir,
					UserId:       userID,
					SourceTypeId: typeID,
					Title:        "Ilminster parish register unique",
					Description:  "notes",
				})); err != nil {
					t.Fatal(err)
				}
				return &engine.SearchCatalogRequest{
					ProjectDir: dir,
					Query:      "Ilminster parish",
					Location:   &engine.WorkspaceLocation{Section: search.SectionSources},
				}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				t.Helper()
				var resp engine.SearchCatalogResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				if len(resp.Hits) == 0 {
					t.Fatal("expected hits")
				}
				top := resp.Hits[0]
				if top.GetKind() != search.KindSource {
					t.Fatalf("kind %s", top.GetKind())
				}
				if top.GetLocation().GetSection() != search.SectionSources || top.GetLocation().GetSourceId() == "" {
					t.Fatalf("location %+v", top.GetLocation())
				}
				if top.GetTitle() == "" {
					t.Fatal("empty title")
				}
			},
		},
		{
			name: "finds seeded type by label",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, _ := sourceFixture(t)
				listOut, err := ListSourceTypes(marshalProto(t, &engine.ListSourceTypesRequest{ProjectDir: dir}))
				if err != nil {
					t.Fatal(err)
				}
				var types engine.ListSourceTypesResponse
				if err := proto.Unmarshal(listOut, &types); err != nil {
					t.Fatal(err)
				}
				if len(types.Types) == 0 {
					t.Fatal("no types")
				}
				label := types.Types[0].GetLabel()
				return &engine.SearchCatalogRequest{
					ProjectDir: dir,
					Query:      label,
					Location:   &engine.WorkspaceLocation{Section: search.SectionSourceTypes},
				}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				t.Helper()
				var resp engine.SearchCatalogResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				found := false
				for _, h := range resp.Hits {
					if h.GetKind() == search.KindSourceType && h.GetLocation().GetTypeId() != "" {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("no source_type hit in %+v", resp.Hits)
				}
			},
		},
		{
			name: "finds seeded field by label",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, _ := sourceFixture(t)
				listOut, err := ListMetadataFields(marshalProto(t, &engine.ListMetadataFieldsRequest{ProjectDir: dir}))
				if err != nil {
					t.Fatal(err)
				}
				var fields engine.ListMetadataFieldsResponse
				if err := proto.Unmarshal(listOut, &fields); err != nil {
					t.Fatal(err)
				}
				if len(fields.Fields) == 0 {
					t.Fatal("no fields")
				}
				return &engine.SearchCatalogRequest{
					ProjectDir: dir,
					Query:      fields.Fields[0].GetLabel(),
					Location:   &engine.WorkspaceLocation{Section: search.SectionSourceFields},
				}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				t.Helper()
				var resp engine.SearchCatalogResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				found := false
				for _, h := range resp.Hits {
					if h.GetKind() == search.KindSourceField && h.GetLocation().GetFieldId() != "" {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("no source_field hit in %+v", resp.Hits)
				}
			},
		},
		{
			name: "recovers title typo",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, typeID := sourceFixture(t)
				if _, err := CreateSource(marshalProto(t, &engine.CreateSourceRequest{
					ProjectDir:   dir,
					UserId:       userID,
					SourceTypeId: typeID,
					Title:        "Ilminster parish register unique",
				})); err != nil {
					t.Fatal(err)
				}
				return &engine.SearchCatalogRequest{
					ProjectDir: dir,
					Query:      "Ilminstr",
					Location:   &engine.WorkspaceLocation{Section: search.SectionSources},
				}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				t.Helper()
				var resp engine.SearchCatalogResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				if len(resp.Hits) == 0 {
					t.Fatal("expected typo hit")
				}
				if resp.Hits[0].GetKind() != search.KindSource {
					t.Fatalf("kind %s", resp.Hits[0].GetKind())
				}
			},
		},
	})
}
