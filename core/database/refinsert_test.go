//go:build cgo

package database

import (
	"errors"
	"testing"

	"github.com/mattn/go-sqlite3"
)

func TestInsertWithRef(t *testing.T) {
	conflict := sqlite3.Error{Code: sqlite3.ErrConstraint, ExtendedCode: sqlite3.ErrConstraintUnique}
	var minted int
	mint := func() (string, error) {
		minted++
		return string(rune('a' + minted - 1)), nil
	}

	t.Run("retries a unique conflict with a fresh ref", func(t *testing.T) {
		minted = 0
		calls := 0
		got, err := InsertWithRef(mint, func(r string) error {
			calls++
			if calls <= 2 {
				return conflict
			}
			return nil
		})
		if err != nil || got != "c" || calls != 3 {
			t.Fatalf("got %q, %v after %d inserts", got, err, calls)
		}
	})
	t.Run("gives up after MaxRefRetries", func(t *testing.T) {
		minted = 0
		calls := 0
		_, err := InsertWithRef(mint, func(string) error { calls++; return conflict })
		if !errors.Is(err, ErrRefsExhausted) || calls != MaxRefRetries {
			t.Fatalf("err %v after %d inserts, want ErrRefsExhausted after %d", err, calls, MaxRefRetries)
		}
	})
	t.Run("returns another insert error at once, unchanged", func(t *testing.T) {
		minted = 0
		boom := errors.New("boom")
		calls := 0
		_, err := InsertWithRef(mint, func(string) error { calls++; return boom })
		if err != boom || calls != 1 {
			t.Fatalf("err %v after %d inserts", err, calls)
		}
	})
	t.Run("returns a mint error at once, unchanged", func(t *testing.T) {
		boom := errors.New("no entropy")
		calls := 0
		_, err := InsertWithRef(func() (string, error) { return "", boom }, func(string) error { calls++; return nil })
		if err != boom || calls != 0 {
			t.Fatalf("err %v after %d inserts", err, calls)
		}
	})
}

func TestUUIDHelpers(t *testing.T) {
	id := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	const want = "01020304-0506-0708-090a-0b0c0d0e0f10"
	if got := UUIDString(id); got != want {
		t.Fatalf("UUIDString = %q", got)
	}
	if got := UUIDJSON(id); got != want {
		t.Fatalf("UUIDJSON = %v", got)
	}
	for _, bad := range [][]byte{nil, {}, {1, 2, 3}, make([]byte, 17)} {
		if UUIDString(bad) != "" || UUIDJSON(bad) != nil {
			t.Fatalf("%v: UUIDString %q, UUIDJSON %v", bad, UUIDString(bad), UUIDJSON(bad))
		}
	}
}

func TestNullString(t *testing.T) {
	if NullString("") != nil || NullString("x") != "x" || NullString(" ") != " " {
		t.Fatal("NullString binds only the empty string as NULL")
	}
}
