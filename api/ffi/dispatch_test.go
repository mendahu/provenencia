package ffi

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mendahu/provenance/api/proto/engine"
	"github.com/mendahu/provenance/core"
	"github.com/mendahu/provenance/core/database"
	"github.com/mendahu/provenance/core/identity"
	"google.golang.org/protobuf/proto"
)

func TestCall(t *testing.T) {
	mustMarshal := func(m proto.Message) []byte {
		t.Helper()
		b, err := proto.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}

	tests := []struct {
		name    string
		method  int32
		in      []byte
		wantErr bool
		check   func(t *testing.T, out []byte)
	}{
		{
			name:   "ping echoes message",
			method: MethodPing,
			in:     mustMarshal(&engine.PingRequest{Message: "hello"}),
			check: func(t *testing.T, out []byte) {
				var resp engine.PingResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				if resp.GetMessage() != "hello" {
					t.Fatalf("got %q", resp.GetMessage())
				}
			},
		},
		{
			name:   "ping echoes empty string",
			method: MethodPing,
			in:     mustMarshal(&engine.PingRequest{Message: ""}),
			check: func(t *testing.T, out []byte) {
				var resp engine.PingResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				if resp.GetMessage() != "" {
					t.Fatalf("got %q", resp.GetMessage())
				}
			},
		},
		{
			name:   "get version matches core.Version",
			method: MethodGetVersion,
			in:     mustMarshal(&engine.GetVersionRequest{}),
			check: func(t *testing.T, out []byte) {
				var resp engine.GetVersionResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				if resp.GetVersion() != core.Version {
					t.Fatalf("got %q want %q", resp.GetVersion(), core.Version)
				}
			},
		},
		{
			name:   "install identity not found",
			method: MethodGetInstallIdentity,
			in:     mustMarshal(&engine.GetInstallIdentityRequest{IdentityDir: t.TempDir()}),
			check: func(t *testing.T, out []byte) {
				var resp engine.GetInstallIdentityResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				if resp.GetFound() {
					t.Fatal("expected not found")
				}
			},
		},
		{
			name:    "unspecified method",
			method:  MethodUnspecified,
			wantErr: true,
		},
		{
			name:    "ping rejects invalid protobuf",
			method:  MethodPing,
			in:      []byte{0xff, 0xff, 0xff, 0xff},
			wantErr: true,
		},
		{
			name:    "unknown method",
			method:  99,
			wantErr: true,
		},
		{
			name:    "complete rejects blank names",
			method:  MethodCompleteOnboarding,
			in:      mustMarshal(&engine.CompleteOnboardingRequest{IdentityDir: t.TempDir(), ParentDir: t.TempDir()}),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := Call(tt.method, tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if tt.check != nil {
				tt.check(t, out)
			}
		})
	}

	t.Run("install identity found", func(t *testing.T) {
		dir := t.TempDir()
		id, err := identity.Mint("Jake")
		if err != nil {
			t.Fatal(err)
		}
		if err := identity.Save(dir, id); err != nil {
			t.Fatal(err)
		}
		out, err := Call(MethodGetInstallIdentity, mustMarshal(&engine.GetInstallIdentityRequest{IdentityDir: dir}))
		if err != nil {
			t.Fatal(err)
		}
		var resp engine.GetInstallIdentityResponse
		if err := proto.Unmarshal(out, &resp); err != nil {
			t.Fatal(err)
		}
		if !resp.GetFound() || resp.GetDisplayName() != "Jake" || resp.GetUserId() != id.UserID.String() {
			t.Fatalf("%+v", resp)
		}
	})

	t.Run("complete then get identity", func(t *testing.T) {
		ident := t.TempDir()
		parent := t.TempDir()
		out, err := Call(MethodCompleteOnboarding, mustMarshal(&engine.CompleteOnboardingRequest{
			IdentityDir: ident,
			ParentDir:   parent,
			DisplayName: "Jake Robins",
			FamilyName:  "Robins Family",
		}))
		if err != nil {
			t.Fatal(err)
		}
		var done engine.CompleteOnboardingResponse
		if err := proto.Unmarshal(out, &done); err != nil {
			t.Fatal(err)
		}
		if done.GetDisplayName() != "Jake Robins" || done.GetUserId() == "" {
			t.Fatalf("%+v", done)
		}
		if filepath.Base(done.GetProjectDir()) != "Robins Family"+database.Suffix {
			t.Fatalf("project %s", done.GetProjectDir())
		}
		if _, err := os.Stat(filepath.Join(done.GetProjectDir(), "provenance.sqlite")); err != nil {
			t.Fatal(err)
		}
		got, err := Call(MethodGetInstallIdentity, mustMarshal(&engine.GetInstallIdentityRequest{IdentityDir: ident}))
		if err != nil {
			t.Fatal(err)
		}
		var idresp engine.GetInstallIdentityResponse
		if err := proto.Unmarshal(got, &idresp); err != nil {
			t.Fatal(err)
		}
		if !idresp.GetFound() || idresp.GetUserId() != done.GetUserId() {
			t.Fatalf("%+v", idresp)
		}
	})

	t.Run("complete already exists", func(t *testing.T) {
		ident := t.TempDir()
		parent := t.TempDir()
		req := mustMarshal(&engine.CompleteOnboardingRequest{
			IdentityDir: ident,
			ParentDir:   parent,
			DisplayName: "Jake",
			FamilyName:  "Dup",
		})
		if _, err := Call(MethodCompleteOnboarding, req); err != nil {
			t.Fatal(err)
		}
		if _, err := Call(MethodCompleteOnboarding, req); err == nil {
			t.Fatal("expected error")
		}
	})
}
