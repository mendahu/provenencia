package database

// NullBlob binds an empty BLOB as SQL NULL, so an optional blob column such as
// a foreign-key id stores NULL rather than a zero-length value.
func NullBlob(b []byte) any {
	if len(b) == 0 {
		return nil
	}
	return b
}
