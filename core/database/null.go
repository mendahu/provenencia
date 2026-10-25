package database

// NullBlob binds an empty BLOB as SQL NULL, so an optional blob column such as
// a foreign-key id stores NULL rather than a zero-length value.
func NullBlob(b []byte) any {
	if len(b) == 0 {
		return nil
	}
	return b
}

// NullString binds an empty string as SQL NULL, so an optional text column
// stores NULL rather than "".
func NullString(s string) any {
	if s == "" {
		return nil
	}
	return s
}
