package handlers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mendahu/provenance/api/proto/engine"
	"github.com/mendahu/provenance/core/database"
	"github.com/mendahu/provenance/core/identity"
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
	want    proto.Message // populated fields must match (proto3 zeros are skipped)
	exact   bool          // proto.Equal(want) including zeros; requires want
	after   func(*testing.T, []byte, proto.Message)
}

func runRPC(t *testing.T, fn func([]byte) ([]byte, error), tests []rpcTest) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
	if _, err := os.Stat(filepath.Join(done.GetProjectDir(), "provenance.sqlite")); err != nil {
		t.Fatal(err)
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
