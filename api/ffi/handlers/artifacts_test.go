package handlers

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/catalogsession"
	"github.com/mendahu/provenencia/core/catalogsession/catalogsessiontest"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/evidencegraphs"
	"github.com/mendahu/provenencia/core/database/sourcemetadata"
	"github.com/mendahu/provenencia/core/ingest"
	"google.golang.org/protobuf/proto"
)

func TestCreateArtifactAndIngest(t *testing.T) {
	runRPC(t, CreateArtifact, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "fileless then ingest",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, typeID := sourceFixture(t)
				cout, err := CreateSource(marshalProto(t, &engine.CreateSourceRequest{
					ProjectDir: dir, UserId: userID, SourceTypeId: typeID, Title: "Photo",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateSourceResponse
				if err := proto.Unmarshal(cout, &created); err != nil {
					t.Fatal(err)
				}
				return &engine.CreateArtifactRequest{
					ProjectDir: dir, UserId: userID, SourceId: created.Source.Id,
					Label: "Front", Description: "front",
				}
			},
			want: nil,
			after: func(t *testing.T, out []byte, req proto.Message) {
				var art engine.CreateArtifactResponse
				if err := proto.Unmarshal(out, &art); err != nil {
					t.Fatal(err)
				}
				if art.Artifact.GetLabel() != "Front" || art.Artifact.GetDescription() != "front" {
					t.Fatalf("%+v", art.Artifact)
				}
				cr := req.(*engine.CreateArtifactRequest)
				assertLatestAuditAction(t, cr.ProjectDir, "create_artifact")
				path := filepath.Join(t.TempDir(), "scan.png")
				writePNG(t, path)
				iout, err := IngestArtifactFile(marshalProto(t, &engine.IngestArtifactFileRequest{
					ProjectDir: cr.ProjectDir, UserId: cr.UserId, ArtifactId: art.Artifact.Id, Path: path,
				}))
				if err != nil {
					t.Fatal(err)
				}
				var ingested engine.IngestArtifactFileResponse
				if err := proto.Unmarshal(iout, &ingested); err != nil {
					t.Fatal(err)
				}
				if ingested.File.GetOriginalFilename() != "scan.png" || ingested.GetReused() {
					t.Fatalf("%+v", &ingested)
				}
				if got := ingested.File.GetTypeIdentifiers(); len(got) != 1 || got[0] != "public.png" {
					t.Fatalf("type identifiers = %v, want [public.png]", got)
				}
				if ingested.Artifact.GetFileId() == "" {
					t.Fatal("expected file_id")
				}
				assertAuditActionPresent(t, cr.ProjectDir, "create_file")
				assertAuditActionPresent(t, cr.ProjectDir, "update_artifact")
				assertLatestAuditAction(t, cr.ProjectDir, "update_artifact")

				path2 := filepath.Join(t.TempDir(), "scan2.png")
				writePNG(t, path2)
				_, err = IngestArtifactFile(marshalProto(t, &engine.IngestArtifactFileRequest{
					ProjectDir: cr.ProjectDir, UserId: cr.UserId, ArtifactId: art.Artifact.Id, Path: path2,
				}))
				if !errors.Is(err, artifacts.ErrFileAlreadyAttached) {
					t.Fatalf("second ingest: %v", err)
				}

				uout, err := UpdateArtifact(marshalProto(t, &engine.UpdateArtifactRequest{
					ProjectDir: cr.ProjectDir, UserId: cr.UserId, ArtifactId: art.Artifact.Id,
					Label: "Front scan", Description: "updated",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var updated engine.UpdateArtifactResponse
				if err := proto.Unmarshal(uout, &updated); err != nil {
					t.Fatal(err)
				}
				if updated.Artifact.GetLabel() != "Front scan" || updated.Artifact.GetDescription() != "updated" {
					t.Fatalf("%+v", updated.Artifact)
				}
			},
		},
	})
}

func TestIngestArtifactFileMissingPath(t *testing.T) {
	runRPC(t, IngestArtifactFile, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "rejects missing path",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, typeID := sourceFixture(t)
				cout, err := CreateSource(marshalProto(t, &engine.CreateSourceRequest{
					ProjectDir: dir, UserId: userID, SourceTypeId: typeID, Title: "Photo",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateSourceResponse
				if err := proto.Unmarshal(cout, &created); err != nil {
					t.Fatal(err)
				}
				aout, err := CreateArtifact(marshalProto(t, &engine.CreateArtifactRequest{
					ProjectDir: dir, UserId: userID, SourceId: created.Source.Id, Label: "Front",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var art engine.CreateArtifactResponse
				if err := proto.Unmarshal(aout, &art); err != nil {
					t.Fatal(err)
				}
				return &engine.IngestArtifactFileRequest{
					ProjectDir: dir,
					UserId:     userID,
					ArtifactId: art.Artifact.Id,
					Path:       filepath.Join(t.TempDir(), "missing-scan.jpg"),
				}
			},
			wantErr:   true,
			wantErrIs: ingest.ErrMissing,
		},
	})
}

func TestDeleteArtifact(t *testing.T) {
	runRPC(t, DeleteArtifact, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "fileless erases",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, typeID := sourceFixture(t)
				cout, err := CreateSource(marshalProto(t, &engine.CreateSourceRequest{
					ProjectDir: dir, UserId: userID, SourceTypeId: typeID, Title: "Photo",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateSourceResponse
				if err := proto.Unmarshal(cout, &created); err != nil {
					t.Fatal(err)
				}
				aout, err := CreateArtifact(marshalProto(t, &engine.CreateArtifactRequest{
					ProjectDir: dir, UserId: userID, SourceId: created.Source.Id, Label: "Front",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var art engine.CreateArtifactResponse
				if err := proto.Unmarshal(aout, &art); err != nil {
					t.Fatal(err)
				}
				return &engine.DeleteArtifactRequest{
					ProjectDir: dir, UserId: userID, ArtifactId: art.Artifact.Id,
				}
			},
			want: &engine.DeleteArtifactResponse{},
		},
		{
			name: "inbound citation refuses",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, typeID := sourceFixture(t)
				cout, err := CreateSource(marshalProto(t, &engine.CreateSourceRequest{
					ProjectDir: dir, UserId: userID, SourceTypeId: typeID, Title: "Photo",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateSourceResponse
				if err := proto.Unmarshal(cout, &created); err != nil {
					t.Fatal(err)
				}
				aout, err := CreateArtifact(marshalProto(t, &engine.CreateArtifactRequest{
					ProjectDir: dir, UserId: userID, SourceId: created.Source.Id, Label: "Cited",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var art engine.CreateArtifactResponse
				if err := proto.Unmarshal(aout, &art); err != nil {
					t.Fatal(err)
				}
				if _, err := CreateCitationWithObservations(marshalProto(t, &engine.CreateCitationWithObservationsRequest{
					ProjectDir:  dir,
					UserId:      userID,
					ArtifactId:  art.Artifact.Id,
					GraphId:     graphLinking(t, dir, art.Artifact.Id),
					LocatorJson: `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`,
				})); err != nil {
					t.Fatal(err)
				}
				return &engine.DeleteArtifactRequest{
					ProjectDir: dir, UserId: userID, ArtifactId: art.Artifact.Id,
				}
			},
			wantErr:   true,
			wantErrIs: artifacts.ErrInUse,
		},
	})
}

func TestCreateArtifactGraphLinks(t *testing.T) {
	f := newGraphFixture(t)
	second := f.createGraph(t, "Jones", f.page1).GetGraph().GetId()
	linkedTo := func(artifactID string) []string {
		var graphs []string
		for _, g := range f.list(t) {
			if slices.Contains(g.GetArtifactIds(), artifactID) {
				graphs = append(graphs, g.GetName())
			}
		}
		return graphs
	}

	everyResp := f.createArtifactResp(t, f.sourceID, "Every", nil)
	assertGraphEffects(t, everyResp.GetEffects(), f.sourceID)
	if got := linkedTo(everyResp.GetArtifact().GetId()); !slices.Equal(got, []string{"Default", "Jones"}) {
		t.Fatalf("absent links, every graph: %v", got)
	}
	chosenResp := f.createArtifactResp(t, f.sourceID, "Chosen", &engine.GraphLinks{GraphIds: []string{second}})
	assertGraphEffects(t, chosenResp.GetEffects(), f.sourceID)
	if got := linkedTo(chosenResp.GetArtifact().GetId()); !slices.Equal(got, []string{"Jones"}) {
		t.Fatalf("chosen graph only: %v", got)
	}
	noneResp := f.createArtifactResp(t, f.sourceID, "None", &engine.GraphLinks{})
	if got := linkedTo(noneResp.GetArtifact().GetId()); len(got) != 0 {
		t.Fatalf("present and empty, none: %v", got)
	}
	if lists := noneResp.GetEffects().GetLists(); slices.Contains(lists, "evidence_graphs") {
		t.Fatalf("no link row was written, but effects list %v", lists)
	}

	srcB, _ := f.otherSource(t)
	foreign := f.graphsOf(t, srcB)[0].GetId()
	err := rpc(t, CreateArtifact, &engine.CreateArtifactRequest{
		ProjectDir: f.dir, UserId: f.userID, SourceId: f.sourceID, Label: "Foreign",
		GraphLinks: &engine.GraphLinks{GraphIds: []string{foreign}},
	}, &engine.CreateArtifactResponse{})
	if !errors.Is(err, evidencegraphs.ErrInvalid) {
		t.Fatalf("another Source's graph: %v", err)
	}
	var arts engine.GetSourceWorkspaceResponse
	mustRPC(t, GetSourceWorkspace, &engine.GetSourceWorkspaceRequest{ProjectDir: f.dir, SourceId: f.sourceID}, &arts)
	for _, a := range arts.GetArtifacts() {
		if a.GetLabel() == "Foreign" {
			t.Fatal("the refused Artifact was created")
		}
	}
}

func TestUpdateArtifactGraphLinks(t *testing.T) {
	f := newGraphFixture(t)
	// page1 stays in Default and Jones throughout, so no unlink here is a
	// graph's last link.
	jones := f.createGraph(t, "Jones", f.page1).GetGraph().GetId()
	def := f.defaultGraphID
	newPage := func(label string) *engine.CreateArtifactResponse {
		return f.createArtifactResp(t, f.sourceID, label, &engine.GraphLinks{GraphIds: []string{def}})
	}
	linkedTo := func(artifactID string) []string {
		var ids []string
		for _, g := range f.list(t) {
			if slices.Contains(g.GetArtifactIds(), artifactID) {
				ids = append(ids, g.GetId())
			}
		}
		return ids
	}
	labelOf := func(artifactID string) string { return artifactLabel(t, f.dir, f.sourceID, artifactID) }
	update := func(artifactID, label string, links *engine.GraphLinks) (*engine.UpdateArtifactResponse, error) {
		var out engine.UpdateArtifactResponse
		err := rpc(t, UpdateArtifact, &engine.UpdateArtifactRequest{
			ProjectDir: f.dir, UserId: f.userID, ArtifactId: artifactID, Label: label, GraphLinks: links,
		}, &out)
		return &out, err
	}

	t.Run("label and links are one revision", func(t *testing.T) {
		created := newPage("P")
		id := created.GetArtifact().GetId()
		resp, err := update(id, "P renamed", &engine.GraphLinks{GraphIds: []string{jones}})
		if err != nil {
			t.Fatal(err)
		}
		if resp.GetEffects().GetRevision() != created.GetEffects().GetRevision()+1 {
			t.Fatalf("revisions %d -> %d", created.GetEffects().GetRevision(), resp.GetEffects().GetRevision())
		}
		if got := linkedTo(id); !slices.Equal(got, []string{jones}) || labelOf(id) != "P renamed" {
			t.Fatalf("linked %v, label %q", got, labelOf(id))
		}
	})
	t.Run("absent links leave the links alone", func(t *testing.T) {
		id := newPage("P2").GetArtifact().GetId()
		if _, err := update(id, "P2 renamed", nil); err != nil {
			t.Fatal(err)
		}
		if got := linkedTo(id); !slices.Equal(got, []string{def}) || labelOf(id) != "P2 renamed" {
			t.Fatalf("linked %v, label %q", got, labelOf(id))
		}
	})
	t.Run("a refused unlink rolls back the label", func(t *testing.T) {
		id := newPage("P3").GetArtifact().GetId()
		mustRPC(t, CreateCitationWithObservations, &engine.CreateCitationWithObservationsRequest{
			ProjectDir: f.dir, UserId: f.userID, ArtifactId: id, GraphId: graphLinking(t, f.dir, id), LocatorJson: validLocatorJSON,
		}, &engine.CreateCitationWithObservationsResponse{})
		_, err := update(id, "P3 renamed", &engine.GraphLinks{})
		if !errors.Is(err, evidencegraphs.ErrArtifactInUse) {
			t.Fatalf("got %v", err)
		}
		if got := linkedTo(id); !slices.Equal(got, []string{def}) || labelOf(id) != "P3" {
			t.Fatalf("linked %v, label %q", got, labelOf(id))
		}
	})
}

// artifactLabel is the label the Source's workspace shows for an Artifact.
func artifactLabel(t *testing.T, dir, sourceID, artifactID string) string {
	t.Helper()
	var ws engine.GetSourceWorkspaceResponse
	mustRPC(t, GetSourceWorkspace, &engine.GetSourceWorkspaceRequest{ProjectDir: dir, SourceId: sourceID}, &ws)
	for _, a := range ws.GetArtifacts() {
		if a.GetId() == artifactID {
			return a.GetLabel()
		}
	}
	t.Fatalf("artifact %s not found", artifactID)
	return ""
}

// TestUpdateArtifactMetadata saves an Artifact's metadata draft with its edit
// in one revision, and returns the Artifact's board.
func TestUpdateArtifactMetadata(t *testing.T) {
	m := newMetadataOwners(t)
	accessed := newUserField(t, m.dir, m.userID, "Accessed")
	link := newField(t, m.dir, m.userID, "URL", "url")
	provider := m.suggested
	id := m.page1
	labelOf := func(t *testing.T) string {
		t.Helper()
		return artifactLabel(t, m.dir, m.sourceID, id)
	}
	set := func(t *testing.T, field, text string) int64 {
		t.Helper()
		var out engine.SetMetadataResponse
		mustRPC(t, SetMetadata, &engine.SetMetadataRequest{
			ProjectDir: m.dir, UserId: m.userID, Owner: m.artifact, FieldId: field, ValueText: text,
		}, &out)
		return out.GetEffects().GetRevision()
	}
	update := func(label string, draft *engine.MetadataDraft, links *engine.GraphLinks) (*engine.UpdateArtifactResponse, error) {
		var out engine.UpdateArtifactResponse
		err := rpc(t, UpdateArtifact, &engine.UpdateArtifactRequest{
			ProjectDir: m.dir, UserId: m.userID, ArtifactId: id, Label: label, GraphLinks: links, Metadata: draft,
		}, &out)
		return &out, err
	}
	board := func(t *testing.T) []*engine.MetadataWorkspaceEntry {
		t.Helper()
		return boardEntries(t, m.dir, m.artifact)
	}
	valueOf := func(t *testing.T, field string) string {
		t.Helper()
		if e := metadataByField(board(t), field); e != nil {
			return e.GetValueText()
		}
		return ""
	}
	holdsNoMetadataChange := func(t *testing.T) {
		t.Helper()
		for _, ty := range latestRevisionTypes(t, m.dir) {
			if strings.HasPrefix(ty, "artifact_metadata") {
				t.Fatalf("metadata changed: %v", latestRevisionTypes(t, m.dir))
			}
		}
	}

	t.Run("saves the edit and the draft in one revision", func(t *testing.T) {
		set(t, provider, "FamilySearch")
		before := set(t, link, "https://example.org/a")
		resp, err := update("Scan (front)", &engine.MetadataDraft{
			Values: []*engine.MetadataValue{
				{FieldId: provider, ValueText: "Ancestry"},
				{FieldId: accessed, ValueText: "12 March 2025"},
			},
			Layout: []*engine.MetadataLayoutRow{
				{FieldId: accessed, SortOrder: 0},
				{FieldId: provider, SortOrder: 1},
			},
		}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if resp.GetEffects().GetRevision() != before+1 {
			t.Fatalf("revision %d -> %d", before, resp.GetEffects().GetRevision())
		}
		want := []string{
			"artifact update", "artifact_metadata delete", "artifact_metadata update",
			"artifact_metadata create", "artifact_metadata_layout create", "artifact_metadata_layout update",
		}
		if got := latestRevisionTypes(t, m.dir); !slices.Equal(got, want) {
			t.Fatalf("changes %v, want %v", got, want)
		}
		if got := metadataFieldIDs(resp.GetMetadata()); !slices.Equal(got, []string{accessed, provider}) {
			t.Fatalf("board order %v, want accessed, provider", got)
		}
		assertReplyBoard(t, m.dir, m.artifact, resp.GetMetadata())
		if labelOf(t) != "Scan (front)" || valueOf(t, provider) != "Ancestry" || valueOf(t, link) != "" {
			t.Fatalf("label %q provider %q url %q", labelOf(t), valueOf(t, provider), valueOf(t, link))
		}
	})
	t.Run("absent leaves metadata, empty clears", func(t *testing.T) {
		resp, err := update("Scan (back)", nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		holdsNoMetadataChange(t)
		assertReplyBoard(t, m.dir, m.artifact, resp.GetMetadata())
		if valueOf(t, provider) != "Ancestry" {
			t.Fatalf("provider %q, want it held", valueOf(t, provider))
		}
		resp, err = update("Scan (back)", &engine.MetadataDraft{}, nil)
		if err != nil {
			t.Fatal(err)
		}
		assertReplyBoard(t, m.dir, m.artifact, resp.GetMetadata())
		for _, e := range resp.GetMetadata() {
			if e.GetHasValue() {
				t.Fatalf("board still holds %+v", e)
			}
		}
	})
	t.Run("nothing changed writes no revision", func(t *testing.T) {
		// page2 has no layout yet, so one SetMetadata leaves provider at 0.
		other := &engine.MetadataOwner{Level: "artifact", Id: m.page2}
		mustRPC(t, SetMetadata, &engine.SetMetadataRequest{
			ProjectDir: m.dir, UserId: m.userID, Owner: other, FieldId: provider, ValueText: "Ancestry",
		}, &engine.SetMetadataResponse{})
		label := artifactLabel(t, m.dir, m.sourceID, m.page2)
		var resp engine.UpdateArtifactResponse
		mustRPC(t, UpdateArtifact, &engine.UpdateArtifactRequest{
			ProjectDir: m.dir, UserId: m.userID, ArtifactId: m.page2, Label: label,
			Metadata: draftOf(provider, "Ancestry"),
		}, &resp)
		if resp.GetEffects().GetRevision() != 0 {
			t.Fatalf("revision %d, want 0", resp.GetEffects().GetRevision())
		}
		assertReplyBoard(t, m.dir, other, resp.GetMetadata())
	})
	t.Run("a refused value names its field and writes nothing", func(t *testing.T) {
		set(t, provider, "Ancestry")
		label := labelOf(t)
		_, err := update("Renamed", &engine.MetadataDraft{Values: []*engine.MetadataValue{
			{FieldId: provider, ValueText: "Other"}, {FieldId: link, ValueText: "not a link"},
		}}, nil)
		if !errors.Is(err, sourcemetadata.ErrInvalid) || !slices.Equal(errParams(t, err), []string{link}) {
			t.Fatalf("got %v", err)
		}
		if labelOf(t) != label || valueOf(t, provider) != "Ancestry" {
			t.Fatalf("label %q provider %q, want %q and Ancestry", labelOf(t), valueOf(t, provider), label)
		}
	})
	t.Run("a refused unlink drops the draft", func(t *testing.T) {
		pageID := m.createArtifact(t, m.sourceID, "P6", &engine.GraphLinks{GraphIds: []string{m.defaultGraphID}})
		mustRPC(t, CreateCitationWithObservations, &engine.CreateCitationWithObservationsRequest{
			ProjectDir: m.dir, UserId: m.userID, ArtifactId: pageID, GraphId: graphLinking(t, m.dir, pageID), LocatorJson: validLocatorJSON,
		}, &engine.CreateCitationWithObservationsResponse{})
		before := latestRevisionTypes(t, m.dir)
		err := rpc(t, UpdateArtifact, &engine.UpdateArtifactRequest{
			ProjectDir: m.dir, UserId: m.userID, ArtifactId: pageID, Label: "P6 renamed",
			GraphLinks: &engine.GraphLinks{}, Metadata: draftOf(provider, "Ancestry"),
		}, &engine.UpdateArtifactResponse{})
		if !errors.Is(err, evidencegraphs.ErrArtifactInUse) {
			t.Fatalf("got %v", err)
		}
		if got := artifactLabel(t, m.dir, m.sourceID, pageID); got != "P6" {
			t.Fatalf("label %q, want P6", got)
		}
		if after := latestRevisionTypes(t, m.dir); !slices.Equal(after, before) {
			t.Fatalf("revision changes %v, want %v", after, before)
		}
		board := boardEntries(t, m.dir, &engine.MetadataOwner{Level: "artifact", Id: pageID})
		if e := metadataByField(board, provider); e != nil && e.GetHasValue() {
			t.Fatalf("provider saved beside the refused unlink: %+v", e)
		}
	})
}

// filelessArtifact is a project with one Artifact that has no File.
func filelessArtifact(t *testing.T) (dir, userID, artifactID string) {
	t.Helper()
	dir, userID, typeID := sourceFixture(t)
	var src engine.CreateSourceResponse
	mustRPC(t, CreateSource, &engine.CreateSourceRequest{ProjectDir: dir, UserId: userID, SourceTypeId: typeID, Title: "Scan"}, &src)
	var art engine.CreateArtifactResponse
	mustRPC(t, CreateArtifact, &engine.CreateArtifactRequest{ProjectDir: dir, UserId: userID, SourceId: src.Source.Id, Label: "Page"}, &art)
	return dir, userID, art.Artifact.Id
}

func TestIngestStagesBeforeSession(t *testing.T) {
	t.Cleanup(func() { _ = catalogsession.CloseAll() })
	path := filepath.Join(t.TempDir(), "page.png")
	writePNG(t, path)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("the copy finishes while another call holds the session", func(t *testing.T) {
		dir, userID, artifactID := filelessArtifact(t)
		release := catalogsessiontest.Hold(t, dir)
		type result struct {
			out []byte
			err error
		}
		done := make(chan result, 1)
		in := marshalProto(t, &engine.IngestArtifactFileRequest{ProjectDir: dir, UserId: userID, ArtifactId: artifactID, Path: path})
		go func() {
			out, err := IngestArtifactFile(in)
			done <- result{out, err}
		}()
		waitForTemp(t, dir, ".ingest-", info.Size())
		release()
		res := <-done
		if res.err != nil {
			t.Fatal(res.err)
		}
		var got engine.IngestArtifactFileResponse
		if err := proto.Unmarshal(res.out, &got); err != nil {
			t.Fatal(err)
		}

		dir2, userID2, artifactID2 := filelessArtifact(t)
		var want engine.IngestArtifactFileResponse
		mustRPC(t, IngestArtifactFile, &engine.IngestArtifactFileRequest{ProjectDir: dir2, UserId: userID2, ArtifactId: artifactID2, Path: path}, &want)
		if got.Reused != want.Reused || got.File.GetRelPath() != want.File.GetRelPath() || got.File.GetByteSize() != want.File.GetByteSize() ||
			got.Artifact.GetLabel() != want.Artifact.GetLabel() || got.File.GetId() == "" {
			t.Fatalf("blocked run %v, unblocked run %v", &got, &want)
		}
		if temps := objectTemps(t, dir, ".ingest-"); len(temps) != 0 {
			t.Fatalf("temps left: %v", temps)
		}
	})

	t.Run("an Artifact that has a File refuses and leaves nothing", func(t *testing.T) {
		dir, userID, artifactID := filelessArtifact(t)
		var first engine.IngestArtifactFileResponse
		mustRPC(t, IngestArtifactFile, &engine.IngestArtifactFileRequest{ProjectDir: dir, UserId: userID, ArtifactId: artifactID, Path: path}, &first)
		before := objectRelPaths(t, dir)
		other := filepath.Join(t.TempDir(), "other.pdf")
		writePDF(t, other)
		_, err := IngestArtifactFile(marshalProto(t, &engine.IngestArtifactFileRequest{ProjectDir: dir, UserId: userID, ArtifactId: artifactID, Path: other}))
		if !errors.Is(err, artifacts.ErrFileAlreadyAttached) {
			t.Fatalf("err = %v, want ErrFileAlreadyAttached", err)
		}
		if got := objectRelPaths(t, dir); !slices.Equal(got, before) {
			t.Fatalf("objects %v, want %v", got, before)
		}
		if temps := objectTemps(t, dir, ".ingest-"); len(temps) != 0 {
			t.Fatalf("temps left: %v", temps)
		}
	})
}

// TestCreateArtifactMetadata saves a metadata draft with the Artifact's create
// in one revision, and refuses a bad draft with nothing written.
func TestCreateArtifactMetadata(t *testing.T) {
	dir, userID, typeID := sourceFixture(t)
	var src engine.CreateSourceResponse
	mustRPC(t, CreateSource, &engine.CreateSourceRequest{ProjectDir: dir, UserId: userID, SourceTypeId: typeID, Title: "Census"}, &src)
	sourceID := src.GetSource().GetId()
	provider := newUserField(t, dir, userID, "Provider")
	link := newField(t, dir, userID, "URL", "url")
	mustRPC(t, AssignTypeField, &engine.AssignTypeFieldRequest{
		ProjectDir: dir, TypeId: typeID, FieldId: link, Level: "artifact",
	}, &engine.AssignTypeFieldResponse{})
	owner := func(id string) *engine.MetadataOwner { return &engine.MetadataOwner{Level: "artifact", Id: id} }
	artifactCount := func() int {
		var ws engine.GetSourceWorkspaceResponse
		mustRPC(t, GetSourceWorkspace, &engine.GetSourceWorkspaceRequest{ProjectDir: dir, SourceId: sourceID}, &ws)
		return len(ws.GetArtifacts())
	}

	var plain engine.CreateArtifactResponse
	mustRPC(t, CreateArtifact, &engine.CreateArtifactRequest{ProjectDir: dir, UserId: userID, SourceId: sourceID, Label: "Plain"}, &plain)
	plainTypes := latestRevisionTypes(t, dir)
	assertReplyBoard(t, dir, owner(plain.GetArtifact().GetId()), plain.GetMetadata())
	if e := metadataByField(plain.GetMetadata(), link); e == nil || !e.GetSuggested() || e.GetHasValue() {
		t.Fatalf("plain create board %+v, want url suggested and empty", plain.GetMetadata())
	}
	if slices.ContainsFunc(plainTypes, func(s string) bool { return s == "artifact_metadata create" }) {
		t.Fatalf("a create without metadata wrote metadata: %v", plainTypes)
	}

	var resp engine.CreateArtifactResponse
	mustRPC(t, CreateArtifact, &engine.CreateArtifactRequest{
		ProjectDir: dir, UserId: userID, SourceId: sourceID, Label: "Page",
		Metadata: draftOf(provider, "FamilySearch"),
	}, &resp)
	got := latestRevisionTypes(t, dir)
	want := append(slices.Clone(plainTypes), "artifact_metadata create", "artifact_metadata_layout create")
	if !slices.Equal(got, want) {
		t.Fatalf("changes %v, want the plain create's %v then the value and layout creates", got, plainTypes)
	}
	assertReplyBoard(t, dir, owner(resp.GetArtifact().GetId()), resp.GetMetadata())
	if e := metadataByField(resp.GetMetadata(), link); e == nil || !e.GetSuggested() || e.GetHasValue() {
		t.Fatalf("board %+v, want url suggested and empty", resp.GetMetadata())
	}
	if e := metadataByField(resp.GetMetadata(), provider); e == nil || e.GetValueText() != "FamilySearch" {
		t.Fatalf("board %+v, want provider FamilySearch", resp.GetMetadata())
	}

	before := artifactCount()
	err := rpc(t, CreateArtifact, &engine.CreateArtifactRequest{
		ProjectDir: dir, UserId: userID, SourceId: sourceID, Label: "Bad",
		Metadata: draftOf("01890000-0000-7000-8000-000000000009", "x"),
	}, &engine.CreateArtifactResponse{})
	if !errors.Is(err, sourcemetadata.ErrInvalid) {
		t.Fatalf("unknown field: %v", err)
	}
	if artifactCount() != before {
		t.Fatal("an Artifact was created beside the refused draft")
	}
	err = rpc(t, CreateArtifact, &engine.CreateArtifactRequest{
		ProjectDir: dir, UserId: userID, SourceId: sourceID, Label: "Bad",
		Metadata: draftOf(link, "not a link"),
	}, &engine.CreateArtifactResponse{})
	if !errors.Is(err, sourcemetadata.ErrInvalid) || !slices.Equal(errParams(t, err), []string{link}) {
		t.Fatalf("bad url: %v, params %v", err, errParams(t, err))
	}
	if artifactCount() != before {
		t.Fatal("an Artifact was created beside the refused url")
	}
	err = rpc(t, CreateArtifact, &engine.CreateArtifactRequest{
		ProjectDir: dir, UserId: userID, SourceId: sourceID, Label: "Bad",
		Metadata: draftOf("nope", "x"),
	}, &engine.CreateArtifactResponse{})
	if !errors.Is(err, errInvalidID) {
		t.Fatalf("malformed field id: %v", err)
	}
	if artifactCount() != before {
		t.Fatal("an Artifact was created beside the malformed draft")
	}
}
