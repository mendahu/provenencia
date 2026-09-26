package urlshape

import (
	"errors"
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{name: "https", in: "https://example.com"},
		{name: "mixed case", in: "HTTPS://Example.COM/Record"},
		{name: "http path", in: "http://example.com/record"},
		{name: "scheme-less www", in: "www.url.com"},
		{name: "scheme-less host path", in: "archives.norfolk.gov.uk/catalogue/PD28-47"},
		{name: "ipv4", in: "http://192.168.1.10/record"},
		{name: "blank", in: "  ", wantErr: true},
		{name: "spaces", in: "https://example.com/a b", wantErr: true},
		{name: "typo scheme", in: "htps://example.com", wantErr: true},
		{name: "file", in: "file:/tmp", wantErr: true},
		{name: "javascript", in: "javascript:alert(1)", wantErr: true},
		{name: "no host", in: "https://", wantErr: true},
		{name: "junk", in: "not a url", wantErr: true},
		{name: "single label", in: "jakerobins", wantErr: true},
		{name: "https single label", in: "https://jakerobins", wantErr: true},
		{name: "empty label", in: "jakerobins.", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.in)
			if tt.wantErr {
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("Validate(%q) = %v, want ErrInvalid", tt.in, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Validate(%q) = %v", tt.in, err)
			}
		})
	}
}

func TestOpenHref(t *testing.T) {
	href, err := OpenHref("www.url.com")
	if err != nil {
		t.Fatal(err)
	}
	if href != "https://www.url.com" {
		t.Fatalf("OpenHref = %q", href)
	}
	href, err = OpenHref("https://example.com")
	if err != nil {
		t.Fatal(err)
	}
	if href != "https://example.com" {
		t.Fatalf("OpenHref https = %q", href)
	}
	if _, err := OpenHref("file:/tmp"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("OpenHref file = %v", err)
	}
	href, err = OpenHref("HTTPS://Example.COM/Record")
	if err != nil {
		t.Fatal(err)
	}
	if href != "https://example.com/record" {
		t.Fatalf("OpenHref mixed case = %q", href)
	}
}

func TestCanonicalLowercases(t *testing.T) {
	got, err := Canonical("HTTPS://Example.COM/Record")
	if err != nil {
		t.Fatal(err)
	}
	if got != "https://example.com/record" {
		t.Fatalf("Canonical = %q", got)
	}
	got, err = Canonical("WWW.URL.COM")
	if err != nil {
		t.Fatal(err)
	}
	if got != "www.url.com" {
		t.Fatalf("Canonical scheme-less = %q", got)
	}
}
