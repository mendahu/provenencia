package handlers

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/catalogsession"
	"github.com/mendahu/provenencia/core/catalogsession/catalogsessiontest"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/evidencegraphs"
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
	labelOf := func(artifactID string) string {
		var ws engine.GetSourceWorkspaceResponse
		mustRPC(t, GetSourceWorkspace, &engine.GetSourceWorkspaceRequest{ProjectDir: f.dir, SourceId: f.sourceID}, &ws)
		for _, a := range ws.GetArtifacts() {
			if a.GetId() == artifactID {
				return a.GetLabel()
			}
		}
		t.Fatalf("artifact %s not found", artifactID)
		return ""
	}
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
