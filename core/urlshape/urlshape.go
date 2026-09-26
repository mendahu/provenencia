// Package urlshape checks host-shaped web addresses for catalog values.
// It does not fetch or resolve the address.
package urlshape

import (
	"errors"
	"net/url"
	"strings"
)

// ErrInvalid is a malformed or disallowed address (no host, bad scheme, spaces).
var ErrInvalid = errors.New("urlshape: invalid")

// Validate accepts http/https with a host, or a scheme-less host
// (www.url.com). Scheme-less values are checked as https:// plus the typed
// string. file, javascript, spaces, and strings with no host fail.
func Validate(s string) error {
	_, err := parse(s)
	return err
}

// OpenHref returns a browser-ready href. Scheme-less input gets an https://
// prefix. The stored catalog value stays whatever the researcher typed.
func OpenHref(s string) (string, error) {
	typed, err := parse(s)
	if err != nil {
		return "", err
	}
	if !hasScheme(typed) {
		return "https://" + typed, nil
	}
	return typed, nil
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
	if u.Hostname() == "" {
		return ErrInvalid
	}
	return nil
}

func hasScheme(s string) bool {
	u, err := url.Parse(strings.TrimSpace(s))
	return err == nil && u.Scheme != ""
}
