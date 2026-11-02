package handlers

import (
	"database/sql"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/catalogsession"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/identity"
	"github.com/mendahu/provenencia/core/onboarding"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// rpcTest is one handler invocation. reqFn wins over req; raw is unparsed bytes.
type rpcTest struct {
	name    string
	req     proto.Message
	reqFn   func(*testing.T) proto.Message
	raw     []byte
	calls   int // default 1; last call is the one wantErr/want/after apply to
	wantErr bool
	// When wantErr is set, optionally require errors.Is(err, wantErrIs).
	wantErrIs error
	want      proto.Message // populated fields must match (proto3 zeros are skipped)
	exact     bool          // proto.Equal(want) including zeros; requires want
	after     func(*testing.T, []byte, proto.Message)
}

func runRPC(t *testing.T, fn func([]byte) ([]byte, error), tests []rpcTest) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() { _ = catalogsession.CloseAll() })
			in, req := tt.input(t)
			n := tt.calls
			if n == 0 {
				n = 1
			}
			var out []byte
			var err error
			for i := 0; i < n; i++ {
				out, err = fn(in)
				if i < n-1 && err != nil {
					t.Fatalf("call %d: %v", i+1, err)
				}
			}
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("got %v want %v", err, tt.wantErrIs)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if tt.want != nil {
				if tt.exact {
					assertWantEqual(t, out, tt.want)
				} else {
					assertWantFields(t, out, tt.want)
				}
			}
			if tt.after != nil {
				tt.after(t, out, req)
			}
		})
	}
}

func (tt rpcTest) input(t *testing.T) ([]byte, proto.Message) {
	t.Helper()
	switch {
	case tt.raw != nil:
		return tt.raw, nil
	case tt.reqFn != nil:
		m := tt.reqFn(t)
		return marshalProto(t, m), m
	case tt.req != nil:
		return marshalProto(t, tt.req), tt.req
	default:
		return nil, nil
	}
}

func marshalProto(t *testing.T, m proto.Message) []byte {
	t.Helper()
	b, err := proto.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// rpc calls fn with req and decodes the reply into resp.
func rpc(t *testing.T, fn func([]byte) ([]byte, error), req, resp proto.Message) error {
	t.Helper()
	out, err := fn(marshalProto(t, req))
	if err != nil {
		return err
	}
	if err := proto.Unmarshal(out, resp); err != nil {
		t.Fatal(err)
	}
	return nil
}

func mustRPC(t *testing.T, fn func([]byte) ([]byte, error), req, resp proto.Message) {
	t.Helper()
	if err := rpc(t, fn, req, resp); err != nil {
		t.Fatal(err)
	}
}

// pngSource is a Source whose one Artifact holds an ingested PNG.
type pngSource struct {
	dir, userID, sourceID, artifactID, fileID string
}

// sourceWithPNG creates a Source titled "Scan" whose one Artifact, labelled
// "Page", holds the PNG ingested as page.png. A test that reads its own Source
// title or Artifact label, or ingests another type, builds its own.
func sourceWithPNG(t *testing.T) pngSource {
	t.Helper()
	dir, userID, typeID := sourceFixture(t)
	var created engine.CreateSourceResponse
	mustRPC(t, CreateSource, &engine.CreateSourceRequest{
		ProjectDir: dir, UserId: userID, SourceTypeId: typeID, Title: "Scan",
	}, &created)
	var art engine.CreateArtifactResponse
	mustRPC(t, CreateArtifact, &engine.CreateArtifactRequest{
		ProjectDir: dir, UserId: userID, SourceId: created.Source.Id, Label: "Page",
	}, &art)
	path := filepath.Join(t.TempDir(), "page.png")
	writePNG(t, path)
	var ingested engine.IngestArtifactFileResponse
	mustRPC(t, IngestArtifactFile, &engine.IngestArtifactFileRequest{
		ProjectDir: dir, UserId: userID, ArtifactId: art.Artifact.Id, Path: path,
	}, &ingested)
	return pngSource{
		dir: dir, userID: userID, sourceID: created.Source.Id,
		artifactID: art.Artifact.Id, fileID: ingested.File.Id,
	}
}

// damageCatalog runs one statement against the project's catalog, as a
// hand-edited or crafted project would, and closes the held session so the next
// read opens fresh.
func damageCatalog(t *testing.T, dir, query string, args ...any) {
	t.Helper()
	err := catalogsession.Do(dir, func(c *database.Catalog) error {
		db, err := c.DB()
		if err != nil {
			return err
		}
		_, err = db.Exec(query, args...)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := catalogsession.CloseAll(); err != nil {
		t.Fatal(err)
	}
}

// objectTemps lists the files at the top of dir/objects whose name starts with
// prefix (".ingest-" or ".stage-"), with their sizes.
func objectTemps(t *testing.T, dir, prefix string) map[string]int64 {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(dir, "objects"))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		t.Fatal(err)
	}
	out := map[string]int64{}
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), prefix) {
			continue
		}
		if info, err := e.Info(); err == nil {
			out[e.Name()] = info.Size()
		}
	}
	return out
}

