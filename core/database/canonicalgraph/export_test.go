package canonicalgraph

import "github.com/mendahu/provenencia/core/database"

// WalkSQL is the query Walk runs for n handles, with its arguments.
func WalkSQL(h Hop, ids [][]byte) (string, []any) {
	query, args := h.query()
	return query + database.SQLInPlaceholders(len(ids)) + `)`, append(args, database.BlobArgs(ids)...)
}

// WalkSourceSQL is the query WalkSource runs, with its arguments.
func WalkSourceSQL(h Hop, sourceID []byte, ids [][]byte) (string, []any) {
	query, args := h.sourceQuery(sourceID)
	return query + database.SQLInPlaceholders(len(ids)) + `)`, append(args, database.BlobArgs(ids)...)
}
