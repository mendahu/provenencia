package database

import "github.com/google/uuid"

// UUIDString is the canonical text of a 16-byte UUID id, or "" for any other
// length.
func UUIDString(id []byte) string {
	u, err := uuid.FromBytes(id)
	if err != nil {
		return ""
	}
	return u.String()
}

// UUIDJSON is UUIDString for a rowchange field: nil unless id is a 16-byte
// UUID, so an unset id is a JSON null.
func UUIDJSON(id []byte) any {
	if s := UUIDString(id); s != "" {
		return s
	}
	return nil
}