// waitForTemp polls dir/objects every 10 ms, for up to 5 s, until a top-level
// file named prefix* of at least minSize bytes exists.
func waitForTemp(t *testing.T, dir, prefix string, minSize int64) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		for _, size := range objectTemps(t, dir, prefix) {
			if size >= minSize {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("no %s* temp of at least %d bytes in %s/objects: %v", prefix, minSize, dir, objectTemps(t, dir, prefix))
}

func assertWantEqual(t *testing.T, out []byte, want proto.Message) {
	t.Helper()
	got := want.ProtoReflect().New().Interface().(proto.Message)
	if err := proto.Unmarshal(out, got); err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func assertWantFields(t *testing.T, out []byte, want proto.Message) {
	t.Helper()
	got := want.ProtoReflect().New().Interface().(proto.Message)
	if err := proto.Unmarshal(out, got); err != nil {
		t.Fatal(err)
	}
	want.ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		gv := got.ProtoReflect().Get(fd)
		if !gv.Equal(v) {
			t.Fatalf("%s: got %v want %v", fd.Name(), gv, v)
		}
		return true
	})
}

func saveIdentity(t *testing.T, name string) (dir string, id identity.Identity) {
	t.Helper()
	dir = t.TempDir()
	id, err := identity.Mint(name)
	if err != nil {
		t.Fatal(err)
	}
	if err := identity.Save(dir, id); err != nil {
		t.Fatal(err)
	}
	return dir, id
}

func sourceFixture(t *testing.T) (projectDir, userID, typeID string) {
	t.Helper()
	res, err := onboarding.Complete(t.TempDir(), t.TempDir(), "Jake", "Sources")
	if err != nil {
		t.Fatal(err)
	}
	listOut, err := ListSourceTypes(marshalProto(t, &engine.ListSourceTypesRequest{ProjectDir: res.ProjectDir}))
	if err != nil {
		t.Fatal(err)
	}
	var types engine.ListSourceTypesResponse
	if err := proto.Unmarshal(listOut, &types); err != nil {
		t.Fatal(err)
	}
	if len(types.Types) == 0 {
		t.Fatal("expected seeded types")
	}
	return res.ProjectDir, res.Identity.UserID.String(), types.Types[0].Id
}

func assertIdentityNotFound(t *testing.T, out []byte, _ proto.Message) {
	t.Helper()
	var resp engine.GetInstallIdentityResponse
	if err := proto.Unmarshal(out, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.GetFound() {
		t.Fatal("expected not found")
	}
}
func assertOnboardingWroteProject(t *testing.T, out []byte, familyBase string) {
	t.Helper()
	var done engine.CompleteOnboardingResponse
	if err := proto.Unmarshal(out, &done); err != nil {
		t.Fatal(err)
	}
	if done.GetUserId() == "" {
		t.Fatal("empty user_id")
	}
	if filepath.Base(done.GetProjectDir()) != familyBase+database.Suffix {
		t.Fatalf("project %s", done.GetProjectDir())
	}
	if _, err := os.Stat(filepath.Join(done.GetProjectDir(), "provenencia.sqlite")); err != nil {
		t.Fatal(err)
	}
	if done.GetProject().GetUuid() == "" {
		t.Fatal("empty project uuid")
	}
}

func assertIdentityMatchesComplete(t *testing.T, out []byte, req proto.Message) {
	t.Helper()
	var done engine.CompleteOnboardingResponse
	if err := proto.Unmarshal(out, &done); err != nil {
		t.Fatal(err)
	}
	cr, ok := req.(*engine.CompleteOnboardingRequest)
	if !ok {
		t.Fatalf("req %T", req)
	}
	got, err := GetInstallIdentity(marshalProto(t, &engine.GetInstallIdentityRequest{IdentityDir: cr.GetIdentityDir()}))
	if err != nil {
		t.Fatal(err)
	}
	assertWantFields(t, got, &engine.GetInstallIdentityResponse{
		Found:  true,
		UserId: done.GetUserId(),
	})
}

func assertActiveMatchesComplete(t *testing.T, out []byte, req proto.Message) {
	t.Helper()
	var done engine.CompleteOnboardingResponse
	if err := proto.Unmarshal(out, &done); err != nil {
		t.Fatal(err)
	}
	cr, ok := req.(*engine.CompleteOnboardingRequest)
	if !ok {
		t.Fatalf("req %T", req)
	}
	got, err := GetActiveProject(marshalProto(t, &engine.GetActiveProjectRequest{IdentityDir: cr.GetIdentityDir()}))
	if err != nil {
		t.Fatal(err)
	}
	assertWantFields(t, got, &engine.GetActiveProjectResponse{
		Found:      true,
		ProjectDir: done.GetProjectDir(),
	})
}

func assertAuditActionPresent(t *testing.T, projectDir, wantAction string) {
	t.Helper()
	err := catalogsession.Do(projectDir, func(c *database.Catalog) error {
		db, err := c.DB()
		if err != nil {
			return err
		}
		var n int
		if err := db.QueryRow(`SELECT COUNT(*) FROM audit_transactions WHERE action_type = ?`, wantAction).Scan(&n); err != nil {
			return err
		}
		if n < 1 {
			t.Fatalf("expected at least one audit transaction with action_type %q", wantAction)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func assertLatestAuditAction(t *testing.T, projectDir, wantAction string) {
	t.Helper()
	err := catalogsession.Do(projectDir, func(c *database.Catalog) error {
		db, err := c.DB()
		if err != nil {
			return err
		}
		var actionType string
		if err := db.QueryRow(`SELECT action_type FROM audit_transactions ORDER BY revision DESC LIMIT 1`).Scan(&actionType); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				t.Fatal("expected audit transaction")
			}
			return err
		}
		if actionType != wantAction {
			t.Fatalf("audit action_type = %q, want %q", actionType, wantAction)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
