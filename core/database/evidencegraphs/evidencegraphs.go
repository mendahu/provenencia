// Package evidencegraphs owns Evidence graphs and their Artifact links.
package evidencegraphs

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/ref"
)

// DefaultName is the name of the graph every Source starts with.
const DefaultName = "Default"

var (
	ErrInvalid       = apperr.New(apperr.CodeEvidenceGraphsInvalid, apperr.KindUser)
	ErrArtifactInUse = apperr.New(apperr.CodeEvidenceGraphsArtifactInUse, apperr.KindUser)
	ErrLastArtifact  = apperr.New(apperr.CodeEvidenceGraphsLastArtifact, apperr.KindUser)
)

const (
	sqlInsert = `INSERT INTO evidence_graphs (id, ref, source_id, name, description)
		VALUES (?, ?, ?, ?, ?)`
	sqlGraphIDs = `SELECT id FROM evidence_graphs WHERE source_id = ? ORDER BY id`

	maxRefRetries = 8
)

// Graph is one evidence_graphs row.
type Graph struct {
	ID          []byte
	Ref         string
	SourceID    []byte
	Name        string
	Description string
}

// CreateDefault inserts the Default graph of a new Source. Its only caller is
// sources.Create, which has already checked the user and the Source.
func CreateDefault(tx *database.Tx, sourceID []byte) (Graph, []rowchange.Change, error) {
	if tx == nil || len(sourceID) != 16 {
		return Graph{}, nil, ErrInvalid
	}
	g, change, err := insert(tx.Tx, sourceID, DefaultName, "")
	if err != nil {
		return Graph{}, nil, err
	}
	return g, []rowchange.Change{change}, nil
}

func insert(tx *sql.Tx, sourceID []byte, name, description string) (Graph, rowchange.Change, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return Graph{}, rowchange.Change{}, err
	}
	idBytes := id[:]

	var graphRef string
	for attempt := 0; attempt < maxRefRetries; attempt++ {
		graphRef, err = ref.Mint(ref.PrefixEvidenceGraph)
		if err != nil {
			return Graph{}, rowchange.Change{}, err
		}
		_, err = tx.Exec(sqlInsert, idBytes, graphRef, sourceID, name, nullStr(description))
		if err == nil {
			break
		}
		if !database.IsUniqueConflict(err) {
			return Graph{}, rowchange.Change{}, mapConstraint(err)
		}
	}
	if err != nil {
		return Graph{}, rowchange.Change{}, ErrInvalid
	}

	change := rowchange.Change{
		EntityType: "evidence_graph",
		EntityID:   append([]byte(nil), idBytes...),
		Action:     rowchange.ActionCreate,
		Fields: rowchange.FullRow(map[string]any{
			"id":          id.String(),
			"ref":         graphRef,
			"source_id":   uuidJSON(sourceID),
			"name":        name,
			"description": nullStr(description),
		}),
	}
	return Graph{
		ID:          append([]byte(nil), idBytes...),
		Ref:         graphRef,
		SourceID:    append([]byte(nil), sourceID...),
		Name:        name,
		Description: description,
	}, change, nil
}

// SoleGraphTx returns the id of the one graph of a Source.
//
// The one bridge between S12-03 and S12-08. S12-08 deletes it.
func SoleGraphTx(tx *sql.Tx, sourceID []byte) ([]byte, error) {
	if tx == nil || len(sourceID) != 16 {
		return nil, ErrInvalid
	}
	ids, err := graphIDsTx(tx, sourceID)
	if err != nil {
		return nil, err
	}
	switch len(ids) {
	case 0:
		return nil, ErrInvalid
	case 1:
		return ids[0], nil
	default:
		return nil, fmt.Errorf("evidencegraphs: source %x has more than one graph", sourceID)
	}
}

// graphIDsTx is the one per-Source graph-id read.
func graphIDsTx(tx *sql.Tx, sourceID []byte) ([][]byte, error) {
	rows, err := tx.Query(sqlGraphIDs, sourceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out [][]byte
	for rows.Next() {
		var id []byte
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func uuidJSON(id []byte) any {
	u, err := uuid.FromBytes(id)
	if err != nil {
		return nil
	}
	return u.String()
}

func mapConstraint(err error) error {
	if database.IsConstraintViolation(err) {
		return ErrInvalid
	}
	return err
}
