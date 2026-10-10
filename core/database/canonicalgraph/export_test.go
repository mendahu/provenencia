package canonicalgraph

import (
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/hops"
)

// WalkSQL is the query Walk runs for n handles, with its arguments.
func WalkSQL(h hops.Hop, ids [][]byte) (string, []any) {
	query, args := hopQuery(h)
	return query + database.SQLInPlaceholders(len(ids)) + `)`, append(args, database.BlobArgs(ids)...)
}

// WalkSourceSQL is the query WalkSource runs, with its arguments.
func WalkSourceSQL(h hops.Hop, sourceID []byte, ids [][]byte) (string, []any) {
	query, args := sourceQuery(h, sourceID)
	return query + database.SQLInPlaceholders(len(ids)) + `)`, append(args, database.BlobArgs(ids)...)
}
