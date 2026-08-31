package ref

import (
	"strings"
	"testing"
)

func TestMint(t *testing.T) {
	tests := []struct {
		name    string
		prefix  string
		wantErr error
	}{
		{name: "usr", prefix: "USR"},
		{name: "const", prefix: PrefixUser},
		{name: "lower", prefix: "usr"},
		{name: "src", prefix: PrefixSource},
		{name: "short", prefix: "US", wantErr: ErrInvalidPrefix},
		{name: "long", prefix: "USER", wantErr: ErrInvalidPrefix},
		{name: "digits", prefix: "US1", wantErr: ErrInvalidPrefix},
		{name: "blank", prefix: "  ", wantErr: ErrInvalidPrefix},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Mint(tt.prefix)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("got %v want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !Valid(got) {
				t.Fatalf("invalid %q", got)
			}
			wantPrefix := strings.ToUpper(strings.TrimSpace(tt.prefix))
			if len(wantPrefix) == 3 {
				if !strings.HasPrefix(got, wantPrefix+"-") {
					t.Fatalf("got %q", got)
				}
			}
		})
	}
}

func TestValid(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{name: "ok", in: "USR-F4N2P", want: true},
		{name: "src", in: "SRC-3K9M2", want: true},
		{name: "bad letter I", in: "USR-F4I2P", want: false},
		{name: "short token", in: "USR-F4N2", want: false},
		{name: "no dash", in: "USRF4N2P", want: false},
		{name: "blank", in: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if Valid(tt.in) != tt.want {
				t.Fatalf("Valid(%q)=%v want %v", tt.in, Valid(tt.in), tt.want)
			}
		})
	}
}

func TestMintUnique(t *testing.T) {
	seen := map[string]struct{}{}
	for i := 0; i < 50; i++ {
		got, err := Mint(PrefixUser)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := seen[got]; ok {
			t.Fatalf("duplicate %q", got)
		}
		seen[got] = struct{}{}
	}
}
