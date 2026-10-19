package evidencegraphs

// SetMintRef replaces the ref source and returns a function that restores it.
func SetMintRef(f func() (string, error)) (restore func()) {
	old := mintRef
	mintRef = f
	return func() { mintRef = old }
}
