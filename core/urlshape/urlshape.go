// Package urlshape checks host-shaped web addresses for catalog values.
// It does not fetch or resolve the address.
package urlshape

import (
	"errors"
	"net"
	"net/url"
	"strings"
)

// ErrInvalid is a malformed or disallowed address (no host, bad scheme, spaces).
var ErrInvalid = errors.New("urlshape: invalid")

// Validate accepts http/https with a host, or a scheme-less host
// (www.url.com). Scheme-less values are checked as https:// plus the typed
// string. file, javascript, spaces, and strings with no host fail.
func Validate(s string) error {
	_, err := Canonical(s)
	return err
}

// Canonical returns the lowercase form used for catalog storage. Mixed-case
// input is accepted when the shape is valid.
func Canonical(s string) (string, error) {
	typed, err := parse(s)
	if err != nil {
		return "", err
	}
	return strings.ToLower(typed), nil
}

// OpenHref returns a browser-ready href. Scheme-less input gets an https://
// prefix. The href is lowercased to match stored catalog values.
func OpenHref(s string) (string, error) {
	canon, err := Canonical(s)
	if err != nil {
		return "", err
	}
	if !hasScheme(canon) {
		return "https://" + canon, nil
	}
	return canon, nil
}

func parse(s string) (string, error) {
	typed := strings.TrimSpace(s)
	if typed == "" || strings.ContainsAny(typed, " \t\n\r") {
		return "", ErrInvalid
	}
	if u, err := url.Parse(typed); err == nil && u.Scheme != "" {
		if err := checkHTTP(u); err != nil {
			return "", err
		}
		return typed, nil
	}
	u, err := url.Parse("https://" + typed)
	if err != nil {
		return "", ErrInvalid
	}
	if err := checkHTTP(u); err != nil {
		return "", err
	}
	return typed, nil
}

func checkHTTP(u *url.URL) error {
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return ErrInvalid
	}
	if !hostShaped(u.Hostname()) {
		return ErrInvalid
	}
	return nil
}

// hostShaped accepts an IP or a dotted name (www.url.com). A single label
// (jakerobins, localhost) is not enough — url.Parse treats those as hosts.
func hostShaped(host string) bool {
	if host == "" {
		return false
	}
	if ip := net.ParseIP(host); ip != nil {
		return true
	}
	labels := strings.Split(host, ".")
	if len(labels) < 2 {
		return false
	}
	for _, label := range labels {
		if label == "" {
			return false
		}
	}
	return true
}

func hasScheme(s string) bool {
	u, err := url.Parse(strings.TrimSpace(s))
	return err == nil && u.Scheme != ""
}
