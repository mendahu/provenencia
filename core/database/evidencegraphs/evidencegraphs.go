// Package evidencegraphs owns Evidence graphs and their Artifact links.
package evidencegraphs

import (
	"bytes"
	"database/sql"
	"errors"
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

	sqlSubjectGraph = `SELECT graph_id FROM subjects WHERE id = ?`
	sqlIsLinked     = `SELECT 1 FROM evidence_graph_artifacts WHERE graph_id = ? AND artifact_id = ?`
	sqlLinkSources  = `SELECT g.source_id, a.source_id FROM evidence_graphs g, artifacts a
		WHERE g.id = ? AND a.id = ?`
	sqlLinkInsert = `INSERT INTO evidence_graph_artifacts (graph_id, artifact_id) VALUES (?, ?)`
	sqlLinkDelete = `DELETE FROM evidence_graph_artifacts WHERE graph_id = ? AND artifact_id = ?`
	sqlLinkCount  = `SELECT COUNT(*) FROM evidence_graph_artifacts WHERE graph_id = ?`
	// The ref is selected for the order and dropped.
	sqlCitationsUsing = `SELECT DISTINCT c.id, c.ref FROM citations c
		JOIN observations o ON o.citation_id = c.id
		JOIN subjects s ON s.id = o.subject_id
		WHERE c.artifact_id = ? AND s.graph_id = ?
		ORDER BY c.ref COLLATE NOCASE`
	sqlMissingRef = `SELECT id FROM evidence_graphs WHERE ref IS NULL`
	sqlSetRef     = `UPDATE evidence_graphs SET ref = ? WHERE id = ?`

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

// SubjectGraphTx returns the graph a Subject belongs to.
func SubjectGraphTx(tx *sql.Tx, subjectID []byte) ([]byte, error) {
	if tx == nil || len(subjectID) != 16 {
		return nil, ErrInvalid
	}
	var id []byte
	err := tx.QueryRow(sqlSubjectGraph, subjectID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInvalid
	}
	if err != nil {
		return nil, err
	}
	return id, nil
}

// IsLinkedTx reports whether the graph and the Artifact are linked.
func IsLinkedTx(tx *sql.Tx, graphID, artifactID []byte) (bool, error) {
	if tx == nil || len(graphID) != 16 || len(artifactID) != 16 {
		return false, ErrInvalid
	}
	var one int
	err := tx.QueryRow(sqlIsLinked, graphID, artifactID).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// CitationsUsingArtifactTx returns the ids of the Citations on the Artifact
// that have at least one Observation on a Subject of the graph, ordered by ref.
func CitationsUsingArtifactTx(tx *sql.Tx, graphID, artifactID []byte) ([][]byte, error) {
	if tx == nil || len(graphID) != 16 || len(artifactID) != 16 {
		return nil, ErrInvalid
	}
	rows, err := tx.Query(sqlCitationsUsing, artifactID, graphID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out [][]byte
	for rows.Next() {
		var id []byte
		var citationRef string
		if err := rows.Scan(&id, &citationRef); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// checkLink validates the shared input of Link and Unlink: the graph and the
// Artifact exist and belong to the same Source.
func checkLink(tx *database.Tx, userID, graphID, artifactID []byte) error {
	if tx == nil || len(graphID) != 16 || len(artifactID) != 16 {
		return ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return err
	}
	var graphSource, artifactSource []byte
	err := tx.QueryRow(sqlLinkSources, graphID, artifactID).Scan(&graphSource, &artifactSource)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrInvalid
	}
	if err != nil {
		return err
	}
	if !bytes.Equal(graphSource, artifactSource) {
		return ErrInvalid
	}
	return nil
}

func linkFields(graphID, artifactID []byte) map[string]any {
	return map[string]any{
		"graph_id":    uuidJSON(graphID),
		"artifact_id": uuidJSON(artifactID),
	}
}

// Link links an Artifact to a graph of its Source. An existing link returns no
// changes.
func Link(tx *database.Tx, userID, graphID, artifactID []byte) ([]rowchange.Change, error) {
	if err := checkLink(tx, userID, graphID, artifactID); err != nil {
		return nil, err
	}
	linked, err := IsLinkedTx(tx.Tx, graphID, artifactID)
	if err != nil {
		return nil, err
	}
	if linked {
		return nil, nil
	}
	if _, err := tx.Exec(sqlLinkInsert, graphID, artifactID); err != nil {
		return nil, mapConstraint(err)
	}
	return []rowchange.Change{{
		EntityType: "evidence_graph_artifact",
		EntityID:   append([]byte(nil), graphID...),
		Action:     rowchange.ActionCreate,
		Fields:     rowchange.FullRow(linkFields(graphID, artifactID)),
	}}, nil
}

// Unlink removes an Artifact from a graph. A link that is not there returns no
// changes. The graph's Citations keep the Artifact linked (ErrArtifactInUse),
// and a graph keeps at least one Artifact (ErrLastArtifact).
func Unlink(tx *database.Tx, userID, graphID, artifactID []byte) ([]rowchange.Change, error) {
	if err := checkLink(tx, userID, graphID, artifactID); err != nil {
		return nil, err
	}
	linked, err := IsLinkedTx(tx.Tx, graphID, artifactID)
	if err != nil {
		return nil, err
	}
	if !linked {
		return nil, nil
	}
	using, err := CitationsUsingArtifactTx(tx.Tx, graphID, artifactID)
	if err != nil {
		return nil, err
	}
	if len(using) > 0 {
		return nil, ErrArtifactInUse
	}
	var n int
	if err := tx.QueryRow(sqlLinkCount, graphID).Scan(&n); err != nil {
		return nil, err
	}
	if n == 1 {
		return nil, ErrLastArtifact
	}
	if _, err := tx.Exec(sqlLinkDelete, graphID, artifactID); err != nil {
		return nil, err
	}
	return []rowchange.Change{{
		EntityType: "evidence_graph_artifact",
		EntityID:   append([]byte(nil), graphID...),
		Action:     rowchange.ActionDelete,
		Fields:     rowchange.DeletedRow(linkFields(graphID, artifactID)),
	}}, nil
}

// LinkArtifact links a new Artifact to the given graphs of its Source; no ids
// means every graph of the Source. A Source with no graph (only direct-SQL
// fixtures) gets no link.
func LinkArtifact(tx *database.Tx, userID, sourceID, artifactID []byte, graphIDs [][]byte) ([]rowchange.Change, error) {
	if tx == nil || len(sourceID) != 16 || len(artifactID) != 16 {
		return nil, ErrInvalid
	}
	if len(graphIDs) == 0 {
		var err error
		graphIDs, err = graphIDsTx(tx.Tx, sourceID)
		if err != nil {
			return nil, err
		}
	} else {
		seen := map[string]bool{}
		for _, id := range graphIDs {
			if len(id) != 16 || seen[string(id)] {
				return nil, ErrInvalid
			}
			seen[string(id)] = true
		}
	}
	var changes []rowchange.Change
	for _, id := range graphIDs {
		ch, err := Link(tx, userID, id, artifactID)
		if err != nil {
			return nil, err
		}
		changes = append(changes, ch...)
	}
	return changes, nil
}

// EnsureRefs mints GPH-… refs for any evidence_graphs row missing one: the
// graphs the format migration backfilled. Unaudited, like users.EnsureRefs.
func EnsureRefs(c *database.Catalog) error {
	db, err := c.DB()
	if err != nil {
		return err
	}
	rows, err := db.Query(sqlMissingRef)
	if err != nil {
		return err
	}
	defer rows.Close()
	var need [][]byte
	for rows.Next() {
		var id []byte
		if err := rows.Scan(&id); err != nil {
			return err
		}
		need = append(need, id)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, id := range need {
		r, err := ref.Mint(ref.PrefixEvidenceGraph)
		if err != nil {
			return err
		}
		if _, err := db.Exec(sqlSetRef, r, id); err != nil {
			return err
		}
	}
	return nil
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
