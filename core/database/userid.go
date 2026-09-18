package database

// RequireUserID rejects empty or non-16-byte user IDs so audited mutations
// always carry attribution. invalid is the caller's domain ErrInvalid (or
// equivalent); this helper does not invent a shared error code.
func RequireUserID(userID []byte, invalid error) error {
	if len(userID) != 16 {
		return invalid
	}
	return nil
}
