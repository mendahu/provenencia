//go:build !cgo

package database

// Stubs for platforms without CGO (e.g. js/wasm). go-sqlite3 does not define
// sqlite3.Error when CGO is disabled; real catalogs always build with CGO.

func IsUniqueConflict(error) bool      { return false }
func IsBusyOrLocked(error) bool        { return false }
func IsConstraint(error) bool          { return false }
func IsConstraintViolation(error) bool { return false }
