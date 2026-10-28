package handlers

import (
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/search"
	"github.com/mendahu/provenencia/core/workspacelocation"
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
					Location:   &engine.WorkspaceLocation{Section: workspacelocation.Sources},
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
					Location:   &engine.WorkspaceLocation{Section: workspacelocation.Sources},
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
				if top.GetLocation().GetSection() != workspacelocation.Sources || top.GetLocation().GetSourceId() == "" {
					t.Fatalf("location %+v", top.GetLocation())
				}
				if top.GetTitle() == "" {
					t.Fatal("empty title")
				}
				if top.GetIconKey() == "" {
					t.Fatal("expected source type icon_key stub")
				}
			},
		},
		func() rpcTest {
			var f graphFixture
			var graphID, graphRef string
			return rpcTest{
				name: "finds an evidence graph by ref",
				reqFn: func(t *testing.T) proto.Message {
					f = newGraphFixture(t)
					g := f.createGraph(t, "Jones household", f.page2).GetGraph()
					graphID, graphRef = g.GetId(), g.GetRef()
					return &engine.SearchCatalogRequest{
						ProjectDir: f.dir,
						Query:      graphRef,
						Location:   &engine.WorkspaceLocation{Section: workspacelocation.Sources},
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
					loc := top.GetLocation()
					if top.GetKind() != search.KindEvidenceGraph || top.GetRef() != graphRef ||
						loc.GetGraphId() != graphID || loc.GetSourceId() != f.sourceID || loc.GetSourceSurface() != workspacelocation.SurfaceGraph {
						t.Fatalf("hit %+v", top)
					}
				},
			}
		}(),
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
					Location:   &engine.WorkspaceLocation{Section: workspacelocation.SourceTypes},
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
						if h.GetIconKey() == "" {
							t.Fatalf("type hit missing icon_key %+v", h)
						}
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
					Location:   &engine.WorkspaceLocation{Section: workspacelocation.Metadata},
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
					if h.GetKind() == search.KindMetadataField && h.GetLocation().GetFieldId() != "" {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("no metadata_field hit in %+v", resp.Hits)
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
					Location:   &engine.WorkspaceLocation{Section: workspacelocation.Sources},
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

func TestSearchLocationProtoRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		in   workspacelocation.Location
	}{
		{
			name: "properties category and property",
			in: workspacelocation.Location{
				Section: "properties", SubjectTypeKey: "person", PropertyID: "prop-1", Title: "Name",
			},
		},
		{
			name: "metadata field",
			in:   workspacelocation.Location{Section: workspacelocation.Metadata, FieldID: "fld-1", Title: "Author"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := locationFromProto(locationToProto(tt.in))
			if got != tt.in {
				t.Fatalf("got %+v want %+v", got, tt.in)
			}
		})
	}
}

func TestSearchCatalogHandleKinds(t *testing.T) {
	dir, handleRef := namedPerson(t)
	search := func(kinds ...string) *engine.SearchCatalogResponse {
		t.Helper()
		out, err := SearchCatalog(marshalProto(t, &engine.SearchCatalogRequest{ProjectDir: dir, Query: "James Robins", Kinds: kinds}))
		if err != nil {
			t.Fatal(err)
		}
		var resp engine.SearchCatalogResponse
		if err := proto.Unmarshal(out, &resp); err != nil {
			t.Fatal(err)
		}
		return &resp
	}
	if hits := search().GetHits(); len(hits) != 1 || hits[0].GetKind() != "person" {
		t.Fatalf("omnibar default: %+v", hits)
	}
	hits := search("person").GetHits()
	if len(hits) != 1 {
		t.Fatalf("got %+v", hits)
	}
	h := hits[0]
	if h.GetKind() != "person" || h.GetRef() != handleRef || h.GetMemberCount() != 1 {
		t.Fatalf("hit %+v", h)
	}
	if h.GetPerson().GetName().GetForm() == "" {
		t.Fatal("person hit has no header")
	}
	if loc := h.GetLocation(); loc.GetSection() != "persons" || loc.GetEntityId() != h.GetId() {
		t.Fatalf("location %+v", loc)
	}
}

// A Source whose cover thumbnail row is damaged to a type off the allowlist is
// still re-indexed by its next write, and its search hit carries no thumbnail.
func TestSearchHitHidesDamagedCoverThumbnail(t *testing.T) {
	f := sourceWithPNG(t)
	var covered engine.SetSourceCoverResponse
	mustRPC(t, SetSourceCover, &engine.SetSourceCoverRequest{
		ProjectDir: f.dir, UserId: f.userID, SourceId: f.sourceID,
		CoverMode: "artifact", PrimaryArtifactId: f.artifactID,
	}, &covered)

	find := func(query string) *engine.SearchHit {
		t.Helper()
		var resp engine.SearchCatalogResponse
		mustRPC(t, SearchCatalog, &engine.SearchCatalogRequest{
			ProjectDir: f.dir,
			Query:      query,
			Location:   &engine.WorkspaceLocation{Section: workspacelocation.Sources},
		}, &resp)
		if len(resp.Hits) == 0 {
			t.Fatalf("no hit for %q", query)
		}
		return resp.Hits[0]
	}
	if find("Scan").GetThumbnailRelPath() == "" {
		t.Fatal("want a thumbnail on the hit before the damage")
	}

	damageCatalog(t, f.dir, `UPDATE files SET media_type = 'application/x-sh'
		WHERE id IN (SELECT derived_file_id FROM file_derivatives)`)
	mustRPC(t, UpdateSource, &engine.UpdateSourceRequest{
		ProjectDir: f.dir, UserId: f.userID, SourceId: f.sourceID,
		SourceTypeId: covered.GetSource().GetSourceTypeId(), Title: "Scanned ledger",
	}, &engine.UpdateSourceResponse{})

	hit := find("Scanned ledger")
	if hit.GetId() != f.sourceID || hit.GetThumbnailRelPath() != "" {
		t.Fatalf("want the Source's hit with an empty thumbnail_rel_path, got %+v", hit)
	}
}
