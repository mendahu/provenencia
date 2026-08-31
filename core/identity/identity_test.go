package identity

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

func TestIdentity(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, dir string)
	}{
		{
			name: "round trip json",
			run: func(t *testing.T, dir string) {
				id, err := Mint("Jake Robins")
				if err != nil {
					t.Fatal(err)
				}
				if err := Save(dir, id); err != nil {
					t.Fatal(err)
				}
				got, err := Load(dir)
				if err != nil {
					t.Fatal(err)
				}
				if got.UserID != id.UserID || got.DisplayName != "Jake Robins" || got.Ref != id.Ref {
					t.Fatalf("got %+v want %+v", got, id)
				}
				raw, err := os.ReadFile(filepath.Join(dir, FileName))
				if err != nil {
					t.Fatal(err)
				}
				var m map[string]string
				if err := json.Unmarshal(raw, &m); err != nil {
					t.Fatal(err)
				}
				if m["user_id"] != id.UserID.String() || m["display_name"] != "Jake Robins" || m["ref"] != id.Ref {
					t.Fatalf("json %v", m)
				}
			},
		},
		{
			name: "load missing is ErrNotFound",
			run: func(t *testing.T, dir string) {
				_, err := Load(dir)
				if err != ErrNotFound {
					t.Fatalf("got %v want %v", err, ErrNotFound)
				}
			},
		},
		{
			name: "save rejects blank name",
			run: func(t *testing.T, dir string) {
				id, err := Mint("Jake")
				if err != nil {
					t.Fatal(err)
				}
				id.DisplayName = "  \t"
				if err := Save(dir, id); err != ErrInvalidName {
					t.Fatalf("got %v want %v", err, ErrInvalidName)
				}
			},
		},
		{
			name: "mint is uuid v7 and unique",
			run: func(t *testing.T, dir string) {
				a, err := Mint("A")
				if err != nil {
					t.Fatal(err)
				}
				b, err := Mint("B")
				if err != nil {
					t.Fatal(err)
				}
				if a.UserID.Version() != 7 || b.UserID.Version() != 7 {
					t.Fatalf("versions %d %d", a.UserID.Version(), b.UserID.Version())
				}
				if a.UserID == b.UserID {
					t.Fatal("ids not unique")
				}
				if a.UserID == uuid.Nil {
					t.Fatal("nil uuid")
				}
				if a.Ref == "" || b.Ref == "" || a.Ref == b.Ref {
					t.Fatalf("refs %q %q", a.Ref, b.Ref)
				}
			},
		},
		{
			name: "load upgrades missing ref",
			run: func(t *testing.T, dir string) {
				id, err := Mint("Jake")
				if err != nil {
					t.Fatal(err)
				}
				raw, err := json.Marshal(map[string]string{
					"user_id":      id.UserID.String(),
					"display_name": "Jake",
				})
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, FileName), raw, 0o600); err != nil {
					t.Fatal(err)
				}
				got, err := Load(dir)
				if err != nil {
					t.Fatal(err)
				}
				if got.Ref == "" || got.UserID != id.UserID {
					t.Fatalf("%+v", got)
				}
				again, err := Load(dir)
				if err != nil {
					t.Fatal(err)
				}
				if again.Ref != got.Ref {
					t.Fatalf("ref changed on reload")
				}
			},
		},
		{
			name: "overwrite keeps user_id name is cache",
			run: func(t *testing.T, dir string) {
				id, err := Mint("Jake")
				if err != nil {
					t.Fatal(err)
				}
				if err := Save(dir, id); err != nil {
					t.Fatal(err)
				}
				id.DisplayName = "Jake R."
				if err := Save(dir, id); err != nil {
					t.Fatal(err)
				}
				got, err := Load(dir)
				if err != nil {
					t.Fatal(err)
				}
				if got.UserID != id.UserID {
					t.Fatalf("user_id changed")
				}
				if got.DisplayName != "Jake R." {
					t.Fatalf("got name %q", got.DisplayName)
				}
			},
		},
		{
			name: "save rejects non v7",
			run: func(t *testing.T, dir string) {
				id := Identity{UserID: uuid.New(), DisplayName: "X"}
				if err := Save(dir, id); err != ErrInvalidID {
					t.Fatalf("got %v want %v", err, ErrInvalidID)
				}
			},
		},
		{
			name: "mint rejects blank name",
			run: func(t *testing.T, dir string) {
				_, err := Mint("   ")
				if err != ErrInvalidName {
					t.Fatalf("got %v want %v", err, ErrInvalidName)
				}
			},
		},
		{
			name: "remove missing is ok",
			run: func(t *testing.T, dir string) {
				if err := Remove(dir); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "remove then load is not found",
			run: func(t *testing.T, dir string) {
				id, err := Mint("Jake")
				if err != nil {
					t.Fatal(err)
				}
				if err := Save(dir, id); err != nil {
					t.Fatal(err)
				}
				if err := Remove(dir); err != nil {
					t.Fatal(err)
				}
				if _, err := Load(dir); err != ErrNotFound {
					t.Fatalf("got %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t, t.TempDir())
		})
	}
}
