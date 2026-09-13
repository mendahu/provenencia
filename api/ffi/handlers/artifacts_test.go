package handlers

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database/artifacts"
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
				path := filepath.Join(t.TempDir(), "scan.jpg")
				if err := os.WriteFile(path, []byte("jpeg-bytes"), 0o644); err != nil {
					t.Fatal(err)
				}
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
				if ingested.File.GetOriginalFilename() != "scan.jpg" || ingested.GetReused() {
					t.Fatalf("%+v", ingested)
				}
				if ingested.Artifact.GetFileId() == "" {
					t.Fatal("expected file_id")
				}
				assertAuditActionPresent(t, cr.ProjectDir, "create_file")
				assertAuditActionPresent(t, cr.ProjectDir, "update_artifact")
				assertLatestAuditAction(t, cr.ProjectDir, "set_source_cover")

				path2 := filepath.Join(t.TempDir(), "scan2.jpg")
				if err := os.WriteFile(path2, []byte("jpeg-two"), 0o644); err != nil {
					t.Fatal(err)
				}
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
			wantErrIs: ingest.ErrInvalid,
		},
	})
}
