package subjects

import (
	"database/sql"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/sqlbatch"
)

// SourceOf is the Source a Subject is on: its id and title.
type SourceOf struct {
	SourceID []byte
	Title    string
}

// SourcesOf reads the Source of each Subject in one query per batch of ids,
// keyed by Subject id as a string. A Subject that does not exist has no entry.
func SourcesOf(q sqlbatch.Querier, ids [][]byte) (map[string]SourceOf, error) {
	out := map[string]SourceOf{}
	err := sqlbatch.Query(q, `SELECT s.id, s.source_id, src.title FROM subjects s
		JOIN sources src ON src.id = s.source_id WHERE s.id IN (`, `)`, nil, database.UniqueBlobIDs(ids),
		func(rows *sql.Rows) error {
			var id, sourceID []byte
			var title string
			if err := rows.Scan(&id, &sourceID, &title); err != nil {
				return err
			}
			out[string(id)] = SourceOf{SourceID: sourceID, Title: title}
			return nil
		})
	if err != nil {
		return nil, err
	}
	return out, nil
}
