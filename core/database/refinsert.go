package database

import (
	"errors"

	"github.com/mendahu/provenencia/core/ref"
)

// MaxRefRetries is how many refs InsertWithRef tries before giving up. A ref
// is a random token, so a collision is rare and a second try almost always wins.
const MaxRefRetries = 8

// ErrRefsExhausted means every ref InsertWithRef minted collided. A package
// maps it to its own invalid-input error.
var ErrRefsExhausted = errors.New("database: ref retries exhausted")

// Minter returns a function that mints a fresh ref with the prefix. A package
// keeps it as a var when a test needs to force a collision.
func Minter(prefix string) func() (string, error) {
	return func() (string, error) { return ref.Mint(prefix) }
}

// InsertWithRef mints a ref and calls insert with it, minting again when
// insert reports a UNIQUE conflict, up to MaxRefRetries times. It returns the
// ref that was stored. A mint error or an insert error that is not a UNIQUE
// conflict returns at once, unchanged; running out of tries returns
// ErrRefsExhausted.
func InsertWithRef(mint func() (string, error), insert func(ref string) error) (string, error) {
	for attempt := 0; attempt < MaxRefRetries; attempt++ {
		r, err := mint()
		if err != nil {
			return "", err
		}
		err = insert(r)
		if err == nil {
			return r, nil
		}
		if !IsUniqueConflict(err) {
			return "", err
		}
	}
	return "", ErrRefsExhausted
}
