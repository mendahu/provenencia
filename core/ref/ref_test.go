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
		{name: "token starting with C", in: "PER-C4N2P", want: true},
		{name: "candidate form", in: "PER-C-7KD45", want: false},
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

func TestMintCandidate(t *testing.T) {
	tests := []struct {
		name    string
		prefix  string
		wantErr error
	}{
		{name: "person", prefix: "PER"},
		{name: "event", prefix: "EVT"},
		{name: "lower", prefix: "plc"},
		{name: "source node", prefix: "SRN"},
		{name: "short", prefix: "PE", wantErr: ErrInvalidPrefix},
		{name: "long", prefix: "PERS", wantErr: ErrInvalidPrefix},
		{name: "digits", prefix: "PE1", wantErr: ErrInvalidPrefix},
		{name: "blank", prefix: "  ", wantErr: ErrInvalidPrefix},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MintCandidate(tt.prefix)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("got %v want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !ValidCandidate(got) {
				t.Fatalf("invalid candidate %q", got)
			}
			// A candidate ref must never pass as a catalog ref.
			if Valid(got) {
				t.Fatalf("%q also matched the catalog form", got)
			}
			wantPrefix := strings.ToUpper(strings.TrimSpace(tt.prefix)) + "-" + candidateMarker + "-"
			if !strings.HasPrefix(got, wantPrefix) {
				t.Fatalf("got %q want prefix %q", got, wantPrefix)
			}
		})
	}
}

func TestValidCandidate(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{name: "ok", in: "PER-C-7KD45", want: true},
		{name: "event", in: "EVT-C-3K9M2", want: true},
		{name: "catalog form", in: "PER-7KD45", want: false},
		{name: "catalog token starting with C", in: "PER-C4N2P", want: false},
		{name: "bad letter I", in: "PER-C-F4I2P", want: false},
		{name: "wrong marker", in: "PER-X-7KD45", want: false},
		{name: "double marker", in: "PER-C-C-7KD45", want: false},
		{name: "lowercase", in: "per-c-7kd45", want: false},
		{name: "short token", in: "PER-C-7KD4", want: false},
		{name: "blank", in: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if ValidCandidate(tt.in) != tt.want {
				t.Fatalf("ValidCandidate(%q)=%v want %v", tt.in, ValidCandidate(tt.in), tt.want)
			}
		})
	}
}

func TestMintCandidateUnique(t *testing.T) {
	seen := map[string]struct{}{}
	for i := 0; i < 50; i++ {
		got, err := MintCandidate("PER")
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := seen[got]; ok {
			t.Fatalf("duplicate %q", got)
		}
		seen[got] = struct{}{}
	}
}

func TestValidatePrefix(t *testing.T) {
	tests := []struct {
		name    string
		prefix  string
		want    string
		wantErr error
	}{
		{name: "person", prefix: "PER", want: "PER"},
		{name: "lower normalizes", prefix: "per", want: "PER"},
		{name: "source node allowed", prefix: "SRN", want: "SRN"},
		{name: "reserved user", prefix: "USR", wantErr: ErrReservedPrefix},
		{name: "reserved source", prefix: "SRC", wantErr: ErrReservedPrefix},
		{name: "reserved artifact", prefix: "ART", wantErr: ErrReservedPrefix},
		{name: "reserved citation", prefix: "CIT", wantErr: ErrReservedPrefix},
		{name: "reserved observation", prefix: "OBS", wantErr: ErrReservedPrefix},
		{name: "reserved lowercase", prefix: "src", wantErr: ErrReservedPrefix},
		// The candidate marker can never be a prefix: prefixes are three letters.
		{name: "candidate marker", prefix: candidateMarker, wantErr: ErrInvalidPrefix},
		{name: "blank", prefix: "  ", wantErr: ErrInvalidPrefix},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidatePrefix(tt.prefix)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("got %v want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}
